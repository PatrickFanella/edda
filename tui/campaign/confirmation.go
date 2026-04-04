package campaign

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PatrickFanella/game-master/internal/world"
	"github.com/PatrickFanella/game-master/tui/styles"
)

// ConfirmationModel shows campaign + character details for final review.
type ConfirmationModel struct {
	name             string
	summary          string
	profile          world.CampaignProfile
	characterProfile world.CharacterProfile
	selected         int // 0 = Begin Adventure, 1 = Go Back
	width            int
	height           int
}

// NewConfirmationModel builds the confirmation review model.
func NewConfirmationModel(name, summary string, profile world.CampaignProfile, charProfile world.CharacterProfile) ConfirmationModel {
	return ConfirmationModel{
		name:             name,
		summary:          summary,
		profile:          profile,
		characterProfile: charProfile,
	}
}

// SetSize updates the layout dimensions.
func (m *ConfirmationModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init implements tea.Model.
func (m ConfirmationModel) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m ConfirmationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyLeft:
			m.selected = 0
			return m, nil
		case tea.KeyDown, tea.KeyRight:
			m.selected = 1
			return m, nil
		case tea.KeyEnter:
			if m.selected == 0 {
				return m, func() tea.Msg {
					return ConfirmedMsg{
						Name:             m.name,
						Summary:          m.summary,
						Profile:          m.profile,
						CharacterProfile: m.characterProfile,
					}
				}
			}
			return m, func() tea.Msg { return ChangeMsg{} }
		case tea.KeyEsc:
			return m, func() tea.Msg { return ChangeMsg{} }
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				m.selected = 1
			case "k":
				m.selected = 0
			}
			return m, nil
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m ConfirmationModel) View() string {
	var b strings.Builder

	b.WriteString(styles.Header.Render("✦ Campaign Summary"))
	b.WriteString("\n\n")
	b.WriteString(styles.Body.Render(fmt.Sprintf("Name: %s", m.name)))
	b.WriteString("\n")
	b.WriteString(styles.Body.Render(fmt.Sprintf("Summary: %s", m.summary)))
	b.WriteString("\n")
	b.WriteString(styles.Body.Render(
		fmt.Sprintf("Genre: %s · Tone: %s", m.profile.Genre, m.profile.Tone),
	))
	b.WriteString("\n")
	if len(m.profile.Themes) > 0 {
		b.WriteString(styles.Body.Render(
			fmt.Sprintf("Themes: %s", strings.Join(m.profile.Themes, ", ")),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.Header.Render("✦ Your Character"))
	b.WriteString("\n\n")
	b.WriteString(styles.Body.Render(fmt.Sprintf("Name: %s", m.characterProfile.Name)))
	b.WriteString("\n")
	b.WriteString(styles.Body.Render(fmt.Sprintf("Concept: %s", m.characterProfile.Concept)))
	b.WriteString("\n")
	b.WriteString(styles.Body.Render(fmt.Sprintf("Background: %s", m.characterProfile.Background)))
	b.WriteString("\n\n")

	// Buttons
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorAccent)
	inactiveStyle := styles.Muted

	beginLabel := "  Begin Adventure  "
	backLabel := "  Go Back  "

	var beginBtn, backBtn string
	if m.selected == 0 {
		beginBtn = activeStyle.Render("[▸" + beginLabel + "]")
		backBtn = inactiveStyle.Render("[ " + backLabel + "]")
	} else {
		beginBtn = inactiveStyle.Render("[ " + beginLabel + "]")
		backBtn = activeStyle.Render("[▸" + backLabel + "]")
	}

	b.WriteString(styles.JoinHorizontal(beginBtn, "    ", backBtn))

	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(b.String())
}
