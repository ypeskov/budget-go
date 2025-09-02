package services

import (
	"fmt"
	"sync"
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/transactions"

	"github.com/shopspring/decimal"
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
	GetExpenseTransactionsForBudget(userId int, categoryIds []int, startDate time.Time, endDate time.Time, transactionIds []int) ([]models.Transaction, error)
}

type TransactionsServiceInstance struct {
	transactionsRepository transactions.Repository
	sm                     *Manager
}

var (
	transactionsInstance *TransactionsServiceInstance
	transactionsOnce     sync.Once
)

func NewTransactionsService(transactionsRepository transactions.Repository, sManager *Manager) TransactionsService {
	transactionsOnce.Do(func() {
		logger.Debug("Creating TransactionsService instance")
		transactionsInstance = &TransactionsServiceInstance{
			transactionsRepository: transactionsRepository,
			sm:                     sManager,
		}
	})

	return transactionsInstance
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
	logger.Debug("GetTransactionsWithAccounts Service")

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
		logger.Error("Error getting transactions", "error", err)
		return nil, err
	}

	var baseCurrency models.Currency
	baseCurrency, err = s.sm.UserSettingsService.GetBaseCurrency(userId)
	if err != nil {
		logger.Error("Error getting base currency", "error", err)
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
			logger.Error("Error calculating amount", "error", err)
			return nil, err
		}

		transactions[i].BaseCurrencyAmount = &amount
	}

	return transactions, nil
}

func (s *TransactionsServiceInstance) GetTemplates(userId int) ([]dto.TemplateDTO, error) {
	logger.Debug("GetTemplates Service")

	templates, err := s.transactionsRepository.GetTemplates(userId)
	if err != nil {
		logger.Error("Error getting templates", "error", err)
		return nil, err
	}

	return templates, nil
}

func (s *TransactionsServiceInstance) DeleteTemplates(templateIds []int, userId int) error {
	logger.Debug("DeleteTemplates Service")

	err := s.transactionsRepository.DeleteTemplates(templateIds, userId)
	if err != nil {
		logger.Error("Error deleting templates", "error", err)
		return err
	}

	return nil
}

func (s *TransactionsServiceInstance) CreateTransaction(transaction models.Transaction, targetAccountID *int, targetAmount *decimal.Decimal) (*models.Transaction, error) {
	logger.Debug("CreateTransaction Service")

	// Validate category ownership
	if transaction.CategoryID != nil && *transaction.CategoryID > 0 {
		isOwner, err := s.sm.CategoriesService.ValidateCategoryOwnership(*transaction.CategoryID, transaction.UserID)
		if err != nil {
			logger.Error("Error validating category ownership", "error", err)
			return nil, err
		}
		if !isOwner {
			logger.Error("User does not own category ID", "categoryID", *transaction.CategoryID)
			return nil, fmt.Errorf("category not found or does not belong to user")
		}
	}

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
	logger.Debug("Creating regular transaction")

	// Get current account balance
	currentBalance, err := s.sm.AccountsService.GetAccountBalance(transaction.AccountID)
	if err != nil {
		logger.Error("Error getting current account balance", "error", err)
		return nil, err
	}

	// Calculate transaction effect and new balance
	effect := s.calculateTransactionEffect(transaction.Amount, transaction.IsIncome, transaction.IsTransfer, false)
	newBalance := currentBalance.Add(effect)

	// Set the new balance in the transaction record
	transaction.NewBalance = &newBalance

	// Update account balance
	err = s.updateAccountBalanceByEffect(transaction.AccountID, effect)
	if err != nil {
		logger.Error("Error updating account balance for new transaction", "error", err)
		return nil, err
	}

	createdTransaction, err := s.transactionsRepository.CreateTransaction(transaction)
	if err != nil {
		logger.Error("Error creating transaction", "error", err)
		// Rollback balance change if transaction creation failed
		rollbackErr := s.updateAccountBalanceByEffect(transaction.AccountID, effect.Neg())
		if rollbackErr != nil {
			logger.Error("Error rolling back account balance", "error", rollbackErr)
		}
		return nil, err
	}

	// Update only affected budgets (recompute), if this is an expense transaction
	if !transaction.IsIncome && !transaction.IsTransfer {
		pairs := []AffectedCategoryDate{}
		if transaction.CategoryID != nil && transaction.DateTime != nil {
			pairs = append(pairs, AffectedCategoryDate{CategoryID: *transaction.CategoryID, Date: *transaction.DateTime})
		}
		if len(pairs) > 0 {
			if err := s.sm.BudgetsService.UpdateBudgetCollectedAmountsForCategories(transaction.UserID, pairs); err != nil {
				logger.Error("Error updating affected budgets after transaction creation", "error", err)
			}
		}
	}

	return createdTransaction, nil
}

