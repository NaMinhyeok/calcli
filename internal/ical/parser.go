package ical

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"

	"github.com/arran4/golang-ical"
)

func ParseEvents(r io.Reader) ([]domain.Event, error) {
	cal, err := ics.ParseCalendar(r)
	if err != nil {
		return nil, err
	}

	var events []domain.Event
	for _, event := range cal.Events() {
		domainEvent := convertToDomainEvent(event)
		events = append(events, domainEvent)
	}

	return events, nil
}

func convertToDomainEvent(event *ics.VEvent) domain.Event {
	domainEvent := domain.Event{}

	if uid := event.GetProperty(ics.ComponentProperty(ics.PropertyUid)); uid != nil {
		domainEvent.UID = uid.Value
	}

	if summary := event.GetProperty(ics.ComponentProperty(ics.PropertySummary)); summary != nil {
		domainEvent.Summary = summary.Value
	}

	if desc := event.GetProperty(ics.ComponentProperty(ics.PropertyDescription)); desc != nil {
		domainEvent.Description = desc.Value
	}

	if location := event.GetProperty(ics.ComponentProperty(ics.PropertyLocation)); location != nil {
		domainEvent.Location = location.Value
	}

	if startTime, err := event.GetStartAt(); err == nil {
		domainEvent.Start = startTime
	}

	if endTime, err := event.GetEndAt(); err == nil {
		domainEvent.End = endTime
	}

	if startTime, err := event.GetStartAt(); err == nil {
		domainEvent.AllDay = startTime.Hour() == 0 && startTime.Minute() == 0 && startTime.Second() == 0
	}

	if rrule := event.GetProperty(ics.ComponentProperty(ics.PropertyRrule)); rrule != nil {
		domainEvent.Recurrence = parseRRULE(rrule.Value)
	}

	return domainEvent
}

func parseRRULE(rruleStr string) *domain.Recurrence {
	if rruleStr == "" {
		return nil
	}

	parts := strings.Split(rruleStr, ";")
	recurrence := &domain.Recurrence{
		Interval: 1,
	}

	for _, part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			continue
		}
		key, value := keyValue[0], keyValue[1]

		switch key {
		case "FREQ":
			recurrence.Frequency = value
		case "INTERVAL":
			if interval, err := strconv.Atoi(value); err == nil {
				recurrence.Interval = interval
			}
		case "COUNT":
			if count, err := strconv.Atoi(value); err == nil {
				recurrence.Count = &count
			}
		case "UNTIL":
			if until, err := time.Parse("20060102T150405Z", value); err == nil {
				recurrence.Until = &until
			}
		}
	}

	return recurrence
}
