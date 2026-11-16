package srd521

import "testing"

type banditTestDice struct {
	values []int
	calls  int
}

func (d *banditTestDice) Intn(n int) int {
	d.calls++
	if len(d.values) == 0 {
		return -1
	}
	v := d.values[0]
	d.values = d.values[1:]
	return v
}
func TestBanditAttacks(t *testing.T) {
	m, version, err := BanditRecord()
	if err != nil {
		t.Fatal(err)
	}
	if m.Available() {
		t.Fatal("bounded runtime must not promote whole catalogue support")
	}
	for _, action := range []string{"scimitar", "light-crossbow"} {
		for _, tc := range []struct {
			name    string
			rolls   []int
			damage  int
			outcome string
			invalid bool
		}{
			{"hit", []int{18, 2}, 4, "hit", false}, {"critical", []int{19, 1, 3}, 7, "critical-hit", false}, {"miss", []int{0}, 0, "miss", false},
			{"bad d20", []int{20}, 0, "", true}, {"bad damage", []int{19, -1}, 0, "", true},
		} {
			t.Run(action+tc.name, func(t *testing.T) {
				d := &banditTestDice{values: tc.rolls}
				r, err := ResolveBanditAttack(BanditProfile{version, action}, 12, d)
				if tc.invalid {
					if err == nil {
						t.Fatal("invalid dice accepted")
					}
					return
				}
				if err != nil || r.Damage != tc.damage || r.Outcome != tc.outcome || r.Modifier.Value != 3 {
					t.Fatalf("%+v %v", r, err)
				}
				if r.SourceIDs[0] != BanditID {
					t.Fatal("wrong source")
				}
			})
		}
	}
	for _, p := range []BanditProfile{{"stale", "scimitar"}, {version, "multiattack"}, {version, ""}} {
		d := &banditTestDice{}
		if _, err := ResolveBanditAttack(p, 12, d); err == nil || d.calls != 0 {
			t.Fatal("unsupported profile rolled")
		}
	}
}
