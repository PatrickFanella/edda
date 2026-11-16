package srd521

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const bardID = "srd-5.2.1:character/"
const bardKey = "srd-5.2.1:mechanic/"

func bardCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-bard.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// This freezes every field, including prerequisite RequiredMechanics, operands,
// recommendations, source ownership, quantities, recovery restrictions and the
// absence of invented scaling. Readable source checks below complement it.
func bardSourceDigest(c CharacterCatalogue) string {
	b, err := json.Marshal(c)
	if err != nil {
		return "marshal error"
	}
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

const bardSourcePin = "63651ddcd8479abb0ffe1527245cc7e16ef443df897a188760f438e3f27ce366"

// These p31–32 facts are checked semantically, independently of the digest.
func bardProseSourceError(c CharacterCatalogue) error {
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		records[strings.TrimPrefix(r.ID, bardID)] = r
	}
	r := records["resource/bard/bardic-inspiration"]
	if r.Capacity != 0 || r.CapacityMechanic != bardKey+"bard/inspiration/charisma-modifier-minimum-one" || len(r.ValuesByLevel) != 0 || !reflect.DeepEqual(r.RequiredMechanics, []string{
		bardKey + "bard/inspiration/charisma-modifier-minimum-one",
		bardKey + "bard/inspiration/long-rest-all",
		bardKey + "bard/inspiration/level-5-short-rest-all",
	}) {
		return fmt.Errorf("Inspiration capacity/recovery differs from source")
	}
	r = records["feature/bard/bardic-inspiration"]
	if !reflect.DeepEqual(r.RequiredMechanics, []string{
		bardKey + "bard/bardic-inspiration", bardKey + "bard/inspiration/bonus-action-other-creature",
		bardKey + "bard/inspiration/target-see-or-hear-bard", bardKey + "bard/inspiration/one-die-per-creature",
		bardKey + "bard/inspiration/failed-d20-test-add-roll-once", bardKey + "bard/inspiration/expended-when-rolled",
		bardKey + "bard/inspiration/range-feet", bardKey + "bard/inspiration/duration-hours",
		bardKey + "bard/inspiration/minimum-uses", bardKey + "bard/inspiration/dice-per-creature", bardKey + "bard/inspiration/uses-per-die",
	}) || !reflect.DeepEqual(r.Parameters, map[string]int{
		bardKey + "bard/inspiration/range-feet": 60, bardKey + "bard/inspiration/duration-hours": 1,
		bardKey + "bard/inspiration/minimum-uses": 1, bardKey + "bard/inspiration/dice-per-creature": 1, bardKey + "bard/inspiration/uses-per-die": 1,
	}) {
		return fmt.Errorf("Inspiration conditions/numbers differ from source")
	}
	r = records["choice/bard/expertise"]
	counts := []CharacterLevelValue{}
	for n := 2; n <= 20; n++ {
		value := 2
		if n >= 9 {
			value = 4
		}
		counts = append(counts, CharacterLevelValue{Level: n, Value: value})
	}
	if r.Select != 2 || !reflect.DeepEqual(r.ValuesByLevel, counts) {
		return fmt.Errorf("Expertise cumulative selections differ from source")
	}
	r = records["prerequisite/bard/equipment-a"]
	if !reflect.DeepEqual(r.Requirement, &CharacterRequirement{
		Key: bardKey + "bard/equipment/a-selected", RecordRefs: []string{bardID + "choice/bard/starting-equipment"}, Parameters: map[string]int{},
	}) || !reflect.DeepEqual(r.RequiredMechanics, []string{bardKey + "bard/equipment/a-selected"}) {
		return fmt.Errorf("equipment instrument condition differs from source")
	}
	return nil
}

