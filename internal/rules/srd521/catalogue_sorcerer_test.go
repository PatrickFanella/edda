package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func sorcererCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-sorcerer.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// Each block freezes the source graph, including ownership, operands and
// complete table cells. Defaults describe serialized zero fields, not missing
// assertions. This oracle does not call the PDF script or load candidate data.
func sorcererSourceAssertions(t *testing.T) []map[string]any {
	t.Helper()
	var expected []map[string]any
	add := func(kind, slug, name, parent, inventory string, page, level int, fields string) {
		r := map[string]any{
			"id":   "srd-5.2.1:character/" + kind + "/sorcerer" + slug,
			"kind": kind, "name": name,
			"parent_id": parent, "level": float64(level),
			"sources":     []any{map[string]any{"inventory_id": inventory, "page": float64(page)}},
			"choice_refs": []any{}, "resource_refs": []any{}, "prerequisite_refs": []any{},
			"required_mechanics": []any{}, "select": float64(0), "option_keys": []any{}, "capacity": float64(0),
		}
		var extra map[string]any
		if err := json.Unmarshal([]byte(fields), &extra); err != nil {
			t.Fatal(err)
		}
		for key, value := range extra {
			r[key] = value
		}
		expected = append(expected, r)
	}
	// Sorcerer, PDF p64.
	add("class", "", "Sorcerer", "", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/entry"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/primary-ability/charisma",
    "srd-5.2.1:mechanic/sorcerer/hit-die-sides",
    "srd-5.2.1:mechanic/saving-throws/constitution",
    "srd-5.2.1:mechanic/saving-throws/charisma",
    "srd-5.2.1:mechanic/weapons/simple",
    "srd-5.2.1:mechanic/sorcerer/armor-training/none"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/hit-die-sides": 6
  }
}`)
	// Entry, PDF p64.
	add("choice", "/entry", "Entry", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/first-class-traits",
    "srd-5.2.1:character/choice/sorcerer/multiclass-traits"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/entry"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/sorcerer/entry/first-class",
    "srd-5.2.1:mechanic/sorcerer/entry/multiclass"
  ]
}`)
	// First Class Traits, PDF p64.
	add("choice", "/first-class-traits", "First Class Traits", "srd-5.2.1:character/choice/sorcerer/entry", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/skills",
    "srd-5.2.1:character/choice/sorcerer/starting-equipment"
  ],
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/sorcerer/first-class-entry"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/first-class-traits"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/sorcerer/all-core-traits-and-level-one-features"
  ]
}`)
	// First Class Entry, PDF p64.
	add("prerequisite", "/first-class-entry", "First Class Entry", "srd-5.2.1:character/choice/sorcerer/first-class-traits", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/first-class-entry-only"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/first-class-entry-only",
    "record_refs": [
      "srd-5.2.1:character/choice/sorcerer/entry",
      "srd-5.2.1:character/class/sorcerer"
    ],
    "parameters": {}
  }
}`)
	// Multiclass Traits, PDF p64.
	add("choice", "/multiclass-traits", "Multiclass Traits", "srd-5.2.1:character/choice/sorcerer/entry", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/sorcerer/multiclass-entry"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/multiclass-traits",
    "srd-5.2.1:mechanic/sorcerer/multiclass/use-multiclass-spell-slot-rules"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/sorcerer/hit-die-and-level-one-features"
  ]
}`)
	// Multiclass Entry, PDF p64.
	add("prerequisite", "/multiclass-entry", "Multiclass Entry", "srd-5.2.1:character/choice/sorcerer/multiclass-traits", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/multiclass-entry-only"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/multiclass-entry-only",
    "record_refs": [
      "srd-5.2.1:character/choice/sorcerer/entry",
      "srd-5.2.1:character/class/sorcerer"
    ],
    "parameters": {}
  }
}`)
	// Skills, PDF p64.
	add("choice", "/skills", "Skills", "srd-5.2.1:character/choice/sorcerer/first-class-traits", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/skills"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/skills/arcana",
    "srd-5.2.1:mechanic/skills/deception",
    "srd-5.2.1:mechanic/skills/insight",
    "srd-5.2.1:mechanic/skills/intimidation",
    "srd-5.2.1:mechanic/skills/persuasion",
    "srd-5.2.1:mechanic/skills/religion"
  ]
}`)
	// Starting Equipment, PDF p64.
	add("choice", "/starting-equipment", "Starting Equipment", "srd-5.2.1:character/choice/sorcerer/first-class-traits", "srd-5.2.1:planning/classes/sorcerer", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/starting-equipment",
    "srd-5.2.1:mechanic/equipment/spear",
    "srd-5.2.1:mechanic/equipment/dagger",
    "srd-5.2.1:mechanic/equipment/arcane-focus-crystal",
    "srd-5.2.1:mechanic/equipment/dungeoneers-pack",
    "srd-5.2.1:mechanic/currency/gp"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/sorcerer/equipment/a",
    "srd-5.2.1:mechanic/sorcerer/equipment/b"
  ],
  "options": [
    {
      "key": "srd-5.2.1:mechanic/sorcerer/equipment/a",
      "quantities": {
        "srd-5.2.1:mechanic/equipment/spear": 1,
        "srd-5.2.1:mechanic/equipment/dagger": 2,
        "srd-5.2.1:mechanic/equipment/arcane-focus-crystal": 1,
        "srd-5.2.1:mechanic/equipment/dungeoneers-pack": 1,
        "srd-5.2.1:mechanic/currency/gp": 28
      }
    },
    {
      "key": "srd-5.2.1:mechanic/sorcerer/equipment/b",
      "quantities": {
        "srd-5.2.1:mechanic/currency/gp": 50
      }
    }
  ]
}`)
	// Sorcerer level 1, PDF p64.
	add("level", "/1", "Sorcerer level 1", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 1, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/spellcasting",
      "srd-5.2.1:character/feature/sorcerer/innate-sorcery"
    ],
    "feature_keys": []
  }
}`)
	// Spellcasting, PDF p64.
	add("feature", "/spellcasting", "Spellcasting", "srd-5.2.1:character/level/sorcerer/1", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/cantrips",
    "srd-5.2.1:character/choice/sorcerer/prepared-spells"
  ],
  "resource_refs": [
    "srd-5.2.1:character/resource/sorcerer/spell-slot-1",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-2",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-3",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-4",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-5",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-6",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-7",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-8",
    "srd-5.2.1:character/resource/sorcerer/spell-slot-9"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spellcasting/charisma",
    "srd-5.2.1:mechanic/sorcerer/spellcasting/arcane-focus",
    "srd-5.2.1:mechanic/sorcerer/spellcasting/always-prepared-excluded-still-sorcerer-spells",
    "srd-5.2.1:mechanic/sorcerer/recommended-cantrip/light",
    "srd-5.2.1:mechanic/sorcerer/recommended-cantrip/prestidigitation",
    "srd-5.2.1:mechanic/sorcerer/recommended-cantrip/shocking-grasp",
    "srd-5.2.1:mechanic/sorcerer/recommended-cantrip/sorcerous-burst",
    "srd-5.2.1:mechanic/sorcerer/recommended-spell/burning-hands",
    "srd-5.2.1:mechanic/sorcerer/recommended-spell/detect-magic",
    "srd-5.2.1:mechanic/sorcerer/spellcasting/level-gain-replace-one-sorcerer-cantrip",
    "srd-5.2.1:mechanic/sorcerer/cantrip-replacements",
    "srd-5.2.1:mechanic/sorcerer/spellcasting/level-gain-replace-one-eligible-sorcerer-spell",
    "srd-5.2.1:mechanic/sorcerer/prepared-replacements",
    "srd-5.2.1:mechanic/sorcerer/feature/spellcasting"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/cantrip-replacements": 1,
    "srd-5.2.1:mechanic/sorcerer/prepared-replacements": 1
  }
}`)
	// Innate Sorcery, PDF p65.
	add("feature", "/innate-sorcery", "Innate Sorcery", "srd-5.2.1:character/level/sorcerer/1", "srd-5.2.1:planning/class-features/sorcerer-level-1-innate-sorcery", 65, 0, `{
  "resource_refs": [
    "srd-5.2.1:character/resource/sorcerer/innate-sorcery"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/innate-sorcery/bonus-action",
    "srd-5.2.1:mechanic/sorcerer/duration-minutes",
    "srd-5.2.1:mechanic/sorcerer/innate-sorcery/sorcerer-spell-save-dc-increase",
    "srd-5.2.1:mechanic/sorcerer/save-dc-increase",
    "srd-5.2.1:mechanic/sorcerer/innate-sorcery/advantage-sorcerer-spell-attacks",
    "srd-5.2.1:mechanic/sorcerer/feature/innate-sorcery"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/duration-minutes": 1,
    "srd-5.2.1:mechanic/sorcerer/save-dc-increase": 1
  }
}`)
	// Sorcerer level 2, PDF p64.
	add("level", "/2", "Sorcerer level 2", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 2, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/font-of-magic",
      "srd-5.2.1:character/feature/sorcerer/metamagic"
    ],
    "feature_keys": []
  }
}`)
	// Font of Magic, PDF p66.
	add("feature", "/font-of-magic", "Font of Magic", "srd-5.2.1:character/level/sorcerer/2", "srd-5.2.1:planning/class-features/sorcerer-level-2-font-of-magic", 66, 0, `{
  "resource_refs": [
    "srd-5.2.1:character/resource/sorcerer/sorcery-points"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/font-of-magic/capped-at-table-maximum",
    "srd-5.2.1:mechanic/sorcerer/font-of-magic/expend-slot-gain-slot-level-points-no-action",
    "srd-5.2.1:mechanic/sorcerer/font-of-magic/bonus-action-spend-unexpended-points-create-slot",
    "srd-5.2.1:mechanic/sorcerer/created-slots-per-conversion",
    "srd-5.2.1:mechanic/sorcerer/font-of-magic/maximum-created-slot-level",
    "srd-5.2.1:mechanic/sorcerer/maximum-created-slot-level",
    "srd-5.2.1:mechanic/sorcerer/font-of-magic/created-slot-vanishes-long-rest",
    "srd-5.2.1:mechanic/sorcerer/create-slot-1-sorcery-point-cost",
    "srd-5.2.1:mechanic/sorcerer/create-slot-1-minimum-sorcerer-level",
    "srd-5.2.1:mechanic/sorcerer/create-slot-2-sorcery-point-cost",
    "srd-5.2.1:mechanic/sorcerer/create-slot-2-minimum-sorcerer-level",
    "srd-5.2.1:mechanic/sorcerer/create-slot-3-sorcery-point-cost",
    "srd-5.2.1:mechanic/sorcerer/create-slot-3-minimum-sorcerer-level",
    "srd-5.2.1:mechanic/sorcerer/create-slot-4-sorcery-point-cost",
    "srd-5.2.1:mechanic/sorcerer/create-slot-4-minimum-sorcerer-level",
    "srd-5.2.1:mechanic/sorcerer/create-slot-5-sorcery-point-cost",
    "srd-5.2.1:mechanic/sorcerer/create-slot-5-minimum-sorcerer-level",
    "srd-5.2.1:mechanic/sorcerer/feature/font-of-magic"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/created-slots-per-conversion": 1,
    "srd-5.2.1:mechanic/sorcerer/maximum-created-slot-level": 5,
    "srd-5.2.1:mechanic/sorcerer/create-slot-1-sorcery-point-cost": 2,
    "srd-5.2.1:mechanic/sorcerer/create-slot-1-minimum-sorcerer-level": 2,
    "srd-5.2.1:mechanic/sorcerer/create-slot-2-sorcery-point-cost": 3,
    "srd-5.2.1:mechanic/sorcerer/create-slot-2-minimum-sorcerer-level": 3,
    "srd-5.2.1:mechanic/sorcerer/create-slot-3-sorcery-point-cost": 5,
    "srd-5.2.1:mechanic/sorcerer/create-slot-3-minimum-sorcerer-level": 5,
    "srd-5.2.1:mechanic/sorcerer/create-slot-4-sorcery-point-cost": 6,
    "srd-5.2.1:mechanic/sorcerer/create-slot-4-minimum-sorcerer-level": 7,
    "srd-5.2.1:mechanic/sorcerer/create-slot-5-sorcery-point-cost": 7,
    "srd-5.2.1:mechanic/sorcerer/create-slot-5-minimum-sorcerer-level": 9
  }
}`)
	// Metamagic, PDF p66.
	add("feature", "/metamagic", "Metamagic", "srd-5.2.1:character/level/sorcerer/2", "srd-5.2.1:planning/class-features/sorcerer-level-2-metamagic", 66, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/metamagic"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/metamagic/chosen-options-modify-own-cast-spells",
    "srd-5.2.1:mechanic/sorcerer/metamagic/spend-option-sorcery-point-cost",
    "srd-5.2.1:mechanic/sorcerer/metamagic/one-per-spell-unless-option-exception",
    "srd-5.2.1:mechanic/sorcerer/options-per-spell",
    "srd-5.2.1:mechanic/sorcerer/metamagic/level-gain-replace-one-with-unknown-option",
    "srd-5.2.1:mechanic/sorcerer/replacements",
    "srd-5.2.1:mechanic/sorcerer/feature/metamagic"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/options-per-spell": 1,
    "srd-5.2.1:mechanic/sorcerer/replacements": 1
  }
}`)
	// Sorcerer level 3, PDF p64.
	add("level", "/3", "Sorcerer level 3", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 3, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/sorcerer-subclass"
    ],
    "feature_keys": []
  }
}`)
	// Sorcerer Subclass, PDF p66.
	add("feature", "/sorcerer-subclass", "Sorcerer Subclass", "srd-5.2.1:character/level/sorcerer/3", "srd-5.2.1:planning/class-features/sorcerer-level-3-sorcerer-subclass", 66, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/subclass"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/feature/sorcerer-subclass"
  ]
}`)
	// Sorcerer level 4, PDF p64.
	add("level", "/4", "Sorcerer level 4", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 4, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Ability Score Improvement, PDF p66.
	add("feature", "/ability-score-improvement", "Ability Score Improvement", "srd-5.2.1:character/level/sorcerer/4", "srd-5.2.1:planning/class-features/sorcerer-level-4-ability-score-improvement", 66, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/ability-score-improvement"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/feature/ability-score-improvement"
  ]
}`)
	// Sorcerer level 5, PDF p64.
	add("level", "/5", "Sorcerer level 5", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 5, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/sorcerous-restoration"
    ],
    "feature_keys": []
  }
}`)
	// Sorcerous Restoration, PDF p66.
	add("feature", "/sorcerous-restoration", "Sorcerous Restoration", "srd-5.2.1:character/level/sorcerer/5", "srd-5.2.1:planning/class-features/sorcerer-level-5-sorcerous-restoration", 66, 0, `{
  "resource_refs": [
    "srd-5.2.1:character/resource/sorcerer/sorcerous-restoration"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/restoration/optional-short-rest-recover-expended-points",
    "srd-5.2.1:mechanic/sorcerer/restoration/maximum-half-sorcerer-level-round-down",
    "srd-5.2.1:mechanic/sorcerer/recovery-level-divisor",
    "srd-5.2.1:mechanic/sorcerer/restoration/once-until-long-rest",
    "srd-5.2.1:mechanic/sorcerer/feature/sorcerous-restoration"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/recovery-level-divisor": 2
  }
}`)
	// Sorcerer level 6, PDF p64.
	add("level", "/6", "Sorcerer level 6", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 6, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression",
    "srd-5.2.1:mechanic/sorcerer/subclass-feature"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [],
    "feature_keys": [
      "srd-5.2.1:mechanic/sorcerer/subclass-feature"
    ]
  }
}`)
	// Sorcerer level 7, PDF p64.
	add("level", "/7", "Sorcerer level 7", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 7, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/sorcery-incarnate"
    ],
    "feature_keys": []
  }
}`)
	// Sorcery Incarnate, PDF p66.
	add("feature", "/sorcery-incarnate", "Sorcery Incarnate", "srd-5.2.1:character/level/sorcerer/7", "srd-5.2.1:planning/class-features/sorcerer-level-7-sorcery-incarnate", 66, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/sorcerer/incarnate-no-uses",
    "srd-5.2.1:character/prerequisite/sorcerer/incarnate-active"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/incarnate/spend-points-bonus-action-activate-innate-sorcery",
    "srd-5.2.1:mechanic/sorcerer/sorcery-point-cost",
    "srd-5.2.1:mechanic/sorcerer/incarnate/innate-active-two-metamagic-options-per-spell",
    "srd-5.2.1:mechanic/sorcerer/options-per-spell",
    "srd-5.2.1:mechanic/sorcerer/feature/sorcery-incarnate"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/sorcery-point-cost": 2,
    "srd-5.2.1:mechanic/sorcerer/options-per-spell": 2
  }
}`)
	// Sorcerer level 8, PDF p64.
	add("level", "/8", "Sorcerer level 8", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 8, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Sorcerer level 9, PDF p64.
	add("level", "/9", "Sorcerer level 9", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 9, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [],
    "feature_keys": []
  }
}`)
	// Sorcerer level 10, PDF p64.
	add("level", "/10", "Sorcerer level 10", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 10, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/metamagic"
    ],
    "feature_keys": []
  }
}`)
	// Sorcerer level 11, PDF p64.
	add("level", "/11", "Sorcerer level 11", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 11, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [],
    "feature_keys": []
  }
}`)
	// Sorcerer level 12, PDF p64.
	add("level", "/12", "Sorcerer level 12", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 12, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Sorcerer level 13, PDF p64.
	add("level", "/13", "Sorcerer level 13", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 13, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [],
    "feature_keys": []
  }
}`)
	// Sorcerer level 14, PDF p64.
	add("level", "/14", "Sorcerer level 14", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 14, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression",
    "srd-5.2.1:mechanic/sorcerer/subclass-feature"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [],
    "feature_keys": [
      "srd-5.2.1:mechanic/sorcerer/subclass-feature"
    ]
  }
}`)
	// Sorcerer level 15, PDF p64.
	add("level", "/15", "Sorcerer level 15", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 15, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [],
    "feature_keys": []
  }
}`)
	// Sorcerer level 16, PDF p64.
	add("level", "/16", "Sorcerer level 16", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 16, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Sorcerer level 17, PDF p64.
	add("level", "/17", "Sorcerer level 17", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 17, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/metamagic"
    ],
    "feature_keys": []
  }
}`)
	// Sorcerer level 18, PDF p64.
	add("level", "/18", "Sorcerer level 18", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 18, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression",
    "srd-5.2.1:mechanic/sorcerer/subclass-feature"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [],
    "feature_keys": [
      "srd-5.2.1:mechanic/sorcerer/subclass-feature"
    ]
  }
}`)
	// Sorcerer level 19, PDF p64.
	add("level", "/19", "Sorcerer level 19", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 19, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/epic-boon"
    ],
    "feature_keys": []
  }
}`)
	// Epic Boon, PDF p66.
	add("feature", "/epic-boon", "Epic Boon", "srd-5.2.1:character/level/sorcerer/19", "srd-5.2.1:planning/class-features/sorcerer-level-19-epic-boon", 66, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/sorcerer/epic-boon"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/recommended-feat/boon-of-dimensional-travel",
    "srd-5.2.1:mechanic/sorcerer/feature/epic-boon"
  ]
}`)
	// Sorcerer level 20, PDF p64.
	add("level", "/20", "Sorcerer level 20", "srd-5.2.1:character/class/sorcerer", "srd-5.2.1:planning/classes/sorcerer", 64, 20, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [
      "srd-5.2.1:character/feature/sorcerer/arcane-apotheosis"
    ],
    "feature_keys": []
  }
}`)
	// Arcane Apotheosis, PDF p66.
	add("feature", "/arcane-apotheosis", "Arcane Apotheosis", "srd-5.2.1:character/level/sorcerer/20", "srd-5.2.1:planning/class-features/sorcerer-level-20-arcane-apotheosis", 66, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/sorcerer/apotheosis-active"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/apotheosis/waive-one-option-point-cost-each-own-turn",
    "srd-5.2.1:mechanic/sorcerer/free-options-per-own-turn",
    "srd-5.2.1:mechanic/sorcerer/feature/arcane-apotheosis"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/sorcerer/free-options-per-own-turn": 1
  }
}`)
	// Cantrips, PDF p64.
	add("choice", "/cantrips", "Cantrips", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/cantrips"
  ],
  "select": 4,
  "option_keys": [
    "srd-5.2.1:mechanic/spells/sorcerer-cantrips"
  ],
  "values_by_level": [
    {
      "level": 1,
      "value": 4
    },
    {
      "level": 2,
      "value": 4
    },
    {
      "level": 3,
      "value": 4
    },
    {
      "level": 4,
      "value": 5
    },
    {
      "level": 5,
      "value": 5
    },
    {
      "level": 6,
      "value": 5
    },
    {
      "level": 7,
      "value": 5
    },
    {
      "level": 8,
      "value": 5
    },
    {
      "level": 9,
      "value": 5
    },
    {
      "level": 10,
      "value": 6
    },
    {
      "level": 11,
      "value": 6
    },
    {
      "level": 12,
      "value": 6
    },
    {
      "level": 13,
      "value": 6
    },
    {
      "level": 14,
      "value": 6
    },
    {
      "level": 15,
      "value": 6
    },
    {
      "level": 16,
      "value": 6
    },
    {
      "level": 17,
      "value": 6
    },
    {
      "level": 18,
      "value": 6
    },
    {
      "level": 19,
      "value": 6
    },
    {
      "level": 20,
      "value": 6
    }
  ]
}`)
	// Prepared Spells, PDF p64.
	add("choice", "/prepared-spells", "Prepared Spells", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/sorcerer/prepared-slot-level",
    "srd-5.2.1:character/prerequisite/sorcerer/initial-prepared-level"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spellcasting/prepare-to-table-count-on-increase"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/spells/sorcerer-list"
  ],
  "values_by_level": [
    {
      "level": 1,
      "value": 2
    },
    {
      "level": 2,
      "value": 4
    },
    {
      "level": 3,
      "value": 6
    },
    {
      "level": 4,
      "value": 7
    },
    {
      "level": 5,
      "value": 9
    },
    {
      "level": 6,
      "value": 10
    },
    {
      "level": 7,
      "value": 11
    },
    {
      "level": 8,
      "value": 12
    },
    {
      "level": 9,
      "value": 14
    },
    {
      "level": 10,
      "value": 15
    },
    {
      "level": 11,
      "value": 16
    },
    {
      "level": 12,
      "value": 16
    },
    {
      "level": 13,
      "value": 17
    },
    {
      "level": 14,
      "value": 17
    },
    {
      "level": 15,
      "value": 18
    },
    {
      "level": 16,
      "value": 18
    },
    {
      "level": 17,
      "value": 19
    },
    {
      "level": 18,
      "value": 20
    },
    {
      "level": 19,
      "value": 21
    },
    {
      "level": 20,
      "value": 22
    }
  ]
}`)
	// Prepared Slot Level, PDF p64.
	add("prerequisite", "/prepared-slot-level", "Prepared Slot Level", "srd-5.2.1:character/choice/sorcerer/prepared-spells", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-level-with-available-slot"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/spell-level-with-available-slot",
    "record_refs": [
      "srd-5.2.1:character/feature/sorcerer/spellcasting",
      "srd-5.2.1:character/choice/sorcerer/prepared-spells"
    ],
    "parameters": {
      "minimum-spell-level": 1
    }
  }
}`)
	// Initial Prepared Level, PDF p64.
	add("prerequisite", "/initial-prepared-level", "Initial Prepared Level", "srd-5.2.1:character/choice/sorcerer/prepared-spells", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/initial-selection-level-one-only"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/initial-selection-level-one-only",
    "record_refs": [
      "srd-5.2.1:character/choice/sorcerer/prepared-spells"
    ],
    "parameters": {
      "spell-level": 1
    }
  }
}`)
	// Spell Slot 1, PDF p64.
	add("resource", "/spell-slot-1", "Spell Slot 1", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-1/long-rest-recover-all"
  ],
  "capacity": 2,
  "values_by_level": [
    {
      "level": 1,
      "value": 2
    },
    {
      "level": 2,
      "value": 3
    },
    {
      "level": 3,
      "value": 4
    },
    {
      "level": 4,
      "value": 4
    },
    {
      "level": 5,
      "value": 4
    },
    {
      "level": 6,
      "value": 4
    },
    {
      "level": 7,
      "value": 4
    },
    {
      "level": 8,
      "value": 4
    },
    {
      "level": 9,
      "value": 4
    },
    {
      "level": 10,
      "value": 4
    },
    {
      "level": 11,
      "value": 4
    },
    {
      "level": 12,
      "value": 4
    },
    {
      "level": 13,
      "value": 4
    },
    {
      "level": 14,
      "value": 4
    },
    {
      "level": 15,
      "value": 4
    },
    {
      "level": 16,
      "value": 4
    },
    {
      "level": 17,
      "value": 4
    },
    {
      "level": 18,
      "value": 4
    },
    {
      "level": 19,
      "value": 4
    },
    {
      "level": 20,
      "value": 4
    }
  ]
}`)
	// Spell Slot 2, PDF p64.
	add("resource", "/spell-slot-2", "Spell Slot 2", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-2/long-rest-recover-all"
  ],
  "capacity": 2,
  "values_by_level": [
    {
      "level": 3,
      "value": 2
    },
    {
      "level": 4,
      "value": 3
    },
    {
      "level": 5,
      "value": 3
    },
    {
      "level": 6,
      "value": 3
    },
    {
      "level": 7,
      "value": 3
    },
    {
      "level": 8,
      "value": 3
    },
    {
      "level": 9,
      "value": 3
    },
    {
      "level": 10,
      "value": 3
    },
    {
      "level": 11,
      "value": 3
    },
    {
      "level": 12,
      "value": 3
    },
    {
      "level": 13,
      "value": 3
    },
    {
      "level": 14,
      "value": 3
    },
    {
      "level": 15,
      "value": 3
    },
    {
      "level": 16,
      "value": 3
    },
    {
      "level": 17,
      "value": 3
    },
    {
      "level": 18,
      "value": 3
    },
    {
      "level": 19,
      "value": 3
    },
    {
      "level": 20,
      "value": 3
    }
  ]
}`)
	// Spell Slot 3, PDF p64.
	add("resource", "/spell-slot-3", "Spell Slot 3", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-3/long-rest-recover-all"
  ],
  "capacity": 2,
  "values_by_level": [
    {
      "level": 5,
      "value": 2
    },
    {
      "level": 6,
      "value": 3
    },
    {
      "level": 7,
      "value": 3
    },
    {
      "level": 8,
      "value": 3
    },
    {
      "level": 9,
      "value": 3
    },
    {
      "level": 10,
      "value": 3
    },
    {
      "level": 11,
      "value": 3
    },
    {
      "level": 12,
      "value": 3
    },
    {
      "level": 13,
      "value": 3
    },
    {
      "level": 14,
      "value": 3
    },
    {
      "level": 15,
      "value": 3
    },
    {
      "level": 16,
      "value": 3
    },
    {
      "level": 17,
      "value": 3
    },
    {
      "level": 18,
      "value": 3
    },
    {
      "level": 19,
      "value": 3
    },
    {
      "level": 20,
      "value": 3
    }
  ]
}`)
	// Spell Slot 4, PDF p64.
	add("resource", "/spell-slot-4", "Spell Slot 4", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-4/long-rest-recover-all"
  ],
  "capacity": 1,
  "values_by_level": [
    {
      "level": 7,
      "value": 1
    },
    {
      "level": 8,
      "value": 2
    },
    {
      "level": 9,
      "value": 3
    },
    {
      "level": 10,
      "value": 3
    },
    {
      "level": 11,
      "value": 3
    },
    {
      "level": 12,
      "value": 3
    },
    {
      "level": 13,
      "value": 3
    },
    {
      "level": 14,
      "value": 3
    },
    {
      "level": 15,
      "value": 3
    },
    {
      "level": 16,
      "value": 3
    },
    {
      "level": 17,
      "value": 3
    },
    {
      "level": 18,
      "value": 3
    },
    {
      "level": 19,
      "value": 3
    },
    {
      "level": 20,
      "value": 3
    }
  ]
}`)
	// Spell Slot 5, PDF p64.
	add("resource", "/spell-slot-5", "Spell Slot 5", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-5/long-rest-recover-all"
  ],
  "capacity": 1,
  "values_by_level": [
    {
      "level": 9,
      "value": 1
    },
    {
      "level": 10,
      "value": 2
    },
    {
      "level": 11,
      "value": 2
    },
    {
      "level": 12,
      "value": 2
    },
    {
      "level": 13,
      "value": 2
    },
    {
      "level": 14,
      "value": 2
    },
    {
      "level": 15,
      "value": 2
    },
    {
      "level": 16,
      "value": 2
    },
    {
      "level": 17,
      "value": 2
    },
    {
      "level": 18,
      "value": 3
    },
    {
      "level": 19,
      "value": 3
    },
    {
      "level": 20,
      "value": 3
    }
  ]
}`)
	// Spell Slot 6, PDF p64.
	add("resource", "/spell-slot-6", "Spell Slot 6", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-6/long-rest-recover-all"
  ],
  "capacity": 1,
  "values_by_level": [
    {
      "level": 11,
      "value": 1
    },
    {
      "level": 12,
      "value": 1
    },
    {
      "level": 13,
      "value": 1
    },
    {
      "level": 14,
      "value": 1
    },
    {
      "level": 15,
      "value": 1
    },
    {
      "level": 16,
      "value": 1
    },
    {
      "level": 17,
      "value": 1
    },
    {
      "level": 18,
      "value": 1
    },
    {
      "level": 19,
      "value": 2
    },
    {
      "level": 20,
      "value": 2
    }
  ]
}`)
	// Spell Slot 7, PDF p64.
	add("resource", "/spell-slot-7", "Spell Slot 7", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-7/long-rest-recover-all"
  ],
  "capacity": 1,
  "values_by_level": [
    {
      "level": 13,
      "value": 1
    },
    {
      "level": 14,
      "value": 1
    },
    {
      "level": 15,
      "value": 1
    },
    {
      "level": 16,
      "value": 1
    },
    {
      "level": 17,
      "value": 1
    },
    {
      "level": 18,
      "value": 1
    },
    {
      "level": 19,
      "value": 1
    },
    {
      "level": 20,
      "value": 2
    }
  ]
}`)
	// Spell Slot 8, PDF p64.
	add("resource", "/spell-slot-8", "Spell Slot 8", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-8/long-rest-recover-all"
  ],
  "capacity": 1,
  "values_by_level": [
    {
      "level": 15,
      "value": 1
    },
    {
      "level": 16,
      "value": 1
    },
    {
      "level": 17,
      "value": 1
    },
    {
      "level": 18,
      "value": 1
    },
    {
      "level": 19,
      "value": 1
    },
    {
      "level": 20,
      "value": 1
    }
  ]
}`)
	// Spell Slot 9, PDF p64.
	add("resource", "/spell-slot-9", "Spell Slot 9", "srd-5.2.1:character/feature/sorcerer/spellcasting", "srd-5.2.1:planning/class-features/sorcerer-level-1-spellcasting", 64, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/spell-slot-9/long-rest-recover-all"
  ],
  "capacity": 1,
  "values_by_level": [
    {
      "level": 17,
      "value": 1
    },
    {
      "level": 18,
      "value": 1
    },
    {
      "level": 19,
      "value": 1
    },
    {
      "level": 20,
      "value": 1
    }
  ]
}`)
	// Innate Sorcery, PDF p65.
	add("resource", "/innate-sorcery", "Innate Sorcery", "srd-5.2.1:character/feature/sorcerer/innate-sorcery", "srd-5.2.1:planning/class-features/sorcerer-level-1-innate-sorcery", 65, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/innate-sorcery/long-rest-recover-all"
  ],
  "capacity": 2,
  "values_by_level": [
    {
      "level": 1,
      "value": 2
    },
    {
      "level": 2,
      "value": 2
    },
    {
      "level": 3,
      "value": 2
    },
    {
      "level": 4,
      "value": 2
    },
    {
      "level": 5,
      "value": 2
    },
    {
      "level": 6,
      "value": 2
    },
    {
      "level": 7,
      "value": 2
    },
    {
      "level": 8,
      "value": 2
    },
    {
      "level": 9,
      "value": 2
    },
    {
      "level": 10,
      "value": 2
    },
    {
      "level": 11,
      "value": 2
    },
    {
      "level": 12,
      "value": 2
    },
    {
      "level": 13,
      "value": 2
    },
    {
      "level": 14,
      "value": 2
    },
    {
      "level": 15,
      "value": 2
    },
    {
      "level": 16,
      "value": 2
    },
    {
      "level": 17,
      "value": 2
    },
    {
      "level": 18,
      "value": 2
    },
    {
      "level": 19,
      "value": 2
    },
    {
      "level": 20,
      "value": 2
    }
  ]
}`)
	// Sorcery Points, PDF p66.
	add("resource", "/sorcery-points", "Sorcery Points", "srd-5.2.1:character/feature/sorcerer/font-of-magic", "srd-5.2.1:planning/class-features/sorcerer-level-2-font-of-magic", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/sorcery-points/long-rest-recover-all"
  ],
  "capacity": 2,
  "values_by_level": [
    {
      "level": 2,
      "value": 2
    },
    {
      "level": 3,
      "value": 3
    },
    {
      "level": 4,
      "value": 4
    },
    {
      "level": 5,
      "value": 5
    },
    {
      "level": 6,
      "value": 6
    },
    {
      "level": 7,
      "value": 7
    },
    {
      "level": 8,
      "value": 8
    },
    {
      "level": 9,
      "value": 9
    },
    {
      "level": 10,
      "value": 10
    },
    {
      "level": 11,
      "value": 11
    },
    {
      "level": 12,
      "value": 12
    },
    {
      "level": 13,
      "value": 13
    },
    {
      "level": 14,
      "value": 14
    },
    {
      "level": 15,
      "value": 15
    },
    {
      "level": 16,
      "value": 16
    },
    {
      "level": 17,
      "value": 17
    },
    {
      "level": 18,
      "value": 18
    },
    {
      "level": 19,
      "value": 19
    },
    {
      "level": 20,
      "value": 20
    }
  ]
}`)
	// Metamagic, PDF p66.
	add("choice", "/metamagic", "Metamagic", "srd-5.2.1:character/feature/sorcerer/metamagic", "srd-5.2.1:planning/class-features/sorcerer-level-2-metamagic", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/metamagic"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/metamagic/options"
  ],
  "values_by_level": [
    {
      "level": 2,
      "value": 2
    },
    {
      "level": 3,
      "value": 2
    },
    {
      "level": 4,
      "value": 2
    },
    {
      "level": 5,
      "value": 2
    },
    {
      "level": 6,
      "value": 2
    },
    {
      "level": 7,
      "value": 2
    },
    {
      "level": 8,
      "value": 2
    },
    {
      "level": 9,
      "value": 2
    },
    {
      "level": 10,
      "value": 4
    },
    {
      "level": 11,
      "value": 4
    },
    {
      "level": 12,
      "value": 4
    },
    {
      "level": 13,
      "value": 4
    },
    {
      "level": 14,
      "value": 4
    },
    {
      "level": 15,
      "value": 4
    },
    {
      "level": 16,
      "value": 4
    },
    {
      "level": 17,
      "value": 6
    },
    {
      "level": 18,
      "value": 6
    },
    {
      "level": 19,
      "value": 6
    },
    {
      "level": 20,
      "value": 6
    }
  ]
}`)
	// Subclass, PDF p66.
	add("choice", "/subclass", "Subclass", "srd-5.2.1:character/feature/sorcerer/sorcerer-subclass", "srd-5.2.1:planning/class-features/sorcerer-level-3-sorcerer-subclass", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/subclass",
    "srd-5.2.1:mechanic/sorcerer/subclass/features-at-own-level-or-lower"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/subclasses/sorcerer",
    "srd-5.2.1:mechanic/subclasses/sorcerer/draconic-sorcery"
  ]
}`)
	// Ability Score Improvement, PDF p66.
	add("choice", "/ability-score-improvement", "Ability Score Improvement", "srd-5.2.1:character/feature/sorcerer/ability-score-improvement", "srd-5.2.1:planning/class-features/sorcerer-level-4-ability-score-improvement", 66, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/sorcerer/ability-score-improvement-qualified"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/ability-score-improvement"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/feats/ability-score-improvement",
    "srd-5.2.1:mechanic/feats/qualified"
  ]
}`)
	// Ability Score Improvement Qualified, PDF p66.
	add("prerequisite", "/ability-score-improvement-qualified", "Ability Score Improvement Qualified", "srd-5.2.1:character/choice/sorcerer/ability-score-improvement", "srd-5.2.1:planning/class-features/sorcerer-level-4-ability-score-improvement", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/selected-feat-qualified"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/selected-feat-qualified",
    "record_refs": [
      "srd-5.2.1:character/choice/sorcerer/ability-score-improvement"
    ],
    "parameters": {}
  }
}`)
	// Epic Boon, PDF p66.
	add("choice", "/epic-boon", "Epic Boon", "srd-5.2.1:character/feature/sorcerer/epic-boon", "srd-5.2.1:planning/class-features/sorcerer-level-19-epic-boon", 66, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/sorcerer/epic-boon-qualified"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/epic-boon"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/feats/epic-boon",
    "srd-5.2.1:mechanic/feats/qualified"
  ]
}`)
	// Epic Boon Qualified, PDF p66.
	add("prerequisite", "/epic-boon-qualified", "Epic Boon Qualified", "srd-5.2.1:character/choice/sorcerer/epic-boon", "srd-5.2.1:planning/class-features/sorcerer-level-19-epic-boon", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/selected-feat-qualified"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/selected-feat-qualified",
    "record_refs": [
      "srd-5.2.1:character/choice/sorcerer/epic-boon"
    ],
    "parameters": {}
  }
}`)
	// Sorcerous Restoration, PDF p66.
	add("resource", "/sorcerous-restoration", "Sorcerous Restoration", "srd-5.2.1:character/feature/sorcerer/sorcerous-restoration", "srd-5.2.1:planning/class-features/sorcerer-level-5-sorcerous-restoration", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/sorcerous-restoration/long-rest-recover-all"
  ],
  "capacity": 1,
  "values_by_level": [
    {
      "level": 5,
      "value": 1
    },
    {
      "level": 6,
      "value": 1
    },
    {
      "level": 7,
      "value": 1
    },
    {
      "level": 8,
      "value": 1
    },
    {
      "level": 9,
      "value": 1
    },
    {
      "level": 10,
      "value": 1
    },
    {
      "level": 11,
      "value": 1
    },
    {
      "level": 12,
      "value": 1
    },
    {
      "level": 13,
      "value": 1
    },
    {
      "level": 14,
      "value": 1
    },
    {
      "level": 15,
      "value": 1
    },
    {
      "level": 16,
      "value": 1
    },
    {
      "level": 17,
      "value": 1
    },
    {
      "level": 18,
      "value": 1
    },
    {
      "level": 19,
      "value": 1
    },
    {
      "level": 20,
      "value": 1
    }
  ]
}`)
	// Incarnate No Uses, PDF p66.
	add("prerequisite", "/incarnate-no-uses", "Incarnate No Uses", "srd-5.2.1:character/feature/sorcerer/sorcery-incarnate", "srd-5.2.1:planning/class-features/sorcerer-level-7-sorcery-incarnate", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/no-innate-sorcery-uses-left"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/no-innate-sorcery-uses-left",
    "record_refs": [
      "srd-5.2.1:character/feature/sorcerer/innate-sorcery"
    ],
    "parameters": {}
  }
}`)
	// Incarnate Active, PDF p66.
	add("prerequisite", "/incarnate-active", "Incarnate Active", "srd-5.2.1:character/feature/sorcerer/sorcery-incarnate", "srd-5.2.1:planning/class-features/sorcerer-level-7-sorcery-incarnate", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/innate-sorcery-active"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/innate-sorcery-active",
    "record_refs": [
      "srd-5.2.1:character/feature/sorcerer/innate-sorcery",
      "srd-5.2.1:character/feature/sorcerer/metamagic"
    ],
    "parameters": {}
  }
}`)
	// Apotheosis Active, PDF p66.
	add("prerequisite", "/apotheosis-active", "Apotheosis Active", "srd-5.2.1:character/feature/sorcerer/arcane-apotheosis", "srd-5.2.1:planning/class-features/sorcerer-level-20-arcane-apotheosis", 66, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/sorcerer/innate-sorcery-active"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/sorcerer/innate-sorcery-active",
    "record_refs": [
      "srd-5.2.1:character/feature/sorcerer/innate-sorcery",
      "srd-5.2.1:character/feature/sorcerer/metamagic"
    ],
    "parameters": {}
  }
}`)
	return expected
}

