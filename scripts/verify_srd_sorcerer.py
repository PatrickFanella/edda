#!/usr/bin/env python3
"""Standalone pinned pp64–66 transcription and complete Sorcerer graph comparator.

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
DATA = ROOT / "internal/rules/srd521/testdata/character-sorcerer.json"


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"

    def extract(layout=False):
        return subprocess.run(
            ["pdftotext", "-f", "64", "-l", "66"]
            + (["-layout"] if layout else [])
            + [str(pdf), "-"],
            check=True,
            capture_output=True,
            text=True,
        ).stdout

    layout = extract(True)
    flat = " ".join(
        re.sub(
            r"(?m)^(?:64|65|66|System Reference Document 5\.2\.1)\s*$", "", extract()
        )
        .replace("\u00ad", "")
        .split()
    )
    records, by = [], {}

    def rec(kind, slug, parent=None, level=0, page=66, name=None):
        ident = C + kind + "/sorcerer" + ("/" + slug if slug else "")
        if kind == "class":
            sources = [
                dict(inventory_id="srd-5.2.1:planning/classes/sorcerer", page=64)
            ]
        elif kind == "feature":
            sources = [
                dict(
                    inventory_id="srd-5.2.1:planning/class-features/sorcerer-level-"
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
            k = K + "sorcerer/" + key.replace("_", "-")
            r.setdefault("parameters", {})[k] = value
            r["required_mechanics"].append(k)

    def choice(slug, parent, count, options, clause, *keys):
        ident = rec("choice", slug, parent)
        by[ident].update(select=count, option_keys=[K + k for k in options])
        rule(ident, clause, *(keys or ("sorcerer/" + slug,)))
        return ident

    def pred(slug, owner, key, refs=(), **numbers):
        ident = rec("prerequisite", slug, owner)
        by[ident]["required_mechanics"] = [K + "sorcerer/" + key]
        by[ident]["requirement"] = dict(
            key=K + "sorcerer/" + key,
            record_refs=list(refs),
            parameters={k.replace("_", "-"): v for k, v in numbers.items()},
        )
        return ident

    def resource(slug, parent, values):
        ident = rec("resource", slug, parent)
        by[ident].update(
            capacity=values[0][1],
            values_by_level=[dict(level=n, value=v) for n, v in values],
        )
        rule(ident, "Long Rest", "sorcerer/" + slug + "/long-rest-recover-all")
        return ident

    cl = rec("class", "", name="Sorcerer")
    rule(cl, "Charisma", "sorcerer/primary-ability/charisma")
    rule(
        cl,
        "D6 per Sorcerer level",
        hit_die_sides=int(re.search(r"D(\d+) per Sorcerer level", flat)[1]),
    )
    rule(
        cl,
        "Constitution and Charisma",
        "saving-throws/constitution",
        "saving-throws/charisma",
    )
    rule(cl, "Simple weapons", "weapons/simple")
    rule(cl, "Armor Training None", "sorcerer/armor-training/none")
    entry = choice(
        "entry",
        cl,
        1,
        ["sorcerer/entry/first-class", "sorcerer/entry/multiclass"],
        "Becoming a Sorcerer",
    )
    first = choice(
        "first-class-traits",
        entry,
        1,
        ["sorcerer/all-core-traits-and-level-one-features"],
        "Gain all the traits in the Core Sorcerer Traits table.",
    )
    pred("first-class-entry", first, "first-class-entry-only", refs=(entry, cl))
    multi = choice(
        "multiclass-traits",
        entry,
        1,
        ["sorcerer/hit-die-and-level-one-features"],
        "Gain the Hit Point Die from the Core Sorcerer Traits table.",
    )
    rule(
        multi,
        "multiclassing rules",
        "sorcerer/multiclass/use-multiclass-spell-slot-rules",
    )
    pred("multiclass-entry", multi, "multiclass-entry-only", refs=(entry, cl))
    skills = [
        "arcana",
        "deception",
        "insight",
        "intimidation",
        "persuasion",
        "religion",
    ]
    assert (
        "Choose 2: Arcana, Deception, Insight, Intimidation, Persuasion, or Religion"
        in flat
    )
    choice("skills", first, 2, ["skills/" + x for x in skills], "Choose 2: Arcana,")
    eq = choice(
        "starting-equipment",
        first,
        1,
        ["sorcerer/equipment/a", "sorcerer/equipment/b"],
        "Choose A or B:",
    )
    assert (
        "Spear, 2 Daggers, Arcane Focus (crystal), Dungeoneer’s Pack, and 28 GP; or (B) 50 GP"
        in flat
    )
    bundles = [
        {
            "equipment/spear": 1,
            "equipment/dagger": 2,
            "equipment/arcane-focus-crystal": 1,
            "equipment/dungeoneers-pack": 1,
            "currency/gp": int(re.search(r"Pack, and (\d+) GP", flat)[1]),
        },
        {"currency/gp": int(re.search(r"or \(B\) (\d+) GP", flat)[1])},
    ]
    by[eq]["options"] = [
        dict(
            key=K + "sorcerer/equipment/" + tag,
            quantities={K + k: v for k, v in b.items()},
        )
        for tag, b in zip(("a", "b"), bundles)
    ]
    by[eq]["required_mechanics"] += [K + k for k in bundles[0]]

    # Every cell comes from the extracted PDF; em dashes mean no capacity.
    rows = re.findall(
        r"^\s*(\d+)\s+\+(\d)\s+(.+?)\s{2,}" + r"\s+".join([r"([\d—]+)"] * 12) + r"\s*$",
        layout,
        re.M,
    )
    assert len(rows) == 20, ("table rows", len(rows))
    features = {}
    wrapped = {
        "Spellcasting,": "Spellcasting, Innate Sorcery",
        "Font of Magic,": "Font of Magic, Metamagic",
        "Ability Score": "Ability Score Improvement",
        "Sorcerous": "Sorcerous Restoration",
    }
    for row in rows:
        n, pb, label, *counts = row
        n = int(n)
        label = wrapped.get(label, label)
        lev = rec("level", str(n), cl, level=n, name=f"Sorcerer level {n}")
        rule(lev, "Sorcerer Features", "sorcerer/class-progression")
        refs, keys = [], []
        for title in label.split(", "):
            if title == "—":
                continue
            if title == "Subclass feature":
                key = K + "sorcerer/subclass-feature"
                keys.append(key)
                by[lev]["required_mechanics"].append(key)
            else:
                slug = title.lower().replace(" ", "-")
                if slug not in features:
                    page = (
                        64
                        if slug == "spellcasting"
                        else 65 if slug == "innate-sorcery" else 66
                    )
                    features[slug] = rec("feature", slug, lev, page=page, name=title)
                refs.append(features[slug])
        by[lev]["progression"] = dict(
            proficiency_bonus=int(pb), feature_refs=refs, feature_keys=keys
        )
    f = features
    spell = f["spellcasting"]
    rule(
        spell, "Charisma is your spellcasting ability", "sorcerer/spellcasting/charisma"
    )
    rule(spell, "You can use an Arcane Focus", "sorcerer/spellcasting/arcane-focus")
    rule(
        spell,
        "those spells don’t count against the number of spells you can prepare",
        "sorcerer/spellcasting/always-prepared-excluded-still-sorcerer-spells",
    )
    rule(
        spell,
        "Light, Prestidigitation, Shocking Grasp, and Sorcerous Burst are recommended.",
        *[
            "sorcerer/recommended-cantrip/" + x
            for x in ("light", "prestidigitation", "shocking-grasp", "sorcerous-burst")
        ],
    )
    rule(
        spell,
        "Burning Hands and Detect Magic are recommended.",
        "sorcerer/recommended-spell/burning-hands",
        "sorcerer/recommended-spell/detect-magic",
    )
    cantrips = choice(
        "cantrips",
        spell,
        4,
        ["spells/sorcerer-cantrips"],
        "You know four Sorcerer cantrips",
    )
    by[cantrips]["values_by_level"] = [
        dict(level=int(r[0]), value=int(r[4])) for r in rows
    ]
    rule(
        spell,
        "Whenever you gain a Sorcerer level, you can replace one of your cantrips",
        "sorcerer/spellcasting/level-gain-replace-one-sorcerer-cantrip",
        cantrip_replacements=1,
    )
    prepared = choice(
        "prepared-spells",
        spell,
        2,
        ["spells/sorcerer-list"],
        "choose two level 1 Sorcerer spells.",
        "sorcerer/spellcasting/prepare-to-table-count-on-increase",
    )
    by[prepared]["values_by_level"] = [
        dict(level=int(r[0]), value=int(r[5])) for r in rows
    ]
    pred(
        "prepared-slot-level",
        prepared,
        "spell-level-with-available-slot",
        refs=(spell, prepared),
        minimum_spell_level=1,
    )
    pred(
        "initial-prepared-level",
        prepared,
        "initial-selection-level-one-only",
        refs=(prepared,),
        spell_level=1,
    )
    rule(
        spell,
        "Whenever you gain a Sorcerer level, you can replace one spell on your list",
        "sorcerer/spellcasting/level-gain-replace-one-eligible-sorcerer-spell",
        prepared_replacements=1,
    )
    for i in range(1, 10):
        resource(
            "spell-slot-" + str(i),
            spell,
            [(int(r[0]), int(r[5 + i])) for r in rows if r[5 + i] != "—"],
        )
    innate = f["innate-sorcery"]
    rule(
        innate,
        "As a Bonus Action, you can unleash that magic for 1 minute",
        "sorcerer/innate-sorcery/bonus-action",
        duration_minutes=int(
            re.search(r"unleash that magic for (\d+) minute", flat)[1]
        ),
    )
    rule(
        innate,
        "The spell save DC of your Sorcerer spells increases by 1.",
        "sorcerer/innate-sorcery/sorcerer-spell-save-dc-increase",
        save_dc_increase=int(re.search(r"spells increases by (\d+)", flat)[1]),
    )
    rule(
        innate,
        "You have Advantage on the attack rolls of Sorcerer spells you cast.",
        "sorcerer/innate-sorcery/advantage-sorcerer-spell-attacks",
    )
    assert "You can use this feature twice" in flat
    resource("innate-sorcery", innate, [(n, 2) for n in range(1, 21)])
    font = f["font-of-magic"]
    resource(
        "sorcery-points", font, [(int(r[0]), int(r[3])) for r in rows if r[3] != "—"]
    )
    rule(
        font,
        "You can’t have more Sorcery Points than the number shown in the table for your level.",
        "sorcerer/font-of-magic/capped-at-table-maximum",
    )
    rule(
        font,
        "expend a spell slot to gain a number of Sorcery Points equal to the slot’s level (no action required)",
        "sorcerer/font-of-magic/expend-slot-gain-slot-level-points-no-action",
    )
    rule(
        font,
        "As a Bonus Action, you can transform unexpended Sorcery Points into one spell slot.",
        "sorcerer/font-of-magic/bonus-action-spend-unexpended-points-create-slot",
        created_slots_per_conversion=1,
    )
    rule(
        font,
        "You can create a spell slot no higher than level 5.",
        "sorcerer/font-of-magic/maximum-created-slot-level",
        maximum_created_slot_level=int(
            re.search(r"slot no higher than level (\d+)", flat)[1]
        ),
    )
    rule(
        font,
        "Any spell slot you create with this feature vanishes when you finish a Long Rest.",
        "sorcerer/font-of-magic/created-slot-vanishes-long-rest",
    )
    conversion = re.findall(
        r"^\s*([1-5])\s+([2-7])\s+([2-9])(?:\s{2,}|$)", layout, re.M
    )
    assert len(conversion) == 5
    for slot, cost, minimum in conversion:
        rule(
            font,
            "Creating Spell Slots",
            **{
                f"create_slot_{slot}_sorcery_point_cost": int(cost),
                f"create_slot_{slot}_minimum_sorcerer_level": int(minimum),
            },
        )
    meta = f["metamagic"]
    options = choice(
        "metamagic",
        meta,
        2,
        ["metamagic/options"],
        "you gain two Metamagic options of your choice",
    )
    by[options]["values_by_level"] = [
        dict(level=n, value=2 if n < 10 else 4 if n < 17 else 6) for n in range(2, 21)
    ]
    rule(
        meta,
        "You use the chosen options to temporarily modify spells you cast.",
        "sorcerer/metamagic/chosen-options-modify-own-cast-spells",
    )
    rule(
        meta,
        "you must spend the number of Sorcery Points that it costs.",
        "sorcerer/metamagic/spend-option-sorcery-point-cost",
    )
    rule(
        meta,
        "only one Metamagic option on a spell when you cast it unless otherwise noted in one of those options.",
        "sorcerer/metamagic/one-per-spell-unless-option-exception",
        options_per_spell=1,
    )
    rule(
        meta,
        "replace one of your Metamagic options with one you don’t know.",
        "sorcerer/metamagic/level-gain-replace-one-with-unknown-option",
        replacements=1,
    )
    assert (
        "two more options at Sorcerer level 10 and two more at Sorcerer level 17"
        in flat
    )
    sub = choice(
        "subclass",
        f["sorcerer-subclass"],
        1,
        ["subclasses/sorcerer", "subclasses/sorcerer/draconic-sorcery"],
        "Draconic Sorcery subclass",
    )
    rule(
        sub,
        "Sorcerer level or lower",
        "sorcerer/subclass/features-at-own-level-or-lower",
    )
    for slug, domains in [
        (
            "ability-score-improvement",
            ["feats/ability-score-improvement", "feats/qualified"],
        ),
        ("epic-boon", ["feats/epic-boon", "feats/qualified"]),
    ]:
        ch = choice(
            slug,
            f[slug],
            1,
            domains,
            "another feat of your choice for which you qualify.",
        )
        pred(slug + "-qualified", ch, "selected-feat-qualified", refs=(ch,))
    rule(
        f["epic-boon"],
        "Boon of Dimensional Travel is recommended.",
        "sorcerer/recommended-feat/boon-of-dimensional-travel",
    )
    restore = f["sorcerous-restoration"]
    rule(
        restore,
        "When you finish a Short Rest, you can regain expended Sorcery Points",
        "sorcerer/restoration/optional-short-rest-recover-expended-points",
    )
    rule(
        restore,
        "half your Sorcerer level (round down)",
        "sorcerer/restoration/maximum-half-sorcerer-level-round-down",
        recovery_level_divisor=2,
    )
    resource("sorcerous-restoration", restore, [(n, 1) for n in range(5, 21)])
    rule(
        restore,
        "you can’t do so again until you finish a Long Rest.",
        "sorcerer/restoration/once-until-long-rest",
    )
    inc = f["sorcery-incarnate"]
    rule(
        inc,
        "spend 2 Sorcery Points when you take the Bonus Action to activate it.",
        "sorcerer/incarnate/spend-points-bonus-action-activate-innate-sorcery",
        sorcery_point_cost=int(re.search(r"spend (\d+) Sorcery Points when", flat)[1]),
    )
    pred("incarnate-no-uses", inc, "no-innate-sorcery-uses-left", refs=(innate,))
    rule(
        inc,
        "while your Innate Sorcery feature is active, you can use up to two of your Metamagic options on each spell you cast.",
        "sorcerer/incarnate/innate-active-two-metamagic-options-per-spell",
        options_per_spell=2,
    )
    pred("incarnate-active", inc, "innate-sorcery-active", refs=(innate, meta))
    apo = f["arcane-apotheosis"]
    rule(
        apo,
        "While your Innate Sorcery feature is active, you can use one Metamagic option on each of your turns without spending Sorcery Points on it.",
        "sorcerer/apotheosis/waive-one-option-point-cost-each-own-turn",
        free_options_per_own_turn=1,
    )
    pred("apotheosis-active", apo, "innate-sorcery-active", refs=(innate, meta))
    for slug, ident in f.items():
        by[ident]["required_mechanics"].append(K + "sorcerer/feature/" + slug)
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
        if r["id"] == "srd-5.2.1:planning/classes/sorcerer"
        or r["id"].startswith("srd-5.2.1:planning/class-features/sorcerer-level-")
    }
    actual = {
        s["inventory_id"]: s["page"]
        for r in c["records"]
        if r["kind"] in ("class", "feature")
        for s in r["sources"]
    }
    assert (
        len(actual) == len(identities) == 11 and actual == identities
    ), "identities/pages"
    compare(c["records"], expected)

    controls = [
        (
            "prerequisite/incarnate-no-uses",
            "requirement",
            "record_refs",
            [C + "feature/sorcerer/spellcasting"],
        ),
        (
            "choice/starting-equipment",
            None,
            "option_keys",
            [K + "sorcerer/equipment/a"],
        ),
        (
            "prerequisite/apotheosis-active",
            "requirement",
            "key",
            K + "sorcerer/innate-sorcery-inactive",
        ),
        (
            "resource/innate-sorcery",
            None,
            "required_mechanics",
            [K + "sorcerer/innate-sorcery/short-rest-recover-all"],
        ),
        (
            "feature/sorcery-incarnate",
            "parameters",
            K + "sorcerer/sorcery-point-cost",
            3,
        ),
        ("feature/innate-sorcery", "parameters", K + "sorcerer/duration-minutes", 10),
        (
            "feature/font-of-magic",
            "parameters",
            K + "sorcerer/create-slot-3-sorcery-point-cost",
            4,
        ),
        ("resource/spell-slot-9", "values_by_level", 0, dict(level=17, value=2)),
        (
            "feature/font-of-magic",
            "parameters",
            K + "sorcerer/create-slot-5-minimum-sorcerer-level",
            7,
        ),
        (
            "feature/arcane-apotheosis",
            "parameters",
            K + "sorcerer/invented-level-scaling",
            2,
        ),
    ]
    for suffix, field, key, value in controls:
        bad = copy.deepcopy(expected)
        kind, slug = suffix.split("/")
        target = next(r for r in bad if r["id"] == C + kind + "/sorcerer/" + slug)
        (target[field] if field else target)[key] = value
        if field == "requirement" and key == "key":
            target["required_mechanics"] = [value]
        if field == "parameters" and key not in target["required_mechanics"]:
            target["required_mechanics"].append(key)
        if suffix == "resource/spell-slot-9":
            target["capacity"] = 2
        if suffix == "choice/starting-equipment":
            target["options"] = target["options"][:1]
        try:
            compare(bad, expected)
        except AssertionError:
            continue
        raise AssertionError("accepted contradiction: " + suffix)
    bad = copy.deepcopy(expected)
    missing = C + "prerequisite/sorcerer/apotheosis-active"
    bad = [r for r in bad if r["id"] != missing]
    next(r for r in bad if r["id"] == C + "feature/sorcerer/arcane-apotheosis")[
        "prerequisite_refs"
    ] = []
    try:
        compare(bad, expected)
    except AssertionError:
        pass
    else:
        raise AssertionError("accepted missing active condition")
    print(
        f"Sorcerer: {len(expected)} graph records, 11 identities, 20 progression rows, 9 slot columns and 5 conversion rows match pinned pp64–66; 11 semantic controls rejected"
    )


if __name__ == "__main__":
    if len(sys.argv) not in (2, 3):
        raise SystemExit("usage: verify_srd_sorcerer.py PINNED.pdf [catalogue.json]")
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) == 3 else DATA)
