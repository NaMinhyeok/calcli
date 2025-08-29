package util

import (
	"fmt"
	"time"
)

// ParseDate parses date strings in various formats
func ParseDate(date string) (time.Time, error) {
	if date == "today" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
	}

	// Try YYYY-MM-DD format
	return time.Parse("2006-01-02", date)
}

// ParseTime parses time strings with TimeProvider injection
func ParseTime(when string, timeProvider TimeProvider) (time.Time, error) {
	// Try different formats
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"15:04", // today at time
	}

	for _, format := range formats {
		if t, err := time.Parse(format, when); err == nil {
			if format == "15:04" {
				now := timeProvider.Now()
				return time.Date(now.Year(), now.Month(), now.Day(),
					t.Hour(), t.Minute(), 0, 0, time.Local), nil
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", when)
}

// TimeProvider interface for dependency injection
type TimeProvider interface {
	Now() time.Time
}
