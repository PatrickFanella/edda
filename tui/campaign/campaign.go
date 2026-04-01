// Package campaign provides the campaign-selection TUI view used during
// Game Master start-up. It presents a list of existing campaigns plus a
// "New campaign" option; when "New campaign" is selected, it shows a multi-step
// Huh form that collects campaign name, genre, and difficulty before proceeding.
package campaign

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/tui/styles"
)

const newCampaignSentinel = "__new__"

// SelectedMsg is sent when the player has chosen a campaign. Campaign
// carries the full selected campaign record (including its ID, name, etc.).
type SelectedMsg struct {
	Campaign statedb.Campaign
}

// item wraps a campaign row for use in the bubbles list.
type item struct {
	id   string // UUID hex string or newCampaignSentinel
	name string
	desc string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.name }

// Model is the Bubble Tea model for campaign selection.
type Model struct {
	campaigns []statedb.Campaign
	list      list.Model
	form       *huh.Form          // non-nil when "New campaign" form is open
	formResult CampaignFormResult  // values collected by the form
	width     int
	height    int
}

// New builds the campaign-selection model from the provided campaigns.
// The "New campaign" option is always appended to the end of the list.
func New(campaigns []statedb.Campaign) Model {
	items := make([]list.Item, 0, len(campaigns)+1)
	for _, c := range campaigns {
		desc := formatCampaignDescription(c)
		items = append(items, item{
			id:   c.ID.String(),
			name: c.Name,
			desc: desc,
		})
	}
	items = append(items, item{
		id:   newCampaignSentinel,
		name: "✦ New Campaign",
		desc: "Create a fresh adventure",
	})

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(styles.ColorAccent).
		BorderForeground(styles.ColorAccent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.ColorAccentDim).
		BorderForeground(styles.ColorAccent)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Select Campaign"
	l.Styles.Title = styles.Header
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Model{
		campaigns: campaigns,
		list:      l,
	}
}

// SetSize implements tui.View.
func (m *Model) SetSize(width, height int) {
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
	if m.form != nil {
		m.form = m.form.WithWidth(width)
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.form != nil {
		return m.updateForm(msg)
	}
	return m.updateList(msg)
}

// View implements tea.Model.
func (m Model) View() string {
	if m.form != nil {
		return m.renderForm()
	}
	return m.renderList()
}

// updateList processes messages while the campaign list is visible.
func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			selected, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			if selected.id == newCampaignSentinel {
				m.form = buildCampaignForm(&m.formResult)
				return m, m.form.Init()
			}
			// Find the matching campaign and emit SelectedMsg.
			for _, c := range m.campaigns {
				if c.ID.String() == selected.id {
					return m, func() tea.Msg { return SelectedMsg{Campaign: c} }
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// updateForm processes messages while the "new campaign" name form is open.
func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		result := m.formResult
		result.Name = strings.TrimSpace(result.Name)
		m.form = nil
		return m, func() tea.Msg { return NewCampaignFormMsg{Result: result} }
	}
	if m.form.State == huh.StateAborted {
		m.form = nil
	}
	return m, cmd
}

// renderList renders the campaign list.
func (m Model) renderList() string {
	inner := m.list.View()
	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(inner)
}

// renderForm renders the new-campaign name form.
func (m Model) renderForm() string {
	title := styles.Header.Render("✦ New Campaign")
	hint := styles.Muted.Render("Enter a name for your new adventure, then press Enter.")

	formView := m.form.View()

	content := styles.JoinVertical(title, "", hint, "", formView)

	// Account for horizontal padding (2 on each side) and clamp to avoid
	// negative widths in very small terminals.
	innerWidth := m.width - 4
	if innerWidth < 0 {
		innerWidth = 0
	}
	return styles.FocusedContainer.
		Width(m.width).
		Height(m.height).
		Render(lipgloss.NewStyle().
			Padding(1, 2).
			Width(innerWidth).
			Render(content))
}


func formatCampaignDescription(c statedb.Campaign) string {
	genre := "Unknown genre"
	if c.Genre.Valid && strings.TrimSpace(c.Genre.String) != "" {
		genre = strings.TrimSpace(c.Genre.String)
	}

	lastPlayed := "Never"
	if c.UpdatedAt.Valid && !c.UpdatedAt.Time.IsZero() {
		lastPlayed = c.UpdatedAt.Time.UTC().Format("2006-01-02")
	}

	status := c.Status
	if strings.TrimSpace(status) == "" {
		status = "unknown"
	}

	return fmt.Sprintf("Genre: %s · Last played: %s · Status: %s", genre, lastPlayed, status)
}
