package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/activationTokens"

	log "github.com/sirupsen/logrus"
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
		log.Debug("Creating ActivationTokenService instance")
		activationTokenInstance = &ActivationTokenServiceInstance{
			tokenRepo:    tokenRepo,
			emailService: emailService,
		}
	})

	return activationTokenInstance
}

func (a *ActivationTokenServiceInstance) CreateActivationToken(userID int) (*models.ActivationToken, error) {
	log.Debug("Creating activation token for user ID: ", userID)

	// Generate a secure random token
	tokenStr, err := generateSecureToken()
	if err != nil {
		log.Error("Error generating secure token: ", err)
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
		log.Error("Error creating activation token: ", err)
		return nil, err
	}

	return createdToken, nil
}

func (a *ActivationTokenServiceInstance) ValidateAndUseToken(tokenStr string) (*models.ActivationToken, error) {
	log.Debug("Validating activation token")

	token, err := a.tokenRepo.GetActivationTokenByToken(tokenStr)
	if err != nil {
		log.Error("Error getting activation token: ", err)
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Delete token after use to prevent reuse
	err = a.tokenRepo.DeleteToken(token.ID)
	if err != nil {
		log.Error("Error deleting used token: ", err)
		return nil, err
	}

	return token, nil
}

func (a *ActivationTokenServiceInstance) SendActivationEmail(user *models.User, token string) error {
	log.Debug("Sending activation email to user: ", user.Email)

	err := a.emailService.SendActivationEmail(user.Email, user.FirstName, token)
	if err != nil {
		log.Error("Error sending activation email: ", err)
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
