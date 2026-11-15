package srd521

import (
	"os"
	"testing"
)

func TestCharacterProjection(t *testing.T) {
	stats, err := os.ReadFile("testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Project(stats, 1, 9, 12, 125)
	if err != nil {
		t.Fatal(err)
	}
	if s.Abilities["charisma"].Modifier.Value != -1 || s.Proficiency.Value != 2 || s.Saves["strength"].Value != 5 || s.Saves["constitution"].Value != 4 || s.Saves["wisdom"].Value != 1 || s.Skills["athletics"].Value != 5 || s.Skills["perception"].Value != 3 || s.Skills["stealth"].Value != 1 || len(s.Skills) != 18 || s.Defenses["armor-class"].Value != 19 || s.MaxHP.Value != 12 || s.HP.Value != 9 || s.Progression.NextLevelXP != 300 || s.Resources[0].Current != 1 {
		t.Fatalf("unexpected projection: %+v", s)
	}
	for _, v := range []Value{s.Saves["strength"], s.Skills["perception"], s.Defenses["armor-class"], s.MaxHP} {
		if len(v.Derivation) < 2 {
			t.Fatal("missing derivation")
		}
		for _, term := range v.Derivation {
			if !contentID.MatchString(term.SourceID) {
				t.Fatalf("invalid source %s", term.SourceID)
			}
		}
	}
	for _, c := range s.Capabilities {
		if c.Playable {
			t.Fatal("read-only projection advertised playable action")
		}
	}
	for _, tc := range []struct{ level, hp, max, xp int }{{2, 9, 12, 125}, {1, 13, 12, 125}, {1, 9, 20, 125}, {1, 9, 12, 300}} {
		if _, err := Project(stats, tc.level, tc.hp, tc.max, tc.xp); err == nil {
			t.Fatalf("unsupported state accepted: %+v", tc)
		}
	}
	if _, err := Project([]byte(`{"strength":16}`), 1, 9, 12, 0); err == nil {
		t.Fatal("legacy stats interpreted as SRD")
	}
}
