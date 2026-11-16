package srd521

// Independent p35 transcription: complete records, including absence of scaling and grants.
func loreSourceRecords() []CharacterRecord {
	return []CharacterRecord{
		{
			ID: "srd-5.2.1:character/subclass/bard/college-of-lore", Kind: "subclass", Name: "College of Lore",
			Sources:      []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclasses/bard-college-of-lore", Page: 35}},
			ParentID:     "srd-5.2.1:character/class/bard",
			ChoiceRefs:   []string{},
			ResourceRefs: []string{},
			PrerequisiteRefs: []string{
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/college-of-lore-level",
			},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/subclasses/college-of-lore",
			},
			OptionKeys: []string{},
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/college-of-lore-level", Kind: "prerequisite", Name: "College of Lore level requirement",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclasses/bard-college-of-lore", Page: 35}},
			ParentID:         "srd-5.2.1:character/subclass/bard/college-of-lore",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/prerequisite/class-level",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/prerequisite/class-level",
				RecordRefs: []string{
					"srd-5.2.1:character/class/bard",
				},
				Parameters: map[string]int{"minimum-level": 3}},
		},
		{
			ID: "srd-5.2.1:character/feature/bard/college-of-lore/bonus-proficiencies", Kind: "feature", Name: "Bonus Proficiencies",
			Sources:  []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-3-bonus-proficiencies", Page: 35}},
			ParentID: "srd-5.2.1:character/subclass/bard/college-of-lore",
			ChoiceRefs: []string{
				"srd-5.2.1:character/choice/bard/college-of-lore/bonus-proficiencies",
			},
			ResourceRefs: []string{},
			PrerequisiteRefs: []string{
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/bonus-proficiencies-level",
			},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/bonus-proficiencies/gain-proficiency-with-selected-skills",
			},
			OptionKeys: []string{},
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/bonus-proficiencies-level", Kind: "prerequisite", Name: "Bonus Proficiencies level requirement",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-3-bonus-proficiencies", Page: 35}},
			ParentID:         "srd-5.2.1:character/feature/bard/college-of-lore/bonus-proficiencies",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/prerequisite/class-level",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/prerequisite/class-level",
				RecordRefs: []string{
					"srd-5.2.1:character/class/bard",
				},
				Parameters: map[string]int{"minimum-level": 3}},
		},
		{
			ID: "srd-5.2.1:character/feature/bard/college-of-lore/cutting-words", Kind: "feature", Name: "Cutting Words",
			Sources:      []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-3-cutting-words", Page: 35}},
			ParentID:     "srd-5.2.1:character/subclass/bard/college-of-lore",
			ChoiceRefs:   []string{},
			ResourceRefs: []string{},
			PrerequisiteRefs: []string{
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/cutting-words-level",
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/cutting-words-inspiration",
			},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/target/creature-you-can-see",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/trigger/damage-roll-or-successful-ability-check-or-attack-roll",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/timing/reaction",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/expend-bardic-inspiration",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/roll-canonical-bardic-inspiration-die",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/subtract-from-creature-roll",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/reduce-damage-or-potentially-turn-success-into-failure",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/range-feet",
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/uses-expended",
			},
			OptionKeys: []string{},
			Parameters: map[string]int{"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/range-feet": 60, "srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/uses-expended": 1},
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/cutting-words-level", Kind: "prerequisite", Name: "Cutting Words level requirement",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-3-cutting-words", Page: 35}},
			ParentID:         "srd-5.2.1:character/feature/bard/college-of-lore/cutting-words",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/prerequisite/class-level",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/prerequisite/class-level",
				RecordRefs: []string{
					"srd-5.2.1:character/class/bard",
				},
				Parameters: map[string]int{"minimum-level": 3}},
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/cutting-words-inspiration", Kind: "prerequisite", Name: "Cutting Words Inspiration requirement",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-3-cutting-words", Page: 35}},
			ParentID:         "srd-5.2.1:character/feature/bard/college-of-lore/cutting-words",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/requires-bardic-inspiration-use",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/requires-bardic-inspiration-use",
				RecordRefs: []string{
					"srd-5.2.1:character/resource/bard/bardic-inspiration",
				},
				Parameters: map[string]int{"uses-required": 1}},
		},
		{
			ID: "srd-5.2.1:character/feature/bard/college-of-lore/magical-discoveries", Kind: "feature", Name: "Magical Discoveries",
			Sources:  []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-6-magical-discoveries", Page: 35}},
			ParentID: "srd-5.2.1:character/subclass/bard/college-of-lore",
			ChoiceRefs: []string{
				"srd-5.2.1:character/choice/bard/college-of-lore/magical-discoveries",
			},
			ResourceRefs: []string{},
			PrerequisiteRefs: []string{
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/magical-discoveries-level",
			},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/learn-chosen-spells",
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/always-have-chosen-spells-prepared",
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/replace-one-on-gaining-bard-level",
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/replacement-meets-same-list-and-cantrip-or-slot-requirements",
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/lists/any-combination",
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/spells-learned",
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/replacements-per-bard-level",
			},
			OptionKeys: []string{},
			Parameters: map[string]int{"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/replacements-per-bard-level": 1, "srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/spells-learned": 2},
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/magical-discoveries-level", Kind: "prerequisite", Name: "Magical Discoveries level requirement",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-6-magical-discoveries", Page: 35}},
			ParentID:         "srd-5.2.1:character/feature/bard/college-of-lore/magical-discoveries",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/prerequisite/class-level",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/prerequisite/class-level",
				RecordRefs: []string{
					"srd-5.2.1:character/class/bard",
				},
				Parameters: map[string]int{"minimum-level": 6}},
		},
		{
			ID: "srd-5.2.1:character/feature/bard/college-of-lore/peerless-skill", Kind: "feature", Name: "Peerless Skill",
			Sources:      []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-14-peerless-skill", Page: 35}},
			ParentID:     "srd-5.2.1:character/subclass/bard/college-of-lore",
			ChoiceRefs:   []string{},
			ResourceRefs: []string{},
			PrerequisiteRefs: []string{
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/peerless-skill-level",
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/peerless-skill-inspiration",
			},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/trigger/own-failed-ability-check-or-attack-roll",
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/expend-bardic-inspiration",
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/roll-canonical-bardic-inspiration-die",
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/add-number-to-d20",
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/potentially-turn-failure-into-success",
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/on-failure-inspiration-not-expended",
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/uses-expended-on-success",
			},
			OptionKeys: []string{},
			Parameters: map[string]int{"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/uses-expended-on-success": 1},
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/peerless-skill-level", Kind: "prerequisite", Name: "Peerless Skill level requirement",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-14-peerless-skill", Page: 35}},
			ParentID:         "srd-5.2.1:character/feature/bard/college-of-lore/peerless-skill",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/prerequisite/class-level",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/prerequisite/class-level",
				RecordRefs: []string{
					"srd-5.2.1:character/class/bard",
				},
				Parameters: map[string]int{"minimum-level": 14}},
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/peerless-skill-inspiration", Kind: "prerequisite", Name: "Peerless Skill Inspiration requirement",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-14-peerless-skill", Page: 35}},
			ParentID:         "srd-5.2.1:character/feature/bard/college-of-lore/peerless-skill",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/requires-bardic-inspiration-use",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/requires-bardic-inspiration-use",
				RecordRefs: []string{
					"srd-5.2.1:character/resource/bard/bardic-inspiration",
				},
				Parameters: map[string]int{"uses-required": 1}},
		},
		{
			ID: "srd-5.2.1:character/choice/bard/college-of-lore/bonus-proficiencies", Kind: "choice", Name: "Bonus Proficiencies choice",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-3-bonus-proficiencies", Page: 35}},
			ParentID:         "srd-5.2.1:character/feature/bard/college-of-lore/bonus-proficiencies",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/bonus-proficiencies/any-skills",
			},
			OptionKeys: []string{
				"srd-5.2.1:mechanic/skills/acrobatics",
				"srd-5.2.1:mechanic/skills/animal-handling",
				"srd-5.2.1:mechanic/skills/arcana",
				"srd-5.2.1:mechanic/skills/athletics",
				"srd-5.2.1:mechanic/skills/deception",
				"srd-5.2.1:mechanic/skills/history",
				"srd-5.2.1:mechanic/skills/insight",
				"srd-5.2.1:mechanic/skills/intimidation",
				"srd-5.2.1:mechanic/skills/investigation",
				"srd-5.2.1:mechanic/skills/medicine",
				"srd-5.2.1:mechanic/skills/nature",
				"srd-5.2.1:mechanic/skills/perception",
				"srd-5.2.1:mechanic/skills/performance",
				"srd-5.2.1:mechanic/skills/persuasion",
				"srd-5.2.1:mechanic/skills/religion",
				"srd-5.2.1:mechanic/skills/sleight-of-hand",
				"srd-5.2.1:mechanic/skills/stealth",
				"srd-5.2.1:mechanic/skills/survival",
			},
			Select: 3,
		},
		{
			ID: "srd-5.2.1:character/choice/bard/college-of-lore/magical-discoveries", Kind: "choice", Name: "Magical Discoveries choice",
			Sources:      []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-6-magical-discoveries", Page: 35}},
			ParentID:     "srd-5.2.1:character/feature/bard/college-of-lore/magical-discoveries",
			ChoiceRefs:   []string{},
			ResourceRefs: []string{},
			PrerequisiteRefs: []string{
				"srd-5.2.1:character/prerequisite/bard/college-of-lore/magical-discoveries-spell-eligibility",
			},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/choose-spells-from-listed-lists-in-any-combination",
			},
			OptionKeys: []string{
				"srd-5.2.1:mechanic/cleric/spell-list/cantrip-or-slot-eligible",
				"srd-5.2.1:mechanic/druid/spell-list/cantrip-or-slot-eligible",
				"srd-5.2.1:mechanic/wizard/spell-list/cantrip-or-slot-eligible",
			},
			Select: 2,
		},
		{
			ID: "srd-5.2.1:character/prerequisite/bard/college-of-lore/magical-discoveries-spell-eligibility", Kind: "prerequisite", Name: "Magical Discoveries spell eligibility",
			Sources:          []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-6-magical-discoveries", Page: 35}},
			ParentID:         "srd-5.2.1:character/choice/bard/college-of-lore/magical-discoveries",
			ChoiceRefs:       []string{},
			ResourceRefs:     []string{},
			PrerequisiteRefs: []string{},
			RequiredMechanics: []string{
				"srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/cantrip-or-level-with-slots-in-bard-features-table",
			},
			OptionKeys: []string{},
			Requirement: &CharacterRequirement{
				Key: "srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/cantrip-or-level-with-slots-in-bard-features-table",
				RecordRefs: []string{
					"srd-5.2.1:character/class/bard",
					"srd-5.2.1:character/resource/bard/spell-slot-1",
					"srd-5.2.1:character/resource/bard/spell-slot-2",
					"srd-5.2.1:character/resource/bard/spell-slot-3",
					"srd-5.2.1:character/resource/bard/spell-slot-4",
					"srd-5.2.1:character/resource/bard/spell-slot-5",
					"srd-5.2.1:character/resource/bard/spell-slot-6",
					"srd-5.2.1:character/resource/bard/spell-slot-7",
					"srd-5.2.1:character/resource/bard/spell-slot-8",
					"srd-5.2.1:character/resource/bard/spell-slot-9",
				},
				Parameters: map[string]int{}},
		},
	}
}
