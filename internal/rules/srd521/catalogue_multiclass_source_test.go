package srd521

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Independent transcription of pp24–26 and class core-trait/multiclass
// continuations. No values are derived from the candidate or production code.
func multiclassSourceExpectations(t *testing.T) ([]MulticlassClass, []MulticlassRequirement, []MulticlassSlots) {
	t.Helper()
	const c = "srd-5.2.1:character/"
	var classes []MulticlassClass
	for _, line := range strings.Split(`barbarian;28;strength;martial-weapons shields;
bard;31;charisma;any-skill musical-instrument light-armor;multiclass-skill multiclass-instrument
cleric;36;wisdom;light-armor medium-armor shields;
druid;41;wisdom;light-armor shields;
fighter;47;strength|dexterity;martial-weapons light-armor medium-armor shields;
monk;49;dexterity+wisdom;;
paladin;53;strength+charisma;martial-weapons light-armor medium-armor shields;
ranger;57;dexterity+wisdom;martial-weapons class-skill light-armor medium-armor shields;multiclass-skill
rogue;61;dexterity;class-skill thieves-tools light-armor;multiclass-skill
sorcerer;64;charisma;;
warlock;70;charisma;light-armor;
wizard;77;intelligence;;`, "\n") {
		f := strings.Split(line, ";")
		page, err := strconv.Atoi(f[1])
		if err != nil {
			t.Fatal(err)
		}
		r := MulticlassClass{ClassRef: c + "class/" + f[0], Source: CharacterSource{InventoryID: "srd-5.2.1:planning/classes/" + f[0], Page: page}, MinimumScore: 13, Proficiencies: strings.Fields(f[3]), ProficiencyOwnerRefs: []string{c + "class/" + f[0]}}
		for _, g := range strings.Split(f[2], "|") {
			r.PrimaryAny = append(r.PrimaryAny, strings.Split(g, "+"))
		}
		for _, owner := range strings.Fields(f[4]) {
			r.ProficiencyOwnerRefs = append(r.ProficiencyOwnerRefs, c+"choice/"+f[0]+"/"+owner)
		}
		classes = append(classes, r)
	}
	all := []string{}
	for _, n := range strings.Fields("barbarian bard cleric druid fighter monk paladin ranger rogue sorcerer warlock wizard") {
		all = append(all, "class/"+n)
	}
	spells := []string{}
	for _, n := range strings.Fields("bard cleric druid paladin ranger sorcerer wizard") {
		spells = append(spells, "feature/"+n+"/spellcasting")
	}
	advancement := []string{}
	for i := 1; i <= 20; i++ {
		advancement = append(advancement, "advancement/"+strconv.Itoa(i))
	}
	var rules []MulticlassRequirement
	add := func(key string, first, last int, clauses string, refs []string, nums map[string]int) {
		r := MulticlassRequirement{Key: "srd-5.2.1:mechanic/multiclass/" + key, FirstPage: first, LastPage: last, Clauses: strings.Split(clauses, ";"), RecordRefs: []string{}, Parameters: map[string]int{}}
		for _, ref := range refs {
			r.RecordRefs = append(r.RecordRefs, c+ref)
		}
		for k, v := range nums {
			r.Parameters[k] = v
		}
		rules = append(rules, r)
	}
	add("prerequisites", 24, 25, "new-and-every-current-class;each-primary-ability-at-least-minimum;alternatives-preserved", nil, nil)
	add("experience-and-proficiency", 25, 25, "use-total-character-level;reuse-global-advancement", advancement, nil)
	add("hit-points-and-dice", 25, 25, "new-class-hp-as-levels-after-first;first-level-hp-only-at-total-level-one;sum-class-hit-dice;same-die-type-may-pool;different-die-types-tracked-separately", all, map[string]int{"first-level-hp-total-level": 1})
	add("proficiencies", 25, 25, "first-level-in-noninitial-class;only-listed-starting-proficiencies;no-other-core-traits", nil, nil)
	add("class-features", 25, 25, "new-class-level-features;check-class-specific-multiclass-rules", all, nil)
	add("armor-class", 25, 25, "multiple-calculations;benefit-from-only-one-at-a-time;monk-unarmored-defense-or-sorcerer-draconic-resilience-example", []string{"feature/monk/unarmored-defense"}, map[string]int{"simultaneous-calculations": 1})
	add("extra-attack", 25, 25, "extra-attack-from-multiple-classes-does-not-stack;maximum-two-unless-feature-explicitly-says-more;fighter-two-extra-attacks-exception;thirsting-blade-pact-weapon-no-additional-attacks-with-extra-attack", []string{"feature/fighter/extra-attack", "feature/fighter/two-extra-attacks", "feature/warlock/eldritch-invocations"}, map[string]int{"ordinary-maximum-attacks": 2})
	add("spellcasting-scope", 25, 25, "combined-and-individual-class-levels;shared-rules-only-with-spellcasting-from-more-than-one-class;one-spellcasting-class-use-own-class-rules", spells, map[string]int{"minimum-spellcasting-classes": 2})
	add("prepared-spells", 25, 26, "prepare-per-class-as-single-class;each-prepared-spell-associated-with-one-class;use-that-class-spellcasting-ability;higher-slots-do-not-unlock-higher-preparations", spells, nil)
	add("cantrips", 25, 25, "power-scaling-total-character-level;unless-spell-says-otherwise", nil, nil)
	add("spell-slots", 25, 26, "sum-full-caster-levels;add-half-paladin-and-ranger-levels-rounded-up;look-up-total-in-shared-table;cast-appropriate-level-spells-from-any-owned-spellcasting-class;higher-slots-only-cast-prepared-lower-level-spells;enhanced-effect-only-if-spell-provides-it", spells, map[string]int{"full-caster-divisor": 1, "paladin-divisor": 2, "ranger-divisor": 2, "round-up": 1})
	add("pact-magic", 26, 26, "requires-warlock-pact-magic-and-spellcasting;pact-slots-cast-prepared-spellcasting-class-spells;spellcasting-slots-cast-prepared-warlock-spells;pact-levels-excluded-from-shared-slot-sum", append([]string{"feature/warlock/pact-magic"}, spells...), nil)
	var rows []MulticlassSlots
	for i, line := range strings.Fields("200000000 300000000 420000000 430000000 432000000 433000000 433100000 433200000 433310000 433320000 433321000 433321000 433321100 433321100 433321110 433321110 433321111 433331111 433332111 433332211") {
		row := MulticlassSlots{Level: i + 1, Slots: []int{}}
		for _, v := range line {
			row.Slots = append(row.Slots, int(v-'0'))
		}
		rows = append(rows, row)
	}
	return classes, rules, rows
}

func multiclassSourceAgrees(t *testing.T, c MulticlassCatalogue) bool {
	t.Helper()
	classes, rules, rows := multiclassSourceExpectations(t)
	terms := []MulticlassSlotTerm{
		{"srd-5.2.1:character/class/bard", 1, "none"},
		{"srd-5.2.1:character/class/cleric", 1, "none"},
		{"srd-5.2.1:character/class/druid", 1, "none"},
		{"srd-5.2.1:character/class/paladin", 2, "up"},
		{"srd-5.2.1:character/class/ranger", 2, "up"},
		{"srd-5.2.1:character/class/sorcerer", 1, "none"},
		{"srd-5.2.1:character/class/wizard", 1, "none"},
	}
	return reflect.DeepEqual(c.SlotTerms, terms) && reflect.DeepEqual(c.Classes, classes) && reflect.DeepEqual(c.Requirements, rules) && reflect.DeepEqual(c.SlotRows, rows)
}
