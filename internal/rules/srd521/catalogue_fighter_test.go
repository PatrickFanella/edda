package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func fighterCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-fighter.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestFighterCatalogueSourceAndProgression(t *testing.T) {
	c := fighterCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("catalogue became available")
	}
	byID := map[string]CharacterRecord{}
	identities := map[string]int{}
	canonical := 0
	levels := map[int]CharacterRecord{}
	for _, r := range c.Records {
		byID[r.ID] = r
		if r.Kind == "class" || r.Kind == "feature" {
			canonical++
			identities[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
		if r.Kind == "level" {
			levels[r.Level] = r
		}
		if r.Kind == "subclass" || r.Kind == "advancement" || strings.Contains(r.ID, "champion") {
			t.Fatalf("scope widened: %s", r.ID)
		}
		for _, key := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(key) {
				t.Fatalf("available: %s", key)
			}
		}
	}
	expected := map[string]int{"srd-5.2.1:planning/classes/fighter": 47}
	features := map[int][]string{1: {"fighting-style", "second-wind", "weapon-mastery"}, 2: {"action-surge", "tactical-mind"}, 3: {"fighter-subclass"}, 4: {"ability-score-improvement"}, 5: {"extra-attack", "tactical-shift"}, 9: {"indomitable", "tactical-master"}, 11: {"two-extra-attacks"}, 13: {"studied-attacks"}, 19: {"epic-boon"}, 20: {"three-extra-attacks"}}
	for n, slugs := range features {
		for _, slug := range slugs {
			page := 48
			if slug == "fighting-style" {
				page = 47
			}
			expected[fmt.Sprintf("srd-5.2.1:planning/class-features/fighter-level-%d-%s", n, slug)] = page
		}
	}
	if canonical != 16 || !reflect.DeepEqual(identities, expected) {
		t.Fatalf("identities/pages: %v", identities)
	}
	// Each tuple independently transcribes p47: proficiency, Second Wind,
	// mastery selections, Action Surge uses, Indomitable uses.
	rows := [20][5]int{{2, 2, 3, 0, 0}, {2, 2, 3, 1, 0}, {2, 2, 3, 1, 0}, {2, 3, 4, 1, 0}, {3, 3, 4, 1, 0}, {3, 3, 4, 1, 0}, {3, 3, 4, 1, 0}, {3, 3, 4, 1, 0}, {4, 3, 4, 1, 1}, {4, 4, 5, 1, 1}, {4, 4, 5, 1, 1}, {4, 4, 5, 1, 1}, {5, 4, 5, 1, 2}, {5, 4, 5, 1, 2}, {5, 4, 5, 1, 2}, {5, 4, 6, 1, 2}, {6, 4, 6, 2, 3}, {6, 4, 6, 2, 3}, {6, 4, 6, 2, 3}, {6, 4, 6, 2, 3}}
	if len(levels) != 20 {
		t.Fatal("missing class rows")
	}
	for i, row := range rows {
		n := i + 1
		r := levels[n]
		p := r.Progression
		if p == nil || p.ProficiencyBonus != row[0] || r.Sources[0].Page != 47 {
			t.Fatalf("row %d", n)
		}
		grants := append([]string{}, features[n]...)
		if n == 6 || n == 8 || n == 12 || n == 14 || n == 16 {
			grants = []string{"ability-score-improvement"}
		}
		if n == 13 {
			grants = []string{"indomitable", "studied-attacks"}
		}
		if n == 17 {
			grants = []string{"action-surge", "indomitable"}
		}
		want := []string{}
		for _, slug := range grants {
			want = append(want, "srd-5.2.1:character/feature/fighter/"+slug)
		}
		if !reflect.DeepEqual(p.FeatureRefs, want) {
			t.Fatalf("row %d grants %v != %v", n, p.FeatureRefs, want)
		}
		keys := []string{}
		if n == 7 || n == 10 || n == 15 || n == 18 {
			keys = []string{"srd-5.2.1:mechanic/fighter/subclass-feature"}
		}
		if !reflect.DeepEqual(p.FeatureKeys, keys) {
			t.Fatalf("subclass slots row %d", n)
		}
		for col, suffix := range []string{"resource/fighter/second-wind", "choice/fighter/weapon-mastery", "resource/fighter/action-surge", "resource/fighter/indomitable"} {
			got := 0
			for _, v := range byID["srd-5.2.1:character/"+suffix].ValuesByLevel {
				if v.Level == n {
					got = v.Value
				}
			}
			if got != row[col+1] {
				t.Fatalf("row %d %s=%d want %d", n, suffix, got, row[col+1])
			}
		}
	}
}

