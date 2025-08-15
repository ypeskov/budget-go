package models

import (
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

type PeriodEnum string

const (
	PeriodDaily   PeriodEnum = "DAILY"
	PeriodWeekly  PeriodEnum = "WEEKLY"
	PeriodMonthly PeriodEnum = "MONTHLY"
	PeriodYearly  PeriodEnum = "YEARLY"
	PeriodCustom  PeriodEnum = "CUSTOM"
)

// ValidatePeriod checks if the given string is a valid period (accepts both upper and lowercase)
func ValidatePeriod(period string) bool {
	switch strings.ToUpper(period) {
	case string(PeriodDaily), string(PeriodWeekly), string(PeriodMonthly), string(PeriodYearly), string(PeriodCustom):
		return true
	default:
		return false
	}
}

// NormalizePeriod converts period to the format expected by the database (uppercase)
func NormalizePeriod(period string) string {
	return strings.ToUpper(period)
}

// FormatPeriodForAPI converts period from database format to API format (lowercase)
func FormatPeriodForAPI(period string) string {
	return strings.ToLower(period)
}

// GetValidPeriods returns all valid period values (lowercase for API compatibility)
func GetValidPeriods() []string {
	return []string{
		"daily", "weekly", "monthly", "yearly", "custom",
	}
}

type Budget struct {
	ID                 *int            `json:"id" db:"id"`
	UserID             int             `json:"userId" db:"user_id"`
	Name               string          `json:"name" db:"name"`
	CurrencyID         int             `json:"currencyId" db:"currency_id"`
	TargetAmount       decimal.Decimal `json:"targetAmount" db:"target_amount"`
	CollectedAmount    decimal.Decimal `json:"collectedAmount" db:"collected_amount"`
	Period             string          `json:"period" db:"period"`
	Repeat             bool            `json:"repeat" db:"repeat"`
	StartDate          *time.Time      `json:"startDate" db:"start_date"`
	EndDate            *time.Time      `json:"endDate" db:"end_date"`
	IncludedCategories *string         `json:"includedCategories" db:"included_categories"`
	Comment            *string         `json:"comment" db:"comment"`
	IsDeleted          bool            `json:"isDeleted" db:"is_deleted"`
	IsArchived         bool            `json:"isArchived" db:"is_archived"`
	CreatedAt          *time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt          *time.Time      `json:"updatedAt" db:"updated_at"`
}
