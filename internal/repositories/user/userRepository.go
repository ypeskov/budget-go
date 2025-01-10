package user

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/models"
)

type RepositoryInterface interface {
	GetAllUsers() ([]*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
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
