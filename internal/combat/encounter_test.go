package combat

import (
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
	"testing"
)

type fixedEncounterDice int

func (d fixedEncounterDice) Intn(int) int { return int(d) }
func TestEncounterEconomyAndInitiative(t *testing.T) {
	p, n := uuid.New(), uuid.New()
	e, err := NewEncounter(EncounterActor{ID: p, HP: 12, MaxHP: 12}, EncounterActor{ID: n, HP: 20, MaxHP: 20}, 0, fixedEncounterDice(9))
	if err != nil {
		t.Fatal(err)
	}
	if e.Active().ID != p || e.Round != 1 || e.Actors[0].Initiative != 10 {
		t.Fatalf("encounter=%+v", e)
	}
	if err := e.Spend(n, srd521.MartialAttack); err == nil {
		t.Fatal("out of turn accepted")
	}
	for _, kind := range []string{srd521.MartialAttack, srd521.ConsumableUse} {
		if err := e.Spend(p, kind); err != nil {
			t.Fatal(err)
		}
		if err := e.Spend(p, kind); err == nil {
			t.Fatalf("double %s accepted", kind)
		}
	}
	for i := 0; i < 6; i++ {
		if err := e.Spend(p, EncounterMove); err != nil {
			t.Fatal(err)
		}
	}
	if err := e.Spend(p, EncounterMove); err == nil {
		t.Fatal("movement exceeded")
	}
	e.Advance()
	if e.Active().ID != n || !e.Budget.Action || !e.Budget.BonusAction || e.Budget.Movement != 30 {
		t.Fatalf("NPC turn=%+v", e)
	}
	e.Advance()
	if e.Round != 2 || e.Active().ID != p {
		t.Fatalf("round=%+v", e)
	}
	if err := e.Spend(p, EncounterDash); err != nil || e.Budget.Movement != 60 || e.Budget.Action {
		t.Fatal("dash did not spend action/add movement")
	}
	e.SetHP(n, 0)
	if e.Status != "completed" {
		t.Fatal("defeat did not complete")
	}
	if err := e.Spend(p, srd521.MartialAttack); err == nil {
		t.Fatal("completed encounter accepted action")
	}
}

func TestReactionConsentAndRecovery(t *testing.T) {
	p, n := uuid.New(), uuid.New()
	e, err := NewEncounter(EncounterActor{ID: p, HP: 12}, EncounterActor{ID: n, HP: 20}, 0, fixedEncounterDice(9))
	if err != nil {
		t.Fatal(err)
	}
	distance, sight := 5, true
	e.Space = &Space{DistanceFeet: &distance, LineOfSight: &sight, Cover: "none", Revision: "scene"}
	e.Advance()
	if err := e.SpendReaction(); err == nil {
		t.Fatal("reaction without trigger")
	}
	e.Reaction.Pending = &ReactionTrigger{ID: uuid.New(), TargetID: n, SpaceRevision: "old"}
	if err := e.SpendReaction(); err == nil || !e.Reaction.Available {
		t.Fatal("stale trigger spent reaction")
	}
	e.Reaction.Pending.SpaceRevision = "scene"
	if err := e.SpendReaction(); err != nil {
		t.Fatal(err)
	}
	if e.Reaction.Available {
		t.Fatal("reaction not spent")
	}
	if err := e.SpendReaction(); err == nil {
		t.Fatal("double reaction")
	}
	e.Reaction.Pending = nil
	e.Advance()
	if !e.Reaction.Available {
		t.Fatal("reaction not restored at start of own turn")
	}
}
