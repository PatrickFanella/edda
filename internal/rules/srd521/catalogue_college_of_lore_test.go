package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func loreCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-college-of-lore-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-bard.json", "character-college-of-lore.json"}) {
		t.Fatal("composition changed")
	}
	composed := bardCatalogue(t)
	b, err = os.ReadFile("testdata/" + manifest.Catalogues[1])
	if err != nil {
		t.Fatal(err)
	}
	batch, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	composed.Records = append(composed.Records, batch.Records...)
	return batch, composed
}

func loreSourceError(c CharacterCatalogue) error {
	want := loreSourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("record count %d, want %d", len(c.Records), len(want))
	}
	for i, r := range c.Records {
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("p35 metadata %s:\n got %+v\nwant %+v", r.ID, r, want[i])
		}
	}
	return nil
}

func TestCollegeOfLoreSourceCompositionAndUnavailable(t *testing.T) {
	batch, c := loreCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Bard accepted")
	}
	if err := validateSource(batch.Source); err != nil {
		t.Fatal(err)
	}
	if err := loreSourceError(batch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Records[:len(c.Records)-len(batch.Records)], bardCatalogue(t).Records) {
		t.Fatal("Bard changed")
	}
	if err := bardProseSourceError(bardCatalogue(t)); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("metadata playable")
	}
	for _, r := range batch.Records {
		for _, k := range r.RequiredMechanics {
			if qualifiedCharacterMechanic(k) {
				t.Fatalf("qualified %s", k)
			}
		}
	}
	c.Records = append(c.Records, batch.Records...)
	if err := c.Validate(repo); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	for _, concept := range []string{"Human Bard", "Human College of Lore", "Human Bard College of Lore"} {
		if _, err := SupportedBuildForConcept(concept, "Acolyte"); err == nil {
			t.Fatal("creation enabled")
		}
	}
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	// Metadata validation cannot affect the qualified fixed spell path.
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
		t.Fatalf("fixed runtime changed: %+v", receipt)
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if before.Coverage() != after.Coverage() || after.FullSupportEligible(repo) {
		t.Fatal("coverage promoted")
	}
}

func TestCollegeOfLoreStructurallyValidContraryControls(t *testing.T) {
	batch, _ := loreCatalogue(t)
	// Every mechanic, number, source and existing operand must be compared by
	// semantics, even when the generic schema intentionally permits the change.
	check := func(name string, mutate func(*CharacterCatalogue)) {
		t.Run(name, func(t *testing.T) {
			b, c := loreCatalogue(t)
			mutate(&b)
			c.Records = append(c.Records[:len(c.Records)-len(b.Records)], b.Records...)
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("control must remain structurally valid: %v", err)
			}
			if loreSourceError(b) == nil {
				t.Fatal("contrary source accepted")
			}
		})
	}
	for i, r := range batch.Records {
		for j := range r.RequiredMechanics {
			check(fmt.Sprintf("mechanic-%d-%d", i, j), func(c *CharacterCatalogue) {
				old := c.Records[i].RequiredMechanics[j]
				c.Records[i].RequiredMechanics[j] = "srd-5.2.1:mechanic/lore/contrary"
				if q := c.Records[i].Requirement; q != nil && q.Key == old {
					q.Key = c.Records[i].RequiredMechanics[j]
				}
				if v, ok := c.Records[i].Parameters[old]; ok {
					delete(c.Records[i].Parameters, old)
					c.Records[i].Parameters[c.Records[i].RequiredMechanics[j]] = v
				}
			})
		}
		for key := range r.Parameters {
			check(fmt.Sprintf("number-%d-%s", i, key), func(c *CharacterCatalogue) { c.Records[i].Parameters[key]++ })
		}
		if r.Requirement != nil {
			for key := range r.Requirement.Parameters {
				check(fmt.Sprintf("predicate-number-%d-%s", i, key), func(c *CharacterCatalogue) { c.Records[i].Requirement.Parameters[key]++ })
			}
			for j := range r.Requirement.RecordRefs {
				check(fmt.Sprintf("existing-wrong-operand-%d-%d", i, j), func(c *CharacterCatalogue) {
					replacement := "srd-5.2.1:character/feature/bard/spellcasting"
					c.Records[i].Requirement.RecordRefs[j] = replacement
				})
			}
		}
		if r.Kind == "feature" {
			check(fmt.Sprintf("source-owner-%d", i), func(c *CharacterCatalogue) {
				oldSource := c.Records[i].Sources[0]
				c.Records[i].Sources = []CharacterSource{{InventoryID: "srd-5.2.1:planning/subclass-features/bard-level-14-peerless-skill", Page: 35}}
				if r.Name == "Peerless Skill" {
					c.Records[i].Sources[0].InventoryID = "srd-5.2.1:planning/subclass-features/bard-level-3-cutting-words"
				}
				for n := range c.Records {
					if n != i && c.Records[n].Sources[0] == oldSource {
						c.Records[n].Sources = c.Records[i].Sources
					}
				}
			})
		}
		if r.Kind == "choice" {
			check(fmt.Sprintf("count-%d", i), func(c *CharacterCatalogue) { c.Records[i].Select++ })
			check(fmt.Sprintf("missing-option-%d", i), func(c *CharacterCatalogue) { c.Records[i].OptionKeys = c.Records[i].OptionKeys[1:] })
			check(fmt.Sprintf("extra-option-%d", i), func(c *CharacterCatalogue) {
				c.Records[i].OptionKeys = append(c.Records[i].OptionKeys, "srd-5.2.1:mechanic/bard/spell-list")
			})
			check(fmt.Sprintf("invented-scaling-%d", i), func(c *CharacterCatalogue) {
				c.Records[i].ValuesByLevel = []CharacterLevelValue{{Level: 6, Value: c.Records[i].Select}, {Level: 14, Value: 4}}
			})
		}
	}
}

func TestCollegeOfLoreSchemaUnknownAndPDF(t *testing.T) {
	batch, c := loreCatalogue(t)
	b, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCharacterCatalogue(strings.NewReader(strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1))); err == nil {
		t.Fatal("unknown field accepted")
	}
	key := "srd-5.2.1:mechanic/lore/unknown"
	c.Records[0].RequiredMechanics = append([]string{key}, c.Records[0].RequiredMechanics...)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), key) {
		t.Fatal("unknown mechanic available")
	}
	pdf := os.Getenv("EDDA_SRD_PDF")
	if pdf == "" {
		t.Skip("set EDDA_SRD_PDF for offline pinned PDF comparison")
	}
	cmd := exec.Command("python3", "../../../scripts/verify_srd_college_of_lore.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF comparison: %v\n%s", err, out)
	}
}
