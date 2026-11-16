package srd521

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func multiclassCatalogue(t *testing.T) MulticlassCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-multiclass.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeMulticlassCatalogue(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMulticlassCompleteSourceAndBoundaries(t *testing.T) {
	c := multiclassCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if !multiclassSourceAgrees(t, c) {
		t.Fatal("independent pp24–26 source expectations disagree")
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("multiclass available")
	}
	b, err := os.ReadFile("../../../docs/investigations/issue-74-character-inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	var inv struct {
		Records []json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(b, &inv); err != nil {
		t.Fatal(err)
	}
	if len(inv.Records) != 344 {
		t.Fatal("named inventory changed")
	}
	if pdf := os.Getenv("EDDA_SRD_PDF"); pdf != "" {
		if out, err := exec.Command("python3", "../../../scripts/verify_srd_multiclass.py", pdf).CombinedOutput(); err != nil {
			t.Fatalf("PDF comparator: %v\n%s", err, out)
		}
	}
	before, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	TestCharacterProjection(t)
	after, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if before.Coverage() != after.Coverage() || after.FullSupportEligible(repo) {
		t.Fatal("coverage promoted")
	}
	// Reuse the existing executable boundary checks as well as this catalogue's
	// unconditional refusal: no class combination has become a supported build.
	for _, concept := range []string{"Human Fighter / Cleric", "Human Wizard", "Human Warlock"} {
		if _, err := SupportedBuildForConcept(concept, "Soldier"); err == nil {
			t.Fatalf("unsupported build enabled: %s", concept)
		}
	}
	for _, fixture := range []struct{ concept, background string }{{"Human Fighter", "Soldier"}, {"Human Cleric", "Acolyte"}} {
		if _, err := SupportedBuildForConcept(fixture.concept, fixture.background); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMulticlassStructurallyValidContradictions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MulticlassCatalogue)
	}{
		{"slot-formula-class", func(c *MulticlassCatalogue) { c.SlotTerms[0].ClassRef = "srd-5.2.1:character/class/warlock" }},
		{"slot-formula-rounding", func(c *MulticlassCatalogue) { c.SlotTerms[3].Rounding = "down" }},
		{"slot-formula-divisor", func(c *MulticlassCatalogue) { c.SlotTerms[0].LevelDivisor = 2 }},
		{"known-wrong-class-ability", func(c *MulticlassCatalogue) { c.Classes[0].PrimaryAny[0][0] = "wisdom" }},
		{"wrong-score", func(c *MulticlassCatalogue) { c.Classes[6].MinimumScore = 12 }},
		{"fighter-and-instead-of-or", func(c *MulticlassCatalogue) { c.Classes[4].PrimaryAny = [][]string{{"strength", "dexterity"}} }},
		{"wrong-proficiency", func(c *MulticlassCatalogue) {
			c.Classes[0].Proficiencies = append(c.Classes[0].Proficiencies, "light-armor")
		}},
		{"wrong-known-owner", func(c *MulticlassCatalogue) {
			c.Requirements[6].RecordRefs[0] = "srd-5.2.1:character/feature/monk/extra-attack"
		}},
		{"wrong-class-level-operand", func(c *MulticlassCatalogue) { c.Requirements[1].RecordRefs = c.Requirements[1].RecordRefs[1:] }},
		{"round-down", func(c *MulticlassCatalogue) {
			c.Requirements[10].Clauses[1] = "add-half-paladin-and-ranger-levels-rounded-down"
		}},
		{"wrong-rounding-divisor", func(c *MulticlassCatalogue) { c.Requirements[10].Parameters["ranger-divisor"] = 3 }},
		{"pool-mixed-dice", func(c *MulticlassCatalogue) { c.Requirements[2].Clauses[4] = "different-die-types-pooled-together" }},
		{"pact-exception-reversed", func(c *MulticlassCatalogue) {
			c.Requirements[11].Clauses[3] = "pact-levels-included-in-shared-slot-sum"
		}},
		{"prepared-ownership", func(c *MulticlassCatalogue) {
			c.Requirements[8].RecordRefs[0] = "srd-5.2.1:character/feature/warlock/pact-magic"
		}},
		{"extra-attack-stacks", func(c *MulticlassCatalogue) { c.Requirements[6].Parameters["ordinary-maximum-attacks"] = 3 }},
		{"ac-stacks", func(c *MulticlassCatalogue) { c.Requirements[5].Parameters["simultaneous-calculations"] = 2 }},
		{"table-cell", func(c *MulticlassCatalogue) { c.SlotRows[19].Slots[8] = 2 }},
		{"invented-grant", func(c *MulticlassCatalogue) {
			c.Requirements[4].Clauses = append(c.Requirements[4].Clauses, "grant-all-starting-equipment")
		}},
		{"invented-scaling", func(c *MulticlassCatalogue) { c.Requirements[9].Clauses[0] = "power-scaling-summed-caster-level" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := multiclassCatalogue(t)
			tt.mutate(&c)
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("control must be structurally valid: %v", err)
			}
			if multiclassSourceAgrees(t, c) {
				t.Fatal("Go source proof accepted contradiction")
			}
			if c.ValidateReadiness(os.DirFS("../../..")) == nil {
				t.Fatal("contradiction became ready")
			}
			if pdf := os.Getenv("EDDA_SRD_PDF"); pdf != "" {
				b, err := json.Marshal(c)
				if err != nil {
					t.Fatal(err)
				}
				file := filepath.Join(t.TempDir(), "candidate.json")
				if err := os.WriteFile(file, b, 0600); err != nil {
					t.Fatal(err)
				}
				if out, err := exec.Command("python3", "../../../scripts/verify_srd_multiclass.py", pdf, file).CombinedOutput(); err == nil {
					t.Fatalf("PDF proof accepted contradiction: %s", out)
				}
			}
		})
	}
}

