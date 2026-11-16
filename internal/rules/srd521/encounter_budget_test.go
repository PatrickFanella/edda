package srd521

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

func encounterTestGroup(cr string, xp, count int64) EncounterMonsterGroup {
	return EncounterMonsterGroup{CR: cr, XP: &xp, Count: count}
}

func encounterFixture(t *testing.T) struct {
	SHA        string     `json:"source_sha256"`
	BudgetPage int        `json:"budget_page"`
	XPPage     int        `json:"xp_page"`
	Budgets    [][3]int64 `json:"budgets"`
	XP         []struct {
		CR string `json:"cr"`
		XP int64  `json:"xp"`
	} `json:"xp"`
} {
	t.Helper()
	var f struct {
		SHA        string     `json:"source_sha256"`
		BudgetPage int        `json:"budget_page"`
		XPPage     int        `json:"xp_page"`
		Budgets    [][3]int64 `json:"budgets"`
		XP         []struct {
			CR string `json:"cr"`
			XP int64  `json:"xp"`
		} `json:"xp"`
	}
	b, err := os.ReadFile("testdata/encounter_budgets.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestEncounterSourceTables(t *testing.T) {
	f := encounterFixture(t)
	if len(f.Budgets) != 20 || len(f.XP) != 35 || f.SHA != artifactSHA256 || f.BudgetPage != 202 || f.XPPage != 256 {
		t.Fatal("invalid source fixture")
	}
	for i, row := range f.Budgets {
		t.Run(fmt.Sprintf("level-%d", i+1), func(t *testing.T) {
			r, err := CalculateEncounterBudget([]EncounterPartyGroup{{i + 1, 1}}, nil)
			if err != nil || r.Budgets != (EncounterBudgets{row[0], row[1], row[2]}) {
				t.Fatalf("%+v %v", r, err)
			}
			if err := validateSource(r.Source); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(r.SourcePages, []int{202, 203, 256}) || r.Party[0].SourceID != EncounterBudgetSourceID || r.Source.Changes == "" {
				t.Fatal("missing provenance")
			}
		})
	}
	for _, row := range f.XP {
		t.Run(fmt.Sprintf("CR-%s-XP-%d", row.CR, row.XP), func(t *testing.T) {
			r, err := CalculateEncounterBudget([]EncounterPartyGroup{{20, 1}}, []EncounterMonsterGroup{encounterTestGroup(row.CR, row.XP, 2)})
			if err != nil || r.TotalXP != 2*row.XP || r.Monsters[0].SourceID != EncounterXPSourceID {
				t.Fatalf("%+v %v", r, err)
			}
			if encounterValidXP(row.CR, row.XP+1) {
				t.Fatal("accepted wrong XP")
			}
		})
	}
}

// Legal XP values have granularity 5, so ±5 is the nearest reachable boundary.
func encounterGroupsForXP(total int64) []EncounterMonsterGroup {
	var groups []EncounterMonsterGroup
	if total%10 != 0 {
		groups = append(groups, encounterTestGroup("1/8", 25, 1))
		total -= 25
	}
	if total > 0 {
		groups = append(groups, encounterTestGroup("0", 10, total/10))
	}
	return groups
}

func TestEncounterDifficultyBoundaries(t *testing.T) {
	f := encounterFixture(t)
	for i, row := range f.Budgets {
		for ceiling, budget := range row {
			for _, delta := range []int64{-5, 0, 5} {
				t.Run(fmt.Sprintf("level-%d-ceiling-%d-delta-%d", i+1, ceiling, delta), func(t *testing.T) {
					total := budget + delta
					want := []string{"low", "moderate", "high"}[ceiling]
					if delta > 0 {
						want = []string{"moderate", "high", "above-high"}[ceiling]
					}
					r, err := CalculateEncounterBudget([]EncounterPartyGroup{{i + 1, 1}}, encounterGroupsForXP(total))
					if err != nil || r.TotalXP != total || r.Difficulty != want {
						t.Fatalf("total %d: got %d %s, want %s, err %v", total, r.TotalXP, r.Difficulty, want, err)
					}
				})
			}
		}
	}
}

func TestEncounterExamplesAndMixedParty(t *testing.T) {
	tests := []struct {
		name       string
		party      []EncounterPartyGroup
		monsters   []EncounterMonsterGroup
		budgets    EncounterBudgets
		total      int64
		difficulty string
	}{
		{"source example 1", []EncounterPartyGroup{{1, 4}}, []EncounterMonsterGroup{encounterTestGroup("1/8", 25, 6)}, EncounterBudgets{200, 300, 400}, 150, "low"},
		{"source example 2", []EncounterPartyGroup{{3, 5}}, []EncounterMonsterGroup{encounterTestGroup("2", 450, 2), encounterTestGroup("1/8", 25, 9)}, EncounterBudgets{750, 1125, 2000}, 1125, "moderate"},
		{"source example 3", []EncounterPartyGroup{{15, 6}}, []EncounterMonsterGroup{encounterTestGroup("17", 18000, 2), encounterTestGroup("9", 5000, 2)}, EncounterBudgets{19800, 32400, 46800}, 46000, "high"},
		{"mixed", []EncounterPartyGroup{{1, 2}, {3, 1}}, []EncounterMonsterGroup{encounterTestGroup("1/4", 50, 2), encounterTestGroup("1/2", 100, 2)}, EncounterBudgets{250, 375, 600}, 300, "moderate"},
		{"empty selection", []EncounterPartyGroup{{1, 1}}, nil, EncounterBudgets{50, 75, 100}, 0, "low"},
		{"explicit zero", []EncounterPartyGroup{{1, 1}}, []EncounterMonsterGroup{encounterTestGroup("0", 0, math.MaxInt64)}, EncounterBudgets{50, 75, 100}, 0, "low"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := CalculateEncounterBudget(tt.party, tt.monsters)
			if err != nil || r.Budgets != tt.budgets || r.TotalXP != tt.total || r.Difficulty != tt.difficulty {
				t.Fatalf("%+v %v", r, err)
			}
			var total int64
			for _, m := range r.Monsters {
				if m.Subtotal != m.XP*m.Count {
					t.Fatal("wrong derivation")
				}
				total += m.Subtotal
			}
			if total != r.TotalXP {
				t.Fatal("wrong sum")
			}
		})
	}
}

