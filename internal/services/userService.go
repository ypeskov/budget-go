package services

import (
	"sync"
	"ypeskov/budget-go/internal/dto"
	"ypeskov/budget-go/internal/errors"
	"ypeskov/budget-go/internal/logger"
	"ypeskov/budget-go/internal/models"
	"ypeskov/budget-go/internal/queue"
	userRepo "ypeskov/budget-go/internal/repositories/user"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAllUsers() ([]*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	RegisterUser(userDTO *dto.UserRegisterRequestDTO,
		currenciesService CurrenciesService,
		activationTokenService ActivationTokenService) (*models.User, error)
	LoginUser(loginDTO *dto.UserLoginDTO) (*models.User, error)
	LoginOrRegisterOAuth(email, firstName, lastName string) (*models.User, error)
	ActivateUser(userID int) error
}

type UserServiceInstance struct {
	userRepo     userRepo.RepositoryInterface
	queueService queue.QueueService
}

var (
	userInstance *UserServiceInstance
	userOnce     sync.Once
)

func NewUserService(userRepo userRepo.RepositoryInterface, queueService queue.QueueService) UserService {
	userOnce.Do(func() {
		logger.Debug("Creating UserService instance")
		userInstance = &UserServiceInstance{
			userRepo:     userRepo,
			queueService: queueService,
		}
	})

	return userInstance
}

func (us *UserServiceInstance) GetAllUsers() ([]*models.User, error) {
	logger.Debug("GetAllUsers service called")
	users, err := us.userRepo.GetAllUsers()
	if err != nil {
		logger.Error("Error getting users", "error", err)
		return nil, err
	}

	return users, nil
}

func (us *UserServiceInstance) GetUserByEmail(email string) (*models.User, error) {
	logger.Debug("GetUserByEmail service called")
	user, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		logger.Error("Error getting user by email", "error", err)
		return nil, err
	}

	return user, nil
}

// CreateUser creates a new user in the repository
// and returns the created user or an error if the creation fails.
// It is used internally by the service and should not be exposed as a public API.
// This method is typically called after validating the user data.
// It is not intended for direct use by clients of the service.
func (us *UserServiceInstance) CreateUser(user *models.User) (*models.User, error) {
	logger.Debug("CreateUser service called")
	createdUser, err := us.userRepo.CreateUser(user)
	if err != nil {
		logger.Error("Error creating user", "error", err)
		return nil, err
	}

	logger.Debug("Created user", "user", createdUser)
	return createdUser, nil
}

func (us *UserServiceInstance) RegisterUser(userDTO *dto.UserRegisterRequestDTO,
	currenciesService CurrenciesService,
	activationTokenService ActivationTokenService) (*models.User, error) {
	logger.Debug("RegisterUser service called")

	// Check if user already exists
	existingUser, err := us.userRepo.GetUserByEmail(userDTO.Email)
	if err == nil && existingUser != nil {
		logger.Error("User already exists with email", "email", userDTO.Email)
		return nil, &errors.UserAlreadyExistsError{Email: userDTO.Email}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing password", "error", err)
		return nil, err
	}

	// Get default currency ID
	defaultCurrency, err := currenciesService.GetCurrencyByCode(models.DefaultCurrency)
	if err != nil {
		logger.Error("Error getting default currency", "error", err)
		// Fallback to currency ID 1 if default currency not found
		defaultCurrency.ID = 1
	}

	// Create new user in INACTIVE state
	newUser := &models.User{
		Email:          userDTO.Email,
		FirstName:      userDTO.FirstName,
		LastName:       userDTO.LastName,
		PasswordHash:   string(hashedPassword),
		IsActive:       false, // User starts as inactive
		BaseCurrencyID: defaultCurrency.ID,
		IsDeleted:      false,
	}

	createdUser, err := us.CreateUser(newUser)
	if err != nil {
		logger.Error("Error creating user", "error", err)
		return nil, err
	}

	// Create activation token
	activationToken, err := activationTokenService.CreateActivationToken(createdUser.ID)
	if err != nil {
		logger.Error("Error creating activation token", "error", err)
		return nil, err
	}

	// Queue activation email to be sent asynchronously
	if us.queueService != nil {
		err = us.queueService.EnqueueActivationEmail(createdUser.Email, createdUser.FirstName, activationToken.Token)
		if err != nil {
			logger.Error("Error queuing activation email", "error", err)
			logger.Warn("User registered but activation email failed to queue")
		} else {
			logger.Info("Activation email queued successfully for user", "email", createdUser.Email)
		}
	} else {
		// Fallback to synchronous email sending if queue service is not available
		err = activationTokenService.SendActivationEmail(createdUser, activationToken.Token)
		if err != nil {
			logger.Error("Error sending activation email", "error", err)
			logger.Warn("User registered but activation email failed to send")
		}
	}

	return createdUser, nil
}

func (us *UserServiceInstance) LoginOrRegisterOAuth(email, firstName, lastName string) (*models.User, error) {
	logger.Debug("LoginOrRegisterOAuth service called")

	existingUser, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		logger.Debug("User not found, creating new user")
		newUser := &models.User{
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
		logger.Error("User not activated", "email", email)
		return nil, &errors.UserNotActivatedError{Email: email}
	}

	return existingUser, nil
}

func (us *UserServiceInstance) LoginUser(loginDTO *dto.UserLoginDTO) (*models.User, error) {
	logger.Debug("LoginUser service called")

	// Get user by email
	user, err := us.userRepo.GetUserByEmail(loginDTO.Email)
	if err != nil {
		logger.Warn("Error getting user by email", "error", err)
		return nil, &errors.UserNotFoundError{Email: loginDTO.Email}
	}

	// Check if user is deleted
	if user.IsDeleted {
		logger.Warn("User is deleted", "email", loginDTO.Email)
		return nil, &errors.UserDeletedError{Email: loginDTO.Email}
	}

	// Check if user is activated
	if !user.IsActive {
		logger.Warn("User is not activated", "email", loginDTO.Email)
		return nil, &errors.UserNotActivatedError{Email: loginDTO.Email}
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginDTO.Password))
	if err != nil {
		logger.Warn("Passwords do not match for user", "email", loginDTO.Email)
		return nil, &errors.InvalidCredentialsError{Email: loginDTO.Email}
	}

	logger.Debug("User login successful", "email", loginDTO.Email)
	return user, nil
}

func (us *UserServiceInstance) ActivateUser(userID int) error {
	logger.Debug("ActivateUser service called for user ID", "userID", userID)

	return us.userRepo.ActivateUser(userID)
}

