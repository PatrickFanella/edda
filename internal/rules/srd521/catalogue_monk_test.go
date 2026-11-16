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

func monkCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-monk.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func monkSourceError(c CharacterCatalogue) error {
	want := monkSourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("Monk record count")
	}
	for i, r := range c.Records {
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("Monk source mismatch %s: got %+v want %+v", r.ID, r, want[i])
		}
	}
	return nil
}
func TestMonkSourceCompositionAndUnavailable(t *testing.T) {
	c := monkCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := monkSourceError(c); err != nil {
		t.Fatal(err)
	}
	if c.ValidateReadiness(repo) == nil {
		t.Fatal("Monk available")
	}
	// Compose each canonical batch once, retaining Druid and Bard/Lore ownership.
	composed := c
	for _, name := range []string{"character-advancement.json", "character-fighter.json", "character-cleric.json", "character-barbarian.json", "character-berserker.json", "character-bard.json", "character-college-of-lore.json", "character-druid.json", "character-champion.json", "character-life-domain.json", "character-backgrounds.json", "character-species.json", "character-origin-feats.json", "character-general-feats.json", "character-epic-boon-feats.json", "character-fighting-styles.json"} {
		b, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		composed.Records = append(composed.Records, batch.Records...)
	}
	if err := composed.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if composed.ValidateReadiness(repo) == nil {
		t.Fatal("composition available")
	}
	composed.Records = append(composed.Records, c.Records...)
	if composed.Validate(repo) == nil {
		t.Fatal("duplicate accepted")
	}
	for _, concept := range []string{"Human Monk", "Human Warrior of the Open Hand"} {
		if _, err := SupportedBuildForConcept(concept, "Acolyte"); err == nil {
			t.Fatal("creation enabled")
		}
	}
	for _, r := range c.Records {
		for _, key := range r.RequiredMechanics {
			if qualifiedCharacterMechanic(key) {
				t.Fatal("mechanic qualified")
			}
		}
	}
	rules, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if rules.FullSupportEligible(repo) {
		t.Fatal("Full Rules promoted")
	}
}

func TestMonkContrarySourceBothProofs(t *testing.T) {
	c := monkCatalogue(t)
	repo := os.DirFS("../../..")
	var controls [][]CharacterRecord
	check := func(name string, mutate func(*CharacterCatalogue)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			b := monkCatalogue(t)
			mutate(&b)
			if err := b.Validate(repo); err != nil {
				t.Fatalf("control must remain structurally valid: %v", err)
			}
			if b.ValidateReadiness(repo) == nil {
				t.Fatal("contrary metadata playable")
			}
			if monkSourceError(b) == nil {
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
				replacement := "srd-5.2.1:mechanic/monk/contrary"
				q.RequiredMechanics[j] = replacement
				if q.Requirement != nil && q.Requirement.Key == old {
					q.Requirement.Key = replacement
				}
				if v, ok := q.Parameters[old]; ok {
					delete(q.Parameters, old)
					q.Parameters[replacement] = v
				}
				for k := range q.Options {
					if v, ok := q.Options[k].Quantities[old]; ok {
						delete(q.Options[k].Quantities, old)
						q.Options[k].Quantities[replacement] = v
					}
				}
				if q.Progression != nil {
					for k, key := range q.Progression.FeatureKeys {
						if key == old {
							q.Progression.FeatureKeys[k] = replacement
						}
					}
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
					b.Records[i].Requirement.RecordRefs[j] = "srd-5.2.1:character/feature/monk/body-and-mind"
				})
			}
		}
		if r.Kind == "choice" {
			check(fmt.Sprintf("count-%d", i), func(b *CharacterCatalogue) {
				q := &b.Records[i]
				q.Select++
				if len(q.ValuesByLevel) > 0 {
					q.ValuesByLevel[0].Value = q.Select
				}
			})
			check(fmt.Sprintf("option-%d", i), func(b *CharacterCatalogue) {
				b.Records[i].OptionKeys = append(b.Records[i].OptionKeys, "srd-5.2.1:mechanic/monk/contrary")
			})
		}
		for j := range r.ValuesByLevel {
			if j > 0 {
				check(fmt.Sprintf("scaling-%d-%d", i, j), func(b *CharacterCatalogue) { b.Records[i].ValuesByLevel[j].Value++ })
			}
		}
		for j, o := range r.Options {
			for key := range o.Quantities {
				check(fmt.Sprintf("quantity-%d-%d-%s", i, j, key), func(b *CharacterCatalogue) { b.Records[i].Options[j].Quantities[key]++ })
			}
		}
		if r.Progression != nil {
			check(fmt.Sprintf("proficiency-%d", i), func(b *CharacterCatalogue) {
				p := b.Records[i].Progression
				p.ProficiencyBonus = 2
				if r.Progression.ProficiencyBonus == 2 {
					p.ProficiencyBonus = 3
				}
			})
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
	cmd := exec.Command("python3", "../../../scripts/verify_srd_monk.py", "--controls", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("independent comparator: %v\n%s", err, out)
	}
	t.Logf("%d structurally valid contrary graphs rejected by both proofs and unavailable", len(controls))
}

func TestMonkStrictSchemaUnknownKeyAndPin(t *testing.T) {
	c := monkCatalogue(t)
	repo := os.DirFS("../../..")
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1), string(b) + ` {}`, strings.Replace(string(b), `"schema_version":1`, `"schema_version":2`, 1)} {
		bad, err := DecodeCharacterCatalogue(strings.NewReader(raw))
		if err == nil && bad.Validate(repo) == nil {
			t.Fatal("invalid schema accepted")
		}
	}
	c.Records[0].RequiredMechanics = append([]string{"srd-5.2.1:mechanic/monk/unknown"}, c.Records[0].RequiredMechanics...)
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil || !strings.Contains(err.Error(), "monk/unknown") {
		t.Fatal("unknown mechanic available")
	}
	pdf := os.Getenv("EDDA_SRD_PDF")
	if pdf == "" {
		t.Skip("set EDDA_SRD_PDF for pinned PDF comparison")
	}
	cmd := exec.Command("python3", "../../../scripts/verify_srd_monk.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF: %v\n%s", err, out)
	}
}

func TestMonkFixedBuildRegressions(t *testing.T) {
	beforeFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	beforeCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	c := monkCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	afterFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	afterCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeFighter, afterFighter) || !reflect.DeepEqual(beforeCleric, afterCleric) || before.Coverage() != after.Coverage() {
		t.Fatal("fixed builds or coverage changed")
	}
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	TestCharacterProjection(t)
	TestLongswordFixture(t)
	TestCureWoundsFixture(t)
}
