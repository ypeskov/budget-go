package errors

import "strconv"

// NotFoundError is an error that indicates a resource was not found
type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return e.Resource + " with ID " + strconv.Itoa(e.ID) + " not found"
}

// InvalidRequestError is an error that indicates a request was invalid
type InvalidRequestError struct {
	Message string
}

func (e *InvalidRequestError) Error() string {
	return e.Message
}

// BadRequestError is an error that indicates a request was invalid
type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}