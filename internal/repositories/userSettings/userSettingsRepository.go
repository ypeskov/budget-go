package userSettings

import (
	"encoding/json"
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetBaseCurrency(userId int) (models.Currency, error)
	UpsertUserSettings(userID int, settingsData map[string]interface{}) (*models.UserSettings, error)
}

type RepositoryInstance struct{}

var (
	db *sqlx.DB
)

func NewUserSettingsRepository(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (r *RepositoryInstance) GetBaseCurrency(userId int) (models.Currency, error) {
	const getBaseCurrencyQuery = `
SELECT base_currency_id as id, currencies.code, currencies.name
FROM users 
JOIN currencies ON users.base_currency_id = currencies.id 
WHERE users.id = $1;`

	var baseCurrency models.Currency
	err := db.Get(&baseCurrency, getBaseCurrencyQuery, userId)
	if err != nil {
		return models.Currency{}, err
	}

	return baseCurrency, nil
}

func (r *RepositoryInstance) UpsertUserSettings(userID int, settingsData map[string]interface{}) (*models.UserSettings, error) {
	settingsJSON, err := json.Marshal(settingsData)
	if err != nil {
		log.Error("Failed to marshal settings data: ", err)
		return nil, err
	}

	// First, try to update existing record
	const updateSettingsQuery = `
		UPDATE user_settings 
		SET settings = $2, updated_at = NOW() 
		WHERE user_id = $1
		RETURNING id, user_id, settings, created_at, updated_at;`

	var userSettings models.UserSettings
	var settingsStr string
	
	err = db.QueryRow(updateSettingsQuery, userID, string(settingsJSON)).Scan(
		&userSettings.ID,
		&userSettings.UserID,
		&settingsStr,
		&userSettings.CreatedAt,
		&userSettings.UpdatedAt,
	)

	// If no rows were affected (user settings don't exist), insert new record
	if err != nil {
		const insertSettingsQuery = `
			INSERT INTO user_settings (user_id, settings, created_at, updated_at) 
			VALUES ($1, $2, NOW(), NOW())
			RETURNING id, user_id, settings, created_at, updated_at;`

		err = db.QueryRow(insertSettingsQuery, userID, string(settingsJSON)).Scan(
			&userSettings.ID,
			&userSettings.UserID,
			&settingsStr,
			&userSettings.CreatedAt,
			&userSettings.UpdatedAt,
		)
		if err != nil {
			log.Error("Failed to insert user settings: ", err)
			return nil, err
		}
	}

	// Parse the settings JSON string back into map
	err = json.Unmarshal([]byte(settingsStr), &userSettings.Settings)
	if err != nil {
		log.Error("Failed to unmarshal settings: ", err)
		return nil, err
	}

	return &userSettings, nil
}
