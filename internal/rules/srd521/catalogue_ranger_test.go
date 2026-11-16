package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func rangerCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-ranger.json")
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
func rangerSourceAssertions(t *testing.T) []map[string]any {
	t.Helper()
	var expected []map[string]any
	add := func(kind, slug, name, parent, inventory string, page, level int, fields string) {
		r := map[string]any{
			"id":   "srd-5.2.1:character/" + kind + "/ranger" + slug,
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
	// Ranger, PDF p57.
	add("class", "", "Ranger", "", "srd-5.2.1:planning/classes/ranger", 57, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/skills",
    "srd-5.2.1:character/choice/ranger/multiclass-skill",
    "srd-5.2.1:character/choice/ranger/starting-equipment"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/primary-ability/dexterity",
    "srd-5.2.1:mechanic/ranger/primary-ability/wisdom",
    "srd-5.2.1:mechanic/ranger/hit-die-sides",
    "srd-5.2.1:mechanic/saving-throws/strength",
    "srd-5.2.1:mechanic/saving-throws/dexterity",
    "srd-5.2.1:mechanic/weapons/simple",
    "srd-5.2.1:mechanic/weapons/martial",
    "srd-5.2.1:mechanic/armor/light",
    "srd-5.2.1:mechanic/armor/medium",
    "srd-5.2.1:mechanic/armor/shields",
    "srd-5.2.1:mechanic/ranger/level-one/all-core-traits-and-features",
    "srd-5.2.1:mechanic/ranger/multiclass/hit-die-martial-weapons-one-class-skill-light-medium-shields",
    "srd-5.2.1:mechanic/ranger/multiclass/level-one-features-multiclass-slot-rules"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/hit-die-sides": 10
  }
}`)
	// Skills, PDF p57.
	add("choice", "/skills", "Skills", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/skills"
  ],
  "select": 3,
  "option_keys": [
    "srd-5.2.1:mechanic/skills/animal-handling",
    "srd-5.2.1:mechanic/skills/athletics",
    "srd-5.2.1:mechanic/skills/insight",
    "srd-5.2.1:mechanic/skills/investigation",
    "srd-5.2.1:mechanic/skills/nature",
    "srd-5.2.1:mechanic/skills/perception",
    "srd-5.2.1:mechanic/skills/stealth",
    "srd-5.2.1:mechanic/skills/survival"
  ]
}`)
	// Multiclass Skill, PDF p57.
	add("choice", "/multiclass-skill", "Multiclass Skill", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/multiclass-entry"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/multiclass-skill"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/skills/animal-handling",
    "srd-5.2.1:mechanic/skills/athletics",
    "srd-5.2.1:mechanic/skills/insight",
    "srd-5.2.1:mechanic/skills/investigation",
    "srd-5.2.1:mechanic/skills/nature",
    "srd-5.2.1:mechanic/skills/perception",
    "srd-5.2.1:mechanic/skills/stealth",
    "srd-5.2.1:mechanic/skills/survival"
  ]
}`)
	// Multiclass Entry, PDF p57.
	add("prerequisite", "/multiclass-entry", "Multiclass Entry", "srd-5.2.1:character/choice/ranger/multiclass-skill", "srd-5.2.1:planning/classes/ranger", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/multiclass-entry-only"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/multiclass-entry-only",
    "record_refs": [
      "srd-5.2.1:character/class/ranger",
      "srd-5.2.1:character/choice/ranger/multiclass-skill"
    ],
    "parameters": {}
  }
}`)
	// Starting Equipment, PDF p57.
	add("choice", "/starting-equipment", "Starting Equipment", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/starting-equipment",
    "srd-5.2.1:mechanic/equipment/studded-leather-armor",
    "srd-5.2.1:mechanic/equipment/scimitar",
    "srd-5.2.1:mechanic/equipment/shortsword",
    "srd-5.2.1:mechanic/equipment/longbow",
    "srd-5.2.1:mechanic/equipment/arrows",
    "srd-5.2.1:mechanic/equipment/quiver",
    "srd-5.2.1:mechanic/equipment/druidic-focus-sprig-of-mistletoe",
    "srd-5.2.1:mechanic/equipment/explorers-pack",
    "srd-5.2.1:mechanic/currency/gp"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/ranger/equipment/a",
    "srd-5.2.1:mechanic/ranger/equipment/b"
  ],
  "options": [
    {
      "key": "srd-5.2.1:mechanic/ranger/equipment/a",
      "quantities": {
        "srd-5.2.1:mechanic/equipment/studded-leather-armor": 1,
        "srd-5.2.1:mechanic/equipment/scimitar": 1,
        "srd-5.2.1:mechanic/equipment/shortsword": 1,
        "srd-5.2.1:mechanic/equipment/longbow": 1,
        "srd-5.2.1:mechanic/equipment/arrows": 20,
        "srd-5.2.1:mechanic/equipment/quiver": 1,
        "srd-5.2.1:mechanic/equipment/druidic-focus-sprig-of-mistletoe": 1,
        "srd-5.2.1:mechanic/equipment/explorers-pack": 1,
        "srd-5.2.1:mechanic/currency/gp": 7
      }
    },
    {
      "key": "srd-5.2.1:mechanic/ranger/equipment/b",
      "quantities": {
        "srd-5.2.1:mechanic/currency/gp": 150
      }
    }
  ]
}`)
	// Ranger level 1, PDF p57.
	add("level", "/1", "Ranger level 1", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 1, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/spellcasting",
      "srd-5.2.1:character/feature/ranger/favored-enemy",
      "srd-5.2.1:character/feature/ranger/weapon-mastery"
    ],
    "feature_keys": []
  }
}`)
	// Spellcasting, PDF p57.
	add("feature", "/spellcasting", "Spellcasting", "srd-5.2.1:character/level/ranger/1", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/prepared-spells"
  ],
  "resource_refs": [
    "srd-5.2.1:character/resource/ranger/spell-slot-1",
    "srd-5.2.1:character/resource/ranger/spell-slot-2",
    "srd-5.2.1:character/resource/ranger/spell-slot-3",
    "srd-5.2.1:character/resource/ranger/spell-slot-4",
    "srd-5.2.1:character/resource/ranger/spell-slot-5"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spellcasting/wisdom",
    "srd-5.2.1:mechanic/ranger/spellcasting/druidic-focus",
    "srd-5.2.1:mechanic/ranger/spellcasting/always-prepared-excluded-still-ranger-spells",
    "srd-5.2.1:mechanic/ranger/recommended-spell/cure-wounds",
    "srd-5.2.1:mechanic/ranger/recommended-spell/ensnaring-strike",
    "srd-5.2.1:mechanic/ranger/spellcasting/long-rest-replace-one-eligible-spell",
    "srd-5.2.1:mechanic/ranger/prepared-replacements",
    "srd-5.2.1:mechanic/ranger/feature/spellcasting"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/prepared-replacements": 1
  }
}`)
	// Favored Enemy, PDF p58.
	add("feature", "/favored-enemy", "Favored Enemy", "srd-5.2.1:character/level/ranger/1", "srd-5.2.1:planning/class-features/ranger-level-1-favored-enemy", 58, 0, `{
  "resource_refs": [
    "srd-5.2.1:character/resource/ranger/favored-enemy"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/favored-enemy/source-5-2-1-always-prepared-hunters-mark",
    "srd-5.2.1:mechanic/spells/hunters-mark",
    "srd-5.2.1:mechanic/ranger/favored-enemy/no-slot-cast",
    "srd-5.2.1:mechanic/ranger/uses-per-cast",
    "srd-5.2.1:mechanic/ranger/feature/favored-enemy"
  ],
  "always_prepared": [
    {
      "level": 1,
      "spell_keys": [
        "srd-5.2.1:mechanic/spells/hunters-mark"
      ]
    }
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/uses-per-cast": 1
  }
}`)
	// Weapon Mastery, PDF p58.
	add("feature", "/weapon-mastery", "Weapon Mastery", "srd-5.2.1:character/level/ranger/1", "srd-5.2.1:planning/class-features/ranger-level-1-weapon-mastery", 58, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/weapon-mastery"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/mastery/long-rest-replace-kinds",
    "srd-5.2.1:mechanic/ranger/feature/weapon-mastery"
  ]
}`)
	// Ranger level 2, PDF p57.
	add("level", "/2", "Ranger level 2", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 2, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/deft-explorer",
      "srd-5.2.1:character/feature/ranger/fighting-style"
    ],
    "feature_keys": []
  }
}`)
	// Deft Explorer, PDF p59.
	add("feature", "/deft-explorer", "Deft Explorer", "srd-5.2.1:character/level/ranger/2", "srd-5.2.1:planning/class-features/ranger-level-2-deft-explorer", 59, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/deft-explorer-expertise",
    "srd-5.2.1:character/choice/ranger/languages"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/feature/deft-explorer"
  ]
}`)
	// Fighting Style, PDF p59.
	add("feature", "/fighting-style", "Fighting Style", "srd-5.2.1:character/level/ranger/2", "srd-5.2.1:planning/class-features/ranger-level-2-fighting-style", 59, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/fighting-style"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/druidic-warrior/replace-one-on-ranger-level",
    "srd-5.2.1:mechanic/ranger/cantrip-replacements",
    "srd-5.2.1:mechanic/ranger/feature/fighting-style"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/cantrip-replacements": 1
  }
}`)
	// Ranger level 3, PDF p57.
	add("level", "/3", "Ranger level 3", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 3, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/ranger-subclass"
    ],
    "feature_keys": []
  }
}`)
	// Ranger Subclass, PDF p59.
	add("feature", "/ranger-subclass", "Ranger Subclass", "srd-5.2.1:character/level/ranger/3", "srd-5.2.1:planning/class-features/ranger-level-3-ranger-subclass", 59, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/subclass"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/subclass/features-at-or-below-ranger-level",
    "srd-5.2.1:mechanic/ranger/feature/ranger-subclass"
  ]
}`)
	// Ranger level 4, PDF p57.
	add("level", "/4", "Ranger level 4", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 4, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 2,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Ability Score Improvement, PDF p59.
	add("feature", "/ability-score-improvement", "Ability Score Improvement", "srd-5.2.1:character/level/ranger/4", "srd-5.2.1:planning/class-features/ranger-level-4-ability-score-improvement", 59, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/ability-score-improvement"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/asi-repeat-at-8-12-16",
    "srd-5.2.1:mechanic/ranger/feature/ability-score-improvement"
  ]
}`)
	// Ranger level 5, PDF p57.
	add("level", "/5", "Ranger level 5", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 5, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/extra-attack"
    ],
    "feature_keys": []
  }
}`)
	// Extra Attack, PDF p59.
	add("feature", "/extra-attack", "Extra Attack", "srd-5.2.1:character/level/ranger/5", "srd-5.2.1:planning/class-features/ranger-level-5-extra-attack", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/extra-attack/own-turn-attack-action",
    "srd-5.2.1:mechanic/ranger/attacks",
    "srd-5.2.1:mechanic/ranger/feature/extra-attack"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/attacks": 2
  }
}`)
	// Ranger level 6, PDF p57.
	add("level", "/6", "Ranger level 6", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 6, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/roving"
    ],
    "feature_keys": []
  }
}`)
	// Roving, PDF p59.
	add("feature", "/roving", "Roving", "srd-5.2.1:character/level/ranger/6", "srd-5.2.1:planning/class-features/ranger-level-6-roving", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/roving-armor"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/roving/speed-increase-when-not-heavy-armor",
    "srd-5.2.1:mechanic/ranger/speed-bonus-feet",
    "srd-5.2.1:mechanic/ranger/roving/climb-and-swim-equal-speed",
    "srd-5.2.1:mechanic/ranger/feature/roving"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/speed-bonus-feet": 10
  }
}`)
	// Ranger level 7, PDF p57.
	add("level", "/7", "Ranger level 7", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 7, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression",
    "srd-5.2.1:mechanic/ranger/subclass-feature"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [],
    "feature_keys": [
      "srd-5.2.1:mechanic/ranger/subclass-feature"
    ]
  }
}`)
	// Ranger level 8, PDF p57.
	add("level", "/8", "Ranger level 8", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 8, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 3,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Ranger level 9, PDF p57.
	add("level", "/9", "Ranger level 9", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 9, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/expertise"
    ],
    "feature_keys": []
  }
}`)
	// Expertise, PDF p59.
	add("feature", "/expertise", "Expertise", "srd-5.2.1:character/level/ranger/9", "srd-5.2.1:planning/class-features/ranger-level-9-expertise", 59, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/expertise-expertise"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/feature/expertise"
  ]
}`)
	// Ranger level 10, PDF p57.
	add("level", "/10", "Ranger level 10", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 10, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/tireless"
    ],
    "feature_keys": []
  }
}`)
	// Tireless, PDF p59.
	add("feature", "/tireless", "Tireless", "srd-5.2.1:character/level/ranger/10", "srd-5.2.1:planning/class-features/ranger-level-10-tireless", 59, 0, `{
  "resource_refs": [
    "srd-5.2.1:character/resource/ranger/tireless"
  ],
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/tireless-exhaustion"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/tireless/magic-action-self-temporary-hp",
    "srd-5.2.1:mechanic/ranger/tireless/temp-hp-die-plus-wisdom-minimum-one",
    "srd-5.2.1:mechanic/ranger/temporary-hp-dice",
    "srd-5.2.1:mechanic/ranger/temporary-hp-die-sides",
    "srd-5.2.1:mechanic/ranger/temporary-hp-minimum",
    "srd-5.2.1:mechanic/ranger/tireless/short-rest-decrease-exhaustion-if-any",
    "srd-5.2.1:mechanic/ranger/exhaustion-reduction",
    "srd-5.2.1:mechanic/ranger/feature/tireless"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/temporary-hp-dice": 1,
    "srd-5.2.1:mechanic/ranger/temporary-hp-die-sides": 8,
    "srd-5.2.1:mechanic/ranger/temporary-hp-minimum": 1,
    "srd-5.2.1:mechanic/ranger/exhaustion-reduction": 1
  }
}`)
	// Ranger level 11, PDF p57.
	add("level", "/11", "Ranger level 11", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 11, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression",
    "srd-5.2.1:mechanic/ranger/subclass-feature"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [],
    "feature_keys": [
      "srd-5.2.1:mechanic/ranger/subclass-feature"
    ]
  }
}`)
	// Ranger level 12, PDF p57.
	add("level", "/12", "Ranger level 12", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 12, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 4,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Ranger level 13, PDF p57.
	add("level", "/13", "Ranger level 13", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 13, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/relentless-hunter"
    ],
    "feature_keys": []
  }
}`)
	// Relentless Hunter, PDF p59.
	add("feature", "/relentless-hunter", "Relentless Hunter", "srd-5.2.1:character/level/ranger/13", "srd-5.2.1:planning/class-features/ranger-level-13-relentless-hunter", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/relentless-mark"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/relentless-hunter/damage-cannot-break-hunters-mark-concentration",
    "srd-5.2.1:mechanic/ranger/feature/relentless-hunter"
  ]
}`)
	// Ranger level 14, PDF p57.
	add("level", "/14", "Ranger level 14", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 14, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/natures-veil"
    ],
    "feature_keys": []
  }
}`)
	// Nature’s Veil, PDF p59.
	add("feature", "/natures-veil", "Nature’s Veil", "srd-5.2.1:character/level/ranger/14", "srd-5.2.1:planning/class-features/ranger-level-14-natures-veil", 59, 0, `{
  "resource_refs": [
    "srd-5.2.1:character/resource/ranger/natures-veil"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/veil/bonus-action-self-invisible-until-end-next-turn",
    "srd-5.2.1:mechanic/ranger/feature/natures-veil"
  ]
}`)
	// Ranger level 15, PDF p57.
	add("level", "/15", "Ranger level 15", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 15, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression",
    "srd-5.2.1:mechanic/ranger/subclass-feature"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [],
    "feature_keys": [
      "srd-5.2.1:mechanic/ranger/subclass-feature"
    ]
  }
}`)
	// Ranger level 16, PDF p57.
	add("level", "/16", "Ranger level 16", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 16, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 5,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/ability-score-improvement"
    ],
    "feature_keys": []
  }
}`)
	// Ranger level 17, PDF p57.
	add("level", "/17", "Ranger level 17", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 17, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/precise-hunter"
    ],
    "feature_keys": []
  }
}`)
	// Precise Hunter, PDF p59.
	add("feature", "/precise-hunter", "Precise Hunter", "srd-5.2.1:character/level/ranger/17", "srd-5.2.1:planning/class-features/ranger-level-17-precise-hunter", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/precise-mark"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/precise-hunter/advantage-on-attacks-against-own-current-mark",
    "srd-5.2.1:mechanic/ranger/feature/precise-hunter"
  ]
}`)
	// Ranger level 18, PDF p57.
	add("level", "/18", "Ranger level 18", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 18, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/feral-senses"
    ],
    "feature_keys": []
  }
}`)
	// Feral Senses, PDF p59.
	add("feature", "/feral-senses", "Feral Senses", "srd-5.2.1:character/level/ranger/18", "srd-5.2.1:planning/class-features/ranger-level-18-feral-senses", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/feral-senses/blindsight",
    "srd-5.2.1:mechanic/ranger/blindsight-feet",
    "srd-5.2.1:mechanic/ranger/feature/feral-senses"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/blindsight-feet": 30
  }
}`)
	// Ranger level 19, PDF p57.
	add("level", "/19", "Ranger level 19", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 19, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/epic-boon"
    ],
    "feature_keys": []
  }
}`)
	// Epic Boon, PDF p59.
	add("feature", "/epic-boon", "Epic Boon", "srd-5.2.1:character/level/ranger/19", "srd-5.2.1:planning/class-features/ranger-level-19-epic-boon", 59, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/epic-boon"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/recommended-feat/boon-of-dimensional-travel",
    "srd-5.2.1:mechanic/ranger/feature/epic-boon"
  ]
}`)
	// Ranger level 20, PDF p57.
	add("level", "/20", "Ranger level 20", "srd-5.2.1:character/class/ranger", "srd-5.2.1:planning/classes/ranger", 57, 20, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/class-progression"
  ],
  "progression": {
    "proficiency_bonus": 6,
    "feature_refs": [
      "srd-5.2.1:character/feature/ranger/foe-slayer"
    ],
    "feature_keys": []
  }
}`)
	// Foe Slayer, PDF p59.
	add("feature", "/foe-slayer", "Foe Slayer", "srd-5.2.1:character/level/ranger/20", "srd-5.2.1:planning/class-features/ranger-level-20-foe-slayer", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/foe-slayer-mark"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/foe-slayer/replace-hunters-mark-damage-die",
    "srd-5.2.1:mechanic/ranger/original-die-sides",
    "srd-5.2.1:mechanic/ranger/replacement-die-sides",
    "srd-5.2.1:mechanic/ranger/feature/foe-slayer"
  ],
  "parameters": {
    "srd-5.2.1:mechanic/ranger/original-die-sides": 6,
    "srd-5.2.1:mechanic/ranger/replacement-die-sides": 10
  }
}`)
	// Prepared Spells, PDF p57.
	add("choice", "/prepared-spells", "Prepared Spells", "srd-5.2.1:character/feature/ranger/spellcasting", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/prepared-slot-level"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spellcasting/prepare-to-table-count-on-increase"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/spells/ranger-list"
  ],
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
      "value": 5
    },
    {
      "level": 5,
      "value": 6
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
      "value": 7
    },
    {
      "level": 9,
      "value": 9
    },
    {
      "level": 10,
      "value": 9
    },
    {
      "level": 11,
      "value": 10
    },
    {
      "level": 12,
      "value": 10
    },
    {
      "level": 13,
      "value": 11
    },
    {
      "level": 14,
      "value": 11
    },
    {
      "level": 15,
      "value": 12
    },
    {
      "level": 16,
      "value": 12
    },
    {
      "level": 17,
      "value": 14
    },
    {
      "level": 18,
      "value": 14
    },
    {
      "level": 19,
      "value": 15
    },
    {
      "level": 20,
      "value": 15
    }
  ]
}`)
	// Prepared Slot Level, PDF p57.
	add("prerequisite", "/prepared-slot-level", "Prepared Slot Level", "srd-5.2.1:character/choice/ranger/prepared-spells", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spell-level-with-available-slot"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/spell-level-with-available-slot",
    "record_refs": [
      "srd-5.2.1:character/feature/ranger/spellcasting",
      "srd-5.2.1:character/choice/ranger/prepared-spells"
    ],
    "parameters": {
      "minimum-spell-level": 1
    }
  }
}`)
	// Spell Slot 1, PDF p57.
	add("resource", "/spell-slot-1", "Spell Slot 1", "srd-5.2.1:character/feature/ranger/spellcasting", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spell-slot-1/long-rest-recover-all"
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
      "value": 3
    },
    {
      "level": 4,
      "value": 3
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
	// Spell Slot 2, PDF p57.
	add("resource", "/spell-slot-2", "Spell Slot 2", "srd-5.2.1:character/feature/ranger/spellcasting", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spell-slot-2/long-rest-recover-all"
  ],
  "capacity": 2,
  "values_by_level": [
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
	// Spell Slot 3, PDF p57.
	add("resource", "/spell-slot-3", "Spell Slot 3", "srd-5.2.1:character/feature/ranger/spellcasting", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spell-slot-3/long-rest-recover-all"
  ],
  "capacity": 2,
  "values_by_level": [
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
	// Spell Slot 4, PDF p57.
	add("resource", "/spell-slot-4", "Spell Slot 4", "srd-5.2.1:character/feature/ranger/spellcasting", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spell-slot-4/long-rest-recover-all"
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
      "value": 2
    },
    {
      "level": 16,
      "value": 2
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
	// Spell Slot 5, PDF p57.
	add("resource", "/spell-slot-5", "Spell Slot 5", "srd-5.2.1:character/feature/ranger/spellcasting", "srd-5.2.1:planning/class-features/ranger-level-1-spellcasting", 57, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/spell-slot-5/long-rest-recover-all"
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
      "value": 2
    },
    {
      "level": 20,
      "value": 2
    }
  ]
}`)
	// Favored Enemy, PDF p58.
	add("resource", "/favored-enemy", "Favored Enemy", "srd-5.2.1:character/feature/ranger/favored-enemy", "srd-5.2.1:planning/class-features/ranger-level-1-favored-enemy", 58, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/favored-enemy/long-rest-recover-all"
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
      "value": 5
    },
    {
      "level": 14,
      "value": 5
    },
    {
      "level": 15,
      "value": 5
    },
    {
      "level": 16,
      "value": 5
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
	// Weapon Mastery, PDF p58.
	add("choice", "/weapon-mastery", "Weapon Mastery", "srd-5.2.1:character/feature/ranger/weapon-mastery", "srd-5.2.1:planning/class-features/ranger-level-1-weapon-mastery", 58, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/mastery-proficiency"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/weapon-mastery"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/weapons/proficient-kinds"
  ]
}`)
	// Mastery Proficiency, PDF p58.
	add("prerequisite", "/mastery-proficiency", "Mastery Proficiency", "srd-5.2.1:character/choice/ranger/weapon-mastery", "srd-5.2.1:planning/class-features/ranger-level-1-weapon-mastery", 58, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/selected-weapon-proficient"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/selected-weapon-proficient",
    "record_refs": [
      "srd-5.2.1:character/choice/ranger/weapon-mastery"
    ],
    "parameters": {}
  }
}`)
	// Deft Explorer Expertise, PDF p59.
	add("choice", "/deft-explorer-expertise", "Deft Explorer Expertise", "srd-5.2.1:character/feature/ranger/deft-explorer", "srd-5.2.1:planning/class-features/ranger-level-2-deft-explorer", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/deft-explorer-eligibility"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/deft-explorer-expertise"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/skills/proficient-without-expertise"
  ]
}`)
	// Deft Explorer Eligibility, PDF p59.
	add("prerequisite", "/deft-explorer-eligibility", "Deft Explorer Eligibility", "srd-5.2.1:character/choice/ranger/deft-explorer-expertise", "srd-5.2.1:planning/class-features/ranger-level-2-deft-explorer", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/proficient-skill-lacking-expertise"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/proficient-skill-lacking-expertise",
    "record_refs": [
      "srd-5.2.1:character/choice/ranger/deft-explorer-expertise"
    ],
    "parameters": {}
  }
}`)
	// Expertise Expertise, PDF p59.
	add("choice", "/expertise-expertise", "Expertise Expertise", "srd-5.2.1:character/feature/ranger/expertise", "srd-5.2.1:planning/class-features/ranger-level-9-expertise", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/expertise-eligibility"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/expertise-expertise"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/skills/proficient-without-expertise"
  ]
}`)
	// Expertise Eligibility, PDF p59.
	add("prerequisite", "/expertise-eligibility", "Expertise Eligibility", "srd-5.2.1:character/choice/ranger/expertise-expertise", "srd-5.2.1:planning/class-features/ranger-level-9-expertise", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/proficient-skill-lacking-expertise"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/proficient-skill-lacking-expertise",
    "record_refs": [
      "srd-5.2.1:character/choice/ranger/expertise-expertise"
    ],
    "parameters": {}
  }
}`)
	// Languages, PDF p59.
	add("choice", "/languages", "Languages", "srd-5.2.1:character/feature/ranger/deft-explorer", "srd-5.2.1:planning/class-features/ranger-level-2-deft-explorer", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/languages"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/languages/character-creation-tables"
  ]
}`)
	// Fighting Style, PDF p59.
	add("choice", "/fighting-style", "Fighting Style", "srd-5.2.1:character/feature/ranger/fighting-style", "srd-5.2.1:planning/class-features/ranger-level-2-fighting-style", 59, 0, `{
  "choice_refs": [
    "srd-5.2.1:character/choice/ranger/druidic-warrior-cantrips"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/fighting-style"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/feats/fighting-style",
    "srd-5.2.1:mechanic/ranger/druidic-warrior"
  ]
}`)
	// Druidic Warrior Cantrips, PDF p59.
	add("choice", "/druidic-warrior-cantrips", "Druidic Warrior Cantrips", "srd-5.2.1:character/choice/ranger/fighting-style", "srd-5.2.1:planning/class-features/ranger-level-2-fighting-style", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/druidic-warrior-selected"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/druidic-warrior-cantrips",
    "srd-5.2.1:mechanic/ranger/druidic-warrior/count-as-ranger-spells-wisdom",
    "srd-5.2.1:mechanic/ranger/recommended-cantrip/guidance",
    "srd-5.2.1:mechanic/ranger/recommended-cantrip/starry-wisp"
  ],
  "select": 2,
  "option_keys": [
    "srd-5.2.1:mechanic/spells/druid-cantrips"
  ]
}`)
	// Druidic Warrior Selected, PDF p59.
	add("prerequisite", "/druidic-warrior-selected", "Druidic Warrior Selected", "srd-5.2.1:character/choice/ranger/druidic-warrior-cantrips", "srd-5.2.1:planning/class-features/ranger-level-2-fighting-style", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/druidic-warrior-selected"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/druidic-warrior-selected",
    "record_refs": [
      "srd-5.2.1:character/choice/ranger/fighting-style"
    ],
    "parameters": {}
  }
}`)
	// Subclass, PDF p59.
	add("choice", "/subclass", "Subclass", "srd-5.2.1:character/feature/ranger/ranger-subclass", "srd-5.2.1:planning/class-features/ranger-level-3-ranger-subclass", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/subclass"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/subclasses/ranger",
    "srd-5.2.1:mechanic/subclasses/ranger/hunter"
  ]
}`)
	// Ability Score Improvement, PDF p59.
	add("choice", "/ability-score-improvement", "Ability Score Improvement", "srd-5.2.1:character/feature/ranger/ability-score-improvement", "srd-5.2.1:planning/class-features/ranger-level-4-ability-score-improvement", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/ability-score-improvement-qualified"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/ability-score-improvement"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/feats/ability-score-improvement",
    "srd-5.2.1:mechanic/feats/qualified"
  ]
}`)
	// Ability Score Improvement Qualified, PDF p59.
	add("prerequisite", "/ability-score-improvement-qualified", "Ability Score Improvement Qualified", "srd-5.2.1:character/choice/ranger/ability-score-improvement", "srd-5.2.1:planning/class-features/ranger-level-4-ability-score-improvement", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/selected-feat-qualified"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/selected-feat-qualified",
    "record_refs": [
      "srd-5.2.1:character/choice/ranger/ability-score-improvement"
    ],
    "parameters": {}
  }
}`)
	// Epic Boon, PDF p59.
	add("choice", "/epic-boon", "Epic Boon", "srd-5.2.1:character/feature/ranger/epic-boon", "srd-5.2.1:planning/class-features/ranger-level-19-epic-boon", 59, 0, `{
  "prerequisite_refs": [
    "srd-5.2.1:character/prerequisite/ranger/epic-boon-qualified"
  ],
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/epic-boon"
  ],
  "select": 1,
  "option_keys": [
    "srd-5.2.1:mechanic/feats/epic-boon",
    "srd-5.2.1:mechanic/feats/qualified"
  ]
}`)
	// Epic Boon Qualified, PDF p59.
	add("prerequisite", "/epic-boon-qualified", "Epic Boon Qualified", "srd-5.2.1:character/choice/ranger/epic-boon", "srd-5.2.1:planning/class-features/ranger-level-19-epic-boon", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/selected-feat-qualified"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/selected-feat-qualified",
    "record_refs": [
      "srd-5.2.1:character/choice/ranger/epic-boon"
    ],
    "parameters": {}
  }
}`)
	// Roving Armor, PDF p59.
	add("prerequisite", "/roving-armor", "Roving Armor", "srd-5.2.1:character/feature/ranger/roving", "srd-5.2.1:planning/class-features/ranger-level-6-roving", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/not-wearing-heavy-armor"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/not-wearing-heavy-armor",
    "record_refs": [
      "srd-5.2.1:character/feature/ranger/roving"
    ],
    "parameters": {}
  }
}`)
	// Tireless, PDF p59.
	add("resource", "/tireless", "Tireless", "srd-5.2.1:character/feature/ranger/tireless", "srd-5.2.1:planning/class-features/ranger-level-10-tireless", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/tireless/long-rest-recover-all",
    "srd-5.2.1:mechanic/ranger/wisdom-modifier-minimum-one"
  ],
  "capacity_mechanic": "srd-5.2.1:mechanic/ranger/wisdom-modifier-minimum-one"
}`)
	// Tireless Exhaustion, PDF p59.
	add("prerequisite", "/tireless-exhaustion", "Tireless Exhaustion", "srd-5.2.1:character/feature/ranger/tireless", "srd-5.2.1:planning/class-features/ranger-level-10-tireless", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/exhaustion-present"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/exhaustion-present",
    "record_refs": [
      "srd-5.2.1:character/feature/ranger/tireless"
    ],
    "parameters": {
      "minimum-exhaustion": 1
    }
  }
}`)
	// Relentless Mark, PDF p59.
	add("prerequisite", "/relentless-mark", "Relentless Mark", "srd-5.2.1:character/feature/ranger/relentless-hunter", "srd-5.2.1:planning/class-features/ranger-level-13-relentless-hunter", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/concentrating-on-hunters-mark"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/concentrating-on-hunters-mark",
    "record_refs": [
      "srd-5.2.1:character/feature/ranger/favored-enemy"
    ],
    "parameters": {}
  }
}`)
	// Natures Veil, PDF p59.
	add("resource", "/natures-veil", "Natures Veil", "srd-5.2.1:character/feature/ranger/natures-veil", "srd-5.2.1:planning/class-features/ranger-level-14-natures-veil", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/natures-veil/long-rest-recover-all",
    "srd-5.2.1:mechanic/ranger/wisdom-modifier-minimum-one"
  ],
  "capacity_mechanic": "srd-5.2.1:mechanic/ranger/wisdom-modifier-minimum-one"
}`)
	// Precise Mark, PDF p59.
	add("prerequisite", "/precise-mark", "Precise Mark", "srd-5.2.1:character/feature/ranger/precise-hunter", "srd-5.2.1:planning/class-features/ranger-level-17-precise-hunter", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/target-is-own-current-hunters-mark"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/target-is-own-current-hunters-mark",
    "record_refs": [
      "srd-5.2.1:character/feature/ranger/favored-enemy"
    ],
    "parameters": {}
  }
}`)
	// Foe Slayer Mark, PDF p59.
	add("prerequisite", "/foe-slayer-mark", "Foe Slayer Mark", "srd-5.2.1:character/feature/ranger/foe-slayer", "srd-5.2.1:planning/class-features/ranger-level-20-foe-slayer", 59, 0, `{
  "required_mechanics": [
    "srd-5.2.1:mechanic/ranger/hunters-mark-damage"
  ],
  "requirement": {
    "key": "srd-5.2.1:mechanic/ranger/hunters-mark-damage",
    "record_refs": [
      "srd-5.2.1:character/feature/ranger/favored-enemy"
    ],
    "parameters": {}
  }
}`)
	return expected
}

