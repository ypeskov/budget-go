package models

type UserCategory struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	ParentID  *int   `db:"parent_id"`
	IsIncome  bool   `db:"is_income"`
	UserID    int    `db:"user_id"`
	IsDeleted bool   `db:"is_deleted"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}
