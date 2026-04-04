package campaign

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/PatrickFanella/game-master/tui/styles"
)

type attributeFormResult struct {
	Genre        string
	SettingStyle string
	Tone         string
}

// AttributesModel is a huh.Form-backed sub-model for selecting campaign
// genre, setting style, and tone via guided dropdowns.
type AttributesModel struct {
	form      *huh.Form
	result    attributeFormResult
	errorText string
	width     int
	height    int
}

// NewAttributesModel builds the three-step attribute selection form.
func NewAttributesModel() AttributesModel {
	return NewAttributesModelWithState("", "", "", "")
}

// NewAttributesModelWithValues builds the attribute form with preselected
// values. It is used when proposal generation fails or the player navigates
// back from the proposal picker so the previous selections are preserved.
func NewAttributesModelWithValues(genre, settingStyle, tone string) AttributesModel {
	return NewAttributesModelWithState(genre, settingStyle, tone, "")
}

// NewAttributesModelWithState builds the attribute form with preselected values
// and an inline error message displayed above the form body when non-empty.
func NewAttributesModelWithState(genre, settingStyle, tone, errorText string) AttributesModel {
	result := attributeFormResult{
		Genre:        genre,
		SettingStyle: settingStyle,
		Tone:         tone,
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Genre").
				Options(
					huh.NewOption("Fantasy", "Fantasy"),
					huh.NewOption("Sci-Fi", "Sci-Fi"),
					huh.NewOption("Horror", "Horror"),
					huh.NewOption("Historical", "Historical"),
					huh.NewOption("Modern", "Modern"),
					huh.NewOption("Post-Apocalyptic", "Post-Apocalyptic"),
					huh.NewOption("Steampunk", "Steampunk"),
				).
				Value(&result.Genre),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Setting Style").
				Options(
					huh.NewOption("Open Wilderness", "Open Wilderness"),
					huh.NewOption("Urban Sprawl", "Urban Sprawl"),
					huh.NewOption("Island Archipelago", "Island Archipelago"),
					huh.NewOption("Underground", "Underground"),
					huh.NewOption("War-Torn Kingdom", "War-Torn Kingdom"),
					huh.NewOption("Peaceful Realm", "Peaceful Realm"),
				).
				Value(&result.SettingStyle),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Tone").
				Options(
					huh.NewOption("Gritty and Dark", "Gritty and Dark"),
					huh.NewOption("Light-Hearted", "Light-Hearted"),
					huh.NewOption("Epic and Grand", "Epic and Grand"),
					huh.NewOption("Mysterious", "Mysterious"),
					huh.NewOption("Humorous", "Humorous"),
					huh.NewOption("Tense and Suspenseful", "Tense and Suspenseful"),
				).
				Value(&result.Tone),
		),
	).WithShowHelp(false)

	return AttributesModel{
		form:      form,
		result:    result,
		errorText: errorText,
	}
}

// SetSize updates the available rendering area.
func (m *AttributesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init implements tea.Model.
func (m AttributesModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update implements tea.Model.
func (m AttributesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	switch m.form.State {
	case huh.StateCompleted:
		return m, func() tea.Msg {
			return AttributesReadyMsg{
				Genre:        m.result.Genre,
				SettingStyle: m.result.SettingStyle,
				Tone:         m.result.Tone,
			}
		}
	case huh.StateAborted:
		return m, func() tea.Msg { return BackMsg{} }
	}

	return m, cmd
}

// View implements tea.Model.
func (m AttributesModel) View() string {
	title := styles.Header.Render("✦ Campaign Attributes")
	body := m.form.View()
	content := styles.JoinVertical(title, "")
	if m.errorText != "" {
		content = styles.JoinVertical(content, styles.StatusError.Render("Error: "+m.errorText), "")
	}
	content = styles.JoinVertical(content, body)
	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(content)
}
