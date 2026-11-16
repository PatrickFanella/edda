package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const championID = "srd-5.2.1:character/"
const championKey = "srd-5.2.1:mechanic/fighter/champion/"

func championCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-champion-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-fighter.json", "character-champion.json"}) {
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

func TestChampionSourceAndImprovements(t *testing.T) {
	batch, c := championCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := batch.Validate(repo); err == nil {
		t.Fatal("unresolved canonical parent accepted")
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("metadata available")
	}
	if batch.SchemaVersion != 1 || batch.Source.SHA256 != c.Source.SHA256 || batch.Source.License != c.Source.License || batch.Source.Attribution != c.Source.Attribution || batch.Source.Changes == "" {
		t.Fatal("source provenance")
	}
	records := map[string]CharacterRecord{}
	identities := map[string]int{}
	for _, r := range batch.Records {
		records[r.ID] = r
		if len(r.Sources) != 1 || r.Sources[0].Page != 49 {
			t.Fatalf("source page: %s", r.ID)
		}
		if r.Kind == "feature" || r.Kind == "subclass" {
			identities[r.Sources[0].InventoryID]++
		}
		for _, key := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(key) {
				t.Fatalf("available: %s", key)
			}
		}
	}
	expected := map[string]int{"srd-5.2.1:planning/subclasses/fighter-champion": 1}
	levels := map[string]int{"improved-critical": 3, "remarkable-athlete": 3, "additional-fighting-style": 7, "heroic-warrior": 10, "superior-critical": 15, "survivor": 18}
	names := map[string]string{"improved-critical": "Improved Critical", "remarkable-athlete": "Remarkable Athlete", "additional-fighting-style": "Additional Fighting Style", "heroic-warrior": "Heroic Warrior", "superior-critical": "Superior Critical", "survivor": "Survivor"}
	for slug, n := range levels {
		expected[fmt.Sprintf("srd-5.2.1:planning/subclass-features/fighter-level-%d-%s", n, slug)] = 1
		r := records[championID+"feature/fighter/champion/"+slug]
		if r.Name != names[slug] || r.ParentID != championID+"subclass/fighter/champion" {
			t.Fatalf("feature identity %s", slug)
		}
		if len(r.PrerequisiteRefs) != 1 {
			t.Fatal("lost level")
		}
		p := records[r.PrerequisiteRefs[0]]
		want := &CharacterRequirement{Key: "srd-5.2.1:mechanic/prerequisite/class-level", RecordRefs: []string{championID + "class/fighter"}, Parameters: map[string]int{"minimum-level": n}}
		if p.ParentID != r.ID || !reflect.DeepEqual(p.Requirement, want) {
			t.Fatalf("level requirement: %s", slug)
		}
	}
	if !reflect.DeepEqual(identities, expected) || len(batch.Records) != 16 {
		t.Fatalf("batch widened: %v, %d", identities, len(batch.Records))
	}
	subclass := records[championID+"subclass/fighter/champion"]
	if subclass.Name != "Champion" || subclass.ParentID != championID+"class/fighter" || len(subclass.PrerequisiteRefs) != 1 || records[subclass.PrerequisiteRefs[0]].Requirement.Parameters["minimum-level"] != 3 {
		t.Fatal("subclass ownership/level")
	}
	// Independently transcribed numeric details, including both ends of each range.
	parameters := map[string]map[string]int{
		"improved-critical":         {"critical/d20-minimum": 19, "critical/d20-maximum": 20},
		"remarkable-athlete":        {"critical-movement/speed-divisor": 2},
		"additional-fighting-style": {}, "heroic-warrior": {},
		"superior-critical": {"critical/d20-minimum": 18, "critical/d20-maximum": 20},
		"survivor":          {"defy-death/d20-minimum": 18, "defy-death/d20-maximum": 20, "defy-death/benefit-of-roll": 20, "heroic-rally/healing-base": 5, "heroic-rally/minimum-current-hit-points": 1},
	}
	requirements := map[string][]string{
		"improved-critical":         {"critical/weapons-and-unarmed-strikes"},
		"remarkable-athlete":        {"advantage/initiative", "advantage/strength-athletics", "critical-movement/optional-immediately-after-critical-hit", "critical-movement/up-to-half-speed", "critical-movement/no-opportunity-attacks"},
		"additional-fighting-style": {"additional-fighting-style"},
		"heroic-warrior":            {"heroic-warrior/optional-self-heroic-inspiration", "heroic-warrior/during-combat", "heroic-warrior/start-of-own-turn", "heroic-warrior/only-without-heroic-inspiration"},
		"superior-critical":         {"critical/weapons-and-unarmed-strikes", "critical/replaces-improved-critical"},
		"survivor":                  {"defy-death/advantage-death-saving-throws", "defy-death/range-grants-benefit-of-20", "heroic-rally/start-of-each-own-turn", "heroic-rally/plus-constitution-modifier", "heroic-rally/only-while-bloodied", "heroic-rally/requires-at-least-one-hit-point"},
	}
	for slug, nums := range parameters {
		r := records[championID+"feature/fighter/champion/"+slug]
		if len(r.Parameters) != len(nums) {
			t.Fatalf("numeric details: %s", slug)
		}
		wantKeys := map[string]bool{}
		for key, value := range nums {
			if r.Parameters[championKey+key] != value {
				t.Fatalf("number %s %s", slug, key)
			}
			wantKeys[championKey+key] = true
		}
		for _, key := range requirements[slug] {
			wantKeys[championKey+key] = true
		}
		gotKeys := map[string]bool{}
		for _, key := range r.RequiredMechanics {
			gotKeys[key] = true
		}
		if !reflect.DeepEqual(gotKeys, wantKeys) {
			t.Fatalf("requirements %s: %v", slug, gotKeys)
		}
	}
	style := records[championID+"choice/fighter/champion/additional-fighting-style"]
	if style.Select != 1 || !reflect.DeepEqual(style.OptionKeys, []string{"srd-5.2.1:mechanic/feats/fighting-style"}) || !reflect.DeepEqual(style.RequiredMechanics, []string{championKey + "additional-fighting-style/another-distinct-feat"}) || !reflect.DeepEqual(records[style.ParentID].ChoiceRefs, []string{style.ID}) {
		t.Fatal("additional style choice")
	}
	resource := records[championID+"resource/fighter/champion/heroic-inspiration"]
	if resource.Capacity != 1 || !reflect.DeepEqual(records[resource.ParentID].ResourceRefs, []string{resource.ID}) || !reflect.DeepEqual(resource.RequiredMechanics, []string{championKey + "heroic-warrior/only-without-heroic-inspiration", championKey + "heroic-warrior/start-of-own-turn", championKey + "heroic-warrior/during-combat", championKey + "heroic-warrior/no-stacking"}) {
		t.Fatal("heroic resource")
	}
	// Composition retains the entire canonical Fighter byte-for-byte as decoded
	// records and never imports global XP or creates duplicate source identities.
	fighter := fighterCatalogue(t)
	if !reflect.DeepEqual(c.Records[:len(fighter.Records)], fighter.Records) {
		t.Fatal("canonical Fighter changed")
	}
	for _, r := range c.Records {
		if r.Kind == "advancement" {
			t.Fatal("global advancement coupled")
		}
	}
}

