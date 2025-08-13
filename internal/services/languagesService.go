package services

import (
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/languages"
)

type LanguagesService interface {
	GetLanguages() ([]models.Language, error)
}

type LanguagesServiceInstance struct {
	languagesRepo languages.Repository
}

func NewLanguagesService(languagesRepo languages.Repository) LanguagesService {
	return &LanguagesServiceInstance{
		languagesRepo: languagesRepo,
	}
}

func (l *LanguagesServiceInstance) GetLanguages() ([]models.Language, error) {
	return l.languagesRepo.GetLanguages()
}