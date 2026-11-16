#!/usr/bin/env python3
"""Standalone pinned-p46 semantic oracle. No Go, catalogue hashes or network needed.

Facts below are an explicit adaptation of the four tables and five features on
p46. Compare all fields, including absent fields; canonical Druid is external.
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
K = M + "druid/circle-of-the-land/"
S = C + "subclass/druid/circle-of-the-land"
F = C + "feature/druid/circle-of-the-land/"
P = C + "prerequisite/druid/circle-of-the-land/"
Q = C + "choice/druid/circle-of-the-land/"
R = C + "resource/druid/circle-of-the-land/"


def source_records():
    records = []

    def record(kind, slug, name, parent, source, mechanics):
        prefix = {"subclass": S.rsplit("/", 1)[0] + "/", "feature": F,
                  "prerequisite": P, "choice": Q, "resource": R}[kind]
        r = dict(id=prefix + slug, kind=kind, name=name,
                 sources=[dict(inventory_id="srd-5.2.1:planning/" + source, page=46)],
                 parent_id=parent, level=0, choice_refs=[], resource_refs=[],
                 prerequisite_refs=[], required_mechanics=mechanics,
                 select=0, option_keys=[], capacity=0)
        records.append(r)
        return r

    def predicate(owner, slug, key, refs, numbers):
        r = record("prerequisite", slug, slug.replace("-", " ").title(), owner["id"],
                   owner["sources"][0]["inventory_id"].split("planning/")[1], [key])
        r["requirement"] = dict(key=key, record_refs=refs, parameters=numbers)
        owner["prerequisite_refs"].append(r["id"])

    def feature(slug, name, level, words, numbers=None):
        r = record("feature", slug, name, S,
                   f"subclass-features/druid-level-{level}-{slug}", [K + slug + "/" + w for w in words])
        if numbers:
            r["parameters"] = {K + slug + "/" + k: v for k, v in numbers.items()}
            r["required_mechanics"] += list(r["parameters"])
        predicate(r, slug + "-level", M + "prerequisite/class-level", [C + "class/druid"], {"minimum-level": level})
        return r

    sub = record("subclass", "circle-of-the-land", "Circle of the Land", C + "class/druid",
                 "subclasses/druid-circle-of-the-land", [M + "druid/subclasses/circle-of-the-land"])
    predicate(sub, "circle-of-the-land-level", M + "prerequisite/class-level", [C + "class/druid"], {"minimum-level": 3})
    spells = feature("circle-of-the-land-spells", "Circle of the Land Spells", 3,
                     ["choose-one-land-on-finishing-long-rest", "prepare-current-land-spells-at-druid-level-and-lower",
                      "conditional-tables-not-cumulative-across-lands"])
    choice = record("choice", "land", "Land", spells["id"], "subclass-features/druid-level-3-circle-of-the-land-spells",
                    [K + "land/choose-on-finishing-long-rest"])
    choice["select"] = 1
    choice["option_keys"] = [K + "land/" + x for x in ("arid", "polar", "temperate", "tropical")]
    spells["choice_refs"] = [choice["id"]]
    tables = {
        "arid": ["blur burning-hands fire-bolt", "fireball", "blight", "wall-of-stone"],
        "polar": ["fog-cloud hold-person ray-of-frost", "sleet-storm", "ice-storm", "cone-of-cold"],
        "temperate": ["misty-step shocking-grasp sleep", "lightning-bolt", "freedom-of-movement", "tree-stride"],
        "tropical": ["acid-splash ray-of-sickness web", "stinking-cloud", "polymorph", "insect-plague"],
    }
    for land, rows in tables.items():
        r = record("feature", "spells-" + land, land.title() + " Land spells", S,
                   "subclass-features/druid-level-3-circle-of-the-land-spells",
                   [K + "spells-" + land + "/only-while-current-land-selected"])
        r["always_prepared"] = [dict(level=n, spell_keys=[M + "spells/" + s for s in row.split()])
                                for n, row in zip((3, 5, 7, 9), rows)]
        r["required_mechanics"] += [s for row in r["always_prepared"] for s in row["spell_keys"]]
        predicate(r, "spells-" + land + "-selected", K + "land/" + land,
                  [choice["id"], spells["id"], C + "class/druid"], {"minimum-level": 3})
    aid = feature("lands-aid", "Land’s Aid", 3,
        ["magic-action", "choose-point", "chosen-creatures-in-sphere/constitution-save-against-own-spell-save-dc",
         "failed-save/necrotic-damage", "successful-save/half-damage", "one-chosen-creature-in-sphere/regains-hit-points",
         "expend-canonical-wild-shape"],
        {"range-feet": 60, "sphere-radius-feet": 10, "wild-shape-uses-expended": 1,
         "dice-sides": 6, "dice-at-level-3": 2, "dice-at-level-10": 3, "dice-at-level-14": 4})
    predicate(aid, "lands-aid-wild-shape", K + "lands-aid/requires-unexpended-wild-shape",
              [C + "resource/druid/wild-shape"], {"minimum-uses": 1})
    recovery = feature("natural-recovery", "Natural Recovery", 6,
        ["cast-one-prepared-circle-spell-of-level-1-or-higher-without-slot", "free-cast/long-rest-recovery",
         "on-finishing-short-rest/choose-expended-slots", "slot-recovery/combined-level-at-most-half-druid-level-rounded-up",
         "slot-recovery/once-until-long-rest"], {"minimum-free-spell-level": 1, "maximum-recovered-slot-level": 5,
         "recovery-budget-divisor": 2, "free-casts-per-long-rest": 1, "slot-recoveries-per-long-rest": 1})
    predicate(recovery, "natural-recovery-operands", K + "natural-recovery/requires-prepared-circle-spell-or-expended-druid-slots",
              [spells["id"], C + "class/druid"] + [C + f"resource/druid/spell-slot-{n}" for n in range(1, 10)], {})
    for slug in ("free-cast", "slot-recovery"):
        r = record("resource", "natural-recovery-" + slug, "Natural Recovery " + slug, recovery["id"],
                   "subclass-features/druid-level-6-natural-recovery", [K + "natural-recovery/" + slug + "/recover-on-long-rest"])
        r["capacity"] = 1
        recovery["resource_refs"].append(r["id"])
    ward = feature("natures-ward", "Nature’s Ward", 10,
                  ["immune-to-poisoned-condition", "resistance/current-land/arid/fire", "resistance/current-land/polar/cold",
                   "resistance/current-land/temperate/lightning", "resistance/current-land/tropical/poison"])
    predicate(ward, "natures-ward-current-land", K + "natures-ward/current-land-determines-resistance", [choice["id"]], {})
    sanctuary = feature("natures-sanctuary", "Nature’s Sanctuary", 14,
        ["magic-action", "spectral-trees-and-vines/cube-on-ground", "ends-on-incapacitated-or-death",
         "self-and-allies-in-area/half-cover", "allies-in-area/gain-current-natures-ward-resistance",
         "bonus-action/move-cube-to-ground-within-range", "expend-canonical-wild-shape"],
        {"range-feet": 120, "cube-side-feet": 15, "duration-minutes": 1, "move-feet": 60, "wild-shape-uses-expended": 1})
    predicate(sanctuary, "natures-sanctuary-operands", K + "natures-sanctuary/requires-unexpended-wild-shape-and-current-ward",
              [C + "resource/druid/wild-shape", ward["id"]], {"minimum-uses": 1})
    return records


def compare(records):
    expected = source_records()
    assert len(records) == len(expected), "p46 record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"p46 semantic mismatch: {want['id']}\nactual={actual}\nexpected={want}"


def controls(records):
    """Same structurally legal mutations used by Go's schema and source proofs."""
    result = []
    for i, r in enumerate(records):
        for j, old in enumerate(r["required_mechanics"]):
            c = copy.deepcopy(records)
            new = K + "contrary"
            c[i]["required_mechanics"][j] = new
            if c[i].get("requirement", {}).get("key") == old:
                c[i]["requirement"]["key"] = new
            if old in c[i].get("parameters", {}):
                c[i]["parameters"][new] = c[i]["parameters"].pop(old)
            for row in c[i].get("always_prepared", []):
                row["spell_keys"] = [new if k == old else k for k in row["spell_keys"]]
            result.append(c)
        for key in r.get("parameters", {}):
            c = copy.deepcopy(records); c[i]["parameters"][key] += 1; result.append(c)
        for key in r.get("requirement", {}).get("parameters", {}):
            c = copy.deepcopy(records); c[i]["requirement"]["parameters"][key] += 1; result.append(c)
        for j in range(len(r.get("requirement", {}).get("record_refs", []))):
            c = copy.deepcopy(records); c[i]["requirement"]["record_refs"][j] = C + "feature/druid/spellcasting"; result.append(c)
        if r["kind"] == "choice":
            c = copy.deepcopy(records); c[i]["select"] += 1; result.append(c)
            c = copy.deepcopy(records); c[i]["option_keys"].pop(); result.append(c)
        if r["kind"] == "resource":
            c = copy.deepcopy(records); c[i]["capacity"] += 1; result.append(c)
        for j in range(len(r.get("always_prepared", []))):
            c = copy.deepcopy(records); c[i]["always_prepared"][j]["level"] += 1; result.append(c)
        if r["kind"] == "feature":
            c = copy.deepcopy(records)
            old = r["sources"][0]["inventory_id"]
            other = "druid-level-3-lands-aid" if "natures-ward" in old else "druid-level-10-natures-ward"
            for record in c:
                if record["sources"][0]["inventory_id"] == old:
                    record["sources"][0]["inventory_id"] = "srd-5.2.1:planning/subclass-features/" + other
            result.append(c)
        if r.get("always_prepared"):
            c = copy.deepcopy(records)
            c[i]["always_prepared"][0]["spell_keys"], c[i]["always_prepared"][1]["spell_keys"] = c[i]["always_prepared"][1]["spell_keys"], c[i]["always_prepared"][0]["spell_keys"]
            result.append(c)
    return result


