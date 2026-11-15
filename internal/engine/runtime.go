package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"git.subcult.tv/subculture-collective/edda/internal/assembly"
	"git.subcult.tv/subculture-collective/edda/internal/bootstrap"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// Engine is the concrete GameEngine implementation used by the TUI.
type Engine struct {
	actionDice         srd521.Dice
	turnDB             transactionBeginner
	baseDB             statedb.DBTX
	turnLock           turnLock
	logger             *slog.Logger
	state              game.StateManager
	queries            statedb.Querier
	assembler          *assembly.ContextAssembler
	processor          *TurnProcessor
	provider           llm.Provider
	postTurnProvider   llm.Provider
	choiceProvider     llm.Provider
	contextTokenBudget int
	tier3              *assembly.Tier3Retriever
	toolFilter         ToolFilter
	embedder           tools.Embedder
	searcher           tools.SearchMemorySearcher
	saveStore          *saves.Store
	summarizer         *journal.Summarizer
	summaryContext     context.Context
}

const recentTurnLimit = 10

// Option configures the Engine during construction.
type Option func(*Engine)

// WithTier3Retriever attaches a semantic memory retriever to the engine.
// When set, ProcessTurn includes relevant memories in the LLM context.
func WithTier3Retriever(t *assembly.Tier3Retriever) Option {
	return func(e *Engine) {
		e.tier3 = t
	}
}

// WithEmbedder attaches a vector embedder to the engine. When set,
// world-building tools automatically embed created entities as memories.
func WithEmbedder(emb tools.Embedder) Option {
	return func(e *Engine) { e.embedder = emb }
}

// WithSearcher attaches a memory searcher for the search_memory tool.
func WithSearcher(s tools.SearchMemorySearcher) Option {
	return func(e *Engine) { e.searcher = s }
}

// WithSaveStore attaches a saves.Store for auto-save after each turn.
func WithSaveStore(s *saves.Store) Option {
	return func(e *Engine) { e.saveStore = s }
}

// WithSummarizer attaches a journal summarizer for auto-summarization after turns.
func WithSummarizer(s *journal.Summarizer) Option {
	return func(e *Engine) { e.summarizer = s }
}

// WithSummaryContext ties automatic summaries to the process lifetime rather
// than to an individual HTTP request. Cancellation stops in-flight summaries.
func WithSummaryContext(ctx context.Context) Option {
	return func(e *Engine) { e.summaryContext = ctx }
}

// WithLogger sets the structured logger for the engine and its subsystems.
func WithLogger(l *slog.Logger) Option {
	return func(e *Engine) { e.logger = l }
}

// EngineLLMProviders contains optional routed providers for specific engine
// flows. Nil fields fall back to the main turn provider.
type EngineLLMProviders struct {
	PostTurnState      llm.Provider
	ChoiceFallback     llm.Provider
	ContextTokenBudget int
}

// WithLLMProviders attaches routed providers for low-risk subflows while the
// main GM turn remains on the provider passed to New.
func WithLLMProviders(providers EngineLLMProviders) Option {
	return func(e *Engine) {
		e.postTurnProvider = providers.PostTurnState
		e.choiceProvider = providers.ChoiceFallback
		e.contextTokenBudget = providers.ContextTokenBudget
	}
}

// New creates a concrete GameEngine backed by the shared game and llm packages.
func New(db statedb.DBTX, provider llm.Provider, llmCfg config.LLMConfig, opts ...Option) (*Engine, error) {
	queries := statedb.New(db)

	e := &Engine{
		state:   game.NewStateManager(db),
		queries: queries,
	}
	if db != nil {
		beginner, ok := db.(transactionBeginner)
		if !ok {
			return nil, fmt.Errorf("database must support transactions for campaign turn locking")
		}
		e.turnDB = beginner
		e.baseDB = db
		switch configured := db.(type) {
		case interface{ Config() *pgxpool.Config }:
			e.turnLock = dedicatedTurnLock(configured.Config().ConnConfig.Copy())
		case interface{ Config() *pgx.ConnConfig }:
			e.turnLock = dedicatedTurnLock(configured.Config().Copy())
		default:
			return nil, fmt.Errorf("database must expose pgx connection configuration for inference-safe turn locking")
		}
	}
	for _, opt := range opts {
		opt(e)
	}
	if e.logger == nil {
		e.logger = slog.Default()
	}
	registry := tools.NewRegistry()
	// Pass db as TimeStore for the advance_time tool (DBTX satisfies TimeStore).
	toolQueries := queries
	var toolTimeDB statedb.DBTX = db
	if db != nil {
		toolTimeDB = toolDB{base: db}
		toolQueries = statedb.New(toolTimeDB)
		registry.AroundExecute(e.journalTool)
	}
	if err := registerAllTools(registry, toolQueries, e.embedder, e.searcher, toolTimeDB); err != nil {
		return nil, fmt.Errorf("register tools: %w", err)
	}

	if e.toolFilter == nil {
		e.toolFilter = NewPhaseToolFilter(registry)
	}

	contextTokenBudget := llmCfg.ContextTokenBudget()
	if e.contextTokenBudget > 0 {
		contextTokenBudget = e.contextTokenBudget
	}
	e.assembler = assembly.NewContextAssembler(registry, assembly.WithTokenBudget(contextTokenBudget))
	e.processor = NewTurnProcessor(provider, registry, tools.NewValidator(registry), e.logger.WithGroup("turns"))
	e.processor.SetPostTurnProvider(e.postTurnProvider)
	e.provider = provider
	if e.choiceProvider == nil {
		e.choiceProvider = provider
	}
	return e, nil
}

