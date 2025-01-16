package services

import (
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/accounts"
)

type AccountsService interface {
	GetUserAccounts(userId int,
		includeHidden bool,
		includeDeleted bool,
		archivedOnly bool) ([]dto.AccountDTO, error)
	GetAccountTypes() ([]models.AccountType, error)
	GetAccountById(id int) (models.Account, error)
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
	archivedOnly bool) ([]dto.AccountDTO, error) {

	userAccounts, err := a.accountsRepo.GetUserAccounts(userId, includeHidden, includeDeleted, archivedOnly)
	if err != nil {
		return nil, err
	}

	return userAccounts, nil
}

func (a *AccountsServiceInstance) GetAccountTypes() ([]models.AccountType, error) {
	accountTypes, err := a.accountsRepo.GetAccountTypes()
	if err != nil {
		return nil, err
	}

	return accountTypes, nil
}

func (a *AccountsServiceInstance) GetAccountById(id int) (models.Account, error) {
	account, err := a.accountsRepo.GetAccountById(id)
	if err != nil {
		return models.Account{}, err
	}
	return account, nil
}
