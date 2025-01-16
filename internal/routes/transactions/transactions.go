package transactions

import (
	"net/http"
	"strconv"

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

	var err error
	var perPage int
	perPageStr := c.QueryParam("per_page")
	if perPageStr == "" {
		perPage = 10
	} else {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid perPage value")
		}
	}

	var page int
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid page value")
		}
	}

	transactions, err := sm.TransactionsService.GetTransactionsWithAccounts(user.ID, sm, perPage, page)
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
