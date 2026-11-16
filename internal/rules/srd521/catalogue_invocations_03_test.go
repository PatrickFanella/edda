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

func invocations03Catalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-invocations-03-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-warlock.json", "character-origin-feats.json", "character-invocations-02.json", "character-invocations-03.json"}) {
		t.Fatal("composition changed")
	}
	composed := warlockCatalogue(t)
	var batch CharacterCatalogue
	for _, file := range manifest.Catalogues[1:] {
		data, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := DecodeCharacterCatalogue(strings.NewReader(string(data)))
		if err != nil {
			t.Fatal(err)
		}
		composed.Records = append(composed.Records, loaded.Records...)
		if file == "character-invocations-03.json" {
			batch = loaded
		}
	}

	return batch, composed
}

func invocations03SourceError(c CharacterCatalogue) error {
	data, err := os.ReadFile("../../../docs/investigations/issue-74-character-inventory.json")
	if err != nil {
		return err
	}
	var inventory struct {
		Source Source `json:"source"`
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		return err
	}
	source := inventory.Source
	source.Changes = "Four Invocation identities on PDF p74 adapted into unavailable metadata with exact prerequisites, nonrepeatability, spell-slot exceptions, pact-weapon Attack action timing and Truesight range. Graph and mechanic keys are Edda additions. Canonical Warlock is external; Pact of the Blade and spells have separate owners. No executable support imported."
	if c.SchemaVersion != 1 || c.Source != source {
		return fmt.Errorf("source envelope differs")
	}
	want := invocations03SourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("record count %d, want %d", len(c.Records), len(want))
	}
	for i, r := range c.Records {
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("p74 metadata %s:\n got %+v\nwant %+v", r.ID, r, want[i])
		}
	}
	return nil
}

