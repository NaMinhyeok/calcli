package app

import (
	"fmt"

	"github.com/NaMinhyeok/calcli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// InteractiveHandler launches the interactive TUI
func InteractiveHandler(reader EventReader) error {
	model := tui.NewModel(reader)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
