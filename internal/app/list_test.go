package app

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

// FakeEventLister provides configurable test data
type FakeEventLister struct {
	events []domain.Event
	err    error
}

func (f FakeEventLister) ListEvents() ([]domain.Event, error) {
	return f.events, f.err
}

func TestListHandler(t *testing.T) {
	tests := []struct {
		name           string
		events         []domain.Event
		expectedOutput string
		expectError    bool
	}{
		{
			name: "single event with location",
			events: []domain.Event{
				{
					UID:      "test-1",
					Summary:  "Test Meeting",
					Start:    time.Date(2025, 8, 27, 14, 0, 0, 0, time.Local),
					End:      time.Date(2025, 8, 27, 15, 0, 0, 0, time.Local),
					Location: "Test Room",
					AllDay:   false,
				},
			},
			expectedOutput: "14:00 - 15:00 Test Meeting\n  @ Test Room\n",
			expectError:    false,
		},
		{
			name: "event without location",
			events: []domain.Event{
				{
					UID:     "test-2",
					Summary: "Quick Call",
					Start:   time.Date(2025, 8, 27, 16, 0, 0, 0, time.Local),
					End:     time.Date(2025, 8, 27, 16, 30, 0, 0, time.Local),
					AllDay:  false,
				},
			},
			expectedOutput: "16:00 - 16:30 Quick Call\n",
			expectError:    false,
		},
		{
			name:           "empty events list",
			events:         []domain.Event{},
			expectedOutput: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			lister := FakeEventLister{events: tt.events}
			formatter := &SimpleEventFormatter{}

			err := ListHandler(lister, formatter, &buf, nil, nil)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			output := buf.String()
			if output != tt.expectedOutput {
				t.Errorf("expected output:\n%q\ngot:\n%q", tt.expectedOutput, output)
			}
		})
	}
}

func TestHardcodedEventLister(t *testing.T) {
	lister := &HardcodedEventLister{}
	events, err := lister.ListEvents()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// Check first event
	if events[0].Summary != "Team Meeting" {
		t.Errorf("expected 'Team Meeting', got %s", events[0].Summary)
	}
	if events[0].Location != "Conference Room A" {
		t.Errorf("expected 'Conference Room A', got %s", events[0].Location)
	}
}

func TestSimpleEventFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := &SimpleEventFormatter{}

	events := []domain.Event{
		{
			Summary:  "Test Event",
			Start:    time.Date(2025, 8, 27, 9, 0, 0, 0, time.Local),
			End:      time.Date(2025, 8, 27, 10, 0, 0, 0, time.Local),
			Location: "Test Location",
		},
	}

	formatter.FormatEvents(events, &buf)

	output := buf.String()
	if !strings.Contains(output, "09:00 - 10:00 Test Event") {
		t.Errorf("output should contain event details, got: %s", output)
	}
	if !strings.Contains(output, "@ Test Location") {
		t.Errorf("output should contain location, got: %s", output)
	}
}

