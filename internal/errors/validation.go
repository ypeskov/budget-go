package errors

import "errors"

// Validation and utility errors
var (
	ErrInvalidPerPage    = errors.New("per_page must be greater than 0")
	ErrInvalidParamValue = errors.New("invalid parameter value")
	ErrInvalidDateValue  = errors.New("invalid date value")
)

// InvalidParamError represents an error for invalid parameter values
type InvalidParamError struct {
	ParamName string
	Value     string
}

func (e *InvalidParamError) Error() string {
	return "invalid value for " + e.ParamName + ": " + e.Value
}

// InvalidDateError represents an error for invalid date values
type InvalidDateError struct {
	ParamName string
	Value     string
}

func (e *InvalidDateError) Error() string {
	return "invalid date value for " + e.ParamName + ": " + e.Value
}