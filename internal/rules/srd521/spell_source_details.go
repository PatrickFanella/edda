package srd521

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// SpellSourceDetails retains the batch's additional source shapes as offline
// metadata. These fields never select or implement a runtime resolver.
type SpellSourceDetails struct {
	Barkskin        *SpellBarkskin     `json:"barkskin,omitempty"`
	Beacon          *SpellBeacon       `json:"beacon,omitempty"`
	Befuddlement    *SpellBefuddlement `json:"befuddlement,omitempty"`
	Curse           *SpellCurse        `json:"curse,omitempty"`
	Tentacles       *SpellTentacles    `json:"tentacles,omitempty"`
	Barrier         *SpellBarrier      `json:"barrier,omitempty"`
	Bless           *SpellBless        `json:"bless,omitempty"`
	Blight          *SpellBlight       `json:"blight,omitempty"`
	Blindness       *SpellBlindness    `json:"blindness,omitempty"`
	Blink           *SpellBlink        `json:"blink,omitempty"`
	Blur            *SpellBlur         `json:"blur,omitempty"`
	Burning         *SpellBurning      `json:"burning,omitempty"`
	CastingSeconds  int                `json:"casting_seconds"`
	CastingTriggers []string           `json:"casting_triggers"`
	EndTiming       string             `json:"end_timing"`
	Rays            *SpellRays         `json:"rays,omitempty"`
	Scrying         *SpellScrying      `json:"scrying,omitempty"`
	Smite           *SpellSmite        `json:"smite,omitempty"`
	Chest           *SpellChest        `json:"chest,omitempty"`
	Vision          *SpellVision       `json:"vision,omitempty"`
	Seeming         *SpellSeeming      `json:"seeming,omitempty"`
	Sending         *SpellSending      `json:"sending,omitempty"`
	Sequester       *SpellSequester    `json:"sequester,omitempty"`
	Shapechange     *SpellShapechange  `json:"shapechange,omitempty"`
	Shatter         *SpellShatter      `json:"shatter,omitempty"`
	Shield          *SpellShield       `json:"shield,omitempty"`
	Faith           *SpellFaith        `json:"faith,omitempty"`
}

// SpellRays describes independently aimed ranged spell attacks.
type SpellRays struct {
	AdditionalPerSlot int      `json:"additional_per_slot"`
	Count             int      `json:"count"`
	Distribution      []string `json:"distribution"`
	AttackPerRay      bool     `json:"attack_per_ray"`
}

// SpellSaveModifier is a source-listed knowledge or physical-connection choice.
type SpellSaveModifier struct {
	Choice   string `json:"choice"`
	Modifier int    `json:"modifier"`
}

// SpellScrying preserves the creature and stationary-location variants.
type SpellScrying struct {
	Variants               []string            `json:"variants"`
	CreatureSamePlane      bool                `json:"creature_same_plane"`
	Save                   string              `json:"save"`
	SaveAwareness          string              `json:"save_awareness"`
	Knowledge              []SpellSaveModifier `json:"knowledge"`
	Connections            []SpellSaveModifier `json:"connections"`
	SuccessRetryHours      int                 `json:"success_retry_hours"`
	SensorDistanceFeet     int                 `json:"sensor_distance_feet"`
	SensorInvisible        bool                `json:"sensor_invisible"`
	SensorIntangible       bool                `json:"sensor_intangible"`
	Senses                 []string            `json:"senses"`
	SensorFollowsCreature  bool                `json:"sensor_follows_creature"`
	VisibleSensor          string              `json:"visible_sensor"`
	LocationPreviouslySeen bool                `json:"location_previously_seen"`
	LocationSensorMoves    bool                `json:"location_sensor_moves"`
	FocusExamples          []string            `json:"focus_examples"`
}

// SpellSmite describes damage followed by a repeating end save.
type SpellSmite struct {
	Save             string `json:"save"`
	SaveTiming       string `json:"save_timing"`
	SuccessEnds      bool   `json:"success_ends"`
	FailureContinues bool   `json:"failure_continues"`
}

