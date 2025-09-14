package app

import (
	"crypto/rand"
	"fmt"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/util"
)

type EventCreator interface {
	CreateEvent(event domain.Event) error
}

type UIDGenerator interface {
	Generate() (string, error)
}

type RealUIDGenerator struct{}

func (g *RealUIDGenerator) Generate() (string, error) {
	return generateUID()
}

func NewHandler(creator EventCreator, timeProvider util.TimeProvider, uidGen UIDGenerator, title, when, duration, location string) error {
	startTime, err := util.ParseTime(when, timeProvider)
	if err != nil {
		return fmt.Errorf("invalid time format: %v", err)
	}

	dur, err := util.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	uid, err := uidGen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate UID: %v", err)
	}

	event := domain.Event{
		UID:      uid,
		Summary:  title,
		Start:    startTime,
		End:      startTime.Add(dur),
		Location: location,
		AllDay:   false,
	}

	return creator.CreateEvent(event)
}

func NewHandlerWithRecurrence(creator EventCreator, timeProvider util.TimeProvider, uidGen UIDGenerator, title, when, duration, location, repeat string, count int, until string) error {
	startTime, err := util.ParseTime(when, timeProvider)
	if err != nil {
		return fmt.Errorf("invalid time format: %v", err)
	}

	dur, err := util.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	uid, err := uidGen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate UID: %v", err)
	}

	event := domain.Event{
		UID:      uid,
		Summary:  title,
		Start:    startTime,
		End:      startTime.Add(dur),
		Location: location,
		AllDay:   false,
	}

	if repeat != "" {
		recurrence, err := parseRecurrenceOptions(repeat, count, until)
		if err != nil {
			return fmt.Errorf("invalid recurrence options: %v", err)
		}
		event.Recurrence = recurrence
	}

	return creator.CreateEvent(event)
}

func parseRecurrenceOptions(repeat string, count int, until string) (*domain.Recurrence, error) {
	var frequency string
	switch repeat {
	case "daily":
		frequency = "DAILY"
	case "weekly":
		frequency = "WEEKLY"
	case "monthly":
		frequency = "MONTHLY"
	case "yearly":
		frequency = "YEARLY"
	default:
		return nil, fmt.Errorf("unsupported repeat pattern: %s. Supported: daily, weekly, monthly, yearly", repeat)
	}

	recurrence := &domain.Recurrence{
		Frequency: frequency,
		Interval:  1,
	}

	if count > 0 && until != "" {
		return nil, fmt.Errorf("cannot specify both count and until date")
	}

	if count > 0 {
		recurrence.Count = &count
	}

	if until != "" {
		untilTime, err := util.ParseDate(until)
		if err != nil {
			return nil, fmt.Errorf("invalid until date: %v", err)
		}
		recurrence.Until = &untilTime
	}

	return recurrence, nil
}

func generateUID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("calcli-%x", b), nil
}
