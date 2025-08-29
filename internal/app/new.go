package app

import (
	"crypto/rand"
	"fmt"
	"time"

	"calcli/internal/domain"
)

type EventCreator interface {
	CreateEvent(event domain.Event) error
}

type TimeProvider interface {
	Now() time.Time
}

type UIDGenerator interface {
	Generate() (string, error)
}

type RealTimeProvider struct{}

func (t *RealTimeProvider) Now() time.Time {
	return time.Now()
}

type RealUIDGenerator struct{}

func (g *RealUIDGenerator) Generate() (string, error) {
	return generateUID()
}

func NewHandler(creator EventCreator, timeProvider TimeProvider, uidGen UIDGenerator, title, when, duration string) error {
	startTime, err := parseTime(when, timeProvider)
	if err != nil {
		return fmt.Errorf("invalid time format: %v", err)
	}

	dur, err := parseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	uid, err := uidGen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate UID: %v", err)
	}

	// Create event
	event := domain.Event{
		UID:     uid,
		Summary: title,
		Start:   startTime,
		End:     startTime.Add(dur),
		AllDay:  false,
	}

	return creator.CreateEvent(event)
}

func parseTime(when string, timeProvider TimeProvider) (time.Time, error) {
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

func parseDuration(duration string) (time.Duration, error) {
	if duration == "" {
		return time.Hour, nil // default 1 hour
	}
	return time.ParseDuration(duration)
}

func generateUID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("calcli-%x", b), nil
}
