package app

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"calcli/internal/domain"
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

			err := ListHandler(lister, formatter, &buf)

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
