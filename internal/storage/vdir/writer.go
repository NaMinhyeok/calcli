package vdir

import (
	"os"
	"path/filepath"

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
