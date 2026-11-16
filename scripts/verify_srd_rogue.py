#!/usr/bin/env python3
"""Independent offline Rogue source graph comparator, pinned SRD pp61–63.

Source clauses below are adapted to opaque mechanic keys; full records, including
absent fields, are compared. No Go invocation, runtime rule or network access.
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
M = "srd-5.2.1:mechanic/"
K = M + "rogue/"
DATA = ROOT / "internal/rules/srd521/testdata/character-rogue.json"


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    layout = subprocess.check_output(["pdftotext", "-f", "61", "-l", "63", "-layout", str(pdf), "-"], text=True)
    # Read columns independently to keep sentences intact across column breaks.
    columns = []
    for page in (61, 62, 63):
        for x in (0, 310):
            columns.append(subprocess.check_output(["pdftotext", "-f", str(page), "-l", str(page),
                "-x", str(x), "-y", "0", "-W", "310", "-H", "792", "-layout", str(pdf), "-"], text=True))
    flat = " ".join(re.sub(r"-\n\s*", "", " ".join(columns)).replace("\u00ad", "").split())
    records, by = [], {}

    def add(kind, slug, parent, name=None, level=0, page=0):
        ident = C + kind + "/rogue" + ("/" + slug if slug else "")
        if kind == "class":
            src = dict(inventory_id="srd-5.2.1:planning/classes/rogue", page=61)
        elif kind == "feature":
            src = dict(inventory_id=f"srd-5.2.1:planning/class-features/rogue-level-{by[parent]['level']}-{slug}", page=page)
        else:
            src = copy.deepcopy(by[parent]["sources"][0])
        r = dict(id=ident, kind=kind, name=name or slug.replace("-", " ").title(), sources=[src],
            parent_id=parent, level=level, choice_refs=[], resource_refs=[], prerequisite_refs=[],
            required_mechanics=[], select=0, option_keys=[], capacity=0)
        records.append(r); by[ident] = r
        if kind in ("choice", "resource", "prerequisite"):
            by[parent][kind + "_refs"].append(ident)
        return ident

    def rule(owner, clause, words, numbers=None):
        assert clause in flat, f"source clause: {clause}"
        r = by[owner]
        r["required_mechanics"] += [K + w for w in words.split()]
        if numbers:
            r.setdefault("parameters", {}).update({K + key: value for key, value in numbers.items()})
            r["required_mechanics"] += [K + key for key in numbers]

    def choice(slug, owner, count, options):
        q = add("choice", slug, owner)
        by[q].update(select=count, option_keys=[M + o for o in options.split()], required_mechanics=[K + slug])
        return q

    def pred(slug, owner, key, refs, numbers=None, option=None):
        p = add("prerequisite", slug, owner)
        by[p]["required_mechanics"] = [K + key]
        by[p]["requirement"] = dict(key=K + key, record_refs=refs, parameters=numbers or {})
        if option:
            by[p]["required_mechanics"].append(K + "predicate/conditional-on-option/" + option)
        return p

    cls = add("class", "", "", "Rogue")
    rule(cls, "Dexterity and Intelligence", "primary-ability/dexterity saves/dexterity saves/intelligence weapons/simple weapons/martial-finesse-or-light tools/thieves-tools armor/light level-1/all-core-traits-and-level-1-features multiclass/hit-die-one-class-skill-thieves-tools-light-armor-and-level-1-features", {"hit-die-sides": 8})
    skills = "acrobatics athletics deception insight intimidation investigation perception persuasion sleight-of-hand stealth".split()
    choice("skills", cls, 4, " ".join("skills/" + s for s in skills))
    choice("multiclass-skill", cls, 1, " ".join("skills/" + s for s in skills))
    eq = choice("starting-equipment", cls, 1, "rogue/equipment/a rogue/equipment/b")
    bundles = [{"leather-armor": 1, "dagger": 2, "shortsword": 1, "shortbow": 1, "arrows": 20, "quiver": 1, "thieves-tools": 1, "burglars-pack": 1, "gp": 8}, {"gp": 100}]
    by[eq]["options"] = []
    for key, bundle in zip(by[eq]["option_keys"], bundles):
        quantities = {M + ("currency/" if item == "gp" else "equipment/") + item: n for item, n in bundle.items()}
        by[eq]["options"].append(dict(key=key, quantities=quantities))
        for k in quantities:
            if k not in by[eq]["required_mechanics"]: by[eq]["required_mechanics"].append(k)
    assert "100 GP" in flat and "and 8 GP" in flat
    # Extract all table cells, not a transcription of only the feature headings.
    rows = re.findall(r"^\s*(\d+)\s+\+(\d)\s+(.+?)\s+(\d+)d6\s*$", layout, re.M)
    assert len(rows) == 20, "20 Rogue table rows"
    table = {}
    for level, pb, features, dice in rows:
        n = int(level)
        table[n] = features
        r = add("level", str(n), cls, f"Rogue level {n}", level=n)
        by[r]["required_mechanics"] = [K + "class-progression", K + "sneak-attack/dice-count", K + "sneak-attack/dice-sides"]
        by[r]["parameters"] = {K + "sneak-attack/dice-count": int(dice), K + "sneak-attack/dice-sides": 6}
        by[r]["progression"] = dict(proficiency_bonus=int(pb), feature_refs=[], feature_keys=[])
    features = [
        (1, "expertise", 61), (1, "sneak-attack", 61), (1, "thieves-cant", 62), (1, "weapon-mastery", 62),
        (2, "cunning-action", 62), (3, "rogue-subclass", 62), (3, "steady-aim", 62), (4, "ability-score-improvement", 63),
        (5, "cunning-strike", 63), (5, "uncanny-dodge", 63), (7, "evasion", 63), (7, "reliable-talent", 63),
        (11, "improved-cunning-strike", 63), (14, "devious-strikes", 63), (15, "slippery-mind", 63),
        (18, "elusive", 63), (19, "epic-boon", 63), (20, "stroke-of-luck", 63)]
    f = {}
    for level, slug, page in features:
        name = "Thieves’ Cant" if slug == "thieves-cant" else slug.replace("-", " ").title().replace(" Of ", " of ")
        f[slug] = add("feature", slug, C + f"level/rogue/{level}", name, page=page)
    names = {by[v]["name"]: v for v in f.values()}
    for n, text in table.items():
        row = by[C + f"level/rogue/{n}"]
        for name in text.split(", "):
            if name == "Subclass feature":
                row["progression"]["feature_keys"].append(K + "subclass-feature")
                row["required_mechanics"].append(K + "subclass-feature")
            else: row["progression"]["feature_refs"].append(names[name])
    ex = f["expertise"]
    rule(ex, "Expertise in two of your skill proficiencies", "expertise/skill-proficiencies-only expertise/recommend-sleight-of-hand-and-stealth-if-proficient", {"expertise/initial-count": 2, "expertise/additional-level": 6, "expertise/additional-count": 2})
    for suffix in ("initial", "level-6"):
        q = choice("expertise-" + suffix, ex, 2, "skills/own-proficiencies")
        pred("expertise-" + suffix, q, "expertise/requires-own-skill-proficiency", [cls, ex], {"minimum-level": 1 if suffix == "initial" else 6})
    sa = f["sneak-attack"]
    rule(sa, "Once per turn", "sneak-attack/once-per-turn sneak-attack/hit-one-creature-with-attack-roll sneak-attack/finesse-or-ranged-weapon sneak-attack/same-damage-type-as-weapon sneak-attack/table-scaling")
    eligibility = choice("sneak-attack-eligibility", sa, 1, "rogue/sneak-attack/advantage rogue/sneak-attack/ally-within-five-feet-not-incapacitated-and-no-disadvantage")
    pred("sneak-attack-eligibility", sa, "sneak-attack/requires-eligible-hit", [sa, eligibility], {"ally-distance-feet": 5})
    rule(sa, "the ally doesn’t have the Incapacitated condition", "sneak-attack/ally-must-not-be-incapacitated sneak-attack/ally-alternative-requires-no-disadvantage")
    rule(f["thieves-cant"], "know Thieves’ Cant and one other language", "language/thieves-cant")
    choice("additional-language", f["thieves-cant"], 1, "languages/character-creation-tables")
    wm = f["weapon-mastery"]
    rule(wm, "Whenever you finish a Long Rest", "weapon-mastery/change-kinds-on-long-rest weapon-mastery/recommend-daggers-shortbows")
    q = choice("weapon-mastery", wm, 2, "weapons/own-proficiencies")
    pred("weapon-mastery-proficiency", q, "weapon-mastery/requires-weapon-proficiency", [cls])
    rule(f["cunning-action"], "On your turn", "cunning-action/bonus-action-on-own-turn")
    choice("cunning-action", f["cunning-action"], 1, "actions/dash actions/disengage actions/hide")
    rule(f["rogue-subclass"], "your Rogue level or lower", "subclass/grant-features-at-own-class-level-or-lower")
    choice("rogue-subclass", f["rogue-subclass"], 1, "rogue/subclasses rogue/subclasses/thief")
    aim = f["steady-aim"]
    rule(aim, "your next attack roll on the current turn", "steady-aim/bonus-action steady-aim/advantage-next-attack-roll-current-turn steady-aim/speed-zero-until-current-turn-end")
    pred("steady-aim-unmoved", aim, "steady-aim/requires-no-movement-this-turn", [cls], {"maximum-feet-moved": 0})
    for slug, opts, clause in [("ability-score-improvement", "feats/ability-score-improvement feats/qualified", "8, 10, 12, and 16"), ("epic-boon", "feats/epic-boon feats/qualified", "Boon of the Night Spirit is recommended")]:
        rule(f[slug], clause, slug + "/qualified-feat-choice")
        q = choice(slug, f[slug], 1, opts)
        pred(slug + "-qualification", q, "feat/requires-qualification", [cls], option="feats/qualified")
    rule(f["epic-boon"], "Boon of the Night Spirit", "epic-boon/recommend-boon-of-the-night-spirit")
    cs = f["cunning-strike"]
    rule(cs, "remove the die before rolling", "cunning-strike/when-dealing-sneak-attack cunning-strike/forgo-dice-before-roll cunning-strike/effect-immediately-after-attack-damage cunning-strike/save-dc-eight-plus-dexterity-modifier-plus-proficiency-bonus", {"cunning-strike/base-effect-limit": 1, "cunning-strike/save-dc-base": 8})
    effects = choice("cunning-strike", cs, 1, "rogue/cunning-strike/poison rogue/cunning-strike/trip rogue/cunning-strike/withdraw")
    ds = f["devious-strikes"]
    rule(ds, "now among your Cunning Strike options", "devious-strikes/add-to-cunning-strike-options")
    extra = choice("devious-strikes", ds, 1, "rogue/cunning-strike/daze rogue/cunning-strike/knock-out rogue/cunning-strike/obscure")
    pred("devious-strikes-options", extra, "devious-strikes/extends-cunning-strike", [effects, cs])
    variants = [
        ("poison", effects, 1, "Constitution saving throw", "constitution-save poisoned-on-failure repeat-save-end-each-target-turn end-on-success requires-poisoners-kit-on-person", {"duration-minutes": 1}),
        ("trip", effects, 1, "Large or smaller", "dexterity-save prone-on-failure target-large-or-smaller", {}),
        ("withdraw", effects, 1, "half your Speed", "immediately-after-attack move-up-to-half-speed no-opportunity-attacks", {"speed-divisor": 2}),
        ("daze", extra, 2, "move or take an action", "constitution-save on-failure-next-turn-only-one-of-move-action-bonus-action", {}),
        ("knock-out", extra, 6, "until it takes any damage", "constitution-save unconscious-on-failure ends-on-any-damage repeat-save-end-each-target-turn end-on-success", {"duration-minutes": 1}),
        ("obscure", extra, 3, "Blinded condition", "dexterity-save blinded-on-failure until-end-target-next-turn", {}),
    ]
    # Bind each variant to its own source block, including the cost and save.
    ends = {"poison": "Trip", "trip": "Withdraw", "withdraw": "Level 5: Uncanny Dodge",
            "daze": "Knock Out", "knock-out": "Obscure", "obscure": "Level 15: Slippery Mind"}
    for slug, owner, cost, clause, words, numbers in variants:
        label = slug.replace("-", " ").title()
        match = re.search(re.escape(label) + r" \(Cost: (\d+)d6\)\. (.*?)" + re.escape(ends[slug]), flat)
        assert match, f"source variant block: {slug}"
        extracted_cost, block = int(match[1]), match[2]
        assert extracted_cost == cost and clause in block, f"cost/condition in {slug}"
        if "constitution-save" in words:
            assert "Constitution saving throw" in block
        if "dexterity-save" in words:
            assert "Dexterity saving throw" in block
        if "duration-minutes" in numbers:
            duration = re.search(r"for (\d+) minute", block)
            assert duration and int(duration[1]) == numbers["duration-minutes"]
        if slug == "poison":
            assert "Poisoner’s Kit on your person" in block
        q = choice("strike-" + slug, owner, 1, "rogue/cunning-strike/" + slug)
        pred("strike-" + slug + "-selected", q, "cunning-strike/" + slug + "/selected", [owner, sa], {"cost-d6": extracted_cost}, "rogue/cunning-strike/" + slug)
        rule(q, clause, " ".join("cunning-strike/" + slug + "/" + w for w in words.split()))
        if slug == "daze":
            turn = choice("daze-restricted-turn", q, 1, "rogue/daze/move rogue/daze/action rogue/daze/bonus-action")
            rule(turn, "on its next turn", "daze/only-on-failed-save-next-turn")
        if numbers:
            pred("strike-" + slug + "-numbers", q, "cunning-strike/" + slug + "/limits", [sa], numbers)
    rule(f["improved-cunning-strike"], "paying the die cost for each effect", "improved-cunning-strike/pay-each-effect-cost", {"improved-cunning-strike/maximum-effects": 2})
    pred("improved-cunning-strike-operands", f["improved-cunning-strike"], "improved-cunning-strike/requires-sneak-attack-and-strike", [sa, effects])
    rule(f["uncanny-dodge"], "round down", "uncanny-dodge/reaction uncanny-dodge/attacker-you-can-see-hits-with-attack-roll uncanny-dodge/halve-attacks-damage-round-down", {"uncanny-dodge/damage-divisor": 2})
    rule(f["evasion"], "You can’t use this feature", "evasion/dexterity-save-for-half-damage evasion/success-no-damage evasion/failure-half-damage evasion/unavailable-while-incapacitated")
    rule(f["reliable-talent"], "skill or tool proficiencies", "reliable-talent/ability-check-using-own-skill-or-tool-proficiency", {"reliable-talent/maximum-d20-roll": 9, "reliable-talent/replacement-roll": 10})
    rule(f["slippery-mind"], "Wisdom and Charisma", "saves/wisdom saves/charisma")
    rule(f["elusive"], "unless you have the Incapacitated", "elusive/no-attack-roll-advantage-against-self-unless-incapacitated")
    luck = f["stroke-of-luck"]
    rule(luck, "If you fail a D20 Test", "stroke-of-luck/failed-d20-test stroke-of-luck/turn-roll-into-twenty", {"stroke-of-luck/replacement-roll": 20})
    res = add("resource", "stroke-of-luck", luck)
    by[res].update(capacity=1, required_mechanics=[K + "stroke-of-luck/recover-on-short-or-long-rest"])
    assert "until you finish a Short or Long Rest" in flat
    return records


def compare(records, expected):
    assert len(records) == len(expected), "record count"
    actual = {r["id"]: r for r in records}
    assert len(actual) == len(records), "duplicate record"
    for r in expected:
        assert actual.get(r["id"]) == r, f"semantic fields: {r['id']}"


def controls(records):
    """Structurally valid contrary facts; Go checks structure before rejection."""
    out = []
    def change(i, fn):
        c = copy.deepcopy(records); fn(c[i]); out.append(c)
    for i, r in enumerate(records):
        for key in r.get("parameters", {}):
            change(i, lambda x, key=key: x["parameters"].__setitem__(key, x["parameters"][key] + 1))
        for key in r.get("requirement", {}).get("parameters", {}):
            change(i, lambda x, key=key: x["requirement"]["parameters"].__setitem__(key, x["requirement"]["parameters"][key] + 1))
        refs = r.get("requirement", {}).get("record_refs", [])
        for j, ref in enumerate(refs):
            other = C + ("feature/rogue/evasion" if ref != C + "feature/rogue/evasion" else "feature/rogue/steady-aim")
            if other not in refs:
                change(i, lambda x, j=j, other=other: x["requirement"]["record_refs"].__setitem__(j, other))
        if r["kind"] == "choice" and len(r["option_keys"]) > 1:
            change(i, lambda x: x["option_keys"].pop()) if not r.get("options") else None
        if r["kind"] == "resource":
            change(i, lambda x: x.__setitem__("capacity", 2))
        if r.get("progression"):
            change(i, lambda x: x["progression"].__setitem__("proficiency_bonus", 3 if x["progression"]["proficiency_bonus"] != 3 else 4))
            change(i, lambda x: (x["required_mechanics"].append(K + "invented-scaling"), x["parameters"].__setitem__(K + "invented-scaling", 2)))
        for j, key in enumerate(r["required_mechanics"]):
            if key not in r.get("parameters", {}) and key != r.get("requirement", {}).get("key") and not r.get("options") and not r.get("progression"):
                change(i, lambda x, j=j: x["required_mechanics"].__setitem__(j, K + "contrary-source"))
                if len(r["required_mechanics"]) > 1:
                    change(i, lambda x, j=j: x["required_mechanics"].pop(j))
    return out


def verify(pdf, candidate=DATA):
    expected = source_records(pdf)
    c = json.loads(candidate.read_text())
    assert set(c) == {"schema_version", "source", "records"}, "catalogue keys"
    assert c["schema_version"] == 1
    pin_source = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())["source"]
    assert set(c["source"]) == set(pin_source), "source keys"
    for key in pin_source:
        if key != "changes":
            assert c["source"][key] == pin_source[key], f"source {key}"
    assert c["source"]["changes"] == "Rogue source identities, all 20 progression rows, nested choices, prerequisites, costs, limits and recovery on PDF pp61-63 adapted into unavailable offline metadata. Mechanic keys and graph are Edda additions. Thief, spell and feat definitions remain opaque. No executable support imported."
    compare(c["records"], expected)
    mutations = controls(c["records"])
    for mutated in mutations:
        try: compare(mutated, expected)
        except AssertionError: pass
        else: raise AssertionError("contrary source accepted")
    print(f"Rogue: {len(expected)} semantic records; 20 extracted table rows; {len(mutations)} contrary controls rejected; unavailable metadata only")


if __name__ == "__main__":
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) > 2 else DATA)
