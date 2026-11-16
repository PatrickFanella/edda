package srd521

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

// SpellRecords is an offline, versioned data contract. It never registers actions.
type SpellRecords struct {
	SchemaVersion int           `json:"schema_version"`
	Source        Source        `json:"source"`
	Records       []SpellRecord `json:"records"`
}

// SpellRecord describes source data or an explicitly bounded existing fixture.
// RequiredMechanics are dependencies, not resolver registrations.
type SpellRecord struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	SourceID          string              `json:"source_id"`
	SourcePage        int                 `json:"source_page"`
	Scope             string              `json:"scope"`
	Level             int                 `json:"level"`
	School            string              `json:"school"`
	Lists             []string            `json:"lists"`
	CastingTime       string              `json:"casting_time"`
	Range             SpellRange          `json:"range"`
	Target            SpellTarget         `json:"target"`
	Components        SpellComponents     `json:"components"`
	Ritual            *bool               `json:"ritual"`
	Concentration     *bool               `json:"concentration"`
	DurationSeconds   int                 `json:"duration_seconds"`
	Duration          string              `json:"duration"`
	Scaling           SpellScaling        `json:"scaling"`
	Effects           []SpellEffect       `json:"effects"`
	RequiredMechanics []string            `json:"required_mechanics"`
	SlotLevels        []int               `json:"slot_levels"`
	SourceDetails     *SpellSourceDetails `json:"source_details,omitempty"`
}

// SpellRange retains the source range even for a self-only fixture.
type SpellRange struct {
	Kind string `json:"kind"`
	Feet int    `json:"feet"`
}

// SpellTarget declares the shape and eligibility of a selected target.
type SpellTarget struct {
	Shape       string `json:"shape"`
	Count       int    `json:"count"`
	Eligibility string `json:"eligibility"`
}

// SpellComponents retains material cost in copper pieces and consumption.
// A nil Material means no material requirement; false flags must be explicit.
type SpellComponents struct {
	Verbal   *bool          `json:"verbal"`
	Somatic  *bool          `json:"somatic"`
	Material *SpellMaterial `json:"material"`
}

// UnmarshalJSON requires an explicit material object or null. Omission cannot
// silently remove a casting requirement from imported data.
func (c *SpellComponents) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if _, ok := fields["material"]; !ok {
		return fmt.Errorf("explicit material object or null required")
	}
	type plain SpellComponents
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode((*plain)(c))
}

// SpellMaterial distinguishes absent cost from a consumed component.
type SpellMaterial struct {
	Description string `json:"description"`
	CostCP      *int   `json:"cost_cp"`
	Consumed    *bool  `json:"consumed"`
}

// SpellScaling identifies per-slot dice additions or references a typed ray
// count increment in SourceDetails.Rays.
// Bounded fixtures use none; their higher-slot catalogue data belongs elsewhere.
type SpellScaling struct {
	Kind       string   `json:"kind"`
	AboveLevel int      `json:"above_level"`
	DiceCount  int      `json:"dice_count"`
	DieSides   int      `json:"die_sides"`
	EffectIDs  []string `json:"effect_ids"`
}

// SpellEffect stores numeric payloads and explicit timing/outcome dependencies.
// HalfInitialOnly means half the initial roll on a miss, with no delayed damage.
type SpellEffect struct {
	ID                 string `json:"id"`
	Mechanic           string `json:"mechanic"`
	Timing             string `json:"timing"`
	DiceCount          int    `json:"dice_count"`
	DieSides           int    `json:"die_sides"`
	DamageType         string `json:"damage_type"`
	Bonus              int    `json:"bonus"`
	AddCastingModifier *bool  `json:"add_casting_modifier"`
	HalfInitialOnly    *bool  `json:"half_initial_only"`
}

