package errors

import "errors"

var (
	ErrNoAccountFound = errors.New("no account found with the provided ID")
)