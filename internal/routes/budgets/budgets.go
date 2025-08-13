package budgets

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/routes/routeErrors"
	"ypeskov/budget-go/internal/services"
	"ypeskov/budget-go/internal/utils"
)

var (
	sm *services.Manager
)

func RegisterBudgetsRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.POST("/add", CreateBudget)
	g.GET("", GetBudgets)
	g.PUT("/:id", UpdateBudget)
	g.DELETE("/:id", DeleteBudget)
	g.PUT("/:id/archive", ArchiveBudget)
	g.GET("/daily-processing", DailyProcessing)
}

func CreateBudget(c echo.Context) error {
	log.Debug("CreateBudget Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	var budgetDTO dto.CreateBudgetDTO
	if err := c.Bind(&budgetDTO); err != nil {
		return utils.LogAndReturnError(c, err, http.StatusBadRequest)
	}

	budget, err := sm.BudgetsService.CreateBudget(budgetDTO, user.ID)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, budget)
}

func UpdateBudget(c echo.Context) error {
	log.Debug("UpdateBudget Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	idStr := c.Param("id")
	if idStr == "" {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Budget ID is required"}, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid budget ID format"}, http.StatusBadRequest)
	}

	var budgetDTO dto.UpdateBudgetDTO
	if err := c.Bind(&budgetDTO); err != nil {
		return utils.LogAndReturnError(c, err, http.StatusBadRequest)
	}

	// Set the ID from the URL parameter
	budgetDTO.ID = id

	budget, err := sm.BudgetsService.UpdateBudget(budgetDTO, user.ID)
	if err != nil {
		if err.Error() == "budget not found" {
			return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "budget", ID: id}, http.StatusNotFound)
		}
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, budget)
}

func GetBudgets(c echo.Context) error {
	log.Debug("GetBudgets Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	include := c.QueryParam("include")
	if include == "" {
		include = "all"
	}

	// Validate include parameter
	if include != "all" && include != "active" && include != "archived" {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid include parameter. Must be 'all', 'active', or 'archived'"}, http.StatusBadRequest)
	}

	budgets, err := sm.BudgetsService.GetUserBudgets(user.ID, include)
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, budgets)
}

func DeleteBudget(c echo.Context) error {
	log.Debug("DeleteBudget Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	idStr := c.Param("id")
	if idStr == "" {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Budget ID is required"}, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid budget ID format"}, http.StatusBadRequest)
	}

	err = sm.BudgetsService.DeleteBudget(id, user.ID)
	if err != nil {
		if err.Error() == "budget not found" {
			return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "budget", ID: id}, http.StatusNotFound)
		}
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Budget deleted successfully",
	})
}

func ArchiveBudget(c echo.Context) error {
	log.Debug("ArchiveBudget Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	idStr := c.Param("id")
	if idStr == "" {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Budget ID is required"}, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Invalid budget ID format"}, http.StatusBadRequest)
	}

	err = sm.BudgetsService.ArchiveBudget(id, user.ID)
	if err != nil {
		if err.Error() == "budget not found" {
			return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "budget", ID: id}, http.StatusNotFound)
		}
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Budget archived successfully",
	})
}

func DailyProcessing(c echo.Context) error {
	log.Debug("DailyProcessing Route")

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return utils.LogAndReturnError(c, &routeErrors.NotFoundError{Resource: "user", ID: 0}, http.StatusBadRequest)
	}

	// Only allow user ID 1 to access this endpoint (admin functionality)
	if user.ID != 1 {
		return utils.LogAndReturnError(c, &routeErrors.BadRequestError{Message: "Forbidden"}, http.StatusForbidden)
	}

	archivedBudgetIDs, err := sm.BudgetsService.ProcessOutdatedBudgets()
	if err != nil {
		return utils.LogAndReturnError(c, err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":           "Daily processing completed",
		"archivedBudgetIds": archivedBudgetIDs,
	})
}
