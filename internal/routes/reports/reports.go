package reports

import (
	"net/http"
	"sort"
	"time"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/services"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

var (
	cfg *config.Config
	sm  *services.Manager
)

func RegisterReportsRoutes(g *echo.Group, cfgGlobal *config.Config, manager *services.Manager) {
	cfg = cfgGlobal
	sm = manager

	g.POST("/cashflow", GetCashFlow)
	g.POST("/balance", GetBalanceReport)
	g.POST("/balance/non-hidden", GetNonHiddenBalance)
	g.POST("/expenses-by-categories", GetExpensesByCategories)
	g.GET("/diagram/:diagram_type/:start_date/:end_date", GetDiagram)
	g.POST("/expenses-data", GetExpensesData)
}

func getUserID(c echo.Context) (int, error) {
	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	return user.ID, nil
}

func GetCashFlow(c echo.Context) error {
	log.Debugf("GetCashFlow request started: %s %s", c.Request().Method, c.Request().URL)

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var input dto.CashFlowReportInputDTO
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	result, err := sm.ReportsService.GetCashFlow(userID, input)
	if err != nil {
		log.Errorf("Error generating cash flow report for user %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating report"})
	}

	log.Debug("GetCashFlow request completed - POST /reports/cashflow")
	return c.JSON(http.StatusOK, result)
}

func GetBalanceReport(c echo.Context) error {
	log.Debugf("GetBalanceReport request started: %s %s", c.Request().Method, c.Request().URL)

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var input dto.BalanceReportInputDTO
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	result, err := sm.ReportsService.GetBalanceReport(userID, input)
	if err != nil {
		log.Errorf("Error generating balance report for user %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating report"})
	}

	log.Debug("GetBalanceReport request completed - POST /reports/balance")
	return c.JSON(http.StatusOK, result)
}

func GetNonHiddenBalance(c echo.Context) error {
	log.Debugf("GetNonHiddenBalance request started: %s %s", c.Request().Method, c.Request().URL)

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var input dto.BalanceReportInputDTO
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	result, err := sm.ReportsService.GetNonHiddenBalanceReport(userID, input)
	if err != nil {
		log.Errorf("Error generating non-hidden balance report for user %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating report"})
	}

	log.Debug("GetNonHiddenBalance request completed - POST /reports/balance/non-hidden")
	return c.JSON(http.StatusOK, result)
}

func GetExpensesByCategories(c echo.Context) error {
	log.Debugf("GetExpensesByCategories request started: %s %s", c.Request().Method, c.Request().URL)

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var input dto.ExpensesReportInputDTO
	if err := c.Bind(&input); err != nil {
		log.Errorf("Error binding expenses report input for user %d: %v", userID, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	result, err := sm.ReportsService.GetExpensesByCategories(userID, input)
	if err != nil {
		log.Errorf("Error generating expenses by categories report for user %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating report"})
	}

	log.Debug("GetExpensesByCategories request completed - POST /reports/expenses-by-categories")
	return c.JSON(http.StatusOK, result)
}

func GetDiagram(c echo.Context) error {
	log.Debugf("GetDiagram request started: %s %s", c.Request().Method, c.Request().URL)

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	diagramType := c.Param("diagram_type")
	startDateStr := c.Param("start_date")
	endDateStr := c.Param("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid start date format"})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid end date format"})
	}

	// Currently only pie diagram is supported
	if diagramType != "pie" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unsupported diagram type"})
	}

	// Get expenses data for the given date range
	data, err := sm.ReportsService.GetExpensesDiagramData(userID, startDate, endDate)
	if err != nil {
		log.Errorf("Error getting expenses data for diagram for user %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating diagram"})
	}

	// Generate pie chart image (as base64 data URL)
	chartImage, err := sm.ChartService.GeneratePieChart(data, "")
	if err != nil {
		log.Errorf("Error generating pie chart for user %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating diagram image"})
	}

	// Return the image in the expected format: { "image": "data:image/png;base64,..." }
	log.Debug("GetDiagram request completed - GET /reports/diagram/:diagram_type/:start_date/:end_date")
	return c.JSON(http.StatusOK, chartImage)
}

func GetExpensesData(c echo.Context) error {
	log.Debugf("GetExpensesData request started: %s %s", c.Request().Method, c.Request().URL)

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var input dto.ExpensesReportInputDTO
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Build parent-level aggregated data, combine small categories, and output in Python-compatible format
	items, err := sm.ReportsService.GetExpensesByCategories(userID, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error getting expenses data"})
	}

	// Aggregate to parent level
	aggregated := aggregateForExpensesData(items)
	// Combine small categories
	aggregated = combineSmallAggregated(aggregated, 0.02)
	// Sort desc
	sort.Slice(aggregated, func(i, j int) bool { return aggregated[i].Amount > aggregated[j].Amount })

	log.Debug("GetExpensesData request completed - POST /reports/expenses-data")
	return c.JSON(http.StatusOK, aggregated)
}

// aggregateForExpensesData groups items by parent category (or itself if top-level) like Python prepare_data
func aggregateForExpensesData(items []dto.ExpensesReportOutputItemDTO) []dto.AggregatedDiagramItemDTO {
	// Map parentID -> total
	totals := make(map[int]dto.AggregatedDiagramItemDTO)

	// Build parent name lookup
	parentNames := make(map[int]string)
	for _, it := range items {
		if it.ParentID == nil { // parent
			parentNames[it.ID] = it.Name
		}
	}

	for _, it := range items {
		var parentID int
		var label string
		if it.ParentID == nil { // parent
			parentID = it.ID
			label = it.Name
		} else { // child -> roll into parent
			parentID = *it.ParentID
			label = parentNames[parentID]
		}

		agg := totals[parentID]
		agg.CategoryID = parentID
		agg.Label = label
		agg.Amount += it.TotalExpenses
		totals[parentID] = agg
	}

	out := make([]dto.AggregatedDiagramItemDTO, 0, len(totals))
	for _, v := range totals {
		out = append(out, v)
	}
	return out
}

// combineSmallAggregated merges small categories into "Other" for the aggregated items
func combineSmallAggregated(items []dto.AggregatedDiagramItemDTO, threshold float64) []dto.AggregatedDiagramItemDTO {
	if len(items) == 0 {
		return items
	}
	var total float64
	for _, it := range items {
		total += it.Amount
	}
	if total <= 0 {
		return items
	}
	large := make([]dto.AggregatedDiagramItemDTO, 0, len(items))
	var other float64
	for _, it := range items {
		if it.Amount/total < threshold {
			other += it.Amount
		} else {
			large = append(large, it)
		}
	}
	if other > 0 {
		large = append(large, dto.AggregatedDiagramItemDTO{CategoryID: 0, Label: "Other", Amount: other})
	}
	return large
}
