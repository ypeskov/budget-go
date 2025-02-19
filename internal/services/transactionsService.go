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
		tratypes []string,
		categoryIds []int,
	) ([]dto.TransactionWithAccount, error)
	GetTemplates(userId int) ([]dto.TemplateDTO, error)
	DeleteTemplates(templateIds []int, userId int) error
	CreateTransaction(transaction models.Transaction) error
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
	transactionTypes []string,
	categoryIds []int,
) ([]dto.TransactionWithAccount, error) {
	log.Debug("GetTransactionsWithAccounts Service")

	transactions, err := s.transactionsRepository.GetTransactionsWithAccounts(userId,
		perPage,
		page,
		accountIds,
		fromDate,
		toDate,
		transactionTypes,
		categoryIds,
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

func (s *TransactionsServiceInstance) GetTemplates(userId int) ([]dto.TemplateDTO, error) {
	log.Debug("GetTemplates Service")

	templates, err := s.transactionsRepository.GetTemplates(userId)
	if err != nil {
		log.Error("Error getting templates: ", err)
		return nil, err
	}

	return templates, nil
}

func (s *TransactionsServiceInstance) DeleteTemplates(templateIds []int, userId int) error {
	log.Debug("DeleteTemplates Service")

	err := s.transactionsRepository.DeleteTemplates(templateIds, userId)
	if err != nil {
		log.Error("Error deleting templates: ", err)
		return err
	}

	return nil
}

func (s *TransactionsServiceInstance) CreateTransaction(transaction models.Transaction) error {
	log.Debug("CreateTransaction Service")

	if transaction.DateTime == nil {
		now := time.Now()
		transaction.DateTime = &now
	}

	if transaction.CreatedAt == nil {
		now := time.Now()
		transaction.CreatedAt = &now
	}

	if transaction.UpdatedAt == nil {
		now := time.Now()
		transaction.UpdatedAt = &now
	}

	if transaction.Notes == nil {
		emptyString := ""
		transaction.Notes = &emptyString
	}

	err := s.transactionsRepository.CreateTransaction(transaction)
	if err != nil {
		log.Error("Error creating transaction: ", err)
		return err
	}

	return nil
}
