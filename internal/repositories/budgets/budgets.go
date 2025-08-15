package budgets

import (
	"fmt"
	"strings"
	"time"
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type Repository interface {
	CreateBudget(budget models.Budget) (*models.Budget, error)
	UpdateBudget(budget models.Budget) error
	GetBudgetByID(budgetID int, userID int) (*models.Budget, error)
	GetUserBudgets(userID int, include string) ([]models.Budget, error)
	DeleteBudget(budgetID int, userID int) error
	ArchiveBudget(budgetID int, userID int) error
	GetBudgetsWithCurrency(userID int, include string) ([]BudgetWithCurrency, error)
	UpdateBudgetCollectedAmount(budgetID int, amount decimal.Decimal) error
	GetOutdatedBudgets() ([]models.Budget, error)
	GetUserCategoriesForBudget(userID int, categoryIDs []int) ([]int, error)
	// GetActiveBudgetsByCategoryAndDate returns budgets for a user whose period covers the given date
	// and include the given category ID in their included_categories list. Includes archived budgets.
	GetActiveBudgetsByCategoryAndDate(userID int, categoryID int, date time.Time) ([]models.Budget, error)
}

type RepositoryInstance struct{}

type BudgetWithCurrency struct {
	models.Budget
	Currency models.Currency `db:"currency"`
}

var db *sqlx.DB

func NewBudgetsRepository(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (r *RepositoryInstance) CreateBudget(budget models.Budget) (*models.Budget, error) {
	const createBudgetQuery = `
INSERT INTO budgets (user_id, name, currency_id, target_amount, collected_amount, period, repeat, 
                     start_date, end_date, included_categories, comment, is_deleted, is_archived, 
                     created_at, updated_at)
VALUES (:user_id, :name, :currency_id, :target_amount, :collected_amount, :period, :repeat, 
        :start_date, :end_date, :included_categories, :comment, :is_deleted, :is_archived, 
        :created_at, :updated_at)
RETURNING id
`

	stmt, err := db.PrepareNamed(createBudgetQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.Get(&id, budget)
	if err != nil {
		return nil, err
	}

	budget.ID = &id
	return &budget, nil
}

func (r *RepositoryInstance) UpdateBudget(budget models.Budget) error {
	const updateBudgetQuery = `
UPDATE budgets SET
    name = :name,
    currency_id = :currency_id,
    target_amount = :target_amount,
    collected_amount = :collected_amount,
    period = :period,
    repeat = :repeat,
    start_date = :start_date,
    end_date = :end_date,
    included_categories = :included_categories,
    comment = :comment,
    updated_at = :updated_at
WHERE id = :id AND user_id = :user_id
`

	_, err := db.NamedExec(updateBudgetQuery, budget)
	return err
}

func (r *RepositoryInstance) GetBudgetByID(budgetID int, userID int) (*models.Budget, error) {
	const getBudgetQuery = `
SELECT id, user_id, name, currency_id, target_amount, collected_amount, period, repeat,
       start_date, end_date, included_categories, comment, is_deleted, is_archived,
       created_at, updated_at
FROM budgets 
WHERE id = $1 AND user_id = $2 AND is_deleted = false
`

	var budget models.Budget
	err := db.Get(&budget, getBudgetQuery, budgetID, userID)
	if err != nil {
		return nil, err
	}

	return &budget, nil
}

func (r *RepositoryInstance) GetUserBudgets(userID int, include string) ([]models.Budget, error) {
	baseQuery := `
SELECT id, user_id, name, currency_id, target_amount, collected_amount, period, repeat,
       start_date, end_date, included_categories, comment, is_deleted, is_archived,
       created_at, updated_at
FROM budgets 
WHERE user_id = $1 AND is_deleted = false
`

	var whereClause string
	switch include {
	case "active":
		whereClause = " AND is_archived = false"
	case "archived":
		whereClause = " AND is_archived = true"
	case "all":
		whereClause = ""
	default:
		return nil, fmt.Errorf("invalid include parameter: %s", include)
	}

	query := baseQuery + whereClause + " ORDER BY is_archived ASC, end_date ASC, name ASC"

	var budgets []models.Budget
	err := db.Select(&budgets, query, userID)
	if err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *RepositoryInstance) GetBudgetsWithCurrency(userID int, include string) ([]BudgetWithCurrency, error) {
	baseQuery := `
SELECT b.id, b.user_id, b.name, b.currency_id, b.target_amount, b.collected_amount, 
       b.period, b.repeat, b.start_date, b.end_date, b.included_categories, b.comment, 
       b.is_deleted, b.is_archived, b.created_at, b.updated_at,
       c.id as "currency.id", c.code as "currency.code", c.name as "currency.name"
FROM budgets b
JOIN currencies c ON b.currency_id = c.id
WHERE b.user_id = $1 AND b.is_deleted = false
`

	var whereClause string
	switch include {
	case "active":
		whereClause = " AND b.is_archived = false"
	case "archived":
		whereClause = " AND b.is_archived = true"
	case "all":
		whereClause = ""
	default:
		return nil, fmt.Errorf("invalid include parameter: %s", include)
	}

	query := baseQuery + whereClause + " ORDER BY b.is_archived ASC, b.end_date ASC, b.name ASC"

	var budgets []BudgetWithCurrency
	err := db.Select(&budgets, query, userID)
	if err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *RepositoryInstance) DeleteBudget(budgetID int, userID int) error {
	const deleteBudgetQuery = `
UPDATE budgets SET is_deleted = true, updated_at = NOW()
WHERE id = $1 AND user_id = $2
`

	result, err := db.Exec(deleteBudgetQuery, budgetID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("budget not found")
	}

	return nil
}

func (r *RepositoryInstance) ArchiveBudget(budgetID int, userID int) error {
	const archiveBudgetQuery = `
UPDATE budgets SET is_archived = true, updated_at = NOW()
WHERE id = $1 AND user_id = $2
`

	result, err := db.Exec(archiveBudgetQuery, budgetID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("budget not found")
	}

	return nil
}

func (r *RepositoryInstance) UpdateBudgetCollectedAmount(budgetID int, amount decimal.Decimal) error {
	const updateAmountQuery = `
UPDATE budgets SET collected_amount = $1, updated_at = NOW()
WHERE id = $2
`

	_, err := db.Exec(updateAmountQuery, amount, budgetID)
	return err
}

func (r *RepositoryInstance) GetOutdatedBudgets() ([]models.Budget, error) {
	const getOutdatedQuery = `
SELECT id, user_id, name, currency_id, target_amount, collected_amount, period, repeat,
       start_date, end_date, included_categories, comment, is_deleted, is_archived,
       created_at, updated_at
FROM budgets 
WHERE end_date < NOW() AND is_archived = false AND is_deleted = false
`

	var budgets []models.Budget
	err := db.Select(&budgets, getOutdatedQuery)
	if err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *RepositoryInstance) GetUserCategoriesForBudget(userID int, categoryIDs []int) ([]int, error) {
	if len(categoryIDs) == 0 {
		return []int{}, nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(categoryIDs))
	args := make([]interface{}, len(categoryIDs)+1)
	args[0] = userID

	for i, id := range categoryIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
SELECT id FROM user_categories 
WHERE user_id = $1 AND id IN (%s) AND is_deleted = false
`, strings.Join(placeholders, ","))

	var validCategories []int
	err := db.Select(&validCategories, query, args...)
	if err != nil {
		return nil, err
	}

	return validCategories, nil
}

// GetActiveBudgetsByCategoryAndDate returns budgets whose period covers the given date that include categoryID.
// Archived budgets are included; deleted budgets are excluded.
func (r *RepositoryInstance) GetActiveBudgetsByCategoryAndDate(userID int, categoryID int, date time.Time) ([]models.Budget, error) {
	// included_categories is stored as comma-separated string of ints
	// Use string_to_array to convert to int[] and check membership with ANY()
	const q = `
SELECT id, user_id, name, currency_id, target_amount, collected_amount, period, repeat,
       start_date, end_date, included_categories, comment, is_deleted, is_archived,
       created_at, updated_at
FROM budgets
WHERE user_id = $1
  AND is_deleted = false
  AND start_date <= $2
  AND end_date > $2
  AND $3 = ANY(string_to_array(NULLIF(included_categories,''), ',')::int[])
`

	var budgets []models.Budget
	if err := db.Select(&budgets, q, userID, date, categoryID); err != nil {
		return nil, err
	}
	return budgets, nil
}