func TestEncounterRejectsInvalidInputs(t *testing.T) {
	party := []EncounterPartyGroup{{1, 1}}
	tests := []struct {
		name     string
		party    []EncounterPartyGroup
		monsters []EncounterMonsterGroup
	}{
		{"nil party", nil, nil}, {"empty party", []EncounterPartyGroup{}, nil},
		{"missing level", []EncounterPartyGroup{{Count: 1}}, nil}, {"negative level", []EncounterPartyGroup{{-1, 1}}, nil}, {"unsupported level", []EncounterPartyGroup{{21, 1}}, nil},
		{"missing count", []EncounterPartyGroup{{Level: 1}}, nil}, {"negative count", []EncounterPartyGroup{{1, -1}}, nil},
		{"party multiplication overflow", []EncounterPartyGroup{{20, math.MaxInt64}}, nil},
		{"party accumulation overflow", []EncounterPartyGroup{{1, math.MaxInt64 / 100}, {1, 1}}, nil},
		{"XP missing", party, []EncounterMonsterGroup{{CR: "0", Count: 1}}},
		{"monster count missing", party, []EncounterMonsterGroup{encounterTestGroup("0", 0, 0)}},
		{"monster count negative", party, []EncounterMonsterGroup{encounterTestGroup("0", 0, -1)}},
		{"negative XP", party, []EncounterMonsterGroup{encounterTestGroup("0", -1, 1)}},
		{"CR0 wrong XP", party, []EncounterMonsterGroup{encounterTestGroup("0", 5, 1)}},
		{"XP mismatch", party, []EncounterMonsterGroup{encounterTestGroup("2", 400, 1)}},
		{"monster multiplication overflow", party, []EncounterMonsterGroup{encounterTestGroup("30", 155000, math.MaxInt64)}},
		{"monster accumulation overflow", party, []EncounterMonsterGroup{encounterTestGroup("0", 10, math.MaxInt64/10), encounterTestGroup("0", 10, 1)}},
	}
	for _, cr := range []string{"", "31", "-1", "0.5", "1/3", "01", "+1", " 1", "1.0"} {
		tests = append(tests, struct {
			name     string
			party    []EncounterPartyGroup
			monsters []EncounterMonsterGroup
		}{"invalid CR " + cr, party, []EncounterMonsterGroup{encounterTestGroup(cr, 200, 1)}})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := CalculateEncounterBudget(tt.party, tt.monsters)
			if err == nil || !reflect.DeepEqual(r, EncounterBudgetResult{}) {
				t.Fatalf("must fail closed: %+v %v", r, err)
			}
		})
	}
}

