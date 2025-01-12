package services

import (
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/categories"
)

type CategoriesService interface {
	GetUserCategories(userId int) ([]models.UserCategory, error)
}

type CategoryServiceInstance struct {
	categoriesRepo categories.Repository
}

func NewCategoriesService(repository categories.Repository) CategoriesService {
	return &CategoryServiceInstance{
		categoriesRepo: repository,
	}
}

func (c *CategoryServiceInstance) GetUserCategories(userId int) ([]models.UserCategory, error) {
	userCategories, err := c.categoriesRepo.GetUserCategories(userId)
	if err != nil {
		return nil, err
	}

	return userCategories, nil
}
