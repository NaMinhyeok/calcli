package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

type FakeEventEditor struct {
	events map[string]domain.Event
	err    error
}

func NewFakeEventEditor() *FakeEventEditor {
	return &FakeEventEditor{
		events: make(map[string]domain.Event),
	}
}

func (f *FakeEventEditor) AddEvent(event domain.Event) {
	f.events[event.UID] = event
}

func (f *FakeEventEditor) FindEventByUID(uid string) (domain.Event, error) {
	if f.err != nil {
		return domain.Event{}, f.err
	}

	event, exists := f.events[uid]
	if !exists {
		return domain.Event{}, fmt.Errorf("event with UID %s not found", uid)
	}

	return event, nil
}

func (f *FakeEventEditor) UpdateEvent(event domain.Event) error {
	if f.err != nil {
		return f.err
	}

	f.events[event.UID] = event
	return nil
}

func TestEditHandler(t *testing.T) {
	originalEvent := domain.Event{
		UID:      "test-1",
		Summary:  "Original Title",
		Start:    time.Date(2025, 9, 3, 10, 0, 0, 0, time.UTC),
		End:      time.Date(2025, 9, 3, 11, 0, 0, 0, time.UTC),
		Location: "Original Location",
	}

	tests := []struct {
		name          string
		uid           string
		options       EditOptions
		expectedTitle string
		expectedStart time.Time
		expectedEnd   time.Time
		expectedLoc   string
		expectError   bool
	}{
		{
			name: "edit title only",
			uid:  "test-1",
			options: EditOptions{
				Title: stringPtr("New Title"),
			},
			expectedTitle: "New Title",
			expectedStart: originalEvent.Start,
			expectedEnd:   originalEvent.End,
			expectedLoc:   "Original Location",
		},
		{
			name: "edit when only",
			uid:  "test-1",
			options: EditOptions{
				When: stringPtr("14:00"),
			},
			expectedTitle: "Original Title",
			expectedStart: time.Date(2025, 8, 29, 14, 0, 0, 0, time.Local), // StubTimeProvider date + new time
			expectedEnd:   time.Date(2025, 8, 29, 15, 0, 0, 0, time.Local), // preserves 1h duration
			expectedLoc:   "Original Location",
		},
		{
			name: "edit duration only",
			uid:  "test-1",
			options: EditOptions{
				Duration: stringPtr("2h"),
			},
			expectedTitle: "Original Title",
			expectedStart: originalEvent.Start,
			expectedEnd:   originalEvent.Start.Add(2 * time.Hour),
			expectedLoc:   "Original Location",
		},
		{
			name: "edit location only",
			uid:  "test-1",
			options: EditOptions{
				Location: stringPtr("New Location"),
			},
			expectedTitle: "Original Title",
			expectedStart: originalEvent.Start,
			expectedEnd:   originalEvent.End,
			expectedLoc:   "New Location",
		},
		{
			name: "edit multiple fields",
			uid:  "test-1",
			options: EditOptions{
				Title:    stringPtr("Multi Edit"),
				Duration: stringPtr("30m"),
				Location: stringPtr("Multi Location"),
			},
			expectedTitle: "Multi Edit",
			expectedStart: originalEvent.Start,
			expectedEnd:   originalEvent.Start.Add(30 * time.Minute),
			expectedLoc:   "Multi Location",
		},
		{
			name:        "event not found",
			uid:         "nonexistent",
			options:     EditOptions{Title: stringPtr("New Title")},
			expectError: true,
		},
		{
			name:        "UID too short",
			uid:         "x",
			options:     EditOptions{Title: stringPtr("New Title")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewFakeEventEditor()
			editor.AddEvent(originalEvent)

			fixedTime := time.Date(2025, 8, 29, 10, 30, 0, 0, time.Local)
			timeProvider := &StubTimeProvider{FixedTime: fixedTime}

			err := EditHandler(editor, timeProvider, tt.uid, tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error but got: %v", err)
				return
			}

			updatedEvent, exists := editor.events[tt.uid]
			if !exists {
				t.Error("event was not found in editor after update")
				return
			}

			if updatedEvent.Summary != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, updatedEvent.Summary)
			}

			if !updatedEvent.Start.Equal(tt.expectedStart) {
				t.Errorf("expected start %v, got %v", tt.expectedStart, updatedEvent.Start)
			}

			if !updatedEvent.End.Equal(tt.expectedEnd) {
				t.Errorf("expected end %v, got %v", tt.expectedEnd, updatedEvent.End)
			}

			if updatedEvent.Location != tt.expectedLoc {
				t.Errorf("expected location %q, got %q", tt.expectedLoc, updatedEvent.Location)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