func rangerSourceDifferences(t *testing.T, c CharacterCatalogue) []string {
	t.Helper()
	var differences []string
	if len(c.Records) != len(rangerSourceAssertions(t)) {
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
	for _, want := range rangerSourceAssertions(t) {
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

func TestRangerCompleteSource(t *testing.T) {
	c := rangerCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if d := rangerSourceDifferences(t, c); len(d) > 0 {
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
		if r.ID == "srd-5.2.1:planning/classes/ranger" || strings.HasPrefix(r.ID, "srd-5.2.1:planning/class-features/ranger-level-") {
			want[r.ID] = r.Page
		}
	}
	got := map[string]int{}
	for _, r := range c.Records {
		if r.Kind == "class" || r.Kind == "feature" {
			got[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
	}
	if len(got) != 18 || !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical identities/pages: %v", got)
	}
	if c.Source.SHA256 != "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87" {
		t.Fatal("pin")
	}
}

func TestRangerGuardsReadinessComposition(t *testing.T) {
	repo := os.DirFS("../../..")
	c := rangerCatalogue(t)
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
		{"wrong page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 58 }},
		{"future grant", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Progression != nil {
					c.Records[i].Progression.FeatureRefs = []string{"srd-5.2.1:character/feature/ranger/foe-slayer"}
					break
				}
			}
		}},
		{"missing owner link", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"missing operand", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].Requirement.RecordRefs = []string{"srd-5.2.1:character/choice/ranger/missing"}
					break
				}
			}
		}},
		{"unrequired number", func(c *CharacterCatalogue) { c.Records[0].Parameters["srd-5.2.1:mechanic/unknown/value"] = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := rangerCatalogue(t)
			tt.change(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("malformed metadata accepted")
			}
		})
	}
	b, err := os.ReadFile("testdata/character-ranger.json")
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
	merged := rangerCatalogue(t)
	for _, file := range []string{"character-advancement.json", "character-fighter.json", "character-cleric.json", "character-barbarian.json", "character-berserker.json", "character-bard.json", "character-college-of-lore.json", "character-druid.json", "character-circle-of-the-land.json", "character-champion.json", "character-life-domain.json", "character-backgrounds.json", "character-species.json", "character-origin-feats.json", "character-general-feats.json", "character-epic-boon-feats.json", "character-fighting-styles.json"} {
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

func TestRangerImportDoesNotPromoteFixedBuilds(t *testing.T) {
	beforeFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	beforeCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	c := rangerCatalogue(t)
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
		if _, err := SupportedBuildForConcept("Human Ranger", background); err == nil {
			t.Fatal("Ranger creation became available")
		}
	}
	// Existing fixed paths retain their numeric/projection assertions too.
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	TestCharacterProjection(t)
}

