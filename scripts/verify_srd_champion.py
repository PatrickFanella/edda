#!/usr/bin/env python3
"""Compare #312 Champion metadata with the pinned PDF p49, offline."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
BASE = "srd-5.2.1:character/"
KEY = "srd-5.2.1:mechanic/fighter/champion/"


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_champion.py /path/to/SRD_CC_v5.2.1.pdf")
    pdf = Path(sys.argv[1])
    if hashlib.sha256(pdf.read_bytes()).hexdigest() != PIN:
        raise SystemExit("PDF pin mismatch")
    # Crop the left column: the adjacent Monk entry is explicitly outside #312.
    text = subprocess.run(
        ["pdftotext", "-f", "49", "-l", "49", "-x", "0", "-y", "0",
         "-W", "310", "-H", "792", "-layout", str(pdf), "-"],
        check=True, capture_output=True, text=True,
    ).stdout
    flat = " ".join(text.split()).replace("Inspi- ration", "Inspiration").replace("Constitu- tion", "Constitution")
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-champion.json").read_text())
    assert c["source"]["sha256"] == PIN
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]) == 16, "record identities"
    inv = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inv["records"] if r.get("catalogue_issue") == 312}
    actual = {s["inventory_id"]: s["page"] for r in c["records"]
              if r["kind"] in ("feature", "subclass") for s in r["sources"]}
    assert len(expected) == 7 and actual == expected, "source identities/pages"
    assert "Fighter Subclass: Champion" in flat and "Monk" not in flat
    headings = re.findall(r"Level (\d+): ([^\n]+)", text)
    assert len(headings) == 6
    for level, name in headings:
        slug = name.strip().lower().replace(" ", "-")
        r = records[BASE + "feature/fighter/champion/" + slug]
        assert r["name"] == name.strip()
        assert records[r["prerequisite_refs"][0]]["requirement"]["parameters"] == {"minimum-level": int(level)}
    def feature(slug):
        return records[BASE + "feature/fighter/champion/" + slug]
    for slug, pattern in [("improved-critical", r"roll of (\d+) or (\d+)"),
                          ("superior-critical", r"(\d+)–(\d+) on the d20")]:
        low, high = map(int, re.search(pattern, flat).groups())
        assert feature(slug)["parameters"] == {KEY + "critical/d20-minimum": low, KEY + "critical/d20-maximum": high}
    assert "up to half your Speed" in flat
    assert feature("remarkable-athlete")["parameters"] == {KEY + "critical-movement/speed-divisor": 2}
    low, high, benefit = map(int, re.search(r"roll (\d+)–(\d+) on a Death Saving Throw, you gain the benefit of rolling a (\d+)", flat).groups())
    base, minimum = map(int, re.search(r"equal to (\d+) plus your Constitution modifier if you are Bloodied and have at least (\d+)", flat).groups())
    assert feature("survivor")["parameters"] == {
        KEY + "defy-death/d20-minimum": low, KEY + "defy-death/d20-maximum": high,
        KEY + "defy-death/benefit-of-roll": benefit, KEY + "heroic-rally/healing-base": base,
        KEY + "heroic-rally/minimum-current-hit-points": minimum,
    }
    choice = records[feature("additional-fighting-style")["choice_refs"][0]]
    assert choice["select"] == 1 and choice["option_keys"] == ["srd-5.2.1:mechanic/feats/fighting-style"]
    resource = records[feature("heroic-warrior")["resource_refs"][0]]
    assert resource["capacity"] == 1  # "without it": no stacking, not a rest use pool.
    for fragment in ["weapons and Unarmed Strikes", "Advantage on Initiative rolls and Strength (Athletics) checks",
                     "immediately after you score a Critical Hit", "without provoking Opportunity Attacks",
                     "another Fighting Style feat of your choice", "During combat, you can give yourself Heroic Inspiration",
                     "whenever you start your turn without it", "Advantage on Death Saving Throws",
                     "At the start of each of your turns", "if you are Bloodied and have at least 1 Hit Point"]:
        assert fragment in flat, f"source anchor missing: {fragment}"
    assert "jump" not in flat.lower(), "unexpected source edition"
    print(f"SRD {PIN}: 7 Champion identities/pages, six levels, numeric improvements and choice/condition anchors match PDF p49")


if __name__ == "__main__":
    main()
