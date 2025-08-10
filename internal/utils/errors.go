package utils

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"ypeskov/budget-go/internal/models"
)

// LogAndReturnError logs detailed error information for developers and returns user-friendly error to client
func LogAndReturnError(c echo.Context, err error, httpStatus int) error {
	// Get request information
	method := c.Request().Method
	path := c.Request().URL.Path

	// Try to get user from context
	var userId interface{} = "unknown"
	if user, ok := c.Get("authenticated_user").(*models.User); ok && user != nil {
		userId = user.ID
	}

	// Log with detailed information
	log.WithFields(log.Fields{
		"method":     method,
		"path":       path,
		"userId":     userId,
		"httpStatus": httpStatus,
		"error":      err.Error(),
		"errorType":  fmt.Sprintf("%T", err),
	}).Error("Request failed")

	// Return informative error to client
	errorMessage := "Internal server error"
	switch httpStatus {
	case http.StatusBadRequest:
		errorMessage = sanitizeValidationError(err.Error())
	case http.StatusNotFound:
		errorMessage = err.Error()
	}

	return c.JSON(httpStatus, map[string]string{
		"error": errorMessage,
	})
}

// sanitizeValidationError converts technical validation errors into user-friendly messages
func sanitizeValidationError(errorStr string) string {
	// Patterns for different types of validation errors
	patterns := []struct {
		pattern     string
		replacement string
	}{
		// strconv.Atoi errors for categoryId
		{`strconv\.Atoi: parsing "([^"]*)" for categoryId: invalid syntax`, "Invalid value for categoryId: expected a number but got '$1'"},
		{`strconv\.Atoi: parsing "([^"]*)".*invalid syntax`, "Invalid number format: expected a number but got '$1'"},
		
		// decimal parsing errors for amount
		{`can't convert ([^\s]+) to decimal`, "Invalid amount format: '$1' is not a valid number"},
		{`decimal: can't convert ([^\s]+) to decimal`, "Invalid amount format: '$1' is not a valid number"},
		
		// JSON unmarshaling errors
		{`json: cannot unmarshal string "([^"]*)" into Go struct field.*\.categoryId of type int`, "Invalid categoryId: expected a number but got '$1'"},
		{`json: cannot unmarshal string "([^"]*)" into Go struct field.*\.amount`, "Invalid amount: expected a number but got '$1'"},
		
		// General JSON binding errors
		{`failed to bind request body:.*json: cannot unmarshal`, "Invalid request format: check your JSON data"},
		{`failed to bind request body:.*`, "Invalid request format"},
		
		// EOF and syntax errors
		{`unexpected end of JSON input`, "Incomplete JSON data"},
		{`invalid character.*looking for beginning of value`, "Invalid JSON format"},
	}

	sanitized := errorStr
	
	// Apply replacement patterns
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		if re.MatchString(sanitized) {
			sanitized = re.ReplaceAllString(sanitized, p.replacement)
			break // Use the first matching pattern
		}
	}

	// If error wasn't handled by patterns, return original text
	// (for custom application errors)
	return sanitized
}