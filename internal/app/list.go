package app

import (
	"fmt"
	"io"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

type EventLister interface {
	ListEvents() ([]domain.Event, error)
}

type EventFormatter interface {
	FormatEvents([]domain.Event, io.Writer)
}

type HardcodedEventLister struct{}

func (h HardcodedEventLister) ListEvents() ([]domain.Event, error) {
	events := []domain.Event{
		{
			UID:         "sample-1",
			Summary:     "Team Meeting",
			Description: "Weekly team sync",
			Start:       time.Date(2025, 8, 27, 10, 0, 0, 0, time.Local),
			End:         time.Date(2025, 8, 27, 11, 0, 0, 0, time.Local),
			Location:    "Conference Room A",
			Categories:  []string{"work"},
			AllDay:      false,
			Calendar:    "work",
		},
		{
			UID:      "sample-2",
			Summary:  "Lunch with John",
			Start:    time.Date(2025, 8, 27, 12, 30, 0, 0, time.Local),
			End:      time.Date(2025, 8, 27, 13, 30, 0, 0, time.Local),
			Location: "Downtown Cafe",
			AllDay:   false,
			Calendar: "personal",
		},
	}
	return events, nil
}

type SimpleEventFormatter struct {
	ShowUID bool
}

func (s SimpleEventFormatter) FormatEvents(events []domain.Event, w io.Writer) {
	for _, event := range events {
		fmt.Fprintf(w, "%s - %s %s",
			event.Start.Format("15:04"),
			event.End.Format("15:04"),
			event.Summary,
		)
		if s.ShowUID {
			fmt.Fprintf(w, " [%s]", event.UID)
		}
		fmt.Fprintf(w, "\n")

		if event.Location != "" {
			fmt.Fprintf(w, "  @ %s\n", event.Location)
		}
	}
}

func ListHandler(lister EventLister, formatter EventFormatter, w io.Writer, from, to *time.Time) error {
	events, err := lister.ListEvents()
	if err != nil {
		return err
	}

	var filteredEvents []domain.Event
	for _, event := range events {
		if event.Recurrence != nil {
			expandedEvents := expandRecurringEvent(event, from, to)
			for _, expandedEvent := range expandedEvents {
				if shouldIncludeEvent(expandedEvent, from, to) {
					filteredEvents = append(filteredEvents, expandedEvent)
				}
			}
		} else {
			if shouldIncludeEvent(event, from, to) {
				filteredEvents = append(filteredEvents, event)
			}
		}
	}

	formatter.FormatEvents(filteredEvents, w)
	return nil
}

func shouldIncludeEvent(event domain.Event, from, to *time.Time) bool {
	if from == nil && to == nil {
		return true
	}

	eventStart := event.Start
	eventEnd := event.End

	if from != nil && eventEnd.Before(*from) {
		return false
	}

	if to != nil && eventStart.After(*to) {
		return false
	}

	return true
}

func expandRecurringEvent(event domain.Event, from, to *time.Time) []domain.Event {
	if event.Recurrence == nil {
		return []domain.Event{event}
	}

	var expandedEvents []domain.Event
	current := event.Start
	duration := event.End.Sub(event.Start)
	count := 0

	for {
		if event.Recurrence.Count != nil && count >= *event.Recurrence.Count {
			break
		}

		if event.Recurrence.Until != nil && current.After(*event.Recurrence.Until) {
			break
		}

		if to != nil && current.After(*to) {
			break
		}

		if from == nil || !current.Add(duration).Before(*from) {
			expandedEvent := event
			expandedEvent.Start = current
			expandedEvent.End = current.Add(duration)
			expandedEvent.Recurrence = nil
			expandedEvents = append(expandedEvents, expandedEvent)
		}

		switch event.Recurrence.Frequency {
		case "DAILY":
			current = current.AddDate(0, 0, event.Recurrence.Interval)
		case "WEEKLY":
			current = current.AddDate(0, 0, 7*event.Recurrence.Interval)
		case "MONTHLY":
			current = current.AddDate(0, event.Recurrence.Interval, 0)
		case "YEARLY":
			current = current.AddDate(event.Recurrence.Interval, 0, 0)
		default:
			break
		}

		count++

		if count > 1000 {
			break
		}
	}

	return expandedEvents
}
