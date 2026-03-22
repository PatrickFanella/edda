package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PatrickFanella/game-master/internal/config"
	"github.com/PatrickFanella/game-master/tui"
	"github.com/PatrickFanella/game-master/tui/character"
	"github.com/PatrickFanella/game-master/tui/inventory"
	"github.com/PatrickFanella/game-master/tui/narrative"
	"github.com/PatrickFanella/game-master/tui/quest"
	"github.com/PatrickFanella/game-master/tui/styles"
)

// appModel is the root Bubble Tea model that hosts all sub-views and handles
// global key bindings (tab switching, quit).
type appModel struct {
	cfg    config.Config
	router *tui.Router
	width  int
	height int
}

func newApp(cfg config.Config) appModel {
	router := tui.NewRouter()

	nv := narrative.New()
	cv := character.New()
	iv := inventory.New()
	qv := quest.New()

	router.Register("Narrative", &nv)
	router.Register("Character", &cv)
	router.Register("Inventory", &iv)
	router.Register("Quests", &qv)

	// Seed the narrative log with example entries.
	nv.AddEntry(narrative.Entry{
		Kind: narrative.KindSystem,
		Text: fmt.Sprintf("Welcome to Game Master  ·  Provider: %s", cfg.LLM.Provider),
	})
	nv.AddEntry(narrative.Entry{
		Kind:    narrative.KindNPC,
		Speaker: "Innkeeper Brynn",
		Text:    "\"Ah, a traveller! You've arrived just in time — there's trouble on the east road.\"",
	})
	nv.AddEntry(narrative.Entry{
		Kind: narrative.KindPlayer,
		Text: "What kind of trouble?",
	})
	nv.AddEntry(narrative.Entry{
		Kind:    narrative.KindNPC,
		Speaker: "Innkeeper Brynn",
		Text:    "\"A merchant went missing three days ago. Cargo and all. Sheriff won't lift a finger.\"",
	})

	return appModel{cfg: cfg, router: router}
}

func (m appModel) Init() tea.Cmd { return nil }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.propagateSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "right", "l":
			m.router.NextTab()
		case "shift+tab", "left", "h":
			m.router.PrevTab()
		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			m.router.GoToTab(idx)
		default:
			cmd := m.router.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// chrome renders the title bar, tab bar, and status bar at the current width.
func (m *appModel) chrome() (titleBar, tabBar, statusBar string) {
	titleBar = styles.TitleBar.Width(m.width).Render(
		"⚔  Game Master" + styles.Muted.Render(
			fmt.Sprintf("  ·  %s", m.cfg.LLM.Provider),
		),
	)
	tabBar = m.renderTabs()
	hints := styles.Muted.Render("tab/←/→ switch view  ·  1–4 jump to view  ·  q quit")
	statusBar = styles.StatusBar.Width(m.width).Render(hints)
	return
}

// propagateSizes pushes the current terminal size down to all sub-views.
func (m *appModel) propagateSizes() {
	titleBar, tabBar, statusBar := m.chrome()

	reserved := lipgloss.Height(titleBar) + lipgloss.Height(tabBar) + lipgloss.Height(statusBar)
	viewHeight := m.height - reserved
	if viewHeight < 1 {
		viewHeight = 1
	}

	m.router.SetSize(m.width, viewHeight)
}

func (m appModel) View() string {
	titleBar, tabBar, statusBar := m.chrome()
	activeView := m.router.View()
	activeView = lipgloss.NewStyle().Width(m.width).Render(activeView)
	return styles.JoinVertical(titleBar, tabBar, activeView, statusBar)
}

func (m appModel) renderTabs() string {
	var tabs []string
	for i, tab := range m.router.Tabs() {
		label := fmt.Sprintf("%d %s", i+1, tab.Name)
		if i == m.router.ActiveTab() {
			tabs = append(tabs, styles.ActiveTab.Render(label))
		} else {
			tabs = append(tabs, styles.Tab.Render(label))
		}
	}
	return styles.JoinHorizontal(tabs...)
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	configPath, err := parseConfigPath(args, os.Getenv("GM_CONFIG"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
	})
	log.SetDefault(logger)

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Errorf("load config: %v", err)
		return 1
	}
	logger.Infof("starting TUI (provider=%s)", cfg.LLM.Provider)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p := tea.NewProgram(
		newApp(cfg),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)

	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received")
	}()

	if _, err := p.Run(); err != nil {
		if ctx.Err() != nil && (errors.Is(err, tea.ErrInterrupted) || errors.Is(err, tea.ErrProgramKilled)) {
			logger.Info("TUI shutdown complete")
			return 0
		}
		logger.Errorf("tui error: %v", err)
		return 1
	}

	logger.Info("TUI stopped")
	return 0
}

func parseConfigPath(args []string, defaultPath string) (string, error) {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := defaultPath
	fs.StringVar(&configPath, "config", configPath, "Path to config file")

	if err := fs.Parse(args); err != nil {
		return "", err
	}
	return configPath, nil
}
