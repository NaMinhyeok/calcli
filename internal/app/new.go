package app

import (
	"crypto/rand"
	"fmt"
	"time"

	"calcli/internal/domain"
	"calcli/internal/util"
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

	dur, err := parseDuration(duration)
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
