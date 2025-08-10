package transactions

import (
	"strings"
	"time"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/models"

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
	GetTransactionDetail(transactionId int, userId int) (*dto.TransactionDetailRaw, error)
	UpdateTransaction(transaction models.Transaction) error
	GetTemplates(userId int) ([]dto.TemplateDTO, error)
	DeleteTemplates(templateIds []int, userId int) error
	CreateTransaction(transaction models.Transaction) error
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
			switch transactionType {
			case "income":
				typeFilters = append(typeFilters, "transactions.is_income = TRUE")
			case "expense":
				typeFilters = append(typeFilters, "transactions.is_income = FALSE")
			case "transfer":
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

func (r *RepositoryInstance) GetTransactionDetail(transactionId int, userId int) (*dto.TransactionDetailRaw, error) {
	params := map[string]interface{}{
		"transaction_id": transactionId,
		"user_id":        userId,
	}

	rows, err := r.db.NamedQuery(getTransactionDetailQuery, params)
	if err != nil {
		return nil, logAndReturnError(err, "Error executing transaction detail query: ")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var transaction dto.TransactionDetailRaw
	if err := rows.StructScan(&transaction); err != nil {
		return nil, logAndReturnError(err, "Error scanning transaction detail: ")
	}

	return &transaction, nil
}

func (r *RepositoryInstance) UpdateTransaction(transaction models.Transaction) error {
	params := map[string]interface{}{
		"id":          *transaction.ID,
		"account_id":  transaction.AccountID,
		"category_id": transaction.CategoryID,
		"amount":      transaction.Amount,
		"label":       transaction.Label,
		"notes":       transaction.Notes,
		"date_time":   transaction.DateTime,
		"is_income":   transaction.IsIncome,
		"is_transfer": transaction.IsTransfer,
		"updated_at":  transaction.UpdatedAt,
	}

	_, err := r.db.NamedExec(updateTransactionQuery, params)
	if err != nil {
		return logAndReturnError(err, "Error executing transaction update: ")
	}

	return nil
}

func logAndReturnError(err error, message string) error {
	log.Error(message, err)
	return err
}
