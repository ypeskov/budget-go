package models

import (
	"strconv"
	"time"
)

type Transaction struct {
	ID                  *int       `db:"id"`
	UserID              int        `db:"user_id"`
	AccountID           int        `db:"account_id"`
	Amount              float64    `db:"amount"`
	NewBalance          *float64   `db:"new_balance"`
	CategoryID          *int       `db:"category_id"`
	Label               string     `db:"label"`
	IsIncome            bool       `db:"is_income"`
	IsTransfer          bool       `db:"is_transfer"`
	LinkedTransactionID *int       `db:"linked_transaction_id"`
	BaseCurrencyAmount  *float64   `db:"base_currency_amount"`
	Notes               *string    `db:"notes"`
	DateTime            *time.Time `db:"date_time"`
	IsDeleted           bool       `db:"is_deleted"`
	CreatedAt           *time.Time `db:"created_at"`
	UpdatedAt           *time.Time `db:"updated_at"`
}

func (t *Transaction) String() string {
	return "[Transaction: " + t.Label + ", Amount: " + strconv.FormatFloat(t.Amount, 'f', -1, 64) + "]"
}
