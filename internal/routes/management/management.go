package management

import (
	"net/http"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/services"

	"github.com/labstack/echo/v4"
	"ypeskov/budget-go/internal/logger"
)


func RegisterManagementRoutes(g *echo.Group, cfg *config.Config, sm *services.Manager) {
	g.GET("/backup", func(c echo.Context) error {
		return triggerBackup(c, sm)
	})
	g.GET("/update-exchange-rates", func(c echo.Context) error {
		return triggerUpdateExchangeRates(c, sm)
	})
}

func triggerBackup(c echo.Context, sm *services.Manager) error {
	logger.Debug("triggerBackup request started", "method", c.Request().Method, "url", c.Request().URL)

	err := sm.QueueService.EnqueueDBBackup()
	if err != nil {
		logger.Error("failed to enqueue db backup task", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue task"})
	}

	logger.Debug("triggerBackup request completed - GET /management/backup")
	return c.JSON(http.StatusAccepted, map[string]string{"message": "backup scheduled"})
}

func triggerUpdateExchangeRates(c echo.Context, sm *services.Manager) error {
	logger.Debug("triggerUpdateExchangeRates request started", "method", c.Request().Method, "url", c.Request().URL)

	err := sm.QueueService.EnqueueExchangeRatesUpdate()
	if err != nil {
		logger.Error("failed to enqueue exchange rates update task", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue task"})
	}
	logger.Debug("triggerUpdateExchangeRates request completed - GET /management/update-exchange-rates")
	return c.JSON(http.StatusAccepted, map[string]string{"message": "exchange rates update scheduled"})
}
