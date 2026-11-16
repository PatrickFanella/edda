#!/usr/bin/env python3
"""Offline p35 semantic comparison for #309; requires only Python and pdftotext.

SOURCE_FACTS explicitly transcribes every metadata field from the pinned page.
The compact c:/m:/p: prefixes expand losslessly; no catalogue digest is used.
"""
import copy
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
PREFIXES = {"c:": "srd-5.2.1:character/", "m:": "srd-5.2.1:mechanic/",
            "p:": "srd-5.2.1:planning/"}

# Source-reviewed facts, not loaded or derived from the catalogue at runtime.
SOURCE_FACTS = [{'id': 'c:subclass/bard/college-of-lore',
  'kind': 'subclass',
  'name': 'College of Lore',
  'sources': [{'inventory_id': 'p:subclasses/bard-college-of-lore', 'page': 35}],
  'parent_id': 'c:class/bard',
  'prerequisite_refs': ['c:prerequisite/bard/college-of-lore/college-of-lore-level'],
  'required_mechanics': ['m:bard/subclasses/college-of-lore']},
 {'id': 'c:prerequisite/bard/college-of-lore/college-of-lore-level',
  'kind': 'prerequisite',
  'name': 'College of Lore level requirement',
  'sources': [{'inventory_id': 'p:subclasses/bard-college-of-lore', 'page': 35}],
  'parent_id': 'c:subclass/bard/college-of-lore',
  'required_mechanics': ['m:prerequisite/class-level'],
  'requirement': {'key': 'm:prerequisite/class-level',
                  'record_refs': ['c:class/bard'],
                  'parameters': {'minimum-level': 3}}},
 {'id': 'c:feature/bard/college-of-lore/bonus-proficiencies',
  'kind': 'feature',
  'name': 'Bonus Proficiencies',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-3-bonus-proficiencies', 'page': 35}],
  'parent_id': 'c:subclass/bard/college-of-lore',
  'choice_refs': ['c:choice/bard/college-of-lore/bonus-proficiencies'],
  'prerequisite_refs': ['c:prerequisite/bard/college-of-lore/bonus-proficiencies-level'],
  'required_mechanics': ['m:bard/college-of-lore/bonus-proficiencies/gain-proficiency-with-selected-skills']},
 {'id': 'c:prerequisite/bard/college-of-lore/bonus-proficiencies-level',
  'kind': 'prerequisite',
  'name': 'Bonus Proficiencies level requirement',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-3-bonus-proficiencies', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/bonus-proficiencies',
  'required_mechanics': ['m:prerequisite/class-level'],
  'requirement': {'key': 'm:prerequisite/class-level',
                  'record_refs': ['c:class/bard'],
                  'parameters': {'minimum-level': 3}}},
 {'id': 'c:feature/bard/college-of-lore/cutting-words',
  'kind': 'feature',
  'name': 'Cutting Words',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-3-cutting-words', 'page': 35}],
  'parent_id': 'c:subclass/bard/college-of-lore',
  'prerequisite_refs': ['c:prerequisite/bard/college-of-lore/cutting-words-level',
                        'c:prerequisite/bard/college-of-lore/cutting-words-inspiration'],
  'required_mechanics': ['m:bard/college-of-lore/cutting-words/target/creature-you-can-see',
                         'm:bard/college-of-lore/cutting-words/trigger/damage-roll-or-successful-ability-check-or-attack-roll',
                         'm:bard/college-of-lore/cutting-words/timing/reaction',
                         'm:bard/college-of-lore/cutting-words/expend-bardic-inspiration',
                         'm:bard/college-of-lore/cutting-words/roll-canonical-bardic-inspiration-die',
                         'm:bard/college-of-lore/cutting-words/subtract-from-creature-roll',
                         'm:bard/college-of-lore/cutting-words/reduce-damage-or-potentially-turn-success-into-failure',
                         'm:bard/college-of-lore/cutting-words/range-feet',
                         'm:bard/college-of-lore/cutting-words/uses-expended'],
  'parameters': {'srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/range-feet': 60,
                 'srd-5.2.1:mechanic/bard/college-of-lore/cutting-words/uses-expended': 1}},
 {'id': 'c:prerequisite/bard/college-of-lore/cutting-words-level',
  'kind': 'prerequisite',
  'name': 'Cutting Words level requirement',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-3-cutting-words', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/cutting-words',
  'required_mechanics': ['m:prerequisite/class-level'],
  'requirement': {'key': 'm:prerequisite/class-level',
                  'record_refs': ['c:class/bard'],
                  'parameters': {'minimum-level': 3}}},
 {'id': 'c:prerequisite/bard/college-of-lore/cutting-words-inspiration',
  'kind': 'prerequisite',
  'name': 'Cutting Words Inspiration requirement',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-3-cutting-words', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/cutting-words',
  'required_mechanics': ['m:bard/college-of-lore/cutting-words/requires-bardic-inspiration-use'],
  'requirement': {'key': 'm:bard/college-of-lore/cutting-words/requires-bardic-inspiration-use',
                  'record_refs': ['c:resource/bard/bardic-inspiration'],
                  'parameters': {'uses-required': 1}}},
 {'id': 'c:feature/bard/college-of-lore/magical-discoveries',
  'kind': 'feature',
  'name': 'Magical Discoveries',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-6-magical-discoveries', 'page': 35}],
  'parent_id': 'c:subclass/bard/college-of-lore',
  'choice_refs': ['c:choice/bard/college-of-lore/magical-discoveries'],
  'prerequisite_refs': ['c:prerequisite/bard/college-of-lore/magical-discoveries-level'],
  'required_mechanics': ['m:bard/college-of-lore/magical-discoveries/learn-chosen-spells',
                         'm:bard/college-of-lore/magical-discoveries/always-have-chosen-spells-prepared',
                         'm:bard/college-of-lore/magical-discoveries/replace-one-on-gaining-bard-level',
                         'm:bard/college-of-lore/magical-discoveries/replacement-meets-same-list-and-cantrip-or-slot-requirements',
                         'm:bard/college-of-lore/magical-discoveries/lists/any-combination',
                         'm:bard/college-of-lore/magical-discoveries/spells-learned',
                         'm:bard/college-of-lore/magical-discoveries/replacements-per-bard-level'],
  'parameters': {'srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/spells-learned': 2,
                 'srd-5.2.1:mechanic/bard/college-of-lore/magical-discoveries/replacements-per-bard-level': 1}},
 {'id': 'c:prerequisite/bard/college-of-lore/magical-discoveries-level',
  'kind': 'prerequisite',
  'name': 'Magical Discoveries level requirement',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-6-magical-discoveries', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/magical-discoveries',
  'required_mechanics': ['m:prerequisite/class-level'],
  'requirement': {'key': 'm:prerequisite/class-level',
                  'record_refs': ['c:class/bard'],
                  'parameters': {'minimum-level': 6}}},
 {'id': 'c:feature/bard/college-of-lore/peerless-skill',
  'kind': 'feature',
  'name': 'Peerless Skill',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-14-peerless-skill', 'page': 35}],
  'parent_id': 'c:subclass/bard/college-of-lore',
  'prerequisite_refs': ['c:prerequisite/bard/college-of-lore/peerless-skill-level',
                        'c:prerequisite/bard/college-of-lore/peerless-skill-inspiration'],
  'required_mechanics': ['m:bard/college-of-lore/peerless-skill/trigger/own-failed-ability-check-or-attack-roll',
                         'm:bard/college-of-lore/peerless-skill/expend-bardic-inspiration',
                         'm:bard/college-of-lore/peerless-skill/roll-canonical-bardic-inspiration-die',
                         'm:bard/college-of-lore/peerless-skill/add-number-to-d20',
                         'm:bard/college-of-lore/peerless-skill/potentially-turn-failure-into-success',
                         'm:bard/college-of-lore/peerless-skill/on-failure-inspiration-not-expended',
                         'm:bard/college-of-lore/peerless-skill/uses-expended-on-success'],
  'parameters': {'srd-5.2.1:mechanic/bard/college-of-lore/peerless-skill/uses-expended-on-success': 1}},
 {'id': 'c:prerequisite/bard/college-of-lore/peerless-skill-level',
  'kind': 'prerequisite',
  'name': 'Peerless Skill level requirement',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-14-peerless-skill', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/peerless-skill',
  'required_mechanics': ['m:prerequisite/class-level'],
  'requirement': {'key': 'm:prerequisite/class-level',
                  'record_refs': ['c:class/bard'],
                  'parameters': {'minimum-level': 14}}},
 {'id': 'c:prerequisite/bard/college-of-lore/peerless-skill-inspiration',
  'kind': 'prerequisite',
  'name': 'Peerless Skill Inspiration requirement',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-14-peerless-skill', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/peerless-skill',
  'required_mechanics': ['m:bard/college-of-lore/peerless-skill/requires-bardic-inspiration-use'],
  'requirement': {'key': 'm:bard/college-of-lore/peerless-skill/requires-bardic-inspiration-use',
                  'record_refs': ['c:resource/bard/bardic-inspiration'],
                  'parameters': {'uses-required': 1}}},
 {'id': 'c:choice/bard/college-of-lore/bonus-proficiencies',
  'kind': 'choice',
  'name': 'Bonus Proficiencies choice',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-3-bonus-proficiencies', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/bonus-proficiencies',
  'required_mechanics': ['m:bard/college-of-lore/bonus-proficiencies/any-skills'],
  'select': 3,
  'option_keys': ['m:skills/acrobatics',
                  'm:skills/animal-handling',
                  'm:skills/arcana',
                  'm:skills/athletics',
                  'm:skills/deception',
                  'm:skills/history',
                  'm:skills/insight',
                  'm:skills/intimidation',
                  'm:skills/investigation',
                  'm:skills/medicine',
                  'm:skills/nature',
                  'm:skills/perception',
                  'm:skills/performance',
                  'm:skills/persuasion',
                  'm:skills/religion',
                  'm:skills/sleight-of-hand',
                  'm:skills/stealth',
                  'm:skills/survival']},
 {'id': 'c:choice/bard/college-of-lore/magical-discoveries',
  'kind': 'choice',
  'name': 'Magical Discoveries choice',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-6-magical-discoveries', 'page': 35}],
  'parent_id': 'c:feature/bard/college-of-lore/magical-discoveries',
  'prerequisite_refs': ['c:prerequisite/bard/college-of-lore/magical-discoveries-spell-eligibility'],
  'required_mechanics': ['m:bard/college-of-lore/magical-discoveries/choose-spells-from-listed-lists-in-any-combination'],
  'select': 2,
  'option_keys': ['m:cleric/spell-list/cantrip-or-slot-eligible',
                  'm:druid/spell-list/cantrip-or-slot-eligible',
                  'm:wizard/spell-list/cantrip-or-slot-eligible']},
 {'id': 'c:prerequisite/bard/college-of-lore/magical-discoveries-spell-eligibility',
  'kind': 'prerequisite',
  'name': 'Magical Discoveries spell eligibility',
  'sources': [{'inventory_id': 'p:subclass-features/bard-level-6-magical-discoveries', 'page': 35}],
  'parent_id': 'c:choice/bard/college-of-lore/magical-discoveries',
  'required_mechanics': ['m:bard/college-of-lore/magical-discoveries/cantrip-or-level-with-slots-in-bard-features-table'],
  'requirement': {'key': 'm:bard/college-of-lore/magical-discoveries/cantrip-or-level-with-slots-in-bard-features-table',
                  'record_refs': ['c:class/bard',
                                  'c:resource/bard/spell-slot-1',
                                  'c:resource/bard/spell-slot-2',
                                  'c:resource/bard/spell-slot-3',
                                  'c:resource/bard/spell-slot-4',
                                  'c:resource/bard/spell-slot-5',
                                  'c:resource/bard/spell-slot-6',
                                  'c:resource/bard/spell-slot-7',
                                  'c:resource/bard/spell-slot-8',
                                  'c:resource/bard/spell-slot-9'],
                  'parameters': {}}}]


