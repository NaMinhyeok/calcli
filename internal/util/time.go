package util

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func ParseDate(date string) (time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	switch date {
	case "today":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	}

	if matched, days := parseRelativeDays(date); matched {
		return today.AddDate(0, 0, days), nil
	}

	if matched, weeks := parseRelativeWeeks(date); matched {
		return today.AddDate(0, 0, weeks*7), nil
	}

	return time.Parse("2006-01-02", date)
}

func parseRelativeDays(date string) (bool, int) {
	re := regexp.MustCompile(`^([+-]?)(\d+)d$`)
	matches := re.FindStringSubmatch(date)
	if len(matches) != 3 {
		return false, 0
	}

	days, err := strconv.Atoi(matches[2])
	if err != nil {
		return false, 0
	}

	if matches[1] == "-" {
		days = -days
	}

	return true, days
}

func parseRelativeWeeks(date string) (bool, int) {
	re := regexp.MustCompile(`^([+-]?)(\d+)w$`)
	matches := re.FindStringSubmatch(date)
	if len(matches) != 3 {
		return false, 0
	}

	weeks, err := strconv.Atoi(matches[2])
	if err != nil {
		return false, 0
	}

	if matches[1] == "-" {
		weeks = -weeks
	}

	return true, weeks
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

func ParseDuration(duration string) (time.Duration, error) {
	if duration == "" {
		return time.Hour, nil // default 1 hour
	}
	return time.ParseDuration(duration)
}
