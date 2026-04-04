// Package tui – launcher model.
package tui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PatrickFanella/game-master/internal/bootstrap"
	"github.com/PatrickFanella/game-master/internal/config"
	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/engine"
	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/logging"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/world"
	"github.com/PatrickFanella/game-master/tui/campaign"
	"github.com/PatrickFanella/game-master/tui/styles"
)

// bootstrapDoneMsg carries the result of the DB bootstrap step.
type bootstrapDoneMsg struct {
	result bootstrap.Result
	err    error
}

// proposalsGeneratedMsg carries generated campaign proposals.
type proposalsGeneratedMsg struct {
	proposals []world.CampaignProposal
	err       error
}

// campaignNamedMsg carries a generated campaign name.
type campaignNamedMsg struct {
	name string
	err  error
}

// campaignLoadedMsg is sent after a campaign has been loaded in the engine.
type campaignLoadedMsg struct {
	c   statedb.Campaign
	err error
}

// launcherState is the internal phase of the Launcher model.
type launcherState int

const (
	launcherLoading launcherState = iota
	launcherSelecting
	launcherChooseMethod
	launcherInterviewing
	launcherAttributes
	launcherGeneratingProposals
	launcherProposals
	launcherGeneratingName
	launcherCharMethod
	launcherCharInterview
	launcherCharForm
	launcherConfirmation
	launcherWorldBuilding
	launcherLoadingCampaign
)

// Launcher is the root Bubble Tea model during start-up and campaign creation.
type Launcher struct {
	cfg      config.Config
	ctx      context.Context
	engine   engine.GameEngine
	queries  statedb.Querier
	provider llm.Provider
	logBuf   *logging.RingBuffer

	user  statedb.User
	state launcherState
	view  View

	campaignProfile  *world.CampaignProfile
	characterProfile *world.CharacterProfile
	proposalAttrs    campaign.AttributesReadyMsg
	campaignName     string
	campaignSummary  string
	openingScene     *world.SceneResult

	spinner spinner.Model
	errMsg  string
	width   int
	height  int
}

// LauncherOption configures the Launcher.
type LauncherOption func(*Launcher)

// WithLogBuffer attaches a RingBuffer so the log panel can display entries.
func WithLogBuffer(buf *logging.RingBuffer) LauncherOption {
	return func(l *Launcher) { l.logBuf = buf }
}

// WithLLMProvider attaches the LLM provider used by the creation wizard.
func WithLLMProvider(provider llm.Provider) LauncherOption {
	return func(l *Launcher) { l.provider = provider }
}

// NewLauncher creates the Launcher model.
func NewLauncher(cfg config.Config, ctx context.Context, queries statedb.Querier) Launcher {
	return NewLauncherWithEngine(cfg, ctx, queries, nil)
}

// NewLauncherWithEngine creates the Launcher model with a game engine dependency.
func NewLauncherWithEngine(cfg config.Config, ctx context.Context, queries statedb.Querier, gameEngine engine.GameEngine, opts ...LauncherOption) Launcher {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorAccent)

	l := Launcher{
		cfg:     cfg,
		ctx:     ctx,
		engine:  gameEngine,
		queries: queries,
		state:   launcherLoading,
		spinner: sp,
	}
	for _, opt := range opts {
		opt(&l)
	}
	return l
}

func (l Launcher) logger() *slog.Logger {
	return slog.Default().WithGroup("launcher")
}

// Init implements tea.Model.
func (l Launcher) Init() tea.Cmd {
	return tea.Batch(l.spinner.Tick, l.runBootstrap())
}

