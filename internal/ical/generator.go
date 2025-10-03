package ical

import (
	"fmt"
	"io"

	"github.com/NaMinhyeok/calcli/internal/domain"

	"github.com/arran4/golang-ical"
)

func GenerateEvent(event domain.Event, w io.Writer) error {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)

	vevent := cal.AddEvent(event.UID)
	vevent.SetSummary(event.Summary)

	if event.Description != "" {
		vevent.SetDescription(event.Description)
	}

	if event.Location != "" {
		vevent.SetLocation(event.Location)
	}

	// Set times
	if event.AllDay {
		vevent.SetAllDayStartAt(event.Start)
		vevent.SetAllDayEndAt(event.End)
	} else {
		vevent.SetStartAt(event.Start)
		vevent.SetEndAt(event.End)
	}

	// Set recurrence rule if present
	if event.Recurrence != nil {
		rrule := buildRRULE(event.Recurrence)
		vevent.SetProperty(ics.ComponentProperty(ics.PropertyRrule), rrule)
	}

	// Write to output
	return cal.SerializeTo(w)
}

func buildRRULE(rec *domain.Recurrence) string {
	if rec == nil {
		return ""
	}

	rrule := fmt.Sprintf("FREQ=%s", rec.Frequency)

	if rec.Interval > 1 {
		rrule += fmt.Sprintf(";INTERVAL=%d", rec.Interval)
	}

	if rec.Count != nil {
		rrule += fmt.Sprintf(";COUNT=%d", *rec.Count)
	}

	if rec.Until != nil {
		rrule += fmt.Sprintf(";UNTIL=%s", rec.Until.Format("20060102T150405Z"))
	}

	return rrule
}