var _ GameEngine = (*Engine)(nil)

func (e *Engine) ProcessTurn(ctx context.Context, campaignID uuid.UUID, playerInput string) (*TurnResult, error) {
	release, err := e.acquireTurn(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	defer release()

	if e.logger == nil {
		e.logger = slog.Default()
	}

	tc := &TurnContext{
		CampaignID:  campaignID,
		PlayerInput: playerInput,
		Logger:      e.logger,
		Started:     time.Now(),
	}

	tc.Logger.Info("process turn started", "campaign_id", campaignID, "input_len", len(playerInput))

	pipeline := NewPipeline(
		e.gatherStage(),
		e.memoryStage(),
		e.assembleStage(),
		e.processStage(),
		e.completeStage(),
	)

	if err := pipeline.Execute(ctx, tc); err != nil && !errors.Is(err, errRecoveredTurn) {
		return nil, e.turnFailure(tc.Ctx, err)
	}

	result := &TurnResult{ActionID: tc.ActionID, Recovered: tc.Recovered, TurnNumber: tc.TurnNumber,
		Narrative:        tc.Narrative,
		AppliedToolCalls: tc.Applied,
		StateChanges:     tc.StateChanges,
		Choices:          tc.Choices,
		CombatActive:     tc.CombatActive,
	}

	tc.Logger.Info("process turn completed",
		"campaign_id", campaignID,
		"duration_ms", time.Since(tc.Started).Milliseconds(),
		"narrative_len", len(result.Narrative),
		"choices", len(result.Choices),
		"tool_calls", len(result.AppliedToolCalls),
	)
	return result, nil
}

// autoSaveIfNeeded creates an auto-save point after each turn and cleans up old ones.
func (e *Engine) autoSaveIfNeeded(ctx context.Context, campaignID uuid.UUID, turnNumber int) {
	if e.saveStore == nil {
		return
	}
	name := fmt.Sprintf("Auto-save (turn %d)", turnNumber)
	if _, err := e.saveStore.CreateSavePoint(ctx, campaignID, name, turnNumber, true); err != nil {
		e.logger.Warn("auto-save: failed to create save point", "campaign_id", campaignID, "error", err)
		return
	}
	if err := e.saveStore.DeleteOldAutoSaves(ctx, campaignID); err != nil {
		e.logger.Warn("auto-save: failed to clean up old auto-saves", "campaign_id", campaignID, "error", err)
	}
}

func (e *Engine) GetGameState(ctx context.Context, campaignID uuid.UUID) (*GameState, error) {
	state, err := e.state.GatherState(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("gather state: %w", err)
	}
	return GameStateFromFull(state), nil
}

func (e *Engine) NewCampaign(ctx context.Context, userID uuid.UUID) (*domain.Campaign, error) {
	return e.state.CreateCampaign(ctx, game.CreateCampaignParams{
		Name:   bootstrap.DefaultCampaignName,
		UserID: userID,
	})
}

func (e *Engine) LoadCampaign(ctx context.Context, campaignID uuid.UUID) error {
	if e.logger == nil {
		e.logger = slog.Default()
	}
	e.logger.Info("load campaign started", "campaign_id", campaignID)
	_, err := e.state.GetCampaignByID(ctx, campaignID)
	if err != nil {
		e.logger.Error("load campaign failed", "campaign_id", campaignID, "error", err)
		return fmt.Errorf("load campaign: %w", err)
	}
	e.logger.Info("load campaign completed", "campaign_id", campaignID)
	return nil
}

// ProcessTurnStream is like ProcessTurn but delivers narrative chunks over
// the returned channel. The channel is closed when processing completes.
// Callers must consume the channel or cancel ctx when abandoning the stream.
//
// Provider fragments are provisional. A narrative_reset status invalidates
// prior fragments; only the persisted TurnResult is committed history.
func (e *Engine) ProcessTurnStream(ctx context.Context, campaignID uuid.UUID, playerInput string) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 16)
	go func() {
		var tc *TurnContext
		defer close(ch)
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("process turn panic: %v", r)
				if tc != nil {
					err = e.turnFailure(tc.Ctx, err)
				}
				sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: err})
			}
		}()

		release, err := e.acquireTurn(ctx, campaignID)
		if err != nil {
			sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: err})
			return
		}
		defer release()

		tc = &TurnContext{CampaignID: campaignID, PlayerInput: playerInput, Logger: e.logger, Started: time.Now()}
		tc.NarrativeCallback = func(text string) {
			if text == "" {
				sendStreamEvent(ctx, ch, StreamEvent{Type: "status", Status: &api.StatusPayload{Stage: "narrative_reset", Description: "Revising provisional narration..."}})
			} else {
				sendStreamEvent(ctx, ch, StreamEvent{Type: "chunk", Text: text})
			}
		}
		emitStatus := func(stage, description string) {
			sendStreamEvent(ctx, ch, StreamEvent{Type: "status", Status: &api.StatusPayload{Stage: stage, Description: description}})
		}
		emitPhase := func(stage, description string, started time.Time) {
			e.emitTimedStatus(ctx, ch, stage, description, started)
		}

		if ctx.Err() != nil {
			return
		}
		gatherStarted := time.Now()
		emitStatus("gathering", "Gathering world state...")
		if err := e.gatherStage()(ctx, tc); err != nil {
			if errors.Is(err, errRecoveredTurn) {
				result := &TurnResult{ActionID: tc.ActionID, Recovered: tc.Recovered, TurnNumber: tc.TurnNumber, Narrative: tc.Narrative, AppliedToolCalls: tc.Applied, StateChanges: tc.StateChanges, Choices: tc.Choices, CombatActive: tc.CombatActive}
				sendStreamEvent(ctx, ch, StreamEvent{Type: "result", Result: result})
				return
			}
			sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: e.turnFailure(tc.Ctx, err)})
			return
		}
		if tc.Ctx == nil {
			tc.Ctx = ctx
		}
		emitPhase("gathering", "Gathering world state complete.", gatherStarted)

		if ctx.Err() != nil {
			return
		}
		memoryStarted := time.Now()
		emitStatus("memory_retrieval", "Retrieving relevant memories...")
		if err := e.memoryStage()(tc.Ctx, tc); err != nil {
			sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: e.turnFailure(tc.Ctx, err)})
			return
		}
		emitPhase("memory_retrieval", "Relevant memories retrieved.", memoryStarted)

		if ctx.Err() != nil {
			return
		}
		assembleStarted := time.Now()
		emitStatus("context_assembly", "Assembling turn context...")
		if err := e.assembleStage()(tc.Ctx, tc); err != nil {
			sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: e.turnFailure(tc.Ctx, err)})
			return
		}
		emitPhase("context_assembly", "Turn context assembled.", assembleStarted)

		if ctx.Err() != nil {
			return
		}
		thinkingStarted := time.Now()
		emitStatus("thinking", "Generating response...")
		localProcessor := *e.processor
		localProcessor.StatusCallback = func(s api.StatusPayload) { sendStreamEvent(ctx, ch, StreamEvent{Type: "status", Status: &s}) }
		narrative, applied, err := localProcessor.ProcessWithRecoveryWithOptions(tc.Ctx, tc.Messages, tc.FilteredTools, TurnProcessorOptions{CandidateCallback: func(text string) error { return e.recordRawCandidate(tc.Ctx, text) }, NarrativeCallback: tc.NarrativeCallback})
		if err != nil {
			sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: e.turnFailure(tc.Ctx, err)})
			return
		}
		emitPhase("thinking", "Response generated.", thinkingStarted)

		tc.Narrative = narrative
		tc.Applied = applied
		if err := e.recordCandidate(tc.Ctx, tc); err != nil {
			sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: e.turnFailure(tc.Ctx, err)})
			return
		}

		if ctx.Err() != nil {
			return
		}
		persistStarted := time.Now()
		emitStatus("finalizing", "Finalizing and persisting turn...")
		if err := e.completeStage()(tc.Ctx, tc); err != nil {
			sendStreamEvent(ctx, ch, StreamEvent{Type: "error", Err: e.turnFailure(tc.Ctx, err)})
			return
		}
		emitPhase("finalizing", "Turn persisted.", persistStarted)

		// Emit combat lifecycle status events based on applied tool calls.
		result := &TurnResult{ActionID: tc.ActionID, Recovered: tc.Recovered, TurnNumber: tc.TurnNumber, Narrative: tc.Narrative, AppliedToolCalls: tc.Applied, StateChanges: tc.StateChanges, Choices: tc.Choices, CombatActive: tc.CombatActive}
		for _, atc := range result.AppliedToolCalls {
			switch atc.Tool {
			case "initiate_combat":
				sendStreamEvent(ctx, ch, StreamEvent{Type: "status", Status: &api.StatusPayload{Stage: "combat_start", Description: "Combat has begun!"}})
			case "resolve_combat":
				sendStreamEvent(ctx, ch, StreamEvent{Type: "status", Status: &api.StatusPayload{Stage: "combat_end", Description: "Combat has ended."}})
			}
		}

		sendStreamEvent(ctx, ch, StreamEvent{Type: "result", Result: result})
	}()
	return ch, nil
}

