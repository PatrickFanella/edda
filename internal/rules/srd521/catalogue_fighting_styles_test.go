package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const styleID = "srd-5.2.1:character/"
const styleKey = "srd-5.2.1:mechanic/"

func fightingStylesCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-fighting-styles.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// Independent transcription of every record, numeric value and condition.
// Structural validation deliberately permits future opaque mechanics; this
// source oracle rejects well-formed metadata that contradicts this batch.
func checkFightingStylesSource(c CharacterCatalogue) error {
	cases := []struct {
		slug, name string
		page       int
		conditions []string
		numbers    map[string]int
	}{
		{"archery", "Archery", 87, []string{"attack-rolls/with-ranged-weapons"}, map[string]int{"attack-roll-bonus": 2}},
		{"defense", "Defense", 88, []string{"armor-class/while-wearing-light-or-medium-or-heavy-armor"}, map[string]int{"armor-class-bonus": 1}},
		{"great-weapon-fighting", "Great Weapon Fighting", 88, []string{"damage-die/optional-treat-result", "damage-die/when-rolling-damage-for-own-attack", "weapon/melee", "weapon/held-with-two-hands", "weapon/property-two-handed-or-versatile"}, map[string]int{"damage-die/minimum-result": 1, "damage-die/maximum-result": 2, "damage-die/treat-as": 3}},
		{"two-weapon-fighting", "Two-Weapon Fighting", 88, []string{"damage/optional-add-ability-modifier", "attack/extra-from-using-light-property-weapon", "damage/only-if-not-already-adding-ability-modifier"}, map[string]int{}},
	}
	if len(c.Records) != 8 {
		return fmt.Errorf("expected four feats and four prerequisites")
	}
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		if records[r.ID].ID != "" {
			return fmt.Errorf("duplicate ID")
		}
		records[r.ID] = r
	}
	for _, tc := range cases {
		rootID := styleID + "feat/" + tc.slug
		predicateID := styleID + "prerequisite/" + tc.slug + "/fighting-style-feature"
		source := []CharacterSource{{InventoryID: "srd-5.2.1:planning/feats/" + tc.slug, Page: tc.page}}
		keys := []string{styleKey + "feats/category/fighting-style", styleKey + "feats/maximum-times-taken"}
		for _, k := range tc.conditions {
			keys = append(keys, styleKey+"feats/"+tc.slug+"/"+k)
		}
		numbers := map[string]int{styleKey + "feats/maximum-times-taken": 1}
		for k, v := range tc.numbers {
			numbers[styleKey+"feats/"+tc.slug+"/"+k] = v
		}
		// Compare key sets independently of JSON ordering.
		for k := range numbers {
			if k != styleKey+"feats/maximum-times-taken" {
				keys = append(keys, k)
			}
		}
		want := CharacterRecord{ID: rootID, Kind: "feat", Name: tc.name, Sources: source,
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{predicateID}, OptionKeys: []string{}, Parameters: numbers}
		got := records[rootID]
		gotKeys := map[string]bool{}
		for _, k := range got.RequiredMechanics {
			if gotKeys[k] {
				return fmt.Errorf("duplicate key")
			}
			gotKeys[k] = true
		}
		wantKeys := map[string]bool{}
		for _, k := range keys {
			wantKeys[k] = true
		}
		got.RequiredMechanics = nil
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(gotKeys, wantKeys) {
			return fmt.Errorf("%s: source values/conditions", tc.slug)
		}
		key := styleKey + "prerequisite/has-fighting-style-feature"
		prerequisite := CharacterRecord{ID: predicateID, Kind: "prerequisite", Name: "Fighting Style feature required", Sources: source, ParentID: rootID,
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, OptionKeys: []string{}, RequiredMechanics: []string{key},
			Requirement: &CharacterRequirement{Key: key, RecordRefs: []string{}, Parameters: map[string]int{}}}
		if !reflect.DeepEqual(records[predicateID], prerequisite) {
			return fmt.Errorf("%s: general feature prerequisite/operands", tc.slug)
		}
	}
	return nil
}

func TestFightingStylesExactSource(t *testing.T) {
	c := fightingStylesCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if err := checkFightingStylesSource(c); err != nil {
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

func TestFightingStylesContrarySource(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"bonus", func(c *CharacterCatalogue) { c.Records[0].Parameters[styleKey+"feats/archery/attack-roll-bonus"] = 3 }},
		{"armor alternatives", func(c *CharacterCatalogue) {
			c.Records[2].RequiredMechanics[2] = styleKey + "feats/defense/any-armor-or-shield"
		}},
		{"2014 reroll", func(c *CharacterCatalogue) {
			c.Records[4].RequiredMechanics[2] = styleKey + "feats/great-weapon-fighting/reroll-ones-and-twos"
		}},
		{"die treatment", func(c *CharacterCatalogue) {
			c.Records[4].Parameters[styleKey+"feats/great-weapon-fighting/damage-die/treat-as"] = 4
		}},
		{"lost two hands", func(c *CharacterCatalogue) {
			c.Records[4].RequiredMechanics = append(c.Records[4].RequiredMechanics[:5], c.Records[4].RequiredMechanics[6:]...)
		}},
		{"lost property alternative", func(c *CharacterCatalogue) {
			c.Records[4].RequiredMechanics[6] = styleKey + "feats/great-weapon-fighting/weapon/two-handed-only"
		}},
		{"double modifier", func(c *CharacterCatalogue) {
			c.Records[6].RequiredMechanics[4] = styleKey + "feats/two-weapon-fighting/always-add-modifier"
		}},
		{"Fighter-only predicate", func(c *CharacterCatalogue) {
			c.Records[1].RequiredMechanics = []string{styleKey + "prerequisite/fighter-fighting-style"}
			c.Records[1].Requirement.Key = c.Records[1].RequiredMechanics[0]
		}},
		{"repeatable", func(c *CharacterCatalogue) { c.Records[0].Parameters[styleKey+"feats/maximum-times-taken"] = 2 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := fightingStylesCatalogue(t)
			tc.mutate(&c)
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("expected structurally valid contrary source: %v", err)
			}
			if err := checkFightingStylesSource(c); err == nil {
				t.Fatal("contrary source accepted")
			}
		})
	}
}

