package services

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"
	budgetRepo "ypeskov/budget-go/internal/repositories/budgets"

	"github.com/shopspring/decimal"
)

type BudgetsService interface {
	CreateBudget(budgetDTO dto.CreateBudgetDTO, userID int) (*models.Budget, error)
	UpdateBudget(budgetDTO dto.UpdateBudgetDTO, userID int) (*models.Budget, error)
	GetUserBudgets(userID int, include string) ([]dto.BudgetResponseDTO, error)
	DeleteBudget(budgetID int, userID int) error
	ArchiveBudget(budgetID int, userID int) error
	ProcessOutdatedBudgets() ([]int, error)
	UpdateBudgetCollectedAmounts(userID int) error
	// UpdateBudgetCollectedAmountsForCategories recalculates only budgets affected by given category/date pairs
	UpdateBudgetCollectedAmountsForCategories(userID int, pairs []AffectedCategoryDate) error
}

type BudgetsServiceInstance struct {
	budgetsRepository budgetRepo.Repository
	sm                *Manager
}

// AffectedCategoryDate describes a category and the date to match budgets' period window
type AffectedCategoryDate struct {
	CategoryID int
	Date       time.Time
}

var (
	budgetsInstance *BudgetsServiceInstance
	budgetsOnce     sync.Once
)

func NewBudgetsService(budgetsRepository budgetRepo.Repository, sManager *Manager) BudgetsService {
	budgetsOnce.Do(func() {
		logger.Debug("Creating BudgetsService instance")
		budgetsInstance = &BudgetsServiceInstance{
			budgetsRepository: budgetsRepository,
			sm:                sManager,
		}
	})

	return budgetsInstance
}

func (s *BudgetsServiceInstance) CreateBudget(budgetDTO dto.CreateBudgetDTO, userID int) (*models.Budget, error) {
	logger.Debug("CreateBudget Service")

	// Validate period
	if !models.ValidatePeriod(budgetDTO.Period) {
		return nil, fmt.Errorf("invalid period: %s. Valid periods are: %v", budgetDTO.Period, models.GetValidPeriods())
	}

	// Validate and filter categories
	validCategories, err := s.budgetsRepository.GetUserCategoriesForBudget(userID, budgetDTO.Categories)
	if err != nil {
		logger.Error("Error validating categories", "error", err)
		return nil, err
	}

	categoriesStr := ConvertCategoryIDsToString(validCategories)

	now := time.Now()
	startDate := budgetDTO.StartDate.ToTime()
	endDate := *budgetDTO.EndDate.ToTime()
	// Add 1 day to include the full end date (similar to FastAPI logic)
	endDate = endDate.AddDate(0, 0, 1)

	budget := models.Budget{
		UserID:             userID,
		Name:               budgetDTO.Name,
		CurrencyID:         budgetDTO.CurrencyID,
		TargetAmount:       budgetDTO.TargetAmount,
		CollectedAmount:    decimal.Zero,
		Period:             models.NormalizePeriod(budgetDTO.Period),
		Repeat:             budgetDTO.Repeat,
		StartDate:          startDate,
		EndDate:            &endDate,
		IncludedCategories: &categoriesStr,
		Comment:            budgetDTO.Comment,
		IsDeleted:          false,
		IsArchived:         false,
		CreatedAt:          &now,
		UpdatedAt:          &now,
	}

	createdBudget, err := s.budgetsRepository.CreateBudget(budget)
	if err != nil {
		logger.Error("Error creating budget", "error", err)
		return nil, err
	}

	// Fill budget with existing transactions
	err = s.fillBudgetWithExistingTransactions(*createdBudget.ID, userID)
	if err != nil {
		logger.Error("Error filling budget with existing transactions", "error", err)
		// Don't fail the creation, just log the error
	}

	return createdBudget, nil
}

