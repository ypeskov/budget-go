package accounts

import (
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetUserAccounts(userId int, includeHidden bool, includeDeleted bool, archivedOnly bool) ([]models.Account, error)
	GetAccountTypes() ([]models.AccountType, error)
}

type RepositoryInstance struct{}

var db *sqlx.DB

func NewAccountsService(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (a *RepositoryInstance) GetUserAccounts(
	userId int,
	includeHidden bool,
	includeDeleted bool,
	archivedOnly bool) ([]models.Account, error) {

	log.Debug("GetUserAccounts repository")
	var getAccountsQuery = `
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
`
	var accounts []models.Account
	var err error
	if archivedOnly {
		getAccountsQuery += `WHERE a.user_id = $1 AND a.archived_at IS NOT NULL`
		err = db.Select(&accounts, getAccountsQuery, userId)
	} else {
		getAccountsQuery += `WHERE a.user_id = $1 AND a.archived_at IS NULL`
		if !includeHidden {
			getAccountsQuery += ` AND a.is_hidden = false`
		}
		if !includeDeleted {
			getAccountsQuery += ` AND a.is_deleted = false`
		}
		err = db.Select(&accounts, getAccountsQuery, userId)
	}
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (a *RepositoryInstance) GetAccountTypes() ([]models.AccountType, error) {
	const getAccountTypesQuery = `
SELECT 
	id, type_name, is_credit, is_deleted, created_at, updated_at
FROM account_types
WHERE is_deleted = false;
`
	var accountTypes []models.AccountType
	err := db.Select(&accountTypes, getAccountTypesQuery)
	if err != nil {
		return nil, err
	}

	return accountTypes, nil
}