func TestFighterNestedChoices(t *testing.T) {
	c := fighterCatalogue(t)
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		records[r.ID] = r
	}
	choice := func(slug string, count int, options ...string) CharacterRecord {
		t.Helper()
		r := records["srd-5.2.1:character/choice/fighter/"+slug]
		want := []string{}
		for _, k := range options {
			want = append(want, "srd-5.2.1:mechanic/"+k)
		}
		if r.Select != count || !reflect.DeepEqual(r.OptionKeys, want) {
			t.Fatalf("choice %s: %+v", slug, r)
		}
		return r
	}
	choice("skills", 2, "skills/acrobatics", "skills/animal-handling", "skills/athletics", "skills/history", "skills/insight", "skills/intimidation", "skills/persuasion", "skills/perception", "skills/survival")
	choice("primary-ability", 1, "abilities/strength", "abilities/dexterity")
	choice("fighting-style", 1, "feats/fighting-style")
	choice("weapon-mastery", 3, "weapons/simple-or-martial/mastery")
	choice("fighter-subclass", 1, "fighter/subclasses", "fighter/subclasses/champion")
	choice("tactical-master", 1, "weapon-mastery/push", "weapon-mastery/sap", "weapon-mastery/slow")
	for _, slug := range []string{"ability-score-improvement", "epic-boon"} {
		key := "feats/" + slug
		r := choice(slug, 1, key, "feats/qualified")
		if len(r.PrerequisiteRefs) != 1 {
			t.Fatal("lost nested qualification")
		}
		p := records[r.PrerequisiteRefs[0]]
		if p.ParentID != r.ID || p.Requirement == nil || p.Requirement.Key != "srd-5.2.1:mechanic/prerequisite/feat-qualification" {
			t.Fatal("lost qualification predicate")
		}
	}
	master := records["srd-5.2.1:character/choice/fighter/tactical-master"]
	if len(master.PrerequisiteRefs) != 1 {
		t.Fatal("lost usable mastery prerequisite")
	}
	predicate := records[master.PrerequisiteRefs[0]].Requirement
	if predicate == nil || predicate.Key != "srd-5.2.1:mechanic/fighter/tactical-master/usable-weapon-mastery" || !reflect.DeepEqual(predicate.RecordRefs, []string{"srd-5.2.1:character/feature/fighter/weapon-mastery"}) {
		t.Fatal("wrong mastery prerequisite")
	}
	for slug, count := range map[string]int{"extra-attack": 2, "two-extra-attacks": 3, "three-extra-attacks": 4} {
		r := records["srd-5.2.1:character/feature/fighter/"+slug]
		if !reflect.DeepEqual(r.Parameters, map[string]int{"srd-5.2.1:mechanic/fighter/attacks-per-attack-action": count}) {
			t.Fatalf("attack count %s", slug)
		}
	}
	for id, keys := range map[string][]string{
		"resource/fighter/second-wind":  {"fighter/second-wind/short-rest", "fighter/second-wind/long-rest"},
		"resource/fighter/action-surge": {"fighter/action-surge/short-rest", "fighter/action-surge/long-rest", "fighter/action-surge/once-per-turn", "fighter/action-surge/except-magic-action"},
		"resource/fighter/indomitable":  {"fighter/indomitable/long-rest"},
		"choice/fighter/fighting-style": {"fighter/fighting-style/replace-on-fighter-level", "fighter/fighting-style/recommended-defense"},
		"choice/fighter/weapon-mastery": {"fighter/weapon-mastery/replace-one-long-rest"},
		"choice/fighter/epic-boon":      {"fighter/epic-boon/recommended-combat-prowess"},
	} {
		r := records["srd-5.2.1:character/"+id]
		for _, key := range keys {
			found := false
			for _, required := range r.RequiredMechanics {
				if required == "srd-5.2.1:mechanic/"+key {
					found = true
				}
			}
			if !found {
				t.Fatalf("lost source requirement %s", key)
			}
		}
	}
	equipment := choice("starting-equipment", 1, "fighter/equipment/a", "fighter/equipment/b", "fighter/equipment/c")
	expected := []map[string]int{{"chain-mail": 1, "greatsword": 1, "flail": 1, "javelin": 8, "dungeoneers-pack": 1, "gp": 4}, {"studded-leather": 1, "scimitar": 1, "shortsword": 1, "longbow": 1, "arrows": 20, "quiver": 1, "dungeoneers-pack": 1, "gp": 11}, {"gp": 155}}
	if len(equipment.Options) != 3 {
		t.Fatal("lost equipment bundles")
	}
	for i, q := range expected {
		want := map[string]int{}
		for k, v := range q {
			want["srd-5.2.1:mechanic/equipment/"+k] = v
		}
		if !reflect.DeepEqual(equipment.Options[i].Quantities, want) {
			t.Fatalf("equipment %d", i)
		}
	}
	// Explicit reference composition reuses canonical IDs exactly once.
	subset := characterFixture(t)
	for _, r := range subset.Records {
		if !reflect.DeepEqual(r, records[r.ID]) {
			t.Fatalf("fixture diverged: %s", r.ID)
		}
	}
}

