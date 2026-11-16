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

func landCatalogue(t *testing.T) (CharacterCatalogue, CharacterCatalogue) {
	t.Helper()
	b, err := os.ReadFile("testdata/character-circle-of-the-land-composition.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Catalogues []string `json:"catalogues"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Catalogues, []string{"character-druid.json", "character-circle-of-the-land.json"}) {
		t.Fatal("composition changed")
	}
	composed := druidCatalogue(t)
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

func landSourceError(c CharacterCatalogue) error {
	want := landSourceRecords()
	if len(c.Records) != len(want) {
		return fmt.Errorf("record count %d, want %d", len(c.Records), len(want))
	}
	for i, r := range c.Records {
		if !reflect.DeepEqual(r, want[i]) {
			return fmt.Errorf("p46 metadata %s:\n got %+v\nwant %+v", r.ID, r, want[i])
		}
	}
	return nil
}

func TestCircleOfTheLandSourceCompositionAndUnavailable(t *testing.T) {
	batch, c := landCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := batch.Validate(repo); err == nil {
		t.Fatal("missing canonical Druid accepted")
	}
	if err := validateSource(batch.Source); err != nil {
		t.Fatal(err)
	}
	if err := landSourceError(batch); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Records[:len(c.Records)-len(batch.Records)], druidCatalogue(t).Records) {
		t.Fatal("Druid changed")
	}
	if differences := druidSourceDifferences(t, druidCatalogue(t)); len(differences) != 0 {
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
	c.Records = append(c.Records, batch.Records...)
	if err := c.Validate(repo); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	for _, concept := range []string{"Human Druid", "Human Circle of the Land", "Human Druid Circle of the Land"} {
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

// The Python oracle emits the same adversarial catalogues for both proofs.
// Every one must pass structural validation before semantic rejection counts.
func TestCircleOfTheLandStructurallyValidContraryControls(t *testing.T) {
	cmd := exec.Command("python3", "-c", `import json,sys; sys.path.insert(0,"../../../scripts"); import verify_srd_circle_of_the_land as v; c=json.load(open("testdata/character-circle-of-the-land.json")); print(json.dumps(v.controls(c["records"])))`)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	var controls [][]CharacterRecord
	if err := json.Unmarshal(out, &controls); err != nil {
		t.Fatal(err)
	}
	if len(controls) < 150 {
		t.Fatalf("insufficient controls: %d", len(controls))
	}
	for i, records := range controls {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			batch, c := landCatalogue(t)
			c.Records = append(c.Records[:len(c.Records)-len(batch.Records)], records...)
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("control structurally invalid: %v", err)
			}
			batch.Records = records
			if landSourceError(batch) == nil {
				t.Fatal("contrary metadata accepted")
			}
		})
	}
	t.Logf("%d structurally valid contrary controls rejected", len(controls))
}

func TestCircleOfTheLandSchemaUnknownAndPDF(t *testing.T) {
	batch, c := landCatalogue(t)
	b, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCharacterCatalogue(strings.NewReader(strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1))); err == nil {
		t.Fatal("unknown field accepted")
	}
	key := "srd-5.2.1:mechanic/land/unknown"
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
	cmd := exec.Command("python3", "../../../scripts/verify_srd_circle_of_the_land.py", pdf)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("PDF comparison: %v\n%s", err, out)
	}
}
