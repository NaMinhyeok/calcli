package vdir

import (
	"io/fs"
	"path/filepath"
	"strings"

	"calcli/internal/domain"
	"calcli/internal/ical"
)

type Reader struct {
	fs   fs.FS
	path string
}

func NewReader(filesystem fs.FS, calendarPath string) *Reader {
	return &Reader{
		fs:   filesystem,
		path: calendarPath,
	}
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

		calendarName := filepath.Base(filepath.Dir(path))
		for i := range events {
			events[i].Calendar = calendarName
		}

		allEvents = append(allEvents, events...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allEvents, nil
}