func TestChampionMetadataGuards(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue, int)
	}{
		{"duplicate identity", func(c *CharacterCatalogue, i int) { c.Records = append(c.Records, c.Records[i]) }},
		{"wrong page", func(c *CharacterCatalogue, i int) { c.Records[i].Sources[0].Page = 48 }},
		{"wrong parent", func(c *CharacterCatalogue, i int) {
			c.Records[i].ParentID = championID + "feature/fighter/fighting-style"
		}},
		{"missing nested choice", func(c *CharacterCatalogue, i int) { c.Records[i+6].ChoiceRefs = []string{} }},
		{"wrong nested source", func(c *CharacterCatalogue, i int) { c.Records[i+8].Sources = c.Records[i].Sources }},
		{"zero choice count", func(c *CharacterCatalogue, i int) { c.Records[i+8].Select = 0 }},
		{"zero resource cap", func(c *CharacterCatalogue, i int) { c.Records[i+11].Capacity = 0 }},
		{"unrequired numeric key", func(c *CharacterCatalogue, i int) { c.Records[i+2].Parameters[championKey+"unknown-number"] = 19 }},
		{"invalid level predicate", func(c *CharacterCatalogue, i int) { c.Records[i+3].Requirement.Parameters["minimum-level"] = -1 }},
		{"missing canonical reference", func(c *CharacterCatalogue, i int) {
			c.Records[i+3].Requirement.RecordRefs = []string{championID + "class/missing"}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch, c := championCatalogue(t)
			tt.mutate(&c, len(c.Records)-len(batch.Records))
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
	_, c := championCatalogue(t)
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// Put an unknown key first so readiness must reject that exact key.
	for _, key := range []string{championKey + "unknown-feature", championKey + "unknown-choice"} {
		c.Records[0].RequiredMechanics = append([]string{key}, c.Records[0].RequiredMechanics...)
		if err := c.Validate(os.DirFS("../../..")); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), key) {
			t.Fatalf("unknown key: %v", err)
		}
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"playable", "implemented", "available"} {
		input := strings.Replace(string(b), `"capacity":1`, `"capacity":1,"`+field+`":true`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("data supplied availability")
		}
	}
	// These established regression fixtures exercise actual fixed creation and
	// unchanged coverage; Champion is never registered or granted by validation.
	for _, build := range []struct{ concept, background string }{{"Human Fighter", "Soldier"}, {"Human Cleric", "Acolyte"}} {
		if _, err := SupportedBuildForConcept(build.concept, build.background); err != nil {
			t.Fatal(err)
		}
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if before.Coverage() != after.Coverage() || after.FullSupportEligible(os.DirFS("../../..")) {
		t.Fatal("metadata promoted coverage")
	}
	for _, concept := range []string{"Human Champion", "Human Fighter Champion"} {
		if _, err := SupportedBuildForConcept(concept, "Soldier"); err == nil {
			t.Fatal("Champion creation promoted")
		}
	}
}
