package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/domain"
	"github.com/PatrickFanella/game-master/internal/llm"
)

const updatePlayerStatusToolName = "update_player_status"

var allowedPlayerStatuses = []string{"healthy", "poisoned", "cursed", "resting", "unconscious", "dead"}

var negativePlayerStatuses = map[string]struct{}{
	"poisoned":    {},
	"cursed":      {},
	"unconscious": {},
}

type statusDuration struct {
	Unit  string `json:"unit"`
	Value string `json:"value"`
}

type playerStatusEntry struct {
	Status   string          `json:"status"`
	Duration *statusDuration `json:"duration,omitempty"`
}

type playerStatusState struct {
	Mode       string              `json:"mode,omitempty"`
	Conditions []playerStatusEntry `json:"conditions"`
}

// UpdatePlayerStatusStore provides player status lookup and persistence for update_player_status.
type UpdatePlayerStatusStore interface {
	GetPlayerCharacterByID(ctx context.Context, playerCharacterID uuid.UUID) (*domain.PlayerCharacter, error)
	UpdatePlayerStatus(ctx context.Context, playerCharacterID uuid.UUID, status string) error
}

// UpdatePlayerStatusTool returns the update_player_status tool definition and JSON schema.
func UpdatePlayerStatusTool() llm.Tool {
	return llm.Tool{
		Name:        updatePlayerStatusToolName,
		Description: "Add or refresh a player status condition, optionally with a duration.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"type":        "string",
					"description": "Status name: healthy, poisoned, cursed, resting, unconscious, or dead.",
					"enum":        allowedPlayerStatuses,
				},
				"duration": map[string]any{
					"type":        "object",
					"description": "Optional duration metadata. Unit must be turns or in_game_time.",
					"properties": map[string]any{
						"unit": map[string]any{
							"type": "string",
							"enum": []string{"turns", "in_game_time"},
						},
						"value": map[string]any{
							"type":        "string",
							"description": "Duration value (e.g., '3' for turns, or '2 hours').",
						},
					},
					"required":             []string{"unit", "value"},
					"additionalProperties": false,
				},
			},
			"required":             []string{"status"},
			"additionalProperties": false,
		},
	}
}

// RegisterUpdatePlayerStatus registers the update_player_status tool and handler.
func RegisterUpdatePlayerStatus(reg *Registry, store UpdatePlayerStatusStore) error {
	if store == nil {
		return errors.New("update_player_status store is required")
	}
	return reg.Register(UpdatePlayerStatusTool(), NewUpdatePlayerStatusHandler(store).Handle)
}

// UpdatePlayerStatusHandler executes update_player_status tool calls.
type UpdatePlayerStatusHandler struct {
	store UpdatePlayerStatusStore
}

// NewUpdatePlayerStatusHandler creates a new update_player_status handler.
func NewUpdatePlayerStatusHandler(store UpdatePlayerStatusStore) *UpdatePlayerStatusHandler {
	return &UpdatePlayerStatusHandler{store: store}
}

