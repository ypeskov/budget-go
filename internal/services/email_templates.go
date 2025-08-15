package services

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"
	"ypeskov/budget-go/internal/config"

	"ypeskov/budget-go/internal/models"

	log "github.com/sirupsen/logrus"
)

//go:embed templates/email/*.html
var emailTemplates embed.FS

type EmailTemplateRenderer struct {
	templates *template.Template
	cfg       *config.Config
}

func NewEmailTemplateRenderer(cfg *config.Config) (*EmailTemplateRenderer, error) {
	templates, err := template.ParseFS(emailTemplates, "templates/email/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	return &EmailTemplateRenderer{
		templates: templates,
		cfg:       cfg,
	}, nil
}

type BackupTemplateData struct {
	Subject   string
	EnvName   string
	DBName    string
	Filename  string
	CreatedAt string
	AppName   string
}

type ExchangeRatesTemplateData struct {
	Subject      string
	EnvName      string
	UpdatedAt    string
	ActualDate   string
	BaseCurrency string
	ServiceName  string
	RateCount    int
	AppName      string
}

type UserActivationTemplateData struct {
	Subject        string
	EnvName        string
	FirstName      string
	ActivationLink string
	AppName        string
}

func (r *EmailTemplateRenderer) RenderBackupNotification(envName, dbName, filename string) (string, error) {
	data := BackupTemplateData{
		Subject:   "Database Backup Created",
		EnvName:   envName,
		DBName:    dbName,
		Filename:  filename,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05 MST"),
		AppName:   r.cfg.AppName,
	}

	return r.renderTemplate("backup_notification.html", data)
}

func (r *EmailTemplateRenderer) RenderExchangeRatesNotification(envName string, exchangeRates *models.ExchangeRates) (string, error) {
	data := ExchangeRatesTemplateData{
		Subject:      "Exchange Rates Updated",
		EnvName:      envName,
		UpdatedAt:    time.Now().Format("2006-01-02 15:04:05 MST"),
		ActualDate:   exchangeRates.ActualDate.Format("2006-01-02"),
		BaseCurrency: exchangeRates.BaseCurrencyCode,
		ServiceName:  exchangeRates.ServiceName,
		RateCount:    len(exchangeRates.Rates),
		AppName:      r.cfg.AppName,
	}

	return r.renderTemplate("exchange_rates_notification.html", data)
}

func (r *EmailTemplateRenderer) RenderUserActivation(firstName, activationToken string) (string, error) {
	activationLink := fmt.Sprintf("%s/activate/%s", r.cfg.FrontendURL, activationToken)
	data := UserActivationTemplateData{
		Subject:        fmt.Sprintf("Activate Your %s Account", r.cfg.AppName),
		EnvName:        r.cfg.Environment,
		FirstName:      firstName,
		ActivationLink: activationLink,
		AppName:        r.cfg.AppName,
	}

	return r.renderTemplate("user_activation.html", data)
}

func (r *EmailTemplateRenderer) renderTemplate(templateName string, data interface{}) (string, error) {
	// Parse base template and the specific template
	tmpl, err := template.New("email").ParseFS(emailTemplates, "templates/email/base.html", "templates/email/"+templateName)
	if err != nil {
		log.Errorf("Failed to parse templates for %s: %v", templateName, err)
		return "", fmt.Errorf("failed to parse templates for %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		log.Errorf("Failed to execute template %s: %v", templateName, err)
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}
	return buf.String(), nil
}
