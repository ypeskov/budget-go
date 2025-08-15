package services

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/exchangeRates"
)

type ExchangeRatesService interface {
	GetExchangeRates() (map[string]map[string]decimal.Decimal, error)
	GetExchangeRateByDate(date time.Time) (map[string]decimal.Decimal, error)
	GetRateBetweenCurrencies(date time.Time, currencyFrom string, currencyTo string) (decimal.Decimal, error)
	CalcAmountFromCurrency(date time.Time, amount decimal.Decimal, currencyFrom string, currencyTo string) (decimal.Decimal, error)
	UpdateExchangeRates(date time.Time) (*models.ExchangeRates, error)
}

type ExchangeRatesServiceInstance struct {
	exchangeRatesRepository exchangeRates.Repository
	cache                   *ExchangeRatesHistoryCache
	currencyBeaconService   *CurrencyBeaconService
	config                  *config.Config
}

const cacheExpiration = time.Hour * 24

type ExchangeRatesHistoryCache struct {
	data           map[string]map[string]decimal.Decimal
	baseCurrencies map[string]string // maps date -> base currency code
	lastUpdate     time.Time
}

var (
	instance *ExchangeRatesServiceInstance
	once     sync.Once
	mu       sync.RWMutex
)

func NewExchangeRatesService(exchangeRatesRepository exchangeRates.Repository, cfg *config.Config) ExchangeRatesService {
	once.Do(func() {
		log.Debug("Creating ExchangeRatesService instance")
		instance = &ExchangeRatesServiceInstance{
			exchangeRatesRepository: exchangeRatesRepository,
			cache: &ExchangeRatesHistoryCache{
				data:           make(map[string]map[string]decimal.Decimal),
				baseCurrencies: make(map[string]string),
			},
			currencyBeaconService: NewCurrencyBeaconService(cfg),
			config:                cfg,
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
		s.cache.baseCurrencies[dateKey] = exchangeRate.BaseCurrencyCode
	}

	s.cache.lastUpdate = time.Now()
}

func (s *ExchangeRatesServiceInstance) isCacheExpired() bool {
	lastUpdateDate := s.cache.lastUpdate.Truncate(24 * time.Hour)
	currentDate := time.Now().Truncate(24 * time.Hour)

	return lastUpdateDate.Add(cacheExpiration).Before(currentDate)
}

func (s *ExchangeRatesServiceInstance) GetExchangeRateByDate(date time.Time) (map[string]decimal.Decimal, error) {
	mu.RLock()
	defer mu.RUnlock()

	// get all date keys from cache that are before or equal to the date
	dateKeys := make([]string, 0, len(s.cache.data))
	for key := range s.cache.data {
		keyDate, err := time.Parse(time.DateOnly, key)
		if err != nil {
			log.Warnf("Invalid date format in cache key: %s", key)
			continue
		}
		if !keyDate.After(date) {
			dateKeys = append(dateKeys, key)
		}
	}

	// sort the date keys in descending order
	sort.Sort(sort.Reverse(sort.StringSlice(dateKeys)))

	// get the first cache value that is before or equal to the date
	for _, key := range dateKeys {
		keyDate, err := time.Parse(time.DateOnly, key)
		if err != nil {
			log.Warnf("Invalid date format in cache key: %s", key)
			continue
		}

		if !keyDate.After(date) {
			return s.cache.data[key], nil
		}
	}

	err := fmt.Errorf("no exchange rates found for any prior date starting from: %s", date.Format(time.DateOnly))
	log.Error(err)
	return nil, err
}

func (s *ExchangeRatesServiceInstance) GetRateBetweenCurrencies(
	date time.Time,
	currencyFrom string,
	currencyTo string,
) (decimal.Decimal, error) {
	// If both currencies are the same, return 1
	if currencyFrom == currencyTo {
		return decimal.NewFromInt(1), nil
	}

	ratesOnDate, err := s.GetExchangeRateByDate(date)
	if err != nil {
		return decimal.Decimal{}, err
	}

	// Get the base currency for this date
	dateKey := s.getDateKeyForRates(date)
	if dateKey == "" {
		return decimal.Decimal{}, fmt.Errorf("no exchange rates found for date: %s", date.Format("2006-01-02"))
	}

	baseCurrency := s.cache.baseCurrencies[dateKey]
	
	// Handle conversions based on the base currency
	if currencyFrom == baseCurrency {
		// Converting from base currency to another currency
		rateTo, ok := ratesOnDate[currencyTo]
		if !ok {
			log.Warnf("No exchange rate found for currency: %s", currencyTo)
			return decimal.Decimal{}, fmt.Errorf("no exchange rate found for currency: %s", currencyTo)
		}
		return rateTo, nil
	} else if currencyTo == baseCurrency {
		// Converting from another currency to base currency
		rateFrom, ok := ratesOnDate[currencyFrom]
		if !ok {
			log.Warnf("No exchange rate found for currency: %s", currencyFrom)
			return decimal.Decimal{}, fmt.Errorf("no exchange rate found for currency: %s", currencyFrom)
		}
		return decimal.NewFromInt(1).Div(rateFrom), nil
	} else {
		// Converting between two non-base currencies
		rateFrom, ok := ratesOnDate[currencyFrom]
		if !ok {
			log.Warnf("No exchange rate found for currency: %s", currencyFrom)
			return decimal.Decimal{}, fmt.Errorf("no exchange rate found for currency: %s", currencyFrom)
		}

		rateTo, ok := ratesOnDate[currencyTo]
		if !ok {
			log.Warnf("No exchange rate found for currency: %s", currencyTo)
			return decimal.Decimal{}, fmt.Errorf("no exchange rate found for currency: %s", currencyTo)
		}

		// Convert through base currency: (1/rateFrom) * rateTo
		return rateTo.Div(rateFrom), nil
	}
}

// getDateKeyForRates finds the appropriate date key for the given date
func (s *ExchangeRatesServiceInstance) getDateKeyForRates(date time.Time) string {
	mu.RLock()
	defer mu.RUnlock()

	// Get all date keys from cache that are before or equal to the date
	dateKeys := make([]string, 0, len(s.cache.data))
	for key := range s.cache.data {
		keyDate, err := time.Parse(time.DateOnly, key)
		if err != nil {
			log.Warnf("Invalid date format in cache key: %s", key)
			continue
		}
		if !keyDate.After(date) {
			dateKeys = append(dateKeys, key)
		}
	}

	if len(dateKeys) == 0 {
		return ""
	}

	// Sort the date keys in descending order and return the first (most recent)
	sort.Sort(sort.Reverse(sort.StringSlice(dateKeys)))
	return dateKeys[0]
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

	return amount.Mul(rate), nil
}

func (s *ExchangeRatesServiceInstance) UpdateExchangeRates(date time.Time) (*models.ExchangeRates, error) {
	log.Infof("Updating exchange rates for %s", date.Format("2006-01-02"))

	// Get exchange rates from external API
	dateStr := date.Format("2006-01-02")
	ratesData, err := s.currencyBeaconService.GetCurrencyRates(dateStr)
	if err != nil {
		log.Errorf("Failed to fetch exchange rates from CurrencyBeacon: %v", err)
		return nil, err
	}

	// Delete existing rates for this date
	err = s.exchangeRatesRepository.DeleteExchangeRatesByDate(dateStr)
	if err != nil {
		log.Errorf("Failed to delete existing exchange rates for %s: %v", dateStr, err)
		return nil, err
	}

	// Parse the actual_date from API response
	actualDate, err := time.Parse("2006-01-02", ratesData["actual_date"].(string))
	if err != nil {
		log.Errorf("Failed to parse actual_date from API response: %v", err)
		return nil, err
	}

	// Create new exchange rates model
	excRates := &models.ExchangeRates{
		Rates:            models.JSONB(ratesData["rates"].(map[string]interface{})),
		ActualDate:       actualDate,
		BaseCurrencyCode: ratesData["base_currency_code"].(string),
		ServiceName:      ratesData["service_name"].(string),
		IsDeleted:        false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to database
	err = s.exchangeRatesRepository.SaveExchangeRates(excRates)
	if err != nil {
		log.Errorf("Failed to save exchange rates: %v", err)
		return nil, err
	}

	// Clear cache to force refresh
	mu.Lock()
	s.cache.data = make(map[string]map[string]decimal.Decimal)
	s.cache.baseCurrencies = make(map[string]string)
	s.cache.lastUpdate = time.Time{}
	mu.Unlock()

	log.Infof("Exchange rates updated successfully for %s", date.Format("2006-01-02"))
	return excRates, nil
}
