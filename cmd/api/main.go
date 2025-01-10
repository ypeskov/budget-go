package main

import (
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.New()

	//log.SetReportCaller(true)

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Invalid log level in config: %s", cfg.LogLevel)
	}
	log.SetLevel(level)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.Debug("Starting server on port ", cfg.Port)

	serverInstance := server.New(cfg)
	err = serverInstance.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
