package services

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/repositories/accounts"
	"ypeskov/budget-go/internal/repositories/categories"
	"ypeskov/budget-go/internal/repositories/currencies"
	"ypeskov/budget-go/internal/repositories/exchangeRates"
	"ypeskov/budget-go/internal/repositories/transactions"
	"ypeskov/budget-go/internal/repositories/user"
	"ypeskov/budget-go/internal/repositories/userSettings"
)

type Manager struct {
	UserService          UserService
	AccountsService      AccountsService
	CategoriesService    CategoriesService
	UserSettingsService  UserSettingsService
	CurrenciesService    CurrenciesService
	TransactionsService  TransactionsService
	ExchangeRatesService ExchangeRatesService
}

var sm *Manager

func NewServicesManager(db *database.Database) *Manager {
	userRepo := user.New(db)
	exchangeRatesRepo := exchangeRates.NewExchangeRatesRepository(db.Db)
	accountsRepo := accounts.NewAccountsService(db.Db)
	categoriesRepo := categories.NewCategoriesRepository(db.Db)
	userSettingsRepo := userSettings.NewUserSettingsRepository(db.Db)
	currenciesRepo := currencies.NewCurrenciesRepository(db.Db)
	transactionsRepo := transactions.NewTransactionsRepository(db.Db)

	sm = &Manager{}

	sm.UserService = NewUserService(userRepo)
	sm.AccountsService = NewAccountsService(accountsRepo, sm)
	sm.CategoriesService = NewCategoriesService(categoriesRepo)
	sm.UserSettingsService = NewUserSettingsService(userSettingsRepo)
	sm.CurrenciesService = NewCurrenciesService(currenciesRepo)
	sm.ExchangeRatesService = NewExchangeRatesService(exchangeRatesRepo)
	sm.TransactionsService = NewTransactionsService(transactionsRepo, sm)

	return sm
}
