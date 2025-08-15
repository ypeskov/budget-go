package jobs

import (
	"context"
	"encoding/json"
	"time"

	"ypeskov/budget-go/internal/services"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

type Handlers struct{ SM *services.Manager }

func (h *Handlers) HandleEmailSend(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	// TODO: integrate SMTP/ESP here. For now, just log.
	log.Infof("Sending email to=%s subject=%s", p.To, p.Subject)
	return nil
}

func (h *Handlers) HandleExchangeRatesDaily(ctx context.Context, t *asynq.Task) error {
	log.Info("Exchange rates update task started")

	// Update exchange rates for today
	today := time.Now()
	exchangeRates, err := h.SM.ExchangeRatesService.UpdateExchangeRates(today)
	if err != nil {
		log.Errorf("Exchange rates update failed: %v", err)
		return err
	}

	log.Infof("Exchange rates updated successfully for %s (base: %s)",
		exchangeRates.ActualDate.Format("2006-01-02"), exchangeRates.BaseCurrencyCode)

	// Send notification email
	err = h.SM.EmailService.SendExchangeRatesUpdateNotification(exchangeRates)
	if err != nil {
		log.Errorf("Failed to send exchange rates notification email: %v", err)
		return err
	}

	log.Info("Exchange rates update task completed successfully")
	return nil
}

func (h *Handlers) HandleDBBackupDaily(ctx context.Context, t *asynq.Task) error {
	log.Info("Starting database backup task")

	backupResult, err := h.SM.BackupService.CreatePostgresBackup()
	if err != nil {
		log.Errorf("Database backup failed: %v", err)
		return err
	}

	log.Infof("Database backup created successfully: %s", backupResult.Filename)

	err = h.SM.EmailService.SendBackupNotification(backupResult)
	if err != nil {
		log.Errorf("Failed to send backup notification email: %v", err)
		return err
	}

	log.Info("Database backup task completed successfully")
	return nil
}

func (h *Handlers) HandleBudgetsDailyProcessing(ctx context.Context, t *asynq.Task) error {
	log.Info("Starting budgets daily processing task")
	_, err := h.SM.BudgetsService.ProcessOutdatedBudgets()
	if err != nil {
		log.Errorf("Budgets daily processing failed: %v", err)
		return err
	}

	log.Info("Budgets daily processing task completed successfully")
	return nil
}

func (h *Handlers) HandleSendActivationEmail(ctx context.Context, t *asynq.Task) error {
	var p ActivationEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Errorf("Failed to unmarshal activation email payload: %v", err)
		return err
	}

	log.Infof("Sending activation email to %s", p.UserEmail)

	err := h.SM.EmailService.SendActivationEmail(p.UserEmail, p.UserName, p.Token)
	if err != nil {
		log.Errorf("Failed to send activation email: %v", err)
		return err
	}

	log.Infof("Activation email sent successfully to %s", p.UserEmail)
	return nil
}