// Handle executes the update_player_status tool.
func (h *UpdatePlayerStatusHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("update_player_status handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("update_player_status store is required")
	}

	playerCharacterID, ok := CurrentPlayerCharacterIDFromContext(ctx)
	if !ok {
		return nil, errors.New("update_player_status requires current player character id in context")
	}

	statusName, err := parseStringArg(args, "status")
	if err != nil {
		return nil, err
	}
	statusName = strings.ToLower(strings.TrimSpace(statusName))
	if !slices.Contains(allowedPlayerStatuses, statusName) {
		return nil, errors.New("status must be one of: healthy, poisoned, cursed, resting, unconscious, dead")
	}

	duration, err := parseOptionalStatusDurationArg(args, "duration")
	if err != nil {
		return nil, err
	}

	playerCharacter, err := h.store.GetPlayerCharacterByID(ctx, playerCharacterID)
	if err != nil {
		return nil, fmt.Errorf("get player character: %w", err)
	}
	if playerCharacter == nil {
		return nil, errors.New("current player character does not exist")
	}

	currentState, err := parsePersistedStatusState(playerCharacter.Status)
	if err != nil {
		return nil, err
	}
	if hasStatus(currentState.Conditions, "dead") && statusName != "dead" {
		return nil, errors.New("cannot update status for a dead player character")
	}

	updatedStatuses := applyStatusUpdate(currentState.Conditions, statusName, duration)
	persistedStatus, err := marshalPersistedStatusState(playerStatusState{
		Mode:       currentState.Mode,
		Conditions: updatedStatuses,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal updated status: %w", err)
	}

	if err := h.store.UpdatePlayerStatus(ctx, playerCharacterID, persistedStatus); err != nil {
		return nil, fmt.Errorf("update player status: %w", err)
	}

	appliedDuration := durationForStatus(updatedStatuses, statusName)
	data := map[string]any{
		"player_character_id": playerCharacterID.String(),
		"status":              statusName,
		"statuses":            updatedStatuses,
	}
	if currentState.Mode != "" {
		data["mode"] = currentState.Mode
	}
	if appliedDuration != nil {
		data["duration"] = appliedDuration
	}
	if statusName == "dead" {
		data["game_over"] = true
	}

	return &ToolResult{
		Success: true,
		Data:    data,
	}, nil
}

func parseOptionalStatusDurationArg(args map[string]any, key string) (*statusDuration, error) {
	raw, ok := args[key]
	if !ok {
		return nil, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", key)
	}
	unit, err := parseStringArg(obj, "unit")
	if err != nil {
		return nil, fmt.Errorf("%s.%w", key, err)
	}
	unit = strings.ToLower(strings.TrimSpace(unit))
	if unit != "turns" && unit != "in_game_time" {
		return nil, fmt.Errorf("%s.unit must be one of: turns, in_game_time", key)
	}
	value, err := parseStringArg(obj, "value")
	if err != nil {
		return nil, fmt.Errorf("%s.%w", key, err)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("%s.value must be a non-empty string", key)
	}
	return &statusDuration{Unit: unit, Value: value}, nil
}

func parsePersistedStatusState(raw string) (playerStatusState, error) {
	state := playerStatusState{Conditions: []playerStatusEntry{}}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return state, nil
	}
	if strings.HasPrefix(trimmed, "{") {
		if err := json.Unmarshal([]byte(trimmed), &state); err != nil {
			return playerStatusState{}, fmt.Errorf("unmarshal player status object: %w", err)
		}
		for i := range state.Conditions {
			state.Conditions[i].Status = strings.ToLower(strings.TrimSpace(state.Conditions[i].Status))
		}
		state.Mode = strings.ToLower(strings.TrimSpace(state.Mode))
		return state, nil
	}
	if !strings.HasPrefix(trimmed, "[") {
		lowerTrimmed := strings.ToLower(trimmed)
		if lowerTrimmed == "active" || lowerTrimmed == "in_combat" || lowerTrimmed == "defeated" {
			state.Mode = lowerTrimmed
			return state, nil
		}
		state.Conditions = []playerStatusEntry{{Status: lowerTrimmed}}
		return state, nil
	}
	if err := json.Unmarshal([]byte(trimmed), &state.Conditions); err != nil {
		return playerStatusState{}, fmt.Errorf("unmarshal player status: %w", err)
	}
	for i := range state.Conditions {
		state.Conditions[i].Status = strings.ToLower(strings.TrimSpace(state.Conditions[i].Status))
	}
	return state, nil
}

func marshalPersistedStatusState(state playerStatusState) (string, error) {
	if state.Mode == "" {
		payload, err := json.Marshal(state.Conditions)
		if err != nil {
			return "", err
		}
		return string(payload), nil
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func applyStatusUpdate(current []playerStatusEntry, next string, duration *statusDuration) []playerStatusEntry {
	if next == "healthy" {
		duration = nil
	}
	updated := make([]playerStatusEntry, 0, len(current)+1)
	for _, status := range current {
		if status.Status == "" {
			continue
		}
		if status.Status == next {
			if duration == nil && next != "healthy" && next != "dead" {
				duration = status.Duration
			}
			continue
		}
		if next == "healthy" {
			if _, negative := negativePlayerStatuses[status.Status]; negative {
				continue
			}
		}
		if next != "healthy" && status.Status == "healthy" {
			continue
		}
		if next == "dead" {
			continue
		}
		updated = append(updated, status)
	}
	switch next {
	case "dead":
		return []playerStatusEntry{{Status: "dead"}}
	case "healthy":
		updated = append(updated, playerStatusEntry{Status: "healthy"})
	default:
		updated = append(updated, playerStatusEntry{Status: next, Duration: duration})
	}
	return updated
}

func hasStatus(statuses []playerStatusEntry, status string) bool {
	for _, s := range statuses {
		if strings.EqualFold(s.Status, status) {
			return true
		}
	}
	return false
}

func durationForStatus(statuses []playerStatusEntry, status string) *statusDuration {
	for _, s := range statuses {
		if strings.EqualFold(s.Status, status) {
			return s.Duration
		}
	}
	return nil
}
