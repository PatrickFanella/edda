#!/usr/bin/env python3
"""#303: independent pp73–74 PDF-to-metadata proof. No Go or network.

Every semantic group is bound to an extracted source clause before constructing
its expected graph. Without a PDF, the same facts test contrary graph controls.
SRD 5.2.1 names and facts adapted under CC-BY-4.0; graph/key design is Edda work.
"""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / "internal/rules/srd521/testdata/character-invocations-02.json"
PIN = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
C = "srd-5.2.1:character/"
M = "srd-5.2.1:mechanic/"
K = M + "warlock/invocations/"


def source_records(pdf=None):
    flat = None
    if pdf:
        assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
        columns = []
        for page in (73, 74):
            for x in (0, 310):
                raw = subprocess.check_output(["pdftotext", "-f", str(page), "-l", str(page),
                    "-layout", "-x", str(x), "-y", "0", "-W", "310", "-H", "792", str(pdf), "-"], text=True)
                columns.append(re.sub(r"-\n\s*", "", raw.replace("\u00ad", "")))
        flat = " ".join(" ".join(columns).split())
    def clause(text):
        if flat is not None:
            assert text in flat, "source clause: " + text
    out = []
    def add(kind, slug, name, parent, source):
        r = dict(id=C+kind+"/"+slug, kind=kind, name=name, sources=[source.copy()],
                 parent_id=parent, level=0, choice_refs=[], resource_refs=[], prerequisite_refs=[],
                 required_mechanics=[], select=0, option_keys=[], capacity=0)
        out.append(r)
        return r
    def rule(r, text, keys, nums=None):
        clause(text)
        r["required_mechanics"] += [K+k for k in keys.split()]
        if nums:
            r.setdefault("parameters", {}).update({K+k:v for k,v in nums.items()})
            r["required_mechanics"] += [K+k for k in nums]
    def pred(r, suffix, key, refs=(), nums=None):
        p = add("prerequisite", r["id"][len(C+"feature/warlock/invocations/"):]+"/"+suffix,
                r["name"]+" "+suffix, r["id"], r["sources"][0])
        p["required_mechanics"] = [K+key]
        p["requirement"] = dict(key=K+key, record_refs=[C+x for x in refs], parameters=nums or {})
        r["prerequisite_refs"].append(p["id"])
        return p
    def choice(r, suffix, count, options, key):
        q = add("choice", r["id"][len(C+"feature/warlock/invocations/"):]+"/"+suffix,
                r["name"]+" "+suffix, r["id"], r["sources"][0])
        q.update(select=count, option_keys=[M+x for x in options.split()], required_mechanics=[K+key])
        r["choice_refs"].append(q["id"])
        return q
    def inv(slug, name, page, level=0, pact=None, repeat=False):
        clause(name)
        r = add("feature", "warlock/invocations/"+slug, name, C+"choice/warlock/invocations",
                dict(inventory_id="srd-5.2.1:planning/invocations/"+slug, page=page))
        r["required_mechanics"] = [K+slug, K+("repeatable/different-eligible-selection" if repeat else "not-repeatable")]
        pred(r, "canonical-eligibility", "canonical-eligibility-and-replacement", ["feature/warlock/eldritch-invocations", "choice/warlock/invocations"])
        if level:
            clause("Level "+str(level)+"+ Warlock")
            pred(r, "level", "minimum-warlock-level", ["class/warlock"], {"minimum-level":level})
        if pact:
            clause("Pact of the "+pact.title()+" Invocation")
            pred(r, "pact", "requires-pact-invocation", ["feature/warlock/invocations/pact-of-the-"+pact])
        return r

    r = inv("investment-of-the-chain-master", "Investment of the Chain Master", 73, 5, "chain")
    rule(r, "When you cast Find Familiar", "investment/when-cast-find-familiar/infuse-summoned-familiar")
    r["required_mechanics"].append(M+"spells/find-familiar")
    rule(r, "either a Fly Speed or a Swim Speed (your choice) of 40 feet", "investment/grant-chosen-speed", {"investment/speed-feet":40})
    choice(r,"movement",1,"movement/fly movement/swim","investment/speed-choice")
    rule(r, "As a Bonus Action, you can command the familiar to take the Attack action", "investment/bonus-action/command-familiar-attack-action")
    rule(r, "Whenever the familiar deals Bludgeoning, Piercing, or Slashing damage", "investment/optional-replace-only-bludgeoning-piercing-slashing")
    clause("make it deal Necrotic or Radiant damage instead")
    choice(r,"damage",1,"damage/necrotic damage/radiant","investment/replacement-damage")
    rule(r, "If the familiar forces a creature to make a saving throw, it uses your spell save DC", "investment/if-familiar-forces-save/use-your-spell-save-dc")
    pred(r,"save-dc","investment/canonical-spell-save-dc",["feature/warlock/pact-magic"])
    rule(r, "When the familiar takes damage, you can take a Reaction to grant it Resistance against that damage", "investment/when-familiar-takes-damage/reaction/resistance-against-that-damage")

    r = inv("lessons-of-the-first-ones", "Lessons of the First Ones", 73, 2, repeat=True)
    rule(r, "gain one Origin feat of your choice", "lessons/gain-origin-feat")
    clause("Each time you do so, choose a different Origin feat")
    q = choice(r,"origin-feat",1,"feats/origin","lessons/different-origin-feat-each-time")
    pred(r,"feat-owners","lessons/available-origin-feat-owners",["feat/alert","feat/magic-initiate","feat/savage-attacker","feat/skilled"])

    r = inv("lifedrinker", "Lifedrinker", 73, 9, "blade")
    rule(r, "Once per turn when you hit a creature with your pact weapon", "lifedrinker/optional-on-hit-creature-with-pact-weapon", {"lifedrinker/maximum-per-turn":1})
    rule(r, "extra 1d6 Necrotic, Psychic, or Radiant damage", "lifedrinker/extra-damage-to-hit-creature", {"lifedrinker/dice":1,"lifedrinker/die-sides":6})
    choice(r,"damage",1,"damage/necrotic damage/psychic damage/radiant","lifedrinker/damage-choice")
    rule(r, "expend one of your Hit Point Dice to roll it", "lifedrinker/optional-expend-own-hit-point-die-and-roll-it", {"lifedrinker/hit-point-dice-cost":1})
    rule(r, "equal to the roll plus your Constitution modifier (minimum of 1 Hit Point)", "lifedrinker/self-healing/expended-die-roll-plus-constitution", {"lifedrinker/minimum-healing":1})
    pred(r,"weapon","lifedrinker/uses-bonded-pact-weapon",["feature/warlock/invocations/pact-of-the-blade"])

    for slug, name, page, level, spell, text in [
        ("mask-of-many-faces","Mask of Many Faces",73,2,"disguise-self","Disguise Self"),
        ("master-of-myriad-forms","Master of Myriad Forms",73,5,"alter-self","Alter Self"),
        ("misty-visions","Misty Visions",74,2,"silent-image","Silent Image"),
        ("one-with-shadows","One with Shadows",74,5,"invisibility","Invisibility"),
        ("otherworldly-leap","Otherworldly Leap",74,2,"jump","Jump")]:
        r = inv(slug,name,page,level)
        clause("You can cast "+text if slug != "one-with-shadows" else "you can cast Invisibility")
        r["required_mechanics"].append(M+"spells/"+spell)
        rule(r,"without expending a spell slot","casting/without-spell-slot")
        if slug in ("one-with-shadows","otherworldly-leap"):
            rule(r,"on yourself without expending","casting/self-only")
        if slug == "one-with-shadows":
            rule(r,"While you’re in an area of Dim Light or Darkness","shadows/only-while-in-dim-light-or-darkness")

    r = inv("pact-of-the-blade","Pact of the Blade",74)
    rule(r,"As a Bonus Action, you can conjure a pact weapon in your hand", "blade/bonus-action/conjure-in-hand-or-bond-touched-magic-weapon")
    clause("a Simple or Martial Melee weapon of your choice")
    choice(r,"bond-method",1,"warlock/pact-weapon/conjured-simple-or-martial-melee warlock/pact-weapon/touched-magic-weapon","blade/bond-alternatives")
    rule(r,"you can’t bond with a magic weapon if someone else is attuned to it or another Warlock is bonded with it", "blade/magic-weapon/no-other-attuned-creature-or-warlock-bond")
    rule(r,"Until the bond ends, you have proficiency with the weapon", "blade/while-bonded/proficient-with-weapon")
    rule(r,"you can use it as a Spellcasting Focus", "blade/while-bonded/spellcasting-focus")
    rule(r,"use your Charisma modifier for the attack and damage rolls instead of using Strength or Dexterity", "blade/attacks/optional-charisma-replaces-strength-or-dexterity-for-attack-and-damage")
    clause("Necrotic, Psychic, or Radiant damage or its normal damage type")
    choice(r,"damage",1,"damage/necrotic damage/psychic damage/radiant warlock/pact-weapon/normal-damage","blade/each-attack/damage-choice")
    rule(r,"Your bond with the weapon ends if you use this feature’s Bonus Action again", "blade/terminate/on-reuse-bonus-action")
    rule(r,"more than 5 feet away from you for 1 minute or more", "blade/terminate/continuous-distance-exceeds-threshold-for-duration", {"blade/distance-feet":5,"blade/duration-minutes":1})
    rule(r,"or if you die. A conjured weapon disappears when the bond ends", "blade/terminate/on-your-death blade/conjured-weapon-disappears-on-bond-end")

    r = inv("pact-of-the-chain","Pact of the Chain",74)
    rule(r,"You learn the Find Familiar spell and can cast it as a Magic action without expending a spell slot", "chain/learn-find-familiar chain/cast-magic-action-without-spell-slot")
    r["required_mechanics"].append(M+"spells/find-familiar")
    clause("normal forms for your familiar or one of the following special forms")
    for text in ("Imp, Pseudodragon, Quasit", "Skeleton, Sphinx of Wonder, Sprite, or Venomous Snake"):
        clause(text)
    choice(r,"familiar",1,"spells/find-familiar/normal-forms creatures/imp creatures/pseudodragon creatures/quasit creatures/skeleton creatures/sphinx-of-wonder creatures/sprite creatures/venomous-snake","chain/choose-on-cast/normal-or-special-form")
    rule(r,"when you take the Attack action", "chain/when-you-take-attack-action")
    rule(r,"forgo one of your own attacks to allow your familiar to make one attack of its own with its Reaction", "chain/optional-forgo-own-attack/familiar-attack-with-its-reaction", {"chain/own-attacks-cost":1,"chain/familiar-attacks":1})

    r = inv("pact-of-the-tome","Pact of the Tome",74)
    rule(r,"conjure forth a book in your hand at the end of a Short or Long Rest", "tome/conjure-book-in-hand/end-short-or-long-rest tome/appearance-your-choice")
    rule(r,"eldritch magic that only you can access", "tome/magic-access-only-you")
    rule(r,"book disappears if you conjure another book with this feature or if you die", "tome/disappear-on-replacement-or-your-death")
    rule(r,"When the book appears, choose three cantrips", "tome/choose-when-book-appears")
    choice(r,"cantrips",3,"spells/any-class/cantrips","tome/cantrip-choice")
    clause("choose two level 1 spells that have the Ritual tag")
    choice(r,"rituals",2,"spells/any-class/level-1-rituals","tome/ritual-choice")
    rule(r,"spells can be from any class’s spell list", "tome/any-class-spell-list")
    rule(r,"they must be spells you don’t already have prepared", "tome/selected-spells-not-already-prepared")
    rule(r,"While the book is on your person, you have the chosen spells prepared", "tome/prepared-only-while-book-on-person")
    rule(r,"they function as Warlock spells for you", "tome/chosen-spells-function-as-warlock-spells")
    pred(r,"casting","tome/canonical-warlock-casting",["feature/warlock/pact-magic"])
    rule(r,"You can use the book as a Spellcasting Focus", "tome/book-spellcasting-focus")

    r = inv("repelling-blast","Repelling Blast",74,2,repeat=True)
    rule(r,"a Warlock Cantrip That Deals Damage via an Attack Roll", "repelling/require-known-warlock-damaging-cantrip-via-attack-roll")
    clause("Choose one of your known Warlock cantrips that requires an attack roll")
    choice(r,"cantrip",1,"spells/known-warlock-cantrips/damage-via-attack-roll","repelling/different-eligible-cantrip-each-time")
    rule(r,"When you hit a Large or smaller creature with that cantrip", "repelling/when-chosen-cantrip-hits-large-or-smaller-creature")
    rule(r,"push the creature up to 10 feet straight away from you", "repelling/optional-push-hit-creature/straight-away", {"repelling/maximum-push-feet":10})
    clause("Each time you do so, choose a different eligible cantrip")
    return out


