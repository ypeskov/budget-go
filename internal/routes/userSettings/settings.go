package userSettings

import (
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/services"
)

var (
	sm *services.Manager
)

func RegisterSettingsRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("/base-currency", GetBaseCurrency)
}

func GetBaseCurrency(c echo.Context) error {
	log.Debug("GetBaseCurrency Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "User not found")
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		log.Error("Failed to get base currency: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get base currency")
	}

	return c.JSON(http.StatusOK, baseCurrency)
}
