package domain

import (
	"testing"
	"time"
)

func TestExpandRecurrence_Daily(t *testing.T) {
	start := time.Date(2025, 9, 1, 10, 0, 0, 0, time.UTC)
	event := Event{
		UID:     "daily-event",
		Summary: "Daily standup",
		Start:   start,
		End:     start.Add(30 * time.Minute),
		Recurrence: &Recurrence{
			Frequency: "DAILY",
			Interval:  1,
		},
	}

	rangeStart := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2025, 9, 5, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	if len(instances) != 5 {
		t.Errorf("Expected 5 instances, got %d", len(instances))
	}

	for i, instance := range instances {
		expectedStart := start.AddDate(0, 0, i)
		if !instance.Start.Equal(expectedStart) {
			t.Errorf("Instance %d: expected start %v, got %v", i, expectedStart, instance.Start)
		}
	}
}

func TestExpandRecurrence_Weekly(t *testing.T) {
	start := time.Date(2025, 9, 1, 14, 0, 0, 0, time.UTC)
	count := 3
	event := Event{
		UID:     "weekly-event",
		Summary: "Weekly meeting",
		Start:   start,
		End:     start.Add(1 * time.Hour),
		Recurrence: &Recurrence{
			Frequency: "WEEKLY",
			Interval:  1,
			Count:     &count,
		},
	}

	rangeStart := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	if len(instances) != 3 {
		t.Errorf("Expected 3 instances (COUNT=3), got %d", len(instances))
	}

	expectedDates := []time.Time{
		time.Date(2025, 9, 1, 14, 0, 0, 0, time.UTC),
		time.Date(2025, 9, 8, 14, 0, 0, 0, time.UTC),
		time.Date(2025, 9, 15, 14, 0, 0, 0, time.UTC),
	}

	for i, instance := range instances {
		if !instance.Start.Equal(expectedDates[i]) {
			t.Errorf("Instance %d: expected start %v, got %v", i, expectedDates[i], instance.Start)
		}
	}
}

func TestExpandRecurrence_Monthly(t *testing.T) {
	start := time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC)
	event := Event{
		UID:     "monthly-event",
		Summary: "Monthly review",
		Start:   start,
		End:     start.Add(2 * time.Hour),
		Recurrence: &Recurrence{
			Frequency: "MONTHLY",
			Interval:  1,
		},
	}

	rangeStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2025, 4, 30, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	if len(instances) != 4 {
		t.Errorf("Expected 4 instances (Jan-Apr), got %d", len(instances))
	}

	expectedMonths := []time.Month{time.January, time.February, time.March, time.April}
	for i, instance := range instances {
		if instance.Start.Month() != expectedMonths[i] {
			t.Errorf("Instance %d: expected month %v, got %v", i, expectedMonths[i], instance.Start.Month())
		}
		if instance.Start.Day() != 15 {
			t.Errorf("Instance %d: expected day 15, got %d", i, instance.Start.Day())
		}
	}
}

func TestExpandRecurrence_UntilDate(t *testing.T) {
	start := time.Date(2025, 9, 1, 10, 0, 0, 0, time.UTC)
	until := time.Date(2025, 9, 10, 23, 59, 59, 0, time.UTC)
	event := Event{
		UID:     "until-event",
		Summary: "Event with UNTIL",
		Start:   start,
		End:     start.Add(1 * time.Hour),
		Recurrence: &Recurrence{
			Frequency: "DAILY",
			Interval:  1,
			Until:     &until,
		},
	}

	rangeStart := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	// Should generate instances from Sept 1 to Sept 10 (10 days)
	if len(instances) != 10 {
		t.Errorf("Expected 10 instances (UNTIL Sept 10), got %d", len(instances))
	}

	for _, instance := range instances {
		if instance.Start.After(until) {
			t.Errorf("Instance start %v is after UNTIL %v", instance.Start, until)
		}
	}
}

func TestExpandRecurrence_NoRecurrence(t *testing.T) {
	start := time.Date(2025, 9, 1, 10, 0, 0, 0, time.UTC)
	event := Event{
		UID:        "single-event",
		Summary:    "Non-recurring event",
		Start:      start,
		End:        start.Add(1 * time.Hour),
		Recurrence: nil,
	}

	rangeStart := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2025, 9, 30, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	if len(instances) != 1 {
		t.Errorf("Expected 1 instance for non-recurring event, got %d", len(instances))
	}

	if !instances[0].Start.Equal(start) {
		t.Errorf("Expected start %v, got %v", start, instances[0].Start)
	}
}

func TestExpandRecurrence_OutsideRange(t *testing.T) {
	start := time.Date(2025, 9, 1, 10, 0, 0, 0, time.UTC)
	event := Event{
		UID:     "daily-event",
		Summary: "Daily standup",
		Start:   start,
		End:     start.Add(30 * time.Minute),
		Recurrence: &Recurrence{
			Frequency: "DAILY",
			Interval:  1,
		},
	}

	// Range that doesn't include the event
	rangeStart := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2025, 10, 5, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	if len(instances) != 5 {
		t.Errorf("Expected 5 instances in October range, got %d", len(instances))
	}
}

func TestExpandRecurrence_Yearly(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	event := Event{
		UID:     "yearly-event",
		Summary: "New Year",
		Start:   start,
		End:     start.Add(24 * time.Hour),
		Recurrence: &Recurrence{
			Frequency: "YEARLY",
			Interval:  1,
		},
	}

	rangeStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2028, 12, 31, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	if len(instances) != 4 {
		t.Errorf("Expected 4 instances (2025-2028), got %d", len(instances))
	}

	expectedYears := []int{2025, 2026, 2027, 2028}
	for i, instance := range instances {
		if instance.Start.Year() != expectedYears[i] {
			t.Errorf("Instance %d: expected year %d, got %d", i, expectedYears[i], instance.Start.Year())
		}
	}
}

func TestExpandRecurrence_IntervalGreaterThanOne(t *testing.T) {
	start := time.Date(2025, 9, 1, 10, 0, 0, 0, time.UTC)
	event := Event{
		UID:     "bi-weekly-event",
		Summary: "Bi-weekly meeting",
		Start:   start,
		End:     start.Add(1 * time.Hour),
		Recurrence: &Recurrence{
			Frequency: "WEEKLY",
			Interval:  2, // Every 2 weeks
		},
	}

	rangeStart := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC)

	instances := ExpandRecurrence(event, rangeStart, rangeEnd)

	// Sept 1, 15, 29, Oct 13, 27 = 5 instances
	if len(instances) < 4 {
		t.Errorf("Expected at least 4 instances for bi-weekly over 2 months, got %d", len(instances))
	}

	// Check that instances are 14 days apart
	for i := 1; i < len(instances); i++ {
		daysDiff := instances[i].Start.Sub(instances[i-1].Start).Hours() / 24
		if daysDiff != 14 {
			t.Errorf("Expected 14 days between instances, got %.0f days", daysDiff)
		}
	}
}
