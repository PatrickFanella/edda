package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	pgvector_go "github.com/pgvector/pgvector-go"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/memory"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// AutoEmbedStore is the narrow persistence interface for auto-embedding.
type AutoEmbedStore interface {
	CreateMemory(ctx context.Context, arg statedb.CreateMemoryParams) (statedb.Memory, error)
}

// AutoEmbedder wraps tool handlers to automatically embed created entities as
// memories. Embedding failures are logged but never propagated to callers.
type AutoEmbedder struct {
	embedder memory.Embedder
	store    AutoEmbedStore
	logger   *log.Logger
}

// NewAutoEmbedder constructs an AutoEmbedder. If embedder is nil, all Embed
// calls become no-ops. If logger is nil, the default logger is used.
func NewAutoEmbedder(embedder memory.Embedder, store AutoEmbedStore, logger *log.Logger) *AutoEmbedder {
	if logger == nil {
		logger = log.Default()
	}
	return &AutoEmbedder{
		embedder: embedder,
		store:    store,
		logger:   logger,
	}
}

// toolMemoryType maps tool names to memory_type values. Tools not present
// in this map are silently skipped during auto-embedding.
var toolMemoryType = map[string]string{
	"create_npc":             "npc",
	"create_location":        "location",
	"create_faction":         "faction",
	"create_lore":            "lore",
	"establish_fact":         "world_fact",
	"create_language":        "language",
	"create_belief_system":   "belief_system",
	"create_culture":         "culture",
	"create_city":            "city",
	"create_economic_system": "economic_system",
}

// Wrap returns a new handler function that delegates to handler and, on
// success, attempts to embed the result as a memory. The original result is
// always returned regardless of embedding outcome.
func (ae *AutoEmbedder) Wrap(toolName string, handler func(ctx context.Context, args map[string]any) (*ToolResult, error)) func(ctx context.Context, args map[string]any) (*ToolResult, error) {
	return func(ctx context.Context, args map[string]any) (result *ToolResult, err error) {
		defer func() {
			if r := recover(); r != nil {
				ae.logger.Error("auto-embed panic recovered", "tool", toolName, "panic", fmt.Sprint(r))
			}
		}()

		result, err = handler(ctx, args)
		if err != nil || result == nil || !result.Success {
			return result, err
		}

		ae.embedResult(ctx, toolName, result)
		return result, err
	}
}

// embedResult extracts a summary from the tool result and persists it as a
// vector memory. All errors are logged, never returned.
func (ae *AutoEmbedder) embedResult(ctx context.Context, toolName string, result *ToolResult) {
	if ae.embedder == nil {
		return
	}

	if result.Data == nil {
		return
	}

	memType, ok := toolMemoryType[toolName]
	if !ok {
		return
	}

	summary := composeSummary(result.Data)
	if summary == "" {
		return
	}

	vec, err := ae.embedder.Embed(ctx, summary)
	if err != nil {
		ae.logger.Error("auto-embed: embedding failed", "tool", toolName, "error", err)
		return
	}

	// Extract campaign_id from result data when available.
	campaignIDStr, _ := result.Data["campaign_id"].(string)
	campaignUUID, parseErr := uuid.Parse(campaignIDStr)
	if parseErr != nil {
		ae.logger.Error("auto-embed: invalid campaign_id in result", "tool", toolName, "campaign_id", campaignIDStr)
		return
	}

	params := statedb.CreateMemoryParams{
		CampaignID: dbutil.ToPgtype(campaignUUID),
		Content:    summary,
		Embedding:  pgvector_go.NewVector(vec),
		MemoryType: memType,
	}

	if _, err := ae.store.CreateMemory(ctx, params); err != nil {
		ae.logger.Error("auto-embed: store failed", "tool", toolName, "error", err)
	}
}

// composeSummary builds a human-readable summary from common entity data
// fields. Returns empty string if no useful content is found.
func composeSummary(data map[string]any) string {
	keys := []string{"name", "title", "description", "content", "fact"}
	var parts []string
	for _, k := range keys {
		if v, ok := data[k].(string); ok && v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, ". ")
}
