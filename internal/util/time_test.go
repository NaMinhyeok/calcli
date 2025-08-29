package util

import (
	"github.com/NaMinhyeok/calcli/internal/testutil"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "today",
			input:    "today",
			expected: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local),
		},
		{
			name:     "YYYY-MM-DD format",
			input:    "2025-08-30",
			expected: time.Date(2025, 8, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "invalid format",
			input:    "invalid-date",
			hasError: true,
		},
		{
			name:     "partial date",
			input:    "2025-08",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)

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

			// For "today" test, check year/month/day only since time.Now() changes
			if tt.input == "today" {
				if result.Year() != tt.expected.Year() || result.Month() != tt.expected.Month() || result.Day() != tt.expected.Day() {
					t.Errorf("expected date components %v, got %v", tt.expected, result)
				}
				if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 {
					t.Errorf("expected time to be 00:00:00, got %02d:%02d:%02d", result.Hour(), result.Minute(), result.Second())
				}
			} else {
				if !result.Equal(tt.expected) {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	fixedTime := time.Date(2025, 8, 29, 10, 30, 0, 0, time.Local)
	timeProvider := &testutil.StubTimeProvider{FixedTime: fixedTime}

	tests := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "full date and time",
			input:    "2025-08-30 14:00",
			expected: time.Date(2025, 8, 30, 14, 0, 0, 0, time.UTC),
		},
		{
			name:     "ISO format",
			input:    "2025-08-30T14:00",
			expected: time.Date(2025, 8, 30, 14, 0, 0, 0, time.UTC),
		},
		{
			name:     "time only uses current date",
			input:    "15:30",
			expected: time.Date(2025, 8, 29, 15, 30, 0, 0, time.Local),
		},
		{
			name:     "invalid format",
			input:    "invalid",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTime(tt.input, timeProvider)

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

			if !result.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
