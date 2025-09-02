package database

import (
	"ypeskov/budget-go/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"ypeskov/budget-go/internal/config"
)

var DbInstance *Database

type Database struct {
	Db    *sqlx.DB
	//DbURL string
}

func New(cfg *config.Config) (*Database, error) {
	if DbInstance != nil {
		logger.Info("Returning existing database instance")
		return DbInstance, nil
	}

	// Build keyword DSN understood by pgx stdlib.
	// NOTE: keep sslmode configurable if needed.
	dsn := "host=" + cfg.DbHost +
		" port=" + cfg.DbPort +
		" user=" + cfg.DbUser +
		" password=" + cfg.DbPassword +
		" dbname=" + cfg.DbName +
		" sslmode=disable"

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// It's a good idea to Ping here to fail fast.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	logger.Info("Connected to database", "database", cfg.DbName)

	DbInstance = &Database{Db: db, }
	return DbInstance, nil
}

// Close gracefully closes the database connection if initialized.
func Close() error {
	if DbInstance != nil && DbInstance.Db != nil {
		return DbInstance.Db.Close()
	}
	return nil
}