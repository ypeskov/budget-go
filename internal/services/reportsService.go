package services

import (
	"sort"
	"strings"
	"sync"
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/repositories/reports"
	"ypeskov/budget-go/internal/utils"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type ReportsService interface {
	GetCashFlow(userID int, input dto.CashFlowReportInputDTO) (*dto.CashFlowReportOutputDTO, error)
	GetBalanceReport(userID int, input dto.BalanceReportInputDTO) ([]dto.BalanceReportOutputDTO, error)
	GetNonHiddenBalanceReport(userID int, input dto.BalanceReportInputDTO) ([]dto.BalanceReportOutputDTO, error)
	GetExpensesByCategories(userID int, input dto.ExpensesReportInputDTO) ([]dto.ExpensesReportOutputItemDTO, error)
	GetExpensesDiagramData(userID int, startDate, endDate time.Time) ([]dto.ExpensesDiagramDataDTO, error)
}

type ReportsServiceInstance struct {
	reportsRepo          *reports.ReportsRepository
	exchangeRatesService ExchangeRatesService
}

var (
	reportsInstance *ReportsServiceInstance
	reportsOnce     sync.Once
)

func NewReportsService(reportsRepo *reports.ReportsRepository, exchangeRatesService ExchangeRatesService) ReportsService {
	reportsOnce.Do(func() {
		log.Debug("Creating ReportsService instance")
		reportsInstance = &ReportsServiceInstance{
			reportsRepo:          reportsRepo,
			exchangeRatesService: exchangeRatesService,
		}
	})

	return reportsInstance
}

func (s *ReportsServiceInstance) GetCashFlow(userID int, input dto.CashFlowReportInputDTO) (*dto.CashFlowReportOutputDTO, error) {
	// Get user's base currency
	baseCurrencyCode, err := s.reportsRepo.GetUserBaseCurrency(userID)
	if err != nil {
		return nil, err
	}

	// Get raw data per account and currency
	rawData, err := s.reportsRepo.GetCashFlowRawData(userID, input)
	if err != nil {
		return nil, err
	}

	totalIncome := make(map[string]decimal.Decimal)
	totalExpenses := make(map[string]decimal.Decimal)
	netFlow := make(map[string]decimal.Decimal)

	// Process each account and convert currency
	for _, data := range rawData {
		// Parse period date for exchange rate lookup (use mid-period date)
		var periodDate time.Time
		switch input.Period {
		case "monthly":
			periodDate, _ = time.Parse("2006-01", data.Period)
			periodDate = periodDate.AddDate(0, 0, 15) // Mid-month
		case "daily":
			periodDate, _ = time.Parse("2006-01-02", data.Period)
		default:
			periodDate, _ = time.Parse("2006-01", data.Period)
			periodDate = periodDate.AddDate(0, 0, 15) // Mid-month
		}

		// Convert income to base currency
		convertedIncome, err := s.exchangeRatesService.CalcAmountFromCurrency(
			periodDate, data.TotalIncome, data.CurrencyCode, baseCurrencyCode)
		if err != nil {
			// If conversion fails, use original amount (this matches FastAPI behavior)
			convertedIncome = data.TotalIncome
		}
		// Convert expenses to base currency
		convertedExpenses, err := s.exchangeRatesService.CalcAmountFromCurrency(
			periodDate, data.TotalExpenses, data.CurrencyCode, baseCurrencyCode)
		if err != nil {
			// If conversion fails, use original amount
			convertedExpenses = data.TotalExpenses
		}

		// Accumulate by period
		if _, exists := totalIncome[data.Period]; !exists {
			totalIncome[data.Period] = decimal.Zero
			totalExpenses[data.Period] = decimal.Zero
		}

		totalIncome[data.Period] = totalIncome[data.Period].Add(convertedIncome)
		totalExpenses[data.Period] = totalExpenses[data.Period].Add(convertedExpenses)
		netFlow[data.Period] = totalIncome[data.Period].Sub(totalExpenses[data.Period])
	}

	// Convert decimals to float64 for JSON marshaling (round to 2 decimal places for financial precision)
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

func (s *ReportsServiceInstance) GetBalanceReport(userID int, input dto.BalanceReportInputDTO) ([]dto.BalanceReportOutputDTO, error) {
	results, err := s.reportsRepo.GetBalanceReport(userID, input)
	if err != nil {
		return nil, err
	}

	// Convert each account balance into user's base currency using the balance date
	// Repository already filled BaseCurrencyCode; use that as conversion target
	for i := range results {
		amountDec := decimal.NewFromFloat(results[i].Balance)
		converted, convErr := s.exchangeRatesService.CalcAmountFromCurrency(
			input.BalanceDate.Time,
			amountDec,
			results[i].CurrencyCode,
			results[i].BaseCurrencyCode,
		)
		if convErr != nil {
			// Fallback to original balance if conversion fails
			converted = amountDec
		}
		results[i].BaseCurrencyBalance, _ = converted.Round(2).Float64()
	}

	return results, nil
}

func (s *ReportsServiceInstance) GetNonHiddenBalanceReport(userID int, input dto.BalanceReportInputDTO) ([]dto.BalanceReportOutputDTO, error) {
	results, err := s.reportsRepo.GetNonHiddenBalanceReport(userID, input)
	if err != nil {
		return nil, err
	}

	for i := range results {
		amountDec := decimal.NewFromFloat(results[i].Balance)
		converted, convErr := s.exchangeRatesService.CalcAmountFromCurrency(
			input.BalanceDate.Time,
			amountDec,
			results[i].CurrencyCode,
			results[i].BaseCurrencyCode,
		)
		if convErr != nil {
			converted = amountDec
		}
		results[i].BaseCurrencyBalance, _ = converted.Round(2).Float64()
	}

	return results, nil
}

func (s *ReportsServiceInstance) GetExpensesByCategories(userID int, input dto.ExpensesReportInputDTO) ([]dto.ExpensesReportOutputItemDTO, error) {
	// Match FastAPI: convert amounts to user's base currency per transaction
	baseCurrency, err := s.reportsRepo.GetUserBaseCurrency(userID)
	if err != nil {
		return nil, err
	}

	// Load flat categories like repository currently does
	categories, err := s.reportsRepo.GetExpensesByCategories(userID, dto.ExpensesReportInputDTO{
		StartDate:           input.StartDate,
		EndDate:             input.EndDate,
		Categories:          input.Categories,
		HideEmptyCategories: false, // get full set
	})
	if err != nil {
		return nil, err
	}

	// Build index for quick lookup
	byID := make(map[int]*dto.ExpensesReportOutputItemDTO)
	for i := range categories {
		categories[i].TotalExpenses = 0
	}
	for i := range categories {
		byID[categories[i].ID] = &categories[i]
	}

	// Fetch raw transactions for conversion
	rawRows, err := s.reportsRepo.GetRawExpensesRows(userID, input)
	if err != nil {
		return nil, err
	}

	// Sum converted amounts into categories
	for _, row := range rawRows {
		// Skip transactions without a category (NULL category_id)
		if row.CategoryID == nil {
			continue
		}

		// Convert each transaction amount to base currency using its date
		converted, convErr := s.exchangeRatesService.CalcAmountFromCurrency(row.DateTime, decimal.NewFromFloat(row.Amount), row.CurrencyCode, baseCurrency)
		if convErr != nil {
			// Fallback to original amount if conversion fails (parity with FastAPI)
			converted = decimal.NewFromFloat(row.Amount)
		}
		if cat, ok := byID[*row.CategoryID]; ok {
			val, _ := converted.Float64()
			cat.TotalExpenses += val
			cat.CurrencyCode = &baseCurrency
		}
	}

	// Apply hideEmptyCategories filter
	result := make([]dto.ExpensesReportOutputItemDTO, 0, len(categories))
	for _, c := range categories {
		if input.HideEmptyCategories && c.TotalExpenses == 0 {
			continue
		}
		result = append(result, c)
	}

	// Order by name (case-insensitive)
	sort.SliceStable(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

func (s *ReportsServiceInstance) GetExpensesDiagramData(userID int, startDate, endDate time.Time) ([]dto.ExpensesDiagramDataDTO, error) {
	input := dto.ExpensesReportInputDTO{
		StartDate:           utils.CustomDate{Time: startDate},
		EndDate:             utils.CustomDate{Time: endDate},
		HideEmptyCategories: false,
	}
	// Fetch detailed category expenses to be able to aggregate like FastAPI prepare_data (with currency conversion)
	items, err := s.GetExpensesByCategories(userID, input)
	if err != nil {
		return nil, err
	}

	// Prepare aggregated data at parent category level (like Python prepare_data)
	data := prepareDiagramData(items)

	// Combine small categories into "Other" like FastAPI implementation
	combined := combineSmallCategories(data, 0.02)

	// Sort by amount descending
	sort.Slice(combined, func(i, j int) bool { return combined[i].Amount > combined[j].Amount })

	// Assign colors deterministically after combining and map proper field names
	colors := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40", "#FF6384", "#C9CBCF"}
	for i := range combined {
		// Ensure field naming parity: use Label instead of categoryName in downstream usage
		// Here we keep CategoryName for internal chart service; the /expenses-data endpoint will output AggregatedDiagramItemDTO
		combined[i].Color = colors[i%len(colors)]
	}

	return combined, nil
}

// prepareDiagramData aggregates child categories into their parent and produces diagram data
// Equivalent to Python prepare_data(categories, category_id=None) but only for top-level parents.
func prepareDiagramData(items []dto.ExpensesReportOutputItemDTO) []dto.ExpensesDiagramDataDTO {
	// Index items by id and by parent id
	// Collect top-level parent categories (ParentID == nil)
	parents := make([]dto.ExpensesReportOutputItemDTO, 0)
	for _, it := range items {
		if it.ParentID == nil { // top-level parent
			parents = append(parents, it)
		}
	}

	// For each parent, sum its own and its direct children amounts
	result := make([]dto.ExpensesDiagramDataDTO, 0, len(parents))
	for _, parent := range parents {
		total := parent.TotalExpenses
		for _, it := range items {
			if it.ParentID != nil && parent.ID == *it.ParentID {
				total += it.TotalExpenses
			}
		}

		result = append(result, dto.ExpensesDiagramDataDTO{
			CategoryName: parent.Name, // label is parent name in Python builder
			Amount:       total,
		})
	}

	return result
}

// combineSmallCategories merges categories contributing less than `threshold` (fraction of total)
// into a single "Other" category. Threshold defaults to 0.02 (2%) in caller.
func combineSmallCategories(data []dto.ExpensesDiagramDataDTO, threshold float64) []dto.ExpensesDiagramDataDTO {
	if len(data) == 0 {
		return data
	}

	// Compute total amount
	var total float64
	for _, d := range data {
		total += d.Amount
	}
	if total <= 0 {
		return data
	}

	// Separate large and small categories
	large := make([]dto.ExpensesDiagramDataDTO, 0, len(data))
	var otherAmount float64
	for _, d := range data {
		size := d.Amount / total
		if size < threshold {
			otherAmount += d.Amount
		} else {
			large = append(large, dto.ExpensesDiagramDataDTO{CategoryName: d.CategoryName, Amount: d.Amount})
		}
	}

	if otherAmount > 0 {
		large = append(large, dto.ExpensesDiagramDataDTO{CategoryName: "Other", Amount: otherAmount})
	}

	return large
}
