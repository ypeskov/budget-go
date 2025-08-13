package jobs

const (
    TaskEmailSend              = "email:send"
    TaskBudgetsUpdateUser      = "budgets:update_user"
    TaskExchangeRatesDaily     = "exchange_rates:daily_update"
    TaskDBBackupDaily          = "db:backup"
    TaskBudgetsDailyProcessing = "budgets:daily_processing"
)

type EmailPayload struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

type BudgetsUpdatePayload struct {
    UserID int `json:"userId"`
}