func (s *BudgetsServiceInstance) UpdateBudget(budgetDTO dto.UpdateBudgetDTO, userID int) (*models.Budget, error) {
	logger.Debug("UpdateBudget Service")

	// Validate period
	if !models.ValidatePeriod(budgetDTO.Period) {
		return nil, fmt.Errorf("invalid period: %s. Valid periods are: %v", budgetDTO.Period, models.GetValidPeriods())
	}

	// Get existing budget to verify ownership
	existingBudget, err := s.budgetsRepository.GetBudgetByID(budgetDTO.ID, userID)
	if err != nil {
		logger.Error("Error getting existing budget", "error", err)
		return nil, fmt.Errorf("budget not found")
	}

	// Validate and filter categories
	validCategories, err := s.budgetsRepository.GetUserCategoriesForBudget(userID, budgetDTO.Categories)
	if err != nil {
		logger.Error("Error validating categories", "error", err)
		return nil, err
	}

	categoriesStr := ConvertCategoryIDsToString(validCategories)

	now := time.Now()
	startDate := budgetDTO.StartDate.ToTime()
	endDate := *budgetDTO.EndDate.ToTime()
	// Add 1 day to include the full end date
	endDate = endDate.AddDate(0, 0, 1)

	budget := models.Budget{
		ID:                 existingBudget.ID,
		UserID:             userID,
		Name:               budgetDTO.Name,
		CurrencyID:         budgetDTO.CurrencyID,
		TargetAmount:       budgetDTO.TargetAmount,
		CollectedAmount:    decimal.Zero, // Reset collected amount on update
		Period:             models.NormalizePeriod(budgetDTO.Period),
		Repeat:             budgetDTO.Repeat,
		StartDate:          startDate,
		EndDate:            &endDate,
		IncludedCategories: &categoriesStr,
		Comment:            budgetDTO.Comment,
		IsDeleted:          existingBudget.IsDeleted,
		IsArchived:         existingBudget.IsArchived,
		CreatedAt:          existingBudget.CreatedAt,
		UpdatedAt:          &now,
	}

	err = s.budgetsRepository.UpdateBudget(budget)
	if err != nil {
		logger.Error("Error updating budget", "error", err)
		return nil, err
	}

	// Fill budget with existing transactions
	err = s.fillBudgetWithExistingTransactions(*budget.ID, userID)
	if err != nil {
		logger.Error("Error filling budget with existing transactions", "error", err)
	}

	return &budget, nil
}

func (s *BudgetsServiceInstance) GetUserBudgets(userID int, include string) ([]dto.BudgetResponseDTO, error) {
	logger.Debug("GetUserBudgets Service")

	budgetsWithCurrency, err := s.budgetsRepository.GetBudgetsWithCurrency(userID, include)
	if err != nil {
		logger.Error("Error getting user budgets", "error", err)
		return nil, err
	}

	budgetDTOs := make([]dto.BudgetResponseDTO, len(budgetsWithCurrency))
	for i, budget := range budgetsWithCurrency {
		// Subtract 1 day from end date for display (reverse of the add operation)
		endDate := *budget.EndDate
		endDate = endDate.AddDate(0, 0, -1)

		categoriesStr := ""
		if budget.IncludedCategories != nil {
			categoriesStr = *budget.IncludedCategories
		}

		budgetDTOs[i] = dto.BudgetResponseDTO{
			ID:                 *budget.ID,
			Name:               budget.Name,
			CurrencyID:         budget.CurrencyID,
			TargetAmount:       budget.TargetAmount,
			CollectedAmount:    budget.CollectedAmount,
			Period:             models.FormatPeriodForAPI(budget.Period),
			Repeat:             budget.Repeat,
			StartDate:          budget.StartDate,
			EndDate:            &endDate,
			IncludedCategories: categoriesStr,
			Comment:            budget.Comment,
			IsArchived:         budget.IsArchived,
			Currency:           budget.Currency,
		}
	}

	return budgetDTOs, nil
}

func (s *BudgetsServiceInstance) DeleteBudget(budgetID int, userID int) error {
	logger.Debug("DeleteBudget Service")

	err := s.budgetsRepository.DeleteBudget(budgetID, userID)
	if err != nil {
		logger.Error("Error deleting budget", "error", err)
		return err
	}

	return nil
}

func (s *BudgetsServiceInstance) ArchiveBudget(budgetID int, userID int) error {
	logger.Debug("ArchiveBudget Service")

	err := s.budgetsRepository.ArchiveBudget(budgetID, userID)
	if err != nil {
		logger.Error("Error archiving budget", "error", err)
		return err
	}

	return nil
}

