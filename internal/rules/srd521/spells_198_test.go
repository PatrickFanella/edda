package srd521

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"testing"
)

func loadSpells198(t *testing.T) SpellRecords {
	t.Helper()
	b, err := os.ReadFile("testdata/spells_198.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := DecodeSpellRecords(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Literal PDF pp.112-114 value fixtures include every choice, effect and scaling
// value, independently of decoder defaults and structural validation.
func TestSpells198SourceValues(t *testing.T) {
	expected := []string{
		`{"id":"srd-5.2.1:content/spells/barkskin","name":"Barkskin","source_id":"srd-5.2.1:content/spells/barkskin","source_page":112,"scope":"complete-source-record","level":2,"school":"transmutation","lists":["druid","ranger"],"casting_time":"bonus-action","range":{"kind":"touch","feet":0},"target":{"shape":"single","count":1,"eligibility":"willing-creature-touch"},"components":{"verbal":true,"somatic":true,"material":{"description":"a handful of bark","cost_cp":0,"consumed":false}},"ritual":false,"concentration":false,"duration":"timed","duration_seconds":3600,"scaling":{"kind":"none","above_level":0,"dice_count":0,"die_sides":0,"effect_ids":[]},"effects":[{"id":"ac","mechanic":"defense.barkskin-minimum-ac","timing":"for-duration","dice_count":0,"die_sides":0,"damage_type":"","bonus":17,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["defense.barkskin-minimum-ac","target.willing-touch","components.material"],"slot_levels":[2,3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-ended","barkskin":{"minimum_ac":17,"only_if_lower":true,"appearance":"bark-like-skin","willing":true}}}`,
		`{"id":"srd-5.2.1:content/spells/beacon-of-hope","name":"Beacon of Hope","source_id":"srd-5.2.1:content/spells/beacon-of-hope","source_page":112,"scope":"complete-source-record","level":3,"school":"abjuration","lists":["cleric"],"casting_time":"action","range":{"kind":"feet","feet":30},"target":{"shape":"selected-creatures","count":0,"eligibility":"creatures-in-range"},"components":{"verbal":true,"somatic":true,"material":null},"ritual":false,"concentration":true,"duration":"timed","duration_seconds":60,"scaling":{"kind":"none","above_level":0,"dice_count":0,"die_sides":0,"effect_ids":[]},"effects":[{"id":"hope","mechanic":"healing.beacon-maximum-and-save-advantage","timing":"for-duration","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["healing.beacon-maximum-and-save-advantage","target.any-number-in-range","concentration.source-duration"],"slot_levels":[3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","beacon":{"save_advantage":["wisdom","death"],"healing":"maximum-hit-points-from-any-healing","target_choice":"any-number-in-range"}}}`,
		`{"id":"srd-5.2.1:content/spells/befuddlement","name":"Befuddlement","source_id":"srd-5.2.1:content/spells/befuddlement","source_page":112,"scope":"complete-source-record","level":8,"school":"enchantment","lists":["bard","druid","warlock","wizard"],"casting_time":"action","range":{"kind":"feet","feet":150},"target":{"shape":"single","count":1,"eligibility":"visible-creature-in-range"},"components":{"verbal":true,"somatic":true,"material":{"description":"a key ring with no keys","cost_cp":0,"consumed":false}},"ritual":false,"concentration":false,"duration":"instantaneous","duration_seconds":0,"scaling":{"kind":"none","above_level":0,"dice_count":0,"die_sides":0,"effect_ids":[]},"effects":[{"id":"damage","mechanic":"damage.befuddlement","timing":"failed-save-half-on-success","dice_count":10,"die_sides":12,"damage_type":"psychic","bonus":0,"add_casting_modifier":false,"half_initial_only":false},{"id":"block","mechanic":"restriction.befuddlement-magic","timing":"failed-save-until-success-or-restoration","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["damage.befuddlement","restriction.befuddlement-magic","save.intelligence-half-only","restriction.spells-and-magic-action","save.repeat-30-days","restoration.greater-restoration-heal-wish","components.material"],"slot_levels":[8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"instantaneous","befuddlement":{"save":"intelligence","save_success":"half-damage-only","prohibited":["cast-spells","magic-action"],"repeat_save_days":30,"repeat_save_timing":"end-of-every-30-days","repeat_success":"ends-effect","ends_by":["greater-restoration","heal","wish"]}}}`,
		`{"id":"srd-5.2.1:content/spells/bestow-curse","name":"Bestow Curse","source_id":"srd-5.2.1:content/spells/bestow-curse","source_page":112,"scope":"complete-source-record","level":3,"school":"necromancy","lists":["bard","cleric","wizard"],"casting_time":"action","range":{"kind":"touch","feet":0},"target":{"shape":"single","count":1,"eligibility":"creature-touch"},"components":{"verbal":true,"somatic":true,"material":null},"ritual":false,"concentration":true,"duration":"timed","duration_seconds":60,"scaling":{"kind":"slot-duration","above_level":3,"dice_count":0,"die_sides":0,"effect_ids":["curse","extra-damage"]},"effects":[{"id":"curse","mechanic":"curse.chosen-effect","timing":"failed-save-for-slot-duration","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false},{"id":"extra-damage","mechanic":"damage.curse-chosen-necrotic","timing":"only-extra-necrotic-choice-on-caster-damage","dice_count":1,"die_sides":8,"damage_type":"necrotic","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["curse.chosen-effect","damage.curse-chosen-necrotic","save.wisdom","curse.four-choices","curse.ability-selection","curse.turn-start-wisdom-dodge","curse.caster-damage-necrotic","scaling.curse-slot-duration-concentration","concentration.source-duration"],"slot_levels":[3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","curse":{"save":"wisdom","save_success":"unaffected","choices":[{"choice":"ability-disadvantage","abilities":["strength","dexterity","constitution","intelligence","wisdom","charisma"],"rolls":["ability-checks","saving-throws"],"save":"not-applicable","timing":"for-duration","failure_action":"not-applicable","triggers":["for-duration"],"dice_count":0,"die_sides":0,"damage_type":"none","roll_modifier":"disadvantage"},{"choice":"attacks-against-caster","abilities":["not-applicable"],"rolls":["attack-rolls-against-caster"],"save":"not-applicable","timing":"for-duration","failure_action":"not-applicable","triggers":["for-duration"],"dice_count":0,"die_sides":0,"damage_type":"none","roll_modifier":"disadvantage"},{"choice":"forced-dodge","abilities":["not-applicable"],"rolls":["not-applicable"],"save":"wisdom","timing":"start-of-each-target-turn-in-combat","failure_action":"dodge","triggers":["start-of-turn-in-combat"],"dice_count":0,"die_sides":0,"damage_type":"none","roll_modifier":"none"},{"choice":"extra-necrotic","abilities":["not-applicable"],"rolls":["not-applicable"],"save":"not-applicable","timing":"when-caster-deals-damage","failure_action":"not-applicable","triggers":["caster-damages-target-with-attack-roll","caster-damages-target-with-spell"],"dice_count":1,"die_sides":8,"damage_type":"necrotic","roll_modifier":"none"}],"choose_count":1,"slot_durations":[{"min_slot":3,"max_slot":3,"seconds":60,"concentration":true,"until_dispelled":false},{"min_slot":4,"max_slot":4,"seconds":600,"concentration":true,"until_dispelled":false},{"min_slot":5,"max_slot":6,"seconds":28800,"concentration":false,"until_dispelled":false},{"min_slot":7,"max_slot":8,"seconds":86400,"concentration":false,"until_dispelled":false},{"min_slot":9,"max_slot":9,"seconds":0,"concentration":false,"until_dispelled":true}],"source_pages":[112,113]}}}`,
		`{"id":"srd-5.2.1:content/spells/black-tentacles","name":"Black Tentacles","source_id":"srd-5.2.1:content/spells/black-tentacles","source_page":113,"scope":"complete-source-record","level":4,"school":"conjuration","lists":["wizard"],"casting_time":"action","range":{"kind":"feet","feet":90},"target":{"shape":"square","count":0,"eligibility":"creatures-in-visible-ground-area"},"components":{"verbal":true,"somatic":true,"material":{"description":"a tentacle","cost_cp":0,"consumed":false}},"ritual":false,"concentration":true,"duration":"timed","duration_seconds":60,"scaling":{"kind":"none","above_level":0,"dice_count":0,"die_sides":0,"effect_ids":[]},"effects":[{"id":"area","mechanic":"area.black-tentacles","timing":"failed-save-on-appearance-entry-or-turn-end-once-per-turn","dice_count":3,"die_sides":6,"damage_type":"bludgeoning","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["area.black-tentacles","area.square-20-visible-ground","terrain.difficult","save.strength-once-per-turn","condition.restrained","check.action-strength-athletics-spell-dc","components.material","concentration.source-duration"],"slot_levels":[4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","tentacles":{"placement":"visible-ground-in-range","difficult_terrain":true,"save":"strength","save_triggers":["on-appearance-in-area","enters-area","ends-turn-in-area"],"saves_per_turn":1,"save_success":"no-damage-or-restraint","failure_condition":"restrained-until-spell-ends","escape_action":"action","escape_check":"strength-athletics","escape_dc":"caster-spell-save-dc","escape_success":"ends-restraint-on-self","appearance":"squirming-ebony-tentacles","side_feet":20}}}`,
		`{"id":"srd-5.2.1:content/spells/blade-barrier","name":"Blade Barrier","source_id":"srd-5.2.1:content/spells/blade-barrier","source_page":113,"scope":"complete-source-record","level":6,"school":"evocation","lists":["cleric"],"casting_time":"action","range":{"kind":"feet","feet":90},"target":{"shape":"wall","count":0,"eligibility":"creatures-in-wall"},"components":{"verbal":true,"somatic":true,"material":null},"ritual":false,"concentration":true,"duration":"timed","duration_seconds":600,"scaling":{"kind":"none","above_level":0,"dice_count":0,"die_sides":0,"effect_ids":[]},"effects":[{"id":"wall","mechanic":"area.blade-barrier","timing":"save-on-appearance-entry-or-turn-end-once-per-turn-half-on-success","dice_count":6,"die_sides":10,"damage_type":"force","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["area.blade-barrier","area.wall-straight-or-ring","terrain.difficult","cover.three-quarters","save.dexterity-half-once-per-turn","concentration.source-duration"],"slot_levels":[6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","barrier":{"forms":[{"choice":"straight","length_feet":100,"diameter_feet":0,"height_feet":20,"thickness_feet":5},{"choice":"ring","length_feet":0,"diameter_feet":60,"height_feet":20,"thickness_feet":5}],"dimensions":"up-to-length-or-diameter","cover":"three-quarters","difficult_terrain":true,"save":"dexterity","save_success":"half-damage","save_triggers":["on-appearance-in-wall","enters-wall","ends-turn-in-wall"],"saves_per_turn":1,"appearance":"whirling-blades-of-magical-energy"}}}`,
		`{"id":"srd-5.2.1:content/spells/bless","name":"Bless","source_id":"srd-5.2.1:content/spells/bless","source_page":113,"scope":"complete-source-record","level":1,"school":"enchantment","lists":["cleric","paladin"],"casting_time":"action","range":{"kind":"feet","feet":30},"target":{"shape":"up-to-count","count":3,"eligibility":"creatures-in-range"},"components":{"verbal":true,"somatic":true,"material":{"description":"a Holy Symbol worth 5+ GP","cost_cp":500,"consumed":false}},"ritual":false,"concentration":true,"duration":"timed","duration_seconds":60,"scaling":{"kind":"slot-targets","above_level":1,"dice_count":0,"die_sides":0,"effect_ids":["bonus"]},"effects":[{"id":"bonus","mechanic":"roll.bless","timing":"each-target-attack-roll-or-saving-throw","dice_count":1,"die_sides":4,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["roll.bless","scaling.targets-per-slot","roll.attack-and-save-d4","components.priced-holy-symbol","components.material","concentration.source-duration"],"slot_levels":[1,2,3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","bless":{"base_targets":3,"additional_targets_per_slot":1,"rolls":["attack-rolls","saving-throws"],"die_sides":4,"roll_modifier":"add","target_choice":"up-to-count-creatures-in-range"}}}`,
		`{"id":"srd-5.2.1:content/spells/blight","name":"Blight","source_id":"srd-5.2.1:content/spells/blight","source_page":113,"scope":"complete-source-record","level":4,"school":"necromancy","lists":["druid","sorcerer","warlock","wizard"],"casting_time":"action","range":{"kind":"feet","feet":30},"target":{"shape":"choice","count":1,"eligibility":"visible-creature-or-nonmagical-noncreature-plant"},"components":{"verbal":true,"somatic":true,"material":null},"ritual":false,"concentration":false,"duration":"instantaneous","duration_seconds":0,"scaling":{"kind":"slot","above_level":4,"dice_count":1,"die_sides":8,"effect_ids":["damage"]},"effects":[{"id":"damage","mechanic":"damage.blight","timing":"creature-failed-save-half-on-success","dice_count":8,"die_sides":8,"damage_type":"necrotic","bonus":0,"add_casting_modifier":false,"half_initial_only":false},{"id":"plant","mechanic":"plant.blight-withering","timing":"nonmagical-noncreature-plant-no-save","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["damage.blight","plant.blight-withering","save.constitution-half","save.plant-automatic-failure","target.nonmagical-noncreature-plant","scaling.per-slot"],"slot_levels":[4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"instantaneous","blight":{"variants":["visible-creature","nonmagical-noncreature-plant"],"save":"constitution","save_success":"half-damage","automatic_failure_types":["plant"],"plant_variant_save":false,"plant_variant_effect":"withers-and-dies","plant_examples":["tree","shrub"]}}}`,
		`{"id":"srd-5.2.1:content/spells/blindness-deafness","name":"Blindness/Deafness","source_id":"srd-5.2.1:content/spells/blindness-deafness","source_page":113,"scope":"complete-source-record","level":2,"school":"transmutation","lists":["bard","cleric","sorcerer","wizard"],"casting_time":"action","range":{"kind":"feet","feet":120},"target":{"shape":"up-to-count","count":1,"eligibility":"visible-creatures-in-range"},"components":{"verbal":true,"somatic":false,"material":null},"ritual":false,"concentration":false,"duration":"timed","duration_seconds":60,"scaling":{"kind":"slot-targets","above_level":2,"dice_count":0,"die_sides":0,"effect_ids":["condition"]},"effects":[{"id":"condition","mechanic":"condition.blindness-or-deafness","timing":"failed-save-until-duration-or-end-turn-success","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["condition.blindness-or-deafness","save.constitution-repeat-end-turn","condition.choice-blinded-deafened","scaling.targets-per-slot"],"slot_levels":[2,3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-ended","blindness":{"conditions":["blinded","deafened"],"choice_per_target":true,"save":"constitution","repeat_save_timing":"end-of-each-target-turn","success":"ends-spell-on-self","base_targets":1,"additional_targets_per_slot":1}}}`,
		`{"id":"srd-5.2.1:content/spells/blink","name":"Blink","source_id":"srd-5.2.1:content/spells/blink","source_page":113,"scope":"complete-source-record","level":3,"school":"transmutation","lists":["sorcerer","wizard"],"casting_time":"action","range":{"kind":"self","feet":0},"target":{"shape":"single","count":1,"eligibility":"self"},"components":{"verbal":true,"somatic":true,"material":null},"ritual":false,"concentration":false,"duration":"timed","duration_seconds":60,"scaling":{"kind":"none","above_level":0,"dice_count":0,"die_sides":0,"effect_ids":[]},"effects":[{"id":"transport","mechanic":"transport.blink","timing":"end-turn-d6-4-to-6-and-return","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["transport.blink","plane.ethereal-perception-interaction","roll.end-turn-d6","transport.blink-return-space"],"slot_levels":[3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-ended","blink":{"roll_die_sides":6,"vanish_min_roll":4,"vanish_max_roll":6,"roll_timing":"end-of-each-caster-turn","destination_plane":"ethereal","already_ethereal":"spell-ends-instantly","perceive_departure_plane":true,"perception_distance_feet":60,"perception_appearance":"shades-of-gray","interaction":"only-other-creatures-on-ethereal-plane","other_plane_perception":"requires-special-ethereal-perception-ability","return_triggers":["start-of-next-caster-turn","spell-ends-while-ethereal"],"return_distance_feet":10,"return_space":"chosen-visible-unoccupied-within-distance-of-departure","no_space_return":"nearest-unoccupied-space","source_pages":[113,114]}}}`,
		`{"id":"srd-5.2.1:content/spells/blur","name":"Blur","source_id":"srd-5.2.1:content/spells/blur","source_page":114,"scope":"complete-source-record","level":2,"school":"illusion","lists":["sorcerer","wizard"],"casting_time":"action","range":{"kind":"self","feet":0},"target":{"shape":"single","count":1,"eligibility":"self"},"components":{"verbal":true,"somatic":false,"material":null},"ritual":false,"concentration":true,"duration":"timed","duration_seconds":60,"scaling":{"kind":"none","above_level":0,"dice_count":0,"die_sides":0,"effect_ids":[]},"effects":[{"id":"blur","mechanic":"defense.blur","timing":"for-duration-unless-blindsight-or-truesight","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["defense.blur","attack.disadvantage","senses.blindsight-truesight-immunity","concentration.source-duration"],"slot_levels":[2,3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"duration-or-concentration-ended","blur":{"attack_rolls_against_caster":"disadvantage","immune_attacker_perception":["blindsight","truesight"],"appearance":"blurred-body"}}}`,
		`{"id":"srd-5.2.1:content/spells/burning-hands","name":"Burning Hands","source_id":"srd-5.2.1:content/spells/burning-hands","source_page":114,"scope":"complete-source-record","level":1,"school":"evocation","lists":["sorcerer","wizard"],"casting_time":"action","range":{"kind":"self","feet":0},"target":{"shape":"cone","count":0,"eligibility":"creatures-and-flammable-unattended-objects"},"components":{"verbal":true,"somatic":true,"material":null},"ritual":false,"concentration":false,"duration":"instantaneous","duration_seconds":0,"scaling":{"kind":"slot","above_level":1,"dice_count":1,"die_sides":6,"effect_ids":["damage"]},"effects":[{"id":"damage","mechanic":"damage.burning-hands","timing":"cone-save-half-on-success","dice_count":3,"die_sides":6,"damage_type":"fire","bonus":0,"add_casting_modifier":false,"half_initial_only":false},{"id":"ignite","mechanic":"object.ignite-unattended-flammable","timing":"on-cast-in-cone","dice_count":0,"die_sides":0,"damage_type":"","bonus":0,"add_casting_modifier":false,"half_initial_only":false}],"required_mechanics":["damage.burning-hands","object.ignite-unattended-flammable","area.cone-15","save.dexterity-half","scaling.per-slot"],"slot_levels":[1,2,3,4,5,6,7,8,9],"source_details":{"casting_seconds":0,"casting_triggers":[],"end_timing":"instantaneous","burning":{"cone_feet":15,"save":"dexterity","save_success":"half-damage","ignited_objects":["flammable","in-cone","not-worn","not-carried"],"appearance":"thin-sheet-of-flames-from-caster"}}}`,
	}
	batch := loadSpells198(t)
	if len(batch.Records) != 12 {
		t.Fatal("exact twelve families required")
	}
	names := []string{"Barkskin", "Beacon of Hope", "Befuddlement", "Bestow Curse", "Black Tentacles", "Blade Barrier", "Bless", "Blight", "Blindness/Deafness", "Blink", "Blur", "Burning Hands"}
	for i, r := range batch.Records {
		t.Run(names[i], func(t *testing.T) {
			var want SpellRecord
			if err := json.Unmarshal([]byte(expected[i]), &want); err != nil {
				t.Fatal(err)
			}
			if r.Name != names[i] || !reflect.DeepEqual(r, want) {
				t.Fatalf("PDF p.%d source values differ", want.SourcePage)
			}
			if r.MechanicsAvailable() {
				t.Fatal("source effect available")
			}
			slots := []int{}
			for level := r.Level; level <= 9; level++ {
				slots = append(slots, level)
			}
			if !slices.Equal(r.SlotLevels, slots) {
				t.Fatal("source slot metadata incomplete")
			}
		})
	}
}

func TestSpells198CompositionBoundaries(t *testing.T) {
	legacy := loadSpellRecords(t)
	batch := loadSpells198(t)
	all, err := ComposeSpellSourceCatalogue(legacy, loadSpells218(t), batch)
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Records) != 25 {
		t.Fatal("expected Acid Arrow and two twelve-family batches")
	}
	for _, r := range all.Records {
		if r.Scope != "complete-source-record" || r.MechanicsAvailable() {
			t.Fatal("source granted availability")
		}
	}
	if _, err := ComposeSpellSourceCatalogue(batch, batch); err == nil {
		t.Fatal("duplicate canonical IDs accepted")
	}
	for _, r := range legacy.Records[1:] {
		if !r.MechanicsAvailable() {
			t.Fatal("bounded availability changed")
		}
		r.SourceDetails = batch.Records[0].SourceDetails
		if r.MechanicsAvailable() {
			t.Fatal("bounded fixture accepted source extension")
		}
	}
	for _, r := range batch.Records {
		r.RequiredMechanics = append(r.RequiredMechanics, "cast.cleric-self-slot-1", "unknown.future")
		if r.MechanicsAvailable() {
			t.Fatal("unsupported source available")
		}
	}
}

// Mutate each field, recursively, rather than only the top-level payload.
// The valid input covers explicit false and zero (including slot 9 seconds).
func TestSpells198MalformedFields(t *testing.T) {
	original, err := os.ReadFile("testdata/spells_198.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(original, &root); err != nil {
		t.Fatal(err)
	}
	for index, record := range root["records"].([]any) {
		r := record.(map[string]any)
		var walk func(map[string]any, string)
		reject := func(label string) {
			t.Helper()
			b, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
				t.Fatalf("record %d accepted %s", index, label)
			}
		}
		walk = func(fields map[string]any, path string) {
			for key, value := range fields {
				fieldPath := path + "/" + key
				delete(fields, key)
				reject("omitted " + fieldPath)
				fields[key] = nil
				reject("null " + fieldPath)
				fields[key] = value
				switch v := value.(type) {
				case float64:
					fields[key] = -1
					reject("negative " + fieldPath)
					fields[key] = value
				case string:
					for _, bad := range []string{"", "   "} {
						fields[key] = bad
						reject("blank " + fieldPath)
					}
					fields[key] = value
				case map[string]any:
					walk(v, fieldPath)
				case []any:
					if len(v) == 0 {
						continue
					}
					for _, bad := range []any{[]any{}, []any{nil}, []any{""}, []any{"   "}, []any{v[0], nil}, []any{v[0], map[string]any{"mixed": true}}} {
						fields[key] = bad
						reject("malformed list " + fieldPath)
					}
					fields[key] = value
					for j, item := range v {
						if nested, ok := item.(map[string]any); ok {
							walk(nested, fmt.Sprintf("%s/%d", fieldPath, j))
						}
					}
				}
			}
		}
		details := r["source_details"].(map[string]any)
		for key, payload := range details {
			if obj, ok := payload.(map[string]any); ok {
				walk(obj, key)
			}
		}
		for _, key := range []string{"source_details", "effects", "required_mechanics", "slot_levels", "lists"} {
			value := r[key]
			delete(r, key)
			reject("omitted " + key)
			r[key] = value
		}
	}
}

