package currencies

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

func RegisterCurrenciesRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("", GetCurrencies)
}

func GetCurrencies(c echo.Context) error {
	log.Debugf("GetCurrencies request started: %s %s", c.Request().Method, c.Request().URL)
	
	currencies, err := sm.CurrenciesService.GetCurrencies()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	var currenciesResponse []models.Currency
	for _, currency := range currencies {
		currenciesResponse = append(currenciesResponse, currency)
	}
	
	log.Debug("GetCurrencies request completed - GET /currencies")
	return c.JSON(http.StatusOK, currenciesResponse)
}
