package srd521

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func characterFixture(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference struct {
		Catalogue string   `json:"catalogue"`
		RecordIDs []string `json:"record_ids"`
	}
	if err := json.Unmarshal(b, &reference); err != nil {
		t.Fatal(err)
	}
	if reference.Catalogue != "character-fighter.json" {
		t.Fatal("unexpected canonical fixture")
	}
	c := fighterCatalogue(t)
	all := c.Records
	c.Records = nil
	for _, id := range reference.RecordIDs {
		found := false
		for _, r := range all {
			if r.ID == id {
				c.Records = append(c.Records, r)
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing fixture reference %s", id)
		}
	}
	return c
}

func TestCharacterCatalogueFighterOne(t *testing.T) {
	c := characterFixture(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if len(c.Records) != 11 {
		t.Fatalf("fixture widened: %d records", len(c.Records))
	}
	levels, features := 0, 0
	pages := map[string]int{"Fighting Style": 47, "Second Wind": 48, "Weapon Mastery": 48}
	for _, r := range c.Records {
		if r.Kind == "level" {
			levels++
			if r.Level != 1 {
				t.Fatal("not level one")
			}
		}
		if r.Kind == "feature" {
			features++
			if pages[r.Name] != r.Sources[0].Page {
				t.Fatalf("wrong feature/page: %+v", r)
			}
		}
		if r.Kind == "resource" && r.Capacity != 2 {
			t.Fatal("level-one Second Wind capacity must be 2")
		}
	}
	if levels != 1 || features != 3 {
		t.Fatalf("levels=%d features=%d", levels, features)
	}
	if err := c.ValidateReadiness(repo); err == nil || !strings.Contains(err.Error(), "unimplemented required key") {
		t.Fatalf("data must stay unavailable: %v", err)
	}
}

func TestCharacterCatalogueRejectsInvalidMetadata(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"schema", func(c *CharacterCatalogue) { c.SchemaVersion = 2 }},
		{"duplicate", func(c *CharacterCatalogue) { c.Records = append(c.Records, c.Records[0]) }},
		{"duplicate class level", func(c *CharacterCatalogue) {
			r := c.Records[1]
			r.ID = "srd-5.2.1:character/level/fighter/one-again"
			c.Records = append(c.Records, r)
		}},
		{"other class feature source", func(c *CharacterCatalogue) {
			c.Records[2].Sources = []CharacterSource{{InventoryID: "srd-5.2.1:planning/class-features/cleric-level-1-divine-order", Page: 37}}
		}},
		{"other level feature source", func(c *CharacterCatalogue) {
			c.Records[2].Sources = []CharacterSource{{InventoryID: "srd-5.2.1:planning/class-features/fighter-level-2-action-surge", Page: 48}}
		}},
		{"nested choice source owner", func(c *CharacterCatalogue) { c.Records[3].Sources = c.Records[4].Sources }},
		{"nested resource source owner", func(c *CharacterCatalogue) { c.Records[5].Sources = c.Records[2].Sources }},
		{"version", func(c *CharacterCatalogue) { c.Source.Version = "5.1" }},
		{"pin", func(c *CharacterCatalogue) { c.Source.SHA256 = "wrong" }},
		{"license", func(c *CharacterCatalogue) { c.Source.License = "MIT" }},
		{"attribution", func(c *CharacterCatalogue) { c.Source.Attribution = "" }},
		{"adaptation", func(c *CharacterCatalogue) { c.Source.Changes = "" }},
		{"zero level", func(c *CharacterCatalogue) { c.Records[1].Level = 0 }},
		{"negative level", func(c *CharacterCatalogue) { c.Records[1].Level = -1 }},
		{"level 21", func(c *CharacterCatalogue) { c.Records[1].Level = 21 }},
		{"level on class", func(c *CharacterCatalogue) { c.Records[0].Level = 1 }},
		{"missing source", func(c *CharacterCatalogue) { c.Records[0].Sources = nil }},
		{"unknown source", func(c *CharacterCatalogue) { c.Records[0].Sources[0].InventoryID = "invented" }},
		{"wrong page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 48 }},
		{"source family", func(c *CharacterCatalogue) {
			c.Records[0].Kind = "feat"
			c.Records[0].ID = "srd-5.2.1:character/feat/fighter"
		}},
		{"missing parent", func(c *CharacterCatalogue) { c.Records[1].ParentID = "" }},
		{"wrong parent", func(c *CharacterCatalogue) { c.Records[1].ParentID = c.Records[2].ID }},
		{"missing choice", func(c *CharacterCatalogue) { c.Records[2].ChoiceRefs = []string{"missing"} }},
		{"duplicate choice", func(c *CharacterCatalogue) {
			c.Records[2].ChoiceRefs = append(c.Records[2].ChoiceRefs, c.Records[2].ChoiceRefs[0])
		}},
		{"wrong choice kind", func(c *CharacterCatalogue) { c.Records[2].ChoiceRefs = []string{c.Records[5].ID} }},
		{"orphan choice", func(c *CharacterCatalogue) { c.Records[2].ChoiceRefs = []string{} }},
		{"choice count", func(c *CharacterCatalogue) { c.Records[3].Select = 0 }},
		{"missing options", func(c *CharacterCatalogue) { c.Records[3].OptionKeys = nil }},
		{"resource capacity", func(c *CharacterCatalogue) { c.Records[5].Capacity = 0 }},
		{"missing resource", func(c *CharacterCatalogue) { c.Records[4].ResourceRefs = []string{"missing"} }},
		{"missing prerequisite", func(c *CharacterCatalogue) { c.Records[0].PrerequisiteRefs = []string{"missing"} }},
		{"missing mechanic", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics = nil }},
		{"invalid mechanic", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics = []string{"anything"} }},
		{"duplicate mechanic", func(c *CharacterCatalogue) {
			c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, c.Records[0].RequiredMechanics[0])
		}},
		{"implicit refs", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := characterFixture(t)
			tt.mutate(&c)
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("accepted invalid metadata")
			}
		})
	}
	if err := characterFixture(t).Validate(nil); err == nil {
		t.Fatal("accepted absent source inventory")
	}
}

