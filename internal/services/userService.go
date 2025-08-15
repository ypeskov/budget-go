package services

import (
	userModel "ypeskov/budget-go/internal/models"
	userRepo "ypeskov/budget-go/internal/repositories/user"
	"ypeskov/budget-go/internal/dto"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAllUsers() ([]*userModel.User, error)
	GetUserByEmail(email string) (*userModel.User, error)
	CreateUser(user *userModel.User) (*userModel.User, error)
	RegisterUser(userDTO *dto.UserRegisterRequestDTO) (*userModel.User, error)
	LoginOrRegisterOAuth(email, firstName, lastName string) (*userModel.User, error)
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

func (us *UserServiceInstance) CreateUser(user *userModel.User) (*userModel.User, error) {
	log.Debug("CreateUser service called")
	createdUser, err := us.userRepo.CreateUser(user)
	if err != nil {
		log.Error("Error creating user: ", err)
		return nil, err
	}

	return createdUser, nil
}

func (us *UserServiceInstance) RegisterUser(userDTO *dto.UserRegisterRequestDTO) (*userModel.User, error) {
	log.Debug("RegisterUser service called")
	
	// Check if user already exists
	existingUser, err := us.userRepo.GetUserByEmail(userDTO.Email)
	if err == nil && existingUser != nil {
		log.Error("User already exists with email: ", userDTO.Email)
		return nil, &UserAlreadyExistsError{Email: userDTO.Email}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Error hashing password: ", err)
		return nil, err
	}

	// Create new user
	newUser := &userModel.User{
		Email:          userDTO.Email,
		FirstName:      userDTO.FirstName,
		LastName:       userDTO.LastName,
		PasswordHash:   string(hashedPassword),
		IsActive:       true,
		BaseCurrencyID: 1, // Default to USD or first currency
		IsDeleted:      false,
	}

	createdUser, err := us.CreateUser(newUser)
	if err != nil {
		log.Error("Error creating user: ", err)
		return nil, err
	}

	return createdUser, nil
}

func (us *UserServiceInstance) LoginOrRegisterOAuth(email, firstName, lastName string) (*userModel.User, error) {
	log.Debug("LoginOrRegisterOAuth service called")
	
	existingUser, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		log.Debug("User not found, creating new user")
		newUser := &userModel.User{
			Email:          email,
			FirstName:      firstName,
			LastName:       lastName,
			PasswordHash:   "",
			IsActive:       true,
			BaseCurrencyID: 1,
			IsDeleted:      false,
		}
		
		return us.CreateUser(newUser)
	}

	if !existingUser.IsActive {
		log.Error("User not activated: ", email)
		return nil, &UserNotActivatedError{Email: email}
	}

	return existingUser, nil
}

type UserNotActivatedError struct {
	Email string
}

func (e *UserNotActivatedError) Error() string {
	return "User not activated: " + e.Email
}

type UserAlreadyExistsError struct {
	Email string
}

func (e *UserAlreadyExistsError) Error() string {
	return "User already exists: " + e.Email
}
