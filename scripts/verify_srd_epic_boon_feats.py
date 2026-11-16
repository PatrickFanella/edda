#!/usr/bin/env python3
"""Compare #298's complete metadata graph with pinned SRD PDF pp87–88."""
import copy
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
KEY = "srd-5.2.1:mechanic/"
CID = "srd-5.2.1:character/"


def expected_records():
    records = []

    def add(kind, slug, name, parent, mechanics, **fields):
        r = dict(id=CID + kind + "/" + slug, kind=kind, name=name,
                 sources=[dict(inventory_id="srd-5.2.1:planning/feats/" + slug.split("/")[0], page=88)],
                 parent_id=parent, level=0, choice_refs=[], resource_refs=[], prerequisite_refs=[],
                 required_mechanics=[KEY + k for k in mechanics], select=0, option_keys=[], capacity=0)
        r.update(fields)
        records.append(r)
        return r

    def pred(owner, part, key, parameters=None, refs=()):
        slug = owner["id"].split("/", 2)[2] + "/" + part
        r = add("prerequisite", slug, part.replace("-", " ").capitalize(), owner["id"],
                ["prerequisite/" + key], requirement=dict(key=KEY + "prerequisite/" + key,
                record_refs=list(refs), parameters={} if parameters is None else parameters))
        owner["prerequisite_refs"].append(r["id"])
        return r

    for slug, name, abilities, benefits, numbers in (
        ("boon-of-combat-prowess", "Boon of Combat Prowess", None,
         ["peerless-aim/optional-replace-miss-with-hit", "peerless-aim/recover-at-start-of-your-next-turn"], {}),
        ("boon-of-dimensional-travel", "Boon of Dimensional Travel", None,
         ["blink-steps/optional-teleport-immediately-after-attack-or-magic-action",
          "blink-steps/destination-unoccupied-space-you-can-see", "blink-steps/maximum-distance-feet"],
         {"blink-steps/maximum-distance-feet": 30}),
        ("boon-of-fate", "Boon of Fate", None,
         ["improve-fate/optional-roll-and-apply-total-as-bonus-or-penalty-to-triggering-d20-roll",
          "improve-fate/dice-count", "improve-fate/die-sides", "improve-fate/other-creature-range-feet",
          "improve-fate/recover-on-roll-initiative-or-finish-short-or-long-rest"],
         {"improve-fate/dice-count": 2, "improve-fate/die-sides": 4, "improve-fate/other-creature-range-feet": 60}),
        ("boon-of-irresistible-offense", "Boon of Irresistible Offense", ["strength", "dexterity"],
         ["overcome-defenses/your-bludgeoning-piercing-slashing-damage-always-ignores-resistance",
          "overwhelming-strike/optional-extra-damage-to-attack-target",
          "overwhelming-strike/extra-damage-equals-selected-ability-score-not-modifier",
          "overwhelming-strike/extra-damage-type-same-as-attack", "overwhelming-strike/attack-d20-result"],
         {"overwhelming-strike/attack-d20-result": 20}),
        ("boon-of-spell-recall", "Boon of Spell Recall", ["intelligence", "wisdom", "charisma"],
         ["free-casting/roll-on-every-spell-cast-with-level-one-to-four-slot",
          "free-casting/dice-count", "free-casting/die-sides",
          "free-casting/do-not-expend-slot-only-if-roll-equals-slot-level"],
         {"free-casting/dice-count": 1, "free-casting/die-sides": 4}),
        ("boon-of-the-night-spirit", "Boon of the Night Spirit", None,
         ["merge-with-shadows/optional-give-self-invisible-condition-as-bonus-action-in-dim-light-or-darkness",
          "merge-with-shadows/ends-immediately-after-you-take-action-or-bonus-action-or-reaction",
          "shadowy-form/resistance-to-all-damage-except-psychic-and-radiant-in-dim-light-or-darkness"], {}),
        ("boon-of-truesight", "Boon of Truesight", None,
         ["truesight/have-truesight", "truesight/range-feet"], {"truesight/range-feet": 60}),
    ):
        base = "feats/" + slug + "/"
        parameters = {KEY + "feats/maximum-times-taken": 1,
                      KEY + base + "ability-score-increase": 1, KEY + base + "maximum-ability-score": 30}
        parameters.update({KEY + base + k: v for k, v in numbers.items()})
        f = add("feat", slug, name, "", ["feats/category/epic-boon", "feats/maximum-times-taken",
                base + "ability-score-increase", base + "maximum-ability-score"] +
                [base + k for k in benefits], parameters=parameters)
        pred(f, "level", "minimum-character-level", {"minimum-level": 19})
        if abilities is None:
            abilities = ["strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"]
        ch = add("choice", slug + "/ability-score", "Ability Score Increase", f["id"],
                 [base + "increase-one-selected-ability", base + "subject-to-maximum-ability-score"],
                 select=1, option_keys=[KEY + "abilities/" + a for a in abilities])
        f["choice_refs"].append(ch["id"])
        if slug == "boon-of-combat-prowess":
            pred(f, "miss", "attack-roll-missed")
        elif slug == "boon-of-dimensional-travel":
            pred(f, "action", "immediately-after-you-take-attack-or-magic-action")
            pred(f, "destination", "teleport-destination-unoccupied-and-visible-to-you")
        elif slug == "boon-of-fate":
            pred(f, "d20-test", "you-or-another-creature-in-range-succeeds-or-fails-d20-test", {"other-creature-range-feet": 60})
        elif slug == "boon-of-irresistible-offense":
            pred(f, "attack-roll", "attack-roll-d20-equals", {"d20-result": 20})
            pred(f, "damage-amount", "extra-damage-equals-ability-score-increased-by-this-feat", refs=[ch["id"]])
        elif slug == "boon-of-spell-recall":
            pred(f, "spellcasting", "has-spellcasting-feature")
            cast = pred(f, "slot", "cast-spell-with-slot-level-in-range", {"minimum-slot-level": 1, "maximum-slot-level": 4})
            pred(f, "roll-match", "die-result-equals-triggering-spell-slot-level", refs=[cast["id"]])
        elif slug == "boon-of-the-night-spirit":
            pred(f, "light", "within-dim-light-or-darkness")
            pred(f, "invisibility-end", "immediately-after-you-take-action-or-bonus-action-or-reaction")
        if slug in ("boon-of-combat-prowess", "boon-of-fate"):
            benefit = "peerless-aim" if slug == "boon-of-combat-prowess" else "improve-fate"
            recovery = "start-of-your-next-turn" if benefit == "peerless-aim" else "roll-initiative-or-finish-short-or-long-rest"
            resource = add("resource", slug + "/" + benefit, benefit.replace("-", " ").title(), f["id"],
                           [base + benefit + "/uses", base + benefit + "/recover-on-" + recovery], capacity=1)
            f["resource_refs"].append(resource["id"])
    return records


