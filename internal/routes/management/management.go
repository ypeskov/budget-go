package management

import (
	"net/http"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/services"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
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
	log.Debugf("triggerBackup request started: %s %s", c.Request().Method, c.Request().URL)

	err := sm.QueueService.EnqueueDBBackup()
	if err != nil {
		log.Errorf("failed to enqueue db backup task: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue task"})
	}

	log.Debug("triggerBackup request completed - GET /management/backup")
	return c.JSON(http.StatusAccepted, map[string]string{"message": "backup scheduled"})
}

func triggerUpdateExchangeRates(c echo.Context, sm *services.Manager) error {
	log.Debugf("triggerUpdateExchangeRates request started: %s %s", c.Request().Method, c.Request().URL)

	err := sm.QueueService.EnqueueExchangeRatesUpdate()
	if err != nil {
		log.Errorf("failed to enqueue exchange rates update task: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue task"})
	}
	log.Debug("triggerUpdateExchangeRates request completed - GET /management/update-exchange-rates")
	return c.JSON(http.StatusAccepted, map[string]string{"message": "exchange rates update scheduled"})
}
