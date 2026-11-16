package combat

import (
	"encoding/json"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestActionSurgeBudget(t *testing.T) {
	for _, first := range []bool{true, false} {
		t.Run(map[bool]string{true: "before ordinary", false: "after ordinary"}[first], func(t *testing.T) {
			actor := uuid.New()
			e := &Encounter{Status: "active", Round: 1, Actors: []EncounterActor{{ID: actor, Player: true, HP: 12}, {ID: uuid.New(), HP: 100}}, Reaction: &ReactionState{Available: false}, Budget: TurnBudget{Action: true, BonusAction: false, Movement: 15, SpellSlotSpent: true}}
			spend := func(kind string) {
				t.Helper()
				if err := e.Spend(actor, kind); err != nil {
					t.Fatal(err)
				}
			}
			if !first {
				spend(srd521.MartialAttack)
			}
			spend(srd521.ActionSurge)
			if e.Budget.BonusAction || e.Budget.Movement != 15 || !e.Budget.SpellSlotSpent || e.Reaction.Available {
				t.Fatal("unrelated budget reset")
			}
			before := *e
			for _, kind := range []string{srd521.ActionSurge, "magic", "unsupported"} {
				if err := e.Spend(actor, kind); err == nil || !reflect.DeepEqual(before, *e) {
					t.Fatalf("invalid %s mutated", kind)
				}
			}
			spend(srd521.MartialAttack)
			if first {
				spend(srd521.MartialAttack)
			}
			if err := e.Spend(actor, srd521.MartialAttack); err == nil {
				t.Fatal("third action")
			}
			if err := e.Spend(actor, srd521.SpellCast); err == nil {
				t.Fatal("surge permits Magic")
			}
			e.Advance()
			if e.Budget.SurgeAction || e.Budget.SurgeUsed {
				t.Fatal("surge carried across turn")
			}
		})
	}
}

func TestActionSurgeMagicAndLegacy(t *testing.T) {
	actor := uuid.New()
	var b TurnBudget
	if err := json.Unmarshal([]byte(`{"action":true,"bonus_action":true,"movement":30}`), &b); err != nil {
		t.Fatal(err)
	}
	e := &Encounter{Status: "active", Round: 1, Actors: []EncounterActor{{ID: actor, HP: 12}}, Budget: b}
	if e.Budget.SurgeAction || e.Budget.SurgeUsed {
		t.Fatal("legacy grants surge")
	}
	if err := e.Spend(actor, srd521.ActionSurge); err != nil {
		t.Fatal(err)
	}
	if err := e.SpendSpell(actor, false); err != nil {
		t.Fatal(err)
	}
	before := e.Budget
	if err := e.Spend(actor, srd521.SpellCast); err == nil || before != e.Budget {
		t.Fatal("Magic consumed surge")
	}
	if err := e.Spend(actor, srd521.MartialAttack); err != nil {
		t.Fatal(err)
	}
}
