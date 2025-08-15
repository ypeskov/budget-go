package reports

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"ypeskov/budget-go/internal/dto"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type ReportsRepository struct {
	db *sqlx.DB
}

func NewReportsRepository(db *sqlx.DB) *ReportsRepository {
	return &ReportsRepository{
		db: db,
	}
}

func (r *ReportsRepository) GetUserBaseCurrency(userID int) (string, error) {
	var baseCurrencyCode string
	err := r.db.Get(&baseCurrencyCode, `
		SELECT c.code 
		FROM currencies c 
		JOIN users u ON u.base_currency_id = c.id 
		WHERE u.id = $1`, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user base currency: %w", err)
	}
	return baseCurrencyCode, nil
}

// CashFlowRawData represents raw cash flow data per account and period
type CashFlowRawData struct {
	AccountID     int
	Period        string
	TotalIncome   decimal.Decimal
	TotalExpenses decimal.Decimal
	CurrencyCode  string
}

func (r *ReportsRepository) GetCashFlowRawData(userID int, input dto.CashFlowReportInputDTO) ([]CashFlowRawData, error) {
	// Build the query based on period - match FastAPI approach
	var periodFormat string
	switch strings.ToLower(input.Period) {
	case "daily":
		periodFormat = "TO_CHAR(t.date_time, 'YYYY-MM-DD')"
	case "weekly":
		periodFormat = "TO_CHAR(t.date_time, 'YYYY-\"W\"IW')"
	case "monthly":
		periodFormat = "TO_CHAR(t.date_time, 'YYYY-MM')"
	case "yearly":
		periodFormat = "TO_CHAR(t.date_time, 'YYYY')"
	default:
		periodFormat = "TO_CHAR(t.date_time, 'YYYY-MM')"
	}

	// Query per account and period like FastAPI does
	query := fmt.Sprintf(`
		SELECT 
			a.id as account_id,
			%s as period,
			COALESCE(SUM(CASE WHEN t.is_income = true THEN t.amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN t.is_income = false THEN t.amount ELSE 0 END), 0) as total_expenses,
			c.code as currency_code
		FROM accounts a
		JOIN currencies c ON a.currency_id = c.id
		LEFT JOIN transactions t ON a.id = t.account_id 
		  AND t.is_deleted = false 
		  AND t.is_transfer = false`, periodFormat)

	args := []interface{}{userID}
	argIndex := 1

	query += fmt.Sprintf(" WHERE a.user_id = $%d", argIndex)

	if input.StartDate != nil && !input.StartDate.IsZero() {
		argIndex++
		query += fmt.Sprintf(" AND (t.date_time IS NULL OR t.date_time >= $%d)", argIndex)
		args = append(args, input.StartDate.Time)
	}

	if input.EndDate != nil && !input.EndDate.IsZero() {
		argIndex++
		query += fmt.Sprintf(" AND (t.date_time IS NULL OR t.date_time <= $%d)", argIndex)
		args = append(args, input.EndDate.Time)
	}

	query += fmt.Sprintf(" GROUP BY a.id, %s, c.code HAVING SUM(CASE WHEN t.is_income = true THEN t.amount ELSE 0 END) + SUM(CASE WHEN t.is_income = false THEN t.amount ELSE 0 END) > 0 ORDER BY %s", periodFormat, periodFormat)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cash flow query: %w", err)
	}
	defer rows.Close()

	var results []CashFlowRawData
	for rows.Next() {
		var accountID int
		var period string
		var income, expenses sql.NullFloat64
		var currencyCode string

		if err := rows.Scan(&accountID, &period, &income, &expenses, &currencyCode); err != nil {
			return nil, fmt.Errorf("failed to scan cash flow row: %w", err)
		}

		// Skip empty periods
		if period == "" {
			continue
		}

		incomeVal := decimal.Zero
		if income.Valid {
			incomeVal = decimal.NewFromFloat(income.Float64)
		}

		expensesVal := decimal.Zero
		if expenses.Valid {
			expensesVal = decimal.NewFromFloat(expenses.Float64)
		}

		results = append(results, CashFlowRawData{
			AccountID:     accountID,
			Period:        period,
			TotalIncome:   incomeVal,
			TotalExpenses: expensesVal,
			CurrencyCode:  currencyCode,
		})
	}

	return results, nil
}

