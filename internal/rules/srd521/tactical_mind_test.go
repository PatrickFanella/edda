package srd521

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTacticalMindSourceRule(t *testing.T) {
	base := CheckReceipt{RulesVersion: "srd-5.2.1", Kind: AbilityCheck, Mode: "ordinary", Outcome: "failure", Total: 9, Dice: []CheckDie{{Sides: 20, Value: 4, Kept: true}}}
	for _, tc := range []struct {
		name          string
		die, dc, cost int
		outcome       string
	}{{"one-equals", 0, 10, 1, "success"}, {"ten-equals", 9, 19, 1, "success"}, {"ten-still-fails", 9, 20, 0, "failure"}} {
		t.Run(tc.name, func(t *testing.T) {
			d := &lifecycleDice{value: tc.die}
			r, err := ResolveTacticalMind(base, tc.dc, 2, "use", d)
			if err != nil {
				t.Fatal(err)
			}
			if d.calls != 1 || r.SecondWindAfter != 2-tc.cost || r.Outcome != tc.outcome || r.Base.Total != 9 || r.Base.Dice[0].Value != 4 {
				t.Fatalf("%+v calls %d", r, d.calls)
			}
			data, _ := json.Marshal(r)
			if strings.Contains(string(data), "difficulty") || strings.Contains(string(data), `"dc"`) {
				t.Fatal(string(data))
			}
		})
	}
	for _, kind := range []string{SavingThrow, MartialAttack} {
		b := base
		b.Kind = kind
		if _, err := ResolveTacticalMind(b, 20, 2, "use", nil); err == nil {
			t.Fatal(kind)
		}
	}
	for _, mode := range []string{"advantage", "disadvantage"} {
		b := base
		b.Mode = mode
		if _, err := ResolveTacticalMind(b, 20, 2, "decline", nil); err == nil {
			t.Fatal(mode)
		}
	}
	for _, value := range []int{-1, 10} {
		if _, err := ResolveTacticalMind(base, 20, 2, "use", &lifecycleDice{value: value}); err == nil {
			t.Fatal(value)
		}
	}
	if r, err := ResolveTacticalMind(base, 20, 2, "decline", nil); err != nil || r.Die != 0 || r.SecondWindAfter != 2 {
		t.Fatal(r, err)
	}
	if _, err := ResolveTacticalMind(base, 20, 0, "decline", nil); err == nil {
		t.Fatal("spent resource")
	}
}