func TestBardProseContraryControls(t *testing.T) {
	if err := bardProseSourceError(bardCatalogue(t)); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(map[string]*CharacterRecord){
		"initial recovery": func(r map[string]*CharacterRecord) {
			r["resource/bard/bardic-inspiration"].RequiredMechanics[1] = bardKey + "bard/inspiration/short-rest-all"
		},
		"visibility": func(r map[string]*CharacterRecord) {
			r["feature/bard/bardic-inspiration"].RequiredMechanics[2] = bardKey + "bard/inspiration/target-see-and-hear-bard"
		},
		"equipment operand": func(r map[string]*CharacterRecord) {
			r["prerequisite/bard/equipment-a"].Requirement.RecordRefs = []string{bardID + "choice/bard/cantrips"}
		},
		"Expertise scaling": func(r map[string]*CharacterRecord) {
			for i := range r["choice/bard/expertise"].ValuesByLevel {
				if r["choice/bard/expertise"].ValuesByLevel[i].Level >= 9 {
					r["choice/bard/expertise"].ValuesByLevel[i].Value = 6
				}
			}
		},
		"range": func(r map[string]*CharacterRecord) {
			r["feature/bard/bardic-inspiration"].Parameters[bardKey+"bard/inspiration/range-feet"] = 600
		},
		"fixed capacity": func(r map[string]*CharacterRecord) {
			r["resource/bard/bardic-inspiration"].CapacityMechanic = ""
			r["resource/bard/bardic-inspiration"].Capacity = 5
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := bardCatalogue(t)
			records := map[string]*CharacterRecord{}
			for i := range c.Records {
				records[strings.TrimPrefix(c.Records[i].ID, bardID)] = &c.Records[i]
			}
			mutate(records)
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("source control must remain structurally valid: %v", err)
			}
			if err := bardProseSourceError(c); err == nil {
				t.Fatal("contrary prose metadata accepted")
			}
		})
	}
}

func TestBardExactSourceAndContraryControls(t *testing.T) {
	c := bardCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if got := bardSourceDigest(c); got != bardSourcePin {
		t.Fatalf("review entire source before updating digest: %s", got)
	}
	// Every required mechanic, including each prerequisite's list, is mandatory
	// in the source oracle even when generic schema validation permits omission.
	for i, r := range c.Records {
		for j := range r.RequiredMechanics {
			fresh := bardCatalogue(t)
			fresh.Records[i].RequiredMechanics = append(fresh.Records[i].RequiredMechanics[:j], fresh.Records[i].RequiredMechanics[j+1:]...)
			if bardSourceDigest(fresh) == bardSourcePin {
				t.Fatalf("source oracle accepted deleted mechanic %s", r.ID)
			}
		}
	}
	cases := map[string]func(*CharacterCatalogue){
		"invented fixed inspiration capacity": func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.CapacityMechanic != "" {
					r.Capacity = 5
				}
			}
		},
		"invented inspiration level scaling": func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.CapacityMechanic != "" {
					r.ValuesByLevel = []CharacterLevelValue{{1, 2}, {5, 3}}
				}
			}
		},
		"unknown formula": func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.CapacityMechanic != "" {
					r.CapacityMechanic = bardKey + "unknown/formula"
				}
			}
		},
		"future feature": func(c *CharacterCatalogue) {
			c.Records[1].Progression.FeatureRefs = []string{bardID + "feature/bard/words-of-creation"}
		},
		"missing nested owner": func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} },
		"unknown predicate operand": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].Requirement.RecordRefs = []string{"missing"}
					break
				}
			}
		},
		"wrong page": func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 32 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			fresh := bardCatalogue(t)
			mutate(&fresh)
			if fresh.Validate(os.DirFS("../../..")) == nil {
				t.Fatal("schema accepted malformed metadata")
			}
			if bardSourceDigest(fresh) == bardSourcePin {
				t.Fatal("source control failed")
			}
		})
	}
	// These are structurally valid but contradict the source. Source assertions,
	// rather than a falsely universal schema rule, must reject them.
	contrary := map[string]func(*CharacterCatalogue){
		"short rest from level one": func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.CapacityMechanic != "" {
					r.RequiredMechanics = append(r.RequiredMechanics, bardKey+"bard/inspiration/level-one-short-rest-all")
				}
			}
		},
		"prepared replacement on long rest": func(c *CharacterCatalogue) {
			c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, bardKey+"bard/spellcasting/replace-prepared-on-long-rest")
		},
		"wrong target distance": func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.ID == bardID+"feature/bard/words-of-creation" {
					r.Parameters[bardKey+"bard/words-of-creation/second-target-distance-feet"] = 30
				}
			}
		},
		"flatten equipment condition": func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.Requirement != nil && r.ID == bardID+"prerequisite/bard/equipment-a" {
					r.Requirement.Key = bardKey + "bard/equipment/any-selected"
				}
			}
		},
		"remove predicate mechanics": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].RequiredMechanics = []string{}
					break
				}
			}
		},
	}
	for name, mutate := range contrary {
		t.Run(name, func(t *testing.T) {
			fresh := bardCatalogue(t)
			mutate(&fresh)
			if bardSourceDigest(fresh) == bardSourcePin {
				t.Fatal("contrary source accepted")
			}
		})
	}
}

