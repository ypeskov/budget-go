package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/routes/routeErrors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
)

// =============================================================================
// UTILITY FUNCTIONS (used by multiple DTOs)
// =============================================================================

// parseCategoryIDFromInterface parses CategoryID from any (can be string or int)
func parseCategoryIdFromInterface(value any) (*int, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return nil, nil
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Error parsing categoryId:", err)
			return nil, err
		}
		return &id, nil
	case int:
		id := int(v)
		return &id, nil
	case float64:
		id := int(v)
		return &id, nil
	default:
		return nil, fmt.Errorf("invalid categoryId type: %T", v)
	}
}

// parseAmountFromInterface parses Amount from any (can be string or float64)
func parseAmountFromInterface(value any) (decimal.Decimal, error) {
	if value == nil {
		return decimal.Zero, nil
	}

	switch v := value.(type) {
	case string:
		amount, err := decimal.NewFromString(v)
		if err != nil {
			log.Error("Error parsing amount:", err)
			return decimal.Zero, err
		}
		return amount, nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case int:
		return decimal.NewFromInt(int64(v)), nil
	default:
		return decimal.Zero, fmt.Errorf("invalid amount type: %T", v)
	}
}

// parseIDFromInterface parses ID from interface{} (can be string or int)
func parseIDFromInterface(value interface{}) (int, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return 0, nil
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Error parsing id:", err)
			return 0, err
		}
		return id, nil
	case int:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid id type: %T", v)
	}
}

// =============================================================================
// TRANSACTION FILTERS DTO AND RELATED FUNCTIONS
// =============================================================================

// TransactionFilters represents filtering parameters for transaction queries
type TransactionFilters struct {
	PerPage          int
	Page             int
	Currencies       []string
	FromDate         time.Time
	ToDate           time.Time
	AccountIds       []int
	TransactionTypes []string
	CategoryIds      []int
}

