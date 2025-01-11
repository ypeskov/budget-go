package services

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/repositories/accounts"
	"ypeskov/budget-go/internal/repositories/user"
)

type Manager struct {
	UserService     UserService
	AccountsService AccountsService
}

func NewServicesManager(db *database.Database) *Manager {
	userRepo := user.New(db)
	accountsRepo := accounts.NewAccountsService(db)

	return &Manager{
		UserService:     NewUserService(userRepo),
		AccountsService: NewAccountsService(accountsRepo),
	}
}
