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
	log.Debug("GetBaseCurrency Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "User not found")
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(user.ID)
	if err != nil {
		log.Error("Failed to get base currency: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get base currency")
	}

	return c.JSON(http.StatusOK, baseCurrency)
}

func GetLanguages(c echo.Context) error {
	log.Debug("GetLanguages Route")

	languages, err := sm.LanguagesService.GetLanguages()
	if err != nil {
		log.Error("Failed to get languages: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get languages")
	}

	return c.JSON(http.StatusOK, languages)
}

func UpdateSettings(c echo.Context) error {
	log.Debug("UpdateSettings Route")

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

	return c.JSON(http.StatusOK, userSettings)
}