def verify(pdf, catalogue=None):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin mismatch"
    columns = [subprocess.check_output(["pdftotext", "-f", "46", "-l", "46", "-x", str(x),
               "-y", "0", "-W", "310", "-H", "792", "-layout", str(pdf), "-"], text=True) for x in (0, 310)]
    text = " ".join(re.sub(r"-\n\s*", "", " ".join(columns).replace("\u00ad", "")).split())
    # Pin anchors establish the inspected page; full graph equality supplies the semantic proof.
    for phrase in ("Circle of the Land", "Land’s Aid", "Natural Recovery", "Nature’s Ward", "Nature’s Sanctuary",
                   "half your Druid level (round up)", "none of them can be level 6+", "Poisoned condition",
                   "Incapacitated condition or die", "Half Cover", "3d6", "4d6"):
        assert phrase in text, phrase
    c = json.loads((catalogue or ROOT / "internal/rules/srd521/testdata/character-circle-of-the-land.json").read_text())
    assert c["schema_version"] == 1 and c["source"]["sha256"] == PIN
    assert c["source"]["license"] == "CC-BY-4.0" and c["source"]["attribution"] and c["source"]["changes"]
    compare(c["records"])
    inventory = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]: r["source_page"] for r in inventory["records"] if r["batch_key"] == "subclass-druid"}
    assert len(expected) == 6
    assert {s["inventory_id"]: s["page"] for r in c["records"] for s in r["sources"]} == expected
    from verify_srd_druid import verify as verify_druid
    verify_druid(pdf)
    mutations = controls(c["records"])
    for variant in mutations:
        try:
            compare(variant)
        except AssertionError:
            continue
        raise AssertionError("contrary source accepted")
    print(f"p46: six inventory identities, {len(c['records'])} complete records; {len(mutations)} contrary controls rejected")


if __name__ == "__main__":
    if len(sys.argv) not in (2, 3):
        raise SystemExit("usage: verify_srd_circle_of_the_land.py PINNED_PDF [CATALOGUE]")
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) == 3 else None)