func (r *ReportsRepository) GetCashFlow(userID int, input dto.CashFlowReportInputDTO) (*dto.CashFlowReportOutputDTO, error) {
	// Get user's base currency
	var baseCurrencyCode string
	err := r.db.Get(&baseCurrencyCode, `
		SELECT c.code 
		FROM currencies c 
		JOIN users u ON u.base_currency_id = c.id 
		WHERE u.id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user base currency: %w", err)
	}

	rawData, err := r.GetCashFlowRawData(userID, input)
	if err != nil {
		return nil, err
	}

	totalIncome := make(map[string]decimal.Decimal)
	totalExpenses := make(map[string]decimal.Decimal)
	netFlow := make(map[string]decimal.Decimal)

	for _, data := range rawData {
		if _, exists := totalIncome[data.Period]; !exists {
			totalIncome[data.Period] = decimal.Zero
		}
		if _, exists := totalExpenses[data.Period]; !exists {
			totalExpenses[data.Period] = decimal.Zero
		}

		// Note: This doesn't do currency conversion - that should be done in the service layer
		totalIncome[data.Period] = totalIncome[data.Period].Add(data.TotalIncome)
		totalExpenses[data.Period] = totalExpenses[data.Period].Add(data.TotalExpenses)
		netFlow[data.Period] = totalIncome[data.Period].Sub(totalExpenses[data.Period])
	}

	// Convert decimals to float64 for JSON marshaling
	totalIncomeFloat := make(map[string]float64)
	totalExpensesFloat := make(map[string]float64)
	netFlowFloat := make(map[string]float64)

	for period, income := range totalIncome {
		totalIncomeFloat[period], _ = income.Round(2).Float64()
	}
	for period, expenses := range totalExpenses {
		totalExpensesFloat[period], _ = expenses.Round(2).Float64()
	}
	for period, net := range netFlow {
		netFlowFloat[period], _ = net.Round(2).Float64()
	}

	return &dto.CashFlowReportOutputDTO{
		Currency:      baseCurrencyCode,
		TotalIncome:   totalIncomeFloat,
		TotalExpenses: totalExpensesFloat,
		NetFlow:       netFlowFloat,
	}, nil
}

func (r *ReportsRepository) GetBalanceReport(userID int, input dto.BalanceReportInputDTO) ([]dto.BalanceReportOutputDTO, error) {
	return r.getBalanceReportWithFilter(userID, input, false)
}

func (r *ReportsRepository) GetNonHiddenBalanceReport(userID int, input dto.BalanceReportInputDTO) ([]dto.BalanceReportOutputDTO, error) {
	return r.getBalanceReportWithFilter(userID, input, true)
}

func (r *ReportsRepository) getBalanceReportWithFilter(userID int, input dto.BalanceReportInputDTO, nonHiddenOnly bool) ([]dto.BalanceReportOutputDTO, error) {
	// Get user's base currency
	var baseCurrencyCode string
	err := r.db.Get(&baseCurrencyCode, `
		SELECT c.code 
		FROM currencies c 
		JOIN users u ON u.base_currency_id = c.id 
		WHERE u.id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user base currency: %w", err)
	}

	// Build args list: [userID, balanceDate, accountIDs..., baseCurrency, reportDate]
	args := []interface{}{userID, input.BalanceDate.Time.Add(24 * time.Hour)}
	argIndex := 2

	// Add account IDs to args first
	var accountPlaceholders []string
	if len(input.AccountIds) > 0 {
		accountPlaceholders = make([]string, len(input.AccountIds))
		for i, accountID := range input.AccountIds {
			argIndex++
			accountPlaceholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, accountID)
		}
	}

	// Add base currency and report date
	argIndex++
	baseCurrencyIndex := argIndex
	args = append(args, baseCurrencyCode)

	argIndex++
	reportDateIndex := argIndex
	args = append(args, input.BalanceDate.Time.Format("2006-01-02"))

	// Build the query
	query := `
		WITH latest_transactions AS (
			SELECT 
				t.account_id,
				t.new_balance,
				ROW_NUMBER() OVER (PARTITION BY t.account_id ORDER BY t.date_time DESC) as rn
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			WHERE a.user_id = $1 
			  AND t.date_time <= $2`

	// Add account filter to CTE if specified
	if len(input.AccountIds) > 0 {
		// Include specified accounts OR accounts marked to show in reports
		query += fmt.Sprintf(" AND (t.account_id IN (%s) OR a.show_in_reports = true)", strings.Join(accountPlaceholders, ","))
	} else {
		// No account filter specified: include only accounts marked to show in reports
		query += " AND a.show_in_reports = true"
	}

	// The main query - we need to filter accounts here too!
	if len(input.AccountIds) > 0 {
		// If specific accounts requested, only get those accounts
		query += fmt.Sprintf(`
		)
		SELECT 
			a.id as account_id,
			a.name as account_name,
			c.code as currency_code,
			COALESCE(lt.new_balance, 0) as balance,
			COALESCE(lt.new_balance, 0) as base_currency_balance,
			$%d as base_currency_code,
			$%d as report_date
		FROM accounts a
		JOIN currencies c ON a.currency_id = c.id
		LEFT JOIN latest_transactions lt ON a.id = lt.account_id AND lt.rn = 1
        WHERE a.user_id = $1 AND (a.id IN (%s) OR a.show_in_reports = true)`, baseCurrencyIndex, reportDateIndex, strings.Join(accountPlaceholders, ","))
	} else {
		// No specific accounts requested
		query += fmt.Sprintf(`
		)
		SELECT 
			a.id as account_id,
			a.name as account_name,
			c.code as currency_code,
			COALESCE(lt.new_balance, 0) as balance,
			COALESCE(lt.new_balance, 0) as base_currency_balance,
			$%d as base_currency_code,
			$%d as report_date
		FROM accounts a
		JOIN currencies c ON a.currency_id = c.id
		LEFT JOIN latest_transactions lt ON a.id = lt.account_id AND lt.rn = 1
        WHERE a.user_id = $1 AND a.show_in_reports = true`, baseCurrencyIndex, reportDateIndex)
	}

	query += " ORDER BY LOWER(a.name)"

	var results []dto.BalanceReportOutputDTO
	err = r.db.Select(&results, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute balance report query: %w", err)
	}

	return results, nil
}

