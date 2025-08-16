package errors

import "errors"

// Authentication and token-related errors
var (
	ErrMissingAuthToken     = errors.New("missing auth-token")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrTokenExpired         = errors.New("token has expired")
)