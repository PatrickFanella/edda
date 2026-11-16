package srd521

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strings"
)

// CharacterCatalogue is metadata only. Validation never registers executable
// rules or changes creation, advancement, action or coverage availability.
type CharacterCatalogue struct {
	SchemaVersion int               `json:"schema_version"`
	Source        Source            `json:"source"`
	Records       []CharacterRecord `json:"records"`
}

// CharacterSource links an exact, one-based PDF page to the pinned planning
// inventory. Nested choices/resources reuse their containing source heading.
type CharacterSource struct {
	InventoryID string `json:"inventory_id"`
	Page        int    `json:"page"`
}

// CharacterRecord is a tagged record. Kind is advancement, class, level, feature, subclass,
// origin, feat, prerequisite, choice or resource. IDs are immutable within this
// schema; the rules namespace and Source.Version must both match the pin.
// Parent ownership forms an acyclic graph. OptionKeys and
// RequiredMechanics name contracts, never executable code supplied by data.
type CharacterRecord struct {
	ID                string            `json:"id"`
	Kind              string            `json:"kind"`
	Name              string            `json:"name"`
	Sources           []CharacterSource `json:"sources"`
	ParentID          string            `json:"parent_id"`
	Level             int               `json:"level"`
	ChoiceRefs        []string          `json:"choice_refs"`
	ResourceRefs      []string          `json:"resource_refs"`
	PrerequisiteRefs  []string          `json:"prerequisite_refs"`
	RequiredMechanics []string          `json:"required_mechanics"`
	// Select is the number of distinct options chosen; OptionKeys identify
	// source-defined option domains without importing adjacent catalogues.
	Select     int      `json:"select"`
	OptionKeys []string `json:"option_keys"`
	// Capacity is a source-defined resource maximum, not a mutable balance.
	Capacity int `json:"capacity"`
	// Advancement holds global character-level table values, never class progression.
	Advancement *CharacterAdvancement `json:"advancement,omitempty"`
	// Progression describes a class table row, independently of global XP.
	Progression *CharacterClassProgression `json:"progression,omitempty"`
	// ValuesByLevel records a choice count or resource capacity at each applicable level.
	ValuesByLevel []CharacterLevelValue `json:"values_by_level,omitempty"`
	// Options retain quantities within alternative equipment bundles.
	Options []CharacterChoiceOption `json:"options,omitempty"`
	// Parameters retain uninterpreted numeric class, feature or feat metadata.
	Parameters map[string]int `json:"parameters,omitempty"`
	// AlwaysPrepared retains fixed spell references at explicit class levels.
	// These are opaque grants, never optional choices or executable spells.
	AlwaysPrepared []CharacterPreparedSpells `json:"always_prepared,omitempty"`
	// Requirement is an unevaluated prerequisite predicate. Its key must also
	// appear in RequiredMechanics. Parameters are source metadata, never code.
	Requirement *CharacterRequirement `json:"requirement,omitempty"`
}

// CharacterPreparedSpells is one fixed domain-spell table row. SpellKeys must
// also be required mechanics; only a qualified consumer could grant them.
type CharacterPreparedSpells struct {
	Level     int      `json:"level"`
	SpellKeys []string `json:"spell_keys"`
}

// CharacterClassProgression records grants from a class table, never XP thresholds.
type CharacterClassProgression struct {
	ProficiencyBonus int      `json:"proficiency_bonus"`
	FeatureRefs      []string `json:"feature_refs"`
	FeatureKeys      []string `json:"feature_keys"`
}

// CharacterLevelValue is a count at an explicit class level, not a runtime value.
type CharacterLevelValue struct {
	Level int `json:"level"`
	Value int `json:"value"`
}

// CharacterChoiceOption retains one alternative and its source quantities.
// Quantity keys are required mechanic keys, not resolved item effects.
type CharacterChoiceOption struct {
	Key        string         `json:"key"`
	Quantities map[string]int `json:"quantities"`
}

// CharacterAdvancement records source numbers. Pointers distinguish an omitted
// value from the valid zero XP threshold at level one.
type CharacterAdvancement struct {
	ExperiencePoints *int `json:"experience_points"`
	ProficiencyBonus *int `json:"proficiency_bonus"`
}

