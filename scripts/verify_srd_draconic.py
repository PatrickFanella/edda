#!/usr/bin/env python3
"""Independent offline pp69–70 source-fact comparison for #317; no Go or network."""
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
BASE = "sorcerer/draconic-sorcery/"


def source_records():
    out = []
    def record(kind, slug, name, source, parent, keys=(), nums=None):
        r = dict(id=C+kind+"/"+BASE+slug, kind=kind, name=name,
                 sources=[dict(inventory_id=P+source, page=70)], parent_id=parent,
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
    subclass = record("subclass", "identity", "Draconic Sorcery",
                      "subclasses/sorcerer-draconic-sorcery", C+"class/sorcerer", ["subclass"])
    subclass["sources"][0]["page"] = 69
    pred(subclass, "level", "class-level", ["class/sorcerer"], {"minimum-level":3})
    def feature(slug, name, level, page, keys, nums=None):
        r = record("feature", slug, name, f"subclass-features/sorcerer-level-{level}-{slug}", subclass["id"], keys, nums)
        r["sources"][0]["page"] = page
        pred(r, "level", "class-level", ["class/sorcerer"], {"minimum-level":level})
        return r
    resilience = feature("draconic-resilience", "Draconic Resilience", 3, 69, [
        "resilience/self/hit-point-maximum-increases-at-level-three",
        "resilience/self/each-additional-sorcerer-level-increases-hit-point-maximum",
        "resilience/self/while-not-wearing-armor/base-ac-ten-plus-dexterity-plus-charisma-modifiers"],
        {"resilience/initial-hp-increase":3,"resilience/hp-per-additional-sorcerer-level":1,"resilience/base-ac-constant":10})
    pred(resilience,"scaling","resilience/count-only-canonical-sorcerer-levels",["class/sorcerer"])
    spells = feature("draconic-spells", "Draconic Spells", 3, 70,
                     ["spells/thereafter-always-prepared-from-specified-sorcerer-level"])
    spells["always_prepared"] = [dict(level=level, spell_keys=[M+"spells/"+s for s in names]) for level,names in [
        (3,["alter-self","chromatic-orb","command","dragons-breath"]), (5,["fear","fly"]),
        (7,["arcane-eye","charm-monster"]), (9,["legend-lore","summon-dragon"])]]
    for row in spells["always_prepared"]:
        spells["required_mechanics"] += row["spell_keys"]
    pred(spells,"preparation","spells/canonical-sorcerer-preparation-slot-and-innate-sorcery-rules",
         ["feature/sorcerer/spellcasting","choice/sorcerer/prepared-spells","feature/sorcerer/innate-sorcery","resource/sorcerer/innate-sorcery"])
    affinity = feature("elemental-affinity","Elemental Affinity",6,70,[
        "affinity/self/resistance-to-chosen-damage-type",
        "affinity/when-you-cast-spell-dealing-chosen-type/optional-add-charisma-modifier-to-one-damage-roll"],
        {"affinity/damage-rolls-maximum":1})
    damage = record("choice","elemental-affinity-damage-type","Elemental Affinity damage type",
                    "subclass-features/sorcerer-level-6-elemental-affinity",affinity["id"],
                    ["affinity/choose-one-draconic-damage-type"])
    damage.update(select=1,option_keys=[M+BASE+"affinity/damage-type/"+x for x in ["acid","cold","fire","lightning","poison"]])
    affinity["choice_refs"].append(damage["id"])
    pred(affinity,"type","affinity/use-selected-type-for-both-benefits",
         ["choice/sorcerer/draconic-sorcery/elemental-affinity-damage-type","feature/sorcerer/spellcasting"])
    wings = feature("dragon-wings","Dragon Wings",14,70,[
        "wings/self/bonus-action-manifest-draconic-wings-on-back","wings/end/one-hour-or-dismiss-no-action",
        "wings/self/fly-speed-for-duration","wings/reuse/finish-long-rest-or-restore-use",
        "wings/restore/expend-sorcery-points-no-action"],
        {"wings/duration-hours":1,"wings/fly-speed-feet":60,"wings/uses":1,"wings/restore-sorcery-points":3})
    pred(wings,"restore","wings/spend-canonical-sorcery-points-with-capacity-and-recovery",
         ["resource/sorcerer/sorcery-points"],{"points-required":3})
    companion = feature("dragon-companion","Dragon Companion",18,70,[
        "companion/cast-summon-dragon/without-material-component",
        "companion/cast-summon-dragon/once-without-spell-slot","companion/free-cast/recover-on-long-rest",
        "companion/whenever-start-casting/optional-no-concentration",
        "companion/if-no-concentration/duration-one-minute-for-that-casting-only"],
        {"companion/free-casts":1,"companion/modified-duration-minutes":1})
    companion["required_mechanics"].append(M+"spells/summon-dragon")
    pred(companion,"casting","companion/use-canonical-sorcerer-spell-and-slot-rules-except-explicit-exemptions",
         ["feature/sorcerer/spellcasting","choice/sorcerer/prepared-spells","resource/sorcerer/spell-slot-5","feature/sorcerer/innate-sorcery"])
    return out


def compare(records):
    expected = source_records()
    assert len(records) == len(expected), "pp69–70 record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"pp69–70 semantic mismatch {want['id']}: {actual} != {want}"


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    def extract(page, right=False):
        args = ["pdftotext", "-f",str(page),"-l",str(page),"-layout"]
        args += ["-x","305" if right else "0","-y","0","-W","310","-H","792"]
        raw = subprocess.check_output(args+[str(pdf),"-"],text=True)
        return " ".join(re.sub(r"-\n\s*", "", raw.replace("\u00ad", "")).split())
    p69, p70 = extract(69,True), extract(70)
    clauses69 = ["Sorcerer Subclass: Draconic Sorcery", "Level 3: Draconic Resilience",
        "Your Hit Point maximum increases by 3", "increases by 1 whenever you gain another Sorcerer level",
        "While you aren’t wearing armor", "base Armor Class equals 10 plus your Dexterity and Charisma modifiers"]
    clauses70 = ["Level 3: Draconic Spells", "When you reach a Sorcerer level specified",
        "thereafter always have the listed spells prepared", "Sorcerer Level Spells",
        "3 Alter Self, Chromatic Orb, Command, Dragon’s Breath", "5 Fear, Fly",
        "7 Arcane Eye, Charm Monster", "9 Legend Lore, Summon Dragon", "Level 6: Elemental Affinity",
        "Choose one of those types: Acid, Cold, Fire, Lightning, or Poison", "Resistance to that damage type",
        "when you cast a spell that deals damage of that type", "can add your Charisma modifier to one damage roll of that spell",
        "Level 14: Dragon Wings", "As a Bonus Action", "wings to appear on your back",
        "wings last for 1 hour or until you dismiss them (no action required)", "For the duration",
        "Fly Speed of 60 feet", "Once you use this feature", "until you finish a Long Rest unless you spend 3 Sorcery Points (no action required) to restore your use",
        "Level 18: Dragon Companion", "cast Summon Dragon without a Material component",
        "cast it once without a spell slot", "regain the ability to cast it in this way when you finish a Long Rest",
        "Whenever you start casting the spell", "can modify it so that it doesn’t require Concentration",
        "If you do so, the spell’s duration becomes 1 minute for that casting"]
    for text, clauses in [(p69,clauses69),(p70,clauses70)]:
        assert all(x in text for x in clauses), [x for x in clauses if x not in text]
    actual = json.loads((ROOT/"internal/rules/srd521/testdata/character-draconic.json").read_text())
    assert actual["schema_version"] == 1 and actual["source"]["sha256"] == PIN
    assert actual["source"]["license"] == "CC-BY-4.0" and actual["source"]["changes"] and actual["source"]["attribution"]
    compare(actual["records"])
    inv = json.loads((ROOT/"docs/investigations/issue-74-character-inventory.json").read_text())
    identities = [r for r in actual["records"] if r["kind"] in ("subclass", "feature")]
    expected = {r["id"]:r["source_page"] for r in inv["records"] if r["batch_key"] == "subclass-sorcerer"}
    assert len(identities) == len(expected) == 6
    assert {s["inventory_id"]:s["page"] for r in identities for s in r["sources"]} == expected
    from verify_srd_sorcerer import verify as verify_sorcerer
    verify_sorcerer(pdf)
    manifest = json.loads((ROOT/"internal/rules/srd521/testdata/character-draconic-composition.json").read_text())
    assert manifest == {"catalogues":["character-sorcerer.json","character-draconic.json"]}
    print(f"pp69–70: six identities and {len(actual['records'])} complete records match independent source facts")


if __name__ == "__main__":
    if len(sys.argv) == 3 and sys.argv[1] == "--controls":
        # Positive baseline prevents an unrelated comparator from rejecting everything.
        compare(json.loads((ROOT/"internal/rules/srd521/testdata/character-draconic.json").read_text())["records"])
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
        raise SystemExit("usage: verify_srd_draconic.py PDF | --controls JSON")
