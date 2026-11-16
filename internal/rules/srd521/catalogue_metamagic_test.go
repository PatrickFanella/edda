package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func metamagicCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-metamagic-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-sorcerer.json", "character-metamagic.json"}) {
		t.Fatal("composition changed")
	}
	composed := sorcererCatalogue(t)
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

func metamagicSourceError(c CharacterCatalogue) error {
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
	source.Changes = "Ten Metamagic identities on PDF pp66–67 adapted into unavailable option metadata. Conditions, alternatives, costs and exceptions are opaque keys and numbers; graph and keys are Edda additions. Canonical Sorcerer is composed externally once. No executable support or spell definitions imported."
	if c.SchemaVersion != 1 || c.Source != source {
		return fmt.Errorf("source envelope differs")
	}
	want := metamagicSourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("record count %d, want %d", len(c.Records), len(want))
	}
	for i, r := range c.Records {
		r.RequiredMechanics = append([]string{}, r.RequiredMechanics...)
		sort.Strings(r.RequiredMechanics)
		sort.Strings(want[i].RequiredMechanics)
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("pp66-67 metadata %s:\n got %+v\nwant %+v", r.ID, r, want[i])
		}
	}
	return nil
}

func TestMetamagicSourceCompositionAndUnavailable(t *testing.T) {
	batch, c := metamagicCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	TestSorcererCompleteSource(t)
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Sorcerer accepted")
	}
	if err := validateSource(batch.Source); err != nil {
		t.Fatal(err)
	}
	if err := metamagicSourceError(batch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Records[:len(c.Records)-len(batch.Records)], sorcererCatalogue(t).Records) {
		t.Fatal("Sorcerer changed")
	}
	if differences := sorcererSourceDifferences(t, sorcererCatalogue(t)); len(differences) > 0 {
		t.Fatal(differences)
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
	duplicateSorcerer := c
	duplicateSorcerer.Records = append(append([]CharacterRecord{}, c.Records...), sorcererCatalogue(t).Records...)
	if duplicateSorcerer.Validate(repo) == nil {
		t.Fatal("duplicate canonical Sorcerer accepted")
	}
	c.Records = append(c.Records, batch.Records...)
	if err := c.Validate(repo); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	for _, concept := range []string{"Human Sorcerer", "Human Metamagic", "Human Sorcerer Metamagic"} {
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

func TestMetamagicContrarySourceBothProofs(t *testing.T) {
	c, _ := metamagicCatalogue(t)
	repo := os.DirFS("../../..")
	var controls [][]CharacterRecord
	check := func(name string, mutate func(*CharacterCatalogue)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			b, composed := metamagicCatalogue(t)
			mutate(&b)
			composed.Records = append(composed.Records[:len(composed.Records)-len(c.Records)], b.Records...)
			if err := composed.Validate(repo); err != nil {
				t.Fatalf("control must remain structurally valid: %v", err)
			}
			if composed.ValidateReadiness(repo) == nil {
				t.Fatal("contrary metadata playable")
			}
			if metamagicSourceError(b) == nil {
				t.Fatal("contrary source accepted")
			}
			controls = append(controls, b.Records)
		})
	}
	for _, key := range []string{"invented/long-rest-recovery", "invented/range-feet"} {
		check(key, func(b *CharacterCatalogue) {
			q := &b.Records[0]
			full := "srd-5.2.1:mechanic/sorcerer/metamagic/" + key
			q.RequiredMechanics = append(q.RequiredMechanics, full)
			if q.Parameters == nil {
				q.Parameters = map[string]int{}
			}
			q.Parameters[full] = 30
		})
	}
	for i, r := range c.Records {
		if r.Kind == "feature" {
			check(fmt.Sprintf("missing-option-%d", i), func(b *CharacterCatalogue) {
				kept := []CharacterRecord{}
				for _, q := range b.Records {
					if q.ID != r.ID && q.ParentID != r.ID {
						kept = append(kept, q)
					}
				}
				b.Records = kept
			})
		}
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
				replacement := "srd-5.2.1:mechanic/sorcerer/contrary"
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
				for _, operand := range []string{"feature/sorcerer/innate-sorcery", "resource/sorcerer/innate-sorcery", "level/sorcerer/20"} {
					check(fmt.Sprintf("operand-%d-%d-%s", i, j, operand), func(b *CharacterCatalogue) {
						b.Records[i].Requirement.RecordRefs[j] = "srd-5.2.1:character/" + operand
					})
				}
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
				b.Records[i].OptionKeys = append(b.Records[i].OptionKeys, "srd-5.2.1:mechanic/sorcerer/contrary")
			})
		}
		if r.Kind == "choice" && len(r.OptionKeys) > 1 {
			check(fmt.Sprintf("missing-alternative-%d", i), func(b *CharacterCatalogue) { b.Records[i].OptionKeys = b.Records[i].OptionKeys[1:] })
			check(fmt.Sprintf("invented-scaling-%d", i), func(b *CharacterCatalogue) {
				b.Records[i].ValuesByLevel = []CharacterLevelValue{{Level: 3, Value: b.Records[i].Select}, {Level: 17, Value: b.Records[i].Select + 1}}
			})
		}
		if r.Kind == "feature" && len(r.RequiredMechanics) > 1 {
			check(fmt.Sprintf("missing-condition-%d", i), func(b *CharacterCatalogue) { b.Records[i].RequiredMechanics = b.Records[i].RequiredMechanics[1:] })
		}

		for j := range r.OptionKeys {
			check(fmt.Sprintf("wrong-option-condition-%d-%d", i, j), func(b *CharacterCatalogue) { b.Records[i].OptionKeys[j] = "srd-5.2.1:mechanic/conditions/charmed" })
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
	positive := exec.Command("python3", "../../../scripts/verify_srd_metamagic.py", "--records", goodPath)
	if out, err := positive.CombinedOutput(); err != nil {
		t.Fatalf("independent positive control: %v\n%s", err, out)
	}
	cmd := exec.Command("python3", "../../../scripts/verify_srd_metamagic.py", "--controls", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("independent comparator: %v\n%s", err, out)
	}
	t.Logf("%d structurally valid contrary graphs rejected by both proofs and unavailable", len(controls))
}

