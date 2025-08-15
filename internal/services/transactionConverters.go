package services

import (
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"

	"github.com/shopspring/decimal"
)

// ConvertTransactionsToResponseList converts a slice of TransactionWithAccount to ResponseTransactionDTO
func ConvertTransactionsToResponseList(transactions []dto.TransactionWithAccount, baseCurrency models.Currency) []dto.ResponseTransactionDTO {
	responseList := make([]dto.ResponseTransactionDTO, 0, len(transactions))
	for _, transaction := range transactions {
		responseList = append(responseList, ConvertToResponseTransaction(transaction, baseCurrency))
	}
	return responseList
}

// ConvertToResponseTransaction converts a single TransactionWithAccount to ResponseTransactionDTO
func ConvertToResponseTransaction(twa dto.TransactionWithAccount, baseCurrency models.Currency) dto.ResponseTransactionDTO {
	var creditLimit decimal.Decimal
	if twa.Account.AccountType.IsCredit {
		if twa.Account.CreditLimit != nil {
			creditLimit = *twa.Account.CreditLimit
		} else {
			creditLimit = decimal.Zero
		}
	} else {
		creditLimit = decimal.Zero
	}

	var balanceInBaseCurrency decimal.Decimal
	if twa.BaseCurrencyAmount != nil {
		balanceInBaseCurrency = *twa.BaseCurrencyAmount
	} else {
		balanceInBaseCurrency = decimal.Zero
	}

	var baseCurrencyAmount *decimal.Decimal
	if twa.BaseCurrencyAmount != nil {
		baseCurrencyAmount = twa.BaseCurrencyAmount
	} else {
		baseCurrencyAmount = &decimal.Zero
	}

	return dto.ResponseTransactionDTO{
		ID:                    *twa.ID,
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
		Account: dto.AccountDTO{
			ID:          twa.Account.ID,
			Name:        twa.Account.Name,
			Balance:     twa.Account.Balance,
			CreditLimit: &creditLimit,
			OpeningDate: twa.Account.OpeningDate,
			Comment:     twa.Account.Comment,
			Currency:    twa.Account.Currency,
			AccountType: twa.Account.AccountType,
		},
		Category: dto.CategoryDTO{
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