package srd521

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

const originID = "srd-5.2.1:character/"
const originKey = "srd-5.2.1:mechanic/"

func originFeatsCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-origin-feats.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// This independent source transcription checks the complete graph, including
// every predicate operand and numeric parameter, rather than just root presence.
func TestOriginFeatsExactSource(t *testing.T) {
	c := originFeatsCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	expected := []CharacterRecord{}
	keys := func(parts ...string) []string {
		result := []string{}
		for _, p := range parts {
			result = append(result, originKey+p)
		}
		return result
	}
	add := func(kind, slug, name, parent string, mechanics ...string) *CharacterRecord {
		expected = append(expected, CharacterRecord{ID: originID + kind + "/" + slug, Kind: kind, Name: name, ParentID: parent,
			Sources:    []CharacterSource{{InventoryID: "srd-5.2.1:planning/feats/" + strings.Split(slug, "/")[0], Page: 87}},
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, OptionKeys: []string{}, RequiredMechanics: keys(mechanics...)})
		return &expected[len(expected)-1]
	}
	// Store indices, since appending may move the backing array.
	feat := func(slug, name, repeat string, mechanics ...string) int {
		i := len(expected)
		add("feat", slug, name, "", append([]string{"feats/" + slug, "feats/category/origin", "feats/repeatability/" + repeat}, mechanics...)...)
		return i
	}
	choice := func(owner int, part, name string, count int, options []string, mechanics ...string) int {
		slug := strings.SplitN(strings.TrimPrefix(expected[owner].ID, originID), "/", 2)[1] + "/" + part
		i := len(expected)
		r := add("choice", slug, name, expected[owner].ID, mechanics...)
		r.Select = count
		r.OptionKeys = keys(options...)
		expected[owner].ChoiceRefs = append(expected[owner].ChoiceRefs, r.ID)
		return i
	}
	pred := func(owner int, part, key string, operands []int, params map[string]int) {
		refs := []string{}
		for _, i := range operands {
			refs = append(refs, expected[i].ID)
		}
		slug := strings.SplitN(strings.TrimPrefix(expected[owner].ID, originID), "/", 2)[1] + "/" + part
		name := strings.ReplaceAll(part, "-", " ")
		name = strings.ToUpper(name[:1]) + name[1:]
		r := add("prerequisite", slug, name, expected[owner].ID, "prerequisite/"+key)
		r.Requirement = &CharacterRequirement{Key: originKey + "prerequisite/" + key, RecordRefs: refs, Parameters: params}
		expected[owner].PrerequisiteRefs = append(expected[owner].PrerequisiteRefs, r.ID)
	}
	a := feat("alert", "Alert", "once", "feats/alert/initiative-proficiency/when-you-roll-initiative", "feats/alert/initiative-proficiency/optional-add-proficiency-bonus", "feats/alert/initiative-swap/optional-exchange-initiative")
	ally := choice(a, "initiative-swap", "Initiative Swap", 1, []string{"targets/willing-ally"}, "feats/alert/initiative-swap/immediately-after-you-roll-initiative")
	pred(ally, "same-combat", "alert-ally-in-same-combat", []int{a, ally}, map[string]int{})
	pred(ally, "self-not-incapacitated", "alert-self-not-incapacitated", []int{a}, map[string]int{})
	pred(ally, "ally-not-incapacitated", "alert-ally-not-incapacitated", []int{ally}, map[string]int{})
	m := feat("magic-initiate", "Magic Initiate", "different-spell-list", "feats/magic-initiate/two-cantrips/learn", "feats/magic-initiate/level-one-spell/always-prepared", "feats/magic-initiate/level-one-spell/cast-using-any-owned-spell-slots", "feats/magic-initiate/spell-change/optional-on-gaining-new-level")
	list := choice(m, "spell-list", "Spell list", 1, []string{"spell-lists/cleric", "spell-lists/druid", "spell-lists/wizard"}, "feats/magic-initiate/choose-one-spell-list")
	choice(m, "spellcasting-ability", "Spellcasting ability", 1, []string{"abilities/intelligence", "abilities/wisdom", "abilities/charisma"}, "feats/magic-initiate/ability/chosen-when-selecting-feat", "feats/magic-initiate/ability/applies-to-this-feats-spells")
	can := choice(m, "cantrips", "Two Cantrips", 2, []string{"spells/selected-list/cantrips"}, "feats/magic-initiate/cantrips/from-selected-list")
	pred(can, "selected-list", "magic-initiate-from-selected-spell-list", []int{list}, map[string]int{"spell-level": 0})
	spell := choice(m, "level-one-spell", "Level 1 Spell", 1, []string{"spells/selected-list/level-one"}, "feats/magic-initiate/level-one-spell/from-same-list-as-cantrips")
	pred(spell, "selected-list", "magic-initiate-from-selected-spell-list", []int{list, can}, map[string]int{"spell-level": 1})
	r := add("resource", "magic-initiate/free-cast", "Level 1 spell without a spell slot", expected[m].ID, "feats/magic-initiate/free-cast/selected-level-one-spell", "feats/magic-initiate/free-cast/without-spell-slot", "feats/magic-initiate/free-cast/regain-on-finish-long-rest")
	r.Capacity = 1
	expected[m].ResourceRefs = append(expected[m].ResourceRefs, r.ID)
	replace := choice(m, "spell-change", "Spell Change", 1, []string{"spells/selected-list/replacement"}, "feats/magic-initiate/spell-change/replace-one-spell-chosen-for-this-feat")
	pred(replace, "new-level", "magic-initiate-on-gaining-new-level", []int{m}, map[string]int{})
	pred(replace, "same-list-and-level", "magic-initiate-replacement-same-level-from-selected-list", []int{list, can, spell, replace}, map[string]int{})
	pred(m, "repeatability", "magic-initiate-different-spell-list-each-time", []int{m, list}, map[string]int{})
	s := feat("savage-attacker", "Savage Attacker", "once", "feats/savage-attacker/optional-on-hit-target-with-weapon", "feats/savage-attacker/roll-weapons-damage-dice-twice", "feats/savage-attacker/use-either-roll-against-target", "feats/savage-attacker/roll-count", "feats/savage-attacker/uses-per-turn")
	expected[s].Parameters = map[string]int{originKey + "feats/savage-attacker/roll-count": 2, originKey + "feats/savage-attacker/uses-per-turn": 1}
	choice(s, "damage-roll", "Weapon damage roll", 1, []string{"feats/savage-attacker/damage-roll/first", "feats/savage-attacker/damage-roll/second"}, "feats/savage-attacker/use-either-roll-against-target")
	sk := feat("skilled", "Skilled", "unrestricted", "feats/skilled/gain-proficiency")
	choice(sk, "proficiencies", "Skill or tool proficiencies", 3, []string{"proficiencies/any-skill", "proficiencies/any-tool"}, "feats/skilled/any-combination-of-skills-or-tools")
	if !reflect.DeepEqual(c.Records, expected) {
		for i, r := range expected {
			if i >= len(c.Records) || !reflect.DeepEqual(c.Records[i], r) {
				t.Errorf("source record %s mismatch", r.ID)
			}
		}
		t.Fatalf("exact four identities and complete nested metadata differ (got %d, want %d)", len(c.Records), len(expected))
	}
}