def compare(records, pdf=None):
    expected = source_records(pdf)
    assert len(records) == len(expected), "record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"semantic mismatch {want['id']}: {actual} != {want}"


def verify(pdf):
    actual = json.loads(DATA.read_text())
    assert actual["schema_version"] == 1 and actual["source"]["sha256"] == PIN
    assert actual["source"]["license"] == "CC-BY-4.0" and actual["source"]["changes"] and actual["source"]["attribution"]
    compare(actual["records"], pdf)
    inv = json.loads((ROOT/"docs/investigations/issue-74-character-inventory.json").read_text())
    expected = {r["id"]:r["source_page"] for r in inv["records"] if r["batch_key"] == "options-invocations-02"}
    identities = [r for r in actual["records"] if r["kind"] == "feature"]
    assert len(identities) == len(expected) == 12
    assert {s["inventory_id"]:s["page"] for r in identities for s in r["sources"]} == expected
    from verify_srd_warlock import verify as verify_warlock
    verify_warlock(pdf)
    print(f"pp73–74: 12 identities and {len(actual['records'])} complete semantic records match PDF facts")


if __name__ == "__main__":
    if len(sys.argv) == 3 and sys.argv[1] == "--controls":
        compare(json.loads(DATA.read_text())["records"])
        controls = json.loads(Path(sys.argv[2]).read_text())
        assert controls
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
        raise SystemExit("usage: verify_srd_invocations_02.py PDF | --controls JSON")
