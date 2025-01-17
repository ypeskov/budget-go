package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/exchangeRates"
)

type ExchangeRatesService interface {
	GetExchangeRates() (map[string]map[string]decimal.Decimal, error)
	GetExchangeRateByDate(date time.Time) (map[string]decimal.Decimal, error)
	GetRateBetweenCurrencies(date time.Time, currencyFrom string, currencyTo string) (decimal.Decimal, error)
	CalcAmountFromCurrency(date time.Time, amount decimal.Decimal, currencyFrom string, currencyTo string) (decimal.Decimal, error)
}

type ExchangeRatesServiceInstance struct {
	exchangeRatesRepository exchangeRates.Repository
	cache                   *ExchangeRatesHistoryCache
}

const cacheExpiration = time.Hour * 24

type ExchangeRatesHistoryCache struct {
	data       map[string]map[string]decimal.Decimal
	lastUpdate time.Time
}

var (
	instance *ExchangeRatesServiceInstance
	once     sync.Once
	mu       sync.RWMutex
)

func NewExchangeRatesService(exchangeRatesRepository exchangeRates.Repository) ExchangeRatesService {
	once.Do(func() {
		log.Debug("Creating ExchangeRatesService instance")
		instance = &ExchangeRatesServiceInstance{
			exchangeRatesRepository: exchangeRatesRepository,
			cache:                   &ExchangeRatesHistoryCache{data: make(map[string]map[string]decimal.Decimal)},
		}
	})

	return instance
}

func (s *ExchangeRatesServiceInstance) fetchExchangeRates() ([]models.ExchangeRates, error) {
	exchangeRates, err := s.exchangeRatesRepository.GetExchangeRates()
	if err != nil {
		log.Error("Error getting exchange rates: ", err)
		return nil, err
	}

	return exchangeRates, nil
}

func (s *ExchangeRatesServiceInstance) GetExchangeRates() (map[string]map[string]decimal.Decimal, error) {
	if s.isCacheExpired() {
		log.Debug("Cache expired, fetching exchange rates")
		exchangeRates, err := s.fetchExchangeRates()
		if err != nil {
			return nil, err
		}

		s.fillCache(exchangeRates)
	}

	return s.cache.data, nil
}

func (s *ExchangeRatesServiceInstance) fillCache(exchangeRates []models.ExchangeRates) {
	mu.Lock()
	defer mu.Unlock()

	for _, exchangeRate := range exchangeRates {
		convertedRates := make(map[string]decimal.Decimal)

		for key, value := range exchangeRate.Rates {
			switch v := value.(type) {
			case float64:
				convertedRates[key] = decimal.NewFromFloat(v)
			case string:
				decimalValue, err := decimal.NewFromString(v)
				if err != nil {
					log.Warnf("Invalid decimal string for key %s: %v", key, value)
					continue
				}
				convertedRates[key] = decimalValue
			default:
				log.Warnf("Unsupported type for key %s: %T", key, value)
			}
		}

		dateKey := exchangeRate.ActualDate.Format("2006-01-02")
		s.cache.data[dateKey] = convertedRates
	}

	s.cache.lastUpdate = time.Now()
}

func (s *ExchangeRatesServiceInstance) isCacheExpired() bool {
	lastUpdateDate := s.cache.lastUpdate.Truncate(24 * time.Hour)
	currentDate := time.Now().Truncate(24 * time.Hour)

	return lastUpdateDate.Add(cacheExpiration).Before(currentDate)
}

func (s *ExchangeRatesServiceInstance) GetExchangeRateByDate(date time.Time) (map[string]decimal.Decimal, error) {
	dateKey := date.Format("2006-01-02")
	ratesOnDate, ok := s.cache.data[dateKey]
	if !ok {
		log.Warnf("No exchange rates found for date: %s", dateKey)
		return nil, fmt.Errorf("no exchange rates found for date: %s", dateKey)
	}

	return ratesOnDate, nil
}

func (s *ExchangeRatesServiceInstance) GetRateBetweenCurrencies(
	date time.Time,
	currencyFrom string,
	currencyTo string,
) (decimal.Decimal, error) {
	ratesOnDate, err := s.GetExchangeRateByDate(date)
	if err != nil {
		return decimal.Decimal{}, err
	}

	fmt.Println(ratesOnDate)

	rateFrom, ok := ratesOnDate[currencyFrom]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("no exchange rate found for currency: %s", currencyFrom)
	}

	rateTo, ok := ratesOnDate[currencyTo]
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("no exchange rate found for currency: %s", currencyTo)
	}

	return rateFrom.Div(rateTo), nil
}

func (s *ExchangeRatesServiceInstance) CalcAmountFromCurrency(
	date time.Time,
	amount decimal.Decimal,
	currencyFrom string,
	currencyTo string,
) (decimal.Decimal, error) {
	// check if cache is empty or expired
	if len(s.cache.data) == 0 || s.isCacheExpired() {
		log.Debug("Cache is empty or expired, fetching exchange rates")
		_, err := s.GetExchangeRates()
		if err != nil {
			return decimal.Decimal{}, err
		}
	}

	rate, err := s.GetRateBetweenCurrencies(date, currencyFrom, currencyTo)
	if err != nil {
		return decimal.Decimal{}, err
	}

	return amount.Div(rate), nil
}
