package srd521

import (
	"os"
	"testing"
)

type attackDice struct {
	values []int
	calls  int
}

func (d *attackDice) Intn(int) int { v := d.values[d.calls]; d.calls++; return v }
func TestLongswordFixture(t *testing.T) {
	stats, err := os.ReadFile("testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	sheet, err := Project(stats, 1, 12, 12, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		rolls   []int
		ac      int
		outcome string
		damage  int
	}{
		{"hit", []int{11, 4}, 10, "hit", 8},
		{"miss", []int{1}, 20, "miss", 0},
		{"natural one", []int{0}, 1, "miss", 0},
		{"critical ignores armor", []int{19, 4, 6}, 99, "critical-hit", 15},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dice := &attackDice{values: tc.rolls}
			r, err := ResolveLongsword(sheet, tc.ac, dice)
			if err != nil {
				t.Fatal(err)
			}
			if r.Outcome != tc.outcome || r.Damage != tc.damage || r.Modifier.Value != 5 || dice.calls != len(tc.rolls) {
				t.Fatalf("receipt=%+v calls=%d", r, dice.calls)
			}
		})
	}
	for _, rolls := range [][]int{{20}, {11, 8}} {
		if _, err := ResolveLongsword(sheet, 10, &attackDice{values: rolls}); err == nil {
			t.Fatal("invalid dice accepted")
		}
	}
}
