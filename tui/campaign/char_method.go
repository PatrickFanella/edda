package campaign

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PatrickFanella/game-master/tui/styles"
)

// charMethodItem is a list item for the character creation method selector.
type charMethodItem struct {
	title string
	desc  string
	value CreationMethod
}

func (i charMethodItem) Title() string       { return i.title }
func (i charMethodItem) Description() string { return i.desc }
func (i charMethodItem) FilterValue() string { return i.title }

// CharMethodModel lets the player choose between conversational description
// and guided step-by-step character creation.
type CharMethodModel struct {
	list   list.Model
	width  int
	height int
}

// NewCharMethodModel builds the character method selection model.
func NewCharMethodModel() CharMethodModel {
	items := []list.Item{
		charMethodItem{
			title: "✦ Describe Your Character",
			desc:  "Tell us about your character through a conversation",
			value: MethodDescribe,
		},
		charMethodItem{
			title: "✦ Guided Creation",
			desc:  "Choose race, class, and background step by step",
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
	l.Title = "How would you like to create your character?"
	l.Styles.Title = styles.Header
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return CharMethodModel{list: l}
}

// SetSize updates the model dimensions.
func (m *CharMethodModel) SetSize(width, height int) {
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
func (m CharMethodModel) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m CharMethodModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			selected, ok := m.list.SelectedItem().(charMethodItem)
			if !ok {
				return m, nil
			}
			method := selected.value
			return m, func() tea.Msg { return CharMethodChosenMsg{Method: method} }
		case tea.KeyEsc:
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m CharMethodModel) View() string {
	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(m.list.View())
}