func (e *Engine) emitTimedStatus(ctx context.Context, ch chan<- StreamEvent, stage, description string, started time.Time) {
	sendStreamEvent(ctx, ch, StreamEvent{Type: "status", Status: &api.StatusPayload{Stage: stage, Description: description}})
	if e != nil && e.logger != nil {
		e.logger.Info("turn phase completed", "stage", stage, "duration_ms", time.Since(started).Milliseconds())
	}
}

// sendStreamEvent releases a blocked publisher when its consumer abandons the
// stream. A canceled stream owes no terminal frame to a disconnected consumer.
func sendStreamEvent(ctx context.Context, ch chan<- StreamEvent, event StreamEvent) {
	if ctx.Err() != nil {
		return
	}
	select {
	case ch <- event:
	case <-ctx.Done():
	}
}

// autoSummarizeIfNeeded triggers async summarization every 10 turns.
func (e *Engine) autoSummarizeIfNeeded(ctx context.Context, campaignID uuid.UUID, turnNumber int) {
	if e.summarizer == nil {
		return
	}
	if turnNumber <= 0 || turnNumber%10 != 0 {
		return
	}

	fromTurn := turnNumber - 9
	if fromTurn < 1 {
		fromTurn = 1
	}
	toTurn := turnNumber

	go func() {
		// A completed request must not cancel its summary, but process shutdown
		// and a fixed deadline must bound the provider and persistence calls.
		parent := e.summaryContext
		if parent == nil {
			parent = context.Background()
		}
		bgCtx, cancel := context.WithTimeout(parent, 2*time.Minute)
		defer cancel()
		if _, err := e.summarizer.Summarize(bgCtx, campaignID, fromTurn, toTurn); err != nil {
			e.logger.Warn("auto-summarize: failed", "campaign_id", campaignID, "from_turn", fromTurn, "to_turn", toTurn, "error", err)
		} else {
			e.logger.Info("auto-summarize: completed", "campaign_id", campaignID, "from_turn", fromTurn, "to_turn", toTurn)
		}
	}()
}

