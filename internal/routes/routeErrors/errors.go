package routeErrors

import "strconv"

type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return e.Resource + " with ID " + strconv.Itoa(e.ID) + " not found"
}

type InvalidRequestError struct {
	Message string
}

func (e *InvalidRequestError) Error() string {
	return e.Message
}
