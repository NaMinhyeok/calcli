package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/NaMinhyeok/calcli/internal/domain"
	"github.com/NaMinhyeok/calcli/internal/render"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the TUI state
type Model struct {
	// Data
	events       []domain.Event
	currentMonth time.Time

	// UI State
	selectedDay   int // 1-31, 0 means no selection
	selectedEvent int // index in filtered events
	width         int
	height        int

	// View mode
	showHelp bool

	// Reader for loading events
	reader EventReader
}

// EventReader interface for loading events
type EventReader interface {
	ListEvents() ([]domain.Event, error)
}

// NewModel creates a new TUI model
func NewModel(reader EventReader) Model {
	return Model{
		currentMonth:  time.Now(),
		selectedDay:   time.Now().Day(),
		selectedEvent: 0,
		showHelp:      false,
		reader:        reader,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return loadEvents(m.reader)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case eventsLoadedMsg:
		m.events = msg.events
		return m, nil

	case errMsg:
		// Handle error (for now, just quit)
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "left", "h":
		if m.showHelp {
			return m, nil
		}
		m.currentMonth = m.currentMonth.AddDate(0, -1, 0)
		m.selectedDay = 1
		return m, loadEvents(m.reader)

	case "right", "l":
		if !m.showHelp {
			m.currentMonth = m.currentMonth.AddDate(0, 1, 0)
			m.selectedDay = 1
			return m, loadEvents(m.reader)
		}
		return m, nil

	case "up", "k":
		if m.selectedEvent > 0 {
			m.selectedEvent--
		}
		return m, nil

	case "down", "j":
		events := m.getEventsForSelectedDay()
		if m.selectedEvent < len(events)-1 {
			m.selectedEvent++
		}
		return m, nil

	case "enter":
		// Select day (cycle through days with events)
		return m, nil

	case "n":
		// TODO: New event (future implementation)
		return m, nil

	case "d":
		// TODO: Delete event (future implementation)
		return m, nil

	case "e":
		// TODO: Edit event (future implementation)
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	var b strings.Builder

	// Title
	b.WriteString(fmt.Sprintf("\n  Calendar - %s %d\n\n", m.currentMonth.Month(), m.currentMonth.Year()))

	// Calendar view
	b.WriteString(m.renderCalendar())
	b.WriteString("\n")

	// Event list
	b.WriteString(m.renderEventList())
	b.WriteString("\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderCalendar() string {
	// Get events for current month
	firstDay := time.Date(m.currentMonth.Year(), m.currentMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Second)

	// Filter events for this month
	var monthEvents []domain.Event
	for _, event := range m.events {
		if !event.Start.Before(firstDay) && !event.Start.After(lastDay) {
			monthEvents = append(monthEvents, event)
		}
	}

	// Expand recurring events
	var expandedEvents []domain.Event
	for _, event := range monthEvents {
		instances := domain.ExpandRecurrence(event, firstDay, lastDay)
		expandedEvents = append(expandedEvents, instances...)
	}

	// Use render.MonthView
	view := render.NewMonthView(m.currentMonth, expandedEvents)
	view.Today = time.Now()

	var buf strings.Builder
	view.Render(&buf)
	return buf.String()
}

func (m Model) renderEventList() string {
	events := m.getEventsForSelectedDay()

	var b strings.Builder
	b.WriteString("Events:\n")
	b.WriteString(strings.Repeat("-", 50))
	b.WriteString("\n")

	if len(events) == 0 {
		b.WriteString("No events for this day.\n")
		return b.String()
	}

	selectedDate := time.Date(m.currentMonth.Year(), m.currentMonth.Month(), m.selectedDay, 0, 0, 0, 0, time.UTC)
	b.WriteString(fmt.Sprintf("\n%s %s:\n", selectedDate.Weekday().String()[:3], selectedDate.Format("2006-01-02")))

	for i, event := range events {
		prefix := "  "
		if i == m.selectedEvent {
			prefix = "> "
		}

		startTime := event.Start.Format("15:04")
		b.WriteString(fmt.Sprintf("%s%s %s\n", prefix, startTime, event.Summary))
	}

	return b.String()
}

func (m Model) renderFooter() string {
	return "  ←/→: month  ↑/↓: event  q: quit  ?: help\n"
}

func (m Model) renderHelp() string {
	var b strings.Builder

	b.WriteString("\n  Calendar Help\n\n")
	b.WriteString("  Navigation:\n")
	b.WriteString("    ← / → : Previous/Next month\n")
	b.WriteString("    ↑ / ↓ : Previous/Next event\n")
	b.WriteString("    h / l : Same as ← / →\n")
	b.WriteString("    j / k : Same as ↑ / ↓\n")
	b.WriteString("\n")
	b.WriteString("  Actions:\n")
	b.WriteString("    n     : New event (coming soon)\n")
	b.WriteString("    e     : Edit event (coming soon)\n")
	b.WriteString("    d     : Delete event (coming soon)\n")
	b.WriteString("\n")
	b.WriteString("  Other:\n")
	b.WriteString("    ?     : Toggle this help\n")
	b.WriteString("    q/Esc : Quit\n")
	b.WriteString("\n")
	b.WriteString("  Press any key to return...\n")

	return b.String()
}

func (m Model) getEventsForSelectedDay() []domain.Event {
	if m.selectedDay == 0 {
		return nil
	}

	selectedDate := time.Date(m.currentMonth.Year(), m.currentMonth.Month(), m.selectedDay, 0, 0, 0, 0, time.UTC)

	var dayEvents []domain.Event
	for _, event := range m.events {
		eventDate := event.Start.Format("2006-01-02")
		selectedDateStr := selectedDate.Format("2006-01-02")

		if eventDate == selectedDateStr {
			dayEvents = append(dayEvents, event)
		}
	}

	return dayEvents
}

// Messages

type eventsLoadedMsg struct {
	events []domain.Event
}

type errMsg struct {
	err error
}

// Commands

func loadEvents(reader EventReader) tea.Cmd {
	return func() tea.Msg {
		events, err := reader.ListEvents()
		if err != nil {
			return errMsg{err}
		}
		return eventsLoadedMsg{events}
	}
}
