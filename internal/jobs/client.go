package jobs

import (
    "context"
    "encoding/json"

    "github.com/hibiken/asynq"
)

func EnqueueEmail(ctx context.Context, c *asynq.Client, p EmailPayload) error {
    b, _ := json.Marshal(p)
    _, err := c.EnqueueContext(ctx, asynq.NewTask(TaskEmailSend, b), asynq.Queue("emails"), asynq.MaxRetry(10))
    return err
}

func EnqueueBudgetsUpdate(ctx context.Context, c *asynq.Client, userID int) error {
    b, _ := json.Marshal(BudgetsUpdatePayload{UserID: userID})
    _, err := c.EnqueueContext(ctx, asynq.NewTask(TaskBudgetsUpdateUser, b), asynq.Queue("default"))
    return err
}
