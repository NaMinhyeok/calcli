package app

import (
	"fmt"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

// Test doubles
type FakeEventCreator struct {
	events []domain.Event
	err    error
}

func (f *FakeEventCreator) CreateEvent(event domain.Event) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, event)
	return nil
}

type StubUIDGenerator struct {
	uid string
	err error
}

func (s *StubUIDGenerator) Generate() (string, error) {
	return s.uid, s.err
}

type StubTimeProvider struct {
	FixedTime time.Time
}

func (s *StubTimeProvider) Now() time.Time {
	return s.FixedTime
}

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		when          string
		duration      string
		location      string
		fixedTime     time.Time
		uid           string
		createErr     error
		uidErr        error
		expectErr     bool
		expectedUID   string
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			name:          "basic event creation",
			title:         "Test Meeting",
			when:          "2025-08-30 14:00",
			duration:      "1h",
			location:      "Conference Room",
			fixedTime:     time.Date(2025, 8, 29, 10, 0, 0, 0, time.UTC),
			uid:           "test-uid-123",
			expectedUID:   "test-uid-123",
			expectedStart: time.Date(2025, 8, 30, 14, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2025, 8, 30, 15, 0, 0, 0, time.UTC),
		},
		{
			name:          "time-only format uses current date",
			title:         "Daily Standup",
			when:          "09:00",
			duration:      "30m",
			location:      "",
			fixedTime:     time.Date(2025, 8, 29, 10, 0, 0, 0, time.Local),
			uid:           "standup-uid",
			expectedUID:   "standup-uid",
			expectedStart: time.Date(2025, 8, 29, 9, 0, 0, 0, time.Local),
			expectedEnd:   time.Date(2025, 8, 29, 9, 30, 0, 0, time.Local),
		},
		{
			name:          "default duration when empty",
			title:         "Quick Chat",
			when:          "2025-09-01 16:00",
			duration:      "",
			location:      "",
			fixedTime:     time.Date(2025, 8, 29, 10, 0, 0, 0, time.UTC),
			uid:           "chat-uid",
			expectedUID:   "chat-uid",
			expectedStart: time.Date(2025, 9, 1, 16, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2025, 9, 1, 17, 0, 0, 0, time.UTC),
		},
		{
			name:      "invalid time format",
			title:     "Bad Time",
			when:      "invalid-time",
			duration:  "1h",
			location:  "",
			fixedTime: time.Date(2025, 8, 29, 10, 0, 0, 0, time.UTC),
			uid:       "bad-uid",
			expectErr: true,
		},
		{
			name:      "invalid duration format",
			title:     "Bad Duration",
			when:      "2025-08-30 14:00",
			duration:  "invalid-duration",
			location:  "",
			fixedTime: time.Date(2025, 8, 29, 10, 0, 0, 0, time.UTC),
			uid:       "bad-duration-uid",
			expectErr: true,
		},
		{
			name:      "UID generation error",
			title:     "UID Fail",
			when:      "2025-08-30 14:00",
			duration:  "1h",
			location:  "",
			fixedTime: time.Date(2025, 8, 29, 10, 0, 0, 0, time.UTC),
			uidErr:    fmt.Errorf("UID generation failed"),
			expectErr: true,
		},
		{
			name:      "event creation error",
			title:     "Create Fail",
			when:      "2025-08-30 14:00",
			duration:  "1h",
			location:  "",
			fixedTime: time.Date(2025, 8, 29, 10, 0, 0, 0, time.UTC),
			uid:       "fail-uid",
			createErr: fmt.Errorf("storage error"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &FakeEventCreator{err: tt.createErr}
			timeProvider := &StubTimeProvider{FixedTime: tt.fixedTime}
			uidGen := &StubUIDGenerator{uid: tt.uid, err: tt.uidErr}

			err := NewHandler(creator, timeProvider, uidGen, tt.title, tt.when, tt.duration, tt.location)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error but got: %v", err)
				return
			}

			if len(creator.events) != 1 {
				t.Errorf("expected 1 event created, got %d", len(creator.events))
				return
			}

			event := creator.events[0]

			if event.UID != tt.expectedUID {
				t.Errorf("expected UID %q, got %q", tt.expectedUID, event.UID)
			}

			if event.Summary != tt.title {
				t.Errorf("expected summary %q, got %q", tt.title, event.Summary)
			}

			if !event.Start.Equal(tt.expectedStart) {
				t.Errorf("expected start time %v, got %v", tt.expectedStart, event.Start)
			}

			if !event.End.Equal(tt.expectedEnd) {
				t.Errorf("expected end time %v, got %v", tt.expectedEnd, event.End)
			}

			if event.AllDay != false {
				t.Errorf("expected AllDay to be false, got %v", event.AllDay)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		hasError bool
	}{
		{
			name:     "empty uses default",
			input:    "",
			expected: time.Hour,
		},
		{
			name:     "valid hour duration",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "valid minute duration",
			input:    "30m",
			expected: 30 * time.Minute,
		},
		{
			name:     "combined duration",
			input:    "1h30m",
			expected: time.Hour + 30*time.Minute,
		},
		{
			name:     "invalid duration",
			input:    "invalid",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error but got: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
