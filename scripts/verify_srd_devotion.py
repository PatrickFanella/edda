#!/usr/bin/env python3
"""Independent offline pp56–57 source-fact comparison for #314; no Go or network."""
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
BASE = "paladin/oath-of-devotion/"


def source_records():
    out = []
    def record(kind, slug, name, source, parent, keys=(), nums=None):
        r = dict(id=C+kind+"/"+BASE+slug, kind=kind, name=name,
                 sources=[dict(inventory_id=P+source, page=56)], parent_id=parent,
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
    subclass = record("subclass", "identity", "Oath of Devotion",
                      "subclasses/paladin-oath-of-devotion", C+"class/paladin", ["subclass"])
    pred(subclass, "level", "class-level", ["class/paladin"], {"minimum-level":3})
    def feature(slug, name, level, keys, nums=None):
        r = record("feature", slug, name, f"subclass-features/paladin-level-{level}-{slug}", subclass["id"], keys, nums)
        if level >= 15:
            r["sources"][0]["page"] = 57
        pred(r, "level", "class-level", ["class/paladin"], {"minimum-level":level})
        return r
    spells = feature("oath-of-devotion-spells", "Oath of Devotion Spells", 3,
                     ["oath-spells/always-prepared-from-specified-paladin-level"])
    spells["always_prepared"] = [dict(level=level, spell_keys=[M+"spells/"+s for s in names]) for level,names in [
        (3,["protection-from-evil-and-good","shield-of-faith"]), (5,["aid","zone-of-truth"]),
        (9,["beacon-of-hope","dispel-magic"]), (13,["freedom-of-movement","guardian-of-faith"]),
        (17,["commune","flame-strike"])]]
    for row in spells["always_prepared"]:
        spells["required_mechanics"] += row["spell_keys"]
    pred(spells,"preparation","oath-spells/canonical-paladin-preparation-rules",["feature/paladin/spellcasting"])
    weapon = feature("sacred-weapon","Sacred Weapon",3,[
        "sacred-weapon/optional-when-taking-attack-action", "sacred-weapon/target/one-held-melee-weapon",
        "sacred-weapon/attack-rolls/add-charisma-modifier-minimum-one",
        "sacred-weapon/each-hit/choose-normal-damage-type-or-radiant",
        "sacred-weapon/end/on-reuse-or-no-longer-carrying-weapon-or-voluntarily-no-action"],
        {"sacred-weapon/channel-divinity-cost":1,"sacred-weapon/duration-minutes":10,
         "sacred-weapon/attack-bonus-minimum":1,"sacred-weapon/bright-light-radius-feet":20,
         "sacred-weapon/dim-light-additional-feet":20})
    pred(weapon,"cost","sacred-weapon/expend-canonical-channel-divinity-with-its-capacity-and-recovery",
         ["resource/paladin/channel-divinity"],{"uses-required":1})
    aura = feature("aura-of-devotion","Aura of Devotion",7,[
        "aura-of-devotion/self-and-allies-in-aura/immune-to-charmed",
        "aura-of-devotion/charmed-ally-entering-aura/condition-has-no-effect-while-there"])
    aura_refs = ["feature/paladin/aura-of-protection","feature/paladin/aura-expansion"]
    pred(aura,"scope","aura-of-devotion/within-canonical-aura-of-protection-including-expansion-and-inactive-condition",aura_refs)
    smite = feature("smite-of-protection","Smite of Protection",15,[
        "smite-of-protection/whenever-cast-divine-smite", "smite-of-protection/self-and-allies-in-aura/half-cover",
        "smite-of-protection/aura-benefit-until-start-of-your-next-turn"])
    pred(smite,"trigger","smite-of-protection/cast-divine-smite-with-canonical-spell-rules",["feature/paladin/paladins-smite"])
    pred(smite,"scope","smite-of-protection/within-canonical-aura-of-protection-including-expansion-and-inactive-condition",aura_refs)
    nimbus = feature("holy-nimbus","Holy Nimbus",20,[
        "holy-nimbus/bonus-action", "holy-nimbus/end/voluntarily-no-action", "holy-nimbus/recover-use/finish-long-rest",
        "holy-nimbus/restore-use/expend-level-five-slot-no-action",
        "holy-nimbus/holy-ward/self-advantage-on-saves-forced-by-fiend-or-undead",
        "holy-nimbus/radiant-damage/enemy-starts-turn-in-aura/charisma-modifier-plus-proficiency-bonus",
        "holy-nimbus/sunlight/aura-filled-with-bright-light-that-is-sunlight"],
        {"holy-nimbus/duration-minutes":10,"holy-nimbus/uses":1,"holy-nimbus/restore-slot-level":5,"holy-nimbus/restore-slot-cost":1})
    pred(nimbus,"scope","holy-nimbus/imbue-canonical-aura-of-protection-including-expansion-and-inactive-condition",aura_refs)
    pred(nimbus,"restore","holy-nimbus/restore-expended-use-with-canonical-slot",["resource/paladin/spell-slot-5"],{"slots-required":1,"slot-level":5})
    return out


def compare(records):
    expected = source_records()
    assert len(records) == len(expected), "pp56–57 record count"
    for actual, want in zip(records, expected):
        assert actual == want, f"pp56–57 semantic mismatch {want['id']}: {actual} != {want}"


def verify(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    def extract(page, right=False):
        args = ["pdftotext", "-f",str(page),"-l",str(page),"-layout"]
        args += ["-x","305" if right else "0","-y","0","-W","310","-H","792"]
        raw = subprocess.check_output(args+[str(pdf),"-"],text=True)
        return " ".join(re.sub(r"-\n\s*", "", raw.replace("\u00ad", "")).split())
    p56, p57 = extract(56,True), extract(57)
    clauses56 = ["Oath of Devotion", "always have the listed spells prepared", "Paladin Level Spells",
        "3 Protection from Evil and Good, Shield of Faith", "5 Aid, Zone of Truth",
        "9 Beacon of Hope, Dispel Magic", "13 Freedom of Movement, Guardian of Faith", "17 Commune, Flame Strike",
        "When you take the Attack action", "expend one use of your Channel Divinity", "one Melee weapon that you are holding",
        "For 10 minutes or until you use this feature again", "add your Charisma modifier to attack rolls",
        "minimum bonus of +1", "each time you hit", "normal damage type or Radiant damage",
        "Bright Light in a 20-foot radius and Dim Light 20 feet beyond that", "end this effect early (no action required)",
        "ends if you aren’t carrying the weapon", "You and your allies have Immunity to the Charmed condition",
        "while in your Aura of Protection", "Charmed ally enters the aura", "no effect on that ally while there"]
    clauses57 = ["Whenever you cast Divine Smite", "you and your allies have Half Cover while in your Aura of Protection",
        "until the start of your next turn", "As a Bonus Action", "imbue your Aura of Protection",
        "10 minutes or until you end them (no action required)", "Once you use this feature",
        "until you finish a Long Rest", "restore your use", "expending a level 5 spell slot (no action required)",
        "Advantage on any saving throw", "forced to make by a Fiend or an Undead",
        "Whenever an enemy starts its turn in the aura", "Radiant damage equal to your Charisma modifier plus your Proficiency Bonus",
        "aura is filled with Bright Light that is sunlight"]
    for text, clauses in [(p56,clauses56),(p57,clauses57)]:
        assert all(x in text for x in clauses), [x for x in clauses if x not in text]
    actual = json.loads((ROOT/"internal/rules/srd521/testdata/character-devotion.json").read_text())
    assert actual["schema_version"] == 1 and actual["source"]["sha256"] == PIN
    assert actual["source"]["license"] == "CC-BY-4.0" and actual["source"]["changes"] and actual["source"]["attribution"]
    compare(actual["records"])
    inv = json.loads((ROOT/"docs/investigations/issue-74-character-inventory.json").read_text())
    identities = [r for r in actual["records"] if r["kind"] in ("subclass", "feature")]
    expected = {r["id"]:r["source_page"] for r in inv["records"] if r["batch_key"] == "subclass-paladin"}
    assert len(identities) == len(expected) == 6
    assert {s["inventory_id"]:s["page"] for r in identities for s in r["sources"]} == expected
    from verify_srd_paladin import verify as verify_paladin
    verify_paladin(pdf)
    print(f"pp56–57: six identities and {len(actual['records'])} complete records match independent source facts")


if __name__ == "__main__":
    if len(sys.argv) == 3 and sys.argv[1] == "--controls":
        # Positive baseline prevents an unrelated comparator from rejecting everything.
        compare(json.loads((ROOT/"internal/rules/srd521/testdata/character-devotion.json").read_text())["records"])
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
        raise SystemExit("usage: verify_srd_devotion.py PDF | --controls JSON")
