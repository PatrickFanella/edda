package campaign

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/world"
	"github.com/PatrickFanella/game-master/tui/styles"
)

// charInterviewResponseMsg carries the result of a CharacterInterviewer
// Start/Step call.
type charInterviewResponseMsg struct {
	message string
	profile *world.CharacterProfile
	done    bool
	err     error
}

// CharInterviewModel wraps world.CharacterInterviewer as a chat-like
// Bubbletea model. It emits CharacterReadyMsg when the interview completes
// or BackMsg on escape.
type CharInterviewModel struct {
	interviewer *world.CharacterInterviewer
	ctx         context.Context
	viewport    viewport.Model
	input       textinput.Model
	messages    []string // chat history lines
	waiting     bool
	spinner     spinner.Model
	profile     *world.CharacterProfile
	done        bool
	err         error
	width       int
	height      int
}

// NewCharInterviewModel creates a CharInterviewModel ready to begin the
// character profile interview.
func NewCharInterviewModel(ctx context.Context, provider llm.Provider, campaignProfile *world.CampaignProfile) CharInterviewModel {
	ci := world.NewCharacterInterviewer(provider, campaignProfile)

	vp := viewport.New(0, 0)

	ti := textinput.New()
	ti.Placeholder = "Describe your character…"
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorAccent)

	return CharInterviewModel{
		interviewer: ci,
		ctx:         ctx,
		viewport:    vp,
		input:       ti,
		spinner:     sp,
	}
}

// SetSize adjusts the viewport and input to the available area.
func (m *CharInterviewModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	const borderPad = 4
	const inputArea = 2
	vpWidth := width - borderPad
	vpHeight := height - borderPad - inputArea
	if vpWidth < 0 {
		vpWidth = 0
	}
	if vpHeight < 0 {
		vpHeight = 0
	}
	m.viewport.Width = vpWidth
	m.viewport.Height = vpHeight
	m.input.Width = vpWidth
	m.refreshViewport()
}

// Init kicks off the interview by calling CharacterInterviewer.Start
// asynchronously.
func (m CharInterviewModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startCmd(),
	)
}

// Update implements tea.Model.
func (m CharInterviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case charInterviewResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.err = msg.err
			m.messages = append(m.messages,
				styles.StatusError.Render(fmt.Sprintf("Error: %v — press Enter to retry", msg.err)))
			m.refreshViewport()
			return m, nil
		}
		m.messages = append(m.messages, styles.Muted.Render("GM: ")+msg.message)
		m.refreshViewport()
		if msg.done && msg.profile != nil {
			m.done = true
			m.profile = msg.profile
			p := *msg.profile
			return m, func() tea.Msg {
				return CharacterReadyMsg{Profile: p}
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if !m.waiting {
				return m, func() tea.Msg { return BackMsg{} }
			}
		case tea.KeyEnter:
			if m.waiting || m.done {
				return m, nil
			}
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			m.err = nil
			m.messages = append(m.messages, styles.Header.Render("You: ")+text)
			m.input.SetValue("")
			m.waiting = true
			m.refreshViewport()
			return m, tea.Batch(m.spinner.Tick, m.stepCmd(text))
		}

	case spinner.TickMsg:
		if m.waiting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m CharInterviewModel) View() string {
	var b strings.Builder

	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	if m.waiting {
		b.WriteString(m.spinner.View() + " Thinking…")
	} else {
		b.WriteString(m.input.View())
	}

	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(b.String())
}

// refreshViewport rebuilds the viewport content from the chat history.
func (m *CharInterviewModel) refreshViewport() {
	content := strings.Join(m.messages, "\n\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

// startCmd returns a tea.Cmd that calls CharacterInterviewer.Start.
func (m *CharInterviewModel) startCmd() tea.Cmd {
	ci := m.interviewer
	ctx := m.ctx
	return func() tea.Msg {
		res, err := ci.Start(ctx)
		if err != nil {
			return charInterviewResponseMsg{err: err}
		}
		return charInterviewResponseMsg{
			message: res.Message,
			profile: res.Profile,
			done:    res.Done,
		}
	}
}

// stepCmd returns a tea.Cmd that calls CharacterInterviewer.Step with the
// given input.
func (m *CharInterviewModel) stepCmd(input string) tea.Cmd {
	ci := m.interviewer
	ctx := m.ctx
	return func() tea.Msg {
		res, err := ci.Step(ctx, input)
		if err != nil {
			return charInterviewResponseMsg{err: err}
		}
		return charInterviewResponseMsg{
			message: res.Message,
			profile: res.Profile,
			done:    res.Done,
		}
	}
}
