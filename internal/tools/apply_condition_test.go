package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRegisterApplyCondition(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterApplyCondition(reg); err != nil {
		t.Fatalf("register apply_condition: %v", err)
	}

	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != applyConditionToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, applyConditionToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 5 {
		t.Fatalf("required field count = %d, want 5", len(required))
	}
}

func TestApplyConditionAddsConditionToCombatant(t *testing.T) {
	playerID := uuid.New()
	enemyID := uuid.New()
	h := NewApplyConditionHandler()

	result, err := h.Handle(context.Background(), map[string]any{
		"target_id":       enemyID.String(),
		"condition":       "poisoned",
		"duration_rounds": 3,
		"source":          "venom blade",
		"combat_state":    baseCombatStateArgs(playerID, enemyID),
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success=true")
	}

	combatant, ok := result.Data["combatant"].(map[string]any)
	if !ok {
		t.Fatalf("combatant type = %T", result.Data["combatant"])
	}
	conditions, ok := combatant["conditions"].([]map[string]any)
	if !ok {
		t.Fatalf("conditions type = %T", combatant["conditions"])
	}
	if len(conditions) != 1 {
		t.Fatalf("conditions count = %d, want 1", len(conditions))
	}
	if name, _ := conditions[0]["name"].(string); name != "poisoned" {
		t.Fatalf("condition name = %q, want poisoned", name)
	}
	if duration, _ := conditions[0]["duration_rounds"].(int); duration != 3 {
		t.Fatalf("condition duration = %d, want 3", duration)
	}
}

func TestApplyConditionRefreshesDuplicateDuration(t *testing.T) {
	playerID := uuid.New()
	enemyID := uuid.New()
	state := baseCombatStateArgs(playerID, enemyID)
	combatants := state["combatants"].([]any)
	enemy := combatants[1].(map[string]any)
	enemy["conditions"] = []any{
		map[string]any{"name": "poisoned", "duration_rounds": 1},
	}

	h := NewApplyConditionHandler()
	result, err := h.Handle(context.Background(), map[string]any{
		"target_id":       enemyID.String(),
		"condition":       "poisoned",
		"duration_rounds": 4,
		"source":          "venom cloud",
		"combat_state":    state,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	combatant := result.Data["combatant"].(map[string]any)
	conditions := combatant["conditions"].([]map[string]any)
	if len(conditions) != 1 {
		t.Fatalf("conditions count = %d, want 1", len(conditions))
	}
	if duration, _ := conditions[0]["duration_rounds"].(int); duration != 4 {
		t.Fatalf("condition duration = %d, want 4", duration)
	}
}

func TestApplyConditionPermanentDuration(t *testing.T) {
	playerID := uuid.New()
	enemyID := uuid.New()
	h := NewApplyConditionHandler()

	result, err := h.Handle(context.Background(), map[string]any{
		"target_id":       enemyID.String(),
		"condition":       "blinded",
		"duration_rounds": -1,
		"source":          "flash powder",
		"combat_state":    baseCombatStateArgs(playerID, enemyID),
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	combatant := result.Data["combatant"].(map[string]any)
	conditions := combatant["conditions"].([]map[string]any)
	if len(conditions) != 1 {
		t.Fatalf("conditions count = %d, want 1", len(conditions))
	}
	if duration, _ := conditions[0]["duration_rounds"].(int); duration != -1 {
		t.Fatalf("condition duration = %d, want -1", duration)
	}

	condition, ok := result.Data["condition"].(map[string]any)
	if !ok {
		t.Fatalf("condition type = %T", result.Data["condition"])
	}
	if name, _ := condition["name"].(string); name != "blinded" {
		t.Fatalf("condition name = %q, want blinded", name)
	}
	if duration, _ := condition["duration_rounds"].(int); duration != -1 {
		t.Fatalf("condition payload duration = %d, want -1", duration)
	}
	if !strings.Contains(result.Narrative, "permanently") {
		t.Fatalf("narrative = %q, want permanently wording", result.Narrative)
	}
}

func TestApplyConditionNilHandler(t *testing.T) {
	var h *ApplyConditionHandler
	_, err := h.Handle(context.Background(), map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "handler is nil") {
		t.Fatalf("expected nil handler error, got %v", err)
	}
}
