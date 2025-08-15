package currencies

import (
	"ypeskov/budget-go/internal/models"

	log "github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetCurrencies() ([]models.Currency, error)
	GetCurrency(id int) (models.Currency, error)
	GetCurrencyByCode(code string) (models.Currency, error)
}

type RepositoryInstance struct{}

var db *sqlx.DB

func NewCurrenciesRepository(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (r *RepositoryInstance) GetCurrencies() ([]models.Currency, error) {
	const getCurrenciesQuery = `SELECT id, code, name FROM currencies WHERE is_deleted = false;`

	var currencies []models.Currency
	err := db.Select(&currencies, getCurrenciesQuery)
	if err != nil {
		log.Error("Failed to get currencies: ", err)
		return nil, err
	}
	return currencies, nil
}

func (r *RepositoryInstance) GetCurrency(id int) (models.Currency, error) {
	const getCurrencyQuery = `SELECT id, code, name, created_at, updated_at FROM currencies WHERE id = $1 AND is_deleted = false;`

	var currency models.Currency
	err := db.Get(&currency, getCurrencyQuery, id)
	if err != nil {
		return models.Currency{}, err
	}

	return currency, nil
}

func (r *RepositoryInstance) GetCurrencyByCode(code string) (models.Currency, error) {
	const getCurrencyByCodeQuery = `SELECT id, code, name, created_at, updated_at
FROM currencies WHERE code = $1 AND is_deleted = false;`

	var currency models.Currency
	err := db.Get(&currency, getCurrencyByCodeQuery, code)
	if err != nil {
		return models.Currency{}, err
	}

	return currency, nil
}
