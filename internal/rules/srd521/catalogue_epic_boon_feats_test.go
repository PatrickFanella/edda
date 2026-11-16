package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const epicID = "srd-5.2.1:character/"
const epicKey = "srd-5.2.1:mechanic/"

func epicBoonsCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-epic-boon-feats.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// Independent Go transcription of the entire p88 graph, including empty fields.
// The oracle never reads the catalogue or the Python comparator.
func expectedEpicBoons() []CharacterRecord {
	var records []CharacterRecord
	keys := func(parts ...string) []string {
		result := []string{}
		for _, p := range parts {
			result = append(result, epicKey+p)
		}
		return result
	}
	add := func(kind, slug, name, parent string, mechanics []string) int {
		i := len(records)
		records = append(records, CharacterRecord{ID: epicID + kind + "/" + slug, Kind: kind, Name: name, ParentID: parent,
			Sources:    []CharacterSource{{InventoryID: "srd-5.2.1:planning/feats/" + strings.Split(slug, "/")[0], Page: 88}},
			ChoiceRefs: []string{}, ResourceRefs: []string{}, PrerequisiteRefs: []string{}, OptionKeys: []string{}, RequiredMechanics: keys(mechanics...)})
		return i
	}
	pred := func(owner int, part, key string, params map[string]int, operands ...string) string {
		slug := strings.SplitN(strings.TrimPrefix(records[owner].ID, epicID), "/", 2)[1]
		name := strings.ReplaceAll(part, "-", " ")
		name = strings.ToUpper(name[:1]) + name[1:]
		i := add("prerequisite", slug+"/"+part, name, records[owner].ID, []string{"prerequisite/" + key})
		if params == nil {
			params = map[string]int{}
		}
		refs := append([]string{}, operands...)
		records[i].Requirement = &CharacterRequirement{Key: epicKey + "prerequisite/" + key, RecordRefs: refs, Parameters: params}
		records[owner].PrerequisiteRefs = append(records[owner].PrerequisiteRefs, records[i].ID)
		return records[i].ID
	}
	for _, tc := range []struct {
		slug, name          string
		abilities, benefits []string
		numbers             map[string]int
	}{
		{"boon-of-combat-prowess", "Boon of Combat Prowess", nil, []string{"peerless-aim/optional-replace-miss-with-hit", "peerless-aim/recover-at-start-of-your-next-turn"}, nil},
		{"boon-of-dimensional-travel", "Boon of Dimensional Travel", nil, []string{"blink-steps/optional-teleport-immediately-after-attack-or-magic-action", "blink-steps/destination-unoccupied-space-you-can-see", "blink-steps/maximum-distance-feet"}, map[string]int{"blink-steps/maximum-distance-feet": 30}},
		{"boon-of-fate", "Boon of Fate", nil, []string{"improve-fate/optional-roll-and-apply-total-as-bonus-or-penalty-to-triggering-d20-roll", "improve-fate/dice-count", "improve-fate/die-sides", "improve-fate/other-creature-range-feet", "improve-fate/recover-on-roll-initiative-or-finish-short-or-long-rest"}, map[string]int{"improve-fate/dice-count": 2, "improve-fate/die-sides": 4, "improve-fate/other-creature-range-feet": 60}},
		{"boon-of-irresistible-offense", "Boon of Irresistible Offense", []string{"strength", "dexterity"}, []string{"overcome-defenses/your-bludgeoning-piercing-slashing-damage-always-ignores-resistance", "overwhelming-strike/optional-extra-damage-to-attack-target", "overwhelming-strike/extra-damage-equals-selected-ability-score-not-modifier", "overwhelming-strike/extra-damage-type-same-as-attack", "overwhelming-strike/attack-d20-result"}, map[string]int{"overwhelming-strike/attack-d20-result": 20}},
		{"boon-of-spell-recall", "Boon of Spell Recall", []string{"intelligence", "wisdom", "charisma"}, []string{"free-casting/roll-on-every-spell-cast-with-level-one-to-four-slot", "free-casting/dice-count", "free-casting/die-sides", "free-casting/do-not-expend-slot-only-if-roll-equals-slot-level"}, map[string]int{"free-casting/dice-count": 1, "free-casting/die-sides": 4}},
		{"boon-of-the-night-spirit", "Boon of the Night Spirit", nil, []string{"merge-with-shadows/optional-give-self-invisible-condition-as-bonus-action-in-dim-light-or-darkness", "merge-with-shadows/ends-immediately-after-you-take-action-or-bonus-action-or-reaction", "shadowy-form/resistance-to-all-damage-except-psychic-and-radiant-in-dim-light-or-darkness"}, nil},
		{"boon-of-truesight", "Boon of Truesight", nil, []string{"truesight/have-truesight", "truesight/range-feet"}, map[string]int{"truesight/range-feet": 60}},
	} {
		base := "feats/" + tc.slug + "/"
		mechanics := []string{"feats/category/epic-boon", "feats/maximum-times-taken", base + "ability-score-increase", base + "maximum-ability-score"}
		for _, benefit := range tc.benefits {
			mechanics = append(mechanics, base+benefit)
		}
		f := add("feat", tc.slug, tc.name, "", mechanics)
		records[f].Parameters = map[string]int{epicKey + "feats/maximum-times-taken": 1, epicKey + base + "ability-score-increase": 1, epicKey + base + "maximum-ability-score": 30}
		for k, v := range tc.numbers {
			records[f].Parameters[epicKey+base+k] = v
		}
		pred(f, "level", "minimum-character-level", map[string]int{"minimum-level": 19})
		abilities := tc.abilities
		if abilities == nil {
			abilities = []string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}
		}
		ch := add("choice", tc.slug+"/ability-score", "Ability Score Increase", records[f].ID, []string{base + "increase-one-selected-ability", base + "subject-to-maximum-ability-score"})
		records[ch].Select = 1
		for _, a := range abilities {
			records[ch].OptionKeys = append(records[ch].OptionKeys, epicKey+"abilities/"+a)
		}
		records[f].ChoiceRefs = append(records[f].ChoiceRefs, records[ch].ID)
		switch tc.slug {
		case "boon-of-combat-prowess":
			pred(f, "miss", "attack-roll-missed", nil)
		case "boon-of-dimensional-travel":
			pred(f, "action", "immediately-after-you-take-attack-or-magic-action", nil)
			pred(f, "destination", "teleport-destination-unoccupied-and-visible-to-you", nil)
		case "boon-of-fate":
			pred(f, "d20-test", "you-or-another-creature-in-range-succeeds-or-fails-d20-test", map[string]int{"other-creature-range-feet": 60})
		case "boon-of-irresistible-offense":
			pred(f, "attack-roll", "attack-roll-d20-equals", map[string]int{"d20-result": 20})
			pred(f, "damage-amount", "extra-damage-equals-ability-score-increased-by-this-feat", nil, records[ch].ID)
		case "boon-of-spell-recall":
			pred(f, "spellcasting", "has-spellcasting-feature", nil)
			slot := pred(f, "slot", "cast-spell-with-slot-level-in-range", map[string]int{"minimum-slot-level": 1, "maximum-slot-level": 4})
			pred(f, "roll-match", "die-result-equals-triggering-spell-slot-level", nil, slot)
		case "boon-of-the-night-spirit":
			pred(f, "light", "within-dim-light-or-darkness", nil)
			pred(f, "invisibility-end", "immediately-after-you-take-action-or-bonus-action-or-reaction", nil)
		}
		if tc.slug == "boon-of-combat-prowess" || tc.slug == "boon-of-fate" {
			benefit, recovery := "peerless-aim", "start-of-your-next-turn"
			if tc.slug == "boon-of-fate" {
				benefit, recovery = "improve-fate", "roll-initiative-or-finish-short-or-long-rest"
			}
			words := strings.Split(benefit, "-")
			for i, w := range words {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
			r := add("resource", tc.slug+"/"+benefit, strings.Join(words, " "), records[f].ID, []string{base + benefit + "/uses", base + benefit + "/recover-on-" + recovery})
			records[r].Capacity = 1
			records[f].ResourceRefs = append(records[f].ResourceRefs, records[r].ID)
		}
	}
	return records
}