func sorcererSourceDifferences(t *testing.T, c CharacterCatalogue) []string {
	t.Helper()
	var differences []string
	if len(c.Records) != len(sorcererSourceAssertions(t)) {
		differences = append(differences, "record count")
	}
	actual := map[string]map[string]any{}
	for _, r := range c.Records {
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(b, &fields); err != nil {
			t.Fatal(err)
		}
		actual[r.ID] = fields
	}
	for _, want := range sorcererSourceAssertions(t) {
		id := want["id"].(string)
		got := actual[id]
		if len(got) != len(want) {
			differences = append(differences, id+"/field-set")
		}
		for field, expected := range want {
			if !reflect.DeepEqual(got[field], expected) {
				differences = append(differences, fmt.Sprintf("%s/%s: got %v, source %v", id, field, got[field], expected))
			}
		}
	}
	return differences
}

func TestSorcererCompleteSource(t *testing.T) {
	c := sorcererCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if d := sorcererSourceDifferences(t, c); len(d) > 0 {
		t.Fatal(strings.Join(d, "\n"))
	}
	// Check exact canonical identities against the independent planning inventory.
	b, err := os.ReadFile("../../../docs/investigations/issue-74-character-inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	var inv struct {
		Records []struct {
			ID   string `json:"id"`
			Page int    `json:"source_page"`
		}
	}
	if err := json.Unmarshal(b, &inv); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{}
	for _, r := range inv.Records {
		if r.ID == "srd-5.2.1:planning/classes/sorcerer" || strings.HasPrefix(r.ID, "srd-5.2.1:planning/class-features/sorcerer-level-") {
			want[r.ID] = r.Page
		}
	}
	got := map[string]int{}
	for _, r := range c.Records {
		if r.Kind == "class" || r.Kind == "feature" {
			got[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
	}
	if len(got) != 11 || !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical identities/pages: %v", got)
	}
	if c.Source.SHA256 != "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87" {
		t.Fatal("pin")
	}
}

func TestSorcererGuardsReadinessComposition(t *testing.T) {
	repo := os.DirFS("../../..")
	c := sorcererCatalogue(t)
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("catalogue enabled support")
	}
	for _, r := range c.Records {
		for _, k := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(k) {
				t.Fatal("qualified", k)
			}
		}
	}
	c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, "srd-5.2.1:mechanic/unknown/future")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("unknown key enabled")
	}
	tests := []struct {
		name   string
		change func(*CharacterCatalogue)
	}{
		{"wrong page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 65 }},
		{"future grant", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Progression != nil {
					c.Records[i].Progression.FeatureRefs = []string{"srd-5.2.1:character/feature/sorcerer/arcane-apotheosis"}
					break
				}
			}
		}},
		{"missing owner link", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"missing operand", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].Requirement.RecordRefs = []string{"srd-5.2.1:character/choice/sorcerer/missing"}
					break
				}
			}
		}},
		{"unrequired number", func(c *CharacterCatalogue) { c.Records[0].Parameters["srd-5.2.1:mechanic/unknown/value"] = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sorcererCatalogue(t)
			tt.change(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("malformed metadata accepted")
			}
		})
	}
	b, err := os.ReadFile("testdata/character-sorcerer.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(string(b), `"kind": "class"`, `"kind": "class", "playable": true`, 1), string(b) + `{}`} {
		if _, err := DecodeCharacterCatalogue(strings.NewReader(bad)); err == nil {
			t.Fatal("strict decode accepted")
		}
	}
	// Composition with existing canonical class metadata stays unavailable and
	// introduces no replacement definitions or executable consumers.
	merged := sorcererCatalogue(t)
	for _, file := range []string{"character-monk.json", "character-ranger.json", "character-advancement.json", "character-fighter.json", "character-cleric.json", "character-barbarian.json", "character-berserker.json", "character-bard.json", "character-college-of-lore.json", "character-druid.json", "character-circle-of-the-land.json", "character-champion.json", "character-life-domain.json", "character-backgrounds.json", "character-species.json", "character-origin-feats.json", "character-general-feats.json", "character-epic-boon-feats.json", "character-fighting-styles.json"} {
		b, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		merged.Records = append(merged.Records, other.Records...)
	}
	if err := merged.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := merged.ValidateReadiness(repo); err == nil {
		t.Fatal("composition qualified")
	}

}

