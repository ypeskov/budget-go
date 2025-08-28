package config

import (
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"ypeskov/budget-go/internal/logger"
)

type Config struct {
	// Server settings
	Port           string `env:"SERVER_PORT" envDefault:"8000"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	FrontendURL    string `env:"FRONTEND_URL" envDefault:"http://https://orgfin.run"`
	SendUserEmails bool   `env:"SEND_USER_EMAILS" envDefault:"false"`

	// Database connection settings
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

	// Database backup settings
	Environment string `env:"ENV" envDefault:"prod"`
	DBBackupDir string `env:"DB_BACKUP_DIR" envDefault:"./backups"`

	// Email settings for notifications
	SMTPHost         string `env:"SMTP_HOST" envDefault:"localhost"`
	SMTPPort         string `env:"SMTP_PORT" envDefault:"587"`
	SMTPUser         string `env:"SMTP_USER" envDefault:""`
	SMTPPassword     string `env:"SMTP_PASSWORD" envDefault:""`
	SMTPFrom         string `env:"SMTP_FROM" envDefault:"noreply@budget-app.com"`
	AdminEmailsRaw   string `env:"ADMINS_NOTIFICATION_EMAILS" envDefault:""`
	EmailFromAddress string `env:"EMAIL_FROM_ADDRESS" envDefault:"noreply@budget-app.com"`

	// CurrencyBeacon API settings
	CurrencyBeaconAPIURL     string `env:"CURRENCYBEACON_API_URL" envDefault:"https://api.currencybeacon.com"`
	CurrencyBeaconAPIKey     string `env:"CURRENCYBEACON_API_KEY" envDefault:""`
	CurrencyBeaconAPIVersion string `env:"CURRENCYBEACON_API_VERSION" envDefault:"v1"`

	// Container detection
	RunningInContainer bool `env:"RUNNING_IN_CONTAINER" envDefault:"false"`

	// Additional settings can be added here as needed
	AppName string `env:"APP_NAME" envDefault:"OrgFin.run"`
}

func New() *Config {
	cfg := &Config{}
	if err := godotenv.Load(".env"); err != nil {
		logger.Warn("No .env file found")
	}

	if err := env.Parse(cfg); err != nil {
		logger.Warn("Failed to parse env")
	}

	// Auto-detect container if not explicitly set
	if !cfg.RunningInContainer {
		cfg.RunningInContainer = isRunningInContainer()
	}

	return cfg
}

// isRunningInContainer detects if the application is running inside a container
func isRunningInContainer() bool {
	// Method 1: Check for /.dockerenv file (Docker containers)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Method 2: Check for Kubernetes service account (Kubernetes pods)
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return true
	}

	// Method 3: Check Kubernetes environment variables
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	// Method 4: Check RUNNING_IN_CONTAINER env var (manual override)
	if os.Getenv("RUNNING_IN_CONTAINER") == "true" {
		return true
	}

	return false
}
