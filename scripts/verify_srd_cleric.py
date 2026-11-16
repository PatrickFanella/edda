#!/usr/bin/env python3
"""Compare #288 Cleric metadata to the local pinned SRD, without network access."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    text = subprocess.run(
        ["pdftotext", "-f", "36", "-l", "38", "-layout", str(pdf), "-"],
        check=True, capture_output=True, text=True,
    ).stdout
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-cleric.json").read_text())
    assert c["source"]["sha256"] == PIN
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]), "duplicate IDs"
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"]
                if r["id"] == "srd-5.2.1:planning/classes/cleric" or
                r["id"].startswith("srd-5.2.1:planning/class-features/cleric-level-")}
    canonical = [r for r in c["records"] if r["kind"] in ("class", "feature")]
    actual = {s["inventory_id"]: s["page"] for r in canonical for s in r["sources"]}
    assert len(canonical) == len(expected) == 12 and actual == expected, "identities/pages"
    # Parse the complete numeric suffix, including every absent spell-slot cell.
    rows = re.findall(r"^\s+(\d{1,2})\s+\+(\d)\s+(.+?)\s{2,}([\d—][\d—\s]+)$", text, re.MULTILINE)
    assert [int(r[0]) for r in rows] == list(range(1, 21)), "table extraction"
    def count(kind, slug, level):
        r = records[f"srd-5.2.1:character/{kind}/cleric/{slug}"]
        return next((v["value"] for v in r["values_by_level"] if v["level"] == level), 0)
    for level, bonus, features, numeric in rows:
        n = int(level)
        cells = [0 if v == "—" else int(v) for v in numeric.split()]
        assert len(cells) == 12, (n, cells)
        p = records[f"srd-5.2.1:character/level/cleric/{n}"]["progression"]
        assert p["proficiency_bonus"] == int(bonus)
        names = [records[ref]["name"] for ref in p["feature_refs"]]
        names += ["Subclass feature" for _ in p["feature_keys"]]
        assert (", ".join(names) or "—") == features, (n, names, features)
        actual = [count("resource", "channel-divinity", n), count("choice", "cantrips", n),
                  count("choice", "prepared-spells", n)]
        actual += [count("resource", f"spell-slot-{j}", n) for j in range(1, 10)]
        assert actual == cells, (n, actual, cells)
    # Bind each nested choice to the role selected in its source parent choice.
    for child, parent, predicate in [
        ("protector", "divine-order", "divine-order/protector-selected"),
        ("thaumaturge", "divine-order", "divine-order/thaumaturge-selected"),
        ("divine-spark", "channel-divinity", "channel-divinity/divine-spark-selected"),
        ("divine-strike", "blessed-strikes", "blessed-strikes/divine-strike-selected"),
    ]:
        choice = records[f"srd-5.2.1:character/choice/cleric/{child}"]
        assert len(choice["prerequisite_refs"]) == 1, child
        prerequisite = records[choice["prerequisite_refs"][0]]
        key = f"srd-5.2.1:mechanic/cleric/{predicate}"
        assert prerequisite["requirement"]["key"] == key, child + " selection predicate"
        assert prerequisite["required_mechanics"] == [key], child + " required selection"
        assert prerequisite["requirement"]["record_refs"] == [f"srd-5.2.1:character/choice/cleric/{parent}"], child
    # These anchors are independent of the JSON and bind nested and conditional
    # choices/numeric metadata to 2024 source wording, rather than 2014 memory.
    flat = " ".join(text.replace("\u00ad", "").split())
    for fragment in [
        "Choose 2: History, Insight,", "Medicine, Persuasion, or", "Religion",
        "Chain Shirt,", "Priest’s Pack, and 7 GP; or (B)", "110 GP",
        "Protector. Trained for battle", "Thaumaturge. You know one extra cantrip",
        "Intelligence (Arcana or Religion)", "minimum of +1",
        "Divine Spark", "Turn Undead", "regain one of its expended uses",
        "levels 7 (2d8), 13 (3d8), and 18 (4d8)", "Constitution saving",
        "Frightened and Incapacitated", "for 1 minute", "you die",
        "minimum of 1d8", "This damage doesn’t end the turn effect",
        "Divine Strike", "Potent Spellcasting", "1d8 Necrotic or Radiant",
        "increases to 2d8", "60 feet", "twice your Wisdom modifier",
        "Cleric spell of level 5 or lower", "Reaction to cast", "Material components",
        "Boon of Fate is recommended", "2d4 Long Rests", "choose Wish",
        "Cleric spell list", "Holy Symbol as", "levels 8,", "12, and 16",
    ]:
        assert fragment in flat, f"source anchor missing: {fragment}"
    print(f"SRD {PIN}: 12 identities/pages, all 20 Cleric rows and nested-choice anchors match pp36–38")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_cleric.py /path/to/SRD_CC_v5.2.1.pdf")
    verify(Path(sys.argv[1]))