// Update implements tea.Model.
func (l Launcher) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		if l.view != nil {
			l.view.SetSize(msg.Width, msg.Height)
		}
		return l, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return l, tea.Quit
		}

	case spinner.TickMsg:
		if l.usesRootSpinner() {
			var cmd tea.Cmd
			l.spinner, cmd = l.spinner.Update(msg)
			return l, cmd
		}

	case bootstrapDoneMsg:
		if msg.err != nil {
			l.errMsg = fmt.Sprintf("Bootstrap failed: %v", msg.err)
			l.state = launcherLoading
			l.view = nil
			return l, nil
		}
		l.user = msg.result.User
		l.logger().Info("bootstrap complete", "campaigns", len(msg.result.Campaigns), "user", msg.result.User.Name)
		picker := campaign.NewPicker(msg.result.Campaigns)
		return l.activateView(launcherSelecting, &picker)

	case campaign.SelectedMsg:
		l.errMsg = ""
		l.openingScene = nil
		l.logger().Info("loading existing campaign", "campaign_id", dbutil.FromPgtype(msg.Campaign.ID), "campaign", msg.Campaign.Name)
		l.state = launcherLoadingCampaign
		l.view = nil
		return l, tea.Batch(l.spinner.Tick, l.runLoadCampaign(msg.Campaign))

	case campaign.NewCampaignMsg:
		l.resetCreationState()
		l.errMsg = ""
		l.logger().Info("starting new campaign wizard")
		method := campaign.NewMethodModel()
		return l.activateView(launcherChooseMethod, &method)

	case campaign.MethodChosenMsg:
		l.errMsg = ""
		l.logger().Info("campaign creation method chosen", "method", creationMethodLabel(msg.Method))
		switch msg.Method {
		case campaign.MethodDescribe:
			if l.provider == nil {
				l.errMsg = "LLM provider unavailable"
				return l, nil
			}
			iv := campaign.NewInterviewModel(l.ctx, l.provider)
			return l.activateView(launcherInterviewing, &iv)
		case campaign.MethodAttributes:
			attrs := campaign.NewAttributesModelWithValues(l.proposalAttrs.Genre, l.proposalAttrs.SettingStyle, l.proposalAttrs.Tone)
			return l.activateView(launcherAttributes, &attrs)
		}

	case campaign.ProfileReadyMsg:
		profile := msg.Profile
		l.campaignProfile = &profile
		l.campaignSummary = summarizeCampaignProfile(profile)
		l.state = launcherGeneratingName
		l.view = nil
		return l, tea.Batch(l.spinner.Tick, l.runGenerateCampaignName(&profile))

	case campaignNamedMsg:
		if msg.err != nil {
			l.errMsg = fmt.Sprintf("Generate campaign name failed: %v", msg.err)
			method := campaign.NewMethodModel()
			return l.activateView(launcherChooseMethod, &method)
		}
		l.campaignName = msg.name
		l.errMsg = ""
		charMethod := campaign.NewCharMethodModel()
		return l.activateView(launcherCharMethod, &charMethod)

	case campaign.AttributesReadyMsg:
		if l.provider == nil {
			l.errMsg = "LLM provider unavailable"
			return l, nil
		}
		l.proposalAttrs = msg
		l.logger().Info("generating campaign proposals", "genre", msg.Genre, "setting_style", msg.SettingStyle, "tone", msg.Tone)
		l.errMsg = ""
		l.state = launcherGeneratingProposals
		l.view = nil
		return l, tea.Batch(l.spinner.Tick, l.runGenerateProposals(msg))

	case proposalsGeneratedMsg:
		if msg.err != nil {
			l.errMsg = fmt.Sprintf("Generate proposals failed: %v", msg.err)
			l.logger().Error("campaign proposal generation failed",
				"genre", l.proposalAttrs.Genre,
				"setting_style", l.proposalAttrs.SettingStyle,
				"tone", l.proposalAttrs.Tone,
				"error", msg.err,
			)
			attrs := campaign.NewAttributesModelWithState(l.proposalAttrs.Genre, l.proposalAttrs.SettingStyle, l.proposalAttrs.Tone, l.errMsg)
			return l.activateView(launcherAttributes, &attrs)
		}
		l.errMsg = ""
		l.logger().Info("campaign proposals generated", "count", len(msg.proposals))
		proposals := campaign.NewProposalsModel(msg.proposals)
		return l.activateView(launcherProposals, &proposals)

	case campaign.ProposalSelectedMsg:
		proposal := msg.Proposal
		l.campaignProfile = &proposal.Profile
		l.campaignName = proposal.Name
		l.campaignSummary = proposal.Summary
		l.errMsg = ""
		l.logger().Info("campaign proposal selected", "campaign", proposal.Name)
		charMethod := campaign.NewCharMethodModel()
		return l.activateView(launcherCharMethod, &charMethod)

	case campaign.CharMethodChosenMsg:
		l.errMsg = ""
		switch msg.Method {
		case campaign.MethodDescribe:
			if l.provider == nil {
				l.errMsg = "LLM provider unavailable"
				return l, nil
			}
			if l.campaignProfile == nil {
				l.errMsg = "Campaign profile missing"
				method := campaign.NewMethodModel()
				return l.activateView(launcherChooseMethod, &method)
			}
			iv := campaign.NewCharInterviewModel(l.ctx, l.provider, l.campaignProfile)
			return l.activateView(launcherCharInterview, &iv)
		case campaign.MethodAttributes:
			form := campaign.NewCharFormModel()
			return l.activateView(launcherCharForm, &form)
		}

	case campaign.CharacterReadyMsg:
		profile := msg.Profile
		l.characterProfile = &profile
		if l.campaignProfile == nil {
			l.errMsg = "Campaign profile missing"
			method := campaign.NewMethodModel()
			return l.activateView(launcherChooseMethod, &method)
		}
		confirm := campaign.NewConfirmationModel(l.campaignName, l.campaignSummary, *l.campaignProfile, profile)
		return l.activateView(launcherConfirmation, &confirm)

	case campaign.ChangeMsg:
		l.errMsg = ""
		charMethod := campaign.NewCharMethodModel()
		return l.activateView(launcherCharMethod, &charMethod)

	case campaign.ConfirmedMsg:
		if l.provider == nil {
			l.errMsg = "LLM provider unavailable"
			return l, nil
		}
		l.errMsg = ""
		profile := msg.Profile
		characterProfile := msg.CharacterProfile
		l.campaignProfile = &profile
		l.characterProfile = &characterProfile
		l.campaignName = msg.Name
		l.campaignSummary = msg.Summary
		l.logger().Info("building new campaign world", "campaign", msg.Name)
		wb := campaign.NewWorldBuildModel(l.ctx, world.NewOrchestrator(l.provider, l.queries), world.OrchestratorInput{
			Name:             msg.Name,
			Summary:          msg.Summary,
			Profile:          &profile,
			CharacterProfile: &characterProfile,
			UserID:           dbutil.FromPgtype(l.user.ID),
		})
		return l.activateView(launcherWorldBuilding, &wb)

	case campaign.WorldReadyMsg:
		l.errMsg = ""
		l.openingScene = msg.Result.Scene
		l.logger().Info("world build complete", "campaign_id", dbutil.FromPgtype(msg.Result.Campaign.ID), "campaign", msg.Result.Campaign.Name)
		l.state = launcherLoadingCampaign
		l.view = nil
		return l, tea.Batch(l.spinner.Tick, l.runLoadCampaign(msg.Result.Campaign))

	case campaign.WorldErrorMsg:
		return l.reloadPicker(fmt.Sprintf("World building failed: %v", msg.Err))

	case campaign.BackMsg:
		return l.handleBack()

	case campaignLoadedMsg:
		if msg.err != nil {
			return l.reloadPicker(fmt.Sprintf("Load campaign failed: %v", msg.err))
		}
		return l.transitionToApp(msg.c, l.openingScene)
	}

	if l.view != nil {
		updated, cmd := l.view.Update(msg)
		if next := normalizeLauncherView(updated); next != nil {
			next.SetSize(l.width, l.height)
			l.view = next
		}
		return l, cmd
	}

	return l, nil
}

