#!/usr/bin/env python3
"""Standalone offline Paladin semantic comparator against pinned PDF pp53–55.

Expected records are a clause-bound source transcription. Table values are
extracted directly from the PDF; no expectations are read from catalogue data
or Go tests. Exact graph comparison includes optional fields and operands.
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
DATA = ROOT / "internal/rules/srd521/testdata/character-paladin.json"


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"

    def extract(*args):
        return subprocess.run(
            ["pdftotext", "-f", "53", "-l", "55", *args, str(pdf), "-"],
            check=True,
            capture_output=True,
            text=True,
        ).stdout

    layout, prose = extract("-layout"), extract()
    flat = " ".join(prose.replace("\u00ad", "").split())
    records, by = [], {}

    def rec(kind, slug, parent="", level=0, page=53, name=None):
        ident = C + kind + "/paladin" + ("/" + slug if slug else "")
        sources = (
            [dict(inventory_id="srd-5.2.1:planning/classes/paladin", page=53)]
            if kind == "class"
            else (
                [
                    dict(
                        inventory_id=f"srd-5.2.1:planning/class-features/paladin-level-{by[parent]['level']}-{slug}",
                        page=page,
                    )
                ]
                if kind == "feature"
                else copy.deepcopy(by[parent]["sources"])
            )
        )
        r = dict(
            id=ident,
            kind=kind,
            name=name or slug.replace("-", " ").title(),
            sources=sources,
            parent_id=parent,
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

    def rule(owner, clause, *keys, **nums):
        assert clause in flat, "missing source clause: " + clause
        r = by[owner]
        r["required_mechanics"].extend(K + k for k in keys)
        for k, v in nums.items():
            key = K + "paladin/" + k.replace("_", "-")
            r.setdefault("parameters", {})[key] = v
            r["required_mechanics"].append(key)

    def choice(slug, parent, count, options, clause, *keys):
        ident = rec("choice", slug, parent)
        by[ident].update(select=count, option_keys=[K + k for k in options])
        rule(ident, clause, *(keys or ("paladin/" + slug,)))
        return ident

    def pred(slug, parent, key, refs=(), **nums):
        ident = rec("prerequisite", slug, parent)
        by[ident]["required_mechanics"] = [K + key]
        by[ident]["requirement"] = dict(
            key=K + key, record_refs=list(refs), parameters=nums
        )
        return ident

    def resource(slug, parent, values, clause, *keys):
        ident = rec("resource", slug, parent)
        by[ident].update(
            capacity=values[0][1],
            values_by_level=[dict(level=n, value=v) for n, v in values],
        )
        rule(ident, clause, *keys)
        return ident

    cl = rec("class", "", name="Paladin")
    rule(
        cl,
        "Strength and Charisma",
        "paladin/primary-ability/strength",
        "paladin/primary-ability/charisma",
    )
    rule(cl, "D10 per Paladin level", hit_die_sides=10)
    rule(cl, "Wisdom and Charisma", "saving-throws/wisdom", "saving-throws/charisma")
    rule(cl, "Simple and Martial weapons", "weapons/simple", "weapons/martial")
    rule(
        cl,
        "mor and Shields",
        "armor/light",
        "armor/medium",
        "armor/heavy",
        "armor/shields",
    )
    rule(
        cl,
        "Gain all the traits in the Core Paladin Traits table",
        "paladin/level-1/all-core-traits-and-features",
    )
    rule(
        cl,
        "Traits table: Hit Point Die",
        "paladin/multiclass/hit-die-martial-weapons-light-medium-armor-shields-only",
        "paladin/multiclass/level-1-features-and-multiclass-spell-slots",
    )
    choice(
        "skills",
        cl,
        2,
        [
            "skills/" + s
            for s in [
                "athletics",
                "insight",
                "intimidation",
                "medicine",
                "persuasion",
                "religion",
            ]
        ],
        "Choose 2: Athletics, Insight, Intimidation, Medicine, Persuasion, or Religion",
    )
    eq = choice(
        "starting-equipment",
        cl,
        1,
        ["paladin/equipment/a", "paladin/equipment/b"],
        "Choose A or B",
    )
    assert (
        "Shield, Longsword, 6 Javelins, Holy Symbol, Priest’s Pack, and 9 GP; or (B) 150 GP"
        in flat
    )
    bundles = [
        {
            "equipment/chain-mail": 1,
            "equipment/shield": 1,
            "equipment/longsword": 1,
            "equipment/javelin": 6,
            "equipment/holy-symbol": 1,
            "equipment/priests-pack": 1,
            "currency/gp": 9,
        },
        {"currency/gp": 150},
    ]
    by[eq]["options"] = [
        dict(
            key=K + "paladin/equipment/" + s,
            quantities={K + k: v for k, v in b.items()},
        )
        for s, b in zip(["a", "b"], bundles)
    ]
    rule(eq, "Chain Mail", *dict.fromkeys(k for b in bundles for k in b))
    features = {
        1: ["lay-on-hands", "spellcasting", "weapon-mastery"],
        2: ["fighting-style", "paladins-smite"],
        3: ["channel-divinity", "paladin-subclass"],
        4: ["ability-score-improvement"],
        5: ["extra-attack", "faithful-steed"],
        6: ["aura-of-protection"],
        9: ["abjure-foes"],
        10: ["aura-of-courage"],
        11: ["radiant-strikes"],
        14: ["restoring-touch"],
        18: ["aura-expansion"],
        19: ["epic-boon"],
    }
    fs = {}
    for n in range(1, 21):
        row = rec("level", str(n), cl, n, name=f"Paladin level {n}")
        rule(row, "Paladin Features", "paladin/class-progression")
        for slug in features.get(n, []):
            page = (
                53
                if slug == "lay-on-hands"
                else (
                    54
                    if slug
                    in (
                        "spellcasting",
                        "weapon-mastery",
                        "fighting-style",
                        "paladins-smite",
                        "channel-divinity",
                    )
                    else 55
                )
            )
            fs[slug] = rec(
                "feature",
                slug,
                row,
                page=page,
                name=(
                    "Paladin’s Smite"
                    if slug == "paladins-smite"
                    else (
                        "Lay On Hands"
                        if slug == "lay-on-hands"
                        else slug.replace("-", " ").title().replace(" Of ", " of ")
                    )
                ),
            )
    table = layout.split("Paladin Features\n")[-1].split("53   System")[0]
    table = re.sub(
        r"Lay On Hands, Spellcasting,\s+([—\d].*?)\n\s+Weapon Mastery",
        r"Lay On Hands, Spellcasting, Weapon Mastery  \1",
        table,
    )
    rows = re.findall(
        r"^\s+(\d{1,2})\s+\+(\d)\s+(.+?)\s{2,}([\d—][\d—\s]+)$", table, re.M
    )
    assert [int(r[0]) for r in rows] == list(range(1, 21)), "all progression rows"
    cells = []
    for n, bonus, names, suffix in rows:
        n = int(n)
        nums = [0 if x == "—" else int(x) for x in suffix.split()]
        assert len(nums) == 7, (n, nums)
        cells.append(nums)
        refs = [fs[s] for s in features.get(n, [])]
        if n in (8, 12, 16):
            refs = [fs["ability-score-improvement"]]
        keys = [K + "paladin/subclass-feature"] if n in (7, 15, 20) else []
        by[C + f"level/paladin/{n}"]["progression"] = dict(
            proficiency_bonus=int(bonus), feature_refs=refs, feature_keys=keys
        )
        by[C + f"level/paladin/{n}"]["required_mechanics"] += keys
        assert names == (
            ", ".join(
                [by[x]["name"] for x in refs] + ["Subclass feature" for _ in keys]
            )
            or "—"
        ), (n, names)
    loh = fs["lay-on-hands"]
    pool = resource(
        "lay-on-hands",
        loh,
        [(n, 5 * n) for n in range(1, 21)],
        "five times your Paladin level",
        "paladin/lay-on-hands/long-rest-replenish-all",
    )
    rule(
        loh,
        "As a Bonus Action, you can touch a creature",
        "paladin/lay-on-hands/bonus-action-touch-including-self-heal-up-to-remaining-pool",
        healing_hp_per_level=5,
    )
    rule(
        loh,
        "those points don’t also restore Hit Points",
        "paladin/lay-on-hands/remove-poisoned-no-healing-for-cost",
        poisoned_removal_cost_hp=5,
    )
    pred(
        "lay-on-hands-pool",
        loh,
        "paladin/lay-on-hands/expend-only-remaining-pool",
        [pool],
    )
    sp = fs["spellcasting"]
    rule(sp, "Charisma is your spellcasting", "paladin/spellcasting/charisma")
    rule(
        sp,
        "use a Holy Symbol as a Spellcasting Focus",
        "paladin/spellcasting/holy-symbol-focus",
    )
    rule(
        sp,
        "replace one spell on your list",
        "paladin/spellcasting/long-rest-replace-one-prepared-spell",
        replaced_spells_per_long_rest=1,
    )
    rule(
        sp,
        "those spells don’t count",
        "paladin/spellcasting/always-prepared-excluded-from-count-still-paladin-spells",
    )
    rule(
        sp,
        "Heroism and Searing Smite are recommended",
        "paladin/recommended-spells/heroism",
        "paladin/recommended-spells/searing-smite",
    )
    prepared = choice(
        "prepared-spells",
        sp,
        cells[0][1],
        ["paladin/spell-list/p55-56/level-1-plus"],
        "choose two level 1 Paladin spells",
        "paladin/spellcasting/add-spells-when-table-count-increases",
    )
    by[prepared]["values_by_level"] = [
        dict(level=n + 1, value=row[1]) for n, row in enumerate(cells)
    ]
    pred(
        "prepared-spells-slot-level",
        prepared,
        "paladin/spellcasting/prepared-level-has-slot",
        [sp],
        **{"minimum-spell-level": 1},
    )
    for j in range(1, 6):
        resource(
            f"spell-slot-{j}",
            sp,
            [(n + 1, row[j + 1]) for n, row in enumerate(cells) if row[j + 1]],
            "You regain all expended slots",
            "paladin/spellcasting/long-rest-all-slots",
            f"paladin/spell-slot-level/{j}",
        )
    wm = fs["weapon-mastery"]
    mastery = choice(
        "weapon-mastery",
        wm,
        2,
        ["weapons/proficient-kinds"],
        "mastery properties of two kinds of weapons",
        "paladin/weapon-mastery/long-rest-replace-kinds",
    )
    pred(
        "mastery-proficiency",
        mastery,
        "paladin/weapon-mastery/requires-weapon-proficiency",
        [cl],
    )
    rule(
        wm,
        "Longswords and Javelins",
        "paladin/weapon-mastery/examples-longswords-javelins",
        "paladin/weapon-mastery/replacement-examples-halberds-flails",
    )
    rule(
        fs["fighting-style"],
        "You gain a Fighting Style feat",
        "paladin/fighting-style/choose-feat-or-blessed-warrior",
    )
    style = choice(
        "fighting-style",
        fs["fighting-style"],
        1,
        ["feats/fighting-style", "paladin/fighting-style/blessed-warrior"],
        "Instead of choosing one of those feats",
    )
    bw = choice(
        "blessed-warrior",
        style,
        2,
        ["cleric/spell-list/cantrips"],
        "You learn two Cleric cantrips",
        "paladin/blessed-warrior/count-as-paladin-charisma",
        "paladin/blessed-warrior/replace-one-cleric-cantrip-on-paladin-level",
        "paladin/blessed-warrior/recommended-guidance-sacred-flame",
    )
    pred(
        "blessed-warrior-selected",
        bw,
        "paladin/fighting-style/blessed-warrior-selected",
        [style],
    )
    for slug, spell, start, clause in [
        (
            "paladins-smite",
            "divine-smite",
            2,
            "You always have the Divine Smite spell prepared",
        ),
        (
            "faithful-steed",
            "find-steed",
            5,
            "always have the Find Steed spell prepared",
        ),
    ]:
        f = fs[slug]
        key = K + "spells/" + spell
        rule(f, clause, "spells/" + spell, "paladin/" + slug + "/always-prepared")
        by[f]["always_prepared"] = [dict(level=start, spell_keys=[key])]
        resource(
            slug + "-free-cast",
            f,
            [(n, 1) for n in range(start, 21)],
            "without expending a spell slot",
            "paladin/" + slug + "/one-slot-free-cast-long-rest-recovery",
        )
    cd = fs["channel-divinity"]
    uses = resource(
        "channel-divinity",
        cd,
        [(n + 1, row[0]) for n, row in enumerate(cells) if row[0]],
        "regain one of its expended uses",
        "paladin/channel-divinity/short-rest-recover-one",
        "paladin/channel-divinity/long-rest-recover-all",
    )
    rule(
        cd,
        "DC equals the spell save DC",
        "paladin/channel-divinity/use-paladin-spell-save-dc",
        short_rest_recovered_uses=1,
    )
    choice(
        "channel-divinity-effect",
        cd,
        1,
        [
            "paladin/channel-divinity/divine-sense",
            "paladin/channel-divinity/other-paladin-feature-effects",
        ],
        "choose which effect from this class to create",
    )
    rule(
        cd,
        "Divine Sense. As a Bonus Action",
        "paladin/divine-sense/bonus-action-expend-one-channel-divinity",
        "paladin/divine-sense/detect-location-and-type-celestials-fiends-undead",
        "paladin/divine-sense/detect-consecrated-desecrated-places-objects-hallow-reference",
        "paladin/divine-sense/ends-if-incapacitated",
        divine_sense_cost_uses=1,
        divine_sense_duration_minutes=10,
        divine_sense_radius_feet=60,
    )
    sub = fs["paladin-subclass"]
    rule(
        sub, "You gain a Paladin subclass of your choice", "paladin/subclass/choose-one"
    )
    choice(
        "paladin-subclass",
        sub,
        1,
        ["paladin/subclasses", "paladin/subclasses/oath-of-devotion"],
        "Oath of Devotion subclass",
    )
    rule(
        sub,
        "features that are of your Paladin level or lower",
        "paladin/subclass/features-at-or-below-paladin-level",
    )
    for slug in ("ability-score-improvement", "epic-boon"):
        rule(fs[slug], "feat of your choice for which you qualify", "paladin/" + slug)
        ch = choice(
            slug,
            fs[slug],
            1,
            ["feats/" + slug, "feats/qualified"],
            "feat of your choice for which you qualify",
        )
        pred(slug + "-qualification", ch, "prerequisite/feat-qualification")
    rule(
        fs["ability-score-improvement"],
        "8, 12, and 16",
        "paladin/asi-repeat-at-8-12-16",
    )
    rule(
        fs["epic-boon"],
        "Truesight is recommended",
        "paladin/epic-boon/recommended-boon-of-truesight",
    )
    rule(
        fs["extra-attack"],
        "attack twice instead of once",
        "paladin/extra-attack/attack-action-own-turn",
        attacks_per_attack_action=2,
    )
    aura = fs["aura-of-protection"]
    rule(
        aura,
        "10-foot Emanation",
        "paladin/aura-of-protection/self-origin-emanation",
        aura_initial_radius_feet=10,
    )
    rule(
        aura,
        "inactive while you have the Incapacitated condition",
        "paladin/aura-of-protection/inactive-if-incapacitated",
    )
    rule(
        aura,
        "minimum bonus of +1",
        "paladin/aura-of-protection/self-and-allies-saving-throws-charisma-modifier",
        aura_minimum_bonus=1,
    )
    rule(
        aura,
        "creature chooses which aura",
        "paladin/aura-of-protection/only-one-aura-benefit-target-chooses",
    )
    ab = fs["abjure-foes"]
    rule(
        ab,
        "As a Magic action, you can expend one use",
        "paladin/abjure-foes/magic-action-present-holy-symbol-or-weapon",
        channel_divinity_cost=1,
    )
    pred(
        "abjure-channel-divinity",
        ab,
        "paladin/abjure-foes/expend-channel-divinity",
        [uses],
        **{"cost-uses": 1},
    )
    rule(
        ab,
        "Charisma modifier (minimum of one creature)",
        "paladin/abjure-foes/visible-targets-charisma-modifier-minimum-one",
        abjure_minimum_targets=1,
        abjure_range_feet=60,
    )
    rule(
        ab,
        "Wisdom saving throw",
        "paladin/abjure-foes/wisdom-save-or-frightened",
        "paladin/abjure-foes/ends-on-any-damage",
        abjure_duration_minutes=1,
    )
    rule(
        ab,
        "move, take an action, or take a Bonus Action",
        "paladin/abjure-foes/frightened-target-only-one-move-action-or-bonus-action-on-turn",
    )
    courage = fs["aura-of-courage"]
    rule(
        courage,
        "You and your allies have Immunity",
        "paladin/aura-of-courage/self-and-allies-frightened-immunity-in-protection-aura",
        "paladin/aura-of-courage/entering-frightened-ally-condition-suppressed-while-in-aura",
    )
    pred(
        "courage-aura",
        courage,
        "paladin/aura-of-courage/within-aura-of-protection",
        [aura],
    )
    rule(
        fs["radiant-strikes"],
        "extra 1d8 Radiant damage",
        "paladin/radiant-strikes/attack-roll-hit-melee-weapon-or-unarmed-strike",
        radiant_damage_dice=1,
        radiant_damage_die_sides=8,
    )
    rt = fs["restoring-touch"]
    rule(
        rt,
        "When you use Lay On Hands",
        "paladin/restoring-touch/with-lay-on-hands-remove-one-or-more-conditions",
        "paladin/restoring-touch/cost-does-not-heal",
        restoring_cost_hp_per_condition=5,
    )
    choice(
        "restoring-touch-conditions",
        rt,
        1,
        [
            "conditions/" + s
            for s in [
                "blinded",
                "charmed",
                "deafened",
                "frightened",
                "paralyzed",
                "stunned",
            ]
        ],
        "Blinded, Charmed, Deafened, Frightened, Paralyzed, or Stunned",
        "paladin/restoring-touch/choose-one-or-more-with-per-condition-cost",
    )
    pred(
        "restoring-healing-pool",
        rt,
        "paladin/restoring-touch/expend-lay-on-hands-pool",
        [pool],
        **{"hp-per-condition": 5},
    )
    rule(
        fs["aura-expansion"],
        "now a 30-foot Emanation",
        "paladin/aura-expansion/replaces-protection-radius",
        aura_expanded_radius_feet=30,
    )
    pred(
        "expanded-aura",
        fs["aura-expansion"],
        "paladin/aura-expansion/changes-aura-of-protection",
        [aura],
    )
    for clause in [
        "Light, Medium, and Heavy ar-",
        "Long Rest",
        "Guidance and Sacred Flame",
        "count as Paladin spells",
        "replace one of these cantrips",
        "You gain an additional use when you reach Paladin level 11",
        "For the next 10 minutes or until you have the Incapacitated condition",
        "know the location",
        "know its creature type",
        "Celestials, Fiends, and Undead",
        "consecrated or desecrated",
        "Hallow spell",
        "can see within 60 feet",
        "Frightened condition for 1 minute or until it takes any damage",
        "no effect on that ally while there",
        "Melee weapon or an Unarmed Strike",
        "5 Hit Points from the healing pool",
        "for each of these conditions",
    ]:
        assert clause in " ".join(
            (prose + layout).replace("\u00ad", "").split()
        ), clause
    # Independent PDF numeric extraction also binds mechanics to actual digits.
    for owner, key, pattern in [
        (cl, "hit-die-sides", r"D(\d+) per Paladin level"),
        (loh, "poisoned-removal-cost-hp", r"expend (\d+) Hit Points from the pool"),
        (cd, "divine-sense-duration-minutes", r"next (\d+) minutes"),
        (cd, "divine-sense-radius-feet", r"those types within (\d+) feet"),
        (aura, "aura-initial-radius-feet", r"aura in a (\d+)-foot"),
        (aura, "aura-minimum-bonus", r"minimum bonus of \+(\d+)"),
        (ab, "abjure-range-feet", r"can see within (\d+) feet"),
        (ab, "abjure-duration-minutes", r"Frightened condition for (\d+) minute"),
        (
            rt,
            "restoring-cost-hp-per-condition",
            r"pend (\d+) Hit Points from the healing pool",
        ),
        (fs["radiant-strikes"], "radiant-damage-dice", r"extra (\d+)d8 Radiant"),
        (fs["radiant-strikes"], "radiant-damage-die-sides", r"extra 1d(\d+) Radiant"),
        (
            fs["aura-expansion"],
            "aura-expanded-radius-feet",
            r"now a (\d+)-foot Emanation",
        ),
    ]:
        match = re.search(pattern, flat)
        assert match, (owner, key, "numeric source extraction")
        by[owner]["parameters"][K + "paladin/" + key] = int(match[1])
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


def controls(records):
    def edit(suffix, field, value):
        bad = copy.deepcopy(records)
        r = next(r for r in bad if r["id"] == C + suffix)
        target = r
        for k in field[:-1]:
            target = target[k]
        target[field[-1]] = value
        return bad

    cases = [
        edit(
            "prerequisite/paladin/restoring-healing-pool",
            ["requirement", "record_refs"],
            [C + "resource/paladin/channel-divinity"],
        ),
        edit(
            "feature/paladin/abjure-foes",
            ["parameters", K + "paladin/abjure-range-feet"],
            61,
        ),
        edit(
            "feature/paladin/restoring-touch",
            ["parameters", K + "paladin/restoring-cost-hp-per-condition"],
            4,
        ),
        edit(
            "resource/paladin/channel-divinity",
            ["required_mechanics"],
            [K + "paladin/channel-divinity/short-rest-recover-all"],
        ),
        edit(
            "feature/paladin/aura-of-protection",
            ["required_mechanics", 2],
            K + "paladin/aura-of-protection/active-if-incapacitated",
        ),
        edit(
            "choice/paladin/restoring-touch-conditions",
            ["option_keys"],
            [K + "conditions/blinded"],
        ),
        edit(
            "resource/paladin/spell-slot-5",
            ["values_by_level"],
            [dict(level=17, value=1), dict(level=19, value=2), dict(level=20, value=2)],
        ),
        edit(
            "feature/paladin/radiant-strikes",
            ["parameters", K + "paladin/radiant-damage-dice"],
            2,
        ),
    ]
    bad = copy.deepcopy(records)
    target = next(r for r in bad if r["id"] == C + "feature/paladin/radiant-strikes")
    key = K + "paladin/radiant-damage-dice-at-level-20"
    target["parameters"][key] = 2
    target["required_mechanics"].append(key)
    cases.append(bad)
    return cases


def verify(pdf, catalogue=DATA):
    expected = source_records(pdf)
    c = json.loads(catalogue.read_text())
    assert (
        c["schema_version"] == 1
        and c["source"]["sha256"] == PIN
        and c["source"]["license"] == "CC-BY-4.0"
    )
    inventory = json.loads(
        (ROOT / "docs/investigations/issue-74-character-inventory.json").read_text()
    )
    identities = {
        r["id"]: r["source_page"]
        for r in inventory["records"]
        if r["id"] == "srd-5.2.1:planning/classes/paladin"
        or r["id"].startswith("srd-5.2.1:planning/class-features/paladin-level-")
    }
    actual = {
        s["inventory_id"]: s["page"]
        for r in c["records"]
        if r["kind"] in ("class", "feature")
        for s in r["sources"]
    }
    assert (
        len(identities) == len(actual) == 18 and identities == actual
    ), "18 exact identities/pages"
    compare(c["records"], expected)
    for bad in controls(expected):
        try:
            compare(bad, expected)
        except AssertionError:
            continue
        raise AssertionError("contrary-source control accepted")
    print(
        "Paladin: 18 identities, 20 full progression rows, complete semantic graph match pinned PDF pp53–55; 9 contrary-source controls rejected"
    )


if __name__ == "__main__":
    if len(sys.argv) not in (2, 3):
        raise SystemExit("usage: verify_srd_paladin.py PINNED.pdf [catalogue.json]")
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) == 3 else DATA)
