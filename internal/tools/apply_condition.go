package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/PatrickFanella/game-master/internal/combat"
	"github.com/PatrickFanella/game-master/internal/llm"
)

const applyConditionToolName = "apply_condition"

// ApplyConditionTool returns the apply_condition tool definition and JSON schema.
func ApplyConditionTool() llm.Tool {
	return llm.Tool{
		Name:        applyConditionToolName,
		Description: "Apply a condition to a combatant, refreshing duration when the condition already exists.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target_id": map[string]any{
					"type":        "string",
					"description": "UUID of the combatant receiving the condition.",
				},
				"condition": map[string]any{
					"type":        "string",
					"description": "Condition name (e.g. stunned, poisoned).",
				},
				"duration_rounds": map[string]any{
					"type":        "integer",
					"description": "Condition duration in rounds. Use -1 for a permanent condition.",
				},
				"source": map[string]any{
					"type":        "string",
					"description": "Source of the condition.",
				},
				"combat_state": map[string]any{
					"type":        "object",
					"description": "Current combat state containing the target combatant.",
				},
			},
			"required":             []string{"target_id", "condition", "duration_rounds", "source", "combat_state"},
			"additionalProperties": false,
		},
	}
}

// RegisterApplyCondition registers the apply_condition tool and handler.
func RegisterApplyCondition(reg *Registry) error {
	return reg.Register(ApplyConditionTool(), NewApplyConditionHandler().Handle)
}

// ApplyConditionHandler executes apply_condition tool calls.
type ApplyConditionHandler struct{}

// NewApplyConditionHandler creates a new apply_condition handler.
func NewApplyConditionHandler() *ApplyConditionHandler {
	return &ApplyConditionHandler{}
}

// Handle executes the apply_condition tool.
func (h *ApplyConditionHandler) Handle(_ context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("apply_condition handler is nil")
	}

	targetID, err := parseUUIDArg(args, "target_id")
	if err != nil {
		return nil, err
	}
	condition, err := parseStringArg(args, "condition")
	if err != nil {
		return nil, err
	}
	durationRounds, err := parseIntArg(args, "duration_rounds")
	if err != nil {
		return nil, err
	}
	if durationRounds < 1 && durationRounds != combat.PermanentDuration {
		return nil, errors.New("duration_rounds must be greater than 0 or -1 for permanent")
	}
	source, err := parseStringArg(args, "source")
	if err != nil {
		return nil, err
	}
	state, err := parseCombatStateArg(args, "combat_state")
	if err != nil {
		return nil, err
	}

	target := combatantByID(state, targetID)
	if target == nil {
		return nil, fmt.Errorf("target combatant %s not found", targetID)
	}

	combat.AddCondition(target, condition, durationRounds)

	narrativeDuration := fmt.Sprintf("for %d rounds", durationRounds)
	if durationRounds == combat.PermanentDuration {
		narrativeDuration = "permanently"
	}

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"combatant":    combatantStateMap(target),
			"combat_state": combatStateToMap(state),
			"condition": map[string]any{
				"target_id":       targetID.String(),
				"name":            condition,
				"duration_rounds": durationRounds,
				"source":          source,
			},
		},
		Narrative: fmt.Sprintf("%s is now %s %s (source: %s).", target.Name, condition, narrativeDuration, source),
	}, nil
}
