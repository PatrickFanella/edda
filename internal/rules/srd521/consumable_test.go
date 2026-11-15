package srd521

import "testing"

func TestHealingPotion(t *testing.T) {
	for _, tc := range []struct {
		name           string
		hp, max, after int
	}{
		{"ordinary", 1, 12, 8}, {"capped", 10, 12, 12}, {"full", 12, 12, 12},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dice := &attackDice{values: []int{1, 2}}
			r, err := ResolveHealingPotion(tc.hp, tc.max, dice)
			if err != nil {
				t.Fatal(err)
			}
			if r.RolledHealing != 7 || r.HPAfter != tc.after || r.Healing != tc.after-tc.hp || dice.calls != 2 {
				t.Fatalf("receipt=%+v", r)
			}
		})
	}
	for _, tc := range []struct {
		hp, max int
		rolls   []int
	}{{0, 12, nil}, {13, 12, nil}, {1, 12, []int{4}}, {1, 12, []int{-1}}} {
		if _, err := ResolveHealingPotion(tc.hp, tc.max, &attackDice{values: tc.rolls}); err == nil {
			t.Fatal("unsupported state/dice accepted")
		}
	}
}