func TestSpells198JSONGuards(t *testing.T) {
	original, err := os.ReadFile("testdata/spells_198.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{
		{`"minimum_ac": 17`, `"minimum_ac": 17, "minimum_ac": 18`},
		{`"minimum_ac": 17`, `"minimum_ac": 17, "MINIMUM_AC": 18`},
		{`"minimum_ac": 17`, `"minimum_ac": 17, "unknown": true`},
		{`"save_advantage": [
            "wisdom",
            "death"
          ]`, `"save_advantage": [ ]`},
		{`"source_id": "srd-5.2.1:content/spells/barkskin"`, `"source_id": "srd-5.2.1:content/spells/unknown"`},
		{`"effect_ids": [
          "damage"
        ]`, `"effect_ids": ["unknown"]`},
	} {
		t.Run(pair[1], func(t *testing.T) {
			b := bytes.Replace(original, []byte(pair[0]), []byte(pair[1]), 1)
			if bytes.Equal(b, original) {
				t.Fatal("fixture mutation did not match")
			}
			if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
				t.Fatal("invalid JSON/source graph accepted")
			}
		})
	}
}

func TestSpells198InvalidTypedVariants(t *testing.T) {
	for _, tc := range []struct {
		name   string
		index  int
		change func(*SpellRecord)
	}{
		{"curse unknown choice", 3, func(s *SpellRecord) { s.SourceDetails.Curse.Choices[0].Choice = "unknown" }},
		{"curse slot concentration", 3, func(s *SpellRecord) { s.SourceDetails.Curse.SlotDurations[2].Concentration = true }},
		{"curse indefinite seconds", 3, func(s *SpellRecord) { s.SourceDetails.Curse.SlotDurations[4].Seconds = 60 }},
		{"wall variant", 5, func(s *SpellRecord) { s.SourceDetails.Barrier.Forms[0].Choice = "unknown" }},
		{"tentacles frequency", 4, func(s *SpellRecord) { s.SourceDetails.Tentacles.SavesPerTurn = 2 }},
		{"blight variant", 7, func(s *SpellRecord) { s.SourceDetails.Blight.Variants[0] = "unknown" }},
		{"condition variant", 8, func(s *SpellRecord) { s.SourceDetails.Blindness.Conditions[0] = "unknown" }},
		{"multiple payloads", 0, func(s *SpellRecord) { s.SourceDetails.Blink = &SpellBlink{} }},
		{"source link", 0, func(s *SpellRecord) { s.SourceID = "unknown" }},
		{"effect graph", 0, func(s *SpellRecord) { s.Effects[0].Mechanic = "unknown" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := loadSpells198(t)
			tc.change(&r.Records[tc.index])
			if r.Validate() == nil || r.Records[tc.index].MechanicsAvailable() {
				t.Fatal("invalid source accepted")
			}
		})
	}
}