func TestInvocations03SourceCompositionAndUnavailable(t *testing.T) {
	batch, c := invocations03Catalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	TestWarlockSourceAndBoundaries(t)
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Warlock accepted")
	}
	if err := validateSource(batch.Source); err != nil {
		t.Fatal(err)
	}
	if err := invocations03SourceError(batch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Records[:len(warlockCatalogue(t).Records)], warlockCatalogue(t).Records) {
		t.Fatal("Warlock changed")
	}
	if err := warlockSourceError(warlockCatalogue(t)); err != nil {
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
	for _, file := range []string{"character-origin-feats.json", "character-invocations-02.json"} {
		data, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		external, err := DecodeCharacterCatalogue(strings.NewReader(string(data)))
		if err != nil {
			t.Fatal(err)
		}
		duplicate := c
		duplicate.Records = append(append([]CharacterRecord{}, c.Records...), external.Records...)
		if duplicate.Validate(repo) == nil {
			t.Fatal("duplicate external owner accepted", file)
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
	for _, concept := range []string{"Human Warlock", "Human Warlock Thirsting Blade", "Human Warlock Witch Sight"} {
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

func TestInvocations03ContrarySourceBothProofs(t *testing.T) {
	c, _ := invocations03Catalogue(t)
	repo := os.DirFS("../../..")
	var controls [][]CharacterRecord
	check := func(name string, mutate func(*CharacterCatalogue)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			b, composed := invocations03Catalogue(t)
			mutate(&b)
			composed.Records = append(composed.Records[:len(composed.Records)-len(c.Records)], b.Records...)
			if err := composed.Validate(repo); err != nil {
				t.Fatalf("control must remain structurally valid: %v", err)
			}
			if composed.ValidateReadiness(repo) == nil {
				t.Fatal("contrary metadata playable")
			}
			if invocations03SourceError(b) == nil {
				t.Fatal("contrary source accepted")
			}
			controls = append(controls, b.Records)
		})
	}
	for _, key := range []string{"invented/long-rest-recovery", "invented/range-feet", "invented/slot-cost", "invented/level-scaling"} {
		check(key, func(b *CharacterCatalogue) {
			q := &b.Records[0]
			full := "srd-5.2.1:mechanic/warlock/invocations03/" + key
			q.RequiredMechanics = append(q.RequiredMechanics, full)
			if q.Parameters == nil {
				q.Parameters = map[string]int{}
			}
			q.Parameters[full] = 30
		})
	}
	check("source-name-variant", func(b *CharacterCatalogue) { b.Records[0].Name = "Thirsting Blade (2014)" })
	for i, r := range c.Records {
		if r.Kind == "prerequisite" {
			check(fmt.Sprintf("missing-predicate-%d", i), func(b *CharacterCatalogue) {
				for n := range b.Records {
					q := &b.Records[n]
					for j, ref := range q.PrerequisiteRefs {
						if ref == r.ID {
							q.PrerequisiteRefs = append(q.PrerequisiteRefs[:j], q.PrerequisiteRefs[j+1:]...)
							break
						}
					}
				}
				b.Records = append(b.Records[:i], b.Records[i+1:]...)
			})
		}
		for j := range r.RequiredMechanics {
			check(fmt.Sprintf("mechanic-%d-%d", i, j), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				old := q.RequiredMechanics[j]
				replacement := "srd-5.2.1:mechanic/warlock/contrary"
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
				for _, operand := range []string{"feature/warlock/pact-magic", "choice/warlock/prepared-spells", "level/warlock/20"} {
					check(fmt.Sprintf("operand-%d-%d-%s", i, j, operand), func(b *CharacterCatalogue) {
						b.Records[i].Requirement.RecordRefs[j] = "srd-5.2.1:character/" + operand
					})
				}
			}
		}
		if r.Kind == "feature" && len(r.RequiredMechanics) > 1 {
			check(fmt.Sprintf("missing-condition-%d", i), func(b *CharacterCatalogue) { b.Records[i].RequiredMechanics = b.Records[i].RequiredMechanics[1:] })
		}

		if r.Kind == "feature" && len(r.RequiredMechanics) > 1 {
			for j, k := range r.RequiredMechanics {
				if _, numeric := r.Parameters[k]; numeric {
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
	good, err := json.Marshal(c.Records)
	if err != nil {
		t.Fatal(err)
	}
	goodPath := filepath.Join(t.TempDir(), "valid.json")
	if err := os.WriteFile(goodPath, good, 0600); err != nil {
		t.Fatal(err)
	}
	positive := exec.Command("python3", "../../../scripts/verify_srd_invocations_03.py", "--records", goodPath)
	if out, err := positive.CombinedOutput(); err != nil {
		t.Fatalf("independent positive control: %v\n%s", err, out)
	}
	cmd := exec.Command("python3", "../../../scripts/verify_srd_invocations_03.py", "--controls", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("independent comparator: %v\n%s", err, out)
	}
	t.Logf("%d structurally valid contrary graphs rejected by both proofs and unavailable", len(controls))
}

func TestInvocations03SchemaUnknownAndPDF(t *testing.T) {
	batch, c := invocations03Catalogue(t)
	b, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1),
		string(b) + ` {}`,
		strings.Replace(string(b), `"schema_version":1`, `"schema_version":2`, 1),
		strings.Replace(string(b), batch.Source.SHA256, strings.Repeat("0", 64), 1),
		strings.Replace(string(b), `"version":"5.2.1"`, `"version":"5.1"`, 1),
	} {
		bad, err := DecodeCharacterCatalogue(strings.NewReader(raw))
		if err == nil {
			bad.Records = append(append([]CharacterRecord{}, c.Records[:len(c.Records)-len(batch.Records)]...), bad.Records...)
		}
		if err == nil && bad.Validate(os.DirFS("../../..")) == nil {
			t.Fatal("invalid source/schema accepted")
		}
	}
	key := "srd-5.2.1:mechanic/invocations03/unknown"
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
	cmd := exec.Command("python3", "../../../scripts/verify_srd_invocations_03.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF comparison: %v\n%s", err, out)
	}
}

// Exercise the exact shared validator, especially the per-source exception:
// a class-feature primary source must not smuggle an invocation as secondary.
func TestInvocations03UniqueOwnership(t *testing.T) {
	repo := os.DirFS("../../..")
	cases := []struct {
		name   string
		change func(*CharacterCatalogue)
	}{
		{"secondary-source", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/feature/warlock/invocations/thirsting-blade" {
					c.Records[i].Sources = append(c.Records[i].Sources, CharacterSource{InventoryID: "srd-5.2.1:planning/invocations/witch-sight", Page: 74})
				}
			}
		}},
		{"secondary-source-bypass", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/feature/warlock/eldritch-invocations" {
					c.Records[i].Sources = append(c.Records[i].Sources, CharacterSource{InventoryID: "srd-5.2.1:planning/invocations/thirsting-blade", Page: 74})
				}
			}
		}},
		{"wrong-slug", func(c *CharacterCatalogue) {
			c.Records[len(c.Records)-len(invocations03SourceRecords())].Sources[0].InventoryID = "srd-5.2.1:planning/invocations/witch-sight"
		}},
		{"wrong-choice", func(c *CharacterCatalogue) {
			c.Records[len(c.Records)-len(invocations03SourceRecords())].ParentID = "srd-5.2.1:character/choice/warlock/prepared-spells"
		}},
		{"wrong-ancestry", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/choice/warlock/invocations" {
					c.Records[i].ParentID = "srd-5.2.1:character/feature/warlock/pact-magic"
				}
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, c := invocations03Catalogue(t)
			tc.change(&c)
			if c.Validate(repo) == nil {
				t.Fatal("invalid invocation ownership accepted")
			}
		})
	}
}
