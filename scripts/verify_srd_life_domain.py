#!/usr/bin/env python3
"""Compare #310 metadata with the local pinned SRD p40, offline."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
BASE = "srd-5.2.1:character/"
KEY = "srd-5.2.1:mechanic/cleric/life-domain/"


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    columns = []
    for x in (0, 310):
        columns.append(subprocess.run(
            ["pdftotext", "-f", "40", "-l", "40", "-x", str(x), "-y", "0",
             "-W", "310", "-H", "792", "-layout", str(pdf), "-"],
            check=True, capture_output=True, text=True,
        ).stdout)
    text = "\n".join(columns)
    # Rejoin source line wrapping before matching exact edition-specific anchors.
    flat = " ".join(text.split())
    for split in ("addi- tional", "crea- tures", "Chan- nel", "in- stead"):
        flat = flat.replace(split, split.replace("- ", ""))
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-life-domain.json").read_text())
    assert c["source"]["sha256"] == PIN
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]) == 13
    inv = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inv["records"] if r.get("catalogue_issue") == 310}
    actual = {s["inventory_id"]: s["page"] for r in c["records"]
              if r["kind"] in ("feature", "subclass") for s in r["sources"]}
    assert len(expected) == 6 and actual == expected, "identities/pages"
    assert "Cleric Subclass: Life Domain" in flat
    headings = re.findall(r"Level (\d+): ([^\n]+)", text)
    assert len(headings) == 5
    for level, name in headings:
        name = name.strip()
        slug = name.lower().replace(" ", "-")
        r = records[BASE + "feature/cleric/life-domain/" + slug]
        assert r["name"] == name
        predicate = records[r["prerequisite_refs"][0]]["requirement"]
        assert predicate == {"key": "srd-5.2.1:mechanic/prerequisite/class-level",
                             "record_refs": [BASE + "class/cleric"],
                             "parameters": {"minimum-level": int(level)}}
    def feature(slug):
        return records[BASE + "feature/cleric/life-domain/" + slug]
    table = re.search(r"Life Domain Spells\s+Cleric\s+Prepared Spells\s+Level\s+(.+?)Level 3: Preserve Life", columns[1], re.S).group(1)
    extracted = re.findall(r"\b([3579])\s+(.+?)(?=\s+[3579]\s+|$)", table.strip(), re.S)
    rows = []
    for level, names in extracted:
        spells = ["srd-5.2.1:mechanic/spells/" + "-".join(name.lower().split()) for name in names.split(",")]
        rows.append({"level": int(level), "spell_keys": spells})
    assert len(rows) == 4 and rows == feature("life-domain-spells")["always_prepared"], "all spell rows"
    assert all(r["kind"] not in ("choice", "resource", "advancement") for r in c["records"])
    bases = re.findall(r"equal (?:to )?(\d+) plus the spell slot’s", flat)
    assert bases == ["2", "2"]
    for slug, suffix, number in [("disciple-of-life", "additional-healing/base", int(bases[0])),
                                 ("blessed-healer", "self-healing/base", int(bases[1]))]:
        assert feature(slug)["parameters"] == {KEY + slug + "/" + suffix: number}
    distance = int(re.search(r"Bloodied creatures within (\d+) feet", flat).group(1))
    assert "five times your Cleric level" in flat and "more than half its Hit Point maximum" in flat
    assert feature("preserve-life")["parameters"] == {
        KEY + "preserve-life/channel-divinity/uses-expended": 1,
        KEY + "preserve-life/healing-pool/cleric-level-multiplier": 5,
        KEY + "preserve-life/targets/range-feet": distance,
        KEY + "preserve-life/healing-cap/hit-point-maximum-divisor": 2,
    }
    count, sides, healing = map(int, re.search(r"restoring (\d+)d(\d+) Hit Points to a creature with a spell, you restore (\d+)", flat).groups())
    assert feature("supreme-healing")["parameters"] == {
        KEY + "supreme-healing/example/dice-count": count,
        KEY + "supreme-healing/example/die-sides": sides,
        KEY + "supreme-healing/example/healing": healing,
    }
    for anchor in [
        "When a spell you cast with a spell slot restores Hit Points to a creature",
        "additional Hit Points on the turn you cast the spell",
        "you thereafter always have the listed spells prepared",
        "As a Magic action, you present your Holy Symbol and expend a use of your Channel Divinity",
        "(which can include you)", "divide those Hit Points among them",
        "Immediately after you cast a spell with a spell slot that restores Hit Points to one or more creatures other than yourself",
        "When you would normally roll one or more dice", "with a spell or Channel Divinity",
        "use the highest number possible for each die",
    ]:
        assert anchor in flat, f"source anchor missing: {anchor}"
    print(f"SRD {PIN}: six identities/pages, five eligibility levels, four fixed spell rows and all numeric/condition anchors match p40")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_life_domain.py /path/to/SRD_CC_v5.2.1.pdf")
    verify(Path(sys.argv[1]))
