package services

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/repositories/user"
)

type Manager struct {
	UserService UserService
}

func NewServicesManager(db *database.Database) *Manager {
	userRepo := user.New(db)

	return &Manager{
		UserService: NewUserService(userRepo),
	}
}
