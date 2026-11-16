#!/usr/bin/env python3
"""Compare #290's canonical Fighter metadata with the pinned PDF, offline."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_fighter.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    text = subprocess.run(
        ["pdftotext", "-f", "47", "-l", "48", "-layout", str(pdf), "-"],
        check=True, capture_output=True, text=True,
    ).stdout
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-fighter.json").read_text())
    assert c["source"]["sha256"] == PIN
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]), "duplicate record IDs"
    inv = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inv["records"]
                if r["id"] == "srd-5.2.1:planning/classes/fighter" or
                r["id"].startswith("srd-5.2.1:planning/class-features/fighter-level-")}
    actual = {s["inventory_id"]: s["page"] for r in c["records"]
              if r["kind"] in ("class", "feature") for s in r["sources"]}
    assert len(expected) == 16 and actual == expected, "source identities/pages"
    rows = re.findall(r"^\s+(\d{1,2})\s+\+(\d)\s+(.+?)\s+(\d)\s+(\d)\s*$", text, re.MULTILINE)
    assert [int(r[0]) for r in rows] == list(range(1, 21)), "PDF table extraction"
    def count(kind, slug, level):
        r = records[f"srd-5.2.1:character/{kind}/fighter/{slug}"]
        return next((v["value"] for v in r["values_by_level"] if v["level"] == level), 0)
    for level, bonus, features, wind, mastery in rows:
        n = int(level)
        p = records[f"srd-5.2.1:character/level/fighter/{n}"]["progression"]
        assert p["proficiency_bonus"] == int(bonus)
        assert count("resource", "second-wind", n) == int(wind)
        assert count("choice", "weapon-mastery", n) == int(mastery)
        names = [records[ref]["name"] for ref in p["feature_refs"]]
        names += ["Subclass feature" for _ in p["feature_keys"]]
        # Resource annotations in the PDF are checked separately below.
        for name, amount in re.findall(r"(Action Surge|Indomitable) \((one|two|three) uses?\)", features):
            slug = name.lower().replace(" ", "-")
            assert count("resource", slug, n) == {"one": 1, "two": 2, "three": 3}[amount]
        headings = re.sub(r" \((?:one|two|three) uses?\)", "", features)
        assert ", ".join(names) == headings, (n, names, headings)
    # Source explicitly gives the initial and improved use limits and recovery.
    for slug, start, changes in [("action-surge", 2, {17: 2}), ("indomitable", 9, {13: 2, 17: 3})]:
        value = 0
        for n in range(1, 21):
            if n == start:
                value = 1
            value = changes.get(n, value)
            assert count("resource", slug, n) == value
    # Anchor nested-domain transcription to the extracted source. Exact option
    # lists, quantities, prerequisites and per-level counts also have Go fixtures.
    flat = " ".join(text.replace("\u00ad", "").split())
    for fragment in ["Choose 2: Acrobatics, Animal", "Handling, Athletics, History,",
                     "Insight, Intimidation, Persua-", "sion, Perception, or Survival",
                     "Choose A, B, or C: (A) Chain", "Mail, Greatsword, Flail, 8 Jav-",
                     "elins, Dungeoneer’s Pack, and", "4 GP; (B) Studded Leather",
                     "Armor, Scimitar, Shortsword,", "Longbow, 20 Arrows, Quiver,",
                     "Dungeoneer’s Pack, and 11", "GP; or (C) 155 GP",
                     "with the Push, Sap, or Slow property", "for which you qualify"]:
        assert fragment in flat, f"source anchor missing: {fragment}"
    print(f"SRD {PIN}: 16 identities/pages, all 20 Fighter rows and choice anchors match PDF pp47-48")


if __name__ == "__main__":
    main()
