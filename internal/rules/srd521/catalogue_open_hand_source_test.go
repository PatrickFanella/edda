package srd521

// P52 graph oracle. These predicates are source metadata, never runtime rules.
func openHandSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/monk/warrior-of-the-open-hand/"
	const base = "monk/warrior-of-the-open-hand/"
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + base + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 52}}, ParentID: parent,
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
	sub := add("subclass", "identity", "Warrior of the Open Hand", "subclasses/monk-warrior-of-the-open-hand", c+"class/monk", []string{"subclass"}, nil)
	pred(sub, "level", "class-level", []string{"class/monk"}, map[string]int{"minimum-level": 3})
	feature := func(slug, name, source string, level int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, source, out[sub].ID, keys, nums)
		pred(i, "level", "class-level", []string{"class/monk"}, map[string]int{"minimum-level": level})
		return i
	}
	hand := feature("open-hand-technique", "Open Hand Technique", "subclass-features/monk-level-3-open-hand-technique", 3, []string{
		"open-hand-technique/optional-on-each-hit", "open-hand-technique/target/hit-creature", "open-hand-technique/push/maximum-distance-feet"}, map[string]int{"open-hand-technique/push/maximum-distance-feet": 15})
	pred(hand, "trigger", "open-hand-technique/hit-with-attack-granted-by-flurry-of-blows", []string{"choice/monk/flurry-of-blows"}, nil)
	choice := add("choice", "open-hand-technique", "Open Hand Technique effects", "subclass-features/monk-level-3-open-hand-technique", out[hand].ID, []string{"open-hand-technique/choose-one-effect-on-target"}, nil)
	out[choice].Select = 1
	out[choice].OptionKeys = []string{m + "open-hand-technique/addle/no-opportunity-attacks-until-start-of-target-next-turn", m + "open-hand-technique/push/strength-save-or-pushed-up-to-15-feet-away-from-you", m + "open-hand-technique/topple/dexterity-save-or-prone"}
	out[hand].ChoiceRefs = append(out[hand].ChoiceRefs, out[choice].ID)
	body := feature("wholeness-of-body", "Wholeness of Body", "subclass-features/monk-level-6-wholeness-of-body", 6, []string{
		"wholeness-of-body/bonus-action", "wholeness-of-body/target/self",
		"wholeness-of-body/regain-hit-points/one-martial-arts-die-plus-wisdom-modifier",
		"wholeness-of-body/uses/wisdom-modifier-minimum-one",
		"wholeness-of-body/recover-all-expended-uses-on-finish-long-rest",
		"wholeness-of-body/healing-minimum", "wholeness-of-body/dice", "wholeness-of-body/uses-minimum", "wholeness-of-body/uses-expended"}, map[string]int{
		"wholeness-of-body/healing-minimum": 1, "wholeness-of-body/dice": 1, "wholeness-of-body/uses-minimum": 1, "wholeness-of-body/uses-expended": 1})
	pred(body, "die", "wholeness-of-body/roll-canonical-martial-arts-die", []string{"feature/monk/martial-arts"}, nil)
	fleet := feature("fleet-step", "Fleet Step", "subclass-features/monk-level-11-fleet-step", 11, []string{"fleet-step/optional-step-of-the-wind-immediately-after-bonus-action"}, nil)
	pred(fleet, "trigger", "fleet-step/bonus-action-other-than-step-of-the-wind", []string{"choice/monk/step-of-the-wind"}, nil)
	pred(fleet, "effect", "fleet-step/use-canonical-step-of-the-wind-with-its-alternatives-and-costs", []string{"choice/monk/step-of-the-wind"}, nil)
	palm := feature("quivering-palm", "Quivering Palm", "subclass-features/monk-level-17-quivering-palm", 17, []string{
		"quivering-palm/optional-on-unarmed-strike-hit-creature",
		"quivering-palm/vibrations/imperceptible-and-harmless-unless-ended-for-damage",
		"quivering-palm/end/action-or-forgo-one-attack-of-attack-action-on-your-turn",
		"quivering-palm/end/requires-same-plane-as-target", "quivering-palm/end/target-constitution-save",
		"quivering-palm/damage/force-on-failure-half-on-success", "quivering-palm/end-harmlessly/no-action-required",
		"quivering-palm/focus-cost", "quivering-palm/forgone-attacks", "quivering-palm/damage-dice", "quivering-palm/damage-die-sides", "quivering-palm/maximum-simultaneous-creatures"}, map[string]int{
		"quivering-palm/focus-cost": 4, "quivering-palm/forgone-attacks": 1, "quivering-palm/damage-dice": 10, "quivering-palm/damage-die-sides": 12, "quivering-palm/maximum-simultaneous-creatures": 1})
	pred(palm, "focus", "quivering-palm/expend-available-focus", []string{"resource/monk/focus-points"}, map[string]int{"points-required": 4})
	pred(palm, "duration", "quivering-palm/vibration-duration-days-equals-monk-level", []string{"class/monk"}, nil)
	return out
}
