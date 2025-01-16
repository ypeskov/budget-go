package services

import (
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/repositories/transactions"
)

type TransactionsService interface {
	GetTransactionsWithAccounts(userId int) ([]dto.TransactionWithAccount, error)
}

type TransactionsServiceInstance struct {
	transactionsRepository transactions.Repository
}

func NewTransactionsService(transactionsRepository transactions.Repository) TransactionsService {
	return &TransactionsServiceInstance{transactionsRepository: transactionsRepository}
}

func (s *TransactionsServiceInstance) GetTransactionsWithAccounts(userId int) ([]dto.TransactionWithAccount, error) {
	return s.transactionsRepository.GetTransactionsWithAccounts(userId)
}
