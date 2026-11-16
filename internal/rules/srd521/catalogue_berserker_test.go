package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const berserkerID = "srd-5.2.1:character/"
const berserkerKey = "srd-5.2.1:mechanic/barbarian/berserker/"

func berserkerCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-berserker-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-barbarian.json", "character-berserker.json"}) {
		t.Fatal("unexpected composition")
	}
	var composed, batch CharacterCatalogue
	for i, name := range manifest.Catalogues {
		b, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateSource(c.Source); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			composed = c
		} else {
			batch = c
			composed.Records = append(composed.Records, c.Records...)
		}
	}
	return batch, composed
}

// Independently transcribed from p30; compare the entire batch so extra or
// missing conditions, resources, numbers and source identities fail closed.
func berserkerSourceError(batch CharacterCatalogue) error {
	levels := []int{3, 6, 10, 14}
	slugs := []string{"frenzy", "mindless-rage", "retaliation", "intimidating-presence"}
	names := []string{"Frenzy", "Mindless Rage", "Retaliation", "Intimidating Presence"}
	requirements := [][]string{
		{"requires-reckless-attack", "requires-active-rage", "first-target-hit-on-own-turn", "strength-based-attack", "extra-damage-dice-count-equals-rage-damage-bonus", "extra-damage-sum-dice", "damage-type-matches-weapon-or-unarmed-strike"},
		{"while-rage-active/immunity-charmed", "while-rage-active/immunity-frightened", "enter-rage/ends-charmed", "enter-rage/ends-frightened"},
		{"trigger-take-damage-from-creature", "damaging-creature-within-range", "optional-reaction", "melee-attack-against-damaging-creature", "weapon-or-unarmed-strike"},
		{"bonus-action", "each-chosen-creature", "emanation-originates-from-self", "wisdom-saving-throw", "save-dc-plus-strength-modifier", "save-dc-plus-proficiency-bonus", "failed-save/frightened", "repeat-save-end-of-each-affected-creature-turn", "successful-repeat-save/ends-effect", "restore-use/expend-rage-use", "restore-use/no-action-required"},
	}
	numbers := []map[string]int{{"damage-die-sides": 6}, nil, {"range-feet": 5, "attack-count": 1}, {"emanation-feet": 30, "save-dc-base": 8, "duration-minutes": 1, "restore-use/rage-cost": 1}}
	records := map[string]CharacterRecord{}
	for _, r := range batch.Records {
		records[r.ID] = r
	}
	if len(records) != 11 || len(batch.Records) != 11 {
		return fmt.Errorf("record count")
	}
	sub := records[berserkerID+"subclass/barbarian/berserker"]
	if sub.Name != "Path of the Berserker" || sub.ParentID != berserkerID+"class/barbarian" || !reflect.DeepEqual(sub.RequiredMechanics, []string{"srd-5.2.1:mechanic/barbarian/subclasses/path-of-the-berserker"}) {
		return fmt.Errorf("subclass")
	}
	check := func(r CharacterRecord, source string, level int) error {
		if !reflect.DeepEqual(r.Sources, []CharacterSource{{InventoryID: "srd-5.2.1:planning/" + source, Page: 30}}) || len(r.PrerequisiteRefs) != 1 {
			return fmt.Errorf("source/prerequisite %s", r.ID)
		}
		p := records[r.PrerequisiteRefs[0]]
		expectedSlug := strings.TrimPrefix(r.ID, berserkerID+r.Kind+"/barbarian/")
		if r.Kind == "subclass" {
			expectedSlug = "berserker/path-of-the-berserker"
		}
		want := &CharacterRequirement{Key: "srd-5.2.1:mechanic/prerequisite/class-level", RecordRefs: []string{berserkerID + "class/barbarian"}, Parameters: map[string]int{"minimum-level": level}}
		if p.ID != berserkerID+"prerequisite/barbarian/"+expectedSlug+"-level" || p.ParentID != r.ID || !reflect.DeepEqual(p.Sources, r.Sources) || !reflect.DeepEqual(p.Requirement, want) || !reflect.DeepEqual(p.RequiredMechanics, []string{want.Key}) {
			return fmt.Errorf("level %s", r.ID)
		}
		return nil
	}
	if err := check(sub, "subclasses/barbarian-path-of-the-berserker", 3); err != nil {
		return err
	}
	for i, slug := range slugs {
		r := records[berserkerID+"feature/barbarian/berserker/"+slug]
		if r.Name != names[i] || r.ParentID != sub.ID || len(r.ChoiceRefs) != 0 {
			return fmt.Errorf("identity %s", slug)
		}
		if err := check(r, fmt.Sprintf("subclass-features/barbarian-level-%d-%s", levels[i], slug), levels[i]); err != nil {
			return err
		}
		wantKeys := map[string]bool{}
		wantNums := map[string]int{}
		for _, k := range requirements[i] {
			wantKeys[berserkerKey+slug+"/"+k] = true
		}
		for k, v := range numbers[i] {
			key := berserkerKey + slug + "/" + k
			wantKeys[key] = true
			wantNums[key] = v
		}
		got := map[string]bool{}
		for _, k := range r.RequiredMechanics {
			got[k] = true
		}
		if !reflect.DeepEqual(wantKeys, got) || len(r.RequiredMechanics) != len(wantKeys) || len(r.Parameters) != len(wantNums) {
			return fmt.Errorf("requirements %s", slug)
		}
		for k, v := range wantNums {
			if r.Parameters[k] != v {
				return fmt.Errorf("number %s", k)
			}
		}
		if slug != "intimidating-presence" && len(r.ResourceRefs) != 0 {
			return fmt.Errorf("extra resource")
		}
	}
	feature := records[berserkerID+"feature/barbarian/berserker/intimidating-presence"]
	resource := records[berserkerID+"resource/barbarian/berserker/intimidating-presence"]
	if resource.Capacity != 1 || len(resource.ValuesByLevel) != 0 || resource.ParentID != feature.ID || !reflect.DeepEqual(resource.Sources, feature.Sources) || !reflect.DeepEqual(feature.ResourceRefs, []string{resource.ID}) || !reflect.DeepEqual(resource.RequiredMechanics, []string{berserkerKey + "intimidating-presence/use/long-rest-reset", berserkerKey + "intimidating-presence/use/restored-by-expending-rage-no-action"}) {
		return fmt.Errorf("use/restoration")
	}
	return nil
}