func TestOriginFeatsCompositionAndGuards(t *testing.T) {
	repo := os.DirFS("../../..")
	c := originFeatsCatalogue(t)
	for _, name := range []string{"advancement", "backgrounds", "cleric", "fighter", "champion"} {
		b, err := os.ReadFile("testdata/character-" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateSource(other.Source); err != nil {
			t.Fatal(err)
		}
		c.Records = append(c.Records, other.Records...)
	}
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("metadata ready")
	}
	cases := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"duplicate", func(c *CharacterCatalogue) { c.Records = append(c.Records, c.Records[0]) }},
		{"page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 88 }},
		{"pin", func(c *CharacterCatalogue) { c.Source.SHA256 = "bad" }},
		{"attribution", func(c *CharacterCatalogue) { c.Source.Attribution = "" }},
		{"adaptation", func(c *CharacterCatalogue) { c.Source.Changes = "" }},
		{"source", func(c *CharacterCatalogue) { c.Records[0].Sources[0].InventoryID = "unknown" }},
		{"family", func(c *CharacterCatalogue) {
			c.Records[0].Sources[0] = CharacterSource{InventoryID: "srd-5.2.1:planning/backgrounds/acolyte", Page: 83}
		}},
		{"ownership", func(c *CharacterCatalogue) { c.Records[1].Sources = c.Records[5].Sources }},
		{"orphan", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"owner", func(c *CharacterCatalogue) { c.Records[1].ParentID = c.Records[5].ID }},
		{"count", func(c *CharacterCatalogue) { c.Records[1].Select = 0 }},
		{"malformed key", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics[0] = "bad key" }},
		{"predicate", func(c *CharacterCatalogue) { c.Records[2].Requirement.Key = originKey + "unknown" }},
		{"operand", func(c *CharacterCatalogue) { c.Records[2].Requirement.RecordRefs = []string{originID + "feat/missing"} }},
		{"duplicate operand", func(c *CharacterCatalogue) {
			c.Records[2].Requirement.RecordRefs = []string{c.Records[0].ID, c.Records[0].ID}
		}},
		{"parameter", func(c *CharacterCatalogue) { c.Records[2].Requirement.Parameters["limit"] = -1 }},
		{"numeric key", func(c *CharacterCatalogue) { c.Records[17].Parameters = map[string]int{originKey + "unknown": 2} }},
		{"numeric zero", func(c *CharacterCatalogue) {
			c.Records[17].Parameters = map[string]int{originKey + "feats/savage-attacker/roll-count": 0}
		}},
		{"numeric owner", func(c *CharacterCatalogue) {
			c.Records[1].Parameters = map[string]int{c.Records[1].RequiredMechanics[0]: 1}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := originFeatsCatalogue(t)
			tc.mutate(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("invalid accepted")
			}
		})
	}
	for _, field := range []string{"available", "implemented", "playable", "unknown"} {
		b, err := json.Marshal(originFeatsCatalogue(t))
		if err != nil {
			t.Fatal(err)
		}
		input := strings.Replace(string(b), `"capacity":0`, `"capacity":0,"`+field+`":true`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("unknown field accepted")
		}
	}
	for _, domain := range []string{"required", "option", "predicate"} {
		c := originFeatsCatalogue(t)
		key := originKey + "feats/unknown"
		switch domain {
		case "required":
			c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, key)
		case "option":
			c.Records[1].OptionKeys = append(c.Records[1].OptionKeys, key)
		case "predicate":
			c.Records[2].RequiredMechanics = []string{key}
			c.Records[2].Requirement.Key = key
		}
		if err := c.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(repo); err == nil {
			t.Fatal("unknown mechanic ready")
		}
	}
}

