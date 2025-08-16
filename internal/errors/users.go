package errors

import "errors"

// Sentinel errors for users
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserNotFoundError represents an error when a user is not found by email
type UserNotFoundError struct {
	Email string
}

func (e *UserNotFoundError) Error() string {
	return "User not found: " + e.Email
}

// UserAlreadyExistsError represents an error when trying to create a user that already exists
type UserAlreadyExistsError struct {
	Email string
}

func (e *UserAlreadyExistsError) Error() string {
	return "User already exists: " + e.Email
}

// UserNotActivatedError represents an error when a user is not activated
type UserNotActivatedError struct {
	Email string
}

func (e *UserNotActivatedError) Error() string {
	return "User not activated: " + e.Email
}

// UserDeletedError represents an error when a user is deleted
type UserDeletedError struct {
	Email string
}

func (e *UserDeletedError) Error() string {
	return "User is deleted: " + e.Email
}

// InvalidCredentialsError represents an error when login credentials are invalid
type InvalidCredentialsError struct {
	Email string
}

func (e *InvalidCredentialsError) Error() string {
	return "Invalid credentials for user: " + e.Email
}