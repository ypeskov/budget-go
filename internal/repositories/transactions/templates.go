package transactions

import (
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/dto"
)

func (r *RepositoryInstance) GetTemplates(userId int) ([]dto.TemplateDTO, error) {
	log.Debug("GetTemplates Repository")
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
		log.Error("Error fetching templates: ", err)
		return nil, err
	}

	var templates []dto.TemplateDTO
	for rows.Next() {
		var template dto.TemplateDTO
		err := rows.Scan(&template.ID, &template.Label, &template.CategoryID, &template.Category.ID, &template.Category.Name)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, nil
}

func (r *RepositoryInstance) DeleteTemplates(templateIds []int, userId int) error {
	log.Debug("DeleteTemplates Repository")
	query := `
		DELETE FROM transaction_templates WHERE id = ANY($1) AND user_id = $2
	`

	_, err := r.db.Exec(query, pq.Array(templateIds), userId)
	if err != nil {
		log.Error("Error deleting templates: ", err)
		return err
	}

	return nil
}