func (s *BudgetsServiceInstance) ProcessOutdatedBudgets() ([]int, error) {
	logger.Debug("ProcessOutdatedBudgets Service")

	outdatedBudgets, err := s.budgetsRepository.GetOutdatedBudgets()
	if err != nil {
		logger.Error("Error getting outdated budgets", "error", err)
		return nil, err
	}

	archivedBudgetIDs := make([]int, 0)

	for _, budget := range outdatedBudgets {
		if budget.Repeat {
			// Create a copy for the next period
			err = s.createCopyOfOutdatedBudget(budget)
			if err != nil { // handle error but continue processing
				logger.Error("Error creating copy of outdated budget", "error", err)
			}
		}

		// Archive the original budget
		err = s.budgetsRepository.ArchiveBudget(*budget.ID, budget.UserID)
		if err != nil {
			logger.Error("Error archiving outdated budget", "error", err)
			continue
		}

		archivedBudgetIDs = append(archivedBudgetIDs, *budget.ID)
	}

	return archivedBudgetIDs, nil
}

func (s *BudgetsServiceInstance) UpdateBudgetCollectedAmounts(userID int) error {
	userBudgets, err := s.budgetsRepository.GetUserBudgets(userID, "all")
	if err != nil {
		return fmt.Errorf("failed to get budgets for user %d: %w", userID, err)
	}

	// Create a shared cache for transaction queries to avoid duplicate database calls
	// Key: "categories:startDate:endDate", Value: transactions
	transactionCache := make(map[string][]models.Transaction)
	cacheMutex := sync.RWMutex{}

	// Process budgets concurrently
	var wg sync.WaitGroup
	errorChan := make(chan error, len(userBudgets))

	// Limit concurrent goroutines to avoid overwhelming the database
	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	for _, budget := range userBudgets {
		if budget.ID != nil {
			wg.Add(1)
			go func(budgetID int, budgetName string) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if err := s.fillBudgetWithExistingTransactionsOptimized(budgetID, userID, transactionCache, &cacheMutex); err != nil {
					logger.Error("Failed to update budget", "budgetID", budgetID, "budgetName", budgetName, "userID", userID, "error", err)
					errorChan <- fmt.Errorf("budget %d (%s): %w", budgetID, budgetName, err)
				}
			}(*budget.ID, budget.Name)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)

	// Collect any errors (non-blocking since channel is closed)
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	// Return first error if any occurred
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// UpdateBudgetCollectedAmountsForCategories finds budgets active at given dates that include the categories,
// de-duplicates budgets, and recomputes their collected amounts (full scan of matching transactions).
func (s *BudgetsServiceInstance) UpdateBudgetCollectedAmountsForCategories(userID int, pairs []AffectedCategoryDate) error {
	if len(pairs) == 0 {
		return nil
	}

	// Collect affected budget IDs
	budgetIDSet := make(map[int]struct{})
	for _, p := range pairs {
		if p.CategoryID == 0 || p.Date.IsZero() {
			continue
		}
		budgets, err := s.budgetsRepository.GetActiveBudgetsByCategoryAndDate(userID, p.CategoryID, p.Date)
		if err != nil {
			logger.Error("failed to get active budgets", "userID", userID, "categoryID", p.CategoryID, "date", p.Date.Format(time.DateOnly), "error", err)
			continue
		}
		for _, b := range budgets {
			if b.ID != nil {
				budgetIDSet[*b.ID] = struct{}{}
			}
		}
	}

	if len(budgetIDSet) == 0 {
		return nil
	}

	// Prepare shared cache for transaction queries
	transactionCache := make(map[string][]models.Transaction)
	cacheMutex := sync.RWMutex{}

	// Process affected budgets concurrently but limited
	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var firstErr error
	var firstErrOnce sync.Once

	for budgetID := range budgetIDSet {
		wg.Add(1)
		id := budgetID
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			if err := s.fillBudgetWithExistingTransactionsOptimized(id, userID, transactionCache, &cacheMutex); err != nil {
				logger.Error("failed to update budget", "budgetID", id, "userID", userID, "error", err)
				firstErrOnce.Do(func() { firstErr = err })
			}
		}()
	}
	wg.Wait()

	return firstErr
}

