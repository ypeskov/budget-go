package services

import (
	"fmt"
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
	UpdateTransaction(transactionDTO dto.PutTransactionDTO, userId int) error
	DeleteTransaction(transactionId int, userId int) error
	GetTemplates(userId int) ([]dto.TemplateDTO, error)
	DeleteTemplates(templateIds []int, userId int) error
	CreateTransaction(transaction models.Transaction, targetAccountID *int, targetAmount *decimal.Decimal) (*models.Transaction, error)
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
			transaction.Amount,
			transaction.Currency.Code,
			baseCurrency.Code,
		)
		if err != nil {
			log.Error("Error calculating amount: ", err)
			return nil, err
		}

		transactions[i].BaseCurrencyAmount = &amount
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

func (s *TransactionsServiceInstance) CreateTransaction(transaction models.Transaction, targetAccountID *int, targetAmount *decimal.Decimal) (*models.Transaction, error) {
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

	if transaction.IsTransfer {
		return s.createTransferTransaction(transaction, targetAccountID, targetAmount)
	} else {
		return s.createRegularTransaction(transaction)
	}
}

func (s *TransactionsServiceInstance) createRegularTransaction(transaction models.Transaction) (*models.Transaction, error) {
	log.Debug("Creating regular transaction")

	// Get current account balance
	currentBalance, err := s.sm.AccountsService.GetAccountBalance(transaction.AccountID)
	if err != nil {
		log.Error("Error getting current account balance: ", err)
		return nil, err
	}

	// Calculate transaction effect and new balance
	effect := s.calculateTransactionEffect(transaction.AccountID, transaction.Amount, transaction.IsIncome, transaction.IsTransfer, false)
	newBalance := currentBalance.Add(effect)
	
	// Set the new balance in the transaction record
	transaction.NewBalance = &newBalance

	// Update account balance
	err = s.updateAccountBalanceByEffect(transaction.AccountID, effect)
	if err != nil {
		log.Error("Error updating account balance for new transaction: ", err)
		return nil, err
	}

	createdTransaction, err := s.transactionsRepository.CreateTransaction(transaction)
	if err != nil {
		log.Error("Error creating transaction: ", err)
		// Rollback balance change if transaction creation failed
		rollbackErr := s.updateAccountBalanceByEffect(transaction.AccountID, effect.Neg())
		if rollbackErr != nil {
			log.Error("Error rolling back account balance: ", rollbackErr)
		}
		return nil, err
	}

	return createdTransaction, nil
}

