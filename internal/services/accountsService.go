package services

import (
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/accounts"
)

type AccountsService interface {
	GetUserAccounts(userId int,
		includeHidden bool,
		includeDeleted bool,
		archivedOnly bool) ([]models.Account, error)
}

type AccountsServiceInstance struct {
	accountsRepo accounts.Repository
}

func NewAccountsService(accountsRepository accounts.Repository) AccountsService {
	return &AccountsServiceInstance{
		accountsRepo: accountsRepository,
	}
}

func (a *AccountsServiceInstance) GetUserAccounts(userId int,
	includeHidden bool,
	includeDeleted bool,
	archivedOnly bool) ([]models.Account, error) {

	userAccounts, err := a.accountsRepo.GetUserAccounts(userId, includeHidden, includeDeleted, archivedOnly)
	if err != nil {
		return nil, err
	}

	return userAccounts, nil
}
