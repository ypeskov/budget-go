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

type CreateCategoryDTO struct {
	Name      string `json:"name"`
	IsIncome  bool   `json:"isIncome"`
	ParentID  *int   `json:"parentId"`
	IsDeleted bool   `json:"isDeleted"`
}

type CategoryDetailDTO struct {
	ID        int                 `json:"id"`
	Name      string              `json:"name"`
	ParentID  *int                `json:"parentId"`
	IsIncome  bool                `json:"isIncome"`
	UserID    int                 `json:"userId"`
	CreatedAt string              `json:"createdAt"`
	UpdatedAt string              `json:"updatedAt"`
	Children  []CategoryDetailDTO `json:"children"`
}
