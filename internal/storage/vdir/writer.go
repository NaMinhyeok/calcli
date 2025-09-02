package vdir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/ical"
)

type Writer struct {
	basePath string
}

func NewWriter(calendarPath string) *Writer {
	return &Writer{
		basePath: calendarPath,
	}
}

func (w *Writer) CreateEvent(event domain.Event) error {
	if err := os.MkdirAll(w.basePath, 0755); err != nil {
		return err
	}

	filename := filepath.Join(w.basePath, event.UID+".ics")
	tmpFile, err := os.CreateTemp(w.basePath, "tmp_*.ics")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if err := ical.GenerateEvent(event, tmpFile); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}

	tmpFile.Close()

	return os.Rename(tmpFile.Name(), filename)
}

func (w *Writer) FindEventByUID(uid string) (domain.Event, error) {
	var foundEvent domain.Event
	var found bool

	err := filepath.WalkDir(w.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".ics") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		events, err := ical.ParseEvents(file)
		if err != nil {
			return nil
		}

		for _, event := range events {
			if event.UID == uid {
				foundEvent = event
				found = true
				return fmt.Errorf("FOUND")
			}
		}

		return nil
	})

	if found {
		return foundEvent, nil
	}

	if err != nil && err.Error() == "FOUND" {
		return foundEvent, nil
	}

	return domain.Event{}, fmt.Errorf("event with UID %s not found", uid)
}

func (w *Writer) UpdateEvent(event domain.Event) error {
	return w.CreateEvent(event)
}
