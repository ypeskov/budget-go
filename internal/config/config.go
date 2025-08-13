package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Port           string `env:"SERVER_PORT" envDefault:"8000"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	DbUser         string `env:"DB_USER" envDefault:"postgres"`
	DbPassword     string `env:"DB_PASSWORD" envDefault:"password"`
	DbHost         string `env:"DB_HOST" envDefault:"localhost"`
	DbPort         string `env:"DB_PORT" envDefault:"5432"`
	DbName         string `env:"DB_NAME" envDefault:"budget"`
	SecretKey      string `env:"SECRET_KEY" envDefault:"SECRET_KEY"`
	GoogleClientID string `env:"GOOGLE_CLIENT_ID" envDefault:""`

    // Background jobs (Asynq)
    RedisAddr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
    Timezone  string `env:"SCHEDULER_TIMEZONE" envDefault:"Europe/Sofia"`

    // Daily schedules (24h clock)
    ExchangeRatesHour   int `env:"DAILY_UPDATE_EXCHANGE_RATES_HOUR" envDefault:"3"`
    ExchangeRatesMinute int `env:"DAILY_UPDATE_EXCHANGE_RATES_MINUTE" envDefault:"0"`
    DBBackupHour        int `env:"DAILY_DB_BACKUP_HOUR" envDefault:"4"`
    DBBackupMinute      int `env:"DAILY_DB_BACKUP_MINUTE" envDefault:"0"`
    BudgetsProcHour     int `env:"DAILY_BUDGETS_PROCESSING_HOUR" envDefault:"2"`
    BudgetsProcMinute   int `env:"DAILY_BUDGETS_PROCESSING_MINUTE" envDefault:"0"`
}

func New() *Config {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found")
	}

	if err := env.Parse(cfg); err != nil {
		log.Warn("Failed to parse env")
	}

	return cfg
}