// CharacterRequirement describes an opaque, source-defined predicate with
// stable record operands and integer parameters (such as minimum levels or
// ability scores). Only a separately qualified mechanic may interpret it.
type CharacterRequirement struct {
	Key        string         `json:"key"`
	RecordRefs []string       `json:"record_refs"`
	Parameters map[string]int `json:"parameters"`
}

var characterID = regexp.MustCompile(`^srd-5\.2\.1:character/(advancement|class|level|feature|subclass|origin|feat|prerequisite|choice|resource)/` + slug + `(/` + slug + `)*$`)
var mechanicKey = regexp.MustCompile(`^srd-5\.2\.1:mechanic/` + slug + `(/` + slug + `)*$`)

// DecodeCharacterCatalogue rejects unknown fields and trailing JSON. Call
// Validate for source and graph checks; decoding conveys no support claim.
func DecodeCharacterCatalogue(reader io.Reader) (CharacterCatalogue, error) {
	var c CharacterCatalogue
	d := json.NewDecoder(reader)
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return c, fmt.Errorf("decode character catalogue: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return c, fmt.Errorf("character catalogue must contain one JSON object")
	}
	return c, nil
}

// Validate checks metadata against the repository's pinned source inventory.
// It permits unresolved mechanics for future data batches. ValidateReadiness
// additionally rejects those keys; neither method enables playable content.
func (c CharacterCatalogue) Validate(repo fs.FS) error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("unsupported character schema %d", c.SchemaVersion)
	}
	if err := validateSource(c.Source); err != nil {
		return err
	}
	if repo == nil {
		return fmt.Errorf("source inventory filesystem required")
	}
	b, err := fs.ReadFile(repo, "docs/investigations/issue-74-character-inventory.json")
	if err != nil {
		return fmt.Errorf("read character inventory: %w", err)
	}
	var inventory struct {
		Source  Source `json:"source"`
		Records []struct {
			ID     string `json:"id"`
			Page   int    `json:"source_page"`
			Family string `json:"family"`
		} `json:"records"`
	}
	if err := json.Unmarshal(b, &inventory); err != nil {
		return fmt.Errorf("decode character inventory: %w", err)
	}
	if err := validateSource(inventory.Source); err != nil {
		return err
	}
	pages := map[string]int{}
	families := map[string]string{}
	for _, s := range inventory.Records {
		pages[s.ID] = s.Page
		families[s.ID] = s.Family
	}
	if len(c.Records) == 0 {
		return fmt.Errorf("character records required")
	}
	records := map[string]CharacterRecord{}
	levels := map[string]bool{}
	for _, r := range c.Records {
		if !characterID.MatchString(r.ID) || records[r.ID].ID != "" || !strings.HasPrefix(r.ID, "srd-5.2.1:character/"+r.Kind+"/") {
			return fmt.Errorf("invalid or duplicate character ID %q", r.ID)
		}
		if strings.TrimSpace(r.Name) == "" || len(r.Sources) == 0 || len(r.RequiredMechanics) == 0 {
			return fmt.Errorf("%s: name, source and required mechanics required", r.ID)
		}
		seenSource := map[CharacterSource]bool{}
		for _, s := range r.Sources {
			if pages[s.InventoryID] != s.Page || s.Page < 1 || s.Page > c.Source.Pages || seenSource[s] {
				return fmt.Errorf("%s: missing, duplicate or mismatched source link", r.ID)
			}
			seenSource[s] = true
		}
		expectedFamily := map[string]string{"advancement": "levels", "class": "classes", "level": "classes", "subclass": "subclasses", "feat": "feats"}[r.Kind]
		if expectedFamily != "" && families[r.Sources[0].InventoryID] != expectedFamily {
			return fmt.Errorf("%s: wrong source family", r.ID)
		}
		family := families[r.Sources[0].InventoryID]
		if r.Kind == "origin" && family != "species" && family != "backgrounds" || r.Kind == "feature" && family != "class-features" && family != "subclass-features" {
			return fmt.Errorf("%s: wrong source family", r.ID)
		}
		isLevel := r.Kind == "level" || r.Kind == "advancement"
		if isLevel && (r.Level < 1 || r.Level > 20) || !isLevel && r.Level != 0 {
			return fmt.Errorf("%s: invalid level", r.ID)
		}
		if isLevel {
			key := fmt.Sprintf("%s:%d", r.ParentID, r.Level)
			if levels[key] {
				return fmt.Errorf("%s: duplicate class level", r.ID)
			}
			levels[key] = true
		}
		if r.Kind == "advancement" {
			a := r.Advancement
			sourceID := fmt.Sprintf("srd-5.2.1:planning/levels/character-level-%d", r.Level)
			if len(r.Sources) != 1 || r.Sources[0].InventoryID != sourceID || r.Sources[0].Page != 23 ||
				r.ID != fmt.Sprintf("srd-5.2.1:character/advancement/%d", r.Level) || r.Name != fmt.Sprintf("Character level %d", r.Level) {
				return fmt.Errorf("%s: advancement identity/source disagrees with level", r.ID)
			}
			if a == nil || a.ExperiencePoints == nil || a.ProficiencyBonus == nil ||
				*a.ExperiencePoints < 0 || (r.Level == 1 && *a.ExperiencePoints != 0) ||
				(r.Level > 1 && *a.ExperiencePoints == 0) || *a.ProficiencyBonus < 2 || *a.ProficiencyBonus > 6 {
				return fmt.Errorf("%s: explicit valid advancement numbers required", r.ID)
			}
			for _, required := range []string{"srd-5.2.1:mechanic/character/experience-threshold", "srd-5.2.1:mechanic/character/proficiency-bonus"} {
				found := false
				for _, key := range r.RequiredMechanics {
					if key == required {
						found = true
					}
				}
				if !found {
					return fmt.Errorf("%s: missing advancement mechanic %s", r.ID, required)
				}
			}
		} else if r.Advancement != nil {
			return fmt.Errorf("%s: advancement numbers on non-advancement", r.ID)
		}
		if r.ChoiceRefs == nil || r.ResourceRefs == nil || r.PrerequisiteRefs == nil || r.OptionKeys == nil {
			return fmt.Errorf("%s: reference arrays must be explicit", r.ID)
		}
		if r.Kind == "choice" {
			if r.Select < 1 || len(r.OptionKeys) == 0 {
				return fmt.Errorf("%s: choice needs count and option keys", r.ID)
			}
		} else if r.Select != 0 || len(r.OptionKeys) != 0 {
			return fmt.Errorf("%s: choice fields on non-choice", r.ID)
		}
		if r.Kind == "resource" {
			if r.Capacity < 1 {
				return fmt.Errorf("%s: resource capacity required", r.ID)
			}
		} else if r.Capacity != 0 {
			return fmt.Errorf("%s: capacity on non-resource", r.ID)
		}
		for _, keys := range [][]string{r.RequiredMechanics, r.OptionKeys} {
			seen := map[string]bool{}
			for _, k := range keys {
				if !mechanicKey.MatchString(k) || seen[k] {
					return fmt.Errorf("%s: invalid or duplicate mechanic key %q", r.ID, k)
				}
				seen[k] = true
			}
		}
		if r.Kind == "prerequisite" {
			if r.Requirement == nil || r.Requirement.RecordRefs == nil || r.Requirement.Parameters == nil {
				return fmt.Errorf("%s: prerequisite predicate required", r.ID)
			}
			found := false
			for _, key := range r.RequiredMechanics {
				if key == r.Requirement.Key {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("%s: prerequisite key must be required", r.ID)
			}
			for key, value := range r.Requirement.Parameters {
				if !familyName.MatchString(key) || value < 0 {
					return fmt.Errorf("%s: invalid prerequisite parameter", r.ID)
				}
			}
		} else if r.Requirement != nil {
			return fmt.Errorf("%s: predicate on non-prerequisite", r.ID)
		}
		if err := validateClassMetadata(r); err != nil {
			return err
		}
		records[r.ID] = r
	}
	edges := map[string][]string{}
	for _, r := range c.Records {
		parent := records[r.ParentID]
		if r.Progression != nil {
			seen := map[string]bool{}
			for _, id := range r.Progression.FeatureRefs {
				target := records[id]
				owner := records[target.ParentID]
				if target.Kind != "feature" || owner.Kind != "level" || owner.ParentID != r.ParentID || owner.Level > r.Level || seen[id] {
					return fmt.Errorf("%s: invalid progression feature", r.ID)
				}
				seen[id] = true
			}
		}
		allowed := false
		switch r.Kind {
		case "advancement", "class", "origin", "feat":
			allowed = r.ParentID == ""
		case "level", "subclass":
			allowed = parent.Kind == "class"
		case "feature":
			allowed = parent.Kind == "level" || parent.Kind == "subclass"
		// A choice may own conditional subchoices; the same source,
		// reciprocal-reference and cycle guards still apply.
		case "choice":
			allowed = parent.ID != "" && parent.Kind != "resource" && parent.Kind != "prerequisite"
		case "resource":
			allowed = parent.ID != "" && parent.Kind != "choice" && parent.Kind != "resource" && parent.Kind != "prerequisite"
		case "prerequisite":
			allowed = parent.ID != "" && parent.Kind != "resource" && parent.Kind != "prerequisite"
		}
		if !allowed {
			return fmt.Errorf("%s: missing or invalid parent", r.ID)
		}
		// Parent links establish ownership, not dependency edges: an owner refers
		// to its nested choice/resource, which must not create a false cycle.
		for kind, refs := range map[string][]string{"choice": r.ChoiceRefs, "resource": r.ResourceRefs, "prerequisite": r.PrerequisiteRefs} {
			seen := map[string]bool{}
			for _, id := range refs {
				target := records[id]
				if target.Kind != kind || seen[id] || target.ParentID != r.ID {
					return fmt.Errorf("%s: missing, duplicate or wrong-owner %s link %q", r.ID, kind, id)
				}
				seen[id] = true
			}
		}
		if r.Requirement != nil {
			seen := map[string]bool{}
			for _, id := range r.Requirement.RecordRefs {
				if records[id].ID == "" || seen[id] {
					return fmt.Errorf("%s: missing or duplicate prerequisite operand", r.ID)
				}
				seen[id] = true
			}
		}
		if r.ParentID != "" {
			edges[r.ID] = append(edges[r.ID], r.ParentID)
		}
		if r.Kind == "choice" || r.Kind == "resource" || r.Kind == "prerequisite" {
			refs := parent.ChoiceRefs
			if r.Kind == "resource" {
				refs = parent.ResourceRefs
			}
			if r.Kind == "prerequisite" {
				refs = parent.PrerequisiteRefs
			}
			found := false
			for _, id := range refs {
				if id == r.ID {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("%s: parent must link nested record", r.ID)
			}
		}
	}
	visiting, done := map[string]bool{}, map[string]bool{}
	var visit func(string) bool
	visit = func(id string) bool {
		if visiting[id] {
			return false
		}
		if done[id] {
			return true
		}
		visiting[id] = true
		for _, next := range edges[id] {
			if !visit(next) {
				return false
			}
		}
		visiting[id] = false
		done[id] = true
		return true
	}
	for id := range records {
		if !visit(id) {
			return fmt.Errorf("%s: cyclic parent links", id)
		}
	}
	// Check ownership only after proving the parent graph is acyclic. Source
	// headings must describe the same class/level as the record's ancestry.
	for _, r := range c.Records {
		parent := records[r.ParentID]
		if len(r.ValuesByLevel) > 0 {
			owner := parent
			for owner.ID != "" && owner.Kind != "level" {
				owner = records[owner.ParentID]
			}
			for _, value := range r.ValuesByLevel {
				if value.Level < owner.Level {
					return fmt.Errorf("%s: level count precedes owning feature", r.ID)
				}
			}
		}
		for _, source := range r.Sources {
			switch r.Kind {
			case "level", "choice", "resource", "prerequisite":
				if !containsCharacterSource(parent.Sources, source) {
					return fmt.Errorf("%s: source does not belong to parent", r.ID)
				}
			case "feature", "subclass":
				owner := parent
				for owner.ID != "" && owner.Kind != "class" {
					owner = records[owner.ParentID]
				}
				if owner.ID == "" {
					return fmt.Errorf("%s: class source owner required", r.ID)
				}
				class := strings.TrimPrefix(owner.Sources[0].InventoryID, "srd-5.2.1:planning/classes/")
				family := "subclasses"
				heading := class + "-"
				if r.Kind == "feature" {
					family = "subclass-features"
					heading += "level-"
					if parent.Kind == "level" {
						family = "class-features"
						heading += fmt.Sprintf("%d-", parent.Level)
					}
				}
				if !strings.HasPrefix(source.InventoryID, "srd-5.2.1:planning/"+family+"/"+heading) {
					return fmt.Errorf("%s: source disagrees with class/level ownership", r.ID)
				}
			}
		}
	}
	return nil
}

func containsCharacterSource(sources []CharacterSource, wanted CharacterSource) bool {
	for _, source := range sources {
		if source == wanted {
			return true
		}
	}
	return false
}

// ValidateReadiness fails closed on every unimplemented feature and choice key.
// There are deliberately no qualified catalogue mechanics yet. Existing fixed
// builds use their own reviewed paths, not this catalogue. Adding a JSON record
// cannot change this code-owned boundary or qualify Full Rules.
func (c CharacterCatalogue) ValidateReadiness(repo fs.FS) error {
	if err := c.Validate(repo); err != nil {
		return err
	}
	for _, r := range c.Records {
		for _, key := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if !qualifiedCharacterMechanic(key) {
				return fmt.Errorf("%s: unimplemented required key %q", r.ID, key)
			}
		}
	}
	return nil
}

