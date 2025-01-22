package models

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type Account struct {
	ID             int              `db:"id" json:"id"`
	UserID         int              `db:"user_id" json:"user_id"`
	AccountTypeId  int              `db:"account_type_id" json:"account_type_id"`
	CurrencyId     int              `db:"currency_id" json:"currency_id"`
	Name           string           `db:"name" json:"name"`
	Balance        decimal.Decimal  `db:"balance" json:"balance"`
	InitialBalance *decimal.Decimal `db:"initial_balance" json:"initial_balance"`
	CreditLimit    *decimal.Decimal `db:"credit_limit" json:"credit_limit"`
	OpeningDate    time.Time        `db:"opening_date" json:"opening_date"`
	Comment        string           `db:"comment" json:"comment"`
	IsHidden       bool             `db:"is_hidden" json:"is_hidden"`
	ShowInReports  bool             `db:"show_in_reports" json:"show_in_reports"`
	IsDeleted      bool             `db:"is_deleted" json:"is_deleted"`
	ArchivedAt     *string          `db:"archived_at" json:"archived_at"`
	CreatedAt      time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at" json:"updated_at"`
}

func (a *Account) MarshalJSON() ([]byte, error) {
	type Alias Account
	alias := &struct {
		Balance        float64 `json:"Balance"`
		InitialBalance float64 `json:"InitialBalance"`
		CreditLimit    float64 `json:"CreditLimit"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	// Convert Balance to float64
	alias.Balance = a.Balance.InexactFloat64()

	// Convert InitialBalance to 0.0 if nil, otherwise to float64
	if a.InitialBalance == nil {
		alias.InitialBalance = 0.0
	} else {
		alias.InitialBalance, _ = a.InitialBalance.Float64()
	}

	// Convert CreditLimit to 0.0 if nil, otherwise to float64
	if a.CreditLimit == nil {
		alias.CreditLimit = 0.0
	} else {
		alias.CreditLimit, _ = a.CreditLimit.Float64()
	}

	return json.Marshal(alias)
}
