package srd521

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

const backgroundID = "srd-5.2.1:character/"
const backgroundKey = "srd-5.2.1:mechanic/"

func backgroundsCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-backgrounds.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestBackgroundsExactSourceChoices(t *testing.T) {
	c := backgroundsCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if len(c.Records) != 18 {
		t.Fatal("unexpected batch size")
	}
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		records[r.ID] = r
		if r.Kind != "origin" && r.Kind != "choice" && r.Kind != "prerequisite" {
			t.Fatal("adjacent import")
		}
		for _, key := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(key) {
				t.Fatalf("available: %s", key)
			}
		}
	}
	cases := []struct {
		slug      string
		abilities []string
		feat      string
		skills    []string
		tool      string
		equipment map[string]int
	}{
		{"acolyte", []string{"intelligence", "wisdom", "charisma"}, "magic-initiate/cleric", []string{"insight", "religion"}, "calligraphers-supplies", map[string]int{"calligraphers-supplies": 1, "book/prayers": 1, "holy-symbol": 1, "parchment/sheets": 10, "robe": 1, "currency/gp": 8}},
		{"criminal", []string{"dexterity", "constitution", "intelligence"}, "alert", []string{"sleight-of-hand", "stealth"}, "thieves-tools", map[string]int{"dagger": 2, "thieves-tools": 1, "crowbar": 1, "pouch": 2, "travelers-clothes": 1, "currency/gp": 16}},
		{"sage", []string{"constitution", "intelligence", "wisdom"}, "magic-initiate/wizard", []string{"arcana", "history"}, "calligraphers-supplies", map[string]int{"quarterstaff": 1, "calligraphers-supplies": 1, "book/history": 1, "parchment/sheets": 8, "robe": 1, "currency/gp": 8}},
		{"soldier", []string{"strength", "dexterity", "constitution"}, "savage-attacker", []string{"athletics", "intimidation"}, "", map[string]int{"spear": 1, "shortbow": 1, "arrow": 20, "gaming-set/selected": 1, "healers-kit": 1, "quiver": 1, "travelers-clothes": 1, "currency/gp": 14}},
	}
	expectedIDs := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.slug, func(t *testing.T) {
			id := func(kind, slug string) string { return backgroundID + kind + "/" + tc.slug + slug }
			get := func(kind, slug string) CharacterRecord {
				r, ok := records[id(kind, slug)]
				if !ok {
					t.Fatalf("missing %s", id(kind, slug))
				}
				expectedIDs[r.ID] = true
				if !reflect.DeepEqual(r.Sources, []CharacterSource{{InventoryID: "srd-5.2.1:planning/backgrounds/" + tc.slug, Page: 83}}) {
					t.Fatal("exact source")
				}
				return r
			}
			root := get("origin", "")
			if root.Name != strings.ToUpper(tc.slug[:1])+tc.slug[1:] || root.ParentID != "" {
				t.Fatal("identity")
			}
			wantRequired := []string{backgroundKey + "backgrounds/" + tc.slug, backgroundKey + "feats/" + tc.feat, backgroundKey + "skills/" + tc.skills[0], backgroundKey + "skills/" + tc.skills[1]}
			if tc.tool != "" {
				wantRequired = append(wantRequired, backgroundKey+"tools/"+tc.tool)
			}
			if !reflect.DeepEqual(root.RequiredMechanics, wantRequired) {
				t.Fatal("fixed feat/skills/tool references")
			}
			ability := get("choice", "/ability-increases")
			wantOptions := []CharacterChoiceOption{}
			for i, a := range tc.abilities {
				for j, b := range tc.abilities {
					if i == j {
						continue
					}
					wantOptions = append(wantOptions, CharacterChoiceOption{Key: backgroundKey + "backgrounds/" + tc.slug + "/ability-increases/" + a + "-2-" + b + "-1", Quantities: map[string]int{backgroundKey + "abilities/" + a: 2, backgroundKey + "abilities/" + b: 1}})
				}
			}
			all := map[string]int{}
			wantAbilityKeys := []string{backgroundKey + "backgrounds/ability-increases/distinct-listed-scores"}
			for _, a := range tc.abilities {
				all[backgroundKey+"abilities/"+a] = 1
				wantAbilityKeys = append(wantAbilityKeys, backgroundKey+"abilities/"+a)
			}
			wantOptions = append(wantOptions, CharacterChoiceOption{Key: backgroundKey + "backgrounds/" + tc.slug + "/ability-increases/all-three-1", Quantities: all})
			keys := []string{}
			for _, o := range wantOptions {
				keys = append(keys, o.Key)
			}
			if ability.Select != 1 || ability.ParentID != root.ID || !reflect.DeepEqual(ability.Options, wantOptions) || !reflect.DeepEqual(ability.OptionKeys, keys) || !reflect.DeepEqual(ability.RequiredMechanics, wantAbilityKeys) {
				t.Fatal("exact ability domain/distributions")
			}
			cap := get("prerequisite", "/ability-score-cap")
			wantCap := &CharacterRequirement{Key: backgroundKey + "prerequisite/background-ability-score-maximum", RecordRefs: []string{ability.ID}, Parameters: map[string]int{"maximum-score": 20}}
			if cap.ParentID != ability.ID || !reflect.DeepEqual(cap.Requirement, wantCap) || !reflect.DeepEqual(ability.PrerequisiteRefs, []string{cap.ID}) || !reflect.DeepEqual(cap.RequiredMechanics, []string{wantCap.Key}) {
				t.Fatal("exact cap predicate")
			}
			eq := get("choice", "/starting-equipment")
			quantities := map[string]int{}
			for k, v := range tc.equipment {
				quantities[backgroundKey+"equipment/"+k] = v
			}
			eqKeys := []string{backgroundKey + "backgrounds/" + tc.slug + "/equipment/a", backgroundKey + "backgrounds/" + tc.slug + "/equipment/b"}
			if eq.Select != 1 || eq.ParentID != root.ID || !reflect.DeepEqual(eq.OptionKeys, eqKeys) || !reflect.DeepEqual(eq.Options, []CharacterChoiceOption{{Key: eqKeys[0], Quantities: quantities}, {Key: eqKeys[1], Quantities: map[string]int{backgroundKey + "equipment/currency/gp": 50}}}) {
				t.Fatal("exact equipment alternatives")
			}
			expectedChoices := []string{ability.ID, eq.ID}
			if tc.slug == "soldier" {
				gaming := get("choice", "/gaming-set")
				if gaming.ParentID != root.ID || gaming.Select != 1 || !reflect.DeepEqual(gaming.OptionKeys, []string{backgroundKey + "tools/gaming-set"}) || !reflect.DeepEqual(gaming.RequiredMechanics, []string{backgroundKey + "backgrounds/soldier/one-kind-of-gaming-set"}) {
					t.Fatal("gaming set choice")
				}
				same := get("prerequisite", "/equipment-gaming-set")
				want := &CharacterRequirement{Key: backgroundKey + "prerequisite/equipment-a-matches-selected-gaming-set", RecordRefs: []string{gaming.ID}, Parameters: map[string]int{}}
				if same.ParentID != eq.ID || !reflect.DeepEqual(same.Requirement, want) || !reflect.DeepEqual(same.RequiredMechanics, []string{want.Key}) || !reflect.DeepEqual(eq.PrerequisiteRefs, []string{same.ID}) {
					t.Fatal("exact conditional matching predicate")
				}
				expectedChoices = append(expectedChoices, gaming.ID)
			} else if len(eq.PrerequisiteRefs) != 0 {
				t.Fatal("unexpected prerequisite")
			}
			if !reflect.DeepEqual(root.ChoiceRefs, expectedChoices) {
				t.Fatal("exact nested choices")
			}
		})
	}
	if len(expectedIDs) != len(records) {
		t.Fatal("extra IDs")
	}
}

