package services

import (
	log "github.com/sirupsen/logrus"
	userModel "ypeskov/budget-go/internal/models"
	userRepo "ypeskov/budget-go/internal/repositories/user"
)

type UserService interface {
	GetAllUsers() ([]*userModel.User, error)
	GetUserByEmail(email string) (*userModel.User, error)
}

type UserServiceInstance struct {
	userRepo userRepo.RepositoryInterface
}

func NewUserService(userRepo userRepo.RepositoryInterface) UserService {
	return &UserServiceInstance{
		userRepo: userRepo,
	}
}

func (us *UserServiceInstance) GetAllUsers() ([]*userModel.User, error) {
	log.Debug("GetAllUsers service called")
	users, err := us.userRepo.GetAllUsers()
	if err != nil {
		log.Error("Error getting users: ", err)
		return nil, err
	}

	return users, nil
}

func (us *UserServiceInstance) GetUserByEmail(email string) (*userModel.User, error) {
	log.Debug("GetUserByEmail service called")
	user, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		log.Error("Error getting user by email: ", err)
		return nil, err
	}

	return user, nil
}
