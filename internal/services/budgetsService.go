package services

import (
	"fmt"
	"strings"
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	budgetRepo "ypeskov/budget-go/internal/repositories/budgets"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type BudgetsService interface {
	CreateBudget(budgetDTO dto.CreateBudgetDTO, userID int) (*models.Budget, error)
	UpdateBudget(budgetDTO dto.UpdateBudgetDTO, userID int) (*models.Budget, error)
	GetUserBudgets(userID int, include string) ([]dto.BudgetResponseDTO, error)
	DeleteBudget(budgetID int, userID int) error
	ArchiveBudget(budgetID int, userID int) error
	ProcessOutdatedBudgets() ([]int, error)
	UpdateBudgetCollectedAmounts(userID int) error
}

type BudgetsServiceInstance struct {
	budgetsRepository budgetRepo.Repository
	sm                *Manager
}

func NewBudgetsService(budgetsRepository budgetRepo.Repository, sManager *Manager) BudgetsService {
	return &BudgetsServiceInstance{budgetsRepository: budgetsRepository, sm: sManager}
}

func (s *BudgetsServiceInstance) CreateBudget(budgetDTO dto.CreateBudgetDTO, userID int) (*models.Budget, error) {
	log.Debug("CreateBudget Service")

	// Validate period
	if !models.ValidatePeriod(budgetDTO.Period) {
		return nil, fmt.Errorf("invalid period: %s. Valid periods are: %v", budgetDTO.Period, models.GetValidPeriods())
	}

	// Validate and filter categories
	validCategories, err := s.budgetsRepository.GetUserCategoriesForBudget(userID, budgetDTO.Categories)
	if err != nil {
		log.Error("Error validating categories: ", err)
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
		log.Error("Error creating budget: ", err)
		return nil, err
	}

	// Fill budget with existing transactions
	err = s.fillBudgetWithExistingTransactions(*createdBudget.ID)
	if err != nil {
		log.Error("Error filling budget with existing transactions: ", err)
		// Don't fail the creation, just log the error
	}

	return createdBudget, nil
}

func (s *BudgetsServiceInstance) UpdateBudget(budgetDTO dto.UpdateBudgetDTO, userID int) (*models.Budget, error) {
	log.Debug("UpdateBudget Service")

	// Validate period
	if !models.ValidatePeriod(budgetDTO.Period) {
		return nil, fmt.Errorf("invalid period: %s. Valid periods are: %v", budgetDTO.Period, models.GetValidPeriods())
	}

	// Get existing budget to verify ownership
	existingBudget, err := s.budgetsRepository.GetBudgetByID(budgetDTO.ID, userID)
	if err != nil {
		log.Error("Error getting existing budget: ", err)
		return nil, fmt.Errorf("budget not found")
	}

	// Validate and filter categories
	validCategories, err := s.budgetsRepository.GetUserCategoriesForBudget(userID, budgetDTO.Categories)
	if err != nil {
		log.Error("Error validating categories: ", err)
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
		log.Error("Error updating budget: ", err)
		return nil, err
	}

	// Fill budget with existing transactions
	err = s.fillBudgetWithExistingTransactions(*budget.ID)
	if err != nil {
		log.Error("Error filling budget with existing transactions: ", err)
	}

	return &budget, nil
}

func (s *BudgetsServiceInstance) GetUserBudgets(userID int, include string) ([]dto.BudgetResponseDTO, error) {
	log.Debug("GetUserBudgets Service")

	budgetsWithCurrency, err := s.budgetsRepository.GetBudgetsWithCurrency(userID, include)
	if err != nil {
		log.Error("Error getting user budgets: ", err)
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
	log.Debug("DeleteBudget Service")

	err := s.budgetsRepository.DeleteBudget(budgetID, userID)
	if err != nil {
		log.Error("Error deleting budget: ", err)
		return err
	}

	return nil
}

func (s *BudgetsServiceInstance) ArchiveBudget(budgetID int, userID int) error {
	log.Debug("ArchiveBudget Service")

	err := s.budgetsRepository.ArchiveBudget(budgetID, userID)
	if err != nil {
		log.Error("Error archiving budget: ", err)
		return err
	}

	return nil
}

func (s *BudgetsServiceInstance) ProcessOutdatedBudgets() ([]int, error) {
	log.Debug("ProcessOutdatedBudgets Service")

	outdatedBudgets, err := s.budgetsRepository.GetOutdatedBudgets()
	if err != nil {
		log.Error("Error getting outdated budgets: ", err)
		return nil, err
	}

	archivedBudgetIDs := make([]int, 0)

	for _, budget := range outdatedBudgets {
		if budget.Repeat {
			// Create a copy for the next period
			err = s.createCopyOfOutdatedBudget(budget)
			if err != nil {
				log.Error("Error creating copy of outdated budget: ", err)
				continue
			}
		}

		// Archive the original budget
		err = s.budgetsRepository.ArchiveBudget(*budget.ID, budget.UserID)
		if err != nil {
			log.Error("Error archiving outdated budget: ", err)
			continue
		}

		archivedBudgetIDs = append(archivedBudgetIDs, *budget.ID)
	}

	return archivedBudgetIDs, nil
}

func (s *BudgetsServiceInstance) UpdateBudgetCollectedAmounts(userID int) error {
	log.Debug("UpdateBudgetCollectedAmounts Service")

	userBudgets, err := s.budgetsRepository.GetUserBudgets(userID, "all")
	if err != nil {
		log.Error("Error getting user budgets: ", err)
		return err
	}

	for _, budget := range userBudgets {
		if budget.ID != nil {
			err = s.fillBudgetWithExistingTransactions(*budget.ID)
			if err != nil {
				log.Error("Error updating budget collected amount: ", err)
				continue
			}
		}
	}

	return nil
}

func (s *BudgetsServiceInstance) fillBudgetWithExistingTransactions(budgetID int) error {
	// Get budget details
	budget, err := s.budgetsRepository.GetBudgetByID(budgetID, 0) // Use 0 as userID since we're using budgetID
	if err != nil {
		return err
	}

	// Parse included categories
	categoryIDs, err := ParseCategoryIDsFromString(*budget.IncludedCategories)
	if err != nil {
		return err
	}

	if len(categoryIDs) == 0 {
		// No categories specified, set collected amount to 0
		return s.budgetsRepository.UpdateBudgetCollectedAmount(budgetID, decimal.Zero)
	}

	// Get base currency for the user
	_, err = s.sm.UserSettingsService.GetBaseCurrency(budget.UserID)
	if err != nil {
		log.Error("Error getting base currency: ", err)
		// Continue with budget currency if base currency fails
	}

	// Get transactions within budget period and categories
	// This is a simplified version - you might want to create a specific repository method
	// For now, we'll set collected amount to 0 and log that this needs implementation
	log.Info("fillBudgetWithExistingTransactions needs full implementation with transaction filtering")

	return s.budgetsRepository.UpdateBudgetCollectedAmount(budgetID, decimal.Zero)
}

func (s *BudgetsServiceInstance) createCopyOfOutdatedBudget(budget models.Budget) error {
	log.Debug("createCopyOfOutdatedBudget Service")

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
		log.Error("Error creating copy of budget: ", err)
		return err
	}

	return nil
}
