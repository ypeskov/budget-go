package transactions

import (
	"strings"
	"time"
	"ypeskov/budget-go/internal/dto"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	GetTransactionsWithAccounts(userId int,
		perPage int,
		page int,
		accountIds []int,
		fromDate time.Time,
		toDate time.Time,
		transactionTypes []string,
		categoryIds []int,
	) ([]dto.TransactionWithAccount, error)
}

type RepositoryInstance struct {
	db *sqlx.DB
}

func NewTransactionsRepository(dbInstance *sqlx.DB) Repository {
	return &RepositoryInstance{
		db: dbInstance,
	}
}

func (r *RepositoryInstance) GetTransactionsWithAccounts(
	userId int,
	perPage int,
	page int,
	accountIds []int,
	fromDate time.Time,
	toDate time.Time,
	transactionTypes []string,
	categoryIds []int,
) ([]dto.TransactionWithAccount, error) {
	query := getTransactionsQuery
	params := map[string]interface{}{
		"user_id":  userId,
		"per_page": perPage,
		"offset":   (page - 1) * perPage,
	}
	filters := buildFilters(accountIds, fromDate, toDate, params, transactionTypes, categoryIds)
	if len(filters) > 0 {
		query += " AND " + filters
	}
	query += ` ORDER BY transactions.date_time DESC LIMIT :per_page OFFSET :offset`

	rows, err := r.db.NamedQuery(query, params)
	if err != nil {
		return nil, logAndReturnError(err, "Error executing query: ")
	}
	defer rows.Close()

	transactions, err := scanTransactions(rows)
	if err != nil {
		return nil, logAndReturnError(err, "Error scanning rows: ")
	}

	updateTransactionsWithAccountData(transactions)

	return transactions, nil
}

func updateTransactionsWithAccountData(transactions []dto.TransactionWithAccount) {
	for i, transaction := range transactions {
		transactions[i].Account.Currency = transaction.Currency
		transactions[i].Account.AccountType = transaction.AccountType
	}
}

func buildFilters(accountIds []int,
	fromDate, toDate time.Time,
	params map[string]interface{},
	transactionTypes []string,
	categoryIds []int,
) string {
	var filters []string

	if len(accountIds) > 0 {
		params["account_ids"] = pq.Array(accountIds)
		filters = append(filters, "transactions.account_id = ANY(:account_ids)")
	}

	if len(categoryIds) > 0 {
		params["category_ids"] = pq.Array(categoryIds)
		filters = append(filters, "transactions.category_id = ANY(:category_ids)")
	}

	if !fromDate.IsZero() {
		filters = append(filters, "transactions.date_time >= :from_date")
		params["from_date"] = fromDate.Format(time.DateOnly)
	}

	if !toDate.IsZero() {
		filters = append(filters, "transactions.date_time < :to_date")
		params["to_date"] = toDate.Add(24 * time.Hour).Format(time.DateOnly)
	}

	if len(transactionTypes) > 0 {
		var typeFilters []string
		for _, transactionType := range transactionTypes {
			if transactionType == "income" {
				typeFilters = append(typeFilters, "transactions.is_income = TRUE")
			} else if transactionType == "expense" {
				typeFilters = append(typeFilters, "transactions.is_income = FALSE")
			} else if transactionType == "transfer" {
				typeFilters = append(typeFilters, "transactions.is_transfer = TRUE")
			}
		}
		if len(typeFilters) > 0 {
			filters = append(filters, "("+strings.Join(typeFilters, " OR ")+")")
		}
	}

	if len(filters) == 0 {
		return ""
	}

	return strings.Join(filters, " AND ")
}

func scanTransactions(rows *sqlx.Rows) ([]dto.TransactionWithAccount, error) {
	var transactions []dto.TransactionWithAccount
	for rows.Next() {
		var transaction dto.TransactionWithAccount
		if err := rows.StructScan(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func logAndReturnError(err error, message string) error {
	log.Error(message, err)
	return err
}
