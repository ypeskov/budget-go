package dto

import (
	"time"
	"ypeskov/budget-go/internal/models"
)

type TransactionWithAccount struct {
	models.Transaction
	Account     AccountDTO     `db:"accounts"`
	Currency    CurrencyDTO    `db:"currencies"`
	AccountType AccountTypeDTO `db:"account_types"`
	Category    *CategoryDTO   `db:"user_categories"`
}

type CreateTransactionDTO struct {
	ID              *int       `json:"id"`
	UserID          *int       `json:"userId"`
	AccountID       int        `json:"accountId"`
	TargetAccountID *int       `json:"targetAccountId"`
	CategoryID      *int       `json:"categoryId"`
	Amount          float64    `json:"amount"`
	TargetAmount    *float64   `json:"targetAmount"`
	Label           string     `json:"label"`
	Notes           *string    `json:"notes"`
	DateTime        *time.Time `json:"dateTime"`
	IsTransfer      bool       `json:"isTransfer"`
	IsIncome        bool       `json:"isIncome"`
}

type UpdateTransactionDTO struct {
	CreateTransactionDTO
	ID     int `json:"id"`
	UserID int `json:"userId"`
}

type ResponseTransactionDTO struct {
	ID                    int         `json:"id"`
	UserID                int         `json:"userId"`
	AccountID             int         `json:"accountId"`
	CategoryID            *int        `json:"categoryId"`
	Amount                float64     `json:"amount"`
	NewBalance            *float64    `json:"newBalance"`
	Label                 string      `json:"label"`
	Notes                 *string     `json:"notes"`
	DateTime              *time.Time  `json:"dateTime"`
	IsTransfer            bool        `json:"isTransfer"`
	IsIncome              bool        `json:"isIncome"`
	BaseCurrencyAmount    *float64    `json:"baseCurrencyAmount"`
	BaseCurrencyCode      *string     `json:"baseCurrencyCode"`
	LinkedTransactionID   *int        `json:"linkedTransactionId"`
	BalanceInBaseCurrency float64     `json:"balanceInBaseCurrency"`
	Category              CategoryDTO `json:"category"`
	Account               AccountDTO  `json:"account"`
}

func TransactionWithAccountToResponseTransactionDTO(twa TransactionWithAccount, baseCurrency models.Currency) ResponseTransactionDTO {
	var creditLimit float64
	if twa.Account.AccountType.IsCredit {
		if twa.Account.CreditLimit != nil {
			creditLimit = *twa.Account.CreditLimit
		} else {
			creditLimit = 0
		}
	} else {
		creditLimit = 0
	}

	var balanceInBaseCurrency float64
	if twa.BaseCurrencyAmount != nil {
		balanceInBaseCurrency = *twa.BaseCurrencyAmount
	} else {
		balanceInBaseCurrency = 0
	}

	var baseCurrencyAmount *float64
	if twa.BaseCurrencyAmount != nil {
		baseCurrencyAmount = twa.BaseCurrencyAmount
	} else {
		zero := 0.0
		baseCurrencyAmount = &zero
	}
	// log.Info(twa.Account.AccountType)
	return ResponseTransactionDTO{
		ID:                    twa.ID,
		UserID:                twa.UserID,
		AccountID:             twa.AccountID,
		CategoryID:            twa.CategoryID,
		Amount:                twa.Amount,
		Label:                 twa.Label,
		Notes:                 twa.Notes,
		DateTime:              twa.DateTime,
		IsTransfer:            twa.IsTransfer,
		IsIncome:              twa.IsIncome,
		LinkedTransactionID:   twa.LinkedTransactionID,
		BaseCurrencyAmount:    baseCurrencyAmount,
		BaseCurrencyCode:      &baseCurrency.Code,
		NewBalance:            twa.NewBalance,
		BalanceInBaseCurrency: balanceInBaseCurrency,
		Account: AccountDTO{
			ID:          twa.Account.ID,
			Name:        twa.Account.Name,
			Balance:     twa.Account.Balance,
			CreditLimit: &creditLimit,
			OpeningDate: twa.Account.OpeningDate,
			Comment:     twa.Account.Comment,
			Currency: CurrencyDTO{
				ID:   twa.Account.Currency.ID,
				Code: twa.Account.Currency.Code,
				Name: twa.Account.Currency.Name,
			},
			AccountType: AccountTypeDTO{
				ID:       twa.Account.AccountType.ID,
				TypeName: twa.Account.AccountType.TypeName,
				IsCredit: twa.Account.AccountType.IsCredit,
			},
		},
		Category: CategoryDTO{
			ID:        twa.Category.ID,
			Name:      twa.Category.Name,
			ParentID:  twa.Category.ParentID,
			IsIncome:  twa.Category.IsIncome,
			UserID:    twa.Category.UserID,
			IsDeleted: twa.Category.IsDeleted,
			CreatedAt: twa.Category.CreatedAt,
			UpdatedAt: twa.Category.UpdatedAt,
		},
	}
}
