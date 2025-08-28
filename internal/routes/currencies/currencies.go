package currencies

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/services"
)

var (
	sm *services.Manager
)

func RegisterCurrenciesRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("", GetCurrencies)
}

func GetCurrencies(c echo.Context) error {
	logger.Debug("GetCurrencies request started", "method", c.Request().Method, "url", c.Request().URL)

	currencies, err := sm.CurrenciesService.GetCurrencies()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	var currenciesResponse []models.Currency
	for _, currency := range currencies {
		currenciesResponse = append(currenciesResponse, currency)
	}

	logger.Debug("GetCurrencies request completed")
	return c.JSON(http.StatusOK, currenciesResponse)
}
