package services

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"sync"
	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/logger"
)

//go:embed templates/email/*.html
var emailTemplates embed.FS

type EmailTemplateRenderer interface {
	RenderBackupNotification(data *BackupTemplateData) (string, error)
	RenderExchangeRatesUpdate(data *ExchangeRatesTemplateData) (string, error)
	RenderActivationEmail(data *ActivationEmailTemplateData) (string, error)
}

type EmailTemplateRendererInstance struct {
	templates *template.Template
	cfg       *config.Config
}

var (
	emailTemplateInstance *EmailTemplateRendererInstance
	emailTemplateOnce     sync.Once
)

func NewEmailTemplateRenderer(cfg *config.Config) (EmailTemplateRenderer, error) {
	var err error
	emailTemplateOnce.Do(func() {
		logger.Debug("Creating EmailTemplateRenderer instance")
		templates, parseErr := template.ParseFS(emailTemplates, "templates/email/*.html")
		if parseErr != nil {
			err = fmt.Errorf("failed to parse email templates: %w", parseErr)
			return
		}
		emailTemplateInstance = &EmailTemplateRendererInstance{
			templates: templates,
			cfg:       cfg,
		}
	})

	if err != nil {
		return nil, err
	}

	return emailTemplateInstance, nil
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

type ActivationEmailTemplateData struct {
	Subject        string
	EnvName        string
	FirstName      string
	ActivationLink string
	AppName        string
}

func (r *EmailTemplateRendererInstance) RenderBackupNotification(data *BackupTemplateData) (string, error) {
	return r.renderTemplate("backup_notification.html", data)
}

func (r *EmailTemplateRendererInstance) RenderExchangeRatesUpdate(data *ExchangeRatesTemplateData) (string, error) {
	return r.renderTemplate("exchange_rates_notification.html", data)
}

func (r *EmailTemplateRendererInstance) RenderActivationEmail(data *ActivationEmailTemplateData) (string, error) {
	return r.renderTemplate("user_activation.html", data)
}

func (r *EmailTemplateRendererInstance) renderTemplate(templateName string, data interface{}) (string, error) {
	// Parse base template and the specific template
	tmpl, err := template.New("email").ParseFS(emailTemplates, "templates/email/base.html", "templates/email/"+templateName)
	if err != nil {
		logger.Error("Failed to parse templates", "templateName", templateName, "error", err)
		return "", fmt.Errorf("failed to parse templates for %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		logger.Error("Failed to execute template", "templateName", templateName, "error", err)
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}
	return buf.String(), nil
}