func TestSorcererImportDoesNotPromoteFixedBuilds(t *testing.T) {
	beforeFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	beforeCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	c := sorcererCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	afterFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	afterCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeFighter, afterFighter) || !reflect.DeepEqual(beforeCleric, afterCleric) {
		t.Fatal("catalogue changed fixed build creation")
	}
	for _, background := range []string{"Acolyte", "Soldier", "Guide", "Hermit"} {
		if _, err := SupportedBuildForConcept("Human Sorcerer", background); err == nil {
			t.Fatal("Sorcerer creation became available")
		}
	}
	// Existing fixed paths retain their numeric/projection assertions too.
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	TestCharacterProjection(t)
}

// These contradictions keep valid ownership, existing operands and positive
// numbers. Structural acceptance must not be mistaken for source agreement.
func TestSorcererContrarySourceControls(t *testing.T) {
	tests := []struct {
		name, id string
		change   func(*CharacterRecord)
	}{
		{"existing wrong operand", "prerequisite/incarnate-no-uses", func(r *CharacterRecord) {
			r.Requirement.RecordRefs = []string{"srd-5.2.1:character/feature/sorcerer/spellcasting"}
		}},
		{"missing equipment alternative", "choice/starting-equipment", func(r *CharacterRecord) { r.OptionKeys = r.OptionKeys[:1]; r.Options = r.Options[:1] }},
		{"wrong condition", "prerequisite/apotheosis-active", func(r *CharacterRecord) {
			r.Requirement.Key = "srd-5.2.1:mechanic/sorcerer/innate-sorcery-inactive"
			r.RequiredMechanics = []string{r.Requirement.Key}
		}},
		{"wrong recovery", "resource/innate-sorcery", func(r *CharacterRecord) {
			r.RequiredMechanics = []string{"srd-5.2.1:mechanic/sorcerer/innate-sorcery/short-rest-recover-all"}
		}},
		{"wrong activation cost", "feature/sorcery-incarnate", func(r *CharacterRecord) { r.Parameters["srd-5.2.1:mechanic/sorcerer/sorcery-point-cost"] = 3 }},
		{"wrong duration", "feature/innate-sorcery", func(r *CharacterRecord) { r.Parameters["srd-5.2.1:mechanic/sorcerer/duration-minutes"] = 10 }},
		{"wrong conversion cost", "feature/font-of-magic", func(r *CharacterRecord) {
			r.Parameters["srd-5.2.1:mechanic/sorcerer/create-slot-3-sorcery-point-cost"] = 4
		}},
		{"wrong conversion minimum", "feature/font-of-magic", func(r *CharacterRecord) {
			r.Parameters["srd-5.2.1:mechanic/sorcerer/create-slot-5-minimum-sorcerer-level"] = 7
		}},
		{"wrong table", "resource/spell-slot-9", func(r *CharacterRecord) { r.ValuesByLevel[0].Value = 2; r.Capacity = 2 }},
		{"invented scaling", "feature/arcane-apotheosis", func(r *CharacterRecord) {
			k := "srd-5.2.1:mechanic/sorcerer/invented-level-scaling"
			r.Parameters[k] = 2
			r.RequiredMechanics = append(r.RequiredMechanics, k)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sorcererCatalogue(t)
			parts := strings.SplitN(tt.id, "/", 2)
			found := false
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/"+parts[0]+"/sorcerer/"+parts[1] {
					tt.change(&c.Records[i])
					found = true
				}
			}
			if !found {
				t.Fatal("missing control target")
			}
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("control must be structurally valid: %v", err)
			}
			if len(sorcererSourceDifferences(t, c)) == 0 {
				t.Fatal("source contradiction accepted")
			}
		})
	}
	t.Run("missing active condition", func(t *testing.T) {
		c := sorcererCatalogue(t)
		id := "srd-5.2.1:character/prerequisite/sorcerer/apotheosis-active"
		var records []CharacterRecord
		for _, r := range c.Records {
			if r.ID == id {
				continue
			}
			if r.ID == "srd-5.2.1:character/feature/sorcerer/arcane-apotheosis" {
				r.PrerequisiteRefs = []string{}
			}
			records = append(records, r)
		}
		c.Records = records
		if err := c.Validate(os.DirFS("../../..")); err != nil {
			t.Fatal(err)
		}
		if len(sorcererSourceDifferences(t, c)) == 0 {
			t.Fatal("missing condition accepted")
		}
	})
}

func TestSorcererSchemaAndDuplicateGuards(t *testing.T) {
	for _, mode := range []string{"schema", "pin", "duplicate", "source", "unknown source"} {
		t.Run(mode, func(t *testing.T) {
			c := sorcererCatalogue(t)
			switch mode {
			case "schema":
				c.SchemaVersion = 2
			case "pin":
				c.Source.SHA256 = strings.Repeat("0", 64)
			case "duplicate":
				c.Records = append(c.Records, c.Records[0])
			case "source":
				c.Records[0].Sources[0].InventoryID = "srd-5.2.1:planning/classes/fighter"
			case "unknown source":
				c.Records[0].Sources[0].InventoryID = "srd-5.2.1:planning/classes/invented"
			}
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("invalid catalogue accepted")
			}
		})
	}
	c := sorcererCatalogue(t)
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(string(b), `"schema_version":1`, `"schema_version":1,"unknown":true`, 1), strings.Replace(string(b), `"proficiency_bonus":2`, `"proficiency_bonus":2,"available":true`, 1)} {
		if _, err := DecodeCharacterCatalogue(strings.NewReader(bad)); err == nil {
			t.Fatal("unknown nested/schema field accepted")
		}
	}
}