func TestListHandlerWithDateRange(t *testing.T) {
	events := []domain.Event{
		{
			UID:     "1",
			Summary: "Past Event",
			Start:   time.Date(2025, 8, 25, 10, 0, 0, 0, time.Local),
			End:     time.Date(2025, 8, 25, 11, 0, 0, 0, time.Local),
		},
		{
			UID:     "2",
			Summary: "Current Event",
			Start:   time.Date(2025, 8, 30, 14, 0, 0, 0, time.Local),
			End:     time.Date(2025, 8, 30, 15, 0, 0, 0, time.Local),
		},
		{
			UID:     "3",
			Summary: "Future Event",
			Start:   time.Date(2025, 9, 5, 16, 0, 0, 0, time.Local),
			End:     time.Date(2025, 9, 5, 17, 0, 0, 0, time.Local),
		},
	}

	tests := []struct {
		name     string
		from     *time.Time
		to       *time.Time
		expected []string
	}{
		{
			name:     "no filter shows all",
			from:     nil,
			to:       nil,
			expected: []string{"Past Event", "Current Event", "Future Event"},
		},
		{
			name:     "from date filters past events",
			from:     timePtr(time.Date(2025, 8, 29, 0, 0, 0, 0, time.Local)),
			to:       nil,
			expected: []string{"Current Event", "Future Event"},
		},
		{
			name:     "to date filters future events",
			from:     nil,
			to:       timePtr(time.Date(2025, 9, 1, 0, 0, 0, 0, time.Local)),
			expected: []string{"Past Event", "Current Event"},
		},
		{
			name:     "from and to date range",
			from:     timePtr(time.Date(2025, 8, 29, 0, 0, 0, 0, time.Local)),
			to:       timePtr(time.Date(2025, 9, 1, 0, 0, 0, 0, time.Local)),
			expected: []string{"Current Event"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lister := FakeEventLister{events: events}
			formatter := &SimpleEventFormatter{}
			var buf bytes.Buffer

			err := ListHandler(lister, formatter, &buf, tt.from, tt.to)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			output := buf.String()
			for _, expectedTitle := range tt.expected {
				if !strings.Contains(output, expectedTitle) {
					t.Errorf("expected output to contain %q, got %q", expectedTitle, output)
				}
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestExpandRecurringEvent(t *testing.T) {
	baseEvent := domain.Event{
		UID:     "recurring-1",
		Summary: "Daily Standup",
		Start:   time.Date(2025, 9, 1, 9, 0, 0, 0, time.Local),
		End:     time.Date(2025, 9, 1, 9, 30, 0, 0, time.Local),
	}

	tests := []struct {
		name           string
		recurrence     *domain.Recurrence
		from           *time.Time
		to             *time.Time
		expectedCount  int
		checkFirstLast func(t *testing.T, events []domain.Event)
	}{
		{
			name: "daily recurrence with count",
			recurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
				Count:     intPtr(3),
			},
			from:          nil,
			to:            nil,
			expectedCount: 3,
			checkFirstLast: func(t *testing.T, events []domain.Event) {
				if events[0].Start.Day() != 1 {
					t.Errorf("first event should be on day 1, got %d", events[0].Start.Day())
				}
				if events[2].Start.Day() != 3 {
					t.Errorf("third event should be on day 3, got %d", events[2].Start.Day())
				}
			},
		},
		{
			name: "weekly recurrence with interval 2",
			recurrence: &domain.Recurrence{
				Frequency: "WEEKLY",
				Interval:  2,
				Count:     intPtr(3),
			},
			from:          nil,
			to:            nil,
			expectedCount: 3,
			checkFirstLast: func(t *testing.T, events []domain.Event) {
				if events[0].Start.Day() != 1 {
					t.Errorf("first event should be on day 1, got %d", events[0].Start.Day())
				}
				if events[1].Start.Day() != 15 {
					t.Errorf("second event should be on day 15, got %d", events[1].Start.Day())
				}
				if events[2].Start.Day() != 29 {
					t.Errorf("third event should be on day 29, got %d", events[2].Start.Day())
				}
			},
		},
		{
			name: "daily recurrence with until date",
			recurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
				Until:     timePtr(time.Date(2025, 9, 3, 23, 59, 59, 0, time.Local)),
			},
			from:          nil,
			to:            nil,
			expectedCount: 3,
			checkFirstLast: func(t *testing.T, events []domain.Event) {
				if events[0].Start.Day() != 1 {
					t.Errorf("first event should be on day 1, got %d", events[0].Start.Day())
				}
				if events[2].Start.Day() != 3 {
					t.Errorf("last event should be on day 3, got %d", events[2].Start.Day())
				}
			},
		},
		{
			name: "daily recurrence with date range filter",
			recurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
				Count:     intPtr(10),
			},
			from:          timePtr(time.Date(2025, 9, 3, 0, 0, 0, 0, time.Local)),
			to:            timePtr(time.Date(2025, 9, 5, 23, 59, 59, 0, time.Local)),
			expectedCount: 3,
			checkFirstLast: func(t *testing.T, events []domain.Event) {
				if events[0].Start.Day() != 3 {
					t.Errorf("first event should be on day 3, got %d", events[0].Start.Day())
				}
				if events[2].Start.Day() != 5 {
					t.Errorf("last event should be on day 5, got %d", events[2].Start.Day())
				}
			},
		},
		{
			name:          "non-recurring event",
			recurrence:    nil,
			from:          nil,
			to:            nil,
			expectedCount: 1,
			checkFirstLast: func(t *testing.T, events []domain.Event) {
				if events[0].Start.Day() != 1 {
					t.Errorf("event should be on day 1, got %d", events[0].Start.Day())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent
			event.Recurrence = tt.recurrence

			rangeStart := time.Time{}
			rangeEnd := time.Now().AddDate(10, 0, 0)

			if tt.from != nil {
				rangeStart = *tt.from
			}
			if tt.to != nil {
				rangeEnd = *tt.to
			}

			expandedEvents := domain.ExpandRecurrence(event, rangeStart, rangeEnd)

			if len(expandedEvents) != tt.expectedCount {
				t.Errorf("expected %d events, got %d", tt.expectedCount, len(expandedEvents))
				return
			}

			for _, expandedEvent := range expandedEvents {
				if expandedEvent.Recurrence != nil {
					t.Error("expanded events should not have recurrence")
				}
				if expandedEvent.Summary != baseEvent.Summary {
					t.Errorf("expanded event should preserve summary, got %s", expandedEvent.Summary)
				}
				duration := expandedEvent.End.Sub(expandedEvent.Start)
				expectedDuration := baseEvent.End.Sub(baseEvent.Start)
				if duration != expectedDuration {
					t.Errorf("expanded event should preserve duration, got %v", duration)
				}
			}

			if tt.checkFirstLast != nil {
				tt.checkFirstLast(t, expandedEvents)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
