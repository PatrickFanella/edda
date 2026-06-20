package engine

import (
	"encoding/json"
	"strings"
)

type DurableClaimKind string

const (
	DurableClaimMovement DurableClaimKind = "movement"
	DurableClaimQuest    DurableClaimKind = "quest"
	DurableClaimFact     DurableClaimKind = "fact"
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

func containsAny(s string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(s, phrase) {
			return true
		}
	}
	return false
}
