package models

type Account struct {
	ID             int      `db:"id"`
	UserID         int      `db:"user_id"`
	AccountTypeID  int      `db:"account_type"`
	CurrencyID     int      `db:"currency_id"`
	Name           string   `db:"name"`
	Balance        float64  `db:"balance"`
	InitialBalance *float64 `db:"initial_balance"`
	CreditLimit    *float64 `db:"credit_limit"`
	OpeningDate    string   `db:"opening_date"`
	Comment        string   `db:"comment"`
	IsHidden       bool     `db:"is_hidden"`
	ShowInReports  bool     `db:"show_in_reports"`
	IsDeleted      bool     `db:"is_deleted"`
	ArchivedAt     *string  `db:"archived_at"`
	CreatedAt      string   `db:"created_at"`
	UpdateAt       string   `db:"updated_at"`

	Currency    Currency    `db:"currency"`
	AccountType AccountType `db:"account_type"`
}
