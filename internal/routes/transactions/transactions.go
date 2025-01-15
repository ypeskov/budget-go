package transactions

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

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
	log.Debug("GetTransactions")
	userRaw := c.Get("user")

	claims, ok := userRaw.(jwt.MapClaims)
	if !ok || claims == nil {
		log.Error("Failed to cast user to jwt.MapClaims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing user")
	}

	user, err := sm.UserService.GetUserByEmail(claims["email"].(string))
	if err != nil {
		log.Error("Error getting user: ", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	transactions, err := sm.TransactionsService.GetTransactionsWithAccounts(user.ID)
	if err != nil {
		log.Error("Error getting transactions: ", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		log.Error("Error getting base currency: ", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	var transactionsDTO []ResponseTransactionDTO
	for _, transaction := range transactions {
		transactionsDTO = append(transactionsDTO, TransactionWithAccountToResponseTransactionDTO(transaction, baseCurrency))
	}

	return c.JSON(http.StatusOK, transactionsDTO)
}