func TestFighterMetadataGuards(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"progression on class", func(c *CharacterCatalogue) { c.Records[0].Progression = c.Records[1].Progression }},
		{"unknown feature", func(c *CharacterCatalogue) { c.Records[1].Progression.FeatureRefs = []string{"missing"} }},
		{"future feature", func(c *CharacterCatalogue) {
			c.Records[1].Progression.FeatureRefs = []string{"srd-5.2.1:character/feature/fighter/three-extra-attacks"}
		}},
		{"unrequired table key", func(c *CharacterCatalogue) {
			c.Records[1].Progression.FeatureKeys = []string{"srd-5.2.1:mechanic/unknown"}
		}},
		{"invalid count", func(c *CharacterCatalogue) { c.Records[5].ValuesByLevel[0].Value = 0 }},
		{"duplicate level count", func(c *CharacterCatalogue) { c.Records[5].ValuesByLevel[1].Level = 1 }},
		{"initial capacity mismatch", func(c *CharacterCatalogue) { c.Records[5].Capacity = 3 }},
		{"resource before owning feature", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/resource/fighter/indomitable" {
					c.Records[i].ValuesByLevel[0].Level = 1
				}
			}
		}},
		{"choice before owning feature", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/choice/fighter/tactical-master" {
					c.Records[i].ValuesByLevel = []CharacterLevelValue{{Level: 1, Value: 1}}
				}
			}
		}},
		{"unknown quantity", func(c *CharacterCatalogue) { c.Records[10].Options[0].Quantities["unknown"] = 1 }},
		{"wrong nested predicate owner", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Kind == "prerequisite" {
					c.Records[i].ParentID = c.Records[0].ID
					break
				}
			}
		}},
		{"resource under choice", func(c *CharacterCatalogue) { c.Records[5].ParentID = c.Records[3].ID }},
		{"null feature grants", func(c *CharacterCatalogue) { c.Records[1].Progression.FeatureRefs = nil }},
		{"duplicate grants", func(c *CharacterCatalogue) {
			c.Records[1].Progression.FeatureRefs = append(c.Records[1].Progression.FeatureRefs, c.Records[1].Progression.FeatureRefs[0])
		}},
		{"unrequired parameter", func(c *CharacterCatalogue) { c.Records[0].Parameters["srd-5.2.1:mechanic/unknown"] = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fighterCatalogue(t)
			tt.mutate(&c)
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
	c := fighterCatalogue(t)
	for _, key := range []string{"srd-5.2.1:mechanic/unknown-fighter-choice", "srd-5.2.1:mechanic/unknown-fighter-feature"} {
		c.Records[0].RequiredMechanics = append([]string{key}, c.Records[0].RequiredMechanics...)
		if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), key) {
			t.Fatalf("unknown key not rejected: %v", err)
		}
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCharacterCatalogue(strings.NewReader(strings.Replace(string(b), `"proficiency_bonus":2`, `"proficiency_bonus":2,"playable":true`, 1))); err == nil {
		t.Fatal("accepted support flag")
	}
}