func (s *TransactionsServiceInstance) createTransferTransaction(transaction models.Transaction,
	targetAccountID *int,
	targetAmount *decimal.Decimal) (*models.Transaction, error) {
	logger.Debug("Creating transfer transaction")

	if targetAccountID == nil {
		return nil, fmt.Errorf("target account ID is required for transfer transactions")
	}

	if targetAmount == nil {
		return nil, fmt.Errorf("target amount is required for transfer transactions")
	}

	// Create source transaction (money going out)
	sourceTransaction := transaction
	sourceTransaction.IsIncome = false // Transfer out is always expense for source

	// Get current balances for both accounts
	sourceCurrentBalance, err := s.sm.AccountsService.GetAccountBalance(sourceTransaction.AccountID)
	if err != nil {
		logger.Error("Error getting source account balance", "error", err)
		return nil, err
	}

	targetCurrentBalance, err := s.sm.AccountsService.GetAccountBalance(*targetAccountID)
	if err != nil {
		logger.Error("Error getting target account balance", "error", err)
		return nil, err
	}

	// Calculate effects and new balances
	sourceEffect := s.calculateTransactionEffect(sourceTransaction.Amount, sourceTransaction.IsIncome, sourceTransaction.IsTransfer, false)
	targetEffect := s.calculateTransactionEffect(*targetAmount, true, true, true) // Transfer in is always income for target

	sourceNewBalance := sourceCurrentBalance.Add(sourceEffect)
	targetNewBalance := targetCurrentBalance.Add(targetEffect)

	// Set new balance in source transaction
	sourceTransaction.NewBalance = &sourceNewBalance

	// Update source account balance
	err = s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect)
	if err != nil {
		logger.Error("Error updating source account balance", "error", err)
		return nil, err
	}

	// Update target account balance
	err = s.updateAccountBalanceByEffect(*targetAccountID, targetEffect)
	if err != nil {
		logger.Error("Error updating target account balance", "error", err)
		// Rollback source account change
		rollbackErr := s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect.Neg())
		if rollbackErr != nil {
			logger.Error("Error rolling back source account balance", "error", rollbackErr)
		}
		return nil, err
	}

	// Create source transaction in database
	createdSourceTx, err := s.transactionsRepository.CreateTransaction(sourceTransaction)
	if err != nil {
		logger.Error("Error creating source transaction", "error", err)
		// Rollback both account balance changes
		rollbackErr1 := s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect.Neg())
		rollbackErr2 := s.updateAccountBalanceByEffect(*targetAccountID, targetEffect.Neg())
		if rollbackErr1 != nil || rollbackErr2 != nil {
			logger.Error("Error rolling back account balances", "error1", rollbackErr1, "error2", rollbackErr2)
		}
		return nil, err
	}

	// Create target transaction (money coming in)
	targetTransaction := models.Transaction{
		UserID:              transaction.UserID,
		AccountID:           *targetAccountID,
		Amount:              *targetAmount,
		CategoryID:          transaction.CategoryID, // Can use same category or make it configurable
		Label:               transaction.Label,      // Use the same label as the source transaction
		IsIncome:            true,                   // Transfer in is always income for target
		IsTransfer:          true,
		LinkedTransactionID: createdSourceTx.ID, // Link to the created source transaction
		NewBalance:          &targetNewBalance,  // Set the new balance for target account
		Notes:               transaction.Notes,
		DateTime:            transaction.DateTime,
		CreatedAt:           transaction.CreatedAt,
		UpdatedAt:           transaction.UpdatedAt,
	}

	// Create target transaction in database
	createdTargetTx, err := s.transactionsRepository.CreateTransaction(targetTransaction)
	if err != nil {
		logger.Error("Error creating target transaction", "error", err)
		// Try to delete the source transaction and rollback balances
		deleteErr := s.transactionsRepository.DeleteTransaction(*createdSourceTx.ID, transaction.UserID)
		rollbackErr1 := s.updateAccountBalanceByEffect(sourceTransaction.AccountID, sourceEffect.Neg())
		rollbackErr2 := s.updateAccountBalanceByEffect(*targetAccountID, targetEffect.Neg())
		if deleteErr != nil || rollbackErr1 != nil || rollbackErr2 != nil {
			logger.Error("Error cleaning up failed transfer", "deleteError", deleteErr, "rollbackError1", rollbackErr1, "rollbackError2", rollbackErr2)
		}
		return nil, err
	}

	// Update source transaction with linked transaction ID
	if createdTargetTx.ID != nil {
		createdSourceTx.LinkedTransactionID = createdTargetTx.ID
		updateErr := s.transactionsRepository.UpdateTransaction(*createdSourceTx)
		if updateErr != nil {
			logger.Error("Error linking transactions", "error", updateErr)
			// Continue anyway, as the transactions are created
		}
	}

	return createdSourceTx, nil
}

