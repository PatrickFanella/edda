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

func hunterCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-hunter-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-ranger.json", "character-hunter.json"}) {
		t.Fatal("composition changed")
	}
	composed := rangerCatalogue(t)
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

func hunterSourceError(c CharacterCatalogue) error {
	want := hunterSourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("record count %d, want %d", len(c.Records), len(want))
	}
	for i, r := range c.Records {
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("p61 metadata %s:\n got %+v\nwant %+v", r.ID, r, want[i])
		}
	}
	return nil
}

func TestHunterSourceCompositionAndUnavailable(t *testing.T) {
	batch, c := hunterCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	TestRangerImportDoesNotPromoteFixedBuilds(t)
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Ranger accepted")
	}
	if err := validateSource(batch.Source); err != nil {
		t.Fatal(err)
	}
	if err := hunterSourceError(batch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Records[:len(c.Records)-len(batch.Records)], rangerCatalogue(t).Records) {
		t.Fatal("Ranger changed")
	}
	TestRangerCompleteSource(t)
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("metadata playable")
	}
	for _, r := range batch.Records {
		for _, k := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(k) {
				t.Fatalf("qualified %s", k)
			}
		}
	}
	duplicateRanger := c
	duplicateRanger.Records = append(append([]CharacterRecord{}, c.Records...), rangerCatalogue(t).Records...)
	if duplicateRanger.Validate(repo) == nil {
		t.Fatal("duplicate canonical Ranger accepted")
	}
	c.Records = append(c.Records, batch.Records...)
	if err := c.Validate(repo); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	for _, concept := range []string{"Human Ranger", "Human Hunter", "Human Ranger Hunter"} {
		if _, err := SupportedBuildForConcept(concept, "Acolyte"); err == nil {
			t.Fatal("creation enabled")
		}
	}
	fighter, err := os.ReadFile("testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Project([]byte(strings.ReplaceAll(string(fighter), `"fighter"`, `"ranger"`)), 1, 1, 10, 0); err == nil {
		t.Fatal("Ranger projection enabled")
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

func TestHunterContrarySourceBothProofs(t *testing.T) {
	c, _ := hunterCatalogue(t)
	repo := os.DirFS("../../..")
	var controls [][]CharacterRecord
	check := func(name string, mutate func(*CharacterCatalogue)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			b, composed := hunterCatalogue(t)
			mutate(&b)
			composed.Records = append(composed.Records[:len(composed.Records)-len(b.Records)], b.Records...)
			if err := composed.Validate(repo); err != nil {
				t.Fatalf("control must remain structurally valid: %v", err)
			}
			if composed.ValidateReadiness(repo) == nil {
				t.Fatal("contrary metadata playable")
			}
			if hunterSourceError(b) == nil {
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
				replacement := "srd-5.2.1:mechanic/ranger/contrary"
				q.RequiredMechanics[j] = replacement
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
					b.Records[i].Requirement.RecordRefs[j] = "srd-5.2.1:character/feature/ranger/spellcasting"
				})
			}
		}
		if r.Kind == "choice" {
			check(fmt.Sprintf("count-%d", i), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				q.Select++
			})
			check(fmt.Sprintf("option-%d", i), func(b *CharacterCatalogue) {
				b.Records[i].OptionKeys = append(b.Records[i].OptionKeys, "srd-5.2.1:mechanic/ranger/contrary")
			})
		}
		if r.Kind == "choice" {
			check(fmt.Sprintf("missing-alternative-%d", i), func(b *CharacterCatalogue) { b.Records[i].OptionKeys = b.Records[i].OptionKeys[1:] })
			check(fmt.Sprintf("invented-scaling-%d", i), func(b *CharacterCatalogue) {
				b.Records[i].ValuesByLevel = []CharacterLevelValue{{Level: 3, Value: 1}, {Level: 17, Value: 2}}
			})
		}
		// Every conditional option and replacement clause is independently required.
		for j, key := range r.RequiredMechanics {
			if len(r.RequiredMechanics) < 2 {
				continue
			}
			if r.Requirement != nil && r.Requirement.Key == key {
				continue
			}
			if _, numeric := r.Parameters[key]; numeric {
				continue
			}
			check(fmt.Sprintf("missing-clause-%d-%d", i, j), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				q.RequiredMechanics = append(q.RequiredMechanics[:j], q.RequiredMechanics[j+1:]...)
			})
		}

		for j := range r.OptionKeys {
			check(fmt.Sprintf("wrong-option-condition-%d-%d", i, j), func(b *CharacterCatalogue) { b.Records[i].OptionKeys[j] = "srd-5.2.1:mechanic/conditions/charmed" })
		}
	}
	b, err := json.Marshal(controls)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "contrary.json")
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("python3", "../../../scripts/verify_srd_hunter.py", "--controls", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("independent comparator: %v\n%s", err, out)
	}
	t.Logf("%d structurally valid contrary graphs rejected by both proofs and unavailable", len(controls))
}

func TestHunterSchemaUnknownAndPDF(t *testing.T) {
	batch, c := hunterCatalogue(t)
	b, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1),
		string(b) + ` {}`,
		strings.Replace(string(b), `"page":61`, `"page":60`, 1),
		strings.Replace(string(b), `"license":"CC-BY-4.0"`, `"license":"unknown"`, 1),
		strings.Replace(string(b), `"schema_version":1`, `"schema_version":2`, 1),
		strings.Replace(string(b), batch.Source.SHA256, strings.Repeat("0", 64), 1),
		strings.Replace(string(b), `"version":"5.2.1"`, `"version":"5.1"`, 1),
	} {
		bad, err := DecodeCharacterCatalogue(strings.NewReader(raw))
		if err == nil {
			bad.Records = append(append([]CharacterRecord{}, rangerCatalogue(t).Records...), bad.Records...)
		}
		if err == nil && bad.Validate(os.DirFS("../../..")) == nil {
			t.Fatal("invalid source/schema accepted")
		}
	}
	key := "srd-5.2.1:mechanic/hunter/unknown"
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
	cmd := exec.Command("python3", "../../../scripts/verify_srd_hunter.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF comparison: %v\n%s", err, out)
	}
}
