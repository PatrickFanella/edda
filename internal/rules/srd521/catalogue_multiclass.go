package srd521

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"slices"
	"strings"
)

// MulticlassCatalogue describes #320's non-indexed source section. It cannot
// register classes, grants, combinations, advancement or executable mechanics.
// Named inventory links retain CharacterCatalogue's unchanged validation.
type MulticlassCatalogue struct {
	SchemaVersion int                     `json:"schema_version"`
	Source        Source                  `json:"source"`
	Scope         MulticlassScope         `json:"scope"`
	Catalogues    []string                `json:"catalogues"`
	Classes       []MulticlassClass       `json:"classes"`
	Requirements  []MulticlassRequirement `json:"requirements"`
	SlotRows      []MulticlassSlots       `json:"slot_rows"`
	SlotTerms     []MulticlassSlotTerm    `json:"slot_terms"`
}

// MulticlassScope binds the explicit non-indexed section to its owning issue.
type MulticlassScope struct {
	Issue     int    `json:"issue"`
	Section   string `json:"section"`
	FirstPage int    `json:"first_page"`
	LastPage  int    `json:"last_page"`
}

// MulticlassClass references canonical core traits and proficiency choices.
// PrimaryAny is OR of AND groups, with MinimumScore applying to each ability.
// Proficiencies describe the restricted multiclass subset, never grants.
type MulticlassClass struct {
	ClassRef             string          `json:"class_ref"`
	Source               CharacterSource `json:"source"`
	MinimumScore         int             `json:"minimum_score"`
	PrimaryAny           [][]string      `json:"primary_any"`
	ProficiencyOwnerRefs []string        `json:"proficiency_owner_refs"`
	Proficiencies        []string        `json:"proficiencies"`
}

// MulticlassRequirement retains uninterpreted conditions, operands and values.
// Key and all clauses are unresolved requirements, never runtime instructions.
type MulticlassRequirement struct {
	Key        string         `json:"key"`
	FirstPage  int            `json:"first_page"`
	LastPage   int            `json:"last_page"`
	Clauses    []string       `json:"clauses"`
	RecordRefs []string       `json:"record_refs"`
	Parameters map[string]int `json:"parameters"`
}

// MulticlassSlotTerm retains a source class-level operand in the shared sum.
// Dividing that class's levels and applying Rounding is a requirement, not code.
type MulticlassSlotTerm struct {
	ClassRef     string `json:"class_ref"`
	LevelDivisor int    `json:"level_divisor"`
	Rounding     string `json:"rounding"`
}

// MulticlassSlots records all nine cells, using zero for source em dashes.
type MulticlassSlots struct {
	Level int   `json:"level"`
	Slots []int `json:"slots"`
}

