package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const generalID = "srd-5.2.1:character/"
const generalKey = "srd-5.2.1:mechanic/"

func generalFeatsCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-general-feats.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// Independent transcription of the complete PDF p87 graph. Structural validity
// alone permits future opaque keys; this oracle also rejects contrary source data.
func expectedGeneralFeats() []CharacterRecord {
	expected := []CharacterRecord{}
	keys := func(parts ...string) []string {
		result := []string{}
		for _, p := range parts {
			result = append(result, generalKey+p)
		}
		return result
	}
	add := func(kind, slug, name, parent string, mechanics ...string) *CharacterRecord {
		expected = append(expected, CharacterRecord{ID: generalID + kind + "/" + slug, Kind: kind, Name: name, ParentID: parent,
			Sources:    []CharacterSource{{InventoryID: "srd-5.2.1:planning/feats/" + strings.Split(slug, "/")[0], Page: 87}},
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, OptionKeys: []string{}, RequiredMechanics: keys(mechanics...)})
		return &expected[len(expected)-1]
	}
	// Store indices, since appending may move the backing array.
	feat := func(slug, name, repeat string, mechanics ...string) int {
		i := len(expected)
		add("feat", slug, name, "", append([]string{"feats/category/general", "feats/repeatability/" + repeat}, mechanics...)...)
		return i
	}
	choice := func(owner int, part, name string, count int, options []string, mechanics ...string) int {
		slug := strings.SplitN(strings.TrimPrefix(expected[owner].ID, generalID), "/", 2)[1] + "/" + part
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
		slug := strings.SplitN(strings.TrimPrefix(expected[owner].ID, generalID), "/", 2)[1] + "/" + part
		name := strings.ReplaceAll(part, "-", " ")
		name = strings.ToUpper(name[:1]) + name[1:]
		r := add("prerequisite", slug, name, expected[owner].ID, "prerequisite/"+key)
		r.Requirement = &CharacterRequirement{Key: generalKey + "prerequisite/" + key, RecordRefs: refs, Parameters: params}
		expected[owner].PrerequisiteRefs = append(expected[owner].PrerequisiteRefs, r.ID)
	}

	a := feat("ability-score-improvement", "Ability Score Improvement", "unrestricted", "feats/ability-score-improvement/maximum-ability-score")
	expected[a].Parameters = map[string]int{generalKey + "feats/ability-score-improvement/maximum-ability-score": 20}
	pred(a, "level", "minimum-character-level", nil, map[string]int{"minimum-level": 4})
	base := "feats/ability-score-improvement/"
	pattern := choice(a, "increase-pattern", "Ability increase pattern", 1, []string{base + "one-score-plus-two", base + "two-scores-plus-one"}, base+"choose-one-increase-pattern", base+"increase-per-score")
	expected[pattern].Options = []CharacterChoiceOption{
		{Key: generalKey + base + "one-score-plus-two", Quantities: map[string]int{generalKey + base + "increase-per-score": 2}},
		{Key: generalKey + base + "two-scores-plus-one", Quantities: map[string]int{generalKey + base + "increase-per-score": 1}},
	}
	for _, tc := range []struct {
		part, name, option string
		count              int
	}{
		{"one-score", "One ability score", "one-score-plus-two", 1},
		{"two-scores", "Two distinct ability scores", "two-scores-plus-one", 2},
	} {
		ch := choice(pattern, tc.part, tc.name, tc.count, []string{"abilities/strength", "abilities/dexterity", "abilities/constitution", "abilities/intelligence", "abilities/wisdom", "abilities/charisma"}, base+"increase-selected-scores", base+"subject-to-maximum-ability-score")
		pred(ch, "selected-pattern", "ability-score-improvement/"+tc.option+"-selected", []int{pattern}, map[string]int{})
	}
	// Grappler defaults to one acquisition under Parts of a Feat, p87.
	g := len(expected)
	add("feat", "grappler", "Grappler", "", "feats/category/general", "feats/maximum-times-taken",
		"feats/grappler/ability-score-increase", "feats/grappler/maximum-ability-score",
		"feats/grappler/punch-and-grab/optional-on-hit-creature", "feats/grappler/punch-and-grab/unarmed-strike",
		"feats/grappler/punch-and-grab/part-of-attack-action-on-your-turn", "feats/grappler/punch-and-grab/use-both-damage-and-grapple-options",
		"feats/grappler/punch-and-grab/uses-per-turn",
		"feats/grappler/attack-advantage/on-attack-rolls-against-creature-grappled-by-you",
		"feats/grappler/fast-wrestler/no-extra-movement-to-move-creature-grappled-by-you",
		"feats/grappler/fast-wrestler/only-if-creature-your-size-or-smaller")
	expected[g].Parameters = map[string]int{generalKey + "feats/maximum-times-taken": 1, generalKey + "feats/grappler/ability-score-increase": 1, generalKey + "feats/grappler/maximum-ability-score": 20, generalKey + "feats/grappler/punch-and-grab/uses-per-turn": 1}
	pred(g, "level", "minimum-character-level", nil, map[string]int{"minimum-level": 4})
	pred(g, "ability", "strength-or-dexterity-at-least", nil, map[string]int{"minimum-score": 13})
	choice(g, "ability-score", "Ability Score Increase", 1, []string{"abilities/strength", "abilities/dexterity"}, "feats/grappler/increase-one-selected-ability", "feats/grappler/subject-to-maximum-ability-score")
	return expected
}

