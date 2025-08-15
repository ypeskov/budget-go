package dto

import (
	"ypeskov/budget-go/internal/utils"
)

// CashFlowReportInputDTO represents input for cash flow report
type CashFlowReportInputDTO struct {
	StartDate *utils.CustomDate `json:"startDate"`
	EndDate   *utils.CustomDate `json:"endDate"`
	Period    string            `json:"period" binding:"required"`
}

// CashFlowReportOutputDTO represents cash flow report output
type CashFlowReportOutputDTO struct {
	Currency      string             `json:"currency"`
	TotalIncome   map[string]float64 `json:"totalIncome"`
	TotalExpenses map[string]float64 `json:"totalExpenses"`
	NetFlow       map[string]float64 `json:"netFlow"`
}

// BalanceReportInputDTO represents input for balance report
type BalanceReportInputDTO struct {
	AccountIds  []int             `json:"account_ids"`
	BalanceDate utils.CustomDate `json:"balanceDate"`
}

// BalanceReportOutputDTO represents balance report output
type BalanceReportOutputDTO struct {
	AccountID           int     `json:"accountId" db:"account_id"`
	AccountName         string  `json:"accountName" db:"account_name"`
	CurrencyCode        string  `json:"currencyCode" db:"currency_code"`
	Balance             float64 `json:"balance" db:"balance"`
	BaseCurrencyBalance float64 `json:"baseCurrencyBalance" db:"base_currency_balance"`
	BaseCurrencyCode    string  `json:"baseCurrencyCode" db:"base_currency_code"`
	ReportDate          string  `json:"reportDate" db:"report_date"`
}

// ExpensesReportInputDTO represents input for expenses report
type ExpensesReportInputDTO struct {
	StartDate            utils.CustomDate `json:"startDate" binding:"required"`
	EndDate              utils.CustomDate `json:"endDate" binding:"required"`
	Categories           []int            `json:"categories"`
	HideEmptyCategories  bool             `json:"hideEmptyCategories"`
}

// ExpensesReportOutputItemDTO represents an item in expenses report
type ExpensesReportOutputItemDTO struct {
	ID            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	ParentID      *int    `json:"parentId" db:"parent_id"`
	ParentName    *string `json:"parentName" db:"parent_name"`
	TotalExpenses float64 `json:"totalExpenses" db:"total_expenses"`
	CurrencyCode  *string `json:"currencyCode" db:"currency_code"`
	IsParent      bool    `json:"isParent" db:"is_parent"`
}

// ExpensesDiagramDataDTO represents data for expenses diagram
type ExpensesDiagramDataDTO struct {
	CategoryName string  `json:"categoryName"`
	Amount       float64 `json:"amount"`
	Color        string  `json:"color"`
}

// ChartImageDTO represents the response for chart generation
type ChartImageDTO struct {
	Image string `json:"image"`
}

// AggregatedDiagramItemDTO represents aggregated parent-category data for diagrams
type AggregatedDiagramItemDTO struct {
    CategoryID int     `json:"category_id"`
    Label      string  `json:"label"`
    Amount     float64 `json:"amount"`
}