func TestCharacterCatalogueKeysStayUnavailable(t *testing.T) {
	c := characterFixture(t)
	for _, r := range c.Records {
		for _, key := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(key) {
				t.Errorf("metadata qualified %s", key)
			}
		}
	}
	for _, key := range []string{"", "srd-5.2.1:mechanic/unknown-feature", "srd-5.2.1:mechanic/unknown-choice"} {
		if qualifiedCharacterMechanic(key) {
			t.Fatalf("unknown key available: %q", key)
		}
	}
	c.Records[0].RequiredMechanics = append([]string{"srd-5.2.1:mechanic/unknown-feature"}, c.Records[0].RequiredMechanics...)
	if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), "unknown-feature") {
		t.Fatalf("unexpected readiness: %v", err)
	}
}

func TestCharacterCatalogueStrictDecode(t *testing.T) {
	c := characterFixture(t)
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{string(b) + " {}", strings.Replace(string(b), `"schema_version":1`, `"schema_version":1,"playable":true`, 1), strings.Replace(string(b), `"capacity":0`, `"capacity":0,"implemented":true`, 1)} {
		if _, err := DecodeCharacterCatalogue(strings.NewReader(text)); err == nil {
			t.Fatal("accepted extra JSON or data-supplied availability")
		}
	}
}

func TestCharacterCatalogueDoesNotPromoteFixedBuilds(t *testing.T) {
	// Metadata validation must leave existing fixed entry points and coverage alone.
	c := characterFixture(t)
	repo := os.DirFS("../../..")
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	for _, build := range []struct{ concept, background string }{{"Human Fighter", "Soldier"}, {"Human Cleric", "Acolyte"}} {
		if _, err := SupportedBuildForConcept(build.concept, build.background); err != nil {
			t.Fatal(err)
		}
	}
	for _, build := range []struct{ concept, background string }{{"Elf Fighter", "Soldier"}, {"Human Fighter", "Sage"}, {"Human Wizard", "Acolyte"}} {
		if _, err := SupportedBuildForConcept(build.concept, build.background); err == nil {
			t.Fatal("metadata widened creation")
		}
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if before.Coverage() != after.Coverage() || after.FullSupportEligible(repo) {
		t.Fatal("metadata promoted coverage")
	}
}

func TestCharacterPrerequisiteContract(t *testing.T) {
	c := characterFixture(t)
	// Synthetic schema case reuses the level-one feature's source. It is not
	// another imported catalogue or a claim of complete prerequisite coverage.
	p := c.Records[2]
	p.ID = "srd-5.2.1:character/prerequisite/fighter/level-one"
	p.Kind = "prerequisite"
	p.ParentID = c.Records[2].ID
	p.ChoiceRefs = []string{}
	p.RequiredMechanics = []string{"srd-5.2.1:mechanic/prerequisite/class-level"}
	p.Requirement = &CharacterRequirement{Key: p.RequiredMechanics[0], RecordRefs: []string{c.Records[0].ID}, Parameters: map[string]int{"minimum-level": 1}}
	c.Records[2].PrerequisiteRefs = []string{p.ID}
	c.Records = append(c.Records, p)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*CharacterRecord){
		func(r *CharacterRecord) { r.Requirement = nil },
		func(r *CharacterRecord) { r.Requirement.Key = "srd-5.2.1:mechanic/not-required" },
		func(r *CharacterRecord) { r.Requirement.RecordRefs = []string{"missing"} },
		func(r *CharacterRecord) { r.Requirement.Parameters["minimum-level"] = -1 },
	} {
		b, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		copy, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		mutate(&copy.Records[len(copy.Records)-1])
		if err := copy.Validate(repo); err == nil {
			t.Fatal("accepted invalid prerequisite")
		}
	}
}
