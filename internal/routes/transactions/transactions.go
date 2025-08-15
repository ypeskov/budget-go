package transactions

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/routes/routeErrors"
	"ypeskov/budget-go/internal/services"
	"ypeskov/budget-go/internal/utils"
)

var (
	sm *services.Manager
)

func RegisterTransactionsRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("", GetTransactions)
	g.GET("/:id", GetTransactionDetail)
	g.PUT("", UpdateTransaction)
	g.DELETE("/:id", DeleteTransaction)
	g.GET("/templates", GetTemplates)
	g.DELETE("/templates", DeleteTemplates)
	g.POST("", CreateTransaction)
}

func GetTransactions(c echo.Context) error {
	log.Debugf("GetTransactions request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	// ParseTransactionFilters extracts and validates query parameters for transaction filtering.
	filters, err := utils.ParseTransactionFilters(c)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusBadRequest)
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
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}
	transactionsDTO := services.ConvertTransactionsToResponseList(transactions, baseCurrency)

	log.Debug("GetTransactions request completed - GET /transactions")
	return c.JSON(http.StatusOK, transactionsDTO)
}

func GetTemplates(c echo.Context) error {
	log.Debugf("GetTemplates request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	templateDTOs, err := sm.TransactionsService.GetTemplates(user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	log.Debug("GetTemplates request completed - GET /transactions/templates")
	return c.JSON(http.StatusOK, templateDTOs)
}

func DeleteTemplates(c echo.Context) error {
	log.Debugf("DeleteTemplates request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	templateIds := c.QueryParam("ids")
	if templateIds == "" {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "[ids] query parameter is required"}, http.StatusBadRequest)
	}

	templateIdsList := strings.Split(templateIds, ",")
	templateIdsInt := make([]int, len(templateIdsList))

	for i, idStr := range templateIdsList {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid template ID format"}, http.StatusBadRequest)
		}
		templateIdsInt[i] = id
	}
	err := sm.TransactionsService.DeleteTemplates(templateIdsInt, user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	log.Debug("DeleteTemplates request completed - DELETE /transactions/templates")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Templates deleted successfully",
	})
}

func CreateTransaction(c echo.Context) error {
	log.Debugf("CreateTransaction request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	transaction := dto.CreateTransactionDTO{
		UserID: &user.ID,
	}
	if err := c.Bind(&transaction); err != nil {
		return utils.LogAndReturnError(c, err, http.StatusBadRequest)
	}

	transactionModel := models.Transaction{
		UserID:     *transaction.UserID,
		AccountID:  transaction.AccountID,
		Amount:     transaction.Amount,
		CategoryID: transaction.CategoryID,
		Label:      transaction.Label,
		IsIncome:   transaction.IsIncome,
		IsTransfer: transaction.IsTransfer,
		Notes:      transaction.Notes,
		DateTime:   transaction.DateTime,
	}

	_, err := sm.TransactionsService.CreateTransaction(transactionModel, transaction.TargetAccountID, transaction.TargetAmount)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	log.Debug("CreateTransaction request completed - POST /transactions")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Transaction created successfully",
	})
}

func GetTransactionDetail(c echo.Context) error {
	log.Debugf("GetTransactionDetail request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	transactionIdStr := c.Param("id")
	if transactionIdStr == "" {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Transaction ID is required"}, http.StatusBadRequest)
	}

	transactionId, err := strconv.Atoi(transactionIdStr)
	if err != nil {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid transaction ID format"}, http.StatusBadRequest)
	}

	transactionDetail, err := sm.TransactionsService.GetTransactionDetail(transactionId, user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	if transactionDetail == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "transaction", ID: transactionId}, http.StatusNotFound)
	}

	log.Debug("GetTransactionDetail request completed - GET /transactions/:id")
	return c.JSON(http.StatusOK, transactionDetail)
}

func UpdateTransaction(c echo.Context) error {
	log.Debugf("UpdateTransaction request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	var transactionDTO dto.PutTransactionDTO
	if err := c.Bind(&transactionDTO); err != nil {
		// Create more informative error with details
		detailedError := fmt.Errorf("failed to bind request body: %w", err)
		return utils.LogAndReturnError(c, detailedError, http.StatusBadRequest)
	}

	if transactionDTO.ID == 0 {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Transaction ID is required"}, http.StatusBadRequest)
	}

	// Check that transaction exists and belongs to user
	existingTransaction, err := sm.TransactionsService.GetTransactionDetail(transactionDTO.ID, user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	if existingTransaction == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "transaction", ID: transactionDTO.ID}, http.StatusNotFound)
	}

	// Update transaction
	err = sm.TransactionsService.UpdateTransaction(transactionDTO, user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	log.Debug("UpdateTransaction request completed - PUT /transactions")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Transaction updated successfully",
	})
}

func DeleteTransaction(c echo.Context) error {
	log.Debugf("DeleteTransaction request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	transactionIdStr := c.Param("id")
	if transactionIdStr == "" {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Transaction ID is required"}, http.StatusBadRequest)
	}

	transactionId, err := strconv.Atoi(transactionIdStr)
	if err != nil {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid transaction ID format"}, http.StatusBadRequest)
	}

	// Check that transaction exists and belongs to user before deleting
	existingTransaction, err := sm.TransactionsService.GetTransactionDetail(transactionId, user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	if existingTransaction == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "transaction", ID: transactionId}, http.StatusNotFound)
	}

	// Delete transaction
	err = sm.TransactionsService.DeleteTransaction(transactionId, user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	log.Debug("DeleteTransaction request completed - DELETE /transactions/:id")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Transaction deleted successfully",
	})
}
