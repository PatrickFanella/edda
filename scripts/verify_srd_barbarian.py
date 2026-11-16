#!/usr/bin/env python3
"""Offline #286 comparison with the exact licensed SRD pp28–30."""
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


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    text = subprocess.run(["pdftotext", "-layout", "-f", "28", "-l", "30", str(pdf), "-"],
                          check=True, capture_output=True, text=True).stdout
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-barbarian.json").read_text())
    assert c["source"]["sha256"] == PIN
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"])
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"]
                if r["id"] == "srd-5.2.1:planning/classes/barbarian" or
                r["id"].startswith("srd-5.2.1:planning/class-features/barbarian-level-")}
    canonical = [r for r in c["records"] if r["kind"] in ("class", "feature")]
    actual = {s["inventory_id"]: s["page"] for r in canonical for s in r["sources"]}
    assert len(canonical) == len(expected) == 21 and actual == expected
    # Numeric suffix and actual source feature names, including both equal-named
    # Improved Brutal Strike headings, are compared to every catalogue row.
    rows = re.findall(r"^\s+(\d{1,2})\s+\+(\d)\s+(.+?)\s{2,}(\d+)\s+\+(\d+)\s+(\d+)\s*$", text, re.M)
    assert [int(row[0]) for row in rows] == list(range(1, 21))
    def count(kind, slug, n):
        return next(v["value"] for v in records[ID + kind + "/barbarian/" + slug]["values_by_level"] if v["level"] == n)
    for level, bonus, names, uses, damage, mastery in rows:
        n = int(level)
        row = records[f"{ID}level/barbarian/{n}"]
        p = row["progression"]
        assert p["proficiency_bonus"] == int(bonus)
        actual = [records[ref]["name"] for ref in p["feature_refs"]]
        actual += ["Subclass feature" for _ in p["feature_keys"]]
        # Title casing in the metadata does not change source identity.
        assert ", ".join(actual).lower() == names.lower(), (n, actual, names)
        assert [count("resource", "rage", n), row["parameters"][KEY + "barbarian/rage/damage-bonus"],
                count("choice", "weapon-mastery", n)] == [int(uses), int(damage), int(mastery)]
    # Column-preserving extraction keeps prose discontinuous. Anchors deliberately
    # follow the printed lines, and do not consume Berserker source headings.
    flat = " ".join(text.replace("\u00ad", "").split())
    anchors = [
        "Choose 2: Animal Handling,", "Athletics, Intimidation, Na-", "ture, Perception, or Survival",
        "4 Handaxes, Explorer’s Pack,", "and 15 GP; or (B) 75 GP", "D12 per Barbarian level",
        "You can enter it as a Bonus Action if", "you aren’t wearing Heavy armor.",
        "regain one ex-", "pended use when you finish a Short Rest", "regain all expended uses",
        "Resistance to Blud-", "geoning, Piercing, and Slashing damage.", "either a weapon or an Unarmed",
        "Advantage on", "Strength checks and Strength saving throws.", "can’t maintain", "can’t cast spells.",
        "lasts until the end of your", "ends early if you don Heavy armor", "Incapacitated condition",
        "Make an attack roll against an enemy.", "Force an enemy to make a saving throw.",
        "Take a Bonus Action to extend your Rage.", "10 minutes.",
        "10 plus your Dexterity and Consti-", "tution modifiers.", "You can use a Shield",
        "Simple or Mar-", "tial Melee weapons of your choice", "change one",
        "Advantage on Dexterity saving", "first attack roll on your turn", "until the start of your next",
        "Acrobatics, Intimidation, Perception,", "Stealth, or Survival.", "els 8, 12, and 16.",
        "attack twice instead of once", "speed increases by 10 feet", "Initiative rolls.",
        "half your Speed.", "chosen attack roll mustn’t", "have Disadvantage.", "extra 1d10 damage of the same type",
        "Forceful Blow.", "pushed 15 feet", "half your Speed straight toward the target without",
        "provoking Opportunity Attacks.", "Hamstring Blow.", "15 feet until the start of your next turn.",
        "most recent one.", "0 Hit Points", "don’t die outright", "10 Constitution saving throw.",
        "twice your Barbarian level.", "DC increases by 5.", "DC resets to 10.",
        "Staggering Blow.", "next saving throw", "Sundering Blow.", "another creature",
        "+5 bonus to the roll.", "only one Sundering Blow bonus.", "When you roll Initiative",
        "can’t do so again until you finish a", "Long Rest.", "without you needing to do anything",
        "Unconscious condition (not just", "2d10.", "two different Brutal",
        "Strike effects", "less than your Strength score", "that score in place of the total.",
        "Boon of", "Irresistible Offense is recommended.", "scores increase by 4, to a maximum of 25.",
    ]
    missing = [anchor for anchor in anchors if anchor not in flat]
    assert not missing, f"source anchors missing: {missing}"
    print(f"SRD {PIN}: 21 identities/pages, 20 complete Barbarian rows and feature anchors match pp28–30")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_barbarian.py /path/to/SRD_CC_v5.2.1.pdf")
    verify(Path(sys.argv[1]))
