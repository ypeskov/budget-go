package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/server"
)


func main() {
	cfg := config.New()

	// Initialize logger
	logger.Init(cfg.LogLevel)

	logger.Info("Starting server on port", "port", cfg.Port)

	serverInstance := server.New(cfg)

	// Run server in a goroutine
	go func() {
		if err := serverInstance.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			logger.Fatal(err.Error())
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := serverInstance.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	if err := database.Close(); err != nil {
		logger.Error("Database close error", "error", err)
	}
}