func TestFightingStylesMetadataGuards(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"pin", func(c *CharacterCatalogue) { c.Source.SHA256 = "bad" }},
		{"attribution", func(c *CharacterCatalogue) { c.Source.Attribution = "" }},
		{"adaptation", func(c *CharacterCatalogue) { c.Source.Changes = "" }},
		{"page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 88 }},
		{"unknown source", func(c *CharacterCatalogue) { c.Records[0].Sources[0].InventoryID = "unknown" }},
		{"duplicate", func(c *CharacterCatalogue) { c.Records = append(c.Records, c.Records[0]) }},
		{"source owner", func(c *CharacterCatalogue) { c.Records[1].Sources = c.Records[2].Sources }},
		{"missing predicate", func(c *CharacterCatalogue) {
			c.Records[0].PrerequisiteRefs = []string{styleID + "prerequisite/missing"}
		}},
		{"orphan", func(c *CharacterCatalogue) { c.Records[0].PrerequisiteRefs = []string{} }},
		{"wrong owner", func(c *CharacterCatalogue) { c.Records[1].ParentID = c.Records[2].ID }},
		{"unknown operand", func(c *CharacterCatalogue) {
			c.Records[1].Requirement.RecordRefs = []string{styleID + "feature/missing"}
		}},
		{"unrequired predicate", func(c *CharacterCatalogue) { c.Records[1].Requirement.Key = styleKey + "unknown" }},
		{"negative predicate", func(c *CharacterCatalogue) { c.Records[1].Requirement.Parameters["minimum-level"] = -1 }},
		{"unrequired number", func(c *CharacterCatalogue) { c.Records[0].Parameters[styleKey+"unknown"] = 2 }},
		{"zero number", func(c *CharacterCatalogue) { c.Records[0].Parameters[styleKey+"feats/maximum-times-taken"] = 0 }},
		{"number on prerequisite", func(c *CharacterCatalogue) {
			c.Records[1].Parameters = map[string]int{styleKey + "prerequisite/has-fighting-style-feature": 1}
		}},
		{"malformed key", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics[0] = "bad" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := fightingStylesCatalogue(t)
			tc.mutate(&c)
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("invalid accepted")
			}
		})
	}
	for _, key := range []string{styleKey + "feats/unknown", styleKey + "prerequisite/unknown"} {
		c := fightingStylesCatalogue(t)
		c.Records[0].RequiredMechanics = append([]string{key}, c.Records[0].RequiredMechanics...)
		if err := c.Validate(os.DirFS("../../..")); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), key) {
			t.Fatalf("unknown readiness: %v", err)
		}
	}
	b, err := json.Marshal(fightingStylesCatalogue(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"available", "playable", "implemented", "category"} {
		input := strings.Replace(string(b), `"capacity":0`, `"capacity":0,"`+field+`":true`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("unknown field accepted")
		}
	}
}

func TestFightingStylesCompositionAndRuntimeBoundary(t *testing.T) {
	repo := os.DirFS("../../..")
	c := fightingStylesCatalogue(t)
	beforeCoverage, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	builds := []struct{ concept, background string }{{"Human Fighter", "Soldier"}, {"Human Cleric", "Acolyte"}}
	beforeBuilds := make([]any, 0, len(builds))
	for _, b := range builds {
		build, err := SupportedBuildForConcept(b.concept, b.background)
		if err != nil {
			t.Fatal(err)
		}
		beforeBuilds = append(beforeBuilds, build)
	}
	for _, name := range []string{"character-advancement.json", "character-fighter.json", "character-champion.json", "character-cleric.json", "character-backgrounds.json"} {
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
	for i, b := range builds {
		build, err := SupportedBuildForConcept(b.concept, b.background)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(beforeBuilds[i], build) {
			t.Fatal("fixed build changed")
		}
	}
	afterCoverage, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if beforeCoverage.Coverage() != afterCoverage.Coverage() || afterCoverage.FullSupportEligible(repo) {
		t.Fatal("coverage promoted")
	}
	// Fixed Defense still contributes +1 AC. The one-hand longsword still
	// uses raw 1 and 2 damage results, +5 attack and a single +3 modifier.
	stats, err := os.ReadFile("testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	sheet, err := Project(stats, 1, 12, 12, 0)
	if err != nil {
		t.Fatal(err)
	}
	if sheet.Defenses["armor-class"].Value != 19 {
		t.Fatal("fixed Defense changed")
	}
	for _, die := range []int{0, 1} {
		dice := &attackDice{values: []int{11, die}}
		receipt, err := ResolveLongsword(sheet, 10, dice)
		if err != nil {
			t.Fatal(err)
		}
		if receipt.Modifier.Value != 5 || receipt.Damage != die+4 || dice.calls != 2 || !reflect.DeepEqual(receipt.DamageDice, []int{die + 1}) {
			t.Fatalf("style leaked: %+v", receipt)
		}
	}
}
