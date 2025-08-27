package domain

import (
	"time"
)

type Event struct {
	UID         string
	Summary     string
	Description string
	Start       time.Time
	End         time.Time
	Location    string
	Categories  []string
	AllDay      bool
	Calendar    string
}

func (e Event) Duration() time.Duration {
	if e.AllDay {
		return 24 * time.Hour
	}
	return e.End.Sub(e.Start)
}
