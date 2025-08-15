package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/models"

	log "github.com/sirupsen/logrus"
)

type EmailService struct {
	cfg              *config.Config
	templateRenderer *EmailTemplateRenderer
}

func NewEmailService(cfg *config.Config) (*EmailService, error) {
	templateRenderer, err := NewEmailTemplateRenderer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize email template renderer: %w", err)
	}

	return &EmailService{
		cfg:              cfg,
		templateRenderer: templateRenderer,
	}, nil
}

type EmailData struct {
	Subject        string
	Recipients     []string
	Body           string
	AttachmentPath string
}

func (s *EmailService) SendBackupNotification(backupResult *BackupResult) error {
	if s.cfg.AdminEmailsRaw == "" {
		log.Error("No admin emails configured for backup notification")
		return fmt.Errorf("no admin emails configured for notifications")
	}

	recipients := s.parseAdminEmails()
	if len(recipients) == 0 {
		log.Error("No valid admin emails found")
		return fmt.Errorf("no valid admin emails found")
	}

	body, err := s.templateRenderer.RenderBackupNotification(s.cfg.Environment, s.cfg.DbName, backupResult.Filename)
	if err != nil {
		log.Errorf("Failed to render backup email template: %v", err)
		return fmt.Errorf("failed to render backup email template: %w", err)
	}

	emailData := &EmailData{
		Subject:        "Database backup created",
		Recipients:     recipients,
		Body:           body,
		AttachmentPath: backupResult.FilePath,
	}

	return s.sendEmail(emailData)
}

func (s *EmailService) SendExchangeRatesUpdateNotification(exchangeRates *models.ExchangeRates) error {
	if s.cfg.AdminEmailsRaw == "" {
		log.Error("No admin emails configured for exchange rates notification")
		return fmt.Errorf("no admin emails configured for notifications")
	}

	recipients := s.parseAdminEmails()
	if len(recipients) == 0 {
		log.Error("No valid admin emails found")
		return fmt.Errorf("no valid admin emails found")
	}

	body, err := s.templateRenderer.RenderExchangeRatesNotification(s.cfg.Environment, exchangeRates)
	if err != nil {
		log.Errorf("Failed to render exchange rates email template: %v", err)
		return fmt.Errorf("failed to render exchange rates email template: %w", err)
	}

	emailData := &EmailData{
		Subject:    "Exchange rates updated",
		Recipients: recipients,
		Body:       body,
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

func (s *EmailService) sendEmail(emailData *EmailData) error {
	if s.cfg.SMTPHost == "" || s.cfg.SMTPUser == "" {
		log.Error("SMTP not configured, would send email:", emailData.Subject)
		log.Error("Email body:", emailData.Body)
		return fmt.Errorf("SMTP not configured")
	}

	// Extract from address once (optimization)
	fromAddr, err := s.extractEmailAddress()
	if err != nil {
		log.Errorf("Failed to extract from email address: %v", err)
		return fmt.Errorf("failed to extract from email address: %w", err)
	}

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// Build base email content once (optimization)
	baseEmailContent, err := s.buildBaseEmailContent(emailData)
	if err != nil {
		log.Errorf("Failed to build base email content: %v", err)
		return fmt.Errorf("failed to build base email content: %w", err)
	}

	// Send individual emails to maintain privacy
	for _, recipient := range emailData.Recipients {
		// Only the To: header differs per recipient
		msg := s.addRecipientHeader(baseEmailContent, recipient)

		err = smtp.SendMail(addr, auth, fromAddr, []string{recipient}, []byte(msg))
		if err != nil {
			log.Errorf("Failed to send email to %s: %v", recipient, err)
			return fmt.Errorf("failed to send email to %s: %w", recipient, err)
		}

		log.Infof("Email '%s' sent to: %s", emailData.Subject, recipient)
	}

	return nil
}

func (s *EmailService) buildBaseEmailContent(emailData *EmailData) (string, error) {
	boundary := "boundary123456789"
	var msg bytes.Buffer

	// Format the From address properly
	fromAddress, err := s.formatFromAddress()
	if err != nil {
		return "", fmt.Errorf("invalid from address: %w", err)
	}

	// Headers (without To: header - will be added per recipient)
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

func (s *EmailService) addRecipientHeader(baseContent, recipient string) string {
	// Add To: header at the beginning
	return fmt.Sprintf("To: %s\r\n%s", recipient, baseContent)
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

func (s *EmailService) SendActivationEmail(toEmail, firstName, activationToken string) error {
	log.Debug("Sending activation email to: ", toEmail)

	activationLink := fmt.Sprintf("%s/activate/%s", s.cfg.FrontendURL, activationToken)


	// For development, just log the activation link instead of sending actual email
	if s.cfg.SendUserEmails == false {
		log.Infof("ACTIVATION EMAIL: Hi %s, please activate your account: %s", firstName, activationLink)
		return nil
	}

	// Production email sending
	subject := "Activate Your Budget Account"
	body := fmt.Sprintf(`
<html>
<body>
<h2>Welcome to Budget!</h2>
<p>Hi %s,</p>
<p>Thank you for registering with Budget. Please activate your account by clicking the link below:</p>
<p><a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Activate Account</a></p>
<p>Or copy and paste this link into your browser:</p>
<p>%s</p>
<p>This link will expire in 24 hours.</p>
<br>
<p>Best regards,<br>The Budget Team</p>
</body>
</html>
`, firstName, activationLink, activationLink)
	emailData := &EmailData{
		Subject:    subject,
		Recipients: []string{toEmail},
		Body:       body,
	}

	return s.sendEmail(emailData)
}
