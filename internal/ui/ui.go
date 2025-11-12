package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the interactive TUI
func Run(model *Model) error {
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
