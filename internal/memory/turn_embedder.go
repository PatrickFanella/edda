package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	pgvector_go "github.com/pgvector/pgvector-go"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// MemoryCreateStore is the narrow persistence interface required by
// TurnEmbedder. It is satisfied by statedb.Queries.
type MemoryCreateStore interface {
	CreateMemory(ctx context.Context, arg statedb.CreateMemoryParams) (statedb.Memory, error)
}

// TurnSummaryInput holds the data needed to embed a single game turn into
// long-term memory. Zero-value optional fields are silently omitted.
type TurnSummaryInput struct {
	CampaignID   uuid.UUID
	PlayerInput  string
	Narrative    string
	ToolsUsed    []string   // tool names invoked during the turn
	StateChanges []string   // human-readable change descriptions
	LocationID   *uuid.UUID // optional
	NPCsInvolved []uuid.UUID
	InGameTime   string // optional, free-form in-game timestamp
}

// TurnEmbedder composes a textual summary of a game turn, embeds it, and
// persists the resulting memory row.
type TurnEmbedder struct {
	embedder Embedder
	store    MemoryCreateStore
}

// NewTurnEmbedder constructs a TurnEmbedder from an Embedder and a store.
func NewTurnEmbedder(embedder Embedder, store MemoryCreateStore) *TurnEmbedder {
	return &TurnEmbedder{embedder: embedder, store: store}
}

// composeSummary builds the plain-text summary stored alongside the vector.
func composeSummary(input TurnSummaryInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Player: %s\nOutcome: %s", input.PlayerInput, input.Narrative)
	if len(input.ToolsUsed) > 0 {
		fmt.Fprintf(&b, "\nActions: %s", strings.Join(input.ToolsUsed, ", "))
	}
	if len(input.StateChanges) > 0 {
		fmt.Fprintf(&b, "\nChanges: %s", strings.Join(input.StateChanges, ", "))
	}
	return b.String()
}

// turnMetadata is the JSON shape stored in CreateMemoryParams.Metadata.
type turnMetadata struct {
	ToolsUsed    []string `json:"tools_used"`
	StateChanges []string `json:"state_changes"`
}

// EmbedTurn embeds a turn summary and persists it as a memory row.
func (te *TurnEmbedder) EmbedTurn(ctx context.Context, input TurnSummaryInput) error {
	summary := composeSummary(input)

	vector, err := te.embedder.Embed(ctx, summary)
	if err != nil {
		return fmt.Errorf("embedding turn summary: %w", err)
	}

	meta := turnMetadata{
		ToolsUsed:    input.ToolsUsed,
		StateChanges: input.StateChanges,
	}
	// Normalise nil slices to empty arrays in JSON output.
	if meta.ToolsUsed == nil {
		meta.ToolsUsed = []string{}
	}
	if meta.StateChanges == nil {
		meta.StateChanges = []string{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshaling turn metadata: %w", err)
	}

	var locationID pgtype.UUID
	if input.LocationID != nil {
		locationID = dbutil.ToPgtype(*input.LocationID)
	}

	npcs := make([]pgtype.UUID, len(input.NPCsInvolved))
	for i, id := range input.NPCsInvolved {
		npcs[i] = dbutil.ToPgtype(id)
	}

	var inGameTime pgtype.Text
	if input.InGameTime != "" {
		inGameTime = pgtype.Text{String: input.InGameTime, Valid: true}
	}

	params := statedb.CreateMemoryParams{
		CampaignID:   dbutil.ToPgtype(input.CampaignID),
		Content:      summary,
		Embedding:    pgvector_go.NewVector(vector),
		MemoryType:   "turn_summary",
		LocationID:   locationID,
		NpcsInvolved: npcs,
		InGameTime:   inGameTime,
		Metadata:     metaJSON,
	}

	if _, err := te.store.CreateMemory(ctx, params); err != nil {
		return fmt.Errorf("storing turn memory: %w", err)
	}
	return nil
}
