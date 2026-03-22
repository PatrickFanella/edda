package tui

import tea "github.com/charmbracelet/bubbletea"

// TabInfo holds metadata about a registered tab.
type TabInfo struct {
	Name string
	View View
}

// Router manages a collection of tabbed views, routing messages
// and rendering to the currently active one.
type Router struct {
	tabs      []TabInfo
	activeIdx int
	width     int
	height    int
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// Register adds a view as a new tab.
func (r *Router) Register(name string, view View) {
	r.tabs = append(r.tabs, TabInfo{Name: name, View: view})
}

// ActiveTab returns the index of the currently active tab.
func (r *Router) ActiveTab() int {
	return r.activeIdx
}

// TabCount returns the number of registered tabs.
func (r *Router) TabCount() int {
	return len(r.tabs)
}

// Tabs returns a copy of all registered tab infos.
func (r *Router) Tabs() []TabInfo {
	out := make([]TabInfo, len(r.tabs))
	copy(out, r.tabs)
	return out
}

// NextTab cycles to the next tab.
func (r *Router) NextTab() {
	if len(r.tabs) == 0 {
		return
	}
	r.activeIdx = (r.activeIdx + 1) % len(r.tabs)
}

// PrevTab cycles to the previous tab.
func (r *Router) PrevTab() {
	if len(r.tabs) == 0 {
		return
	}
	r.activeIdx = (r.activeIdx - 1 + len(r.tabs)) % len(r.tabs)
}

// GoToTab switches to the tab at the given index (0-based).
func (r *Router) GoToTab(idx int) {
	if idx >= 0 && idx < len(r.tabs) {
		r.activeIdx = idx
	}
}

// SetSize propagates dimensions to all registered views.
func (r *Router) SetSize(width, height int) {
	r.width = width
	r.height = height
	for _, tab := range r.tabs {
		tab.View.SetSize(width, height)
	}
}

// ActiveView returns the currently active View, or nil if none registered.
func (r *Router) ActiveView() View {
	if len(r.tabs) == 0 {
		return nil
	}
	return r.tabs[r.activeIdx].View
}

// Update routes a message to the active view and returns any command.
func (r *Router) Update(msg tea.Msg) tea.Cmd {
	if len(r.tabs) == 0 {
		return nil
	}
	active := r.tabs[r.activeIdx].View
	updated, cmd := active.Update(msg)
	if v, ok := updated.(View); ok {
		r.tabs[r.activeIdx].View = v
	}
	return cmd
}

// View renders the active view.
func (r *Router) View() string {
	if len(r.tabs) == 0 {
		return ""
	}
	return r.tabs[r.activeIdx].View.View()
}
