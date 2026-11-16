package srd521

// Independent transcription of the four p74 headings. The complete structs
// deliberately include empty fields so invented choices, pools or scaling fail.
func invocations03SourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/warlock/invocations/"
	var out []CharacterRecord
	for _, fact := range []struct {
		slug, name string
		level      int
	}{
		{"thirsting-blade", "Thirsting Blade", 5},
		{"visions-of-distant-realms", "Visions of Distant Realms", 9},
		{"whispers-of-the-grave", "Whispers of the Grave", 7},
		{"witch-sight", "Witch Sight", 15},
	} {
		source := []CharacterSource{{InventoryID: "srd-5.2.1:planning/invocations/" + fact.slug, Page: 74}}
		feature := CharacterRecord{ID: c + "feature/warlock/invocations/" + fact.slug, Kind: "feature", Name: fact.name,
			Sources: source, ParentID: c + "choice/warlock/invocations", ChoiceRefs: []string{}, ResourceRefs: []string{},
			PrerequisiteRefs: []string{}, RequiredMechanics: []string{m + fact.slug, m + "not-repeatable"}, OptionKeys: []string{}}
		var predicates []CharacterRecord
		pred := func(suffix, key string, refs []string, nums map[string]int) {
			r := CharacterRecord{ID: c + "prerequisite/warlock/invocations/" + fact.slug + "-" + suffix, Kind: "prerequisite", Name: fact.name + " " + suffix,
				Sources: source, ParentID: feature.ID, ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{},
				RequiredMechanics: []string{m + key}, OptionKeys: []string{},
				Requirement: &CharacterRequirement{Key: m + key, RecordRefs: refs, Parameters: nums}}
			feature.PrerequisiteRefs = append(feature.PrerequisiteRefs, r.ID)
			predicates = append(predicates, r)
		}
		pred("level", "warlock-minimum-level", []string{c + "class/warlock"}, map[string]int{"minimum-level": fact.level})
		pred("selection", "requires-canonical-invocation-selection", []string{c + "choice/warlock/invocations"}, map[string]int{})
		switch fact.slug {
		case "thirsting-blade":
			pred("pact", "requires-pact-of-the-blade-invocation", []string{c + "feature/warlock/invocations/pact-of-the-blade"}, map[string]int{})
			feature.RequiredMechanics = append(feature.RequiredMechanics, m+"thirsting-blade/extra-attack-for-pact-weapon-only", m+"thirsting-blade/when-take-attack-action-on-your-turn", m+"thirsting-blade/attacks-instead-of-once")
			feature.Parameters = map[string]int{m + "thirsting-blade/attacks-instead-of-once": 2}
		case "visions-of-distant-realms":
			feature.RequiredMechanics = append(feature.RequiredMechanics, "srd-5.2.1:mechanic/spells/arcane-eye", m+"visions-of-distant-realms/can-cast", m+"visions-of-distant-realms/without-expending-spell-slot")
		case "whispers-of-the-grave":
			feature.RequiredMechanics = append(feature.RequiredMechanics, "srd-5.2.1:mechanic/spells/speak-with-dead", m+"whispers-of-the-grave/can-cast", m+"whispers-of-the-grave/without-expending-spell-slot")
		case "witch-sight":
			feature.RequiredMechanics = append(feature.RequiredMechanics, m+"witch-sight/truesight", m+"witch-sight/range-feet")
			feature.Parameters = map[string]int{m + "witch-sight/range-feet": 30}
		}
		out = append(out, feature)
		out = append(out, predicates...)
	}
	return out
}
