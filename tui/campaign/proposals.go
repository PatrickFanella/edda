package campaign

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PatrickFanella/game-master/internal/world"
	"github.com/PatrickFanella/game-master/tui/styles"
)

// proposalItem wraps a CampaignProposal for the bubbles list.
type proposalItem struct {
	proposal world.CampaignProposal
}

func (i proposalItem) Title() string       { return i.proposal.Name }
func (i proposalItem) Description() string { return i.proposal.Summary }
func (i proposalItem) FilterValue() string { return i.proposal.Name }

// ProposalsModel displays generated campaign proposals for selection.
type ProposalsModel struct {
	proposals []world.CampaignProposal
	list      list.Model
	width     int
	height    int
}

// NewProposalsModel builds the proposal-selection model.
func NewProposalsModel(proposals []world.CampaignProposal) ProposalsModel {
	items := make([]list.Item, len(proposals))
	for i, p := range proposals {
		items[i] = proposalItem{proposal: p}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(styles.ColorAccent).
		BorderForeground(styles.ColorAccent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.ColorAccentDim).
		BorderForeground(styles.ColorAccent)

	l := list.New(items, delegate, 0, 0)
	l.Title = "✦ Choose Your Campaign"
	l.Styles.Title = styles.Header
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return ProposalsModel{
		proposals: proposals,
		list:      l,
	}
}

// SetSize updates the layout dimensions.
func (m *ProposalsModel) SetSize(width, height int) {
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
func (m ProposalsModel) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m ProposalsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			selected, ok := m.list.SelectedItem().(proposalItem)
			if !ok {
				return m, nil
			}
			p := selected.proposal
			return m, func() tea.Msg { return ProposalSelectedMsg{Proposal: p} }
		case tea.KeyEsc:
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m ProposalsModel) View() string {
	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(m.list.View())
}
