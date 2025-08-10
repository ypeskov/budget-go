package dto

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
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

	Currency    CurrencyDTO    `json:"currency" db:"currency"`
	AccountType AccountTypeDTO `json:"accountType" db:"account_type"`
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

type AccountTypeDTO struct {
	ID       int    `json:"id" db:"id"`
	TypeName string `json:"type_name" db:"type_name"`
	IsCredit bool   `json:"is_credit" db:"is_credit"`
}

func AccountTypeToDTO(accountType models.AccountType) AccountTypeDTO {
	return AccountTypeDTO{
		ID:       accountType.ID,
		TypeName: accountType.TypeName,
		IsCredit: accountType.IsCredit,
	}
}

func AccountToDTO(account models.Account) AccountDTO {
	accountDTO := AccountDTO{
		ID:             account.ID,
		UserID:         account.UserID,
		AccountTypeId:  account.AccountTypeId,
		CurrencyId:     account.CurrencyId,
		Name:           account.Name,
		Balance:        account.Balance,
		InitialBalance: account.InitialBalance,
		CreditLimit:    account.CreditLimit,
		OpeningDate:    account.OpeningDate.Format(time.RFC3339),
		Comment:        account.Comment,
		IsHidden:       account.IsHidden,
		ShowInReports:  account.ShowInReports,
		IsDeleted:      account.IsDeleted,
		ArchivedAt:     account.ArchivedAt,
		CreatedAt:      account.CreatedAt.Format(time.RFC3339),
		UpdateAt:       account.UpdatedAt.Format(time.RFC3339),

		AccountType: AccountTypeDTO{
			ID: account.AccountTypeId,
		},
	}

	return accountDTO
}
