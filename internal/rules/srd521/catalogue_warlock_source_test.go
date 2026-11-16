package srd521

import (
	"fmt"
	"strings"
)

// Independent pp70–72 source transcription. Builders describe ownership and
// operands explicitly; no fixture, Python oracle, or digest supplies these facts.
func warlockSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/"
	const k = m + "warlock/"
	cls := c + "class/warlock"
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
		id := c + kind + "/warlock"
		if slug != "" {
			id += "/" + slug
		}
		src := CharacterSource{InventoryID: "srd-5.2.1:planning/classes/warlock", Page: 70}
		if kind == "feature" {
			src = CharacterSource{InventoryID: fmt.Sprintf("srd-5.2.1:planning/class-features/warlock-level-%d-%s", out[by[parent]].Level, slug), Page: page}
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
		if len(nums) > 0 && out[i].Parameters == nil {
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
	add("class", "", "", "Warlock", 0, 0)
	rule(cls, "primary-ability/charisma saves/wisdom saves/charisma weapons/simple armor/light level-1/all-core-traits-and-level-1-features multiclass/hit-die-light-armor-level-1-features multiclass/consult-multiclass-spell-slot-rules", map[string]int{"hit-die-sides": 8})
	choice("skills", cls, 2, "skills/arcana skills/deception skills/history skills/intimidation skills/investigation skills/nature skills/religion")
	eq := choice("starting-equipment", cls, 1, "warlock/equipment/a warlock/equipment/b")
	for j, bundle := range []map[string]int{
		{"equipment/leather-armor": 1, "equipment/sickle": 1, "equipment/dagger": 2, "equipment/arcane-focus-orb": 1, "equipment/book-occult-lore": 1, "equipment/scholars-pack": 1, "currency/gp": 15},
		{"currency/gp": 100},
	} {
		o := CharacterChoiceOption{Key: out[by[eq]].OptionKeys[j], Quantities: map[string]int{}}
		// Stable source order for the mechanic set, independent of map iteration.
		for _, key := range []string{"equipment/leather-armor", "equipment/sickle", "equipment/dagger", "equipment/arcane-focus-orb", "equipment/book-occult-lore", "equipment/scholars-pack", "currency/gp"} {
			if n, ok := bundle[key]; ok {
				o.Quantities[m+key] = n
				if j == 0 {
					out[by[eq]].RequiredMechanics = append(out[by[eq]].RequiredMechanics, m+key)
				}
			}
		}
		out[by[eq]].Options = append(out[by[eq]].Options, o)
	}
	rows := []string{"eldritch-invocations pact-magic", "magical-cunning", "warlock-subclass", "ability-score-improvement", "", "subclass", "", "ability-score-improvement", "contact-patron", "subclass", "mystic-arcanum", "ability-score-improvement", "mystic-arcanum", "subclass", "mystic-arcanum", "ability-score-improvement", "mystic-arcanum", "", "epic-boon", "eldritch-master"}
	// Columns: invocations, cantrips, prepared, Pact slots, slot level.
	table := [][5]int{
		{1, 2, 2, 1, 1},   // level 1
		{3, 2, 3, 2, 1},   // level 2
		{3, 2, 4, 2, 2},   // level 3
		{3, 3, 5, 2, 2},   // level 4
		{5, 3, 6, 2, 3},   // level 5
		{5, 3, 7, 2, 3},   // level 6
		{6, 3, 8, 2, 4},   // level 7
		{6, 3, 9, 2, 4},   // level 8
		{7, 3, 10, 2, 5},  // level 9
		{7, 4, 10, 2, 5},  // level 10
		{7, 4, 11, 3, 5},  // level 11
		{8, 4, 11, 3, 5},  // level 12
		{8, 4, 12, 3, 5},  // level 13
		{8, 4, 12, 3, 5},  // level 14
		{9, 4, 13, 3, 5},  // level 15
		{9, 4, 13, 3, 5},  // level 16
		{9, 4, 14, 4, 5},  // level 17
		{10, 4, 14, 4, 5}, // level 18
		{10, 4, 15, 4, 5}, // level 19
		{10, 4, 15, 4, 5}, // level 20
	}
	columns := []string{"invocations/count", "cantrips/count", "prepared-spells/count", "pact-slots/count", "pact-slots/level"}
	for i, cells := range table {
		n := i + 1
		id := add("level", fmt.Sprint(n), cls, fmt.Sprintf("Warlock level %d", n), n, 0)
		nums := map[string]int{}
		for j, key := range columns {
			nums[key] = cells[j]
		}
		rule(id, "class-progression", nums)
		out[by[id]].Progression = &CharacterClassProgression{ProficiencyBonus: 2 + (n-1)/4, FeatureRefs: []string{}, FeatureKeys: []string{}}
	}
	f := map[string]string{}
	for _, v := range []struct {
		slug        string
		level, page int
	}{
		{"eldritch-invocations", 1, 70}, {"pact-magic", 1, 71}, {"magical-cunning", 2, 72}, {"warlock-subclass", 3, 72}, {"ability-score-improvement", 4, 72}, {"contact-patron", 9, 72}, {"mystic-arcanum", 11, 72}, {"epic-boon", 19, 72}, {"eldritch-master", 20, 72},
	} {
		f[v.slug] = add("feature", v.slug, c+fmt.Sprintf("level/warlock/%d", v.level), "", 0, v.page)
	}
	for i, row := range rows {
		r := &out[by[c+fmt.Sprintf("level/warlock/%d", i+1)]]
		for _, slug := range strings.Fields(row) {
			if slug == "subclass" {
				r.Progression.FeatureKeys = append(r.Progression.FeatureKeys, k+"subclass-feature")
				r.RequiredMechanics = append(r.RequiredMechanics, k+"subclass-feature")
			} else {
				r.Progression.FeatureRefs = append(r.Progression.FeatureRefs, f[slug])
			}
		}
		if i == 10 || i == 12 || i == 14 || i == 16 {
			rule(r.ID, "", map[string]int{"mystic-arcanum/spell-level": 6 + (i-10)/2})
		}
	}
	inv := f["eldritch-invocations"]
	rule(inv, "invocations/table-count invocations/gain-more-on-table-increase invocations/recommend-pact-of-the-tome invocations/replace-one-on-warlock-level invocations/cannot-replace-prerequisite-of-owned-invocation invocations/no-duplicates-unless-description-allows", map[string]int{"invocations/replacement-count": 1})
	q := choice("invocations", inv, 1, "warlock/invocation-options")
	for i, cells := range table {
		out[by[q]].ValuesByLevel = append(out[by[q]].ValuesByLevel, CharacterLevelValue{Level: i + 1, Value: cells[0]})
	}
	pred("invocation-eligibility", q, "invocations/must-meet-all-prerequisites", []string{cls, inv}, nil, "")
	pred("invocation-replacement", q, "invocations/replacement-must-qualify-and-not-be-required-by-owned-invocation", []string{inv, q}, map[string]int{"maximum-replacements": 1}, "")
	pm := f["pact-magic"]
	rule(pm, "spell-list/warlock spellcasting/charisma spellcasting/arcane-focus pact-slots/distinct-from-spellcasting pact-slots/all-same-level pact-slots/expended-to-cast-level-one-through-five pact-slots/cast-at-slot-level pact-slots/recover-all-on-short-or-long-rest cantrips/learn-on-table-increase cantrips/recommend-eldritch-blast-and-prestidigitation cantrips/replace-one-on-warlock-level prepared-spells/recommend-charm-person-and-hex prepared-spells/add-to-table-count-on-increase prepared-spells/replace-one-on-warlock-level prepared-spells/other-warlock-features-always-prepared-excluded-from-count-but-warlock-spells", map[string]int{"cantrips/replacement-count": 1, "prepared-spells/replacement-count": 1, "pact-slots/minimum-spell-level": 1, "pact-slots/maximum-spell-level": 5})
	can := choice("cantrips", pm, 2, "spells/warlock-cantrips")
	prep := choice("prepared-spells", pm, 2, "spells/warlock-level-one-through-five")
	slots := add("resource", "pact-slots", pm, "Pact Magic slots", 0, 0)
	out[by[slots]].Capacity = 1
	ruleKey := k + "pact-slots/recover-all-on-short-or-long-rest"
	out[by[slots]].RequiredMechanics = []string{ruleKey}
	for i, cells := range table {
		for j, id := range []string{can, prep, slots} {
			out[by[id]].ValuesByLevel = append(out[by[id]].ValuesByLevel, CharacterLevelValue{Level: i + 1, Value: cells[j+1]})
		}
	}
	pred("cantrip-eligibility", can, "cantrips/warlock-list-only", []string{cls, pm}, nil, "")
	pred("prepared-spell-eligibility", prep, "prepared-spells/warlock-list-level-at-most-current-pact-slot-level", []string{cls, pm, slots}, map[string]int{"minimum-spell-level": 1}, "")
	cunning := f["magical-cunning"]
	rule(cunning, "magical-cunning/rite-at-end-recover-expended-pact-slots magical-cunning/half-maximum-round-up", map[string]int{"magical-cunning/rite-minutes": 1, "magical-cunning/maximum-divisor": 2})
	res := add("resource", "magical-cunning", cunning, "", 0, 0)
	out[by[res]].Capacity = 1
	out[by[res]].RequiredMechanics = []string{k + "magical-cunning/recover-on-long-rest"}
	pred("magical-cunning-recovery", cunning, "magical-cunning/expended-pact-slots-only-up-to-half-maximum-rounded-up", []string{slots, res}, map[string]int{"maximum-divisor": 2}, "")
	rule(f["warlock-subclass"], "subclass/grant-features-at-own-warlock-level-or-lower", nil)
	choice("warlock-subclass", f["warlock-subclass"], 1, "warlock/subclasses warlock/subclasses/fiend-patron")
	for _, slug := range []string{"ability-score-improvement", "epic-boon"} {
		rule(f[slug], slug+"/qualified-feat-choice", nil)
		q := choice(slug, f[slug], 1, "feats/"+slug+" feats/qualified")
		pred(slug+"-qualification", q, "feat/requires-qualification", []string{cls}, nil, "feats/qualified")
	}
	rule(f["epic-boon"], "epic-boon/recommend-boon-of-fate", nil)
	contact := f["contact-patron"]
	rule(contact, "spells/contact-other-plane contact-patron/always-prepared", nil)
	out[by[contact]].AlwaysPrepared = []CharacterPreparedSpells{{Level: 9, SpellKeys: []string{k + "spells/contact-other-plane"}}}
	modes := choice("contact-patron-casting", contact, 1, "warlock/contact-patron/ordinary-spellcasting warlock/contact-patron/free-patron-contact")
	pred("contact-patron-free", modes, "contact-patron/free-cast-contact-own-patron-auto-save-success", []string{contact, c + "resource/warlock/contact-patron"}, nil, "warlock/contact-patron/free-patron-contact")
	res = add("resource", "contact-patron", contact, "", 0, 0)
	out[by[res]].Capacity = 1
	out[by[res]].RequiredMechanics = []string{k + "contact-patron/free-cast-no-slot-recover-on-long-rest"}
	arc := f["mystic-arcanum"]
	rule(arc, "mystic-arcanum/warlock-spell-list mystic-arcanum/no-slot-for-free-cast mystic-arcanum/replace-one-on-warlock-level mystic-arcanum/replacement-same-spell-level mystic-arcanum/recover-all-uses-on-long-rest", map[string]int{"mystic-arcanum/replacement-count": 1})
	for i, level := range []int{11, 13, 15, 17} {
		spell := 6 + i
		slug := fmt.Sprintf("arcanum-%d", spell)
		q := choice(slug, arc, 1, fmt.Sprintf("spells/warlock-level-%d", spell))
		pred(slug+"-eligibility", q, "mystic-arcanum/warlock-spell-at-specific-level", []string{cls, arc}, map[string]int{"minimum-warlock-level": level, "spell-level": spell}, "")
		res := add("resource", slug, arc, fmt.Sprintf("Level %d arcanum free use", spell), 0, 0)
		out[by[res]].Capacity = 1
		out[by[res]].RequiredMechanics = []string{k + "mystic-arcanum/separate-free-use-recover-on-long-rest"}
		pred(slug+"-casting", arc, "mystic-arcanum/free-use-for-selected-arcanum-only", []string{q, res}, map[string]int{"minimum-warlock-level": level, "spell-level": spell}, "")
		pred(slug+"-replacement", q, "mystic-arcanum/replace-with-same-level-warlock-spell", []string{q, arc}, map[string]int{"spell-level": spell}, "")
	}
	rule(f["eldritch-master"], "eldritch-master/when-using-magical-cunning eldritch-master/recover-all-expended-pact-slots", nil)
	pred("eldritch-master-recovery", f["eldritch-master"], "eldritch-master/changes-cunning-recovery-only", []string{cunning, slots}, nil, "")
	return out
}
