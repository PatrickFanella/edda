package srd521

// P82 graph oracle. These predicates are source metadata, never runtime rules.
func evokerSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/wizard/evoker/"
	const base = "wizard/evoker/"
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + base + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 82}}, ParentID: parent,
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
		out[n].Requirement = &CharacterRequirement{Key: m + key, RecordRefs: operands, Parameters: nums}
		out[owner].PrerequisiteRefs = append(out[owner].PrerequisiteRefs, out[n].ID)
	}
	sub := add("subclass", "identity", "Evoker", "subclasses/wizard-evoker", c+"class/wizard", []string{"subclass"}, nil)
	pred(sub, "level", "class-level", []string{"class/wizard"}, map[string]int{"minimum-level": 3})
	pred(sub, "selection", "requires-canonical-evoker-subclass-selection", []string{"choice/wizard/wizard-subclass"}, nil)
	feature := func(slug, name, source string, level int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, source, out[sub].ID, keys, nums)
		pred(i, "level", "class-level", []string{"class/wizard"}, map[string]int{"minimum-level": level})
		return i
	}
	choice := func(owner int, slug, name string, count int, options []string) int {
		i := add("choice", slug, name, out[owner].Sources[0].InventoryID[len("srd-5.2.1:planning/"):], out[owner].ID, []string{"choose/" + slug}, nil)
		out[i].Select = count
		for _, option := range options {
			out[i].OptionKeys = append(out[i].OptionKeys, m+option)
		}
		out[owner].ChoiceRefs = append(out[owner].ChoiceRefs, out[i].ID)
		return i
	}

	savant := feature("evocation-savant", "Evocation Savant", "subclass-features/wizard-level-3-evocation-savant", 3, []string{"savant/add-chosen-spells-to-spellbook-for-free"}, nil)
	initial := choice(savant, "savant-initial", "Initial Evocation spells", 2, []string{"savant/wizard-evocation-spells-no-higher-than-level-two"})
	pred(initial, "book", "savant/initial-wizard-evocation-spells-maximum-level", []string{"choice/wizard/spellbook-initial"}, map[string]int{"maximum-spell-level": 2})
	later := choice(savant, "savant-new-slots", "New spell-slot level Evocation spell", 1, []string{"savant/wizard-evocation-spell-at-a-level-with-slots"})
	pred(later, "trigger", "savant/whenever-gain-access-to-new-slot-level-in-this-class", []string{"feature/wizard/spellcasting", "choice/wizard/spellbook-level-gain", "class/wizard"}, nil)
	potent := feature("potent-cantrip", "Potent Cantrip", "subclass-features/wizard-level-3-potent-cantrip", 3, []string{
		"potent/when-cast-damaging-cantrip-at-creature", "potent/target-takes-half-cantrip-damage-if-any", "potent/no-additional-cantrip-effect", "potent/damage-divisor"}, map[string]int{"potent/damage-divisor": 2})
	choice(potent, "potent-trigger", "Alternative Potent Cantrip triggers", 1, []string{"potent/miss-with-attack-roll", "potent/target-succeeds-on-save-against-cantrip"})
	feature("sculpt-spells", "Sculpt Spells", "subclass-features/wizard-level-6-sculpt-spells", 6, []string{
		"sculpt/when-cast-evocation-affecting-other-creatures-you-can-see",
		"sculpt/may-choose-number-of-them-equal-to-base-plus-spell-level",
		"sculpt/chosen-creatures-automatically-succeed-on-saves-against-spell",
		"sculpt/no-damage-only-if-normally-half-on-successful-save",
		"sculpt/creature-count-base"}, map[string]int{"sculpt/creature-count-base": 1})
	empowered := feature("empowered-evocation", "Empowered Evocation", "subclass-features/wizard-level-10-empowered-evocation", 10, []string{
		"empowered/whenever-cast-wizard-evocation-spell",
		"empowered/may-add-intelligence-modifier-to-one-damage-roll-of-that-spell",
		"empowered/damage-roll-count"}, map[string]int{"empowered/damage-roll-count": 1})
	pred(empowered, "ability", "empowered/canonical-intelligence-spellcasting-ability", []string{"feature/wizard/spellcasting"}, nil)
	over := feature("overchannel", "Overchannel", "subclass-features/wizard-level-14-overchannel", 14, []string{
		"overchannel/when-cast-damaging-wizard-spell-with-qualifying-slot",
		"overchannel/may-deal-maximum-spell-damage-on-turn-you-cast-it",
		"overchannel/first-use-no-adverse-effect",
		"overchannel/reuse-before-finish-long-rest-necrotic-self-damage",
		"overchannel/initial-reuse-dice-per-spell-slot-level",
		"overchannel/immediately-after-casting",
		"overchannel/ignores-resistance-and-immunity",
		"overchannel/each-further-use-before-finish-long-rest-increases-damage-per-spell-level",
		"overchannel/minimum-slot-level", "overchannel/maximum-slot-level", "overchannel/initial-reuse-dice",
		"overchannel/die-sides", "overchannel/further-use-dice-increase"}, map[string]int{
		"overchannel/minimum-slot-level": 1, "overchannel/maximum-slot-level": 5,
		"overchannel/initial-reuse-dice": 2, "overchannel/die-sides": 12, "overchannel/further-use-dice-increase": 1})
	pred(over, "slots", "overchannel/uses-canonical-wizard-slot-owner-not-prepared-count-or-character-level", []string{"feature/wizard/spellcasting", "class/wizard"}, nil)
	return out
}
