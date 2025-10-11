package vdir

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/ical"
	"github.com/NaMinhyeok/calcli/internal/storage/cache"
)

type Reader struct {
	fs    fs.FS
	path  string
	cache *cache.EventCache
}

func NewReader(filesystem fs.FS, calendarPath string) *Reader {
	return &Reader{
		fs:    filesystem,
		path:  calendarPath,
		cache: nil, // Cache is optional
	}
}

// WithCache configures the reader to use caching
func (r *Reader) WithCache(c *cache.EventCache) *Reader {
	r.cache = c
	return r
}

func (r *Reader) ListEvents() ([]domain.Event, error) {
	var allEvents []domain.Event

	err := fs.WalkDir(r.fs, r.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".ics") {
			return nil
		}

		calendarName := filepath.Base(filepath.Dir(path))

		// Load events (with or without cache)
		return r.loadEventsFromFile(path, calendarName, &allEvents)
	})

	if err != nil {
		return nil, err
	}

	return allEvents, nil
}

// loadEventsFromFile loads events from a file, using cache if available
func (r *Reader) loadEventsFromFile(path string, calendarName string, allEvents *[]domain.Event) error {
	file, err := r.fs.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	events, err := ical.ParseEvents(file)
	if err != nil {
		// Skip files that can't be parsed, but continue processing others
		return nil
	}

	// Set calendar name for all events
	for i := range events {
		events[i].Calendar = calendarName
	}

	*allEvents = append(*allEvents, events...)
	return nil
}
