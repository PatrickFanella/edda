#!/usr/bin/env python3
"""Independent offline p52 source-fact comparison for #313; no Go or network."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
C = "srd-5.2.1:character/"
M = "srd-5.2.1:mechanic/"
P = "srd-5.2.1:planning/"
BASE = "monk/warrior-of-the-open-hand/"


def source_records():
    out = []
    def record(kind, slug, name, source, parent, keys=(), nums=None):
        r = dict(id=C+kind+"/"+BASE+slug, kind=kind, name=name,
                 sources=[dict(inventory_id=P+source, page=52)], parent_id=parent,
                 level=0, choice_refs=[], resource_refs=[], prerequisite_refs=[],
                 required_mechanics=[M+BASE+k for k in keys], select=0,
                 option_keys=[], capacity=0)
        if nums:
            r["parameters"] = {M+BASE+k:v for k,v in nums.items()}
            r["required_mechanics"] += list(r["parameters"])
        out.append(r)
        return r
    def pred(owner, suffix, key, refs, nums=None):
        r = record("prerequisite", owner["id"].split("/")[-1]+"-"+suffix,
                   owner["name"]+" "+suffix, owner["sources"][0]["inventory_id"][len(P):],
                   owner["id"], [key])
        r["requirement"] = dict(key=M+BASE+key, record_refs=[C+x for x in refs], parameters=nums or {})
        owner["prerequisite_refs"].append(r["id"])
    subclass = record("subclass", "identity", "Warrior of the Open Hand",
                      "subclasses/monk-warrior-of-the-open-hand", C+"class/monk", ["subclass"])
    pred(subclass, "level", "class-level", ["class/monk"], {"minimum-level":3})
    def feature(slug, name, level, keys, nums=None):
        r = record("feature", slug, name, f"subclass-features/monk-level-{level}-{slug}", subclass["id"], keys, nums)
        pred(r, "level", "class-level", ["class/monk"], {"minimum-level":level})
        return r
    hand = feature("open-hand-technique", "Open Hand Technique", 3,
                   ["open-hand-technique/optional-on-each-hit", "open-hand-technique/target/hit-creature"],
                   {"open-hand-technique/push/maximum-distance-feet":15})
    pred(hand, "trigger", "open-hand-technique/hit-with-attack-granted-by-flurry-of-blows",
         ["choice/monk/flurry-of-blows"])
    ch = record("choice", "open-hand-technique", "Open Hand Technique effects",
                "subclass-features/monk-level-3-open-hand-technique", hand["id"], ["open-hand-technique/choose-one-effect-on-target"])
    ch["select"] = 1
    ch["option_keys"] = [M+BASE+"open-hand-technique/"+x for x in [
        "addle/no-opportunity-attacks-until-start-of-target-next-turn",
        "push/strength-save-or-pushed-up-to-15-feet-away-from-you",
        "topple/dexterity-save-or-prone"]]
    hand["choice_refs"].append(ch["id"])
    body = feature("wholeness-of-body", "Wholeness of Body", 6,
                   ["wholeness-of-body/bonus-action", "wholeness-of-body/target/self",
                    "wholeness-of-body/regain-hit-points/one-martial-arts-die-plus-wisdom-modifier",
                    "wholeness-of-body/uses/wisdom-modifier-minimum-one",
                    "wholeness-of-body/recover-all-expended-uses-on-finish-long-rest"],
                   {"wholeness-of-body/healing-minimum":1, "wholeness-of-body/dice":1,
                    "wholeness-of-body/uses-minimum":1, "wholeness-of-body/uses-expended":1})
    pred(body, "die", "wholeness-of-body/roll-canonical-martial-arts-die", ["feature/monk/martial-arts"])
    fleet = feature("fleet-step", "Fleet Step", 11, ["fleet-step/optional-step-of-the-wind-immediately-after-bonus-action"])
    pred(fleet, "trigger", "fleet-step/bonus-action-other-than-step-of-the-wind", ["choice/monk/step-of-the-wind"])
    pred(fleet, "effect", "fleet-step/use-canonical-step-of-the-wind-with-its-alternatives-and-costs", ["choice/monk/step-of-the-wind"])
    palm = feature("quivering-palm", "Quivering Palm", 17,
                   ["quivering-palm/optional-on-unarmed-strike-hit-creature",
                    "quivering-palm/vibrations/imperceptible-and-harmless-unless-ended-for-damage",
                    "quivering-palm/end/action-or-forgo-one-attack-of-attack-action-on-your-turn",
                    "quivering-palm/end/requires-same-plane-as-target",
                    "quivering-palm/end/target-constitution-save",
                    "quivering-palm/damage/force-on-failure-half-on-success",
                    "quivering-palm/end-harmlessly/no-action-required"],
                   {"quivering-palm/focus-cost":4, "quivering-palm/forgone-attacks":1,
                    "quivering-palm/damage-dice":10, "quivering-palm/damage-die-sides":12,
                    "quivering-palm/maximum-simultaneous-creatures":1})
    pred(palm, "focus", "quivering-palm/expend-available-focus", ["resource/monk/focus-points"], {"points-required":4})
    pred(palm, "duration", "quivering-palm/vibration-duration-days-equals-monk-level", ["class/monk"])
    return out


def compare(records):
    expected = source_records()
    assert len(records) == len(expected), "p52 record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"p52 semantic mismatch {want['id']}: {actual} != {want}"


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    raw = subprocess.check_output(["pdftotext", "-f", "52", "-l", "52", "-x", "305", "-y", "0", "-W", "310", "-H", "792", "-layout", str(pdf), "-"], text=True)
    text = " ".join(re.sub(r"-\n\s*", "", raw.replace("\u00ad", "")).split())
    clauses = ["Warrior of the", "Open Hand", "Level 3: Open Hand Technique", "Whenever you hit a creature", "attack granted by your Flurry of Blows", "one of the following effects on that target", "Addle", "can’t make Opportunity Attacks", "until the start of its next turn", "Strength saving throw", "up to 15 feet away from you", "Dexterity saving throw", "Prone condition", "Level 6: Wholeness of Body", "heal yourself", "As a Bonus Action", "roll your Martial Arts die", "plus your Wisdom modifier", "minimum of 1 Hit Point", "number of times equal to your Wisdom modifier", "minimum of once", "regain all expended uses", "finish a Long Rest", "Level 11: Fleet Step", "Bonus Action other than Step of the Wind", "immediately after that Bonus Action", "Level 17: Quivering Palm", "hit a creature with an Unarmed Strike", "expend 4 Focus Points", "imperceptible vibrations", "number of days equal to your Monk level", "harmless unless you take an action", "Alternatively", "Attack action on your turn", "forgo one of the attacks", "same plane of existence", "Constitution saving throw", "10d12 Force damage", "half as much damage on a successful one", "only one creature", "at a time", "end the vibrations harmlessly", "no action required"]
    assert all(x in text for x in clauses), [x for x in clauses if x not in text]
    actual = json.loads((ROOT/"internal/rules/srd521/testdata/character-open-hand.json").read_text())
    assert actual["schema_version"] == 1 and actual["source"]["sha256"] == PIN
    assert actual["source"]["license"] == "CC-BY-4.0" and actual["source"]["changes"] and actual["source"]["attribution"]
    compare(actual["records"])
    inv = json.loads((ROOT/"docs/investigations/issue-74-character-inventory.json").read_text())
    identities = [r for r in actual["records"] if r["kind"] in ("subclass", "feature")]
    expected = {r["id"]:r["source_page"] for r in inv["records"] if r["batch_key"] == "subclass-monk"}
    assert len(identities) == len(expected) == 5
    assert {s["inventory_id"]:s["page"] for r in identities for s in r["sources"]} == expected
    from verify_srd_monk import verify as verify_monk
    verify_monk(pdf)
    print(f"p52: five identities and {len(actual['records'])} complete records match independent source facts")


if __name__ == "__main__":
    if len(sys.argv) == 3 and sys.argv[1] == "--controls":
        controls = json.loads(Path(sys.argv[2]).read_text())
        for records in controls:
            try:
                compare(records)
            except AssertionError:
                continue
            raise AssertionError("contrary graph accepted")
        print(f"{len(controls)} contrary graphs rejected")
    elif len(sys.argv) == 2:
        verify(Path(sys.argv[1]))
    else:
        raise SystemExit("usage: verify_srd_open_hand.py PDF | --controls JSON")
