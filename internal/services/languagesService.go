package services

import (
	"sync"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/repositories/languages"
)

type LanguagesService interface {
	GetLanguages() ([]models.Language, error)
}

type LanguagesServiceInstance struct {
	languagesRepo languages.Repository
}

var (
	languagesInstance *LanguagesServiceInstance
	languagesOnce     sync.Once
)

func NewLanguagesService(languagesRepo languages.Repository) LanguagesService {
	languagesOnce.Do(func() {
		logger.Debug("Creating LanguagesService instance")
		languagesInstance = &LanguagesServiceInstance{
			languagesRepo: languagesRepo,
		}
	})

	return languagesInstance
}

func (l *LanguagesServiceInstance) GetLanguages() ([]models.Language, error) {
	return l.languagesRepo.GetLanguages()
}
