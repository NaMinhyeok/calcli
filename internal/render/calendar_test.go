package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

func TestMonthView_Render(t *testing.T) {
	date := time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC)
	events := []domain.Event{
		{
			UID:     "event1",
			Summary: "Meeting",
			Start:   time.Date(2025, time.September, 10, 14, 0, 0, 0, time.UTC),
			End:     time.Date(2025, time.September, 10, 15, 0, 0, 0, time.UTC),
		},
		{
			UID:     "event2",
			Summary: "Lunch",
			Start:   time.Date(2025, time.September, 15, 12, 0, 0, 0, time.UTC),
			End:     time.Date(2025, time.September, 15, 13, 0, 0, 0, time.UTC),
		},
	}

	view := NewMonthView(date, events)
	view.Today = time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC) // Set known today

	var buf bytes.Buffer
	err := view.Render(&buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "September 2025") {
		t.Error("Expected month/year header")
	}

	// Check weekday headers
	if !strings.Contains(output, "Su") || !strings.Contains(output, "Mo") {
		t.Error("Expected weekday headers")
	}

	// Check that days are present
	if !strings.Contains(output, "10") || !strings.Contains(output, "15") {
		t.Error("Expected day numbers")
	}
}

func TestMonthView_RenderWithEvents(t *testing.T) {
	date := time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC)
	events := []domain.Event{
		{
			UID:     "event1",
			Summary: "Morning standup",
			Start:   time.Date(2025, time.September, 5, 9, 0, 0, 0, time.UTC),
			End:     time.Date(2025, time.September, 5, 9, 30, 0, 0, time.UTC),
		},
		{
			UID:     "event2",
			Summary: "Team lunch",
			Start:   time.Date(2025, time.September, 5, 12, 0, 0, 0, time.UTC),
			End:     time.Date(2025, time.September, 5, 13, 0, 0, 0, time.UTC),
		},
	}

	view := NewMonthView(date, events)
	view.Today = time.Date(2025, time.September, 10, 0, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	err := view.RenderWithEvents(&buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Check calendar is rendered
	if !strings.Contains(output, "September 2025") {
		t.Error("Expected calendar header")
	}

	// Check events section
	if !strings.Contains(output, "Events:") {
		t.Error("Expected events section")
	}

	// Check event details
	if !strings.Contains(output, "Morning standup") {
		t.Error("Expected first event")
	}

	if !strings.Contains(output, "Team lunch") {
		t.Error("Expected second event")
	}

	// Check date grouping
	if !strings.Contains(output, "2025-09-05") {
		t.Error("Expected date grouping")
	}
}

func TestMonthView_NoEvents(t *testing.T) {
	date := time.Date(2025, time.October, 1, 0, 0, 0, 0, time.UTC)
	events := []domain.Event{}

	view := NewMonthView(date, events)

	var buf bytes.Buffer
	err := view.RenderWithEvents(&buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Should show calendar
	if !strings.Contains(output, "October 2025") {
		t.Error("Expected calendar header")
	}

	// Should show no events message
	if !strings.Contains(output, "No events this month") {
		t.Error("Expected no events message")
	}
}

func TestMonthView_EventMarkers(t *testing.T) {
	date := time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC)
	events := []domain.Event{
		{
			UID:     "event1",
			Summary: "Event on 5th",
			Start:   time.Date(2025, time.September, 5, 10, 0, 0, 0, time.UTC),
			End:     time.Date(2025, time.September, 5, 11, 0, 0, 0, time.UTC),
		},
	}

	view := NewMonthView(date, events)
	view.Today = time.Date(2025, time.September, 10, 0, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	err := view.Render(&buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Check for today marker (*)
	lines := strings.Split(output, "\n")
	foundToday := false
	for _, line := range lines {
		if strings.Contains(line, "10*") {
			foundToday = true
			break
		}
	}

	if !foundToday {
		t.Error("Expected today marker (*) on day 10")
	}
}

func TestDateKey(t *testing.T) {
	date := time.Date(2025, time.September, 5, 14, 30, 0, 0, time.UTC)
	key := dateKey(date)

	expected := "2025-09-05"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}
}

func TestMonthView_DifferentMonths(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month time.Month
	}{
		{"January", 2025, time.January},
		{"February", 2025, time.February},
		{"December", 2025, time.December},
		{"February Leap Year", 2024, time.February},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := time.Date(tt.year, tt.month, 1, 0, 0, 0, 0, time.UTC)
			view := NewMonthView(date, nil)

			var buf bytes.Buffer
			err := view.Render(&buf)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.month.String()) {
				t.Errorf("Expected month %s in output", tt.month)
			}
		})
	}
}