def expand(value):
    if isinstance(value, str):
        for short, prefix in PREFIXES.items():
            if value.startswith(short):
                return prefix + value[len(short):]
        return value
    if isinstance(value, list):
        return [expand(v) for v in value]
    if isinstance(value, dict):
        return {expand(k): expand(v) for k, v in value.items()}
    return value


def expected_records():
    result = expand(SOURCE_FACTS)
    for r in result:
        for k in ("choice_refs", "resource_refs", "prerequisite_refs", "option_keys"):
            r.setdefault(k, [])
        for k in ("level", "select", "capacity"):
            r.setdefault(k, 0)
    return result


def compare(records):
    expected = expected_records()
    assert len(records) == len(expected), "p35 record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"p35 semantic mismatch: {want['id']}\nactual={actual}\nexpected={want}"


def contrary_controls(records):
    # These variants are also exercised through Go's structural validator.
    controls = []
    for i, r in enumerate(records):
        for j, key in enumerate(r["required_mechanics"]):
            c = copy.deepcopy(records)
            replacement = "srd-5.2.1:mechanic/lore/contrary"
            c[i]["required_mechanics"][j] = replacement
            if c[i].get("requirement", {}).get("key") == key:
                c[i]["requirement"]["key"] = replacement
            if key in c[i].get("parameters", {}):
                c[i]["parameters"][replacement] = c[i]["parameters"].pop(key)
            controls.append(c)
        for field in ("parameters",):
            for key in r.get(field, {}):
                c = copy.deepcopy(records); c[i][field][key] += 1; controls.append(c)
        if "requirement" in r:
            for key in r["requirement"]["parameters"]:
                c = copy.deepcopy(records); c[i]["requirement"]["parameters"][key] += 1; controls.append(c)
            for j in range(len(r["requirement"]["record_refs"])):
                c = copy.deepcopy(records)
                c[i]["requirement"]["record_refs"][j] = PREFIXES["c:"] + "feature/bard/spellcasting"
                controls.append(c)
        if r["kind"] == "feature":
            c = copy.deepcopy(records)
            other = "cutting-words" if r["name"] == "Peerless Skill" else "peerless-skill"
            level = 3 if other == "cutting-words" else 14
            old = c[i]["sources"][0]["inventory_id"]
            for record in c:
                if record["sources"][0]["inventory_id"] == old:
                    record["sources"][0]["inventory_id"] = PREFIXES["p:"] + f"subclass-features/bard-level-{level}-{other}"
            controls.append(c)
        if r["kind"] == "choice":
            c = copy.deepcopy(records); c[i]["select"] += 1; controls.append(c)
            c = copy.deepcopy(records); c[i]["option_keys"].pop(); controls.append(c)
            c = copy.deepcopy(records); c[i]["option_keys"].append(PREFIXES["m:"] + "bard/spell-list"); controls.append(c)
            c = copy.deepcopy(records); c[i]["values_by_level"] = [{"level": 6, "value": r["select"]}, {"level": 14, "value": 4}]; controls.append(c)
    for c in controls:
        try:
            compare(c)
        except AssertionError:
            continue
        raise AssertionError("contrary source metadata accepted")
    return len(controls)


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    columns = [subprocess.check_output(["pdftotext", "-f", "35", "-l", "35", "-x", str(x),
               "-y", "0", "-W", "310", "-H", "792", "-layout", str(pdf), "-"], text=True)
               for x in (0, 310)]
    text = " ".join(re.sub(r"-\n\s*", "", " ".join(columns).replace("\u00ad", "")).split())
    anchors = ["College of Lore", "Level 3: Bonus Proficiencies", "three skills of your choice",
               "Level 3: Cutting Words", "creature that you can see", "within 60 feet",
               "makes a damage roll or succeeds on an ability check or attack roll",
               "take a Reaction", "expend one use", "Bardic Inspiration die",
               "subtract the number rolled from the creature’s roll",
               "reducing the damage or potentially turning the success into a failure",
               "Level 6: Magical Discoveries", "two spells of your choice",
               "Cleric, Druid, or Wizard spell list", "any combination thereof",
               "must be a cantrip or a spell for which you have spell slots",
               "as shown in the Bard Features table", "always have the chosen spells prepared",
               "whenever you gain a Bard level", "replace one", "meets these requirements",
               "Level 14: Peerless Skill", "ability check or attack roll and fail",
               "add the number rolled to the d20", "On a failure", "isn’t expended"]
    assert all(a in text for a in anchors), [a for a in anchors if a not in text]
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-college-of-lore.json").read_text())
    assert c["schema_version"] == 1 and c["source"]["sha256"] == PIN
    assert c["source"]["license"] == "CC-BY-4.0" and c["source"]["attribution"] and c["source"]["changes"]
    compare(c["records"])
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r["batch_key"] == "subclass-bard"}
    identities = [r for r in c["records"] if r["kind"] in ("subclass", "feature")]
    assert len(identities) == len(expected) == 5
    assert {s["inventory_id"]: s["page"] for r in identities for s in r["sources"]} == expected
    # Reused resource: canonical ownership, die progression, rest and slot/initiative
    # recovery are checked by the existing complete Bard comparator, not copied.
    from verify_srd_bard import verify as verify_bard
    verify_bard(pdf)
    n = contrary_controls(c["records"])
    print(f"p35: five identities, all 15 records and semantic fields match; {n} contrary controls rejected")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_college_of_lore.py /path/to/SRD_CC_v5.2.1.pdf")
    verify(Path(sys.argv[1]))
