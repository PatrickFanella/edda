package srd521

import (
	"fmt"
	"strings"
)

// P61 graph oracle. These predicates are source metadata, never runtime rules.
func hunterSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/ranger/hunter/"
	const base = "ranger/hunter/"
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + base + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 61}}, ParentID: parent,
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
	sub := add("subclass", "identity", "Hunter", "subclasses/ranger-hunter", c+"class/ranger", []string{"subclass"}, nil)
	pred(sub, "level", "class-level", []string{"class/ranger"}, map[string]int{"minimum-level": 3})
	feature := func(slug, name string, level int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, fmt.Sprintf("subclass-features/ranger-level-%d-%s", level, slug), out[sub].ID, keys, nums)
		pred(i, "level", "class-level", []string{"class/ranger"}, map[string]int{"minimum-level": level})
		return i
	}
	lore := feature("hunters-lore", "Hunter’s Lore", 3, []string{"hunters-lore/know-whether-marked-creature-has-immunities-resistances-vulnerabilities-and-what-they-are"}, nil)
	pred(lore, "mark", "target-marked-by-your-hunters-mark", []string{"feature/ranger/favored-enemy"}, nil)
	prey := feature("hunters-prey", "Hunter’s Prey", 3, []string{"hunters-prey/gain-one-chosen-option"}, nil)
	choice := func(owner int, slug, name string, options []string) {
		i := add("choice", slug, name, out[owner].Sources[0].InventoryID[len("srd-5.2.1:planning/"):], out[owner].ID, []string{"choose-one", "replace-chosen-option-with-other-on-finish-short-or-long-rest"}, nil)
		out[i].Select = 1
		for _, option := range options {
			out[i].OptionKeys = append(out[i].OptionKeys, m+option)
		}
		out[owner].ChoiceRefs = append(out[owner].ChoiceRefs, out[i].ID)
	}
	option := func(slug, name, source, choiceSlug string, keys []string, nums map[string]int) {
		owner := c + "choice/" + base + choiceSlug
		keys = append([]string{slug + "/only-while-chosen"}, keys...)
		i := add("prerequisite", slug, name, source, owner, keys, nil)
		if nums == nil {
			nums = map[string]int{}
		}
		out[i].Requirement = &CharacterRequirement{Key: m + keys[0], RecordRefs: []string{owner}, Parameters: map[string]int{}}
		for key, value := range nums {
			out[i].Requirement.Parameters[key[strings.Index(key, "/")+1:]] = value
		}
		for j := range out {
			if out[j].ID == owner {
				out[j].PrerequisiteRefs = append(out[j].PrerequisiteRefs, out[i].ID)
			}
		}
	}
	choice(prey, "hunters-prey", "Hunter’s Prey options", []string{"colossus-slayer", "horde-breaker"})
	option("colossus-slayer", "Colossus Slayer", "subclass-features/ranger-level-3-hunters-prey", "hunters-prey", []string{
		"colossus-slayer/on-weapon-hit-creature", "colossus-slayer/weapon-extra-damage-to-hit-target",
		"colossus-slayer/only-if-target-missing-any-hit-points", "colossus-slayer/once-per-turn",
		"colossus-slayer/damage-dice", "colossus-slayer/damage-die-sides"}, map[string]int{"colossus-slayer/damage-dice": 1, "colossus-slayer/damage-die-sides": 8})
	option("horde-breaker", "Horde Breaker", "subclass-features/ranger-level-3-hunters-prey", "hunters-prey", []string{
		"horde-breaker/once-on-each-of-your-turns-when-making-weapon-attack",
		"horde-breaker/optional-another-attack-with-same-weapon",
		"horde-breaker/target-different-creature-within-weapon-range-not-attacked-this-turn",
		"horde-breaker/target-within-feet-of-original-target"}, map[string]int{"horde-breaker/target-within-feet-of-original-target": 5})
	defense := feature("defensive-tactics", "Defensive Tactics", 7, []string{"defensive-tactics/gain-one-chosen-option"}, nil)
	choice(defense, "defensive-tactics", "Defensive Tactics options", []string{"escape-the-horde", "multiattack-defense"})
	option("escape-the-horde", "Escape the Horde", "subclass-features/ranger-level-7-defensive-tactics", "defensive-tactics", []string{"escape-the-horde/opportunity-attacks-against-you-have-disadvantage"}, nil)
	option("multiattack-defense", "Multiattack Defense", "subclass-features/ranger-level-7-defensive-tactics", "defensive-tactics", []string{
		"multiattack-defense/when-creature-hits-you-with-attack-roll", "multiattack-defense/that-creature-disadvantage-on-all-other-attack-rolls-against-you-this-turn"}, nil)
	super := feature("superior-hunters-prey", "Superior Hunter’s Prey", 11, []string{
		"superior-hunters-prey/once-per-turn-when-you-deal-damage-to-marked-creature",
		"superior-hunters-prey/optional-deal-that-spells-extra-damage",
		"superior-hunters-prey/target-different-creature-you-can-see",
		"superior-hunters-prey/target-within-feet-of-first-creature"}, map[string]int{"superior-hunters-prey/target-within-feet-of-first-creature": 30})
	pred(super, "mark", "target-marked-by-your-hunters-mark", []string{"feature/ranger/favored-enemy"}, nil)
	pred(super, "damage", "superior-hunters-prey/use-current-canonical-hunters-mark-extra-damage", []string{"feature/ranger/favored-enemy", "feature/ranger/foe-slayer"}, nil)
	feature("superior-hunters-defense", "Superior Hunter’s Defense", 15, []string{
		"superior-hunters-defense/when-you-take-damage", "superior-hunters-defense/optional-reaction",
		"superior-hunters-defense/self-resistance-to-triggering-damage-and-any-other-damage-of-same-type",
		"superior-hunters-defense/until-end-of-current-turn"}, nil)
	return out
}
