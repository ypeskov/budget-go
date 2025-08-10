package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/config"
)

var DbInstance *Database

type Database struct {
	Db    *sqlx.DB
	DbUrl string
}

func New(cfg *config.Config) (*Database, error) {
	if DbInstance != nil {
		log.Info("Returning existing database instance")
		return DbInstance, nil
	}

	DbInstance = &Database{
		DbUrl: "host=" + cfg.DbHost + " port=" + cfg.DbPort + " user=" + cfg.DbUser + " password=" +
			cfg.DbPassword + " dbname=" + cfg.DbName + " sslmode=disable",
	}

	db, err := sqlx.Connect("postgres", DbInstance.DbUrl)
	if err != nil {
		return nil, err
	}
	log.Infof("Connected to database %s\n", cfg.DbName)

	DbInstance.Db = db

	return DbInstance, nil
}

// Close gracefully closes the database connection if initialized
func Close() error {
    if DbInstance != nil && DbInstance.Db != nil {
        return DbInstance.Db.Close()
    }
    return nil
}
