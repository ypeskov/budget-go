package jobs

const (
	TaskEmailSend              = "email:send"
	TaskExchangeRatesDaily     = "exchange_rates:daily_update"
	TaskDBBackupDaily          = "db:backup"
	TaskBudgetsDailyProcessing = "budgets:daily_processing"
	TaskSendActivationEmail    = "email:send_activation"
)

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type BudgetsUpdatePayload struct {
	UserID int `json:"userId"`
}

type ActivationEmailPayload struct {
	UserEmail string `json:"userEmail"`
	UserName  string `json:"userName"`
	Token     string `json:"token"`
}
