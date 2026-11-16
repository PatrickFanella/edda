package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const lifeDomainID = "srd-5.2.1:character/"
const lifeDomainKey = "srd-5.2.1:mechanic/cleric/life-domain/"

func lifeDomainCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-life-domain-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-cleric.json", "character-life-domain.json"}) {
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

func TestLifeDomainSourceAndMechanics(t *testing.T) {
	batch, c := lifeDomainCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Cleric accepted")
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("metadata available")
	}
	if !reflect.DeepEqual(c.Records[:len(c.Records)-len(batch.Records)], clericCatalogue(t).Records) {
		t.Fatal("canonical Cleric changed")
	}
	records := map[string]CharacterRecord{}
	identities := map[string]int{}
	for _, r := range batch.Records {
		records[r.ID] = r
		if len(r.Sources) != 1 || r.Sources[0].Page != 40 {
			t.Fatal("source page")
		}
		if r.Kind == "feature" || r.Kind == "subclass" {
			identities[r.Sources[0].InventoryID]++
		}
		if r.Kind == "choice" || r.Kind == "resource" || r.Kind == "advancement" {
			t.Fatal("scope widened")
		}
		for _, k := range r.RequiredMechanics {
			if qualifiedCharacterMechanic(k) {
				t.Fatal("available mechanic")
			}
		}
	}
	expected := map[string]int{"srd-5.2.1:planning/subclasses/cleric-life-domain": 1}
	for slug, n := range map[string]int{"disciple-of-life": 3, "life-domain-spells": 3, "preserve-life": 3, "blessed-healer": 6, "supreme-healing": 17} {
		expected[fmt.Sprintf("srd-5.2.1:planning/subclass-features/cleric-level-%d-%s", n, slug)] = 1
		r := records[lifeDomainID+"feature/cleric/life-domain/"+slug]
		if r.Name != map[string]string{"disciple-of-life": "Disciple of Life", "life-domain-spells": "Life Domain Spells", "preserve-life": "Preserve Life", "blessed-healer": "Blessed Healer", "supreme-healing": "Supreme Healing"}[slug] || r.ParentID != lifeDomainID+"subclass/cleric/life-domain" {
			t.Fatal("identity/ownership")
		}
		p := records[r.PrerequisiteRefs[0]]
		want := &CharacterRequirement{Key: "srd-5.2.1:mechanic/prerequisite/class-level", RecordRefs: []string{lifeDomainID + "class/cleric"}, Parameters: map[string]int{"minimum-level": n}}
		if p.ParentID != r.ID || !reflect.DeepEqual(p.Requirement, want) {
			t.Fatalf("level predicate %s", slug)
		}
	}
	if !reflect.DeepEqual(identities, expected) || len(batch.Records) != 13 {
		t.Fatalf("scope %v %d", identities, len(batch.Records))
	}
	sub := records[lifeDomainID+"subclass/cleric/life-domain"]
	if sub.Name != "Life Domain" || sub.ParentID != lifeDomainID+"class/cleric" || !reflect.DeepEqual(records[sub.PrerequisiteRefs[0]].Requirement, &CharacterRequirement{Key: "srd-5.2.1:mechanic/prerequisite/class-level", RecordRefs: []string{lifeDomainID + "class/cleric"}, Parameters: map[string]int{"minimum-level": 3}}) {
		t.Fatal("subclass prerequisite")
	}
	// Independent transcription of p40, including timing and the spell-slot and
	// Bloodied restrictions that differ from older editions.
	keys := map[string][]string{
		"disciple-of-life":   {"spell-cast-with-spell-slot", "restores-hit-points-to-creature", "additional-healing/on-casting-turn", "additional-healing/plus-slot-level"},
		"life-domain-spells": {"always-prepared/thereafter-at-cleric-level"},
		"preserve-life":      {"magic-action", "present-holy-symbol", "expend-channel-divinity-use", "healing-pool/per-cleric-level", "targets/bloodied-creatures", "targets/can-include-self", "targets/divide-hit-points", "healing-cap/half-target-hit-point-maximum"},
		"blessed-healer":     {"spell-cast-with-spell-slot", "restores-hit-points-to-one-or-more-other-creatures", "self-healing/immediately-after-casting", "self-healing/plus-slot-level"},
		"supreme-healing":    {"when-would-roll-one-or-more-healing-dice", "healing/from-spell-or-channel-divinity", "healing/creature", "healing/replace-each-die-with-maximum"},
	}
	nums := map[string]map[string]int{
		"disciple-of-life": {"additional-healing/base": 2}, "life-domain-spells": {},
		"preserve-life":  {"channel-divinity/uses-expended": 1, "healing-pool/cleric-level-multiplier": 5, "targets/range-feet": 30, "healing-cap/hit-point-maximum-divisor": 2},
		"blessed-healer": {"self-healing/base": 2}, "supreme-healing": {"example/dice-count": 2, "example/die-sides": 6, "example/healing": 12},
	}
	spells := [][]string{{"aid", "bless", "cure-wounds", "lesser-restoration"}, {"mass-healing-word", "revivify"}, {"aura-of-life", "death-ward"}, {"greater-restoration", "mass-cure-wounds"}}
	rows := []CharacterPreparedSpells{}
	for i, level := range []int{3, 5, 7, 9} {
		row := CharacterPreparedSpells{Level: level}
		for _, spell := range spells[i] {
			row.SpellKeys = append(row.SpellKeys, "srd-5.2.1:mechanic/spells/"+spell)
		}
		rows = append(rows, row)
	}
	for slug, ks := range keys {
		r := records[lifeDomainID+"feature/cleric/life-domain/"+slug]
		wantKeys := map[string]bool{}
		wantNums := map[string]int{}
		prefix := lifeDomainKey + slug + "/"
		for _, k := range ks {
			wantKeys[prefix+k] = true
		}
		for k, v := range nums[slug] {
			wantKeys[prefix+k] = true
			wantNums[prefix+k] = v
		}
		if slug == "life-domain-spells" {
			if !reflect.DeepEqual(r.AlwaysPrepared, rows) {
				t.Fatal("fixed spell table")
			}
			for _, row := range rows {
				for _, k := range row.SpellKeys {
					wantKeys[k] = true
				}
			}
		} else if len(r.AlwaysPrepared) != 0 {
			t.Fatal("unexpected spell grant")
		}
		gotKeys := map[string]bool{}
		for _, k := range r.RequiredMechanics {
			gotKeys[k] = true
		}
		if !reflect.DeepEqual(gotKeys, wantKeys) || len(r.Parameters) != len(wantNums) {
			t.Fatalf("mechanics %s", slug)
		}
		for k, v := range wantNums {
			if r.Parameters[k] != v {
				t.Fatalf("numeric %s", k)
			}
		}
	}
	r := records[lifeDomainID+"feature/cleric/life-domain/preserve-life"]
	if len(r.PrerequisiteRefs) != 2 || !reflect.DeepEqual(records[r.PrerequisiteRefs[1]].Requirement, &CharacterRequirement{Key: lifeDomainKey + "preserve-life/requires-channel-divinity-use", RecordRefs: []string{lifeDomainID + "resource/cleric/channel-divinity"}, Parameters: map[string]int{"uses-required": 1}}) {
		t.Fatal("canonical resource operand")
	}
}

