package user

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/models"
)

type RepositoryInterface interface {
	GetAllUsers() ([]*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
}

type RepositoryInstance struct {
	db *database.Database
}

func New(dbInstance *database.Database) RepositoryInterface {
	return &RepositoryInstance{
		db: dbInstance,
	}
}

func (u *RepositoryInstance) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	err := u.db.Db.Select(&users, "SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u *RepositoryInstance) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := u.db.Db.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *RepositoryInstance) CreateUser(user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, first_name, last_name, password_hash, is_active, base_currency_id, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	
	err := u.db.Db.QueryRow(query, user.Email, user.FirstName, user.LastName, user.PasswordHash, user.IsActive, user.BaseCurrencyID, user.IsDeleted).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}
