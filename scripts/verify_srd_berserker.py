#!/usr/bin/env python3
"""Compare #308 metadata against the pinned PDF p30, without network access."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
ID = "srd-5.2.1:character/"
KEY = "srd-5.2.1:mechanic/barbarian/berserker/"


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    text = subprocess.run(
        ["pdftotext", "-f", "30", "-l", "30", "-x", "310", "-y", "0",
         "-W", "302", "-H", "792", "-layout", str(pdf), "-"],
        check=True, capture_output=True, text=True,
    ).stdout
    # Exclude the preceding Primal Champion text in the same column.
    text = text[text.index("Barbarian Subclass:"):]
    flat = " ".join(text.split())
    flat = re.sub(r"(\w)- (\w)", r"\1\2", flat)
    c = json.loads((ROOT / "internal/rules/srd521/testdata/character-berserker.json").read_text())
    assert c["source"]["sha256"] == PIN
    records = {r["id"]: r for r in c["records"]}
    assert len(records) == len(c["records"]) == 11
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r.get("catalogue_issue") == 308}
    actual = {s["inventory_id"]: s["page"] for r in c["records"]
              if r["kind"] in ("subclass", "feature") for s in r["sources"]}
    assert len(expected) == 5 and actual == expected, "source identities/pages"
    sub = records[ID + "subclass/barbarian/berserker"]
    assert "Barbarian Subclass: " + sub["name"] in flat
    for r in c["records"]:
        if r["kind"] == "prerequisite":
            assert r["required_mechanics"] == ["srd-5.2.1:mechanic/prerequisite/class-level"], "prerequisite mechanics"
    headings = re.findall(r"Level (\d+): ([^\n]+)", text)
    assert headings == [("3", "Frenzy"), ("6", "Mindless Rage"), ("10", "Retaliation"), ("14", "Intimidating Presence")]
    for level, name in headings:
        slug = name.lower().replace(" ", "-")
        r = records[ID + "feature/barbarian/berserker/" + slug]
        assert r["name"] == name and r["parent_id"] == sub["id"]
        assert records[r["prerequisite_refs"][0]]["requirement"]["parameters"] == {"minimum-level": int(level)}

    # Each opaque condition has an independent source anchor. Numerical values
    # are extracted below rather than accepted from catalogue metadata.
    anchors = {
        "frenzy": {
            "requires-reckless-attack": "If you use Reckless Attack",
            "requires-active-rage": "while your Rage is active",
            "first-target-hit-on-own-turn": "the first target you hit on your turn",
            "strength-based-attack": "with a Strength-based attack",
            "extra-damage-dice-count-equals-rage-damage-bonus": "a number of d6s equal to your Rage Damage bonus",
            "extra-damage-sum-dice": "add them together",
            "damage-type-matches-weapon-or-unarmed-strike": "same type as the weapon or Unarmed Strike used for the attack",
        },
        "mindless-rage": {
            "while-rage-active/immunity-charmed": "Immunity to the Charmed",
            "while-rage-active/immunity-frightened": "Frightened conditions while your Rage is active",
            "enter-rage/ends-charmed": "If you’re Charmed or Frightened when you enter your Rage, the condition ends on you",
            "enter-rage/ends-frightened": "If you’re Charmed or Frightened when you enter your Rage, the condition ends on you",
        },
        "retaliation": {
            "trigger-take-damage-from-creature": "When you take damage from a creature",
            "damaging-creature-within-range": "within 5 feet of you",
            "optional-reaction": "you can take a Reaction",
            "melee-attack-against-damaging-creature": "one melee attack against that creature",
            "weapon-or-unarmed-strike": "using a weapon or an Unarmed Strike",
        },
        "intimidating-presence": {
            "bonus-action": "As a Bonus Action",
            "each-chosen-creature": "each creature of your choice",
            "emanation-originates-from-self": "Emanation originating from you",
            "wisdom-saving-throw": "Wisdom saving throw",
            "save-dc-plus-strength-modifier": "plus your Strength modifier",
            "save-dc-plus-proficiency-bonus": "and Proficiency Bonus",
            "failed-save/frightened": "On a failed save, a creature has the Frightened condition",
            "repeat-save-end-of-each-affected-creature-turn": "At the end of each of the Frightened creature’s turns, the creature repeats the save",
            "successful-repeat-save/ends-effect": "ending the effect on itself on a success",
            "restore-use/expend-rage-use": "expend a use of your Rage",
            "restore-use/no-action-required": "no action required",
        },
    }
    d6 = int(re.search(r"number of d(\d+)s", flat)[1])
    feet = int(re.search(r"within (\d+) feet", flat)[1])
    emanation = int(re.search(r"(\d+)-foot Emanation", flat)[1])
    dc = int(re.search(r"DC (\d+) plus", flat)[1])
    duration = int(re.search(r"for (\d+) minute", flat)[1])
    numbers = {
        "frenzy": {"damage-die-sides": d6},
        "mindless-rage": {},
        "retaliation": {"range-feet": feet, "attack-count": 1},  # source: one
        "intimidating-presence": {"emanation-feet": emanation, "save-dc-base": dc,
                                  "duration-minutes": duration, "restore-use/rage-cost": 1},  # a use
    }
    for slug, conditions in anchors.items():
        r = records[ID + "feature/barbarian/berserker/" + slug]
        for anchor in conditions.values():
            assert anchor in flat, f"source anchor missing: {anchor}"
        prefix = KEY + slug + "/"
        assert set(r["required_mechanics"]) == {prefix + k for k in conditions.keys() | numbers[slug].keys()}
        assert r.get("parameters", {}) == {prefix + k: v for k, v in numbers[slug].items()}
    resource = records[ID + "resource/barbarian/berserker/intimidating-presence"]
    assert resource["capacity"] == 1
    assert not resource.get("values_by_level"), "unexpected resource scaling"
    assert "Once you use this feature, you can’t use it again until you finish a Long Rest unless" in flat
    assert "expend a use of your Rage (no action required) to restore your use of it" in flat
    assert resource["required_mechanics"] == [KEY + "intimidating-presence/use/long-rest-reset",
                                             KEY + "intimidating-presence/use/restored-by-expending-rage-no-action"]
    print(f"SRD {PIN}: 5 Berserker identities/pages, four levels, nested conditions, numbers and use restoration match p30")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_berserker.py /path/to/SRD_CC_v5.2.1.pdf")
    verify(Path(sys.argv[1]))
