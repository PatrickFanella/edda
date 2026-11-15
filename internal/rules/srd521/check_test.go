package srd521

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type sequenceDice struct {
	values []int
	calls  int
}

func (d *sequenceDice) Intn(n int) int {
	if n != 20 || d.calls >= len(d.values) {
		panic("unexpected dice request")
	}
	v := d.values[d.calls]
	d.calls++
	return v
}

func checkSheet(t *testing.T, expertise ...string) *CharacterSheet {
	t.Helper()
	stats, err := os.ReadFile("testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct{ SRD521 Build }
	if err := json.Unmarshal(stats, &data); err != nil {
		t.Fatal(err)
	}
	data.SRD521.Expertise = expertise
	data.SRD521.Abilities["charisma"] = 9 // Odd negative scores must round down.
	stats, err = json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	sheet, err := Project(stats, 1, 9, 12, 125)
	if err != nil {
		t.Fatal(err)
	}
	return sheet
}

func TestResolveCheck(t *testing.T) {
	for _, tc := range []struct {
		name             string
		check            Check
		rolls            []int
		expertise        []string
		dc, total, bonus int
		mode             string
		kept             int
		success          bool
	}{
		{"ordinary proficient skill", Check{Kind: AbilityCheck, Ability: "strength", Skill: "athletics"}, []int{9}, nil, 15, 15, 5, "ordinary", 0, true},
		{"ability only excludes skill proficiency", Check{Kind: AbilityCheck, Ability: "strength"}, []int{9}, nil, 15, 13, 3, "ordinary", 0, false},
		{"untrained skill", Check{Kind: AbilityCheck, Ability: "intelligence", Skill: "arcana"}, []int{9}, nil, 10, 10, 0, "ordinary", 0, true},
		{"negative modifier rounds down", Check{Kind: AbilityCheck, Ability: "charisma"}, []int{9}, nil, 10, 9, -1, "ordinary", 0, false},
		{"advantage keeps second", Check{Kind: AbilityCheck, Ability: "strength", Skill: "athletics", Advantage: true}, []int{2, 14}, nil, 15, 20, 5, "advantage", 1, true},
		{"advantage keeps first", Check{Kind: AbilityCheck, Ability: "strength", Advantage: true}, []int{14, 2}, nil, 15, 18, 3, "advantage", 0, true},
		{"disadvantage keeps second", Check{Kind: SavingThrow, Ability: "constitution", Disadvantage: true}, []int{16, 3}, nil, 12, 8, 4, "disadvantage", 1, false},
		{"disadvantage keeps first", Check{Kind: SavingThrow, Ability: "constitution", Disadvantage: true}, []int{3, 16}, nil, 12, 8, 4, "disadvantage", 0, false},
		{"advantage tie keeps one", Check{Kind: SavingThrow, Ability: "constitution", Advantage: true}, []int{9, 9}, nil, 12, 14, 4, "advantage", 0, true},
		{"advantage cancels disadvantage", Check{Kind: AbilityCheck, Ability: "strength", Advantage: true, Disadvantage: true}, []int{9}, nil, 13, 13, 3, "ordinary", 0, true},
		{"expertise doubles proficiency once", Check{Kind: AbilityCheck, Ability: "wisdom", Skill: "perception"}, []int{9}, []string{"perception"}, 15, 15, 5, "ordinary", 0, true},
		{"expertise does not affect saves", Check{Kind: SavingThrow, Ability: "wisdom"}, []int{9}, []string{"perception"}, 15, 11, 1, "ordinary", 0, false},
		{"expertise does not affect another skill", Check{Kind: AbilityCheck, Ability: "wisdom", Skill: "insight"}, []int{9}, []string{"perception"}, 15, 11, 1, "ordinary", 0, false},
		{"natural 20 can fail", Check{Kind: AbilityCheck, Ability: "charisma"}, []int{19}, nil, 30, 19, -1, "ordinary", 0, false},
		{"natural 1 can succeed", Check{Kind: SavingThrow, Ability: "strength"}, []int{0}, nil, 5, 6, 5, "ordinary", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dice := &sequenceDice{values: tc.rolls}
			sheet := checkSheet(t, tc.expertise...)
			got, err := ResolveCheck(sheet, tc.check, Difficulty{Value: tc.dc, PlayerKnown: true, OutcomeKnown: true}, dice)
			if err != nil {
				t.Fatal(err)
			}
			r := got.PlayerReceipt()
			if got.Success() != tc.success || r.Total != tc.total || r.Modifier.Value != tc.bonus || r.Mode != tc.mode || r.DC == nil || *r.DC != tc.dc || len(r.Dice) != len(tc.rolls) || dice.calls != len(tc.rolls) {
				t.Fatalf("unexpected resolution: %+v, success=%v", r, got.Success())
			}
			wantOutcome := "failure"
			if tc.success {
				wantOutcome = "success"
			}
			if r.Outcome != wantOutcome {
				t.Fatalf("outcome = %s", r.Outcome)
			}
			for i, die := range r.Dice {
				if die.Value != tc.rolls[i]+1 || die.Sides != 20 || die.Kept != (i == tc.kept) {
					t.Fatalf("die %d = %+v", i, die)
				}
			}
			for _, id := range r.SourceIDs {
				if !contentID.MatchString(id) {
					t.Fatalf("invalid source ID: %s", id)
				}
			}
			if len(r.Modifier.Derivation) == 0 {
				t.Fatal("missing modifier terms")
			}
		})
	}
}

