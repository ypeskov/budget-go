package transactions

import (
	"fmt"
	"ypeskov/budget-go/internal/models"

	log "github.com/sirupsen/logrus"
)

func (r *RepositoryInstance) CreateTransaction(transaction models.Transaction) (*models.Transaction, error) {
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
		RETURNING id, user_id, account_id, amount, category_id, label, is_income, is_transfer, 
				  linked_transaction_id, base_currency_amount, notes, date_time, created_at, updated_at, is_deleted
	`
	
	rows, err := r.db.NamedQuery(query, transaction)
	if err != nil {
		return nil, logAndReturnError(err, "Error creating transaction: ")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, logAndReturnError(fmt.Errorf("no rows returned"), "Error: no rows returned after creating transaction: ")
	}

	var createdTransaction models.Transaction
	err = rows.StructScan(&createdTransaction)
	if err != nil {
		return nil, logAndReturnError(err, "Error scanning created transaction: ")
	}

	log.Debugf("Transaction created: %+v", createdTransaction)

	return &createdTransaction, nil
}
