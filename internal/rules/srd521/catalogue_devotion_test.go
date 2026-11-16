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

func devotionCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-devotion-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-paladin.json", "character-devotion.json"}) {
		t.Fatal("composition changed")
	}
	composed := paladinCatalogue(t)
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

func devotionSourceError(c CharacterCatalogue) error {
	want := devotionSourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("record count %d, want %d", len(c.Records), len(want))
	}
	for i, r := range c.Records {
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("pp56–57 metadata %s:\n got %+v\nwant %+v", r.ID, r, want[i])
		}
	}
	return nil
}

func TestDevotionSourceCompositionAndUnavailable(t *testing.T) {
	batch, c := devotionCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	TestPaladinImportDoesNotPromoteFixedBuilds(t)
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Paladin accepted")
	}
	if err := validateSource(batch.Source); err != nil {
		t.Fatal(err)
	}
	if err := devotionSourceError(batch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Records[:len(c.Records)-len(batch.Records)], paladinCatalogue(t).Records) {
		t.Fatal("Paladin changed")
	}
	if !paladinMatches(paladinCatalogue(t)) {
		t.Fatal("canonical Paladin source changed")
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
	duplicatePaladin := c
	duplicatePaladin.Records = append(append([]CharacterRecord{}, c.Records...), paladinCatalogue(t).Records...)
	if duplicatePaladin.Validate(repo) == nil {
		t.Fatal("duplicate canonical Paladin accepted")
	}
	c.Records = append(c.Records, batch.Records...)
	if err := c.Validate(repo); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	for _, concept := range []string{"Human Paladin", "Human Oath of Devotion", "Human Paladin Oath of Devotion"} {
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

func TestDevotionContrarySourceBothProofs(t *testing.T) {
	c, _ := devotionCatalogue(t)
	repo := os.DirFS("../../..")
	var controls [][]CharacterRecord
	check := func(name string, mutate func(*CharacterCatalogue)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			b, composed := devotionCatalogue(t)
			mutate(&b)
			composed.Records = append(composed.Records[:len(composed.Records)-len(b.Records)], b.Records...)
			if err := composed.Validate(repo); err != nil {
				t.Fatalf("control must remain structurally valid: %v", err)
			}
			if composed.ValidateReadiness(repo) == nil {
				t.Fatal("contrary metadata playable")
			}
			if devotionSourceError(b) == nil {
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
				replacement := "srd-5.2.1:mechanic/paladin/contrary"
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
					b.Records[i].Requirement.RecordRefs[j] = "srd-5.2.1:character/feature/paladin/lay-on-hands"
				})
			}
		}
		for j := range r.AlwaysPrepared {
			check(fmt.Sprintf("table-row-level-%d-%d", i, j), func(b *CharacterCatalogue) { b.Records[i].AlwaysPrepared[j].Level++ })
			check(fmt.Sprintf("table-spell-%d-%d", i, j), func(b *CharacterCatalogue) {
				rows := b.Records[i].AlwaysPrepared
				other := (j + 1) % len(rows)
				rows[j].SpellKeys[0], rows[other].SpellKeys[0] = rows[other].SpellKeys[0], rows[j].SpellKeys[0]
			})
			check(fmt.Sprintf("missing-table-row-%d-%d", i, j), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				q.AlwaysPrepared = append(q.AlwaysPrepared[:j], q.AlwaysPrepared[j+1:]...)
			})
		}
		if r.Kind == "feature" {
			check(fmt.Sprintf("invented-feature-scaling-%d", i), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				key := "srd-5.2.1:mechanic/paladin/oath-of-devotion/invented-scaling/per-paladin-level"
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
	b, err := json.Marshal(controls)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "contrary.json")
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("python3", "../../../scripts/verify_srd_devotion.py", "--controls", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("independent comparator: %v\n%s", err, out)
	}
	t.Logf("%d structurally valid contrary graphs rejected by both proofs and unavailable", len(controls))
}

func TestDevotionSchemaUnknownAndPDF(t *testing.T) {
	batch, c := devotionCatalogue(t)
	b, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1),
		strings.Replace(string(b), `"requirement":{`, `"requirement":{"execute":true,`, 1),
		string(b) + ` {}`,
		strings.Replace(string(b), `"schema_version":1`, `"schema_version":2`, 1),
		strings.Replace(string(b), batch.Source.SHA256, strings.Repeat("0", 64), 1),
		strings.Replace(string(b), `"version":"5.2.1"`, `"version":"5.1"`, 1),
	} {
		bad, err := DecodeCharacterCatalogue(strings.NewReader(raw))
		if err == nil {
			bad.Records = append(append([]CharacterRecord{}, paladinCatalogue(t).Records...), bad.Records...)
		}
		if err == nil && bad.Validate(os.DirFS("../../..")) == nil {
			t.Fatal("invalid source/schema accepted")
		}
	}
	key := "srd-5.2.1:mechanic/devotion/unknown"
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
	cmd := exec.Command("python3", "../../../scripts/verify_srd_devotion.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF comparison: %v\n%s", err, out)
	}
}
