package userSettings

import (
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetBaseCurrency(userId int) (models.Currency, error)
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
