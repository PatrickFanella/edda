#!/usr/bin/env python3
"""Compare #301 Origin feat metadata with pinned SRD 5.2.1 PDF p87, offline."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
ID = "srd-5.2.1:character/"
KEY = "srd-5.2.1:mechanic/"


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_origin_feats.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    columns = []
    for x in (0, 310):
        text = subprocess.run(
            ["pdftotext", "-f", "87", "-l", "87", "-x", str(x), "-y", "0",
             "-W", "310", "-H", "792", "-layout", str(pdf), "-"],
            capture_output=True, text=True, check=True,
        ).stdout
        columns.append(" ".join(re.sub(r"-\s+", "", text).split()))
    left, right = columns
    origin = left.split("Origin Feats Alert ", 1)[1] + " " + right.split(" General Feats", 1)[0]
    assert "A feat can be taken only once unless" in left
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-origin-feats.json").read_text())
    assert c["source"]["sha256"] == PIN and c["source"]["license"] == "CC-BY-4.0"
    assert c["source"]["attribution"] and c["source"]["changes"]
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]) == 21
    roots = [r for r in c["records"] if r["kind"] == "feat"]
    inv = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inv["records"] if r.get("catalogue_issue") == 301}
    actual = {s["inventory_id"]: s["page"] for r in roots for s in r["sources"]}
    assert len(expected) == 4 and actual == expected
    for r in c["records"]:
        assert r["kind"] in ("feat", "choice", "prerequisite", "resource")
        slug = r["id"].removeprefix(ID).split("/")[1]
        assert r["sources"] == [{"inventory_id": "srd-5.2.1:planning/feats/" + slug, "page": 87}]
    for r, name, repeat in zip(roots, ["Alert", "Magic Initiate", "Savage Attacker", "Skilled"],
                               ["once", "different-spell-list", "once", "unrestricted"], strict=True):
        assert r["name"] == name and name + " Origin Feat" in ("Alert " + origin)
        assert KEY + "feats/category/origin" in r["required_mechanics"]
        assert KEY + "feats/repeatability/" + repeat in r["required_mechanics"]

    def rec(kind, slug):
        return records[ID + kind + "/" + slug]

    # Every benefit and conditional predicate has a source anchor. Go fixtures
    # independently compare every graph operand, key, count and parameter.
    anchors = {
        "When you roll Initiative, you can add your Proficiency Bonus to the roll.": ("feat", "alert"),
        "Immediately after you roll Initiative, you can swap your Initiative with the Initiative": ("choice", "alert/initiative-swap"),
        "of one willing ally in the same combat.": ("prerequisite", "alert/initiative-swap/same-combat"),
        "You can’t make this swap if you or the ally has the Incapacitated condition.": ("prerequisite", "alert/initiative-swap/self-not-incapacitated"),
        "Two Cantrips. You learn two cantrips of your choice from the Cleric, Druid, or Wizard spell list.": ("choice", "magic-initiate/cantrips"),
        "Intelligence, Wisdom, or Charisma is your spellcasting ability for this feat’s spells (choose when you select this feat).": ("choice", "magic-initiate/spellcasting-ability"),
        "Choose a level 1 spell from the same list you selected for this feat’s cantrips.": ("prerequisite", "magic-initiate/level-one-spell/selected-list"),
        "You always have that spell prepared.": ("feat", "magic-initiate"),
        "You can cast it once without a spell slot": ("resource", "magic-initiate/free-cast"),
        "when you finish a Long Rest.": ("resource", "magic-initiate/free-cast"),
        "You can also cast the spell using any spell slots you have.": ("feat", "magic-initiate"),
        "Whenever you gain a new level, you can replace one of the spells you chose for this feat": ("prerequisite", "magic-initiate/spell-change/new-level"),
        "with a different spell of the same level from the chosen spell list.": ("prerequisite", "magic-initiate/spell-change/same-list-and-level"),
        "you must choose a different spell list each time.": ("prerequisite", "magic-initiate/repeatability"),
        "Once per turn when you hit a target with a weapon, you can roll the weapon’s damage dice twice and use either roll against the target.": ("feat", "savage-attacker"),
        "You gain proficiency in any combination of three skills or tools of your choice.": ("choice", "skilled/proficiencies"),
    }
    for anchor, (kind, slug) in anchors.items():
        assert anchor in origin, f"missing PDF anchor: {anchor}"
        assert rec(kind, slug)["required_mechanics"]
    assert origin.count("Repeatable. You can take this feat more than once") == 2
    for slug, count, options in [
        ("alert/initiative-swap", 1, ["targets/willing-ally"]),
        ("magic-initiate/spell-list", 1, ["spell-lists/cleric", "spell-lists/druid", "spell-lists/wizard"]),
        ("magic-initiate/spellcasting-ability", 1, ["abilities/intelligence", "abilities/wisdom", "abilities/charisma"]),
        ("magic-initiate/cantrips", 2, ["spells/selected-list/cantrips"]),
        ("magic-initiate/level-one-spell", 1, ["spells/selected-list/level-one"]),
        ("magic-initiate/spell-change", 1, ["spells/selected-list/replacement"]),
        ("savage-attacker/damage-roll", 1, ["feats/savage-attacker/damage-roll/first", "feats/savage-attacker/damage-roll/second"]),
        ("skilled/proficiencies", 3, ["proficiencies/any-skill", "proficiencies/any-tool"]),
    ]:
        r = rec("choice", slug)
        assert r["select"] == count and r["option_keys"] == [KEY + k for k in options]
    assert rec("resource", "magic-initiate/free-cast")["capacity"] == 1
    assert rec("prerequisite", "magic-initiate/cantrips/selected-list")["requirement"]["parameters"] == {"spell-level": 0}
    assert rec("prerequisite", "magic-initiate/level-one-spell/selected-list")["requirement"]["parameters"] == {"spell-level": 1}
    assert rec("feat", "savage-attacker")["parameters"] == {
        KEY + "feats/savage-attacker/roll-count": 2, KEY + "feats/savage-attacker/uses-per-turn": 1}
    print(f"SRD {PIN}: exact four Origin feat identities, p87, choices, numeric limits, conditions, recovery and repeatability anchors match")


if __name__ == "__main__":
    main()
