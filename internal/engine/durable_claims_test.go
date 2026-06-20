package engine

import "testing"

func TestAuditDurableClaimsFlagsMissingMovement(t *testing.T) {
	issues := AuditDurableClaims("You arrive at the Lower Tier Corridor.", nil, []string{"move_player", "create_location"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimMovement {
		t.Fatalf("expected movement issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsAcceptsCreateLocationMovePlayerHere(t *testing.T) {
	applied := []AppliedToolCall{{Tool: "create_location", Result: []byte(`{"move_player_here":true,"location_id":"00000000-0000-0000-0000-000000000001"}`)}}
	issues := AuditDurableClaims("You arrive at the Lower Tier Corridor.", applied, []string{"move_player", "create_location"})
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsMissingQuest(t *testing.T) {
	issues := AuditDurableClaims("A new quest is added to your journal: Find the Core.", nil, []string{"create_quest"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimQuest {
		t.Fatalf("expected quest issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsMissingFact(t *testing.T) {
	issues := AuditDurableClaims("You now know the Core Archive is unstable.", nil, []string{"establish_fact"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimFact {
		t.Fatalf("expected fact issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsSuppressesMatchingToolForEachKind(t *testing.T) {
	issues := AuditDurableClaims(
		"You arrive at the Lower Tier Corridor. A new quest is added to your journal. You now know the Core Archive is unstable.",
		[]AppliedToolCall{
			{Tool: "move_player"},
			{Tool: "create_quest"},
			{Tool: "establish_fact"},
		},
		[]string{"move_player", "create_quest", "establish_fact"},
	)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %#v", issues)
	}
}
