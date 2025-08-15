package dto

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"ypeskov/budget-go/internal/models"
)

type AccountDTO struct {
	ID                    int              `json:"id" db:"id"`
	UserID                int              `json:"userId" db:"user_id"`
	AccountTypeId         int              `json:"accountTypeId" db:"account_type_id"`
	CurrencyId            int              `json:"currencyId" db:"currency_id"`
	Name                  string           `json:"name" db:"name"`
	Balance               decimal.Decimal  `json:"balance" db:"balance"`
	InitialBalance        *decimal.Decimal `json:"initialBalance" db:"initial_balance"`
	CreditLimit           *decimal.Decimal `json:"creditLimit" db:"credit_limit"`
	OpeningDate           string           `json:"openingDate" db:"opening_date"`
	Comment               string           `json:"comment" db:"comment"`
	IsHidden              bool             `json:"isHidden" db:"is_hidden"`
	ShowInReports         bool             `json:"showInReports" db:"show_in_reports"`
	IsDeleted             bool             `json:"isDeleted" db:"is_deleted"`
	ArchivedAt            *string          `json:"archivedAt" db:"archived_at"`
	CreatedAt             string           `json:"createdAt" db:"created_at"`
	UpdateAt              string           `json:"updatedAt" db:"updated_at"`
	BalanceInBaseCurrency *decimal.Decimal `json:"balanceInBaseCurrency" db:"balance_in_base_currency"`

	Currency    models.Currency    `json:"currency" db:"currency"`
	AccountType models.AccountType `json:"accountType" db:"account_type"`
}

func (a *AccountDTO) MarshalJSON() ([]byte, error) {
	type Alias AccountDTO
	var initialBalance *float64
	if a.InitialBalance != nil {
		val, _ := a.InitialBalance.Float64()
		initialBalance = &val
	}
	var creditLimit *float64
	if a.CreditLimit != nil {
		val, _ := a.CreditLimit.Float64()
		creditLimit = &val
	}
	var balanceInBaseCurrency *float64
	if a.BalanceInBaseCurrency != nil {
		val, _ := a.BalanceInBaseCurrency.Float64()
		balanceInBaseCurrency = &val
	}

	return json.Marshal(&struct {
		Balance               float64  `json:"balance"`
		InitialBalance        *float64 `json:"initialBalance"`
		CreditLimit           *float64 `json:"creditLimit"`
		BalanceInBaseCurrency *float64 `json:"balanceInBaseCurrency"`
		*Alias
	}{
		Balance:               a.Balance.InexactFloat64(),
		InitialBalance:        initialBalance,
		CreditLimit:           creditLimit,
		BalanceInBaseCurrency: balanceInBaseCurrency,
		Alias:                 (*Alias)(a),
	})
}

// AccountTypeDTO has been consolidated with models.AccountType
// Use models.AccountType directly for all account type operations
