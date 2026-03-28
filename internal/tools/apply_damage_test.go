package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRegisterApplyDamage(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterApplyDamage(reg); err != nil {
		t.Fatalf("register apply_damage: %v", err)
	}

	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != applyDamageToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, applyDamageToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 5 {
		t.Fatalf("required field count = %d, want 5", len(required))
	}
}

func TestApplyDamageReducesHPAndRecordsDamageType(t *testing.T) {
	playerID := uuid.New()
	enemyID := uuid.New()
	h := NewApplyDamageHandler()

	result, err := h.Handle(context.Background(), map[string]any{
		"target_id":    enemyID.String(),
		"amount":       4,
		"damage_type":  "fire",
		"source":       "burning hands",
		"combat_state": baseCombatStateArgs(playerID, enemyID),
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
	if hp, _ := combatant["hp"].(int); hp != 5 {
		t.Fatalf("enemy HP = %v, want 5", combatant["hp"])
	}
	if status, _ := combatant["status"].(string); status != "alive" {
		t.Fatalf("enemy status = %v, want alive", combatant["status"])
	}

	damage, ok := result.Data["damage"].(map[string]any)
	if !ok {
		t.Fatalf("damage type = %T", result.Data["damage"])
	}
	if got, _ := damage["damage_type"].(string); got != "fire" {
		t.Fatalf("damage_type = %q, want fire", got)
	}
	if got, _ := damage["source"].(string); got != "burning hands" {
		t.Fatalf("source = %q, want burning hands", got)
	}
	if got, _ := damage["applied_amount"].(int); got != 4 {
		t.Fatalf("applied_amount = %d, want 4", got)
	}
}

func TestApplyDamageClampsAtZeroAndKillsNPC(t *testing.T) {
	playerID := uuid.New()
	enemyID := uuid.New()
	h := NewApplyDamageHandler()

	result, err := h.Handle(context.Background(), map[string]any{
		"target_id":    enemyID.String(),
		"amount":       50,
		"damage_type":  "slashing",
		"source":       "greatsword",
		"combat_state": baseCombatStateArgs(playerID, enemyID),
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	combatant := result.Data["combatant"].(map[string]any)
	if hp, _ := combatant["hp"].(int); hp != 0 {
		t.Fatalf("enemy HP = %v, want 0", combatant["hp"])
	}
	if status, _ := combatant["status"].(string); status != "dead" {
		t.Fatalf("enemy status = %v, want dead", combatant["status"])
	}
}

func TestApplyDamagePlayerAtZeroBecomesUnconscious(t *testing.T) {
	playerID := uuid.New()
	enemyID := uuid.New()
	h := NewApplyDamageHandler()

	result, err := h.Handle(context.Background(), map[string]any{
		"target_id":    playerID.String(),
		"amount":       24,
		"damage_type":  "necrotic",
		"source":       "death ray",
		"combat_state": baseCombatStateArgs(playerID, enemyID),
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	combatant := result.Data["combatant"].(map[string]any)
	if hp, _ := combatant["hp"].(int); hp != 0 {
		t.Fatalf("player HP = %v, want 0", combatant["hp"])
	}
	if status, _ := combatant["status"].(string); status != "unconscious" {
		t.Fatalf("player status = %v, want unconscious", combatant["status"])
	}
}

func TestApplyDamageNilHandler(t *testing.T) {
	var h *ApplyDamageHandler
	_, err := h.Handle(context.Background(), map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "handler is nil") {
		t.Fatalf("expected nil handler error, got %v", err)
	}
}

func TestApplyDamageDeadTargetNoOpMetadataAndNarrative(t *testing.T) {
	playerID := uuid.New()
	enemyID := uuid.New()
	state := baseCombatStateArgs(playerID, enemyID)
	combatants := state["combatants"].([]any)
	enemy := combatants[1].(map[string]any)
	enemy["hp"] = 0
	enemy["status"] = "dead"

	h := NewApplyDamageHandler()
	result, err := h.Handle(context.Background(), map[string]any{
		"target_id":    enemyID.String(),
		"amount":       7,
		"damage_type":  "fire",
		"source":       "fire bolt",
		"combat_state": state,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	damage := result.Data["damage"].(map[string]any)
	if got, _ := damage["applied_amount"].(int); got != 0 {
		t.Fatalf("applied_amount = %d, want 0", got)
	}
	combatant := result.Data["combatant"].(map[string]any)
	if hp, _ := combatant["hp"].(int); hp != 0 {
		t.Fatalf("enemy HP = %d, want 0", hp)
	}
	if !strings.Contains(result.Narrative, "already dead") {
		t.Fatalf("narrative = %q, want already dead", result.Narrative)
	}
}
