package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/server"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	// log.SetReportCaller(true)
}

func main() {
	cfg := config.New()

	// log.SetReportCaller(true)

	lvlStr := strings.TrimSpace(strings.ToLower(cfg.LogLevel))
	level, err := log.ParseLevel(lvlStr)
	if err != nil {
		log.Fatalf("Invalid log level in config: %s", cfg.LogLevel)
	}
	log.SetLevel(level)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.Debug("Starting server on port ", cfg.Port)

	serverInstance := server.New(cfg)

	// Run server in a goroutine
	go func() {
		if err := serverInstance.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := serverInstance.Shutdown(ctx); err != nil {
		log.Errorf("Server shutdown error: %v", err)
	}

	if err := database.Close(); err != nil {
		log.Errorf("Database close error: %v", err)
	}
}