func (r *ReportsRepository) GetExpensesByCategories(userID int, input dto.ExpensesReportInputDTO) ([]dto.ExpensesReportOutputItemDTO, error) {
	// Get user's base currency
	baseCurrencyCode, err := r.GetUserBaseCurrency(userID)
	if err != nil {
		return nil, err
	}

	// First, get all user categories like Python does
	allCategoriesQuery := `
		SELECT 
			cat.id,
			cat.name,
			cat.parent_id,
			parent_cat.name as parent_name,
			CASE WHEN cat.parent_id IS NULL THEN true ELSE false END as is_parent
		FROM user_categories cat
		LEFT JOIN user_categories parent_cat ON cat.parent_id = parent_cat.id
		WHERE cat.user_id = $1 
		  AND cat.is_deleted = false 
		  AND cat.is_income = false`

	args := []interface{}{userID}

	if len(input.Categories) > 0 {
		placeholders := make([]string, len(input.Categories))
		for i := range input.Categories {
			placeholders[i] = fmt.Sprintf("$%d", len(args)+1+i)
		}
		allCategoriesQuery += fmt.Sprintf(" AND cat.id IN (%s)", strings.Join(placeholders, ","))
		for _, catID := range input.Categories {
			args = append(args, catID)
		}
	}

	allCategoriesQuery += " ORDER BY cat.name"

	// Execute query to get all categories
	type CategoryRow struct {
		ID         int     `db:"id"`
		Name       string  `db:"name"`
		ParentID   *int    `db:"parent_id"`
		ParentName *string `db:"parent_name"`
		IsParent   bool    `db:"is_parent"`
	}

	var allCategories []CategoryRow
	err = r.db.Select(&allCategories, allCategoriesQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user categories: %w", err)
	}

	// Create results map initialized with zero expenses
	results := make(map[int]dto.ExpensesReportOutputItemDTO)
	for _, cat := range allCategories {
		displayName := cat.Name
		if cat.ParentID != nil {
			// Format child categories like Python: "Parent >> Child"
			displayName = fmt.Sprintf("%s >> %s", *cat.ParentName, cat.Name)
		}

		results[cat.ID] = dto.ExpensesReportOutputItemDTO{
			ID:            cat.ID,
			Name:          displayName,
			ParentID:      cat.ParentID,
			ParentName:    cat.ParentName,
			TotalExpenses: 0.0,
			CurrencyCode:  nil,
			IsParent:      cat.IsParent,
		}
	}

	// Now get actual expenses for the date range
	expensesQuery := `
		SELECT 
			t.category_id,
			ABS(t.amount) as amount,
			c.code as currency_code
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		JOIN currencies c ON a.currency_id = c.id
		WHERE a.user_id = $1
		  AND t.date_time >= $2
		  AND t.date_time <= $3
		  AND t.is_income = false
		  AND t.is_deleted = false
		  AND t.is_transfer = false`

	expensesArgs := []interface{}{userID, input.StartDate.Time, input.EndDate.Time}

	if len(input.Categories) > 0 {
		placeholders := make([]string, len(input.Categories))
		for i := range input.Categories {
			placeholders[i] = fmt.Sprintf("$%d", len(expensesArgs)+1+i)
		}
		expensesQuery += fmt.Sprintf(" AND t.category_id IN (%s)", strings.Join(placeholders, ","))
		for _, catID := range input.Categories {
			expensesArgs = append(expensesArgs, catID)
		}
	}

	type ExpenseRow struct {
		CategoryID   *int    `db:"category_id"`
		Amount       float64 `db:"amount"`
		CurrencyCode string  `db:"currency_code"`
	}

	var expenses []ExpenseRow
	err = r.db.Select(&expenses, expensesQuery, expensesArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Aggregate expenses by category (without currency conversion for now)
	categoryExpenses := make(map[int]float64)
	for _, expense := range expenses {
		// Skip transactions without a category (NULL category_id)
		if expense.CategoryID == nil {
			continue
		}
		categoryExpenses[*expense.CategoryID] += expense.Amount
	}

	// Update results with actual expenses
	for categoryID, totalExpense := range categoryExpenses {
		if category, exists := results[categoryID]; exists {
			category.TotalExpenses = totalExpense
			category.CurrencyCode = &baseCurrencyCode
			results[categoryID] = category
		}
	}

	// Convert map to slice - initialize with empty slice to avoid null JSON
	finalResults := make([]dto.ExpensesReportOutputItemDTO, 0)
	for _, result := range results {
		// Apply hide empty categories filter
		if input.HideEmptyCategories && result.TotalExpenses == 0 {
			continue
		}
		finalResults = append(finalResults, result)
	}

	return finalResults, nil
}

// ExpenseRawRow represents a single expense transaction row for conversion/aggregation in services
type ExpenseRawRow struct {
	CategoryID   *int      `db:"category_id"`
	Amount       float64   `db:"amount"`
	CurrencyCode string    `db:"currency_code"`
	DateTime     time.Time `db:"date_time"`
}

// GetRawExpensesRows returns per-transaction expenses for the given period and optional category filter.
// This is used by services to perform currency conversion like the FastAPI implementation.
func (r *ReportsRepository) GetRawExpensesRows(userID int, input dto.ExpensesReportInputDTO) ([]ExpenseRawRow, error) {
	query := `
        SELECT 
            t.category_id,
            ABS(t.amount) as amount,
            c.code as currency_code,
            t.date_time
        FROM transactions t
        JOIN accounts a ON t.account_id = a.id
        JOIN currencies c ON a.currency_id = c.id
        WHERE a.user_id = $1
          AND t.date_time >= $2
          AND t.date_time < $3
          AND t.is_income = false
          AND t.is_deleted = false
          AND t.is_transfer = false`

	args := []interface{}{userID, input.StartDate.Time, input.EndDate.Time.Add(24 * time.Hour)}

	if len(input.Categories) > 0 {
		placeholders := make([]string, len(input.Categories))
		for i := range input.Categories {
			placeholders[i] = fmt.Sprintf("$%d", len(args)+1+i)
		}
		query += fmt.Sprintf(" AND t.category_id IN (%s)", strings.Join(placeholders, ","))
		for _, catID := range input.Categories {
			args = append(args, catID)
		}
	}

	var rows []ExpenseRawRow
	err := r.db.Select(&rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw expenses: %w", err)
	}

	return rows, nil
}

func (r *ReportsRepository) GetExpensesDiagramData(userID int, input dto.ExpensesReportInputDTO) ([]dto.ExpensesDiagramDataDTO, error) {
	expenses, err := r.GetExpensesByCategories(userID, input)
	if err != nil {
		return nil, err
	}

	// Convert expenses to diagram data format
	var diagramData []dto.ExpensesDiagramDataDTO
	colors := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40", "#FF6384", "#C9CBCF"}

	for i, expense := range expenses {
		if expense.TotalExpenses > 0 {
			diagramData = append(diagramData, dto.ExpensesDiagramDataDTO{
				CategoryName: expense.Name,
				Amount:       expense.TotalExpenses,
				Color:        colors[i%len(colors)],
			})
		}
	}

	return diagramData, nil
}