// SpellChest preserves both material objects and the recall/return lifecycle.
type SpellChest struct {
	DecayTiming            string   `json:"decay_timing"`
	ChestSizeFeet          []int    `json:"chest_size_feet"`
	ChestMinimumCostCP     int      `json:"chest_minimum_cost_cp"`
	ReplicaMinimumCostCP   int      `json:"replica_minimum_cost_cp"`
	ReplicaSize            string   `json:"replica_size"`
	Materials              string   `json:"materials"`
	BothTouched            bool     `json:"both_touched"`
	CapacityCubicFeet      int      `json:"capacity_cubic_feet"`
	Contents               string   `json:"contents"`
	Plane                  string   `json:"plane"`
	RecallAction           string   `json:"recall_action"`
	RecallTouch            string   `json:"recall_touch"`
	RecallDistanceFeet     int      `json:"recall_distance_feet"`
	RecallSpace            string   `json:"recall_space"`
	ReturnAction           string   `json:"return_action"`
	ReturnTouch            []string `json:"return_touch"`
	DecayAfterDays         int      `json:"decay_after_days"`
	DailyCumulativePercent int      `json:"daily_cumulative_percent"`
	EndTriggers            []string `json:"end_triggers"`
	EndsOnEthereal         string   `json:"ends_on_ethereal"`
}

// SpellVision preserves invisible and Ethereal perception without granting it.
type SpellVision struct {
	Subjects           []string `json:"subjects"`
	InvisibleAsVisible bool     `json:"invisible_as_visible"`
	SeeEthereal        bool     `json:"see_ethereal"`
	EtherealAppearance string   `json:"ethereal_appearance"`
}

// SpellSeeming preserves each target's appearance choices and discovery checks.
type SpellSeeming struct {
	Save                  string   `json:"save"`
	SaveEligibility       string   `json:"save_eligibility"`
	SaveSuccess           string   `json:"save_success"`
	AppearanceChoices     []string `json:"appearance_choices"`
	Affects               []string `json:"affects"`
	HeightDeltaFeet       int      `json:"height_delta_feet"`
	WeightChoices         []string `json:"weight_choices"`
	RetainLimbArrangement bool     `json:"retain_limb_arrangement"`
	OtherAppearance       string   `json:"other_appearance"`
	PhysicalInspection    string   `json:"physical_inspection"`
	DiscoveryAction       string   `json:"discovery_action"`
	DiscoveryCheck        string   `json:"discovery_check"`
	DiscoveryDC           string   `json:"discovery_dc"`
	DiscoverySuccess      string   `json:"discovery_success"`
}

// SpellSending records addressing alternatives, reply and recipient blocking.
type SpellSending struct {
	ReplyMaxWords            int      `json:"reply_max_words"`
	BlockTiming              string   `json:"block_timing"`
	AddressChoices           []string `json:"address_choices"`
	MaxWords                 int      `json:"max_words"`
	Delivery                 string   `json:"delivery"`
	RecognizeKnownSender     bool     `json:"recognize_known_sender"`
	UnderstandMeaning        bool     `json:"understand_meaning"`
	Reply                    string   `json:"reply"`
	CrossPlaneFailurePercent int      `json:"cross_plane_failure_percent"`
	SenderKnowsFailure       bool     `json:"sender_knows_failure"`
	RecipientMayBlock        bool     `json:"recipient_may_block"`
	BlockHours               int      `json:"block_hours"`
	BlockedAttempt           string   `json:"blocked_attempt"`
}

// SpellSequester preserves protections, creature suspension and early endings.
type SpellSequester struct {
	TargetChoices       []string `json:"target_choices"`
	Condition           string   `json:"condition"`
	Protections         []string `json:"protections"`
	CreatureCondition   string   `json:"creature_condition"`
	CreatureAges        bool     `json:"creature_ages"`
	UnneededSustenance  []string `json:"unneeded_sustenance"`
	EarlyEndCondition   string   `json:"early_end_condition"`
	EarlyEndRadiusMiles int      `json:"early_end_radius_miles"`
	EarlyEndExamples    []string `json:"early_end_examples"`
	DamageEnds          bool     `json:"damage_ends"`
}

