package srd521

import "testing"

func TestNewSupportedMartialBuild(t *testing.T) {
	build, err := NewSupportedMartialBuild("Human", "Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	stats, err := build.Stats()
	if err != nil {
		t.Fatal(err)
	}
	sheet, err := Project(stats, 1, 12, 12, 0)
	if err != nil {
		t.Fatal(err)
	}
	if sheet.Defenses["armor-class"].Value != 19 || sheet.Skills["athletics"].Value != 5 || sheet.Resources[0].Current != 2 || sheet.MaxHP.Value != 12 || build.OriginFeat != "Savage Attacker" || build.SpeciesFeat == "" || len(build.Proficiencies) != 8 || len(build.WeaponMastery) != 3 || build.Gold != 29 {
		t.Fatalf("unexpected derived build: AC=%d athletics=%d HP=%d build=%+v", sheet.Defenses["armor-class"].Value, sheet.Skills["athletics"].Value, sheet.MaxHP.Value, build)
	}
}

func TestNewSupportedMartialBuildRejectsIncompleteChoices(t *testing.T) {
	for _, tc := range []struct{ concept, background string }{{"elf fighter", "Soldier"}, {"Human Fighter", ""}, {"Human", "Soldier"}, {"Human Wizard Fighter", "Soldier"}, {"Human Wizard", "Soldier"}} {
		t.Run(tc.concept+tc.background, func(t *testing.T) {
			if _, err := MartialBuildForConcept(tc.concept, tc.background); err == nil {
				t.Fatal("unsupported build accepted")
			}
		})
	}
}
