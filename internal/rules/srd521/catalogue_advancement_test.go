package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func advancementCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-advancement.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCharacterAdvancementSourceTable(t *testing.T) {
	c := advancementCatalogue(t)
	repo := os.DirFS("../../..")
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	// Independent transcription of the three columns on pinned PDF page 23.
	xp := []int{0, 300, 900, 2700, 6500, 14000, 23000, 34000, 48000, 64000, 85000, 100000, 120000, 140000, 165000, 195000, 225000, 265000, 305000, 355000}
	proficiency := []int{2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 6, 6, 6, 6}
	if len(c.Records) != 20 {
		t.Fatalf("want exactly 20 global records, got %d", len(c.Records))
	}
	seen := map[int]bool{}
	for _, r := range c.Records {
		if seen[r.Level] || r.Level < 1 || r.Level > 20 {
			t.Fatalf("duplicate/invalid level %d", r.Level)
		}
		seen[r.Level] = true
		if r.Kind != "advancement" || r.ID != fmt.Sprintf("srd-5.2.1:character/advancement/%d", r.Level) || r.Name != fmt.Sprintf("Character level %d", r.Level) || r.ParentID != "" ||
			!reflect.DeepEqual(r.Sources, []CharacterSource{{InventoryID: fmt.Sprintf("srd-5.2.1:planning/levels/character-level-%d", r.Level), Page: 23}}) ||
			*r.Advancement.ExperiencePoints != xp[r.Level-1] || *r.Advancement.ProficiencyBonus != proficiency[r.Level-1] {
			t.Fatalf("source table mismatch: %+v", r)
		}
		if len(r.ChoiceRefs)+len(r.ResourceRefs)+len(r.PrerequisiteRefs)+len(r.OptionKeys) != 0 {
			t.Fatal("table has no nested choices or resources")
		}
		if err := (CharacterCatalogue{SchemaVersion: c.SchemaVersion, Source: c.Source, Records: []CharacterRecord{r}}).ValidateReadiness(repo); err == nil || !strings.Contains(err.Error(), "unimplemented required key") {
			t.Fatalf("row became ready: %v", err)
		}
	}
	// All twenty unique in-range rows imply completeness, irrespective of ordering.
	for level := 1; level <= 20; level++ {
		if !seen[level] {
			t.Fatalf("missing level %d", level)
		}
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) || after.FullSupportEligible(repo) {
		t.Fatal("catalogue changed runtime registration/coverage")
	}
	c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, "srd-5.2.1:mechanic/unknown-advancement")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if qualifiedCharacterMechanic("srd-5.2.1:mechanic/unknown-advancement") || c.ValidateReadiness(repo) == nil {
		t.Fatal("unknown mechanic became available")
	}
}

func TestCharacterAdvancementRejectsInvalidMetadata(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"missing numbers", func(c *CharacterCatalogue) { c.Records[0].Advancement = nil }},
		{"missing XP", func(c *CharacterCatalogue) { c.Records[0].Advancement.ExperiencePoints = nil }},
		{"missing bonus", func(c *CharacterCatalogue) { c.Records[0].Advancement.ProficiencyBonus = nil }},
		{"negative XP", func(c *CharacterCatalogue) { *c.Records[0].Advancement.ExperiencePoints = -1 }},
		{"nonzero first XP", func(c *CharacterCatalogue) { *c.Records[0].Advancement.ExperiencePoints = 1 }},
		{"zero later XP", func(c *CharacterCatalogue) { *c.Records[1].Advancement.ExperiencePoints = 0 }},
		{"low bonus", func(c *CharacterCatalogue) { *c.Records[0].Advancement.ProficiencyBonus = 1 }},
		{"high bonus", func(c *CharacterCatalogue) { *c.Records[0].Advancement.ProficiencyBonus = 7 }},
		{"zero level", func(c *CharacterCatalogue) { c.Records[0].Level = 0 }},
		{"level 21", func(c *CharacterCatalogue) { c.Records[0].Level = 21 }},
		{"wrong source level", func(c *CharacterCatalogue) { c.Records[0].Sources = c.Records[1].Sources }},
		{"wrong page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 24 }},
		{"wrong name", func(c *CharacterCatalogue) { c.Records[0].Name = "Fighter level 1" }},
		{"duplicate", func(c *CharacterCatalogue) { c.Records = append(c.Records, c.Records[0]) }},
		{"invented class owner", func(c *CharacterCatalogue) { c.Records[0].ParentID = "srd-5.2.1:character/class/fighter" }},
		{"missing XP mechanic", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics = c.Records[0].RequiredMechanics[1:] }},
		{"missing bonus mechanic", func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics = c.Records[0].RequiredMechanics[:1] }},
		{"implicit choice refs", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = nil }},
		{"unresolved choice", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{"missing"} }},
		{"class level without owner", func(c *CharacterCatalogue) {
			c.Records[0].Kind = "level"
			c.Records[0].ID = "srd-5.2.1:character/level/fighter/1"
			c.Records[0].Advancement = nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := advancementCatalogue(t)
			tc.mutate(&c)
			if c.Validate(os.DirFS("../../..")) == nil {
				t.Fatal("accepted invalid metadata")
			}
		})
	}
	c := characterFixture(t)
	c.Records[1].Advancement = advancementCatalogue(t).Records[0].Advancement
	if c.Validate(os.DirFS("../../..")) == nil {
		t.Fatal("accepted global numbers on class row")
	}
}

func TestCharacterAdvancementStrictNumbers(t *testing.T) {
	c := advancementCatalogue(t)
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []struct{ name, old string }{{"experience_points", "300"}, {"proficiency_bonus", "2"}} {
		for _, value := range []string{`"300"`, `300.5`, `true`, `9223372036854775808`} {
			text := strings.Replace(string(b), `"`+field.name+`":`+field.old, `"`+field.name+`":`+value, 1)
			if _, err := DecodeCharacterCatalogue(strings.NewReader(text)); err == nil {
				t.Fatalf("accepted noninteger %s %s", field.name, value)
			}
		}
	}
	for _, field := range []string{`"playable":true,`, `"implemented":true,`} {
		text := strings.Replace(string(b), `"experience_points":0`, field+`"experience_points":0`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(text)); err == nil {
			t.Fatal("data supplied availability")
		}
	}
}
