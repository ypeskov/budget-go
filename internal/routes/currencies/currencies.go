package currencies

import (
	"net/http"

	"github.com/labstack/echo/v4"

	settings "ypeskov/budget-go/internal/routes/userSettings"
	"ypeskov/budget-go/internal/services"
)

var (
	cm *services.Manager
)

func RegisterCurrenciesRoutes(g *echo.Group, manager *services.Manager) {
	cm = manager

	g.GET("", GetCurrencies)
}

func GetCurrencies(c echo.Context) error {
	currencies, err := cm.CurrenciesService.GetCurrencies()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	var currenciesResponse []settings.BaseCurrencyDTO
	for _, currency := range currencies {
		currenciesResponse = append(currenciesResponse, settings.BaseCurrencyDTO{
			ID:   currency.ID,
			Code: currency.Code,
			Name: currency.Name,
		})
	}
	return c.JSON(http.StatusOK, currenciesResponse)
}
