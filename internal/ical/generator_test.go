package ical

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

func TestGenerateEvent(t *testing.T) {
	event := domain.Event{
		UID:         "test-event-1",
		Summary:     "Test Meeting",
		Description: "A test meeting",
		Location:    "Conference Room",
		Start:       time.Date(2025, 8, 30, 14, 0, 0, 0, time.UTC),
		End:         time.Date(2025, 8, 30, 15, 0, 0, 0, time.UTC),
		AllDay:      false,
	}

	var buf bytes.Buffer
	err := GenerateEvent(event, &buf)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	output := buf.String()

	// Check required components
	if !strings.Contains(output, "BEGIN:VCALENDAR") {
		t.Error("output should contain BEGIN:VCALENDAR")
	}

	if !strings.Contains(output, "BEGIN:VEVENT") {
		t.Error("output should contain BEGIN:VEVENT")
	}

	if !strings.Contains(output, "SUMMARY:Test Meeting") {
		t.Error("output should contain SUMMARY")
	}

	if !strings.Contains(output, "UID:test-event-1") {
		t.Error("output should contain UID")
	}
}

func TestGenerateEvent_WithRecurrence(t *testing.T) {
	tests := []struct {
		name       string
		recurrence *domain.Recurrence
		wantRRULE  string
	}{
		{
			name: "daily with count",
			recurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
				Count:     intPtr(5),
			},
			wantRRULE: "RRULE:FREQ=DAILY;COUNT=5",
		},
		{
			name: "weekly with interval",
			recurrence: &domain.Recurrence{
				Frequency: "WEEKLY",
				Interval:  2,
				Count:     intPtr(10),
			},
			wantRRULE: "RRULE:FREQ=WEEKLY;INTERVAL=2;COUNT=10",
		},
		{
			name: "monthly with until",
			recurrence: &domain.Recurrence{
				Frequency: "MONTHLY",
				Interval:  1,
				Until:     timePtr(time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)),
			},
			wantRRULE: "RRULE:FREQ=MONTHLY;UNTIL=20251231T235959Z",
		},
		{
			name: "daily with default interval",
			recurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
			},
			wantRRULE: "RRULE:FREQ=DAILY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := domain.Event{
				UID:        "recurring-event",
				Summary:    "Recurring Meeting",
				Start:      time.Date(2025, 9, 1, 10, 0, 0, 0, time.UTC),
				End:        time.Date(2025, 9, 1, 11, 0, 0, 0, time.UTC),
				Recurrence: tt.recurrence,
			}

			var buf bytes.Buffer
			err := GenerateEvent(event, &buf)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantRRULE) {
				t.Errorf("output should contain %q\nGot output:\n%s", tt.wantRRULE, output)
			}
		})
	}
}

func TestBuildRRULE(t *testing.T) {
	tests := []struct {
		name       string
		recurrence *domain.Recurrence
		want       string
	}{
		{
			name:       "nil recurrence",
			recurrence: nil,
			want:       "",
		},
		{
			name: "simple daily",
			recurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
			},
			want: "FREQ=DAILY",
		},
		{
			name: "weekly with interval 2",
			recurrence: &domain.Recurrence{
				Frequency: "WEEKLY",
				Interval:  2,
			},
			want: "FREQ=WEEKLY;INTERVAL=2",
		},
		{
			name: "with count",
			recurrence: &domain.Recurrence{
				Frequency: "DAILY",
				Interval:  1,
				Count:     intPtr(10),
			},
			want: "FREQ=DAILY;COUNT=10",
		},
		{
			name: "with until",
			recurrence: &domain.Recurrence{
				Frequency: "MONTHLY",
				Interval:  1,
				Until:     timePtr(time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)),
			},
			want: "FREQ=MONTHLY;UNTIL=20251231T235959Z",
		},
		{
			name: "complex rule",
			recurrence: &domain.Recurrence{
				Frequency: "WEEKLY",
				Interval:  2,
				Count:     intPtr(5),
			},
			want: "FREQ=WEEKLY;INTERVAL=2;COUNT=5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildRRULE(tt.recurrence)
			if got != tt.want {
				t.Errorf("buildRRULE() = %q, want %q", got, tt.want)
			}
		})
	}
}
