package accounts

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"ypeskov/budget-go/internal/config"
	appErrors "ypeskov/budget-go/internal/errors"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/services"

	"ypeskov/budget-go/internal/logger"

	"github.com/labstack/echo/v4"
)

var (
	// cfg *config.Config
	sm *services.Manager
)

func RegisterAccountsRoutes(g *echo.Group, cfgGlobal *config.Config, manager *services.Manager) {
	// cfg = cfgGlobal
	sm = manager

	g.GET("", GetAccounts)
	g.GET("/types", GetAccountsTypes)
	g.GET("/:id", GetAccountById)
	g.POST("", CreateAccount)
	g.PUT("/:id", UpdateAccount)
}

func GetAccounts(c echo.Context) error {
	logger.Debug("GetAccounts request started", "method", c.Request().Method, "url", c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		logger.Warn("Authenticated user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var includeHidden, includeDeleted, archivedOnly bool

	if c.QueryParam("includeHidden") == "true" {
		includeHidden = true
	} else {
		includeHidden = false
	}

	if c.QueryParam("includeDeleted") == "true" {
		includeDeleted = true
	} else {
		includeDeleted = false
	}

	if c.QueryParam("archivedOnly") == "true" {
		archivedOnly = true
	} else {
		archivedOnly = false
	}

	userAccounts, err := sm.AccountsService.GetUserAccounts(user.ID, sm, includeHidden, includeDeleted, archivedOnly)
	if err != nil {
		logger.Error("Error getting user accounts: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	logger.Debug("GetAccounts request completed - GET /accounts")
	return c.JSON(http.StatusOK, userAccounts)
}

func GetAccountsTypes(c echo.Context) error {
	logger.Debug("GetAccountsTypes request started", "method", c.Request().Method, "url", c.Request().URL)
	accountTypes, err := sm.AccountsService.GetAccountTypes()
	if err != nil {
		logger.Error("Error getting account types: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	var accountTypesDTO []models.AccountType
	for _, accountType := range accountTypes {
		accountTypesDTO = append(accountTypesDTO, accountType)
	}

	logger.Debug("GetAccountsTypes request completed - GET /accounts/types")
	return c.JSON(http.StatusOK, accountTypesDTO)
}

func GetAccountById(c echo.Context) error {
	logger.Debug("GetAccountById request started", "method", c.Request().Method, "url", c.Request().URL)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid account ID")
	}

	account, err := sm.AccountsService.GetAccountById(id)
	if err != nil {
		logger.Error("Error getting account by ID: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	logger.Debug("GetAccountById request completed - GET /accounts/:id")
	return c.JSON(http.StatusOK, account)
}

func CreateAccount(c echo.Context) error {
	logger.Debug("CreateAccount request started", "method", c.Request().Method, "url", c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		logger.Warn("Authenticated user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	account, err := prepareAccount(c)
	if err != nil {
		return err
	}

	account.UserID = user.ID

	createdAccount, err := sm.AccountsService.CreateAccount(account)
	if err != nil {
		logger.Error("Error creating account: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	logger.Debug("CreateAccount request completed - POST /accounts")
	return c.JSON(http.StatusOK, createdAccount)
}

func UpdateAccount(c echo.Context) error {
	logger.Debug("UpdateAccount request started", "method", c.Request().Method, "url", c.Request().URL)

	account, err := prepareAccount(c)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid account ID", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid account ID")
	}

	account.ID = id

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		logger.Warn("Authenticated user not found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	account.UserID = user.ID

	updatedAccount, err := sm.AccountsService.UpdateAccount(account)
	if err != nil {
		if errors.Is(err, appErrors.ErrNoAccountFound) {
			logger.Error("No account found with the provided ID", "accountID", account.ID)
			return c.String(http.StatusNotFound, "not found")
		}
		logger.Error("Error updating account: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	logger.Debug("UpdateAccount request completed - PUT /accounts/:id")
	return c.JSON(http.StatusOK, updatedAccount)
}

func prepareAccount(c echo.Context) (models.Account, error) {
	var rawInput map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&rawInput); err != nil {
		logger.Error("Error decoding JSON: ", err)
		return models.Account{}, echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}

	if rawOpeningDate, ok := rawInput["opening_date"].(string); ok {
		parsedTime, err := time.Parse("2006-01-02T15:04", rawOpeningDate)
		if err != nil {
			logger.Error("Error parsing opening_date: ", err)
			return models.Account{}, echo.NewHTTPError(http.StatusBadRequest, "Invalid opening_date format")
		}
		rawInput["opening_date"] = parsedTime
	}

	rawBytes, err := json.Marshal(rawInput)
	if err != nil {
		logger.Error("Error marshaling data: ", err)
		return models.Account{}, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	account := new(models.Account)
	if err := json.Unmarshal(rawBytes, account); err != nil {
		logger.Error("Error unmarshaling data: ", err)
		return models.Account{}, echo.NewHTTPError(http.StatusBadRequest, "Invalid data structure")
	}

	return *account, nil
}
