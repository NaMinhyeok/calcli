package vdir

import (
	"testing"
	"testing/fstest"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

func TestReader_ListEvents(t *testing.T) {
	// Create a test filesystem with sample .ics files
	testFS := fstest.MapFS{
		"home/event1.ics": &fstest.MapFile{
			Data: []byte(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:home-event-1
SUMMARY:Home Meeting
DTSTART:20250828T100000Z
DTEND:20250828T110000Z
END:VEVENT
END:VCALENDAR`),
		},
		"work/event2.ics": &fstest.MapFile{
			Data: []byte(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:work-event-1
SUMMARY:Work Meeting
DTSTART:20250828T140000Z
DTEND:20250828T150000Z
END:VEVENT
END:VCALENDAR`),
		},
		"home/not-ics.txt": &fstest.MapFile{
			Data: []byte("This is not an ICS file"),
		},
	}

	tests := []struct {
		name        string
		path        string
		wantErr     bool
		wantLen     int
		checkEvents func(t *testing.T, events []domain.Event)
	}{
		{
			name:    "read home calendar",
			path:    "home",
			wantErr: false,
			wantLen: 1,
			checkEvents: func(t *testing.T, events []domain.Event) {
				e := events[0]
				if e.Summary != "Home Meeting" {
					t.Errorf("expected 'Home Meeting', got %s", e.Summary)
				}
				if e.Calendar != "home" {
					t.Errorf("expected calendar 'home', got %s", e.Calendar)
				}
			},
		},
		{
			name:    "read work calendar",
			path:    "work",
			wantErr: false,
			wantLen: 1,
			checkEvents: func(t *testing.T, events []domain.Event) {
				e := events[0]
				if e.Summary != "Work Meeting" {
					t.Errorf("expected 'Work Meeting', got %s", e.Summary)
				}
				if e.Calendar != "work" {
					t.Errorf("expected calendar 'work', got %s", e.Calendar)
				}
			},
		},
		{
			name:    "nonexistent path",
			path:    "nonexistent",
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewReader(testFS, tt.path)
			events, err := reader.ListEvents()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if len(events) != tt.wantLen {
				t.Errorf("expected %d events, got %d", tt.wantLen, len(events))
			}

			if tt.checkEvents != nil && len(events) > 0 {
				tt.checkEvents(t, events)
			}
		})
	}
}
