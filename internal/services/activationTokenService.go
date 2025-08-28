package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/activationTokens"
)

type ActivationTokenService interface {
	CreateActivationToken(userID int) (*models.ActivationToken, error)
	ValidateAndUseToken(tokenStr string) (*models.ActivationToken, error)
	SendActivationEmail(user *models.User, token string) error
}

type ActivationTokenServiceInstance struct {
	tokenRepo    activationTokens.RepositoryInterface
	emailService EmailService
}

var (
	activationTokenInstance *ActivationTokenServiceInstance
	activationTokenOnce     sync.Once
)

func NewActivationTokenService(tokenRepo activationTokens.RepositoryInterface, emailService EmailService) ActivationTokenService {
	activationTokenOnce.Do(func() {
		logger.Debug("Creating ActivationTokenService instance")
		activationTokenInstance = &ActivationTokenServiceInstance{
			tokenRepo:    tokenRepo,
			emailService: emailService,
		}
	})

	return activationTokenInstance
}

func (a *ActivationTokenServiceInstance) CreateActivationToken(userID int) (*models.ActivationToken, error) {
	logger.Debug("Creating activation token for user ID", "userID", userID)

	// Generate a secure random token
	tokenStr, err := generateSecureToken()
	if err != nil {
		logger.Error("Error generating secure token", "error", err)
		return nil, err
	}

	// Set expiration to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)

	token := &models.ActivationToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}

	createdToken, err := a.tokenRepo.CreateActivationToken(token)
	if err != nil {
		logger.Error("Error creating activation token", "error", err)
		return nil, err
	}

	return createdToken, nil
}

func (a *ActivationTokenServiceInstance) ValidateAndUseToken(tokenStr string) (*models.ActivationToken, error) {
	logger.Debug("Validating activation token")

	token, err := a.tokenRepo.GetActivationTokenByToken(tokenStr)
	if err != nil {
		logger.Error("Error getting activation token", "error", err)
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Delete token after use to prevent reuse
	err = a.tokenRepo.DeleteToken(token.ID)
	if err != nil {
		logger.Error("Error deleting used token", "error", err)
		return nil, err
	}

	return token, nil
}

func (a *ActivationTokenServiceInstance) SendActivationEmail(user *models.User, token string) error {
	logger.Debug("Sending activation email to user", "email", user.Email)

	err := a.emailService.SendActivationEmail(user.Email, user.FirstName, token)
	if err != nil {
		logger.Error("Error sending activation email", "error", err)
		return err
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
// limited to 32 characters to match database constraint
func generateSecureToken() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
