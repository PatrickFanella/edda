package srd521

// P76 graph oracle. These predicates are source metadata, never runtime rules.
func fiendSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/warlock/fiend-patron/"
	const base = "warlock/fiend-patron/"
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + base + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 76}}, ParentID: parent,
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, RequiredMechanics: []string{}, OptionKeys: []string{}}
		for _, k := range keys {
			r.RequiredMechanics = append(r.RequiredMechanics, m+k)
		}
		// Numeric order is source order, supplied by keys, independent of map iteration.
		if nums != nil {
			r.Parameters = map[string]int{}
			for _, k := range keys {
				if v, ok := nums[k]; ok {
					r.Parameters[m+k] = v
				}
			}
		}
		out = append(out, r)
		return len(out) - 1
	}
	pred := func(owner int, suffix, key string, refs []string, nums map[string]int) {
		o := out[owner]
		slug := o.ID[len(c+o.Kind+"/"+base):]
		n := add("prerequisite", slug+"-"+suffix, o.Name+" "+suffix, o.Sources[0].InventoryID[len("srd-5.2.1:planning/"):], o.ID, []string{key}, nil)
		if nums == nil {
			nums = map[string]int{}
		}
		operands := []string{}
		for _, r := range refs {
			operands = append(operands, c+r)
		}
		out[n].Sources[0].Page = o.Sources[0].Page
		out[n].Requirement = &CharacterRequirement{Key: m + key, RecordRefs: operands, Parameters: nums}
		out[owner].PrerequisiteRefs = append(out[owner].PrerequisiteRefs, out[n].ID)
	}
	sub := add("subclass", "identity", "Fiend Patron", "subclasses/warlock-fiend-patron", c+"class/warlock", []string{"subclass"}, nil)
	pred(sub, "level", "class-level", []string{"class/warlock"}, map[string]int{"minimum-level": 3})
	feature := func(slug, name, source string, level int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, source, out[sub].ID, keys, nums)
		pred(i, "level", "class-level", []string{"class/warlock"}, map[string]int{"minimum-level": level})
		return i
	}
	blessing := feature("dark-ones-blessing", "Dark One’s Blessing", "subclass-features/warlock-level-3-dark-ones-blessing", 3, []string{
		"blessing/when-you-reduce-enemy-to-zero-hit-points",
		"blessing/alternative/someone-else-reduces-enemy-within-range-of-you-to-zero-hit-points",
		"blessing/self-temporary-hit-points/charisma-modifier-plus-warlock-level-minimum-one",
		"blessing/other-killer-enemy-range-feet", "blessing/temporary-hit-points-minimum"}, map[string]int{
		"blessing/other-killer-enemy-range-feet": 10, "blessing/temporary-hit-points-minimum": 1})
	pred(blessing, "formula", "blessing/formula-uses-canonical-warlock-level", []string{"class/warlock"}, nil)
	spells := feature("fiend-spells", "Fiend Spells", "subclass-features/warlock-level-3-fiend-spells", 3, []string{"fiend-spells/thereafter-always-prepared-from-specified-warlock-level"}, nil)
	out[spells].AlwaysPrepared = []CharacterPreparedSpells{
		{Level: 3, SpellKeys: []string{"srd-5.2.1:mechanic/spells/burning-hands", "srd-5.2.1:mechanic/spells/command", "srd-5.2.1:mechanic/spells/scorching-ray", "srd-5.2.1:mechanic/spells/suggestion"}},
		{Level: 5, SpellKeys: []string{"srd-5.2.1:mechanic/spells/fireball", "srd-5.2.1:mechanic/spells/stinking-cloud"}},
		{Level: 7, SpellKeys: []string{"srd-5.2.1:mechanic/spells/fire-shield", "srd-5.2.1:mechanic/spells/wall-of-fire"}},
		{Level: 9, SpellKeys: []string{"srd-5.2.1:mechanic/spells/geas", "srd-5.2.1:mechanic/spells/insect-plague"}},
	}
	for _, row := range out[spells].AlwaysPrepared {
		out[spells].RequiredMechanics = append(out[spells].RequiredMechanics, row.SpellKeys...)
	}
	pred(spells, "preparation", "fiend-spells/canonical-pact-magic-preparation-and-spellcasting", []string{"feature/warlock/pact-magic", "choice/warlock/prepared-spells"}, nil)
	feature("dark-ones-own-luck", "Dark One’s Own Luck", "subclass-features/warlock-level-6-dark-ones-own-luck", 6, []string{
		"luck/optional-on-your-ability-check-or-saving-throw", "luck/add-die-to-your-roll",
		"luck/after-seeing-roll-before-any-roll-effects", "luck/uses-equal-charisma-modifier-minimum-one",
		"luck/recover-all-expended-uses-on-finish-long-rest", "luck/dice", "luck/die-sides", "luck/minimum-uses", "luck/maximum-uses-per-roll"}, map[string]int{
		"luck/dice": 1, "luck/die-sides": 10, "luck/minimum-uses": 1, "luck/maximum-uses-per-roll": 1})
	resilience := feature("fiendish-resilience", "Fiendish Resilience", "subclass-features/warlock-level-10-fiendish-resilience", 10, []string{
		"resilience/choose-on-finish-short-or-long-rest", "resilience/self-resistance-to-chosen-type",
		"resilience/until-choose-different-type-with-this-feature"}, nil)
	choice := add("choice", "resistance-type", "Fiendish Resilience damage type", "subclass-features/warlock-level-10-fiendish-resilience", out[resilience].ID, []string{"resilience/choose-one-damage-type-other-than-force"}, nil)
	out[choice].Select = 1
	for _, typ := range []string{"acid", "bludgeoning", "cold", "fire", "lightning", "necrotic", "piercing", "poison", "psychic", "radiant", "slashing", "thunder"} {
		out[choice].OptionKeys = append(out[choice].OptionKeys, m+"resilience/damage-type/"+typ)
	}
	out[resilience].ChoiceRefs = append(out[resilience].ChoiceRefs, out[choice].ID)
	hurl := feature("hurl-through-hell", "Hurl Through Hell", "subclass-features/warlock-level-14-hurl-through-hell", 14, []string{
		"hurl/optional-on-hit-creature-with-attack-roll", "hurl/target-charisma-save-against-your-spell-save-dc",
		"hurl/on-failed-save/target-disappears-through-lower-planes", "hurl/on-failed-save/psychic-damage-only-if-target-not-fiend",
		"hurl/on-failed-save/incapacitated-until-end-of-your-next-turn",
		"hurl/return-at-end-of-your-next-turn/previous-space-or-nearest-unoccupied-space",
		"hurl/recover-use-on-finish-long-rest", "hurl/restore-use/expend-pact-magic-slot-no-action",
		"hurl/maximum-uses-per-turn", "hurl/damage-dice", "hurl/damage-die-sides", "hurl/uses", "hurl/restore-slot-cost"}, map[string]int{
		"hurl/maximum-uses-per-turn": 1, "hurl/damage-dice": 8, "hurl/damage-die-sides": 10, "hurl/uses": 1, "hurl/restore-slot-cost": 1})
	pred(hurl, "save", "hurl/canonical-charisma-spell-save-dc", []string{"feature/warlock/pact-magic"}, nil)
	pred(hurl, "restore", "hurl/restore-expended-use-with-canonical-pact-slot", []string{"resource/warlock/pact-slots", "feature/warlock/pact-magic"}, map[string]int{"slots-required": 1})
	return out
}
