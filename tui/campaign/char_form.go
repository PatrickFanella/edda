package campaign

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/PatrickFanella/game-master/internal/world"
	"github.com/PatrickFanella/game-master/tui/styles"
)

type charFormResult struct {
	Name       string
	Race       string
	Class      string
	Background string
	Alignment  string
}

// CharFormModel is a huh.Form-backed sub-model for building a D&D 5e
// character through guided attribute selection.
type CharFormModel struct {
	form   *huh.Form
	result charFormResult
	width  int
	height int
}

func selectOptions(values []string) []huh.Option[string] {
	opts := make([]huh.Option[string], len(values))
	for i, v := range values {
		opts[i] = huh.NewOption(v, v)
	}
	return opts
}

// NewCharFormModel builds the five-step character creation form.
func NewCharFormModel() CharFormModel {
	var result charFormResult

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Character Name").
				Value(&result.Name).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return huh.ErrTimeout // non-nil to block progression
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Race").
				Options(selectOptions(world.D5ERaces)...).
				Value(&result.Race),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Class").
				Options(selectOptions(world.D5EClasses)...).
				Value(&result.Class),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Background").
				Options(selectOptions(world.D5EBackgrounds)...).
				Value(&result.Background),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Alignment").
				Options(selectOptions(world.D5EAlignments)...).
				Value(&result.Alignment),
		),
	).WithShowHelp(false)

	return CharFormModel{
		form:   form,
		result: result,
	}
}

// SetSize updates the available rendering area.
func (m *CharFormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init implements tea.Model.
func (m CharFormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update implements tea.Model.
func (m CharFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	switch m.form.State {
	case huh.StateCompleted:
		profile := world.BuildCharacterFromAttributes(world.CharacterAttributes{
			Name:       strings.TrimSpace(m.result.Name),
			Race:       m.result.Race,
			Class:      m.result.Class,
			Background: m.result.Background,
			Alignment:  m.result.Alignment,
		})
		return m, func() tea.Msg {
			return CharacterReadyMsg{Profile: *profile}
		}
	case huh.StateAborted:
		return m, func() tea.Msg { return BackMsg{} }
	}

	return m, cmd
}

// View implements tea.Model.
func (m CharFormModel) View() string {
	title := styles.Header.Render("✦ Character Creation")
	body := m.form.View()
	content := styles.JoinVertical(title, "", body)
	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(content)
}
