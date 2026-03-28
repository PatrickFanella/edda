package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/PatrickFanella/game-master/internal/combat"
	"github.com/PatrickFanella/game-master/internal/llm"
)

const applyDamageToolName = "apply_damage"

// ApplyDamageTool returns the apply_damage tool definition and JSON schema.
func ApplyDamageTool() llm.Tool {
	return llm.Tool{
		Name:        applyDamageToolName,
		Description: "Apply damage to a combatant, update HP/status, and return the updated combatant and combat state.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target_id": map[string]any{
					"type":        "string",
					"description": "UUID of the combatant receiving damage.",
				},
				"amount": map[string]any{
					"type":        "integer",
					"description": "Amount of damage to apply.",
				},
				"damage_type": map[string]any{
					"type":        "string",
					"description": "Damage type (e.g. slashing, fire, psychic).",
				},
				"source": map[string]any{
					"type":        "string",
					"description": "Source of the damage.",
				},
				"combat_state": map[string]any{
					"type":        "object",
					"description": "Current combat state containing the target combatant.",
				},
			},
			"required":             []string{"target_id", "amount", "damage_type", "source", "combat_state"},
			"additionalProperties": false,
		},
	}
}

// RegisterApplyDamage registers the apply_damage tool and handler.
func RegisterApplyDamage(reg *Registry) error {
	return reg.Register(ApplyDamageTool(), NewApplyDamageHandler().Handle)
}

// ApplyDamageHandler executes apply_damage tool calls.
type ApplyDamageHandler struct{}

// NewApplyDamageHandler creates a new apply_damage handler.
func NewApplyDamageHandler() *ApplyDamageHandler {
	return &ApplyDamageHandler{}
}

// Handle executes the apply_damage tool.
func (h *ApplyDamageHandler) Handle(_ context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("apply_damage handler is nil")
	}

	targetID, err := parseUUIDArg(args, "target_id")
	if err != nil {
		return nil, err
	}
	amount, err := parseIntArg(args, "amount")
	if err != nil {
		return nil, err
	}
	if amount < 0 {
		return nil, errors.New("amount must be greater than or equal to 0")
	}
	damageType, err := parseStringArg(args, "damage_type")
	if err != nil {
		return nil, err
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

	hpBefore := target.HP
	combat.ApplyDamage(target, amount)

	damage := map[string]any{
		"target_id":   targetID.String(),
		"amount":      amount,
		"damage_type": damageType,
		"source":      source,
		"hp_before":   hpBefore,
		"hp_after":    target.HP,
	}

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"combatant":    combatantStateMap(target),
			"combat_state": combatStateToMap(state),
			"damage":       damage,
		},
		Narrative: fmt.Sprintf("%s takes %d %s damage from %s. HP is now %d/%d.", target.Name, amount, damageType, source, target.HP, target.MaxHP),
	}, nil
}

func combatantStateMap(c *combat.Combatant) map[string]any {
	conditions := make([]map[string]any, 0, len(c.Conditions))
	for _, cond := range c.Conditions {
		conditions = append(conditions, map[string]any{
			"name":            cond.Name,
			"duration_rounds": cond.DurationRounds,
		})
	}

	return map[string]any{
		"entity_id":   c.EntityID.String(),
		"entity_type": string(c.EntityType),
		"name":        c.Name,
		"hp":          c.HP,
		"max_hp":      c.MaxHP,
		"conditions":  conditions,
		"status":      string(c.Status),
	}
}