func TestBerserkerExactSourceAndComposition(t *testing.T) {
	batch, c := berserkerCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := berserkerSourceError(batch); err != nil {
		t.Fatal(err)
	}
	if err := batch.Validate(repo); err == nil {
		t.Fatal("canonical parent required")
	}
	canonical := barbarianCatalogue(t)
	if !reflect.DeepEqual(c.Records[:len(canonical.Records)], canonical.Records) {
		t.Fatal("canonical Barbarian changed")
	}
	// All 20 rows and Rage Damage improvements remain canonical references.
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("metadata promoted")
	}
	for _, r := range batch.Records {
		for _, k := range r.RequiredMechanics {
			if qualifiedCharacterMechanic(k) {
				t.Fatal(k)
			}
		}
	}
	if batch.Source.SHA256 != canonical.Source.SHA256 || batch.Source.Attribution != canonical.Source.Attribution || batch.Source.License != "CC-BY-4.0" || batch.Source.Changes == "" {
		t.Fatal("provenance")
	}
}

func TestBerserkerContrarySource(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"wrong page", func(c *CharacterCatalogue) { c.Records[2].Sources[0].Page = 29 }},
		{"wrong level", func(c *CharacterCatalogue) { c.Records[3].Requirement.Parameters["minimum-level"] = 4 }},
		{"legacy exhaustion frenzy", func(c *CharacterCatalogue) {
			c.Records[2].RequiredMechanics = append(c.Records[2].RequiredMechanics, berserkerKey+"frenzy/exhaustion")
		}},
		{"fixed frenzy dice", func(c *CharacterCatalogue) {
			c.Records[2].RequiredMechanics[4] = berserkerKey + "frenzy/fixed-dice-count"
		}},
		{"missing condition termination", func(c *CharacterCatalogue) { c.Records[4].RequiredMechanics = c.Records[4].RequiredMechanics[:2] }},
		{"retaliation range", func(c *CharacterCatalogue) { c.Records[6].Parameters[berserkerKey+"retaliation/range-feet"] = 10 }},
		{"charisma dc", func(c *CharacterCatalogue) {
			c.Records[8].RequiredMechanics[4] = berserkerKey + "intimidating-presence/save-dc-plus-charisma-modifier"
		}},
		{"missing repeated save", func(c *CharacterCatalogue) {
			c.Records[8].RequiredMechanics[7] = berserkerKey + "intimidating-presence/no-repeat-save"
		}},
		{"wrong duration", func(c *CharacterCatalogue) {
			c.Records[8].Parameters[berserkerKey+"intimidating-presence/duration-minutes"] = 10
		}},
		{"missing rage restoration", func(c *CharacterCatalogue) { c.Records[10].RequiredMechanics = c.Records[10].RequiredMechanics[:1] }},
		{"short rest recharge", func(c *CharacterCatalogue) {
			c.Records[10].RequiredMechanics[0] = berserkerKey + "intimidating-presence/use/short-rest-reset"
		}},
		{"resource scaling", func(c *CharacterCatalogue) {
			c.Records[10].ValuesByLevel = []CharacterLevelValue{{Level: 14, Value: 1}, {Level: 15, Value: 2}}
		}},
		{"prerequisite mechanic", func(c *CharacterCatalogue) {
			c.Records[3].RequiredMechanics = append(c.Records[3].RequiredMechanics, berserkerKey+"frenzy/requires-heavy-armor")
		}},
		{"extra identity", func(c *CharacterCatalogue) { c.Records = append(c.Records, c.Records[2]) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch, _ := berserkerCatalogue(t)
			tt.mutate(&batch)
			if err := berserkerSourceError(batch); err == nil {
				t.Fatal("contrary source accepted")
			}
		})
	}
}

