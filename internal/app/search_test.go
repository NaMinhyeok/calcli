package app

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

type FakeEventSearcher struct {
	events []domain.Event
	err    error
}

func (f *FakeEventSearcher) ListEvents() ([]domain.Event, error) {
	return f.events, f.err
}

func TestSearchHandler(t *testing.T) {
	events := []domain.Event{
		{
			UID:         "1",
			Summary:     "Team Meeting",
			Description: "Weekly team sync",
			Location:    "Conference Room A",
			Start:       time.Date(2025, 8, 30, 10, 0, 0, 0, time.UTC),
			End:         time.Date(2025, 8, 30, 11, 0, 0, 0, time.UTC),
		},
		{
			UID:         "2",
			Summary:     "1:1 with Manager",
			Description: "Personal development meeting discussion",
			Location:    "Office",
			Start:       time.Date(2025, 8, 30, 14, 0, 0, 0, time.UTC),
			End:         time.Date(2025, 8, 30, 15, 0, 0, 0, time.UTC),
		},
		{
			UID:         "3",
			Summary:     "Project Review",
			Description: "Review Q3 meeting outcomes",
			Location:    "Meeting Room B",
			Start:       time.Date(2025, 8, 31, 9, 0, 0, 0, time.UTC),
			End:         time.Date(2025, 8, 31, 10, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name           string
		query          string
		field          SearchField
		expectedCount  int
		expectedTitles []string
	}{
		{
			name:           "search any field for 'meeting'",
			query:          "meeting",
			field:          SearchFieldAny,
			expectedCount:  3, // "Team Meeting", "1:1 with Manager" (desc), "Project Review" (desc has "meeting")
			expectedTitles: []string{"Team Meeting", "1:1 with Manager", "Project Review"},
		},
		{
			name:           "search title only for 'team'",
			query:          "team",
			field:          SearchFieldTitle,
			expectedCount:  1,
			expectedTitles: []string{"Team Meeting"},
		},
		{
			name:           "search description for 'development'",
			query:          "development",
			field:          SearchFieldDesc,
			expectedCount:  1,
			expectedTitles: []string{"1:1 with Manager"},
		},
		{
			name:           "search location for 'room'",
			query:          "room",
			field:          SearchFieldLocation,
			expectedCount:  2,
			expectedTitles: []string{"Team Meeting", "Project Review"},
		},
		{
			name:          "no matches",
			query:         "nonexistent",
			field:         SearchFieldAny,
			expectedCount: 0,
		},
		{
			name:           "case insensitive search",
			query:          "TEAM",
			field:          SearchFieldTitle,
			expectedCount:  1,
			expectedTitles: []string{"Team Meeting"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searcher := &FakeEventSearcher{events: events}
			formatter := &SimpleEventFormatter{}
			var buf bytes.Buffer

			err := SearchHandler(searcher, formatter, &buf, tt.query, tt.field)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			output := buf.String()

			// Check that all expected titles appear in output
			for _, expectedTitle := range tt.expectedTitles {
				if !strings.Contains(output, expectedTitle) {
					t.Errorf("expected output to contain %q, but it didn't. Output: %q", expectedTitle, output)
				}
			}

			// Count actual matches by counting how many expected titles appear
			matchCount := 0
			for _, expectedTitle := range tt.expectedTitles {
				if strings.Contains(output, expectedTitle) {
					matchCount++
				}
			}

			if matchCount != tt.expectedCount {
				t.Errorf("expected %d matches, got %d. Output: %q", tt.expectedCount, matchCount, output)
			}
		})
	}
}

func TestMatchesEvent(t *testing.T) {
	event := domain.Event{
		Summary:     "Team Meeting",
		Description: "Weekly sync with the team",
		Location:    "Conference Room",
	}

	tests := []struct {
		name     string
		query    string
		field    SearchField
		expected bool
	}{
		{"title match", "meeting", SearchFieldTitle, true},
		{"title no match", "lunch", SearchFieldTitle, false},
		{"desc match", "sync", SearchFieldDesc, true},
		{"desc no match", "project", SearchFieldDesc, false},
		{"location match", "conference", SearchFieldLocation, true},
		{"location no match", "office", SearchFieldLocation, false},
		{"any field title match", "team", SearchFieldAny, true},
		{"any field desc match", "weekly", SearchFieldAny, true},
		{"any field location match", "room", SearchFieldAny, true},
		{"any field no match", "nonexistent", SearchFieldAny, false},
		{"case insensitive", "TEAM", SearchFieldTitle, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesEvent(event, tt.query, tt.field)
			if result != tt.expected {
				t.Errorf("matchesEvent(%q, %q) = %v, want %v", tt.query, tt.field, result, tt.expected)
			}
		})
	}
}
