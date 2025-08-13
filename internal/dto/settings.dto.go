package dto

// BaseCurrencyDTO has been consolidated with models.Currency
// Use models.Currency directly for all currency operations

type UpdateSettingsDTO struct {
	Language string `json:"language" validate:"required"`
}
