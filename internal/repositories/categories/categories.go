package categories

import (
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetUserCategories(userId int) ([]models.UserCategory, error)
	CreateCategory(category models.UserCategory) (*models.UserCategory, error)
	ValidateCategoryOwnership(categoryId int, userId int) (bool, error)
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

func (r *RepositoryInstance) CreateCategory(category models.UserCategory) (*models.UserCategory, error) {
	const createCategoryQuery = `
		INSERT INTO user_categories (name, parent_id, is_income, user_id, is_deleted, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, name, parent_id, is_income, user_id, is_deleted, created_at, updated_at
	`

	var createdCategory models.UserCategory
	err := db.QueryRow(
		createCategoryQuery,
		category.Name,
		category.ParentID,
		category.IsIncome,
		category.UserID,
		category.IsDeleted,
	).Scan(
		&createdCategory.ID,
		&createdCategory.Name,
		&createdCategory.ParentID,
		&createdCategory.IsIncome,
		&createdCategory.UserID,
		&createdCategory.IsDeleted,
		&createdCategory.CreatedAt,
		&createdCategory.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// If it has a parent, get the parent name for the response
	if createdCategory.ParentID != nil {
		const getParentNameQuery = `SELECT name FROM user_categories WHERE id = $1 AND user_id = $2`
		var parentName string
		err := db.Get(&parentName, getParentNameQuery, *createdCategory.ParentID, category.UserID)
		if err == nil {
			createdCategory.ParentName = &parentName
		}
	}

	return &createdCategory, nil
}

func (r *RepositoryInstance) ValidateCategoryOwnership(categoryId int, userId int) (bool, error) {
	const validateOwnershipQuery = `
SELECT COUNT(*) FROM user_categories 
WHERE id = $1 AND user_id = $2 AND is_deleted = false
`
	var count int
	err := db.Get(&count, validateOwnershipQuery, categoryId, userId)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
