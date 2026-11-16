package srd521

// P64 graph oracle. These predicates are source metadata, never runtime rules.
func thiefSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/rogue/thief/"
	const base = "rogue/thief/"
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + base + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 64}}, ParentID: parent,
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
	sub := add("subclass", "identity", "Thief", "subclasses/rogue-thief", c+"class/rogue", []string{"subclass"}, nil)
	pred(sub, "level", "class-level", []string{"class/rogue"}, map[string]int{"minimum-level": 3})
	pred(sub, "selection", "requires-canonical-thief-subclass-selection", []string{"choice/rogue/rogue-subclass"}, nil)
	feature := func(slug, name, source string, level int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, source, out[sub].ID, keys, nums)
		pred(i, "level", "class-level", []string{"class/rogue"}, map[string]int{"minimum-level": level})
		return i
	}
	choice := func(owner int, slug, name string, options []string) int {
		i := add("choice", slug, name, out[owner].Sources[0].InventoryID[len("srd-5.2.1:planning/"):], out[owner].ID, []string{"choose-one/" + slug}, nil)
		out[i].Select = 1
		for _, option := range options {
			out[i].OptionKeys = append(out[i].OptionKeys, m+option)
		}
		out[owner].ChoiceRefs = append(out[owner].ChoiceRefs, out[i].ID)
		return i
	}
	fast := feature("fast-hands", "Fast Hands", "subclass-features/rogue-level-3-fast-hands", 3, []string{"fast-hands/bonus-action"}, nil)
	action := choice(fast, "fast-hands", "Fast Hands options", []string{"fast-hands/sleight-of-hand", "fast-hands/use-an-object"})
	hand := choice(action, "sleight-of-hand", "Sleight of Hand targets", []string{"sleight-of-hand/pick-lock-with-thieves-tools", "sleight-of-hand/disarm-trap-with-thieves-tools", "sleight-of-hand/pick-pocket"})
	pred(hand, "check", "sleight-of-hand/dexterity-sleight-of-hand-check", []string{"choice/rogue/thief/sleight-of-hand"}, nil)
	obj := choice(action, "use-an-object", "Use an Object actions", []string{"use-an-object/utilize-action", "use-an-object/magic-action-use-magic-item-requiring-that-action"})
	pred(obj, "selection", "fast-hands/conditional-on-use-an-object-option", []string{"choice/rogue/thief/fast-hands"}, nil)
	pred(hand, "selection", "fast-hands/conditional-on-sleight-of-hand-option", []string{"choice/rogue/thief/fast-hands"}, nil)
	feature("second-story-work", "Second-Story Work", "subclass-features/rogue-level-3-second-story-work", 3, []string{
		"climber/climb-speed-equals-speed", "jumper/may-use-dexterity-instead-of-strength-for-jump-distance"}, nil)
	sneak := feature("supreme-sneak", "Supreme Sneak", "subclass-features/rogue-level-9-supreme-sneak", 9, []string{"supreme-sneak/add-cunning-strike-option"}, nil)
	strike := choice(sneak, "stealth-attack", "Stealth Attack", []string{"stealth-attack/retain-hide-invisible-if-end-turn-behind-required-cover"})
	pred(strike, "strike", "stealth-attack/extends-canonical-cunning-strike-pay-sneak-attack-dice-before-roll", []string{"choice/rogue/cunning-strike", "feature/rogue/sneak-attack"}, map[string]int{"cost-dice": 1, "die-sides": 6})
	pred(strike, "hide", "stealth-attack/requires-hide-actions-invisible-condition-on-self", []string{"choice/rogue/cunning-action"}, nil)
	cover := choice(strike, "stealth-cover", "Stealth Attack end-turn cover", []string{"stealth-attack/end-turn-behind-three-quarters-cover", "stealth-attack/end-turn-behind-total-cover"})
	pred(cover, "timing", "stealth-attack/this-attack-does-not-end-that-condition-only-if-end-turn-behind-cover", []string{"choice/rogue/thief/stealth-attack"}, nil)
	device := feature("use-magic-device", "Use Magic Device", "subclass-features/rogue-level-13-use-magic-device", 13, []string{
		"attunement/up-to-maximum-magic-items-at-once",
		"charges/whenever-using-magic-item-property-that-expends-charges-roll-die",
		"charges/on-success-use-property-without-expending-charges",
		"scrolls/any-spell-scroll-intelligence-spellcasting-ability",
		"attunement/maximum-items", "charges/dice-count", "charges/die-sides", "charges/success-roll"}, map[string]int{
		"attunement/maximum-items": 4, "charges/dice-count": 1, "charges/die-sides": 6, "charges/success-roll": 6})
	scroll := choice(device, "scrolls", "Spell Scroll level cases", []string{"scrolls/cantrip-or-level-one-cast-reliably", "scrolls/higher-level-requires-intelligence-arcana-check-dc-base-plus-spell-level"})
	pred(scroll, "level", "scrolls/reliable-through-level-otherwise-higher-level-check", []string{"choice/rogue/thief/scrolls"}, map[string]int{"reliable-maximum-spell-level": 1, "check-dc-base": 10})
	result := choice(scroll, "scroll-check", "Higher-level scroll check outcomes", []string{"scrolls/success-cast-spell-from-scroll", "scrolls/failure-scroll-disintegrates"})
	pred(result, "condition", "scrolls/outcomes-only-for-higher-level-scroll-check", []string{"choice/rogue/thief/scrolls"}, nil)
	feature("thiefs-reflexes", "Thief’s Reflexes", "subclass-features/rogue-level-17-thiefs-reflexes", 17, []string{
		"reflexes/may-take-two-turns-first-round-any-combat", "reflexes/first-turn-normal-initiative", "reflexes/second-turn-own-initiative-minus-offset",
		"reflexes/turn-count", "reflexes/combat-round", "reflexes/second-turn-initiative-offset"}, map[string]int{
		"reflexes/turn-count": 2, "reflexes/combat-round": 1, "reflexes/second-turn-initiative-offset": 10})
	return out
}
