package srd521

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func loadSpells218(t *testing.T) SpellRecords {
	t.Helper()
	b, err := os.ReadFile("testdata/spells_218.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := DecodeSpellRecords(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Values transcribed from pinned PDF pp.159–162, not inferred from runtime
// fixtures. The table deliberately includes Seeming/Shield continuation pages.
func TestSpells218SourceHeaders(t *testing.T) {
	expected := []struct {
		name                              string
		page, level                       int
		school, lists, casting, rangeKind string
		feet                              int
		shape                             string
		count                             int
		eligibility, material             string
		cost                              int
		consumed, somatic, concentration  bool
		duration                          string
		seconds                           int
	}{
		{"Scorching Ray", 159, 2, "evocation", "sorcerer wizard", "action", "feet", 120, "rays", 3, "targets-in-range", "", 0, false, true, false, "instantaneous", 0},
		{"Scrying", 160, 5, "divination", "bard cleric druid warlock wizard", "10-minutes", "self", 0, "choice", 1, "same-plane-creature-or-seen-location", "a focus worth 1,000+ GP, such as a crystal ball, mirror, or water-filled font", 100000, false, true, true, "timed", 600},
		{"Searing Smite", 160, 1, "evocation", "paladin", "bonus-action", "self", 0, "single", 1, "hit-target", "", 0, false, false, false, "timed", 60},
		{"Secret Chest", 160, 4, "conjuration", "wizard", "action", "touch", 0, "single", 1, "chest-and-replica", "a rare-material chest worth 5,000+ GP and a Tiny replica made from the same materials worth 50+ GP", 505000, false, true, false, "until-dispelled", 0},
		{"See Invisibility", 160, 2, "divination", "bard sorcerer wizard", "action", "self", 0, "single", 1, "self", "a pinch of talc", 0, false, true, false, "timed", 3600},
		{"Seeming", 160, 5, "illusion", "bard sorcerer wizard", "action", "feet", 30, "selected-creatures", 0, "visible-creatures-in-range", "", 0, false, true, false, "timed", 28800},
		{"Sending", 161, 3, "divination", "bard cleric wizard", "action", "unlimited", 0, "single", 1, "target-in-range", "a copper wire", 0, false, true, false, "instantaneous", 0},
		{"Sequester", 161, 7, "transmutation", "wizard", "action", "touch", 0, "single", 1, "object-or-willing-creature", "gem dust worth 5,000+ GP", 500000, true, true, false, "until-dispelled", 0},
		{"Shapechange", 161, 9, "transmutation", "druid wizard", "action", "self", 0, "single", 1, "self", "a jade circlet worth 1,500+ GP", 150000, false, true, true, "timed", 3600},
		{"Shatter", 161, 2, "evocation", "bard sorcerer wizard", "action", "feet", 60, "sphere", 0, "creatures-and-eligible-objects", "a chip of mica", 0, false, true, false, "instantaneous", 0},
		{"Shield", 161, 1, "abjuration", "sorcerer wizard", "reaction", "self", 0, "single", 1, "self", "", 0, false, true, false, "timed", 6},
		{"Shield of Faith", 162, 1, "abjuration", "cleric paladin", "bonus-action", "feet", 60, "single", 1, "creature-in-range", "a prayer scroll", 0, false, true, true, "timed", 600},
	}
	r := loadSpells218(t)
	if len(r.Records) != 12 || r.Source.SHA256 != artifactSHA256 || r.Source.Attribution != Attribution || r.Source.License != "CC-BY-4.0" || r.Source.Changes == "" {
		t.Fatal("batch size or pin/license/adaptation changed")
	}
	for i, want := range expected {
		t.Run(want.name, func(t *testing.T) {
			s := r.Records[i]
			id := prefix + "spells/" + strings.ReplaceAll(strings.ToLower(want.name), " ", "-")
			if s.Name != want.name || s.ID != id || s.SourceID != id || s.SourcePage != want.page || s.Scope != "complete-source-record" || s.Level != want.level || s.School != want.school || !slices.Equal(s.Lists, strings.Fields(want.lists)) || s.CastingTime != want.casting || s.Range != (SpellRange{want.rangeKind, want.feet}) || s.Target != (SpellTarget{want.shape, want.count, want.eligibility}) {
				t.Fatal("source identity/casting/target mismatch")
			}
			if !*s.Components.Verbal || *s.Components.Somatic != want.somatic || *s.Concentration != want.concentration || *s.Ritual || s.Duration != want.duration || s.DurationSeconds != want.seconds {
				t.Fatal("source flags/duration mismatch")
			}
			if want.material == "" {
				if s.Components.Material != nil {
					t.Fatal("unexpected material")
				}
			} else {
				m := s.Components.Material
				if m == nil || m.Description != want.material || *m.CostCP != want.cost || *m.Consumed != want.consumed {
					t.Fatal("source material mismatch")
				}
			}
			slots := []int{}
			for slot := want.level; slot <= 9; slot++ {
				slots = append(slots, slot)
			}
			if !slices.Equal(s.SlotLevels, slots) || s.MechanicsAvailable() {
				t.Fatal("source slot metadata or fail-closed behavior changed")
			}
		})
	}
}

// These explicit source-value fixtures retain every nested choice and source
// numeric value on pp.159–162. Strict decoding separately requires all fields.
func TestSpells218SourceChoices(t *testing.T) {
	expected := []string{
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"instantaneous","rays":{"additional_per_slot":1,"count":3,"distribution":["one-target","several-targets"],"attack_per_ray":true}}`,
		`{"casting_seconds":600,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","scrying":{"variants":["creature","seen-location"],"creature_same_plane":true,"save":"wisdom","save_awareness":"unaware-of-spell-but-feels-uneasy","knowledge":[{"choice":"secondhand-heard-of","modifier":5},{"choice":"firsthand-met","modifier":0},{"choice":"extensive-known-well","modifier":-5}],"connections":[{"choice":"picture-or-other-likeness","modifier":-2},{"choice":"garment-or-other-possession","modifier":-4},{"choice":"body-part-lock-of-hair-or-bit-of-nail","modifier":-10}],"success_retry_hours":24,"sensor_distance_feet":10,"sensor_invisible":true,"sensor_intangible":true,"senses":["see-as-present","hear-as-present"],"sensor_follows_creature":true,"visible_sensor":"luminous-fist-sized-orb","location_previously_seen":true,"location_sensor_moves":false,"focus_examples":["crystal-ball","mirror","water-filled-font"]}}`,
		`{"casting_seconds":0,"casting_triggers":["immediately-after-hit-with-melee-weapon","immediately-after-hit-with-unarmed-strike"],"end_timing":"duration-ended","smite":{"save":"constitution","save_timing":"after-start-of-turn-damage","success_ends":true,"failure_continues":true}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"dispelled-or-explicit-end","chest":{"decay_timing":"end-of-each-day","chest_size_feet":[3,2,2],"chest_minimum_cost_cp":500000,"replica_minimum_cost_cp":5000,"replica_size":"tiny","materials":"rare-materials-same-for-chest-and-replica","both_touched":true,"capacity_cubic_feet":12,"contents":"nonliving-material","plane":"ethereal","recall_action":"magic","recall_touch":"replica","recall_distance_feet":5,"recall_space":"unoccupied-ground","return_action":"magic","return_touch":["chest","replica"],"decay_after_days":60,"daily_cumulative_percent":5,"end_triggers":["daily-cumulative-chance-after-60-days","cast-spell-again","replica-destroyed"],"ends_on_ethereal":"chest-remains-for-anyone-to-find"}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-ended","vision":{"subjects":["creatures","objects"],"invisible_as_visible":true,"see_ethereal":true,"ethereal_appearance":"ghostly"}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-ended","seeming":{"save":"charisma","save_eligibility":"unwilling-target","save_success":"unaffected","appearance_choices":["same-for-all","different-per-target"],"affects":["bodies","equipment"],"height_delta_feet":1,"weight_choices":["heavier","lighter"],"retain_limb_arrangement":true,"other_appearance":"caster-choice","physical_inspection":"fails-objects-pass-through-added-features","discovery_action":"study","discovery_check":"intelligence-investigation","discovery_dc":"caster-spell-save-dc","discovery_success":"aware-target-is-disguised"}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"instantaneous","sending":{"reply_max_words":25,"block_timing":"upon-receiving-message","address_choices":["creature-caster-met","creature-described-by-someone-who-met-it"],"max_words":25,"delivery":"heard-in-target-mind","recognize_known_sender":true,"understand_meaning":true,"reply":"immediate-in-like-manner","cross_plane_failure_percent":5,"sender_knows_failure":true,"recipient_may_block":true,"block_hours":8,"blocked_attempt":"sender-learns-block-and-spell-fails"}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"dispelled-or-explicit-end","sequester":{"target_choices":["object","willing-creature"],"condition":"invisible","protections":["cannot-be-targeted-by-divination-spells","cannot-be-detected-by-magic","cannot-be-viewed-remotely-by-magic"],"creature_condition":"unconscious-suspended-animation","creature_ages":false,"unneeded_sustenance":["food","water","air"],"early_end_condition":"optional-caster-chosen-condition-occurs-or-is-visible-within-radius","early_end_radius_miles":1,"early_end_examples":["after-1000-years","when-tarrasque-awakens"],"damage_ends":true}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","shapechange":{"form_cr_limit":"no-higher-than-caster-level-or-challenge-rating","previously_seen":true,"excluded_types":["construct","undead"],"change_action":"magic-to-different-eligible-form","temporary_hp":"first-form-hit-points","temporary_hp_granted":"on-cast-only","temporary_hp_ends_with_spell":true,"replace_statistics":"chosen-form-stat-block-except-retained","retained":["creature-type","alignment","personality","intelligence","wisdom","charisma","hit-points","hit-point-dice","proficiencies","ability-to-communicate"],"retain_spellcasting_if_present":true,"equipment_choices":["drops-to-ground","changes-size-and-shape-to-fit-while-in-form"]}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"instantaneous","shatter":{"center":"chosen-point-in-range","radius_feet":10,"save":"constitution","save_success":"half-damage","disadvantage_types":["construct"],"objects":["nonmagical","not-worn","not-carried","in-area"],"object_damage":"same-damage-no-save","sound":"loud-noise"}}`,
		`{"casting_seconds":0,"casting_triggers":["hit-by-attack-roll","targeted-by-magic-missile"],"end_timing":"start-of-caster-next-turn","shield":{"includes_triggering_attack":true,"no_damage_from":["magic-missile"],"appearance":"imperceptible-force-barrier"}}`,
		`{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","faith":{"target_choice":"creature-caster-chooses-within-range","appearance":"shimmering-field"}}`,
	}
	for i, s := range loadSpells218(t).Records {
		t.Run(s.Name, func(t *testing.T) {
			var want SpellSourceDetails
			if err := json.Unmarshal([]byte(expected[i]), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(s.SourceDetails, &want) {
				t.Fatalf("PDF p.%d choices mismatch: %+v", s.SourcePage, s.SourceDetails)
			}
		})
	}
}

func TestSpells218EffectsScalingDependencies(t *testing.T) {
	// id:dependency:timing:dice:sides:type:bonus, in source order.
	expected := [][]string{
		{"ray:attack.ranged-spell-per-ray:on-hit-per-ray:2:6:fire:0"},
		{"creature-sensor:divination.scrying-creature:failed-wisdom-save:0:0::0", "location-sensor:divination.scrying-location:on-cast-location-mode:0:0::0"},
		{"initial:damage.searing-smite-hit:as-hit-extra-attack-damage:1:6:fire:0", "recurring:damage.searing-smite-turn:start-each-target-turn-before-save:1:6:fire:0"},
		{"hide:transport.secret-chest-ethereal:on-cast-touch-both:0:0::0", "recall:transport.secret-chest-recall:magic-action-touch-replica:0:0::0", "return:transport.secret-chest-return:magic-action-touch-both:0:0::0"},
		{"vision:senses.invisible-and-ethereal:for-duration:0:0::0"},
		{"disguise:illusion.seeming-appearance:on-cast-willing-or-failed-save:0:0::0"},
		{"message:communication.sending:on-cast-unless-blocked-or-cross-plane-failure:0:0::0", "reply:communication.sending-reply:immediate:0:0::0"},
		{"conceal:protection.sequester:for-duration:0:0::0", "suspend:condition.sequester-creature:creature-only-for-duration:0:0::0"},
		{"form:transformation.shapechange:on-cast-and-magic-action-change:0:0::0", "temporary-hp:temporary-hp.first-shapechange-form:on-cast-only:0:0::0"},
		{"area-damage:damage.shatter:failed-save-or-eligible-object:3:8:thunder:0"},
		{"ac:defense.shield-ac:until-start-next-caster-turn:0:0::5", "missiles:defense.shield-magic-missile:until-start-next-caster-turn:0:0::0"},
		{"ac:defense.shield-of-faith-creature:for-duration:0:0::2"},
	}
	deps := []string{
		"attack.ranged-spell-per-ray damage.fire scaling.per-slot-ray",
		"components.material concentration.source-duration divination.scrying-creature divination.scrying-location save.wisdom-knowledge-connection-modifiers divination.scrying-success-24h-lockout sensor.invisible-intangible-moving-or-stationary cast.long-600-seconds",
		"damage.searing-smite-hit damage.searing-smite-turn cast.bonus-action-after-hit save.constitution-end-on-success scaling.per-slot",
		"components.material transport.secret-chest-ethereal transport.secret-chest-recall transport.secret-chest-return duration.secret-chest-cumulative-decay duration.secret-chest-recast-replica-end components.two-priced-objects",
		"components.material senses.invisible-and-ethereal",
		"illusion.seeming-appearance save.charisma-unwilling check.study-investigation-spell-dc illusion.physical-inspection",
		"components.material communication.sending communication.sending-reply communication.cross-plane-5-percent-failure communication.recipient-8h-block",
		"components.material protection.sequester condition.sequester-creature duration.sequester-condition-or-damage components.cost-consumption",
		"components.material concentration.source-duration transformation.shapechange temporary-hp.first-shapechange-form transformation.eligible-seen-cr-excluding-construct-undead transformation.retained-statistics transformation.equipment-choice",
		"components.material damage.shatter save.constitution-half save.construct-disadvantage area.sphere-10-feet damage.shatter-unattended-nonmagical-objects scaling.per-slot",
		"defense.shield-ac defense.shield-magic-missile reaction.hit-or-magic-missile defense.retroactive-triggering-attack",
		"components.material concentration.source-duration defense.shield-of-faith-creature",
	}
	for i, s := range loadSpells218(t).Records {
		t.Run(s.Name, func(t *testing.T) {
			got := []string{}
			for _, e := range s.Effects {
				got = append(got, fmt.Sprintf("%s:%s:%s:%d:%d:%s:%d", e.ID, e.Mechanic, e.Timing, e.DiceCount, e.DieSides, e.DamageType, e.Bonus))
				if *e.AddCastingModifier || *e.HalfInitialOnly {
					t.Fatal("unexpected modifier/miss behavior")
				}
			}
			if !slices.Equal(got, expected[i]) || !slices.Equal(s.RequiredMechanics, strings.Fields(deps[i])) {
				t.Fatal("source effect/dependency mismatch")
			}
			want := SpellScaling{Kind: "none", EffectIDs: []string{}}
			switch s.Name {
			case "Scorching Ray":
				want = SpellScaling{"slot-count", 2, 0, 0, []string{"ray"}}
			case "Searing Smite":
				want = SpellScaling{"slot", 1, 1, 6, []string{"initial", "recurring"}}
			case "Shatter":
				want = SpellScaling{"slot", 2, 1, 8, []string{"area-damage"}}
			}
			if !reflect.DeepEqual(s.Scaling, want) {
				t.Fatal("source scaling mismatch")
			}
		})
	}
}

func TestSpells218CompositionSeparatesCompatibility(t *testing.T) {
	batch := loadSpells218(t)
	legacy := loadSpellRecords(t)
	composed, err := ComposeSpellSourceCatalogue(legacy, batch)
	if err != nil {
		t.Fatal(err)
	}
	if len(composed.Records) != 13 {
		t.Fatal("expected Acid Arrow plus exact twelve source families")
	}
	count := 0
	for _, s := range composed.Records {
		if s.Scope != "complete-source-record" || s.MechanicsAvailable() {
			t.Fatal("source composition granted compatibility availability")
		}
		if s.Name == "Shield of Faith" {
			count++
			if s.Target.Eligibility != "creature-in-range" || len(s.SlotLevels) != 9 {
				t.Fatal("full source replaced by fixture")
			}
		}
	}
	if count != 1 || !legacy.Records[1].MechanicsAvailable() || !legacy.Records[2].MechanicsAvailable() {
		t.Fatal("canonical duplicate or compatibility changed")
	}
	if _, err := ComposeSpellSourceCatalogue(batch, batch); err == nil {
		t.Fatal("duplicate source accepted")
	}
}

func TestSpells218RejectMissingSourceDetails(t *testing.T) {
	original, err := os.ReadFile("testdata/spells_218.json")
	if err != nil {
		t.Fatal(err)
	}
	for index, s := range loadSpells218(t).Records {
		t.Run(s.Name, func(t *testing.T) {
			var root map[string]any
			if err := json.Unmarshal(original, &root); err != nil {
				t.Fatal(err)
			}
			record := root["records"].([]any)[index].(map[string]any)
			details := record["source_details"].(map[string]any)
			// Omit every individual field in the selected payload, including fields
			// whose intended source value is false, zero, or an empty list.
			paths := [][]string{{"source_details"}, {"required_mechanics"}, {"effects"}, {"components", "material"}}
			for key, value := range details {
				paths = append(paths, []string{"source_details", key})
				if payload, ok := value.(map[string]any); ok {
					for field := range payload {
						paths = append(paths, []string{"source_details", key, field})
					}
				}
			}
			for _, path := range paths {
				t.Run(strings.Join(path, "/"), func(t *testing.T) {
					var fresh map[string]any
					if err := json.Unmarshal(original, &fresh); err != nil {
						t.Fatal(err)
					}
					target := fresh["records"].([]any)[index].(map[string]any)
					for _, key := range path[:len(path)-1] {
						target = target[key].(map[string]any)
					}
					delete(target, path[len(path)-1])
					b, err := json.Marshal(fresh)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
						t.Fatal("missing source field accepted")
					}
				})
			}
		})
	}
}

func TestSpells218InvalidAndUnknownUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		index  int
		change func(*SpellRecord)
	}{
		{"unknown variant", 1, func(s *SpellRecord) { s.SourceDetails.Scrying.Variants[0] = "invented" }},
		{"missing knowledge row", 1, func(s *SpellRecord) { s.SourceDetails.Scrying.Knowledge = s.SourceDetails.Scrying.Knowledge[:2] }},
		{"invalid long cast", 1, func(s *SpellRecord) { s.SourceDetails.CastingSeconds = 0 }},
		{"invalid component cost", 3, func(s *SpellRecord) { *s.Components.Material.CostCP = -1 }},
		{"invalid decay", 3, func(s *SpellRecord) { s.SourceDetails.Chest.DailyCumulativePercent = 101 }},
		{"invalid delivery chance", 6, func(s *SpellRecord) { s.SourceDetails.Sending.CrossPlaneFailurePercent = 101 }},
		{"missing retained statistics", 8, func(s *SpellRecord) { s.SourceDetails.Shapechange.Retained = nil }},
		{"invalid sphere", 9, func(s *SpellRecord) { s.SourceDetails.Shatter.RadiusFeet = -1 }},
		{"invalid reaction", 10, func(s *SpellRecord) { s.SourceDetails.CastingTriggers = []string{"invented"} }},
		{"invalid effect reference", 9, func(s *SpellRecord) { s.Scaling.EffectIDs = []string{"missing"} }},
		{"missing dependency", 0, func(s *SpellRecord) { s.RequiredMechanics = nil }},
		{"missing effect dependency", 0, func(s *SpellRecord) { s.RequiredMechanics = []string{"unknown"} }},
		{"mismatched payload", 0, func(s *SpellRecord) { s.SourceDetails.Faith = &SpellFaith{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := loadSpells218(t)
			s := &r.Records[tc.index]
			tc.change(s)
			if s.MechanicsAvailable() || r.Validate() == nil {
				t.Fatal("invalid source accepted or available")
			}
		})
	}
	for _, s := range loadSpells218(t).Records {
		t.Run(s.Name, func(t *testing.T) {
			for _, key := range []string{"unknown.future-mechanic", "defense.shield-of-faith-self", "cast.cleric-self-slot-1"} {
				s.RequiredMechanics = append(s.RequiredMechanics, key)
				if s.MechanicsAvailable() {
					t.Fatal("unsupported source record became available")
				}
			}
		})
	}
	// Metadata must never be usable to smuggle widened behavior into either
	// exact bounded compatibility record, even with otherwise supported keys.
	for _, s := range loadSpellRecords(t).Records[1:] {
		s.SourceDetails = loadSpells218(t).Records[11].SourceDetails
		if s.MechanicsAvailable() {
			t.Fatal("bounded fixture accepted source extension")
		}
	}
}

func TestSpells218RejectInvalidScryingLabels(t *testing.T) {
	original, err := os.ReadFile("testdata/spells_218.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range []string{"knowledge", "connections"} {
		for _, choice := range []any{nil, "", "   "} {
			t.Run(fmt.Sprintf("%s/%v", group, choice), func(t *testing.T) {
				var root map[string]any
				if err := json.Unmarshal(original, &root); err != nil {
					t.Fatal(err)
				}
				scrying := root["records"].([]any)[1].(map[string]any)["source_details"].(map[string]any)["scrying"].(map[string]any)
				scrying[group].([]any)[0].(map[string]any)["choice"] = choice
				b, err := json.Marshal(root)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
					t.Fatal("invalid scrying label accepted")
				}
			})
		}
	}
}

