package userSettings

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/dto"
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
	userRaw := c.Get("user")

	claims, ok := userRaw.(jwt.MapClaims)
	if !ok || claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing user")
	}

	userId := int(claims["id"].(float64))

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(userId)
	if err != nil {
		log.Error("Failed to get base currency: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get base currency")
	}

	return c.JSON(http.StatusOK, dto.BaseCurrencyDTO{
		ID:   baseCurrency.ID,
		Code: baseCurrency.Code,
		Name: baseCurrency.Name,
	})
}
