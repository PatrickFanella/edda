#!/usr/bin/env python3
"""Standalone pinned pp57–59 transcription and complete Ranger graph comparator.

No Go helpers, candidate-derived expectations, hashes of records, or runtime
consumers. Prose clauses bind semantic fields; table cells come from the PDF.
"""

import copy
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
C = "srd-5.2.1:character/"
K = "srd-5.2.1:mechanic/"
DATA = ROOT / "internal/rules/srd521/testdata/character-ranger.json"


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"

    def extract(layout=False):
        return subprocess.run(
            ["pdftotext", "-f", "57", "-l", "59"]
            + (["-layout"] if layout else [])
            + [str(pdf), "-"],
            check=True,
            capture_output=True,
            text=True,
        ).stdout

    layout = extract(True)
    flat = " ".join(extract().replace("\u00ad", "").split())
    records, by = [], {}

    def rec(kind, slug, parent=None, level=0, page=59, name=None):
        ident = C + kind + "/ranger" + ("/" + slug if slug else "")
        if kind == "class":
            sources = [dict(inventory_id="srd-5.2.1:planning/classes/ranger", page=57)]
        elif kind == "feature":
            sources = [
                dict(
                    inventory_id="srd-5.2.1:planning/class-features/ranger-level-"
                    + str(by[parent]["level"])
                    + "-"
                    + slug,
                    page=page,
                )
            ]
        else:
            sources = copy.deepcopy(by[parent]["sources"])
        r = dict(
            id=ident,
            kind=kind,
            name=name or slug.replace("-", " ").title(),
            sources=sources,
            parent_id=parent or "",
            level=level,
            choice_refs=[],
            resource_refs=[],
            prerequisite_refs=[],
            required_mechanics=[],
            select=0,
            option_keys=[],
            capacity=0,
        )
        records.append(r)
        by[ident] = r
        if kind in ("choice", "resource", "prerequisite"):
            by[parent][kind + "_refs"].append(ident)
        return ident

    def rule(owner, clause, *keys, **numbers):
        assert clause in flat, "source clause: " + clause
        r = by[owner]
        r["required_mechanics"] += [K + k for k in keys]
        for key, value in numbers.items():
            k = K + "ranger/" + key.replace("_", "-")
            r.setdefault("parameters", {})[k] = value
            r["required_mechanics"].append(k)

    def choice(slug, parent, count, options, clause, *keys):
        ident = rec("choice", slug, parent)
        by[ident].update(select=count, option_keys=[K + k for k in options])
        rule(ident, clause, *(keys or ("ranger/" + slug,)))
        return ident

    def pred(slug, owner, key, refs=(), **numbers):
        ident = rec("prerequisite", slug, owner)
        by[ident]["required_mechanics"] = [K + "ranger/" + key]
        by[ident]["requirement"] = dict(
            key=K + "ranger/" + key,
            record_refs=list(refs),
            parameters={k.replace("_", "-"): v for k, v in numbers.items()},
        )
        return ident

    def resource(slug, parent, values=None, symbolic=False):
        ident = rec("resource", slug, parent)
        keys = ["ranger/" + slug + "/long-rest-recover-all"]
        if symbolic:
            by[ident]["capacity_mechanic"] = K + "ranger/wisdom-modifier-minimum-one"
            keys += ["ranger/wisdom-modifier-minimum-one"]
        else:
            by[ident].update(
                capacity=values[0][1],
                values_by_level=[dict(level=n, value=v) for n, v in values],
            )
        rule(ident, "Long Rest", *keys)
        return ident

    cl = rec("class", "", name="Ranger")
    rule(
        cl,
        "Dexterity and Wisdom",
        "ranger/primary-ability/dexterity",
        "ranger/primary-ability/wisdom",
    )
    rule(
        cl,
        "D10 per Ranger level",
        hit_die_sides=int(re.search(r"D(\d+) per Ranger level", flat)[1]),
    )
    rule(
        cl,
        "Strength and Dexterity",
        "saving-throws/strength",
        "saving-throws/dexterity",
    )
    rule(cl, "Simple and Martial weapons", "weapons/simple", "weapons/martial")
    rule(
        cl,
        "Light and Medium armor and Shields",
        "armor/light",
        "armor/medium",
        "armor/shields",
    )
    rule(
        cl,
        "Gain all the traits in the Core Ranger Traits table.",
        "ranger/level-one/all-core-traits-and-features",
    )
    rule(
        cl,
        "Hit Point Die, proficiency with Martial weapons, proficiency in one skill of your choice from the Ranger’s skill list",
        "ranger/multiclass/hit-die-martial-weapons-one-class-skill-light-medium-shields",
        "ranger/multiclass/level-one-features-multiclass-slot-rules",
    )
    skills = [
        "animal-handling",
        "athletics",
        "insight",
        "investigation",
        "nature",
        "perception",
        "stealth",
        "survival",
    ]
    assert (
        "Choose 3: Animal Handling, Athletics, Insight, Investigation, Nature, Perception, Stealth, or Survival"
        in flat
    )
    choice(
        "skills", cl, 3, ["skills/" + s for s in skills], "Choose 3: Animal Handling,"
    )
    mc = choice(
        "multiclass-skill",
        cl,
        1,
        ["skills/" + s for s in skills],
        "proficiency in one skill of your choice from the Ranger’s skill list",
    )
    pred("multiclass-entry", mc, "multiclass-entry-only", refs=(cl, mc))
    eq = choice(
        "starting-equipment",
        cl,
        1,
        ["ranger/equipment/a", "ranger/equipment/b"],
        "or (B) 150 GP",
    )
    assert (
        "Leather Armor, Scimitar, Shortsword, Longbow, 20 Arrows, Quiver, Druidic Focus (sprig of mistletoe), Explorer’s Pack, and 7 GP"
        in flat
    )
    bundles = [
        {
            "equipment/" + s: n
            for s, n in [
                ("studded-leather-armor", 1),
                ("scimitar", 1),
                ("shortsword", 1),
                ("longbow", 1),
                ("arrows", 20),
                ("quiver", 1),
                ("druidic-focus-sprig-of-mistletoe", 1),
                ("explorers-pack", 1),
            ]
        },
        {},
    ]
    bundles[0]["currency/gp"] = int(re.search(r"Pack, and (\d+) GP", flat)[1])
    bundles[1]["currency/gp"] = int(re.search(r"or \(B\) (\d+) GP", flat)[1])
    by[eq]["options"] = [
        dict(
            key=K + "ranger/equipment/" + tag,
            quantities={K + k: v for k, v in b.items()},
        )
        for tag, b in zip(("a", "b"), bundles)
    ]
    by[eq]["required_mechanics"] += [K + k for k in bundles[0]]

    # Extract every numeric table column. Wrapped level-one feature text is
    # restored before matching; no counts are taken from the candidate graph.
    table = re.sub(
        r"(\n\s*1\s+\+2\s+Spellcasting, Favored Enemy,)([^\n]*)\n\s*Weapon Mastery",
        r"\1 Weapon Mastery\2",
        layout,
    )
    rows = re.findall(
        r"^\s*(\d+)\s+\+(\d)\s+(.+?)\s{2,}(\d+)\s+(\d+)\s+(\d+)\s+([\d—]+)\s+([\d—]+)\s+([\d—]+)\s+([\d—]+)\s*$",
        table,
        re.M,
    )
    assert len(rows) == 20, ("table rows", len(rows))
    features = {}
    for row in rows:
        n, pb, label, *counts = row
        n = int(n)
        lev = rec("level", str(n), cl, level=n, name=f"Ranger level {n}")
        rule(lev, "Ranger Features", "ranger/class-progression")
        refs, keys = [], []
        for title in label.split(", "):
            if title == "Subclass feature":
                key = K + "ranger/subclass-feature"
                keys.append(key)
                by[lev]["required_mechanics"].append(key)
            else:
                slug = title.lower().replace("’", "").replace(" ", "-")
                if slug not in features:
                    page = (
                        57
                        if slug == "spellcasting"
                        else 58
                        if slug in ("favored-enemy", "weapon-mastery")
                        else 59
                    )
                    features[slug] = rec("feature", slug, lev, page=page, name=title)
                refs.append(features[slug])
        by[lev]["progression"] = dict(
            proficiency_bonus=int(pb), feature_refs=refs, feature_keys=keys
        )
    f = features
    spell = f["spellcasting"]
    rule(spell, "Wisdom is your spellcasting ability", "ranger/spellcasting/wisdom")
    rule(spell, "You can use a Druidic Focus", "ranger/spellcasting/druidic-focus")
    rule(
        spell,
        "those spells don’t count against the number of spells you can prepare",
        "ranger/spellcasting/always-prepared-excluded-still-ranger-spells",
    )
    rule(
        spell,
        "Cure Wounds and Ensnaring Strike are recommended.",
        "ranger/recommended-spell/cure-wounds",
        "ranger/recommended-spell/ensnaring-strike",
    )
    prepared = choice(
        "prepared-spells",
        spell,
        2,
        ["spells/ranger-list"],
        "choose two level 1 Ranger spells.",
        "ranger/spellcasting/prepare-to-table-count-on-increase",
    )
    by[prepared]["values_by_level"] = [
        dict(level=int(r[0]), value=int(r[4])) for r in rows
    ]
    pred(
        "prepared-slot-level",
        prepared,
        "spell-level-with-available-slot",
        refs=(spell, prepared),
        minimum_spell_level=1,
    )
    rule(
        spell,
        "Whenever you finish a Long Rest, you can replace one spell on your list with another Ranger spell for which you have spell slots.",
        "ranger/spellcasting/long-rest-replace-one-eligible-spell",
        prepared_replacements=1,
    )
    for i in range(1, 6):
        resource(
            "spell-slot-" + str(i),
            spell,
            [(int(r[0]), int(r[4 + i])) for r in rows if r[4 + i] != "—"],
        )
    enemy = f["favored-enemy"]
    rule(
        enemy,
        "You always have the Hunter’s Mark spell prepared.",
        "ranger/favored-enemy/source-5-2-1-always-prepared-hunters-mark",
        "spells/hunters-mark",
    )
    by[enemy]["always_prepared"] = [
        dict(level=1, spell_keys=[K + "spells/hunters-mark"])
    ]
    rule(
        enemy,
        "You can cast it twice without expending a spell slot",
        "ranger/favored-enemy/no-slot-cast",
        uses_per_cast=1,
    )
    resource("favored-enemy", enemy, [(int(r[0]), int(r[3])) for r in rows])
    mastery = choice(
        "weapon-mastery",
        f["weapon-mastery"],
        2,
        ["weapons/proficient-kinds"],
        "mastery properties of two kinds of weapons",
    )
    pred("mastery-proficiency", mastery, "selected-weapon-proficient", refs=(mastery,))
    rule(
        f["weapon-mastery"],
        "Whenever you finish a Long Rest, you can change the kinds of weapons you chose.",
        "ranger/mastery/long-rest-replace-kinds",
    )
    for slug, count, clause in [
        ("deft-explorer", 1, "Choose one of your skill proficiencies"),
        ("expertise", 2, "Choose two of your skill proficiencies"),
    ]:
        ch = choice(
            slug + "-expertise",
            f[slug],
            count,
            ["skills/proficient-without-expertise"],
            clause,
        )
        pred(
            slug + "-eligibility", ch, "proficient-skill-lacking-expertise", refs=(ch,)
        )
    choice(
        "languages",
        f["deft-explorer"],
        2,
        ["languages/character-creation-tables"],
        "You know two languages of your choice",
    )
    style = choice(
        "fighting-style",
        f["fighting-style"],
        1,
        ["feats/fighting-style", "ranger/druidic-warrior"],
        "Instead of choosing one of those feats",
    )
    cantrip = choice(
        "druidic-warrior-cantrips",
        style,
        2,
        ["spells/druid-cantrips"],
        "You learn two Druid cantrips of your choice",
    )
    pred("druidic-warrior-selected", cantrip, "druidic-warrior-selected", refs=(style,))
    rule(
        cantrip,
        "Wisdom is your spellcasting ability for them.",
        "ranger/druidic-warrior/count-as-ranger-spells-wisdom",
    )
    rule(
        f["fighting-style"],
        "Whenever you gain a Ranger level, you can replace one of these cantrips with another Druid cantrip.",
        "ranger/druidic-warrior/replace-one-on-ranger-level",
        cantrip_replacements=1,
    )
    rule(
        cantrip,
        "Guidance and Starry Wisp are recommended.",
        "ranger/recommended-cantrip/guidance",
        "ranger/recommended-cantrip/starry-wisp",
    )
    choice(
        "subclass",
        f["ranger-subclass"],
        1,
        ["subclasses/ranger", "subclasses/ranger/hunter"],
        "The Hunter subclass is detailed after this class’s description.",
    )
    rule(
        f["ranger-subclass"],
        "features that are of your Ranger level or lower.",
        "ranger/subclass/features-at-or-below-ranger-level",
    )
    for slug, options, clause in [
        (
            "ability-score-improvement",
            ["feats/ability-score-improvement", "feats/qualified"],
            "You gain the Ability Score Improvement feat",
        ),
        (
            "epic-boon",
            ["feats/epic-boon", "feats/qualified"],
            "You gain an Epic Boon feat",
        ),
    ]:
        ch = choice(slug, f[slug], 1, options, clause)
        pred(slug + "-qualified", ch, "selected-feat-qualified", refs=(ch,))
    rule(
        f["ability-score-improvement"],
        "You gain this feature again at Ranger levels 8, 12, and 16.",
        "ranger/asi-repeat-at-8-12-16",
    )
    rule(
        f["epic-boon"],
        "Boon of Dimensional Travel is recommended.",
        "ranger/recommended-feat/boon-of-dimensional-travel",
    )
    rule(
        f["extra-attack"],
        "attack twice instead of once whenever you take the Attack action on your turn.",
        "ranger/extra-attack/own-turn-attack-action",
        attacks=2,
    )
    rule(
        f["roving"],
        "Your Speed increases by 10 feet while you aren’t wearing Heavy armor.",
        "ranger/roving/speed-increase-when-not-heavy-armor",
        speed_bonus_feet=int(re.search(r"Speed increases by (\d+) feet", flat)[1]),
    )
    pred("roving-armor", f["roving"], "not-wearing-heavy-armor", refs=(f["roving"],))
    rule(
        f["roving"],
        "Climb Speed and a Swim Speed equal to your Speed.",
        "ranger/roving/climb-and-swim-equal-speed",
    )
    tire = f["tireless"]
    rule(tire, "As a Magic action", "ranger/tireless/magic-action-self-temporary-hp")
    rule(
        tire,
        "1d8 plus your Wisdom modifier (minimum of 1)",
        "ranger/tireless/temp-hp-die-plus-wisdom-minimum-one",
        temporary_hp_dice=1,
        temporary_hp_die_sides=int(re.search(r"1d(\d+) plus your Wisdom", flat)[1]),
        temporary_hp_minimum=1,
    )
    resource("tireless", tire, symbolic=True)
    rule(
        tire,
        "Whenever you finish a Short Rest, your Exhaustion level, if any, decreases by 1.",
        "ranger/tireless/short-rest-decrease-exhaustion-if-any",
        exhaustion_reduction=1,
    )
    pred(
        "tireless-exhaustion",
        tire,
        "exhaustion-present",
        refs=(tire,),
        minimum_exhaustion=1,
    )
    rule(
        f["relentless-hunter"],
        "Taking damage can’t break your Concentration on Hunter’s Mark.",
        "ranger/relentless-hunter/damage-cannot-break-hunters-mark-concentration",
    )
    pred(
        "relentless-mark",
        f["relentless-hunter"],
        "concentrating-on-hunters-mark",
        refs=(enemy,),
    )
    veil = f["natures-veil"]
    rule(
        veil,
        "As a Bonus Action, you can give yourself the Invisible condition until the end of your next turn.",
        "ranger/veil/bonus-action-self-invisible-until-end-next-turn",
    )
    resource("natures-veil", veil, symbolic=True)
    rule(
        f["precise-hunter"],
        "Advantage on attack rolls against the creature currently marked by your Hunter’s Mark.",
        "ranger/precise-hunter/advantage-on-attacks-against-own-current-mark",
    )
    pred(
        "precise-mark",
        f["precise-hunter"],
        "target-is-own-current-hunters-mark",
        refs=(enemy,),
    )
    rule(
        f["feral-senses"],
        "Blindsight with a range of 30 feet.",
        "ranger/feral-senses/blindsight",
        blindsight_feet=int(
            re.search(r"Blindsight with a range of (\d+) feet", flat)[1]
        ),
    )
    rule(
        f["foe-slayer"],
        "Hunter’s Mark is a d10 rather than a d6.",
        "ranger/foe-slayer/replace-hunters-mark-damage-die",
        original_die_sides=6,
        replacement_die_sides=10,
    )
    pred("foe-slayer-mark", f["foe-slayer"], "hunters-mark-damage", refs=(enemy,))
    # Feature headings are unavailable requirements even when all details live
    # in their nested choices. This does not qualify any implementation.
    for slug, ident in features.items():
        by[ident]["required_mechanics"].append(K + "ranger/feature/" + slug)
    assert (
        "You can use this action a number of times equal to your Wisdom modifier (minimum of once)"
        in flat
    )
    assert (
        "You can use this feature a number of times equal to your Wisdom modifier (minimum of once)"
        in flat
    )
    return records