func TestBackgroundsCompositionAndGuards(t *testing.T) {
	repo := os.DirFS("../../..")
	c := backgroundsCatalogue(t)
	// Compose each canonical catalogue once; the Fighter level-one manifest is
	// a subset reference and must not be concatenated as a second definition.
	for _, name := range []string{"character-advancement.json", "character-fighter.json", "character-champion.json"} {
		b, err := os.ReadFile("testdata/" + name)
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
		t.Fatal("catalogue ready")
	}
	cases := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"duplicate", func(c *CharacterCatalogue) { c.Records = append(c.Records, c.Records[0]) }},
		{"page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 84 }},
		{"pin", func(c *CharacterCatalogue) { c.Source.SHA256 = "bad" }},
		{"attribution", func(c *CharacterCatalogue) { c.Source.Attribution = "" }},
		{"adaptation", func(c *CharacterCatalogue) { c.Source.Changes = "" }},
		{"source ownership", func(c *CharacterCatalogue) { c.Records[1].Sources = c.Records[4].Sources }},
		{"unknown source", func(c *CharacterCatalogue) { c.Records[0].Sources[0].InventoryID = "unknown" }},
		{"missing choice", func(c *CharacterCatalogue) {
			c.Records[0].ChoiceRefs = []string{backgroundID + "choice/acolyte/missing"}
		}},
		{"orphan", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"wrong owner", func(c *CharacterCatalogue) { c.Records[1].ParentID = c.Records[4].ID }},
		{"count", func(c *CharacterCatalogue) { c.Records[1].Select = 0 }},
		{"quantity", func(c *CharacterCatalogue) { c.Records[3].Options[0].Quantities[backgroundKey+"equipment/robe"] = 0 }},
		{"unknown quantity", func(c *CharacterCatalogue) { c.Records[3].Options[0].Quantities[backgroundKey+"equipment/unknown"] = 1 }},
		{"unknown option", func(c *CharacterCatalogue) { c.Records[1].Options[0].Key = backgroundKey + "unknown-option" }},
		{"malformed key", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics[0] = "bad key" }},
		{"predicate", func(c *CharacterCatalogue) { c.Records[2].Requirement.Key = backgroundKey + "unknown-predicate" }},
		{"operand", func(c *CharacterCatalogue) {
			c.Records[2].Requirement.RecordRefs = []string{backgroundID + "choice/missing"}
		}},
		{"negative limit", func(c *CharacterCatalogue) { c.Records[2].Requirement.Parameters["maximum-score"] = -1 }},
		{"cycle", func(c *CharacterCatalogue) {
			c.Records[2].ParentID = c.Records[1].ID
			c.Records[1].ParentID = c.Records[2].ID
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := backgroundsCatalogue(t)
			tc.mutate(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("invalid accepted")
			}
		})
	}
	for _, key := range []string{backgroundKey + "backgrounds/unknown-feature", backgroundKey + "backgrounds/unknown-choice"} {
		c := backgroundsCatalogue(t)
		c.Records[0].RequiredMechanics = append([]string{key}, c.Records[0].RequiredMechanics...)
		if err := c.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(repo); err == nil || !strings.Contains(err.Error(), key) {
			t.Fatalf("unknown readiness: %v", err)
		}
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"available", "implemented", "playable"} {
		input := strings.Replace(string(b), `"capacity":0`, `"capacity":0,"`+field+`":true`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("support flag accepted")
		}
	}
}

// Validate the actual background batch between fixed-runtime snapshots. Existing
// creation tests independently assert AC, skills, feats, resources and currency.
func TestBackgroundsLeaveFixedRuntimeUnchanged(t *testing.T) {
	repo := os.DirFS("../../..")
	c := backgroundsCatalogue(t)
	coverageBefore, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		concept, background string
		gold                int
	}{{"Human Fighter", "Soldier", 29}, {"Human Cleric", "Acolyte", 118}} {
		before, err := SupportedBuildForConcept(tc.concept, tc.background)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(repo); err == nil {
			t.Fatal("backgrounds ready")
		}
		after, err := SupportedBuildForConcept(tc.concept, tc.background)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) || after.Gold != tc.gold {
			t.Fatal("fixed creation changed")
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
	if coverageBefore.Coverage() != coverageAfter.Coverage() || coverageAfter.FullSupportEligible(repo) {
		t.Fatal("coverage promoted")
	}
}
