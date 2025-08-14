package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/models"

	log "github.com/sirupsen/logrus"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

type EmailData struct {
	Subject        string
	Recipients     []string
	Body           string
	AttachmentPath string
}

func (s *EmailService) SendBackupNotification(backupResult *BackupResult) error {
	if s.cfg.AdminEmailsRaw == "" {
		log.Warn("No admin emails configured for backup notification")
		return nil
	}

	recipients := s.parseAdminEmails()
	if len(recipients) == 0 {
		log.Warn("No valid admin emails found")
		return nil
	}

	emailData := &EmailData{
		Subject:        "Database backup created",
		Recipients:     recipients,
		Body:           s.generateBackupEmailBody(backupResult),
		AttachmentPath: backupResult.FilePath,
	}

	return s.sendEmail(emailData)
}

func (s *EmailService) SendExchangeRatesUpdateNotification(exchangeRates *models.ExchangeRates) error {
	if s.cfg.AdminEmailsRaw == "" {
		log.Warn("No admin emails configured for exchange rates notification")
		return nil
	}

	recipients := s.parseAdminEmails()
	if len(recipients) == 0 {
		log.Warn("No valid admin emails found")
		return nil
	}

	emailData := &EmailData{
		Subject:    "Exchange rates updated",
		Recipients: recipients,
		Body:       s.generateExchangeRatesEmailBody(exchangeRates),
	}

	return s.sendEmail(emailData)
}

func (s *EmailService) parseAdminEmails() []string {
	if s.cfg.AdminEmailsRaw == "" {
		return []string{}
	}

	emails := strings.Split(s.cfg.AdminEmailsRaw, ",")
	var validEmails []string

	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email != "" && strings.Contains(email, "@") {
			validEmails = append(validEmails, email)
		}
	}

	return validEmails
}

func (s *EmailService) generateBackupEmailBody(backupResult *BackupResult) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Database Backup Created</title>
</head>
<body>
    <h2>Database Backup Successfully Created</h2>
    <p>A new database backup has been created for your Budget Go application.</p>
    
    <h3>Backup Details:</h3>
    <ul>
        <li><strong>Environment:</strong> {{.EnvName}}</li>
        <li><strong>Database:</strong> {{.DBName}}</li>
        <li><strong>Backup File:</strong> {{.Filename}}</li>
    </ul>
    
    <p>The backup file is attached to this email.</p>
    
    <hr>
    <p><small>This is an automated message from your Budget Go application.</small></p>
</body>
</html>`

	data := struct {
		EnvName  string
		DBName   string
		Filename string
	}{
		EnvName:  s.cfg.Environment,
		DBName:   s.cfg.DbName,
		Filename: backupResult.Filename,
	}

	t, err := template.New("backup").Parse(tmpl)
	if err != nil {
		log.Errorf("Failed to parse email template: %v", err)
		return fmt.Sprintf("Database backup created: %s for environment %s", backupResult.Filename, s.cfg.Environment)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		log.Errorf("Failed to execute email template: %v", err)
		return fmt.Sprintf("Database backup created: %s for environment %s", backupResult.Filename, s.cfg.Environment)
	}

	return buf.String()
}

func (s *EmailService) generateExchangeRatesEmailBody(exchangeRates *models.ExchangeRates) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Exchange Rates Updated</title>
</head>
<body>
    <h2>Exchange Rates Successfully Updated</h2>
    <p>The exchange rates have been successfully updated for your Budget Go application.</p>
    
    <h3>Update Details:</h3>
    <ul>
        <li><strong>Environment:</strong> {{.EnvName}}</li>
        <li><strong>Date:</strong> {{.UpdatedAt}}</li>
        <li><strong>Actual Date:</strong> {{.ActualDate}}</li>
        <li><strong>Base Currency:</strong> {{.BaseCurrency}}</li>
        <li><strong>Service:</strong> {{.ServiceName}}</li>
        <li><strong>Rate Count:</strong> {{.RateCount}}</li>
    </ul>
    
    <hr>
    <p><small>This is an automated message from your Budget Go application.</small></p>
</body>
</html>`

	data := struct {
		EnvName      string
		UpdatedAt    string
		ActualDate   string
		BaseCurrency string
		ServiceName  string
		RateCount    int
	}{
		EnvName:      s.cfg.Environment,
		UpdatedAt:    time.Now().Format("2006-01-02 15:04:05 MST"),
		ActualDate:   exchangeRates.ActualDate.Format("2006-01-02"),
		BaseCurrency: exchangeRates.BaseCurrencyCode,
		ServiceName:  exchangeRates.ServiceName,
		RateCount:    len(exchangeRates.Rates),
	}

	t, err := template.New("exchange-rates").Parse(tmpl)
	if err != nil {
		log.Errorf("Failed to parse email template: %v", err)
		return fmt.Sprintf("Exchange rates updated at %s for environment %s", data.UpdatedAt, s.cfg.Environment)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		log.Errorf("Failed to execute email template: %v", err)
		return fmt.Sprintf("Exchange rates updated at %s for environment %s", data.UpdatedAt, s.cfg.Environment)
	}

	return buf.String()
}

