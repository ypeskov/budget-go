package accounts

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/services"
)

var (
	cfg *config.Config
	sm  *services.Manager
)

func RegisterAccountsRoutes(g *echo.Group, cfgGlobal *config.Config, manager *services.Manager) {
	cfg = cfgGlobal
	sm = manager

	g.GET("", GetAccounts)
}


func GetAccounts(c echo.Context) error {
	log.Debug("GetAccounts")
	userRaw := c.Get("user")

	claims, ok := userRaw.(jwt.MapClaims)
	if !ok || claims == nil {
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

	userAccounts, err := sm.AccountsService.GetUserAccounts(user.ID)
	if err != nil {
		log.Error("Error getting user accounts: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, userAccounts)
}
