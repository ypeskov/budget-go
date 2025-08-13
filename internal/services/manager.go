package services

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/repositories/accounts"
	"ypeskov/budget-go/internal/repositories/budgets"
	"ypeskov/budget-go/internal/repositories/categories"
	"ypeskov/budget-go/internal/repositories/currencies"
	"ypeskov/budget-go/internal/repositories/exchangeRates"
	"ypeskov/budget-go/internal/repositories/languages"
	"ypeskov/budget-go/internal/repositories/reports"
	"ypeskov/budget-go/internal/repositories/transactions"
	"ypeskov/budget-go/internal/repositories/user"
	"ypeskov/budget-go/internal/repositories/userSettings"
)

type Manager struct {
	UserService          UserService
	AccountsService      AccountsService
	BudgetsService       BudgetsService
	CategoriesService    CategoriesService
	UserSettingsService  UserSettingsService
	CurrenciesService    CurrenciesService
	LanguagesService     LanguagesService
	TransactionsService  TransactionsService
	ExchangeRatesService ExchangeRatesService
	ReportsService       ReportsService
	ChartService         ChartService
}

var sm *Manager

func NewServicesManager(db *database.Database) *Manager {
	userRepo := user.New(db)
	exchangeRatesRepo := exchangeRates.NewExchangeRatesRepository(db.Db)
	accountsRepo := accounts.NewAccountsService(db.Db)
	budgetsRepo := budgets.NewBudgetsRepository(db.Db)
	categoriesRepo := categories.NewCategoriesRepository(db.Db)
	userSettingsRepo := userSettings.NewUserSettingsRepository(db.Db)
	currenciesRepo := currencies.NewCurrenciesRepository(db.Db)
	languagesRepo := languages.NewLanguagesRepository(db.Db)
	transactionsRepo := transactions.NewTransactionsRepository(db.Db)
	reportsRepo := reports.NewReportsRepository(db.Db)

	sm = &Manager{}

	sm.UserService = NewUserService(userRepo)
	sm.AccountsService = NewAccountsService(accountsRepo, sm)
	sm.BudgetsService = NewBudgetsService(budgetsRepo, sm)
	sm.CategoriesService = NewCategoriesService(categoriesRepo)
	sm.UserSettingsService = NewUserSettingsService(userSettingsRepo)
	sm.CurrenciesService = NewCurrenciesService(currenciesRepo)
	sm.LanguagesService = NewLanguagesService(languagesRepo)
	sm.ExchangeRatesService = NewExchangeRatesService(exchangeRatesRepo)
	sm.TransactionsService = NewTransactionsService(transactionsRepo, sm)
	sm.ReportsService = NewReportsService(reportsRepo, sm.ExchangeRatesService)
	sm.ChartService = NewChartService()

	return sm
}
