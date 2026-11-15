package combat

import "testing"

func TestSpatialRulings(t *testing.T) {
	feet, far, invalid, sight, blocked := 5, 10, -1, true, false
	for _, tc := range []struct {
		name  string
		space Space
		legal bool
		bonus int
	}{
		{"ordinary", Space{DistanceFeet: &feet, LineOfSight: &sight, Cover: "none"}, true, 0},
		{"half cover", Space{DistanceFeet: &feet, LineOfSight: &sight, Cover: "half"}, true, 2},
		{"three quarters", Space{DistanceFeet: &feet, LineOfSight: &sight, Cover: "three-quarters"}, true, 5},
		{"total cover", Space{DistanceFeet: &feet, LineOfSight: &sight, Cover: "total"}, false, 0},
		{"out of reach", Space{DistanceFeet: &far, LineOfSight: &sight, Cover: "none"}, false, 0},
		{"unseen", Space{DistanceFeet: &feet, LineOfSight: &blocked, Cover: "none"}, false, 0},
		{"unknown sight", Space{DistanceFeet: &feet, Cover: "none"}, false, 0},
		{"unknown distance", Space{LineOfSight: &sight, Cover: "none"}, false, 0},
		{"invalid distance", Space{DistanceFeet: &invalid, LineOfSight: &sight, Cover: "none"}, false, 0},
		{"unknown cover", Space{DistanceFeet: &feet, LineOfSight: &sight}, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if (tc.space.Melee() == nil) != tc.legal || tc.space.CoverBonus() != tc.bonus {
				t.Fatalf("ruling=%+v", tc.space)
			}
		})
	}
}

func TestMovementCopiesAndExactDistance(t *testing.T) {
	distance, sight := 10, true
	original := Space{DistanceFeet: &distance, LineOfSight: &sight, Cover: "none"}
	copy := original
	if err := copy.Move(EncounterApproach); err != nil {
		t.Fatal(err)
	}
	if *original.DistanceFeet != 10 || *copy.DistanceFeet != 5 {
		t.Fatal("movement changed the preview's source")
	}
	if err := copy.Move(EncounterRetreat); err != nil || *copy.DistanceFeet != 10 {
		t.Fatal("retreat did not change distance")
	}
	copy.Cover = "total"
	if err := copy.Move(EncounterApproach); err == nil {
		t.Fatal("moved through unresolved path")
	}
}