func TestMulticlassStrictContract(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MulticlassCatalogue)
	}{
		{"schema", func(c *MulticlassCatalogue) { c.SchemaVersion = 2 }},
		{"source", func(c *MulticlassCatalogue) { c.Source.SHA256 = "bad" }},
		{"attribution", func(c *MulticlassCatalogue) { c.Source.Attribution = "" }},
		{"adaptation", func(c *MulticlassCatalogue) { c.Source.Changes = "" }},
		{"scope-issue", func(c *MulticlassCatalogue) { c.Scope.Issue = 284 }},
		{"scope-range", func(c *MulticlassCatalogue) { c.Scope.LastPage = 27 }},
		{"scope-name", func(c *MulticlassCatalogue) { c.Scope.Section = "trinkets" }},
		{"requirement-page", func(c *MulticlassCatalogue) { c.Requirements[0].FirstPage = 23 }},
		{"unknown-scope", func(c *MulticlassCatalogue) { c.Requirements[0].Key += "-unknown" }},
		{"unknown-source", func(c *MulticlassCatalogue) { c.Classes[0].Source.InventoryID = "unknown" }},
		{"wrong-page", func(c *MulticlassCatalogue) { c.Classes[0].Source.Page++ }},
		{"missing-class", func(c *MulticlassCatalogue) { c.Classes = c.Classes[:11] }},
		{"duplicate-class", func(c *MulticlassCatalogue) { c.Classes[1] = c.Classes[0] }},
		{"duplicate-canonical", func(c *MulticlassCatalogue) { c.Catalogues = append(c.Catalogues, c.Catalogues[0]) }},
		{"duplicate-advancement", func(c *MulticlassCatalogue) { c.Catalogues = append(c.Catalogues, "character-advancement.json") }},
		{"missing-advancement", func(c *MulticlassCatalogue) { c.Catalogues = c.Catalogues[:12] }},
		{"unknown-record", func(c *MulticlassCatalogue) {
			c.Requirements[0].RecordRefs = []string{"srd-5.2.1:character/class/unknown"}
		}},
		{"duplicate-requirement", func(c *MulticlassCatalogue) { c.Requirements[1] = c.Requirements[0] }},
		{"empty-alternative", func(c *MulticlassCatalogue) { c.Classes[0].PrimaryAny = [][]string{{}} }},
		{"unknown-ability", func(c *MulticlassCatalogue) { c.Classes[0].PrimaryAny[0][0] = "luck" }},
		{"unknown-proficiency", func(c *MulticlassCatalogue) { c.Classes[0].Proficiencies = []string{"heavy-armor"} }},
		{"missing-table", func(c *MulticlassCatalogue) { c.SlotRows = c.SlotRows[:19] }},
		{"duplicate-level", func(c *MulticlassCatalogue) { c.SlotRows[1].Level = 1 }},
		{"wrong-width", func(c *MulticlassCatalogue) { c.SlotRows[0].Slots = c.SlotRows[0].Slots[:8] }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := multiclassCatalogue(t)
			tt.mutate(&c)
			if c.Validate(os.DirFS("../../..")) == nil {
				t.Fatal("invalid contract accepted")
			}
		})
	}
	b, err := os.ReadFile("testdata/character-multiclass.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{string(b) + " {}", strings.Replace(string(b), `"schema_version":`, `"runtime_grants":[],"schema_version":`, 1), strings.Replace(string(b), `"minimum_score":`, `"automatic_grant":true,"minimum_score":`, 1)} {
		if _, err := DecodeMulticlassCatalogue(strings.NewReader(raw)); err == nil {
			t.Fatal("unknown fields/trailing JSON accepted")
		}
	}
	c := multiclassCatalogue(t)
	c.Requirements[0].Clauses = append(c.Requirements[0].Clauses, "unknown-mechanic")
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if c.ValidateReadiness(os.DirFS("../../..")) == nil {
		t.Fatal("unknown mechanic ready")
	}
}
