package services

import (
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/currencies"
)

type CurrenciesService interface {
	GetCurrencies() ([]models.Currency, error)
	GetCurrency(id int) (models.Currency, error)
}

type CurrenciesServiceInstance struct {
	currenciesRepo currencies.Repository
}

func NewCurrenciesService(currenciesRepo currencies.Repository) CurrenciesService {
	return &CurrenciesServiceInstance{
		currenciesRepo: currenciesRepo,
	}
}

func (c *CurrenciesServiceInstance) GetCurrencies() ([]models.Currency, error) {
	return c.currenciesRepo.GetCurrencies()
}

func (c *CurrenciesServiceInstance) GetCurrency(id int) (models.Currency, error) {
	return c.currenciesRepo.GetCurrency(id)
}