func checkEpicBoonsSource(c CharacterCatalogue) error {
	if !reflect.DeepEqual(c.Records, expectedEpicBoons()) {
		return fmt.Errorf("Epic Boons differ from complete page 88 graph")
	}
	return nil
}

func TestEpicBoonsExactSource(t *testing.T) {
	c := epicBoonsCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if err := checkEpicBoonsSource(c); err != nil {
		t.Fatal(err)
	}
	for _, r := range c.Records {
		for _, k := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(k) {
				t.Fatalf("qualified %s", k)
			}
		}
	}
}

// Every mutation here still passes the generic schema. Only complete source
// equality can reject these contrary mechanics, operands and invented scaling.
func TestEpicBoonsContrarySource(t *testing.T) {
	original := epicBoonsCatalogue(t)
	for i, r := range original.Records {
		cases := map[string]func(*CharacterRecord){
			"invented mechanic": func(r *CharacterRecord) {
				r.RequiredMechanics = append(r.RequiredMechanics, epicKey+"invented/scaling")
			},
			"renamed identity": func(r *CharacterRecord) { r.Name += " contrary" },
		}
		for j := range r.RequiredMechanics {
			cases[fmt.Sprintf("mechanic-%d", j)] = func(r *CharacterRecord) {
				old := r.RequiredMechanics[j]
				r.RequiredMechanics[j] = epicKey + "contrary/mechanic"
				if value, ok := r.Parameters[old]; ok {
					delete(r.Parameters, old)
					r.Parameters[r.RequiredMechanics[j]] = value
				}
				if r.Requirement != nil {
					r.Requirement.Key = r.RequiredMechanics[j]
				}
			}
		}
		for k := range r.Parameters {
			cases["number-"+k] = func(r *CharacterRecord) { r.Parameters[k]++ }
		}
		if r.Kind == "resource" {
			cases["capacity"] = func(r *CharacterRecord) { r.Capacity++ }
		}
		if r.Kind == "choice" {
			cases["ability alternatives"] = func(r *CharacterRecord) { r.OptionKeys = r.OptionKeys[:1] }
			cases["ability count"] = func(r *CharacterRecord) { r.Select = 2 }
			cases["invented scaling"] = func(r *CharacterRecord) {
				r.ValuesByLevel = []CharacterLevelValue{{Level: 19, Value: 1}, {Level: 20, Value: 2}}
			}
		}
		if r.Requirement != nil {
			for k := range r.Requirement.Parameters {
				cases["predicate number-"+k] = func(r *CharacterRecord) { r.Requirement.Parameters[k]++ }
			}
			cases["invented predicate body"] = func(r *CharacterRecord) { r.Requirement.Parameters["invented-scaling"] = 2 }
			cases["operand"] = func(r *CharacterRecord) { r.Requirement.RecordRefs = []string{original.Records[0].ID} }
			if len(r.Requirement.RecordRefs) > 0 {
				cases["missing operand"] = func(r *CharacterRecord) { r.Requirement.RecordRefs = []string{} }
			}
		}
		for name, mutate := range cases {
			t.Run(fmt.Sprintf("%d/%s", i, name), func(t *testing.T) {
				c := epicBoonsCatalogue(t)
				mutate(&c.Records[i])
				if err := c.Validate(os.DirFS("../../..")); err != nil {
					t.Fatalf("contrary fixture must be structurally valid: %v", err)
				}
				if err := checkEpicBoonsSource(c); err == nil {
					t.Fatal("contrary source accepted")
				}
			})
		}
	}
	// Move a valid subtree to another owner, updating every reciprocal link and
	// source. The schema accepts the graph, but the source oracle must reject it.
	c := epicBoonsCatalogue(t)
	ch := c.Records[2]
	c.Records[0].ChoiceRefs = []string{}
	c.Records[5].ChoiceRefs = append(c.Records[5].ChoiceRefs, ch.ID)
	c.Records[2].ParentID = c.Records[5].ID
	c.Records[2].Sources = c.Records[5].Sources
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if err := checkEpicBoonsSource(c); err == nil {
		t.Fatal("changed ownership accepted")
	}
}

