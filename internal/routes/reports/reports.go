package reports

import (
	"github.com/labstack/echo/v4"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/services"
)

var (
	cfg *config.Config
	sm  *services.Manager
)

func RegisterReportsRoutes(g *echo.Group, cfgGlobal *config.Config, manager *services.Manager) {
	cfg = cfgGlobal
	sm = manager

	g.POST("/balance/non-hidden", GetNonHiddenBalance)
}

func GetNonHiddenBalance(c echo.Context) error {
	// {"accountId": 22, "accountName": "BGN Cash", "currencyCode": "BGN", "balance": 2510.0, "baseCurrencyBalance": 1316.0079574964034, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}, {"accountId": 41, "accountName": "DSK credit card", "currencyCode": "BGN", "balance": -992.19, "baseCurrencyBalance": -520.2111296208592, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}, {"accountId": 17, "accountName": "DSK Main", "currencyCode": "BGN", "balance": 15356.62, "baseCurrencyBalance": 8051.567378585027, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}, {"accountId": 18, "accountName": "DSK Virtual", "currencyCode": "BGN", "balance": 74.42, "baseCurrencyBalance": 39.018849480829616, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}, {"accountId": 42, "accountName": "DSK zalog", "currencyCode": "BGN", "balance": 1200.0, "baseCurrencyBalance": 629.1671509942964, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}, {"accountId": 23, "accountName": "EUR Cash", "currencyCode": "EUR", "balance": 3000.0, "baseCurrencyBalance": 3076.359961064358, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}, {"accountId": 19, "accountName": "Monobank", "currencyCode": "UAH", "balance": 707.5, "baseCurrencyBalance": 16.72284799831154, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}, {"accountId": 38, "accountName": "\u0412\u0430\u0443\u0447\u0435\u0440\u044b \u0437\u0430 \u0445\u0440\u0430\u043d\u0430", "currencyCode": "BGN", "balance": 0.0, "baseCurrencyBalance": 0.0, "baseCurrencyCode": "USD", "reportDate": "2025-01-12"}
	type Accounts []struct {
		AccountID           int     `json:"accountId"`
		AccountName         string  `json:"accountName"`
		CurrencyCode        string  `json:"currencyCode"`
		Balance             float64 `json:"balance"`
		BaseCurrencyBalance float64 `json:"baseCurrencyBalance"`
		BaseCurrencyCode    string  `json:"baseCurrencyCode"`
		ReportDate          string  `json:"reportDate"`
	}

	accounts := Accounts{
		{
			AccountID:           22,
			AccountName:         "BGN Cash",
			CurrencyCode:        "BGN",
			Balance:             2510.0,
			BaseCurrencyBalance: 1316.0079574964034,
			BaseCurrencyCode:    "USD",
			ReportDate:          "2025-01-12",
		},
		{
			AccountID:           18,
			AccountName:         "USD cash",
			CurrencyCode:        "USD",
			Balance:             3000.0,
			BaseCurrencyBalance: 1316.0079574964034,
			BaseCurrencyCode:    "USD",
			ReportDate:          "2025-01-12",
		},
	}

	return c.JSON(200, accounts)
}