func TestMetamagicSchemaUnknownAndPDF(t *testing.T) {
	batch, c := metamagicCatalogue(t)
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
			bad.Records = append(append([]CharacterRecord{}, sorcererCatalogue(t).Records...), bad.Records...)
		}
		if err == nil && bad.Validate(os.DirFS("../../..")) == nil {
			t.Fatal("invalid source/schema accepted")
		}
	}
	key := "srd-5.2.1:mechanic/metamagic/unknown"
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
	cmd := exec.Command("python3", "../../../scripts/verify_srd_metamagic.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF comparison: %v\n%s", err, out)
	}
}

// metamagicSourceRecords is an independent Go transcription of pp66–67.
// It neither reads the JSON fixture nor executes the standalone Python proof.
func metamagicSourceRecords() []CharacterRecord {
	const c = "srd-5.2.1:character/"
	const m = "srd-5.2.1:mechanic/sorcerer/metamagic/"
	type row struct {
		slug, clauses string
		numbers       map[string]int
		alternatives  []string
	}
	rows := []row{
		{"careful-spell", "when-cast-forces-other-creatures-saving-throw choose-up-to-charisma-modifier-creatures chosen-auto-succeed-save chosen-no-damage-only-if-normally-half-on-success", map[string]int{"sorcery-point-cost": 1, "minimum-creatures": 1}, nil},
		{"distant-spell", "when-cast choose-one-range-alternative", map[string]int{"sorcery-point-cost": 1, "minimum-range-feet": 5, "range-multiplier": 2, "touch-range-feet": 30}, []string{"range-at-least-minimum-double", "range-touch-becomes-fixed-feet"}},
		{"empowered-spell", "when-roll-spell-damage reroll-up-to-charisma-modifier-damage-dice must-use-new-rolls may-use-after-different-metamagic-during-casting", map[string]int{"sorcery-point-cost": 1, "minimum-dice": 1}, nil},
		{"extended-spell", "when-cast-duration-at-least-minimum double-duration-up-to-maximum if-concentration-advantage-all-saves-to-maintain-it", map[string]int{"sorcery-point-cost": 1, "minimum-duration-minutes": 1, "duration-multiplier": 2, "maximum-duration-hours": 24}, nil},
		{"heightened-spell", "when-cast-forces-creature-saving-throw one-spell-target-disadvantage-all-saves-against-spell", map[string]int{"sorcery-point-cost": 2, "target-count": 1}, nil},
		{"quickened-spell", "when-cast-action-casting-time change-to-bonus-action-for-this-casting cannot-if-already-cast-level-one-plus-spell-current-turn cannot-cast-level-one-plus-spell-after-use-this-turn", map[string]int{"sorcery-point-cost": 2}, nil},
		{"seeking-spell", "when-spell-attack-roll-misses reroll-attack-d20 must-use-new-roll may-use-after-different-metamagic-during-casting", map[string]int{"sorcery-point-cost": 1, "die-sides": 20}, nil},
		{"subtle-spell", "when-cast omit-verbal-somatic-material-components retain-consumed-material-components retain-material-components-with-specified-cost", map[string]int{"sorcery-point-cost": 1}, nil},
		{"transmuted-spell", "when-cast-dealing-listed-damage-type change-that-type-to-one-other-listed-type", map[string]int{"sorcery-point-cost": 1}, []string{"acid", "cold", "fire", "lightning", "poison", "thunder"}},
		{"twinned-spell", "when-cast-spell-upcastable-with-higher-slot-to-target-additional-creature increase-effective-spell-level-not-slot-level", map[string]int{"sorcery-point-cost": 1, "effective-level-increase": 1}, nil},
	}
	out := []CharacterRecord{}
	for i, row := range rows {
		name := strings.ReplaceAll(row.slug, "-", " ")
		words := strings.Fields(name)
		for j, w := range words {
			words[j] = strings.ToUpper(w[:1]) + w[1:]
		}
		name = strings.Join(words, " ")
		page := 67
		if i == 0 {
			page = 66
		}
		record := func(kind, slug, label, parent string) CharacterRecord {
			return CharacterRecord{ID: c + kind + "/sorcerer/metamagic/" + slug, Kind: kind, Name: label, ParentID: parent, Sources: []CharacterSource{{InventoryID: "srd-5.2.1:planning/metamagic/" + row.slug, Page: page}}, ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, RequiredMechanics: []string{}, OptionKeys: []string{}}
		}
		prefix := m + row.slug + "/"
		feature := record("feature", row.slug, name, c+"choice/sorcerer/metamagic")
		for _, k := range strings.Fields(row.clauses) {
			feature.RequiredMechanics = append(feature.RequiredMechanics, prefix+k)
		}
		feature.Parameters = map[string]int{}
		// Numeric order is irrelevant to meaning; sort in the oracle comparison.
		for k, v := range row.numbers {
			feature.Parameters[prefix+k] = v
			feature.RequiredMechanics = append(feature.RequiredMechanics, prefix+k)
		}
		pred := record("prerequisite", row.slug+"-owners", name+" canonical owners", feature.ID)
		key := prefix + "chosen-option-spend-canonical-points-modify-own-cast-spell"
		pred.RequiredMechanics = []string{key}
		pred.Requirement = &CharacterRequirement{Key: key, RecordRefs: []string{c + "choice/sorcerer/metamagic", c + "resource/sorcerer/sorcery-points", c + "feature/sorcerer/spellcasting", c + "choice/sorcerer/prepared-spells"}, Parameters: map[string]int{}}
		if row.slug == "twinned-spell" {
			for n := 1; n <= 9; n++ {
				pred.Requirement.RecordRefs = append(pred.Requirement.RecordRefs, fmt.Sprintf("%sresource/sorcerer/spell-slot-%d", c, n))
			}
		}
		feature.PrerequisiteRefs = append(feature.PrerequisiteRefs, pred.ID)
		if len(row.alternatives) > 0 {
			choice := record("choice", row.slug+"-alternatives", name+" alternatives", feature.ID)
			choice.RequiredMechanics = []string{prefix + "choose-alternative"}
			choice.Select = 1
			for _, k := range row.alternatives {
				choice.OptionKeys = append(choice.OptionKeys, prefix+k)
			}
			feature.ChoiceRefs = append(feature.ChoiceRefs, choice.ID)
			out = append(out, feature, pred, choice)
		} else {
			out = append(out, feature, pred)
		}
	}
	return out
}

