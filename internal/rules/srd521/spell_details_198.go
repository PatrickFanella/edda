package srd521

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// SpellBarkskin retains typed SRD source facts without implementing them.
type SpellBarkskin struct {
	MinimumAC   int    `json:"minimum_ac"`
	OnlyIfLower bool   `json:"only_if_lower"`
	Appearance  string `json:"appearance"`
	Willing     bool   `json:"willing"`
}

// SpellBeacon retains typed SRD source facts without implementing them.
type SpellBeacon struct {
	SaveAdvantage []string `json:"save_advantage"`
	Healing       string   `json:"healing"`
	TargetChoice  string   `json:"target_choice"`
}

// SpellBefuddlement retains typed SRD source facts without implementing them.
type SpellBefuddlement struct {
	Save             string   `json:"save"`
	SaveSuccess      string   `json:"save_success"`
	Prohibited       []string `json:"prohibited"`
	RepeatSaveDays   int      `json:"repeat_save_days"`
	RepeatSaveTiming string   `json:"repeat_save_timing"`
	RepeatSuccess    string   `json:"repeat_success"`
	EndsBy           []string `json:"ends_by"`
}

// SpellCurseChoicesEntry retains typed SRD source facts without implementing them.
type SpellCurseChoicesEntry struct {
	RollModifier  string   `json:"roll_modifier"`
	Choice        string   `json:"choice"`
	Abilities     []string `json:"abilities"`
	Rolls         []string `json:"rolls"`
	Save          string   `json:"save"`
	Timing        string   `json:"timing"`
	FailureAction string   `json:"failure_action"`
	Triggers      []string `json:"triggers"`
	DiceCount     int      `json:"dice_count"`
	DieSides      int      `json:"die_sides"`
	DamageType    string   `json:"damage_type"`
}

// SpellCurseSlotDurationsEntry retains typed SRD source facts without implementing them.
type SpellCurseSlotDurationsEntry struct {
	MinSlot        int  `json:"min_slot"`
	MaxSlot        int  `json:"max_slot"`
	Seconds        int  `json:"seconds"`
	Concentration  bool `json:"concentration"`
	UntilDispelled bool `json:"until_dispelled"`
}

// SpellCurse retains typed SRD source facts without implementing them.
type SpellCurse struct {
	SourcePages   []int                          `json:"source_pages"`
	Save          string                         `json:"save"`
	SaveSuccess   string                         `json:"save_success"`
	Choices       []SpellCurseChoicesEntry       `json:"choices"`
	ChooseCount   int                            `json:"choose_count"`
	SlotDurations []SpellCurseSlotDurationsEntry `json:"slot_durations"`
}

// SpellTentacles retains typed SRD source facts without implementing them.
type SpellTentacles struct {
	SideFeet         int      `json:"side_feet"`
	Placement        string   `json:"placement"`
	DifficultTerrain bool     `json:"difficult_terrain"`
	Save             string   `json:"save"`
	SaveTriggers     []string `json:"save_triggers"`
	SavesPerTurn     int      `json:"saves_per_turn"`
	SaveSuccess      string   `json:"save_success"`
	FailureCondition string   `json:"failure_condition"`
	EscapeAction     string   `json:"escape_action"`
	EscapeCheck      string   `json:"escape_check"`
	EscapeDC         string   `json:"escape_dc"`
	EscapeSuccess    string   `json:"escape_success"`
	Appearance       string   `json:"appearance"`
}

// SpellBarrierFormsEntry retains typed SRD source facts without implementing them.
type SpellBarrierFormsEntry struct {
	Choice        string `json:"choice"`
	LengthFeet    int    `json:"length_feet"`
	DiameterFeet  int    `json:"diameter_feet"`
	HeightFeet    int    `json:"height_feet"`
	ThicknessFeet int    `json:"thickness_feet"`
}

// SpellBarrier retains typed SRD source facts without implementing them.
type SpellBarrier struct {
	Forms            []SpellBarrierFormsEntry `json:"forms"`
	Dimensions       string                   `json:"dimensions"`
	Cover            string                   `json:"cover"`
	DifficultTerrain bool                     `json:"difficult_terrain"`
	Save             string                   `json:"save"`
	SaveSuccess      string                   `json:"save_success"`
	SaveTriggers     []string                 `json:"save_triggers"`
	SavesPerTurn     int                      `json:"saves_per_turn"`
	Appearance       string                   `json:"appearance"`
}

