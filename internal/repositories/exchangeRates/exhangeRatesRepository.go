package exchangeRates

import (
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var db *sqlx.DB

type Repository interface {
	GetExchangeRates() ([]models.ExchangeRates, error)
}

type RepositoryInstance struct{}

func NewExchangeRatesRepository(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (r *RepositoryInstance) GetExchangeRates() ([]models.ExchangeRates, error) {
	log.Debug("GetExchangeRates Repository")
	const query = `
	SELECT id AS id, rates AS rates, actual_date AS actual_date, base_currency_code AS base_currency_code, 
			service_name AS service_name, is_deleted AS is_deleted, created_at AS created_at, updated_at AS updated_at
	FROM exchange_rates
	WHERE is_deleted = false
	`

	var exchangeRates []models.ExchangeRates
	err := db.Select(&exchangeRates, query)
	if err != nil {
		log.Error("Error getting exchange rates: ", err)
		return nil, err
	}

	return exchangeRates, nil
}
