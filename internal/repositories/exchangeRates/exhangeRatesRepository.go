package exchangeRates

import (
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var db *sqlx.DB

type Repository interface {
	GetExchangeRates() ([]models.ExchangeRates, error)
	SaveExchangeRates(rates *models.ExchangeRates) error
	DeleteExchangeRatesByDate(date string) error
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

func (r *RepositoryInstance) SaveExchangeRates(rates *models.ExchangeRates) error {
	log.Debug("SaveExchangeRates Repository")
	const query = `
		INSERT INTO exchange_rates (rates, actual_date, base_currency_code, service_name, is_deleted, created_at, updated_at)
		VALUES (:rates, :actual_date, :base_currency_code, :service_name, :is_deleted, :created_at, :updated_at)
	`

	_, err := db.NamedExec(query, rates)
	if err != nil {
		log.Error("Error saving exchange rates: ", err)
		return err
	}

	log.Debug("Exchange rates saved successfully")
	return nil
}

func (r *RepositoryInstance) DeleteExchangeRatesByDate(date string) error {
	log.Debug("DeleteExchangeRatesByDate Repository")
	const query = `DELETE FROM exchange_rates WHERE actual_date = $1`

	result, err := db.Exec(query, date)
	if err != nil {
		log.Error("Error deleting exchange rates: ", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Debugf("Deleted %d exchange rate records for date %s", rowsAffected, date)
	return nil
}
