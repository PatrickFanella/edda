package srd521

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const barbarianID = "srd-5.2.1:character/"
const barbarianKey = "srd-5.2.1:mechanic/"

func barbarianCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-barbarian.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestBarbarianSourceAndEveryTableCell(t *testing.T) {
	c := barbarianCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	byID := map[string]CharacterRecord{}
	identities := map[string]int{}
	for _, r := range c.Records {
		byID[r.ID] = r
		if r.Kind == "class" || r.Kind == "feature" {
			identities[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
		if r.Kind == "advancement" || r.Kind == "subclass" || r.Kind == "feat" || r.Kind == "origin" {
			t.Fatalf("scope widened: %s", r.ID)
		}
	}
	// Independent transcription of the 20 named headings; the two identical
	// Improved Brutal Strike names retain different level-owned identities.
	headings := map[int][]string{1: {"rage", "unarmored-defense", "weapon-mastery"}, 2: {"danger-sense", "reckless-attack"}, 3: {"barbarian-subclass", "primal-knowledge"}, 4: {"ability-score-improvement"}, 5: {"extra-attack", "fast-movement"}, 7: {"feral-instinct", "instinctive-pounce"}, 9: {"brutal-strike"}, 11: {"relentless-rage"}, 13: {"improved-brutal-strike"}, 15: {"persistent-rage"}, 17: {"improved-brutal-strike"}, 18: {"indomitable-might"}, 19: {"epic-boon"}, 20: {"primal-champion"}}
	wantIDs := map[string]int{"srd-5.2.1:planning/classes/barbarian": 28}
	for level, names := range headings {
		for _, name := range names {
			page := 29
			if name == "rage" {
				page = 28
			}
			if level >= 11 {
				page = 30
			}
			wantIDs[fmt.Sprintf("srd-5.2.1:planning/class-features/barbarian-level-%d-%s", level, name)] = page
		}
	}
	if len(identities) != 21 || !reflect.DeepEqual(identities, wantIDs) {
		t.Fatalf("identities/pages: %v", identities)
	}
	// p28: proficiency, uses, damage bonus, mastery selections. Repeated cells
	// are intentional and test the table-only improvements independently.
	rows := [20][4]int{{2, 2, 2, 2}, {2, 2, 2, 2}, {2, 3, 2, 2}, {2, 3, 2, 3}, {3, 3, 2, 3}, {3, 4, 2, 3}, {3, 4, 2, 3}, {3, 4, 2, 3}, {4, 4, 3, 3}, {4, 4, 3, 4}, {4, 4, 3, 4}, {4, 5, 3, 4}, {5, 5, 3, 4}, {5, 5, 3, 4}, {5, 5, 3, 4}, {5, 5, 4, 4}, {6, 6, 4, 4}, {6, 6, 4, 4}, {6, 6, 4, 4}, {6, 6, 4, 4}}
	count := 0
	for _, r := range c.Records {
		if r.Kind == "level" {
			count++
		}
	}
	if count != 20 {
		t.Fatal("incomplete progression")
	}
	for i, cells := range rows {
		n := i + 1
		r := byID[fmt.Sprintf("%slevel/barbarian/%d", barbarianID, n)]
		if r.Progression == nil || r.Progression.ProficiencyBonus != cells[0] || !reflect.DeepEqual(r.Parameters, map[string]int{barbarianKey + "barbarian/rage/damage-bonus": cells[2]}) {
			t.Fatalf("row %d numbers", n)
		}
		slugs := headings[n]
		if n == 8 || n == 12 || n == 16 {
			slugs = []string{"ability-score-improvement"}
		}
		refs := []string{}
		for _, slug := range slugs {
			if slug == "improved-brutal-strike" {
				slug += fmt.Sprintf("-%d", n)
			}
			refs = append(refs, barbarianID+"feature/barbarian/"+slug)
		}
		keys := []string{}
		if n == 6 || n == 10 || n == 14 {
			keys = []string{barbarianKey + "barbarian/subclass-feature"}
		}
		if !reflect.DeepEqual(r.Progression.FeatureRefs, refs) || !reflect.DeepEqual(r.Progression.FeatureKeys, keys) {
			t.Fatalf("row %d grants", n)
		}
		for col, id := range []string{"resource/barbarian/rage", "choice/barbarian/weapon-mastery"} {
			values := byID[barbarianID+id].ValuesByLevel
			want := cells[1]
			if col == 1 {
				want = cells[3]
			}
			if len(values) != 20 || values[i] != (CharacterLevelValue{Level: n, Value: want}) {
				t.Fatalf("row %d %s", n, id)
			}
		}
	}
}

func TestBarbarianExactSourceContract(t *testing.T) {
	// Each reviewed record is frozen in full, including source ownership,
	// reciprocal refs, every required key, option, operand and integer. The
	// readable contract below also makes the option/predicate/numeric assertions
	// reviewable without relying on presence-only tests or runtime execution.
	c := barbarianCatalogue(t)
	b, err := json.Marshal(c.Records)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "8f1caeb556f2d083841e6bfe769ffe2ed92228d64577236ca7a0f3dc08ff2795" {
		t.Fatal("exact source contract changed: review all source facts before updating digest")
	}
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		records[r.ID] = r
	}
	for slug, want := range map[string][]string{
		"skills":                          {"skills/animal-handling", "skills/athletics", "skills/intimidation", "skills/nature", "skills/perception", "skills/survival"},
		"primal-knowledge-skill":          {"skills/animal-handling", "skills/athletics", "skills/intimidation", "skills/nature", "skills/perception", "skills/survival"},
		"primal-knowledge-strength-check": {"skills/acrobatics", "skills/intimidation", "skills/perception", "skills/stealth", "skills/survival"},
		"starting-equipment":              {"barbarian/equipment/a", "barbarian/equipment/b"},
		"weapon-mastery":                  {"weapons/simple-or-martial-melee/mastery"},
		"barbarian-subclass":              {"barbarian/subclasses", "barbarian/subclasses/path-of-the-berserker"},
		"ability-score-improvement":       {"feats/ability-score-improvement", "feats/qualified"}, "epic-boon": {"feats/epic-boon", "feats/qualified"},
		"rage-extension":        {"barbarian/rage/attack-roll-against-enemy", "barbarian/rage/force-enemy-saving-throw", "barbarian/rage/bonus-action-extension"},
		"brutal-strike-effects": {"barbarian/brutal-strike/forceful-blow", "barbarian/brutal-strike/hamstring-blow"},
		"forceful-blow":         {"barbarian/brutal-strike/forceful-blow"}, "hamstring-blow": {"barbarian/brutal-strike/hamstring-blow"},
		"improved-brutal-strike-13": {"barbarian/brutal-strike/staggering-blow", "barbarian/brutal-strike/sundering-blow"},
	} {
		r := records[barbarianID+"choice/barbarian/"+slug]
		keys := []string{}
		for _, v := range want {
			keys = append(keys, barbarianKey+v)
		}
		count := 1
		if slug == "skills" || slug == "weapon-mastery" {
			count = 2
		}
		if r.Select != count || !reflect.DeepEqual(r.OptionKeys, keys) {
			t.Fatalf("choice %s", slug)
		}
	}
	for _, r := range c.Records {
		if r.Kind == "resource" && r.ID != barbarianID+"resource/barbarian/rage" {
			t.Fatal("numeric bonus modeled as resource")
		}
	}
	effects := records[barbarianID+"choice/barbarian/brutal-strike-effects"]
	for i, v := range effects.ValuesByLevel {
		want := 1
		if v.Level >= 17 {
			want = 2
		}
		if v.Level != i+9 || v.Value != want {
			t.Fatal("effect improvement")
		}
	}
	if len(effects.ValuesByLevel) != 12 {
		t.Fatal("effect count rows")
	}
}

func TestBarbarianGuardsCompositionAndNoSupport(t *testing.T) {
	repo := os.DirFS("../../..")
	tests := map[string]func(*CharacterCatalogue){
		"unknown grant": func(c *CharacterCatalogue) { c.Records[1].Progression.FeatureRefs = []string{"missing"} },
		"future grant": func(c *CharacterCatalogue) {
			c.Records[1].Progression.FeatureRefs = []string{barbarianID + "feature/barbarian/primal-champion"}
		},
		"wrong page":                  func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 30 },
		"unrequired numeric":          func(c *CharacterCatalogue) { c.Records[1].Parameters[barbarianKey+"unknown"] = 1 },
		"zero numeric":                func(c *CharacterCatalogue) { c.Records[1].Parameters[barbarianKey+"barbarian/rage/damage-bonus"] = 0 },
		"numeric without progression": func(c *CharacterCatalogue) { c.Records[1].Progression = nil },
		"invalid operand": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].Requirement.RecordRefs = []string{"missing"}
					break
				}
			}
		},
		"missing reciprocal": func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} },
		"wrong feature owner": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == barbarianID+"feature/barbarian/improved-brutal-strike-17" {
					c.Records[i].ParentID = barbarianID + "level/barbarian/13"
				}
			}
		},
		"count before feature": func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == barbarianID+"choice/barbarian/brutal-strike-effects" {
					c.Records[i].ValuesByLevel[0].Level = 1
				}
			}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			c := barbarianCatalogue(t)
			mutate(&c)
			if c.Validate(repo) == nil {
				t.Fatal("accepted malformed metadata")
			}
		})
	}
	c := barbarianCatalogue(t)
	c.Records[0].RequiredMechanics = append([]string{barbarianKey + "unknown/barbarian"}, c.Records[0].RequiredMechanics...)
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil || !strings.Contains(err.Error(), "unknown/barbarian") {
		t.Fatal("unknown mechanics enabled")
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCharacterCatalogue(strings.NewReader(strings.Replace(string(b), `"schema_version":1`, `"schema_version":1,"playable":true`, 1))); err == nil {
		t.Fatal("unknown field accepted")
	}
	c = barbarianCatalogue(t)
	for _, name := range []string{"character-advancement.json", "character-fighter.json", "character-champion.json", "character-cleric.json", "character-life-domain.json", "character-backgrounds.json", "character-fighting-styles.json"} {
		b, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		c.Records = append(c.Records, other.Records...)
	}
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if c.ValidateReadiness(repo) == nil {
		t.Fatal("catalogue became playable")
	}
	coverage, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if coverage.FullSupportEligible(repo) {
		t.Fatal("Full Rules promoted")
	}
}

