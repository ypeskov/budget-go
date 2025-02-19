package transactions

import (
	"net/http"
	"strconv"
	"strings"

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
	g.GET("/templates", GetTemplates)
	g.DELETE("/templates", DeleteTemplates)
	g.POST("", CreateTransaction)
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
		filters.CategoryIds,
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

func GetTemplates(c echo.Context) error {
	log.Debug("GetTemplates Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return logAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	templateDTOs, err := sm.TransactionsService.GetTemplates(user.ID)
	if err != nil {
		return logAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, templateDTOs)
}

func DeleteTemplates(c echo.Context) error {
	log.Debug("DeleteTemplates Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return logAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	templateIds := c.QueryParam("ids")
	if templateIds == "" {
		return logAndReturnError(c, &routeErrors.BadRequestError{Message: "[ids] query parameter is required"}, http.StatusBadRequest)
	}

	templateIdsList := strings.Split(templateIds, ",")
	templateIdsInt := make([]int, len(templateIdsList))

	for i, idStr := range templateIdsList {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			return logAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid template ID format"}, http.StatusBadRequest)
		}
		templateIdsInt[i] = id
	}
	err := sm.TransactionsService.DeleteTemplates(templateIdsInt, user.ID)
	if err != nil {
		return logAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Templates deleted successfully",
	})
}

func logAndReturnError(c echo.Context, err error, httpStatus int) error {
	log.Error("Error: ", err)
	return c.JSON(httpStatus, map[string]string{
		"error": "Internal server error",
	})
}

func CreateTransaction(c echo.Context) error {
	log.Debug("CreateTransaction Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return logAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	transaction := dto.CreateTransactionDTO{
		UserID: &user.ID,
	}
	if err := c.Bind(&transaction); err != nil {
		return logAndReturnError(c, err, http.StatusBadRequest)
	}

	transactionModel := models.Transaction{
		UserID:     *transaction.UserID,
		AccountID:  transaction.AccountID,
		Amount:     transaction.Amount,
		CategoryID: transaction.CategoryID,
		Label:      transaction.Label,
	}

	err := sm.TransactionsService.CreateTransaction(transactionModel)
	if err != nil {
		return logAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Transaction created successfully",
	})
}
