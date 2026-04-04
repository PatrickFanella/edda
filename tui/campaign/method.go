package campaign

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PatrickFanella/game-master/tui/styles"
)

// methodItem is a list item for the campaign creation method selector.
type methodItem struct {
	title string
	desc  string
	value CreationMethod
}

func (i methodItem) Title() string       { return i.title }
func (i methodItem) Description() string { return i.desc }
func (i methodItem) FilterValue() string { return i.title }

// MethodModel lets the player choose between free-form description
// and guided attribute selection for campaign creation.
type MethodModel struct {
	list   list.Model
	width  int
	height int
}

// NewMethodModel builds the method selection model.
func NewMethodModel() MethodModel {
	items := []list.Item{
		methodItem{
			title: "✦ Describe Your World",
			desc:  "Tell us about your ideal campaign and we'll build it",
			value: MethodDescribe,
		},
		methodItem{
			title: "✦ Choose Attributes",
			desc:  "Select genre, setting, and tone from a menu",
			value: MethodAttributes,
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(styles.ColorAccent).
		BorderForeground(styles.ColorAccent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.ColorAccentDim).
		BorderForeground(styles.ColorAccent)

	l := list.New(items, delegate, 0, 0)
	l.Title = "How would you like to create your campaign?"
	l.Styles.Title = styles.Header
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return MethodModel{list: l}
}

// SetSize updates the model dimensions.
func (m *MethodModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listWidth := width - 4
	if listWidth < 0 {
		listWidth = 0
	}
	listHeight := height - 4
	if listHeight < 0 {
		listHeight = 0
	}
	m.list.SetSize(listWidth, listHeight)
}

// Init implements tea.Model.
func (m MethodModel) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m MethodModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			selected, ok := m.list.SelectedItem().(methodItem)
			if !ok {
				return m, nil
			}
			method := selected.value
			return m, func() tea.Msg { return MethodChosenMsg{Method: method} }
		case tea.KeyEsc:
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m MethodModel) View() string {
	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(m.list.View())
}
