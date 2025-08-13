package utils

import (
	"fmt"
	"strings"
	"time"
)

// CustomDate handles both date-only (YYYY-MM-DD) and datetime formats
type CustomDate struct {
	time.Time
}

// UnmarshalJSON handles parsing of both date and datetime formats
func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	dateStr := strings.Trim(string(data), `"`)
	
	// Try parsing as date-only format first (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		cd.Time = t
		return nil
	}
	
	// Try parsing as RFC3339 datetime format
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		cd.Time = t
		return nil
	}
	
	// Try parsing as ISO 8601 datetime format without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", dateStr); err == nil {
		cd.Time = t
		return nil
	}
	
	// Try parsing as common datetime format
	if t, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
		cd.Time = t
		return nil
	}
	
	return fmt.Errorf("cannot parse date: %s", dateStr)
}

// MarshalJSON outputs in ISO format
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + cd.Time.Format("2006-01-02T15:04:05Z07:00") + `"`), nil
}

// ToTime returns the underlying time.Time
func (cd *CustomDate) ToTime() *time.Time {
	return &cd.Time
}