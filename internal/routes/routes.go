package routes

import (
	"net/http"
	"ypeskov/budget-go/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"ypeskov/budget-go/internal/config"
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

	return e
}

func Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