func (s *TransactionsServiceInstance) GetTransactionDetail(transactionId int, userId int) (*dto.TransactionDetailDTO, error) {
	logger.Debug("GetTransactionDetail Service")

	transactionRaw, err := s.transactionsRepository.GetTransactionDetail(transactionId, userId)
	if err != nil {
		logger.Error("Error getting transaction detail", "error", err)
		return nil, err
	}

	if transactionRaw == nil {
		return nil, nil
	}

	baseCurrency, err := s.sm.UserSettingsService.GetBaseCurrency(userId)
	if err != nil {
		logger.Error("Error getting base currency", "error", err)
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
			logger.Error("Error calculating base currency amount", "error", err)
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
	if raw.Category != nil {
		categoryDetail = dto.CategoryDetailDTO{
			Name:      getStringValue(raw.Category.Name),
			ParentID:  raw.Category.ParentID,
			IsIncome:  getBoolValue(raw.Category.IsIncome),
			ID:        getIntValue(raw.Category.ID),
			UserID:    getIntValue(raw.Category.UserID),
			CreatedAt: getStringValue(raw.Category.CreatedAt),
			UpdatedAt: getStringValue(raw.Category.UpdatedAt),
			Children:  []dto.CategoryDetailDTO{},
		}
	}

	// Handle linked transaction
	var linkedTransaction *dto.TransactionDetailDTO
	if raw.LinkedTransaction != nil && raw.LinkedTransaction.ID != nil {
		linkedTransaction = convertLinkedTransactionToDetail(raw.LinkedTransaction, baseCurrencyCode)
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
		User: dto.UserRegisterResponseDTO{
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
		LinkedTransaction:   linkedTransaction,
	}
}

func convertLinkedTransactionToDetail(linkedTx *models.NullableTransaction, baseCurrencyCode string) *dto.TransactionDetailDTO {
	var notes string
	if linkedTx.Notes != nil {
		notes = *linkedTx.Notes
	}

	var newBalance decimal.Decimal
	if linkedTx.NewBalance != nil {
		newBalance = *linkedTx.NewBalance
	} else {
		newBalance = decimal.Zero
	}

	var baseCurrencyAmount decimal.Decimal
	if linkedTx.BaseCurrencyAmount != nil {
		baseCurrencyAmount = *linkedTx.BaseCurrencyAmount
	} else {
		baseCurrencyAmount = decimal.Zero
	}

	var amount decimal.Decimal
	if linkedTx.Amount != nil {
		amount = *linkedTx.Amount
	} else {
		amount = decimal.Zero
	}

	var label string
	if linkedTx.Label != nil {
		label = *linkedTx.Label
	}

	return &dto.TransactionDetailDTO{
		ID:                  getIntValue(linkedTx.ID),
		AccountID:           getIntValue(linkedTx.AccountID),
		CategoryID:          linkedTx.CategoryID,
		Amount:              amount,
		Label:               label,
		Notes:               notes,
		DateTime:            linkedTx.DateTime,
		IsTransfer:          getBoolValue(linkedTx.IsTransfer),
		IsIncome:            getBoolValue(linkedTx.IsIncome),
		UserID:              getIntValue(linkedTx.UserID),
		BaseCurrencyAmount:  baseCurrencyAmount,
		BaseCurrencyCode:    baseCurrencyCode,
		NewBalance:          newBalance,
		LinkedTransactionID: linkedTx.LinkedTransactionID,
		// Note: We don't include nested linked transactions to avoid infinite recursion
		LinkedTransaction: nil,
	}
}

func (s *TransactionsServiceInstance) UpdateTransaction(transactionDTO dto.PutTransactionDTO, userId int) error {
	logger.Debug("UpdateTransaction Service")

	// Validate category ownership
	if transactionDTO.CategoryID != nil && *transactionDTO.CategoryID > 0 {
		isOwner, err := s.sm.CategoriesService.ValidateCategoryOwnership(*transactionDTO.CategoryID, userId)
		if err != nil {
			logger.Error("Error validating category ownership", "error", err)
			return err
		}
		if !isOwner {
			logger.Error("User does not own category ID", "categoryID", *transactionDTO.CategoryID)
			return fmt.Errorf("category not found or does not belong to user")
		}
	}

	// Get the existing transaction to compare values
	existingTransaction, err := s.transactionsRepository.GetTransactionDetail(transactionDTO.ID, userId)
	if err != nil {
		logger.Error("Error getting existing transaction", "error", err)
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

	// Preserve linked transaction ID if it exists
	if existingTransaction.LinkedTransactionID != nil {
		transaction.LinkedTransactionID = existingTransaction.LinkedTransactionID
	}

	transaction.Notes = transactionDTO.Notes

	// Handle account balance updates (including target account changes for transfers)
	err = s.handleAccountBalanceUpdates(existingTransaction, &transaction, transactionDTO.TargetAccountID)
	if err != nil {
		logger.Error("Error handling account balance updates", "error", err)
		return err
	}

	// Calculate and set the new balance for the updated transaction
	currentBalance, err := s.sm.AccountsService.GetAccountBalance(transaction.AccountID)
	if err != nil {
		logger.Error("Error getting current balance for updated transaction", "error", err)
		return err
	}
	transaction.NewBalance = &currentBalance

	// Call repository for update
	err = s.transactionsRepository.UpdateTransaction(transaction)
	if err != nil {
		logger.Error("Error updating transaction", "error", err)
		return err
	}

	// Handle transfer transactions - update the linked transaction
	if existingTransaction.IsTransfer && transaction.IsTransfer && existingTransaction.LinkedTransactionID != nil {
		err = s.updateLinkedTransferTransaction(existingTransaction, &transaction, transactionDTO.TargetAmount, transactionDTO.TargetAccountID)
		if err != nil {
			logger.Error("Error updating linked transfer transaction", "error", err)
			return err
		}
	}

	// Update only affected budgets (recompute) for expense impact
	// Consider both old and new values if either side is expense and not transfer
	var pairs []AffectedCategoryDate
	if !existingTransaction.IsIncome && !existingTransaction.IsTransfer && existingTransaction.CategoryID != nil && existingTransaction.DateTime != nil {
		pairs = append(pairs, AffectedCategoryDate{CategoryID: *existingTransaction.CategoryID, Date: *existingTransaction.DateTime})
	}
	if !transaction.IsIncome && !transaction.IsTransfer && transaction.CategoryID != nil && transaction.DateTime != nil {
		pairs = append(pairs, AffectedCategoryDate{CategoryID: *transaction.CategoryID, Date: *transaction.DateTime})
	}
	if len(pairs) > 0 {
		if err := s.sm.BudgetsService.UpdateBudgetCollectedAmountsForCategories(userId, pairs); err != nil {
			logger.Error("Error updating affected budgets after transaction update", "error", err)
		}
	}

	return nil
}

func (s *TransactionsServiceInstance) DeleteTransaction(transactionId int, userId int) error {
	logger.Debug("DeleteTransaction Service")

	// Get the existing transaction to handle balance updates
	existingTransaction, err := s.transactionsRepository.GetTransactionDetail(transactionId, userId)
	if err != nil {
		logger.Error("Error getting existing transaction", "error", err)
		return err
	}
	if existingTransaction == nil {
		return fmt.Errorf("transaction not found")
	}

	// Handle account balance updates before deletion
	err = s.handleAccountBalanceOnDelete(existingTransaction)
	if err != nil {
		logger.Error("Error handling account balance on delete", "error", err)
		return err
	}

	err = s.transactionsRepository.DeleteTransaction(transactionId, userId)
	if err != nil {
		logger.Error("Error deleting transaction", "error", err)
		return err
	}

	// Update only affected budgets (recompute) if this was an expense transaction
	if !existingTransaction.IsIncome && !existingTransaction.IsTransfer && existingTransaction.CategoryID != nil && existingTransaction.DateTime != nil {
		pairs := []AffectedCategoryDate{{CategoryID: *existingTransaction.CategoryID, Date: *existingTransaction.DateTime}}
		if err := s.sm.BudgetsService.UpdateBudgetCollectedAmountsForCategories(userId, pairs); err != nil {
			logger.Error("Error updating affected budgets after transaction deletion", "error", err)
		}
	}

	return nil
}

// handleAccountBalanceUpdates handles balance changes when a transaction is updated
func (s *TransactionsServiceInstance) handleAccountBalanceUpdates(oldTx *dto.TransactionDetailRaw, newTx *models.Transaction, newTargetAccountID *int) error {
	// Calculate the balance effect changes
	oldEffect := s.calculateTransactionEffect(oldTx.Amount, oldTx.IsIncome, oldTx.IsTransfer, false)
	newEffect := s.calculateTransactionEffect(newTx.Amount, newTx.IsIncome, newTx.IsTransfer, false)

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
				linkedEffect := s.calculateTransactionEffect(linkedTx.Amount, linkedTx.IsIncome, linkedTx.IsTransfer, true)
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
				linkedEffect := s.calculateTransactionEffect(newTx.Amount, !newTx.IsIncome, newTx.IsTransfer, true)
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
				// Handle target account change for transfers
				if newTargetAccountID != nil && linkedTx.AccountID != *newTargetAccountID {
					// Target account changed - move balance from old to new target account
					oldLinkedEffect := s.calculateTransactionEffect(linkedTx.Amount, linkedTx.IsIncome, linkedTx.IsTransfer, true)
					newLinkedEffect := s.calculateTransactionEffect(newTx.Amount, !newTx.IsIncome, newTx.IsTransfer, true)
					
					// Reverse effect from old target account
					err = s.updateAccountBalanceByEffect(linkedTx.AccountID, oldLinkedEffect.Neg())
					if err != nil {
						return err
					}
					
					// Apply effect to new target account
					err = s.updateAccountBalanceByEffect(*newTargetAccountID, newLinkedEffect)
					if err != nil {
						return err
					}
				} else {
					// Same target account, just handle amount changes
					oldLinkedEffect := s.calculateTransactionEffect(oldTx.Amount, !oldTx.IsIncome, linkedTx.IsTransfer, true)
					newLinkedEffect := s.calculateTransactionEffect(newTx.Amount, !newTx.IsIncome, linkedTx.IsTransfer, true)
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
	}

	return nil
}

// handleAccountBalanceOnDelete handles balance changes when a transaction is deleted
func (s *TransactionsServiceInstance) handleAccountBalanceOnDelete(tx *dto.TransactionDetailRaw) error {
	// Reverse the transaction effect
	effect := s.calculateTransactionEffect(tx.Amount, tx.IsIncome, tx.IsTransfer, false)
	err := s.updateAccountBalanceByEffect(tx.AccountID, effect.Neg())
	if err != nil {
		return err
	}

	// Handle linked transaction for transfers
	if tx.IsTransfer && tx.LinkedTransactionID != nil {
		linkedTx, err := s.transactionsRepository.GetTransactionDetail(*tx.LinkedTransactionID, tx.UserID)
		if err == nil && linkedTx != nil {
			linkedEffect := s.calculateTransactionEffect(linkedTx.Amount, linkedTx.IsIncome, linkedTx.IsTransfer, true)
			err = s.updateAccountBalanceByEffect(linkedTx.AccountID, linkedEffect.Neg())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// calculateTransactionEffect calculates how a transaction affects account balance
func (s *TransactionsServiceInstance) calculateTransactionEffect(amount decimal.Decimal, isIncome, isTransfer, isLinkedTransaction bool) decimal.Decimal {
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

// updateLinkedTransferTransaction updates the linked transaction for a transfer
func (s *TransactionsServiceInstance) updateLinkedTransferTransaction(existingTx *dto.TransactionDetailRaw,
	updatedSourceTx *models.Transaction,
	targetAmount *decimal.Decimal,
	newTargetAccountID *int) error {
	// Get the linked transaction details
	linkedTx, err := s.transactionsRepository.GetTransactionDetail(*existingTx.LinkedTransactionID, existingTx.UserID)
	if err != nil {
		return fmt.Errorf("error getting linked transaction: %w", err)
	}
	if linkedTx == nil {
		return fmt.Errorf("linked transaction not found")
	}

	// Determine the amount for the linked transaction
	linkedAmount := updatedSourceTx.Amount // Default to same amount as source
	if targetAmount != nil {
		linkedAmount = *targetAmount // Use target amount if specified
	}

	// sync dates of both transactions
	if updatedSourceTx.DateTime != nil {
		linkedTx.DateTime = updatedSourceTx.DateTime
	}

	// sync labels of both transactions
	linkedTx.Label = updatedSourceTx.Label

	// sync notes of both transactions
	linkedTx.Notes = updatedSourceTx.Notes

	// Determine which account to use for the linked transaction
	targetAccountID := linkedTx.AccountID
	if newTargetAccountID != nil {
		targetAccountID = *newTargetAccountID
	}

	// Get current balance for the target account (already updated by handleAccountBalanceUpdates)
	linkedCurrentBalance, err := s.sm.AccountsService.GetAccountBalance(targetAccountID)
	if err != nil {
		return fmt.Errorf("error getting linked account balance: %w", err)
	}

	// Create updated linked transaction
	now := time.Now()
	updatedLinkedTx := models.Transaction{
		ID:                  linkedTx.ID,
		UserID:              linkedTx.UserID,
		AccountID:           targetAccountID,          // Use new target account if changed
		Amount:              linkedAmount,             // Use the determined amount
		CategoryID:          linkedTx.CategoryID,      // Keep existing category
		Label:               updatedSourceTx.Label,    // Sync label with source
		IsIncome:            linkedTx.IsIncome,        // Keep original direction
		IsTransfer:          linkedTx.IsTransfer,
		LinkedTransactionID: updatedSourceTx.ID,       // Link back to source
		Notes:               updatedSourceTx.Notes,    // Sync notes with source
		DateTime:            updatedSourceTx.DateTime, // Sync datetime with source
		NewBalance:          &linkedCurrentBalance,    // Current balance after updates
		UpdatedAt:           &now,
	}

	// Update the linked transaction
	err = s.transactionsRepository.UpdateTransaction(updatedLinkedTx)
	if err != nil {
		return fmt.Errorf("error updating linked transaction: %w", err)
	}

	return nil
}

func (s *TransactionsServiceInstance) GetExpenseTransactionsForBudget(userId int, categoryIds []int, startDate time.Time, endDate time.Time, transactionIds []int) ([]models.Transaction, error) {
	txList, err := s.transactionsRepository.GetExpenseTransactionsForBudget(userId, categoryIds, startDate, endDate, transactionIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense transactions for user %d: %w", userId, err)
	}

	return txList, nil
}
