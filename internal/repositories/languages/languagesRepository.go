package languages

import (
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetLanguages() ([]models.Language, error)
}

type RepositoryInstance struct{}

var db *sqlx.DB

func NewLanguagesRepository(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (r *RepositoryInstance) GetLanguages() ([]models.Language, error) {
	const getLanguagesQuery = `SELECT id, name, code, is_deleted, created_at, updated_at FROM languages WHERE is_deleted = false ORDER BY name;`

	var languages []models.Language
	err := db.Select(&languages, getLanguagesQuery)
	if err != nil {
		logger.Error("Failed to get languages from database", "error", err)
		return nil, err
	}

	return languages, nil
}
