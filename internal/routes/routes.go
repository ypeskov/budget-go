package routes

import (
	log "github.com/sirupsen/logrus"
	"net/http"
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
	// authRoutesGroup.Use(echojwt.WithConfig(jwtConfig))
	auth.RegisterAuthRoutes(authRoutesGroup, cfg, servicesManager)

	// Routes that require JWT
	protectedRoutes := e.Group("")
	protectedRoutes.Use(customMiddleware.CheckJWT)

	accountsRoutesGroup := protectedRoutes.Group("/accounts")
	accounts.RegisterAccountsRoutes(accountsRoutesGroup, cfg, servicesManager)

	//e.Use(customMiddleware.CheckJWT)
	e.GET("/test", func(c echo.Context) error {
		log.Info(c)
		return c.String(http.StatusOK, "Test")
	})

	return e
}

func Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
