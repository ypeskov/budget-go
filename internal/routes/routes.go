package routes

import (
	"net/http"
	"ypeskov/budget-go/internal/routes/categories"
	"ypeskov/budget-go/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"ypeskov/budget-go/internal/config"
	customMiddleware "ypeskov/budget-go/internal/middleware"
	"ypeskov/budget-go/internal/routes/accounts"
	"ypeskov/budget-go/internal/routes/auth"
)

func RegisterRoutes(cfg *config.Config, servicesManager *services.Manager) *echo.Echo {
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))

	e.GET("/health", Health)

	authRoutesGroup := e.Group("/auth")
	auth.RegisterAuthRoutes(authRoutesGroup, cfg, servicesManager)

	// Routes that require JWT
	protectedRoutes := e.Group("")
	protectedRoutes.Use(customMiddleware.CheckJWT)

	accountsRoutesGroup := protectedRoutes.Group("/accounts")
	accounts.RegisterAccountsRoutes(accountsRoutesGroup, cfg, servicesManager)

	categoriesRoutesGroup := protectedRoutes.Group("/categories")
	categories.RegisterCategoriesRoutes(categoriesRoutesGroup, cfg, servicesManager)

	return e
}

func Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