// Plausible contrary readings supplement exhaustive field mutations.
func TestEpicBoonsContraryBenefitReadings(t *testing.T) {
	for _, tc := range []struct{ name, from, to string }{
		{"hit instead of miss", "prerequisite/attack-roll-missed", "prerequisite/attack-roll-hit"},
		{"two actions required", "prerequisite/immediately-after-you-take-attack-or-magic-action", "prerequisite/immediately-after-you-take-attack-and-magic-actions"},
		{"fate only succeeds", "prerequisite/you-or-another-creature-in-range-succeeds-or-fails-d20-test", "prerequisite/you-or-another-creature-in-range-succeeds-d20-test"},
		{"offense ignores immunity", "feats/boon-of-irresistible-offense/overcome-defenses/your-bludgeoning-piercing-slashing-damage-always-ignores-resistance", "feats/boon-of-irresistible-offense/overcome-defenses/your-damage-ignores-immunity"},
		{"ability modifier", "feats/boon-of-irresistible-offense/overwhelming-strike/extra-damage-equals-selected-ability-score-not-modifier", "feats/boon-of-irresistible-offense/overwhelming-strike/extra-damage-equals-ability-modifier"},
		{"spell recall result less than slot", "prerequisite/die-result-equals-triggering-spell-slot-level", "prerequisite/die-result-less-than-triggering-spell-slot-level"},
		{"night requires both light states", "prerequisite/within-dim-light-or-darkness", "prerequisite/within-dim-light-and-darkness"},
		{"night loses reaction termination", "prerequisite/immediately-after-you-take-action-or-bonus-action-or-reaction", "prerequisite/after-you-take-action-or-bonus-action"},
		{"night resists psychic", "feats/boon-of-the-night-spirit/shadowy-form/resistance-to-all-damage-except-psychic-and-radiant-in-dim-light-or-darkness", "feats/boon-of-the-night-spirit/shadowy-form/resistance-to-all-damage-except-radiant-in-dim-light-or-darkness"},
		{"combat recovers on any turn", "feats/boon-of-combat-prowess/peerless-aim/recover-at-start-of-your-next-turn", "feats/boon-of-combat-prowess/peerless-aim/recover-at-start-of-any-turn"},
		{"fate only long rest recovery", "feats/boon-of-fate/improve-fate/recover-on-roll-initiative-or-finish-short-or-long-rest", "feats/boon-of-fate/improve-fate/recover-on-long-rest"},
		{"night involuntary invisibility", "feats/boon-of-the-night-spirit/merge-with-shadows/optional-give-self-invisible-condition-as-bonus-action-in-dim-light-or-darkness", "feats/boon-of-the-night-spirit/merge-with-shadows/automatically-invisible-in-dim-light-or-darkness"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := epicBoonsCatalogue(t)
			found := false
			for i := range c.Records {
				r := &c.Records[i]
				for j, k := range r.RequiredMechanics {
					if k == epicKey+tc.from {
						r.RequiredMechanics[j] = epicKey + tc.to
						found = true
					}
				}
				if r.Requirement != nil && r.Requirement.Key == epicKey+tc.from {
					r.Requirement.Key = epicKey + tc.to
				}
			}
			if !found {
				t.Fatal("fixture failed to mutate source")
			}
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatal(err)
			}
			if err := checkEpicBoonsSource(c); err == nil {
				t.Fatal("contrary benefit reading accepted")
			}
		})
	}
}

