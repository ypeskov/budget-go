package models

import (
	"encoding/json"
	"time"
)

type Currency struct {
	ID        int       `db:"id"`
	Code      string    `db:"code"`
	Name      string    `db:"name"`
	IsDeleted bool      `db:"is_deleted"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// MarshalJSON customizes JSON output to exclude internal metadata fields
func (c *Currency) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID   int    `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	}{
		ID:   c.ID,
		Code: c.Code,
		Name: c.Name,
	})
}
