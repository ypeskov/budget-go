package transactions

var getTransactionsQuery = `
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

	COALESCE(user_categories.id, NULL) AS "user_categories.id",COALESCE(user_categories.name, '') AS "user_categories.name", 
	COALESCE(user_categories.is_deleted, FALSE) AS "user_categories.is_deleted", 
	COALESCE(user_categories.created_at, NULL) AS "user_categories.created_at", 
	COALESCE(user_categories.updated_at, NULL) AS "user_categories.updated_at"
FROM transactions 
LEFT JOIN accounts ON transactions.account_id = accounts.id
LEFT JOIN currencies ON accounts.currency_id = currencies.id
LEFT JOIN account_types ON accounts.account_type_id = account_types.id
LEFT JOIN user_categories ON transactions.category_id = user_categories.id

WHERE transactions.user_id = :user_id
`
