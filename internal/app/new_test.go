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

func TestNewHandlerWithRecurrence(t *testing.T) {
	tests := []struct {
		name               string
		title              string
		when               string
		duration           string
		location           string
		repeat             string
		count              int
		until              string
		fixedTime          time.Time
		uid                string
		expectErr          bool
		expectedRecurrence *domain.Recurrence
	}{
		{
			name:      "daily recurrence with count",
			title:     "Daily Standup",
			when:      "2025-09-01 09:00",
			duration:  "30m",
			location:  "Office",
			repeat:    "daily",
			count:     5,
			until:     "",
			fixedTime: time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
			uid:       "daily-uid",
			expectedRecurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
				Count:     intPtr(5),
				Until:     nil,
			},
		},
		{
			name:      "weekly recurrence with until",
			title:     "Team Meeting",
			when:      "2025-09-01 14:00",
			duration:  "1h",
			location:  "Conference Room",
			repeat:    "weekly",
			count:     0,
			until:     "2025-12-31",
			fixedTime: time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
			uid:       "weekly-uid",
			expectedRecurrence: &domain.Recurrence{
				Frequency: "WEEKLY",
				Interval:  1,
				Count:     nil,
				Until:     timePtr(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name:      "monthly recurrence",
			title:     "Monthly Review",
			when:      "2025-09-01 16:00",
			duration:  "2h",
			location:  "Boardroom",
			repeat:    "monthly",
			count:     12,
			until:     "",
			fixedTime: time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
			uid:       "monthly-uid",
			expectedRecurrence: &domain.Recurrence{
				Frequency: "MONTHLY",
				Interval:  1,
				Count:     intPtr(12),
				Until:     nil,
			},
		},
		{
			name:               "no recurrence",
			title:              "One-time Meeting",
			when:               "2025-09-01 10:00",
			duration:           "1h",
			location:           "Room A",
			repeat:             "",
			count:              0,
			until:              "",
			fixedTime:          time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
			uid:                "one-time-uid",
			expectedRecurrence: nil,
		},
		{
			name:      "invalid repeat pattern",
			title:     "Bad Repeat",
			when:      "2025-09-01 10:00",
			duration:  "1h",
			location:  "",
			repeat:    "invalid",
			count:     5,
			until:     "",
			fixedTime: time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
			uid:       "bad-uid",
			expectErr: true,
		},
		{
			name:      "both count and until specified",
			title:     "Conflict Test",
			when:      "2025-09-01 10:00",
			duration:  "1h",
			location:  "",
			repeat:    "daily",
			count:     5,
			until:     "2025-12-31",
			fixedTime: time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
			uid:       "conflict-uid",
			expectErr: true,
		},
		{
			name:      "invalid until date",
			title:     "Bad Until",
			when:      "2025-09-01 10:00",
			duration:  "1h",
			location:  "",
			repeat:    "daily",
			count:     0,
			until:     "invalid-date",
			fixedTime: time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
			uid:       "bad-until-uid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &FakeEventCreator{}
			timeProvider := &StubTimeProvider{FixedTime: tt.fixedTime}
			uidGen := &StubUIDGenerator{uid: tt.uid}

			err := NewHandlerWithRecurrence(creator, timeProvider, uidGen, tt.title, tt.when, tt.duration, tt.location, tt.repeat, tt.count, tt.until)

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

			if event.Summary != tt.title {
				t.Errorf("expected summary %q, got %q", tt.title, event.Summary)
			}

			if (event.Recurrence == nil) != (tt.expectedRecurrence == nil) {
				t.Errorf("recurrence presence mismatch: expected %v, got %v", tt.expectedRecurrence != nil, event.Recurrence != nil)
				return
			}

			if event.Recurrence != nil && tt.expectedRecurrence != nil {
				if event.Recurrence.Frequency != tt.expectedRecurrence.Frequency {
					t.Errorf("expected frequency %s, got %s", tt.expectedRecurrence.Frequency, event.Recurrence.Frequency)
				}

				if event.Recurrence.Interval != tt.expectedRecurrence.Interval {
					t.Errorf("expected interval %d, got %d", tt.expectedRecurrence.Interval, event.Recurrence.Interval)
				}

				if (event.Recurrence.Count == nil) != (tt.expectedRecurrence.Count == nil) {
					t.Errorf("count pointer mismatch: expected %v, got %v", tt.expectedRecurrence.Count, event.Recurrence.Count)
				} else if event.Recurrence.Count != nil && *event.Recurrence.Count != *tt.expectedRecurrence.Count {
					t.Errorf("expected count %d, got %d", *tt.expectedRecurrence.Count, *event.Recurrence.Count)
				}

				if (event.Recurrence.Until == nil) != (tt.expectedRecurrence.Until == nil) {
					t.Errorf("until pointer mismatch: expected %v, got %v", tt.expectedRecurrence.Until, event.Recurrence.Until)
				} else if event.Recurrence.Until != nil && !event.Recurrence.Until.Equal(*tt.expectedRecurrence.Until) {
					t.Errorf("expected until %v, got %v", *tt.expectedRecurrence.Until, *event.Recurrence.Until)
				}
			}
		})
	}
}
