package transactions

import (
	"ypeskov/budget-go/internal/models"

	log "github.com/sirupsen/logrus"
)

func (r *RepositoryInstance) CreateTransaction(transaction models.Transaction) error {
	query := `
		INSERT INTO transactions (
			user_id,
			account_id,
			amount,
			category_id,
			label,
			is_income,
			is_transfer,
			linked_transaction_id,
			base_currency_amount,
			notes,
			date_time,
			created_at,
			updated_at,
			is_deleted
		)
		VALUES (
			:user_id,
			:account_id,
			:amount,
			:category_id,
			:label,
			:is_income,
			:is_transfer,
			:linked_transaction_id,
			:base_currency_amount,
			:notes,
			:date_time,
			:created_at,
			:updated_at,
			:is_deleted
		)
	`
	_, err := r.db.NamedExec(query, transaction)
	if err != nil {
		return logAndReturnError(err, "Error creating transaction: ")
	}

	log.Debugf("Transaction created: %+v", transaction)

	return nil
}