// SpellShapechange preserves eligible forms, retained statistics and equipment choices.
type SpellShapechange struct {
	FormCRLimit                 string   `json:"form_cr_limit"`
	PreviouslySeen              bool     `json:"previously_seen"`
	ExcludedTypes               []string `json:"excluded_types"`
	ChangeAction                string   `json:"change_action"`
	TemporaryHP                 string   `json:"temporary_hp"`
	TemporaryHPGranted          string   `json:"temporary_hp_granted"`
	TemporaryHPEndsWithSpell    bool     `json:"temporary_hp_ends_with_spell"`
	ReplaceStatistics           string   `json:"replace_statistics"`
	Retained                    []string `json:"retained"`
	RetainSpellcastingIfPresent bool     `json:"retain_spellcasting_if_present"`
	EquipmentChoices            []string `json:"equipment_choices"`
}

// SpellShatter preserves area, saves and the separate object eligibility rule.
type SpellShatter struct {
	Center            string   `json:"center"`
	RadiusFeet        int      `json:"radius_feet"`
	Save              string   `json:"save"`
	SaveSuccess       string   `json:"save_success"`
	DisadvantageTypes []string `json:"disadvantage_types"`
	Objects           []string `json:"objects"`
	ObjectDamage      string   `json:"object_damage"`
	Sound             string   `json:"sound"`
}

// SpellShield describes the triggering attack and Magic Missile protection.
type SpellShield struct {
	IncludesTriggeringAttack bool     `json:"includes_triggering_attack"`
	NoDamageFrom             []string `json:"no_damage_from"`
	Appearance               string   `json:"appearance"`
}

// SpellFaith describes the full source target, separately from the self fixture.
type SpellFaith struct {
	TargetChoice string `json:"target_choice"`
	Appearance   string `json:"appearance"`
}

// UnmarshalJSON requires explicit fields within every selected typed payload.
// This prevents omitted false/zero source facts silently taking Go defaults.
func (s *SpellSourceDetails) UnmarshalJSON(data []byte) error {
	type plain SpellSourceDetails
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
	for _, key := range []string{"casting_seconds", "casting_triggers", "end_timing"} {
		if len(fields[key]) == 0 || bytes.Equal(bytes.TrimSpace(fields[key]), []byte("null")) {
			return fmt.Errorf("source details require %s", key)
		}
	}
	templates := map[string]any{"rays": SpellRays{}, "scrying": SpellScrying{}, "smite": SpellSmite{}, "chest": SpellChest{}, "vision": SpellVision{}, "seeming": SpellSeeming{}, "sending": SpellSending{}, "sequester": SpellSequester{}, "shapechange": SpellShapechange{}, "shatter": SpellShatter{}, "shield": SpellShield{}, "faith": SpellFaith{}}
	for key, template := range spell198Templates() {
		if raw, ok := fields[key]; ok {
			if err := requireSpell198Fields(raw, reflect.TypeOf(template)); err != nil {
				return err
			}
		}
	}
	for key, template := range templates {
		raw, ok := fields[key]
		if !ok {
			continue
		}
		b, err := json.Marshal(template)
		if err != nil {
			return err
		}
		var required, actual map[string]json.RawMessage
		if err := json.Unmarshal(b, &required); err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &actual); err != nil {
			return err
		}
		for field := range required {
			value := bytes.TrimSpace(actual[field])
			if len(value) == 0 || bytes.Equal(value, []byte("null")) {
				return fmt.Errorf("%s requires %s", key, field)
			}
			var content any
			if err := json.Unmarshal(value, &content); err != nil {
				return err
			}
			switch content := content.(type) {
			case string:
				if strings.TrimSpace(content) == "" {
					return fmt.Errorf("%s has invalid %s", key, field)
				}
			case []any:
				if len(content) == 0 {
					return fmt.Errorf("%s has empty %s", key, field)
				}
				for _, item := range content {
					text, isText := item.(string)
					if item == nil || (isText && strings.TrimSpace(text) == "") {
						return fmt.Errorf("%s has invalid %s entry", key, field)
					}
				}
			case float64:
				if content < 0 {
					return fmt.Errorf("%s has invalid %s", key, field)
				}
			}
		}
	}
	if decoded.Scrying != nil {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(fields["scrying"], &raw); err != nil {
			return err
		}
		for _, group := range []string{"knowledge", "connections"} {
			var rows []map[string]json.RawMessage
			if err := json.Unmarshal(raw[group], &rows); err != nil {
				return err
			}
			for _, row := range rows {
				var choice string
				if err := json.Unmarshal(row["choice"], &choice); err != nil || strings.TrimSpace(choice) == "" || len(row["modifier"]) == 0 || bytes.Equal(bytes.TrimSpace(row["modifier"]), []byte("null")) {
					return fmt.Errorf("incomplete scrying choice")
				}
			}
		}
	}
	*s = SpellSourceDetails(decoded)
	return nil
}