func TestBardIdentitiesRowsAndSymbolicResource(t *testing.T) {
	c := bardCatalogue(t)
	if err := bardProseSourceError(c); err != nil {
		t.Fatal(err)
	}
	records := map[string]CharacterRecord{}
	identities := map[string]int{}
	for _, r := range c.Records {
		records[r.ID] = r
		if r.Kind == "class" || r.Kind == "feature" {
			identities[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
		if !strings.Contains(r.ID, "/bard") || r.Kind == "subclass" {
			t.Fatal("adjacent definition imported")
		}
	}
	headings := map[int][]string{1: {"bardic-inspiration", "spellcasting"}, 2: {"expertise", "jack-of-all-trades"}, 3: {"bard-subclass"}, 4: {"ability-score-improvement"}, 5: {"font-of-inspiration"}, 7: {"countercharm"}, 10: {"magical-secrets"}, 18: {"superior-inspiration"}, 19: {"epic-boon"}, 20: {"words-of-creation"}}
	want := map[string]int{"srd-5.2.1:planning/classes/bard": 31}
	for n, slugs := range headings {
		for _, s := range slugs {
			p := 32
			if s == "bardic-inspiration" {
				p = 31
			}
			if n >= 7 {
				p = 33
			}
			want[fmt.Sprintf("srd-5.2.1:planning/class-features/bard-level-%d-%s", n, s)] = p
		}
	}
	if !reflect.DeepEqual(identities, want) || len(identities) != 13 {
		t.Fatalf("identities/pages: %v", identities)
	}
	// p31: die, cantrips, prepared spells and all nine spell-slot columns.
	rows := [20][12]int{
		{6, 2, 4, 2, 0, 0, 0, 0, 0, 0, 0, 0}, {6, 2, 5, 3, 0, 0, 0, 0, 0, 0, 0, 0}, {6, 2, 6, 4, 2, 0, 0, 0, 0, 0, 0, 0}, {6, 3, 7, 4, 3, 0, 0, 0, 0, 0, 0, 0},
		{8, 3, 9, 4, 3, 2, 0, 0, 0, 0, 0, 0}, {8, 3, 10, 4, 3, 3, 0, 0, 0, 0, 0, 0}, {8, 3, 11, 4, 3, 3, 1, 0, 0, 0, 0, 0}, {8, 3, 12, 4, 3, 3, 2, 0, 0, 0, 0, 0},
		{8, 3, 14, 4, 3, 3, 3, 1, 0, 0, 0, 0}, {10, 4, 15, 4, 3, 3, 3, 2, 0, 0, 0, 0}, {10, 4, 16, 4, 3, 3, 3, 2, 1, 0, 0, 0}, {10, 4, 16, 4, 3, 3, 3, 2, 1, 0, 0, 0},
		{10, 4, 17, 4, 3, 3, 3, 2, 1, 1, 0, 0}, {10, 4, 17, 4, 3, 3, 3, 2, 1, 1, 0, 0}, {12, 4, 18, 4, 3, 3, 3, 2, 1, 1, 1, 0}, {12, 4, 18, 4, 3, 3, 3, 2, 1, 1, 1, 0},
		{12, 4, 19, 4, 3, 3, 3, 2, 1, 1, 1, 1}, {12, 4, 20, 4, 3, 3, 3, 3, 1, 1, 1, 1}, {12, 4, 21, 4, 3, 3, 3, 3, 2, 1, 1, 1}, {12, 4, 22, 4, 3, 3, 3, 3, 2, 2, 1, 1},
	}
	count := func(id string, n int) int {
		for _, v := range records[bardID+id].ValuesByLevel {
			if v.Level == n {
				return v.Value
			}
		}
		return 0
	}
	levels := 0
	for _, r := range c.Records {
		if r.Kind == "level" {
			levels++
		}
	}
	if levels != 20 {
		t.Fatal("missing rows")
	}
	for i, row := range rows {
		n := i + 1
		r := records[fmt.Sprintf("%slevel/bard/%d", bardID, n)]
		p := r.Progression
		if p == nil || p.ProficiencyBonus != 2+(n-1)/4 || r.Parameters[bardKey+"bard/bardic-die-sides"] != row[0] {
			t.Fatalf("row %d", n)
		}
		fs := headings[n]
		if n == 8 || n == 12 || n == 16 {
			fs = []string{"ability-score-improvement"}
		}
		if n == 9 {
			fs = []string{"expertise"}
		}
		refs := []string{}
		for _, f := range fs {
			refs = append(refs, bardID+"feature/bard/"+f)
		}
		keys := []string{}
		if n == 6 || n == 14 {
			keys = []string{bardKey + "bard/subclass-feature"}
		}
		if !reflect.DeepEqual(p.FeatureRefs, refs) || !reflect.DeepEqual(p.FeatureKeys, keys) {
			t.Fatalf("row %d feature improvement", n)
		}
		ids := []string{"choice/bard/cantrips", "choice/bard/prepared-spells"}
		for j := 1; j <= 9; j++ {
			ids = append(ids, fmt.Sprintf("resource/bard/spell-slot-%d", j))
		}
		for j, id := range ids {
			if count(id, n) != row[j+1] {
				t.Fatalf("row %d %s", n, id)
			}
		}
	}
	r := records[bardID+"resource/bard/bardic-inspiration"]
	if r.Capacity != 0 || len(r.ValuesByLevel) != 0 || r.CapacityMechanic != bardKey+"bard/inspiration/charisma-modifier-minimum-one" {
		t.Fatal("invented fixed or scaling capacity")
	}
}

func TestBardCompositionAndUnavailable(t *testing.T) {
	c := bardCatalogue(t)
	repo := os.DirFS("../../..")
	for _, name := range []string{"advancement", "fighter", "champion", "cleric", "life-domain", "barbarian", "berserker", "backgrounds", "origin-feats", "fighting-styles", "species"} {
		b, err := os.ReadFile("testdata/character-" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		c.Records = append(c.Records, other.Records...)
	}
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if c.ValidateReadiness(repo) == nil {
		t.Fatal("catalogue promoted readiness")
	}
	c.Records = append(c.Records, c.Records[0])
	if c.Validate(repo) == nil {
		t.Fatal("duplicate composition accepted")
	}
	c = bardCatalogue(t)
	c.Records[0].RequiredMechanics = append([]string{bardKey + "unknown/bard"}, c.Records[0].RequiredMechanics...)
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil || !strings.Contains(err.Error(), "unknown/bard") {
		t.Fatal("unknown mechanics available")
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCharacterCatalogue(strings.NewReader(strings.Replace(string(b), `"schema_version":1`, `"schema_version":1,"playable":true`, 1))); err == nil {
		t.Fatal("unknown support field accepted")
	}
	coverage, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if coverage.FullSupportEligible(repo) {
		t.Fatal("Full Rules promoted")
	}
}

func TestBardExactNestedChoicesNumbersAndPredicates(t *testing.T) {
	c := bardCatalogue(t)
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		records[r.ID] = r
	}
	// Independent source review surface: exact counts/domains, quantities,
	// numeric parameters, fixed spells and every prerequisite with its operands.
	expected := map[string]string{
		"srd-5.2.1:character/class/bard":                                                `{"parameters":{"srd-5.2.1:mechanic/bard/hit-die-sides":8}}`,
		"srd-5.2.1:character/feature/bard/bardic-inspiration":                           `{"parameters":{"srd-5.2.1:mechanic/bard/inspiration/dice-per-creature":1,"srd-5.2.1:mechanic/bard/inspiration/duration-hours":1,"srd-5.2.1:mechanic/bard/inspiration/minimum-uses":1,"srd-5.2.1:mechanic/bard/inspiration/range-feet":60,"srd-5.2.1:mechanic/bard/inspiration/uses-per-die":1}}`,
		"srd-5.2.1:character/feature/bard/spellcasting":                                 `{"parameters":{"srd-5.2.1:mechanic/bard/spellcasting/cantrips-replaced-per-level":1,"srd-5.2.1:mechanic/bard/spellcasting/prepared-replaced-per-level":1}}`,
		"srd-5.2.1:character/feature/bard/font-of-inspiration":                          `{"parameters":{"srd-5.2.1:mechanic/bard/font-of-inspiration/slots-expended":1,"srd-5.2.1:mechanic/bard/font-of-inspiration/uses-regained":1}}`,
		"srd-5.2.1:character/feature/bard/countercharm":                                 `{"parameters":{"srd-5.2.1:mechanic/bard/countercharm/range-feet":30}}`,
		"srd-5.2.1:character/feature/bard/superior-inspiration":                         `{"parameters":{"srd-5.2.1:mechanic/bard/superior-inspiration/recovery-threshold":2}}`,
		"srd-5.2.1:character/feature/bard/words-of-creation":                            `{"always_prepared":[{"level":20,"spell_keys":["srd-5.2.1:mechanic/spells/power-word-heal","srd-5.2.1:mechanic/spells/power-word-kill"]}],"parameters":{"srd-5.2.1:mechanic/bard/words-of-creation/additional-targets":1,"srd-5.2.1:mechanic/bard/words-of-creation/second-target-distance-feet":10}}`,
		"srd-5.2.1:character/choice/bard/skills":                                        `{"option_keys":["srd-5.2.1:mechanic/skills/acrobatics","srd-5.2.1:mechanic/skills/animal-handling","srd-5.2.1:mechanic/skills/arcana","srd-5.2.1:mechanic/skills/athletics","srd-5.2.1:mechanic/skills/deception","srd-5.2.1:mechanic/skills/history","srd-5.2.1:mechanic/skills/insight","srd-5.2.1:mechanic/skills/intimidation","srd-5.2.1:mechanic/skills/investigation","srd-5.2.1:mechanic/skills/medicine","srd-5.2.1:mechanic/skills/nature","srd-5.2.1:mechanic/skills/perception","srd-5.2.1:mechanic/skills/performance","srd-5.2.1:mechanic/skills/persuasion","srd-5.2.1:mechanic/skills/religion","srd-5.2.1:mechanic/skills/sleight-of-hand","srd-5.2.1:mechanic/skills/stealth","srd-5.2.1:mechanic/skills/survival"],"select":3}`,
		"srd-5.2.1:character/choice/bard/musical-instruments":                           `{"option_keys":["srd-5.2.1:mechanic/tools/musical-instruments"],"select":3}`,
		"srd-5.2.1:character/choice/bard/starting-equipment":                            `{"option_keys":["srd-5.2.1:mechanic/bard/equipment/a","srd-5.2.1:mechanic/bard/equipment/b"],"options":[{"key":"srd-5.2.1:mechanic/bard/equipment/a","quantities":{"srd-5.2.1:mechanic/currency/gp":19,"srd-5.2.1:mechanic/equipment/dagger":2,"srd-5.2.1:mechanic/equipment/entertainers-pack":1,"srd-5.2.1:mechanic/equipment/leather-armor":1,"srd-5.2.1:mechanic/tools/musical-instruments":1}},{"key":"srd-5.2.1:mechanic/bard/equipment/b","quantities":{"srd-5.2.1:mechanic/currency/gp":90}}],"select":1}`,
		"srd-5.2.1:character/choice/bard/equipment-instrument":                          `{"option_keys":["srd-5.2.1:mechanic/tools/musical-instruments"],"select":1}`,
		"srd-5.2.1:character/prerequisite/bard/equipment-a":                             `{"required_mechanics":["srd-5.2.1:mechanic/bard/equipment/a-selected"],"requirement":{"key":"srd-5.2.1:mechanic/bard/equipment/a-selected","parameters":{},"record_refs":["srd-5.2.1:character/choice/bard/starting-equipment"]}}`,
		"srd-5.2.1:character/choice/bard/multiclass-traits":                             `{"option_keys":["srd-5.2.1:mechanic/bard/multiclass/core-traits-subset"],"select":1}`,
		"srd-5.2.1:character/prerequisite/bard/multiclass":                              `{"required_mechanics":["srd-5.2.1:mechanic/character/multiclass-entry"],"requirement":{"key":"srd-5.2.1:mechanic/character/multiclass-entry","parameters":{},"record_refs":["srd-5.2.1:character/class/bard"]}}`,
		"srd-5.2.1:character/choice/bard/multiclass-skill":                              `{"option_keys":["srd-5.2.1:mechanic/skills/acrobatics","srd-5.2.1:mechanic/skills/animal-handling","srd-5.2.1:mechanic/skills/arcana","srd-5.2.1:mechanic/skills/athletics","srd-5.2.1:mechanic/skills/deception","srd-5.2.1:mechanic/skills/history","srd-5.2.1:mechanic/skills/insight","srd-5.2.1:mechanic/skills/intimidation","srd-5.2.1:mechanic/skills/investigation","srd-5.2.1:mechanic/skills/medicine","srd-5.2.1:mechanic/skills/nature","srd-5.2.1:mechanic/skills/perception","srd-5.2.1:mechanic/skills/performance","srd-5.2.1:mechanic/skills/persuasion","srd-5.2.1:mechanic/skills/religion","srd-5.2.1:mechanic/skills/sleight-of-hand","srd-5.2.1:mechanic/skills/stealth","srd-5.2.1:mechanic/skills/survival"],"select":1}`,
		"srd-5.2.1:character/choice/bard/multiclass-instrument":                         `{"option_keys":["srd-5.2.1:mechanic/tools/musical-instruments"],"select":1}`,
		"srd-5.2.1:character/choice/bard/cantrips":                                      `{"option_keys":["srd-5.2.1:mechanic/bard/spell-list/p33-35/cantrips"],"select":2}`,
		"srd-5.2.1:character/choice/bard/prepared-spells":                               `{"option_keys":["srd-5.2.1:mechanic/bard/spell-list/p33-35/level-1-plus"],"select":4}`,
		"srd-5.2.1:character/prerequisite/bard/prepared-slot-eligibility":               `{"required_mechanics":["srd-5.2.1:mechanic/bard/spellcasting/has-slot-for-chosen-spell"],"requirement":{"key":"srd-5.2.1:mechanic/bard/spellcasting/has-slot-for-chosen-spell","parameters":{},"record_refs":["srd-5.2.1:character/feature/bard/spellcasting"]}}`,
		"srd-5.2.1:character/choice/bard/expertise":                                     `{"option_keys":["srd-5.2.1:mechanic/skills/acrobatics","srd-5.2.1:mechanic/skills/animal-handling","srd-5.2.1:mechanic/skills/arcana","srd-5.2.1:mechanic/skills/athletics","srd-5.2.1:mechanic/skills/deception","srd-5.2.1:mechanic/skills/history","srd-5.2.1:mechanic/skills/insight","srd-5.2.1:mechanic/skills/intimidation","srd-5.2.1:mechanic/skills/investigation","srd-5.2.1:mechanic/skills/medicine","srd-5.2.1:mechanic/skills/nature","srd-5.2.1:mechanic/skills/perception","srd-5.2.1:mechanic/skills/performance","srd-5.2.1:mechanic/skills/persuasion","srd-5.2.1:mechanic/skills/religion","srd-5.2.1:mechanic/skills/sleight-of-hand","srd-5.2.1:mechanic/skills/stealth","srd-5.2.1:mechanic/skills/survival"],"select":2}`,
		"srd-5.2.1:character/prerequisite/bard/expertise-proficiency":                   `{"required_mechanics":["srd-5.2.1:mechanic/character/skill-proficient"],"requirement":{"key":"srd-5.2.1:mechanic/character/skill-proficient","parameters":{},"record_refs":["srd-5.2.1:character/class/bard"]}}`,
		"srd-5.2.1:character/prerequisite/bard/jack-of-all-trades":                      `{"required_mechanics":["srd-5.2.1:mechanic/bard/jack-of-all-trades/ability-check-uses-untrained-skill-and-not-proficiency"],"requirement":{"key":"srd-5.2.1:mechanic/bard/jack-of-all-trades/ability-check-uses-untrained-skill-and-not-proficiency","parameters":{},"record_refs":["srd-5.2.1:character/class/bard"]}}`,
		"srd-5.2.1:character/choice/bard/bard-subclass":                                 `{"option_keys":["srd-5.2.1:mechanic/bard/subclasses","srd-5.2.1:mechanic/bard/subclasses/college-of-lore"],"select":1}`,
		"srd-5.2.1:character/choice/bard/ability-score-improvement":                     `{"option_keys":["srd-5.2.1:mechanic/feats/ability-score-improvement","srd-5.2.1:mechanic/feats/qualified"],"select":1}`,
		"srd-5.2.1:character/prerequisite/bard/ability-score-improvement-qualification": `{"required_mechanics":["srd-5.2.1:mechanic/prerequisite/feat-qualification"],"requirement":{"key":"srd-5.2.1:mechanic/prerequisite/feat-qualification","parameters":{},"record_refs":[]}}`,
		"srd-5.2.1:character/choice/bard/epic-boon":                                     `{"option_keys":["srd-5.2.1:mechanic/feats/epic-boon","srd-5.2.1:mechanic/feats/qualified"],"select":1}`,
		"srd-5.2.1:character/prerequisite/bard/epic-boon-qualification":                 `{"required_mechanics":["srd-5.2.1:mechanic/prerequisite/feat-qualification"],"requirement":{"key":"srd-5.2.1:mechanic/prerequisite/feat-qualification","parameters":{},"record_refs":[]}}`,
		"srd-5.2.1:character/prerequisite/bard/font-expended-use":                       `{"required_mechanics":["srd-5.2.1:mechanic/bard/inspiration/has-expended-use"],"requirement":{"key":"srd-5.2.1:mechanic/bard/inspiration/has-expended-use","parameters":{},"record_refs":["srd-5.2.1:character/feature/bard/bardic-inspiration","srd-5.2.1:character/feature/bard/spellcasting"]}}`,
		"srd-5.2.1:character/prerequisite/bard/countercharm":                            `{"required_mechanics":["srd-5.2.1:mechanic/bard/countercharm/self-or-creature-failed-save-against-charmed-or-frightened"],"requirement":{"key":"srd-5.2.1:mechanic/bard/countercharm/self-or-creature-failed-save-against-charmed-or-frightened","parameters":{"range-feet":30},"record_refs":[]}}`,
		"srd-5.2.1:character/choice/bard/magical-secrets":                               `{"option_keys":["srd-5.2.1:mechanic/bard/spell-list/p33-35/level-1-plus","srd-5.2.1:mechanic/cleric/spell-list/level-1-plus","srd-5.2.1:mechanic/druid/spell-list/level-1-plus","srd-5.2.1:mechanic/wizard/spell-list/level-1-plus"],"select":1}`,
		"srd-5.2.1:character/prerequisite/bard/magical-secrets":                         `{"required_mechanics":["srd-5.2.1:mechanic/bard/magical-secrets/bard-level-at-least-10-and-prepared-count-increases-or-replacement"],"requirement":{"key":"srd-5.2.1:mechanic/bard/magical-secrets/bard-level-at-least-10-and-prepared-count-increases-or-replacement","parameters":{"minimum-bard-level":10},"record_refs":["srd-5.2.1:character/choice/bard/prepared-spells","srd-5.2.1:character/feature/bard/spellcasting"]}}`,
		"srd-5.2.1:character/prerequisite/bard/superior-inspiration":                    `{"required_mechanics":["srd-5.2.1:mechanic/bard/inspiration/remaining-uses-less-than-two"],"requirement":{"key":"srd-5.2.1:mechanic/bard/inspiration/remaining-uses-less-than-two","parameters":{"uses-threshold":2},"record_refs":["srd-5.2.1:character/feature/bard/bardic-inspiration"]}}`,
		"srd-5.2.1:character/prerequisite/bard/words-second-target":                     `{"required_mechanics":["srd-5.2.1:mechanic/bard/words-of-creation/second-creature-within-ten-feet-of-first"],"requirement":{"key":"srd-5.2.1:mechanic/bard/words-of-creation/second-creature-within-ten-feet-of-first","parameters":{"distance-feet":10},"record_refs":[]}}`,
	}
	for id, want := range expected {
		b, err := json.Marshal(records[id])
		if err != nil {
			t.Fatal(err)
		}
		var actual, subset map[string]json.RawMessage
		if err := json.Unmarshal(b, &actual); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(want), &subset); err != nil {
			t.Fatal(err)
		}
		for field, value := range subset {
			var got, w any
			if err := json.Unmarshal(actual[field], &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(value, &w); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, w) {
				t.Fatalf("%s %s differs from source", id, field)
			}
		}
	}
}