// DecodeMulticlassCatalogue rejects unknown or duplicate members, null values,
// non-integer numbers and trailing JSON. Explicit integer zero remains valid.
func DecodeMulticlassCatalogue(r io.Reader) (MulticlassCatalogue, error) {
	var c MulticlassCatalogue
	data, err := io.ReadAll(r)
	if err != nil {
		return c, fmt.Errorf("read multiclass: %w", err)
	}
	members := json.NewDecoder(bytes.NewReader(data))
	members.UseNumber()
	if err := uniqueCatalogueJSON(members, true); err != nil {
		return c, fmt.Errorf("decode multiclass: %w", err)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return c, fmt.Errorf("decode multiclass: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return c, fmt.Errorf("multiclass requires one JSON object")
	}
	return c, nil
}

var multiclassClause = regexp.MustCompile(`^` + slug + `$`)

var multiclassNames = []string{"barbarian", "bard", "cleric", "druid", "fighter", "monk", "paladin", "ranger", "rogue", "sorcerer", "warlock", "wizard"}

// Validate checks the narrow source scope and canonical composition. Semantic
// agreement is independently proved by Go source expectations and the PDF tool.
func (c MulticlassCatalogue) Validate(repo fs.FS) error {
	if c.SchemaVersion != 1 || c.Scope != (MulticlassScope{320, "multiclassing", 24, 26}) {
		return fmt.Errorf("unsupported multiclass schema/source scope")
	}
	if err := validateSource(c.Source); err != nil {
		return err
	}
	if repo == nil {
		return fmt.Errorf("canonical filesystem required")
	}
	wantFiles := []string{}
	for _, n := range multiclassNames {
		wantFiles = append(wantFiles, "character-"+n+".json")
	}
	wantFiles = append(wantFiles, "character-advancement.json")
	if !slices.Equal(c.Catalogues, wantFiles) {
		return fmt.Errorf("canonical class/advancement composition must occur once in source order")
	}
	var canonical CharacterCatalogue
	for i, file := range c.Catalogues {
		f, err := repo.Open("internal/rules/srd521/testdata/" + file)
		if err != nil {
			return fmt.Errorf("open canonical: %w", err)
		}
		batch, err := DecodeCharacterCatalogue(f)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if err := validateSource(batch.Source); err != nil {
			return fmt.Errorf("canonical %s: %w", file, err)
		}
		if i == 0 {
			canonical = batch
		} else {
			canonical.Records = append(canonical.Records, batch.Records...)
		}
	}
	if err := canonical.Validate(repo); err != nil {
		return err
	}
	owners := map[string]CharacterRecord{}
	for _, r := range canonical.Records {
		owners[r.ID] = r
	}
	if len(c.Classes) != 12 {
		return fmt.Errorf("all twelve class references required")
	}
	for i, cl := range c.Classes {
		ref := "srd-5.2.1:character/class/" + multiclassNames[i]
		owner := owners[ref]
		if cl.ClassRef != ref || len(owner.Sources) == 0 || cl.Source != owner.Sources[0] || cl.MinimumScore < 1 || cl.MinimumScore > 30 || len(cl.PrimaryAny) == 0 {
			return fmt.Errorf("invalid multiclass class/source/prerequisite %s", cl.ClassRef)
		}
		groups := map[string]bool{}
		for _, group := range cl.PrimaryAny {
			if len(group) == 0 {
				return fmt.Errorf("empty primary alternative")
			}
			seen := map[string]bool{}
			for _, a := range group {
				if !slices.Contains([]string{"strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}, a) || seen[a] {
					return fmt.Errorf("invalid primary ability")
				}
				seen[a] = true
			}
			key := strings.Join(group, "+")
			if groups[key] {
				return fmt.Errorf("duplicate primary alternative")
			}
			groups[key] = true
		}
		if len(cl.ProficiencyOwnerRefs) == 0 || cl.ProficiencyOwnerRefs[0] != ref {
			return fmt.Errorf("canonical proficiency owner required")
		}
		seen := map[string]bool{}
		for _, r := range cl.ProficiencyOwnerRefs {
			if owners[r].ID == "" || seen[r] || !multiclassOwnedBy(owners, r, ref) {
				return fmt.Errorf("invalid proficiency owner %s", r)
			}
			seen[r] = true
		}
		seen = map[string]bool{}
		for _, p := range cl.Proficiencies {
			if !slices.Contains([]string{"martial-weapons", "shields", "light-armor", "medium-armor", "any-skill", "class-skill", "musical-instrument", "thieves-tools"}, p) || seen[p] {
				return fmt.Errorf("invalid proficiency subset")
			}
			seen[p] = true
		}
	}
	scopes := map[string][2]int{"prerequisites": {24, 25}, "experience-and-proficiency": {25, 25}, "hit-points-and-dice": {25, 25}, "proficiencies": {25, 25}, "class-features": {25, 25}, "armor-class": {25, 25}, "extra-attack": {25, 25}, "spellcasting-scope": {25, 25}, "prepared-spells": {25, 26}, "cantrips": {25, 25}, "spell-slots": {25, 26}, "pact-magic": {26, 26}}
	if len(c.Requirements) != len(scopes) {
		return fmt.Errorf("complete section requirements required")
	}
	seen := map[string]bool{}
	for _, r := range c.Requirements {
		scope, ok := scopes[strings.TrimPrefix(r.Key, "srd-5.2.1:mechanic/multiclass/")]
		if !ok || !strings.HasPrefix(r.Key, "srd-5.2.1:mechanic/multiclass/") || seen[r.Key] || scope != [2]int{r.FirstPage, r.LastPage} || len(r.Clauses) == 0 || r.RecordRefs == nil || r.Parameters == nil {
			return fmt.Errorf("invalid/duplicate requirement scope %s", r.Key)
		}
		seen[r.Key] = true
		clauses := map[string]bool{}
		for _, clause := range r.Clauses {
			if !multiclassClause.MatchString(clause) || clauses[clause] {
				return fmt.Errorf("invalid/duplicate clause")
			}
			clauses[clause] = true
		}
		refs := map[string]bool{}
		for _, ref := range r.RecordRefs {
			if owners[ref].ID == "" || refs[ref] {
				return fmt.Errorf("missing/duplicate operand %s", ref)
			}
			refs[ref] = true
		}
		for k, v := range r.Parameters {
			if !multiclassClause.MatchString(k) || v < 1 {
				return fmt.Errorf("invalid requirement value")
			}
		}
	}
	if len(c.SlotTerms) != 7 {
		return fmt.Errorf("seven slot class operands required")
	}
	terms := map[string]bool{}
	for _, term := range c.SlotTerms {
		if owners[term.ClassRef].Kind != "class" || terms[term.ClassRef] || term.LevelDivisor < 1 || term.LevelDivisor > 3 || !slices.Contains([]string{"none", "up", "down"}, term.Rounding) {
			return fmt.Errorf("invalid slot class operand")
		}
		terms[term.ClassRef] = true
	}
	if len(c.SlotRows) != 20 {
		return fmt.Errorf("twenty slot rows required")
	}
	for i, row := range c.SlotRows {
		if row.Level != i+1 || len(row.Slots) != 9 {
			return fmt.Errorf("invalid slot row")
		}
		for _, v := range row.Slots {
			if v < 0 || v > 4 {
				return fmt.Errorf("invalid slot cell")
			}
		}
	}
	return nil
}

// The validated canonical graph is acyclic; choices can be nested below traits.
func multiclassOwnedBy(owners map[string]CharacterRecord, id, class string) bool {
	for id != "" {
		if id == class {
			return true
		}
		id = owners[id].ParentID
	}
	return false
}

// ValidateReadiness always refuses multiclass combinations, even if every
// source fact is correct. No metadata can supply executable support.
func (c MulticlassCatalogue) ValidateReadiness(repo fs.FS) error {
	if err := c.Validate(repo); err != nil {
		return err
	}
	return fmt.Errorf("unimplemented required mechanic %s: no supported multiclass combinations", c.Requirements[0].Key)
}
