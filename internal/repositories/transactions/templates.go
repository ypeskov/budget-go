package transactions

import (
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/logger"
)

func (r *RepositoryInstance) GetTemplates(userId int) ([]dto.TemplateDTO, error) {
	logger.Debug("GetTemplates Repository")
	query := `
		SELECT 
			tt.id,
			tt.label,
			tt.category_id,
			uc.id,
			uc.name
		FROM transaction_templates tt
		JOIN user_categories uc ON tt.category_id = uc.id
		WHERE uc.user_id = $1
	`

	rows, err := r.db.Queryx(query, userId)
	if err != nil {
		logger.Error("Error fetching templates", "error", err)
		return nil, err
	}
	defer rows.Close()

	templates := make([]dto.TemplateDTO, 0)
	for rows.Next() {
		var template dto.TemplateDTO
		err := rows.Scan(&template.ID, &template.Label, &template.CategoryID, &template.Category.ID, &template.Category.Name)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over templates rows", "error", err)
		return nil, err
	}

	return templates, nil
}

func (r *RepositoryInstance) DeleteTemplates(templateIds []int, userId int) error {
	logger.Debug("DeleteTemplates Repository")
	query := `
		DELETE FROM transaction_templates WHERE id = ANY($1) AND user_id = $2
	`

	_, err := r.db.Exec(query, templateIds, userId)
	if err != nil {
		logger.Error("Error deleting templates", "error", err)
		return err
	}

	return nil
}