// ParseTransactionFilters extracts and validates query parameters for transaction filtering
func ParseTransactionFilters(c echo.Context) (*TransactionFilters, error) {
	perPage, err := getPerPage(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	page, err := getPage(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	currencies, err := getCurrencies(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	fromDate, err := getFromDate(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	toDate, err := getToDate(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	accountIds, err := getAccountIds(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	transactionTypes, err := getTransactionTypes(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	categoryIds, err := getCategoryIds(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}

	return &TransactionFilters{
		PerPage:          perPage,
		Page:             page,
		Currencies:       currencies,
		FromDate:         fromDate,
		ToDate:           toDate,
		AccountIds:       accountIds,
		TransactionTypes: transactionTypes,
		CategoryIds:      categoryIds,
	}, nil
}

// Helper functions for parsing query parameters
func getPerPage(c echo.Context) (int, error) {
	perPage, err := getQueryParamAsInt(c, "per_page", 10)
	if err != nil {
		return 0, err
	}
	if perPage < 1 {
		return 0, errors.New("per_page must be greater than 0")
	}
	return perPage, nil
}

func getPage(c echo.Context) (int, error) {
	return getQueryParamAsInt(c, "page", 20)
}

func getCurrencies(c echo.Context) ([]string, error) {
	return getQueryParamAsStringSlice(c, "currencies")
}

func getFromDate(c echo.Context) (time.Time, error) {
	return getQueryParamAsTime(c, "from_date")
}

func getToDate(c echo.Context) (time.Time, error) {
	return getQueryParamAsTime(c, "to_date")
}

func getAccountIds(c echo.Context) ([]int, error) {
	return getQueryParamAsIntSlice(c, "accounts")
}

func getCategoryIds(c echo.Context) ([]int, error) {
	return getQueryParamAsIntSlice(c, "categories")
}

func getTransactionTypes(c echo.Context) ([]string, error) {
	return getQueryParamAsStringSlice(c, "types")
}

func getQueryParamAsInt(c echo.Context, paramName string, defaultValue int) (int, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return defaultValue, nil // Return the default value if the parameter is not provided
	}
	val, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, errors.New("invalid value for " + paramName)
	}
	return val, nil
}

func getQueryParamAsIntSlice(c echo.Context, paramName string) ([]int, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return []int{}, nil
	}
	strSlice := strings.Split(paramStr, ",")
	var intSlice []int
	for _, str := range strSlice {
		val, err := strconv.Atoi(str)
		if err != nil {
			log.Warn("invalid value for " + paramName + ": [" + str + "] skipping")
			continue
		}
		intSlice = append(intSlice, val)
	}

	return intSlice, nil
}

func getQueryParamAsStringSlice(c echo.Context, paramName string) ([]string, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return []string{}, nil
	}
	return strings.Split(paramStr, ","), nil
}

func getQueryParamAsTime(c echo.Context, paramName string) (time.Time, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return time.Time{}, nil
	}
	parsedTime, err := time.Parse(time.DateOnly, paramStr)
	if err != nil {
		return time.Time{}, errors.New("invalid value for " + paramName)
	}
	return parsedTime, nil
}

// =============================================================================
// TRANSACTION DTOs AND THEIR METHODS
// =============================================================================

type TransactionWithAccount struct {
	models.Transaction
	Account     AccountDTO          `db:"accounts"`
	Currency    models.Currency     `db:"currencies"`
	AccountType models.AccountType  `db:"account_types"`
	Category    *CategoryDTO        `db:"user_categories"`
}

type CreateTransactionDTO struct {
	ID              *int             `json:"id"`
	UserID          *int             `json:"userId"`
	AccountID       int              `json:"accountId"`
	TargetAccountID *int             `json:"targetAccountId"`
	CategoryID      *int             `json:"categoryId"`
	Amount          decimal.Decimal  `json:"amount"`
	TargetAmount    *decimal.Decimal `json:"targetAmount"`
	Label           string           `json:"label"`
	Notes           *string          `json:"notes"`
	DateTime        *time.Time       `json:"dateTime"`
	IsTransfer      bool             `json:"isTransfer"`
	IsIncome        bool             `json:"isIncome"`
}

func (c *CreateTransactionDTO) UnmarshalJSON(data []byte) error {
	type Alias CreateTransactionDTO
	aux := &struct {
		CategoryID interface{} `json:"categoryId"`
		Amount     interface{} `json:"amount"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle CategoryID using helper function
	categoryID, err := parseCategoryIdFromInterface(aux.CategoryID)
	if err != nil {
		return err
	}
	c.CategoryID = categoryID

	// Handle Amount using helper function
	amount, err := parseAmountFromInterface(aux.Amount)
	if err != nil {
		return err
	}
	c.Amount = amount

	return nil
}

func (c *CreateTransactionDTO) MarshalJSON() ([]byte, error) {
	type Alias CreateTransactionDTO
	var targetAmount *float64
	if c.TargetAmount != nil {
		val, _ := c.TargetAmount.Float64()
		targetAmount = &val
	}

	return json.Marshal(&struct {
		Amount       float64  `json:"amount"`
		TargetAmount *float64 `json:"targetAmount"`
		*Alias
	}{
		Amount:       c.Amount.InexactFloat64(),
		TargetAmount: targetAmount,
		Alias:        (*Alias)(c),
	})
}

type PutTransactionDTO struct {
	CreateTransactionDTO
	ID         int   `json:"id"`
	IsTemplate *bool `json:"isTemplate"`
}

func (p *PutTransactionDTO) UnmarshalJSON(data []byte) error {
	type Alias PutTransactionDTO
	aux := &struct {
		CategoryID interface{} `json:"categoryId"`
		Amount     interface{} `json:"amount"`
		ID         interface{} `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle CategoryID using helper function
	categoryID, err := parseCategoryIdFromInterface(aux.CategoryID)
	if err != nil {
		return err
	}
	p.CategoryID = categoryID

	// Handle Amount using helper function
	amount, err := parseAmountFromInterface(aux.Amount)
	if err != nil {
		return err
	}
	p.Amount = amount

	// Handle ID using helper function
	id, err := parseIDFromInterface(aux.ID)
	if err != nil {
		return err
	}
	if id != 0 {
		p.ID = id
		// Keep embedded pointer in sync if present
		p.CreateTransactionDTO.ID = &id
	} else if p.CreateTransactionDTO.ID != nil && p.ID == 0 {
		// If decoder populated embedded ID, mirror it to top-level
		p.ID = *p.CreateTransactionDTO.ID
	}

	return nil
}

type ResponseTransactionDTO struct {
	ID                    int              `json:"id"`
	UserID                int              `json:"userId"`
	AccountID             int              `json:"accountId"`
	CategoryID            *int             `json:"categoryId"`
	Amount                decimal.Decimal  `json:"amount"`
	NewBalance            *decimal.Decimal `json:"newBalance"`
	Label                 string           `json:"label"`
	Notes                 *string          `json:"notes"`
	DateTime              *time.Time       `json:"dateTime"`
	IsTransfer            bool             `json:"isTransfer"`
	IsIncome              bool             `json:"isIncome"`
	BaseCurrencyAmount    *decimal.Decimal `json:"baseCurrencyAmount"`
	BaseCurrencyCode      *string          `json:"baseCurrencyCode"`
	LinkedTransactionID   *int             `json:"linkedTransactionId"`
	BalanceInBaseCurrency decimal.Decimal  `json:"balanceInBaseCurrency"`
	Category              CategoryDTO      `json:"category"`
	Account               AccountDTO       `json:"account"`
}

func (r *ResponseTransactionDTO) MarshalJSON() ([]byte, error) {
	type Alias ResponseTransactionDTO
	var newBalance *float64
	if r.NewBalance != nil {
		val, _ := r.NewBalance.Float64()
		newBalance = &val
	}
	var baseCurrencyAmount *float64
	if r.BaseCurrencyAmount != nil {
		val, _ := r.BaseCurrencyAmount.Float64()
		baseCurrencyAmount = &val
	}

	return json.Marshal(&struct {
		Amount                float64  `json:"amount"`
		NewBalance            *float64 `json:"newBalance"`
		BaseCurrencyAmount    *float64 `json:"baseCurrencyAmount"`
		BalanceInBaseCurrency float64  `json:"balanceInBaseCurrency"`
		*Alias
	}{
		Amount:                r.Amount.InexactFloat64(),
		NewBalance:            newBalance,
		BaseCurrencyAmount:    baseCurrencyAmount,
		BalanceInBaseCurrency: r.BalanceInBaseCurrency.InexactFloat64(),
		Alias:                 (*Alias)(r),
	})
}

type TransactionDetailDTO struct {
	ID                  int               `json:"id"`
	AccountID           int               `json:"accountId"`
	TargetAccountID     *int              `json:"targetAccountId"`
	CategoryID          *int              `json:"categoryId"`
	Amount              decimal.Decimal   `json:"amount"`
	TargetAmount        *decimal.Decimal  `json:"targetAmount"`
	Label               string            `json:"label"`
	Notes               string            `json:"notes"`
	DateTime            *time.Time        `json:"dateTime"`
	IsTransfer          bool              `json:"isTransfer"`
	IsIncome            bool              `json:"isIncome"`
	IsTemplate          *bool             `json:"isTemplate"`
	UserID              int               `json:"userId"`
	User                UserDTO           `json:"user"`
	Account             AccountDetailDTO  `json:"account"`
	BaseCurrencyAmount  decimal.Decimal   `json:"baseCurrencyAmount"`
	BaseCurrencyCode    string            `json:"baseCurrencyCode"`
	NewBalance          decimal.Decimal   `json:"newBalance"`
	Category            CategoryDetailDTO `json:"category"`
	LinkedTransactionID *int              `json:"linkedTransactionId"`
}

func (t *TransactionDetailDTO) MarshalJSON() ([]byte, error) {
	type Alias TransactionDetailDTO
	var targetAmount *float64
	if t.TargetAmount != nil {
		val, _ := t.TargetAmount.Float64()
		targetAmount = &val
	}

	return json.Marshal(&struct {
		Amount             float64  `json:"amount"`
		TargetAmount       *float64 `json:"targetAmount"`
		BaseCurrencyAmount float64  `json:"baseCurrencyAmount"`
		NewBalance         float64  `json:"newBalance"`
		*Alias
	}{
		Amount:             t.Amount.InexactFloat64(),
		TargetAmount:       targetAmount,
		BaseCurrencyAmount: t.BaseCurrencyAmount.InexactFloat64(),
		NewBalance:         t.NewBalance.InexactFloat64(),
		Alias:              (*Alias)(t),
	})
}

// =============================================================================
// SUPPORTING DTOs AND THEIR METHODS
// =============================================================================

type UserDTO struct {
	Email     string `json:"email"`
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type AccountDetailDTO struct {
	UserID                int                  `json:"userId"`
	AccountTypeID         int                  `json:"accountTypeId"`
	CurrencyID            int                  `json:"currencyId"`
	InitialBalance        decimal.Decimal      `json:"initialBalance"`
	Balance               decimal.Decimal      `json:"balance"`
	CreditLimit           decimal.Decimal      `json:"creditLimit"`
	Name                  string               `json:"name"`
	OpeningDate           *time.Time           `json:"openingDate"`
	Comment               string               `json:"comment"`
	IsHidden              bool                 `json:"isHidden"`
	ShowInReports         bool                 `json:"showInReports"`
	ID                    int                  `json:"id"`
	Currency              models.Currency      `json:"currency"`
	AccountType           AccountTypeDetailDTO `json:"accountType"`
	IsDeleted             bool                 `json:"isDeleted"`
	IsArchived            bool                 `json:"isArchived"`
	BalanceInBaseCurrency decimal.Decimal      `json:"balanceInBaseCurrency"`
	ArchivedAt            *string              `json:"archivedAt"`
}

func (a *AccountDetailDTO) MarshalJSON() ([]byte, error) {
	type Alias AccountDetailDTO
	return json.Marshal(&struct {
		InitialBalance        float64 `json:"initialBalance"`
		Balance               float64 `json:"balance"`
		CreditLimit           float64 `json:"creditLimit"`
		BalanceInBaseCurrency float64 `json:"balanceInBaseCurrency"`
		*Alias
	}{
		InitialBalance:        a.InitialBalance.InexactFloat64(),
		Balance:               a.Balance.InexactFloat64(),
		CreditLimit:           a.CreditLimit.InexactFloat64(),
		BalanceInBaseCurrency: a.BalanceInBaseCurrency.InexactFloat64(),
		Alias:                 (*Alias)(a),
	})
}

type AccountTypeDetailDTO struct {
	ID       int    `json:"id"`
	TypeName string `json:"type_name"`
	IsCredit bool   `json:"is_credit"`
}

type CategoryDetailDTO struct {
	Name      string              `json:"name"`
	ParentID  *int                `json:"parentId"`
	IsIncome  bool                `json:"isIncome"`
	ID        int                 `json:"id"`
	UserID    int                 `json:"userId"`
	CreatedAt string              `json:"createdAt"`
	UpdatedAt string              `json:"updatedAt"`
	Children  []CategoryDetailDTO `json:"children"`
}

// =============================================================================
// RAW DATA DTOs (for database operations)
// =============================================================================

type TransactionDetailRaw struct {
	models.Transaction
	User        models.User         `db:"users"`
	Account     models.Account      `db:"accounts"`
	Currency    models.Currency     `db:"currencies"`
	AccountType models.AccountType  `db:"account_types"`
	Category    *CategoryRaw        `db:"user_categories"`
}

type CategoryRaw struct {
	ID        *int    `db:"id"`
	Name      string  `db:"name"`
	ParentID  *int    `db:"parent_id"`
	IsIncome  bool    `db:"is_income"`
	UserID    *int    `db:"user_id"`
	IsDeleted bool    `db:"is_deleted"`
	CreatedAt *string `db:"created_at"`
	UpdatedAt *string `db:"updated_at"`
}

// =============================================================================
// CONVERSION FUNCTIONS
// =============================================================================

// ConvertTransactionsToResponseList converts a slice of TransactionWithAccount to ResponseTransactionDTO
func ConvertTransactionsToResponseList(transactions []TransactionWithAccount, baseCurrency models.Currency) []ResponseTransactionDTO {
	responseList := make([]ResponseTransactionDTO, 0, len(transactions))
	for _, transaction := range transactions {
		responseList = append(responseList, ConvertToResponseTransaction(transaction, baseCurrency))
	}
	return responseList
}

// ConvertToResponseTransaction converts a single TransactionWithAccount to ResponseTransactionDTO
func ConvertToResponseTransaction(twa TransactionWithAccount, baseCurrency models.Currency) ResponseTransactionDTO {
	var creditLimit decimal.Decimal
	if twa.Account.AccountType.IsCredit {
		if twa.Account.CreditLimit != nil {
			creditLimit = *twa.Account.CreditLimit
		} else {
			creditLimit = decimal.Zero
		}
	} else {
		creditLimit = decimal.Zero
	}

	var balanceInBaseCurrency decimal.Decimal
	if twa.BaseCurrencyAmount != nil {
		balanceInBaseCurrency = *twa.BaseCurrencyAmount
	} else {
		balanceInBaseCurrency = decimal.Zero
	}

	var baseCurrencyAmount *decimal.Decimal
	if twa.BaseCurrencyAmount != nil {
		baseCurrencyAmount = twa.BaseCurrencyAmount
	} else {
		baseCurrencyAmount = &decimal.Zero
	}
	// log.Info(twa.Account.AccountType)
	return ResponseTransactionDTO{
		ID:                    *twa.ID,
		UserID:                twa.UserID,
		AccountID:             twa.AccountID,
		CategoryID:            twa.CategoryID,
		Amount:                twa.Amount,
		Label:                 twa.Label,
		Notes:                 twa.Notes,
		DateTime:              twa.DateTime,
		IsTransfer:            twa.IsTransfer,
		IsIncome:              twa.IsIncome,
		LinkedTransactionID:   twa.LinkedTransactionID,
		BaseCurrencyAmount:    baseCurrencyAmount,
		BaseCurrencyCode:      &baseCurrency.Code,
		NewBalance:            twa.NewBalance,
		BalanceInBaseCurrency: balanceInBaseCurrency,
		Account: AccountDTO{
			ID:          twa.Account.ID,
			Name:        twa.Account.Name,
			Balance:     twa.Account.Balance,
			CreditLimit: &creditLimit,
			OpeningDate: twa.Account.OpeningDate,
			Comment:     twa.Account.Comment,
			Currency: twa.Account.Currency,
			AccountType: twa.Account.AccountType,
		},
		Category: CategoryDTO{
			ID:        twa.Category.ID,
			Name:      twa.Category.Name,
			ParentID:  twa.Category.ParentID,
			IsIncome:  twa.Category.IsIncome,
			UserID:    twa.Category.UserID,
			IsDeleted: twa.Category.IsDeleted,
			CreatedAt: twa.Category.CreatedAt,
			UpdatedAt: twa.Category.UpdatedAt,
		},
	}
}
