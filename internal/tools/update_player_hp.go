package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
)

const updatePlayerHPToolName = "update_player_hp"

// UpdatePlayerHPStore provides player lookup and HP persistence for update_player_hp.
type UpdatePlayerHPStore interface {
	GetPlayerCharacterByID(ctx context.Context, playerCharacterID uuid.UUID) (*domain.PlayerCharacter, error)
	UpdatePlayerCurrentHP(ctx context.Context, playerCharacterID uuid.UUID, hp int) error
}

// UpdatePlayerHPTool returns the update_player_hp tool definition and JSON schema.
func UpdatePlayerHPTool() llm.Tool {
	return llm.Tool{
		Name:        updatePlayerHPToolName,
		Description: "Update the player character's current HP from an exploration hazard, healing, or other non-combat effect. This changes current HP only; do not use it to change maximum HP.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"hp": map[string]any{
					"type":        "integer",
					"description": "Current HP after the effect is applied.",
				},
			},
			"required":             []string{"hp"},
			"additionalProperties": false,
		},
	}
}

// RegisterUpdatePlayerHP registers the update_player_hp tool and handler.
func RegisterUpdatePlayerHP(reg *Registry, store UpdatePlayerHPStore) error {
	if store == nil {
		return errors.New("update_player_hp store is required")
	}
	return reg.Register(UpdatePlayerHPTool(), NewUpdatePlayerHPHandler(store).Handle)
}

// UpdatePlayerHPHandler executes update_player_hp tool calls.
type UpdatePlayerHPHandler struct {
	store UpdatePlayerHPStore
}

// NewUpdatePlayerHPHandler creates a new update_player_hp handler.
func NewUpdatePlayerHPHandler(store UpdatePlayerHPStore) *UpdatePlayerHPHandler {
	return &UpdatePlayerHPHandler{store: store}
}

// Handle executes the update_player_hp tool.
func (h *UpdatePlayerHPHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("update_player_hp handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("update_player_hp store is required")
	}
	for key := range args {
		if key != "hp" {
			return nil, fmt.Errorf("%s is not supported by update_player_hp", key)
		}
	}

	playerCharacterID, ok := CurrentPlayerCharacterIDFromContext(ctx)
	if !ok {
		return nil, errors.New("update_player_hp requires current player character id in context")
	}

	hp, err := parseIntArg(args, "hp")
	if err != nil {
		return nil, err
	}
	if hp < 0 {
		return nil, errors.New("hp must be greater than or equal to 0")
	}

	playerCharacter, err := h.store.GetPlayerCharacterByID(ctx, playerCharacterID)
	if err != nil {
		return nil, fmt.Errorf("get player character: %w", err)
	}
	if playerCharacter == nil {
		return nil, errors.New("current player character does not exist")
	}

	maxHP := playerCharacter.MaxHP
	if hp > maxHP {
		return nil, errors.New("hp must be less than or equal to current max_hp")
	}

	oldHP := playerCharacter.HP
	oldMaxHP := playerCharacter.MaxHP
	if err := h.store.UpdatePlayerCurrentHP(ctx, playerCharacterID, hp); err != nil {
		return nil, fmt.Errorf("update player hp: %w", err)
	}

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"player_character_id": playerCharacterID.String(),
			"hp":                  hp,
			"max_hp":              maxHP,
			"old_hp":              oldHP,
			"old_max_hp":          oldMaxHP,
		},
		Narrative: fmt.Sprintf("Player HP updated from %d/%d to %d/%d.", oldHP, oldMaxHP, hp, maxHP),
	}, nil
}
