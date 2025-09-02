package main

import (
	"log"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/logger"

	"ypeskov/budget-go/internal/database"

	"github.com/jmoiron/sqlx"
)

func main() {
	cfg := config.New()
	logger.Init(cfg.LogLevel)

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer func(Db *sqlx.DB) {
		err := Db.Close()
		if err != nil {

		}
	}(db.Db)

	log.Println("🧹 Cleaning DB...")

	// execute multiple statements in order
	tx, err := db.Db.Begin()
	if err != nil {
		log.Fatalf("failed to begin tx: %v", err)
	}

	if _, err := tx.Exec(`DROP SCHEMA public CASCADE`); err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to drop schema: %v", err)
	}
	if _, err := tx.Exec(`CREATE SCHEMA public`); err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to create schema: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit failed: %v", err)
	}

	log.Println("✅ DB fully wiped and public schema recreated")
}