package dto

type CategoryDTO struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ParentID  *int   `json:"parentId"`
	IsIncome  bool   `json:"isIncome"`
	UserID    int    `json:"userId"`
	IsDeleted bool   `json:"isDeleted"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
