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
	GetTransactionDetail(transactionId int, userId int) (*dto.TransactionDetailDTO, error)
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

func (s *TransactionsServiceInstance) GetTransactionDetail(transactionId int, userId int) (*dto.TransactionDetailDTO, error) {
	log.Debug("GetTransactionDetail Service")

	transactionRaw, err := s.transactionsRepository.GetTransactionDetail(transactionId, userId)
	if err != nil {
		log.Error("Error getting transaction detail: ", err)
		return nil, err
	}

	if transactionRaw == nil {
		return nil, nil
	}

	baseCurrency, err := s.sm.UserSettingsService.GetBaseCurrency(userId)
	if err != nil {
		log.Error("Error getting base currency: ", err)
		return nil, err
	}

	var baseCurrencyAmount decimal.Decimal
	if transactionRaw.BaseCurrencyAmount != nil {
		baseCurrencyAmount = decimal.NewFromFloat(*transactionRaw.BaseCurrencyAmount)
	} else {
		amount, err := s.sm.ExchangeRatesService.CalcAmountFromCurrency(
			*transactionRaw.DateTime,
			decimal.NewFromFloat(transactionRaw.Amount),
			transactionRaw.Currency.Code,
			baseCurrency.Code,
		)
		if err != nil {
			log.Error("Error calculating base currency amount: ", err)
			baseCurrencyAmount = decimal.Zero
		} else {
			baseCurrencyAmount = amount
		}
	}

	transactionDetail := convertRawToTransactionDetail(transactionRaw, baseCurrency.Code, baseCurrencyAmount)
	return transactionDetail, nil
}

func convertRawToTransactionDetail(raw *dto.TransactionDetailRaw, baseCurrencyCode string, baseCurrencyAmount decimal.Decimal) *dto.TransactionDetailDTO {
	var notes string
	if raw.Notes != nil {
		notes = *raw.Notes
	}

	var newBalance decimal.Decimal
	if raw.NewBalance != nil {
		newBalance = decimal.NewFromFloat(*raw.NewBalance)
	} else {
		newBalance = decimal.Zero
	}

	var initialBalance decimal.Decimal
	if raw.Account.InitialBalance != nil {
		initialBalance = *raw.Account.InitialBalance
	} else {
		initialBalance = decimal.Zero
	}

	var creditLimit decimal.Decimal
	if raw.Account.CreditLimit != nil {
		creditLimit = *raw.Account.CreditLimit
	} else {
		creditLimit = decimal.Zero
	}

	var categoryDetail dto.CategoryDetailDTO
	if raw.Category != nil && raw.Category.ID != nil {
		var createdAt, updatedAt string
		if raw.Category.CreatedAt != nil {
			createdAt = *raw.Category.CreatedAt
		}
		if raw.Category.UpdatedAt != nil {
			updatedAt = *raw.Category.UpdatedAt
		}

		var userId int
		if raw.Category.UserID != nil {
			userId = *raw.Category.UserID
		}

		categoryDetail = dto.CategoryDetailDTO{
			Name:      raw.Category.Name,
			ParentID:  raw.Category.ParentID,
			IsIncome:  raw.Category.IsIncome,
			ID:        *raw.Category.ID,
			UserID:    userId,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Children:  []dto.CategoryDetailDTO{},
		}
	}

	return &dto.TransactionDetailDTO{
		ID:              *raw.ID,
		AccountID:       raw.AccountID,
		TargetAccountID: nil,
		CategoryID:      raw.CategoryID,
		Amount:          decimal.NewFromFloat(raw.Amount),
		TargetAmount:    nil,
		Label:           raw.Label,
		Notes:           notes,
		DateTime:        raw.DateTime,
		IsTransfer:      raw.IsTransfer,
		IsIncome:        raw.IsIncome,
		IsTemplate:      nil,
		UserID:          raw.UserID,
		User: dto.UserDTO{
			Email:     raw.User.Email,
			ID:        raw.User.ID,
			FirstName: raw.User.FirstName,
			LastName:  raw.User.LastName,
		},
		Account: dto.AccountDetailDTO{
			UserID:         raw.Account.UserID,
			AccountTypeID:  raw.Account.AccountTypeId,
			CurrencyID:     raw.Account.CurrencyId,
			InitialBalance: initialBalance,
			Balance:        raw.Account.Balance,
			CreditLimit:    creditLimit,
			Name:           raw.Account.Name,
			OpeningDate:    &raw.Account.OpeningDate,
			Comment:        raw.Account.Comment,
			IsHidden:       raw.Account.IsHidden,
			ShowInReports:  raw.Account.ShowInReports,
			ID:             raw.Account.ID,
			Currency:       raw.Currency,
			AccountType: dto.AccountTypeDetailDTO{
				ID:       raw.AccountType.ID,
				TypeName: raw.AccountType.TypeName,
				IsCredit: raw.AccountType.IsCredit,
			},
			IsDeleted:             raw.Account.IsDeleted,
			IsArchived:            false,
			BalanceInBaseCurrency: decimal.Zero,
			ArchivedAt:            raw.Account.ArchivedAt,
		},
		BaseCurrencyAmount:  baseCurrencyAmount,
		BaseCurrencyCode:    baseCurrencyCode,
		NewBalance:          newBalance,
		Category:            categoryDetail,
		LinkedTransactionID: raw.LinkedTransactionID,
	}
}
