package app

import (
	"fmt"
	"io"

	"github.com/NaMinhyeok/calcli/internal/config"
)

func CalendarsHandler(cfg *config.Config, output io.Writer) error {
	for name, calendar := range cfg.Calendars {
		fmt.Fprintf(output, "%s: %s\n", name, calendar.Path)
	}
	return nil
}
