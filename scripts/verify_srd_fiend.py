#!/usr/bin/env python3
"""Independent offline p76 source-fact comparison for #318; no Go or network."""
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
BASE = "warlock/fiend-patron/"


def source_records():
    out = []
    def record(kind, slug, name, source, parent, keys=(), nums=None):
        r = dict(id=C+kind+"/"+BASE+slug, kind=kind, name=name,
                 sources=[dict(inventory_id=P+source, page=76)], parent_id=parent,
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
        r["sources"][0]["page"] = owner["sources"][0]["page"]
        r["requirement"] = dict(key=M+BASE+key, record_refs=[C+x for x in refs], parameters=nums or {})
        owner["prerequisite_refs"].append(r["id"])
    subclass = record("subclass", "identity", "Fiend Patron", "subclasses/warlock-fiend-patron", C+"class/warlock", ["subclass"])
    pred(subclass,"level","class-level",["class/warlock"],{"minimum-level":3})
    def feature(slug,name,level,keys,nums=None):
        r = record("feature",slug,name,f"subclass-features/warlock-level-{level}-{slug}",subclass["id"],keys,nums)
        pred(r,"level","class-level",["class/warlock"],{"minimum-level":level})
        return r
    blessing = feature("dark-ones-blessing","Dark One’s Blessing",3,[
        "blessing/when-you-reduce-enemy-to-zero-hit-points",
        "blessing/alternative/someone-else-reduces-enemy-within-range-of-you-to-zero-hit-points",
        "blessing/self-temporary-hit-points/charisma-modifier-plus-warlock-level-minimum-one"],
        {"blessing/other-killer-enemy-range-feet":10,"blessing/temporary-hit-points-minimum":1})
    pred(blessing,"formula","blessing/formula-uses-canonical-warlock-level",["class/warlock"])
    spells = feature("fiend-spells","Fiend Spells",3,["fiend-spells/thereafter-always-prepared-from-specified-warlock-level"])
    spells["always_prepared"] = [dict(level=n,spell_keys=[M+"spells/"+x for x in names]) for n,names in [
        (3,["burning-hands","command","scorching-ray","suggestion"]),
        (5,["fireball","stinking-cloud"]),(7,["fire-shield","wall-of-fire"]),(9,["geas","insect-plague"])]]
    for row in spells["always_prepared"]:
        spells["required_mechanics"] += row["spell_keys"]
    pred(spells,"preparation","fiend-spells/canonical-pact-magic-preparation-and-spellcasting",["feature/warlock/pact-magic","choice/warlock/prepared-spells"])
    feature("dark-ones-own-luck","Dark One’s Own Luck",6,[
        "luck/optional-on-your-ability-check-or-saving-throw","luck/add-die-to-your-roll",
        "luck/after-seeing-roll-before-any-roll-effects","luck/uses-equal-charisma-modifier-minimum-one",
        "luck/recover-all-expended-uses-on-finish-long-rest"],
        {"luck/dice":1,"luck/die-sides":10,"luck/minimum-uses":1,"luck/maximum-uses-per-roll":1})
    resilience = feature("fiendish-resilience","Fiendish Resilience",10,[
        "resilience/choose-on-finish-short-or-long-rest","resilience/self-resistance-to-chosen-type",
        "resilience/until-choose-different-type-with-this-feature"])
    choice = record("choice","resistance-type","Fiendish Resilience damage type",
                    "subclass-features/warlock-level-10-fiendish-resilience",resilience["id"],
                    ["resilience/choose-one-damage-type-other-than-force"])
    choice["select"] = 1
    choice["option_keys"] = [M+BASE+"resilience/damage-type/"+x for x in
        ["acid","bludgeoning","cold","fire","lightning","necrotic","piercing","poison","psychic","radiant","slashing","thunder"]]
    resilience["choice_refs"].append(choice["id"])
    hurl = feature("hurl-through-hell","Hurl Through Hell",14,[
        "hurl/optional-on-hit-creature-with-attack-roll","hurl/target-charisma-save-against-your-spell-save-dc",
        "hurl/on-failed-save/target-disappears-through-lower-planes","hurl/on-failed-save/psychic-damage-only-if-target-not-fiend",
        "hurl/on-failed-save/incapacitated-until-end-of-your-next-turn",
        "hurl/return-at-end-of-your-next-turn/previous-space-or-nearest-unoccupied-space",
        "hurl/recover-use-on-finish-long-rest","hurl/restore-use/expend-pact-magic-slot-no-action"],
        {"hurl/maximum-uses-per-turn":1,"hurl/damage-dice":8,"hurl/damage-die-sides":10,"hurl/uses":1,"hurl/restore-slot-cost":1})
    pred(hurl,"save","hurl/canonical-charisma-spell-save-dc",["feature/warlock/pact-magic"])
    pred(hurl,"restore","hurl/restore-expended-use-with-canonical-pact-slot",["resource/warlock/pact-slots","feature/warlock/pact-magic"],{"slots-required":1})
    return out

def compare(records):
    expected = source_records()
    assert len(records) == len(expected), "p76 record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"p76 semantic mismatch {want['id']}: {actual} != {want}"


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    def extract(page, right=False):
        args = ["pdftotext", "-f",str(page),"-l",str(page),"-layout"]
        args += ["-x","305" if right else "0","-y","0","-W","310","-H","792"]
        raw = subprocess.check_output(args+[str(pdf),"-"],text=True)
        return " ".join(re.sub(r"-\n\s*", "", raw.replace("\u00ad", "")).split())
    left, right = extract(76), extract(76,True)
    clauses_left = ["Warlock Subclass: Fiend Patron", "Level 3: Dark One’s Blessing",
        "When you reduce an enemy to 0 Hit Points", "Temporary Hit Points equal to your Charisma modifier plus your Warlock level",
        "minimum of 1 Temporary Hit Point", "someone else reduces an enemy within 10 feet of you to 0 Hit Points",
        "Level 3: Fiend Spells", "Warlock level specified in the Fiend Spells table", "thereafter always have the listed spells prepared"]
    clauses_right = ["Fiend Spells Warlock Level Spells", "3 Burning Hands, Command, Scorching Ray, Suggestion",
        "5 Fireball, Stinking Cloud", "7 Fire Shield, Wall of Fire", "9 Geas, Insect Plague",
        "Level 6: Dark One’s Own Luck", "When you make an ability check or a saving throw",
        "add 1d10 to your roll", "after seeing the roll but before any of the roll’s effects occur",
        "number of times equal to your Charisma modifier (minimum of once)", "no more than once per roll",
        "regain all expended uses when you finish a Long Rest", "Level 10: Fiendish Resilience",
        "Choose one damage type, other than Force", "whenever you finish a Short or Long Rest",
        "Resistance to that damage type until you choose a different one with this feature",
        "Level 14: Hurl Through Hell", "Once per turn when you hit a creature with an attack roll",
        "target must succeed on a Charisma saving throw against your spell save DC",
        "or the target disappears", "target takes 8d10 Psychic damage if it isn’t a Fiend",
        "Incapacitated condition until the end of your next turn", "returns to the space it previously occupied or the nearest unoccupied space",
        "Once you use this feature", "until you finish a Long Rest unless you expend a Pact Magic spell slot (no action required) to restore your use"]
    for text, clauses in [(left,clauses_left),(right,clauses_right)]:
        assert all(x in text for x in clauses), [x for x in clauses if x not in text]
    actual = json.loads((ROOT/"internal/rules/srd521/testdata/character-fiend.json").read_text())
    assert actual["schema_version"] == 1 and actual["source"]["sha256"] == PIN
    assert actual["source"]["license"] == "CC-BY-4.0" and actual["source"]["changes"] and actual["source"]["attribution"]
    compare(actual["records"])
    inv = json.loads((ROOT/"docs/investigations/issue-74-character-inventory.json").read_text())
    identities = [r for r in actual["records"] if r["kind"] in ("subclass", "feature")]
    expected = {r["id"]:r["source_page"] for r in inv["records"] if r["batch_key"] == "subclass-warlock"}
    assert len(identities) == len(expected) == 6
    assert {s["inventory_id"]:s["page"] for r in identities for s in r["sources"]} == expected
    from verify_srd_warlock import verify as verify_warlock
    verify_warlock(pdf)
    print(f"p76: six identities and {len(actual['records'])} complete records match independent source facts")


if __name__ == "__main__":
    if len(sys.argv) == 3 and sys.argv[1] == "--controls":
        # Positive baseline prevents an unrelated comparator from rejecting everything.
        compare(json.loads((ROOT/"internal/rules/srd521/testdata/character-fiend.json").read_text())["records"])
        controls = json.loads(Path(sys.argv[2]).read_text())
        assert controls, "no contrary graphs supplied"
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
        raise SystemExit("usage: verify_srd_fiend.py PDF | --controls JSON")