func checkGeneralFeatsSource(c CharacterCatalogue) error {
	if !reflect.DeepEqual(c.Records, expectedGeneralFeats()) {
		return fmt.Errorf("general feats differ from complete page 87 source graph")
	}
	return nil
}

func TestGeneralFeatsExactSource(t *testing.T) {
	c := generalFeatsCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if err := checkGeneralFeatsSource(c); err != nil {
		t.Fatal(err)
	}
	for _, r := range c.Records {
		for _, k := range r.RequiredMechanics {
			if qualifiedCharacterMechanic(k) {
				t.Fatalf("qualified %s", k)
			}
		}
	}
}
func TestGeneralFeatsContrarySource(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"category", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics[0] = generalKey + "feats/category/origin" }},
		{"ASI repeatability", func(c *CharacterCatalogue) {
			c.Records[0].RequiredMechanics[1] = generalKey + "feats/repeatability/once"
		}},
		{"ASI cap", func(c *CharacterCatalogue) {
			c.Records[0].Parameters[generalKey+"feats/ability-score-improvement/maximum-ability-score"] = 22
		}},
		{"ASI level", func(c *CharacterCatalogue) { c.Records[1].Requirement.Parameters["minimum-level"] = 3 }},
		{"ASI alternative amount", func(c *CharacterCatalogue) {
			c.Records[2].Options[0].Quantities[generalKey+"feats/ability-score-improvement/increase-per-score"] = 1
		}},
		{"ASI two distinct scores", func(c *CharacterCatalogue) { c.Records[5].Select = 1 }},
		{"ASI ability alternatives", func(c *CharacterCatalogue) { c.Records[3].OptionKeys = c.Records[3].OptionKeys[:5] }},
		{"ASI branch condition", func(c *CharacterCatalogue) {
			c.Records[4].Requirement.Key = generalKey + "prerequisite/always"
			c.Records[4].RequiredMechanics = []string{c.Records[4].Requirement.Key}
		}},
		{"ASI branch operand", func(c *CharacterCatalogue) { c.Records[6].Requirement.RecordRefs = []string{c.Records[0].ID} }},
		{"Grappler repeatability", func(c *CharacterCatalogue) { c.Records[7].Parameters[generalKey+"feats/maximum-times-taken"] = 2 }},
		{"Grappler increase", func(c *CharacterCatalogue) {
			c.Records[7].Parameters[generalKey+"feats/grappler/ability-score-increase"] = 2
		}},
		{"Grappler cap", func(c *CharacterCatalogue) {
			c.Records[7].Parameters[generalKey+"feats/grappler/maximum-ability-score"] = 21
		}},
		{"Grappler per turn", func(c *CharacterCatalogue) {
			c.Records[7].Parameters[generalKey+"feats/grappler/punch-and-grab/uses-per-turn"] = 2
		}},
		{"Grappler your turn", func(c *CharacterCatalogue) {
			c.Records[7].RequiredMechanics[6] = generalKey + "feats/grappler/any-turn"
		}},
		{"Grappler own grapple", func(c *CharacterCatalogue) {
			c.Records[7].RequiredMechanics[9] = generalKey + "feats/grappler/attack-advantage/any-grapple"
		}},
		{"Grappler size condition", func(c *CharacterCatalogue) { c.Records[7].RequiredMechanics = c.Records[7].RequiredMechanics[:11] }},
		{"Grappler level", func(c *CharacterCatalogue) { c.Records[8].Requirement.Parameters["minimum-level"] = 1 }},
		{"Grappler ability OR", func(c *CharacterCatalogue) {
			c.Records[9].Requirement.Key = generalKey + "prerequisite/strength-and-dexterity-at-least"
			c.Records[9].RequiredMechanics = []string{c.Records[9].Requirement.Key}
		}},
		{"Grappler threshold", func(c *CharacterCatalogue) { c.Records[9].Requirement.Parameters["minimum-score"] = 12 }},
		{"Grappler ability choice", func(c *CharacterCatalogue) { c.Records[10].OptionKeys[1] = generalKey + "abilities/constitution" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := generalFeatsCatalogue(t)
			tc.mutate(&c)
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("contrary fixture must remain structurally valid: %v", err)
			}
			if err := checkGeneralFeatsSource(c); err == nil {
				t.Fatal("contrary source accepted")
			}
		})
	}
}

