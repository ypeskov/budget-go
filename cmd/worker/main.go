package main

import (
    "log"

    "github.com/hibiken/asynq"
    "ypeskov/budget-go/internal/config"
    "ypeskov/budget-go/internal/database"
    "ypeskov/budget-go/internal/jobs"
    "ypeskov/budget-go/internal/services"
)

func main() {
    cfg := config.New()
    db, err := database.New(cfg)
    if err != nil { log.Fatal(err) }
    sm := services.NewServicesManager(db, cfg)

    srv := asynq.NewServer(asynq.RedisClientOpt{Addr: cfg.RedisAddr}, asynq.Config{
        Concurrency: 20,
        Queues:      map[string]int{"emails": 5, "default": 10},
    })

    h := &jobs.Handlers{SM: sm}
    mux := asynq.NewServeMux()
    mux.HandleFunc(jobs.TaskEmailSend, h.HandleEmailSend)
    mux.HandleFunc(jobs.TaskBudgetsUpdateUser, h.HandleBudgetsUpdateUser)
    mux.HandleFunc(jobs.TaskExchangeRatesDaily, h.HandleExchangeRatesDaily)
    mux.HandleFunc(jobs.TaskDBBackupDaily, h.HandleDBBackupDaily)
    mux.HandleFunc(jobs.TaskBudgetsDailyProcessing, h.HandleBudgetsDailyProcessing)

    // Run blocks and processes jobs until the process receives a shutdown signal
    if err := srv.Run(mux); err != nil { log.Fatal(err) }
}
