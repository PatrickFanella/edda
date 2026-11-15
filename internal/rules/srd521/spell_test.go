package srd521

import (
	"encoding/json"
	"os"
	"testing"
)

type spellDice struct{ value, calls int }

func (d *spellDice) Intn(int) int { d.calls++; return d.value }
func TestCureWoundsFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/cleric1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name          string
		hp, die, want int
	}{{"healing", 1, 0, 6}, {"capped", 9, 7, 10}, {"full", 10, 0, 10}} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := Project(data, 1, tc.hp, 10, 0)
			if err != nil {
				t.Fatal(err)
			}
			d := &spellDice{value: tc.die}
			r, err := ResolveCureWounds(s, d)
			if err != nil {
				t.Fatal(err)
			}
			if r.HPAfter != tc.want || r.SlotsAfter != 1 || d.calls != 2 || r.Modifier.Value != 3 {
				t.Fatalf("receipt=%+v calls=%d", r, d.calls)
			}
			if s.Saves["wisdom"].Value != 5 || s.Saves["strength"].Value != 0 || s.Defenses["armor-class"].Value != 11 {
				t.Fatal("wrong cleric projection")
			}
		})
	}
	for _, field := range []string{"prepared", "slots", "verbal", "somatic"} {
		t.Run(field, func(t *testing.T) {
			var b map[string]map[string]any
			if err := json.Unmarshal(data, &b); err != nil {
				t.Fatal(err)
			}
			c := b["srd521"]["casting"].(map[string]any)
			if field == "slots" {
				c[field] = 0
			} else {
				c[field] = false
			}
			changed, err := json.Marshal(b)
			if err != nil {
				t.Fatal(err)
			}
			s, err := Project(changed, 1, 1, 10, 0)
			if err != nil {
				t.Fatal(err)
			}
			d := &spellDice{}
			if _, err := ResolveCureWounds(s, d); err == nil || d.calls != 0 {
				t.Fatal("unavailable cast rolled")
			}
		})
	}
	s, err := Project(data, 1, 1, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []int{-1, 8} {
		if _, err := ResolveCureWounds(s, &spellDice{value: v}); err == nil {
			t.Fatal("bad d8 accepted")
		}
	}
}
