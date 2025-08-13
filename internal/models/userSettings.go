package models

import "time"

type UserSettings struct {
	ID        int                    `json:"id" db:"id"`
	UserID    int                    `json:"user_id" db:"user_id"`
	Settings  map[string]interface{} `json:"settings" db:"settings"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}