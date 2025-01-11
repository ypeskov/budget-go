package models

type Currency struct {
	ID        int    `db:"id"`
	Code      string `db:"code"`
	Name      string `db:"name"`
	IsDeleted bool   `db:"is_deleted"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}
