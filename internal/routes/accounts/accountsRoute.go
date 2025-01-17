package accounts

import (
	"net/http"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/services"

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
	log.Debug("GetAccounts Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		log.Errorf("User %v not found\n", user.Email)
		return echo.NewHTTPError(http.StatusInternalServerError, "User not found")
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
		log.Error("Error getting user accounts: ", err)
		return c.String(http.StatusInternalServerError, "Internal server error")
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
