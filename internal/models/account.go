package models

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Account struct {
	ID             int              `db:"id"`
	UserID         int              `db:"user_id"`
	AccountTypeId  int              `db:"account_type_id"`
	CurrencyId     int              `db:"currency_id"`
	Name           string           `db:"name"`
	Balance        decimal.Decimal  `db:"balance"`
	InitialBalance *decimal.Decimal `db:"initial_balance"`
	CreditLimit    *decimal.Decimal `db:"credit_limit"`
	OpeningDate    string           `db:"opening_date"`
	Comment        string           `db:"comment"`
	IsHidden       bool             `db:"is_hidden"`
	ShowInReports  bool             `db:"show_in_reports"`
	IsDeleted      bool             `db:"is_deleted"`
	ArchivedAt     *string          `db:"archived_at"`
	CreatedAt      string           `db:"created_at"`
	UpdateAt       string           `db:"updated_at"`
}

func (a *Account) String() string {
	return "[Account: " + a.Name + ", Balance: " + a.Balance.String() + "]"
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