func TestBarbarianExactNumbersPredicatesAndEquipment(t *testing.T) {
	c := barbarianCatalogue(t)
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		records[r.ID] = r
	}
	numbers := map[string]map[string]int{
		"class/barbarian":                             {"hit-die-sides": 12},
		"feature/barbarian/rage":                      {"rage/maximum-duration-minutes": 10, "rage/short-rest-recovery-uses": 1},
		"feature/barbarian/unarmored-defense":         {"unarmored-defense/base-ac": 10},
		"feature/barbarian/extra-attack":              {"extra-attack/attacks-per-action": 2},
		"feature/barbarian/fast-movement":             {"fast-movement/speed-increase-feet": 10},
		"feature/barbarian/instinctive-pounce":        {"instinctive-pounce/speed-divisor": 2},
		"feature/barbarian/brutal-strike":             {"brutal-strike/damage-dice": 1, "brutal-strike/damage-die-sides": 10, "brutal-strike/effect-count": 1, "brutal-strike/forceful-blow-push-feet": 15, "brutal-strike/forceful-blow-follow-speed-divisor": 2, "brutal-strike/hamstring-blow-speed-reduction-feet": 15},
		"feature/barbarian/relentless-rage":           {"relentless-rage/initial-and-reset-dc": 10, "relentless-rage/dc-increase": 5, "relentless-rage/hp-class-level-multiplier": 2},
		"feature/barbarian/improved-brutal-strike-13": {"brutal-strike/sundering-blow-attack-bonus": 5},
		"feature/barbarian/persistent-rage":           {"persistent-rage/duration-minutes": 10, "persistent-rage/initiative-recoveries-per-long-rest": 1},
		"feature/barbarian/improved-brutal-strike-17": {"brutal-strike/damage-dice": 2, "brutal-strike/damage-die-sides": 10, "brutal-strike/effect-count": 2},
		"feature/barbarian/primal-champion":           {"primal-champion/score-increase": 4, "primal-champion/score-maximum": 25},
	}
	for id, values := range numbers {
		want := map[string]int{}
		for key, v := range values {
			want[barbarianKey+"barbarian/"+key] = v
		}
		if !reflect.DeepEqual(records[barbarianID+id].Parameters, want) {
			t.Fatalf("numeric %s", id)
		}
	}
	// All predicates, including conditional selection and exact record operands.
	predicates := map[string]struct {
		owner, key string
		refs       []string
		parameters map[string]int
	}{
		"rage-entry":              {"feature/rage", "barbarian/rage/not-wearing-heavy-armor", []string{}, map[string]int{}},
		"unarmored-defense":       {"feature/unarmored-defense", "barbarian/unarmored-defense/not-wearing-any-armor", []string{}, map[string]int{}},
		"danger-sense":            {"feature/danger-sense", "barbarian/danger-sense/not-incapacitated", []string{}, map[string]int{}},
		"primal-knowledge-raging": {"choice/primal-knowledge-strength-check", "barbarian/rage/active", []string{"feature/rage"}, map[string]int{}},
		"ability-score-improvement-qualification": {"choice/ability-score-improvement", "prerequisite/feat-qualification", []string{}, map[string]int{}},
		"epic-boon-qualification":                 {"choice/epic-boon", "prerequisite/feat-qualification", []string{}, map[string]int{}},
		"fast-movement":                           {"feature/fast-movement", "barbarian/fast-movement/not-wearing-heavy-armor", []string{}, map[string]int{}},
		"brutal-strike":                           {"feature/brutal-strike", "barbarian/brutal-strike/reckless-attack-used-and-chosen-roll-not-disadvantaged", []string{"feature/reckless-attack"}, map[string]int{}},
		"forceful-blow-selected":                  {"choice/forceful-blow", "barbarian/brutal-strike/forceful-blow-selected", []string{"choice/brutal-strike-effects"}, map[string]int{}},
		"hamstring-blow-selected":                 {"choice/hamstring-blow", "barbarian/brutal-strike/hamstring-blow-selected", []string{"choice/brutal-strike-effects"}, map[string]int{}},
		"relentless-rage":                         {"feature/relentless-rage", "barbarian/relentless-rage/drop-to-zero-hp-while-raging-and-not-die-outright", []string{"feature/rage"}, map[string]int{"hit-points": 0}},
		"improved-brutal-strike-13":               {"choice/improved-brutal-strike-13", "barbarian/brutal-strike/uses-base-feature", []string{"feature/brutal-strike"}, map[string]int{}},
		"improved-brutal-strike-17":               {"feature/improved-brutal-strike-17", "barbarian/brutal-strike/uses-base-and-expanded-options", []string{"feature/brutal-strike", "feature/improved-brutal-strike-13"}, map[string]int{}},
		"indomitable-might":                       {"feature/indomitable-might", "barbarian/indomitable-might/total-less-than-strength-score", []string{}, map[string]int{}},
	}
	expand := func(short string) string {
		parts := strings.SplitN(short, "/", 2)
		return barbarianID + parts[0] + "/barbarian/" + parts[1]
	}
	count := 0
	for _, r := range c.Records {
		if r.Kind == "prerequisite" {
			count++
		}
	}
	if count != len(predicates) {
		t.Fatal("unexpected predicate")
	}
	for slug, p := range predicates {
		r := records[barbarianID+"prerequisite/barbarian/"+slug]
		refs := []string{}
		for _, ref := range p.refs {
			refs = append(refs, expand(ref))
		}
		want := &CharacterRequirement{Key: barbarianKey + p.key, RecordRefs: refs, Parameters: p.parameters}
		if r.ParentID != expand(p.owner) || !reflect.DeepEqual(r.Requirement, want) || !reflect.DeepEqual(r.RequiredMechanics, []string{barbarianKey + p.key}) {
			t.Fatalf("predicate %s", slug)
		}
	}
	equipment := records[barbarianID+"choice/barbarian/starting-equipment"]
	expected := []map[string]int{{"greataxe": 1, "handaxe": 4, "explorers-pack": 1, "gp": 15}, {"gp": 75}}
	if len(equipment.Options) != 2 {
		t.Fatal("equipment alternatives")
	}
	for i, q := range expected {
		want := map[string]int{}
		for k, v := range q {
			want[barbarianKey+"equipment/"+k] = v
		}
		if equipment.Options[i].Key != equipment.OptionKeys[i] || !reflect.DeepEqual(equipment.Options[i].Quantities, want) {
			t.Fatalf("equipment %d", i)
		}
	}
}
