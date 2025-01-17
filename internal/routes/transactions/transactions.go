package transactions

import (
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/routes/routeErrors"
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
		return logAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	// parseTransactionFilters extracts and validates query parameters for transaction filtering.
	filters, err := parseTransactionFilters(c)
	if err != nil {
		return logAndReturnError(c, err, http.StatusBadRequest)
	}

	transactions, err := sm.TransactionsService.GetTransactionsWithAccounts(
		user.ID,
		sm,
		filters.PerPage,
		filters.Page,
		filters.AccountIds,
		filters.FromDate,
		filters.ToDate,
		filters.TransactionTypes,
	)
	if err != nil {
		return logAndReturnError(c, err, http.StatusInternalServerError)
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		return logAndReturnError(c, err, http.StatusInternalServerError)
	}
	transactionsDTO := convertTransactionsToDTO(transactions, baseCurrency)

	return c.JSON(http.StatusOK, transactionsDTO)
}

func convertTransactionsToDTO(transactions []dto.TransactionWithAccount, baseCurrency models.Currency) []dto.ResponseTransactionDTO {
	var transactionsDTO []dto.ResponseTransactionDTO
	for _, transaction := range transactions {
		transactionsDTO = append(transactionsDTO, dto.TransactionWithAccountToResponseTransactionDTO(transaction, baseCurrency))
	}
	return transactionsDTO
}

func logAndReturnError(c echo.Context, err error, httpStatus int) error {
	return c.JSON(httpStatus, map[string]string{
		"error": err.Error(),
	})
}