func TestSpells198RequiredNumbersAndFlags(t *testing.T) {
	original, err := os.ReadFile("testdata/spells_198.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(original, &root); err != nil {
		t.Fatal(err)
	}
	for index, item := range root["records"].([]any) {
		r := item.(map[string]any)
		objects := []map[string]any{r, r["range"].(map[string]any), r["target"].(map[string]any), r["scaling"].(map[string]any), r["components"].(map[string]any)}
		for _, effect := range r["effects"].([]any) {
			objects = append(objects, effect.(map[string]any))
		}
		if material, ok := r["components"].(map[string]any)["material"].(map[string]any); ok {
			objects = append(objects, material)
		}
		for _, obj := range objects {
			for key, value := range obj {
				switch value.(type) {
				case float64, bool:
				default:
					continue
				}
				for _, omit := range []bool{true, false} {
					if omit {
						delete(obj, key)
					} else {
						obj[key] = nil
					}
					b, err := json.Marshal(root)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
						t.Fatalf("record %d accepted missing/null %s", index, key)
					}
					obj[key] = value
				}
			}
		}
	}
}

func TestSpells198WhitespaceEmptyListsAndNestedGuards(t *testing.T) {
	original, err := os.ReadFile("testdata/spells_198.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(original, &root); err != nil {
		t.Fatal(err)
	}
	var walk func(any)
	walk = func(node any) {
		switch node := node.(type) {
		case map[string]any:
			for key, value := range node {
				if list, ok := value.([]any); ok && len(list) > 0 {
					node[key] = "empty-list-marker"
					b, err := json.Marshal(root)
					node[key] = value
					if err != nil {
						t.Fatal(err)
					}
					b = bytes.Replace(b, []byte(`"empty-list-marker"`), []byte(`[ ]`), 1)
					// Empty scaling effect references are valid only for kind=none, which
					// already has an empty list and therefore never reaches this branch.
					if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
						t.Fatalf("accepted whitespace empty %s", key)
					}
				}
				walk(value)
			}
		case []any:
			for _, item := range node {
				walk(item)
			}
		}
	}
	walk(root)
	for _, replacement := range []string{
		`"choice": "ability-disadvantage", "choice": "unknown"`,
		`"choice": "ability-disadvantage", "CHOICE": "unknown"`,
		`"choice": "ability-disadvantage", "unknown": true`,
	} {
		b := bytes.Replace(original, []byte(`"choice": "ability-disadvantage"`), []byte(replacement), 1)
		if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
			t.Fatal("nested duplicate/alias/unknown accepted")
		}
	}
}

func TestSpells198RejectContradictoryTargetScaling(t *testing.T) {
	for _, index := range []int{6, 8} {
		for _, invalid := range []string{"zero increment", "mismatched base"} {
			t.Run(fmt.Sprintf("%d/%s", index, invalid), func(t *testing.T) {
				r := loadSpells198(t)
				s := &r.Records[index]
				if d := s.SourceDetails.Bless; d != nil {
					if invalid == "zero increment" {
						d.AdditionalTargetsPerSlot = 0
					} else {
						d.BaseTargets = 99
					}
				} else {
					d := s.SourceDetails.Blindness
					if invalid == "zero increment" {
						d.AdditionalTargetsPerSlot = 0
					} else {
						d.BaseTargets = 99
					}
				}
				if err := r.Validate(); err == nil {
					t.Error("contradictory target scaling validated")
				}
				b, err := json.Marshal(r)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := DecodeSpellRecords(bytes.NewReader(b)); err == nil {
					t.Error("contradictory target scaling decoded")
				}
			})
		}
	}
}
