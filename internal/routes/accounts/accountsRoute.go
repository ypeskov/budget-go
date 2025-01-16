package accounts

import (
	"net/http"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
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
}

func GetAccounts(c echo.Context) error {
	log.Debug("GetAccounts")
	userRaw := c.Get("user")

	claims, ok := userRaw.(jwt.MapClaims)
	if !ok || claims == nil {
		log.Error("Failed to cast user to jwt.MapClaims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing user")
	}

	email, emailOk := claims["email"].(string)
	if !emailOk {
		log.Error("Email not found in claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}

	user, err := sm.UserService.GetUserByEmail(email)
	if err != nil {
		log.Error("Error getting user by email: ", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	// baseCurrency, err := sm.CurrenciesService.GetCurrency(user.BaseCurrencyID)
	// if err != nil {
	// 	log.Error("Error getting base currency: ", err)
	// 	return c.String(http.StatusInternalServerError, "Internal server error")
	// }

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

	userAccounts, err := sm.AccountsService.GetUserAccounts(user.ID, includeHidden, includeDeleted, archivedOnly)
	if err != nil {
		log.Error("Error getting user accounts: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}
	for i, account := range userAccounts {
		if account.InitialBalance == nil {
			zero := 0.0
			userAccounts[i].InitialBalance = &zero
		}

		if account.CreditLimit == nil {
			zero := 0.0
			userAccounts[i].CreditLimit = &zero
		}

		// TODO: remove this after base currency is implemented
		if account.BalanceInBaseCurrency == nil {
			zero := 0.0
			userAccounts[i].BalanceInBaseCurrency = &zero
		}
	}

	return c.JSON(http.StatusOK, userAccounts)
}

func GetAccountsTypes(c echo.Context) error {
	log.Debug("GetAccountsTypes Route")
	accountTypes, err := sm.AccountsService.GetAccountTypes()
	if err != nil {
		log.Error("Error getting account types: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	var accountTypesDTO []dto.AccountTypeDTO
	for _, accountType := range accountTypes {
		accountTypesDTO = append(accountTypesDTO, dto.AccountTypeToDTO(accountType))
	}

	return c.JSON(http.StatusOK, accountTypesDTO)
}
