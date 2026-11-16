#!/usr/bin/env python3
"""Compare #306 background metadata to pinned SRD 5.2.1 PDF p83, offline."""

import hashlib
import itertools
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
        raise SystemExit("usage: verify_srd_backgrounds.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    columns = []
    for x in (0, 310):
        text = subprocess.run(
            ["pdftotext", "-f", "83", "-l", "83", "-x", str(x), "-y", "0",
             "-W", "310", "-H", "792", "-layout", str(pdf), "-"],
            capture_output=True, text=True, check=True,
        ).stdout
        columns.append(" ".join(re.sub(r"-\s+", "", text).split()))
    left, right = columns
    assert "Increase one by 2 and another one by 1, or increase all three by 1." in left
    assert "None of these increases can raise a score above 20." in left
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-backgrounds.json").read_text())
    assert c["source"]["sha256"] == PIN and c["source"]["license"] == "CC-BY-4.0"
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]) == 18
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r.get("catalogue_issue") == 306}
    actual = {s["inventory_id"]: s["page"] for r in c["records"] if r["kind"] == "origin" for s in r["sources"]}
    assert len(expected) == 4 and actual == expected, "exact four identities/pages"
    names = ["Acolyte", "Criminal", "Sage", "Soldier"]
    # Exact source labels map to opaque references, without importing item definitions.
    items = {
        "Calligrapher’s Supplies": "calligraphers-supplies", "Book (prayers)": "book/prayers",
        "Holy Symbol": "holy-symbol", "Robe": "robe", "Thieves’ Tools": "thieves-tools",
        "Daggers": "dagger", "Crowbar": "crowbar", "Pouches": "pouch",
        "Traveler’s Clothes": "travelers-clothes", "Quarterstaff": "quarterstaff",
        "Book (history)": "book/history", "Spear": "spear", "Shortbow": "shortbow",
        "Arrows": "arrow", "Gaming Set (same as above)": "gaming-set/selected",
        "Healer’s Kit": "healers-kit", "Quiver": "quiver",
    }
    feat_keys = {"Magic Initiate (Cleric)": "magic-initiate/cleric", "Alert": "alert",
                 "Magic Initiate (Wizard)": "magic-initiate/wizard", "Savage Attacker": "savage-attacker"}
    for i, name in enumerate(names):
        column = left if i < 2 else right
        block = column.split(name + " Ability Scores: ", 1)[1]
        abilities, block = block.split(" Feat: ", 1)
        abilities = abilities.lower().split(", ")
        feat, block = block.split(" (see “Feats”) Skill Proficiencies: ", 1)
        skills, block = block.split(" Tool Proficiency: ", 1)
        tool, equipment = block.split(" Equipment: Choose A or B: (A) ", 1)
        package, equipment = equipment.split("; or (B) ", 1)
        gold = int(re.match(r"(\d+) GP", equipment)[1])
        slug = name.lower()
        def record(kind, part=""):
            r = records[ID + kind + "/" + slug + ("/" + part if part else "")]
            assert r["sources"] == [{"inventory_id": "srd-5.2.1:planning/backgrounds/" + slug, "page": 83}]
            return r
        root = record("origin")
        assert root["name"] == name
        fixed = [KEY + "backgrounds/" + slug, KEY + "feats/" + feat_keys[feat]]
        fixed += [KEY + "skills/" + s.lower().replace(" ", "-") for s in skills.split(" and ")]
        if slug != "soldier":
            fixed += [KEY + "tools/" + items[tool]]
        assert root["required_mechanics"] == fixed
        ability = record("choice", "ability-increases")
        options = [{"key": KEY + "backgrounds/" + slug + "/ability-increases/" + a + "-2-" + b + "-1",
                    "quantities": {KEY + "abilities/" + a: 2, KEY + "abilities/" + b: 1}}
                   for a, b in itertools.permutations(abilities, 2)]
        options.append({"key": KEY + "backgrounds/" + slug + "/ability-increases/all-three-1",
                        "quantities": {KEY + "abilities/" + a: 1 for a in abilities}})
        assert ability["select"] == 1 and ability["options"] == options
        assert ability["option_keys"] == [o["key"] for o in options]
        assert record("prerequisite", "ability-score-cap")["requirement"] == {
            "key": KEY + "prerequisite/background-ability-score-maximum",
            "record_refs": [ability["id"]], "parameters": {"maximum-score": 20}}
        quantities = {}
        for entry in package.split(", "):
            if match := re.fullmatch(r"Parchment \((\d+) sheets\)", entry):
                key, count = "parchment/sheets", int(match[1])
            elif match := re.fullmatch(r"(\d+) GP", entry):
                key, count = "currency/gp", int(match[1])
            else:
                match = re.fullmatch(r"(?:(\d+) )?(.+)", entry)
                key, count = items[match[2]], int(match[1] or 1)
            quantities[KEY + "equipment/" + key] = count
        eq = record("choice", "starting-equipment")
        keys = [KEY + "backgrounds/" + slug + "/equipment/" + v for v in ("a", "b")]
        assert eq["select"] == 1 and eq["option_keys"] == keys
        assert eq["options"] == [{"key": keys[0], "quantities": quantities},
                                  {"key": keys[1], "quantities": {KEY + "equipment/currency/gp": gold}}]
        if slug == "soldier":
            assert tool == "Choose one kind of Gaming Set (see “Equipment”)"
            gaming = record("choice", "gaming-set")
            assert gaming["select"] == 1 and gaming["option_keys"] == [KEY + "tools/gaming-set"]
            assert record("prerequisite", "equipment-gaming-set")["requirement"] == {
                "key": KEY + "prerequisite/equipment-a-matches-selected-gaming-set",
                "record_refs": [gaming["id"]], "parameters": {}}
    print(f"SRD {PIN}: four background identities/pages, exact ability domains/distributions/cap, feat/skill/tool references, equipment quantities/currency and Soldier matching choice agree with PDF p83")


if __name__ == "__main__":
    main()
