package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
)

const reviseFactToolName = "revise_fact"

var (
	ErrWorldFactNotFound      = domain.ErrWorldFactNotFound
	ErrWorldFactSuperseded    = domain.ErrWorldFactSuperseded
	ErrWorldFactWrongCampaign = errors.New("world fact does not belong to campaign")
)

type ReviseWorldFactCommand = domain.ReviseWorldFactCommand

type ReviseWorldFactResult = domain.ReviseWorldFactResult

// ReviseFactStore persists fact supersession.
type ReviseFactStore interface {
	ReviseWorldFact(ctx context.Context, cmd ReviseWorldFactCommand) (*ReviseWorldFactResult, error)
}

// ReviseFactTool returns the revise_fact tool definition and JSON schema.
func ReviseFactTool() llm.Tool {
	return llm.Tool{
		Name:        reviseFactToolName,
		Description: "Revise a canonical world fact by creating a new fact that supersedes the old one. The old fact is marked as superseded and excluded from active queries.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"campaign_id": map[string]any{
					"type":        "string",
					"description": "Campaign UUID that owns the fact. Usually supplied by tool context; explicit values must match context.",
				},
				"fact_id": map[string]any{
					"type":        "string",
					"description": "UUID of the existing world fact to supersede.",
				},
				"new_fact": map[string]any{
					"type":        "string",
					"description": "The revised fact text that will replace the old one.",
				},
				"reveal_to_player": map[string]any{
					"type":        "boolean",
					"description": "If true, the player character becomes aware of this fact. Defaults to false.",
				},
			},
			"required":             []string{"fact_id", "new_fact"},
			"additionalProperties": false,
		},
	}
}

// RegisterReviseFact registers the revise_fact tool in the registry.
func RegisterReviseFact(reg *Registry, factStore ReviseFactStore, memoryStore MemoryStore, embedder Embedder) error {
	if factStore == nil {
		return errors.New("revise_fact fact store is required")
	}
	tool := ReviseFactTool()
	handler := NewReviseFactHandler(factStore, memoryStore, embedder)
	return reg.Register(tool, func(ctx context.Context, args map[string]any) (*ToolResult, error) {
		return handler.Handle(ctx, args)
	})
}

// ReviseFactHandler handles the revise_fact tool call.
type ReviseFactHandler struct {
	factStore   ReviseFactStore
	memoryStore MemoryStore
	embedder    Embedder
}

// NewReviseFactHandler creates a new ReviseFactHandler.
func NewReviseFactHandler(factStore ReviseFactStore, memoryStore MemoryStore, embedder Embedder) *ReviseFactHandler {
	return &ReviseFactHandler{
		factStore:   factStore,
		memoryStore: memoryStore,
		embedder:    embedder,
	}
}

// Handle executes the revise_fact tool.
func (h *ReviseFactHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("revise_fact handler is nil")
	}
	if h.factStore == nil {
		return nil, errors.New("revise_fact fact store is required")
	}

	factID, err := parseUUIDArg(args, "fact_id")
	if err != nil {
		return nil, err
	}
	campaignID, err := campaignIDForReviseFact(ctx, args)
	if err != nil {
		return nil, err
	}
	newFactText, err := parseStringArg(args, "new_fact")
	if err != nil {
		return nil, err
	}

	revealToPlayer, err := parseBoolArg(args, "reveal_to_player")
	if err != nil {
		return nil, err
	}

	result, err := h.factStore.ReviseWorldFact(ctx, ReviseWorldFactCommand{CampaignID: campaignID, FactID: factID, NewFact: newFactText, RevealToPlayer: revealToPlayer})
	if err != nil {
		if errors.Is(err, ErrWorldFactNotFound) {
			return nil, errors.New("fact_id does not reference an existing world fact")
		}
		if errors.Is(err, ErrWorldFactSuperseded) {
			return nil, fmt.Errorf("fact %s is already superseded and cannot be revised", factID)
		}
		return nil, fmt.Errorf("revise world fact: %w", err)
	}

	newFactID := result.NewFact.ID
	resultCampaignID := result.NewFact.CampaignID

	if h.embedder != nil && h.memoryStore != nil {
		if err := h.embedRevisedFactMemory(ctx, resultCampaignID, factID, newFactID, newFactText, result.NewFact.Category); err != nil {
			return &ToolResult{Success: true, Data: map[string]any{"new_fact_id": newFactID.String(), "old_fact_id": factID.String(), "campaign_id": resultCampaignID.String(), "new_fact": newFactText, "category": result.NewFact.Category, "source": reviseFactToolName, "supersedes": factID.String(), "player_known": result.PlayerKnownPropagated, "memory_warning": err.Error()}, Narrative: fmt.Sprintf("World fact revised: %q supersedes fact %s. Memory embedding failed: %v", newFactText, factID, err)}, nil
		}
	}

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"new_fact_id":  newFactID.String(),
			"old_fact_id":  factID.String(),
			"campaign_id":  resultCampaignID.String(),
			"new_fact":     newFactText,
			"category":     result.NewFact.Category,
			"source":       reviseFactToolName,
			"supersedes":   factID.String(),
			"player_known": result.PlayerKnownPropagated,
		},
		Narrative: fmt.Sprintf("World fact revised: %q supersedes fact %s.", newFactText, factID),
	}, nil
}

func campaignIDForReviseFact(ctx context.Context, args map[string]any) (uuid.UUID, error) {
	contextCampaignID, hasContextCampaignID := CurrentCampaignIDFromContext(ctx)
	raw, hasArg := args["campaign_id"]
	if !hasArg {
		if hasContextCampaignID {
			return contextCampaignID, nil
		}
		return uuid.Nil, errors.New("campaign_id is required")
	}
	argCampaignID, err := parseUUIDArg(map[string]any{"campaign_id": raw}, "campaign_id")
	if err != nil {
		return uuid.Nil, err
	}
	if hasContextCampaignID && argCampaignID != contextCampaignID {
		return uuid.Nil, errors.New("campaign_id does not match current campaign context")
	}
	return argCampaignID, nil
}

func (h *ReviseFactHandler) embedRevisedFactMemory(
	ctx context.Context,
	campaignID uuid.UUID,
	oldFactID uuid.UUID,
	newFactID uuid.UUID,
	newFactText string,
	category string,
) error {
	memoryContent := fmt.Sprintf("World fact revised. Category: %s. New fact: %s (supersedes fact %s)", category, newFactText, oldFactID)
	embedding, err := h.embedder.Embed(ctx, memoryContent)
	if err != nil {
		return fmt.Errorf("embed revised fact memory: %w", err)
	}
	metadata, err := json.Marshal(map[string]any{
		"fact_id":     newFactID.String(),
		"old_fact_id": oldFactID.String(),
		"category":    category,
		"source":      reviseFactToolName,
		"supersedes":  oldFactID.String(),
	})
	if err != nil {
		return fmt.Errorf("marshal revised fact memory metadata: %w", err)
	}
	if err := h.memoryStore.CreateMemory(ctx, CreateMemoryParams{
		CampaignID: campaignID,
		Content:    memoryContent,
		Embedding:  embedding,
		MemoryType: string(domain.MemoryTypeWorldFact),
		Metadata:   metadata,
	}); err != nil {
		return fmt.Errorf("create revised fact memory: %w", err)
	}
	return nil
}
