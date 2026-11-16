package srd521

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func speciesCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	f, err := os.Open("testdata/character-species.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := DecodeCharacterCatalogue(f)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestSpeciesExactSourceFacts(t *testing.T) {
	c := speciesCatalogue(t)
	names := []string{"Dragonborn", "Dwarf", "Elf", "Gnome", "Goliath", "Halfling", "Human", "Orc", "Tiefling"}
	pages := []int{84, 84, 84, 85, 85, 86, 86, 86, 86}
	i := 0
	for _, r := range c.Records {
		if r.Kind == "origin" {
			if i >= 9 || r.Name != names[i] || r.Sources[0].Page != pages[i] || r.ID != "srd-5.2.1:character/origin/"+strings.ToLower(names[i]) {
				t.Fatalf("identity %s", r.ID)
			}
			i++
		} else if r.Kind != "choice" && r.Kind != "prerequisite" && r.Kind != "resource" {
			t.Fatalf("adjacent definition %s", r.ID)
		}
	}
	if i != 9 {
		t.Fatal("nine identities required")
	}
	// The reviewed source transcript asserts every trait key, condition operand,
	// numerical row, grant, recovery key, option set and owner, not just references.
	var facts strings.Builder
	for _, r := range c.Records {
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		facts.Write(b)
		facts.WriteByte(10)
	}
	want, err := os.ReadFile("testdata/species-source-facts.txt")
	if err != nil {
		t.Fatal(err)
	}
	if facts.String() != string(want) {
		t.Fatal("species differs from reviewed exact source transcript; inspect source before updating")
	}
	// Compose canonical batches in memory; there is no copied subset manifest.
	backgrounds := backgroundsCatalogue(t)
	composed := backgrounds
	composed.Records = append(append([]CharacterRecord{}, backgrounds.Records...), c.Records...)
	if err := composed.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(composed.Records[:len(backgrounds.Records)], backgrounds.Records) {
		t.Fatal("composition changed background records")
	}
}

func TestSpeciesMetadataGuards(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"symbolic resource", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].CapacityMechanic != "" {
					c.Records[i].Capacity = 1
					return
				}
			}
		}},
		{"unknown capacity expression", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].CapacityMechanic != "" {
					c.Records[i].CapacityMechanic = "srd-5.2.1:mechanic/unknown"
					return
				}
			}
		}},
		{"unrequired capacity expression", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].CapacityMechanic != "" {
					r := &c.Records[i]
					r.RequiredMechanics = r.RequiredMechanics[:len(r.RequiredMechanics)-1]
					return
				}
			}
		}},
		{"trait on choice", func(c *CharacterCatalogue) { c.Records[1].Species = c.Records[0].Species }},
		{"source pin", func(c *CharacterCatalogue) { c.Source.SHA256 = "invalid" }},
		{"source page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 85 }},
		{"source identity", func(c *CharacterCatalogue) { c.Records[0].Sources[0].InventoryID = "srd-5.2.1:planning/species/dwarf" }},
		{"category", func(c *CharacterCatalogue) { c.Records[0].Species.Category = "background" }},
		{"creature type", func(c *CharacterCatalogue) { c.Records[0].Species.CreatureType = "dragon" }},
		{"speed", func(c *CharacterCatalogue) { c.Records[0].Species.SpeedFeet = 0 }},
		{"size", func(c *CharacterCatalogue) { c.Records[0].Species.Sizes[0].Key = "large" }},
		{"missing trait key", func(c *CharacterCatalogue) { c.Records[0].Species.Traits[0].Key = "srd-5.2.1:mechanic/unknown" }},
		{"unrequired number", func(c *CharacterCatalogue) {
			c.Records[0].Species.Traits[0].Parameters["srd-5.2.1:mechanic/unknown"] = 3
		}},
		{"negative number", func(c *CharacterCatalogue) {
			c.Records[0].Species.Traits[0].Parameters["srd-5.2.1:mechanic/species/dragonborn/darkvision/range-feet"] = -1
		}},
		{"bad condition", func(c *CharacterCatalogue) {
			c.Records[0].Species.Traits[1].PrerequisiteRefs = []string{c.Records[0].ID}
		}},
		{"bad resource", func(c *CharacterCatalogue) { c.Records[0].Species.Traits[0].ResourceRef = c.Records[0].ID }},
		{"zero selection", func(c *CharacterCatalogue) { c.Records[1].Select = 0 }},
		{"missing parent link", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"predicate option", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if q := c.Records[i].Requirement; q != nil && q.OptionKey != "" {
					q.OptionKey = "srd-5.2.1:mechanic/dragon/missing"
					return
				}
			}
		}},
		{"predicate operands", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if q := c.Records[i].Requirement; q != nil {
					q.RecordRefs = []string{c.Records[0].ID}
					return
				}
			}
		}},
		{"predicate parameters", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if q := c.Records[i].Requirement; q != nil {
					q.Parameters["unexpected"] = 1
					return
				}
			}
		}},
		{"character level", func(c *CharacterCatalogue) {
			for i := range c.Records[0].Species.Traits {
				q := &c.Records[0].Species.Traits[i]
				if len(q.CharacterLevels) > 0 {
					q.CharacterLevels[0].Level = 21
					return
				}
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := speciesCatalogue(t)
			tt.mutate(&c)
			if c.Validate(os.DirFS("../../..")) == nil {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
	c := speciesCatalogue(t)
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"playable", "implemented", "available", "class_levels"} {
		input := strings.Replace(string(b), `"category":"species"`, `"`+field+`":true,"category":"species"`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("unknown field accepted")
		}
	}
	c.Records[0].RequiredMechanics = append([]string{"srd-5.2.1:mechanic/unknown-species-effect"}, c.Records[0].RequiredMechanics...)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), "unknown-species-effect") {
		t.Fatalf("readiness: %v", err)
	}
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, build := range []struct{ concept, background string }{{"Human Fighter", "Soldier"}, {"Human Cleric", "Acolyte"}} {
		if _, err := SupportedBuildForConcept(build.concept, build.background); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"Elf Fighter", "Dwarf Cleric", "Tiefling Fighter"} {
		if _, err := SupportedBuildForConcept(name, "Soldier"); err == nil {
			t.Fatal("species creation enabled")
		}
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if before.Coverage() != after.Coverage() || after.FullSupportEligible(os.DirFS("../../..")) {
		t.Fatal("metadata promoted coverage")
	}
}

func TestSpeciesRejectNullNumericOperands(t *testing.T) {
	b, err := os.ReadFile("testdata/character-species.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}
	for _, value := range root["records"].([]any) {
		record := value.(map[string]any)
		species, ok := record["species"].(map[string]any)
		if !ok {
			continue
		}
		for _, value := range species["traits"].([]any) {
			trait := value.(map[string]any)
			maps := []map[string]any{trait["parameters"].(map[string]any)}
			for _, value := range trait["character_levels"].([]any) {
				maps = append(maps, value.(map[string]any)["parameters"].(map[string]any))
			}
			for _, numbers := range maps {
				for key, original := range numbers {
					numbers[key] = nil
					changed, err := json.Marshal(root)
					numbers[key] = original
					if err != nil {
						t.Fatal(err)
					}
					if _, err := DecodeCharacterCatalogue(strings.NewReader(string(changed))); err == nil {
						t.Fatalf("accepted null numeric operand %s in %s", key, record["id"])
					}
				}
			}
		}
	}
}