func TestGeneralFeatsCompositionAndGuards(t *testing.T) {
	repo := os.DirFS("../../..")
	c := generalFeatsCatalogue(t)
	for _, name := range []string{"advancement", "backgrounds", "cleric", "fighter", "champion", "life-domain", "barbarian", "species", "origin-feats", "fighting-styles"} {
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
		{"ownership", func(c *CharacterCatalogue) { c.Records[1].Sources = c.Records[7].Sources }},
		{"orphan", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"owner", func(c *CharacterCatalogue) { c.Records[1].ParentID = c.Records[5].ID }},
		{"count", func(c *CharacterCatalogue) { c.Records[2].Select = 0 }},
		{"malformed key", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics[0] = "bad key" }},
		{"predicate", func(c *CharacterCatalogue) { c.Records[1].Requirement.Key = generalKey + "unknown" }},
		{"operand", func(c *CharacterCatalogue) {
			c.Records[1].Requirement.RecordRefs = []string{generalID + "feat/missing"}
		}},
		{"duplicate operand", func(c *CharacterCatalogue) {
			c.Records[1].Requirement.RecordRefs = []string{c.Records[0].ID, c.Records[0].ID}
		}},
		{"parameter", func(c *CharacterCatalogue) { c.Records[1].Requirement.Parameters["limit"] = -1 }},
		{"numeric key", func(c *CharacterCatalogue) { c.Records[7].Parameters = map[string]int{generalKey + "unknown": 2} }},
		{"numeric zero", func(c *CharacterCatalogue) {
			c.Records[7].Parameters = map[string]int{generalKey + "feats/grappler/ability-score-increase": 0}
		}},
		{"numeric owner", func(c *CharacterCatalogue) {
			c.Records[1].Parameters = map[string]int{c.Records[1].RequiredMechanics[0]: 1}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := generalFeatsCatalogue(t)
			tc.mutate(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("invalid accepted")
			}
		})
	}
	for _, field := range []string{"available", "implemented", "playable", "unknown"} {
		b, err := json.Marshal(generalFeatsCatalogue(t))
		if err != nil {
			t.Fatal(err)
		}
		input := strings.Replace(string(b), `"capacity":0`, `"capacity":0,"`+field+`":true`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("unknown field accepted")
		}
	}
	for _, domain := range []string{"required", "option", "predicate"} {
		c := generalFeatsCatalogue(t)
		key := generalKey + "feats/unknown"
		switch domain {
		case "required":
			c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, key)
		case "option":
			c.Records[3].OptionKeys = append(c.Records[3].OptionKeys, key)
		case "predicate":
			c.Records[1].RequiredMechanics = []string{key}
			c.Records[1].Requirement.Key = key
		}
		if err := c.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(repo); err == nil {
			t.Fatal("unknown mechanic ready")
		}
	}
}

func TestGeneralFeatsLeaveFixedRuntimeUnchanged(t *testing.T) {
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
		c := generalFeatsCatalogue(t)
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