func (s *TransactionsServiceInstance) createTransferTransaction(transaction models.Transaction, targetAccountID *int, targetAmount *decimal.Decimal) (*models.Transaction, error) {
	log.Debug("Creating transfer transaction")

	if targetAccountID == nil {
		return nil, fmt.Errorf("target account ID is required for transfer transactions")
	}

	if targetAmount == nil {
		targetAmount = &transaction.Amount // Use same amount if not specified
	}

	// Create source transaction (money going out)
	sourceTransaction := transaction
	sourceTransaction.IsIncome = false // Transfer out is always expense for source

	// Get current balances for both accounts
	sourceCurrentBalance, err := s.sm.AccountsService.GetAccountBalance(sourceTransaction.AccountID)
	if err != nil {
		log.Error("Error getting source account balance: ", err)
		return nil, err
	}

	targetCurrentBalance, err := s.sm.AccountsService.GetAccountBalance(*targetAccountID)
	if err != nil {
		log.Error("Error getting target account balance: ", err)
		return nil, err
	}

	// Calculate effects and new balances
	sourceEffect := s.calculateTransactionEffect(sourceTransaction.AccountID, sourceTransaction.Amount, sourceTransaction.IsIncome, sourceTransaction.IsTransfer, false)
	targetEffect := s.calculateTransactionEffect(*targetAccountID, *targetAmount, true, true, true) // Transfer in is always income for target
	
	sourceNewBalance := sourceCurrentBalance.Add(sourceEffect)
	targetNewBalance := targetCurrentBalance.Add(targetEffect)

	// Set new balance in source transaction
	sourceTransaction.NewBalance = &sourceNewBalance

	// Update source account balance
	err = s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect)
	if err != nil {
		log.Error("Error updating source account balance: ", err)
		return nil, err
	}

	// Update target account balance
	err = s.updateAccountBalanceByEffect(*targetAccountID, targetEffect)
	if err != nil {
		log.Error("Error updating target account balance: ", err)
		// Rollback source account change
		rollbackErr := s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect.Neg())
		if rollbackErr != nil {
			log.Error("Error rolling back source account balance: ", rollbackErr)
		}
		return nil, err
	}

	// Create source transaction in database
	createdSourceTx, err := s.transactionsRepository.CreateTransaction(sourceTransaction)
	if err != nil {
		log.Error("Error creating source transaction: ", err)
		// Rollback both account balance changes
		rollbackErr1 := s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect.Neg())
		rollbackErr2 := s.updateAccountBalanceByEffect(*targetAccountID, targetEffect.Neg())
		if rollbackErr1 != nil || rollbackErr2 != nil {
			log.Error("Error rolling back account balances: ", rollbackErr1, rollbackErr2)
		}
		return nil, err
	}

	// Create target transaction (money coming in)
	targetTransaction := models.Transaction{
		UserID:              transaction.UserID,
		AccountID:           *targetAccountID,
		Amount:              *targetAmount,
		CategoryID:          transaction.CategoryID, // Can use same category or make it configurable
		Label:               transaction.Label, // Use the same label as the source transaction
		IsIncome:            true, // Transfer in is always income for target
		IsTransfer:          true,
		LinkedTransactionID: createdSourceTx.ID, // Link to the created source transaction
		NewBalance:          &targetNewBalance, // Set the new balance for target account
		Notes:               transaction.Notes,
		DateTime:            transaction.DateTime,
		CreatedAt:           transaction.CreatedAt,
		UpdatedAt:           transaction.UpdatedAt,
	}

	// Create target transaction in database
	createdTargetTx, err := s.transactionsRepository.CreateTransaction(targetTransaction)
	if err != nil {
		log.Error("Error creating target transaction: ", err)
		// Try to delete the source transaction and rollback balances
		deleteErr := s.transactionsRepository.DeleteTransaction(*createdSourceTx.ID, transaction.UserID)
		rollbackErr1 := s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect.Neg())
		rollbackErr2 := s.updateAccountBalanceByEffect(*targetAccountID, targetEffect.Neg())
		if deleteErr != nil || rollbackErr1 != nil || rollbackErr2 != nil {
			log.Error("Error cleaning up failed transfer: ", deleteErr, rollbackErr1, rollbackErr2)
		}
		return nil, err
	}

	// Update source transaction with linked transaction ID
	if createdTargetTx.ID != nil {
		createdSourceTx.LinkedTransactionID = createdTargetTx.ID
		updateErr := s.transactionsRepository.UpdateTransaction(*createdSourceTx)
		if updateErr != nil {
			log.Error("Error linking transactions: ", updateErr)
			// Continue anyway, as the transactions are created
		}
	}

	return createdSourceTx, nil
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
		baseCurrencyAmount = *transactionRaw.BaseCurrencyAmount
	} else {
		amount, err := s.sm.ExchangeRatesService.CalcAmountFromCurrency(
			*transactionRaw.DateTime,
			transactionRaw.Amount,
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
		newBalance = *raw.NewBalance
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
		Amount:          raw.Amount,
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

func (s *TransactionsServiceInstance) UpdateTransaction(transactionDTO dto.PutTransactionDTO, userId int) error {
	log.Debug("UpdateTransaction Service")

	// Get the existing transaction to compare values
	existingTransaction, err := s.transactionsRepository.GetTransactionDetail(transactionDTO.ID, userId)
	if err != nil {
		log.Error("Error getting existing transaction: ", err)
		return err
	}
	if existingTransaction == nil {
		return fmt.Errorf("transaction not found")
	}

	now := time.Now()
	transaction := models.Transaction{
		ID:         &transactionDTO.ID,
		UserID:     userId,
		AccountID:  transactionDTO.AccountID,
		Amount:     transactionDTO.Amount,
		CategoryID: transactionDTO.CategoryID,
		Label:      transactionDTO.Label,
		IsIncome:   transactionDTO.IsIncome,
		IsTransfer: transactionDTO.IsTransfer,
		DateTime:   transactionDTO.DateTime,
		UpdatedAt:  &now,
	}

	transaction.Notes = transactionDTO.Notes

	// Handle account balance updates
	err = s.handleAccountBalanceUpdates(existingTransaction, &transaction)
	if err != nil {
		log.Error("Error handling account balance updates: ", err)
		return err
	}

	// Calculate and set the new balance for the updated transaction
	currentBalance, err := s.sm.AccountsService.GetAccountBalance(transaction.AccountID)
	if err != nil {
		log.Error("Error getting current balance for updated transaction: ", err)
		return err
	}
	transaction.NewBalance = &currentBalance

	// Call repository for update
	err = s.transactionsRepository.UpdateTransaction(transaction)
	if err != nil {
		log.Error("Error updating transaction: ", err)
		return err
	}

	return nil
}

func (s *TransactionsServiceInstance) DeleteTransaction(transactionId int, userId int) error {
	log.Debug("DeleteTransaction Service")

	// Get the existing transaction to handle balance updates
	existingTransaction, err := s.transactionsRepository.GetTransactionDetail(transactionId, userId)
	if err != nil {
		log.Error("Error getting existing transaction: ", err)
		return err
	}
	if existingTransaction == nil {
		return fmt.Errorf("transaction not found")
	}

	// Handle account balance updates before deletion
	err = s.handleAccountBalanceOnDelete(existingTransaction)
	if err != nil {
		log.Error("Error handling account balance on delete: ", err)
		return err
	}

	err = s.transactionsRepository.DeleteTransaction(transactionId, userId)
	if err != nil {
		log.Error("Error deleting transaction: ", err)
		return err
	}

	return nil
}

// handleAccountBalanceUpdates handles balance changes when a transaction is updated
func (s *TransactionsServiceInstance) handleAccountBalanceUpdates(oldTx *dto.TransactionDetailRaw, newTx *models.Transaction) error {
	// Calculate the balance effect changes
	oldEffect := s.calculateTransactionEffect(oldTx.AccountID, oldTx.Amount, oldTx.IsIncome, oldTx.IsTransfer, false)
	newEffect := s.calculateTransactionEffect(newTx.AccountID, newTx.Amount, newTx.IsIncome, newTx.IsTransfer, false)

	// Handle account changes
	if oldTx.AccountID != newTx.AccountID {
		// Transaction moved between accounts
		// Reverse old effect on old account
		err := s.updateAccountBalanceByEffect(oldTx.AccountID, oldEffect.Neg())
		if err != nil {
			return err
		}

		// Apply new effect on new account
		err = s.updateAccountBalanceByEffect(newTx.AccountID, newEffect)
		if err != nil {
			return err
		}

		// Handle transfer transactions affecting linked accounts
		if oldTx.IsTransfer && oldTx.LinkedTransactionID != nil {
			// Handle old transfer's linked account
			linkedTx, err := s.transactionsRepository.GetTransactionDetail(*oldTx.LinkedTransactionID, oldTx.UserID)
			if err == nil && linkedTx != nil {
				linkedEffect := s.calculateTransactionEffect(linkedTx.AccountID, linkedTx.Amount, linkedTx.IsIncome, linkedTx.IsTransfer, true)
				err = s.updateAccountBalanceByEffect(linkedTx.AccountID, linkedEffect.Neg())
				if err != nil {
					return err
				}
			}
		}

		if newTx.IsTransfer && newTx.LinkedTransactionID != nil {
			// Handle new transfer's linked account
			linkedTx, err := s.transactionsRepository.GetTransactionDetail(*newTx.LinkedTransactionID, newTx.UserID)
			if err == nil && linkedTx != nil {
				linkedEffect := s.calculateTransactionEffect(linkedTx.AccountID, newTx.Amount, !newTx.IsIncome, newTx.IsTransfer, true)
				err = s.updateAccountBalanceByEffect(linkedTx.AccountID, linkedEffect)
				if err != nil {
					return err
				}
			}
		}
	} else {
		// Same account, just update the balance difference
		balanceDifference := newEffect.Sub(oldEffect)
		if !balanceDifference.IsZero() {
			err := s.updateAccountBalanceByEffect(newTx.AccountID, balanceDifference)
			if err != nil {
				return err
			}
		}

		// Handle transfer amount changes on linked account
		if oldTx.IsTransfer && newTx.IsTransfer && oldTx.LinkedTransactionID != nil {
			linkedTx, err := s.transactionsRepository.GetTransactionDetail(*oldTx.LinkedTransactionID, oldTx.UserID)
			if err == nil && linkedTx != nil {
				oldLinkedEffect := s.calculateTransactionEffect(linkedTx.AccountID, oldTx.Amount, !oldTx.IsIncome, linkedTx.IsTransfer, true)
				newLinkedEffect := s.calculateTransactionEffect(linkedTx.AccountID, newTx.Amount, !newTx.IsIncome, linkedTx.IsTransfer, true)
				linkedDifference := newLinkedEffect.Sub(oldLinkedEffect)
				
				if !linkedDifference.IsZero() {
					err = s.updateAccountBalanceByEffect(linkedTx.AccountID, linkedDifference)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// handleAccountBalanceOnDelete handles balance changes when a transaction is deleted
func (s *TransactionsServiceInstance) handleAccountBalanceOnDelete(tx *dto.TransactionDetailRaw) error {
	// Reverse the transaction effect
	effect := s.calculateTransactionEffect(tx.AccountID, tx.Amount, tx.IsIncome, tx.IsTransfer, false)
	err := s.updateAccountBalanceByEffect(tx.AccountID, effect.Neg())
	if err != nil {
		return err
	}

	// Handle linked transaction for transfers
	if tx.IsTransfer && tx.LinkedTransactionID != nil {
		linkedTx, err := s.transactionsRepository.GetTransactionDetail(*tx.LinkedTransactionID, tx.UserID)
		if err == nil && linkedTx != nil {
			linkedEffect := s.calculateTransactionEffect(linkedTx.AccountID, linkedTx.Amount, linkedTx.IsIncome, linkedTx.IsTransfer, true)
			err = s.updateAccountBalanceByEffect(linkedTx.AccountID, linkedEffect.Neg())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// calculateTransactionEffect calculates how a transaction affects account balance
func (s *TransactionsServiceInstance) calculateTransactionEffect(accountID int, amount decimal.Decimal, isIncome, isTransfer, isLinkedTransaction bool) decimal.Decimal {
	if isTransfer {
		if isLinkedTransaction {
			// For linked transactions in transfers, it's always the opposite effect
			return amount
		} else {
			// For the originating transaction in transfers, it's always negative (money going out)
			return amount.Neg()
		}
	}
	
	if isIncome {
		return amount // Positive effect
	} else {
		return amount.Neg() // Negative effect
	}
}

// updateAccountBalanceByEffect updates an account's balance by applying the effect
func (s *TransactionsServiceInstance) updateAccountBalanceByEffect(accountID int, effect decimal.Decimal) error {
	if effect.IsZero() {
		return nil // No change needed
	}

	currentBalance, err := s.sm.AccountsService.GetAccountBalance(accountID)
	if err != nil {
		return err
	}

	newBalance := currentBalance.Add(effect)
	return s.sm.AccountsService.UpdateAccountBalance(accountID, newBalance)
}
