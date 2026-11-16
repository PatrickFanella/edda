package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const clericID = "srd-5.2.1:character/"
const clericKey = "srd-5.2.1:mechanic/"

func clericCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-cleric.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestClericCatalogueSourceAndRows(t *testing.T) {
	c := clericCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	records := map[string]CharacterRecord{}
	identities := map[string]int{}
	canonical := 0
	for _, r := range c.Records {
		records[r.ID] = r
		if r.Kind == "class" || r.Kind == "feature" {
			canonical++
			identities[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
		if r.Kind == "subclass" || r.Kind == "advancement" || !strings.HasPrefix(r.ID, clericID+r.Kind+"/cleric") {
			t.Fatalf("scope widened: %s", r.ID)
		}
	}
	expected := map[string]int{"srd-5.2.1:planning/classes/cleric": 36}
	features := map[int][]string{1: {"spellcasting", "divine-order"}, 2: {"channel-divinity"}, 3: {"cleric-subclass"}, 4: {"ability-score-improvement"}, 5: {"sear-undead"}, 7: {"blessed-strikes"}, 10: {"divine-intervention"}, 14: {"improved-blessed-strikes"}, 19: {"epic-boon"}, 20: {"greater-divine-intervention"}}
	for n, slugs := range features {
		for _, slug := range slugs {
			page := 37
			if slug == "spellcasting" {
				page = 36
			}
			if n >= 7 {
				page = 38
			}
			expected[fmt.Sprintf("srd-5.2.1:planning/class-features/cleric-level-%d-%s", n, slug)] = page
		}
	}
	if canonical != 12 || !reflect.DeepEqual(identities, expected) {
		t.Fatalf("identities/pages: %v", identities)
	}
	// Independent transcription of p36: proficiency, Channel Divinity, cantrips,
	// prepared spells, then all nine spell-slot columns (zero means absent).
	rows := [20][13]int{
		{2, 0, 3, 4, 2, 0, 0, 0, 0, 0, 0, 0, 0}, {2, 2, 3, 5, 3, 0, 0, 0, 0, 0, 0, 0, 0},
		{2, 2, 3, 6, 4, 2, 0, 0, 0, 0, 0, 0, 0}, {2, 2, 4, 7, 4, 3, 0, 0, 0, 0, 0, 0, 0},
		{3, 2, 4, 9, 4, 3, 2, 0, 0, 0, 0, 0, 0}, {3, 3, 4, 10, 4, 3, 3, 0, 0, 0, 0, 0, 0},
		{3, 3, 4, 11, 4, 3, 3, 1, 0, 0, 0, 0, 0}, {3, 3, 4, 12, 4, 3, 3, 2, 0, 0, 0, 0, 0},
		{4, 3, 4, 14, 4, 3, 3, 3, 1, 0, 0, 0, 0}, {4, 3, 5, 15, 4, 3, 3, 3, 2, 0, 0, 0, 0},
		{4, 3, 5, 16, 4, 3, 3, 3, 2, 1, 0, 0, 0}, {4, 3, 5, 16, 4, 3, 3, 3, 2, 1, 0, 0, 0},
		{5, 3, 5, 17, 4, 3, 3, 3, 2, 1, 1, 0, 0}, {5, 3, 5, 17, 4, 3, 3, 3, 2, 1, 1, 0, 0},
		{5, 3, 5, 18, 4, 3, 3, 3, 2, 1, 1, 1, 0}, {5, 3, 5, 18, 4, 3, 3, 3, 2, 1, 1, 1, 0},
		{6, 3, 5, 19, 4, 3, 3, 3, 2, 1, 1, 1, 1}, {6, 4, 5, 20, 4, 3, 3, 3, 3, 1, 1, 1, 1},
		{6, 4, 5, 21, 4, 3, 3, 3, 3, 2, 1, 1, 1}, {6, 4, 5, 22, 4, 3, 3, 3, 3, 2, 2, 1, 1},
	}
	count := func(id string, n int) int {
		for _, v := range records[clericID+id].ValuesByLevel {
			if v.Level == n {
				return v.Value
			}
		}
		return 0
	}
	levels := 0
	for _, r := range c.Records {
		if r.Kind == "level" {
			levels++
		}
	}
	if levels != 20 {
		t.Fatal("missing class rows")
	}
	for i, row := range rows {
		n := i + 1
		r := records[fmt.Sprintf("%slevel/cleric/%d", clericID, n)]
		p := r.Progression
		if p == nil || p.ProficiencyBonus != row[0] || r.Sources[0].Page != 36 {
			t.Fatalf("row %d", n)
		}
		grants := features[n]
		if n == 8 || n == 12 || n == 16 {
			grants = []string{"ability-score-improvement"}
		}
		want := []string{}
		for _, slug := range grants {
			want = append(want, clericID+"feature/cleric/"+slug)
		}
		if !reflect.DeepEqual(p.FeatureRefs, want) {
			t.Fatalf("row %d feature refs", n)
		}
		keys := []string{}
		if n == 6 || n == 17 {
			keys = []string{clericKey + "cleric/subclass-feature"}
		}
		if !reflect.DeepEqual(p.FeatureKeys, keys) {
			t.Fatalf("row %d subclass slot", n)
		}
		ids := []string{"resource/cleric/channel-divinity", "choice/cleric/cantrips", "choice/cleric/prepared-spells"}
		for j := 1; j <= 9; j++ {
			ids = append(ids, fmt.Sprintf("resource/cleric/spell-slot-%d", j))
		}
		for j, id := range ids {
			if got := count(id, n); got != row[j+1] {
				t.Fatalf("row %d %s=%d want %d", n, id, got, row[j+1])
			}
		}
		uses := 0
		if n >= 10 {
			uses = 1
		}
		if count("resource/cleric/divine-intervention", n) != uses {
			t.Fatalf("Intervention row %d", n)
		}
	}
}

func TestClericNestedChoicesAndNumbers(t *testing.T) {
	c := clericCatalogue(t)
	records := map[string]CharacterRecord{}
	for _, r := range c.Records {
		records[r.ID] = r
	}
	choice := func(slug string, count int, options ...string) CharacterRecord {
		t.Helper()
		r := records[clericID+"choice/cleric/"+slug]
		want := []string{}
		for _, s := range options {
			want = append(want, clericKey+s)
		}
		if r.Select != count || !reflect.DeepEqual(r.OptionKeys, want) {
			t.Fatalf("choice %s: %+v", slug, r)
		}
		return r
	}
	choice("skills", 2, "skills/history", "skills/insight", "skills/medicine", "skills/persuasion", "skills/religion")
	choice("cantrips", 3, "cleric/spell-list/p38-40/cantrips")
	choice("prepared-spells", 4, "cleric/spell-list/p38-40/level-1-plus")
	order := choice("divine-order", 1, "cleric/divine-order/protector", "cleric/divine-order/thaumaturge")
	pro := choice("protector", 1, "cleric/divine-order/protector/grant-martial-weapons-and-heavy-armor")
	th := choice("thaumaturge", 1, "cleric/spell-list/p38-40/cantrips")
	ch := choice("channel-divinity", 1, "cleric/channel-divinity/divine-spark", "cleric/channel-divinity/turn-undead")
	spark := choice("divine-spark", 1, "cleric/divine-spark/healing", "damage/necrotic", "damage/radiant")
	bl := choice("blessed-strikes", 1, "cleric/blessed-strikes/divine-strike", "cleric/blessed-strikes/potent-spellcasting")
	ds := choice("divine-strike", 1, "damage/necrotic", "damage/radiant")
	for _, option := range []struct {
		child, parent CharacterRecord
		predicate     string
	}{
		{pro, order, "cleric/divine-order/protector-selected"},
		{th, order, "cleric/divine-order/thaumaturge-selected"},
		{spark, ch, "cleric/channel-divinity/divine-spark-selected"},
		{ds, bl, "cleric/blessed-strikes/divine-strike-selected"},
	} {
		child, parent := option.child, option.parent
		if child.ParentID != parent.ID || len(child.PrerequisiteRefs) != 1 {
			t.Fatal("flattened nested option")
		}
		prerequisite := records[child.PrerequisiteRefs[0]]
		req := prerequisite.Requirement
		if req == nil || req.Key != clericKey+option.predicate || !reflect.DeepEqual(prerequisite.RequiredMechanics, []string{clericKey + option.predicate}) || !reflect.DeepEqual(req.RecordRefs, []string{parent.ID}) {
			t.Fatalf("lost conditional selection for %s", child.ID)
		}
	}
	choice("cleric-subclass", 1, "cleric/subclasses", "cleric/subclasses/life-domain")
	for _, slug := range []string{"ability-score-improvement", "epic-boon"} {
		r := choice(slug, 1, "feats/"+slug, "feats/qualified")
		if len(r.PrerequisiteRefs) != 1 || records[r.PrerequisiteRefs[0]].Requirement.Key != clericKey+"prerequisite/feat-qualification" {
			t.Fatal("lost feat qualification")
		}
	}
	choice("divine-intervention-spell", 1, "cleric/spell-list/p38-40/level-5-or-lower")
	req := records[clericID+"prerequisite/cleric/divine-intervention-non-reaction"].Requirement
	if req.Key != clericKey+"cleric/divine-intervention/non-reaction-casting-time" || req.Parameters["maximum-spell-level"] != 5 {
		t.Fatal("intervention eligibility")
	}
	equipment := choice("starting-equipment", 1, "cleric/equipment/a", "cleric/equipment/b")
	expected := []map[string]int{{"equipment/chain-shirt": 1, "equipment/shield": 1, "equipment/mace": 1, "equipment/holy-symbol": 1, "equipment/priests-pack": 1, "currency/gp": 7}, {"currency/gp": 110}}
	if len(equipment.Options) != 2 {
		t.Fatal("equipment alternatives")
	}
	for i, q := range expected {
		want := map[string]int{}
		for k, v := range q {
			want[clericKey+k] = v
		}
		if !reflect.DeepEqual(equipment.Options[i].Quantities, want) {
			t.Fatal("equipment quantities")
		}
	}
	params := map[string]map[string]int{
		"class/cleric":                               {"hit-die-sides": 8},
		"feature/cleric/divine-order":                {"thaumaturge-extra-cantrips": 1, "thaumaturge-minimum-check-bonus": 1},
		"feature/cleric/channel-divinity":            {"divine-spark-range-feet": 30, "divine-spark-die-sides": 8, "turn-undead-range-feet": 30, "turn-undead-duration-minutes": 1, "short-rest-recovered-uses": 1, "divine-spark-dice-at-level-2": 1, "divine-spark-dice-at-level-7": 2, "divine-spark-dice-at-level-13": 3, "divine-spark-dice-at-level-18": 4},
		"feature/cleric/sear-undead":                 {"sear-undead-die-sides": 8, "sear-undead-minimum-dice": 1},
		"feature/cleric/blessed-strikes":             {"divine-strike-dice": 1, "divine-strike-die-sides": 8},
		"feature/cleric/improved-blessed-strikes":    {"divine-strike-dice": 2, "divine-strike-die-sides": 8, "potent-range-feet": 60, "potent-wisdom-multiplier": 2},
		"feature/cleric/greater-divine-intervention": {"wish-recovery-dice": 2, "wish-recovery-die-sides": 4},
	}
	for id, want := range params {
		got := records[clericID+id].Parameters
		expected := map[string]int{}
		for k, v := range want {
			expected[clericKey+"cleric/"+k] = v
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("parameters %s", id)
		}
	}
	// Required semantics are metadata too: deleting an opaque restriction must
	// not turn this source transcription into an apparently complete fixture.
	required := map[string][]string{
		"feature/cleric/spellcasting":                {"cleric/spellcasting/wisdom", "cleric/spellcasting/holy-symbol-focus-p97", "cleric/spellcasting/cleric-spell-list-p38-40", "cleric/spellcasting/replace-one-cantrip-on-cleric-level", "cleric/spellcasting/long-rest-all-slots", "cleric/spellcasting/long-rest-replace-prepared-spells", "cleric/spellcasting/always-prepared-excluded-from-count"},
		"choice/cleric/protector":                    {"weapons/martial", "armor/heavy"},
		"choice/cleric/thaumaturge":                  {"cleric/divine-order/thaumaturge/extra-cantrip", "cleric/divine-order/thaumaturge/arcana-religion-wisdom-modifier-minimum-1"},
		"choice/cleric/divine-spark":                 {"cleric/divine-spark/magic-action", "cleric/divine-spark/holy-symbol-other-visible-creature-30-feet", "cleric/divine-spark/d8-plus-wisdom-modifier", "cleric/divine-spark/constitution-save-half-round-down"},
		"resource/cleric/channel-divinity":           {"cleric/channel-divinity/short-rest-recover-one", "cleric/channel-divinity/long-rest-recover-all", "cleric/channel-divinity/save-dc-spellcasting"},
		"feature/cleric/channel-divinity":            {"cleric/turn-undead/magic-action-holy-symbol", "cleric/turn-undead/chosen-undead-within-30-feet", "cleric/turn-undead/wisdom-save", "cleric/turn-undead/frightened-incapacitated-1-minute", "cleric/turn-undead/move-away-on-turns", "cleric/turn-undead/ends-on-damage-or-cleric-incapacitated-or-dead"},
		"feature/cleric/sear-undead":                 {"cleric/sear-undead/on-turn-undead-failed-save", "cleric/sear-undead/d8-count-wisdom-modifier-minimum-1", "cleric/sear-undead/radiant-damage", "cleric/sear-undead/damage-does-not-end-turn-effect"},
		"choice/cleric/divine-strike":                {"cleric/divine-strike/once-on-own-turn-weapon-attack-roll-hit", "cleric/divine-strike/extra-1d8"},
		"feature/cleric/blessed-strikes":             {"cleric/potent-spellcasting/add-wisdom-modifier-to-cleric-cantrip-damage"},
		"feature/cleric/improved-blessed-strikes":    {"cleric/improved-blessed-strikes/improves-only-chosen-option", "cleric/improved-blessed-strikes/divine-strike-extra-2d8", "cleric/improved-blessed-strikes/potent-on-cleric-cantrip-damage", "cleric/improved-blessed-strikes/potent-self-or-other-within-60-feet", "cleric/improved-blessed-strikes/potent-temporary-hp-twice-wisdom-modifier"},
		"choice/cleric/divine-intervention-spell":    {"cleric/divine-intervention/magic-action-cast-same-action", "cleric/divine-intervention/no-slot-or-material-components"},
		"resource/cleric/divine-intervention":        {"cleric/divine-intervention/long-rest-recovery"},
		"feature/cleric/greater-divine-intervention": {"cleric/greater-divine-intervention/wish-option", "cleric/greater-divine-intervention/wish-locks-divine-intervention-2d4-long-rests"},
	}
	for id, keys := range required {
		got := records[clericID+id].RequiredMechanics
		for _, key := range keys {
			found := false
			for _, k := range got {
				if k == clericKey+key {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s missing %s", id, key)
			}
		}
	}
	multiclass := records[clericID+"prerequisite/cleric/multiclass-eligibility"].Requirement
	if multiclass == nil || multiclass.Parameters["minimum-wisdom"] != 13 || multiclass.Parameters["minimum-current-class-primary-abilities"] != 13 {
		t.Fatal("multiclass eligibility")
	}

	var view struct {
		Catalogue string   `json:"catalogue"`
		RecordIDs []string `json:"record_ids"`
	}
	b, err := os.ReadFile("testdata/character-cleric1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &view); err != nil {
		t.Fatal(err)
	}
	if view.Catalogue != "character-cleric.json" {
		t.Fatal("wrong canonical catalogue")
	}
	seen := map[string]bool{}
	for _, id := range view.RecordIDs {
		if records[id].ID == "" || seen[id] {
			t.Fatal("duplicate/missing composition record")
		}
		seen[id] = true
	}
	for _, id := range []string{"class/cleric", "feature/cleric/spellcasting", "feature/cleric/divine-order", "choice/cleric/protector", "choice/cleric/thaumaturge"} {
		if !seen[clericID+id] {
			t.Fatal("incomplete level-one composition")
		}
	}
}

func TestClericGuardsAndReadiness(t *testing.T) {
	repo := os.DirFS("../../..")
	c := clericCatalogue(t)
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("catalogue enabled support")
	}
	for _, r := range c.Records {
		for _, key := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(key) {
				t.Fatalf("qualified %s", key)
			}
		}
	}
	// Unknown keys remain valid metadata but never acquire executable support.
	c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, clericKey+"unknown/future")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("unknown key became available")
	}
	tests := []struct {
		name   string
		mutate func(*CharacterCatalogue)
	}{
		{"wrong source page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 37 }},
		{"future grant", func(c *CharacterCatalogue) {
			c.Records[1].Progression.FeatureRefs = []string{clericID + "feature/cleric/greater-divine-intervention"}
		}},
		{"missing nested owner link", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].ID == clericID+"choice/cleric/divine-order" {
					c.Records[i].ChoiceRefs = []string{}
				}
			}
		}},
		{"cycle", func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.ID == clericID+"choice/cleric/protector" {
					r.ParentID = r.ID
					r.ChoiceRefs = []string{r.ID}
				}
			}
		}},
		{"missing prerequisite operand", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].Requirement.RecordRefs = []string{clericID + "choice/cleric/missing"}
					break
				}
			}
		}},
		{"count before owning feature", func(c *CharacterCatalogue) {
			for i := range c.Records {
				r := &c.Records[i]
				if r.ID == clericID+"resource/cleric/channel-divinity" {
					r.ValuesByLevel[0].Level = 1
				}
			}
		}},
		{"unrequired numeric key", func(c *CharacterCatalogue) { c.Records[0].Parameters[clericKey+"unknown/value"] = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := clericCatalogue(t)
			tt.mutate(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("accepted malformed metadata")
			}
		})
	}
	b, err := os.ReadFile("testdata/character-cleric.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCharacterCatalogue(strings.NewReader(strings.Replace(string(b), `"kind": "class"`, `"kind": "class", "playable": true`, 1))); err == nil {
		t.Fatal("data-supplied support accepted")
	}
}