// SpellBless retains typed SRD source facts without implementing them.
type SpellBless struct {
	BaseTargets              int      `json:"base_targets"`
	AdditionalTargetsPerSlot int      `json:"additional_targets_per_slot"`
	Rolls                    []string `json:"rolls"`
	DieSides                 int      `json:"die_sides"`
	RollModifier             string   `json:"roll_modifier"`
	TargetChoice             string   `json:"target_choice"`
}

// SpellBlight retains typed SRD source facts without implementing them.
type SpellBlight struct {
	Variants              []string `json:"variants"`
	Save                  string   `json:"save"`
	SaveSuccess           string   `json:"save_success"`
	AutomaticFailureTypes []string `json:"automatic_failure_types"`
	PlantVariantSave      bool     `json:"plant_variant_save"`
	PlantVariantEffect    string   `json:"plant_variant_effect"`
	PlantExamples         []string `json:"plant_examples"`
}

// SpellBlindness retains typed SRD source facts without implementing them.
type SpellBlindness struct {
	Conditions               []string `json:"conditions"`
	ChoicePerTarget          bool     `json:"choice_per_target"`
	Save                     string   `json:"save"`
	RepeatSaveTiming         string   `json:"repeat_save_timing"`
	Success                  string   `json:"success"`
	BaseTargets              int      `json:"base_targets"`
	AdditionalTargetsPerSlot int      `json:"additional_targets_per_slot"`
}

// SpellBlink retains typed SRD source facts without implementing them.
type SpellBlink struct {
	SourcePages            []int    `json:"source_pages"`
	RollDieSides           int      `json:"roll_die_sides"`
	VanishMinRoll          int      `json:"vanish_min_roll"`
	VanishMaxRoll          int      `json:"vanish_max_roll"`
	RollTiming             string   `json:"roll_timing"`
	DestinationPlane       string   `json:"destination_plane"`
	AlreadyEthereal        string   `json:"already_ethereal"`
	PerceiveDeparturePlane bool     `json:"perceive_departure_plane"`
	PerceptionDistanceFeet int      `json:"perception_distance_feet"`
	PerceptionAppearance   string   `json:"perception_appearance"`
	Interaction            string   `json:"interaction"`
	OtherPlanePerception   string   `json:"other_plane_perception"`
	ReturnTriggers         []string `json:"return_triggers"`
	ReturnDistanceFeet     int      `json:"return_distance_feet"`
	ReturnSpace            string   `json:"return_space"`
	NoSpaceReturn          string   `json:"no_space_return"`
}

// SpellBlur retains typed SRD source facts without implementing them.
type SpellBlur struct {
	AttackRollsAgainstCaster string   `json:"attack_rolls_against_caster"`
	ImmuneAttackerPerception []string `json:"immune_attacker_perception"`
	Appearance               string   `json:"appearance"`
}

// SpellBurning retains typed SRD source facts without implementing them.
type SpellBurning struct {
	ConeFeet       int      `json:"cone_feet"`
	Save           string   `json:"save"`
	SaveSuccess    string   `json:"save_success"`
	IgnitedObjects []string `json:"ignited_objects"`
	Appearance     string   `json:"appearance"`
}

// Reflection here walks only the concrete schema types. It enforces presence at
// every nesting level without treating source-valid zero/false as absent.
func requireSpell198Fields(raw []byte, typ reflect.Type) error {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("required source value missing or null")
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return err
		}
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			key := strings.Split(f.Tag.Get("json"), ",")[0]
			if err := requireSpell198Fields(fields[key], f.Type); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	case reflect.Slice:
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		if len(values) == 0 {
			return fmt.Errorf("required source list empty")
		}
		for _, v := range values {
			if err := requireSpell198Fields(v, typ.Elem()); err != nil {
				return err
			}
		}
	case reflect.String:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return err
		}
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("blank source label")
		}
	case reflect.Int:
		var value int
		if err := json.Unmarshal(raw, &value); err != nil {
			return err
		}
		// This batch has no signed numeric quantities. Earlier Scrying
		// modifiers use their existing decoder and retain negative values.
		if value < 0 {
			return fmt.Errorf("negative source quantity")
		}
	default:
		value := reflect.New(typ).Interface()
		if err := json.Unmarshal(raw, value); err != nil {
			return err
		}
	}
	return nil
}

func spell198Templates() map[string]any {
	return map[string]any{
		"barkskin":     SpellBarkskin{},
		"beacon":       SpellBeacon{},
		"befuddlement": SpellBefuddlement{},
		"curse":        SpellCurse{},
		"tentacles":    SpellTentacles{},
		"barrier":      SpellBarrier{},
		"bless":        SpellBless{},
		"blight":       SpellBlight{},
		"blindness":    SpellBlindness{},
		"blink":        SpellBlink{},
		"blur":         SpellBlur{},
		"burning":      SpellBurning{},
	}
}

