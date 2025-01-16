package dto

type CurrencyDTO struct {
	ID        int    `json:"id" db:"id"`
	Code      string `json:"code" db:"code"`
	Name      string `json:"name" db:"name"`
	CreatedAt string `json:"createdAt" db:"created_at"`
	UpdatedAt string `json:"updatedAt" db:"updated_at"`
}
