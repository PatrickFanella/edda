package srd521

// P56–57 graph oracle. These predicates are source metadata, never runtime rules.
func devotionSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/paladin/oath-of-devotion/"
	const base = "paladin/oath-of-devotion/"
	var out []CharacterRecord
	add := func(kind, slug, name, source, parent string, keys []string, nums map[string]int) int {
		r := CharacterRecord{ID: c + kind + "/" + base + slug, Kind: kind, Name: name,
			Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 56}}, ParentID: parent,
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
	sub := add("subclass", "identity", "Oath of Devotion", "subclasses/paladin-oath-of-devotion", c+"class/paladin", []string{"subclass"}, nil)
	pred(sub, "level", "class-level", []string{"class/paladin"}, map[string]int{"minimum-level": 3})
	feature := func(slug, name, source string, level int, keys []string, nums map[string]int) int {
		i := add("feature", slug, name, source, out[sub].ID, keys, nums)
		if level >= 15 {
			out[i].Sources[0].Page = 57
		}
		pred(i, "level", "class-level", []string{"class/paladin"}, map[string]int{"minimum-level": level})
		return i
	}
	spells := feature("oath-of-devotion-spells", "Oath of Devotion Spells", "subclass-features/paladin-level-3-oath-of-devotion-spells", 3, []string{"oath-spells/always-prepared-from-specified-paladin-level"}, nil)
	out[spells].AlwaysPrepared = []CharacterPreparedSpells{
		{Level: 3, SpellKeys: []string{"srd-5.2.1:mechanic/spells/protection-from-evil-and-good", "srd-5.2.1:mechanic/spells/shield-of-faith"}},
		{Level: 5, SpellKeys: []string{"srd-5.2.1:mechanic/spells/aid", "srd-5.2.1:mechanic/spells/zone-of-truth"}},
		{Level: 9, SpellKeys: []string{"srd-5.2.1:mechanic/spells/beacon-of-hope", "srd-5.2.1:mechanic/spells/dispel-magic"}},
		{Level: 13, SpellKeys: []string{"srd-5.2.1:mechanic/spells/freedom-of-movement", "srd-5.2.1:mechanic/spells/guardian-of-faith"}},
		{Level: 17, SpellKeys: []string{"srd-5.2.1:mechanic/spells/commune", "srd-5.2.1:mechanic/spells/flame-strike"}},
	}
	for _, row := range out[spells].AlwaysPrepared {
		out[spells].RequiredMechanics = append(out[spells].RequiredMechanics, row.SpellKeys...)
	}
	pred(spells, "preparation", "oath-spells/canonical-paladin-preparation-rules", []string{"feature/paladin/spellcasting"}, nil)
	weapon := feature("sacred-weapon", "Sacred Weapon", "subclass-features/paladin-level-3-sacred-weapon", 3, []string{
		"sacred-weapon/optional-when-taking-attack-action", "sacred-weapon/target/one-held-melee-weapon",
		"sacred-weapon/attack-rolls/add-charisma-modifier-minimum-one",
		"sacred-weapon/each-hit/choose-normal-damage-type-or-radiant",
		"sacred-weapon/end/on-reuse-or-no-longer-carrying-weapon-or-voluntarily-no-action",
		"sacred-weapon/channel-divinity-cost", "sacred-weapon/duration-minutes", "sacred-weapon/attack-bonus-minimum",
		"sacred-weapon/bright-light-radius-feet", "sacred-weapon/dim-light-additional-feet"}, map[string]int{
		"sacred-weapon/channel-divinity-cost": 1, "sacred-weapon/duration-minutes": 10, "sacred-weapon/attack-bonus-minimum": 1,
		"sacred-weapon/bright-light-radius-feet": 20, "sacred-weapon/dim-light-additional-feet": 20})
	pred(weapon, "cost", "sacred-weapon/expend-canonical-channel-divinity-with-its-capacity-and-recovery", []string{"resource/paladin/channel-divinity"}, map[string]int{"uses-required": 1})
	aura := feature("aura-of-devotion", "Aura of Devotion", "subclass-features/paladin-level-7-aura-of-devotion", 7, []string{
		"aura-of-devotion/self-and-allies-in-aura/immune-to-charmed",
		"aura-of-devotion/charmed-ally-entering-aura/condition-has-no-effect-while-there"}, nil)
	pred(aura, "scope", "aura-of-devotion/within-canonical-aura-of-protection-including-expansion-and-inactive-condition", []string{"feature/paladin/aura-of-protection", "feature/paladin/aura-expansion"}, nil)
	smite := feature("smite-of-protection", "Smite of Protection", "subclass-features/paladin-level-15-smite-of-protection", 15, []string{
		"smite-of-protection/whenever-cast-divine-smite", "smite-of-protection/self-and-allies-in-aura/half-cover",
		"smite-of-protection/aura-benefit-until-start-of-your-next-turn"}, nil)
	pred(smite, "trigger", "smite-of-protection/cast-divine-smite-with-canonical-spell-rules", []string{"feature/paladin/paladins-smite"}, nil)
	pred(smite, "scope", "smite-of-protection/within-canonical-aura-of-protection-including-expansion-and-inactive-condition", []string{"feature/paladin/aura-of-protection", "feature/paladin/aura-expansion"}, nil)
	nimbus := feature("holy-nimbus", "Holy Nimbus", "subclass-features/paladin-level-20-holy-nimbus", 20, []string{
		"holy-nimbus/bonus-action", "holy-nimbus/end/voluntarily-no-action", "holy-nimbus/recover-use/finish-long-rest",
		"holy-nimbus/restore-use/expend-level-five-slot-no-action",
		"holy-nimbus/holy-ward/self-advantage-on-saves-forced-by-fiend-or-undead",
		"holy-nimbus/radiant-damage/enemy-starts-turn-in-aura/charisma-modifier-plus-proficiency-bonus",
		"holy-nimbus/sunlight/aura-filled-with-bright-light-that-is-sunlight",
		"holy-nimbus/duration-minutes", "holy-nimbus/uses", "holy-nimbus/restore-slot-level", "holy-nimbus/restore-slot-cost"}, map[string]int{
		"holy-nimbus/duration-minutes": 10, "holy-nimbus/uses": 1, "holy-nimbus/restore-slot-level": 5, "holy-nimbus/restore-slot-cost": 1})
	pred(nimbus, "scope", "holy-nimbus/imbue-canonical-aura-of-protection-including-expansion-and-inactive-condition", []string{"feature/paladin/aura-of-protection", "feature/paladin/aura-expansion"}, nil)
	pred(nimbus, "restore", "holy-nimbus/restore-expended-use-with-canonical-slot", []string{"resource/paladin/spell-slot-5"}, map[string]int{"slots-required": 1, "slot-level": 5})
	return out
}
