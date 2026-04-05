package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/memory"
)

// SearchMemorySearcher is the narrow interface required by the search_memory
// tool. It is satisfied by *memory.Searcher.
type SearchMemorySearcher interface {
	SearchSimilar(ctx context.Context, campaignID uuid.UUID, query string, limit int) ([]memory.MemoryResult, error)
}

// SearchMemoryTool returns the search_memory tool definition.
func SearchMemoryTool() llm.Tool {
	return llm.Tool{
		Name:        "search_memory",
		Description: "Search campaign memories for information relevant to a query. Returns the most semantically similar stored memories about NPCs, locations, factions, lore, and world facts.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "What to search for (e.g., 'tavern keeper in the northern district', 'dragon attack on the village', 'ancient prophecy').",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of memories to return (1-10). Defaults to 5.",
				},
			},
			"required":             []string{"query"},
			"additionalProperties": false,
		},
	}
}

// RegisterSearchMemory registers the search_memory tool with the given registry.
func RegisterSearchMemory(reg *Registry, searcher SearchMemorySearcher) error {
	if searcher == nil {
		return errors.New("search_memory: searcher is required")
	}
	handler := &SearchMemoryHandler{searcher: searcher}
	return reg.Register(SearchMemoryTool(), handler.Handle)
}

// SearchMemoryHandler executes search_memory tool calls.
type SearchMemoryHandler struct {
	searcher SearchMemorySearcher
}

const (
	defaultSearchMemoryLimit = 5
	maxSearchMemoryLimit     = 10
)

// Handle processes a search_memory tool call.
func (h *SearchMemoryHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	query, err := parseStringArg(args, "query")
	if err != nil {
		return nil, err
	}

	limit := defaultSearchMemoryLimit
	if _, ok := args["limit"]; ok {
		parsed, parseErr := parseIntArg(args, "limit")
		if parseErr != nil {
			return nil, parseErr
		}
		if parsed >= 1 && parsed <= maxSearchMemoryLimit {
			limit = parsed
		}
	}

	campaignID, ok := CurrentCampaignIDFromContext(ctx)
	if !ok {
		return nil, errors.New("search_memory requires campaign_id in context")
	}

	results, err := h.searcher.SearchSimilar(ctx, campaignID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search memories: %w", err)
	}

	if len(results) == 0 {
		return &ToolResult{
			Success:   true,
			Data:      map[string]any{"count": 0, "memories": []any{}},
			Narrative: "No relevant memories found.",
		}, nil
	}

	memories := make([]any, len(results))
	var parts []string
	for i, r := range results {
		memories[i] = map[string]any{
			"content":     r.Content,
			"memory_type": r.MemoryType,
			"relevance":   1 - r.Distance,
		}
		parts = append(parts, fmt.Sprintf("- [%s] %s", r.MemoryType, r.Content))
	}

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"count":    len(results),
			"memories": memories,
		},
		Narrative: fmt.Sprintf("Found %d relevant memories:\n%s", len(results), strings.Join(parts, "\n")),
	}, nil
}
