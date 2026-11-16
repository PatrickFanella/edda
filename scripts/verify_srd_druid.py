#!/usr/bin/env python3
"""Offline semantic comparison of the Druid catalogue with pinned PDF pp41–43.

The expected graph is a source transcription, not loaded from the catalogue.
Each prose-to-mechanic mapping below requires its source clause; table cells
are extracted from the PDF. Every serialized field is compared, with paths.
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
DATA = ROOT / "internal/rules/srd521/testdata/character-druid.json"


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    text = subprocess.run(
        ["pdftotext", "-f", "41", "-l", "43", "-layout", str(pdf), "-"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout
    prose = subprocess.run(
        ["pdftotext", "-f", "41", "-l", "43", str(pdf), "-"],
        check=True,
        capture_output=True,
        text=True,
    ).stdout
    flat = " ".join((prose + text).replace("\u00ad", "").split())
    records = []
    by = {}

    def record(kind, slug, parent=None, level=0, page=None, name=None):
        ident = f"{C}{kind}/druid" + (f"/{slug}" if slug else "")
        if kind == "class":
            sources = [{"inventory_id": "srd-5.2.1:planning/classes/druid", "page": 41}]
        elif kind == "feature":
            sources = [
                {
                    "inventory_id": f"srd-5.2.1:planning/class-features/druid-level-{by[parent]['level']}-{slug}",
                    "page": page,
                }
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
        # Checking the clause alone is insufficient: attach its semantic keys and
        # numeric operands to the expected record that is compared below.
        assert clause in flat, f"source clause missing: {clause}"
        r = by[owner]
        r["required_mechanics"] += [K + k for k in keys]
        for key, value in numbers.items():
            k = K + "druid/" + key.replace("_", "-")
            r.setdefault("parameters", {})[k] = value
            r["required_mechanics"].append(k)

    def choice(slug, parent, count, options, clause, *keys):
        ident = record("choice", slug, parent)
        by[ident].update(select=count, option_keys=[K + k for k in options])
        rule(ident, clause, *keys or ("druid/" + slug,))
        return ident

    def predicate(slug, parent, key, refs=(), **parameters):
        ident = record("prerequisite", slug, parent)
        by[ident]["required_mechanics"] = [K + key]
        by[ident]["requirement"] = dict(
            key=K + key, record_refs=list(refs), parameters=parameters
        )
        return ident

    def resource(slug, parent, values, clause, *keys):
        ident = record("resource", slug, parent)
        by[ident].update(
            capacity=values[0][1],
            values_by_level=[dict(level=n, value=v) for n, v in values],
        )
        rule(ident, clause, *keys)
        return ident

    cl = record("class", "", name="Druid")
    rule(cl, "Primary Ability Wisdom", "druid/primary-ability/wisdom")
    rule(cl, "D8 per Druid level", hit_die_sides=8)
    rule(
        cl,
        "Intelligence and Wisdom",
        "saving-throws/intelligence",
        "saving-throws/wisdom",
    )
    rule(cl, "Simple weapons", "weapons/simple")
    rule(cl, "Herbalism Kit", "tools/herbalism-kit")
    rule(cl, "Light armor and Shields", "armor/light", "armor/shields")
    rule(
        cl,
        "Gain the following traits from the Core Druid",
        "druid/multiclass/hit-die-light-armor-shields-only",
        "druid/multiclass/level-1-features-and-multiclass-spell-slots",
        "druid/level-1/all-core-traits-and-features",
    )
    skills = choice(
        "skills",
        cl,
        2,
        [
            "skills/" + s
            for s in [
                "animal-handling",
                "arcana",
                "insight",
                "medicine",
                "nature",
                "perception",
                "religion",
                "survival",
            ]
        ],
        "Choose 2: Animal Handling,",
    )
    rule(skills, "Arcana, Insight, Medicine, Na-", "druid/skills/source-eight-options")
    eq = choice(
        "starting-equipment",
        cl,
        1,
        ["druid/equipment/a", "druid/equipment/b"],
        "or (B) 50 GP",
    )
    bundles = [
        {
            "equipment/leather-armor": 1,
            "equipment/shield": 1,
            "equipment/sickle": 1,
            "equipment/druidic-focus-quarterstaff": 1,
            "equipment/explorers-pack": 1,
            "equipment/herbalism-kit": 1,
            "currency/gp": 9,
        },
        {"currency/gp": 50},
    ]
    by[eq]["options"] = [
        dict(
            key=K + "druid/equipment/" + s, quantities={K + k: v for k, v in b.items()}
        )
        for s, b in zip(["a", "b"], bundles)
    ]
    rule(
        eq,
        "Pack, Herbalism Kit, and 9 GP;",
        *dict.fromkeys(k for b in bundles for k in b),
    )
    features = {
        1: ["spellcasting", "druidic", "primal-order"],
        2: ["wild-shape", "wild-companion"],
        3: ["druid-subclass"],
        4: ["ability-score-improvement"],
        5: ["wild-resurgence"],
        7: ["elemental-fury"],
        15: ["improved-elemental-fury"],
        18: ["beast-spells"],
        19: ["epic-boon"],
        20: ["archdruid"],
    }
    fs = {}
    for n in range(1, 21):
        row = record("level", str(n), cl, n, name=f"Druid level {n}")
        rule(row, "Druid Features", "druid/class-progression")
        for slug in features.get(n, []):
            page = (
                41
                if slug == "spellcasting"
                else 42
                if slug in ("druidic", "primal-order", "wild-shape")
                else 43
            )
            fs[slug] = record(
                "feature",
                slug,
                row,
                page=page,
                name=slug.replace("-", " ")
                .title()
                .replace("Ability Score Improvement", "Ability Score Improvement"),
            )
    # Wrapped feature names are normalized before extracting all 20 rows.
    table = text.split("Druid Features\n")[-1].split("41    System")[0]
    table = re.sub(
        r"Spellcasting, Druidic,\s+([—\d].*?)\n\s+Primal Order",
        r"Spellcasting, Druidic, Primal Order  \1",
        table,
    )
    table = re.sub(
        r"Wild Shape,\s+([—\d].*?)\n\s+Wild Companion",
        r"Wild Shape, Wild Companion  \1",
        table,
    )
    rows = re.findall(
        r"^\s+(\d{1,2})\s+\+(\d)\s+(.+?)\s{2,}([\d—][\d—\s]+)$", table, re.M
    )
    assert [int(r[0]) for r in rows] == list(range(1, 21)), "progression extraction"
    cells = []
    for n, bonus, names, suffix in rows:
        n = int(n)
        nums = [0 if v == "—" else int(v) for v in suffix.split()]
        assert len(nums) == 12, (n, nums)
        cells.append(nums)
        refs = [fs[s] for s in features.get(n, [])]
        if n in (8, 12, 16):
            refs = [fs["ability-score-improvement"]]
        keys = [K + "druid/subclass-feature"] if n in (6, 10, 14) else []
        r = by[f"{C}level/druid/{n}"]
        r["progression"] = dict(
            proficiency_bonus=int(bonus), feature_refs=refs, feature_keys=keys
        )
        r["required_mechanics"] += keys
        expected_names = [by[x]["name"] for x in refs] + [
            "Subclass feature" for _ in keys
        ]
        assert (", ".join(expected_names) or "—") == names, (n, names, expected_names)
    sp = fs["spellcasting"]
    rule(sp, "Wisdom is your spellcasting", "druid/spellcasting/wisdom")
    rule(
        sp,
        "as a Spellcasting Focus for your Druid spells",
        "druid/spellcasting/druidic-focus",
    )
    rule(
        sp,
        "another cantrip of your choice from the Druid spell",
        "druid/spellcasting/replace-one-cantrip-on-druid-level",
    )
    rule(
        sp,
        "those spells otherwise count as",
        "druid/spellcasting/always-prepared-excluded-from-count-still-druid-spells",
    )
    rule(
        sp,
        "replacing any of the spells with other",
        "druid/spellcasting/long-rest-replace-any-prepared-spells",
    )
    rule(
        sp,
        "Druidcraft and Produce Flame are recommended",
        "druid/recommended-cantrips/druidcraft",
        "druid/recommended-cantrips/produce-flame",
    )
    rule(
        sp,
        "Cure Wounds, Faerie Fire, and Thunderwave are",
        "druid/recommended-spells/animal-friendship",
        "druid/recommended-spells/cure-wounds",
        "druid/recommended-spells/faerie-fire",
        "druid/recommended-spells/thunderwave",
    )
    for slug, col, domain in [
        ("cantrips", 1, "cantrips"),
        ("prepared-spells", 2, "level-1-plus"),
    ]:
        ch = choice(
            slug,
            sp,
            cells[0][col],
            ["druid/spell-list/p44-45/" + domain],
            "from the Druid spell list",
        )
        by[ch]["values_by_level"] = [
            dict(level=n + 1, value=row[col]) for n, row in enumerate(cells)
        ]
        if col == 2:
            predicate(
                "prepared-spells-slot-level",
                ch,
                "druid/spellcasting/prepared-level-has-slot",
                [sp],
                **{"minimum-spell-level": 1},
            )
    for j in range(1, 10):
        resource(
            f"spell-slot-{j}",
            sp,
            [(n + 1, row[j + 2]) for n, row in enumerate(cells) if row[j + 2]],
            "You regain all expended slots",
            "druid/spellcasting/long-rest-all-slots",
            f"druid/spell-slot-level/{j}",
        )
    dr = fs["druidic"]
    rule(dr, "You know Druidic, the secret language of Druids", "languages/druidic")
    rule(
        dr,
        "prepared. You can use Druidic",
        "druid/druidic/always-prepared-speak-with-animals",
    )
    by[dr]["always_prepared"] = [
        dict(
            level=1, spell_keys=[K + "druid/druidic/always-prepared-speak-with-animals"]
        )
    ]
    rule(
        dr,
        "automatically spot",
        "druid/druidic/knowers-automatically-spot-hidden-messages",
    )
    rule(
        dr,
        "can’t decipher it without magic",
        "druid/druidic/nonknower-investigation-intelligence-dc-15-spot-not-decipher-without-magic",
        hidden_message_dc=15,
    )
    po = fs["primal-order"]
    rule(po, "sacred roles of your choice", "druid/primal-order/choose-one")
    order = choice(
        "primal-order",
        po,
        1,
        ["druid/primal-order/magician", "druid/primal-order/warden"],
        "Magician. You know one extra cantrip",
    )
    magician = choice(
        "magician",
        order,
        1,
        ["druid/spell-list/p44-45/cantrips"],
        "Intelligence (Arcana or Nature) checks",
        "druid/magician/extra-cantrip",
        "druid/magician/arcana-nature-wisdom-modifier-minimum-1",
    )
    rule(po, "minimum bonus of +1", magician_extra_cantrips=1, magician_minimum_bonus=1)
    warden = choice(
        "warden",
        order,
        1,
        ["druid/warden/grant-martial-weapons-medium-armor"],
        "with Martial weapons and training with Medium",
        "weapons/martial",
        "armor/medium",
    )
    for slug, ch in [("magician", magician), ("warden", warden)]:
        predicate(
            slug + "-selected", ch, "druid/primal-order/" + slug + "-selected", [order]
        )
    ws = fs["wild-shape"]
    rule(
        ws,
        "As a Bonus Action, you shape-shift",
        "druid/wild-shape/bonus-action-learned-beast-form",
    )
    rule(
        ws,
        "half your Druid level",
        "druid/wild-shape/duration-hours-half-druid-level",
        duration_level_divisor=2,
    )
    rule(
        ws,
        "or until you use Wild Shape again",
        "druid/wild-shape/ends-on-reuse-incapacitated-death",
    )
    rule(
        ws,
        "leave the form early as a Bonus Action",
        "druid/wild-shape/bonus-action-end-early",
    )
    resource(
        "wild-shape",
        ws,
        [(n + 1, row[0]) for n, row in enumerate(cells) if row[0]],
        "Short Rest, and you regain all",
        "druid/wild-shape/short-rest-recover-one",
        "druid/wild-shape/long-rest-recover-all",
    )
    rule(ws, "You regain one expended use", short_rest_recovered_uses=1)
    forms_rows = re.findall(
        r"[ \t]+(2|4|8)\s+(4|6|8)\s+(1/4|1/2|1)\s+(No|Yes)\s*$", text, re.M
    )
    assert len(forms_rows) == 3, "Beast Shapes extraction"
    forms = choice(
        "known-forms",
        ws,
        4,
        [
            "druid/beasts/eligible-stat-blocks",
            "druid/beasts/other-sources-with-gm-permission",
        ],
        "other sources for eligible Beasts if the Game Master",
    )
    by[forms]["values_by_level"] = [
        dict(
            level=n,
            value=int(next(row[1] for row in reversed(forms_rows) if int(row[0]) <= n)),
        )
        for n in range(2, 21)
    ]
    for n, count, cr, fly in forms_rows:
        numerator, denominator = (1, int(cr.split("/")[1])) if "/" in cr else (1, 1)
        rule(
            ws,
            "Beast Shapes",
            **{
                f"forms_at_level_{n}": int(count),
                f"cr_numerator_at_level_{n}": numerator,
                f"cr_denominator_at_level_{n}": denominator,
            },
        )
    predicate(
        "known-form-eligibility",
        forms,
        "druid/wild-shape/beast-cr-limit-no-fly-before-level-8",
        [ws],
        **{"fly-speed-minimum-druid-level": 8},
    )
    rule(
        forms,
        "replace one of your known forms",
        "druid/wild-shape/long-rest-replace-one-known-form",
    )
    rule(
        forms,
        "The Rat, Riding Horse, Spider,",
        "druid/recommended-forms/rat",
        "druid/recommended-forms/riding-horse",
        "druid/recommended-forms/spider",
        "druid/recommended-forms/wolf",
    )
    rule(
        ws,
        "retain your personality, memories, and ability to",
        "druid/wild-shape/retain-personality-memories-speech",
    )
    rule(
        ws,
        "Points equal to your Druid level",
        "druid/wild-shape/temporary-hp-equal-druid-level",
    )
    rule(
        ws,
        "by the Beast’s stat block, but you retain",
        "druid/wild-shape/replace-statistics-with-beast",
        "druid/wild-shape/retain-type-hp-hit-dice-intelligence-wisdom-charisma-class-features-languages-feats",
    )
    rule(
        ws,
        "ficiency Bonus for them, in addition to gaining",
        "druid/wild-shape/retain-skills-saves-use-own-proficiency-bonus-gain-beast-proficiencies",
    )
    rule(
        ws,
        "use the one in the stat block",
        "druid/wild-shape/use-higher-beast-skill-save-modifier",
    )
    rule(
        ws,
        "shifting doesn’t break your Concentration",
        "druid/wild-shape/no-spellcasting-retain-concentration-and-prior-spells",
    )
    rule(
        ws,
        "mined by the form’s limbs rather than your own",
        "druid/wild-shape/object-handling-form-limbs",
    )
    equipment = choice(
        "shape-equipment",
        ws,
        1,
        [
            "druid/equipment/fall-in-space",
            "druid/equipment/merge",
            "druid/equipment/wear",
        ],
        "falls in your space, merges into your new form",
    )
    rule(
        equipment,
        "Worn equipment functions as nor-",
        "druid/wild-shape/worn-equipment-functions-normally-gm-size-shape-practicality",
        "druid/wild-shape/equipment-does-not-resize-reshape",
        "druid/wild-shape/unwearable-must-fall-or-merge",
        "druid/wild-shape/merged-equipment-no-effect",
    )
    wc = fs["wild-companion"]
    rule(
        wc,
        "As a Magic action",
        "druid/wild-companion/magic-action-find-familiar-no-material-components",
    )
    rule(
        wc,
        "familiar is Fey and disappears",
        "druid/wild-companion/fey-disappears-on-long-rest",
    )
    choice(
        "wild-companion-cost",
        wc,
        1,
        ["druid/cost/one-spell-slot", "druid/cost/one-wild-shape-use"],
        "expend a spell slot or a use of Wild Shape",
    )
    sub = fs["druid-subclass"]
    rule(
        sub,
        "features that are of your Druid level or lower",
        "druid/subclass-features-at-or-below-druid-level",
    )
    choice(
        "druid-subclass",
        sub,
        1,
        ["druid/subclasses", "druid/subclasses/circle-of-the-land"],
        "of the Land subclass is detailed",
    )
    for slug, domain, recommendation in [
        ("ability-score-improvement", "ability-score-improvement", ""),
        (
            "epic-boon",
            "epic-boon",
            "druid/epic-boon/recommended-boon-of-dimensional-travel",
        ),
    ]:
        rule(fs[slug], "feat of your choice for which you qualify", "druid/" + slug)
        ch = choice(
            slug,
            fs[slug],
            1,
            ["feats/" + domain, "feats/qualified"],
            "feat of your choice for which you qualify",
        )
        predicate(slug + "-qualification", ch, "prerequisite/feat-qualification")
        if recommendation:
            rule(fs[slug], "mensional Travel is recommended", recommendation)
        else:
            rule(fs[slug], "12, and 16", "druid/asi-repeat-at-8-12-16")
    wr = fs["wild-resurgence"]
    rule(
        wr,
        "Once on each of your turns",
        "druid/wild-resurgence/once-own-turn-zero-uses-expend-slot-regain-one-no-action",
        regained_uses=1,
    )
    choice(
        "wild-resurgence-conversion",
        wr,
        1,
        [
            "druid/conversion/spell-slot-to-one-wild-shape",
            "druid/conversion/one-wild-shape-to-level-1-slot",
        ],
        "give yourself a level 1 spell",
    )
    predicate(
        "wild-resurgence-empty",
        wr,
        "druid/wild-resurgence/slot-to-use-requires-zero-uses",
        [C + "resource/druid/wild-shape"],
        **{"maximum-remaining-uses": 0},
    )
    resource(
        "wild-resurgence-slot-conversion",
        wr,
        [(n, 1) for n in range(5, 21)],
        "can’t do so again until you finish a Long",
        "druid/wild-resurgence/one-use-to-level-1-slot-no-action-once-long-rest",
    )
    ef = fs["elemental-fury"]
    rule(
        ef,
        "one of the following options of your choice",
        "druid/elemental-fury/choose-one",
    )
    fury = choice(
        "elemental-fury",
        ef,
        1,
        [
            "druid/elemental-fury/potent-spellcasting",
            "druid/elemental-fury/primal-strike",
        ],
        "Potent Spellcasting. Add your Wisdom modifier",
    )
    potent = choice(
        "potent-spellcasting",
        fury,
        1,
        ["druid/potent-spellcasting/add-wisdom-to-druid-cantrip-damage"],
        "to the damage you deal with any Druid cantrip",
    )
    strike = choice(
        "primal-strike",
        fury,
        1,
        ["damage/" + d for d in ["cold", "fire", "lightning", "thunder"]],
        "Lightning, or Thunder damage (choose when you",
        "druid/primal-strike/once-own-turn-weapon-or-wild-shape-beast-attack-roll-hit-creature-choose-on-hit",
    )
    rule(ef, "an extra 1d8 Cold, Fire", primal_strike_dice=1, primal_strike_die_sides=8)
    for slug, ch in [("potent-spellcasting", potent), ("primal-strike", strike)]:
        predicate(
            slug + "-selected", ch, "druid/elemental-fury/" + slug + "-selected", [fury]
        )
    improved = fs["improved-elemental-fury"]
    rule(
        improved,
        "The option you chose for Elemental Fury",
        "druid/improved-elemental-fury/improves-only-chosen-option",
    )
    predicate(
        "improved-fury-choice", improved, "druid/requires-elemental-fury-choice", [fury]
    )
    rule(
        improved,
        "range increases by 300 feet",
        "druid/improved-elemental-fury/potent-druid-cantrip-minimum-range-10-add-300-feet",
        potent_minimum_range_feet=10,
        potent_added_range_feet=300,
    )
    rule(
        improved,
        "Strike increases to 2d8",
        "druid/improved-elemental-fury/primal-strike-extra-2d8",
        primal_strike_dice=2,
        primal_strike_die_sides=8,
    )
    bs = fs["beast-spells"]
    rule(
        bs,
        "While using Wild Shape, you can cast spells in Beast",
        "druid/beast-spells/cast-in-wild-shape",
    )
    predicate(
        "beast-spells-components",
        bs,
        "druid/beast-spells/exclude-costed-or-consumed-material",
        [ws],
    )
    ar = fs["archdruid"]
    rule(
        ar,
        "Whenever you roll Initia-",
        "druid/archdruid/initiative-zero-wild-shape-regain-one",
        initiative_recovered_uses=1,
    )
    rule(
        ar,
        "with each use contributing 2 spell levels",
        "druid/archdruid/nature-magician-no-action-convert-unexpended-uses-single-slot",
        spell_levels_per_use=2,
    )
    choice(
        "nature-magician-uses",
        ar,
        1,
        ["druid/archdruid/choose-number-of-unexpended-wild-shape-uses"],
        "a number of your unexpended uses of Wild Shape",
    )
    resource(
        "nature-magician",
        ar,
        [(20, 1)],
        "Once you use this benefit",
        "druid/archdruid/nature-magician-once-long-rest",
    )
    predicate(
        "nature-magician-unexpended",
        ar,
        "druid/archdruid/requires-unexpended-wild-shape",
        [C + "resource/druid/wild-shape"],
        **{"minimum-uses": 1},
    )
    rule(
        ar,
        "that pass, your body ages only one year",
        "druid/archdruid/longevity-ten-years-per-body-year",
        elapsed_years=10,
        body_years=1,
    )
    # Independent numeric extraction binds the prose quantities to the actual
    # PDF digits as well as to their containing mechanic, not just an anchor.
    for owner, key, pattern in [
        (cl, "hit-die-sides", r"D(\d+) per Druid level"),
        (dr, "hidden-message-dc", r"DC (\d+) Intelligence"),
        (po, "magician-minimum-bonus", r"minimum bonus of \+(\d+)"),
        (ef, "primal-strike-dice", r"an extra (\d+)d8 Cold"),
        (ef, "primal-strike-die-sides", r"an extra 1d(\d+) Cold"),
        (improved, "potent-minimum-range-feet", r"range of (\d+) feet or greater"),
        (improved, "potent-added-range-feet", r"range increases by (\d+) feet"),
        (improved, "primal-strike-dice", r"Strike increases to (\d+)d8"),
        (improved, "primal-strike-die-sides", r"Strike increases to 2d(\d+)"),
        (ar, "spell-levels-per-use", r"each use contributing (\d+) spell levels"),
    ]:
        match = re.search(pattern, flat)
        assert match, (owner, key, "numeric source extraction")
        by[owner]["parameters"][K + "druid/" + key] = int(match[1])
    core = " ".join(prose.replace("\u00ad", "").split())
    assert re.search(
        r"Choose 2: Animal Handling, Arcana, Insight, Medicine, Nature, Perception, Religion, or Survival",
        core,
    ), "complete skill domain"
    assert (
        "Leather Armor, Shield, Sickle, Druidic Focus (Quarterstaff), Explorer’s Pack, Herbalism Kit, and 9 GP"
        in core
    ), "equipment A contents"
    money_a = re.search(r"Herbalism Kit, and (\d+) GP", core)
    money_b = re.search(r"or \(B\) (\d+) GP", core)
    assert money_a and money_b, "equipment currency extraction"
    by[eq]["options"][0]["quantities"][K + "currency/gp"] = int(money_a[1])
    by[eq]["options"][1]["quantities"][K + "currency/gp"] = int(money_b[1])
    return records


def compare(actual, expected, path="records"):
    if type(actual) is not type(expected):
        raise AssertionError(f"{path}: type mismatch")
    if isinstance(expected, dict):
        assert actual.keys() == expected.keys(), f"{path}: field set"
        for k in expected:
            compare(actual[k], expected[k], path + "/" + k)
    elif isinstance(expected, list):
        assert len(actual) == len(expected), f"{path}: length"
        for i, (a, e) in enumerate(zip(actual, expected)):
            compare(a, e, f"{path}/{i}")
    else:
        assert actual == expected, f"{path}: {actual!r} != source {expected!r}"


def verify(pdf, catalogue=DATA):
    expected = source_records(pdf)
    c = json.loads(catalogue.read_text())
    assert c["source"]["sha256"] == PIN and c["source"]["license"] == "CC-BY-4.0"
    inventory = json.loads(
        (ROOT / "docs/investigations/issue-74-character-inventory.json").read_text()
    )
    identities = {
        r["id"]: r["source_page"]
        for r in inventory["records"]
        if r["id"] == "srd-5.2.1:planning/classes/druid"
        or r["id"].startswith("srd-5.2.1:planning/class-features/druid-level-")
    }
    actual = {
        s["inventory_id"]: s["page"]
        for r in c["records"]
        if r["kind"] in ("class", "feature")
        for s in r["sources"]
    }
    assert len(identities) == len(actual) == 14 and actual == identities, (
        "source identities/pages"
    )
    compare(c["records"], expected)
    # Contrary-source records are still structurally valid metadata: predicate
    # operands point at existing records, numbers stay positive, keys are opaque.
    controls = [
        (
            "feature/druid/improved-elemental-fury",
            "parameters",
            K + "druid/potent-added-range-feet",
            301,
        ),
        (
            "prerequisite/druid/magician-selected",
            "requirement",
            "record_refs",
            [C + "choice/druid/elemental-fury"],
        ),
        ("choice/druid/primal-strike", None, "option_keys", [K + "damage/acid"]),
        (
            "resource/druid/wild-shape",
            None,
            "required_mechanics",
            [K + "druid/wild-shape/short-rest-recover-all"],
        ),
    ]
    for suffix, field, key, value in controls:
        bad = copy.deepcopy(c["records"])
        target = next(r for r in bad if r["id"] == C + suffix)
        (target[field] if field else target)[key] = value
        try:
            compare(bad, expected)
        except AssertionError:
            continue
        raise AssertionError("contrary-source control accepted: " + suffix)
    print(
        f"SRD {PIN}: 14 Druid identities, 20 full rows, every graph field, mechanic, predicate/operand, option, number and recovery match pp41–43; 4 contrary-source controls rejected"
    )


if __name__ == "__main__":
    if len(sys.argv) not in (2, 3):
        raise SystemExit("usage: verify_srd_druid.py PINNED.pdf [catalogue.json]")
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) == 3 else DATA)
