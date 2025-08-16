package services

import (
	"sync"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/userSettings"

	log "github.com/sirupsen/logrus"
)

type UserSettingsService interface {
	GetBaseCurrency(userId int) (models.Currency, error)
	UpdateUserSettings(userID int, settingsData map[string]interface{}) (*models.UserSettings, error)
	GetUserSettings(userID int) (*models.UserSettings, error)
}

type UserSettingsServiceInstance struct {
	userSettingsRepo userSettings.Repository
}

var (
	userSettingsInstance *UserSettingsServiceInstance
	userSettingsOnce     sync.Once
)

func NewUserSettingsService(userSettingsRepo userSettings.Repository) UserSettingsService {
	userSettingsOnce.Do(func() {
		log.Debug("Creating UserSettingsService instance")
		userSettingsInstance = &UserSettingsServiceInstance{
			userSettingsRepo: userSettingsRepo,
		}
	})

	return userSettingsInstance
}

func (u *UserSettingsServiceInstance) GetBaseCurrency(userId int) (models.Currency, error) {
	baseCurrency, err := u.userSettingsRepo.GetBaseCurrency(userId)
	if err != nil {
		log.Error("Failed to get base currency: ", err)
		return models.Currency{}, err
	}

	return baseCurrency, nil
}

func (u *UserSettingsServiceInstance) UpdateUserSettings(userID int, settingsData map[string]interface{}) (*models.UserSettings, error) {
	userSettings, err := u.userSettingsRepo.UpsertUserSettings(userID, settingsData)
	if err != nil {
		log.Error("Failed to update user settings: ", err)
		return nil, err
	}

	return userSettings, nil
}

func (u *UserSettingsServiceInstance) GetUserSettings(userID int) (*models.UserSettings, error) {
	userSettings, err := u.userSettingsRepo.GetUserSettings(userID)
	if err != nil {
		log.Error("Failed to get user settings: ", err)
		return nil, err
	}

	return userSettings, nil
}
