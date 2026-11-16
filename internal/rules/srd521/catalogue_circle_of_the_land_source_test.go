package srd521

import "strings"

// Independent p46 transcription. These builders express source ownership,
// conditional tables, operands and all numbers; they never read the JSON or
// the Python oracle. Empty fields stay explicit so unexpected grants fail.
func landSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/"
	const k = m + "druid/circle-of-the-land/"
	const sub = c + "subclass/druid/circle-of-the-land"
	const f = c + "feature/druid/circle-of-the-land/"
	const p = c + "prerequisite/druid/circle-of-the-land/"
	const q = c + "choice/druid/circle-of-the-land/"
	const res = c + "resource/druid/circle-of-the-land/"
	var records []CharacterRecord
	add := func(id, kind, name, parent, source string, keys []string) int {
		records = append(records, CharacterRecord{ID: id, Kind: kind, Name: name, ParentID: parent,
			Sources:    []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 46}},
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, OptionKeys: []string{}, RequiredMechanics: keys})
		return len(records) - 1
	}
	predicate := func(owner int, slug, key string, refs []string, numbers map[string]int) {
		// Labels are editorial metadata; source semantics reside in the predicate.
		words := strings.Split(slug, "-")
		for i, w := range words {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
		source := strings.TrimPrefix(records[owner].Sources[0].InventoryID, "srd-5.2.1:planning/")
		id := add(p+slug, "prerequisite", strings.Join(words, " "), records[owner].ID, source, []string{key})
		records[id].Requirement = &CharacterRequirement{Key: key, RecordRefs: refs, Parameters: numbers}
		records[owner].PrerequisiteRefs = append(records[owner].PrerequisiteRefs, records[id].ID)
	}
	feature := func(slug, name, source string, level int, semantics []string, numbers [][2]any) int {
		keys := []string{}
		for _, s := range semantics {
			keys = append(keys, k+slug+"/"+s)
		}
		id := add(f+slug, "feature", name, sub, "subclass-features/druid-level-"+source, keys)
		if len(numbers) > 0 {
			records[id].Parameters = map[string]int{}
		}
		for _, pair := range numbers {
			key := k + slug + "/" + pair[0].(string)
			records[id].RequiredMechanics = append(records[id].RequiredMechanics, key)
			records[id].Parameters[key] = pair[1].(int)
		}
		predicate(id, slug+"-level", m+"prerequisite/class-level", []string{c + "class/druid"}, map[string]int{"minimum-level": level})
		return id
	}
	id := add(sub, "subclass", "Circle of the Land", c+"class/druid", "subclasses/druid-circle-of-the-land", []string{m + "druid/subclasses/circle-of-the-land"})
	predicate(id, "circle-of-the-land-level", m+"prerequisite/class-level", []string{c + "class/druid"}, map[string]int{"minimum-level": 3})
	spells := feature("circle-of-the-land-spells", "Circle of the Land Spells", "3-circle-of-the-land-spells", 3, []string{
		"choose-one-land-on-finishing-long-rest", "prepare-current-land-spells-at-druid-level-and-lower", "conditional-tables-not-cumulative-across-lands"}, nil)
	choice := add(q+"land", "choice", "Land", f+"circle-of-the-land-spells", "subclass-features/druid-level-3-circle-of-the-land-spells", []string{k + "land/choose-on-finishing-long-rest"})
	records[choice].Select = 1
	records[choice].OptionKeys = []string{k + "land/arid", k + "land/polar", k + "land/temperate", k + "land/tropical"}
	records[spells].ChoiceRefs = []string{q + "land"}
	tables := []struct {
		land, name string
		rows       [4]string
	}{
		{"arid", "Arid", [4]string{"blur burning-hands fire-bolt", "fireball", "blight", "wall-of-stone"}},
		{"polar", "Polar", [4]string{"fog-cloud hold-person ray-of-frost", "sleet-storm", "ice-storm", "cone-of-cold"}},
		{"temperate", "Temperate", [4]string{"misty-step shocking-grasp sleep", "lightning-bolt", "freedom-of-movement", "tree-stride"}},
		{"tropical", "Tropical", [4]string{"acid-splash ray-of-sickness web", "stinking-cloud", "polymorph", "insect-plague"}},
	}
	for _, table := range tables {
		slug := "spells-" + table.land
		id := add(f+slug, "feature", table.name+" Land spells", sub, "subclass-features/druid-level-3-circle-of-the-land-spells", []string{k + slug + "/only-while-current-land-selected"})
		for i, row := range table.rows {
			keys := []string{}
			for _, spell := range strings.Fields(row) {
				keys = append(keys, m+"spells/"+spell)
			}
			records[id].AlwaysPrepared = append(records[id].AlwaysPrepared, CharacterPreparedSpells{Level: 3 + 2*i, SpellKeys: keys})
			records[id].RequiredMechanics = append(records[id].RequiredMechanics, keys...)
		}
		predicate(id, slug+"-selected", k+"land/"+table.land, []string{q + "land", f + "circle-of-the-land-spells", c + "class/druid"}, map[string]int{"minimum-level": 3})
	}
	aid := feature("lands-aid", "Land’s Aid", "3-lands-aid", 3, []string{
		"magic-action", "choose-point", "chosen-creatures-in-sphere/constitution-save-against-own-spell-save-dc",
		"failed-save/necrotic-damage", "successful-save/half-damage", "one-chosen-creature-in-sphere/regains-hit-points",
		"expend-canonical-wild-shape"}, [][2]any{
		{"range-feet", 60}, {"sphere-radius-feet", 10}, {"wild-shape-uses-expended", 1}, {"dice-sides", 6}, {"dice-at-level-3", 2}, {"dice-at-level-10", 3}, {"dice-at-level-14", 4}})
	predicate(aid, "lands-aid-wild-shape", k+"lands-aid/requires-unexpended-wild-shape", []string{c + "resource/druid/wild-shape"}, map[string]int{"minimum-uses": 1})
	recovery := feature("natural-recovery", "Natural Recovery", "6-natural-recovery", 6, []string{
		"cast-one-prepared-circle-spell-of-level-1-or-higher-without-slot", "free-cast/long-rest-recovery",
		"on-finishing-short-rest/choose-expended-slots", "slot-recovery/combined-level-at-most-half-druid-level-rounded-up", "slot-recovery/once-until-long-rest"}, [][2]any{
		{"minimum-free-spell-level", 1}, {"maximum-recovered-slot-level", 5}, {"recovery-budget-divisor", 2}, {"free-casts-per-long-rest", 1}, {"slot-recoveries-per-long-rest", 1}})
	predicate(recovery, "natural-recovery-operands", k+"natural-recovery/requires-prepared-circle-spell-or-expended-druid-slots", []string{
		f + "circle-of-the-land-spells", c + "class/druid", c + "resource/druid/spell-slot-1", c + "resource/druid/spell-slot-2", c + "resource/druid/spell-slot-3",
		c + "resource/druid/spell-slot-4", c + "resource/druid/spell-slot-5", c + "resource/druid/spell-slot-6", c + "resource/druid/spell-slot-7", c + "resource/druid/spell-slot-8", c + "resource/druid/spell-slot-9"}, map[string]int{})
	for _, slug := range []string{"free-cast", "slot-recovery"} {
		id := add(res+"natural-recovery-"+slug, "resource", "Natural Recovery "+slug, f+"natural-recovery", "subclass-features/druid-level-6-natural-recovery", []string{k + "natural-recovery/" + slug + "/recover-on-long-rest"})
		records[id].Capacity = 1
		records[recovery].ResourceRefs = append(records[recovery].ResourceRefs, records[id].ID)
	}
	ward := feature("natures-ward", "Nature’s Ward", "10-natures-ward", 10, []string{
		"immune-to-poisoned-condition", "resistance/current-land/arid/fire", "resistance/current-land/polar/cold",
		"resistance/current-land/temperate/lightning", "resistance/current-land/tropical/poison"}, nil)
	predicate(ward, "natures-ward-current-land", k+"natures-ward/current-land-determines-resistance", []string{q + "land"}, map[string]int{})
	sanctuary := feature("natures-sanctuary", "Nature’s Sanctuary", "14-natures-sanctuary", 14, []string{
		"magic-action", "spectral-trees-and-vines/cube-on-ground", "ends-on-incapacitated-or-death", "self-and-allies-in-area/half-cover",
		"allies-in-area/gain-current-natures-ward-resistance", "bonus-action/move-cube-to-ground-within-range", "expend-canonical-wild-shape"}, [][2]any{
		{"range-feet", 120}, {"cube-side-feet", 15}, {"duration-minutes", 1}, {"move-feet", 60}, {"wild-shape-uses-expended", 1}})
	predicate(sanctuary, "natures-sanctuary-operands", k+"natures-sanctuary/requires-unexpended-wild-shape-and-current-ward", []string{c + "resource/druid/wild-shape", f + "natures-ward"}, map[string]int{"minimum-uses": 1})
	return records
}
