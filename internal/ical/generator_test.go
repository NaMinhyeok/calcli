package ical

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"calcli/internal/domain"
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
