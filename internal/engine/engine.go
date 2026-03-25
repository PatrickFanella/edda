// Package engine defines the core GameEngine interface that both the TUI and
// API server consume. It provides the primary entry points for processing
// player turns, managing campaigns, and querying game state.
package engine

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/domain"
)

// GameEngine is the primary interface consumed by the TUI and API server.
// It orchestrates turn processing, campaign management, and state queries.
type GameEngine interface {
	// ProcessTurn processes the player's input for the given campaign,
	// returning a TurnResult that contains the narrative response,
	// any tool calls that were applied, suggested choices, and state
	// changes.
	ProcessTurn(ctx context.Context, campaignID uuid.UUID, playerInput string) (*TurnResult, error)

	// GetGameState returns the current state of the specified campaign.
	GetGameState(ctx context.Context, campaignID uuid.UUID) (*GameState, error)

	// NewCampaign creates a new campaign owned by the given user.
	NewCampaign(ctx context.Context, userID uuid.UUID) (*domain.Campaign, error)

	// LoadCampaign loads an existing campaign into the engine so that
	// subsequent calls to ProcessTurn and GetGameState can operate on it.
	LoadCampaign(ctx context.Context, campaignID uuid.UUID) error
}

// ---------------------------------------------------------------------------
// Turn result types
// ---------------------------------------------------------------------------

// TurnResult holds the outcome of a single player turn.
type TurnResult struct {
	// Narrative is the descriptive text generated for this turn.
	Narrative string
	// AppliedToolCalls lists the tool invocations that were executed
	// during turn processing.
	AppliedToolCalls []AppliedToolCall
	// Choices contains suggested actions the player may take next.
	Choices []Choice
	// StateChanges describes modifications made to the game state as a
	// result of this turn.
	StateChanges []StateChange
}

// AppliedToolCall records a single tool invocation that occurred while
// processing a turn.
type AppliedToolCall struct {
	// Tool is the name of the tool that was invoked.
	Tool string
	// Arguments holds the input parameters passed to the tool,
	// serialized as JSON.
	Arguments json.RawMessage
	// Result holds the output returned by the tool, serialized as JSON.
	Result json.RawMessage
}

// Choice represents a suggested action the player can take.
type Choice struct {
	// ID is a stable identifier for this choice, suitable for
	// programmatic selection.
	ID string
	// Text is the human-readable description shown to the player.
	Text string
}

// StateChange records a single modification to the game state.
type StateChange struct {
	// Entity is the kind of entity that was modified (e.g. "location",
	// "npc", "quest").
	Entity string
	// EntityID is the unique identifier of the modified entity.
	EntityID uuid.UUID
	// Field is the name of the field that changed.
	Field string
	// OldValue is the previous value, serialized as JSON.
	OldValue json.RawMessage
	// NewValue is the updated value, serialized as JSON.
	NewValue json.RawMessage
}

// ---------------------------------------------------------------------------
// Game state
// ---------------------------------------------------------------------------

// GameState represents the current state of a campaign as seen by the player.
type GameState struct {
	// CurrentLocation is the player's current location in the game world.
	CurrentLocation domain.Location
	// PlayerCharacter is the player's character in this campaign.
	PlayerCharacter domain.PlayerCharacter
	// NPCsPresent lists the NPCs at the player's current location.
	NPCsPresent []domain.NPC
	// ActiveQuests lists the quests the player is currently pursuing.
	ActiveQuests []domain.Quest
}
