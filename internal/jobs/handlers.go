package jobs

import (
	"context"
	"encoding/json"
	"time"

	"ypeskov/budget-go/internal/queue"
	"ypeskov/budget-go/internal/services"

	"github.com/hibiken/asynq"
	"ypeskov/budget-go/internal/logger"
)

type Handlers struct{ SM *services.Manager }

func (h *Handlers) HandleEmailSend(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	// TODO: integrate SMTP/ESP here. For now, just log.
	logger.Info("Sending email", "to", p.To, "subject", p.Subject)
	return nil
}

func (h *Handlers) HandleExchangeRatesDaily(ctx context.Context, t *asynq.Task) error {
	logger.Info("Exchange rates update task started")

	// Update exchange rates for today
	today := time.Now()
	exchangeRates, err := h.SM.ExchangeRatesService.UpdateExchangeRates(today)
	if err != nil {
		logger.Error("Exchange rates update failed", "error", err)
		return err
	}

	logger.Info("Exchange rates updated successfully", "date", exchangeRates.ActualDate.Format("2006-01-02"), "base", exchangeRates.BaseCurrencyCode)

	// Send notification email
	err = h.SM.EmailService.SendExchangeRatesUpdateNotification(exchangeRates)
	if err != nil {
		logger.Error("Failed to send exchange rates notification email", "error", err)
		return err
	}

	logger.Info("Exchange rates update task completed successfully")
	return nil
}

func (h *Handlers) HandleDBBackupDaily(ctx context.Context, t *asynq.Task) error {
	logger.Info("Starting database backup task")

	backupResult, err := h.SM.BackupService.CreatePostgresBackup()
	if err != nil {
		logger.Error("Database backup failed", "error", err)
		return err
	}

	logger.Info("Database backup created successfully", "filename", backupResult.Filename)

	err = h.SM.EmailService.SendBackupNotification(backupResult)
	if err != nil {
		logger.Error("Failed to send backup notification email", "error", err)
		return err
	}

	logger.Info("Database backup task completed successfully")
	return nil
}

func (h *Handlers) HandleBudgetsDailyProcessing(ctx context.Context, t *asynq.Task) error {
	logger.Info("Starting budgets daily processing task")
	_, err := h.SM.BudgetsService.ProcessOutdatedBudgets()
	if err != nil {
		logger.Error("Budgets daily processing failed", "error", err)
		return err
	}

	logger.Info("Budgets daily processing task completed successfully")
	return nil
}

func (h *Handlers) HandleSendActivationEmail(ctx context.Context, t *asynq.Task) error {
	var p queue.ActivationEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		logger.Error("Failed to unmarshal activation email payload", "error", err)
		return err
	}

	logger.Info("Sending activation email", "email", p.UserEmail)

	err := h.SM.EmailService.SendActivationEmail(p.UserEmail, p.UserName, p.Token)
	if err != nil {
		logger.Error("Failed to send activation email", "error", err)
		return err
	}

	logger.Info("Activation email sent successfully", "email", p.UserEmail)
	return nil
}
