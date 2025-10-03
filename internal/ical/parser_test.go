package ical

import (
	"strings"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

func TestParseEvents(t *testing.T) {
	tests := []struct {
		name       string
		icsData    string
		wantErr    bool
		wantLen    int
		checkEvent func(t *testing.T, events []domain.Event)
	}{
		{
			name: "single basic event",
			icsData: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:test-event-1
SUMMARY:Test Meeting
DESCRIPTION:A test meeting
LOCATION:Conference Room
DTSTART:20250827T140000Z
DTEND:20250827T150000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: false,
			wantLen: 1,
			checkEvent: func(t *testing.T, events []domain.Event) {
				e := events[0]
				if e.UID != "test-event-1" {
					t.Errorf("expected UID 'test-event-1', got %s", e.UID)
				}
				if e.Summary != "Test Meeting" {
					t.Errorf("expected Summary 'Test Meeting', got %s", e.Summary)
				}
				if e.Location != "Conference Room" {
					t.Errorf("expected Location 'Conference Room', got %s", e.Location)
				}
				if e.AllDay {
					t.Error("expected non-all-day event")
				}
			},
		},
		{
			name: "all-day event",
			icsData: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:allday-1
SUMMARY:All Day Event
DTSTART:20250827T000000Z
DTEND:20250828T000000Z
END:VEVENT
END:VCALENDAR`,
			wantErr: false,
			wantLen: 1,
			checkEvent: func(t *testing.T, events []domain.Event) {
				e := events[0]
				if e.Summary != "All Day Event" {
					t.Errorf("expected Summary 'All Day Event', got %s", e.Summary)
				}
				if !e.AllDay {
					t.Error("expected all-day event")
				}
			},
		},
		{
			name: "empty calendar",
			icsData: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
END:VCALENDAR`,
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "invalid ics data",
			icsData: "invalid ics content",
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.icsData)
			events, err := ParseEvents(reader)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if len(events) != tt.wantLen {
				t.Errorf("expected %d events, got %d", tt.wantLen, len(events))
			}

			if tt.checkEvent != nil && len(events) > 0 {
				tt.checkEvent(t, events)
			}
		})
	}
}

func TestParseRRULE(t *testing.T) {
	tests := []struct {
		name     string
		rrule    string
		expected *domain.Recurrence
	}{
		{
			name:  "daily recurrence with count",
			rrule: "FREQ=DAILY;COUNT=10",
			expected: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
				Count:     intPtr(10),
				Until:     nil,
			},
		},
		{
			name:  "weekly recurrence with interval",
			rrule: "FREQ=WEEKLY;INTERVAL=2",
			expected: &domain.Recurrence{
				Frequency: "WEEKLY",
				Interval:  2,
				Count:     nil,
				Until:     nil,
			},
		},
		{
			name:  "monthly recurrence with until",
			rrule: "FREQ=MONTHLY;UNTIL=20251231T235959Z",
			expected: &domain.Recurrence{
				Frequency: "MONTHLY",
				Interval:  1,
				Count:     nil,
				Until:     timePtr(time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)),
			},
		},
		{
			name:  "complex recurrence",
			rrule: "FREQ=DAILY;INTERVAL=3;COUNT=5",
			expected: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  3,
				Count:     intPtr(5),
				Until:     nil,
			},
		},
		{
			name:     "empty rrule",
			rrule:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRRULE(tt.rrule)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Error("expected recurrence, got nil")
				return
			}

			if result.Frequency != tt.expected.Frequency {
				t.Errorf("expected Frequency %s, got %s", tt.expected.Frequency, result.Frequency)
			}

			if result.Interval != tt.expected.Interval {
				t.Errorf("expected Interval %d, got %d", tt.expected.Interval, result.Interval)
			}

			if (result.Count == nil) != (tt.expected.Count == nil) {
				t.Errorf("Count pointer mismatch: expected %v, got %v", tt.expected.Count, result.Count)
			} else if result.Count != nil && *result.Count != *tt.expected.Count {
				t.Errorf("expected Count %d, got %d", *tt.expected.Count, *result.Count)
			}

			if (result.Until == nil) != (tt.expected.Until == nil) {
				t.Errorf("Until pointer mismatch: expected %v, got %v", tt.expected.Until, result.Until)
			} else if result.Until != nil && !result.Until.Equal(*tt.expected.Until) {
				t.Errorf("expected Until %v, got %v", *tt.expected.Until, *result.Until)
			}
		})
	}
}
