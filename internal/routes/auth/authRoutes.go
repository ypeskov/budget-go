package auth

import (
	"fmt"
	"net/http"
	"strings"

	"ypeskov/budget-go/internal/middleware"
	"ypeskov/budget-go/internal/utils"

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

	signedToken, err := utils.GenerateAccessToken(user, cfg)
	if err != nil {
		log.Error("Error generating token: ", err)
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

	// Accept both Authorization: Bearer and legacy auth-token header
	var token string
	if authz := c.Request().Header.Get("Authorization"); authz != "" {
		if parts := strings.SplitN(authz, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token = parts[1]
		}
	}
	if token == "" {
		token = c.Request().Header.Get("auth-token")
	}

	claims, err := middleware.GetUserFromToken(token, cfg)
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
	fmt.Printf("user %+v\n", user)

	// Get user's actual base currency
	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		log.Error("Error getting user base currency: ", err)
		// Fallback to USD if we can't get the base currency
		baseCurrency.Code = "USD"
	}

	// Get user settings including language
	settings := map[string]string{}
	userSettings, err := sm.UserSettingsService.GetUserSettings(user.ID)
	if err != nil {
		log.Error("Error getting user settings: ", err)
		// Fallback to default language if we can't get settings
		settings["language"] = "en"
	} else {
		// Extract language from settings, default to "en" if not found
		if lang, ok := userSettings.Settings["language"].(string); ok {
			settings["language"] = lang
		} else {
			settings["language"] = "en"
		}
	}

	return c.JSON(http.StatusOK, ProfileDTO{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Settings:  settings,
		BaseCurrency: baseCurrency.Code,
	})
}

func comparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
