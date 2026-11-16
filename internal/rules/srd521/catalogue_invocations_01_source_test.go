package srd521

// Independent pp72–73 graph transcription. No fixture or Python output is read.
func invocations01SourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/warlock/invocations/"
	const b = "warlock/invocations/"
	var out []CharacterRecord
	add := func(kind, slug, name, source string, page int, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + b + slug, Kind: kind, Name: name, Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/invocations/" + source, Page: page}}, ParentID: parent, ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, OptionKeys: []string{}, RequiredMechanics: []string{}}
		for _, k := range keys {
			r.RequiredMechanics = append(r.RequiredMechanics, m+k)
		}
		if len(nums) > 0 {
			r.Parameters = map[string]int{}
			for k, v := range nums {
				r.Parameters[m+k] = v
			}
		}
		// Numeric keys are appended in the explicit source order below.
		out = append(out, r)
		return len(out) - 1
	}
	numbers := func(i int, keys ...string) {
		for _, k := range keys {
			out[i].RequiredMechanics = append(out[i].RequiredMechanics, m+k)
		}
	}
	pred := func(i int, suffix, key string, refs []string, nums map[string]int) {
		o := out[i]
		slug := o.ID[len(c+o.Kind+"/"+b):]
		n := add("prerequisite", slug+"-"+suffix, o.Name+" "+suffix, o.Sources[0].InventoryID[len("srd-5.2.1:planning/invocations/"):], o.Sources[0].Page, o.ID, []string{key}, nil)
		operands := []string{}
		for _, ref := range refs {
			operands = append(operands, c+ref)
		}
		if nums == nil {
			nums = map[string]int{}
		}
		out[n].Requirement = &CharacterRequirement{Key: m + key, RecordRefs: operands, Parameters: nums}
		out[i].PrerequisiteRefs = append(out[i].PrerequisiteRefs, out[n].ID)
	}
	feature := func(slug, name string, page, level int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, slug, page, c+"choice/warlock/invocations", keys, nums)
		if level > 0 {
			pred(i, "level", "class-level", []string{"class/warlock"}, map[string]int{"minimum-level": level})
		}
		return i
	}
	choice := func(i int, suffix string, keys, options []string) int {
		o := out[i]
		slug := o.ID[len(c+o.Kind+"/"+b):]
		n := add("choice", slug+"-"+suffix, o.Name+" "+suffix, slug, o.Sources[0].Page, o.ID, keys, nil)
		out[n].Select = 1
		for _, k := range options {
			out[n].OptionKeys = append(out[n].OptionKeys, "srd-5.2.1:mechanic/"+k)
		}
		out[i].ChoiceRefs = append(out[i].ChoiceRefs, out[n].ID)
		return n
	}
	a := feature("agonizing-blast", "Agonizing Blast", 72, 2, []string{
		"agonizing-blast/add-charisma-modifier-to-chosen-spell-damage-rolls",
		"agonizing-blast/repeatable-different-eligible-cantrip-each-time",
	}, nil)
	q := choice(a, "cantrip", []string{"agonizing-blast/choose-known-warlock-cantrip-that-deals-damage"}, []string{"warlock/known-damaging-cantrips"})
	pred(q, "eligibility", "known-damaging-warlock-cantrip", []string{"choice/warlock/cantrips", "feature/warlock/pact-magic"}, nil)
	a = feature("armor-of-shadows", "Armor of Shadows", 72, 0, []string{"armor-of-shadows/cast-on-self-without-spell-slot"}, nil)
	out[a].RequiredMechanics = append(out[a].RequiredMechanics, "srd-5.2.1:mechanic/spells/mage-armor")
	pred(a, "spellcasting", "canonical-pact-magic-spellcasting", []string{"feature/warlock/pact-magic"}, nil)
	a = feature("ascendant-step", "Ascendant Step", 72, 5, []string{"ascendant-step/cast-on-self-without-spell-slot"}, nil)
	out[a].RequiredMechanics = append(out[a].RequiredMechanics, "srd-5.2.1:mechanic/spells/levitate")
	pred(a, "spellcasting", "canonical-pact-magic-spellcasting", []string{"feature/warlock/pact-magic"}, nil)
	a = feature("devils-sight", "Devil’s Sight", 72, 2, []string{"devils-sight/see-normally-in-dim-light-and-darkness-magical-and-nonmagical"}, map[string]int{"devils-sight/range-feet": 120})
	numbers(a, "devils-sight/range-feet")
	a = feature("devouring-blade", "Devouring Blade", 72, 12, []string{
		"requires/thirsting-blade-invocation",
		"devouring-blade/thirsting-blade-extra-attack-replacement",
	}, map[string]int{"devouring-blade/extra-attacks": 2, "devouring-blade/replaced-extra-attacks": 1})
	numbers(a, "devouring-blade/extra-attacks", "devouring-blade/replaced-extra-attacks")
	pred(a, "invocation", "requires/thirsting-blade-invocation", []string{"choice/warlock/invocations"}, nil)
	feature("eldritch-mind", "Eldritch Mind", 72, 0, []string{"eldritch-mind/advantage-on-constitution-saves-to-maintain-concentration"}, nil)
	a = feature("eldritch-smite", "Eldritch Smite", 72, 5, []string{
		"requires/pact-of-the-blade-invocation",
		"eldritch-smite/optional-on-hit-creature-with-your-pact-weapon",
		"eldritch-smite/expend-pact-magic-slot",
		"eldritch-smite/extra-force-damage/base-plus-dice-per-slot-level",
		"eldritch-smite/optional-prone-only-if-target-huge-or-smaller",
	}, map[string]int{
		"eldritch-smite/maximum-uses-per-turn": 1,
		"eldritch-smite/slot-cost":             1,
		"eldritch-smite/base-dice":             1,
		"eldritch-smite/dice-per-slot-level":   1,
		"eldritch-smite/die-sides":             8,
	})
	numbers(a, "eldritch-smite/maximum-uses-per-turn", "eldritch-smite/slot-cost", "eldritch-smite/base-dice", "eldritch-smite/dice-per-slot-level", "eldritch-smite/die-sides")
	pred(a, "invocation", "requires/pact-of-the-blade-invocation", []string{"choice/warlock/invocations"}, nil)
	pred(a, "slot", "eldritch-smite/canonical-slot-cost-and-level", []string{"resource/warlock/pact-slots", "feature/warlock/pact-magic"}, map[string]int{"slots-required": 1})
	choice(a, "prone", []string{"eldritch-smite/choose-whether-to-apply-prone"}, []string{"conditions/prone", "warlock/invocations/eldritch-smite/no-prone"})
	a = feature("eldritch-spear", "Eldritch Spear", 73, 2, []string{
		"eldritch-spear/on-cast-increase-chosen-spell-range-by-feet-per-warlock-level",
		"eldritch-spear/repeatable-different-eligible-cantrip-each-time",
	}, map[string]int{"eldritch-spear/feet-per-warlock-level": 30})
	numbers(a, "eldritch-spear/feet-per-warlock-level")
	pred(a, "formula", "eldritch-spear/range-formula-warlock-level", []string{"class/warlock"}, nil)
	q = choice(a, "cantrip", []string{
		"eldritch-spear/choose-known-damaging-warlock-cantrip-with-range-at-least-ten-feet",
	}, []string{"warlock/known-damaging-cantrips-with-range-at-least-ten-feet"})
	pred(q, "eligibility", "known-damaging-warlock-cantrip-minimum-range", []string{"choice/warlock/cantrips", "feature/warlock/pact-magic"}, map[string]int{"minimum-range-feet": 10})
	a = feature("fiendish-vigor", "Fiendish Vigor", 73, 2, []string{
		"fiendish-vigor/cast-false-life-on-self-without-spell-slot",
		"fiendish-vigor/temporary-hit-points-use-highest-number-on-die-without-rolling",
	}, nil)
	out[a].RequiredMechanics = append(out[a].RequiredMechanics, "srd-5.2.1:mechanic/spells/false-life")
	pred(a, "spellcasting", "canonical-pact-magic-spellcasting", []string{"feature/warlock/pact-magic"}, nil)
	a = feature("gaze-of-two-minds", "Gaze of Two Minds", 73, 5, []string{
		"gaze-of-two-minds/bonus-action-touch-willing-creature-perceive-through-its-senses",
		"gaze-of-two-minds/duration-until-end-of-your-next-turn",
		"gaze-of-two-minds/subsequent-turn-bonus-action-maintain-only-on-same-plane",
		"gaze-of-two-minds/maintenance-extends-until-end-of-your-next-turn-otherwise-ends",
		"gaze-of-two-minds/benefit-from-target-special-senses",
		"gaze-of-two-minds/while-perceiving-cast-from-either-space-only-within-range",
	}, map[string]int{"gaze-of-two-minds/casting-range-feet": 60})
	numbers(a, "gaze-of-two-minds/casting-range-feet")
	choice(a, "target", []string{"gaze-of-two-minds/choose-willing-touched-creature"}, []string{"targets/willing-touched-creature"})
	choice(a, "casting-origin", []string{"gaze-of-two-minds/choose-spell-origin-within-range"}, []string{"spaces/self", "spaces/connected-creature"})
	pred(a, "spellcasting", "canonical-pact-magic-spellcasting", []string{"feature/warlock/pact-magic"}, nil)
	a = feature("gift-of-the-depths", "Gift of the Depths", 73, 5, []string{
		"gift-of-the-depths/breathe-underwater",
		"gift-of-the-depths/swim-speed-equals-your-speed",
		"gift-of-the-depths/cast-water-breathing-without-spell-slot",
		"gift-of-the-depths/recover-casting-on-finish-long-rest",
	}, map[string]int{"gift-of-the-depths/casting-uses": 1})
	numbers(a, "gift-of-the-depths/casting-uses")
	out[a].RequiredMechanics = append(out[a].RequiredMechanics, "srd-5.2.1:mechanic/spells/water-breathing")
	pred(a, "spellcasting", "canonical-pact-magic-spellcasting", []string{"feature/warlock/pact-magic"}, nil)
	a = feature("gift-of-the-protectors", "Gift of the Protectors", 73, 9, []string{
		"requires/pact-of-the-tome-invocation",
		"gift-of-the-protectors/new-page-in-book-of-shadows-when-conjured",
		"gift-of-the-protectors/with-your-permission-creature-takes-action-to-write-own-name",
		"gift-of-the-protectors/name-capacity-equals-charisma-modifier-minimum-one",
		"gift-of-the-protectors/named-creature-reduced-to-zero-hit-points-but-not-killed-outright",
		"gift-of-the-protectors/trigger-drops-creature-to-one-hit-point-instead",
		"gift-of-the-protectors/after-trigger-no-creature-benefits-until-you-finish-long-rest",
		"gift-of-the-protectors/magic-action-touch-page-to-erase-name",
	}, map[string]int{
		"gift-of-the-protectors/minimum-names":          1,
		"gift-of-the-protectors/replacement-hit-points": 1,
		"gift-of-the-protectors/shared-trigger-uses":    1,
	})
	numbers(a, "gift-of-the-protectors/minimum-names", "gift-of-the-protectors/replacement-hit-points", "gift-of-the-protectors/shared-trigger-uses")
	pred(a, "invocation", "requires/pact-of-the-tome-invocation", []string{"choice/warlock/invocations"}, nil)
	choice(a, "names", []string{
		"gift-of-the-protectors/choose-permitted-creature-names-up-to-symbolic-capacity",
	}, []string{"targets/permitted-creature-writing-own-name"})
	choice(a, "erase", []string{"gift-of-the-protectors/choose-name-to-erase"}, []string{"targets/name-on-touched-page"})
	return out
}