func validSpellTarget(t SpellTarget) bool {
	switch t.Shape {
	case "square":
		return t.Count == 0 && t.Eligibility == "creatures-in-visible-ground-area"
	case "wall":
		return t.Count == 0 && t.Eligibility == "creatures-in-wall"
	case "cone":
		return t.Count == 0 && t.Eligibility == "creatures-and-flammable-unattended-objects"
	case "up-to-count":
		return (t.Count == 3 && t.Eligibility == "creatures-in-range") || (t.Count == 1 && t.Eligibility == "visible-creatures-in-range")
	case "single":
		return t.Count == 1 && slices.Contains([]string{"target-in-range", "creature-in-range", "conscious-self", "self", "hit-target", "object-or-willing-creature", "chest-and-replica", "willing-creature-touch", "creature-touch", "visible-creature-in-range"}, t.Eligibility)
	case "rays":
		return t.Count == 3 && t.Eligibility == "targets-in-range"
	case "choice":
		return t.Count == 1 && (t.Eligibility == "same-plane-creature-or-seen-location" || t.Eligibility == "visible-creature-or-nonmagical-noncreature-plant")
	case "selected-creatures":
		return t.Count == 0 && (t.Eligibility == "visible-creatures-in-range" || t.Eligibility == "creatures-in-range")
	case "sphere":
		return t.Count == 0 && t.Eligibility == "creatures-and-eligible-objects"
	default:
		return false
	}
}