func TestLifeDomainMetadataGuards(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue, int)
	}{
		{"duplicate identity", func(c *CharacterCatalogue, i int) { c.Records = append(c.Records, c.Records[i]) }},
		{"source page", func(c *CharacterCatalogue, i int) { c.Records[i].Sources[0].Page = 39 }},
		{"source owner", func(c *CharacterCatalogue, i int) { c.Records[i+3].Sources = c.Records[i].Sources }},
		{"missing ref", func(c *CharacterCatalogue, i int) {
			c.Records[i+1].Requirement.RecordRefs = []string{lifeDomainID + "class/missing"}
		}},
		{"wrong parent", func(c *CharacterCatalogue, i int) {
			c.Records[i].ParentID = lifeDomainID + "feature/cleric/spellcasting"
		}},
		{"lost reciprocal ref", func(c *CharacterCatalogue, i int) { c.Records[i].PrerequisiteRefs = []string{} }},
		{"negative predicate", func(c *CharacterCatalogue, i int) { c.Records[i+1].Requirement.Parameters["minimum-level"] = -1 }},
		{"unrequired numeric", func(c *CharacterCatalogue, i int) { c.Records[i+2].Parameters[lifeDomainKey+"unknown"] = 2 }},
		{"zero numeric", func(c *CharacterCatalogue, i int) {
			c.Records[i+2].Parameters[lifeDomainKey+"disciple-of-life/additional-healing/base"] = 0
		}},
		{"duplicate row", func(c *CharacterCatalogue, i int) { c.Records[i+4].AlwaysPrepared[1].Level = 3 }},
		{"invalid row level", func(c *CharacterCatalogue, i int) { c.Records[i+4].AlwaysPrepared[0].Level = 0 }},
		{"late row", func(c *CharacterCatalogue, i int) { c.Records[i+4].AlwaysPrepared[3].Level = 21 }},
		{"empty row", func(c *CharacterCatalogue, i int) { c.Records[i+4].AlwaysPrepared[0].SpellKeys = nil }},
		{"unrequired spell", func(c *CharacterCatalogue, i int) {
			c.Records[i+4].AlwaysPrepared[0].SpellKeys[0] = lifeDomainKey + "unknown"
		}},
		{"duplicate spell", func(c *CharacterCatalogue, i int) {
			c.Records[i+4].AlwaysPrepared[0].SpellKeys[0] = c.Records[i+4].AlwaysPrepared[0].SpellKeys[1]
		}},
		{"grant on subclass", func(c *CharacterCatalogue, i int) { c.Records[i].AlwaysPrepared = c.Records[i+4].AlwaysPrepared }},
		{"optional grant", func(c *CharacterCatalogue, i int) { c.Records[i+4].Select = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch, c := lifeDomainCatalogue(t)
			tt.mutate(&c, len(c.Records)-len(batch.Records))
			if err := c.Validate(os.DirFS("../../..")); err == nil {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
	_, c := lifeDomainCatalogue(t)
	for _, key := range []string{lifeDomainKey + "unknown-feature", "srd-5.2.1:mechanic/spells/unknown-spell"} {
		c.Records[0].RequiredMechanics = append([]string{key}, c.Records[0].RequiredMechanics...)
		if err := c.Validate(os.DirFS("../../..")); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), key) {
			t.Fatal("unknown key available")
		}
	}
	b, _ := json.Marshal(c)
	for _, field := range []string{"playable", "available", "implemented"} {
		input := strings.Replace(string(b), `"spell_keys":`, `"`+field+`":true,"spell_keys":`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("unknown nested field accepted")
		}
	}
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SupportedBuildForConcept("Human Cleric", "Acolyte"); err != nil {
		t.Fatal(err)
	}
	for _, concept := range []string{"Human Life Domain", "Human Cleric Life Domain"} {
		if _, err := SupportedBuildForConcept(concept, "Acolyte"); err == nil {
			t.Fatal("subclass creation enabled")
		}
	}
	// Exercise the existing fixed Cure Wounds path after catalogue validation:
	// 2d8 + 3 remains 5 at minimum dice, without a Life bonus or maximization.
	data, err := os.ReadFile("testdata/cleric1.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := Project(data, 1, 1, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	dice := &spellDice{}
	receipt, err := ResolveCureWounds(state, dice)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.HPAfter != 6 || dice.calls != 2 || receipt.SlotsAfter != 1 {
		t.Fatalf("fixed spell changed: %+v", receipt)
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if before.Coverage() != after.Coverage() || after.FullSupportEligible(os.DirFS("../../..")) {
		t.Fatal("coverage promoted")
	}
}
