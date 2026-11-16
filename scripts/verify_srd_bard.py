#!/usr/bin/env python3
"""Offline #287 comparison with the pinned licensed SRD pp31–33; no network."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
ID = 'srd-5.2.1:character/'
KEY = 'srd-5.2.1:mechanic/'


def prose_facts(records):
    # Lossless prefix abbreviations keep the explicit source facts readable.
    def compact(value):
        if isinstance(value, str):
            for prefix, short in ((ID, 'c:'), (KEY, 'm:'), ('srd-5.2.1:planning/', 'p:')):
                if value.startswith(prefix):
                    return short + value[len(prefix):]
            assert not value.startswith(('c:', 'm:', 'p:')), 'reserved source abbreviation'
            return value
        if isinstance(value, list):
            return [compact(item) for item in value]
        if isinstance(value, dict):
            return {compact(key): compact(item) for key, item in value.items()}
        return value

    result = {}
    for record in records.values():
        # Spell counts are compared with every printed table cell below.
        table_count = record['id'] in {
            ID + 'choice/bard/cantrips', ID + 'choice/bard/prepared-spells',
            *(ID + f'resource/bard/spell-slot-{n}' for n in range(1, 10)),
        }
        facts = {key: value for key, value in record.items()
                 if key != 'id' and not (key == 'values_by_level' and table_count)
                 and not (record['kind'] == 'level' and key in ('progression', 'parameters'))
                 and not (value is None or (type(value) in (str, list, dict) and not value)
                          or (type(value) is int and value == 0))}
        result[compact(record['id'])] = compact(facts)
    return result


# Explicit source facts reviewed against pp31–33. Unlike a digest, these name
# each mechanic, predicate operand, choice, number, owner and recovery rule.
PROSE_FACTS = {
    'c:class/bard': {'kind': 'class', 'name': 'Bard', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'choice_refs': ['c:choice/bard/skills', 'c:choice/bard/musical-instruments', 'c:choice/bard/starting-equipment', 'c:choice/bard/multiclass-traits'], 'required_mechanics': ['m:bard/primary-ability/charisma', 'm:bard/saving-throws/dexterity-charisma', 'm:weapons/simple', 'm:armor/light', 'm:bard/level-one/all-core-traits-and-level-one-features', 'm:bard/hit-die-sides'], 'parameters': {'m:bard/hit-die-sides': 8}},
    'c:level/bard/1': {'kind': 'level', 'name': 'Bard level 1', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 1, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/2': {'kind': 'level', 'name': 'Bard level 2', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 2, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/3': {'kind': 'level', 'name': 'Bard level 3', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 3, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/4': {'kind': 'level', 'name': 'Bard level 4', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 4, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/5': {'kind': 'level', 'name': 'Bard level 5', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 5, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/6': {'kind': 'level', 'name': 'Bard level 6', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 6, 'required_mechanics': ['m:bard/class-progression', 'm:bard/subclass-feature', 'm:bard/bardic-die-sides']},
    'c:level/bard/7': {'kind': 'level', 'name': 'Bard level 7', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 7, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/8': {'kind': 'level', 'name': 'Bard level 8', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 8, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/9': {'kind': 'level', 'name': 'Bard level 9', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 9, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/10': {'kind': 'level', 'name': 'Bard level 10', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 10, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/11': {'kind': 'level', 'name': 'Bard level 11', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 11, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/12': {'kind': 'level', 'name': 'Bard level 12', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 12, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/13': {'kind': 'level', 'name': 'Bard level 13', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 13, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/14': {'kind': 'level', 'name': 'Bard level 14', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 14, 'required_mechanics': ['m:bard/class-progression', 'm:bard/subclass-feature', 'm:bard/bardic-die-sides']},
    'c:level/bard/15': {'kind': 'level', 'name': 'Bard level 15', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 15, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/16': {'kind': 'level', 'name': 'Bard level 16', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 16, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/17': {'kind': 'level', 'name': 'Bard level 17', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 17, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/18': {'kind': 'level', 'name': 'Bard level 18', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 18, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/19': {'kind': 'level', 'name': 'Bard level 19', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 19, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:level/bard/20': {'kind': 'level', 'name': 'Bard level 20', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'level': 20, 'required_mechanics': ['m:bard/class-progression', 'm:bard/bardic-die-sides']},
    'c:feature/bard/bardic-inspiration': {'kind': 'feature', 'name': 'Bardic Inspiration', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-bardic-inspiration', 'page': 31}], 'parent_id': 'c:level/bard/1', 'resource_refs': ['c:resource/bard/bardic-inspiration'], 'required_mechanics': ['m:bard/bardic-inspiration', 'm:bard/inspiration/bonus-action-other-creature', 'm:bard/inspiration/target-see-or-hear-bard', 'm:bard/inspiration/one-die-per-creature', 'm:bard/inspiration/failed-d20-test-add-roll-once', 'm:bard/inspiration/expended-when-rolled', 'm:bard/inspiration/range-feet', 'm:bard/inspiration/duration-hours', 'm:bard/inspiration/minimum-uses', 'm:bard/inspiration/dice-per-creature', 'm:bard/inspiration/uses-per-die'], 'parameters': {'m:bard/inspiration/range-feet': 60, 'm:bard/inspiration/duration-hours': 1, 'm:bard/inspiration/minimum-uses': 1, 'm:bard/inspiration/dice-per-creature': 1, 'm:bard/inspiration/uses-per-die': 1}},
    'c:feature/bard/spellcasting': {'kind': 'feature', 'name': 'Spellcasting', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:level/bard/1', 'choice_refs': ['c:choice/bard/cantrips', 'c:choice/bard/prepared-spells'], 'resource_refs': ['c:resource/bard/spell-slot-1', 'c:resource/bard/spell-slot-2', 'c:resource/bard/spell-slot-3', 'c:resource/bard/spell-slot-4', 'c:resource/bard/spell-slot-5', 'c:resource/bard/spell-slot-6', 'c:resource/bard/spell-slot-7', 'c:resource/bard/spell-slot-8', 'c:resource/bard/spell-slot-9'], 'required_mechanics': ['m:bard/spellcasting', 'm:bard/spellcasting/charisma', 'm:bard/spellcasting/musical-instrument-focus', 'm:bard/spellcasting/replace-one-cantrip-on-bard-level', 'm:bard/spellcasting/replace-one-prepared-on-bard-level-with-slot-eligible-bard-spell', 'm:bard/spellcasting/increase-prepared-to-table-count-slot-eligible', 'm:bard/spellcasting/always-prepared-excluded-from-count-count-as-bard', 'm:bard/spellcasting/cantrips-replaced-per-level', 'm:bard/spellcasting/prepared-replaced-per-level'], 'parameters': {'m:bard/spellcasting/cantrips-replaced-per-level': 1, 'm:bard/spellcasting/prepared-replaced-per-level': 1}},
    'c:feature/bard/expertise': {'kind': 'feature', 'name': 'Expertise', 'sources': [{'inventory_id': 'p:class-features/bard-level-2-expertise', 'page': 32}], 'parent_id': 'c:level/bard/2', 'choice_refs': ['c:choice/bard/expertise'], 'required_mechanics': ['m:bard/expertise']},
    'c:feature/bard/jack-of-all-trades': {'kind': 'feature', 'name': 'Jack of All Trades', 'sources': [{'inventory_id': 'p:class-features/bard-level-2-jack-of-all-trades', 'page': 32}], 'parent_id': 'c:level/bard/2', 'prerequisite_refs': ['c:prerequisite/bard/jack-of-all-trades'], 'required_mechanics': ['m:bard/jack-of-all-trades', 'm:bard/jack-of-all-trades/half-proficiency-round-down']},
    'c:feature/bard/bard-subclass': {'kind': 'feature', 'name': 'Bard Subclass', 'sources': [{'inventory_id': 'p:class-features/bard-level-3-bard-subclass', 'page': 32}], 'parent_id': 'c:level/bard/3', 'choice_refs': ['c:choice/bard/bard-subclass'], 'required_mechanics': ['m:bard/bard-subclass', 'm:bard/subclass/features-at-or-below-bard-level']},
    'c:feature/bard/ability-score-improvement': {'kind': 'feature', 'name': 'Ability Score Improvement', 'sources': [{'inventory_id': 'p:class-features/bard-level-4-ability-score-improvement', 'page': 32}], 'parent_id': 'c:level/bard/4', 'choice_refs': ['c:choice/bard/ability-score-improvement'], 'required_mechanics': ['m:bard/ability-score-improvement', 'm:bard/ability-score-improvement/repeats-levels-8-12-16']},
    'c:feature/bard/font-of-inspiration': {'kind': 'feature', 'name': 'Font of Inspiration', 'sources': [{'inventory_id': 'p:class-features/bard-level-5-font-of-inspiration', 'page': 32}], 'parent_id': 'c:level/bard/5', 'prerequisite_refs': ['c:prerequisite/bard/font-expended-use'], 'required_mechanics': ['m:bard/font-of-inspiration', 'm:bard/inspiration/short-or-long-rest-all', 'm:bard/font-of-inspiration/expend-any-spell-slot-no-action-regain-one', 'm:bard/font-of-inspiration/slots-expended', 'm:bard/font-of-inspiration/uses-regained'], 'parameters': {'m:bard/font-of-inspiration/slots-expended': 1, 'm:bard/font-of-inspiration/uses-regained': 1}},
    'c:feature/bard/countercharm': {'kind': 'feature', 'name': 'Countercharm', 'sources': [{'inventory_id': 'p:class-features/bard-level-7-countercharm', 'page': 33}], 'parent_id': 'c:level/bard/7', 'prerequisite_refs': ['c:prerequisite/bard/countercharm'], 'required_mechanics': ['m:bard/countercharm', 'm:bard/countercharm/reaction-reroll-save-with-advantage', 'm:bard/countercharm/range-feet'], 'parameters': {'m:bard/countercharm/range-feet': 30}},
    'c:feature/bard/magical-secrets': {'kind': 'feature', 'name': 'Magical Secrets', 'sources': [{'inventory_id': 'p:class-features/bard-level-10-magical-secrets', 'page': 33}], 'parent_id': 'c:level/bard/10', 'choice_refs': ['c:choice/bard/magical-secrets'], 'required_mechanics': ['m:bard/magical-secrets']},
    'c:feature/bard/superior-inspiration': {'kind': 'feature', 'name': 'Superior Inspiration', 'sources': [{'inventory_id': 'p:class-features/bard-level-18-superior-inspiration', 'page': 33}], 'parent_id': 'c:level/bard/18', 'prerequisite_refs': ['c:prerequisite/bard/superior-inspiration'], 'required_mechanics': ['m:bard/superior-inspiration', 'm:bard/superior-inspiration/initiative-regain-expended-until-two', 'm:bard/superior-inspiration/recovery-threshold'], 'parameters': {'m:bard/superior-inspiration/recovery-threshold': 2}},
    'c:feature/bard/epic-boon': {'kind': 'feature', 'name': 'Epic Boon', 'sources': [{'inventory_id': 'p:class-features/bard-level-19-epic-boon', 'page': 33}], 'parent_id': 'c:level/bard/19', 'choice_refs': ['c:choice/bard/epic-boon'], 'required_mechanics': ['m:bard/epic-boon', 'm:bard/recommendation/boon-of-spell-recall']},
    'c:feature/bard/words-of-creation': {'kind': 'feature', 'name': 'Words of Creation', 'sources': [{'inventory_id': 'p:class-features/bard-level-20-words-of-creation', 'page': 33}], 'parent_id': 'c:level/bard/20', 'prerequisite_refs': ['c:prerequisite/bard/words-second-target'], 'required_mechanics': ['m:bard/words-of-creation', 'm:spells/power-word-heal', 'm:spells/power-word-kill', 'm:bard/words-of-creation/optional-second-target-when-casting-either-spell', 'm:bard/words-of-creation/second-target-distance-feet', 'm:bard/words-of-creation/additional-targets'], 'always_prepared': [{'level': 20, 'spell_keys': ['m:spells/power-word-heal', 'm:spells/power-word-kill']}], 'parameters': {'m:bard/words-of-creation/second-target-distance-feet': 10, 'm:bard/words-of-creation/additional-targets': 1}},
    'c:choice/bard/skills': {'kind': 'choice', 'name': 'Skills', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'required_mechanics': ['m:bard/skills'], 'select': 3, 'option_keys': ['m:skills/acrobatics', 'm:skills/animal-handling', 'm:skills/arcana', 'm:skills/athletics', 'm:skills/deception', 'm:skills/history', 'm:skills/insight', 'm:skills/intimidation', 'm:skills/investigation', 'm:skills/medicine', 'm:skills/nature', 'm:skills/perception', 'm:skills/performance', 'm:skills/persuasion', 'm:skills/religion', 'm:skills/sleight-of-hand', 'm:skills/stealth', 'm:skills/survival']},
    'c:choice/bard/musical-instruments': {'kind': 'choice', 'name': 'Musical Instruments', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'required_mechanics': ['m:bard/musical-instruments'], 'select': 3, 'option_keys': ['m:tools/musical-instruments']},
    'c:choice/bard/starting-equipment': {'kind': 'choice', 'name': 'Starting Equipment', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'choice_refs': ['c:choice/bard/equipment-instrument'], 'required_mechanics': ['m:bard/starting-equipment', 'm:equipment/leather-armor', 'm:equipment/dagger', 'm:tools/musical-instruments', 'm:equipment/entertainers-pack', 'm:currency/gp'], 'select': 1, 'option_keys': ['m:bard/equipment/a', 'm:bard/equipment/b'], 'options': [{'key': 'm:bard/equipment/a', 'quantities': {'m:equipment/leather-armor': 1, 'm:equipment/dagger': 2, 'm:tools/musical-instruments': 1, 'm:equipment/entertainers-pack': 1, 'm:currency/gp': 19}}, {'key': 'm:bard/equipment/b', 'quantities': {'m:currency/gp': 90}}]},
    'c:choice/bard/equipment-instrument': {'kind': 'choice', 'name': 'Equipment Instrument', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:choice/bard/starting-equipment', 'prerequisite_refs': ['c:prerequisite/bard/equipment-a'], 'required_mechanics': ['m:bard/equipment-instrument'], 'select': 1, 'option_keys': ['m:tools/musical-instruments']},
    'c:prerequisite/bard/equipment-a': {'kind': 'prerequisite', 'name': 'Equipment A', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:choice/bard/equipment-instrument', 'required_mechanics': ['m:bard/equipment/a-selected'], 'requirement': {'key': 'm:bard/equipment/a-selected', 'record_refs': ['c:choice/bard/starting-equipment'], 'parameters': {}}},
    'c:choice/bard/multiclass-traits': {'kind': 'choice', 'name': 'Multiclass Traits', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:class/bard', 'choice_refs': ['c:choice/bard/multiclass-skill', 'c:choice/bard/multiclass-instrument'], 'prerequisite_refs': ['c:prerequisite/bard/multiclass'], 'required_mechanics': ['m:bard/multiclass-traits', 'm:bard/multiclass/d8-light-armor-one-skill-one-instrument-level-one-features', 'm:bard/multiclass/spell-slots-character-creation'], 'select': 1, 'option_keys': ['m:bard/multiclass/core-traits-subset']},
    'c:prerequisite/bard/multiclass': {'kind': 'prerequisite', 'name': 'Multiclass', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:choice/bard/multiclass-traits', 'required_mechanics': ['m:character/multiclass-entry'], 'requirement': {'key': 'm:character/multiclass-entry', 'record_refs': ['c:class/bard'], 'parameters': {}}},
    'c:choice/bard/multiclass-skill': {'kind': 'choice', 'name': 'Multiclass Skill', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:choice/bard/multiclass-traits', 'required_mechanics': ['m:bard/multiclass-skill'], 'select': 1, 'option_keys': ['m:skills/acrobatics', 'm:skills/animal-handling', 'm:skills/arcana', 'm:skills/athletics', 'm:skills/deception', 'm:skills/history', 'm:skills/insight', 'm:skills/intimidation', 'm:skills/investigation', 'm:skills/medicine', 'm:skills/nature', 'm:skills/perception', 'm:skills/performance', 'm:skills/persuasion', 'm:skills/religion', 'm:skills/sleight-of-hand', 'm:skills/stealth', 'm:skills/survival']},
    'c:choice/bard/multiclass-instrument': {'kind': 'choice', 'name': 'Multiclass Instrument', 'sources': [{'inventory_id': 'p:classes/bard', 'page': 31}], 'parent_id': 'c:choice/bard/multiclass-traits', 'required_mechanics': ['m:bard/multiclass-instrument'], 'select': 1, 'option_keys': ['m:tools/musical-instruments']},
    'c:resource/bard/bardic-inspiration': {'kind': 'resource', 'name': 'Bardic Inspiration', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-bardic-inspiration', 'page': 31}], 'parent_id': 'c:feature/bard/bardic-inspiration', 'required_mechanics': ['m:bard/inspiration/charisma-modifier-minimum-one', 'm:bard/inspiration/long-rest-all', 'm:bard/inspiration/level-5-short-rest-all'], 'capacity_mechanic': 'm:bard/inspiration/charisma-modifier-minimum-one'},
    'c:choice/bard/cantrips': {'kind': 'choice', 'name': 'Cantrips', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/cantrips', 'm:bard/recommendation/dancing-lights', 'm:bard/recommendation/vicious-mockery'], 'select': 2, 'option_keys': ['m:bard/spell-list/p33-35/cantrips']},
    'c:choice/bard/prepared-spells': {'kind': 'choice', 'name': 'Prepared Spells', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'prerequisite_refs': ['c:prerequisite/bard/prepared-slot-eligibility'], 'required_mechanics': ['m:bard/prepared-spells', 'm:bard/recommendation/charm-person', 'm:bard/recommendation/color-spray', 'm:bard/recommendation/dissonant-whispers', 'm:bard/recommendation/healing-word'], 'select': 4, 'option_keys': ['m:bard/spell-list/p33-35/level-1-plus']},
    'c:prerequisite/bard/prepared-slot-eligibility': {'kind': 'prerequisite', 'name': 'Prepared Slot Eligibility', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:choice/bard/prepared-spells', 'required_mechanics': ['m:bard/spellcasting/has-slot-for-chosen-spell'], 'requirement': {'key': 'm:bard/spellcasting/has-slot-for-chosen-spell', 'record_refs': ['c:feature/bard/spellcasting'], 'parameters': {}}},
    'c:resource/bard/spell-slot-1': {'kind': 'resource', 'name': 'Spell Slot 1', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 2},
    'c:resource/bard/spell-slot-2': {'kind': 'resource', 'name': 'Spell Slot 2', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 2},
    'c:resource/bard/spell-slot-3': {'kind': 'resource', 'name': 'Spell Slot 3', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 2},
    'c:resource/bard/spell-slot-4': {'kind': 'resource', 'name': 'Spell Slot 4', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 1},
    'c:resource/bard/spell-slot-5': {'kind': 'resource', 'name': 'Spell Slot 5', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 1},
    'c:resource/bard/spell-slot-6': {'kind': 'resource', 'name': 'Spell Slot 6', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 1},
    'c:resource/bard/spell-slot-7': {'kind': 'resource', 'name': 'Spell Slot 7', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 1},
    'c:resource/bard/spell-slot-8': {'kind': 'resource', 'name': 'Spell Slot 8', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 1},
    'c:resource/bard/spell-slot-9': {'kind': 'resource', 'name': 'Spell Slot 9', 'sources': [{'inventory_id': 'p:class-features/bard-level-1-spellcasting', 'page': 32}], 'parent_id': 'c:feature/bard/spellcasting', 'required_mechanics': ['m:bard/spellcasting/long-rest-all-slots'], 'capacity': 1},
    'c:choice/bard/expertise': {'kind': 'choice', 'name': 'Expertise', 'sources': [{'inventory_id': 'p:class-features/bard-level-2-expertise', 'page': 32}], 'parent_id': 'c:feature/bard/expertise', 'prerequisite_refs': ['c:prerequisite/bard/expertise-proficiency'], 'required_mechanics': ['m:bard/expertise', 'm:bard/expertise/recommended-performance-persuasion-if-proficient', 'm:bard/expertise/two-additional-at-level-9'], 'select': 2, 'option_keys': ['m:skills/acrobatics', 'm:skills/animal-handling', 'm:skills/arcana', 'm:skills/athletics', 'm:skills/deception', 'm:skills/history', 'm:skills/insight', 'm:skills/intimidation', 'm:skills/investigation', 'm:skills/medicine', 'm:skills/nature', 'm:skills/perception', 'm:skills/performance', 'm:skills/persuasion', 'm:skills/religion', 'm:skills/sleight-of-hand', 'm:skills/stealth', 'm:skills/survival'], 'values_by_level': [{'level': 2, 'value': 2}, {'level': 3, 'value': 2}, {'level': 4, 'value': 2}, {'level': 5, 'value': 2}, {'level': 6, 'value': 2}, {'level': 7, 'value': 2}, {'level': 8, 'value': 2}, {'level': 9, 'value': 4}, {'level': 10, 'value': 4}, {'level': 11, 'value': 4}, {'level': 12, 'value': 4}, {'level': 13, 'value': 4}, {'level': 14, 'value': 4}, {'level': 15, 'value': 4}, {'level': 16, 'value': 4}, {'level': 17, 'value': 4}, {'level': 18, 'value': 4}, {'level': 19, 'value': 4}, {'level': 20, 'value': 4}]},
    'c:prerequisite/bard/expertise-proficiency': {'kind': 'prerequisite', 'name': 'Expertise Proficiency', 'sources': [{'inventory_id': 'p:class-features/bard-level-2-expertise', 'page': 32}], 'parent_id': 'c:choice/bard/expertise', 'required_mechanics': ['m:character/skill-proficient'], 'requirement': {'key': 'm:character/skill-proficient', 'record_refs': ['c:class/bard'], 'parameters': {}}},
    'c:prerequisite/bard/jack-of-all-trades': {'kind': 'prerequisite', 'name': 'Jack Of All Trades', 'sources': [{'inventory_id': 'p:class-features/bard-level-2-jack-of-all-trades', 'page': 32}], 'parent_id': 'c:feature/bard/jack-of-all-trades', 'required_mechanics': ['m:bard/jack-of-all-trades/ability-check-uses-untrained-skill-and-not-proficiency'], 'requirement': {'key': 'm:bard/jack-of-all-trades/ability-check-uses-untrained-skill-and-not-proficiency', 'record_refs': ['c:class/bard'], 'parameters': {}}},
    'c:choice/bard/bard-subclass': {'kind': 'choice', 'name': 'Bard Subclass', 'sources': [{'inventory_id': 'p:class-features/bard-level-3-bard-subclass', 'page': 32}], 'parent_id': 'c:feature/bard/bard-subclass', 'required_mechanics': ['m:bard/bard-subclass'], 'select': 1, 'option_keys': ['m:bard/subclasses', 'm:bard/subclasses/college-of-lore']},
    'c:choice/bard/ability-score-improvement': {'kind': 'choice', 'name': 'Ability Score Improvement', 'sources': [{'inventory_id': 'p:class-features/bard-level-4-ability-score-improvement', 'page': 32}], 'parent_id': 'c:feature/bard/ability-score-improvement', 'prerequisite_refs': ['c:prerequisite/bard/ability-score-improvement-qualification'], 'required_mechanics': ['m:bard/ability-score-improvement'], 'select': 1, 'option_keys': ['m:feats/ability-score-improvement', 'm:feats/qualified']},
    'c:prerequisite/bard/ability-score-improvement-qualification': {'kind': 'prerequisite', 'name': 'Ability Score Improvement Qualification', 'sources': [{'inventory_id': 'p:class-features/bard-level-4-ability-score-improvement', 'page': 32}], 'parent_id': 'c:choice/bard/ability-score-improvement', 'required_mechanics': ['m:prerequisite/feat-qualification'], 'requirement': {'key': 'm:prerequisite/feat-qualification', 'record_refs': [], 'parameters': {}}},
    'c:choice/bard/epic-boon': {'kind': 'choice', 'name': 'Epic Boon', 'sources': [{'inventory_id': 'p:class-features/bard-level-19-epic-boon', 'page': 33}], 'parent_id': 'c:feature/bard/epic-boon', 'prerequisite_refs': ['c:prerequisite/bard/epic-boon-qualification'], 'required_mechanics': ['m:bard/epic-boon'], 'select': 1, 'option_keys': ['m:feats/epic-boon', 'm:feats/qualified']},
    'c:prerequisite/bard/epic-boon-qualification': {'kind': 'prerequisite', 'name': 'Epic Boon Qualification', 'sources': [{'inventory_id': 'p:class-features/bard-level-19-epic-boon', 'page': 33}], 'parent_id': 'c:choice/bard/epic-boon', 'required_mechanics': ['m:prerequisite/feat-qualification'], 'requirement': {'key': 'm:prerequisite/feat-qualification', 'record_refs': [], 'parameters': {}}},
    'c:prerequisite/bard/font-expended-use': {'kind': 'prerequisite', 'name': 'Font Expended Use', 'sources': [{'inventory_id': 'p:class-features/bard-level-5-font-of-inspiration', 'page': 32}], 'parent_id': 'c:feature/bard/font-of-inspiration', 'required_mechanics': ['m:bard/inspiration/has-expended-use'], 'requirement': {'key': 'm:bard/inspiration/has-expended-use', 'record_refs': ['c:feature/bard/bardic-inspiration', 'c:feature/bard/spellcasting'], 'parameters': {}}},
    'c:prerequisite/bard/countercharm': {'kind': 'prerequisite', 'name': 'Countercharm', 'sources': [{'inventory_id': 'p:class-features/bard-level-7-countercharm', 'page': 33}], 'parent_id': 'c:feature/bard/countercharm', 'required_mechanics': ['m:bard/countercharm/self-or-creature-failed-save-against-charmed-or-frightened'], 'requirement': {'key': 'm:bard/countercharm/self-or-creature-failed-save-against-charmed-or-frightened', 'record_refs': [], 'parameters': {'range-feet': 30}}},
    'c:choice/bard/magical-secrets': {'kind': 'choice', 'name': 'Magical Secrets', 'sources': [{'inventory_id': 'p:class-features/bard-level-10-magical-secrets', 'page': 33}], 'parent_id': 'c:feature/bard/magical-secrets', 'prerequisite_refs': ['c:prerequisite/bard/magical-secrets'], 'required_mechanics': ['m:bard/magical-secrets', 'm:bard/magical-secrets/new-or-replaced-prepared-count-as-bard', 'm:bard/magical-secrets/any-new-prepared-spells-from-listed-lists', 'm:bard/magical-secrets/replacements-from-listed-lists'], 'select': 1, 'option_keys': ['m:bard/spell-list/p33-35/level-1-plus', 'm:cleric/spell-list/level-1-plus', 'm:druid/spell-list/level-1-plus', 'm:wizard/spell-list/level-1-plus']},
    'c:prerequisite/bard/magical-secrets': {'kind': 'prerequisite', 'name': 'Magical Secrets', 'sources': [{'inventory_id': 'p:class-features/bard-level-10-magical-secrets', 'page': 33}], 'parent_id': 'c:choice/bard/magical-secrets', 'required_mechanics': ['m:bard/magical-secrets/bard-level-at-least-10-and-prepared-count-increases-or-replacement'], 'requirement': {'key': 'm:bard/magical-secrets/bard-level-at-least-10-and-prepared-count-increases-or-replacement', 'record_refs': ['c:choice/bard/prepared-spells', 'c:feature/bard/spellcasting'], 'parameters': {'minimum-bard-level': 10}}},
    'c:prerequisite/bard/superior-inspiration': {'kind': 'prerequisite', 'name': 'Superior Inspiration', 'sources': [{'inventory_id': 'p:class-features/bard-level-18-superior-inspiration', 'page': 33}], 'parent_id': 'c:feature/bard/superior-inspiration', 'required_mechanics': ['m:bard/inspiration/remaining-uses-less-than-two'], 'requirement': {'key': 'm:bard/inspiration/remaining-uses-less-than-two', 'record_refs': ['c:feature/bard/bardic-inspiration'], 'parameters': {'uses-threshold': 2}}},
    'c:prerequisite/bard/words-second-target': {'kind': 'prerequisite', 'name': 'Words Second Target', 'sources': [{'inventory_id': 'p:class-features/bard-level-20-words-of-creation', 'page': 33}], 'parent_id': 'c:feature/bard/words-of-creation', 'required_mechanics': ['m:bard/words-of-creation/second-creature-within-ten-feet-of-first'], 'requirement': {'key': 'm:bard/words-of-creation/second-creature-within-ten-feet-of-first', 'record_refs': [], 'parameters': {'distance-feet': 10}}},
}


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, 'PDF pin mismatch'
    text = subprocess.check_output(['pdftotext', '-f', '31', '-l', '33', '-layout', str(pdf), '-'], text=True)
    c = json.loads((ROOT / 'internal/rules/srd521/testdata/character-bard.json').read_text())
    assert c['source']['sha256'] == PIN
    records = {r['id']: r for r in c['records']}
    assert len(records) == len(c['records']), 'duplicate canonical ID'
    inventory = json.loads((ROOT / 'docs/investigations/issue-74-character-inventory.json').read_text())
    expected = {r['id']: r['source_page'] for r in inventory['records'] if r['batch_key'] == 'class-bard'}
    canonical = [r for r in c['records'] if r['kind'] in ('class', 'feature')]
    assert len(canonical) == len(expected) == 13
    assert {s['inventory_id']: s['page'] for r in canonical for s in r['sources']} == expected
    rows = re.findall(r'^\s+(\d{1,2})\s+\+(\d)\s+(.+?)\s{2,}D(\d+)\s+(\d)\s+(\d+)\s+([\d —]+)\s*$', text, re.M)
    assert [int(r[0]) for r in rows] == list(range(1, 21)), '20 source rows required'

    def count(kind, slug, level):
        return next((v['value'] for v in records[ID + kind + '/bard/' + slug]['values_by_level'] if v['level'] == level), 0)

    for level, bonus, names, die, cantrips, prepared, slots in rows:
        n = int(level)
        row = records[f'{ID}level/bard/{n}']
        p = row['progression']
        assert row['sources'][0]['page'] == 31
        assert p['proficiency_bonus'] == int(bonus)
        assert row['parameters'] == {KEY + 'bard/bardic-die-sides': int(die)}
        assert p['feature_keys'] == ([KEY + 'bard/subclass-feature'] if n in (6, 14) else []), 'subclass row requirement'
        actual = [records[ref]['name'] for ref in p['feature_refs']]
        actual += ['Subclass feature' for _ in p['feature_keys']]
        # Rows 1 and 2 wrap onto a second printed line.
        if n == 1:
            names += ' Spellcasting'
        if n == 2:
            names += ' Jack of All Trades'
        assert ', '.join(actual).lower().replace(',', '') == names.lower().replace(',', '').strip().replace('—', ''), (n, actual, names)
        assert count('choice', 'cantrips', n) == int(cantrips)
        assert count('choice', 'prepared-spells', n) == int(prepared)
        expected_slots = [0 if v == '—' else int(v) for v in slots.split()]
        assert len(expected_slots) == 9
        assert [count('resource', f'spell-slot-{j}', n) for j in range(1, 10)] == expected_slots
    # Extract prose by column to avoid interleaving opposite-column sentences.
    columns = [subprocess.check_output(['pdftotext', '-f', str(p), '-l', str(p), '-x', str(x), '-y', '0',
                '-W', '310', '-H', '792', '-layout', str(pdf), '-'], text=True) for p in (31, 32, 33) for x in (0, 310)]
    flat = ' '.join(re.sub(r'-\n\s*', '', ' '.join(columns).replace('\u00ad', '')).split())
    anchors = [
        'D8 per Bard level', 'Dexterity and Charisma', 'Choose any 3 skills', 'Choose 3 Musical Instruments',
        'Simple weapons', 'Light armor', '2 Daggers', '19 GP', '90 GP', 'one skill', 'one Musical Instrument',
        'As a Bonus Action', '60 feet', 'who can see or hear you', 'only one Bardic Inspiration die at a time',
        'within the next hour', 'fails a D20 Test', 'expended when it’s rolled', 'Charisma modifier (minimum of once)',
        'd8 at level 5', 'd10 at level 10', 'd12 at level 15',
        'Dancing Lights and Vicious', 'Mockery are recommended', 'replace one of your cantrips', 'levels 4 and 10',
        'four level 1', 'Charm Person, Color', 'Spray, Dissonant Whispers, and Healing Word',
        'must be of a level for which you have spell slots', 'always have prepared', 'don’t count',
        'replace one spell', 'Charisma is your spellcasting', 'Musical Instrument as a Spellcasting Focus',
        'Expertise', 'two of your skill proficiencies', 'Performance and Persuasion', 'level 9', 'two more',
        'half your Proficiency Bonus (round down)', 'skill proficiency you lack', 'doesn’t otherwise use your Proficiency Bonus',
        'College of Lore', 'Bard level or lower', 'levels 8, 12, and 16', 'for which you qualify',
        'Short or Long Rest', 'expend a spell slot (no action', 'regain one expended use',
        '30 feet', 'Charmed or Frightened', 'Reaction', 'rerolled', 'Advantage',
        'including this level', 'Prepared Spells number', 'increases', 'Bard, Cleric, Druid, and Wizard spell lists',
        'whenever you replace', 'count as Bard spells', 'roll Initiative', 'until you have two', 'fewer than that',
        'Boon of Spell Recall', 'Power Word Heal and Power Word Kill', 'always', 'second creature', 'within 10 feet of the first target',
    ]
    missing = [a for a in anchors if a not in flat]
    assert not missing, f'missing source anchors: {missing}'
    assert json.dumps(prose_facts(records), sort_keys=True) == json.dumps(PROSE_FACTS, sort_keys=True), 'explicit Bard prose metadata'
    print(f'SRD {PIN}: 13 identities/pages, 20 Bard rows, all nine slot columns and explicit prose metadata match pp31–33')


if __name__ == '__main__':
    if len(sys.argv) != 2:
        raise SystemExit('usage: verify_srd_bard.py /path/to/SRD_CC_v5.2.1.pdf')
    verify(Path(sys.argv[1]))