func (s SpellRecord) validateSpell198() (bool, error) {
	names := map[string]string{"barkskin": "barkskin", "beacon": "beacon-of-hope", "befuddlement": "befuddlement", "curse": "bestow-curse", "tentacles": "black-tentacles", "barrier": "blade-barrier", "bless": "bless", "blight": "blight", "blindness": "blindness-deafness", "blink": "blink", "blur": "blur", "burning": "burning-hands"}
	expected := ""
	for key, name := range names {
		if s.ID == prefix+"spells/"+name {
			expected = key
		}
	}
	if s.SourceDetails == nil {
		if expected != "" {
			return true, fmt.Errorf("batch 198 details required")
		}
		return false, nil
	}
	d := s.SourceDetails
	payloads := map[string]any{
		"barkskin":     d.Barkskin,
		"beacon":       d.Beacon,
		"befuddlement": d.Befuddlement,
		"curse":        d.Curse,
		"tentacles":    d.Tentacles,
		"barrier":      d.Barrier,
		"bless":        d.Bless,
		"blight":       d.Blight,
		"blindness":    d.Blindness,
		"blink":        d.Blink,
		"blur":         d.Blur,
		"burning":      d.Burning,
	}
	count := 0
	for key, value := range payloads {
		if reflect.ValueOf(value).IsNil() {
			continue
		}
		count++
		if expected != key {
			return true, fmt.Errorf("mismatched batch 198 payload")
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return true, err
		}
		if err := requireSpell198Fields(raw, reflect.TypeOf(value).Elem()); err != nil {
			return true, err
		}
	}
	if expected == "" && count == 0 {
		return false, nil
	}
	if count != 1 || s.Scope != "complete-source-record" || d.CastingSeconds != 0 || len(d.CastingTriggers) != 0 || d.CastingTriggers == nil || strings.TrimSpace(d.EndTiming) == "" {
		return true, fmt.Errorf("invalid batch 198 source details")
	}
	if d.Rays != nil || d.Scrying != nil || d.Smite != nil || d.Chest != nil || d.Vision != nil || d.Seeming != nil || d.Sending != nil || d.Sequester != nil || d.Shapechange != nil || d.Shatter != nil || d.Shield != nil || d.Faith != nil {
		return true, fmt.Errorf("multiple source payloads")
	}
	if d.Curse != nil {
		c := d.Curse
		choices := []string{}
		for _, v := range c.Choices {
			choices = append(choices, v.Choice)
		}
		if !slices.Equal(c.SourcePages, []int{112, 113}) {
			return true, fmt.Errorf("invalid curse source pages")
		}
		if !slices.Equal(choices, []string{"ability-disadvantage", "attacks-against-caster", "forced-dodge", "extra-necrotic"}) || c.ChooseCount != 1 || len(c.SlotDurations) != 5 {
			return true, fmt.Errorf("invalid curse choices")
		}
		want := []SpellCurseSlotDurationsEntry{{3, 3, 60, true, false}, {4, 4, 600, true, false}, {5, 6, 28800, false, false}, {7, 8, 86400, false, false}, {9, 9, 0, false, true}}
		if !slices.Equal(c.SlotDurations, want) {
			return true, fmt.Errorf("invalid curse slot durations")
		}
	}
	if d.Barrier != nil {
		b := d.Barrier
		if len(b.Forms) != 2 || b.Forms[0].Choice != "straight" || b.Forms[1].Choice != "ring" || b.SavesPerTurn != 1 {
			return true, fmt.Errorf("invalid wall forms or frequency")
		}
	}
	if d.Tentacles != nil && (d.Tentacles.SavesPerTurn != 1 || d.Tentacles.SideFeet <= 0) {
		return true, fmt.Errorf("invalid tentacles area or frequency")
	}
	if d.Blink != nil && !slices.Equal(d.Blink.SourcePages, []int{113, 114}) {
		return true, fmt.Errorf("invalid blink source pages")
	}
	if d.Blight != nil && !slices.Equal(d.Blight.Variants, []string{"visible-creature", "nonmagical-noncreature-plant"}) {
		return true, fmt.Errorf("invalid blight alternatives")
	}
	if d.Blindness != nil && !slices.Equal(d.Blindness.Conditions, []string{"blinded", "deafened"}) {
		return true, fmt.Errorf("invalid condition alternatives")
	}
	return true, nil
}
