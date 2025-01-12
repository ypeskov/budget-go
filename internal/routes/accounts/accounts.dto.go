package accounts

import (
	"github.com/shopspring/decimal"
	"ypeskov/budget-go/internal/models"
)

type UserAccountDTO struct {
	ID             int              `json:"id"`
	UserID         int              `json:"userId"`
	AccountTypeId  int              `json:"accountTypeId"`
	CurrencyId     int              `json:"currencyId"`
	Name           string           `json:"name"`
	Balance        decimal.Decimal  `json:"balance"`
	InitialBalance *decimal.Decimal `json:"initialBalance"`
	CreditLimit    *decimal.Decimal `json:"creditLimit"`
	OpeningDate    string           `json:"openingDate"`
	Comment        string           `json:"comment"`
	IsHidden       bool             `json:"isHidden"`
	ShowInReports  bool             `json:"showInReports"`
	IsDeleted      bool             `json:"isDeleted"`
	ArchivedAt     *string          `json:"archivedAt"`
	CreatedAt      string           `json:"createdAt"`
	UpdateAt       string           `json:"updatedAt"`

	Currency struct {
		ID   int    `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"currency"`

	AccountType struct {
		ID       int    `json:"id"`
		TypeName string `json:"type_name"`
		IsCredit bool   `json:"is_credit"`
	} `json:"accountType"`
}

func AccountToDTO(account models.Account) UserAccountDTO {
	return UserAccountDTO{
		ID:             account.ID,
		UserID:         account.UserID,
		AccountTypeId:  account.AccountTypeId,
		CurrencyId:     account.CurrencyId,
		Name:           account.Name,
		Balance:        account.Balance,
		InitialBalance: account.InitialBalance,
		CreditLimit:    account.CreditLimit,
		OpeningDate:    account.OpeningDate,
		Comment:        account.Comment,
		IsHidden:       account.IsHidden,
		ShowInReports:  account.ShowInReports,
		IsDeleted:      account.IsDeleted,
		ArchivedAt:     account.ArchivedAt,
		CreatedAt:      account.CreatedAt,
		UpdateAt:       account.UpdateAt,

		Currency: struct {
			ID   int    `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		}{
			ID:   account.Currency.ID,
			Code: account.Currency.Code,
			Name: account.Currency.Name,
		},
		AccountType: struct {
			ID       int    `json:"id"`
			TypeName string `json:"type_name"`
			IsCredit bool   `json:"is_credit"`
		}{
			ID:       account.AccountType.ID,
			TypeName: account.AccountType.TypeName,
			IsCredit: account.AccountType.IsCredit,
		},
	}
}
