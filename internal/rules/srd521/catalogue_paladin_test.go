package srd521

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

const pc = "srd-5.2.1:character/"
const pm = "srd-5.2.1:mechanic/"

// Frozen source graph: every row, link, predicate, alternative and quantity.
// Expectations never load the imported JSON or invoke the PDF comparator.
// Empty fields are deliberate assertions too.
var paladinSourceGraph = []CharacterRecord{
	{
		ID:   pc + "class/paladin",
		Kind: "class",
		Name: "Paladin",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID: "",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/skills",
			pc + "choice/paladin/starting-equipment",
		},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/primary-ability/strength",
			pm + "paladin/primary-ability/charisma",
			pm + "paladin/hit-die-sides",
			pm + "saving-throws/wisdom",
			pm + "saving-throws/charisma",
			pm + "weapons/simple",
			pm + "weapons/martial",
			pm + "armor/light",
			pm + "armor/medium",
			pm + "armor/heavy",
			pm + "armor/shields",
			pm + "paladin/level-1/all-core-traits-and-features",
			pm + "paladin/multiclass/hit-die-martial-weapons-light-medium-armor-shields-only",
			pm + "paladin/multiclass/level-1-features-and-multiclass-spell-slots",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/hit-die-sides": 10},
	},
	{
		ID:   pc + "choice/paladin/skills",
		Kind: "choice",
		Name: "Skills",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/skills",
		},
		Select: 2,
		OptionKeys: []string{
			pm + "skills/athletics",
			pm + "skills/insight",
			pm + "skills/intimidation",
			pm + "skills/medicine",
			pm + "skills/persuasion",
			pm + "skills/religion",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "choice/paladin/starting-equipment",
		Kind: "choice",
		Name: "Starting Equipment",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/starting-equipment",
			pm + "equipment/chain-mail",
			pm + "equipment/shield",
			pm + "equipment/longsword",
			pm + "equipment/javelin",
			pm + "equipment/holy-symbol",
			pm + "equipment/priests-pack",
			pm + "currency/gp",
		},
		Select: 1,
		OptionKeys: []string{
			pm + "paladin/equipment/a",
			pm + "paladin/equipment/b",
		},
		Capacity: 0,
		Options: []CharacterChoiceOption{
			{Key: pm + "paladin/equipment/a",
				Quantities: map[string]int{pm + "equipment/chain-mail": 1, pm + "equipment/shield": 1, pm + "equipment/longsword": 1, pm + "equipment/javelin": 6, pm + "equipment/holy-symbol": 1, pm + "equipment/priests-pack": 1, pm + "currency/gp": 9}},
			{Key: pm + "paladin/equipment/b",
				Quantities: map[string]int{pm + "currency/gp": 150}},
		},
	},
	{
		ID:   pc + "level/paladin/1",
		Kind: "level",
		Name: "Paladin level 1",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            1,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 2,
			FeatureRefs: []string{
				pc + "feature/paladin/lay-on-hands",
				pc + "feature/paladin/spellcasting",
				pc + "feature/paladin/weapon-mastery",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/lay-on-hands",
		Kind: "feature",
		Name: "Lay On Hands",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-lay-on-hands",
				Page: 53},
		},
		ParentID:   pc + "level/paladin/1",
		Level:      0,
		ChoiceRefs: []string{},
		ResourceRefs: []string{
			pc + "resource/paladin/lay-on-hands",
		},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/lay-on-hands-pool",
		},
		RequiredMechanics: []string{
			pm + "paladin/lay-on-hands/bonus-action-touch-including-self-heal-up-to-remaining-pool",
			pm + "paladin/healing-hp-per-level",
			pm + "paladin/lay-on-hands/remove-poisoned-no-healing-for-cost",
			pm + "paladin/poisoned-removal-cost-hp",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/healing-hp-per-level": 5, pm + "paladin/poisoned-removal-cost-hp": 5},
	},
	{
		ID:   pc + "feature/paladin/spellcasting",
		Kind: "feature",
		Name: "Spellcasting",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID: pc + "level/paladin/1",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/prepared-spells",
		},
		ResourceRefs: []string{
			pc + "resource/paladin/spell-slot-1",
			pc + "resource/paladin/spell-slot-2",
			pc + "resource/paladin/spell-slot-3",
			pc + "resource/paladin/spell-slot-4",
			pc + "resource/paladin/spell-slot-5",
		},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/charisma",
			pm + "paladin/spellcasting/holy-symbol-focus",
			pm + "paladin/spellcasting/long-rest-replace-one-prepared-spell",
			pm + "paladin/replaced-spells-per-long-rest",
			pm + "paladin/spellcasting/always-prepared-excluded-from-count-still-paladin-spells",
			pm + "paladin/recommended-spells/heroism",
			pm + "paladin/recommended-spells/searing-smite",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/replaced-spells-per-long-rest": 1},
	},
	{
		ID:   pc + "feature/paladin/weapon-mastery",
		Kind: "feature",
		Name: "Weapon Mastery",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-weapon-mastery",
				Page: 54},
		},
		ParentID: pc + "level/paladin/1",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/weapon-mastery",
		},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/weapon-mastery/examples-longswords-javelins",
			pm + "paladin/weapon-mastery/replacement-examples-halberds-flails",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
	},
	{
		ID:   pc + "level/paladin/2",
		Kind: "level",
		Name: "Paladin level 2",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            2,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 2,
			FeatureRefs: []string{
				pc + "feature/paladin/fighting-style",
				pc + "feature/paladin/paladins-smite",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/fighting-style",
		Kind: "feature",
		Name: "Fighting Style",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-2-fighting-style",
				Page: 54},
		},
		ParentID: pc + "level/paladin/2",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/fighting-style",
		},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/fighting-style/choose-feat-or-blessed-warrior",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
	},
	{
		ID:   pc + "feature/paladin/paladins-smite",
		Kind: "feature",
		Name: "Paladin’s Smite",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-2-paladins-smite",
				Page: 54},
		},
		ParentID:   pc + "level/paladin/2",
		Level:      0,
		ChoiceRefs: []string{},
		ResourceRefs: []string{
			pc + "resource/paladin/paladins-smite-free-cast",
		},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "spells/divine-smite",
			pm + "paladin/paladins-smite/always-prepared",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		AlwaysPrepared: []CharacterPreparedSpells{
			{Level: 2,
				SpellKeys: []string{
					pm + "spells/divine-smite",
				}},
		},
	},
	{
		ID:   pc + "level/paladin/3",
		Kind: "level",
		Name: "Paladin level 3",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            3,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 2,
			FeatureRefs: []string{
				pc + "feature/paladin/channel-divinity",
				pc + "feature/paladin/paladin-subclass",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/channel-divinity",
		Kind: "feature",
		Name: "Channel Divinity",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-3-channel-divinity",
				Page: 54},
		},
		ParentID: pc + "level/paladin/3",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/channel-divinity-effect",
		},
		ResourceRefs: []string{
			pc + "resource/paladin/channel-divinity",
		},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/channel-divinity/use-paladin-spell-save-dc",
			pm + "paladin/short-rest-recovered-uses",
			pm + "paladin/divine-sense/bonus-action-expend-one-channel-divinity",
			pm + "paladin/divine-sense/detect-location-and-type-celestials-fiends-undead",
			pm + "paladin/divine-sense/detect-consecrated-desecrated-places-objects-hallow-reference",
			pm + "paladin/divine-sense/ends-if-incapacitated",
			pm + "paladin/divine-sense-cost-uses",
			pm + "paladin/divine-sense-duration-minutes",
			pm + "paladin/divine-sense-radius-feet",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/short-rest-recovered-uses": 1, pm + "paladin/divine-sense-cost-uses": 1, pm + "paladin/divine-sense-duration-minutes": 10, pm + "paladin/divine-sense-radius-feet": 60},
	},
	{
		ID:   pc + "feature/paladin/paladin-subclass",
		Kind: "feature",
		Name: "Paladin Subclass",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-3-paladin-subclass",
				Page: 55},
		},
		ParentID: pc + "level/paladin/3",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/paladin-subclass",
		},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/subclass/choose-one",
			pm + "paladin/subclass/features-at-or-below-paladin-level",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
	},
	{
		ID:   pc + "level/paladin/4",
		Kind: "level",
		Name: "Paladin level 4",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            4,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 2,
			FeatureRefs: []string{
				pc + "feature/paladin/ability-score-improvement",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/ability-score-improvement",
		Kind: "feature",
		Name: "Ability Score Improvement",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-4-ability-score-improvement",
				Page: 55},
		},
		ParentID: pc + "level/paladin/4",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/ability-score-improvement",
		},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/ability-score-improvement",
			pm + "paladin/asi-repeat-at-8-12-16",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
	},
	{
		ID:   pc + "level/paladin/5",
		Kind: "level",
		Name: "Paladin level 5",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            5,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 3,
			FeatureRefs: []string{
				pc + "feature/paladin/extra-attack",
				pc + "feature/paladin/faithful-steed",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/extra-attack",
		Kind: "feature",
		Name: "Extra Attack",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-5-extra-attack",
				Page: 55},
		},
		ParentID:         pc + "level/paladin/5",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/extra-attack/attack-action-own-turn",
			pm + "paladin/attacks-per-attack-action",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/attacks-per-attack-action": 2},
	},
	{
		ID:   pc + "feature/paladin/faithful-steed",
		Kind: "feature",
		Name: "Faithful Steed",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-5-faithful-steed",
				Page: 55},
		},
		ParentID:   pc + "level/paladin/5",
		Level:      0,
		ChoiceRefs: []string{},
		ResourceRefs: []string{
			pc + "resource/paladin/faithful-steed-free-cast",
		},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "spells/find-steed",
			pm + "paladin/faithful-steed/always-prepared",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		AlwaysPrepared: []CharacterPreparedSpells{
			{Level: 5,
				SpellKeys: []string{
					pm + "spells/find-steed",
				}},
		},
	},
	{
		ID:   pc + "level/paladin/6",
		Kind: "level",
		Name: "Paladin level 6",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            6,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 3,
			FeatureRefs: []string{
				pc + "feature/paladin/aura-of-protection",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/aura-of-protection",
		Kind: "feature",
		Name: "Aura of Protection",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-6-aura-of-protection",
				Page: 55},
		},
		ParentID:         pc + "level/paladin/6",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/aura-of-protection/self-origin-emanation",
			pm + "paladin/aura-initial-radius-feet",
			pm + "paladin/aura-of-protection/inactive-if-incapacitated",
			pm + "paladin/aura-of-protection/self-and-allies-saving-throws-charisma-modifier",
			pm + "paladin/aura-minimum-bonus",
			pm + "paladin/aura-of-protection/only-one-aura-benefit-target-chooses",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/aura-initial-radius-feet": 10, pm + "paladin/aura-minimum-bonus": 1},
	},
	{
		ID:   pc + "level/paladin/7",
		Kind: "level",
		Name: "Paladin level 7",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            7,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
			pm + "paladin/subclass-feature",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 3,
			FeatureRefs: []string{},
			FeatureKeys: []string{
				pm + "paladin/subclass-feature",
			}},
	},
	{
		ID:   pc + "level/paladin/8",
		Kind: "level",
		Name: "Paladin level 8",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            8,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 3,
			FeatureRefs: []string{
				pc + "feature/paladin/ability-score-improvement",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "level/paladin/9",
		Kind: "level",
		Name: "Paladin level 9",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            9,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 4,
			FeatureRefs: []string{
				pc + "feature/paladin/abjure-foes",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/abjure-foes",
		Kind: "feature",
		Name: "Abjure Foes",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-9-abjure-foes",
				Page: 55},
		},
		ParentID:     pc + "level/paladin/9",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/abjure-channel-divinity",
		},
		RequiredMechanics: []string{
			pm + "paladin/abjure-foes/magic-action-present-holy-symbol-or-weapon",
			pm + "paladin/channel-divinity-cost",
			pm + "paladin/abjure-foes/visible-targets-charisma-modifier-minimum-one",
			pm + "paladin/abjure-minimum-targets",
			pm + "paladin/abjure-range-feet",
			pm + "paladin/abjure-foes/wisdom-save-or-frightened",
			pm + "paladin/abjure-foes/ends-on-any-damage",
			pm + "paladin/abjure-duration-minutes",
			pm + "paladin/abjure-foes/frightened-target-only-one-move-action-or-bonus-action-on-turn",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/channel-divinity-cost": 1, pm + "paladin/abjure-minimum-targets": 1, pm + "paladin/abjure-range-feet": 60, pm + "paladin/abjure-duration-minutes": 1},
	},
	{
		ID:   pc + "level/paladin/10",
		Kind: "level",
		Name: "Paladin level 10",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            10,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 4,
			FeatureRefs: []string{
				pc + "feature/paladin/aura-of-courage",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/aura-of-courage",
		Kind: "feature",
		Name: "Aura of Courage",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-10-aura-of-courage",
				Page: 55},
		},
		ParentID:     pc + "level/paladin/10",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/courage-aura",
		},
		RequiredMechanics: []string{
			pm + "paladin/aura-of-courage/self-and-allies-frightened-immunity-in-protection-aura",
			pm + "paladin/aura-of-courage/entering-frightened-ally-condition-suppressed-while-in-aura",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
	},
	{
		ID:   pc + "level/paladin/11",
		Kind: "level",
		Name: "Paladin level 11",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            11,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 4,
			FeatureRefs: []string{
				pc + "feature/paladin/radiant-strikes",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/radiant-strikes",
		Kind: "feature",
		Name: "Radiant Strikes",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-11-radiant-strikes",
				Page: 55},
		},
		ParentID:         pc + "level/paladin/11",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/radiant-strikes/attack-roll-hit-melee-weapon-or-unarmed-strike",
			pm + "paladin/radiant-damage-dice",
			pm + "paladin/radiant-damage-die-sides",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/radiant-damage-dice": 1, pm + "paladin/radiant-damage-die-sides": 8},
	},
	{
		ID:   pc + "level/paladin/12",
		Kind: "level",
		Name: "Paladin level 12",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            12,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 4,
			FeatureRefs: []string{
				pc + "feature/paladin/ability-score-improvement",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "level/paladin/13",
		Kind: "level",
		Name: "Paladin level 13",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            13,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 5,
			FeatureRefs: []string{},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "level/paladin/14",
		Kind: "level",
		Name: "Paladin level 14",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            14,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 5,
			FeatureRefs: []string{
				pc + "feature/paladin/restoring-touch",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/restoring-touch",
		Kind: "feature",
		Name: "Restoring Touch",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-14-restoring-touch",
				Page: 55},
		},
		ParentID: pc + "level/paladin/14",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/restoring-touch-conditions",
		},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/restoring-healing-pool",
		},
		RequiredMechanics: []string{
			pm + "paladin/restoring-touch/with-lay-on-hands-remove-one-or-more-conditions",
			pm + "paladin/restoring-touch/cost-does-not-heal",
			pm + "paladin/restoring-cost-hp-per-condition",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/restoring-cost-hp-per-condition": 5},
	},
	{
		ID:   pc + "level/paladin/15",
		Kind: "level",
		Name: "Paladin level 15",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            15,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
			pm + "paladin/subclass-feature",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 5,
			FeatureRefs: []string{},
			FeatureKeys: []string{
				pm + "paladin/subclass-feature",
			}},
	},
	{
		ID:   pc + "level/paladin/16",
		Kind: "level",
		Name: "Paladin level 16",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            16,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 5,
			FeatureRefs: []string{
				pc + "feature/paladin/ability-score-improvement",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "level/paladin/17",
		Kind: "level",
		Name: "Paladin level 17",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            17,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 6,
			FeatureRefs: []string{},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "level/paladin/18",
		Kind: "level",
		Name: "Paladin level 18",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            18,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 6,
			FeatureRefs: []string{
				pc + "feature/paladin/aura-expansion",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/aura-expansion",
		Kind: "feature",
		Name: "Aura Expansion",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-18-aura-expansion",
				Page: 55},
		},
		ParentID:     pc + "level/paladin/18",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/expanded-aura",
		},
		RequiredMechanics: []string{
			pm + "paladin/aura-expansion/replaces-protection-radius",
			pm + "paladin/aura-expanded-radius-feet",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Parameters: map[string]int{pm + "paladin/aura-expanded-radius-feet": 30},
	},
	{
		ID:   pc + "level/paladin/19",
		Kind: "level",
		Name: "Paladin level 19",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            19,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 6,
			FeatureRefs: []string{
				pc + "feature/paladin/epic-boon",
			},
			FeatureKeys: []string{}},
	},
	{
		ID:   pc + "feature/paladin/epic-boon",
		Kind: "feature",
		Name: "Epic Boon",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-19-epic-boon",
				Page: 55},
		},
		ParentID: pc + "level/paladin/19",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/epic-boon",
		},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/epic-boon",
			pm + "paladin/epic-boon/recommended-boon-of-truesight",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
	},
	{
		ID:   pc + "level/paladin/20",
		Kind: "level",
		Name: "Paladin level 20",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/classes/paladin",
				Page: 53},
		},
		ParentID:         pc + "class/paladin",
		Level:            20,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/class-progression",
			pm + "paladin/subclass-feature",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Progression: &CharacterClassProgression{ProficiencyBonus: 6,
			FeatureRefs: []string{},
			FeatureKeys: []string{
				pm + "paladin/subclass-feature",
			}},
	},
	{
		ID:   pc + "resource/paladin/lay-on-hands",
		Kind: "resource",
		Name: "Lay On Hands",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-lay-on-hands",
				Page: 53},
		},
		ParentID:         pc + "feature/paladin/lay-on-hands",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/lay-on-hands/long-rest-replenish-all",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   5,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 1,
				Value: 5},
			{Level: 2,
				Value: 10},
			{Level: 3,
				Value: 15},
			{Level: 4,
				Value: 20},
			{Level: 5,
				Value: 25},
			{Level: 6,
				Value: 30},
			{Level: 7,
				Value: 35},
			{Level: 8,
				Value: 40},
			{Level: 9,
				Value: 45},
			{Level: 10,
				Value: 50},
			{Level: 11,
				Value: 55},
			{Level: 12,
				Value: 60},
			{Level: 13,
				Value: 65},
			{Level: 14,
				Value: 70},
			{Level: 15,
				Value: 75},
			{Level: 16,
				Value: 80},
			{Level: 17,
				Value: 85},
			{Level: 18,
				Value: 90},
			{Level: 19,
				Value: 95},
			{Level: 20,
				Value: 100},
		},
	},
	{
		ID:   pc + "prerequisite/paladin/lay-on-hands-pool",
		Kind: "prerequisite",
		Name: "Lay On Hands Pool",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-lay-on-hands",
				Page: 53},
		},
		ParentID:         pc + "feature/paladin/lay-on-hands",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/lay-on-hands/expend-only-remaining-pool",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/lay-on-hands/expend-only-remaining-pool",
			RecordRefs: []string{
				pc + "resource/paladin/lay-on-hands",
			},
			Parameters: map[string]int{}},
	},
	{
		ID:   pc + "choice/paladin/prepared-spells",
		Kind: "choice",
		Name: "Prepared Spells",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID:     pc + "feature/paladin/spellcasting",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/prepared-spells-slot-level",
		},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/add-spells-when-table-count-increases",
		},
		Select: 2,
		OptionKeys: []string{
			pm + "paladin/spell-list/p55-56/level-1-plus",
		},
		Capacity: 0,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 1,
				Value: 2},
			{Level: 2,
				Value: 3},
			{Level: 3,
				Value: 4},
			{Level: 4,
				Value: 5},
			{Level: 5,
				Value: 6},
			{Level: 6,
				Value: 6},
			{Level: 7,
				Value: 7},
			{Level: 8,
				Value: 7},
			{Level: 9,
				Value: 9},
			{Level: 10,
				Value: 9},
			{Level: 11,
				Value: 10},
			{Level: 12,
				Value: 10},
			{Level: 13,
				Value: 11},
			{Level: 14,
				Value: 11},
			{Level: 15,
				Value: 12},
			{Level: 16,
				Value: 12},
			{Level: 17,
				Value: 14},
			{Level: 18,
				Value: 14},
			{Level: 19,
				Value: 15},
			{Level: 20,
				Value: 15},
		},
	},
	{
		ID:   pc + "prerequisite/paladin/prepared-spells-slot-level",
		Kind: "prerequisite",
		Name: "Prepared Spells Slot Level",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID:         pc + "choice/paladin/prepared-spells",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/prepared-level-has-slot",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/spellcasting/prepared-level-has-slot",
			RecordRefs: []string{
				pc + "feature/paladin/spellcasting",
			},
			Parameters: map[string]int{"minimum-spell-level": 1}},
	},
	{
		ID:   pc + "resource/paladin/spell-slot-1",
		Kind: "resource",
		Name: "Spell Slot 1",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/spellcasting",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/long-rest-all-slots",
			pm + "paladin/spell-slot-level/1",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   2,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 1,
				Value: 2},
			{Level: 2,
				Value: 2},
			{Level: 3,
				Value: 3},
			{Level: 4,
				Value: 3},
			{Level: 5,
				Value: 4},
			{Level: 6,
				Value: 4},
			{Level: 7,
				Value: 4},
			{Level: 8,
				Value: 4},
			{Level: 9,
				Value: 4},
			{Level: 10,
				Value: 4},
			{Level: 11,
				Value: 4},
			{Level: 12,
				Value: 4},
			{Level: 13,
				Value: 4},
			{Level: 14,
				Value: 4},
			{Level: 15,
				Value: 4},
			{Level: 16,
				Value: 4},
			{Level: 17,
				Value: 4},
			{Level: 18,
				Value: 4},
			{Level: 19,
				Value: 4},
			{Level: 20,
				Value: 4},
		},
	},
	{
		ID:   pc + "resource/paladin/spell-slot-2",
		Kind: "resource",
		Name: "Spell Slot 2",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/spellcasting",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/long-rest-all-slots",
			pm + "paladin/spell-slot-level/2",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   2,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 5,
				Value: 2},
			{Level: 6,
				Value: 2},
			{Level: 7,
				Value: 3},
			{Level: 8,
				Value: 3},
			{Level: 9,
				Value: 3},
			{Level: 10,
				Value: 3},
			{Level: 11,
				Value: 3},
			{Level: 12,
				Value: 3},
			{Level: 13,
				Value: 3},
			{Level: 14,
				Value: 3},
			{Level: 15,
				Value: 3},
			{Level: 16,
				Value: 3},
			{Level: 17,
				Value: 3},
			{Level: 18,
				Value: 3},
			{Level: 19,
				Value: 3},
			{Level: 20,
				Value: 3},
		},
	},
	{
		ID:   pc + "resource/paladin/spell-slot-3",
		Kind: "resource",
		Name: "Spell Slot 3",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/spellcasting",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/long-rest-all-slots",
			pm + "paladin/spell-slot-level/3",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   2,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 9,
				Value: 2},
			{Level: 10,
				Value: 2},
			{Level: 11,
				Value: 3},
			{Level: 12,
				Value: 3},
			{Level: 13,
				Value: 3},
			{Level: 14,
				Value: 3},
			{Level: 15,
				Value: 3},
			{Level: 16,
				Value: 3},
			{Level: 17,
				Value: 3},
			{Level: 18,
				Value: 3},
			{Level: 19,
				Value: 3},
			{Level: 20,
				Value: 3},
		},
	},
	{
		ID:   pc + "resource/paladin/spell-slot-4",
		Kind: "resource",
		Name: "Spell Slot 4",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/spellcasting",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/long-rest-all-slots",
			pm + "paladin/spell-slot-level/4",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   1,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 13,
				Value: 1},
			{Level: 14,
				Value: 1},
			{Level: 15,
				Value: 2},
			{Level: 16,
				Value: 2},
			{Level: 17,
				Value: 3},
			{Level: 18,
				Value: 3},
			{Level: 19,
				Value: 3},
			{Level: 20,
				Value: 3},
		},
	},
	{
		ID:   pc + "resource/paladin/spell-slot-5",
		Kind: "resource",
		Name: "Spell Slot 5",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-spellcasting",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/spellcasting",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/spellcasting/long-rest-all-slots",
			pm + "paladin/spell-slot-level/5",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   1,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 17,
				Value: 1},
			{Level: 18,
				Value: 1},
			{Level: 19,
				Value: 2},
			{Level: 20,
				Value: 2},
		},
	},
	{
		ID:   pc + "choice/paladin/weapon-mastery",
		Kind: "choice",
		Name: "Weapon Mastery",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-weapon-mastery",
				Page: 54},
		},
		ParentID:     pc + "feature/paladin/weapon-mastery",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/mastery-proficiency",
		},
		RequiredMechanics: []string{
			pm + "paladin/weapon-mastery/long-rest-replace-kinds",
		},
		Select: 2,
		OptionKeys: []string{
			pm + "weapons/proficient-kinds",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "prerequisite/paladin/mastery-proficiency",
		Kind: "prerequisite",
		Name: "Mastery Proficiency",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-1-weapon-mastery",
				Page: 54},
		},
		ParentID:         pc + "choice/paladin/weapon-mastery",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/weapon-mastery/requires-weapon-proficiency",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/weapon-mastery/requires-weapon-proficiency",
			RecordRefs: []string{
				pc + "class/paladin",
			},
			Parameters: map[string]int{}},
	},
	{
		ID:   pc + "choice/paladin/fighting-style",
		Kind: "choice",
		Name: "Fighting Style",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-2-fighting-style",
				Page: 54},
		},
		ParentID: pc + "feature/paladin/fighting-style",
		Level:    0,
		ChoiceRefs: []string{
			pc + "choice/paladin/blessed-warrior",
		},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/fighting-style",
		},
		Select: 1,
		OptionKeys: []string{
			pm + "feats/fighting-style",
			pm + "paladin/fighting-style/blessed-warrior",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "choice/paladin/blessed-warrior",
		Kind: "choice",
		Name: "Blessed Warrior",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-2-fighting-style",
				Page: 54},
		},
		ParentID:     pc + "choice/paladin/fighting-style",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/blessed-warrior-selected",
		},
		RequiredMechanics: []string{
			pm + "paladin/blessed-warrior/count-as-paladin-charisma",
			pm + "paladin/blessed-warrior/replace-one-cleric-cantrip-on-paladin-level",
			pm + "paladin/blessed-warrior/recommended-guidance-sacred-flame",
		},
		Select: 2,
		OptionKeys: []string{
			pm + "cleric/spell-list/cantrips",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "prerequisite/paladin/blessed-warrior-selected",
		Kind: "prerequisite",
		Name: "Blessed Warrior Selected",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-2-fighting-style",
				Page: 54},
		},
		ParentID:         pc + "choice/paladin/blessed-warrior",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/fighting-style/blessed-warrior-selected",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/fighting-style/blessed-warrior-selected",
			RecordRefs: []string{
				pc + "choice/paladin/fighting-style",
			},
			Parameters: map[string]int{}},
	},
	{
		ID:   pc + "resource/paladin/paladins-smite-free-cast",
		Kind: "resource",
		Name: "Paladins Smite Free Cast",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-2-paladins-smite",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/paladins-smite",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/paladins-smite/one-slot-free-cast-long-rest-recovery",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   1,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 2,
				Value: 1},
			{Level: 3,
				Value: 1},
			{Level: 4,
				Value: 1},
			{Level: 5,
				Value: 1},
			{Level: 6,
				Value: 1},
			{Level: 7,
				Value: 1},
			{Level: 8,
				Value: 1},
			{Level: 9,
				Value: 1},
			{Level: 10,
				Value: 1},
			{Level: 11,
				Value: 1},
			{Level: 12,
				Value: 1},
			{Level: 13,
				Value: 1},
			{Level: 14,
				Value: 1},
			{Level: 15,
				Value: 1},
			{Level: 16,
				Value: 1},
			{Level: 17,
				Value: 1},
			{Level: 18,
				Value: 1},
			{Level: 19,
				Value: 1},
			{Level: 20,
				Value: 1},
		},
	},
	{
		ID:   pc + "resource/paladin/faithful-steed-free-cast",
		Kind: "resource",
		Name: "Faithful Steed Free Cast",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-5-faithful-steed",
				Page: 55},
		},
		ParentID:         pc + "feature/paladin/faithful-steed",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/faithful-steed/one-slot-free-cast-long-rest-recovery",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   1,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 5,
				Value: 1},
			{Level: 6,
				Value: 1},
			{Level: 7,
				Value: 1},
			{Level: 8,
				Value: 1},
			{Level: 9,
				Value: 1},
			{Level: 10,
				Value: 1},
			{Level: 11,
				Value: 1},
			{Level: 12,
				Value: 1},
			{Level: 13,
				Value: 1},
			{Level: 14,
				Value: 1},
			{Level: 15,
				Value: 1},
			{Level: 16,
				Value: 1},
			{Level: 17,
				Value: 1},
			{Level: 18,
				Value: 1},
			{Level: 19,
				Value: 1},
			{Level: 20,
				Value: 1},
		},
	},
	{
		ID:   pc + "resource/paladin/channel-divinity",
		Kind: "resource",
		Name: "Channel Divinity",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-3-channel-divinity",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/channel-divinity",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/channel-divinity/short-rest-recover-one",
			pm + "paladin/channel-divinity/long-rest-recover-all",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   2,
		ValuesByLevel: []CharacterLevelValue{
			{Level: 3,
				Value: 2},
			{Level: 4,
				Value: 2},
			{Level: 5,
				Value: 2},
			{Level: 6,
				Value: 2},
			{Level: 7,
				Value: 2},
			{Level: 8,
				Value: 2},
			{Level: 9,
				Value: 2},
			{Level: 10,
				Value: 2},
			{Level: 11,
				Value: 3},
			{Level: 12,
				Value: 3},
			{Level: 13,
				Value: 3},
			{Level: 14,
				Value: 3},
			{Level: 15,
				Value: 3},
			{Level: 16,
				Value: 3},
			{Level: 17,
				Value: 3},
			{Level: 18,
				Value: 3},
			{Level: 19,
				Value: 3},
			{Level: 20,
				Value: 3},
		},
	},
	{
		ID:   pc + "choice/paladin/channel-divinity-effect",
		Kind: "choice",
		Name: "Channel Divinity Effect",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-3-channel-divinity",
				Page: 54},
		},
		ParentID:         pc + "feature/paladin/channel-divinity",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/channel-divinity-effect",
		},
		Select: 1,
		OptionKeys: []string{
			pm + "paladin/channel-divinity/divine-sense",
			pm + "paladin/channel-divinity/other-paladin-feature-effects",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "choice/paladin/paladin-subclass",
		Kind: "choice",
		Name: "Paladin Subclass",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-3-paladin-subclass",
				Page: 55},
		},
		ParentID:         pc + "feature/paladin/paladin-subclass",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/paladin-subclass",
		},
		Select: 1,
		OptionKeys: []string{
			pm + "paladin/subclasses",
			pm + "paladin/subclasses/oath-of-devotion",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "choice/paladin/ability-score-improvement",
		Kind: "choice",
		Name: "Ability Score Improvement",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-4-ability-score-improvement",
				Page: 55},
		},
		ParentID:     pc + "feature/paladin/ability-score-improvement",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/ability-score-improvement-qualification",
		},
		RequiredMechanics: []string{
			pm + "paladin/ability-score-improvement",
		},
		Select: 1,
		OptionKeys: []string{
			pm + "feats/ability-score-improvement",
			pm + "feats/qualified",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "prerequisite/paladin/ability-score-improvement-qualification",
		Kind: "prerequisite",
		Name: "Ability Score Improvement Qualification",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-4-ability-score-improvement",
				Page: 55},
		},
		ParentID:         pc + "choice/paladin/ability-score-improvement",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "prerequisite/feat-qualification",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "prerequisite/feat-qualification",
			RecordRefs: []string{},
			Parameters: map[string]int{}},
	},
	{
		ID:   pc + "choice/paladin/epic-boon",
		Kind: "choice",
		Name: "Epic Boon",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-19-epic-boon",
				Page: 55},
		},
		ParentID:     pc + "feature/paladin/epic-boon",
		Level:        0,
		ChoiceRefs:   []string{},
		ResourceRefs: []string{},
		PrerequisiteRefs: []string{
			pc + "prerequisite/paladin/epic-boon-qualification",
		},
		RequiredMechanics: []string{
			pm + "paladin/epic-boon",
		},
		Select: 1,
		OptionKeys: []string{
			pm + "feats/epic-boon",
			pm + "feats/qualified",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "prerequisite/paladin/epic-boon-qualification",
		Kind: "prerequisite",
		Name: "Epic Boon Qualification",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-19-epic-boon",
				Page: 55},
		},
		ParentID:         pc + "choice/paladin/epic-boon",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "prerequisite/feat-qualification",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "prerequisite/feat-qualification",
			RecordRefs: []string{},
			Parameters: map[string]int{}},
	},
	{
		ID:   pc + "prerequisite/paladin/abjure-channel-divinity",
		Kind: "prerequisite",
		Name: "Abjure Channel Divinity",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-9-abjure-foes",
				Page: 55},
		},
		ParentID:         pc + "feature/paladin/abjure-foes",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/abjure-foes/expend-channel-divinity",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/abjure-foes/expend-channel-divinity",
			RecordRefs: []string{
				pc + "resource/paladin/channel-divinity",
			},
			Parameters: map[string]int{"cost-uses": 1}},
	},
	{
		ID:   pc + "prerequisite/paladin/courage-aura",
		Kind: "prerequisite",
		Name: "Courage Aura",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-10-aura-of-courage",
				Page: 55},
		},
		ParentID:         pc + "feature/paladin/aura-of-courage",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/aura-of-courage/within-aura-of-protection",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/aura-of-courage/within-aura-of-protection",
			RecordRefs: []string{
				pc + "feature/paladin/aura-of-protection",
			},
			Parameters: map[string]int{}},
	},
	{
		ID:   pc + "choice/paladin/restoring-touch-conditions",
		Kind: "choice",
		Name: "Restoring Touch Conditions",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-14-restoring-touch",
				Page: 55},
		},
		ParentID:         pc + "feature/paladin/restoring-touch",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/restoring-touch/choose-one-or-more-with-per-condition-cost",
		},
		Select: 1,
		OptionKeys: []string{
			pm + "conditions/blinded",
			pm + "conditions/charmed",
			pm + "conditions/deafened",
			pm + "conditions/frightened",
			pm + "conditions/paralyzed",
			pm + "conditions/stunned",
		},
		Capacity: 0,
	},
	{
		ID:   pc + "prerequisite/paladin/restoring-healing-pool",
		Kind: "prerequisite",
		Name: "Restoring Healing Pool",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-14-restoring-touch",
				Page: 55},
		},
		ParentID:         pc + "feature/paladin/restoring-touch",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/restoring-touch/expend-lay-on-hands-pool",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/restoring-touch/expend-lay-on-hands-pool",
			RecordRefs: []string{
				pc + "resource/paladin/lay-on-hands",
			},
			Parameters: map[string]int{"hp-per-condition": 5}},
	},
	{
		ID:   pc + "prerequisite/paladin/expanded-aura",
		Kind: "prerequisite",
		Name: "Expanded Aura",
		Sources: []CharacterSource{
			{InventoryID: "srd-5.2.1:planning/class-features/paladin-level-18-aura-expansion",
				Page: 55},
		},
		ParentID:         pc + "feature/paladin/aura-expansion",
		Level:            0,
		ChoiceRefs:       []string{},
		ResourceRefs:     []string{},
		PrerequisiteRefs: []string{},
		RequiredMechanics: []string{
			pm + "paladin/aura-expansion/changes-aura-of-protection",
		},
		Select:     0,
		OptionKeys: []string{},
		Capacity:   0,
		Requirement: &CharacterRequirement{Key: pm + "paladin/aura-expansion/changes-aura-of-protection",
			RecordRefs: []string{
				pc + "feature/paladin/aura-of-protection",
			},
			Parameters: map[string]int{}},
	},
}

func paladinCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-paladin.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func paladinMatches(c CharacterCatalogue) bool {
	return reflect.DeepEqual(c.Records, paladinSourceGraph)
}
func TestPaladinCompleteSource(t *testing.T) {
	c := paladinCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if !paladinMatches(c) {
		for i, r := range c.Records {
			if i >= len(paladinSourceGraph) || !reflect.DeepEqual(r, paladinSourceGraph[i]) {
				t.Errorf("source graph differs at %s: %+v", r.ID, r)
			}
		}
		t.Fatal("complete source graph differs")
	}
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
	want, got := map[string]int{}, map[string]int{}
	for _, r := range inv.Records {
		if r.ID == "srd-5.2.1:planning/classes/paladin" || strings.HasPrefix(r.ID, "srd-5.2.1:planning/class-features/paladin-level-") {
			want[r.ID] = r.Page
		}
	}
	for _, r := range c.Records {
		if r.Kind == "class" || r.Kind == "feature" {
			got[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
	}
	if len(got) != 18 || !reflect.DeepEqual(got, want) {
		t.Fatalf("18 identities/pages: %v", got)
	}
}

func TestPaladinContrarySourceControls(t *testing.T) {
	tests := []struct {
		name, id string
		change   func(*CharacterRecord)
	}{
		{"existing wrong operand", "prerequisite/paladin/restoring-healing-pool", func(r *CharacterRecord) {
			r.Requirement.RecordRefs = []string{pc + "resource/paladin/channel-divinity"}
		}},
		{"range", "feature/paladin/abjure-foes", func(r *CharacterRecord) { r.Parameters[pm+"paladin/abjure-range-feet"] = 61 }},
		{"cost", "feature/paladin/restoring-touch", func(r *CharacterRecord) { r.Parameters[pm+"paladin/restoring-cost-hp-per-condition"] = 4 }},
		{"recovery", "resource/paladin/channel-divinity", func(r *CharacterRecord) {
			r.RequiredMechanics = []string{pm + "paladin/channel-divinity/short-rest-recover-all"}
		}},
		{"condition", "feature/paladin/aura-of-protection", func(r *CharacterRecord) {
			r.RequiredMechanics[2] = pm + "paladin/aura-of-protection/active-if-incapacitated"
		}},
		{"missing variant", "choice/paladin/restoring-touch-conditions", func(r *CharacterRecord) { r.OptionKeys = []string{pm + "conditions/blinded"} }},
		{"missing table cell", "resource/paladin/spell-slot-5", func(r *CharacterRecord) { r.ValuesByLevel = append(r.ValuesByLevel[:1], r.ValuesByLevel[2:]...) }},
		{"invented scaling", "feature/paladin/radiant-strikes", func(r *CharacterRecord) { r.Parameters[pm+"paladin/radiant-damage-dice"] = 2 }},
		{"invented level scaling", "feature/paladin/radiant-strikes", func(r *CharacterRecord) {
			key := pm + "paladin/radiant-damage-dice-at-level-20"
			r.Parameters[key] = 2
			r.RequiredMechanics = append(r.RequiredMechanics, key)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := paladinCatalogue(t)
			found := false
			for i := range c.Records {
				if c.Records[i].ID == pc+tt.id {
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
			if paladinMatches(c) {
				t.Fatal("contrary source accepted")
			}
		})
	}
}
func TestPaladinGuardsReadinessComposition(t *testing.T) {
	repo := os.DirFS("../../..")
	c := paladinCatalogue(t)
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
		{"schema version", func(c *CharacterCatalogue) { c.SchemaVersion = 2 }},
		{"wrong pin", func(c *CharacterCatalogue) { c.Source.SHA256 = strings.Repeat("0", 64) }},
		{"wrong page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 54 }},
		{"future grant", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Progression != nil {
					c.Records[i].Progression.FeatureRefs = []string{"srd-5.2.1:character/feature/paladin/epic-boon"}
					break
				}
			}
		}},
		{"missing owner link", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"missing operand", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].Requirement.RecordRefs = []string{"srd-5.2.1:character/choice/paladin/missing"}
					break
				}
			}
		}},
		{"unrequired number", func(c *CharacterCatalogue) { c.Records[0].Parameters["srd-5.2.1:mechanic/unknown/value"] = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := paladinCatalogue(t)
			tt.change(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("malformed metadata accepted")
			}
		})
	}
	b, err := os.ReadFile("testdata/character-paladin.json")
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
	for _, file := range []string{"character-cleric.json", "character-fighter.json", "character-barbarian.json"} {
		b, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		merged := paladinCatalogue(t)
		merged.Records = append(merged.Records, other.Records...)
		if err := merged.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := merged.ValidateReadiness(repo); err == nil {
			t.Fatal("composition qualified")
		}
	}
}

func TestPaladinImportDoesNotPromoteFixedBuilds(t *testing.T) {
	beforeFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	beforeCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	c := paladinCatalogue(t)
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
		if _, err := SupportedBuildForConcept("Human Paladin", background); err == nil {
			t.Fatal("Paladin creation became available")
		}
	}
	// Existing fixed paths retain their numeric/projection assertions too.
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	TestCharacterProjection(t)
}

func TestPaladinCanonicalComposition(t *testing.T) {
	combined := paladinCatalogue(t)
	for _, file := range []string{
		"character-fighter.json", "character-cleric.json", "character-barbarian.json",
		"character-bard.json", "character-druid.json", "character-college-of-lore.json",
	} {
		b, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		combined.Records = append(combined.Records, c.Records...)
	}
	repo := os.DirFS("../../..")
	if err := combined.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := combined.ValidateReadiness(repo); err == nil {
		t.Fatal("composition enabled support")
	}
	// A second copy of any canonical definition must fail closed.
	combined.Records = append(combined.Records, paladinCatalogue(t).Records[0])
	if err := combined.Validate(repo); err == nil {
		t.Fatal("duplicate canonical definition accepted")
	}
}
