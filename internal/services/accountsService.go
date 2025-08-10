package services

import (
	"errors"
	"fmt"
	"time"
	"ypeskov/budget-go/internal/dto"
	customErrors "ypeskov/budget-go/internal/errors"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/accounts"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type AccountsService interface {
	GetUserAccounts(userId int,
		sm *Manager,
		includeHidden bool,
		includeDeleted bool,
		archivedOnly bool) ([]dto.AccountDTO, error)
	GetAccountTypes() ([]models.AccountType, error)
	GetAccountById(id int) (*dto.AccountDTO, error)
	CreateAccount(account models.Account) (dto.AccountDTO, error)
	UpdateAccount(account models.Account) (dto.AccountDTO, error)
}

type AccountsServiceInstance struct {
	accountsRepo accounts.Repository
	sm           *Manager
}

func NewAccountsService(accountsRepository accounts.Repository, sManager *Manager) AccountsService {
	return &AccountsServiceInstance{
		accountsRepo: accountsRepository,
		sm:           sManager,
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
			zero := decimal.Zero
			userAccounts[i].InitialBalance = &zero
		}

		if account.CreditLimit == nil {
			zero := decimal.Zero
			userAccounts[i].CreditLimit = &zero
		}

		amount, err := sm.ExchangeRatesService.CalcAmountFromCurrency(
			time.Now(),
			account.Balance,
			account.Currency.Code,
			baseCurrency.Code,
		)
		if err != nil {
			return nil, err
		}
		userAccounts[i].BalanceInBaseCurrency = &amount
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

func (a *AccountsServiceInstance) GetAccountById(id int) (*dto.AccountDTO, error) {
	account, err := a.accountsRepo.GetAccountById(id)
	if err != nil {
		return nil, err
	}

	// Ensure numeric fields are not nil to avoid nulls in JSON
	if account.InitialBalance == nil {
		zero := decimal.NewFromFloat(0)
		account.InitialBalance = &zero
	}
	if account.CreditLimit == nil {
		zero := decimal.NewFromFloat(0)
		account.CreditLimit = &zero
	}

	// Enrich with currency details and balance in base currency
	accountDTO, err := buildAccountDTO(account)
	if err != nil {
		return nil, err
	}

	// Enrich account type details (type_name, is_credit)
	accountTypes, err := a.accountsRepo.GetAccountTypes()
	if err != nil {
		return nil, err
	}

	for _, at := range accountTypes {
		if at.ID == accountDTO.AccountTypeId {
			accountDTO.AccountType.TypeName = at.TypeName
			accountDTO.AccountType.IsCredit = at.IsCredit
			break
		}
	}

	return &accountDTO, nil
}

func (a *AccountsServiceInstance) CreateAccount(account models.Account) (dto.AccountDTO, error) {
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	if account.InitialBalance == nil {
		zero := decimal.NewFromFloat(0)
		account.InitialBalance = &zero
	}

	if account.CreditLimit == nil {
		zero := decimal.NewFromFloat(0)
		account.CreditLimit = &zero
	}

	newAccount, err := a.accountsRepo.CreateAccount(account)
	if err != nil {
		return dto.AccountDTO{}, err
	}

	accountDto, err := buildAccountDTO(newAccount)
	if err != nil {
		return dto.AccountDTO{}, err
	}

	return accountDto, nil
}

func buildAccountDTO(account models.Account) (dto.AccountDTO, error) {
	accountDto := dto.AccountToDTO(account)

	baseCurrency, err := sm.UserSettingsService.GetBaseCurrency(accountDto.UserID)
	if err != nil {
		return dto.AccountDTO{}, err
	}

	accountCurrency, err := sm.CurrenciesService.GetCurrency(accountDto.CurrencyId)
	if err != nil {
		return dto.AccountDTO{}, err
	}

	accountDto.Currency = *buildCurrencyDTO(accountCurrency)

	amount, err := sm.ExchangeRatesService.CalcAmountFromCurrency(
		time.Now(),
		accountDto.Balance,
		accountCurrency.Code,
		baseCurrency.Code,
	)
	if err != nil {
		return dto.AccountDTO{}, fmt.Errorf("failed to calculate amount from currency: %w", err)
	}

	accountDto.BalanceInBaseCurrency = &amount

	return accountDto, nil
}

func (a *AccountsServiceInstance) UpdateAccount(account models.Account) (dto.AccountDTO, error) {
	log.Debug("UpdateAccount Service")
	account.UpdatedAt = time.Now()

	if account.InitialBalance == nil {
		zero := decimal.NewFromFloat(0)
		account.InitialBalance = &zero
	}

	if account.CreditLimit == nil {
		zero := decimal.NewFromFloat(0)
		account.CreditLimit = &zero
	}

	updatedAccount, err := a.accountsRepo.UpdateAccount(account)
	if err != nil {
		if errors.Is(err, customErrors.ErrNoAccountFound) {
			log.Errorf("No account found with the provided ID: %v", account.ID)
			return dto.AccountDTO{}, customErrors.ErrNoAccountFound
		}
		return dto.AccountDTO{}, err
	}

	accountDto, err := buildAccountDTO(updatedAccount)
	if err != nil {
		return dto.AccountDTO{}, err
	}

	return accountDto, nil
}

func buildCurrencyDTO(currency models.Currency) *dto.CurrencyDTO {
	fmt.Println(currency)
	fmt.Println("--------------------------------")
	return &dto.CurrencyDTO{
		ID:        currency.ID,
		Code:      currency.Code,
		Name:      currency.Name,
		CreatedAt: currency.CreatedAt.Format(time.RFC3339),
		UpdatedAt: currency.UpdatedAt.Format(time.RFC3339),
	}
}
