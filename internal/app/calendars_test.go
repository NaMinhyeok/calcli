package app

import (
	"bytes"
	"testing"

	"calcli/internal/config"
)

func TestCalendarsHandler(t *testing.T) {
	cfg := &config.Config{
		Calendars: map[string]config.CalendarConfig{
			"home": {
				Path:     "/home/user/.calcli/home",
				Color:    "blue",
				ReadOnly: false,
			},
			"work": {
				Path:     "/home/user/.calcli/work",
				Color:    "red",
				ReadOnly: true,
			},
		},
		Defaults: config.DefaultsConfig{
			DefaultCalendar: "home",
		},
	}

	var buf bytes.Buffer
	err := CalendarsHandler(cfg, &buf)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	output := buf.String()
	expectedStrings := []string{
		"home: /home/user/.calcli/home",
		"work: /home/user/.calcli/work",
	}

	for _, expected := range expectedStrings {
		if !containsString(output, expected) {
			t.Errorf("output should contain %q, got: %q", expected, output)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && indexString(s, substr) >= 0
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
