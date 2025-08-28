package services

import (
	"sync"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/currencies"
)

type CurrenciesService interface {
	GetCurrencies() ([]models.Currency, error)
	GetCurrency(id int) (models.Currency, error)
	GetCurrencyByCode(code string) (models.Currency, error)
}

type CurrenciesServiceInstance struct {
	currenciesRepo currencies.Repository
}

var (
	currenciesInstance *CurrenciesServiceInstance
	currenciesOnce     sync.Once
)

func NewCurrenciesService(currenciesRepo currencies.Repository) CurrenciesService {
	currenciesOnce.Do(func() {
		logger.Debug("Creating CurrenciesService instance")
		currenciesInstance = &CurrenciesServiceInstance{
			currenciesRepo: currenciesRepo,
		}
	})

	return currenciesInstance
}

func (c *CurrenciesServiceInstance) GetCurrencies() ([]models.Currency, error) {
	return c.currenciesRepo.GetCurrencies()
}

func (c *CurrenciesServiceInstance) GetCurrency(id int) (models.Currency, error) {
	currency, err := c.currenciesRepo.GetCurrency(id)
	if err != nil {
		return models.Currency{}, err
	}

	return currency, nil
}

func (c *CurrenciesServiceInstance) GetCurrencyByCode(code string) (models.Currency, error) {
	currency, err := c.currenciesRepo.GetCurrencyByCode(code)
	if err != nil {
		return models.Currency{}, err
	}

	return currency, nil
}
