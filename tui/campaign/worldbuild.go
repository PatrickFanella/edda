package campaign

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PatrickFanella/game-master/internal/world"
	"github.com/PatrickFanella/game-master/tui/styles"
)

// worldStageMsg carries an intermediate progress update from the orchestrator.
type worldStageMsg struct {
	stage string
}

// worldDoneMsg is the internal message sent when the orchestrator finishes.
type worldDoneMsg struct {
	result *world.OrchestratorResult
	err    error
}

// WorldBuildModel shows a spinner while the orchestrator generates the world.
type WorldBuildModel struct {
	ctx          context.Context
	orchestrator *world.Orchestrator
	input        world.OrchestratorInput
	spinner      spinner.Model
	stage        string
	progressCh   chan string
	done         bool
	err          error
	width        int
	height       int
}

// NewWorldBuildModel builds the world-generation model.
func NewWorldBuildModel(ctx context.Context, orchestrator *world.Orchestrator, input world.OrchestratorInput) WorldBuildModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorAccent)

	return WorldBuildModel{
		ctx:          ctx,
		orchestrator: orchestrator,
		input:        input,
		spinner:      s,
		stage:        "Building your world…",
		progressCh:   make(chan string, 1),
	}
}

// SetSize updates the layout dimensions.
func (m *WorldBuildModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init implements tea.Model.
func (m WorldBuildModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.listenForStage(), m.runOrchestrator())
}

// runOrchestrator returns a tea.Cmd that blocks until generation completes.
func (m WorldBuildModel) runOrchestrator() tea.Cmd {
	ctx := m.ctx
	orch := m.orchestrator
	input := m.input
	progressCh := m.progressCh
	return func() tea.Msg {
		result, err := orch.Run(ctx, input, func(stage string) {
			publishStage(progressCh, stage)
		})
		close(progressCh)
		return worldDoneMsg{result: result, err: err}
	}
}

// listenForStage waits for the next progress update from the orchestrator.
func (m WorldBuildModel) listenForStage() tea.Cmd {
	progressCh := m.progressCh
	return func() tea.Msg {
		stage, ok := <-progressCh
		if !ok {
			return nil
		}
		return worldStageMsg{stage: stage}
	}
}

// Update implements tea.Model.
func (m WorldBuildModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case worldStageMsg:
		if msg.stage != "" {
			m.stage = msg.stage
		}
		if m.done {
			return m, nil
		}
		return m, m.listenForStage()
	case worldDoneMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
			return m, func() tea.Msg { return WorldErrorMsg{Err: msg.err} }
		}
		result := msg.result
		return m, func() tea.Msg { return WorldReadyMsg{Result: result} }
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View implements tea.Model.
func (m WorldBuildModel) View() string {
	var content string
	if m.err != nil {
		content = styles.StatusError.Render("Error: " + m.err.Error())
	} else {
		content = m.spinner.View() + "  " + styles.Body.Render(m.stage)
	}

	return styles.Container.
		Width(m.width).
		Height(m.height).
		Render(styles.Place(m.width, m.height, content))
}

func publishStage(ch chan string, stage string) {
	select {
	case ch <- stage:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- stage:
	default:
	}
}
