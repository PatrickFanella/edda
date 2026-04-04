// Package campaign provides the campaign-selection and creation wizard TUI views.
package campaign

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/tui/styles"
)

const newCampaignSentinel = "__new__"

// item wraps a campaign row for use in the bubbles list.
type item struct {
	id   string // UUID hex string or newCampaignSentinel
	name string
	desc string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.name }

// Picker is the Bubble Tea model for campaign selection. It shows existing
// campaigns plus a "New Campaign" sentinel. It emits SelectedMsg when an
// existing campaign is chosen or NewCampaignMsg for new creation.
type Picker struct {
	campaigns []statedb.Campaign
	list      list.Model
	width     int
	height    int
}

// NewPicker builds the campaign-selection model from the provided campaigns.
// The "New campaign" option is always appended to the end of the list.
func NewPicker(campaigns []statedb.Campaign) Picker {
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

	return Picker{
		campaigns: campaigns,
		list:      l,
	}
}

// SetSize implements tui.View.
func (m *Picker) SetSize(width, height int) {
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
func (m Picker) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			selected, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			if selected.id == newCampaignSentinel {
				return m, func() tea.Msg { return NewCampaignMsg{} }
			}
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

// View implements tea.Model.
func (m Picker) View() string {
	inner := m.list.View()
	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(inner)
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
