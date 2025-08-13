package dto

import (
	"time"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/utils"
	"github.com/shopspring/decimal"
)

type CreateBudgetDTO struct {
	Name         string                `json:"name" validate:"required"`
	CurrencyID   int                   `json:"currencyId" validate:"required"`
	TargetAmount decimal.Decimal       `json:"targetAmount" validate:"required"`
	Period       string                `json:"period" validate:"required"`
	Repeat       bool                  `json:"repeat"`
	StartDate    *utils.CustomDate     `json:"startDate" validate:"required"`
	EndDate      *utils.CustomDate     `json:"endDate" validate:"required"`
	Categories   []int                 `json:"categories"`
	Comment      *string               `json:"comment"`
}

type UpdateBudgetDTO struct {
	ID           int                   `json:"id" validate:"required"`
	Name         string                `json:"name" validate:"required"`
	CurrencyID   int                   `json:"currencyId" validate:"required"`
	TargetAmount decimal.Decimal       `json:"targetAmount" validate:"required"`
	Period       string                `json:"period" validate:"required"`
	Repeat       bool                  `json:"repeat"`
	StartDate    *utils.CustomDate     `json:"startDate" validate:"required"`
	EndDate      *utils.CustomDate     `json:"endDate" validate:"required"`
	Categories   []int                 `json:"categories"`
	Comment      *string               `json:"comment"`
}

type BudgetResponseDTO struct {
	ID                 int                 `json:"id"`
	Name               string              `json:"name"`
	CurrencyID         int                 `json:"currencyId"`
	TargetAmount       decimal.Decimal     `json:"targetAmount"`
	CollectedAmount    decimal.Decimal     `json:"collectedAmount"`
	Period             string              `json:"period"`
	Repeat             bool                `json:"repeat"`
	StartDate          *time.Time          `json:"startDate"`
	EndDate            *time.Time          `json:"endDate"`
	IncludedCategories string              `json:"includedCategories"`
	Comment            *string             `json:"comment"`
	IsArchived         bool                `json:"isArchived"`
	Currency           models.Currency     `json:"currency"`
}

type BudgetListFilters struct {
	Include string `query:"include"`
}