// View implements tea.Model.
func (l Launcher) View() string {
	switch l.state {
	case launcherLoading:
		if l.errMsg != "" {
			return styles.StatusError.Render("⚠  " + l.errMsg)
		}
		return l.loadingView("Connecting to database…")
	case launcherGeneratingProposals:
		return l.loadingView("Drafting campaign proposals…")
	case launcherGeneratingName:
		return l.loadingView("Naming your campaign…")
	case launcherLoadingCampaign:
		return l.loadingView("Loading campaign…")
	}

	if l.view == nil {
		return ""
	}
	if l.errMsg == "" {
		return l.view.View()
	}
	return styles.JoinVertical(
		styles.StatusError.Render("⚠  "+l.errMsg),
		l.view.View(),
	)
}

func (l Launcher) usesRootSpinner() bool {
	switch l.state {
	case launcherLoading, launcherGeneratingProposals, launcherGeneratingName, launcherLoadingCampaign:
		return true
	default:
		return false
	}
}

func (l Launcher) activateView(state launcherState, view View) (Launcher, tea.Cmd) {
	l.state = state
	l.view = view
	if view != nil {
		view.SetSize(l.width, l.height)
		return l, view.Init()
	}
	return l, nil
}

func (l Launcher) handleBack() (tea.Model, tea.Cmd) {
	l.errMsg = ""
	switch l.state {
	case launcherChooseMethod:
		return l.reloadPicker("")
	case launcherInterviewing, launcherAttributes:
		method := campaign.NewMethodModel()
		return l.activateView(launcherChooseMethod, &method)
	case launcherProposals:
		attrs := campaign.NewAttributesModelWithValues(l.proposalAttrs.Genre, l.proposalAttrs.SettingStyle, l.proposalAttrs.Tone)
		return l.activateView(launcherAttributes, &attrs)
	case launcherCharMethod:
		method := campaign.NewMethodModel()
		return l.activateView(launcherChooseMethod, &method)
	case launcherCharInterview, launcherCharForm, launcherConfirmation:
		charMethod := campaign.NewCharMethodModel()
		return l.activateView(launcherCharMethod, &charMethod)
	default:
		return l, nil
	}
}

func (l Launcher) reloadPicker(errMsg string) (tea.Model, tea.Cmd) {
	l.errMsg = errMsg
	l.state = launcherLoading
	l.view = nil
	l.openingScene = nil
	return l, tea.Batch(l.spinner.Tick, l.runBootstrap())
}