func (s *BudgetsServiceInstance) fillBudgetWithExistingTransactions(budgetID int, userID int) error {
	// Get budget details
	budget, err := s.budgetsRepository.GetBudgetByID(budgetID, userID)
	if err != nil {
		return fmt.Errorf("failed to get budget %d for user %d: %w", budgetID, userID, err)
	}

	// Parse included categories
	categoryIDs, err := ParseCategoryIDsFromString(*budget.IncludedCategories)
	if err != nil {
		return fmt.Errorf("failed to parse category IDs '%s' for budget %d: %w", *budget.IncludedCategories, budgetID, err)
	}

	if len(categoryIDs) == 0 {
		return s.budgetsRepository.UpdateBudgetCollectedAmount(budgetID, decimal.Zero)
	}

	// Get expense transactions for this budget
	var transactionIds []int
	transactions, err := s.sm.TransactionsService.GetExpenseTransactionsForBudget(
		budget.UserID, categoryIDs, *budget.StartDate, *budget.EndDate, transactionIds)
	if err != nil {
		return fmt.Errorf("failed to get expense transactions for budget %d (user=%d, categories=%v, start=%v, end=%v): %w",
			budgetID, budget.UserID, categoryIDs, budget.StartDate, budget.EndDate, err)
	}

	// Calculate total collected amount in budget's currency
	totalAmount := decimal.Zero
	for _, transaction := range transactions {
		// Convert transaction amount to budget currency
		convertedAmount, err := s.convertTransactionAmountToBudgetCurrency(transaction, budget.CurrencyID)
		if err != nil {
			return fmt.Errorf("failed to convert transaction %d (amount=%s) to budget %d currency: %w",
				*transaction.ID, transaction.Amount.String(), budgetID, err)
		}

		totalAmount = totalAmount.Add(convertedAmount)
	}

	// Update budget collected amount
	err = s.budgetsRepository.UpdateBudgetCollectedAmount(budgetID, totalAmount)
	if err != nil {
		return fmt.Errorf("failed to update collected amount for budget %d to %s: %w", budgetID, totalAmount.String(), err)
	}

	return nil
}

func (s *BudgetsServiceInstance) convertTransactionAmountToBudgetCurrency(transaction models.Transaction, budgetCurrencyID int) (decimal.Decimal, error) {
	// Get transaction account to know its currency
	accountDTO, err := s.sm.AccountsService.GetAccountById(transaction.AccountID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get account %d for transaction %d: %w", transaction.AccountID, *transaction.ID, err)
	}

	// If transaction account currency matches budget currency, no conversion needed
	if accountDTO.CurrencyId == budgetCurrencyID {
		return transaction.Amount, nil
	}

	// If transaction has base_currency_amount and user's base currency matches budget currency
	userBaseCurrency, err := s.sm.UserSettingsService.GetBaseCurrency(transaction.UserID)
	if err == nil && transaction.BaseCurrencyAmount != nil && userBaseCurrency.ID == budgetCurrencyID {
		return *transaction.BaseCurrencyAmount, nil
	}

	// Get budget currency details to get the currency code
	budgetCurrency, err := s.sm.CurrenciesService.GetCurrency(budgetCurrencyID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get budget currency %d: %w", budgetCurrencyID, err)
	}

	// Need currency conversion - use exchange rate service with currency codes
	convertedAmount, err := s.sm.ExchangeRatesService.CalcAmountFromCurrency(
		*transaction.DateTime, transaction.Amount, accountDTO.Currency.Code, budgetCurrency.Code)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to convert %s %s to %s on %v: %w",
			transaction.Amount.String(), accountDTO.Currency.Code, budgetCurrency.Code, transaction.DateTime.Format("2006-01-02"), err)
	}

	return convertedAmount, nil
}

