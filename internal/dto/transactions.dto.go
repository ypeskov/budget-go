package dto

import (
	"encoding/json"
	"time"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/utils"

	"github.com/shopspring/decimal"
)

// =============================================================================
// TRANSACTION DTOs AND THEIR METHODS
// =============================================================================

type TransactionWithAccount struct {
	models.Transaction
	Account     AccountDTO         `db:"accounts"`
	Currency    models.Currency    `db:"currencies"`
	AccountType models.AccountType `db:"account_types"`
	Category    *CategoryDTO       `db:"user_categories"`
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
	categoryID, err := utils.ParseCategoryIdFromInterface(aux.CategoryID)
	if err != nil {
		return err
	}
	c.CategoryID = categoryID

	// Handle Amount using helper function
	amount, err := utils.ParseAmountFromInterface(aux.Amount)
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
	// Explicitly define all fields instead of embedding to fix JSON unmarshaling
	ID              int              `json:"id"`
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
	IsTemplate      *bool            `json:"isTemplate"`
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
	categoryID, err := utils.ParseCategoryIdFromInterface(aux.CategoryID)
	if err != nil {
		return err
	}
	p.CategoryID = categoryID

	// Handle Amount using helper function
	amount, err := utils.ParseAmountFromInterface(aux.Amount)
	if err != nil {
		return err
	}
	p.Amount = amount

	// Handle ID using helper function
	id, err := utils.ParseIDFromInterface(aux.ID)
	if err != nil {
		return err
	}
	if id != 0 {
		p.ID = id
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
	ID                  int                     `json:"id"`
	AccountID           int                     `json:"accountId"`
	TargetAccountID     *int                    `json:"targetAccountId"`
	CategoryID          *int                    `json:"categoryId"`
	Amount              decimal.Decimal         `json:"amount"`
	TargetAmount        *decimal.Decimal        `json:"targetAmount"`
	Label               string                  `json:"label"`
	Notes               string                  `json:"notes"`
	DateTime            *time.Time              `json:"dateTime"`
	IsTransfer          bool                    `json:"isTransfer"`
	IsIncome            bool                    `json:"isIncome"`
	IsTemplate          *bool                   `json:"isTemplate"`
	UserID              int                     `json:"userId"`
	User                UserRegisterResponseDTO `json:"user"`
	Account             AccountDetailDTO        `json:"account"`
	BaseCurrencyAmount  decimal.Decimal         `json:"baseCurrencyAmount"`
	BaseCurrencyCode    string                  `json:"baseCurrencyCode"`
	NewBalance          decimal.Decimal         `json:"newBalance"`
	Category            CategoryDetailDTO       `json:"category"`
	LinkedTransactionID *int                    `json:"linkedTransactionId"`
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

// =============================================================================
// RAW DATA DTOs (for database operations)
// =============================================================================

type TransactionDetailRaw struct {
	models.Transaction
	User        models.User        `db:"users"`
	Account     models.Account     `db:"accounts"`
	Currency    models.Currency    `db:"currencies"`
	AccountType models.AccountType `db:"account_types"`
	Category    *models.UserCategory `db:"user_categories"`
}


