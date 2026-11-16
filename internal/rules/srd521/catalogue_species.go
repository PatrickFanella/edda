package srd521

import (
	"encoding/json"
	"fmt"
)

// CharacterSpecies describes source anatomy and traits. It is never a runtime grant.
// Sizes are alternatives at species selection, not simultaneous sizes.
type CharacterSpecies struct {
	Category      string                  `json:"category"`
	CreatureType  string                  `json:"creature_type"`
	SpeedFeet     int                     `json:"speed_feet"`
	Sizes         []CharacterSpeciesSize  `json:"sizes"`
	SizeChoiceRef string                  `json:"size_choice_ref,omitempty"`
	Traits        []CharacterSpeciesTrait `json:"traits"`
}

// CharacterSpeciesSize retains the source approximate height range in feet.
type CharacterSpeciesSize struct {
	Key               string `json:"key"`
	MinimumHeightFeet int    `json:"minimum_height_feet"`
	MaximumHeightFeet int    `json:"maximum_height_feet"`
}

// CharacterSpeciesTrait retains fixed grants and their conditions. Keys and
// numeric parameter keys must be required by the owning origin. CharacterLevels
// are replacement parameter rows at total character levels, never class levels.
// Prerequisites are conjunctive; an empty list means an automatic trait.
type CharacterSpeciesTrait struct {
	Key               string                  `json:"key"`
	PrerequisiteRefs  []string                `json:"prerequisite_refs"`
	RequiredMechanics []string                `json:"required_mechanics"`
	Parameters        CharacterSpeciesNumbers `json:"parameters"`
	CharacterLevels   []CharacterSpeciesLevel `json:"character_levels"`
	ResourceRef       string                  `json:"resource_ref,omitempty"`
}

// CharacterSpeciesNumbers preserves explicit zero while refusing null operands.
type CharacterSpeciesNumbers map[string]int

// UnmarshalJSON requires each supplied operand to be an explicit integer.
func (n *CharacterSpeciesNumbers) UnmarshalJSON(data []byte) error {
	var values map[string]*int
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if values == nil {
		return fmt.Errorf("species numbers require an object")
	}
	decoded := make(CharacterSpeciesNumbers, len(values))
	for key, value := range values {
		if value == nil {
			return fmt.Errorf("species numeric operand %q must be an explicit integer", key)
		}
		decoded[key] = *value
	}
	*n = decoded
	return nil
}

// CharacterSpeciesLevel replaces named numeric parameters at a character level.
type CharacterSpeciesLevel struct {
	Level      int                     `json:"level"`
	Parameters CharacterSpeciesNumbers `json:"parameters"`
}

const speciesSelectedOption = "srd-5.2.1:mechanic/species/option-selected"

func validateSpecies(r CharacterRecord) error {
	s := r.Species
	if r.Kind != "origin" || len(r.Sources) != 1 || r.Sources[0].InventoryID != "srd-5.2.1:planning/species/"+r.ID[len("srd-5.2.1:character/origin/"):] || s.Category != "species" || s.CreatureType != "humanoid" || s.SpeedFeet <= 0 || len(s.Sizes) == 0 || len(s.Traits) == 0 {
		return fmt.Errorf("%s: invalid species header", r.ID)
	}
	required := map[string]bool{}
	for _, k := range r.RequiredMechanics {
		required[k] = true
	}
	sizes := map[string]bool{}
	for _, size := range s.Sizes {
		if (size.Key != "small" && size.Key != "medium") || sizes[size.Key] || size.MinimumHeightFeet < 1 || size.MaximumHeightFeet < size.MinimumHeightFeet {
			return fmt.Errorf("%s: invalid species size", r.ID)
		}
		sizes[size.Key] = true
	}
	if (len(s.Sizes) > 1) != (s.SizeChoiceRef != "") {
		return fmt.Errorf("%s: size choice required", r.ID)
	}
	seen := map[string]bool{}
	for _, trait := range s.Traits {
		if !required[trait.Key] || seen[trait.Key] || trait.PrerequisiteRefs == nil || len(trait.RequiredMechanics) == 0 || trait.Parameters == nil || trait.CharacterLevels == nil {
			return fmt.Errorf("%s: invalid species trait", r.ID)
		}
		seen[trait.Key] = true
		keys := map[string]bool{}
		for _, k := range trait.RequiredMechanics {
			if !required[k] || keys[k] {
				return fmt.Errorf("%s: trait key must be required and unique", r.ID)
			}
			keys[k] = true
		}
		check := func(values map[string]int) bool {
			for k, v := range values {
				if !keys[k] || v < 0 {
					return false
				}
			}
			return true
		}
		if !check(trait.Parameters) {
			return fmt.Errorf("%s: invalid trait numbers", r.ID)
		}
		previous := 0
		for _, row := range trait.CharacterLevels {
			if row.Level <= previous || row.Level > 20 || len(row.Parameters) == 0 || !check(row.Parameters) {
				return fmt.Errorf("%s: invalid character-level trait row", r.ID)
			}
			previous = row.Level
		}
	}
	return nil
}

func validateSpeciesLinks(r CharacterRecord, records map[string]CharacterRecord) error {
	if q := r.Requirement; q != nil && (q.OptionKey != "" || q.Key == speciesSelectedOption) {
		if q.Key != speciesSelectedOption || len(q.RecordRefs) != 1 || len(q.Parameters) != 0 {
			return fmt.Errorf("%s: invalid selected-option predicate", r.ID)
		}
		choice := records[q.RecordRefs[0]]
		found := false
		for _, k := range choice.OptionKeys {
			if k == q.OptionKey {
				found = true
			}
		}
		if choice.Kind != "choice" || !found {
			return fmt.Errorf("%s: missing selected option", r.ID)
		}
		root := func(v CharacterRecord) string {
			for n := 0; v.ParentID != "" && n < len(records); n++ {
				v = records[v.ParentID]
			}
			return v.ID
		}
		if root(r) != root(choice) {
			return fmt.Errorf("%s: selected option crosses owner", r.ID)
		}
	}
	if r.Species == nil {
		return nil
	}
	s := r.Species
	if s.SizeChoiceRef != "" {
		c := records[s.SizeChoiceRef]
		if c.Kind != "choice" || c.ParentID != r.ID || c.Select != 1 || len(c.OptionKeys) != len(s.Sizes) {
			return fmt.Errorf("%s: invalid size choice", r.ID)
		}
		for i, size := range s.Sizes {
			if c.OptionKeys[i] != "srd-5.2.1:mechanic/size/"+size.Key {
				return fmt.Errorf("%s: size alternatives disagree", r.ID)
			}
		}
	}
	for _, t := range s.Traits {
		seen := map[string]bool{}
		for _, id := range t.PrerequisiteRefs {
			q := records[id]
			if q.Kind != "prerequisite" || q.ParentID != r.ID || seen[id] {
				return fmt.Errorf("%s: invalid trait prerequisite", r.ID)
			}
			seen[id] = true
		}
		if t.ResourceRef != "" {
			q := records[t.ResourceRef]
			if q.Kind != "resource" || q.ParentID != r.ID {
				return fmt.Errorf("%s: invalid trait resource", r.ID)
			}
		}
	}
	return nil
}
