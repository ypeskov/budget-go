package models

import "fmt"

type User struct {
	ID             int    `db:"id"`
	Email          string `db:"email"`
	FirsName       string `db:"first_name"`
	LastName       string `db:"last_name"`
	PasswordHash   string `db:"password_hash"`
	IsActive       bool   `db:"is_active"`
	BaseCurrencyID int    `db:"base_currency_id"`
	IsDeleted      bool   `db:"is_deleted"`
	CreatedAt      string `db:"created_at"`
	UpdatedAt      string `db:"updated_at"`
}

func (u *User) String() string {
	return fmt.Sprintf("[User: %d, %s, %s]", u.ID, u.FirsName, u.LastName)
}
