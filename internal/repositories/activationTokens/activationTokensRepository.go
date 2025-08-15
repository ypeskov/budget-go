package activationTokens

import (
	"time"
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/models"
)

type RepositoryInterface interface {
	CreateActivationToken(token *models.ActivationToken) (*models.ActivationToken, error)
	GetActivationTokenByToken(token string) (*models.ActivationToken, error)
	DeleteToken(tokenID int) error
}

type RepositoryInstance struct {
	db *database.Database
}

func New(dbInstance *database.Database) RepositoryInterface {
	return &RepositoryInstance{
		db: dbInstance,
	}
}

func (r *RepositoryInstance) CreateActivationToken(token *models.ActivationToken) (*models.ActivationToken, error) {
	query := `
		INSERT INTO activation_tokens (user_id, token, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	token.CreatedAt = now
	token.UpdatedAt = now

	err := r.db.Db.QueryRow(query, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt, token.UpdatedAt).Scan(&token.ID)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *RepositoryInstance) GetActivationTokenByToken(tokenStr string) (*models.ActivationToken, error) {
	var token models.ActivationToken
	query := `SELECT id, user_id, token, expires_at, created_at, updated_at 
			  FROM activation_tokens 
			  WHERE token = $1 AND expires_at > NOW()`

	err := r.db.Db.Get(&token, query, tokenStr)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *RepositoryInstance) DeleteToken(tokenID int) error {
	query := `DELETE FROM activation_tokens WHERE id = $1`
	_, err := r.db.Db.Exec(query, tokenID)
	return err
}