// fillBudgetWithExistingTransactionsOptimized is a cache-aware version for concurrent processing
func (s *BudgetsServiceInstance) fillBudgetWithExistingTransactionsOptimized(budgetID int, userID int, transactionCache map[string][]models.Transaction, cacheMutex *sync.RWMutex) error {
	// Get budget details
	budget, err := s.budgetsRepository.GetBudgetByID(budgetID, userID)
	if err != nil {
		return fmt.Errorf("failed to get budget %d for user %d: %w", budgetID, userID, err)
	}

	// Parse included categories
	categoryIDs, err := ParseCategoryIDsFromString(*budget.IncludedCategories)
	if err != nil {
		return fmt.Errorf("failed to parse category IDs '%s' for budget %d: %w", *budget.IncludedCategories, budgetID, err)
	}

	if len(categoryIDs) == 0 {
		return s.budgetsRepository.UpdateBudgetCollectedAmount(budgetID, decimal.Zero)
	}

	// Create cache key for this query
	cacheKey := fmt.Sprintf("%v:%s:%s", categoryIDs, budget.StartDate.Format("2006-01-02"), budget.EndDate.Format("2006-01-02"))

	// Try to get transactions from cache first
	cacheMutex.RLock()
	transactions, found := transactionCache[cacheKey]
	cacheMutex.RUnlock()

	if !found {
		// Not in cache, fetch from database
		var transactionIds []int
		transactions, err = s.sm.TransactionsService.GetExpenseTransactionsForBudget(
			budget.UserID, categoryIDs, *budget.StartDate, *budget.EndDate, transactionIds)
		if err != nil {
			return fmt.Errorf("failed to get expense transactions for budget %d (user=%d, categories=%v, start=%v, end=%v): %w",
				budgetID, budget.UserID, categoryIDs, budget.StartDate, budget.EndDate, err)
		}

		// Store in cache for other budgets
		cacheMutex.Lock()
		transactionCache[cacheKey] = transactions
		cacheMutex.Unlock()
	}

	// Calculate total collected amount in budget's currency
	totalAmount := decimal.Zero
	for _, transaction := range transactions {
		// Convert transaction amount to budget currency
		convertedAmount, err := s.convertTransactionAmountToBudgetCurrency(transaction, budget.CurrencyID)
		if err != nil {
			return fmt.Errorf("failed to convert transaction %d (amount=%s) to budget %d currency: %w",
				*transaction.ID, transaction.Amount.String(), budgetID, err)
		}

		totalAmount = totalAmount.Add(convertedAmount)
	}

	// Update budget collected amount
	err = s.budgetsRepository.UpdateBudgetCollectedAmount(budgetID, totalAmount)
	if err != nil {
		return fmt.Errorf("failed to update collected amount for budget %d to %s: %w", budgetID, totalAmount.String(), err)
	}

	return nil
}

func (s *BudgetsServiceInstance) createCopyOfOutdatedBudget(budget models.Budget) error {
	logger.Debug("createCopyOfOutdatedBudget Service")

	endDate := *budget.EndDate
	var newStartDate, newEndDate time.Time

	switch strings.ToUpper(budget.Period) {
	case string(models.PeriodDaily):
		newStartDate = endDate
		newEndDate = endDate.AddDate(0, 0, 1)
	case string(models.PeriodWeekly):
		newStartDate = endDate
		newEndDate = endDate.AddDate(0, 0, 7)
	case string(models.PeriodMonthly):
		newStartDate = endDate
		newEndDate = endDate.AddDate(0, 1, 0)
	case string(models.PeriodYearly):
		newStartDate = endDate
		newEndDate = endDate.AddDate(1, 0, 0)
	default:
		return fmt.Errorf("invalid period for auto-renewal: %s", budget.Period)
	}

	now := time.Now()
	copyName := fmt.Sprintf("%s (copy)", budget.Name)

	newBudget := models.Budget{
		UserID:             budget.UserID,
		Name:               copyName,
		CurrencyID:         budget.CurrencyID,
		TargetAmount:       budget.TargetAmount,
		CollectedAmount:    decimal.Zero,
		Period:             budget.Period,
		Repeat:             budget.Repeat,
		StartDate:          &newStartDate,
		EndDate:            &newEndDate,
		IncludedCategories: budget.IncludedCategories,
		Comment:            budget.Comment,
		IsDeleted:          false,
		IsArchived:         false,
		CreatedAt:          &now,
		UpdatedAt:          &now,
	}

	_, err := s.budgetsRepository.CreateBudget(newBudget)
	if err != nil {
		logger.Error("Error creating copy of budget", "error", err)
		return err
	}

	return nil
}