func TestEpicBoonsCompositionAndGuards(t *testing.T) {
	repo := os.DirFS("../../..")
	c := epicBoonsCatalogue(t)
	for _, name := range []string{"advancement", "backgrounds", "cleric", "fighter", "champion", "life-domain", "barbarian", "berserker", "species", "origin-feats", "general-feats", "fighting-styles"} {
		b, err := os.ReadFile("testdata/character-" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateSource(other.Source); err != nil {
			t.Fatal(err)
		}
		c.Records = append(c.Records, other.Records...)
	}
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("metadata ready")
	}
	for name, mutate := range map[string]func(*CharacterCatalogue){
		"duplicate":   func(c *CharacterCatalogue) { c.Records = append(c.Records, c.Records[0]) },
		"page":        func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 87 },
		"pin":         func(c *CharacterCatalogue) { c.Source.SHA256 = "bad" },
		"license":     func(c *CharacterCatalogue) { c.Source.License = "MIT" },
		"attribution": func(c *CharacterCatalogue) { c.Source.Attribution = "" },
		"adaptation":  func(c *CharacterCatalogue) { c.Source.Changes = "" },
		"family": func(c *CharacterCatalogue) {
			c.Records[0].Sources[0] = CharacterSource{InventoryID: "srd-5.2.1:planning/backgrounds/acolyte", Page: 83}
		},
		"ownership":        func(c *CharacterCatalogue) { c.Records[2].ParentID = c.Records[5].ID },
		"source ownership": func(c *CharacterCatalogue) { c.Records[2].Sources = c.Records[5].Sources },
		"orphan":           func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} },
		"count":            func(c *CharacterCatalogue) { c.Records[2].Select = 0 },
		"malformed":        func(c *CharacterCatalogue) { c.Records[0].RequiredMechanics[0] = "bad key" },
		"predicate key":    func(c *CharacterCatalogue) { c.Records[1].Requirement.Key = epicKey + "unknown" },
		"operand":          func(c *CharacterCatalogue) { c.Records[1].Requirement.RecordRefs = []string{epicID + "feat/missing"} },
		"duplicate operand": func(c *CharacterCatalogue) {
			c.Records[1].Requirement.RecordRefs = []string{c.Records[0].ID, c.Records[0].ID}
		},
		"negative predicate": func(c *CharacterCatalogue) { c.Records[1].Requirement.Parameters["minimum-level"] = -1 },
		"numeric key":        func(c *CharacterCatalogue) { c.Records[0].Parameters[epicKey+"unknown"] = 2 },
		"numeric zero":       func(c *CharacterCatalogue) { c.Records[0].Parameters[epicKey+"feats/maximum-times-taken"] = 0 },
		"schema":             func(c *CharacterCatalogue) { c.SchemaVersion = 2 },
	} {
		t.Run(name, func(t *testing.T) {
			c := epicBoonsCatalogue(t)
			mutate(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("invalid accepted")
			}
		})
	}
	for _, field := range []string{"available", "implemented", "playable", "unknown"} {
		b, err := json.Marshal(epicBoonsCatalogue(t))
		if err != nil {
			t.Fatal(err)
		}
		input := strings.Replace(string(b), `"capacity":0`, `"capacity":0,"`+field+`":true`, 1)
		if _, err := DecodeCharacterCatalogue(strings.NewReader(input)); err == nil {
			t.Fatal("unknown field accepted")
		}
	}
	for _, domain := range []string{"required", "option", "predicate"} {
		c := epicBoonsCatalogue(t)
		key := epicKey + "feats/unknown"
		switch domain {
		case "required":
			c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, key)
		case "option":
			c.Records[2].OptionKeys = append(c.Records[2].OptionKeys, key)
		case "predicate":
			c.Records[1].RequiredMechanics = []string{key}
			c.Records[1].Requirement.Key = key
		}
		if err := c.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(repo); err == nil {
			t.Fatal("unknown mechanic ready")
		}
	}
	b, err := json.Marshal(epicBoonsCatalogue(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCharacterCatalogue(strings.NewReader(string(b) + ` {}`)); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}
func TestEpicBoonsLeaveFixedRuntimeUnchanged(t *testing.T) {
	repo := os.DirFS("../../..")
	coverageBefore, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		concept, background, feat string
		hp, gold                  int
	}{
		{"Human Fighter", "Soldier", "Savage Attacker", 12, 29},
		{"Human Cleric", "Acolyte", "Magic Initiate (Cleric; Wisdom)", 10, 118},
	} {
		before, err := SupportedBuildForConcept(tc.concept, tc.background)
		if err != nil {
			t.Fatal(err)
		}
		stats, err := before.Stats()
		if err != nil {
			t.Fatal(err)
		}
		sheetBefore, err := Project(stats, 1, tc.hp, tc.hp, 0)
		if err != nil {
			t.Fatal(err)
		}
		c := epicBoonsCatalogue(t)
		if err := c.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := c.ValidateReadiness(repo); err == nil {
			t.Fatal("metadata ready")
		}
		after, err := SupportedBuildForConcept(tc.concept, tc.background)
		if err != nil {
			t.Fatal(err)
		}
		statsAfter, err := after.Stats()
		if err != nil {
			t.Fatal(err)
		}
		sheetAfter, err := Project(statsAfter, 1, tc.hp, tc.hp, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(sheetBefore, sheetAfter) || after.Gold != tc.gold || after.OriginFeat != tc.feat {
			t.Fatalf("fixed runtime changed: %+v", after)
		}
		if tc.background == "Acolyte" {
			if sheetAfter.Spellcasting.Choices.FeatUses != 1 {
				t.Fatal("feat uses changed")
			}
			for _, spell := range sheetAfter.Spellcasting.Spells {
				if spell.ID == "detect-magic" && spell.Playable {
					t.Fatal("feat effect enabled")
				}
			}
		}
		for _, background := range []string{"Criminal", "Sage"} {
			if _, err := SupportedBuildForConcept(tc.concept, background); err == nil {
				t.Fatal("creation widened")
			}
		}
	}
	coverageAfter, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(coverageBefore, coverageAfter) || coverageAfter.FullSupportEligible(repo) {
		t.Fatal("coverage promoted")
	}
}
