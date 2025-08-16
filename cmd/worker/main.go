package main

import (
	"strings"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/constants"
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/jobs"
	"ypeskov/budget-go/internal/services"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.New()
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	lvlStr := strings.TrimSpace(strings.ToLower(cfg.LogLevel))
	level, err := log.ParseLevel(lvlStr)
	if err != nil {
		log.Fatalf("Invalid log level in config: %s", cfg.LogLevel)
	}
	log.SetLevel(level)

	sm, err := services.NewServicesManager(db, cfg)
	if err != nil {
		log.Fatal(err)
	}

	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: cfg.RedisAddr}, asynq.Config{
		Concurrency: 20,
		Queues:      map[string]int{"emails": 5, "default": 10},
	})

	h := &jobs.Handlers{SM: sm}
	mux := asynq.NewServeMux()
	mux.HandleFunc(constants.TaskEmailSend, h.HandleEmailSend)
	mux.HandleFunc(constants.TaskSendActivationEmail, h.HandleSendActivationEmail)
	mux.HandleFunc(constants.TaskExchangeRatesDaily, h.HandleExchangeRatesDaily)
	mux.HandleFunc(constants.TaskDBBackupDaily, h.HandleDBBackupDaily)
	mux.HandleFunc(constants.TaskBudgetsDailyProcessing, h.HandleBudgetsDailyProcessing)

	// Run blocks and processes jobs until the process receives a shutdown signal
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