def column(pdf, page, x):
    result = subprocess.run(["pdftotext", "-f", str(page), "-l", str(page), "-x", str(x),
                             "-y", "0", "-W", "310", "-H", "792", "-layout", str(pdf), "-"],
                            capture_output=True, text=True, check=True)
    return " ".join(re.sub(r"-\s+", "", result.stdout).split()).replace(" 88 System Reference Document 5.2.1", "")


def verify_graph(c):
    assert c["schema_version"] == 1
    source = c["source"]
    assert source["sha256"] == PIN and source["license"] == "CC-BY-4.0"
    assert source["attribution"] and source["changes"]
    records = c["records"]
    assert len(records) == len({r["id"] for r in records})
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r.get("catalogue_issue") == 298}
    actual = {s["inventory_id"]: s["page"] for r in records if r["kind"] == "feat" for s in r["sources"]}
    assert len(expected) == 7 and actual == expected
    assert json.dumps(records, sort_keys=True) == json.dumps(expected_records(), sort_keys=True), "complete page 88 source graph"


def contrary_source_fixtures(catalogue):
    """Exercise the comparator itself with contrary but plausible source graphs."""
    fixtures = []
    for i, record in enumerate(catalogue["records"]):
        def change(field, value):
            if record.get(field) == value:
                return
            c = copy.deepcopy(catalogue)
            c["records"][i][field] = value
            fixtures.append(c)
        change("required_mechanics", record["required_mechanics"] + [KEY + "invented/scaling"])
        change("parent_id", CID + "feat/boon-of-truesight")
        for j in range(len(record["required_mechanics"])):
            mechanics = record["required_mechanics"].copy()
            mechanics[j] = KEY + "contrary/mechanic"
            change("required_mechanics", mechanics)
        for key in record.get("parameters", {}):
            params = record["parameters"].copy()
            params[key] += 1
            change("parameters", params)
        if record["kind"] == "choice":
            change("option_keys", record["option_keys"][:1])
            change("select", 2)
            change("values_by_level", [{"level": 19, "value": 1}, {"level": 20, "value": 2}])
        if record["kind"] == "resource":
            change("capacity", 2)
        if "requirement" in record:
            req = copy.deepcopy(record["requirement"])
            req["record_refs"] = [CID + "feat/boon-of-truesight"]
            change("requirement", req)
            for key in record["requirement"]["parameters"]:
                req = copy.deepcopy(record["requirement"])
                req["parameters"][key] += 1
                change("requirement", req)
            req = copy.deepcopy(record["requirement"])
            req["parameters"]["invented-scaling"] = 2
            change("requirement", req)
            if record["requirement"]["record_refs"]:
                req = copy.deepcopy(record["requirement"])
                req["record_refs"] = []
                change("requirement", req)
    for fixture in fixtures:
        try:
            verify_graph(fixture)
        except AssertionError:
            continue
        raise AssertionError("comparator accepted contrary-source fixture")
    print(f"Rejected {len(fixtures)} contrary-source comparator fixtures")


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_epic_boon_feats.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    assert "A feat can be taken only once unless its description states otherwise" in column(pdf, 87, 0)
    text = column(pdf, 88, 0).split("Epic Boon Feats ", 1)[1] + " " + column(pdf, 88, 310)
    # Exact source text, including Fate's continuation at the top of the right column.
    assert text == SOURCE_TEXT, ("source text changed", text)
    catalogue = json.loads((ROOT / "internal/rules/srd521/testdata/character-epic-boon-feats.json").read_text())
    verify_graph(catalogue)
    contrary_source_fixtures(catalogue)
    print("SRD " + PIN + ": exact seven Epic Boon identities, complete PDF pp87–88 source and complete metadata graph agree")