func TestBerserkerUnavailableAndSchemaGuards(t *testing.T) {
	_, c := berserkerCatalogue(t)
	repo := os.DirFS("../../..")
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	key := berserkerKey + "unknown-key"
	c.Records[0].RequiredMechanics = append([]string{key}, c.Records[0].RequiredMechanics...)
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil || !strings.Contains(err.Error(), key) {
		t.Fatalf("unknown key: %v", err)
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"available", "implemented", "playable"} {
		input := strings.Replace(string(b), `"capacity":1`, `"capacity":1,"`+field+`":true`, 1)
		if input == string(b) {
			t.Fatal("mutation missed")
		}
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("availability accepted")
		}
	}
	for _, concept := range []string{"Human Berserker", "Human Barbarian Path of the Berserker"} {
		if _, err := SupportedBuildForConcept(concept, "Soldier"); err == nil {
			t.Fatal("creation promoted")
		}
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if before.Coverage() != after.Coverage() || after.FullSupportEligible(repo) {
		t.Fatal("coverage promoted")
	}
}

func TestBerserkerGraphGuards(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue, int)
	}{
		{"duplicate", func(c *CharacterCatalogue, i int) { c.Records = append(c.Records, c.Records[i]) }},
		{"wrong page", func(c *CharacterCatalogue, i int) { c.Records[i+2].Sources[0].Page = 29 }},
		{"wrong class", func(c *CharacterCatalogue, i int) { c.Records[i].ParentID = berserkerID + "class/fighter" }},
		{"missing resource link", func(c *CharacterCatalogue, i int) { c.Records[i+8].ResourceRefs = []string{} }},
		{"wrong nested source", func(c *CharacterCatalogue, i int) { c.Records[i+10].Sources = c.Records[i+2].Sources }},
		{"missing level operand", func(c *CharacterCatalogue, i int) {
			c.Records[i+3].Requirement.RecordRefs = []string{berserkerID + "class/missing"}
		}},
		{"unrequired number", func(c *CharacterCatalogue, i int) { c.Records[i+2].Parameters[berserkerKey+"unknown-number"] = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch, c := berserkerCatalogue(t)
			tt.mutate(&c, len(c.Records)-len(batch.Records))
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("invalid graph accepted")
			}
		})
	}
}
