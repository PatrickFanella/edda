#!/usr/bin/env python3
"""Standalone pinned-PDF Warlock semantic comparator. CC-BY-4.0 adaptation.

No Go, fixture-derived expectations, runtime evaluation or network access.
Opaque Invocation, Fiend, spell and feat definitions stay outside this batch.
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
K = M + "warlock/"
DATA = ROOT / "internal/rules/srd521/testdata/character-warlock.json"
CHANGES = "Warlock source identities, all 20 progression rows, nested choices, prerequisites, Pact Magic and separate arcanum resources on PDF pp70-72 adapted into unavailable offline metadata. Mechanic keys and graph are Edda additions. Invocation, Fiend, spell and feat definitions remain opaque. No executable support imported."


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    def extract(page, x=None):
        args = ["pdftotext", "-f", str(page), "-l", str(page), "-layout"]
        if x is not None:
            args += ["-x",str(x),"-y","0","-W","310","-H","792"]
        return subprocess.check_output(args + [str(pdf), "-"], text=True)
    layout = extract(71)
    columns = [extract(p,x) for p in (70,71,72) for x in (0,310)]
    flat = " ".join(re.sub(r"-\n\s*", "", " ".join(columns)).replace("\u00ad","").split())
    def clause(text):
        assert text in flat, "source clause: " + text
    records, by = [], {}
    def add(kind, slug, parent, name=None, level=0, page=0):
        ident = C + kind + "/warlock" + ("/" + slug if slug else "")
        if kind == "class":
            source = dict(inventory_id="srd-5.2.1:planning/classes/warlock", page=70)
        elif kind == "feature":
            source = dict(inventory_id=f"srd-5.2.1:planning/class-features/warlock-level-{by[parent]['level']}-{slug}",page=page)
        else:
            source = copy.deepcopy(by[parent]["sources"][0])
        r = dict(id=ident,kind=kind,name=name or slug.replace("-"," ").title(),sources=[source],
                 parent_id=parent,level=level,choice_refs=[],resource_refs=[],prerequisite_refs=[],
                 required_mechanics=[],select=0,option_keys=[],capacity=0)
        records.append(r); by[ident]=r
        if kind in ("choice","resource","prerequisite"):
            by[parent][kind+"_refs"].append(ident)
        return ident
    def rule(owner, text, words, nums=None):
        clause(text)
        r=by[owner]
        r["required_mechanics"] += [K+w for w in words.split()]
        if nums:
            r.setdefault("parameters",{}).update({K+k:v for k,v in nums.items()})
            r["required_mechanics"] += [K+k for k in nums]
    def choice(slug,owner,count,options):
        q=add("choice",slug,owner)
        by[q].update(select=count,option_keys=[M+w for w in options.split()],required_mechanics=[K+slug])
        return q
    def pred(slug,owner,key,refs,nums=None,option=None):
        p=add("prerequisite",slug,owner)
        by[p]["required_mechanics"]=[K+key]
        by[p]["requirement"]=dict(key=K+key,record_refs=refs,parameters=nums or {})
        if option:
            by[p]["required_mechanics"].append(K+"predicate/conditional-on-option/"+option)
        return p
    def resource(slug,owner,name=None,key=None):
        r=add("resource",slug,owner,name)
        by[r].update(capacity=1,required_mechanics=[K+key] if key else [])
        return r
    cls=add("class","","","Warlock")
    rule(cls,"Wisdom and Charisma","primary-ability/charisma saves/wisdom saves/charisma weapons/simple armor/light level-1/all-core-traits-and-level-1-features multiclass/hit-die-light-armor-level-1-features multiclass/consult-multiclass-spell-slot-rules",{"hit-die-sides":8})
    for text in ("D8 per Warlock level","Simple weapons","Light armor","Hit Point Die and training with Light","determine your available spell slots"):
        clause(text)
    # Exact skill and item domains transcribed from the core table, not adjacent definitions.
    choice("skills",cls,2,"skills/arcana skills/deception skills/history skills/intimidation skills/investigation skills/nature skills/religion")
    clause("Choose 2: Arcana, Deception, History, Intimidation, Investigation, Nature, or Religion")
    eq=choice("starting-equipment",cls,1,"warlock/equipment/a warlock/equipment/b")
    clause("Armor, Sickle, 2 Daggers")
    clause("cane Focus (orb), Book (occult lore), Scholar’s Pack, and 15 GP; or (B) 100 GP")
    bundles=[{"equipment/leather-armor":1,"equipment/sickle":1,"equipment/dagger":2,"equipment/arcane-focus-orb":1,"equipment/book-occult-lore":1,"equipment/scholars-pack":1,"currency/gp":15},{"currency/gp":100}]
    by[eq]["options"]=[dict(key=key,quantities={M+k:v for k,v in b.items()}) for key,b in zip(by[eq]["option_keys"],bundles)]
    by[eq]["required_mechanics"] += [M+k for k in bundles[0]]
    # Parse every actual numeric cell plus the feature text from page 71.
    rows=re.findall(r"^\s*(\d+)\s+\+(\d)\s+(.+?)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s*$",layout,re.M)
    assert len(rows)==20, "20 extracted table rows"
    table={}
    keys=["invocations/count","cantrips/count","prepared-spells/count","pact-slots/count","pact-slots/level"]
    for level,pb,features,*cells in rows:
        n=int(level); cells=list(map(int,cells)); table[n]=(features,cells)
        ident=add("level",str(n),cls,f"Warlock level {n}",level=n)
        rule(ident,"Warlock Features","class-progression",dict(zip(keys,cells)))
        by[ident]["progression"]=dict(proficiency_bonus=int(pb),feature_refs=[],feature_keys=[])
    f={}
    for level,slug,page in [(1,"eldritch-invocations",70),(1,"pact-magic",71),(2,"magical-cunning",72),(3,"warlock-subclass",72),(4,"ability-score-improvement",72),(9,"contact-patron",72),(11,"mystic-arcanum",72),(19,"epic-boon",72),(20,"eldritch-master",72)]:
        f[slug]=add("feature",slug,C+f"level/warlock/{level}",page=page)
    names={by[ident]["name"]:ident for ident in f.values()}
    for n,(features,cells) in table.items():
        r=by[C+f"level/warlock/{n}"]
        for name in features.split(", "):
            if name=="—": continue
            if name=="Subclass feature":
                r["progression"]["feature_keys"].append(K+"subclass-feature")
                r["required_mechanics"].append(K+"subclass-feature")
            elif name.startswith("Mystic Arcanum"):
                spell=int(re.fullmatch(r"Mystic Arcanum \(level (\d) spell\)",name)[1])
                r["progression"]["feature_refs"].append(f["mystic-arcanum"])
                rule(r["id"],name,"",{"mystic-arcanum/spell-level":spell})
            else: r["progression"]["feature_refs"].append(names[name])
    inv=f["eldritch-invocations"]
    rule(inv,"You can’t pick the same invocation more than once","invocations/table-count invocations/gain-more-on-table-increase invocations/recommend-pact-of-the-tome invocations/replace-one-on-warlock-level invocations/cannot-replace-prerequisite-of-owned-invocation invocations/no-duplicates-unless-description-allows",{"invocations/replacement-count":1})
    for text in ("such as Pact of the Tome","you must meet it to learn that invocation","you can replace one of your invocations","requisite for another invocation that you have","unless its description says otherwise"):
        clause(text)
    q=choice("invocations",inv,1,"warlock/invocation-options")
    by[q]["values_by_level"]=[dict(level=n,value=cells[0]) for n,(_,cells) in table.items()]
    pred("invocation-eligibility",q,"invocations/must-meet-all-prerequisites",[cls,inv])
    pred("invocation-replacement",q,"invocations/replacement-must-qualify-and-not-be-required-by-owned-invocation",[inv,q],{"maximum-replacements":1})
    pm=f["pact-magic"]
    rule(pm,"Charisma is the spellcasting","spell-list/warlock spellcasting/charisma spellcasting/arcane-focus pact-slots/distinct-from-spellcasting pact-slots/all-same-level pact-slots/expended-to-cast-level-one-through-five pact-slots/cast-at-slot-level pact-slots/recover-all-on-short-or-long-rest cantrips/learn-on-table-increase cantrips/recommend-eldritch-blast-and-prestidigitation cantrips/replace-one-on-warlock-level prepared-spells/recommend-charm-person-and-hex prepared-spells/add-to-table-count-on-increase prepared-spells/replace-one-on-warlock-level prepared-spells/other-warlock-features-always-prepared-excluded-from-count-but-warlock-spells",{"cantrips/replacement-count":1,"prepared-spells/replacement-count":1,"pact-slots/minimum-spell-level":1,"pact-slots/maximum-spell-level":5})
    for text in ("Eldritch Blast and Prestidigitation","Charm Person and Hex","all of which are the same level","finish a Short or Long Rest","you can replace one of your cantrips","you can replace one spell on your list","always have prepared, those spells don’t count","but those spells otherwise count as Warlock spells for you","Arcane Focus as a Spellcasting Focus","you cast it as a level 3 spell","must be of a level no higher"):
        clause(text)
    can=choice("cantrips",pm,2,"spells/warlock-cantrips")
    prep=choice("prepared-spells",pm,2,"spells/warlock-level-one-through-five")
    slots=resource("pact-slots",pm,"Pact Magic slots","pact-slots/recover-all-on-short-or-long-rest")
    for j,ident in enumerate((can,prep,slots),1):
        by[ident]["values_by_level"]=[dict(level=n,value=cells[j]) for n,(_,cells) in table.items()]
    pred("cantrip-eligibility",can,"cantrips/warlock-list-only",[cls,pm])
    pred("prepared-spell-eligibility",prep,"prepared-spells/warlock-list-level-at-most-current-pact-slot-level",[cls,pm,slots],{"minimum-spell-level":1})
    cunning=f["magical-cunning"]
    rule(cunning,"half your maximum (round up)","magical-cunning/rite-at-end-recover-expended-pact-slots magical-cunning/half-maximum-round-up",{"magical-cunning/rite-minutes":1,"magical-cunning/maximum-divisor":2})
    clause("rite for 1 minute. At the end of it")
    clause("Once you use this feature, you can’t do so again until you finish a Long Rest")
    res=resource("magical-cunning",cunning,key="magical-cunning/recover-on-long-rest")
    pred("magical-cunning-recovery",cunning,"magical-cunning/expended-pact-slots-only-up-to-half-maximum-rounded-up",[slots,res],{"maximum-divisor":2})
    rule(f["warlock-subclass"],"of your Warlock level or lower","subclass/grant-features-at-own-warlock-level-or-lower")
    choice("warlock-subclass",f["warlock-subclass"],1,"warlock/subclasses warlock/subclasses/fiend-patron")
    clause("Fiend Patron subclass")
    for slug in ("ability-score-improvement","epic-boon"):
        rule(f[slug],"feat of your choice for which you qualify",slug+"/qualified-feat-choice")
        q=choice(slug,f[slug],1,"feats/"+slug+" feats/qualified")
        pred(slug+"-qualification",q,"feat/requires-qualification",[cls],option="feats/qualified")
    rule(f["epic-boon"],"Fate is recommended","epic-boon/recommend-boon-of-fate")
    contact=f["contact-patron"]
    rule(contact,"always have the Contact Other Plane","spells/contact-other-plane contact-patron/always-prepared")
    by[contact]["always_prepared"]=[dict(level=9,spell_keys=[K+"spells/contact-other-plane"])]
    modes=choice("contact-patron-casting",contact,1,"warlock/contact-patron/ordinary-spellcasting warlock/contact-patron/free-patron-contact")
    clause("without expending a spell slot to contact your patron")
    clause("automatically succeed on the spell’s saving throw")
    clause("Once you cast the spell with this feature")
    pred("contact-patron-free",modes,"contact-patron/free-cast-contact-own-patron-auto-save-success",[contact,C+"resource/warlock/contact-patron"],option="warlock/contact-patron/free-patron-contact")
    resource("contact-patron",contact,key="contact-patron/free-cast-no-slot-recover-on-long-rest")
    arc=f["mystic-arcanum"]
    rule(arc,"You regain all uses of your Mystic Arcanum","mystic-arcanum/warlock-spell-list mystic-arcanum/no-slot-for-free-cast mystic-arcanum/replace-one-on-warlock-level mystic-arcanum/replacement-same-spell-level mystic-arcanum/recover-all-uses-on-long-rest",{"mystic-arcanum/replacement-count":1})
    clause("once without expending a spell slot")
    clause("finish a Long Rest before you can cast it in this way again")
    clause("replace one of your arcanum spells with another Warlock spell of the same level")
    for level in (11,13,15,17):
        spell=by[C+f"level/warlock/{level}"]["parameters"][K+"mystic-arcanum/spell-level"]
        slug=f"arcanum-{spell}"
        q=choice(slug,arc,1,f"spells/warlock-level-{spell}")
        pred(slug+"-eligibility",q,"mystic-arcanum/warlock-spell-at-specific-level",[cls,arc],{"minimum-warlock-level":level,"spell-level":spell})
        res=resource(slug,arc,f"Level {spell} arcanum free use","mystic-arcanum/separate-free-use-recover-on-long-rest")
        pred(slug+"-casting",arc,"mystic-arcanum/free-use-for-selected-arcanum-only",[q,res],{"minimum-warlock-level":level,"spell-level":spell})
        pred(slug+"-replacement",q,"mystic-arcanum/replace-with-same-level-warlock-spell",[q,arc],{"spell-level":spell})
    rule(f["eldritch-master"],"When you use your Magical Cunning feature, you regain all your expended Pact Magic spell slots","eldritch-master/when-using-magical-cunning eldritch-master/recover-all-expended-pact-slots")
    pred("eldritch-master-recovery",f["eldritch-master"],"eldritch-master/changes-cunning-recovery-only",[cunning,slots])
    return records

def compare(records, expected):
    assert len(records) == len(expected), "record count"
    actual = {r["id"]: r for r in records}
    assert len(actual) == len(records), "duplicate record"
    for r in expected:
        r = copy.deepcopy(r)
        got = copy.deepcopy(actual.get(r["id"]))
        if got is not None: got["required_mechanics"] = sorted(got["required_mechanics"])
        r["required_mechanics"] = sorted(r["required_mechanics"])
        assert got == r, f"semantic fields: {r['id']}"


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
        for j, row in enumerate(r.get("values_by_level", [])):
            change(i, lambda x, j=j: x["values_by_level"][j].__setitem__("value", x["values_by_level"][j]["value"] + 1)) if j else None
        for j, o in enumerate(r.get("options", [])):
            for key in o["quantities"]:
                change(i, lambda x, j=j, key=key: x["options"][j]["quantities"].__setitem__(key, x["options"][j]["quantities"][key] + 1))
        refs = r.get("requirement", {}).get("record_refs", [])
        for j, ref in enumerate(refs):
            other = C + ("feature/warlock/pact-magic" if ref != C + "feature/warlock/pact-magic" else "feature/warlock/magical-cunning")
            if other not in refs:
                change(i, lambda x, j=j, other=other: x["requirement"]["record_refs"].__setitem__(j, other))
        if r["kind"] == "choice" and len(r["option_keys"]) > 1:
            change(i, lambda x: x["option_keys"].pop()) if not r.get("options") else None
        if r["kind"] == "resource":
            change(i, lambda x: (x.__setitem__("capacity", 2), x["values_by_level"][0].__setitem__("value", 2)) if x.get("values_by_level") else x.__setitem__("capacity", 2))
        if r.get("progression"):
            change(i, lambda x: x["progression"].__setitem__("proficiency_bonus", 3 if x["progression"]["proficiency_bonus"] != 3 else 4))
            change(i, lambda x: (x["required_mechanics"].append(K + "invented-scaling"), x["parameters"].__setitem__(K + "invented-scaling", 2)))
        for j, key in enumerate(r["required_mechanics"]):
            if key not in r.get("parameters", {}) and key != r.get("requirement", {}).get("key") and not r.get("options") and not r.get("progression") and not any(key in row["spell_keys"] for row in r.get("always_prepared", [])):
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
    assert c["source"]["changes"] == CHANGES
    compare(c["records"], expected)
    mutations = controls(c["records"])
    for mutated in mutations:
        try: compare(mutated, expected)
        except AssertionError: pass
        else: raise AssertionError("contrary source accepted")
    print(f"Warlock: {len(expected)} semantic records; 20 extracted table rows; {len(mutations)} contrary controls rejected; unavailable metadata only")


if __name__ == "__main__":
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) > 2 else DATA)
