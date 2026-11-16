#!/usr/bin/env python3
"""Independent p61 Hunter metadata proof; no Go, network or runtime rules."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
C = "srd-5.2.1:character/"
M = "srd-5.2.1:mechanic/ranger/hunter/"
P = "srd-5.2.1:planning/"
BASE = "ranger/hunter/"


def source_records(facts=None):
    # Constants also support deterministic contradiction checks without a PDF.
    facts = facts or dict(dice=1, sides=8, nearby=5, visible=30)
    out = []

    def record(kind, slug, name, source, parent, keys=(), nums=None):
        r = dict(
            id=C + kind + "/" + BASE + slug,
            kind=kind,
            name=name,
            sources=[dict(inventory_id=P + source, page=61)],
            parent_id=parent,
            level=0,
            choice_refs=[],
            resource_refs=[],
            prerequisite_refs=[],
            required_mechanics=[M + k for k in keys],
            select=0,
            option_keys=[],
            capacity=0,
        )
        if nums:
            r["parameters"] = {M + k: v for k, v in nums.items()}
        out.append(r)
        return r

    def pred(owner, suffix, key, refs, nums=None):
        r = record(
            "prerequisite",
            owner["id"].split("/")[-1] + "-" + suffix,
            owner["name"] + " " + suffix,
            owner["sources"][0]["inventory_id"][len(P) :],
            owner["id"],
            [key],
        )
        r["requirement"] = dict(
            key=M + key, record_refs=[C + x for x in refs], parameters=nums or {}
        )
        owner["prerequisite_refs"].append(r["id"])

    sub = record(
        "subclass",
        "identity",
        "Hunter",
        "subclasses/ranger-hunter",
        C + "class/ranger",
        ["subclass"],
    )
    pred(sub, "level", "class-level", ["class/ranger"], {"minimum-level": 3})

    def feature(slug, name, level, keys, nums=None):
        r = record(
            "feature",
            slug,
            name,
            f"subclass-features/ranger-level-{level}-{slug}",
            sub["id"],
            keys,
            nums,
        )
        pred(r, "level", "class-level", ["class/ranger"], {"minimum-level": level})
        return r

    def choice(owner, slug, name, options):
        r = record(
            "choice",
            slug,
            name,
            owner["sources"][0]["inventory_id"][len(P) :],
            owner["id"],
            [
                "choose-one",
                "replace-chosen-option-with-other-on-finish-short-or-long-rest",
            ],
        )
        r["select"] = 1
        r["option_keys"] = [M + x for x in options]
        owner["choice_refs"].append(r["id"])
        return r

    def option(owner, slug, name, keys, nums=None):
        keys = [slug + "/only-while-chosen"] + keys
        r = record(
            "prerequisite",
            slug,
            name,
            owner["sources"][0]["inventory_id"][len(P) :],
            owner["id"],
            keys,
        )
        r["requirement"] = dict(
            key=M + keys[0],
            record_refs=[owner["id"]],
            parameters={k.split("/")[-1]: v for k, v in (nums or {}).items()},
        )
        owner["prerequisite_refs"].append(r["id"])

    lore = feature(
        "hunters-lore",
        "Hunter’s Lore",
        3,
        [
            "hunters-lore/know-whether-marked-creature-has-immunities-resistances-vulnerabilities-and-what-they-are"
        ],
    )
    pred(
        lore,
        "mark",
        "target-marked-by-your-hunters-mark",
        ["feature/ranger/favored-enemy"],
    )
    prey = feature(
        "hunters-prey", "Hunter’s Prey", 3, ["hunters-prey/gain-one-chosen-option"]
    )
    ch = choice(
        prey,
        "hunters-prey",
        "Hunter’s Prey options",
        ["colossus-slayer", "horde-breaker"],
    )
    option(
        ch,
        "colossus-slayer",
        "Colossus Slayer",
        [
            "colossus-slayer/on-weapon-hit-creature",
            "colossus-slayer/weapon-extra-damage-to-hit-target",
            "colossus-slayer/only-if-target-missing-any-hit-points",
            "colossus-slayer/once-per-turn",
            "colossus-slayer/damage-dice",
            "colossus-slayer/damage-die-sides",
        ],
        {
            "colossus-slayer/damage-dice": facts["dice"],
            "colossus-slayer/damage-die-sides": facts["sides"],
        },
    )
    option(
        ch,
        "horde-breaker",
        "Horde Breaker",
        [
            "horde-breaker/once-on-each-of-your-turns-when-making-weapon-attack",
            "horde-breaker/optional-another-attack-with-same-weapon",
            "horde-breaker/target-different-creature-within-weapon-range-not-attacked-this-turn",
            "horde-breaker/target-within-feet-of-original-target",
        ],
        {"horde-breaker/target-within-feet-of-original-target": facts["nearby"]},
    )
    defense = feature(
        "defensive-tactics",
        "Defensive Tactics",
        7,
        ["defensive-tactics/gain-one-chosen-option"],
    )
    ch = choice(
        defense,
        "defensive-tactics",
        "Defensive Tactics options",
        ["escape-the-horde", "multiattack-defense"],
    )
    option(
        ch,
        "escape-the-horde",
        "Escape the Horde",
        ["escape-the-horde/opportunity-attacks-against-you-have-disadvantage"],
    )
    option(
        ch,
        "multiattack-defense",
        "Multiattack Defense",
        [
            "multiattack-defense/when-creature-hits-you-with-attack-roll",
            "multiattack-defense/that-creature-disadvantage-on-all-other-attack-rolls-against-you-this-turn",
        ],
    )
    sup = feature(
        "superior-hunters-prey",
        "Superior Hunter’s Prey",
        11,
        [
            "superior-hunters-prey/once-per-turn-when-you-deal-damage-to-marked-creature",
            "superior-hunters-prey/optional-deal-that-spells-extra-damage",
            "superior-hunters-prey/target-different-creature-you-can-see",
            "superior-hunters-prey/target-within-feet-of-first-creature",
        ],
        {
            "superior-hunters-prey/target-within-feet-of-first-creature": facts[
                "visible"
            ]
        },
    )
    pred(
        sup,
        "mark",
        "target-marked-by-your-hunters-mark",
        ["feature/ranger/favored-enemy"],
    )
    pred(
        sup,
        "damage",
        "superior-hunters-prey/use-current-canonical-hunters-mark-extra-damage",
        ["feature/ranger/favored-enemy", "feature/ranger/foe-slayer"],
    )
    feature(
        "superior-hunters-defense",
        "Superior Hunter’s Defense",
        15,
        [
            "superior-hunters-defense/when-you-take-damage",
            "superior-hunters-defense/optional-reaction",
            "superior-hunters-defense/self-resistance-to-triggering-damage-and-any-other-damage-of-same-type",
            "superior-hunters-defense/until-end-of-current-turn",
        ],
    )
    return out


def compare(records, facts=None):
    expected = source_records(facts)
    assert len(records) == len(expected), "p61 record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"p61 semantic mismatch {want['id']}: {actual} != {want}"


def verify(pdf, candidate=None):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    raw = subprocess.check_output(
        [
            "pdftotext",
            "-f",
            "61",
            "-l",
            "61",
            "-x",
            "0",
            "-y",
            "0",
            "-W",
            "305",
            "-H",
            "792",
            "-layout",
            str(pdf),
            "-",
        ],
        text=True,
    )
    text = " ".join(re.sub(r"-\n\s*", "", raw.replace("\u00ad", "")).split())
    clauses = [
        "Ranger Subclass: Hunter",
        "Level 3: Hunter’s Lore",
        "While a creature is marked by your Hunter’s Mark",
        "whether that creature has any Immunities, Resistances, or Vulnerabilities",
        "you know what they are",
        "Level 3: Hunter’s Prey",
        "Colossus Slayer",
        "When you hit a creature with a weapon",
        "damage to the target if it’s missing any of its Hit Points",
        "only once per turn",
        "Horde Breaker",
        "Once on each of your turns",
        "make an attack with a weapon",
        "another attack with the same weapon",
        "different creature",
        "within the weapon’s range",
        "you haven’t attacked this turn",
        "Level 7: Defensive Tactics",
        "Escape the Horde",
        "Opportunity Attacks have Disadvantage against you",
        "Multiattack Defense",
        "When a creature hits you with an attack roll",
        "that creature has Disadvantage on all other attack rolls against you this turn",
        "Level 11: Superior Hunter’s Prey",
        "Once per turn when you deal damage to a creature marked by your Hunter’s Mark",
        "that spell’s extra damage to a different creature",
        "you can see",
        "of the first creature",
        "Level 15: Superior Hunter’s Defense",
        "When you take damage",
        "take a Reaction",
        "give yourself Resistance to that damage",
        "any other damage of the same type until the end of the current turn",
    ]
    assert all(x in text for x in clauses), [x for x in clauses if x not in text]
    assert text.count("one of the following feature options of your choice") == 2
    assert text.count("Whenever you finish a Short or Long Rest") == 2
    assert text.count("replace the chosen option with the other one") == 2
    dice, sides = map(int, re.search(r"extra (\d+)d(\d+) damage", text).groups())
    facts = dict(
        dice=dice,
        sides=sides,
        nearby=int(re.search(r"within (\d+) feet of the original target", text)[1]),
        visible=int(re.search(r"see within (\d+) feet of the first creature", text)[1]),
    )
    actual = json.loads(
        (
            candidate or ROOT / "internal/rules/srd521/testdata/character-hunter.json"
        ).read_text()
    )
    assert actual["schema_version"] == 1 and actual["source"]["sha256"] == PIN
    assert (
        actual["source"]["license"] == "CC-BY-4.0"
        and actual["source"]["changes"]
        and actual["source"]["attribution"]
    )
    compare(actual["records"], facts)
    inv = json.loads(
        (ROOT / "docs/investigations/issue-74-character-inventory.json").read_text()
    )
    identities = [r for r in actual["records"] if r["kind"] in ("subclass", "feature")]
    expected = {
        r["id"]: r["source_page"]
        for r in inv["records"]
        if r["batch_key"] == "subclass-ranger"
    }
    assert len(identities) == len(expected) == 6
    assert {
        s["inventory_id"]: s["page"] for r in identities for s in r["sources"]
    } == expected
    manifest = json.loads(
        (
            ROOT / "internal/rules/srd521/testdata/character-hunter-composition.json"
        ).read_text()
    )
    assert manifest["catalogues"] == ["character-ranger.json", "character-hunter.json"]
    from verify_srd_ranger import verify as verify_ranger

    verify_ranger(pdf)
    print(
        f"p61: six identities and {len(actual['records'])} complete records match independent source facts"
    )


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
    elif len(sys.argv) in (2, 3):
        verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) == 3 else None)
    else:
        raise SystemExit(
            "usage: verify_srd_hunter.py PDF [CANDIDATE] | --controls JSON"
        )
