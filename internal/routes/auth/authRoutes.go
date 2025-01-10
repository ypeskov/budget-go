package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/services"
)

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type jwtCustomClaims struct {
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

	claims := &jwtCustomClaims{
		Id: user.ID,
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
		"token": signedToken,
	})
}

func comparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
