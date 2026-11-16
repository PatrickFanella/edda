#!/usr/bin/env python3
"""Compare #299's four feat identities and complete benefits to the pinned PDF."""

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


def column(pdf, page, x):
    text = subprocess.run(
        ["pdftotext", "-f", str(page), "-l", str(page), "-x", str(x),
         "-y", "0", "-W", "310", "-H", "792", "-layout", str(pdf), "-"],
        capture_output=True, text=True, check=True,
    ).stdout
    return " ".join(re.sub(r"-\s+", "", text).split())


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_fighting_styles.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    p87 = column(pdf, 87, 0)
    assert "A feat can be taken only once unless its description states otherwise" in p87
    archery = column(pdf, 87, 310).split("Fighting Style Feats Archery ", 1)[1]
    archery = archery.split("87 System Reference", 1)[0].strip()
    p88 = column(pdf, 88, 0)
    prefix = "Fighting Style Feat (Prerequisite: Fighting Style Feature) "
    cases = [
        ("archery", "Archery", 87, archery,
         "You gain a +2 bonus to attack rolls you make with Ranged weapons.",
         ["attack-rolls/with-ranged-weapons"], {"attack-roll-bonus": 2}),
        ("defense", "Defense", 88,
         p88.split("Defense ", 1)[1].split(" Great Weapon Fighting", 1)[0],
         "While you’re wearing Light, Medium, or Heavy armor, you gain a +1 bonus to Armor Class.",
         ["armor-class/while-wearing-light-or-medium-or-heavy-armor"], {"armor-class-bonus": 1}),
        ("great-weapon-fighting", "Great Weapon Fighting", 88,
         p88.split("Great Weapon Fighting ", 1)[1].split(" Two-Weapon Fighting", 1)[0],
         "When you roll damage for an attack you make with a Melee weapon that you are holding with two hands, you can treat any 1 or 2 on a damage die as a 3. The weapon must have the Two-Handed or Versatile property to gain this benefit.",
         ["damage-die/optional-treat-result", "damage-die/when-rolling-damage-for-own-attack",
          "weapon/melee", "weapon/held-with-two-hands", "weapon/property-two-handed-or-versatile"],
         {"damage-die/minimum-result": 1, "damage-die/maximum-result": 2, "damage-die/treat-as": 3}),
        ("two-weapon-fighting", "Two-Weapon Fighting", 88,
         p88.split("Two-Weapon Fighting ", 1)[1].split(" Epic Boon Feats", 1)[0],
         "When you make an extra attack as a result of using a weapon that has the Light property, you can add your ability modifier to the damage of that attack if you aren’t already adding it to the damage.",
         ["damage/optional-add-ability-modifier", "attack/extra-from-using-light-property-weapon",
          "damage/only-if-not-already-adding-ability-modifier"], {}),
    ]
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-fighting-styles.json").read_text())
    assert c["source"]["sha256"] == PIN and c["source"]["license"] == "CC-BY-4.0"
    assert c["source"]["attribution"] and c["source"]["changes"]
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]) == 8
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r.get("catalogue_issue") == 299}
    actual = {s["inventory_id"]: s["page"] for r in c["records"] if r["kind"] == "feat" for s in r["sources"]}
    assert len(expected) == 4 and actual == expected, "exact four identities/pages"
    expected_ids = set()
    for slug, name, page, extracted, benefit, conditions, numbers in cases:
        assert extracted == prefix + benefit, (slug, "source text changed", extracted)
        root_id = ID + "feat/" + slug
        predicate_id = ID + "prerequisite/" + slug + "/fighting-style-feature"
        expected_ids.update((root_id, predicate_id))
        r, p = records[root_id], records[predicate_id]
        source = [{"inventory_id": "srd-5.2.1:planning/feats/" + slug, "page": page}]
        assert r["name"] == name and r["parent_id"] == "" and r["kind"] == "feat"
        assert r["sources"] == p["sources"] == source
        assert r["prerequisite_refs"] == [predicate_id]
        assert r["choice_refs"] == r["resource_refs"] == r["option_keys"] == []
        assert r["select"] == r["capacity"] == r["level"] == 0
        base = KEY + "feats/" + slug + "/"
        nums = {KEY + "feats/maximum-times-taken": 1, **{base + k: v for k, v in numbers.items()}}
        assert r["parameters"] == nums
        required = [KEY + "feats/category/fighting-style", KEY + "feats/maximum-times-taken"]
        required += [base + k for k in conditions] + [base + k for k in numbers]
        assert r["required_mechanics"] == required
        key = KEY + "prerequisite/has-fighting-style-feature"
        assert p["parent_id"] == root_id and p["kind"] == "prerequisite"
        assert p["required_mechanics"] == [key]
        assert p["requirement"] == {"key": key, "record_refs": [], "parameters": {}}
    assert set(records) == expected_ids, "adjacent import"
    print("SRD " + PIN + ": exact four Fighting Style identities/pages; category, general prerequisite, acquisition limit, all benefits/numbers/conditions agree with PDF pp87-88")


if __name__ == "__main__":
    main()
