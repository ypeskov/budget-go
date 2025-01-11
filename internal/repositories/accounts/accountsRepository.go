package accounts

import (
	"ypeskov/budget-go/internal/database"
	"ypeskov/budget-go/internal/models"
)

type Repository interface {
	GetUserAccounts(userId int) ([]models.Account, error)
}

type RepositoryInstance struct {
	db *database.Database
}

func NewAccountsService(dbInstance *database.Database) Repository {
	return &RepositoryInstance{
		db: dbInstance,
	}
}

func (a *RepositoryInstance) GetUserAccounts(userId int) ([]models.Account, error) {
	const getAccountsQuery = `
SELECT 
    a.id, a.user_id, a.name, a.balance, a.credit_limit, a.opening_date, a.comment,
    a.currency_id, a.account_type_id,
    a.is_hidden, a.show_in_reports, a.is_deleted, a.archived_at, a.created_at, a.updated_at,
    c.id AS "currency.id", c.code AS "currency.code", c.name AS "currency.name", 
    c.created_at AS "currency.created_at", c.updated_at AS "currency.updated_at",
    at.id AS "account_type.id", at.type_name AS "account_type.type_name"
FROM accounts a
JOIN currencies c ON a.currency_id = c.id
JOIN account_types at ON a.account_type_id = at.id
WHERE a.user_id = $1 AND a.is_deleted = false;
`
	var accounts []models.Account
	err := a.db.Db.Select(&accounts, getAccountsQuery, userId)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}
