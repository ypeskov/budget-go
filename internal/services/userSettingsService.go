package services

import (
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/userSettings"

	log "github.com/sirupsen/logrus"
)

type UserSettingsService interface {
	GetBaseCurrency(userId int) (models.Currency, error)
}

type UserSettingsServiceInstance struct {
	userSettingsRepo userSettings.Repository
}

func NewUserSettingsService(userSettingsRepo userSettings.Repository) UserSettingsService {
	return &UserSettingsServiceInstance{
		userSettingsRepo: userSettingsRepo,
	}
}

func (u *UserSettingsServiceInstance) GetBaseCurrency(userId int) (models.Currency, error) {
	baseCurrency, err := u.userSettingsRepo.GetBaseCurrency(userId)
	if err != nil {
		log.Error("Failed to get base currency: ", err)
		return models.Currency{}, err
	}

	return baseCurrency, nil
}
