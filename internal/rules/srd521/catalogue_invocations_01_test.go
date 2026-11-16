package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func invocations01Catalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-invocations-01-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-warlock.json", "character-invocations-01.json"}) {
		t.Fatal("composition changed")
	}
	composed := warlockCatalogue(t)
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

func invocations01SourceError(c CharacterCatalogue) error {
	want := invocations01SourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("record count %d, want %d", len(c.Records), len(want))
	}
	for i, r := range c.Records {
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("pp72–73 metadata %s:\n got %+v\nwant %+v", r.ID, r, want[i])
		}
	}
	return nil
}

func TestInvocations01SourceCompositionAndUnavailable(t *testing.T) {
	batch, c := invocations01Catalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	TestWarlockSourceAndBoundaries(t)
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Warlock accepted")
	}
	if err := validateSource(batch.Source); err != nil {
		t.Fatal(err)
	}
	if err := invocations01SourceError(batch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Records[:len(c.Records)-len(batch.Records)], warlockCatalogue(t).Records) {
		t.Fatal("Warlock changed")
	}
	if err := warlockSourceError(warlockCatalogue(t)); err != nil {
		t.Fatal("canonical Warlock source changed")
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
	duplicateWarlock := c
	duplicateWarlock.Records = append(append([]CharacterRecord{}, c.Records...), warlockCatalogue(t).Records...)
	if duplicateWarlock.Validate(repo) == nil {
		t.Fatal("duplicate canonical Warlock accepted")
	}
	c.Records = append(c.Records, batch.Records...)
	if err := c.Validate(repo); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	for _, concept := range []string{"Human Warlock", "Human Warlock Agonizing Blast", "Human Warlock Gift of the Protectors"} {
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

func TestInvocations01ContrarySourceBothProofs(t *testing.T) {
	c, _ := invocations01Catalogue(t)
	repo := os.DirFS("../../..")
	var controls [][]CharacterRecord
	check := func(name string, mutate func(*CharacterCatalogue)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			b, composed := invocations01Catalogue(t)
			mutate(&b)
			composed.Records = append(composed.Records[:len(warlockCatalogue(t).Records)], b.Records...)
			if err := composed.Validate(repo); err != nil {
				t.Fatalf("control must remain structurally valid: %v", err)
			}
			if composed.ValidateReadiness(repo) == nil {
				t.Fatal("contrary metadata playable")
			}
			if invocations01SourceError(b) == nil {
				t.Fatal("contrary source accepted")
			}
			controls = append(controls, b.Records)
		})
	}
	for i, r := range c.Records {
		for j := range r.RequiredMechanics {
			check(fmt.Sprintf("mechanic-%d-%d", i, j), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				old := q.RequiredMechanics[j]
				replacement := "srd-5.2.1:mechanic/warlock/contrary"
				q.RequiredMechanics[j] = replacement
				for row := range q.AlwaysPrepared {
					for col, key := range q.AlwaysPrepared[row].SpellKeys {
						if key == old {
							q.AlwaysPrepared[row].SpellKeys[col] = replacement
						}
					}
				}
				if q.Requirement != nil && q.Requirement.Key == old {
					q.Requirement.Key = replacement
				}
				if v, ok := q.Parameters[old]; ok {
					delete(q.Parameters, old)
					q.Parameters[replacement] = v
				}
			})
		}
		for key := range r.Parameters {
			check(fmt.Sprintf("number-%d-%s", i, key), func(b *CharacterCatalogue) { b.Records[i].Parameters[key]++ })
		}
		if r.Requirement != nil {
			for key := range r.Requirement.Parameters {
				check(fmt.Sprintf("predicate-number-%d-%s", i, key), func(b *CharacterCatalogue) { b.Records[i].Requirement.Parameters[key]++ })
			}
			for j := range r.Requirement.RecordRefs {
				check(fmt.Sprintf("operand-%d-%d", i, j), func(b *CharacterCatalogue) {
					b.Records[i].Requirement.RecordRefs[j] = "srd-5.2.1:character/feature/warlock/eldritch-invocations"
				})
			}
		}
		for j := range r.PrerequisiteRefs {
			check(fmt.Sprintf("omitted-prerequisite-%d-%d", i, j), func(b *CharacterCatalogue) {
				id := b.Records[i].PrerequisiteRefs[j]
				q := &b.Records[i]
				q.PrerequisiteRefs = append(q.PrerequisiteRefs[:j], q.PrerequisiteRefs[j+1:]...)
				kept := []CharacterRecord{}
				for _, q := range b.Records {
					if q.ID != id {
						kept = append(kept, q)
					}
				}
				b.Records = kept
			})
		}

		for j := range r.OptionKeys {
			check(fmt.Sprintf("wrong-option-%d-%d", i, j), func(b *CharacterCatalogue) {
				b.Records[i].OptionKeys[j] = "srd-5.2.1:mechanic/warlock/wrong-choice"
			})
			if len(r.OptionKeys) > 1 {
				check(fmt.Sprintf("missing-option-%d-%d", i, j), func(b *CharacterCatalogue) {
					q := &b.Records[i]
					q.OptionKeys = append(q.OptionKeys[:j], q.OptionKeys[j+1:]...)
				})
			}
		}
		if len(r.ChoiceRefs) > 0 {
			check(fmt.Sprintf("missing-nested-choice-%d", i), func(b *CharacterCatalogue) {
				id := b.Records[i].ChoiceRefs[0]
				b.Records[i].ChoiceRefs = b.Records[i].ChoiceRefs[1:]
				removed := map[string]bool{id: true}
				for _, q := range b.Records {
					if q.ParentID == id {
						removed[q.ID] = true
					}
				}
				kept := []CharacterRecord{}
				for _, q := range b.Records {
					if !removed[q.ID] {
						kept = append(kept, q)
					}
				}
				b.Records = kept
			})
		}

		if r.Kind == "feature" {
			check(fmt.Sprintf("invented-feature-scaling-%d", i), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				key := "srd-5.2.1:mechanic/warlock/invocations/invented-scaling/per-warlock-level"
				q.RequiredMechanics = append(q.RequiredMechanics, key)
				if q.Parameters == nil {
					q.Parameters = map[string]int{}
				}
				q.Parameters[key] = 2
			})
		}

		if r.Kind == "feature" && len(r.RequiredMechanics) > 1 {
			check(fmt.Sprintf("missing-condition-%d", i), func(b *CharacterCatalogue) { b.Records[i].RequiredMechanics = b.Records[i].RequiredMechanics[1:] })
		}

		if r.Kind == "feature" && len(r.RequiredMechanics) > 1 {
			for j, k := range r.RequiredMechanics {
				if _, numeric := r.Parameters[k]; numeric || strings.HasPrefix(k, "srd-5.2.1:mechanic/spells/") {
					continue
				}
				check(fmt.Sprintf("omitted-clause-%d-%d", i, j), func(b *CharacterCatalogue) {
					q := &b.Records[i]
					q.RequiredMechanics = append(q.RequiredMechanics[:j], q.RequiredMechanics[j+1:]...)
				})
			}
		}

	}
	check("previous-edition-fiendish-vigor-roll", func(b *CharacterCatalogue) {
		for i := range b.Records {
			q := &b.Records[i]
			if strings.HasSuffix(q.ID, "/fiendish-vigor") {
				q.RequiredMechanics[1] = "srd-5.2.1:mechanic/warlock/invocations/fiendish-vigor/roll-temporary-hit-points"
			}
		}
	})
	b, err := json.Marshal(controls)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "contrary.json")
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("python3", "../../../scripts/verify_srd_invocations_01.py", "--controls", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("independent comparator: %v\n%s", err, out)
	}
	t.Logf("%d structurally valid contrary graphs rejected by both proofs and unavailable", len(controls))
}

