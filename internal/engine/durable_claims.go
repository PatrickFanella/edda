package engine

import (
	"encoding/json"
	"strings"
)

type DurableClaimKind string

const (
	DurableClaimMovement         DurableClaimKind = "movement"
	DurableClaimQuest            DurableClaimKind = "quest"
	DurableClaimFact             DurableClaimKind = "fact"
	DurableClaimInventoryCreated DurableClaimKind = "inventory_created"
	DurableClaimInventoryRemoved DurableClaimKind = "inventory_removed"
	DurableClaimInventoryUpdated DurableClaimKind = "inventory_updated"
	DurableClaimCombatStarted    DurableClaimKind = "combat_started"
	DurableClaimCombatResolved   DurableClaimKind = "combat_resolved"
)

type DurableClaimIssue struct {
	Kind    DurableClaimKind
	Message string
}

func AuditDurableClaims(narrative string, applied []AppliedToolCall, advertised []string) []DurableClaimIssue {
	lower := strings.ToLower(narrative)
	_ = advertised

	var sawMovePlayer bool
	var sawCreateLocationMove bool
	var sawQuestTool bool
	var sawFactTool bool
	var sawInventoryCreateTool bool
	var sawInventoryRemoveTool bool
	var sawInventoryUpdateTool bool
	var sawCombatStartTool bool
	var sawCombatResolveTool bool

	for _, call := range applied {
		switch call.Tool {
		case "move_player":
			sawMovePlayer = true
		case "create_location":
			if hasMovePlayerHere(call.Result) {
				sawCreateLocationMove = true
			}
		case "create_quest", "update_quest", "complete_objective":
			sawQuestTool = true
		case "establish_fact", "revise_fact":
			sawFactTool = true
		case "add_item", "create_item":
			sawInventoryCreateTool = true
		case "remove_item":
			sawInventoryRemoveTool = true
		case "modify_item", "update_item":
			sawInventoryUpdateTool = true
		case "initiate_combat":
			sawCombatStartTool = true
		case "resolve_combat":
			sawCombatResolveTool = true
		}
	}

	issues := make([]DurableClaimIssue, 0, 3)
	if claimsMovement(lower) && !sawMovePlayer && !sawCreateLocationMove {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimMovement, Message: "narrative claims movement without matching applied state tool"})
	}
	if claimsQuest(lower) && !sawQuestTool {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimQuest, Message: "narrative claims quest journal change without matching applied quest tool"})
	}
	if claimsFact(lower) && !sawFactTool {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimFact, Message: "narrative claims durable knowledge without matching applied fact tool"})
	}
	if claimsInventoryCreated(lower) && !sawInventoryCreateTool {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimInventoryCreated, Message: "narrative claims inventory creation without matching applied inventory create tool"})
	}
	if claimsInventoryRemoved(lower) && !sawInventoryRemoveTool {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimInventoryRemoved, Message: "narrative claims inventory removal without matching applied inventory remove tool"})
	}
	if claimsInventoryUpdated(lower) && !sawInventoryUpdateTool {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimInventoryUpdated, Message: "narrative claims inventory update without matching applied inventory update tool"})
	}
	if claimsCombatStarted(lower) && !sawCombatStartTool {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimCombatStarted, Message: "narrative claims combat start without matching applied combat tool"})
	}
	if claimsCombatResolved(lower) && !sawCombatResolveTool {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimCombatResolved, Message: "narrative claims combat resolution without matching applied combat tool"})
	}
	return issues
}

func hasMovePlayerHere(result json.RawMessage) bool {
	var data map[string]any
	if err := json.Unmarshal(result, &data); err != nil {
		return false
	}
	moved, _ := data["move_player_here"].(bool)
	return moved
}

func claimsMovement(s string) bool {
	return containsAny(s, []string{
		"you arrive at ",
		"you arrive in ",
		"you enter ",
		"you step into ",
		"you return to ",
		"you make your way back to ",
		"you step through the doorway",
		"you step through the threshold",
		"you step through the gate",
		"you step through the hatch",
		"you step through the portal",
	})
}

func claimsQuest(s string) bool {
	return containsAny(s, []string{
		"new quest",
		"quest added",
		"added to your journal",
	})
}

func claimsFact(s string) bool {
	return containsAny(s, []string{
		"you now know",
		"you learn that",
		"you confirm that",
	})
}

func claimsInventoryCreated(s string) bool {
	return containsAny(s, []string{
		"added to your inventory",
		"now in your inventory",
		"you add it to your inventory",
		"you place it in your inventory",
		"you place it in your pack",
		"you place it in your pouch",
		"you place it in your bag",
		"you pocket the ",
	})
}

func claimsInventoryRemoved(s string) bool {
	return containsAny(s, []string{
		"removed from your inventory",
		"leaves your inventory",
		"no longer in your inventory",
		"you remove it from your inventory",
		"you drop the item",
		"you drop the item from your inventory",
		"you hand over the item",
		"you hand over the item from your inventory",
		"you consume the item",
		"you consume the item from your inventory",
	})
}

func claimsInventoryUpdated(s string) bool {
	return containsAny(s, []string{
		"updated in your inventory",
		"modified in your inventory",
	})
}

func claimsCombatStarted(s string) bool {
	if containsAny(s, []string{
		"combat begins",
		"combat starts",
		"roll initiative",
		"the fight begins",
		"the battle begins",
	}) {
		return true
	}
	if strings.Contains(s, "drone") && containsAny(s, []string{
		"locks onto you",
	}) {
		return true
	}
	return false
}

func claimsCombatResolved(s string) bool {
	return containsAny(s, []string{
		"combat ends",
		"combat is over",
		"the fight ends",
		"the battle ends",
		"encounter resolved",
		"you defeat the combat",
		"you defeat the encounter",
		"the drone is disabled",
	})
}

func containsAny(s string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(s, phrase) {
			return true
		}
	}
	return false
}