func (s *EmailService) sendEmail(emailData *EmailData) error {
	if s.cfg.SMTPHost == "" || s.cfg.SMTPUser == "" {
		log.Info("SMTP not configured, would send email:", emailData.Subject)
		log.Info("Email body:", emailData.Body)
		if emailData.AttachmentPath != "" {
			log.Info("Would attach file:", emailData.AttachmentPath)
		}
		return nil
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	for _, recipient := range emailData.Recipients {
		msg, err := s.buildMimeEmailWithAttachment(recipient, emailData)
		if err != nil {
			log.Errorf("Failed to build email for %s: %v", recipient, err)
			return fmt.Errorf("failed to build email for %s: %w", recipient, err)
		}

		// Extract just the email address for SMTP envelope
		fromAddr, err := s.extractEmailAddress()
		if err != nil {
			log.Errorf("Failed to extract from email address: %v", err)
			return fmt.Errorf("failed to extract from email address: %w", err)
		}

		addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
		err = smtp.SendMail(addr, auth, fromAddr, []string{recipient}, []byte(msg))
		if err != nil {
			log.Errorf("Failed to send email to %s: %v", recipient, err)
			return fmt.Errorf("failed to send email to %s: %w", recipient, err)
		}

		log.Infof("Email '%s' sent to: %s", emailData.Subject, recipient)
	}

	return nil
}

func (s *EmailService) buildMimeEmailWithAttachment(recipient string, emailData *EmailData) (string, error) {
	boundary := "boundary123456789"

	var msg bytes.Buffer

	// Format the From address properly
	fromAddress, err := s.formatFromAddress()
	if err != nil {
		return "", fmt.Errorf("invalid from address: %w", err)
	}

	// Headers
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromAddress))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", emailData.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")

	// HTML body part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(emailData.Body)
	msg.WriteString("\r\n")

	// Attachment part
	if emailData.AttachmentPath != "" {
		filename := filepath.Base(emailData.AttachmentPath)

		// Read file
		fileData, err := os.ReadFile(emailData.AttachmentPath)
		if err != nil {
			return "", fmt.Errorf("failed to read attachment file: %w", err)
		}

		// Encode file to base64
		encodedFile := base64.StdEncoding.EncodeToString(fileData)

		// Get MIME type
		mimeType := mime.TypeByExtension(filepath.Ext(filename))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", mimeType, filename))
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
		msg.WriteString("\r\n")

		// Write base64 content in chunks of 76 characters
		for i := 0; i < len(encodedFile); i += 76 {
			end := i + 76
			if end > len(encodedFile) {
				end = len(encodedFile)
			}
			msg.WriteString(encodedFile[i:end])
			msg.WriteString("\r\n")
		}
	}

	// End boundary
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return msg.String(), nil
}

func (s *EmailService) formatFromAddress() (string, error) {
	// Parse the from address to ensure it's properly formatted
	addr, err := mail.ParseAddress(s.cfg.EmailFromAddress)
	if err != nil {
		// If parsing fails, it might be a simple email without display name
		// Try to use it as-is if it looks like a valid email
		if strings.Contains(s.cfg.EmailFromAddress, "@") && !strings.Contains(s.cfg.EmailFromAddress, "<") {
			return s.cfg.EmailFromAddress, nil
		}
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	// Return the properly formatted address
	return addr.String(), nil
}

func (s *EmailService) extractEmailAddress() (string, error) {
	// Parse the from address and extract just the email part for SMTP envelope
	addr, err := mail.ParseAddress(s.cfg.EmailFromAddress)
	if err != nil {
		// If parsing fails, it might be a simple email without display name
		if strings.Contains(s.cfg.EmailFromAddress, "@") && !strings.Contains(s.cfg.EmailFromAddress, "<") {
			return s.cfg.EmailFromAddress, nil
		}
		return "", fmt.Errorf("invalid email format: %w", err)
	}
	
	// Return just the email address without display name
	return addr.Address, nil
}
