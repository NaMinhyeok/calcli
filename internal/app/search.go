package app

import (
	"io"
	"strings"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

type EventSearcher interface {
	ListEvents() ([]domain.Event, error)
}

type SearchField string

const (
	SearchFieldAny      SearchField = "any"
	SearchFieldTitle    SearchField = "title"
	SearchFieldDesc     SearchField = "desc"
	SearchFieldLocation SearchField = "location"
)

func SearchHandler(searcher EventSearcher, formatter EventFormatter, output io.Writer, query string, field SearchField) error {
	events, err := searcher.ListEvents()
	if err != nil {
		return err
	}

	// Expand recurring events for a reasonable search range (1 year from now)
	rangeStart := time.Now().AddDate(0, 0, -30) // 30 days in the past
	rangeEnd := time.Now().AddDate(1, 0, 0)     // 1 year in the future

	var expandedEvents []domain.Event
	for _, event := range events {
		if event.Recurrence != nil {
			instances := domain.ExpandRecurrence(event, rangeStart, rangeEnd)
			expandedEvents = append(expandedEvents, instances...)
		} else {
			expandedEvents = append(expandedEvents, event)
		}
	}

	var matches []domain.Event
	for _, event := range expandedEvents {
		if matchesEvent(event, query, field) {
			matches = append(matches, event)
		}
	}

	formatter.FormatEvents(matches, output)
	return nil
}

func matchesEvent(event domain.Event, query string, field SearchField) bool {
	query = strings.ToLower(query)

	switch field {
	case SearchFieldTitle:
		return strings.Contains(strings.ToLower(event.Summary), query)
	case SearchFieldDesc:
		return strings.Contains(strings.ToLower(event.Description), query)
	case SearchFieldLocation:
		return strings.Contains(strings.ToLower(event.Location), query)
	case SearchFieldAny:
		return strings.Contains(strings.ToLower(event.Summary), query) ||
			strings.Contains(strings.ToLower(event.Description), query) ||
			strings.Contains(strings.ToLower(event.Location), query)
	default:
		return false
	}
}
