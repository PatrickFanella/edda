package srd521

import (
	"encoding/json"
	"os"
	"testing"
)

func TestEquipmentRules(t *testing.T) {
	for _, tc := range []struct {
		item, kind string
		equipped   bool
		seconds    int
		armor      string
		shield     bool
	}{{"chain-mail", EquipmentRemove, true, 300, "none", true}, {"chain-mail", EquipmentEquip, false, 600, "chain-mail", true}, {"shield", EquipmentRemove, true, 6, "chain-mail", false}, {"shield", EquipmentEquip, false, 6, "chain-mail", true}} {
		t.Run(tc.item+tc.kind, func(t *testing.T) {
			data, err := os.ReadFile("testdata/fighter1.json")
			if err != nil {
				t.Fatal(err)
			}
			var e struct {
				Build Build `json:"srd521"`
			}
			if err = json.Unmarshal(data, &e); err != nil {
				t.Fatal(err)
			}
			if tc.item == "chain-mail" && !tc.equipped {
				e.Build.Armor = "none"
			}
			if tc.item == "shield" {
				e.Build.Shield = tc.equipped
			}
			r, err := ChangeEquipment(&e.Build, tc.kind, tc.item, tc.equipped, 28800)
			if err != nil {
				t.Fatal(err)
			}
			if r.TimeAfter-r.TimeBefore != tc.seconds || e.Build.Armor != tc.armor || e.Build.Shield != tc.shield {
				t.Fatalf("receipt=%+v build=%+v", r, e.Build)
			}
			stats, _ := json.Marshal(map[string]any{"srd521": e.Build})
			sheet, err := Project(stats, 1, 12, 12, 0)
			if err != nil {
				t.Fatal(err)
			}
			want := 11
			if tc.armor == "chain-mail" {
				want = 17
			}
			if tc.shield {
				want += 2
			}
			if sheet.Defenses["armor-class"].Value != want {
				t.Fatalf("AC=%+v", sheet.Defenses)
			}
		})
	}
	b := &Build{Class: "cleric", Armor: "none", Abilities: map[string]int{"strength": 16}, Casting: &CastingBuild{Concentration: &Concentration{ID: "effect", Spell: "shield-of-faith", StartedAt: 0, ExpiresAt: 600}}}
	if _, err := ChangeEquipment(b, EquipmentEquip, "chain-mail", false, 595); err == nil {
		t.Fatal("untrained heavy armor accepted")
	}
	r, err := ChangeEquipment(b, EquipmentEquip, "shield", false, 595)
	if err != nil || r.Expired == nil || b.Casting.Concentration != nil {
		t.Fatalf("expiry=%+v err=%v", r, err)
	}
	for _, item := range []string{"plate", "chain-mail+1", "cursed-shield"} {
		if _, err := ChangeEquipment(b, EquipmentRemove, item, true, 601); err == nil {
			t.Fatal("unsupported equipment accepted")
		}
	}
}
