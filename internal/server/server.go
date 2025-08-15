package server

import (
	"fmt"
	"net/http"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/routes"
	"ypeskov/budget-go/internal/services"
)

func New(cfg *config.Config) *http.Server {
	db, err := database.New(cfg)
	if err != nil {
		panic(err)
	}

	servicesManager, err := services.NewServicesManager(db, cfg)
	if err != nil {
		panic(err)
	}

	e := routes.RegisterRoutes(cfg, servicesManager)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: e,
	}
}