func TestSpells218RejectInvalidRequiredLists(t *testing.T) {
	original, err := os.ReadFile("testdata/spells_218.json")
	if err != nil {
		t.Fatal(err)
	}
	for index, spell := range loadSpells218(t).Records {
		var root map[string]any
		if err := json.Unmarshal(original, &root); err != nil {
			t.Fatal(err)
		}
		details := root["records"].([]any)[index].(map[string]any)["source_details"].(map[string]any)
		for payloadName, payload := range details {
			fields, ok := payload.(map[string]any)
			if !ok {
				continue
			}
			for field, value := range fields {
				list, ok := value.([]any)
				if !ok {
					continue
				}
				invalid := []string{`[]`, `[ ]`, `[null]`}
				if _, stringList := list[0].(string); stringList {
					invalid = append(invalid, `[""]`, `["   "]`, `["valid",null]`)
				}
				for _, raw := range invalid {
					t.Run(spell.Name+"/"+payloadName+"/"+field+"/"+raw, func(t *testing.T) {
						fields[field] = "invalid-required-list-marker"
						b, err := json.Marshal(root)
						fields[field] = value
						if err != nil {
							t.Fatal(err)
						}
						b = bytes.Replace(b, []byte(`"invalid-required-list-marker"`), []byte(raw), 1)
						if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
							t.Fatal("invalid required list accepted")
						}
					})
				}
			}
		}
	}
}
