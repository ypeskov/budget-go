package dto

import (
	"ypeskov/budget-go/internal/models"
)

type AccountDTO struct {
	ID                    int      `json:"id"`
	UserID                int      `json:"userId"`
	AccountTypeId         int      `json:"accountTypeId"`
	CurrencyId            int      `json:"currencyId"`
	Name                  string   `json:"name"`
	Balance               float64  `json:"balance"`
	InitialBalance        float64  `json:"initialBalance"`
	CreditLimit           *float64 `json:"creditLimit"`
	OpeningDate           string   `json:"openingDate"`
	Comment               string   `json:"comment"`
	IsHidden              bool     `json:"isHidden"`
	ShowInReports         bool     `json:"showInReports"`
	IsDeleted             bool     `json:"isDeleted"`
	ArchivedAt            *string  `json:"archivedAt"`
	CreatedAt             string   `json:"createdAt"`
	UpdateAt              string   `json:"updatedAt"`
	BalanceInBaseCurrency *float64 `json:"balanceInBaseCurrency"`

	Currency    CurrencyDTO    `json:"currency"`
	AccountType AccountTypeDTO `json:"accountType"`
}

func AccountToDTO(account models.Account, baseCurrency models.Currency, accountType models.AccountType, currency models.Currency) AccountDTO {
	balance, _ := account.Balance.Float64()

	var initialBalance float64
	if account.InitialBalance != nil {
		if val, ok := account.InitialBalance.Float64(); ok {
			initialBalance = val
		}
	} else {
		initialBalance = 0
	}

	var creditLimit float64
	if account.CreditLimit != nil {
		if val, ok := account.CreditLimit.Float64(); ok {
			creditLimit = val
		}
	} else {
		creditLimit = 0
	}

	return AccountDTO{
		ID:             account.ID,
		UserID:         account.UserID,
		AccountTypeId:  account.AccountTypeId,
		CurrencyId:     account.CurrencyId,
		Name:           account.Name,
		Balance:        balance,
		InitialBalance: initialBalance,
		CreditLimit:    &creditLimit,
		OpeningDate:    account.OpeningDate,
		Comment:        account.Comment,
		IsHidden:       account.IsHidden,
		ShowInReports:  account.ShowInReports,
		IsDeleted:      account.IsDeleted,
		ArchivedAt:     account.ArchivedAt,
		CreatedAt:      account.CreatedAt,
		UpdateAt:       account.UpdateAt,

		Currency: CurrencyDTO{
			ID:   currency.ID,
			Code: currency.Code,
			Name: currency.Name,
		},
		AccountType: AccountTypeDTO{
			ID:       accountType.ID,
			TypeName: accountType.TypeName,
			IsCredit: accountType.IsCredit,
		},
		BalanceInBaseCurrency: nil,
	}
}

type AccountTypeDTO struct {
	ID       int    `json:"id"`
	TypeName string `json:"type_name"`
	IsCredit bool   `json:"is_credit"`
}

func AccountTypeToDTO(accountType models.AccountType) AccountTypeDTO {
	return AccountTypeDTO{
		ID:       accountType.ID,
		TypeName: accountType.TypeName,
		IsCredit: accountType.IsCredit,
	}
}
