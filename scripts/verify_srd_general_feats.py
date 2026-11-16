#!/usr/bin/env python3
"""Compare #300's exact identities, alternatives and benefits to pinned PDF p87."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
KEY = "srd-5.2.1:mechanic/"


def column(pdf, x):
    result = subprocess.run(
        ["pdftotext", "-f", "87", "-l", "87", "-x", str(x), "-y", "0",
         "-W", "310", "-H", "792", "-layout", str(pdf), "-"],
        capture_output=True, text=True, check=True,
    )
    return " ".join(re.sub(r"-\s+", "", result.stdout).split())


def expected_records():
    """Transcribe the complete graph from the source text checked below."""
    cid = "srd-5.2.1:character/"

    def record(kind, slug, name, parent, mechanics, **fields):
        return {
            "id": cid + kind + "/" + slug, "kind": kind, "name": name,
            "sources": [{"inventory_id": "srd-5.2.1:planning/feats/" + slug.split("/")[0], "page": 87}],
            "parent_id": parent, "level": 0, "choice_refs": [],
            "resource_refs": [], "prerequisite_refs": [],
            "required_mechanics": [KEY + k for k in mechanics],
            "select": 0, "option_keys": [], "capacity": 0, **fields,
        }

    def predicate(slug, name, parent, key, refs=(), parameters=None):
        return record("prerequisite", slug, name, parent, ["prerequisite/" + key],
                      requirement={"key": KEY + "prerequisite/" + key,
                                   "record_refs": list(refs),
                                   "parameters": {} if parameters is None else parameters})

    asi = "ability-score-improvement"
    base = "feats/" + asi + "/"
    feat_id = cid + "feat/" + asi
    pattern_id = cid + "choice/" + asi + "/increase-pattern"
    result = [
        record("feat", asi, "Ability Score Improvement", "",
               ["feats/category/general", "feats/repeatability/unrestricted", base + "maximum-ability-score"],
               choice_refs=[pattern_id], prerequisite_refs=[cid + "prerequisite/" + asi + "/level"],
               parameters={KEY + base + "maximum-ability-score": 20}),
        predicate(asi + "/level", "Level", feat_id, "minimum-character-level", parameters={"minimum-level": 4}),
        record("choice", asi + "/increase-pattern", "Ability increase pattern", feat_id,
               [base + "choose-one-increase-pattern", base + "increase-per-score"], select=1,
               choice_refs=[pattern_id + "/one-score", pattern_id + "/two-scores"],
               option_keys=[KEY + base + "one-score-plus-two", KEY + base + "two-scores-plus-one"],
               options=[
                   {"key": KEY + base + "one-score-plus-two", "quantities": {KEY + base + "increase-per-score": 2}},
                   {"key": KEY + base + "two-scores-plus-one", "quantities": {KEY + base + "increase-per-score": 1}},
               ]),
    ]
    for part, name, count, option in (
        ("one-score", "One ability score", 1, "one-score-plus-two"),
        ("two-scores", "Two distinct ability scores", 2, "two-scores-plus-one"),
    ):
        slug = asi + "/increase-pattern/" + part
        result.append(record(
            "choice", slug, name, pattern_id,
            [base + "increase-selected-scores", base + "subject-to-maximum-ability-score"],
            select=count, option_keys=[KEY + "abilities/" + ability for ability in
                                      ("strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma")],
            prerequisite_refs=[cid + "prerequisite/" + slug + "/selected-pattern"],
        ))
        result.append(predicate(slug + "/selected-pattern", "Selected pattern", cid + "choice/" + slug,
                                asi + "/" + option + "-selected", refs=[pattern_id]))
    feat_id = cid + "feat/grappler"
    result.extend([
        record("feat", "grappler", "Grappler", "", [
            "feats/category/general", "feats/maximum-times-taken",
            "feats/grappler/ability-score-increase", "feats/grappler/maximum-ability-score",
            "feats/grappler/punch-and-grab/optional-on-hit-creature", "feats/grappler/punch-and-grab/unarmed-strike",
            "feats/grappler/punch-and-grab/part-of-attack-action-on-your-turn",
            "feats/grappler/punch-and-grab/use-both-damage-and-grapple-options", "feats/grappler/punch-and-grab/uses-per-turn",
            "feats/grappler/attack-advantage/on-attack-rolls-against-creature-grappled-by-you",
            "feats/grappler/fast-wrestler/no-extra-movement-to-move-creature-grappled-by-you",
            "feats/grappler/fast-wrestler/only-if-creature-your-size-or-smaller",
        ], choice_refs=[cid + "choice/grappler/ability-score"],
            prerequisite_refs=[cid + "prerequisite/grappler/level", cid + "prerequisite/grappler/ability"],
            parameters={KEY + "feats/maximum-times-taken": 1, KEY + "feats/grappler/ability-score-increase": 1,
                        KEY + "feats/grappler/maximum-ability-score": 20, KEY + "feats/grappler/punch-and-grab/uses-per-turn": 1}),
        predicate("grappler/level", "Level", feat_id, "minimum-character-level", parameters={"minimum-level": 4}),
        predicate("grappler/ability", "Ability", feat_id, "strength-or-dexterity-at-least", parameters={"minimum-score": 13}),
        record("choice", "grappler/ability-score", "Ability Score Increase", feat_id,
               ["feats/grappler/increase-one-selected-ability", "feats/grappler/subject-to-maximum-ability-score"],
               select=1, option_keys=[KEY + "abilities/strength", KEY + "abilities/dexterity"]),
    ])
    return result


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_general_feats.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    assert "A feat can be taken only once unless its description states otherwise" in column(pdf, 0)
    text = column(pdf, 310).split("General Feats ", 1)[1].split(" Fighting Style Feats", 1)[0]
    assert text == (
        "Ability Score Improvement General Feat (Prerequisite: Level 4+) "
        "Increase one ability score of your choice by 2, or increase two ability scores of your choice by 1. "
        "This feat can’t increase an ability score above 20. Repeatable. You can take this feat more than once. "
        "Grappler General Feat (Prerequisite: Level 4+, Strength or Dexterity 13+) "
        "You gain the following benefits. Ability Score Increase. Increase your Strength or Dexterity score by 1, to a maximum of 20. "
        "Punch and Grab. When you hit a creature with an Unarmed Strike as part of the Attack action on your turn, "
        "you can use both the Damage and the Grapple option. You can use this benefit only once per turn. "
        "Attack Advantage. You have Advantage on attack rolls against a creature Grappled by you. "
        "Fast Wrestler. You don’t have to spend extra movement to move a creature Grappled by you if the creature is your size or smaller."
    ), ("source text changed", text)
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-general-feats.json").read_text())
    assert c["schema_version"] == 1
    source = c["source"]
    assert source["sha256"] == PIN and source["license"] == "CC-BY-4.0"
    assert source["attribution"] and source["changes"]
    records = c["records"]
    assert len(records) == len({r["id"] for r in records}) == 11
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r.get("catalogue_issue") == 300}
    actual = {s["inventory_id"]: s["page"] for r in records if r["kind"] == "feat" for s in r["sources"]}
    assert len(expected) == 2 and actual == expected
    for r in records:
        assert r["sources"] == [{"inventory_id": "srd-5.2.1:planning/feats/" + r["id"].split("/")[2], "page": 87}]
        assert r["kind"] in ("feat", "choice", "prerequisite")
    # Canonical JSON distinguishes explicit integers from Boolean/float aliases
    # while ignoring object property order. Extra fields and scaling also fail.
    assert json.dumps(records, sort_keys=True) == json.dumps(expected_records(), sort_keys=True), "complete page 87 source graph"
    print("SRD " + PIN + ": exact two General feat identities on PDF p87; complete source text, prerequisites, alternatives, nested choices, caps and conditional benefits agree")


if __name__ == "__main__":
    main()