func (l *Launcher) resetCreationState() {
	l.campaignProfile = nil
	l.characterProfile = nil
	l.proposalAttrs = campaign.AttributesReadyMsg{}
	l.campaignName = ""
	l.campaignSummary = ""
	l.openingScene = nil
}

// loadingView renders a centred spinner with a message.
func (l Launcher) loadingView(msg string) string {
	line := l.spinner.View() + "  " + styles.Body.Render(msg)
	if l.width > 0 && l.height > 0 {
		return styles.Place(l.width, l.height, line)
	}
	return line
}

// transitionToApp creates the main App model and returns it as the new model.
func (l Launcher) transitionToApp(c statedb.Campaign, scene *world.SceneResult) (tea.Model, tea.Cmd) {
	app := NewAppWithEngine(l.cfg, c, l.ctx, l.engine, l.logBuf)
	app.seedOpeningScene(scene)
	l.openingScene = nil
	return app, app.Init()
}

// runBootstrap returns a tea.Cmd that runs the DB bootstrap asynchronously.
func (l Launcher) runBootstrap() tea.Cmd {
	ctx := l.ctx
	queries := l.queries
	return func() tea.Msg {
		result, err := bootstrap.Run(ctx, queries)
		return bootstrapDoneMsg{result: result, err: err}
	}
}

func (l Launcher) runGenerateProposals(attrs campaign.AttributesReadyMsg) tea.Cmd {
	ctx := l.ctx
	provider := l.provider
	return func() tea.Msg {
		generator := world.NewProposalGenerator(provider)
		proposals, err := generator.Generate(ctx, attrs.Genre, attrs.SettingStyle, attrs.Tone)
		return proposalsGeneratedMsg{proposals: proposals, err: err}
	}
}

func (l Launcher) runGenerateCampaignName(profile *world.CampaignProfile) tea.Cmd {
	ctx := l.ctx
	provider := l.provider
	return func() tea.Msg {
		name, err := world.GenerateCampaignName(ctx, provider, profile)
		return campaignNamedMsg{name: name, err: err}
	}
}

// runLoadCampaign returns a tea.Cmd that loads a campaign in the game engine
// before transitioning to the main app.
func (l Launcher) runLoadCampaign(c statedb.Campaign) tea.Cmd {
	ctx := l.ctx
	gameEngine := l.engine
	return func() tea.Msg {
		if gameEngine == nil {
			return campaignLoadedMsg{c: c, err: nil}
		}
		err := gameEngine.LoadCampaign(ctx, dbutil.FromPgtype(c.ID))
		return campaignLoadedMsg{c: c, err: err}
	}
}

func summarizeCampaignProfile(profile world.CampaignProfile) string {
	parts := make([]string, 0, 5)
	if profile.Tone != "" || profile.Genre != "" {
		parts = append(parts, strings.TrimSpace(profile.Tone+" "+profile.Genre))
	}
	if len(profile.Themes) > 0 {
		parts = append(parts, "themes of "+strings.Join(profile.Themes, ", "))
	}
	if profile.WorldType != "" {
		parts = append(parts, "set in a "+profile.WorldType+" world")
	}
	if profile.DangerLevel != "" {
		parts = append(parts, "with "+profile.DangerLevel+" danger")
	}
	if profile.PoliticalComplexity != "" {
		parts = append(parts, "and "+profile.PoliticalComplexity+" politics")
	}
	if len(parts) == 0 {
		return "A new adventure awaits."
	}
	return strings.TrimSpace(strings.Join(parts, " ")) + "."
}

func normalizeLauncherView(model tea.Model) View {
	switch v := model.(type) {
	case campaign.Picker:
		return &v
	case *campaign.Picker:
		return v
	case campaign.MethodModel:
		return &v
	case *campaign.MethodModel:
		return v
	case campaign.InterviewModel:
		return &v
	case *campaign.InterviewModel:
		return v
	case campaign.AttributesModel:
		return &v
	case *campaign.AttributesModel:
		return v
	case campaign.ProposalsModel:
		return &v
	case *campaign.ProposalsModel:
		return v
	case campaign.CharMethodModel:
		return &v
	case *campaign.CharMethodModel:
		return v
	case campaign.CharInterviewModel:
		return &v
	case *campaign.CharInterviewModel:
		return v
	case campaign.CharFormModel:
		return &v
	case *campaign.CharFormModel:
		return v
	case campaign.ConfirmationModel:
		return &v
	case *campaign.ConfirmationModel:
		return v
	case campaign.WorldBuildModel:
		return &v
	case *campaign.WorldBuildModel:
		return v
	default:
		return nil
	}
}

func creationMethodLabel(method campaign.CreationMethod) string {
	switch method {
	case campaign.MethodDescribe:
		return "describe"
	case campaign.MethodAttributes:
		return "attributes"
	default:
		return "unknown"
	}
}
