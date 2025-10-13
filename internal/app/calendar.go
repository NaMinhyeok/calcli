package app

import (
	"io"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/render"
)

// EventReader interface for reading events
type EventReader interface {
	ListEvents() ([]domain.Event, error)
}

// CalendarHandler displays a month calendar view with events
func CalendarHandler(reader EventReader, output io.Writer, date *time.Time) error {
	// Default to current month if no date provided
	targetDate := time.Now()
	if date != nil {
		targetDate = *date
	}

	// Get all events for the month
	firstDay := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Second)

	events, err := reader.ListEvents()
	if err != nil {
		return err
	}

	// Filter events for this month
	var monthEvents []domain.Event
	for _, event := range events {
		if !event.Start.Before(firstDay) && !event.Start.After(lastDay) {
			monthEvents = append(monthEvents, event)
		}
	}

	// Expand recurring events within the month
	var expandedEvents []domain.Event
	for _, event := range monthEvents {
		instances := domain.ExpandRecurrence(event, firstDay, lastDay)
		expandedEvents = append(expandedEvents, instances...)
	}

	// Render calendar
	view := render.NewMonthView(targetDate, expandedEvents)
	return view.RenderWithEvents(output)
}
