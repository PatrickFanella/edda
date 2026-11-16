package srd521

import (
	"fmt"
	"strings"
)

// Independent pp61–63 source transcription. Builders describe ownership and
// operands explicitly; no fixture, Python oracle, or digest supplies these facts.
func rogueSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/"
	const k = m + "rogue/"
	cls := c + "class/rogue"
	var out []CharacterRecord
	by := map[string]int{}
	title := func(s string) string {
		words := strings.Split(s, "-")
		for i, w := range words {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
		return strings.Join(words, " ")
	}
	add := func(kind, slug, parent, name string, level, page int) string {
		id := c + kind + "/rogue"
		if slug != "" {
			id += "/" + slug
		}
		src := CharacterSource{InventoryID: "srd-5.2.1:planning/classes/rogue", Page: 61}
		if kind == "feature" {
			src = CharacterSource{InventoryID: fmt.Sprintf("srd-5.2.1:planning/class-features/rogue-level-%d-%s", out[by[parent]].Level, slug), Page: page}
		} else if parent != "" {
			src = out[by[parent]].Sources[0]
		}
		if name == "" {
			name = title(slug)
		}
		by[id] = len(out)
		out = append(out, CharacterRecord{ID: id, Kind: kind, Name: name, ParentID: parent, Level: level, Sources: []CharacterSource{src}, ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, RequiredMechanics: []string{}, OptionKeys: []string{}})
		if kind == "choice" {
			out[by[parent]].ChoiceRefs = append(out[by[parent]].ChoiceRefs, id)
		}
		if kind == "prerequisite" {
			out[by[parent]].PrerequisiteRefs = append(out[by[parent]].PrerequisiteRefs, id)
		}
		if kind == "resource" {
			out[by[parent]].ResourceRefs = append(out[by[parent]].ResourceRefs, id)
		}
		return id
	}
	rule := func(id, words string, nums map[string]int) {
		i := by[id]
		for _, s := range strings.Fields(words) {
			out[i].RequiredMechanics = append(out[i].RequiredMechanics, k+s)
		}
		if len(nums) > 0 {
			out[i].Parameters = map[string]int{}
		}
		// Map order is irrelevant; the full comparator normalizes mechanic sets.
		for key, n := range nums {
			out[i].Parameters[k+key] = n
			out[i].RequiredMechanics = append(out[i].RequiredMechanics, k+key)
		}
	}
	choice := func(slug, parent string, count int, options string) string {
		id := add("choice", slug, parent, "", 0, 0)
		i := by[id]
		out[i].Select = count
		out[i].RequiredMechanics = []string{k + slug}
		for _, s := range strings.Fields(options) {
			out[i].OptionKeys = append(out[i].OptionKeys, m+s)
		}
		return id
	}
	pred := func(slug, parent, key string, refs []string, nums map[string]int, option string) {
		id := add("prerequisite", slug, parent, "", 0, 0)
		i := by[id]
		if nums == nil {
			nums = map[string]int{}
		}
		out[i].RequiredMechanics = []string{k + key}
		out[i].Requirement = &CharacterRequirement{Key: k + key, RecordRefs: refs, Parameters: nums}
		if option != "" {
			out[i].RequiredMechanics = append(out[i].RequiredMechanics, k+"predicate/conditional-on-option/"+option)
		}
	}
	add("class", "", "", "Rogue", 0, 0)
	rule(cls, "primary-ability/dexterity saves/dexterity saves/intelligence weapons/simple weapons/martial-finesse-or-light tools/thieves-tools armor/light level-1/all-core-traits-and-level-1-features multiclass/hit-die-one-class-skill-thieves-tools-light-armor-and-level-1-features", map[string]int{"hit-die-sides": 8})
	skills := "skills/acrobatics skills/athletics skills/deception skills/insight skills/intimidation skills/investigation skills/perception skills/persuasion skills/sleight-of-hand skills/stealth"
	choice("skills", cls, 4, skills)
	choice("multiclass-skill", cls, 1, skills)
	eq := choice("starting-equipment", cls, 1, "rogue/equipment/a rogue/equipment/b")
	// Preserve source item order as well as quantities for inspectable bundles.
	bundles := [][]struct {
		key string
		n   int
	}{
		{{"equipment/leather-armor", 1}, {"equipment/dagger", 2}, {"equipment/shortsword", 1}, {"equipment/shortbow", 1}, {"equipment/arrows", 20}, {"equipment/quiver", 1}, {"equipment/thieves-tools", 1}, {"equipment/burglars-pack", 1}, {"currency/gp", 8}},
		{{"currency/gp", 100}},
	}
	for j, b := range bundles {
		q := CharacterChoiceOption{Key: out[by[eq]].OptionKeys[j], Quantities: map[string]int{}}
		for _, v := range b {
			q.Quantities[m+v.key] = v.n
			if j == 0 {
				out[by[eq]].RequiredMechanics = append(out[by[eq]].RequiredMechanics, m+v.key)
			}
		}
		out[by[eq]].Options = append(out[by[eq]].Options, q)
	}
	// All 20 table rows: repeated ASIs, Expertise and subclass slots are grants
	// of the existing identity, never invented additional source identities.
	rows := []string{
		"expertise sneak-attack thieves-cant weapon-mastery", "cunning-action", "rogue-subclass steady-aim", "ability-score-improvement", "cunning-strike uncanny-dodge", "expertise", "evasion reliable-talent", "ability-score-improvement", "subclass", "ability-score-improvement", "improved-cunning-strike", "ability-score-improvement", "subclass", "devious-strikes", "slippery-mind", "ability-score-improvement", "subclass", "elusive", "epic-boon", "stroke-of-luck",
	}
	for i := range rows {
		n := i + 1
		id := add("level", fmt.Sprint(n), cls, fmt.Sprintf("Rogue level %d", n), n, 0)
		rule(id, "class-progression", map[string]int{"sneak-attack/dice-count": (n + 1) / 2, "sneak-attack/dice-sides": 6})
		out[by[id]].Progression = &CharacterClassProgression{ProficiencyBonus: 2 + (n-1)/4, FeatureRefs: []string{}, FeatureKeys: []string{}}
	}
	features := []struct {
		slug        string
		level, page int
	}{
		{"expertise", 1, 61}, {"sneak-attack", 1, 61}, {"thieves-cant", 1, 62}, {"weapon-mastery", 1, 62}, {"cunning-action", 2, 62}, {"rogue-subclass", 3, 62}, {"steady-aim", 3, 62}, {"ability-score-improvement", 4, 63}, {"cunning-strike", 5, 63}, {"uncanny-dodge", 5, 63}, {"evasion", 7, 63}, {"reliable-talent", 7, 63}, {"improved-cunning-strike", 11, 63}, {"devious-strikes", 14, 63}, {"slippery-mind", 15, 63}, {"elusive", 18, 63}, {"epic-boon", 19, 63}, {"stroke-of-luck", 20, 63},
	}
	f := map[string]string{}
	for _, v := range features {
		name := strings.ReplaceAll(title(v.slug), " Of ", " of ")
		if v.slug == "thieves-cant" {
			name = "Thieves’ Cant"
		}
		f[v.slug] = add("feature", v.slug, c+fmt.Sprintf("level/rogue/%d", v.level), name, 0, v.page)
	}
	for i, row := range rows {
		r := &out[by[c+fmt.Sprintf("level/rogue/%d", i+1)]]
		for _, slug := range strings.Fields(row) {
			if slug == "subclass" {
				r.Progression.FeatureKeys = append(r.Progression.FeatureKeys, k+"subclass-feature")
				r.RequiredMechanics = append(r.RequiredMechanics, k+"subclass-feature")
			} else {
				r.Progression.FeatureRefs = append(r.Progression.FeatureRefs, f[slug])
			}
		}
	}
	ex := f["expertise"]
	rule(ex, "expertise/skill-proficiencies-only expertise/recommend-sleight-of-hand-and-stealth-if-proficient", map[string]int{"expertise/initial-count": 2, "expertise/additional-level": 6, "expertise/additional-count": 2})
	for i, s := range []string{"initial", "level-6"} {
		q := choice("expertise-"+s, ex, 2, "skills/own-proficiencies")
		level := 1
		if i == 1 {
			level = 6
		}
		pred("expertise-"+s, q, "expertise/requires-own-skill-proficiency", []string{cls, ex}, map[string]int{"minimum-level": level}, "")
	}
	sa := f["sneak-attack"]
	rule(sa, "sneak-attack/once-per-turn sneak-attack/hit-one-creature-with-attack-roll sneak-attack/finesse-or-ranged-weapon sneak-attack/same-damage-type-as-weapon sneak-attack/table-scaling", nil)
	eligibility := choice("sneak-attack-eligibility", sa, 1, "rogue/sneak-attack/advantage rogue/sneak-attack/ally-within-five-feet-not-incapacitated-and-no-disadvantage")
	pred("sneak-attack-eligibility", sa, "sneak-attack/requires-eligible-hit", []string{sa, eligibility}, map[string]int{"ally-distance-feet": 5}, "")
	rule(sa, "sneak-attack/ally-must-not-be-incapacitated sneak-attack/ally-alternative-requires-no-disadvantage", nil)
	rule(f["thieves-cant"], "language/thieves-cant", nil)
	choice("additional-language", f["thieves-cant"], 1, "languages/character-creation-tables")
	wm := f["weapon-mastery"]
	rule(wm, "weapon-mastery/change-kinds-on-long-rest weapon-mastery/recommend-daggers-shortbows", nil)
	q := choice("weapon-mastery", wm, 2, "weapons/own-proficiencies")
	pred("weapon-mastery-proficiency", q, "weapon-mastery/requires-weapon-proficiency", []string{cls}, nil, "")
	rule(f["cunning-action"], "cunning-action/bonus-action-on-own-turn", nil)
	choice("cunning-action", f["cunning-action"], 1, "actions/dash actions/disengage actions/hide")
	rule(f["rogue-subclass"], "subclass/grant-features-at-own-class-level-or-lower", nil)
	choice("rogue-subclass", f["rogue-subclass"], 1, "rogue/subclasses rogue/subclasses/thief")
	aim := f["steady-aim"]
	rule(aim, "steady-aim/bonus-action steady-aim/advantage-next-attack-roll-current-turn steady-aim/speed-zero-until-current-turn-end", nil)
	pred("steady-aim-unmoved", aim, "steady-aim/requires-no-movement-this-turn", []string{cls}, map[string]int{"maximum-feet-moved": 0}, "")
	for _, s := range []string{"ability-score-improvement", "epic-boon"} {
		rule(f[s], s+"/qualified-feat-choice", nil)
		opt := "feats/" + s
		ch := choice(s, f[s], 1, opt+" feats/qualified")
		pred(s+"-qualification", ch, "feat/requires-qualification", []string{cls}, nil, "feats/qualified")
	}
	rule(f["epic-boon"], "epic-boon/recommend-boon-of-the-night-spirit", nil)
	cs := f["cunning-strike"]
	rule(cs, "cunning-strike/when-dealing-sneak-attack cunning-strike/forgo-dice-before-roll cunning-strike/effect-immediately-after-attack-damage cunning-strike/save-dc-eight-plus-dexterity-modifier-plus-proficiency-bonus", map[string]int{"cunning-strike/base-effect-limit": 1, "cunning-strike/save-dc-base": 8})
	effects := choice("cunning-strike", cs, 1, "rogue/cunning-strike/poison rogue/cunning-strike/trip rogue/cunning-strike/withdraw")
	ds := f["devious-strikes"]
	rule(ds, "devious-strikes/add-to-cunning-strike-options", nil)
	extra := choice("devious-strikes", ds, 1, "rogue/cunning-strike/daze rogue/cunning-strike/knock-out rogue/cunning-strike/obscure")
	pred("devious-strikes-options", extra, "devious-strikes/extends-cunning-strike", []string{effects, cs}, nil, "")
	variants := []struct {
		slug, owner, words string
		cost               int
		nums               map[string]int
	}{
		{"poison", effects, "constitution-save poisoned-on-failure repeat-save-end-each-target-turn end-on-success requires-poisoners-kit-on-person", 1, map[string]int{"duration-minutes": 1}},
		{"trip", effects, "dexterity-save prone-on-failure target-large-or-smaller", 1, nil},
		{"withdraw", effects, "immediately-after-attack move-up-to-half-speed no-opportunity-attacks", 1, map[string]int{"speed-divisor": 2}},
		{"daze", extra, "constitution-save on-failure-next-turn-only-one-of-move-action-bonus-action", 2, nil},
		{"knock-out", extra, "constitution-save unconscious-on-failure ends-on-any-damage repeat-save-end-each-target-turn end-on-success", 6, map[string]int{"duration-minutes": 1}},
		{"obscure", extra, "dexterity-save blinded-on-failure until-end-target-next-turn", 3, nil},
	}
	for _, v := range variants {
		ch := choice("strike-"+v.slug, v.owner, 1, "rogue/cunning-strike/"+v.slug)
		pred("strike-"+v.slug+"-selected", ch, "cunning-strike/"+v.slug+"/selected", []string{v.owner, sa}, map[string]int{"cost-d6": v.cost}, "rogue/cunning-strike/"+v.slug)
		words := []string{}
		for _, w := range strings.Fields(v.words) {
			words = append(words, "cunning-strike/"+v.slug+"/"+w)
		}
		rule(ch, strings.Join(words, " "), nil)
		if v.slug == "daze" {
			turn := choice("daze-restricted-turn", ch, 1, "rogue/daze/move rogue/daze/action rogue/daze/bonus-action")
			rule(turn, "daze/only-on-failed-save-next-turn", nil)
		}
		if v.nums != nil {
			pred("strike-"+v.slug+"-numbers", ch, "cunning-strike/"+v.slug+"/limits", []string{sa}, v.nums, "")
		}
	}
	rule(f["improved-cunning-strike"], "improved-cunning-strike/pay-each-effect-cost", map[string]int{"improved-cunning-strike/maximum-effects": 2})
	pred("improved-cunning-strike-operands", f["improved-cunning-strike"], "improved-cunning-strike/requires-sneak-attack-and-strike", []string{sa, effects}, nil, "")
	rule(f["uncanny-dodge"], "uncanny-dodge/reaction uncanny-dodge/attacker-you-can-see-hits-with-attack-roll uncanny-dodge/halve-attacks-damage-round-down", map[string]int{"uncanny-dodge/damage-divisor": 2})
	rule(f["evasion"], "evasion/dexterity-save-for-half-damage evasion/success-no-damage evasion/failure-half-damage evasion/unavailable-while-incapacitated", nil)
	rule(f["reliable-talent"], "reliable-talent/ability-check-using-own-skill-or-tool-proficiency", map[string]int{"reliable-talent/maximum-d20-roll": 9, "reliable-talent/replacement-roll": 10})
	rule(f["slippery-mind"], "saves/wisdom saves/charisma", nil)
	rule(f["elusive"], "elusive/no-attack-roll-advantage-against-self-unless-incapacitated", nil)
	luck := f["stroke-of-luck"]
	rule(luck, "stroke-of-luck/failed-d20-test stroke-of-luck/turn-roll-into-twenty", map[string]int{"stroke-of-luck/replacement-roll": 20})
	res := add("resource", "stroke-of-luck", luck, "", 0, 0)
	out[by[res]].Capacity = 1
	out[by[res]].RequiredMechanics = []string{k + "stroke-of-luck/recover-on-short-or-long-rest"}
	return out
}
