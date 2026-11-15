package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

// newAPIEngine keeps the API startup wiring shared with its integration checks.
func newAPIEngine(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, roster *llm.Roster, saveStore *saves.Store) (*engine.Engine, error) {
	opts := []engine.Option{
		engine.WithLogger(slog.Default().WithGroup("engine")),
		engine.WithSaveStore(saveStore),
		engine.WithSummarizer(journal.NewSummarizer(roster.Provider(llm.FlowSummary), journal.NewStore(pool))),
		engine.WithSummaryContext(ctx),
		engine.WithLLMProviders(engine.EngineLLMProviders{
			PostTurnState:      roster.Provider(llm.FlowPostTurnState),
			ChoiceFallback:     roster.Provider(llm.FlowChoiceFallback),
			ContextTokenBudget: roster.Info(llm.FlowGMTurn).ContextTokenBudget,
		}),
	}
	memoryOpts, err := engine.SemanticMemoryOptions(ctx, cfg.LLM, pool, statedb.New(pool), slog.Default().WithGroup("memory"))
	if err != nil {
		return nil, fmt.Errorf("configure semantic campaign memory: %w", err)
	}
	opts = append(opts, memoryOpts...)
	return engine.New(pool, roster.Provider(llm.FlowGMTurn), cfg.LLM, opts...)
}