func TestInvocations01SchemaUnknownAndPDF(t *testing.T) {
	batch, c := invocations01Catalogue(t)
	b, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1),
		strings.Replace(string(b), `"select":1`, `"execute":true,"select":1`, 1),
		strings.Replace(string(b), `"requirement":{`, `"requirement":{"execute":true,`, 1),
		string(b) + ` {}`,
		strings.Replace(string(b), `"schema_version":1`, `"schema_version":2`, 1),
		strings.Replace(string(b), batch.Source.SHA256, strings.Repeat("0", 64), 1),
		strings.Replace(string(b), `"version":"5.2.1"`, `"version":"5.1"`, 1),
	} {
		bad, err := DecodeCharacterCatalogue(strings.NewReader(raw))
		if err == nil {
			bad.Records = append(append([]CharacterRecord{}, warlockCatalogue(t).Records...), bad.Records...)
		}
		if err == nil && bad.Validate(os.DirFS("../../..")) == nil {
			t.Fatal("invalid source/schema accepted")
		}
	}
	key := "srd-5.2.1:mechanic/invocations01/unknown"
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
	cmd := exec.Command("python3", "../../../scripts/verify_srd_invocations_01.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF comparison: %v\n%s", err, out)
	}
}

func TestInvocations01OwnershipExtensionGuards(t *testing.T) {
	repo := os.DirFS("../../..")
	for name, mutate := range map[string]func(*CharacterCatalogue){
		"other-choice-parent": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if strings.HasSuffix(c.Records[i].ID, "/agonizing-blast") {
					c.Records[i].ParentID = "srd-5.2.1:character/choice/warlock/cantrips"
				}
			}
		},
		"level-parent": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if strings.HasSuffix(c.Records[i].ID, "/agonizing-blast") {
					c.Records[i].ParentID = "srd-5.2.1:character/level/warlock/1"
				}
			}
		},
		"wrong-invocation-source": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if strings.HasSuffix(c.Records[i].ID, "/agonizing-blast") {
					c.Records[i].Sources[0].InventoryID = "srd-5.2.1:planning/invocations/eldritch-mind"
				}
			}
		},
		"nested-source": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Kind == "choice" && strings.HasSuffix(c.Records[i].ID, "/agonizing-blast-cantrip") {
					c.Records[i].Sources[0].Page = 73
				}
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, c := invocations01Catalogue(t)
			mutate(&c)
			if c.Validate(repo) == nil {
				t.Fatal("invalid invocation ownership accepted")
			}
		})
	}
}

func TestInvocations01RejectsSecondarySourceOnClassFeature(t *testing.T) {
	c := warlockCatalogue(t)
	for i := range c.Records {
		if c.Records[i].ID == "srd-5.2.1:character/feature/warlock/pact-magic" {
			c.Records[i].Sources = append(c.Records[i].Sources, CharacterSource{
				InventoryID: "srd-5.2.1:planning/invocations/agonizing-blast", Page: 72,
			})
		}
	}
	if err := c.Validate(os.DirFS("../../..")); err == nil {
		t.Fatal("class feature accepted an unrelated secondary invocation source")
	}
}
