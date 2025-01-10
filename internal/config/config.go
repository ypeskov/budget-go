package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Port       string `env:"SERVER_PORT" envDefault:"8000"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info"`
	DbUser     string `env:"DB_USER" envDefault:"postgres"`
	DbPassword string `env:"DB_PASSWORD" envDefault:"password"`
	DbHost     string `env:"DB_HOST" envDefault:"localhost"`
	DbPort     string `env:"DB_PORT" envDefault:"5432"`
	DbName     string `env:"DB_NAME" envDefault:"budget"`
	SecretKey  string `env:"SECRET_KEY" envDefault:"SECRET_KEY"`
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
