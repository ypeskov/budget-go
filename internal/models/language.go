package models

import "time"

type Language struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Code      string    `json:"code" db:"code"`
	IsDeleted bool      `json:"isDeleted" db:"is_deleted"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}