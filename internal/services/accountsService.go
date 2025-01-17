package services

import (
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/accounts"

	"github.com/shopspring/decimal"
)

type AccountsService interface {
	GetUserAccounts(userId int,
		sm *Manager,
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

func (a *AccountsServiceInstance) GetUserAccounts(
	userId int,
	sm *Manager,
	includeHidden bool,
	includeDeleted bool,
	archivedOnly bool) ([]dto.AccountDTO, error) {

	userAccounts, err := a.accountsRepo.GetUserAccounts(userId, includeHidden, includeDeleted, archivedOnly)
	if err != nil {
		return nil, err
	}

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(userId)
	if err != nil {
		return nil, err
	}

	for i, account := range userAccounts {
		if account.InitialBalance == nil {
			zero := 0.0
			userAccounts[i].InitialBalance = &zero
		}

		if account.CreditLimit == nil {
			zero := 0.0
			userAccounts[i].CreditLimit = &zero
		}

		amount, err := sm.ExchangeRatesService.CalcAmountFromCurrency(
			time.Now(),
			decimal.NewFromFloat(account.Balance),
			account.Currency.Code,
			baseCurrency.Code,
		)
		if err != nil {
			return nil, err
		}
		amountValue := amount.InexactFloat64()
		userAccounts[i].BalanceInBaseCurrency = &amountValue
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
