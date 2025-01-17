package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type ExchangeRates struct {
	ID               uint      `db:"id"`
	Rates            JSONB     `db:"rates"`
	ActualDate       time.Time `db:"actual_date"`
	BaseCurrencyCode string    `db:"base_currency_code"`
	ServiceName      string    `db:"service_name"`
	IsDeleted        bool      `db:"is_deleted"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// JSONB is a custom type for handling PostgreSQL JSONB
type JSONB map[string]interface{}

func (j *JSONB) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (er *ExchangeRates) String() string {
	return fmt.Sprintf("ExchangeRates{ID: %d, Rates: %v, ActualDate: %s, BaseCurrencyCode: %s, ServiceName: %s, IsDeleted: %t, CreatedAt: %s, UpdatedAt: %s}",
		er.ID, er.Rates, er.ActualDate, er.BaseCurrencyCode, er.ServiceName, er.IsDeleted, er.CreatedAt, er.UpdatedAt)
}
