package util

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "today",
			input:    "today",
			expected: today,
		},
		{
			name:     "tomorrow",
			input:    "tomorrow",
			expected: today.AddDate(0, 0, 1),
		},
		{
			name:     "yesterday",
			input:    "yesterday",
			expected: today.AddDate(0, 0, -1),
		},
		{
			name:     "+3d",
			input:    "+3d",
			expected: today.AddDate(0, 0, 3),
		},
		{
			name:     "-2d",
			input:    "-2d",
			expected: today.AddDate(0, 0, -2),
		},
		{
			name:     "5d (implicit positive)",
			input:    "5d",
			expected: today.AddDate(0, 0, 5),
		},
		{
			name:     "+2w",
			input:    "+2w",
			expected: today.AddDate(0, 0, 14),
		},
		{
			name:     "-1w",
			input:    "-1w",
			expected: today.AddDate(0, 0, -7),
		},
		{
			name:     "3w",
			input:    "3w",
			expected: today.AddDate(0, 0, 21),
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

// StubTimeProvider for testing
type StubTimeProvider struct {
	FixedTime time.Time
}

func (s *StubTimeProvider) Now() time.Time {
	return s.FixedTime
}

func TestParseTime(t *testing.T) {
	fixedTime := time.Date(2025, 8, 29, 10, 30, 0, 0, time.Local)
	timeProvider := &StubTimeProvider{FixedTime: fixedTime}

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
			result, err := ParseDuration(tt.input)

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
