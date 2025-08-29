package util

import (
	"fmt"
	"time"
)

func ParseDate(date string) (time.Time, error) {
	if date == "today" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
	}

	// Try YYYY-MM-DD format
	return time.Parse("2006-01-02", date)
}

func ParseTime(when string, timeProvider TimeProvider) (time.Time, error) {
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

type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

func (t *RealTimeProvider) Now() time.Time {
	return time.Now()
}
