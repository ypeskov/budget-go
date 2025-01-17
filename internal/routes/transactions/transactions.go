package transactions

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/services"
)

var (
	sm *services.Manager
)

func RegisterTransactionsRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("", GetTransactions)
}

func GetTransactions(c echo.Context) error {
	log.Debug("GetTransactions Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "User not found")
	}

	perPage, err := getPerPage(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid per_page value")
	}

	page, err := getPage(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page value")
	}

	accountIds, err := getAccountIds(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid account_ids value")
	}

	fromDate, err := getFromDate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid from_date value")
	}

	toDate, err := getToDate(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid to_date value")
	}

	currencies, err := getCurrencies(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid currencies value")
	}

	fmt.Println("accountIds", accountIds)
	fmt.Println("fromDate", fromDate)
	fmt.Println("toDate", toDate)
	fmt.Println("currencies", currencies)

	transactions, err := sm.TransactionsService.GetTransactionsWithAccounts(user.ID, sm, perPage, page, accountIds)
	if err != nil {
		log.Error("Error getting transactions: ", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		log.Error("Error getting base currency: ", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	var transactionsDTO []dto.ResponseTransactionDTO
	for _, transaction := range transactions {
		transactionsDTO = append(transactionsDTO, dto.TransactionWithAccountToResponseTransactionDTO(transaction, baseCurrency))
	}

	return c.JSON(http.StatusOK, transactionsDTO)
}
