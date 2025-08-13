package jobs

import (
    "context"
    "encoding/json"

    "github.com/hibiken/asynq"
    log "github.com/sirupsen/logrus"
    "ypeskov/budget-go/internal/services"
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

func (h *Handlers) HandleBudgetsUpdateUser(ctx context.Context, t *asynq.Task) error {
    var p BudgetsUpdatePayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    log.Infof("Updating budgets for user=%d", p.UserID)
    return h.SM.BudgetsService.UpdateBudgetCollectedAmounts(p.UserID)
}

func (h *Handlers) HandleExchangeRatesDaily(ctx context.Context, t *asynq.Task) error {
    // touch exchange rates cache/populate; replace with actual fetch if needed
    _, err := h.SM.ExchangeRatesService.GetExchangeRates()
    return err
}

func (h *Handlers) HandleDBBackupDaily(ctx context.Context, t *asynq.Task) error {
    // TODO: invoke backup script or function
    log.Info("DB backup task executed (implement script call)")
    return nil
}

func (h *Handlers) HandleBudgetsDailyProcessing(ctx context.Context, t *asynq.Task) error {
    _, err := h.SM.BudgetsService.ProcessOutdatedBudgets()
    return err
}
