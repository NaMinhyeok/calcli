package vdir

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

func TestWriter_CreateEvent(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	writer := NewWriter(tmpDir)

	event := domain.Event{
		UID:         "test-create-1",
		Summary:     "Test Event",
		Description: "A test event for writing",
		Location:    "Test Location",
		Start:       time.Date(2025, 8, 30, 10, 0, 0, 0, time.UTC),
		End:         time.Date(2025, 8, 30, 11, 0, 0, 0, time.UTC),
		AllDay:      false,
	}

	// Write event
	err := writer.CreateEvent(event)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check that file was created
	expectedPath := filepath.Join(tmpDir, "test-create-1.ics")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("expected ICS file to be created")
	}

	// Read back and verify content
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Errorf("failed to read created file: %v", err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Error("expected non-empty file content")
	}

	// Basic content validation
	expectedStrings := []string{
		"BEGIN:VCALENDAR",
		"BEGIN:VEVENT",
		"UID:test-create-1",
		"SUMMARY:Test Event",
		"END:VEVENT",
		"END:VCALENDAR",
	}

	for _, expected := range expectedStrings {
		if !containsString(content, expected) {
			t.Errorf("expected content to contain %q", expected)
		}
	}
}

// containsString checks if a string contains a substring (helper)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && indexString(s, substr) >= 0
}

// indexString finds the index of substr in s (helper)
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
