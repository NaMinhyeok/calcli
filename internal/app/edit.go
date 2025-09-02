package app

import (
	"fmt"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/util"
)

type EventEditor interface {
	FindEventByUID(uid string) (domain.Event, error)
	UpdateEvent(event domain.Event) error
}

type EditOptions struct {
	Title    *string
	When     *string
	Duration *string
	Location *string
}

func EditHandler(editor EventEditor, timeProvider util.TimeProvider, uid string, options EditOptions) error {
	// Basic UID validation
	if len(uid) < 3 {
		return fmt.Errorf("UID '%s' is too short. UIDs should be at least 3 characters long", uid)
	}

	event, err := editor.FindEventByUID(uid)
	if err != nil {
		return fmt.Errorf("no event found with UID '%s'. Use 'calcli list --show-uid' or 'calcli search --show-uid <query>' to find valid UIDs", uid)
	}

	if options.Title != nil {
		event.Summary = *options.Title
	}

	if options.When != nil {
		startTime, err := util.ParseTime(*options.When, timeProvider)
		if err != nil {
			return fmt.Errorf("invalid time format: %v", err)
		}

		duration := event.End.Sub(event.Start)
		event.Start = startTime
		event.End = startTime.Add(duration)
	}

	if options.Duration != nil {
		dur, err := util.ParseDuration(*options.Duration)
		if err != nil {
			return fmt.Errorf("invalid duration: %v", err)
		}
		event.End = event.Start.Add(dur)
	}

	if options.Location != nil {
		event.Location = *options.Location
	}

	if err := editor.UpdateEvent(event); err != nil {
		return fmt.Errorf("failed to update event: %v", err)
	}

	return nil
}
