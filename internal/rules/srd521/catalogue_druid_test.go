package srd521

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func druidCatalogue(t *testing.T) CharacterCatalogue {
	t.Helper()
	b, err := os.ReadFile("testdata/character-druid.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// Explicit source expectations for every serialized field, including empty
// arrays, source ownership, predicates and their operands. These are frozen
// assertions, never loaded from the imported JSON or a record digest. Table
// counts include all nine slot columns (absent capacities have no level entry).
var druidSourceAssertions = []string{
	`{"id":"srd-5.2.1:character/class/druid","kind":"class","name":"Druid","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/skills","srd-5.2.1:character/choice/druid/starting-equipment"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/primary-ability/wisdom","srd-5.2.1:mechanic/druid/hit-die-sides","srd-5.2.1:mechanic/saving-throws/intelligence","srd-5.2.1:mechanic/saving-throws/wisdom","srd-5.2.1:mechanic/weapons/simple","srd-5.2.1:mechanic/tools/herbalism-kit","srd-5.2.1:mechanic/armor/light","srd-5.2.1:mechanic/armor/shields","srd-5.2.1:mechanic/druid/multiclass/hit-die-light-armor-shields-only","srd-5.2.1:mechanic/druid/multiclass/level-1-features-and-multiclass-spell-slots","srd-5.2.1:mechanic/druid/level-1/all-core-traits-and-features"],"select":0,"option_keys":[],"capacity":0,"parameters":{"srd-5.2.1:mechanic/druid/hit-die-sides":8}}`,
	`{"id":"srd-5.2.1:character/choice/druid/skills","kind":"choice","name":"Skills","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/skills","srd-5.2.1:mechanic/druid/skills/source-eight-options"],"select":2,"option_keys":["srd-5.2.1:mechanic/skills/animal-handling","srd-5.2.1:mechanic/skills/arcana","srd-5.2.1:mechanic/skills/insight","srd-5.2.1:mechanic/skills/medicine","srd-5.2.1:mechanic/skills/nature","srd-5.2.1:mechanic/skills/perception","srd-5.2.1:mechanic/skills/religion","srd-5.2.1:mechanic/skills/survival"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/starting-equipment","kind":"choice","name":"Starting Equipment","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/starting-equipment","srd-5.2.1:mechanic/equipment/leather-armor","srd-5.2.1:mechanic/equipment/shield","srd-5.2.1:mechanic/equipment/sickle","srd-5.2.1:mechanic/equipment/druidic-focus-quarterstaff","srd-5.2.1:mechanic/equipment/explorers-pack","srd-5.2.1:mechanic/equipment/herbalism-kit","srd-5.2.1:mechanic/currency/gp"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/equipment/a","srd-5.2.1:mechanic/druid/equipment/b"],"capacity":0,"options":[{"key":"srd-5.2.1:mechanic/druid/equipment/a","quantities":{"srd-5.2.1:mechanic/equipment/leather-armor":1,"srd-5.2.1:mechanic/equipment/shield":1,"srd-5.2.1:mechanic/equipment/sickle":1,"srd-5.2.1:mechanic/equipment/druidic-focus-quarterstaff":1,"srd-5.2.1:mechanic/equipment/explorers-pack":1,"srd-5.2.1:mechanic/equipment/herbalism-kit":1,"srd-5.2.1:mechanic/currency/gp":9}},{"key":"srd-5.2.1:mechanic/druid/equipment/b","quantities":{"srd-5.2.1:mechanic/currency/gp":50}}]}`,
	`{"id":"srd-5.2.1:character/level/druid/1","kind":"level","name":"Druid level 1","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":1,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":2,"feature_refs":["srd-5.2.1:character/feature/druid/spellcasting","srd-5.2.1:character/feature/druid/druidic","srd-5.2.1:character/feature/druid/primal-order"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/spellcasting","kind":"feature","name":"Spellcasting","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/level/druid/1","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/cantrips","srd-5.2.1:character/choice/druid/prepared-spells"],"resource_refs":["srd-5.2.1:character/resource/druid/spell-slot-1","srd-5.2.1:character/resource/druid/spell-slot-2","srd-5.2.1:character/resource/druid/spell-slot-3","srd-5.2.1:character/resource/druid/spell-slot-4","srd-5.2.1:character/resource/druid/spell-slot-5","srd-5.2.1:character/resource/druid/spell-slot-6","srd-5.2.1:character/resource/druid/spell-slot-7","srd-5.2.1:character/resource/druid/spell-slot-8","srd-5.2.1:character/resource/druid/spell-slot-9"],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/wisdom","srd-5.2.1:mechanic/druid/spellcasting/druidic-focus","srd-5.2.1:mechanic/druid/spellcasting/replace-one-cantrip-on-druid-level","srd-5.2.1:mechanic/druid/spellcasting/always-prepared-excluded-from-count-still-druid-spells","srd-5.2.1:mechanic/druid/spellcasting/long-rest-replace-any-prepared-spells","srd-5.2.1:mechanic/druid/recommended-cantrips/druidcraft","srd-5.2.1:mechanic/druid/recommended-cantrips/produce-flame","srd-5.2.1:mechanic/druid/recommended-spells/animal-friendship","srd-5.2.1:mechanic/druid/recommended-spells/cure-wounds","srd-5.2.1:mechanic/druid/recommended-spells/faerie-fire","srd-5.2.1:mechanic/druid/recommended-spells/thunderwave"],"select":0,"option_keys":[],"capacity":0}`,
	`{"id":"srd-5.2.1:character/feature/druid/druidic","kind":"feature","name":"Druidic","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-druidic","page":42}],"parent_id":"srd-5.2.1:character/level/druid/1","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/languages/druidic","srd-5.2.1:mechanic/druid/druidic/always-prepared-speak-with-animals","srd-5.2.1:mechanic/druid/druidic/knowers-automatically-spot-hidden-messages","srd-5.2.1:mechanic/druid/druidic/nonknower-investigation-intelligence-dc-15-spot-not-decipher-without-magic","srd-5.2.1:mechanic/druid/hidden-message-dc"],"select":0,"option_keys":[],"capacity":0,"always_prepared":[{"level":1,"spell_keys":["srd-5.2.1:mechanic/druid/druidic/always-prepared-speak-with-animals"]}],"parameters":{"srd-5.2.1:mechanic/druid/hidden-message-dc":15}}`,
	`{"id":"srd-5.2.1:character/feature/druid/primal-order","kind":"feature","name":"Primal Order","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-primal-order","page":42}],"parent_id":"srd-5.2.1:character/level/druid/1","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/primal-order"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/primal-order/choose-one","srd-5.2.1:mechanic/druid/magician-extra-cantrips","srd-5.2.1:mechanic/druid/magician-minimum-bonus"],"select":0,"option_keys":[],"capacity":0,"parameters":{"srd-5.2.1:mechanic/druid/magician-extra-cantrips":1,"srd-5.2.1:mechanic/druid/magician-minimum-bonus":1}}`,
	`{"id":"srd-5.2.1:character/level/druid/2","kind":"level","name":"Druid level 2","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":2,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":2,"feature_refs":["srd-5.2.1:character/feature/druid/wild-shape","srd-5.2.1:character/feature/druid/wild-companion"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/wild-shape","kind":"feature","name":"Wild Shape","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-2-wild-shape","page":42}],"parent_id":"srd-5.2.1:character/level/druid/2","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/known-forms","srd-5.2.1:character/choice/druid/shape-equipment"],"resource_refs":["srd-5.2.1:character/resource/druid/wild-shape"],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-shape/bonus-action-learned-beast-form","srd-5.2.1:mechanic/druid/wild-shape/duration-hours-half-druid-level","srd-5.2.1:mechanic/druid/duration-level-divisor","srd-5.2.1:mechanic/druid/wild-shape/ends-on-reuse-incapacitated-death","srd-5.2.1:mechanic/druid/wild-shape/bonus-action-end-early","srd-5.2.1:mechanic/druid/short-rest-recovered-uses","srd-5.2.1:mechanic/druid/forms-at-level-2","srd-5.2.1:mechanic/druid/cr-numerator-at-level-2","srd-5.2.1:mechanic/druid/cr-denominator-at-level-2","srd-5.2.1:mechanic/druid/forms-at-level-4","srd-5.2.1:mechanic/druid/cr-numerator-at-level-4","srd-5.2.1:mechanic/druid/cr-denominator-at-level-4","srd-5.2.1:mechanic/druid/forms-at-level-8","srd-5.2.1:mechanic/druid/cr-numerator-at-level-8","srd-5.2.1:mechanic/druid/cr-denominator-at-level-8","srd-5.2.1:mechanic/druid/wild-shape/retain-personality-memories-speech","srd-5.2.1:mechanic/druid/wild-shape/temporary-hp-equal-druid-level","srd-5.2.1:mechanic/druid/wild-shape/replace-statistics-with-beast","srd-5.2.1:mechanic/druid/wild-shape/retain-type-hp-hit-dice-intelligence-wisdom-charisma-class-features-languages-feats","srd-5.2.1:mechanic/druid/wild-shape/retain-skills-saves-use-own-proficiency-bonus-gain-beast-proficiencies","srd-5.2.1:mechanic/druid/wild-shape/use-higher-beast-skill-save-modifier","srd-5.2.1:mechanic/druid/wild-shape/no-spellcasting-retain-concentration-and-prior-spells","srd-5.2.1:mechanic/druid/wild-shape/object-handling-form-limbs"],"select":0,"option_keys":[],"capacity":0,"parameters":{"srd-5.2.1:mechanic/druid/duration-level-divisor":2,"srd-5.2.1:mechanic/druid/short-rest-recovered-uses":1,"srd-5.2.1:mechanic/druid/forms-at-level-2":4,"srd-5.2.1:mechanic/druid/cr-numerator-at-level-2":1,"srd-5.2.1:mechanic/druid/cr-denominator-at-level-2":4,"srd-5.2.1:mechanic/druid/forms-at-level-4":6,"srd-5.2.1:mechanic/druid/cr-numerator-at-level-4":1,"srd-5.2.1:mechanic/druid/cr-denominator-at-level-4":2,"srd-5.2.1:mechanic/druid/forms-at-level-8":8,"srd-5.2.1:mechanic/druid/cr-numerator-at-level-8":1,"srd-5.2.1:mechanic/druid/cr-denominator-at-level-8":1}}`,
	`{"id":"srd-5.2.1:character/feature/druid/wild-companion","kind":"feature","name":"Wild Companion","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-2-wild-companion","page":43}],"parent_id":"srd-5.2.1:character/level/druid/2","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/wild-companion-cost"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-companion/magic-action-find-familiar-no-material-components","srd-5.2.1:mechanic/druid/wild-companion/fey-disappears-on-long-rest"],"select":0,"option_keys":[],"capacity":0}`,
	`{"id":"srd-5.2.1:character/level/druid/3","kind":"level","name":"Druid level 3","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":3,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":2,"feature_refs":["srd-5.2.1:character/feature/druid/druid-subclass"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/druid-subclass","kind":"feature","name":"Druid Subclass","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-3-druid-subclass","page":43}],"parent_id":"srd-5.2.1:character/level/druid/3","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/druid-subclass"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/subclass-features-at-or-below-druid-level"],"select":0,"option_keys":[],"capacity":0}`,
	`{"id":"srd-5.2.1:character/level/druid/4","kind":"level","name":"Druid level 4","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":4,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":2,"feature_refs":["srd-5.2.1:character/feature/druid/ability-score-improvement"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/ability-score-improvement","kind":"feature","name":"Ability Score Improvement","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-4-ability-score-improvement","page":43}],"parent_id":"srd-5.2.1:character/level/druid/4","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/ability-score-improvement"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/ability-score-improvement","srd-5.2.1:mechanic/druid/asi-repeat-at-8-12-16"],"select":0,"option_keys":[],"capacity":0}`,
	`{"id":"srd-5.2.1:character/level/druid/5","kind":"level","name":"Druid level 5","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":5,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":3,"feature_refs":["srd-5.2.1:character/feature/druid/wild-resurgence"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/wild-resurgence","kind":"feature","name":"Wild Resurgence","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-5-wild-resurgence","page":43}],"parent_id":"srd-5.2.1:character/level/druid/5","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/wild-resurgence-conversion"],"resource_refs":["srd-5.2.1:character/resource/druid/wild-resurgence-slot-conversion"],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/wild-resurgence-empty"],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-resurgence/once-own-turn-zero-uses-expend-slot-regain-one-no-action","srd-5.2.1:mechanic/druid/regained-uses"],"select":0,"option_keys":[],"capacity":0,"parameters":{"srd-5.2.1:mechanic/druid/regained-uses":1}}`,
	`{"id":"srd-5.2.1:character/level/druid/6","kind":"level","name":"Druid level 6","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":6,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression","srd-5.2.1:mechanic/druid/subclass-feature"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":3,"feature_refs":[],"feature_keys":["srd-5.2.1:mechanic/druid/subclass-feature"]}}`,
	`{"id":"srd-5.2.1:character/level/druid/7","kind":"level","name":"Druid level 7","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":7,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":3,"feature_refs":["srd-5.2.1:character/feature/druid/elemental-fury"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/elemental-fury","kind":"feature","name":"Elemental Fury","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-7-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/level/druid/7","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/elemental-fury"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/elemental-fury/choose-one","srd-5.2.1:mechanic/druid/primal-strike-dice","srd-5.2.1:mechanic/druid/primal-strike-die-sides"],"select":0,"option_keys":[],"capacity":0,"parameters":{"srd-5.2.1:mechanic/druid/primal-strike-dice":1,"srd-5.2.1:mechanic/druid/primal-strike-die-sides":8}}`,
	`{"id":"srd-5.2.1:character/level/druid/8","kind":"level","name":"Druid level 8","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":8,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":3,"feature_refs":["srd-5.2.1:character/feature/druid/ability-score-improvement"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/level/druid/9","kind":"level","name":"Druid level 9","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":9,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":4,"feature_refs":[],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/level/druid/10","kind":"level","name":"Druid level 10","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":10,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression","srd-5.2.1:mechanic/druid/subclass-feature"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":4,"feature_refs":[],"feature_keys":["srd-5.2.1:mechanic/druid/subclass-feature"]}}`,
	`{"id":"srd-5.2.1:character/level/druid/11","kind":"level","name":"Druid level 11","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":11,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":4,"feature_refs":[],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/level/druid/12","kind":"level","name":"Druid level 12","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":12,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":4,"feature_refs":["srd-5.2.1:character/feature/druid/ability-score-improvement"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/level/druid/13","kind":"level","name":"Druid level 13","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":13,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":5,"feature_refs":[],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/level/druid/14","kind":"level","name":"Druid level 14","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":14,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression","srd-5.2.1:mechanic/druid/subclass-feature"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":5,"feature_refs":[],"feature_keys":["srd-5.2.1:mechanic/druid/subclass-feature"]}}`,
	`{"id":"srd-5.2.1:character/level/druid/15","kind":"level","name":"Druid level 15","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":15,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":5,"feature_refs":["srd-5.2.1:character/feature/druid/improved-elemental-fury"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/improved-elemental-fury","kind":"feature","name":"Improved Elemental Fury","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-15-improved-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/level/druid/15","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/improved-fury-choice"],"required_mechanics":["srd-5.2.1:mechanic/druid/improved-elemental-fury/improves-only-chosen-option","srd-5.2.1:mechanic/druid/improved-elemental-fury/potent-druid-cantrip-minimum-range-10-add-300-feet","srd-5.2.1:mechanic/druid/potent-minimum-range-feet","srd-5.2.1:mechanic/druid/potent-added-range-feet","srd-5.2.1:mechanic/druid/improved-elemental-fury/primal-strike-extra-2d8","srd-5.2.1:mechanic/druid/primal-strike-dice","srd-5.2.1:mechanic/druid/primal-strike-die-sides"],"select":0,"option_keys":[],"capacity":0,"parameters":{"srd-5.2.1:mechanic/druid/potent-minimum-range-feet":10,"srd-5.2.1:mechanic/druid/potent-added-range-feet":300,"srd-5.2.1:mechanic/druid/primal-strike-dice":2,"srd-5.2.1:mechanic/druid/primal-strike-die-sides":8}}`,
	`{"id":"srd-5.2.1:character/level/druid/16","kind":"level","name":"Druid level 16","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":16,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":5,"feature_refs":["srd-5.2.1:character/feature/druid/ability-score-improvement"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/level/druid/17","kind":"level","name":"Druid level 17","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":17,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":6,"feature_refs":[],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/level/druid/18","kind":"level","name":"Druid level 18","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":18,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":6,"feature_refs":["srd-5.2.1:character/feature/druid/beast-spells"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/beast-spells","kind":"feature","name":"Beast Spells","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-18-beast-spells","page":43}],"parent_id":"srd-5.2.1:character/level/druid/18","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/beast-spells-components"],"required_mechanics":["srd-5.2.1:mechanic/druid/beast-spells/cast-in-wild-shape"],"select":0,"option_keys":[],"capacity":0}`,
	`{"id":"srd-5.2.1:character/level/druid/19","kind":"level","name":"Druid level 19","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":19,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":6,"feature_refs":["srd-5.2.1:character/feature/druid/epic-boon"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/epic-boon","kind":"feature","name":"Epic Boon","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-19-epic-boon","page":43}],"parent_id":"srd-5.2.1:character/level/druid/19","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/epic-boon"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/epic-boon","srd-5.2.1:mechanic/druid/epic-boon/recommended-boon-of-dimensional-travel"],"select":0,"option_keys":[],"capacity":0}`,
	`{"id":"srd-5.2.1:character/level/druid/20","kind":"level","name":"Druid level 20","sources":[{"inventory_id":"srd-5.2.1:planning/classes/druid","page":41}],"parent_id":"srd-5.2.1:character/class/druid","level":20,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/class-progression"],"select":0,"option_keys":[],"capacity":0,"progression":{"proficiency_bonus":6,"feature_refs":["srd-5.2.1:character/feature/druid/archdruid"],"feature_keys":[]}}`,
	`{"id":"srd-5.2.1:character/feature/druid/archdruid","kind":"feature","name":"Archdruid","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-20-archdruid","page":43}],"parent_id":"srd-5.2.1:character/level/druid/20","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/nature-magician-uses"],"resource_refs":["srd-5.2.1:character/resource/druid/nature-magician"],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/nature-magician-unexpended"],"required_mechanics":["srd-5.2.1:mechanic/druid/archdruid/initiative-zero-wild-shape-regain-one","srd-5.2.1:mechanic/druid/initiative-recovered-uses","srd-5.2.1:mechanic/druid/archdruid/nature-magician-no-action-convert-unexpended-uses-single-slot","srd-5.2.1:mechanic/druid/spell-levels-per-use","srd-5.2.1:mechanic/druid/archdruid/longevity-ten-years-per-body-year","srd-5.2.1:mechanic/druid/elapsed-years","srd-5.2.1:mechanic/druid/body-years"],"select":0,"option_keys":[],"capacity":0,"parameters":{"srd-5.2.1:mechanic/druid/initiative-recovered-uses":1,"srd-5.2.1:mechanic/druid/spell-levels-per-use":2,"srd-5.2.1:mechanic/druid/elapsed-years":10,"srd-5.2.1:mechanic/druid/body-years":1}}`,
	`{"id":"srd-5.2.1:character/choice/druid/cantrips","kind":"choice","name":"Cantrips","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/cantrips"],"select":2,"option_keys":["srd-5.2.1:mechanic/druid/spell-list/p44-45/cantrips"],"capacity":0,"values_by_level":[{"level":1,"value":2},{"level":2,"value":2},{"level":3,"value":2},{"level":4,"value":3},{"level":5,"value":3},{"level":6,"value":3},{"level":7,"value":3},{"level":8,"value":3},{"level":9,"value":3},{"level":10,"value":4},{"level":11,"value":4},{"level":12,"value":4},{"level":13,"value":4},{"level":14,"value":4},{"level":15,"value":4},{"level":16,"value":4},{"level":17,"value":4},{"level":18,"value":4},{"level":19,"value":4},{"level":20,"value":4}]}`,
	`{"id":"srd-5.2.1:character/choice/druid/prepared-spells","kind":"choice","name":"Prepared Spells","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/prepared-spells-slot-level"],"required_mechanics":["srd-5.2.1:mechanic/druid/prepared-spells"],"select":4,"option_keys":["srd-5.2.1:mechanic/druid/spell-list/p44-45/level-1-plus"],"capacity":0,"values_by_level":[{"level":1,"value":4},{"level":2,"value":5},{"level":3,"value":6},{"level":4,"value":7},{"level":5,"value":9},{"level":6,"value":10},{"level":7,"value":11},{"level":8,"value":12},{"level":9,"value":14},{"level":10,"value":15},{"level":11,"value":16},{"level":12,"value":16},{"level":13,"value":17},{"level":14,"value":17},{"level":15,"value":18},{"level":16,"value":18},{"level":17,"value":19},{"level":18,"value":20},{"level":19,"value":21},{"level":20,"value":22}]}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/prepared-spells-slot-level","kind":"prerequisite","name":"Prepared Spells Slot Level","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/choice/druid/prepared-spells","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/prepared-level-has-slot"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/spellcasting/prepared-level-has-slot","record_refs":["srd-5.2.1:character/feature/druid/spellcasting"],"parameters":{"minimum-spell-level":1}}}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-1","kind":"resource","name":"Spell Slot 1","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/1"],"select":0,"option_keys":[],"capacity":2,"values_by_level":[{"level":1,"value":2},{"level":2,"value":3},{"level":3,"value":4},{"level":4,"value":4},{"level":5,"value":4},{"level":6,"value":4},{"level":7,"value":4},{"level":8,"value":4},{"level":9,"value":4},{"level":10,"value":4},{"level":11,"value":4},{"level":12,"value":4},{"level":13,"value":4},{"level":14,"value":4},{"level":15,"value":4},{"level":16,"value":4},{"level":17,"value":4},{"level":18,"value":4},{"level":19,"value":4},{"level":20,"value":4}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-2","kind":"resource","name":"Spell Slot 2","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/2"],"select":0,"option_keys":[],"capacity":2,"values_by_level":[{"level":3,"value":2},{"level":4,"value":3},{"level":5,"value":3},{"level":6,"value":3},{"level":7,"value":3},{"level":8,"value":3},{"level":9,"value":3},{"level":10,"value":3},{"level":11,"value":3},{"level":12,"value":3},{"level":13,"value":3},{"level":14,"value":3},{"level":15,"value":3},{"level":16,"value":3},{"level":17,"value":3},{"level":18,"value":3},{"level":19,"value":3},{"level":20,"value":3}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-3","kind":"resource","name":"Spell Slot 3","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/3"],"select":0,"option_keys":[],"capacity":2,"values_by_level":[{"level":5,"value":2},{"level":6,"value":3},{"level":7,"value":3},{"level":8,"value":3},{"level":9,"value":3},{"level":10,"value":3},{"level":11,"value":3},{"level":12,"value":3},{"level":13,"value":3},{"level":14,"value":3},{"level":15,"value":3},{"level":16,"value":3},{"level":17,"value":3},{"level":18,"value":3},{"level":19,"value":3},{"level":20,"value":3}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-4","kind":"resource","name":"Spell Slot 4","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/4"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":7,"value":1},{"level":8,"value":2},{"level":9,"value":3},{"level":10,"value":3},{"level":11,"value":3},{"level":12,"value":3},{"level":13,"value":3},{"level":14,"value":3},{"level":15,"value":3},{"level":16,"value":3},{"level":17,"value":3},{"level":18,"value":3},{"level":19,"value":3},{"level":20,"value":3}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-5","kind":"resource","name":"Spell Slot 5","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/5"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":9,"value":1},{"level":10,"value":2},{"level":11,"value":2},{"level":12,"value":2},{"level":13,"value":2},{"level":14,"value":2},{"level":15,"value":2},{"level":16,"value":2},{"level":17,"value":2},{"level":18,"value":3},{"level":19,"value":3},{"level":20,"value":3}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-6","kind":"resource","name":"Spell Slot 6","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/6"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":11,"value":1},{"level":12,"value":1},{"level":13,"value":1},{"level":14,"value":1},{"level":15,"value":1},{"level":16,"value":1},{"level":17,"value":1},{"level":18,"value":1},{"level":19,"value":2},{"level":20,"value":2}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-7","kind":"resource","name":"Spell Slot 7","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/7"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":13,"value":1},{"level":14,"value":1},{"level":15,"value":1},{"level":16,"value":1},{"level":17,"value":1},{"level":18,"value":1},{"level":19,"value":1},{"level":20,"value":2}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-8","kind":"resource","name":"Spell Slot 8","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/8"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":15,"value":1},{"level":16,"value":1},{"level":17,"value":1},{"level":18,"value":1},{"level":19,"value":1},{"level":20,"value":1}]}`,
	`{"id":"srd-5.2.1:character/resource/druid/spell-slot-9","kind":"resource","name":"Spell Slot 9","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-spellcasting","page":41}],"parent_id":"srd-5.2.1:character/feature/druid/spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/spellcasting/long-rest-all-slots","srd-5.2.1:mechanic/druid/spell-slot-level/9"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":17,"value":1},{"level":18,"value":1},{"level":19,"value":1},{"level":20,"value":1}]}`,
	`{"id":"srd-5.2.1:character/choice/druid/primal-order","kind":"choice","name":"Primal Order","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-primal-order","page":42}],"parent_id":"srd-5.2.1:character/feature/druid/primal-order","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/magician","srd-5.2.1:character/choice/druid/warden"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/primal-order"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/primal-order/magician","srd-5.2.1:mechanic/druid/primal-order/warden"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/magician","kind":"choice","name":"Magician","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-primal-order","page":42}],"parent_id":"srd-5.2.1:character/choice/druid/primal-order","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/magician-selected"],"required_mechanics":["srd-5.2.1:mechanic/druid/magician/extra-cantrip","srd-5.2.1:mechanic/druid/magician/arcana-nature-wisdom-modifier-minimum-1"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/spell-list/p44-45/cantrips"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/warden","kind":"choice","name":"Warden","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-primal-order","page":42}],"parent_id":"srd-5.2.1:character/choice/druid/primal-order","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/warden-selected"],"required_mechanics":["srd-5.2.1:mechanic/weapons/martial","srd-5.2.1:mechanic/armor/medium"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/warden/grant-martial-weapons-medium-armor"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/magician-selected","kind":"prerequisite","name":"Magician Selected","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-primal-order","page":42}],"parent_id":"srd-5.2.1:character/choice/druid/magician","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/primal-order/magician-selected"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/primal-order/magician-selected","record_refs":["srd-5.2.1:character/choice/druid/primal-order"],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/warden-selected","kind":"prerequisite","name":"Warden Selected","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-1-primal-order","page":42}],"parent_id":"srd-5.2.1:character/choice/druid/warden","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/primal-order/warden-selected"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/primal-order/warden-selected","record_refs":["srd-5.2.1:character/choice/druid/primal-order"],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/resource/druid/wild-shape","kind":"resource","name":"Wild Shape","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-2-wild-shape","page":42}],"parent_id":"srd-5.2.1:character/feature/druid/wild-shape","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-shape/short-rest-recover-one","srd-5.2.1:mechanic/druid/wild-shape/long-rest-recover-all"],"select":0,"option_keys":[],"capacity":2,"values_by_level":[{"level":2,"value":2},{"level":3,"value":2},{"level":4,"value":2},{"level":5,"value":2},{"level":6,"value":3},{"level":7,"value":3},{"level":8,"value":3},{"level":9,"value":3},{"level":10,"value":3},{"level":11,"value":3},{"level":12,"value":3},{"level":13,"value":3},{"level":14,"value":3},{"level":15,"value":3},{"level":16,"value":3},{"level":17,"value":4},{"level":18,"value":4},{"level":19,"value":4},{"level":20,"value":4}]}`,
	`{"id":"srd-5.2.1:character/choice/druid/known-forms","kind":"choice","name":"Known Forms","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-2-wild-shape","page":42}],"parent_id":"srd-5.2.1:character/feature/druid/wild-shape","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/known-form-eligibility"],"required_mechanics":["srd-5.2.1:mechanic/druid/known-forms","srd-5.2.1:mechanic/druid/wild-shape/long-rest-replace-one-known-form","srd-5.2.1:mechanic/druid/recommended-forms/rat","srd-5.2.1:mechanic/druid/recommended-forms/riding-horse","srd-5.2.1:mechanic/druid/recommended-forms/spider","srd-5.2.1:mechanic/druid/recommended-forms/wolf"],"select":4,"option_keys":["srd-5.2.1:mechanic/druid/beasts/eligible-stat-blocks","srd-5.2.1:mechanic/druid/beasts/other-sources-with-gm-permission"],"capacity":0,"values_by_level":[{"level":2,"value":4},{"level":3,"value":4},{"level":4,"value":6},{"level":5,"value":6},{"level":6,"value":6},{"level":7,"value":6},{"level":8,"value":8},{"level":9,"value":8},{"level":10,"value":8},{"level":11,"value":8},{"level":12,"value":8},{"level":13,"value":8},{"level":14,"value":8},{"level":15,"value":8},{"level":16,"value":8},{"level":17,"value":8},{"level":18,"value":8},{"level":19,"value":8},{"level":20,"value":8}]}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/known-form-eligibility","kind":"prerequisite","name":"Known Form Eligibility","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-2-wild-shape","page":42}],"parent_id":"srd-5.2.1:character/choice/druid/known-forms","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-shape/beast-cr-limit-no-fly-before-level-8"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/wild-shape/beast-cr-limit-no-fly-before-level-8","record_refs":["srd-5.2.1:character/feature/druid/wild-shape"],"parameters":{"fly-speed-minimum-druid-level":8}}}`,
	`{"id":"srd-5.2.1:character/choice/druid/shape-equipment","kind":"choice","name":"Shape Equipment","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-2-wild-shape","page":42}],"parent_id":"srd-5.2.1:character/feature/druid/wild-shape","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/shape-equipment","srd-5.2.1:mechanic/druid/wild-shape/worn-equipment-functions-normally-gm-size-shape-practicality","srd-5.2.1:mechanic/druid/wild-shape/equipment-does-not-resize-reshape","srd-5.2.1:mechanic/druid/wild-shape/unwearable-must-fall-or-merge","srd-5.2.1:mechanic/druid/wild-shape/merged-equipment-no-effect"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/equipment/fall-in-space","srd-5.2.1:mechanic/druid/equipment/merge","srd-5.2.1:mechanic/druid/equipment/wear"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/wild-companion-cost","kind":"choice","name":"Wild Companion Cost","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-2-wild-companion","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/wild-companion","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-companion-cost"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/cost/one-spell-slot","srd-5.2.1:mechanic/druid/cost/one-wild-shape-use"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/druid-subclass","kind":"choice","name":"Druid Subclass","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-3-druid-subclass","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/druid-subclass","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/druid-subclass"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/subclasses","srd-5.2.1:mechanic/druid/subclasses/circle-of-the-land"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/ability-score-improvement","kind":"choice","name":"Ability Score Improvement","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-4-ability-score-improvement","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/ability-score-improvement","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/ability-score-improvement-qualification"],"required_mechanics":["srd-5.2.1:mechanic/druid/ability-score-improvement"],"select":1,"option_keys":["srd-5.2.1:mechanic/feats/ability-score-improvement","srd-5.2.1:mechanic/feats/qualified"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/ability-score-improvement-qualification","kind":"prerequisite","name":"Ability Score Improvement Qualification","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-4-ability-score-improvement","page":43}],"parent_id":"srd-5.2.1:character/choice/druid/ability-score-improvement","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/prerequisite/feat-qualification"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/prerequisite/feat-qualification","record_refs":[],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/choice/druid/epic-boon","kind":"choice","name":"Epic Boon","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-19-epic-boon","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/epic-boon","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/epic-boon-qualification"],"required_mechanics":["srd-5.2.1:mechanic/druid/epic-boon"],"select":1,"option_keys":["srd-5.2.1:mechanic/feats/epic-boon","srd-5.2.1:mechanic/feats/qualified"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/epic-boon-qualification","kind":"prerequisite","name":"Epic Boon Qualification","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-19-epic-boon","page":43}],"parent_id":"srd-5.2.1:character/choice/druid/epic-boon","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/prerequisite/feat-qualification"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/prerequisite/feat-qualification","record_refs":[],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/choice/druid/wild-resurgence-conversion","kind":"choice","name":"Wild Resurgence Conversion","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-5-wild-resurgence","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/wild-resurgence","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-resurgence-conversion"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/conversion/spell-slot-to-one-wild-shape","srd-5.2.1:mechanic/druid/conversion/one-wild-shape-to-level-1-slot"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/wild-resurgence-empty","kind":"prerequisite","name":"Wild Resurgence Empty","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-5-wild-resurgence","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/wild-resurgence","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-resurgence/slot-to-use-requires-zero-uses"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/wild-resurgence/slot-to-use-requires-zero-uses","record_refs":["srd-5.2.1:character/resource/druid/wild-shape"],"parameters":{"maximum-remaining-uses":0}}}`,
	`{"id":"srd-5.2.1:character/resource/druid/wild-resurgence-slot-conversion","kind":"resource","name":"Wild Resurgence Slot Conversion","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-5-wild-resurgence","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/wild-resurgence","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/wild-resurgence/one-use-to-level-1-slot-no-action-once-long-rest"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":5,"value":1},{"level":6,"value":1},{"level":7,"value":1},{"level":8,"value":1},{"level":9,"value":1},{"level":10,"value":1},{"level":11,"value":1},{"level":12,"value":1},{"level":13,"value":1},{"level":14,"value":1},{"level":15,"value":1},{"level":16,"value":1},{"level":17,"value":1},{"level":18,"value":1},{"level":19,"value":1},{"level":20,"value":1}]}`,
	`{"id":"srd-5.2.1:character/choice/druid/elemental-fury","kind":"choice","name":"Elemental Fury","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-7-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/elemental-fury","level":0,"choice_refs":["srd-5.2.1:character/choice/druid/potent-spellcasting","srd-5.2.1:character/choice/druid/primal-strike"],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/elemental-fury"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/elemental-fury/potent-spellcasting","srd-5.2.1:mechanic/druid/elemental-fury/primal-strike"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/potent-spellcasting","kind":"choice","name":"Potent Spellcasting","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-7-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/choice/druid/elemental-fury","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/potent-spellcasting-selected"],"required_mechanics":["srd-5.2.1:mechanic/druid/potent-spellcasting"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/potent-spellcasting/add-wisdom-to-druid-cantrip-damage"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/choice/druid/primal-strike","kind":"choice","name":"Primal Strike","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-7-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/choice/druid/elemental-fury","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":["srd-5.2.1:character/prerequisite/druid/primal-strike-selected"],"required_mechanics":["srd-5.2.1:mechanic/druid/primal-strike/once-own-turn-weapon-or-wild-shape-beast-attack-roll-hit-creature-choose-on-hit"],"select":1,"option_keys":["srd-5.2.1:mechanic/damage/cold","srd-5.2.1:mechanic/damage/fire","srd-5.2.1:mechanic/damage/lightning","srd-5.2.1:mechanic/damage/thunder"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/potent-spellcasting-selected","kind":"prerequisite","name":"Potent Spellcasting Selected","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-7-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/choice/druid/potent-spellcasting","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/elemental-fury/potent-spellcasting-selected"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/elemental-fury/potent-spellcasting-selected","record_refs":["srd-5.2.1:character/choice/druid/elemental-fury"],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/primal-strike-selected","kind":"prerequisite","name":"Primal Strike Selected","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-7-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/choice/druid/primal-strike","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/elemental-fury/primal-strike-selected"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/elemental-fury/primal-strike-selected","record_refs":["srd-5.2.1:character/choice/druid/elemental-fury"],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/improved-fury-choice","kind":"prerequisite","name":"Improved Fury Choice","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-15-improved-elemental-fury","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/improved-elemental-fury","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/requires-elemental-fury-choice"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/requires-elemental-fury-choice","record_refs":["srd-5.2.1:character/choice/druid/elemental-fury"],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/beast-spells-components","kind":"prerequisite","name":"Beast Spells Components","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-18-beast-spells","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/beast-spells","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/beast-spells/exclude-costed-or-consumed-material"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/beast-spells/exclude-costed-or-consumed-material","record_refs":["srd-5.2.1:character/feature/druid/wild-shape"],"parameters":{}}}`,
	`{"id":"srd-5.2.1:character/choice/druid/nature-magician-uses","kind":"choice","name":"Nature Magician Uses","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-20-archdruid","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/archdruid","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/nature-magician-uses"],"select":1,"option_keys":["srd-5.2.1:mechanic/druid/archdruid/choose-number-of-unexpended-wild-shape-uses"],"capacity":0}`,
	`{"id":"srd-5.2.1:character/resource/druid/nature-magician","kind":"resource","name":"Nature Magician","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-20-archdruid","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/archdruid","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/archdruid/nature-magician-once-long-rest"],"select":0,"option_keys":[],"capacity":1,"values_by_level":[{"level":20,"value":1}]}`,
	`{"id":"srd-5.2.1:character/prerequisite/druid/nature-magician-unexpended","kind":"prerequisite","name":"Nature Magician Unexpended","sources":[{"inventory_id":"srd-5.2.1:planning/class-features/druid-level-20-archdruid","page":43}],"parent_id":"srd-5.2.1:character/feature/druid/archdruid","level":0,"choice_refs":[],"resource_refs":[],"prerequisite_refs":[],"required_mechanics":["srd-5.2.1:mechanic/druid/archdruid/requires-unexpended-wild-shape"],"select":0,"option_keys":[],"capacity":0,"requirement":{"key":"srd-5.2.1:mechanic/druid/archdruid/requires-unexpended-wild-shape","record_refs":["srd-5.2.1:character/resource/druid/wild-shape"],"parameters":{"minimum-uses":1}}}`,
}

// druidSourceDifferences compares individual semantic fields and reports paths;
// controls below must remain schema-valid yet fail these source assertions.
func druidSourceDifferences(t *testing.T, c CharacterCatalogue) []string {
	t.Helper()
	var differences []string
	if len(c.Records) != len(druidSourceAssertions) {
		differences = append(differences, "record count")
	}
	actual := map[string]map[string]any{}
	for _, r := range c.Records {
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(b, &fields); err != nil {
			t.Fatal(err)
		}
		actual[r.ID] = fields
	}
	for _, assertion := range druidSourceAssertions {
		var want map[string]any
		if err := json.Unmarshal([]byte(assertion), &want); err != nil {
			t.Fatal(err)
		}
		id := want["id"].(string)
		got := actual[id]
		if len(got) != len(want) {
			differences = append(differences, id+"/field-set")
		}
		for field, expected := range want {
			if !reflect.DeepEqual(got[field], expected) {
				differences = append(differences, fmt.Sprintf("%s/%s: got %v, source %v", id, field, got[field], expected))
			}
		}
	}
	return differences
}

func TestDruidCompleteSource(t *testing.T) {
	c := druidCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	if d := druidSourceDifferences(t, c); len(d) > 0 {
		t.Fatal(strings.Join(d, "\n"))
	}
	// Check exact canonical identities against the independent planning inventory.
	b, err := os.ReadFile("../../../docs/investigations/issue-74-character-inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	var inv struct {
		Records []struct {
			ID   string `json:"id"`
			Page int    `json:"source_page"`
		}
	}
	if err := json.Unmarshal(b, &inv); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{}
	for _, r := range inv.Records {
		if r.ID == "srd-5.2.1:planning/classes/druid" || strings.HasPrefix(r.ID, "srd-5.2.1:planning/class-features/druid-level-") {
			want[r.ID] = r.Page
		}
	}
	got := map[string]int{}
	for _, r := range c.Records {
		if r.Kind == "class" || r.Kind == "feature" {
			got[r.Sources[0].InventoryID] = r.Sources[0].Page
		}
	}
	if len(got) != 14 || !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical identities/pages: %v", got)
	}
	if c.Source.SHA256 != "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87" {
		t.Fatal("pin")
	}
}

func TestDruidContrarySourceControls(t *testing.T) {
	tests := []struct {
		name, id string
		change   func(*CharacterRecord)
	}{
		{"numeric range", "feature/druid/improved-elemental-fury", func(r *CharacterRecord) { r.Parameters["srd-5.2.1:mechanic/druid/potent-added-range-feet"] = 301 }},
		{"redirected selection operand", "prerequisite/druid/magician-selected", func(r *CharacterRecord) {
			r.Requirement.RecordRefs = []string{"srd-5.2.1:character/choice/druid/elemental-fury"}
		}},
		{"contrary damage option", "choice/druid/primal-strike", func(r *CharacterRecord) { r.OptionKeys = []string{"srd-5.2.1:mechanic/damage/acid"} }},
		{"contrary recovery", "resource/druid/wild-shape", func(r *CharacterRecord) {
			r.RequiredMechanics = []string{"srd-5.2.1:mechanic/druid/wild-shape/short-rest-recover-all"}
		}},
		{"slot capacity", "resource/druid/spell-slot-9", func(r *CharacterRecord) { r.ValuesByLevel[1].Value = 2 }},
		{"form CR", "feature/druid/wild-shape", func(r *CharacterRecord) { r.Parameters["srd-5.2.1:mechanic/druid/cr-denominator-at-level-2"] = 2 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := druidCatalogue(t)
			found := false
			for i := range c.Records {
				if c.Records[i].ID == "srd-5.2.1:character/"+tt.id {
					tt.change(&c.Records[i])
					found = true
				}
			}
			if !found {
				t.Fatal("control target missing")
			}
			if err := c.Validate(os.DirFS("../../..")); err != nil {
				t.Fatalf("control must be structurally valid: %v", err)
			}
			if len(druidSourceDifferences(t, c)) == 0 {
				t.Fatal("source proof accepted contrary mechanics")
			}
		})
	}
}

func TestDruidGuardsReadinessComposition(t *testing.T) {
	repo := os.DirFS("../../..")
	c := druidCatalogue(t)
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("catalogue enabled support")
	}
	for _, r := range c.Records {
		for _, k := range append(append([]string{}, r.RequiredMechanics...), r.OptionKeys...) {
			if qualifiedCharacterMechanic(k) {
				t.Fatal("qualified", k)
			}
		}
	}
	c.Records[0].RequiredMechanics = append(c.Records[0].RequiredMechanics, "srd-5.2.1:mechanic/unknown/future")
	if err := c.Validate(repo); err != nil {
		t.Fatal(err)
	}
	if err := c.ValidateReadiness(repo); err == nil {
		t.Fatal("unknown key enabled")
	}
	tests := []struct {
		name   string
		change func(*CharacterCatalogue)
	}{
		{"wrong page", func(c *CharacterCatalogue) { c.Records[0].Sources[0].Page = 42 }},
		{"future grant", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Progression != nil {
					c.Records[i].Progression.FeatureRefs = []string{"srd-5.2.1:character/feature/druid/archdruid"}
					break
				}
			}
		}},
		{"missing owner link", func(c *CharacterCatalogue) { c.Records[0].ChoiceRefs = []string{} }},
		{"missing operand", func(c *CharacterCatalogue) {
			for i := range c.Records {
				if c.Records[i].Requirement != nil {
					c.Records[i].Requirement.RecordRefs = []string{"srd-5.2.1:character/choice/druid/missing"}
					break
				}
			}
		}},
		{"unrequired number", func(c *CharacterCatalogue) { c.Records[0].Parameters["srd-5.2.1:mechanic/unknown/value"] = 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := druidCatalogue(t)
			tt.change(&c)
			if err := c.Validate(repo); err == nil {
				t.Fatal("malformed metadata accepted")
			}
		})
	}
	b, err := os.ReadFile("testdata/character-druid.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(string(b), `"kind": "class"`, `"kind": "class", "playable": true`, 1), string(b) + `{}`} {
		if _, err := DecodeCharacterCatalogue(strings.NewReader(bad)); err == nil {
			t.Fatal("strict decode accepted")
		}
	}
	// Composition with existing canonical class metadata stays unavailable and
	// introduces no replacement definitions or executable consumers.
	for _, file := range []string{"character-cleric.json", "character-fighter.json", "character-barbarian.json"} {
		b, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		other, err := DecodeCharacterCatalogue(strings.NewReader(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		merged := druidCatalogue(t)
		merged.Records = append(merged.Records, other.Records...)
		if err := merged.Validate(repo); err != nil {
			t.Fatal(err)
		}
		if err := merged.ValidateReadiness(repo); err == nil {
			t.Fatal("composition qualified")
		}
	}
}

func TestDruidImportDoesNotPromoteFixedBuilds(t *testing.T) {
	beforeFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	beforeCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	c := druidCatalogue(t)
	if err := c.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	afterFighter, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	afterCleric, err := SupportedBuildForConcept("Human Cleric", "Acolyte")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeFighter, afterFighter) || !reflect.DeepEqual(beforeCleric, afterCleric) {
		t.Fatal("catalogue changed fixed build creation")
	}
	for _, background := range []string{"Acolyte", "Soldier", "Guide", "Hermit"} {
		if _, err := SupportedBuildForConcept("Human Druid", background); err == nil {
			t.Fatal("Druid creation became available")
		}
	}
	// Existing fixed paths retain their numeric/projection assertions too.
	TestNewSupportedMartialBuild(t)
	TestSupportedCasterCreation(t)
	TestCharacterProjection(t)
}
