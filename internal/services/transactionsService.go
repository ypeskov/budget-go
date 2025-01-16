package services

import (
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/repositories/transactions"

	log "github.com/sirupsen/logrus"
)

type TransactionsService interface {
	GetTransactionsWithAccounts(userId int, sm *Manager, perPage int, page int) ([]dto.TransactionWithAccount, error)
}

type TransactionsServiceInstance struct {
	transactionsRepository transactions.Repository
}

func NewTransactionsService(transactionsRepository transactions.Repository) TransactionsService {
	return &TransactionsServiceInstance{transactionsRepository: transactionsRepository}
}

func (s *TransactionsServiceInstance) GetTransactionsWithAccounts(userId int, sm *Manager, perPage int, page int) ([]dto.TransactionWithAccount, error) {
	log.Debug("GetTransactionsWithAccounts Service")

	transactions, err := s.transactionsRepository.GetTransactionsWithAccounts(userId, perPage, page)
	if err != nil {
		log.Error("Error getting transactions: ", err)
		return nil, err
	}

	// TODO: Convert transactions to base currency
	for _, transaction := range transactions {
		var zero float64 = 0
		transaction.BaseCurrencyAmount = &zero
	}

	return transactions, nil
}