func TestOriginFeatsLeaveFixedRuntimeUnchanged(t *testing.T) {
	repo := os.DirFS("../../..")
	coverageBefore, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		concept, background, feat string
		hp, gold                  int
	}{
		{"Human Fighter", "Soldier", "Savage Attacker", 12, 29},
		{"Human Cleric", "Acolyte", "Magic Initiate (Cleric; Wisdom)", 10, 118},
	} {
		before, err := SupportedBuildForConcept(tc.concept, tc.background)
		if err != nil {
			t.Fatal(err)
		}
		stats, err := before.Stats()
		if err != nil {
			t.Fatal(err)
		}
		sheetBefore, err := Project(stats, 1, tc.hp, tc.hp, 0)
		if err != nil {
			t.Fatal(err)
		}
		c := originFeatsCatalogue(t)
		if err := c.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(repo); err == nil {
			t.Fatal("metadata ready")
		}
		after, err := SupportedBuildForConcept(tc.concept, tc.background)
		if err != nil {
			t.Fatal(err)
		}
		statsAfter, err := after.Stats()
		if err != nil {
			t.Fatal(err)
		}
		sheetAfter, err := Project(statsAfter, 1, tc.hp, tc.hp, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(sheetBefore, sheetAfter) || after.Gold != tc.gold || after.OriginFeat != tc.feat {
			t.Fatalf("fixed runtime changed: %+v", after)
		}
		if tc.background == "Acolyte" {
			if sheetAfter.Spellcasting.Choices.FeatUses != 1 {
				t.Fatal("feat uses changed")
			}
			for _, spell := range sheetAfter.Spellcasting.Spells {
				if spell.ID == "detect-magic" && spell.Playable {
					t.Fatal("feat effect enabled")
				}
			}
		}
		for _, background := range []string{"Criminal", "Sage"} {
			if _, err := SupportedBuildForConcept(tc.concept, background); err == nil {
				t.Fatal("creation widened")
			}
		}
	}
	coverageAfter, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(coverageBefore, coverageAfter) || coverageAfter.FullSupportEligible(repo) {
		t.Fatal("coverage promoted")
	}
}
