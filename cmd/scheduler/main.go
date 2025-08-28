package main

import (
	"fmt"
	"time"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/constants"
	"ypeskov/budget-go/internal/logger"

	"github.com/hibiken/asynq"
)

func main() {
	cfg := config.New()
	logger.Init(cfg.LogLevel)

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		logger.Fatal("failed to load timezone", "timezone", cfg.Timezone, "error", err)
	}

	sch := asynq.NewScheduler(asynq.RedisClientOpt{Addr: cfg.RedisAddr}, &asynq.SchedulerOpts{Location: loc})

	// Build cron expressions: "minute hour * * *"
	ex := fmt.Sprintf("%d %d * * *", cfg.ExchangeRatesMinute, cfg.ExchangeRatesHour)
	db := fmt.Sprintf("%d %d * * *", cfg.DBBackupMinute, cfg.DBBackupHour)
	bud := fmt.Sprintf("%d %d * * *", cfg.BudgetsProcMinute, cfg.BudgetsProcHour)

	if _, err := sch.Register(ex, asynq.NewTask(constants.TaskExchangeRatesDaily, nil)); err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Info("Scheduled task to run at cron", "task", constants.TaskExchangeRatesDaily, "cron", ex)
	}

	if _, err := sch.Register(db, asynq.NewTask(constants.TaskDBBackupDaily, nil)); err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Info("Scheduled task to run at cron", "task", constants.TaskDBBackupDaily, "cron", db)
	}

	if _, err := sch.Register(bud, asynq.NewTask(constants.TaskBudgetsDailyProcessing, nil)); err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Info("Scheduled task to run at cron", "task", constants.TaskBudgetsDailyProcessing, "cron", bud)
	}

	if err := sch.Run(); err != nil {
		logger.Fatal(err.Error())
	}
}
