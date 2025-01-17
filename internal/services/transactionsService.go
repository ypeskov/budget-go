package services

import (
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/transactions"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type TransactionsService interface {
	GetTransactionsWithAccounts(userId int,
		sm *Manager,
		perPage int,
		page int,
		accountIds []int,
		fromDate time.Time,
		toDate time.Time,
		types []string,
	) ([]dto.TransactionWithAccount, error)
}

type TransactionsServiceInstance struct {
	transactionsRepository transactions.Repository
	sm                     *Manager
}

func NewTransactionsService(transactionsRepository transactions.Repository, sManager *Manager) TransactionsService {
	return &TransactionsServiceInstance{transactionsRepository: transactionsRepository, sm: sManager}
}

func (s *TransactionsServiceInstance) GetTransactionsWithAccounts(userId int,
	sm *Manager,
	perPage int,
	page int,
	accountIds []int,
	fromDate time.Time,
	toDate time.Time,
	types []string,
) ([]dto.TransactionWithAccount, error) {
	log.Debug("GetTransactionsWithAccounts Service")

	transactions, err := s.transactionsRepository.GetTransactionsWithAccounts(userId,
		perPage,
		page,
		accountIds,
		fromDate,
		toDate,
		types,
	)
	if err != nil {
		log.Error("Error getting transactions: ", err)
		return nil, err
	}

	var baseCurrency models.Currency
	baseCurrency, err = s.sm.UserSettingsService.GetBaseCurrency(userId)
	if err != nil {
		log.Error("Error getting base currency: ", err)
		return nil, err
	}

	for i, transaction := range transactions {
		amount, err := s.sm.ExchangeRatesService.CalcAmountFromCurrency(
			*transaction.DateTime,
			decimal.NewFromFloat(transaction.Amount),
			transaction.Currency.Code,
			baseCurrency.Code,
		)
		if err != nil {
			log.Error("Error calculating amount: ", err)
			return nil, err
		}

		amountFloat := amount.InexactFloat64()
		transactions[i].BaseCurrencyAmount = &amountFloat
	}

	return transactions, nil
}