func TestEncounterNumericPresence(t *testing.T) {
	for _, raw := range []string{`{"cr":"0","count":1}`, `{"cr":"0","xp":null,"count":1}`, `{"cr":"0","xp":0}`, `{"cr":"0","xp":0,"count":null}`} {
		t.Run(raw, func(t *testing.T) {
			var m EncounterMonsterGroup
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				t.Fatal(err)
			}
			if _, err := CalculateEncounterBudget([]EncounterPartyGroup{{1, 1}}, []EncounterMonsterGroup{m}); err == nil {
				t.Fatal("missing numeric data accepted")
			}
		})
	}
	for _, raw := range []string{`{"xp":1.5}`, `{"xp":9223372036854775808}`, `{"count":1.5}`, `{"count":9223372036854775808}`, `{"xp":"0"}`} {
		var m EncounterMonsterGroup
		if json.Unmarshal([]byte(raw), &m) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestEncounterTypedMonsterAndIsolation(t *testing.T) {
	c, err := LoadMonsters()
	if err != nil {
		t.Fatal(err)
	}
	m := c.Monsters[0]
	before, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if m.Available() {
		t.Fatal("fixture unexpectedly playable")
	}
	group, err := EncounterMonsterFromRecord(m, 4)
	if err != nil {
		t.Fatal(err)
	}
	party := []EncounterPartyGroup{{1, 4}}
	monsters := []EncounterMonsterGroup{group}
	r, err := CalculateEncounterBudget(party, monsters)
	if err != nil {
		t.Fatal(err)
	}
	if r.TotalXP != 100 || r.Monsters[0].MonsterID != m.ID || r.Party[0].Subtotal != (EncounterBudgets{200, 300, 400}) {
		t.Fatalf("wrong typed derivation: %+v", r)
	}
	r.Party[0].Count = 99
	r.Monsters[0].XP = 999
	r.SourcePages[0] = 999
	if party[0].Count != 4 || *group.XP != 25 {
		t.Fatal("result aliases input")
	}
	*group.XP = 50
	if m.XP != 25 {
		t.Fatal("input aliases catalogue")
	}
	after, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) || m.Available() {
		t.Fatal("catalogue changed")
	}
	for _, count := range []int64{0, -1} {
		if _, err := EncounterMonsterFromRecord(m, count); err == nil {
			t.Fatal("invalid count accepted")
		}
	}
	m.XP = 26
	if _, err := EncounterMonsterFromRecord(m, 1); err == nil {
		t.Fatal("mismatched record XP accepted")
	}
	m.ID = ""
	if _, err := EncounterMonsterFromRecord(m, 1); err == nil || !strings.Contains(err.Error(), "encounter monster") {
		t.Fatal("invalid typed record accepted")
	}
}
