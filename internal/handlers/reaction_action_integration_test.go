//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
)

func pendingReaction(t *testing.T) *spatialFixture {
	t.Helper()
	f := newSpatialFixture(t)
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"none","tactic":"withdraw"`)
	f.confirm(f.prepare(combat.EncounterStart))
	f.confirm(f.prepare(combat.EncounterEndTurn))
	saved, err := game.ReadEncounter(context.Background(), testPool, f.campaign)
	f.must(err)
	if saved.Reaction.Pending == nil || !saved.Reaction.Available || saved.Active().Player || *saved.Space.DistanceFeet != 5 {
		t.Fatalf("trigger not saved before withdrawal: %+v", saved)
	}
	if f.dice.calls != 3 {
		t.Fatalf("reaction silently rolled: %d", f.dice.calls)
	}
	if f.prepare(srd521.MartialAttack).Available || f.prepare(combat.EncounterEndTurn).Available {
		t.Fatal("turn bypassed pending choice")
	}
	return f
}

func TestIntegrationReactionConfirmOnceAndReconnect(t *testing.T) {
	f := pendingReaction(t)
	p := f.prepare(combat.EncounterReact)
	if p.Encounter.NPCTactic != "" {
		t.Fatal("unplayed NPC tactic leaked")
	}
	for _, actor := range p.Encounter.Actors {
		if !actor.Player && (actor.HP != 0 || actor.MaxHP != 0) {
			t.Fatal("private HP leaked")
		}
	}
	if !p.Available || len(p.Costs) != 1 || p.Costs[0].ResourceID != "reaction" || p.Costs[0].Amount != 1 {
		t.Fatalf("reaction preview: %+v", p)
	}
	calls := f.dice.calls
	request, result := f.confirm(p)
	var payload struct {
		Receipt      srd521.AttackReceipt `json:"receipt"`
		ReactionCost int                  `json:"reaction_cost"`
	}
	f.must(json.Unmarshal(result.AppliedToolCalls[0].Result, &payload))
	if payload.ReactionCost != 1 || payload.Receipt.Damage != 4 || f.dice.calls != calls+2 {
		t.Fatalf("reaction receipt=%+v calls=%d", payload, f.dice.calls)
	}
	saved, err := game.ReadEncounter(context.Background(), testPool, f.campaign)
	f.must(err)
	if saved.Reaction.Pending != nil || !saved.Active().Player || *saved.Space.DistanceFeet != 10 || saved.Round != 2 || !saved.Reaction.Available {
		t.Fatalf("withdrawal/recovery=%+v", saved)
	}
	// A fresh engine models reconnect after a committed response with a lost reply.
	calls = f.dice.calls
	replay, err := f.engine().ConfirmAction(context.Background(), f.campaign, request)
	f.must(err)
	if f.dice.calls != calls || !sameTurnResult(result, replay) {
		t.Fatal("reconnect rerolled or moved twice")
	}
	request.ActionID = uuid.New()
	if _, err := f.engine().ConfirmAction(context.Background(), f.campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale reaction response: %v", err)
	}
	if f.prepare(combat.EncounterReact).Available {
		t.Fatal("reaction without pending trigger")
	}
}

func TestIntegrationReactionDeclineAndStaleRuling(t *testing.T) {
	f := pendingReaction(t)
	p := f.prepare(combat.EncounterReact)
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"total","tactic":"withdraw"`)
	calls := f.dice.calls
	_, err := f.engine().ConfirmAction(context.Background(), f.campaign, engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent, TargetID: f.target})
	if !errors.Is(err, engine.ErrActionRefresh) || f.dice.calls != calls {
		t.Fatalf("stale reaction rolled: %v", err)
	}
	if f.prepare(combat.EncounterReact).Available {
		t.Fatal("changed trigger still eligible")
	}
	request, result := f.confirm(f.prepare(combat.EncounterDecline))
	if f.dice.calls != calls {
		t.Fatal("decline rolled")
	}
	saved, err := game.ReadEncounter(context.Background(), testPool, f.campaign)
	f.must(err)
	if saved.Reaction.Pending != nil || !saved.Active().Player || *saved.Space.DistanceFeet != 5 {
		t.Fatal("decline invented path through total cover")
	}
	replay, err := f.engine().ConfirmAction(context.Background(), f.campaign, request)
	f.must(err)
	if f.dice.calls != calls || !sameTurnResult(result, replay) {
		t.Fatal("decline replay repeated")
	}
}

func TestIntegrationReactionDeclineAndEquipmentChange(t *testing.T) {
	f := pendingReaction(t)
	p := f.prepare(combat.EncounterReact)
	_, err := testPool.Exec(context.Background(), `UPDATE items SET equipped=false WHERE id=$1`, f.weapon)
	f.must(err)
	_, err = f.engine().ConfirmAction(context.Background(), f.campaign, engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent, TargetID: f.target})
	if !errors.Is(err, engine.ErrActionRefresh) || f.dice.calls != 3 {
		t.Fatalf("unequipped reaction: %v", err)
	}
	f.confirm(f.prepare(combat.EncounterDecline))
	saved, err := game.ReadEncounter(context.Background(), testPool, f.campaign)
	f.must(err)
	if *saved.Space.DistanceFeet != 10 || saved.Reaction.Pending != nil || f.dice.calls != 3 {
		t.Fatal("decline did not finish the step without dice")
	}
}

func TestIntegrationReactionNeedsEligibleTrigger(t *testing.T) {
	f := newSpatialFixture(t)
	if f.prepare(combat.EncounterReact).Available || f.prepare(combat.EncounterDecline).Available {
		t.Fatal("reaction available without encounter")
	}
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"none","tactic":"withdraw"`)
	f.confirm(f.prepare(combat.EncounterStart))
	_, err := testPool.Exec(context.Background(), `UPDATE items SET equipped=false WHERE id=$1`, f.weapon)
	f.must(err)
	f.confirm(f.prepare(combat.EncounterEndTurn))
	saved, err := game.ReadEncounter(context.Background(), testPool, f.campaign)
	f.must(err)
	if saved.Reaction.Pending != nil || !saved.Active().Player || *saved.Space.DistanceFeet != 10 || f.dice.calls != 3 {
		t.Fatal("offered unsupported unarmed reaction")
	}
}

func TestIntegrationReactionLeadingNPCAndDefeat(t *testing.T) {
	f := newSpatialFixture(t)
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"none","tactic":"withdraw"`)
	f.dice.d20 = []int{0, 19, 18}
	_, err := testPool.Exec(context.Background(), `UPDATE npcs SET hp=4 WHERE id=$1`, f.target)
	f.must(err)
	f.confirm(f.prepare(combat.EncounterStart))
	p := f.prepare(combat.EncounterReact)
	if p.Encounter.Round != 1 || p.Encounter.Active().Player || p.Encounter.Reaction.Pending == nil {
		t.Fatal("leading NPC trigger skipped")
	}
	f.confirm(p)
	saved, err := game.ReadEncounter(context.Background(), testPool, f.campaign)
	f.must(err)
	if saved.Status != "completed" || saved.Reaction.Pending != nil || *saved.Space.DistanceFeet != 5 {
		t.Fatal("defeated foe completed suspended movement")
	}
}