func qualifiedCharacterMechanic(string) bool { return false }

func validateClassMetadata(r CharacterRecord) error {
	required := map[string]bool{}
	for _, key := range r.RequiredMechanics {
		required[key] = true
	}
	previousPreparedLevel := 0
	seenSpells := map[string]bool{}
	for _, row := range r.AlwaysPrepared {
		if r.Kind != "feature" || row.Level <= previousPreparedLevel || row.Level > 20 || len(row.SpellKeys) == 0 {
			return fmt.Errorf("%s: invalid always-prepared row", r.ID)
		}
		previousPreparedLevel = row.Level
		for _, key := range row.SpellKeys {
			if !mechanicKey.MatchString(key) || !required[key] || seenSpells[key] {
				return fmt.Errorf("%s: fixed spell key must be required and unique", r.ID)
			}
			seenSpells[key] = true
		}
	}
	if p := r.Progression; p != nil {
		if r.Kind != "level" || p.ProficiencyBonus < 2 || p.ProficiencyBonus > 6 || p.FeatureRefs == nil || p.FeatureKeys == nil {
			return fmt.Errorf("%s: invalid class progression", r.ID)
		}
		seen := map[string]bool{}
		for _, key := range p.FeatureKeys {
			if !required[key] || seen[key] {
				return fmt.Errorf("%s: progression key must be required and unique", r.ID)
			}
			seen[key] = true
		}
	}
	if len(r.ValuesByLevel) > 0 {
		if r.Kind != "resource" && r.Kind != "choice" {
			return fmt.Errorf("%s: level counts require resource or choice", r.ID)
		}
		previous := 0
		for _, v := range r.ValuesByLevel {
			if v.Level <= previous || v.Level > 20 || v.Value < 1 {
				return fmt.Errorf("%s: invalid level count", r.ID)
			}
			previous = v.Level
		}
		initial := r.Capacity
		if r.Kind == "choice" {
			initial = r.Select
		}
		if r.ValuesByLevel[0].Value != initial {
			return fmt.Errorf("%s: initial count mismatch", r.ID)
		}
	}
	seen := map[string]bool{}
	for _, option := range r.Options {
		if r.Kind != "choice" || seen[option.Key] {
			return fmt.Errorf("%s: invalid option", r.ID)
		}
		found := false
		for _, key := range r.OptionKeys {
			if key == option.Key {
				found = true
			}
		}
		if !found || len(option.Quantities) == 0 {
			return fmt.Errorf("%s: option needs domain and quantities", r.ID)
		}
		seen[option.Key] = true
		for key, value := range option.Quantities {
			if !required[key] || value < 1 {
				return fmt.Errorf("%s: quantity key must be required", r.ID)
			}
		}
	}
	if len(r.Parameters) > 0 && r.Kind != "class" && r.Kind != "feature" && r.Kind != "feat" {
		return fmt.Errorf("%s: parameters require class, feature or feat", r.ID)
	}
	for key, value := range r.Parameters {
		if !required[key] || value < 1 {
			return fmt.Errorf("%s: parameter key must be required", r.ID)
		}
	}
	return nil
}
