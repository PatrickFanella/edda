package srd521

// P73–74 semantic graph oracle. All clauses are unavailable source metadata.
// Kept as Go constants so ordinary tests do not depend on Python or the PDF.
func invocations02SourceRecords() []CharacterRecord {
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, page int, keys []string) *CharacterRecord {
		out = append(out, CharacterRecord{ID: "srd-5.2.1:character/" + kind + "/" + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/invocations/" + source, Page: page}}, ParentID: parent,
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, RequiredMechanics: keys, OptionKeys: []string{}})
		return &out[len(out)-1]
	}
	var r *CharacterRecord

	r = add("feature", "warlock/invocations/investment-of-the-chain-master", "Investment of the Chain Master", "investment-of-the-chain-master", "srd-5.2.1:character/choice/warlock/invocations", 73, []string{
		"srd-5.2.1:mechanic/warlock/invocations/investment-of-the-chain-master",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/warlock/invocations/investment/when-cast-find-familiar/infuse-summoned-familiar",
		"srd-5.2.1:mechanic/spells/find-familiar",
		"srd-5.2.1:mechanic/warlock/invocations/investment/grant-chosen-speed",
		"srd-5.2.1:mechanic/warlock/invocations/investment/speed-feet",
		"srd-5.2.1:mechanic/warlock/invocations/investment/bonus-action/command-familiar-attack-action",
		"srd-5.2.1:mechanic/warlock/invocations/investment/optional-replace-only-bludgeoning-piercing-slashing",
		"srd-5.2.1:mechanic/warlock/invocations/investment/if-familiar-forces-save/use-your-spell-save-dc",
		"srd-5.2.1:mechanic/warlock/invocations/investment/when-familiar-takes-damage/reaction/resistance-against-that-damage",
	})
	r.ChoiceRefs = []string{
		"srd-5.2.1:character/choice/investment-of-the-chain-master/movement",
		"srd-5.2.1:character/choice/investment-of-the-chain-master/damage",
	}
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/investment-of-the-chain-master/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/investment-of-the-chain-master/level",
		"srd-5.2.1:character/prerequisite/investment-of-the-chain-master/pact",
		"srd-5.2.1:character/prerequisite/investment-of-the-chain-master/save-dc",
	}
	r.Parameters = map[string]int{"srd-5.2.1:mechanic/warlock/invocations/investment/speed-feet": 40}

	r = add("prerequisite", "investment-of-the-chain-master/canonical-eligibility", "Investment of the Chain Master canonical-eligibility", "investment-of-the-chain-master", "srd-5.2.1:character/feature/warlock/invocations/investment-of-the-chain-master", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "investment-of-the-chain-master/level", "Investment of the Chain Master level", "investment-of-the-chain-master", "srd-5.2.1:character/feature/warlock/invocations/investment-of-the-chain-master", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 5}}

	r = add("prerequisite", "investment-of-the-chain-master/pact", "Investment of the Chain Master pact", "investment-of-the-chain-master", "srd-5.2.1:character/feature/warlock/invocations/investment-of-the-chain-master", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/requires-pact-invocation"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/requires-pact-invocation", RecordRefs: []string{"srd-5.2.1:character/feature/warlock/invocations/pact-of-the-chain"}, Parameters: map[string]int{}}

	r = add("choice", "investment-of-the-chain-master/movement", "Investment of the Chain Master movement", "investment-of-the-chain-master", "srd-5.2.1:character/feature/warlock/invocations/investment-of-the-chain-master", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/investment/speed-choice"})
	r.OptionKeys = []string{"srd-5.2.1:mechanic/movement/fly", "srd-5.2.1:mechanic/movement/swim"}
	r.Select = 1

	r = add("choice", "investment-of-the-chain-master/damage", "Investment of the Chain Master damage", "investment-of-the-chain-master", "srd-5.2.1:character/feature/warlock/invocations/investment-of-the-chain-master", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/investment/replacement-damage"})
	r.OptionKeys = []string{"srd-5.2.1:mechanic/damage/necrotic", "srd-5.2.1:mechanic/damage/radiant"}
	r.Select = 1

	r = add("prerequisite", "investment-of-the-chain-master/save-dc", "Investment of the Chain Master save-dc", "investment-of-the-chain-master", "srd-5.2.1:character/feature/warlock/invocations/investment-of-the-chain-master", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/investment/canonical-spell-save-dc"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/investment/canonical-spell-save-dc", RecordRefs: []string{"srd-5.2.1:character/feature/warlock/pact-magic"}, Parameters: map[string]int{}}

	r = add("feature", "warlock/invocations/lessons-of-the-first-ones", "Lessons of the First Ones", "lessons-of-the-first-ones", "srd-5.2.1:character/choice/warlock/invocations", 73, []string{
		"srd-5.2.1:mechanic/warlock/invocations/lessons-of-the-first-ones",
		"srd-5.2.1:mechanic/warlock/invocations/repeatable/different-eligible-selection",
		"srd-5.2.1:mechanic/warlock/invocations/lessons/gain-origin-feat",
	})
	r.ChoiceRefs = []string{"srd-5.2.1:character/choice/lessons-of-the-first-ones/origin-feat"}
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/lessons-of-the-first-ones/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/lessons-of-the-first-ones/level",
		"srd-5.2.1:character/prerequisite/lessons-of-the-first-ones/feat-owners",
	}

	r = add("prerequisite", "lessons-of-the-first-ones/canonical-eligibility", "Lessons of the First Ones canonical-eligibility", "lessons-of-the-first-ones", "srd-5.2.1:character/feature/warlock/invocations/lessons-of-the-first-ones", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "lessons-of-the-first-ones/level", "Lessons of the First Ones level", "lessons-of-the-first-ones", "srd-5.2.1:character/feature/warlock/invocations/lessons-of-the-first-ones", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 2}}

	r = add("choice", "lessons-of-the-first-ones/origin-feat", "Lessons of the First Ones origin-feat", "lessons-of-the-first-ones", "srd-5.2.1:character/feature/warlock/invocations/lessons-of-the-first-ones", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/lessons/different-origin-feat-each-time"})
	r.OptionKeys = []string{"srd-5.2.1:mechanic/feats/origin"}
	r.Select = 1

	r = add("prerequisite", "lessons-of-the-first-ones/feat-owners", "Lessons of the First Ones feat-owners", "lessons-of-the-first-ones", "srd-5.2.1:character/feature/warlock/invocations/lessons-of-the-first-ones", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/lessons/available-origin-feat-owners"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/lessons/available-origin-feat-owners", RecordRefs: []string{
		"srd-5.2.1:character/feat/alert",
		"srd-5.2.1:character/feat/magic-initiate",
		"srd-5.2.1:character/feat/savage-attacker",
		"srd-5.2.1:character/feat/skilled",
	}, Parameters: map[string]int{}}

	r = add("feature", "warlock/invocations/lifedrinker", "Lifedrinker", "lifedrinker", "srd-5.2.1:character/choice/warlock/invocations", 73, []string{
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/optional-on-hit-creature-with-pact-weapon",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/maximum-per-turn",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/extra-damage-to-hit-creature",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/dice",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/die-sides",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/optional-expend-own-hit-point-die-and-roll-it",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/hit-point-dice-cost",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/self-healing/expended-die-roll-plus-constitution",
		"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/minimum-healing",
	})
	r.ChoiceRefs = []string{"srd-5.2.1:character/choice/lifedrinker/damage"}
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/lifedrinker/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/lifedrinker/level",
		"srd-5.2.1:character/prerequisite/lifedrinker/pact",
		"srd-5.2.1:character/prerequisite/lifedrinker/weapon",
	}
	r.Parameters = map[string]int{"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/maximum-per-turn": 1, "srd-5.2.1:mechanic/warlock/invocations/lifedrinker/dice": 1, "srd-5.2.1:mechanic/warlock/invocations/lifedrinker/die-sides": 6, "srd-5.2.1:mechanic/warlock/invocations/lifedrinker/hit-point-dice-cost": 1, "srd-5.2.1:mechanic/warlock/invocations/lifedrinker/minimum-healing": 1}

	r = add("prerequisite", "lifedrinker/canonical-eligibility", "Lifedrinker canonical-eligibility", "lifedrinker", "srd-5.2.1:character/feature/warlock/invocations/lifedrinker", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "lifedrinker/level", "Lifedrinker level", "lifedrinker", "srd-5.2.1:character/feature/warlock/invocations/lifedrinker", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 9}}

	r = add("prerequisite", "lifedrinker/pact", "Lifedrinker pact", "lifedrinker", "srd-5.2.1:character/feature/warlock/invocations/lifedrinker", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/requires-pact-invocation"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/requires-pact-invocation", RecordRefs: []string{"srd-5.2.1:character/feature/warlock/invocations/pact-of-the-blade"}, Parameters: map[string]int{}}

	r = add("choice", "lifedrinker/damage", "Lifedrinker damage", "lifedrinker", "srd-5.2.1:character/feature/warlock/invocations/lifedrinker", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/damage-choice"})
	r.OptionKeys = []string{
		"srd-5.2.1:mechanic/damage/necrotic",
		"srd-5.2.1:mechanic/damage/psychic",
		"srd-5.2.1:mechanic/damage/radiant",
	}
	r.Select = 1

	r = add("prerequisite", "lifedrinker/weapon", "Lifedrinker weapon", "lifedrinker", "srd-5.2.1:character/feature/warlock/invocations/lifedrinker", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/lifedrinker/uses-bonded-pact-weapon"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/lifedrinker/uses-bonded-pact-weapon", RecordRefs: []string{"srd-5.2.1:character/feature/warlock/invocations/pact-of-the-blade"}, Parameters: map[string]int{}}

	r = add("feature", "warlock/invocations/mask-of-many-faces", "Mask of Many Faces", "mask-of-many-faces", "srd-5.2.1:character/choice/warlock/invocations", 73, []string{
		"srd-5.2.1:mechanic/warlock/invocations/mask-of-many-faces",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/spells/disguise-self",
		"srd-5.2.1:mechanic/warlock/invocations/casting/without-spell-slot",
	})
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/mask-of-many-faces/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/mask-of-many-faces/level",
	}

	r = add("prerequisite", "mask-of-many-faces/canonical-eligibility", "Mask of Many Faces canonical-eligibility", "mask-of-many-faces", "srd-5.2.1:character/feature/warlock/invocations/mask-of-many-faces", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "mask-of-many-faces/level", "Mask of Many Faces level", "mask-of-many-faces", "srd-5.2.1:character/feature/warlock/invocations/mask-of-many-faces", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 2}}

	r = add("feature", "warlock/invocations/master-of-myriad-forms", "Master of Myriad Forms", "master-of-myriad-forms", "srd-5.2.1:character/choice/warlock/invocations", 73, []string{
		"srd-5.2.1:mechanic/warlock/invocations/master-of-myriad-forms",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/spells/alter-self",
		"srd-5.2.1:mechanic/warlock/invocations/casting/without-spell-slot",
	})
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/master-of-myriad-forms/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/master-of-myriad-forms/level",
	}

	r = add("prerequisite", "master-of-myriad-forms/canonical-eligibility", "Master of Myriad Forms canonical-eligibility", "master-of-myriad-forms", "srd-5.2.1:character/feature/warlock/invocations/master-of-myriad-forms", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "master-of-myriad-forms/level", "Master of Myriad Forms level", "master-of-myriad-forms", "srd-5.2.1:character/feature/warlock/invocations/master-of-myriad-forms", 73, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 5}}

	r = add("feature", "warlock/invocations/misty-visions", "Misty Visions", "misty-visions", "srd-5.2.1:character/choice/warlock/invocations", 74, []string{
		"srd-5.2.1:mechanic/warlock/invocations/misty-visions",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/spells/silent-image",
		"srd-5.2.1:mechanic/warlock/invocations/casting/without-spell-slot",
	})
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/misty-visions/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/misty-visions/level",
	}

	r = add("prerequisite", "misty-visions/canonical-eligibility", "Misty Visions canonical-eligibility", "misty-visions", "srd-5.2.1:character/feature/warlock/invocations/misty-visions", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "misty-visions/level", "Misty Visions level", "misty-visions", "srd-5.2.1:character/feature/warlock/invocations/misty-visions", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 2}}

	r = add("feature", "warlock/invocations/one-with-shadows", "One with Shadows", "one-with-shadows", "srd-5.2.1:character/choice/warlock/invocations", 74, []string{
		"srd-5.2.1:mechanic/warlock/invocations/one-with-shadows",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/spells/invisibility",
		"srd-5.2.1:mechanic/warlock/invocations/casting/without-spell-slot",
		"srd-5.2.1:mechanic/warlock/invocations/casting/self-only",
		"srd-5.2.1:mechanic/warlock/invocations/shadows/only-while-in-dim-light-or-darkness",
	})
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/one-with-shadows/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/one-with-shadows/level",
	}

	r = add("prerequisite", "one-with-shadows/canonical-eligibility", "One with Shadows canonical-eligibility", "one-with-shadows", "srd-5.2.1:character/feature/warlock/invocations/one-with-shadows", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "one-with-shadows/level", "One with Shadows level", "one-with-shadows", "srd-5.2.1:character/feature/warlock/invocations/one-with-shadows", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 5}}

	r = add("feature", "warlock/invocations/otherworldly-leap", "Otherworldly Leap", "otherworldly-leap", "srd-5.2.1:character/choice/warlock/invocations", 74, []string{
		"srd-5.2.1:mechanic/warlock/invocations/otherworldly-leap",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/spells/jump",
		"srd-5.2.1:mechanic/warlock/invocations/casting/without-spell-slot",
		"srd-5.2.1:mechanic/warlock/invocations/casting/self-only",
	})
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/otherworldly-leap/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/otherworldly-leap/level",
	}

	r = add("prerequisite", "otherworldly-leap/canonical-eligibility", "Otherworldly Leap canonical-eligibility", "otherworldly-leap", "srd-5.2.1:character/feature/warlock/invocations/otherworldly-leap", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "otherworldly-leap/level", "Otherworldly Leap level", "otherworldly-leap", "srd-5.2.1:character/feature/warlock/invocations/otherworldly-leap", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 2}}

	r = add("feature", "warlock/invocations/pact-of-the-blade", "Pact of the Blade", "pact-of-the-blade", "srd-5.2.1:character/choice/warlock/invocations", 74, []string{
		"srd-5.2.1:mechanic/warlock/invocations/pact-of-the-blade",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/warlock/invocations/blade/bonus-action/conjure-in-hand-or-bond-touched-magic-weapon",
		"srd-5.2.1:mechanic/warlock/invocations/blade/magic-weapon/no-other-attuned-creature-or-warlock-bond",
		"srd-5.2.1:mechanic/warlock/invocations/blade/while-bonded/proficient-with-weapon",
		"srd-5.2.1:mechanic/warlock/invocations/blade/while-bonded/spellcasting-focus",
		"srd-5.2.1:mechanic/warlock/invocations/blade/attacks/optional-charisma-replaces-strength-or-dexterity-for-attack-and-damage",
		"srd-5.2.1:mechanic/warlock/invocations/blade/terminate/on-reuse-bonus-action",
		"srd-5.2.1:mechanic/warlock/invocations/blade/terminate/continuous-distance-exceeds-threshold-for-duration",
		"srd-5.2.1:mechanic/warlock/invocations/blade/distance-feet",
		"srd-5.2.1:mechanic/warlock/invocations/blade/duration-minutes",
		"srd-5.2.1:mechanic/warlock/invocations/blade/terminate/on-your-death",
		"srd-5.2.1:mechanic/warlock/invocations/blade/conjured-weapon-disappears-on-bond-end",
	})
	r.ChoiceRefs = []string{
		"srd-5.2.1:character/choice/pact-of-the-blade/bond-method",
		"srd-5.2.1:character/choice/pact-of-the-blade/damage",
	}
	r.PrerequisiteRefs = []string{"srd-5.2.1:character/prerequisite/pact-of-the-blade/canonical-eligibility"}
	r.Parameters = map[string]int{"srd-5.2.1:mechanic/warlock/invocations/blade/distance-feet": 5, "srd-5.2.1:mechanic/warlock/invocations/blade/duration-minutes": 1}

	r = add("prerequisite", "pact-of-the-blade/canonical-eligibility", "Pact of the Blade canonical-eligibility", "pact-of-the-blade", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-blade", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("choice", "pact-of-the-blade/bond-method", "Pact of the Blade bond-method", "pact-of-the-blade", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-blade", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/blade/bond-alternatives"})
	r.OptionKeys = []string{
		"srd-5.2.1:mechanic/warlock/pact-weapon/conjured-simple-or-martial-melee",
		"srd-5.2.1:mechanic/warlock/pact-weapon/touched-magic-weapon",
	}
	r.Select = 1

	r = add("choice", "pact-of-the-blade/damage", "Pact of the Blade damage", "pact-of-the-blade", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-blade", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/blade/each-attack/damage-choice"})
	r.OptionKeys = []string{
		"srd-5.2.1:mechanic/damage/necrotic",
		"srd-5.2.1:mechanic/damage/psychic",
		"srd-5.2.1:mechanic/damage/radiant",
		"srd-5.2.1:mechanic/warlock/pact-weapon/normal-damage",
	}
	r.Select = 1

	r = add("feature", "warlock/invocations/pact-of-the-chain", "Pact of the Chain", "pact-of-the-chain", "srd-5.2.1:character/choice/warlock/invocations", 74, []string{
		"srd-5.2.1:mechanic/warlock/invocations/pact-of-the-chain",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/warlock/invocations/chain/learn-find-familiar",
		"srd-5.2.1:mechanic/warlock/invocations/chain/cast-magic-action-without-spell-slot",
		"srd-5.2.1:mechanic/spells/find-familiar",
		"srd-5.2.1:mechanic/warlock/invocations/chain/when-you-take-attack-action",
		"srd-5.2.1:mechanic/warlock/invocations/chain/optional-forgo-own-attack/familiar-attack-with-its-reaction",
		"srd-5.2.1:mechanic/warlock/invocations/chain/own-attacks-cost",
		"srd-5.2.1:mechanic/warlock/invocations/chain/familiar-attacks",
	})
	r.ChoiceRefs = []string{"srd-5.2.1:character/choice/pact-of-the-chain/familiar"}
	r.PrerequisiteRefs = []string{"srd-5.2.1:character/prerequisite/pact-of-the-chain/canonical-eligibility"}
	r.Parameters = map[string]int{"srd-5.2.1:mechanic/warlock/invocations/chain/own-attacks-cost": 1, "srd-5.2.1:mechanic/warlock/invocations/chain/familiar-attacks": 1}

	r = add("prerequisite", "pact-of-the-chain/canonical-eligibility", "Pact of the Chain canonical-eligibility", "pact-of-the-chain", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-chain", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("choice", "pact-of-the-chain/familiar", "Pact of the Chain familiar", "pact-of-the-chain", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-chain", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/chain/choose-on-cast/normal-or-special-form"})
	r.OptionKeys = []string{
		"srd-5.2.1:mechanic/spells/find-familiar/normal-forms",
		"srd-5.2.1:mechanic/creatures/imp",
		"srd-5.2.1:mechanic/creatures/pseudodragon",
		"srd-5.2.1:mechanic/creatures/quasit",
		"srd-5.2.1:mechanic/creatures/skeleton",
		"srd-5.2.1:mechanic/creatures/sphinx-of-wonder",
		"srd-5.2.1:mechanic/creatures/sprite",
		"srd-5.2.1:mechanic/creatures/venomous-snake",
	}
	r.Select = 1

	r = add("feature", "warlock/invocations/pact-of-the-tome", "Pact of the Tome", "pact-of-the-tome", "srd-5.2.1:character/choice/warlock/invocations", 74, []string{
		"srd-5.2.1:mechanic/warlock/invocations/pact-of-the-tome",
		"srd-5.2.1:mechanic/warlock/invocations/not-repeatable",
		"srd-5.2.1:mechanic/warlock/invocations/tome/conjure-book-in-hand/end-short-or-long-rest",
		"srd-5.2.1:mechanic/warlock/invocations/tome/appearance-your-choice",
		"srd-5.2.1:mechanic/warlock/invocations/tome/magic-access-only-you",
		"srd-5.2.1:mechanic/warlock/invocations/tome/disappear-on-replacement-or-your-death",
		"srd-5.2.1:mechanic/warlock/invocations/tome/choose-when-book-appears",
		"srd-5.2.1:mechanic/warlock/invocations/tome/any-class-spell-list",
		"srd-5.2.1:mechanic/warlock/invocations/tome/selected-spells-not-already-prepared",
		"srd-5.2.1:mechanic/warlock/invocations/tome/prepared-only-while-book-on-person",
		"srd-5.2.1:mechanic/warlock/invocations/tome/chosen-spells-function-as-warlock-spells",
		"srd-5.2.1:mechanic/warlock/invocations/tome/book-spellcasting-focus",
	})
	r.ChoiceRefs = []string{
		"srd-5.2.1:character/choice/pact-of-the-tome/cantrips",
		"srd-5.2.1:character/choice/pact-of-the-tome/rituals",
	}
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/pact-of-the-tome/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/pact-of-the-tome/casting",
	}

	r = add("prerequisite", "pact-of-the-tome/canonical-eligibility", "Pact of the Tome canonical-eligibility", "pact-of-the-tome", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-tome", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("choice", "pact-of-the-tome/cantrips", "Pact of the Tome cantrips", "pact-of-the-tome", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-tome", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/tome/cantrip-choice"})
	r.OptionKeys = []string{"srd-5.2.1:mechanic/spells/any-class/cantrips"}
	r.Select = 3

	r = add("choice", "pact-of-the-tome/rituals", "Pact of the Tome rituals", "pact-of-the-tome", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-tome", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/tome/ritual-choice"})
	r.OptionKeys = []string{"srd-5.2.1:mechanic/spells/any-class/level-1-rituals"}
	r.Select = 2

	r = add("prerequisite", "pact-of-the-tome/casting", "Pact of the Tome casting", "pact-of-the-tome", "srd-5.2.1:character/feature/warlock/invocations/pact-of-the-tome", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/tome/canonical-warlock-casting"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/tome/canonical-warlock-casting", RecordRefs: []string{"srd-5.2.1:character/feature/warlock/pact-magic"}, Parameters: map[string]int{}}

	r = add("feature", "warlock/invocations/repelling-blast", "Repelling Blast", "repelling-blast", "srd-5.2.1:character/choice/warlock/invocations", 74, []string{
		"srd-5.2.1:mechanic/warlock/invocations/repelling-blast",
		"srd-5.2.1:mechanic/warlock/invocations/repeatable/different-eligible-selection",
		"srd-5.2.1:mechanic/warlock/invocations/repelling/require-known-warlock-damaging-cantrip-via-attack-roll",
		"srd-5.2.1:mechanic/warlock/invocations/repelling/when-chosen-cantrip-hits-large-or-smaller-creature",
		"srd-5.2.1:mechanic/warlock/invocations/repelling/optional-push-hit-creature/straight-away",
		"srd-5.2.1:mechanic/warlock/invocations/repelling/maximum-push-feet",
	})
	r.ChoiceRefs = []string{"srd-5.2.1:character/choice/repelling-blast/cantrip"}
	r.PrerequisiteRefs = []string{
		"srd-5.2.1:character/prerequisite/repelling-blast/canonical-eligibility",
		"srd-5.2.1:character/prerequisite/repelling-blast/level",
	}
	r.Parameters = map[string]int{"srd-5.2.1:mechanic/warlock/invocations/repelling/maximum-push-feet": 10}

	r = add("prerequisite", "repelling-blast/canonical-eligibility", "Repelling Blast canonical-eligibility", "repelling-blast", "srd-5.2.1:character/feature/warlock/invocations/repelling-blast", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/canonical-eligibility-and-replacement", RecordRefs: []string{
		"srd-5.2.1:character/feature/warlock/eldritch-invocations",
		"srd-5.2.1:character/choice/warlock/invocations",
	}, Parameters: map[string]int{}}

	r = add("prerequisite", "repelling-blast/level", "Repelling Blast level", "repelling-blast", "srd-5.2.1:character/feature/warlock/invocations/repelling-blast", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level"})
	r.Requirement = &CharacterRequirement{Key: "srd-5.2.1:mechanic/warlock/invocations/minimum-warlock-level", RecordRefs: []string{"srd-5.2.1:character/class/warlock"}, Parameters: map[string]int{"minimum-level": 2}}

	r = add("choice", "repelling-blast/cantrip", "Repelling Blast cantrip", "repelling-blast", "srd-5.2.1:character/feature/warlock/invocations/repelling-blast", 74, []string{"srd-5.2.1:mechanic/warlock/invocations/repelling/different-eligible-cantrip-each-time"})
	r.OptionKeys = []string{"srd-5.2.1:mechanic/spells/known-warlock-cantrips/damage-via-attack-roll"}
	r.Select = 1
	return out
}
