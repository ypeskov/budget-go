package services

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/repositories/accounts"
	"ypeskov/budget-go/internal/repositories/categories"
	"ypeskov/budget-go/internal/repositories/currencies"
	"ypeskov/budget-go/internal/repositories/transactions"
	"ypeskov/budget-go/internal/repositories/user"
	"ypeskov/budget-go/internal/repositories/userSettings"
)

type Manager struct {
	UserService         UserService
	AccountsService     AccountsService
	CategoriesService   CategoriesService
	UserSettingsService UserSettingsService
	CurrenciesService   CurrenciesService
	TransactionsService TransactionsService
}

func NewServicesManager(db *database.Database) *Manager {
	userRepo := user.New(db)
	accountsRepo := accounts.NewAccountsService(db.Db)
	categoriesRepo := categories.NewCategoriesRepository(db.Db)
	userSettingsRepo := userSettings.NewUserSettingsRepository(db.Db)
	currenciesRepo := currencies.NewCurrenciesRepository(db.Db)
	transactionsRepo := transactions.NewTransactionsRepository(db.Db)

	return &Manager{
		UserService:         NewUserService(userRepo),
		AccountsService:     NewAccountsService(accountsRepo),
		CategoriesService:   NewCategoriesService(categoriesRepo),
		UserSettingsService: NewUserSettingsService(userSettingsRepo),
		CurrenciesService:   NewCurrenciesService(currenciesRepo),
		TransactionsService: NewTransactionsService(transactionsRepo),
	}
}
