package services

import (
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
)

// AccountToDTO converts a models.Account to dto.AccountDTO
func AccountToDTO(account models.Account) dto.AccountDTO {
	accountDTO := dto.AccountDTO{
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

		AccountType: models.AccountType{
			ID: account.AccountTypeId,
		},
	}

	return accountDTO
}
