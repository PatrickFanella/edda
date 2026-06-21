package tools

import (
	"context"
	"errors"
	"fmt"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
)

const movePlayerToolName = "move_player"

// MovePlayerStore moves a player character through campaign-scoped state.
type MovePlayerStore interface {
	MovePlayer(ctx context.Context, cmd domain.MovePlayerCommand) (*domain.MovePlayerResult, error)
}

// MovePlayerTool returns the move_player tool definition and JSON schema.
func MovePlayerTool() llm.Tool {
	return llm.Tool{
		Name:        movePlayerToolName,
		Description: "Move the player character to a connected location.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location_id": map[string]any{
					"type":        "string",
					"description": "Destination location UUID.",
				},
			},
			"required":             []string{"location_id"},
			"additionalProperties": false,
		},
	}
}

// RegisterMovePlayer registers the move_player tool and handler.
func RegisterMovePlayer(reg *Registry, store MovePlayerStore) error {
	if store == nil {
		return errors.New("move_player store is required")
	}
	return reg.Register(MovePlayerTool(), NewMovePlayerHandler(store).Handle)
}

// MovePlayerHandler executes move_player tool calls.
type MovePlayerHandler struct {
	store MovePlayerStore
}

// NewMovePlayerHandler creates a new move_player handler.
func NewMovePlayerHandler(store MovePlayerStore) *MovePlayerHandler {
	return &MovePlayerHandler{store: store}
}

// Handle executes the move_player tool.
func (h *MovePlayerHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("move_player handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("move_player store is required")
	}

	targetLocationID, err := parseUUIDArg(args, "location_id")
	if err != nil {
		return nil, err
	}

	currentLocationID, ok := CurrentLocationIDFromContext(ctx)
	if !ok {
		return nil, errors.New("move_player requires current location id in context")
	}

	playerCharacterID, ok := CurrentPlayerCharacterIDFromContext(ctx)
	if !ok {
		return nil, errors.New("move_player requires current player character id in context")
	}
	campaignID, ok := CurrentCampaignIDFromContext(ctx)
	if !ok {
		return nil, errors.New("move_player requires current campaign id in context")
	}

	result, err := h.store.MovePlayer(ctx, domain.MovePlayerCommand{CampaignID: campaignID, PlayerCharacterID: playerCharacterID, CurrentLocationID: currentLocationID, TargetLocationID: targetLocationID})
	if err != nil {
		return nil, fmt.Errorf("move player: %w", err)
	}

	narrative := fmt.Sprintf("Player moved to %s.", result.ToLocationName)
	if result.TravelTime != "" && result.TimeWarning == "" {
		narrative = fmt.Sprintf("Player moved to %s. The journey took %s. It is now Day %d, %02d:%02d.", result.ToLocationName, result.TravelTime, result.Day, result.Hour, result.Minute)
	}
	if result.TimeWarning != "" {
		narrative += " Time update warning: " + result.TimeWarning + "."
	}
	if result.VisitedWarning != "" {
		narrative += " Visit marker warning: " + result.VisitedWarning + "."
	}

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"location_id":         targetLocationID.String(),
			"campaign_id":         campaignID.String(),
			"player_character_id": playerCharacterID.String(),
			"name":                result.ToLocationName,
			"description":         result.ToLocationDescription,
			"location_type":       result.ToLocationType,
			"travel_time":         result.TravelTime,
			"day":                 result.Day,
			"hour":                result.Hour,
			"minute":              result.Minute,
			"visited_marked":      result.VisitedMarked,
			"visited_warning":     result.VisitedWarning,
			"time_warning":        result.TimeWarning,
		},
		Narrative: narrative,
	}, nil
}
