package dto

type TemplateDTO struct {
	ID         int    `json:"id"`
	Label      string `json:"label"`
	CategoryID int    `json:"categoryId"`
	Category   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}