def compare(actual, expected, path="records"):
    assert type(actual) is type(expected), path + ": type"
    if isinstance(expected, dict):
        assert actual.keys() == expected.keys(), path + ": fields"
        for k in expected:
            compare(actual[k], expected[k], path + "/" + k)
    elif isinstance(expected, list):
        assert len(actual) == len(expected), path + ": length"
        for i, (a, e) in enumerate(zip(actual, expected)):
            compare(a, e, f"{path}/{i}")
    else:
        assert actual == expected, f"{path}: {actual!r} != source {expected!r}"


def verify(pdf, catalogue=DATA):
    expected = source_records(pdf)
    c = json.loads(catalogue.read_text())
    assert c["schema_version"] == 1 and c["source"]["sha256"] == PIN
    assert (
        c["source"]["license"] == "CC-BY-4.0"
        and c["source"]["attribution"]
        and c["source"]["changes"]
    )
    inventory = json.loads(
        (ROOT / "docs/investigations/issue-74-character-inventory.json").read_text()
    )
    identities = {
        r["id"]: r["source_page"]
        for r in inventory["records"]
        if r["id"] == "srd-5.2.1:planning/classes/ranger"
        or r["id"].startswith("srd-5.2.1:planning/class-features/ranger-level-")
    }
    actual = {
        s["inventory_id"]: s["page"]
        for r in c["records"]
        if r["kind"] in ("class", "feature")
        for s in r["sources"]
    }
    assert len(actual) == len(identities) == 18 and actual == identities, (
        "identities/pages"
    )
    compare(c["records"], expected)
    controls = [
        (
            "prerequisite/precise-mark",
            "requirement",
            "record_refs",
            [C + "feature/ranger/spellcasting"],
        ),
        (
            "prerequisite/roving-armor",
            "requirement",
            "key",
            K + "ranger/wearing-heavy-armor",
        ),
        (
            "resource/tireless",
            None,
            "required_mechanics",
            [
                K + "ranger/tireless/short-rest-recover-all",
                K + "ranger/wisdom-modifier-minimum-one",
            ],
        ),
        ("feature/favored-enemy", "parameters", K + "ranger/uses-per-cast", 2),
        ("feature/feral-senses", "parameters", K + "ranger/blindsight-feet", 60),
        ("resource/favored-enemy", "values_by_level", 1, dict(level=2, value=3)),
        (
            "feature/favored-enemy",
            None,
            "required_mechanics",
            [
                K + "ranger/favored-enemy/2014-creature-type-variant",
                K + "spells/hunters-mark",
                K + "ranger/uses-per-cast",
            ],
        ),
        ("feature/foe-slayer", "parameters", K + "ranger/invented-level-scaling", 2),
    ]
    for suffix, field, key, value in controls:
        bad = copy.deepcopy(expected)
        kind, slug = suffix.split("/")
        target = next(r for r in bad if r["id"] == C + kind + "/ranger/" + slug)
        (target[field] if field else target)[key] = value
        if field == "requirement" and key == "key":
            target["required_mechanics"] = [value]
        if field == "parameters" and key not in target["required_mechanics"]:
            target["required_mechanics"].append(key)
        try:
            compare(bad, expected)
        except AssertionError:
            continue
        raise AssertionError("accepted contrary source: " + suffix)
    print(
        f"Ranger: {len(expected)} records, 18 identities, all 20 table rows and all graph fields match pinned pp57–59; 8 contrary-source controls rejected"
    )


if __name__ == "__main__":
    if len(sys.argv) not in (2, 3):
        raise SystemExit("usage: verify_srd_ranger.py PINNED.pdf [catalogue.json]")
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) == 3 else DATA)