// DecodeSpellRecords rejects malformed data and unknown JSON fields offline.
// Unknown mechanic names are retained but remain unavailable.
func DecodeSpellRecords(r io.Reader) (SpellRecords, error) {
	var records SpellRecords
	data, err := io.ReadAll(r)
	if err != nil {
		return records, fmt.Errorf("read spell records: %w", err)
	}
	// Reuse the catalogue JSON guard: duplicate and case-aliased properties
	// must fail before encoding/json can silently take the last value.
	if err := uniqueItemJSON(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return records, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&records); err != nil {
		return records, fmt.Errorf("decode spell records: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return SpellRecords{}, fmt.Errorf("spell records require one JSON object")
	}
	if err := records.Validate(); err != nil {
		return SpellRecords{}, err
	}
	return records, nil
}

// Validate checks structure, provenance and internal dependencies, not conformance.
func (r SpellRecords) Validate() error {
	if r.SchemaVersion != 1 {
		return fmt.Errorf("unsupported spell schema %d", r.SchemaVersion)
	}
	if err := validateSource(r.Source); err != nil {
		return err
	}
	if len(r.Records) == 0 {
		return fmt.Errorf("spell records required")
	}
	seen := map[string]bool{}
	for _, s := range r.Records {
		if seen[s.ID] || !strings.HasPrefix(s.ID, prefix+"spells/") || !contentID.MatchString(s.ID) || s.SourceID != s.ID || s.SourcePage < 107 || s.SourcePage > 175 || strings.TrimSpace(s.Name) == "" {
			return fmt.Errorf("invalid, duplicate or source-less spell %q", s.ID)
		}
		seen[s.ID] = true
		if err := s.validate(); err != nil {
			return fmt.Errorf("%s: %w", s.ID, err)
		}
	}
	return nil
}

func (s SpellRecord) validate() error {
	if !slices.Contains([]string{"complete-source-record", "bounded-self-level-1"}, s.Scope) || s.Level < 0 || s.Level > 9 || !slices.Contains(strings.Fields("abjuration conjuration divination enchantment evocation illusion necromancy transmutation"), s.School) || len(s.Lists) == 0 || !uniqueNames(s.Lists) {
		return fmt.Errorf("invalid scope, level, school or lists")
	}
	for _, class := range s.Lists {
		if !slices.Contains(strings.Fields("bard cleric druid paladin ranger sorcerer warlock wizard"), class) {
			return fmt.Errorf("invalid spell list")
		}
	}
	if !slices.Contains([]string{"action", "bonus-action", "reaction", "10-minutes"}, s.CastingTime) || !slices.Contains([]string{"feet", "touch", "self", "unlimited"}, s.Range.Kind) || s.Range.Feet < 0 || (s.Range.Kind == "feet" && s.Range.Feet == 0) || (s.Range.Kind != "feet" && s.Range.Feet != 0) || !validSpellTarget(s.Target) {
		return fmt.Errorf("invalid casting, range or target shape")
	}
	if s.Components.Verbal == nil || s.Components.Somatic == nil || s.Ritual == nil || s.Concentration == nil {
		return fmt.Errorf("explicit component, ritual and concentration flags required")
	}
	if m := s.Components.Material; m != nil {
		if strings.TrimSpace(m.Description) == "" || m.CostCP == nil || *m.CostCP < 0 || m.Consumed == nil {
			return fmt.Errorf("invalid material")
		}
	}
	if !slices.Contains([]string{"instantaneous", "timed", "until-dispelled"}, s.Duration) || s.DurationSeconds < 0 || (s.Duration == "instantaneous" && (s.DurationSeconds != 0 || *s.Concentration)) || (s.Duration == "timed" && s.DurationSeconds == 0) {
		return fmt.Errorf("invalid duration")
	}
	if s.Duration == "until-dispelled" && (s.DurationSeconds != 0 || *s.Concentration) {
		return fmt.Errorf("invalid indefinite duration")
	}
	if err := s.validateSourceDetails(); err != nil {
		return err
	}
	if len(s.RequiredMechanics) == 0 || !uniqueNames(s.RequiredMechanics) || len(s.Effects) == 0 || len(s.SlotLevels) == 0 {
		return fmt.Errorf("mechanics, effects and slots required")
	}
	last := -1
	for _, level := range s.SlotLevels {
		if level < s.Level || level > 9 || level <= last {
			return fmt.Errorf("invalid slot levels")
		}
		last = level
	}
	effects := map[string]bool{}
	for _, e := range s.Effects {
		if e.ID == "" || effects[e.ID] || !slices.Contains(s.RequiredMechanics, e.Mechanic) || e.Timing == "" || e.DiceCount < 0 || (e.DiceCount > 0 && !slices.Contains([]int{4, 6, 8, 10, 12, 20}, e.DieSides)) || (e.DiceCount == 0 && e.DieSides != 0) || e.AddCastingModifier == nil || e.HalfInitialOnly == nil {
			return fmt.Errorf("invalid effect")
		}
		effects[e.ID] = true
	}
	sc := s.Scaling
	switch sc.Kind {
	case "none":
		if sc.AboveLevel != 0 || sc.DiceCount != 0 || sc.DieSides != 0 || len(sc.EffectIDs) != 0 {
			return fmt.Errorf("unexpected scaling payload")
		}
	case "slot-targets", "slot-duration":
		if sc.AboveLevel != s.Level || sc.DiceCount != 0 || sc.DieSides != 0 || len(sc.EffectIDs) == 0 || !uniqueNames(sc.EffectIDs) || s.SourceDetails == nil {
			return fmt.Errorf("invalid typed slot scaling")
		}
		d := s.SourceDetails
		if sc.Kind == "slot-targets" {
			base, additional := 0, 0
			if d.Bless != nil {
				base, additional = d.Bless.BaseTargets, d.Bless.AdditionalTargetsPerSlot
			} else if d.Blindness != nil {
				base, additional = d.Blindness.BaseTargets, d.Blindness.AdditionalTargetsPerSlot
			}
			if base != s.Target.Count || additional <= 0 || !slices.Contains(s.RequiredMechanics, "scaling.targets-per-slot") {
				return fmt.Errorf("target scaling requires consistent base and positive increment")
			}
		}
		if sc.Kind == "slot-duration" && (d.Curse == nil || !slices.Contains(s.RequiredMechanics, "scaling.curse-slot-duration-concentration")) {
			return fmt.Errorf("duration scaling requires typed slot table")
		}
		for _, id := range sc.EffectIDs {
			if !effects[id] {
				return fmt.Errorf("unknown scaling effect")
			}
		}
	case "slot-count":
		if sc.AboveLevel != s.Level || sc.DiceCount != 0 || sc.DieSides != 0 || s.SourceDetails == nil || s.SourceDetails.Rays == nil || s.SourceDetails.Rays.AdditionalPerSlot <= 0 || len(sc.EffectIDs) != 1 || !effects[sc.EffectIDs[0]] || !slices.Contains(s.RequiredMechanics, "scaling.per-slot-ray") {
			return fmt.Errorf("invalid ray scaling")
		}
	case "slot":
		if sc.AboveLevel != s.Level || sc.DiceCount <= 0 || !slices.Contains([]int{4, 6, 8, 10, 12, 20}, sc.DieSides) || !uniqueNames(sc.EffectIDs) || len(sc.EffectIDs) == 0 || !slices.Contains(s.RequiredMechanics, "scaling.per-slot") {
			return fmt.Errorf("invalid slot scaling")
		}
		for _, id := range sc.EffectIDs {
			if !effects[id] {
				return fmt.Errorf("unknown scaling effect")
			}
		}
	default:
		return fmt.Errorf("unsupported scaling shape")
	}
	if s.Scope == "bounded-self-level-1" && (s.Level != 1 || s.Target.Shape != "single" || s.Target.Count != 1 || s.Target.Eligibility != "conscious-self" || !slices.Equal(s.SlotLevels, []int{1}) || sc.Kind != "none") {
		return fmt.Errorf("bounded fixture cannot widen targets or slots")
	}
	return nil
}

func uniqueNames(names []string) bool {
	seen := map[string]bool{}
	for _, name := range names {
		if strings.TrimSpace(name) == "" || seen[name] {
			return false
		}
		seen[name] = true
	}
	return true
}

// MechanicsAvailable reports only the two existing bounded fixture dependencies.
// It is not permission to cast. Complete source records and unknown/unsupported
// keys always return false, independently of their presence in a catalogue.
func (s SpellRecord) MechanicsAvailable() bool {
	if s.validate() != nil || s.Scope != "bounded-self-level-1" || s.SourceDetails != nil {
		return false
	}
	if !*s.Components.Verbal || !*s.Components.Somatic || *s.Ritual || s.School != "abjuration" || len(s.Effects) != 1 {
		return false
	}
	e := s.Effects[0]
	if e.DamageType != "" || *e.HalfInitialOnly {
		return false
	}
	var required []string
	switch s.ID {
	case prefix + "spells/cure-wounds":
		if s.Name != "Cure Wounds" || s.SourcePage != 121 || s.CastingTime != "action" || s.Range != (SpellRange{"touch", 0}) || s.Components.Material != nil || *s.Concentration || s.Duration != "instantaneous" || s.DurationSeconds != 0 || !slices.Equal(s.Lists, []string{"bard", "cleric", "druid", "paladin", "ranger"}) || e.ID != "healing" || e.Mechanic != "healing.cure-wounds-self" || e.Timing != "immediate" || e.DiceCount != 2 || e.DieSides != 8 || e.Bonus != 0 || !*e.AddCastingModifier {
			return false
		}
		required = []string{"cast.cleric-self-slot-1", "healing.cure-wounds-self"}
	case prefix + "spells/shield-of-faith":
		m := s.Components.Material
		if s.Name != "Shield of Faith" || s.SourcePage != 162 || s.CastingTime != "bonus-action" || s.Range != (SpellRange{"feet", 60}) || m == nil || m.Description != "a prayer scroll" || *m.CostCP != 0 || *m.Consumed || !*s.Concentration || s.Duration != "timed" || s.DurationSeconds != 600 || !slices.Equal(s.Lists, []string{"cleric", "paladin"}) || e.ID != "ac" || e.Mechanic != "defense.shield-of-faith-self" || e.Timing != "while-concentrating" || e.DiceCount != 0 || e.DieSides != 0 || e.Bonus != 2 || *e.AddCastingModifier {
			return false
		}
		required = []string{"cast.cleric-self-slot-1", "defense.shield-of-faith-self", "concentration.self-600-seconds"}
	default:
		return false
	}
	return s.SourceID == s.ID && s.SourcePage >= 107 && s.SourcePage <= 175 && slices.Equal(s.RequiredMechanics, required)
}

// ComposeSpellSourceCatalogue combines offline source catalogues, excluding
// compatibility fixtures. Duplicate complete source IDs are errors, never wins
// based on input order. It does not register or grant playable actions.
func ComposeSpellSourceCatalogue(catalogues ...SpellRecords) (SpellRecords, error) {
	if len(catalogues) == 0 {
		return SpellRecords{}, fmt.Errorf("source catalogues required")
	}
	result := SpellRecords{SchemaVersion: 1, Source: catalogues[0].Source}
	result.Source.Changes = "SRD spell facts adapted into a composed offline source catalogue; bounded compatibility fixtures excluded. No executable effects or availability grants."
	for _, catalogue := range catalogues {
		if err := catalogue.Validate(); err != nil {
			return SpellRecords{}, err
		}
		for _, record := range catalogue.Records {
			if record.Scope == "complete-source-record" {
				result.Records = append(result.Records, record)
			}
		}
	}
	if err := result.Validate(); err != nil {
		return SpellRecords{}, err
	}
	return result, nil
}

// UnmarshalJSON distinguishes required numeric zero values from omitted/null
// fields; pointer flags and structural requirements are checked by Validate.
func (s *SpellRecord) UnmarshalJSON(data []byte) error {
	type plain SpellRecord
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var decoded plain
	if err := d.Decode(&decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	require := func(fields map[string]json.RawMessage, keys ...string) error {
		for _, key := range keys {
			raw := bytes.TrimSpace(fields[key])
			if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
				return fmt.Errorf("spell requires explicit %s", key)
			}
		}
		return nil
	}
	if err := require(fields, "level", "source_page", "duration_seconds"); err != nil {
		return err
	}
	for object, keys := range map[string][]string{"range": {"feet"}, "target": {"count"}, "scaling": {"above_level", "dice_count", "die_sides"}} {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(fields[object], &nested); err != nil {
			return err
		}
		if err := require(nested, keys...); err != nil {
			return err
		}
	}
	var effects []map[string]json.RawMessage
	if err := json.Unmarshal(fields["effects"], &effects); err != nil {
		return err
	}
	for _, effect := range effects {
		if err := require(effect, "dice_count", "die_sides", "bonus"); err != nil {
			return err
		}
	}
	*s = SpellRecord(decoded)
	return nil
}