func TestCheckReceiptDisclosure(t *testing.T) {
	for _, dcKnown := range []bool{false, true} {
		for _, outcomeKnown := range []bool{false, true} {
			sheet := checkSheet(t)
			got, err := ResolveCheck(sheet, Check{Kind: SavingThrow, Ability: "strength"}, Difficulty{Value: 17, PlayerKnown: dcKnown, OutcomeKnown: outcomeKnown}, &sequenceDice{values: []int{9}})
			if err != nil {
				t.Fatal(err)
			}
			r := got.PlayerReceipt()
			if (r.DC != nil) != dcKnown || (r.Outcome != "unknown") != outcomeKnown || got.Success() {
				t.Fatalf("disclosure flags not respected: %+v", r)
			}
			encoded, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			if _, ok := fields["dc"]; ok != dcKnown {
				t.Fatalf("secret DC in transport: %s", encoded)
			}
			for _, key := range []string{"difficulty", "margin", "success", "facts", "reason", "narrative"} {
				if _, ok := fields[key]; ok {
					t.Fatalf("unexpected private field %s", key)
				}
			}
			private, err := json.Marshal(got)
			if err != nil || string(private) != string(encoded) {
				t.Fatalf("resolution did not default to its player-safe receipt: %s, %v", private, err)
			}
			// Receipt mutations cannot alter a later read or the original sheet.
			original := got.PlayerReceipt()
			r.Dice[0].Value = 99
			r.Modifier.Derivation[0].Value = 99
			r.SourceIDs[0] = "hidden fact"
			if r.DC != nil {
				*r.DC = 99
			}
			if !reflect.DeepEqual(original, got.PlayerReceipt()) {
				t.Fatal("receipt mutated the resolution")
			}
			if sheet.Saves["strength"].Derivation[0].Value != 3 {
				t.Fatal("receipt mutated the character sheet")
			}
			sheet.Saves["strength"].Derivation[0].Value = 99
			if !reflect.DeepEqual(original, got.PlayerReceipt()) {
				t.Fatal("character sheet mutated the resolved receipt")
			}
		}
	}
}

func TestCheckRejectsUnsupportedInputBeforeRolling(t *testing.T) {
	for _, check := range []Check{
		{Kind: "attack", Ability: "strength"},
		{Kind: SavingThrow, Ability: "death"},
		{Kind: SavingThrow, Ability: "wisdom", Skill: "perception"},
		{Kind: AbilityCheck, Ability: "wisdom", Skill: "unknown hidden entity"},
		{Kind: AbilityCheck, Ability: "strength", Skill: "perception"},
		{Kind: AbilityCheck, Ability: "luck"},
	} {
		dice := &sequenceDice{}
		if _, err := ResolveCheck(checkSheet(t), check, Difficulty{Value: 15}, dice); err == nil || dice.calls != 0 {
			t.Fatalf("invalid check rolled: %+v, %v", check, err)
		}
	}
	check := Check{Kind: AbilityCheck, Ability: "strength"}
	for _, value := range []int{-1, 20} {
		if _, err := ResolveCheck(checkSheet(t), check, Difficulty{Value: 15}, &sequenceDice{values: []int{value}}); err == nil {
			t.Fatal("accepted invalid die")
		}
	}
	if _, err := ResolveCheck(nil, check, Difficulty{}, &sequenceDice{}); err == nil {
		t.Fatal("accepted missing sheet")
	}
	if _, err := ResolveCheck(checkSheet(t), check, Difficulty{}, nil); err == nil {
		t.Fatal("accepted missing dice")
	}
	if _, err := ResolveCheck(checkSheet(t), check, Difficulty{Value: -1}, &sequenceDice{}); err == nil {
		t.Fatal("accepted invalid difficulty")
	}
}

func TestExpertiseRequiresRecordedProficiency(t *testing.T) {
	for _, expertise := range []string{`["arcana"]`, `["perception","perception"]`, `["unknown"]`} {
		stats := []byte(`{"srd521":{"class":"fighter","fighting_style":"defense","abilities":{"strength":16,"dexterity":12,"constitution":14,"intelligence":10,"wisdom":13,"charisma":8},"skills":["perception"],"expertise":` + expertise + `,"armor":"chain-mail","second_wind":1,"hit_dice":1}}`)
		if _, err := Project(stats, 1, 9, 12, 125); err == nil {
			t.Fatalf("accepted invalid or duplicate expertise: %s", expertise)
		}
	}
}
