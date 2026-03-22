package tui

import tea "github.com/charmbracelet/bubbletea"

// View extends tea.Model with size management for tabbed views.
type View interface {
	tea.Model
	SetSize(width, height int)
}
