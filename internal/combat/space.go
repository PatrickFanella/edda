package combat

import "fmt"

// Space records the declared, symmetric spatial facts for the two-actor fixture.
// Nil/empty facts mean unresolved, never a guessed distance or clear sight line.
type Space struct {
	DistanceFeet *int   `json:"distance_feet"`
	LineOfSight  *bool  `json:"line_of_sight"`
	Cover        string `json:"cover"`
	Revision     string `json:"revision"`
}

// Validate requires an exact distance, sight ruling and recognized cover ruling.
func (s Space) Validate() error {
	if s.DistanceFeet == nil || *s.DistanceFeet < 0 || s.LineOfSight == nil {
		return fmt.Errorf("Clarify the exact distance and line of sight before rolling.")
	}
	switch s.Cover {
	case "none", "half", "three-quarters", "total":
		return nil
	default:
		return fmt.Errorf("Clarify cover before rolling.")
	}
}

// Melee validates ordinary five-foot reach. Unseen attacks remain unsupported.
func (s Space) Melee() error {
	if err := s.Validate(); err != nil {
		return err
	}
	if !*s.LineOfSight {
		return fmt.Errorf("No clear line of sight; unseen attacks require adjudication.")
	}
	if s.Cover == "total" {
		return fmt.Errorf("Total cover blocks this attack.")
	}
	if *s.DistanceFeet > 5 {
		return fmt.Errorf("Out of reach: this melee attack requires 5 feet or less.")
	}
	return nil
}

// CoverBonus is the symmetric fixture's declared bonus to Armor Class.
func (s Space) CoverBonus() int {
	switch s.Cover {
	case "half":
		return 2
	case "three-quarters":
		return 5
	}
	return 0
}

// Move changes an exact distance by five feet without inventing a grid position.
func (s *Space) Move(kind string) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if !*s.LineOfSight || s.Cover != "none" {
		return fmt.Errorf("Clarify an unobstructed path before moving.")
	}
	distance := *s.DistanceFeet
	switch kind {
	case EncounterApproach:
		if distance < 5 {
			return fmt.Errorf("Already within 5 feet; cannot approach further.")
		}
		distance -= 5
	case EncounterRetreat:
		distance += 5
	default:
		return fmt.Errorf("Unsupported spatial movement.")
	}
	s.DistanceFeet = &distance
	return nil
}
