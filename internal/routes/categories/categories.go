package categories

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"ypeskov/budget-go/internal/logger"

	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/services"
)

var (
	sm *services.Manager
)

func RegisterCategoriesRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("", GetCategories)
	g.GET("/grouped", GetGroupedCategories)
	g.POST("", CreateCategory)
}

func GetCategories(c echo.Context) error {
	logger.Debug("GetCategories request started", "method", c.Request().Method, "url", c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	userCategories, err := sm.CategoriesService.GetUserCategories(user.ID)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	var categories []dto.CategoryDTO
	for i := range userCategories {
		category := dto.CategoryDTO{
			ID:         userCategories[i].ID,
			Name:       userCategories[i].Name,
			ParentID:   userCategories[i].ParentID,
			ParentName: userCategories[i].ParentName,
			IsIncome:   userCategories[i].IsIncome,
			UserID:     userCategories[i].UserID,
			IsDeleted:  userCategories[i].IsDeleted,
			CreatedAt:  userCategories[i].CreatedAt,
			UpdatedAt:  userCategories[i].UpdatedAt,
		}
		categories = append(categories, category)
	}

	logger.Debug("GetCategories request completed")
	return c.JSON(200, categories)
}

func GetGroupedCategories(c echo.Context) error {
	logger.Debug("GetGroupedCategories request started", "method", c.Request().Method, "url", c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	groupedCategories, err := sm.CategoriesService.GetUserCategoriesGrouped(user.ID)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	logger.Debug("GetGroupedCategories request completed")
	return c.JSON(200, groupedCategories)
}

func CreateCategory(c echo.Context) error {
	logger.Debug("CreateCategory request started", "method", c.Request().Method, "url", c.Request().URL)

	user, ok := c.Get("authenticated_user").(*models.User)
	if !ok || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var createDTO dto.CreateCategoryDTO
	if err := c.Bind(&createDTO); err != nil {
		logger.Error("Failed to bind create category DTO: ", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Basic validation
	if createDTO.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Category name is required")
	}

	createdCategory, err := sm.CategoriesService.CreateCategory(
		createDTO.Name,
		createDTO.IsIncome,
		createDTO.ParentID,
		user.ID,
	)
	if err != nil {
		logger.Error("Failed to create category: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create category")
	}

	// Convert to DTO format for response
	categoryDTO := dto.CategoryDTO{
		ID:         createdCategory.ID,
		Name:       createdCategory.Name,
		ParentID:   createdCategory.ParentID,
		ParentName: createdCategory.ParentName,
		IsIncome:   createdCategory.IsIncome,
		UserID:     createdCategory.UserID,
		IsDeleted:  createdCategory.IsDeleted,
		CreatedAt:  createdCategory.CreatedAt,
		UpdatedAt:  createdCategory.UpdatedAt,
	}

	logger.Debug("CreateCategory request completed")
	return c.JSON(http.StatusCreated, categoryDTO)
}
