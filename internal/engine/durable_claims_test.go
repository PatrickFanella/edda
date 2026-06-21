package engine

import "testing"

func TestAuditDurableClaimsFlagsMissingMovement(t *testing.T) {
	issues := AuditDurableClaims("You arrive at the Lower Tier Corridor.", nil, []string{"move_player", "create_location"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimMovement {
		t.Fatalf("expected movement issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsReturnMovement(t *testing.T) {
	issues := AuditDurableClaims("You return to the Entry Chamber.", nil, []string{"move_player"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimMovement {
		t.Fatalf("expected movement issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsThresholdMovement(t *testing.T) {
	issues := AuditDurableClaims("You step through the threshold into the Entry Chamber.", nil, []string{"move_player"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimMovement {
		t.Fatalf("expected movement issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsRejectsReturnViewAgainAsMovement(t *testing.T) {
	issues := AuditDurableClaims("The drone comes into view again.", nil, []string{"move_player"})
	if len(issues) != 0 {
		t.Fatalf("expected no movement issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsRejectsInstructionStepthroughAsMovement(t *testing.T) {
	issues := AuditDurableClaims("You step through the instructions carefully.", nil, []string{"move_player"})
	if len(issues) != 0 {
		t.Fatalf("expected no movement issue, got %#v", issues)
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

func TestAuditDurableClaimsFlagsMissingInventoryCreated(t *testing.T) {
	issues := AuditDurableClaims("The spanner is now added to your inventory.", nil, []string{"add_item"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimInventoryCreated {
		t.Fatalf("expected inventory issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsAcceptsInventoryCreateTool(t *testing.T) {
	issues := AuditDurableClaims("The spanner is now added to your inventory.", []AppliedToolCall{{Tool: "add_item"}}, []string{"add_item"})
	if len(issues) != 0 {
		t.Fatalf("expected no inventory issues, got %#v", issues)
	}
}

func TestAuditDurableClaimsDoesNotSatisfyInventoryCreateWithRemoveTool(t *testing.T) {
	issues := AuditDurableClaims("The spanner is now added to your inventory.", []AppliedToolCall{{Tool: "remove_item"}}, []string{"remove_item"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimInventoryCreated {
		t.Fatalf("expected inventory create issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsMissingCombatStarted(t *testing.T) {
	issues := AuditDurableClaims("Combat begins as the drone locks onto you.", nil, []string{"initiate_combat"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimCombatStarted {
		t.Fatalf("expected combat issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsRejectsAttackNarrationAsCombatStart(t *testing.T) {
	issues := AuditDurableClaims("The raider attacks you.", nil, []string{"initiate_combat"})
	if len(issues) != 0 {
		t.Fatalf("expected no combat issues, got %#v", issues)
	}
}

func TestAuditDurableClaimsAcceptsCombatStartTool(t *testing.T) {
	issues := AuditDurableClaims("Combat begins as the drone locks onto you.", []AppliedToolCall{{Tool: "initiate_combat"}}, []string{"initiate_combat"})
	if len(issues) != 0 {
		t.Fatalf("expected no combat issues, got %#v", issues)
	}
}

func TestAuditDurableClaimsDoesNotSatisfyCombatStartWithResolveTool(t *testing.T) {
	issues := AuditDurableClaims("Combat begins.", []AppliedToolCall{{Tool: "resolve_combat"}}, []string{"resolve_combat"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimCombatStarted {
		t.Fatalf("expected combat started issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsMissingCombatResolved(t *testing.T) {
	issues := AuditDurableClaims("You defeat the drone and combat is over.", nil, []string{"resolve_combat"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimCombatResolved {
		t.Fatalf("expected combat resolved issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsExplicitInventoryRemoval(t *testing.T) {
	issues := AuditDurableClaims("The medkit is removed from your inventory.", nil, []string{"remove_item"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimInventoryRemoved {
		t.Fatalf("expected inventory removed issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsExplicitCombatResolution(t *testing.T) {
	issues := AuditDurableClaims("The combat ends and the drone collapses.", nil, []string{"resolve_combat"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimCombatResolved {
		t.Fatalf("expected combat resolved issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsRejectsNonInventoryPlacementAsInventoryCreate(t *testing.T) {
	for _, narrative := range []string{
		"You place your hand on the panel.",
		"You place the key in the lock.",
	} {
		if issues := AuditDurableClaims(narrative, nil, []string{"add_item"}); len(issues) != 0 {
			t.Fatalf("expected no inventory issues for %q, got %#v", narrative, issues)
		}
	}
}

func TestAuditDurableClaimsAcceptsCombatResolveTool(t *testing.T) {
	issues := AuditDurableClaims("You defeat the drone and combat is over.", []AppliedToolCall{{Tool: "resolve_combat"}}, []string{"resolve_combat"})
	if len(issues) != 0 {
		t.Fatalf("expected no combat issues, got %#v", issues)
	}
}

func TestAuditDurableClaimsRejectsFalsePositiveInventoryAndCombatClaims(t *testing.T) {
	for _, tc := range []struct {
		name      string
		narrative string
	}{
		{name: "stairs", narrative: "You take the stairs."},
		{name: "trail", narrative: "You pick up the trail."},
		{name: "sealed door", narrative: "A sealed door blocks the route."},
		{name: "hostile look", narrative: "The guard gives you a hostile look."},
		{name: "door view", narrative: "The entry hall stays in sight."},
		{name: "drop subject", narrative: "You drop the subject."},
		{name: "hand over railing", narrative: "You hand over the railing."},
		{name: "consume story", narrative: "You consume the story eagerly."},
		{name: "defeat lock", narrative: "You defeat the lock with patience."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if issues := AuditDurableClaims(tc.narrative, nil, []string{"add_item", "initiate_combat"}); len(issues) != 0 {
				t.Fatalf("expected no durable claim issues, got %#v", issues)
			}
		})
	}
}

func TestAuditDurableClaimsFlagsDroneLockCombatStart(t *testing.T) {
	issues := AuditDurableClaims("The drone locks onto you.", nil, []string{"initiate_combat"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimCombatStarted {
		t.Fatalf("expected combat start issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsRejectsOpticSensorsFlareAsCombatStart(t *testing.T) {
	issues := AuditDurableClaims("The drone's optic sensors flare red.", nil, []string{"initiate_combat"})
	if len(issues) != 0 {
		t.Fatalf("expected no combat issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsRejectsHostileDroneWithoutLock(t *testing.T) {
	issues := AuditDurableClaims("The drone hangs in the corridor, a hostile presence overhead.", nil, []string{"initiate_combat"})
	if len(issues) != 0 {
		t.Fatalf("expected no combat issue, got %#v", issues)
	}
}
