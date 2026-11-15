package srd521

import (
	"encoding/json"
	"testing"
)

func TestFighterAdvancement(t *testing.T) {
	for _, choice := range []string{"fixed", "roll"} {
		t.Run(choice, func(t *testing.T) {
			creation, err := SupportedBuildForConcept("Human Fighter", "Soldier")
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
			sheet, err := Project(stats, 1, 9, 12, 300)
			if err != nil {
				t.Fatal(err)
			}
			dice := &lifecycleDice{value: 4}
			r, err := AdvanceFighter(&envelope.Build, sheet, choice, dice)
			if err != nil {
				t.Fatal(err)
			}
			expected, calls := 20, 0
			if choice == "roll" {
				expected, calls = 19, 1
			}
			data, err := json.Marshal(envelope)
			if err != nil {
				t.Fatal(err)
			}
			after, err := Project(data, 2, 9, expected, 300)
			if err != nil {
				t.Fatal(err)
			}
			if r.MaxHPAfter != expected || after.MaxHP.Value != expected || after.HP.Value != 9 || after.Resources[1].Maximum != 2 || after.Resources[1].Current != 2 || after.Progression.NextLevelXP != 900 || dice.calls != calls {
				t.Fatalf("receipt=%+v sheet=%+v calls=%d", r, after, dice.calls)
			}
			if _, err := AdvanceFighter(&envelope.Build, after, choice, dice); err == nil {
				t.Fatal("advanced twice")
			}
			if _, err := Project(data, 3, 9, expected, 900); err == nil {
				t.Fatal("unsupported level accepted")
			}
		})
	}
}

func TestFighterAdvancementRejectsMissingChoicesAndXP(t *testing.T) {
	for _, tc := range []struct {
		choice string
		xp     int
	}{{"", 300}, {"invented", 300}, {"fixed", 299}} {
		t.Run(tc.choice, func(t *testing.T) {
			creation, err := SupportedBuildForConcept("Human Fighter", "Soldier")
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
			sheet, err := Project(stats, 1, 12, 12, tc.xp)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := AdvanceFighter(&envelope.Build, sheet, tc.choice, &lifecycleDice{}); err == nil {
				t.Fatal("invalid transition accepted")
			}
		})
	}
}
