package userSettings

import (
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

func RegisterSettingsRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("/base-currency", GetBaseCurrency)
	g.GET("/languages", GetLanguages)
	g.POST("", UpdateSettings)
}

func GetBaseCurrency(c echo.Context) error {
	log.Debugf("GetBaseCurrency request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "User not found")
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		log.Error("Failed to get base currency: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get base currency")
	}

	log.Debug("GetBaseCurrency request completed - GET /settings/base-currency")
	return c.JSON(http.StatusOK, baseCurrency)
}

func GetLanguages(c echo.Context) error {
	log.Debugf("GetLanguages request started: %s %s", c.Request().Method, c.Request().URL)

	languages, err := sm.LanguagesService.GetLanguages()
	if err != nil {
		log.Error("Failed to get languages: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get languages")
	}

	log.Debug("GetLanguages request completed - GET /settings/languages")
	return c.JSON(http.StatusOK, languages)
}

func UpdateSettings(c echo.Context) error {
	log.Debugf("UpdateSettings request started: %s %s", c.Request().Method, c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "User not found")
	}

	var settingsDTO dto.UpdateSettingsDTO
	if err := c.Bind(&settingsDTO); err != nil {
		log.Error("Failed to bind settings DTO: ", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Basic validation - check if language is provided
	if settingsDTO.Language == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Language is required")
	}

	// Convert DTO to settings map
	settingsData := map[string]interface{}{
		"language": settingsDTO.Language,
	}

	userSettings, err := sm.UserSettingsService.UpdateUserSettings(user.ID, settingsData)
	if err != nil {
		log.Error("Failed to update user settings: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update settings")
	}

	log.Debug("UpdateSettings request completed - POST /settings")
	return c.JSON(http.StatusOK, userSettings)
}
