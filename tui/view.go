package tui

import tea "github.com/charmbracelet/bubbletea"

// View extends tea.Model with size management for tabbed views.
type View interface {
	tea.Model
	SetSize(width, height int)
}

// GlobalShortcutSuppressor is an optional view contract for telling the root app
// to forward conflicting shortcuts to the active view instead of handling them
// globally. Views that do not need this behavior can ignore the interface.
type GlobalShortcutSuppressor interface {
	SuppressGlobalShortcuts() bool
}
