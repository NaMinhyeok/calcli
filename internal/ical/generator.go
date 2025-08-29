package ical

import (
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

	// Write to output
	return cal.SerializeTo(w)
}
