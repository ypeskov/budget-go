package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	logrus "github.com/sirupsen/logrus"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/jobs"
)

func main() {
	cfg := config.New()

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("failed to load timezone %s: %v", cfg.Timezone, err)
	}

	sch := asynq.NewScheduler(asynq.RedisClientOpt{Addr: cfg.RedisAddr}, &asynq.SchedulerOpts{Location: loc})

	// Build cron expressions: "minute hour * * *"
	ex := fmt.Sprintf("%d %d * * *", cfg.ExchangeRatesMinute, cfg.ExchangeRatesHour)
	db := fmt.Sprintf("%d %d * * *", cfg.DBBackupMinute, cfg.DBBackupHour)
	bud := fmt.Sprintf("%d %d * * *", cfg.BudgetsProcMinute, cfg.BudgetsProcHour)

	if _, err := sch.Register(ex, asynq.NewTask(jobs.TaskExchangeRatesDaily, nil)); err != nil {
		log.Fatal(err)
	} else {
		logrus.Infof("Scheduled task '%s' to run at cron '%s'", jobs.TaskExchangeRatesDaily, ex)
	}

	if _, err := sch.Register(db, asynq.NewTask(jobs.TaskDBBackupDaily, nil)); err != nil {
		log.Fatal(err)
	} else {
		logrus.Infof("Scheduled task '%s' to run at cron '%s'", jobs.TaskDBBackupDaily, db)
	}

	if _, err := sch.Register(bud, asynq.NewTask(jobs.TaskBudgetsDailyProcessing, nil)); err != nil {
		log.Fatal(err)
	} else {
		logrus.Infof("Scheduled task '%s' to run at cron '%s'", jobs.TaskBudgetsDailyProcessing, bud)
	}

	if err := sch.Run(); err != nil {
		log.Fatal(err)
	}
}
