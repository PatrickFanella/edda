package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"git.subcult.tv/subculture-collective/edda/internal/assembly"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/memory"
)

type memoryDimensionDB interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// SemanticMemoryOptions builds server startup options for retrieval, recording
// by world-building tools, and search_memory, independently of the GM provider.
// Configuration and schema errors stop startup. Provider availability is checked
// on use; existing retrieval and recording paths report failures without undoing
// a turn or a durable world-state write.
func SemanticMemoryOptions(ctx context.Context, cfg config.LLMConfig, db memoryDimensionDB, store memory.MemorySearchStore, logger *slog.Logger) ([]Option, error) {
	if logger == nil {
		logger = slog.Default()
	}
	embedding, err := cfg.ResolveMemoryEmbedding()
	if err != nil {
		return nil, err
	}
	if !embedding.Enabled {
		if cfg.MemoryEmbedding != nil {
			logger.Warn("semantic campaign memory disabled: llm.memoryembedding.enabled is false")
		} else {
			logger.Warn("semantic campaign memory unavailable: no embedding provider configured; set llm.memoryembedding.enabled and endpoint/model/dimension")
		}
		return nil, nil
	}
	if err := memory.ValidateMemoryEmbeddingDimension(ctx, db, embedding.Dimension); err != nil {
		return nil, fmt.Errorf("semantic campaign memory unavailable: %w", err)
	}
	embedder := memory.NewOllamaEmbedder(embedding.Endpoint, embedding.Model,
		memory.WithOllamaEmbedderDimension(embedding.Dimension),
		memory.WithOllamaEmbedderTimeout(cfg.Ollama.RequestTimeout()),
		memory.WithOllamaEmbedderAPIKey(embedding.APIKey),
	)
	searcher := memory.NewSearcher(embedder, store)
	tier3 := assembly.NewTier3Retriever(searcher, 5, logger.WithGroup("tier3"))
	logger.Info("semantic campaign memory configured; embedding availability checked on use", "provider", embedding.Provider, "dimension", embedding.Dimension)
	return []Option{WithTier3Retriever(tier3), WithEmbedder(embedder), WithSearcher(searcher)}, nil
}