SOURCE_TEXT = (
    'Boon of Combat Prowess Epic Boon Feat (Prerequisite: Level 19+) You gain the following benefits. '
    'Ability Score Increase. Increase one ability score of your choice by 1, to a maximum of 30. Peerless '
    'Aim. When you miss with an attack roll, you can hit instead. Once you use this benefit, you can’t '
    'use it again until the start of your next turn. Boon of Dimensional Travel Epic Boon Feat '
    '(Prerequisite: Level 19+) You gain the following benefits. Ability Score Increase. Increase one '
    'ability score of your choice by 1, to a maximum of 30. Blink Steps. Immediately after you take the '
    'Attack action or the Magic action, you can teleport up to 30 feet to an unoccupied space you can '
    'see. Boon of Fate Epic Boon Feat (Prerequisite: Level 19+) You gain the following benefits. Ability '
    'Score Increase. Increase one ability score of your choice by 1, to a maximum of 30. Improve Fate. '
    'When you or another creature within 60 feet of you succeeds on or fails a D20 Test, you can roll 2d4 '
    'and apply the total rolled as a bonus or penalty to the d20 roll. Once you use this benefit, you '
    'can’t use it again until you roll Initiative or finish a Short or Long Rest. Boon of Irresistible '
    'Offense Epic Boon Feat (Prerequisite: Level 19+) You gain the following benefits. Ability Score '
    'Increase. Increase your Strength or Dexterity score by 1, to a maximum of 30. Overcome Defenses. The '
    'Bludgeoning, Piercing, and Slashing damage you deal always ignores Resistance. Overwhelming Strike. '
    'When you roll a 20 on the d20 for an attack roll, you can deal extra damage to the target equal to '
    'the ability score increased by this feat. The extra damage’s type is the same as the attack’s type. '
    'Boon of Spell Recall Epic Boon Feat (Prerequisite: Level 19+, Spellcasting Feature) You gain the '
    'following benefits. Ability Score Increase. Increase your Intelligence, Wisdom, or Charisma score by '
    '1, to a maximum of 30. Free Casting. Whenever you cast a spell with a level 1–4 spell slot, roll '
    '1d4. If the number you roll is the same as the slot’s level, the slot isn’t expended. Boon of the '
    'Night Spirit Epic Boon Feat (Prerequisite: Level 19+) You gain the following benefits. Ability Score '
    'Increase. Increase one ability score of your choice by 1, to a maximum of 30. Merge with Shadows. '
    'While within Dim Light or Darkness, you can give yourself the Invisible condition as a Bonus Action. '
    'The condition ends on you immediately after you take an action, a Bonus Action, or a Reaction. '
    'Shadowy Form. While within Dim Light or Darkness, you have Resistance to all damage except Psychic '
    'and Radiant. Boon of Truesight Epic Boon Feat (Prerequisite: Level 19+) You gain the following '
    'benefits. Ability Score Increase. Increase one ability score of your choice by 1, to a maximum of '
    '30. Truesight. You have Truesight with a range of 60 feet.'
)

if __name__ == "__main__":
    main()
