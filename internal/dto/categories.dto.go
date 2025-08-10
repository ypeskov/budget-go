package dto

type CategoryDTO struct {
	ID         *int    `json:"id" db:"id"`
	Name       *string `json:"name" db:"name"`
	ParentID   *int    `json:"parentId" db:"parent_id"`
	ParentName *string `json:"parentName" db:"parent_name"`
	IsIncome   *bool   `json:"isIncome" db:"is_income"`
	UserID     *int    `json:"userId" db:"user_id"`
	IsDeleted  *bool   `json:"isDeleted" db:"is_deleted"`
	CreatedAt  *string `json:"createdAt" db:"created_at"`
	UpdatedAt  *string `json:"updatedAt" db:"updated_at"`
}
