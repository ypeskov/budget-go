package dto

import (
	"ypeskov/budget-go/internal/models"
)

type AccountDTO struct {
	ID                    int      `json:"id" db:"id"`
	UserID                int      `json:"userId" db:"user_id"`
	AccountTypeId         int      `json:"accountTypeId" db:"account_type_id"`
	CurrencyId            int      `json:"currencyId" db:"currency_id"`
	Name                  string   `json:"name" db:"name"`
	Balance               float64  `json:"balance" db:"balance"`
	InitialBalance        *float64 `json:"initialBalance" db:"initial_balance"`
	CreditLimit           *float64 `json:"creditLimit" db:"credit_limit"`
	OpeningDate           string   `json:"openingDate" db:"opening_date"`
	Comment               string   `json:"comment" db:"comment"`
	IsHidden              bool     `json:"isHidden" db:"is_hidden"`
	ShowInReports         bool     `json:"showInReports" db:"show_in_reports"`
	IsDeleted             bool     `json:"isDeleted" db:"is_deleted"`
	ArchivedAt            *string  `json:"archivedAt" db:"archived_at"`
	CreatedAt             string   `json:"createdAt" db:"created_at"`
	UpdateAt              string   `json:"updatedAt" db:"updated_at"`
	BalanceInBaseCurrency *float64 `json:"balanceInBaseCurrency" db:"balance_in_base_currency"`

	Currency    CurrencyDTO    `json:"currency" db:"currency"`
	AccountType AccountTypeDTO `json:"accountType" db:"account_type"`
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
