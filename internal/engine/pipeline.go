package engine

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/domain"
	"github.com/PatrickFanella/game-master/internal/game"
	"github.com/PatrickFanella/game-master/internal/llm"
)

// TurnContext carries all data accumulated during turn processing.
// Stages may enrich Ctx (e.g. attaching campaign/player IDs); later stages
// automatically see the updated value because Pipeline reads tc.Ctx before
// each invocation.
type TurnContext struct {
	Ctx           context.Context
	CampaignID    uuid.UUID
	PlayerInput   string
	State         *game.GameState
	RecentLogs    []domain.SessionLog
	Memories      []string
	Messages      []llm.Message
	AllTools      []llm.Tool
	FilteredTools []llm.Tool
	Narrative     string
	Applied       []AppliedToolCall
	Choices       []Choice
	CombatActive  bool
	TurnNumber    int
	ToolCallsJSON json.RawMessage
	Logger        *slog.Logger
	Started       time.Time
}

// Stage is a single step in the turn pipeline.
type Stage func(ctx context.Context, tc *TurnContext) error

// Pipeline executes stages in order, short-circuiting on error.
type Pipeline struct {
	stages []Stage
}

// NewPipeline creates a pipeline from the given stages.
func NewPipeline(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Execute runs each stage sequentially, returning the first error encountered.
// Each stage receives tc.Ctx so that context enrichments from earlier stages
// (e.g. campaign/player IDs) are visible to later ones.
func (p *Pipeline) Execute(ctx context.Context, tc *TurnContext) error {
	if tc.Ctx == nil {
		tc.Ctx = ctx
	}
	for _, stage := range p.stages {
		if err := stage(tc.Ctx, tc); err != nil {
			return err
		}
	}
	return nil
}