// All controls retain valid graph links, positive numbers and opaque keys.
// Structural validation must pass before the semantic oracle rejects them.
func TestRangerContrarySourceControls(t *testing.T) {
	tests := []struct {
		name, id string
		change   func(*CharacterRecord)
	}{
		{"existing wrong operand", "prerequisite/precise-mark", func(r *CharacterRecord) {
			r.Requirement.RecordRefs = []string{"srd-5.2.1:character/feature/ranger/spellcasting"}
		}},
		{"wrong condition", "prerequisite/roving-armor", func(r *CharacterRecord) {
			r.Requirement.Key = "srd-5.2.1:mechanic/ranger/wearing-heavy-armor"
			r.RequiredMechanics = []string{r.Requirement.Key}
		}},
		{"wrong recovery", "resource/tireless", func(r *CharacterRecord) {
			r.RequiredMechanics[0] = "srd-5.2.1:mechanic/ranger/tireless/short-rest-recover-all"
		}},
		{"wrong cost", "feature/favored-enemy", func(r *CharacterRecord) { r.Parameters["srd-5.2.1:mechanic/ranger/uses-per-cast"] = 2 }},
		{"wrong range", "feature/feral-senses", func(r *CharacterRecord) { r.Parameters["srd-5.2.1:mechanic/ranger/blindsight-feet"] = 60 }},
		{"wrong table", "resource/favored-enemy", func(r *CharacterRecord) { r.ValuesByLevel[1].Value = 3 }},
		{"wrong variant", "feature/favored-enemy", func(r *CharacterRecord) {
			r.RequiredMechanics[0] = "srd-5.2.1:mechanic/ranger/favored-enemy/2014-creature-type-variant"
		}},
		{"invented scaling", "feature/foe-slayer", func(r *CharacterRecord) {
			k := "srd-5.2.1:mechanic/ranger/invented-level-scaling"
			r.RequiredMechanics = append(r.RequiredMechanics, k)
			r.Parameters[k] = 2
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := rangerCatalogue(t)
			found := false
			parts := strings.SplitN(tt.id, "/", 2)
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/"+parts[0]+"/ranger/"+parts[1] {
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
			if len(rangerSourceDifferences(t, c)) == 0 {
				t.Fatal("source contradiction accepted")
			}
		})
	}
}

func TestRangerSymbolicCapacityGuards(t *testing.T) {
	for _, mode := range []string{"unknown formula", "fixed plus formula", "scaling plus formula", "not required", "nonresource"} {
		t.Run(mode, func(t *testing.T) {
			c := rangerCatalogue(t)
			for i := range c.Records {
				r := &c.Records[i]
				if r.ID != "srd-5.2.1:character/resource/ranger/tireless" {
					continue
				}
				switch mode {
				case "unknown formula":
					r.CapacityMechanic = "srd-5.2.1:mechanic/ranger/unbounded-wisdom"
				case "fixed plus formula":
					r.Capacity = 1
				case "scaling plus formula":
					r.ValuesByLevel = []CharacterLevelValue{{Level: 10, Value: 1}}
				case "not required":
					r.RequiredMechanics = r.RequiredMechanics[:1]
				case "nonresource":
					r.Kind = "feature"
				}
			}
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("invalid symbolic capacity accepted")
			}
		})
	}
}
