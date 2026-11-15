package srd521

import (
	"encoding/json"
	"testing"
)

type lifecycleDice struct{ value, calls int }

func (d *lifecycleDice) Intn(int) int { d.calls++; return d.value }
func casterBuild(t *testing.T) (*Build, *CharacterSheet) {
	t.Helper()
	creation, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	stats, err := creation.Stats()
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Build Build `json:"srd521"`
	}
	if err := json.Unmarshal(stats, &envelope); err != nil {
		t.Fatal(err)
	}
	sheet, err := Project(stats, 1, 2, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	return &envelope.Build, sheet
}
func TestCasterLifecycleReplacementExpiryAndRest(t *testing.T) {
	b, sheet := casterBuild(t)
	dice := &lifecycleDice{value: 3}
	first, err := Lifecycle(b, sheet, SpellCast, "shield-of-faith", "first", 28800, dice)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Lifecycle(b, sheet, SpellCast, "shield-of-faith", "second", 28860, dice)
	if err != nil {
		t.Fatal(err)
	}
	if second.Before.ID != "first" || second.After.ID != "second" || first.After.ExpiresAt != 29400 || b.Casting.Slots != 0 || dice.calls != 0 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
	end, err := Lifecycle(b, sheet, EndConcentration, "", "end", 28860, dice)
	if err != nil || end.Before.ID != "second" || end.After != nil {
		t.Fatalf("end=%+v err=%v", end, err)
	}
	if _, err := Lifecycle(b, sheet, EndConcentration, "", "end-again", 28860, dice); err == nil {
		t.Fatal("ended twice")
	}
	b.Casting.Slots = 1
	if _, err := Lifecycle(b, sheet, SpellCast, "shield-of-faith", "third", 28860, dice); err != nil {
		t.Fatal(err)
	}
	for now := 28860; now < 29460; now += 60 {
		if _, err := Lifecycle(b, sheet, WaitMinute, "", "", now, dice); err != nil {
			t.Fatal(err)
		}
	}
	if b.Casting.Concentration != nil {
		t.Fatal("effect did not expire")
	}
	r, err := Lifecycle(b, sheet, ShortRest, "hit-die", "short", 29460, dice)
	if err != nil {
		t.Fatal(err)
	}
	if r.HPAfter != 8 || b.HitDice != 0 || b.Casting.Slots != 0 || r.TimeAfter != 33060 || r.HealingDie != 4 {
		t.Fatalf("short rest=%+v", r)
	}
	if _, err := Lifecycle(b, sheet, ShortRest, "hit-die", "again", 33060, dice); err == nil {
		t.Fatal("spent missing Hit Die")
	}
	r, err = Lifecycle(b, sheet, LongRest, "", "long", 33060, dice)
	if err != nil {
		t.Fatal(err)
	}
	if r.HPAfter != 10 || b.HitDice != 1 || b.Casting.Slots != 2 || *b.Casting.LastLongRest != 61860 {
		t.Fatalf("long rest=%+v", r)
	}
	if _, err := Lifecycle(b, sheet, LongRest, "", "early", 61860+16*3600-1, dice); err == nil {
		t.Fatal("rest started too soon")
	}
	if _, err := Lifecycle(b, sheet, LongRest, "", "allowed", 61860+16*3600, dice); err != nil {
		t.Fatal(err)
	}
}

func TestConcentrationDamageChecks(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		damage, hp, roll, dc, calls int
		kept                        bool
	}{
		{"no damage", 0, 2, 0, 0, 0, true}, {"success", 1, 1, 7, 10, 1, true},
		{"failure", 1, 1, 6, 10, 1, false}, {"half damage", 25, 1, 10, 12, 1, true},
		{"DC cap", 100, 1, 19, 30, 1, false}, {"unconscious", 2, 0, 19, 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, sheet := casterBuild(t)
			b.Casting.Concentration = &Concentration{ID: "effect", Spell: "shield-of-faith", StartedAt: 28800, ExpiresAt: 29400}
			dice := &lifecycleDice{value: tc.roll}
			r, err := ConcentrationDamage(b, sheet, tc.damage, tc.hp, 28806, dice)
			if err != nil {
				t.Fatal(err)
			}
			if dice.calls != tc.calls || (b.Casting.Concentration != nil) != tc.kept {
				t.Fatalf("receipt=%+v casting=%+v calls=%d", r, b.Casting, dice.calls)
			}
			if tc.calls > 0 && (r.Check == nil || *r.Check.DC != tc.dc || r.Check.Ability != "constitution") {
				t.Fatalf("receipt=%+v", r)
			}
		})
	}
}