func nextTurnNumber(logs []domain.SessionLog) int {
	if len(logs) == 0 {
		return 1
	}
	return logs[len(logs)-1].TurnNumber + 1
}

func marshalAppliedToolCalls(applied []AppliedToolCall) (json.RawMessage, error) {
	if applied == nil {
		applied = []AppliedToolCall{}
	}
	data, err := json.Marshal(applied)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// questToolNames are tools that modify quest state, triggering a history snapshot.
var questToolNames = map[string]struct{}{
	"create_quest":       {},
	"create_subquest":    {},
	"update_quest":       {},
	"complete_objective": {},
	"branch_quest":       {},
	"link_quest_entity":  {},
}

// snapshotQuestsIfNeeded creates quest history entries when quest tools were invoked.
func (e *Engine) snapshotQuestsIfNeeded(ctx context.Context, campaignID uuid.UUID, applied []AppliedToolCall) {
	if e.queries == nil {
		return
	}
	hasQuestTool := false
	for _, atc := range applied {
		if _, ok := questToolNames[atc.Tool]; ok {
			hasQuestTool = true
			break
		}
	}
	if !hasQuestTool {
		return
	}

	pgCampaignID := dbutil.ToPgtype(campaignID)
	quests, err := e.queries.ListActiveQuests(ctx, pgCampaignID)
	if err != nil {
		e.logger.Warn("quest snapshot: failed to list quests", "error", err)
		return
	}
	for _, q := range quests {
		snapshot := fmt.Sprintf("Status: %s | Title: %s", q.Status, q.Title)
		if _, err := e.queries.CreateQuestHistoryEntry(ctx, statedb.CreateQuestHistoryEntryParams{
			QuestID:  q.ID,
			Snapshot: snapshot,
		}); err != nil {
			e.logger.Warn("quest snapshot: failed to create history entry", "quest_id", q.ID, "error", err)
		}
	}
}