func TestMetamagicOwnershipGuards(t *testing.T) {
	repo := os.DirFS("../../..")
	tests := map[string]func(*CharacterCatalogue){
		"wrong-owner": func(b *CharacterCatalogue) {
			b.Records[0].ParentID = "srd-5.2.1:character/choice/sorcerer/prepared-spells"
		},
		"ordinary-level-owner": func(b *CharacterCatalogue) { b.Records[0].ParentID = "srd-5.2.1:character/level/sorcerer/2" },
		"wrong-source-slug":    func(b *CharacterCatalogue) { b.Records[0].Sources = b.Records[2].Sources },
		"extra-source": func(b *CharacterCatalogue) {
			b.Records[0].Sources = append(b.Records[0].Sources, b.Records[2].Sources...)
		},
		"wrong-feature-id": func(b *CharacterCatalogue) { b.Records[0].ID += "-other" },
		"feature-source-as-option": func(b *CharacterCatalogue) {
			b.Records[0].Sources = []CharacterSource{{InventoryID: "srd-5.2.1:planning/class-features/sorcerer-level-2-metamagic", Page: 66}}
		},
		"source-page": func(b *CharacterCatalogue) { b.Records[0].Sources[0].Page = 67 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			b, c := metamagicCatalogue(t)
			mutate(&b)
			c.Records = append(c.Records[:len(c.Records)-len(b.Records)], b.Records...)
			if c.Validate(repo) == nil {
				t.Fatal("invalid option ancestry accepted")
			}
		})
	}
	// Even canonical IDs cannot conceal a changed class ancestry or choice owner.
	for _, change := range []string{"choice-parent", "class-ancestry"} {
		t.Run(change, func(t *testing.T) {
			_, c := metamagicCatalogue(t)
			for i := range c.Records {
				r := &c.Records[i]
				if change == "choice-parent" && r.ID == "srd-5.2.1:character/choice/sorcerer/metamagic" {
					r.ParentID = "srd-5.2.1:character/feature/sorcerer/spellcasting"
				}
				if change == "class-ancestry" && r.ID == "srd-5.2.1:character/class/sorcerer" {
					r.Sources = []CharacterSource{{InventoryID: "srd-5.2.1:planning/classes/wizard", Page: 78}}
				}
			}
			if c.Validate(repo) == nil {
				t.Fatal("contrary canonical ancestry accepted")
			}
		})
	}
}

func TestMetamagicSecondarySourceCannotBypassOptionGuards(t *testing.T) {
	_, c := metamagicCatalogue(t)
	for i := range c.Records {
		r := &c.Records[i]
		if r.ID == "srd-5.2.1:character/feature/sorcerer/innate-sorcery" {
			r.Sources = append(r.Sources, CharacterSource{InventoryID: "srd-5.2.1:planning/metamagic/careful-spell", Page: 66})
		}
	}
	err := c.Validate(os.DirFS("../../.."))
	if err == nil || !strings.Contains(err.Error(), "Metamagic requires one source") {
		t.Fatalf("secondary source bypass: %v", err)
	}
}
