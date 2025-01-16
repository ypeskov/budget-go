package categories

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/services"
)

var (
	sm *services.Manager
)

func RegisterCategoriesRoutes(g *echo.Group, manager *services.Manager) {
	sm = manager

	g.GET("", GetCategories)
}

func GetCategories(c echo.Context) error {
	log.Debug("GetCategories Route")
	userRaw := c.Get("user")

	claims, ok := userRaw.(jwt.MapClaims)
	if !ok || claims == nil {
		log.Error("Failed to cast user to jwt.MapClaims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing user")
	}

	email, emailOk := claims["email"].(string)
	if !emailOk {
		log.Error("Email not found in claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}

	user, err := sm.UserService.GetUserByEmail(email)
	if err != nil {
		log.Error("Error getting user by email: ", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	userCategories, err := sm.CategoriesService.GetUserCategories(user.ID)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	var categories []dto.CategoryDTO
	for i := range userCategories {
		category := dto.CategoryDTO{
			ID:        userCategories[i].ID,
			Name:      userCategories[i].Name,
			ParentID:  userCategories[i].ParentID,
			IsIncome:  userCategories[i].IsIncome,
			UserID:    userCategories[i].UserID,
			IsDeleted: userCategories[i].IsDeleted,
			CreatedAt: userCategories[i].CreatedAt,
			UpdatedAt: userCategories[i].UpdatedAt,
		}
		categories = append(categories, category)
	}

	return c.JSON(200, categories)
}
