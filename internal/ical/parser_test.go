package ical

import (
	"strings"
	"testing"

	"calcli/internal/domain"
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
