package management

import (
	"net/http"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/jobs"
	"ypeskov/budget-go/internal/services"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

var (
	asynqClient *asynq.Client
)

func RegisterManagementRoutes(g *echo.Group, cfg *config.Config, _ *services.Manager) {
	// Lazily init a singleton asynq client using cfg
	if asynqClient == nil {
		asynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	}

	g.GET("/backup", triggerBackup)
	g.GET("/update-exchange-rates", triggerUpdateExchangeRates)
}

func triggerBackup(c echo.Context) error {
	log.Debugf("triggerBackup request started: %s %s", c.Request().Method, c.Request().URL)

	// Enqueue DB backup task
	if asynqClient == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "queue not initialized"})
	}

	_, err := asynqClient.Enqueue(asynq.NewTask(jobs.TaskDBBackupDaily, nil), asynq.Queue("default"))
	if err != nil {
		log.Errorf("failed to enqueue db backup task: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue task"})
	}

	log.Debug("triggerBackup request completed - GET /management/backup")
	return c.JSON(http.StatusAccepted, map[string]string{"message": "backup scheduled"})
}

func triggerUpdateExchangeRates(c echo.Context) error {
	log.Debugf("triggerUpdateExchangeRates request started: %s %s", c.Request().Method, c.Request().URL)

	if asynqClient == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "queue not initialized"})
	}
	_, err := asynqClient.Enqueue(asynq.NewTask(jobs.TaskExchangeRatesDaily, nil), asynq.Queue("default"))
	if err != nil {
		log.Errorf("failed to enqueue exchange rates update task: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue task"})
	}
	log.Debug("triggerUpdateExchangeRates request completed - GET /management/update-exchange-rates")
	return c.JSON(http.StatusAccepted, map[string]string{"message": "exchange rates update scheduled"})
}
