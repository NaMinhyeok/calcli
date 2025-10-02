package domain

import (
	"time"
)

// ExpandRecurrence generates instances of a recurring event within a time range
func ExpandRecurrence(event Event, rangeStart, rangeEnd time.Time) []Event {
	if event.Recurrence == nil {
		return []Event{event}
	}

	var instances []Event
	current := event.Start

	// Maximum iterations to prevent infinite loops
	const maxIterations = 5000

	for i := 0; i < maxIterations; i++ {
		// Check if we've passed the range end
		if current.After(rangeEnd) {
			break
		}

		// Check UNTIL condition
		if event.Recurrence.Until != nil && current.After(*event.Recurrence.Until) {
			break
		}

		// Check COUNT condition
		if event.Recurrence.Count != nil && i >= *event.Recurrence.Count {
			break
		}

		// If instance is within range, add it
		if !current.Before(rangeStart) && !current.After(rangeEnd) {
			instance := event
			instance.Start = current
			instance.End = current.Add(event.Duration())
			instance.Recurrence = nil // Expanded instances are not recurring
			instances = append(instances, instance)
		}

		// Move to next occurrence
		current = nextOccurrence(current, event.Recurrence)
	}

	return instances
}

func nextOccurrence(current time.Time, rec *Recurrence) time.Time {
	interval := rec.Interval
	if interval == 0 {
		interval = 1
	}

	switch rec.Frequency {
	case "DAILY":
		return current.AddDate(0, 0, interval)
	case "WEEKLY":
		return current.AddDate(0, 0, 7*interval)
	case "MONTHLY":
		return current.AddDate(0, interval, 0)
	case "YEARLY":
		return current.AddDate(interval, 0, 0)
	default:
		return current.AddDate(0, 0, interval)
	}
}
