package auth

import (
	"net/http"
	"time"
	"ypeskov/budget-go/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/services"
)

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JWTCustomClaims struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var (
	cfg *config.Config
	sm  *services.Manager
)

func RegisterAuthRoutes(g *echo.Group, cfgGlobal *config.Config, manager *services.Manager) {
	cfg = cfgGlobal
	sm = manager

	g.POST("/login", Login)
	g.GET("/profile", Profile)
}

func Login(c echo.Context) error {
	u := new(UserLogin)
	if err := c.Bind(u); err != nil {
		log.Error(err)
		return c.String(http.StatusBadRequest, "Bad request")
	}

	user, err := sm.UserService.GetUserByEmail(u.Email)
	if err != nil {
		log.Error("Error getting user by email: ", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	if !comparePassword(user.PasswordHash, u.Password) {
		log.Error("Passwords do not match")
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	claims := &JWTCustomClaims{
		Id:    user.ID,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if token == nil {
		log.Error("Error creating token")
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	signedToken, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		log.Error("Error signing token: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"accessToken": signedToken,
		"tokenType":   "Bearer",
	})
}

type ProfileDTO struct {
	ID           int               `json:"id"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Email        string            `json:"email"`
	Settings     map[string]string `json:"settings"`
	BaseCurrency string            `json:"baseCurrency"`
}

func Profile(c echo.Context) error {
	cfg := c.Get("config").(*config.Config)

	claims, err := middleware.GetUserFromToken(c.Request().Header.Get("auth-token"), cfg)
	if err != nil || claims == nil {
		log.Error("Failed to cast user to jwt.MapClaims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing user")
	}

	email, emailOk := claims["email"].(string)
	if !emailOk {
		log.Error("Email not found in claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}

	user, err := sm.UserService.GetUserByEmail(email)
	if err != nil {
		log.Error("Error getting user by email: ", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	return c.JSON(http.StatusOK, ProfileDTO{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Settings: map[string]string{
			"language": "uk",
		},
		BaseCurrency: "USD",
	})
}

func comparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
