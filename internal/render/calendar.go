package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

// MonthView renders a month calendar grid
type MonthView struct {
	Year   int
	Month  time.Month
	Events []domain.Event
	Today  time.Time
}

// NewMonthView creates a month view for a given date
func NewMonthView(date time.Time, events []domain.Event) *MonthView {
	return &MonthView{
		Year:   date.Year(),
		Month:  date.Month(),
		Events: events,
		Today:  time.Now(),
	}
}

// Render outputs the calendar to the writer
func (m *MonthView) Render(w io.Writer) error {
	// Header: Month Year
	fmt.Fprintf(w, "\n  %s %d\n\n", m.Month.String(), m.Year)

	// Weekday headers
	weekdays := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for _, day := range weekdays {
		fmt.Fprintf(w, " %3s", day)
	}
	fmt.Fprintln(w)

	// Separator
	fmt.Fprintln(w, strings.Repeat("-", 28))

	// Get first day of month and number of days
	firstDay := time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	numDays := lastDay.Day()

	// Starting weekday (0 = Sunday)
	startWeekday := int(firstDay.Weekday())

	// Build event map for quick lookup
	eventMap := m.buildEventMap()

	// Render calendar grid
	currentDay := 1
	for week := 0; week < 6; week++ { // Maximum 6 weeks
		if currentDay > numDays {
			break
		}

		for weekday := 0; weekday < 7; weekday++ {
			if week == 0 && weekday < startWeekday {
				// Empty cell before month starts
				fmt.Fprintf(w, "    ")
			} else if currentDay > numDays {
				// Empty cell after month ends
				fmt.Fprintf(w, "    ")
			} else {
				// Render day
				date := time.Date(m.Year, m.Month, currentDay, 0, 0, 0, 0, time.UTC)
				m.renderDay(w, currentDay, date, eventMap)
				currentDay++
			}
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	return nil
}

func (m *MonthView) renderDay(w io.Writer, day int, date time.Time, eventMap map[string]int) {
	hasEvents := eventMap[dateKey(date)] > 0
	isToday := m.isToday(date)

	var marker string
	if isToday {
		marker = "*" // Today marker
	} else if hasEvents {
		marker = "•" // Has events
	} else {
		marker = " "
	}

	fmt.Fprintf(w, " %2d%s", day, marker)
}

func (m *MonthView) isToday(date time.Time) bool {
	y1, m1, d1 := m.Today.Date()
	y2, m2, d2 := date.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func (m *MonthView) buildEventMap() map[string]int {
	eventMap := make(map[string]int)
	for _, event := range m.Events {
		key := dateKey(event.Start)
		eventMap[key]++
	}
	return eventMap
}

func dateKey(t time.Time) string {
	return t.Format("2006-01-02")
}

// RenderWithEvents renders the calendar with event list below
func (m *MonthView) RenderWithEvents(w io.Writer) error {
	if err := m.Render(w); err != nil {
		return err
	}

	// Group events by date
	eventsByDate := m.groupEventsByDate()

	if len(eventsByDate) == 0 {
		fmt.Fprintln(w, "No events this month.")
		return nil
	}

	fmt.Fprintln(w, "Events:")
	fmt.Fprintln(w, strings.Repeat("-", 50))

	// Sort dates and display events
	dates := make([]string, 0, len(eventsByDate))
	for date := range eventsByDate {
		dates = append(dates, date)
	}

	// Simple sort by date string (YYYY-MM-DD format sorts correctly)
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[i] > dates[j] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	for _, dateStr := range dates {
		events := eventsByDate[dateStr]
		date, _ := time.Parse("2006-01-02", dateStr)
		fmt.Fprintf(w, "\n%s %s:\n", date.Weekday().String()[:3], dateStr)

		for _, event := range events {
			startTime := event.Start.Format("15:04")
			fmt.Fprintf(w, "  %s %s\n", startTime, event.Summary)
		}
	}

	fmt.Fprintln(w)
	return nil
}

func (m *MonthView) groupEventsByDate() map[string][]domain.Event {
	groups := make(map[string][]domain.Event)
	for _, event := range m.Events {
		key := dateKey(event.Start)
		groups[key] = append(groups[key], event)
	}
	return groups
}
