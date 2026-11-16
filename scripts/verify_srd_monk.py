#!/usr/bin/env python3
"""Offline Monk semantic and actual PDF table comparison for issue #291.

Source facts are transcribed independently of the imported JSON and frozen Go
graph. Requires Python 3 and pdftotext; no network, Go or catalogue digests.
"""

# Source transcription of SRD 5.2.1 pp49-52. No runtime catalogue inputs.
C = "srd-5.2.1:character/"
M = "srd-5.2.1:mechanic/"
P = "srd-5.2.1:planning/"


def source_records():
    out = []

    def add(kind, slug, name, page, source, parent="", keys=(), nums=None):
        r = dict(
            id=C + kind + "/monk" + ("/" + slug if slug else ""),
            kind=kind,
            name=name,
            sources=[dict(inventory_id=P + source, page=page)],
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
            nums = {k: v for k, v in nums.items() if v != 0}
            r["parameters"] = {M + "monk/" + k: v for k, v in nums.items()}
            r["required_mechanics"] += list(r["parameters"])
        out.append(r)
        return r

    cls = add(
        "class",
        "",
        "Monk",
        49,
        "classes/monk",
        keys=[
            "monk/primary-abilities/dexterity-and-wisdom",
            "saving-throws/strength",
            "saving-throws/dexterity",
            "weapons/simple",
            "weapons/martial-light",
            "armor/none",
            "monk/level-1/all-core-traits-and-features",
            "monk/multiclass/hit-point-die-and-level-1-features-only",
        ],
        nums={"hit-die-sides": 8},
    )

    def choice(owner, slug, name, opts, keys=(), count=1, nums=None):
        if not keys:
            keys = ["monk/" + slug + "/choose-one"]
        r = add(
            "choice",
            slug,
            name,
            owner["sources"][0]["page"],
            owner["sources"][0]["inventory_id"][len(P) :],
            owner["id"],
            keys,
            nums,
        )
        r["select"] = count
        r["option_keys"] = [M + k for k in opts]
        owner["choice_refs"].append(r["id"])
        return r

    def pred(owner, slug, key, refs=(), nums=None):
        r = add(
            "prerequisite",
            slug,
            slug.replace("-", " ").title(),
            owner["sources"][0]["page"],
            owner["sources"][0]["inventory_id"][len(P) :],
            owner["id"],
            [key],
        )
        r["requirement"] = dict(
            key=M + key, record_refs=[C + x for x in refs], parameters=nums or {}
        )
        owner["prerequisite_refs"].append(r["id"])
        return r

    choice(
        cls,
        "skills",
        "Skills",
        [
            "skills/" + x
            for x in [
                "acrobatics",
                "athletics",
                "history",
                "insight",
                "religion",
                "stealth",
            ]
        ],
        ["monk/skills"],
        2,
    )
    choice(
        cls,
        "tools",
        "Tool proficiency",
        ["tools/any-artisans-tools", "tools/any-musical-instrument"],
        ["monk/tools/choose-one-type"],
    )
    eq = choice(
        cls,
        "starting-equipment",
        "Starting Equipment",
        ["monk/equipment/a", "monk/equipment/b"],
        ["monk/equipment/a/uses-chosen-proficient-tool"],
    )
    eq["options"] = [
        dict(
            key=M + "monk/equipment/a",
            quantities={
                M + "equipment/spear": 1,
                M + "equipment/dagger": 5,
                M + "monk/equipment/chosen-proficient-tool": 1,
                M + "equipment/explorers-pack": 1,
                M + "currency/gp": 11,
            },
        ),
        dict(key=M + "monk/equipment/b", quantities={M + "currency/gp": 50}),
    ]
    eq["required_mechanics"] += list(eq["options"][0]["quantities"])
    pred(
        eq,
        "equipment-tool",
        "monk/equipment/tool-matches-proficiency",
        ["choice/monk/tools"],
    )
    features = {}

    def feat(slug, name, level, page, keys, nums=None):
        r = add(
            "feature",
            slug,
            name,
            page,
            f"class-features/monk-level-{level}-{slug}",
            C + f"level/monk/{level}",
            ["monk/" + slug + "/" + k for k in keys],
            {slug + "/" + k: v for k, v in (nums or {}).items()},
        )
        features[slug] = (level, r)
        return r

    ma = feat(
        "martial-arts",
        "Martial Arts",
        1,
        50,
        [
            "monk-weapons/simple-melee-or-martial-melee-light",
            "bonus-unarmed-strike/bonus-action",
            "die/replace-normal-unarmed-or-monk-weapon-damage",
            "die/scales-with-monk-table",
            "dexterous-attacks/dexterity-instead-of-strength-attack-and-damage",
            "dexterous-attacks/grapple-or-shove-save-dc-dexterity-instead-of-strength",
        ],
    )
    pred(
        ma,
        "martial-arts-eligibility",
        "monk/martial-arts/unarmed-or-only-monk-weapons-and-no-armor-or-shield",
        ["class/monk"],
    )
    ud = feat(
        "unarmored-defense",
        "Unarmored Defense",
        1,
        50,
        ["base-ac-plus-dexterity-and-wisdom-modifiers"],
        {"base-ac": 10},
    )
    pred(
        ud,
        "unarmored-defense-eligibility",
        "monk/unarmored-defense/no-armor-or-shield",
        ["class/monk"],
    )
    focus = feat(
        "monks-focus",
        "Monk’s Focus",
        2,
        50,
        [
            "points/fuel-features",
            "save-dc/base-plus-wisdom-modifier-and-proficiency-bonus",
        ],
        {"save-dc/base": 8},
    )
    res = add(
        "resource",
        "focus-points",
        "Focus Points",
        50,
        "class-features/monk-level-2-monks-focus",
        focus["id"],
        [
            "monk/focus-points/regain-all-expended-on-short-or-long-rest",
            "monk/focus-points/expended-unavailable-until-recovered",
        ],
    )
    res["capacity"] = 2
    res["values_by_level"] = [dict(level=i, value=i) for i in range(2, 21)]
    focus["resource_refs"].append(res["id"])
    fl = choice(
        focus,
        "focus-actions",
        "Focus actions",
        [
            "monk/focus/flurry-of-blows",
            "monk/focus/patient-defense",
            "monk/focus/step-of-the-wind",
        ],
        ["monk/focus/known-three-actions-not-a-permanent-selection"],
    )

    def action(slug, opts, keys, nums):
        # Numeric action facts belong to the owning feature under schema v1.
        for k, v in nums.items():
            key = M + "monk/" + slug + "/" + k
            focus.setdefault("parameters", {})[key] = v
            focus["required_mechanics"].append(key)
        return choice(
            fl,
            slug,
            slug.replace("-", " ").title(),
            ["monk/" + slug + "/" + x for x in opts],
            ["monk/" + slug + "/" + x for x in keys],
        )

    a = action(
        "flurry-of-blows",
        ["paid"],
        ["bonus-action", "paid/two-unarmed-strikes"],
        {"cost": 1, "strikes": 2},
    )
    pred(
        a,
        "flurry-focus",
        "monk/flurry-of-blows/available-focus",
        ["resource/monk/focus-points"],
        {"points-required": 1},
    )
    a = action(
        "patient-defense",
        ["free-disengage", "paid-disengage-and-dodge"],
        ["bonus-action", "free/no-focus-cost", "paid/expend-focus"],
        {"paid-cost": 1},
    )
    pred(
        a,
        "patient-defense-paid",
        "monk/patient-defense/paid-option-requires-focus",
        ["resource/monk/focus-points"],
        {"points-required": 1},
    )
    a = action(
        "step-of-the-wind",
        ["free-dash", "paid-disengage-and-dash"],
        ["bonus-action", "free/no-focus-cost", "paid/jump-distance-doubled-for-turn"],
        {"paid-cost": 1, "paid-jump-multiplier": 2},
    )
    pred(
        a,
        "step-of-the-wind-paid",
        "monk/step-of-the-wind/paid-option-requires-focus",
        ["resource/monk/focus-points"],
        {"points-required": 1},
    )
    um = feat(
        "unarmored-movement",
        "Unarmored Movement",
        2,
        51,
        ["speed-increase-scales-with-table"],
    )
    pred(
        um,
        "unarmored-movement-eligibility",
        "monk/unarmored-movement/no-armor-or-shield",
        ["class/monk"],
    )
    met = feat(
        "uncanny-metabolism",
        "Uncanny Metabolism",
        2,
        51,
        [
            "on-initiative/optional-regain-all-focus",
            "on-use/heal-monk-level-plus-martial-arts-die",
            "once-per-long-rest",
        ],
    )
    rr = add(
        "resource",
        "uncanny-metabolism",
        "Uncanny Metabolism use",
        51,
        "class-features/monk-level-2-uncanny-metabolism",
        met["id"],
        ["monk/uncanny-metabolism/regain-use-on-long-rest"],
    )
    rr["capacity"] = 1
    met["resource_refs"].append(rr["id"])
    d = feat(
        "deflect-attacks",
        "Deflect Attacks",
        3,
        51,
        [
            "reaction/on-attack-roll-hit-including-bludgeoning-piercing-or-slashing",
            "reduce-total-damage/die-plus-dexterity-modifier-plus-monk-level",
            "redirect/same-damage-type-as-attack",
            "redirect/dexterity-save-or-two-martial-arts-dice-plus-dexterity-modifier",
            "redirect/melee-visible-target",
            "redirect/ranged-visible-target-not-total-cover",
        ],
        {
            "reduction-dice": 1,
            "reduction-die-sides": 10,
            "redirect-cost": 1,
            "redirect-dice": 2,
            "melee-range-feet": 5,
            "ranged-range-feet": 60,
        },
    )
    pred(
        d,
        "deflect-redirect",
        "monk/deflect-attacks/reduced-damage-zero-and-available-focus",
        ["resource/monk/focus-points"],
        {"remaining-damage": 0, "points-required": 1},
    )
    choice(
        d,
        "redirect-target",
        "Redirect target",
        ["monk/deflect-attacks/melee-target", "monk/deflect-attacks/ranged-target"],
        ["monk/deflect-attacks/target-domain-depends-on-incoming-attack"],
    )
    sub = feat(
        "monk-subclass",
        "Monk Subclass",
        3,
        51,
        ["gain-subclass-features-at-or-below-monk-level"],
    )
    choice(
        sub,
        "subclass",
        "Monk Subclass",
        ["monk/subclasses/warrior-of-the-open-hand"],
        ["monk/subclass/source-named-option-not-imported"],
    )
    asi = feat(
        "ability-score-improvement",
        "Ability Score Improvement",
        4,
        51,
        ["gain-asi-feat-or-another-qualified-feat", "repeat-at-levels-8-12-16"],
    )
    ch = choice(
        asi,
        "ability-score-improvement",
        "Feat choice",
        ["feats/ability-score-improvement", "feats/another-qualified-feat"],
        ["monk/asi/one-feat-at-each-occurrence"],
    )
    ch["values_by_level"] = [dict(level=i, value=1) for i in (4, 8, 12, 16)]
    pred(ch, "asi-eligibility", "monk/asi/chosen-feat-prerequisites", ["class/monk"])
    feat(
        "slow-fall",
        "Slow Fall",
        4,
        51,
        ["reaction/when-falling-reduce-fall-damage-by-multiplier-times-monk-level"],
        {"level-multiplier": 5},
    )
    feat(
        "extra-attack",
        "Extra Attack",
        5,
        51,
        ["attack-action-on-own-turn/replace-one-attack-with-two"],
        {"attacks": 2},
    )
    st = feat(
        "stunning-strike",
        "Stunning Strike",
        5,
        51,
        [
            "once-per-turn/on-hit-creature-with-monk-weapon-or-unarmed-strike",
            "target/constitution-save",
            "failed-save/stunned-until-start-of-own-next-turn",
            "successful-save/speed-halved-until-start-of-own-next-turn",
            "successful-save/next-attack-against-target-before-then-has-advantage",
        ],
        {"focus-cost": 1, "speed-divisor": 2},
    )
    pred(
        st,
        "stunning-strike-focus",
        "monk/stunning-strike/available-focus",
        ["resource/monk/focus-points"],
        {"points-required": 1},
    )
    emp = feat(
        "empowered-strikes",
        "Empowered Strikes",
        6,
        51,
        ["whenever-unarmed-strike-deals-damage/choose-type"],
    )
    choice(
        emp,
        "empowered-strikes",
        "Damage type",
        ["damage/force", "monk/unarmed-strike/normal-damage-type"],
    )
    ev = feat(
        "evasion",
        "Evasion",
        7,
        51,
        ["dexterity-save-for-half-damage/success-none-failure-half"],
    )
    pred(ev, "evasion-eligibility", "monk/evasion/not-incapacitated", ["class/monk"])
    ac = feat(
        "acrobatic-movement",
        "Acrobatic Movement",
        9,
        51,
        ["vertical-surfaces-and-liquids-on-own-turn/no-falling-during-movement"],
    )
    pred(
        ac,
        "acrobatic-movement-eligibility",
        "monk/acrobatic-movement/no-armor-or-shield",
        ["class/monk"],
    )
    hf = feat(
        "heightened-focus",
        "Heightened Focus",
        10,
        51,
        [
            "flurry/three-strikes-instead-of-two-for-one-focus",
            "patient-defense/paid-use-gains-two-martial-arts-dice-temporary-hp",
            "step-of-the-wind/paid-use-move-chosen-creature-with-self-until-end-of-turn",
            "step-of-the-wind/carried-creature-movement-no-opportunity-attacks",
        ],
        {
            "flurry-cost": 1,
            "flurry-strikes": 3,
            "patient-defense-temporary-hp-dice": 2,
            "step-of-the-wind-range-feet": 5,
        },
    )
    hch = choice(
        hf,
        "heightened-wind-target",
        "Heightened Step of the Wind target",
        [
            "monk/heightened-wind/no-passenger",
            "monk/heightened-wind/willing-large-or-smaller-creature",
        ],
    )
    pred(
        hch,
        "heightened-wind-eligibility",
        "monk/heightened-wind/paid-step-willing-large-or-smaller-within-range",
        ["choice/monk/step-of-the-wind", "resource/monk/focus-points"],
        {"range-feet": 5, "points-required": 1},
    )
    # The heading is p51; the Patient Defense and Step of the Wind continuation is p52.
    hf["required_mechanics"].append(
        M + "monk/heightened-focus/continued-on-pdf-page-52"
    )
    sr = feat(
        "self-restoration",
        "Self-Restoration",
        10,
        52,
        [
            "end-of-each-own-turn/remove-one-condition-from-self",
            "no-exhaustion-from-forgoing-food-and-drink",
        ],
        {"conditions-removed": 1},
    )
    choice(
        sr,
        "self-restoration",
        "Condition to remove",
        ["conditions/charmed", "conditions/frightened", "conditions/poisoned"],
    )
    de = feat(
        "deflect-energy",
        "Deflect Energy",
        13,
        52,
        ["deflect-attacks/any-damage-type-instead-of-bludgeoning-piercing-slashing"],
    )
    pred(
        de,
        "deflect-energy-feature",
        "monk/deflect-energy/modifies-deflect-attacks",
        ["feature/monk/deflect-attacks"],
    )
    ds = feat(
        "disciplined-survivor",
        "Disciplined Survivor",
        14,
        52,
        [
            "proficiency/all-saving-throws",
            "on-failed-saving-throw/expend-focus-reroll-must-use-new-roll",
        ],
        {"reroll-cost": 1},
    )
    pred(
        ds,
        "disciplined-survivor-focus",
        "monk/disciplined-survivor/available-focus",
        ["resource/monk/focus-points"],
        {"points-required": 1},
    )
    pf = feat(
        "perfect-focus",
        "Perfect Focus",
        15,
        52,
        [
            "on-initiative/when-not-using-uncanny-metabolism-regain-to-four-if-three-or-fewer"
        ],
        {"threshold": 3, "regain-to": 4},
    )
    pred(
        pf,
        "perfect-focus-eligibility",
        "monk/perfect-focus/initiative-not-using-metabolism-and-focus-at-most-threshold",
        ["feature/monk/uncanny-metabolism", "resource/monk/focus-points"],
        {"maximum-current-points": 3},
    )
    sd = feat(
        "superior-defense",
        "Superior Defense",
        18,
        52,
        [
            "start-of-own-turn/expend-focus",
            "resistance/all-damage-except-force",
            "ends/one-minute-or-incapacitated",
        ],
        {"focus-cost": 3, "duration-minutes": 1},
    )
    pred(
        sd,
        "superior-defense-focus",
        "monk/superior-defense/available-focus",
        ["resource/monk/focus-points"],
        {"points-required": 3},
    )
    eb = feat(
        "epic-boon",
        "Epic Boon",
        19,
        52,
        [
            "gain-epic-boon-or-another-qualified-feat",
            "recommended/boon-of-irresistible-offense",
        ],
    )
    ch = choice(
        eb,
        "epic-boon",
        "Epic Boon feat",
        ["feats/epic-boon", "feats/another-qualified-feat"],
    )
    pred(
        ch,
        "epic-boon-eligibility",
        "monk/epic-boon/chosen-feat-prerequisites",
        ["class/monk"],
    )
    feat(
        "body-and-mind",
        "Body and Mind",
        20,
        52,
        ["increase-dexterity-and-wisdom-scores"],
        {"increase": 4, "maximum": 25},
    )
    rows = [
        "martial-arts,unarmored-defense",
        "monks-focus,unarmored-movement,uncanny-metabolism",
        "deflect-attacks,monk-subclass",
        "ability-score-improvement,slow-fall",
        "extra-attack,stunning-strike",
        "empowered-strikes,subclass",
        "evasion",
        "ability-score-improvement",
        "acrobatic-movement",
        "heightened-focus,self-restoration",
        "subclass",
        "ability-score-improvement",
        "deflect-energy",
        "disciplined-survivor",
        "perfect-focus",
        "ability-score-improvement",
        "subclass",
        "superior-defense",
        "epic-boon",
        "body-and-mind",
    ]
    for i, fs in enumerate(rows, 1):
        r = add(
            "level",
            str(i),
            f"Monk level {i}",
            49,
            "classes/monk",
            cls["id"],
            ["monk/class-progression"],
            {
                "table/martial-arts-dice": 1,
                "table/martial-arts-die-sides": 6
                if i < 5
                else 8
                if i < 11
                else 10
                if i < 17
                else 12,
                "table/focus-points": 0 if i == 1 else i,
                "table/unarmored-movement-feet": 0
                if i == 1
                else 10
                if i < 6
                else 15
                if i < 10
                else 20
                if i < 14
                else 25
                if i < 18
                else 30,
            },
        )
        if i == 1:
            r["required_mechanics"] += [
                M + "monk/table/focus-points/absent",
                M + "monk/table/unarmored-movement/absent",
            ]
        r["level"] = i
        r["progression"] = dict(
            proficiency_bonus=2 + (i - 1) // 4,
            feature_refs=[
                features[f][1]["id"] for f in fs.split(",") if f != "subclass"
            ],
            feature_keys=[M + "monk/subclass-feature"]
            if "subclass" in fs.split(",")
            else [],
        )
        r["required_mechanics"] += r["progression"]["feature_keys"]
    return out


def compare(records):
    expected = source_records()
    assert len(records) == len(expected), "Monk record count"
    for actual, want in zip(records, expected):
        assert actual == want, (
            f"Monk semantic mismatch: {want['id']}\nactual={actual}\nexpected={want}"
        )


def contrary_controls(records):
    import copy

    controls = []
    for i, r in enumerate(records):
        for j, key in enumerate(r["required_mechanics"]):
            c = copy.deepcopy(records)
            replacement = M + "monk/contrary"
            c[i]["required_mechanics"][j] = replacement
            if c[i].get("requirement", {}).get("key") == key:
                c[i]["requirement"]["key"] = replacement
            if key in c[i].get("parameters", {}):
                c[i]["parameters"][replacement] = c[i]["parameters"].pop(key)
            for option in c[i].get("options", []):
                if key in option["quantities"]:
                    option["quantities"][replacement] = option["quantities"].pop(key)
            if c[i].get("progression"):
                c[i]["progression"]["feature_keys"] = [
                    replacement if k == key else k
                    for k in c[i]["progression"]["feature_keys"]
                ]
            controls.append(c)
        for key in r.get("parameters", {}):
            c = copy.deepcopy(records)
            c[i]["parameters"][key] += 1
            controls.append(c)
        if "requirement" in r:
            for key in r["requirement"]["parameters"]:
                c = copy.deepcopy(records)
                c[i]["requirement"]["parameters"][key] += 1
                controls.append(c)
            for j in range(len(r["requirement"]["record_refs"])):
                c = copy.deepcopy(records)
                c[i]["requirement"]["record_refs"][j] = C + "feature/monk/body-and-mind"
                controls.append(c)
        if r["kind"] == "choice":
            c = copy.deepcopy(records)
            c[i]["select"] += 1
            if c[i].get("values_by_level"):
                c[i]["values_by_level"][0]["value"] = c[i]["select"]
            controls.append(c)
            c = copy.deepcopy(records)
            c[i]["option_keys"].append(M + "monk/contrary")
            controls.append(c)
        for j in range(len(r.get("values_by_level", []))):
            if j:  # Initial count must still match capacity/select.
                c = copy.deepcopy(records)
                c[i]["values_by_level"][j]["value"] += 1
                controls.append(c)
        for j, option in enumerate(r.get("options", [])):
            for key in option["quantities"]:
                c = copy.deepcopy(records)
                c[i]["options"][j]["quantities"][key] += 1
                controls.append(c)
        if r.get("progression"):
            c = copy.deepcopy(records)
            c[i]["progression"]["proficiency_bonus"] = (
                3 if r["progression"]["proficiency_bonus"] == 2 else 2
            )
            controls.append(c)
    for c in controls:
        try:
            compare(c)
        except AssertionError:
            continue
        raise AssertionError("contrary metadata accepted")
    return controls


def verify(pdf):
    import hashlib, json, re, subprocess
    from pathlib import Path

    root = Path(__file__).resolve().parents[1]
    pin = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == pin, "PDF pin"
    # Read the actual pinned pages. No adjacent Champion/Open Hand definitions.
    columns = []
    for page in range(49, 53):
        for x in [310] if page == 49 else [0] if page == 52 else [0, 310]:
            columns.append(
                subprocess.check_output(
                    [
                        "pdftotext",
                        "-f",
                        str(page),
                        "-l",
                        str(page),
                        "-x",
                        str(x),
                        "-y",
                        "0",
                        "-W",
                        "310",
                        "-H",
                        "792",
                        "-layout",
                        str(pdf),
                        "-",
                    ],
                    text=True,
                )
            )
    text = " ".join(
        re.sub(r"-\n\s*", "", " ".join(columns).replace("\u00ad", "")).split()
    )
    anchors = [
        "Core Monk Traits",
        "Dexterity and Wisdom",
        "Choose 2: Acrobatics",
        "Simple weapons and Martial",
        "Choose one type of Artisan’s",
        "Spear, 5",
        "11 GP",
        "50 GP",
        "Bonus Unarmed Strike",
        "Simple Melee weapons",
        "Martial Melee weapons",
        "Light property",
        "Grapple or Shove",
        "Dexterity modifier instead of your Strength modifier",
        "10 plus your",
        "Short or Long Rest",
        "8 plus your Wisdom modifier and Proficiency Bonus",
        "two Unarmed Strikes as a Bonus Action",
        "Disengage and the Dodge",
        "Disengage and Dash",
        "jump distance is doubled for the turn",
        "roll Initiative",
        "Monk level plus the number rolled",
        "finish a Long Rest",
        "1d10 plus",
        "damage to 0",
        "within 5",
        "within 60",
        "Total Cover",
        "two rolls of",
        "same type dealt by the attack",
        "Monk level or lower",
        "levels 8, 12, and 16",
        "five times your Monk level",
        "twice instead of once",
        "Once per turn",
        "Constitution saving throw",
        "Stunned condition until the start of your next turn",
        "Speed is halved",
        "next attack roll",
        "Advantage",
        "Force damage or its normal damage type",
        "Dexterity saving throw to take only half",
        "no damage if you succeed",
        "half damage if you fail",
        "Incapacitated condition",
        "vertical surfaces and across liquids",
        "three Unarmed Strikes",
        "Temporary Hit Points equal to two rolls",
        "willing creature within 5",
        "Large or smaller",
        "end of your turn",
        "doesn’t provoke Opportunity Attacks",
        "Charmed, Frightened, or Poisoned",
        "food and drink",
        "Exhaustion",
        "any damage type",
        "all saving throws",
        "must use the new roll",
        "don’t use Uncanny Metabolism",
        "4 if you have 3 or fewer",
        "3 Focus Points",
        "1 minute",
        "except Force damage",
        "Boon of Irresistible Offense",
        "increase by 4",
        "maximum of 25",
    ]
    assert all(a in text for a in anchors), [a for a in anchors if a not in text]
    c = json.loads(
        (root / "internal/rules/srd521/testdata/character-monk.json").read_text()
    )
    assert c["schema_version"] == 1 and c["source"]["sha256"] == pin
    assert (
        c["source"]["license"] == "CC-BY-4.0"
        and c["source"]["attribution"]
        and c["source"]["changes"]
    )
    compare(c["records"])
    inventory = json.loads(
        (root / "docs/investigations/issue-74-character-inventory.json").read_text()
    )
    expected = {
        r["id"]: r["source_page"]
        for r in inventory["records"]
        if r["batch_key"] == "class-monk"
    }
    identities = [r for r in c["records"] if r["kind"] in ("class", "feature")]
    assert len(identities) == len(expected) == 23
    assert {
        s["inventory_id"]: s["page"] for r in identities for s in r["sources"]
    } == expected
    table = subprocess.check_output(
        ["pdftotext", "-f", "50", "-l", "50", "-layout", str(pdf), "-"], text=True
    ).split("Level 1: Martial Arts")[0]
    rows = []
    for line in table.splitlines():
        m = re.match(
            r"\s*(\d+)\s+\+(\d+)\s+(.+?)\s+(1d\d+)\s+(—|\d+)\s+(—|\+\d+ ft\.)\s*$", line
        )
        if m:
            rows.append(m.groups())
        elif rows and "Uncanny Metabolism" in line:
            a = list(rows[-1])
            a[2] += " Uncanny Metabolism"
            rows[-1] = tuple(a)
    assert len(rows) == 20, "all PDF table rows"
    levels = [r for r in c["records"] if r["kind"] == "level"]
    by_id = {r["id"]: r for r in c["records"]}
    for r, (level, pb, names, die, focus, speed) in zip(levels, rows):
        assert r["level"] == int(level) and r["progression"][
            "proficiency_bonus"
        ] == int(pb)
        got = [by_id[k]["name"] for k in r["progression"]["feature_refs"]]
        if r["progression"]["feature_keys"]:
            got += ["Subclass feature"]
        assert ", ".join(got) == names, (got, names)
        p = r["parameters"]
        assert p[M + "monk/table/martial-arts-dice"] == 1
        assert p[M + "monk/table/martial-arts-die-sides"] == int(die[2:])
        assert p.get(M + "monk/table/focus-points", 0) == (
            0 if focus == "—" else int(focus)
        )
        assert p.get(M + "monk/table/unarmored-movement-feet", 0) == (
            0 if speed == "—" else int(speed[1:].split()[0])
        )
    controls = contrary_controls(c["records"])
    print(
        f"Monk pp49-52: 23 identities, {len(c['records'])} complete semantic records, 20 actual PDF table rows; {len(controls)} contrary controls rejected"
    )


if __name__ == "__main__":
    import sys
    from pathlib import Path

    if len(sys.argv) == 3 and sys.argv[1] == "--controls":
        import json

        controls = json.loads(Path(sys.argv[2]).read_text())
        for records in controls:
            try:
                compare(records)
            except AssertionError:
                continue
            raise AssertionError("Go contrary graph accepted")
        print(f"{len(controls)} Go structurally valid controls rejected independently")
        raise SystemExit(0)
    if len(sys.argv) != 2:
        raise SystemExit("usage: verify_srd_monk.py /path/to/SRD_CC_v5.2.1.pdf")
    verify(Path(sys.argv[1]))
