package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func warlockCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-warlock.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func warlockSourceError(c CharacterCatalogue) error {
	expected := warlockSourceRecords()
	if len(c.Records) != len(expected) {
		return fmt.Errorf("record count: got %d want %d", len(c.Records), len(expected))
	}
	actual := map[string]CharacterRecord{}
	for _, r := range c.Records {
		if _, ok := actual[r.ID]; ok {
			return fmt.Errorf("duplicate %s", r.ID)
		}
		actual[r.ID] = r
	}
	for _, want := range expected {
		got, ok := actual[want.ID]
		if !ok {
			return fmt.Errorf("missing %s", want.ID)
		}
		// Required mechanics form a set. Preserve all other arrays and complete
		// structs, so new scaling, wrong operands and omitted fields are rejected.
		got.RequiredMechanics = slices.Clone(got.RequiredMechanics)
		want.RequiredMechanics = slices.Clone(want.RequiredMechanics)
		slices.Sort(got.RequiredMechanics)
		slices.Sort(want.RequiredMechanics)
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("source metadata %s:\n got %+v\nwant %+v", want.ID, got, want)
		}
	}
	return nil
}

func TestWarlockSourceAndBoundaries(t *testing.T) {
	c := warlockCatalogue(t)
	repo := os.DirFS("../../..")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := warlockSourceError(c); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("Warlock playable")
	}
	sources := map[CharacterSource]bool{}
	levels := 0
	for _, r := range c.Records {
		for _, s := range r.Sources {
			sources[s] = true
		}
		if r.Kind == "level" {
			levels++
		}
		for _, k := range r.RequiredMechanics {
			if qualifiedCharacterMechanic(k) {
				t.Fatal("qualified", k)
			}
		}
	}
	if len(sources) != 10 || levels != 20 {
		t.Fatalf("identities=%d levels=%d", len(sources), levels)
	}
	// Canonical composition includes each batch once, never a fixture copy of
	// another class or a replacement identity. Duplicate imports fail closed.
	composed := c
	for _, file := range []string{"character-species.json", "character-backgrounds.json", "character-origin-feats.json", "character-general-feats.json", "character-fighting-styles.json", "character-epic-boon-feats.json", "character-monk.json", "character-ranger.json", "character-rogue.json", "character-advancement.json", "character-barbarian.json", "character-bard.json", "character-cleric.json", "character-druid.json", "character-fighter.json", "character-berserker.json", "character-champion.json", "character-life-domain.json", "character-college-of-lore.json", "character-circle-of-the-land.json"} {
		b, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		composed.Records = append(composed.Records, other.Records...)
	}
	if err := composed.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := composed.ValidateReadiness(repo); err == nil {
		t.Fatal("composition playable")
	}
	composed.Records = append(composed.Records, c.Records...)
	if err := composed.Validate(repo); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	for _, background := range []string{"Soldier", "Acolyte", "Criminal"} {
		if _, err := SupportedBuildForConcept("Human Warlock", background); err == nil {
			t.Fatal("Warlock creation enabled")
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
}

func TestWarlockStructurallyValidContraryControls(t *testing.T) {
	cmd := exec.Command("python3", "-c", `import json,sys;sys.path.insert(0,"../../../scripts");import verify_srd_warlock as v;c=json.load(open("testdata/character-warlock.json"));print(json.dumps(v.controls(c["records"])))`)
	b, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	var controls [][]CharacterRecord
	if err := json.Unmarshal(b, &controls); err != nil {
		t.Fatal(err)
	}
	if len(controls) < 150 {
		t.Fatalf("too few controls: %d", len(controls))
	}
	for i, records := range controls {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			c := warlockCatalogue(t)
			c.Records = records
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("structurally invalid control: %v", err)
			}
			if warlockSourceError(c) == nil {
				t.Fatal("contrary metadata accepted")
			}
			if err := c.ValidateReadiness(os.DirFS("../../..")); err == nil {
				t.Fatal("contrary metadata playable")
			}
		})
	}
	t.Logf("%d structurally valid contrary controls rejected", len(controls))
}

func TestWarlockStrictSchemaSourceUnknown(t *testing.T) {
	c := warlockCatalogue(t)
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{string(b) + `{}`, strings.Replace(string(b), `"kind":`, `"playable":true,"kind":`, 1), strings.Replace(string(b), `"proficiency_bonus":`, `"derived":true,"proficiency_bonus":`, 1), strings.Replace(string(b), `"record_refs":`, `"execute":true,"record_refs":`, 1)} {
		if _, err := DecodeCharacterCatalogue(strings.NewReader(bad)); err == nil {
			t.Fatal("strict decoder accepted unknown/trailing data")
		}
	}
	for _, mutate := range []func(*CharacterCatalogue){
		func(c *CharacterCatalogue) { c.SchemaVersion = 2 },
		func(c *CharacterCatalogue) { c.Source.SHA256 = "wrong" },
		func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 71 },
		func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics = []string{"unknown"} },
		func(c *CharacterCatalogue) { c.Records[1].Select = 0 },
		func(c *CharacterCatalogue) {
			c.Records[0].ChoiceRefs = append(c.Records[0].ChoiceRefs, "srd-5.2.1:character/choice/warlock/missing")
		},
	} {
		fresh := warlockCatalogue(t)
		mutate(&fresh)
		if err := fresh.Validate(os.DirFS("../../..")); err == nil {
			t.Fatal("invalid structure/source accepted")
		}
	}
	fresh := warlockCatalogue(t)
	key := "srd-5.2.1:mechanic/warlock/unknown"
	fresh.Records[0].RequiredMechanics = append([]string{key}, fresh.Records[0].RequiredMechanics...)
	if err := fresh.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if err := fresh.ValidateReadiness(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), key) {
		t.Fatal("unknown key available")
	}
}
