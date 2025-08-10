package categories

import (
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetUserCategories(userId int) ([]models.UserCategory, error)
}

type RepositoryInstance struct{}

var db *sqlx.DB

func NewCategoriesRepository(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (r *RepositoryInstance) GetUserCategories(userId int) ([]models.UserCategory, error) {
    const getUserCategoriesQuery = `
SELECT 
    c.id,
    c.name,
    c.parent_id,
    p.name AS parent_name,
    c.is_income,
    c.user_id,
    c.is_deleted,
    c.created_at,
    c.updated_at
FROM user_categories c
LEFT JOIN user_categories p ON p.id = c.parent_id AND p.user_id = c.user_id AND p.is_deleted = false
WHERE c.user_id = $1 AND c.is_deleted = false
ORDER BY LOWER(c.name) ASC;
`
	var categories []models.UserCategory
	err := db.Select(&categories, getUserCategoriesQuery, userId)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