func (s SpellRecord) validateSourceDetails() error {
	if handled, err := s.validateSpell198(); handled || err != nil {
		return err
	}
	d := s.SourceDetails
	if d == nil {
		if s.Scope == "complete-source-record" && slices.Contains([]string{"Scorching Ray", "Scrying", "Searing Smite", "Secret Chest", "See Invisibility", "Seeming", "Sending", "Sequester", "Shapechange", "Shatter", "Shield", "Shield of Faith"}, s.Name) {
			return fmt.Errorf("batch source details required")
		}
		if s.CastingTime == "reaction" || s.CastingTime == "10-minutes" || s.Duration == "until-dispelled" || s.Range.Kind == "unlimited" || s.Target.Shape != "single" {
			return fmt.Errorf("source details required for extended shape")
		}
		return nil
	}
	if s.Scope != "complete-source-record" || d.EndTiming == "" || d.CastingSeconds < 0 || d.CastingTriggers == nil {
		return fmt.Errorf("invalid source details")
	}
	if (s.CastingTime == "10-minutes") != (d.CastingSeconds == 600) || (s.CastingTime == "reaction" && len(d.CastingTriggers) == 0) {
		return fmt.Errorf("invalid casting details")
	}
	payloads := []bool{d.Rays != nil, d.Scrying != nil, d.Smite != nil, d.Chest != nil, d.Vision != nil, d.Seeming != nil, d.Sending != nil, d.Sequester != nil, d.Shapechange != nil, d.Shatter != nil, d.Shield != nil, d.Faith != nil}
	names := []string{"scorching-ray", "scrying", "searing-smite", "secret-chest", "see-invisibility", "seeming", "sending", "sequester", "shapechange", "shatter", "shield", "shield-of-faith"}
	count := 0
	for i, present := range payloads {
		if present {
			count++
			if s.ID != prefix+"spells/"+names[i] {
				return fmt.Errorf("mismatched source payload")
			}
		}
	}
	if count != 1 {
		return fmt.Errorf("exactly one batch payload required")
	}
	if d.Scrying != nil {
		if len(d.Scrying.Variants) != 2 || len(d.Scrying.Knowledge) != 3 || len(d.Scrying.Connections) != 3 || d.Scrying.Save != "wisdom" || d.Scrying.SensorDistanceFeet <= 0 || d.Scrying.SuccessRetryHours <= 0 {
			return fmt.Errorf("invalid scrying choices")
		}
	}
	if d.Chest != nil && (len(d.Chest.ChestSizeFeet) != 3 || d.Chest.ChestMinimumCostCP <= 0 || d.Chest.ReplicaMinimumCostCP <= 0 || d.Chest.DailyCumulativePercent <= 0 || d.Chest.DailyCumulativePercent > 100) {
		return fmt.Errorf("invalid chest dimensions, costs or decay")
	}
	if d.Shapechange != nil && (len(d.Shapechange.ExcludedTypes) != 2 || len(d.Shapechange.Retained) == 0 || len(d.Shapechange.EquipmentChoices) != 2) {
		return fmt.Errorf("invalid form choices")
	}
	if d.Shatter != nil && (d.Shatter.RadiusFeet <= 0 || d.Shatter.Save != "constitution") {
		return fmt.Errorf("invalid shatter area or save")
	}
	if d.Sending != nil && (d.Sending.MaxWords <= 0 || d.Sending.CrossPlaneFailurePercent < 0 || d.Sending.CrossPlaneFailurePercent > 100 || d.Sending.BlockHours <= 0) {
		return fmt.Errorf("invalid sending limits")
	}
	if d.Rays != nil && (d.Rays.Count != 3 || d.Rays.AdditionalPerSlot <= 0 || !slices.Equal(d.Rays.Distribution, []string{"one-target", "several-targets"}) || !d.Rays.AttackPerRay) {
		return fmt.Errorf("invalid ray choices")
	}
	if d.Scrying != nil && (!slices.Equal(d.Scrying.Variants, []string{"creature", "seen-location"}) || !d.Scrying.CreatureSamePlane || !d.Scrying.LocationPreviouslySeen || d.Scrying.LocationSensorMoves) {
		return fmt.Errorf("invalid scrying variants")
	}
	if d.Smite != nil && (d.Smite.Save != "constitution" || d.Smite.SaveTiming != "after-start-of-turn-damage" || !d.Smite.SuccessEnds || !d.Smite.FailureContinues || len(d.CastingTriggers) != 2) {
		return fmt.Errorf("invalid smite save or trigger")
	}
	if d.Seeming != nil && (d.Seeming.HeightDeltaFeet < 0 || d.Seeming.Save != "charisma" || len(d.Seeming.AppearanceChoices) != 2 || len(d.Seeming.WeightChoices) != 2) {
		return fmt.Errorf("invalid seeming choices")
	}
	if d.Sequester != nil && (d.Sequester.EarlyEndRadiusMiles <= 0 || !slices.Equal(d.Sequester.TargetChoices, []string{"object", "willing-creature"}) || d.Sequester.CreatureAges || !slices.Equal(d.Sequester.UnneededSustenance, []string{"food", "water", "air"})) {
		return fmt.Errorf("invalid sequester choices")
	}
	if d.Shield != nil && (!slices.Equal(d.CastingTriggers, []string{"hit-by-attack-roll", "targeted-by-magic-missile"}) || d.EndTiming != "start-of-caster-next-turn") {
		return fmt.Errorf("invalid shield reaction")
	}

	return nil
}
