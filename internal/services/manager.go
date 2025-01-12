package services

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/repositories/accounts"
	"ypeskov/budget-go/internal/repositories/categories"
	"ypeskov/budget-go/internal/repositories/user"
)

type Manager struct {
	UserService       UserService
	AccountsService   AccountsService
	CategoriesService CategoriesService
}

func NewServicesManager(db *database.Database) *Manager {
	userRepo := user.New(db)
	accountsRepo := accounts.NewAccountsService(db)
	categoriesRepo := categories.NewCategoriesRepository(db)

	return &Manager{
		UserService:       NewUserService(userRepo),
		AccountsService:   NewAccountsService(accountsRepo),
		CategoriesService: NewCategoriesService(categoriesRepo),
	}
}
