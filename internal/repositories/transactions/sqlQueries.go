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
AND transactions.is_deleted = FALSE
`

var getTransactionDetailQuery = `
SELECT 
	transactions.id, transactions.user_id, transactions.account_id, transactions.category_id, 
	transactions.amount, transactions.new_balance, transactions.label, transactions.notes, transactions.date_time, 
	transactions.is_income, transactions.is_transfer, transactions.linked_transaction_id, 
	transactions.base_currency_amount, transactions.is_deleted, transactions.created_at, 
	transactions.updated_at,

	users.id AS "users.id", users.email AS "users.email", users.first_name AS "users.first_name", 
	users.last_name AS "users.last_name",

	accounts.id AS "accounts.id", accounts.user_id AS "accounts.user_id", accounts.name AS "accounts.name", 
	accounts.balance AS "accounts.balance", accounts.initial_balance AS "accounts.initial_balance",
	accounts.credit_limit AS "accounts.credit_limit", accounts.opening_date AS "accounts.opening_date", 
	accounts.comment AS "accounts.comment", accounts.currency_id AS "accounts.currency_id", 
	accounts.account_type_id AS "accounts.account_type_id", accounts.is_hidden AS "accounts.is_hidden", 
	accounts.show_in_reports AS "accounts.show_in_reports", accounts.is_deleted AS "accounts.is_deleted", 
	accounts.archived_at AS "accounts.archived_at", accounts.created_at AS "accounts.created_at", 
	accounts.updated_at AS "accounts.updated_at", 

	currencies.id AS "currencies.id", currencies.code AS "currencies.code", currencies.name AS "currencies.name", 
	account_types.id AS "account_types.id", account_types.type_name AS "account_types.type_name", 
	account_types.is_credit AS "account_types.is_credit", 

	user_categories.id AS "user_categories.id",
	user_categories.name AS "user_categories.name", 
	user_categories.parent_id AS "user_categories.parent_id",
	user_categories.is_income AS "user_categories.is_income",
	user_categories.user_id AS "user_categories.user_id",
	user_categories.is_deleted AS "user_categories.is_deleted", 
	user_categories.created_at AS "user_categories.created_at", 
	user_categories.updated_at AS "user_categories.updated_at",

	linked_transactions.id AS "linked_transactions.id",
	linked_transactions.user_id AS "linked_transactions.user_id",
	linked_transactions.account_id AS "linked_transactions.account_id",
	linked_transactions.category_id AS "linked_transactions.category_id",
	linked_transactions.amount AS "linked_transactions.amount",
	linked_transactions.new_balance AS "linked_transactions.new_balance",
	linked_transactions.label AS "linked_transactions.label",
	linked_transactions.notes AS "linked_transactions.notes",
	linked_transactions.date_time AS "linked_transactions.date_time",
	linked_transactions.is_income AS "linked_transactions.is_income",
	linked_transactions.is_transfer AS "linked_transactions.is_transfer",
	linked_transactions.linked_transaction_id AS "linked_transactions.linked_transaction_id",
	linked_transactions.base_currency_amount AS "linked_transactions.base_currency_amount",
	linked_transactions.is_deleted AS "linked_transactions.is_deleted",
	linked_transactions.created_at AS "linked_transactions.created_at",
	linked_transactions.updated_at AS "linked_transactions.updated_at"
FROM transactions 
LEFT JOIN users ON transactions.user_id = users.id
LEFT JOIN accounts ON transactions.account_id = accounts.id
LEFT JOIN currencies ON accounts.currency_id = currencies.id
LEFT JOIN account_types ON accounts.account_type_id = account_types.id
LEFT JOIN user_categories ON transactions.category_id = user_categories.id
LEFT JOIN transactions AS linked_transactions ON transactions.linked_transaction_id = linked_transactions.id

WHERE transactions.id = :transaction_id AND transactions.user_id = :user_id
`

var updateTransactionQuery = `
UPDATE transactions 
SET account_id = :account_id,
    category_id = :category_id,
    amount = :amount,
    new_balance = :new_balance,
    label = :label,
    notes = :notes,
    date_time = :date_time,
    is_income = :is_income,
    is_transfer = :is_transfer,
    linked_transaction_id = :linked_transaction_id,
    updated_at = :updated_at
WHERE id = :id
`

var deleteTransactionQuery = `
UPDATE transactions 
SET is_deleted = TRUE,
    updated_at = :updated_at
WHERE id = :id AND user_id = :user_id AND is_deleted = FALSE
`
