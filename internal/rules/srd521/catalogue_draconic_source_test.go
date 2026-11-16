package srd521

import "fmt"

// P69–70 graph oracle. These predicates are source metadata, never runtime rules.
func draconicSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/sorcerer/draconic-sorcery/"
	const base = "sorcerer/draconic-sorcery/"
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + base + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 70}}, ParentID: parent,
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
	sub := add("subclass", "identity", "Draconic Sorcery", "subclasses/sorcerer-draconic-sorcery", c+"class/sorcerer", []string{"subclass"}, nil)
	out[sub].Sources[0].Page = 69
	pred(sub, "level", "class-level", []string{"class/sorcerer"}, map[string]int{"minimum-level": 3})
	feature := func(slug, name string, level, page int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, "subclass-features/sorcerer-level-"+fmt.Sprint(level)+"-"+slug, out[sub].ID, keys, nums)
		out[i].Sources[0].Page = page
		pred(i, "level", "class-level", []string{"class/sorcerer"}, map[string]int{"minimum-level": level})
		return i
	}
	resilience := feature("draconic-resilience", "Draconic Resilience", 3, 69, []string{
		"resilience/self/hit-point-maximum-increases-at-level-three",
		"resilience/self/each-additional-sorcerer-level-increases-hit-point-maximum",
		"resilience/self/while-not-wearing-armor/base-ac-ten-plus-dexterity-plus-charisma-modifiers",
		"resilience/initial-hp-increase", "resilience/hp-per-additional-sorcerer-level", "resilience/base-ac-constant"}, map[string]int{
		"resilience/initial-hp-increase": 3, "resilience/hp-per-additional-sorcerer-level": 1, "resilience/base-ac-constant": 10})
	pred(resilience, "scaling", "resilience/count-only-canonical-sorcerer-levels", []string{"class/sorcerer"}, nil)
	spells := feature("draconic-spells", "Draconic Spells", 3, 70, []string{"spells/thereafter-always-prepared-from-specified-sorcerer-level"}, nil)
	out[spells].AlwaysPrepared = []CharacterPreparedSpells{
		{Level: 3, SpellKeys: []string{"srd-5.2.1:mechanic/spells/alter-self", "srd-5.2.1:mechanic/spells/chromatic-orb", "srd-5.2.1:mechanic/spells/command", "srd-5.2.1:mechanic/spells/dragons-breath"}},
		{Level: 5, SpellKeys: []string{"srd-5.2.1:mechanic/spells/fear", "srd-5.2.1:mechanic/spells/fly"}},
		{Level: 7, SpellKeys: []string{"srd-5.2.1:mechanic/spells/arcane-eye", "srd-5.2.1:mechanic/spells/charm-monster"}},
		{Level: 9, SpellKeys: []string{"srd-5.2.1:mechanic/spells/legend-lore", "srd-5.2.1:mechanic/spells/summon-dragon"}},
	}
	for _, row := range out[spells].AlwaysPrepared {
		out[spells].RequiredMechanics = append(out[spells].RequiredMechanics, row.SpellKeys...)
	}
	pred(spells, "preparation", "spells/canonical-sorcerer-preparation-slot-and-innate-sorcery-rules", []string{"feature/sorcerer/spellcasting", "choice/sorcerer/prepared-spells", "feature/sorcerer/innate-sorcery", "resource/sorcerer/innate-sorcery"}, nil)
	affinity := feature("elemental-affinity", "Elemental Affinity", 6, 70, []string{
		"affinity/self/resistance-to-chosen-damage-type",
		"affinity/when-you-cast-spell-dealing-chosen-type/optional-add-charisma-modifier-to-one-damage-roll",
		"affinity/damage-rolls-maximum"}, map[string]int{"affinity/damage-rolls-maximum": 1})
	choice := add("choice", "elemental-affinity-damage-type", "Elemental Affinity damage type", "subclass-features/sorcerer-level-6-elemental-affinity", out[affinity].ID, []string{"affinity/choose-one-draconic-damage-type"}, nil)
	out[choice].Select = 1
	for _, damage := range []string{"acid", "cold", "fire", "lightning", "poison"} {
		out[choice].OptionKeys = append(out[choice].OptionKeys, m+"affinity/damage-type/"+damage)
	}
	out[affinity].ChoiceRefs = append(out[affinity].ChoiceRefs, out[choice].ID)
	pred(affinity, "type", "affinity/use-selected-type-for-both-benefits", []string{"choice/sorcerer/draconic-sorcery/elemental-affinity-damage-type", "feature/sorcerer/spellcasting"}, nil)
	wings := feature("dragon-wings", "Dragon Wings", 14, 70, []string{
		"wings/self/bonus-action-manifest-draconic-wings-on-back", "wings/end/one-hour-or-dismiss-no-action",
		"wings/self/fly-speed-for-duration", "wings/reuse/finish-long-rest-or-restore-use",
		"wings/restore/expend-sorcery-points-no-action", "wings/duration-hours", "wings/fly-speed-feet", "wings/uses", "wings/restore-sorcery-points"}, map[string]int{
		"wings/duration-hours": 1, "wings/fly-speed-feet": 60, "wings/uses": 1, "wings/restore-sorcery-points": 3})
	pred(wings, "restore", "wings/spend-canonical-sorcery-points-with-capacity-and-recovery", []string{"resource/sorcerer/sorcery-points"}, map[string]int{"points-required": 3})
	companion := feature("dragon-companion", "Dragon Companion", 18, 70, []string{
		"companion/cast-summon-dragon/without-material-component",
		"companion/cast-summon-dragon/once-without-spell-slot", "companion/free-cast/recover-on-long-rest",
		"companion/whenever-start-casting/optional-no-concentration",
		"companion/if-no-concentration/duration-one-minute-for-that-casting-only",
		"companion/free-casts", "companion/modified-duration-minutes"}, map[string]int{
		"companion/free-casts": 1, "companion/modified-duration-minutes": 1})
	out[companion].RequiredMechanics = append(out[companion].RequiredMechanics, "srd-5.2.1:mechanic/spells/summon-dragon")
	pred(companion, "casting", "companion/use-canonical-sorcerer-spell-and-slot-rules-except-explicit-exemptions", []string{"feature/sorcerer/spellcasting", "choice/sorcerer/prepared-spells", "resource/sorcerer/spell-slot-5", "feature/sorcerer/innate-sorcery"}, nil)
	return out
}
