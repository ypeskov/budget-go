package transactions

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetTransactionsWithAccounts(userId int) ([]TransactionWithAccount, error)
}

type RepositoryInstance struct{}

var db *sqlx.DB

func NewTransactionsRepository(dbInstance *sqlx.DB) Repository {
	db = dbInstance
	return &RepositoryInstance{}
}

func (r *RepositoryInstance) GetTransactionsWithAccounts(userId int) ([]TransactionWithAccount, error) {
	const getTransactionsQuery = `
	SELECT 
		transactions.id, transactions.user_id, transactions.account_id, transactions.category_id, 
		transactions.amount, transactions.new_balance, transactions.label, transactions.notes, transactions.date_time, 
		transactions.is_income, transactions.is_transfer, transactions.linked_transaction_id, 
		transactions.base_currency_amount, transactions.is_deleted, transactions.created_at, 
		transactions.updated_at, 
		accounts.id AS "accounts.id", accounts.name AS "accounts.name", accounts.balance AS "accounts.balance", 
		accounts.credit_limit AS "accounts.credit_limit", accounts.opening_date AS "accounts.opening_date", 
		accounts.comment AS "accounts.comment", accounts.currency_id AS "accounts.currency_id", 
		accounts.account_type_id AS "accounts.account_type_id", accounts.is_hidden AS "accounts.is_hidden", 
		accounts.show_in_reports AS "accounts.show_in_reports", accounts.is_deleted AS "accounts.is_deleted", 
		accounts.archived_at AS "accounts.archived_at", accounts.created_at AS "accounts.created_at", 
		accounts.updated_at AS "accounts.updated_at", 
		currencies.id AS "currencies.id", currencies.code AS "currencies.code", currencies.name AS "currencies.name", 
		account_types.id AS "account_types.id", account_types.type_name AS "account_types.type_name", 
		account_types.is_credit AS "account_types.is_credit", 
		user_categories.id AS "user_categories.id", user_categories.name AS "user_categories.name", 
		user_categories.is_deleted AS "user_categories.is_deleted", 
		user_categories.created_at AS "user_categories.created_at", user_categories.updated_at AS "user_categories.updated_at"
	FROM transactions 
	LEFT JOIN accounts ON transactions.account_id = accounts.id
	LEFT JOIN currencies ON accounts.currency_id = currencies.id
	LEFT JOIN account_types ON accounts.account_type_id = account_types.id
	LEFT JOIN user_categories ON transactions.category_id = user_categories.id
	WHERE transactions.user_id = $1
	ORDER BY transactions.date_time DESC
	LIMIT 10`

	var transactions []TransactionWithAccount
	err := db.Select(&transactions, getTransactionsQuery, userId)
	if err != nil {
		log.Error("Error getting transactions: ", err)
		return nil, err
	}
	for _, transaction := range transactions {
		log.Info(transaction.Account.Currency)
	}

	return transactions, nil
}
