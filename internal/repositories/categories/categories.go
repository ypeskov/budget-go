package categories

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/models"
)

type Repository interface {
	GetUserCategories(userId int) ([]models.UserCategory, error)
}

type RepositoryInstance struct {
	db *database.Database
}

func NewCategoriesRepository(dbInstance *database.Database) Repository {
	return &RepositoryInstance{
		db: dbInstance,
	}
}

func (r *RepositoryInstance) GetUserCategories(userId int) ([]models.UserCategory, error) {
	const getUserCategoriesQuery = `
SELECT 
	c.id, c.name, c.parent_id, c.is_income, c.user_id, c.is_deleted, c.created_at, c.updated_at
FROM user_categories c
WHERE c.user_id = $1 AND c.is_deleted = false;
`
	var categories []models.UserCategory
	err := r.db.Db.Select(&categories, getUserCategoriesQuery, userId)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
