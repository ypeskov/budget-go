package accounts

import (
	"database/sql"
	"errors"

	"ypeskov/budget-go/internal/dto"
	customErrors "ypeskov/budget-go/internal/errors"
	"ypeskov/budget-go/internal/models"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetUserAccounts(userId int, includeHidden bool, includeDeleted bool, archivedOnly bool) ([]dto.AccountDTO, error)
	GetAccountTypes() ([]models.AccountType, error)
	GetAccountById(id int) (models.Account, error)
	CreateAccount(account models.Account) (models.Account, error)
	UpdateAccount(account models.Account) (models.Account, error)
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
	archivedOnly bool) ([]dto.AccountDTO, error) {

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
	var accounts []dto.AccountDTO
	var err error
	if archivedOnly {
		getAccountsQuery += `WHERE a.user_id = $1 AND a.archived_at IS NOT NULL`
	} else {
		getAccountsQuery += `WHERE a.user_id = $1 AND a.archived_at IS NULL `
		if !includeHidden {
			getAccountsQuery += ` AND a.is_hidden = false`
		}
		if !includeDeleted {
			getAccountsQuery += ` AND a.is_deleted = false`
		}

	}
	getAccountsQuery += ` ORDER BY a.name`
	err = db.Select(&accounts, getAccountsQuery, userId)
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

func (a *RepositoryInstance) GetAccountById(id int) (models.Account, error) {
	const getAccountByIdQuery = `
	SELECT id, user_id, name, balance, credit_limit, opening_date, comment,
		currency_id, account_type_id, is_hidden, show_in_reports, is_deleted, archived_at, created_at, updated_at
	FROM accounts 
	WHERE id = $1
	`
	var account models.Account
	err := db.Get(&account, getAccountByIdQuery, id)
	if err != nil {
		log.Error("Error getting account by id: ", err)
		return models.Account{}, err
	}

	return account, nil
}

func (a *RepositoryInstance) CreateAccount(account models.Account) (models.Account, error) {
	const insertAccountQuery = `
INSERT INTO accounts (user_id, name, balance, account_type_id, currency_id, initial_balance, credit_limit, opening_date, comment, is_hidden, show_in_reports, is_deleted, archived_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING id, user_id, name, balance, account_type_id, currency_id, initial_balance, credit_limit, opening_date, comment, is_hidden, show_in_reports, is_deleted, archived_at, created_at, updated_at
`
	var newAccount models.Account
	err := db.Get(
		&newAccount,
		insertAccountQuery,
		account.UserID,         // $1
		account.Name,           // $2
		account.Balance,        // $3
		account.AccountTypeId,  // $4
		account.CurrencyId,     // $5
		account.InitialBalance, // $6
		account.CreditLimit,    // $7
		account.OpeningDate,    // $8
		account.Comment,        // $9
		account.IsHidden,       // $10
		account.ShowInReports,  // $11
		account.IsDeleted,      // $12
		account.ArchivedAt,     // $13
		account.CreatedAt,      // $14
		account.UpdatedAt,      // $15
	)
	if err != nil {
		return models.Account{}, err
	}

	return newAccount, nil
}

func (a *RepositoryInstance) UpdateAccount(account models.Account) (models.Account, error) {
	log.Debug("UpdateAccount Repository")
	const updateAccountQuery = `
UPDATE accounts
SET user_id = $1, name = $2, balance = $3, account_type_id = $4, currency_id = $5, 
    initial_balance = $6, credit_limit = $7, opening_date = $8, comment = $9,
    is_hidden = $10, show_in_reports = $11, is_deleted = $12, archived_at = $13, 
    updated_at = $14
WHERE id = $15
RETURNING id, user_id, name, balance, account_type_id, currency_id, initial_balance, credit_limit, opening_date, comment, is_hidden, show_in_reports, is_deleted, archived_at, created_at, updated_at
`
	var updatedAccount models.Account
	err := db.Get(
		&updatedAccount,
		updateAccountQuery,
		account.UserID,         // $1
		account.Name,           // $2
		account.Balance,        // $3
		account.AccountTypeId,  // $4
		account.CurrencyId,     // $5
		account.InitialBalance, // $6
		account.CreditLimit,    // $7
		account.OpeningDate,    // $8
		account.Comment,        // $9
		account.IsHidden,       // $10
		account.ShowInReports,  // $11
		account.IsDeleted,      // $12
		account.ArchivedAt,     // $13
		account.UpdatedAt,      // $14
		account.ID,             // $15
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Errorf("No account found with the provided ID: %v", account.ID)
			return models.Account{}, customErrors.ErrNoAccountFound
		}
		log.Error("Error updating account: ", err)
		return models.Account{}, err
	}

	return updatedAccount, nil
}
