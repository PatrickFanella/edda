#!/usr/bin/env python3
"""Independent offline Wizard source graph comparator, pinned SRD pp77–79.

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
K = M + "wizard/"
DATA = ROOT / "internal/rules/srd521/testdata/character-wizard.json"


def source_records(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN, "PDF pin"
    layout = subprocess.check_output(["pdftotext", "-f", "77", "-l", "79", "-layout", str(pdf), "-"], text=True)
    # Read columns independently to keep sentences intact across column breaks.
    columns = []
    for page in (77, 78, 79):
        for x in (0, 310):
            columns.append(subprocess.check_output(["pdftotext", "-f", str(page), "-l", str(page),
                "-x", str(x), "-y", "0", "-W", "310", "-H", "792", "-layout", str(pdf), "-"], text=True))
    flat = " ".join(re.sub(r"-\n\s*", "", " ".join(columns)).replace("\u00ad", "").split())
    records, by = [], {}

    def add(kind, slug, parent, name=None, level=0, page=0):
        ident = C + kind + "/wizard" + ("/" + slug if slug else "")
        if kind == "class":
            src = dict(inventory_id="srd-5.2.1:planning/classes/wizard", page=77)
        elif kind == "feature":
            src = dict(inventory_id=f"srd-5.2.1:planning/class-features/wizard-level-{by[parent]['level']}-{slug}", page=page)
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


    cls = add('class', '', '', 'Wizard')
    rule(cls, 'Intelligence and Wisdom', 'primary-ability/intelligence saves/intelligence saves/wisdom weapons/simple armor/none level-1/all-core-traits-and-level-1-features multiclass/hit-die-and-level-1-features multiclass/slots-from-character-creation', {'hit-die-sides': 6})
    choice('skills', cls, 2, 'skills/arcana skills/history skills/insight skills/investigation skills/medicine skills/nature skills/religion')
    eq = choice('starting-equipment', cls, 1, 'wizard/equipment/a wizard/equipment/b')
    by[eq]['options'] = []
    for key, bundle in zip(by[eq]['option_keys'], [{'dagger':2,'arcane-focus-quarterstaff':1,'robe':1,'spellbook':1,'scholars-pack':1,'gp':5},{'gp':55}]):
        quantities = {M+('currency/' if item == 'gp' else 'equipment/')+item:n for item,n in bundle.items()}
        by[eq]['options'].append(dict(key=key, quantities=quantities))
        for k in quantities:
            if k not in by[eq]['required_mechanics']: by[eq]['required_mechanics'].append(k)
    assert '2 Daggers' in flat and '55 GP' in flat and 'and 5 GP' in flat
    rows = re.findall(r'^\s*(\d+)\s+\+(\d)\s+(.+?)\s{2,}(\d+)\s+(\d+)\s+([\d—\s]+)$', layout, re.M)
    assert len(rows) == 20, f'20 Wizard rows: {len(rows)}'
    table = {}
    for level,pb,features,cantrips,prepared,slots in rows:
        n=int(level); cells=slots.split(); assert len(cells)==9
        r=add('level',str(n),cls,f'Wizard level {n}',level=n)
        nums={'cantrips':int(cantrips),'prepared-spells':int(prepared)}
        nums.update({f'spell-slots-level-{i+1}':0 if v=='—' else int(v) for i,v in enumerate(cells)})
        rule(r, 'Wizard Features', 'class-progression')
        pred('table-'+str(n),r,'table/cantrips-prepared-and-nine-slot-columns',[cls],nums)
        by[r]['progression']=dict(proficiency_bonus=int(pb),feature_refs=[],feature_keys=[])
        table[n]=features
    f={}
    for level,slug,page in [(1,'spellcasting',77),(1,'ritual-adept',78),(1,'arcane-recovery',78),(2,'scholar',78),(3,'wizard-subclass',78),(4,'ability-score-improvement',79),(5,'memorize-spell',79),(18,'spell-mastery',79),(19,'epic-boon',79),(20,'signature-spells',79)]:
        f[slug]=add('feature',slug,C+f'level/wizard/{level}',slug.replace('-',' ').title(),page=page)
    names={by[v]['name']:v for v in f.values()}
    for n,text in table.items():
        # The first row wraps Ritual Adept and Arcane Recovery onto two lines.
        labels=text.split(', ')
        if n==1:
            continuation=re.search(r'^\s+(Arcane Recovery)\s*$',layout,re.M)
            assert continuation, 'first row continuation'
            labels=[v.rstrip(',') for v in labels if v]+[continuation[1]]
        for label in labels:
            r=by[C+f'level/wizard/{n}']
            if label=='Subclass feature':
                r['progression']['feature_keys'].append(K+'subclass-feature'); r['required_mechanics'].append(K+'subclass-feature')
            elif label!='—': r['progression']['feature_refs'].append(names[label])
    sc=f['spellcasting']
    rule(sc,'Whenever you finish a Long Rest','spellcasting/wizard-spell-list spellcasting/intelligence spellcasting/arcane-focus-or-spellbook spellcasting/cantrip-replace-one-on-long-rest spellcasting/add-cantrip-at-levels-four-and-ten spellcasting/regain-all-slots-on-long-rest spellcasting/prepare-from-spellbook-with-available-slot-level spellcasting/add-prepared-when-table-increases spellcasting/always-prepared-excluded-from-count-but-wizard-spells spellcasting/change-any-prepared-from-book-on-long-rest',{'spellcasting/initial-cantrips':3,'spellcasting/initial-prepared':4,'spellcasting/cantrip-replacement-count':1})
    can=choice('cantrips',sc,3,'spells/wizard-cantrips')
    rule(can,'Light, Mage Hand, and Ray of Frost','cantrips/recommend-light-mage-hand-ray-of-frost')
    by[can]['values_by_level']=[dict(level=n,value=3 if n<4 else 4 if n<10 else 5) for n in range(1,21)]
    book=choice('spellbook-initial',sc,6,'spells/wizard-level-one')
    rule(book,'Detect Magic, Feather Fall','spellbook/recommend-detect-magic-feather-fall-mage-armor-magic-missile-sleep-thunderwave spellbook/contains-known-level-one-plus-spells spellbook/appearance-and-materials-chosen spellbook/tiny spellbook/read-only-self-or-identify',None)
    rule(sc,'contains 100 pages','spellbook/dimensions',{'spellbook/weight-pounds':3,'spellbook/pages':100})
    gain=choice('spellbook-level-gain',sc,2,'spells/wizard-level-one-plus')
    rule(gain,'after 1','spellbook/two-per-wizard-level-after-one')
    pred('spellbook-level-gain',gain,'spellbook/requires-own-table-slot-level',[cls,sc],{'minimum-wizard-level':2})
    prep=choice('prepared-spells',sc,4,'spells/own-wizard-spellbook')
    by[prep]['values_by_level']=[dict(level=int(row[0]),value=int(row[4])) for row in rows]
    pred('prepared-slot-level',prep,'preparation/requires-book-and-own-slot-level',[book,sc],{'minimum-spell-level':1})
    copyspell=choice('copy-found-spell',sc,1,'spells/found-wizard-level-one-plus')
    pred('copy-found-spell',copyspell,'copy/requires-preparable-level-and-time',[prep,book],{'hours-per-spell-level':2,'gp-per-spell-level':50,'minimum-spell-level':1})
    rule(copyspell,'Afterward you can prepare','copy/after-transcription-prepare-like-other-book-spells copy/source-can-be-spell-scroll')
    backup=choice('copy-own-book',sc,1,'spells/own-wizard-spellbook')
    pred('copy-own-book',backup,'copy/own-known-book-to-another-book',[book],{'hours-per-spell-level':1,'gp-per-spell-level':10})
    recovery=choice('replace-lost-book',sc,1,'spells/own-prepared-wizard-spells')
    pred('replace-lost-book',recovery,'copy/lost-book-only-prepared-spells',[prep,backup],{'hours-per-spell-level':1,'gp-per-spell-level':10})
    rule(recovery,'Filling out the remainder','copy/remainder-requires-finding-new-spells')
    ritual=f['ritual-adept']
    rule(ritual,'needn’t have the spell prepared','ritual/not-required-prepared ritual/must-read-from-book')
    pred('ritual-eligibility',ritual,'ritual/requires-ritual-tag-and-in-own-book',[book,sc])
    arc=f['arcane-recovery']
    rule(arc,'When you finish a Short Rest','arcane-recovery/study-spellbook-on-short-rest arcane-recovery/choose-expended-slots arcane-recovery/combined-level-half-wizard-level-round-up',{'arcane-recovery/budget-divisor':2,'arcane-recovery/maximum-slot-level':5})
    pred('arcane-recovery-budget',arc,'arcane-recovery/requires-book-and-class-budget',[cls,book],{'level-divisor':2,'round-up':1,'maximum-slot-level':5})
    res=add('resource','arcane-recovery',arc)
    by[res].update(capacity=1,required_mechanics=[K+'arcane-recovery/recover-on-long-rest'])
    rule(f['scholar'],'skills in which you have proficiency','scholar/choose-proficient-listed-skill-for-expertise')
    scholar=choice('scholar',f['scholar'],1,'skills/arcana skills/history skills/investigation skills/medicine skills/nature skills/religion')
    rule(scholar,'You have Expertise in the chosen skill','scholar/expertise-in-chosen-skill')
    pred('scholar-proficiency',scholar,'scholar/requires-proficiency',[cls,scholar])
    rule(f['wizard-subclass'],'of your Wizard level or lower','subclass/grant-features-at-own-wizard-level-or-lower')
    choice('wizard-subclass',f['wizard-subclass'],1,'wizard/subclasses wizard/subclasses/evoker')
    for slug in ('ability-score-improvement','epic-boon'):
        feat=choice(slug,f[slug],1,'feats/'+slug+' feats/qualified')
        pred(slug+'-qualification',feat,'feat/requires-qualification',[cls],option='feats/qualified')
    rule(f['ability-score-improvement'],'8, 12, and 16','asi/repeat-at-eight-twelve-sixteen')
    rule(f['epic-boon'],'Boon of Spell Recall is recommended','epic-boon/recommend-boon-of-spell-recall')
    mem=f['memorize-spell']
    rule(mem,'Whenever you finish a Short Rest','memorize/on-short-rest-study-book memorize/replace-one-prepared-spellcasting-wizard-level-one-plus-with-another-level-one-plus-from-book',{'memorize/replacement-count':1,'memorize/minimum-spell-level':1})
    pred('memorize-replacement',mem,'memorize/requires-spellcasting-prepared-and-book',[prep,book,sc])
    mastery=f['spell-mastery']
    rule(mastery,'at their lowest level','mastery/always-prepared mastery/at-will-lowest-level-no-slot mastery/higher-level-requires-slot mastery/long-rest-study-book-replace-one-with-eligible-same-level',{'mastery/replacement-count':1})
    for level in (1,2):
        ch=choice(f'mastery-level-{level}',mastery,1,'spells/own-wizard-spellbook')
        pred(f'mastery-level-{level}',ch,'mastery/requires-book-spell-of-level-and-action-casting-time',[book,sc],{'spell-level':level,'casting-time-actions':1})
    sig=f['signature-spells']
    ch=choice('signature-spells',sig,2,'spells/own-wizard-spellbook-level-three')
    pred('signature-eligibility',ch,'signature/requires-two-distinct-level-three-book-spells',[book,sc],{'spell-level':3})
    rule(sig,'You always have these spells','signature/always-prepared signature/each-selected-spell-separate-free-cast signature/higher-level-requires-slot')
    for which in ('first','second'):
        res=add('resource','signature-'+which,sig)
        by[res].update(capacity=1,required_mechanics=[K+'signature/recover-on-short-or-long-rest',K+'signature/free-cast-at-level-three-no-slot',K+'signature/bound-to-'+which+'-distinct-selected-spell'])
        pred('signature-'+which+'-binding',sig,'signature/resource-bound-to-selected-spell',[ch,res],{'selection-index':1 if which=='first' else 2,'cast-level':3})
    assert '2 hours and costs 50 GP' in flat and '1 hour and 10 GP' in flat
    assert 'half your Wizard level (round up)' in flat
    assert 'none of the slots can be level 6 or higher' in flat
    assert 'Short or Long Rest' in flat

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
        for j, option in enumerate(r.get("options", [])):
            for key in option["quantities"]:
                change(i, lambda x,j=j,key=key: x["options"][j]["quantities"].__setitem__(key,x["options"][j]["quantities"][key]+1))
        for j in range(1,len(r.get("values_by_level", []))):
            change(i, lambda x,j=j: x["values_by_level"][j].__setitem__("value",x["values_by_level"][j]["value"]+1))
        for key in r.get("parameters", {}):
            change(i, lambda x, key=key: x["parameters"].__setitem__(key, x["parameters"][key] + 1))
        for key in r.get("requirement", {}).get("parameters", {}):
            change(i, lambda x, key=key: x["requirement"]["parameters"].__setitem__(key, x["requirement"]["parameters"][key] + 1))
        refs = r.get("requirement", {}).get("record_refs", [])
        for j, ref in enumerate(refs):
            other = C + ("feature/wizard/ritual-adept" if ref != C + "feature/wizard/ritual-adept" else "feature/wizard/scholar")
            if other not in refs:
                change(i, lambda x, j=j, other=other: x["requirement"]["record_refs"].__setitem__(j, other))
        if r["kind"] == "choice" and len(r["option_keys"]) > 1:
            change(i, lambda x: x["option_keys"].pop()) if not r.get("options") else None
        if r["kind"] == "resource":
            change(i, lambda x: x.__setitem__("capacity", 2))
        if r.get("progression"):
            change(i, lambda x: x["progression"].__setitem__("proficiency_bonus", 3 if x["progression"]["proficiency_bonus"] != 3 else 4))
            change(i, lambda x: (x["required_mechanics"].append(K + "invented-scaling"), x.setdefault("parameters", {}).__setitem__(K + "invented-scaling", 2)))
        for j, key in enumerate(r["required_mechanics"]):
            if key not in r.get("parameters", {}) and key != r.get("requirement", {}).get("key") and not r.get("options") and not r.get("progression"):
                change(i, lambda x, j=j: x["required_mechanics"].__setitem__(j, K + "contrary-source"))
                if len(r["required_mechanics"]) > 1:
                    change(i, lambda x, j=j: x["required_mechanics"].pop(j))
    return out


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        assert key not in result, f"duplicate JSON key: {key}"
        result[key] = value
    return result


def verify(pdf, candidate=DATA):
    expected = source_records(pdf)
    c = json.loads(candidate.read_text(), object_pairs_hook=unique_object)
    assert set(c) == {"schema_version", "source", "records"}, "catalogue keys"
    assert c["schema_version"] == 1
    pin_source = json.loads((ROOT / "docs/investigations/issue-74-character-inventory.json").read_text())["source"]
    assert set(c["source"]) == set(pin_source), "source keys"
    for key in pin_source:
        if key != "changes":
            assert c["source"][key] == pin_source[key], f"source {key}"
    assert c["source"]["changes"] == "Wizard source identities, all 20 progression rows, nested choices, prerequisites, costs, limits and recovery on PDF pp77-79 adapted into unavailable offline metadata. Mechanic keys and graph are Edda additions. Evoker, spell and feat definitions remain opaque. No executable support imported."
    compare(c["records"], expected)
    mutations = controls(c["records"])
    for mutated in mutations:
        try: compare(mutated, expected)
        except AssertionError: pass
        else: raise AssertionError("contrary source accepted")
    print(f"Wizard: {len(expected)} semantic records; 20 extracted table rows; {len(mutations)} contrary controls rejected; unavailable metadata only")


if __name__ == "__main__":
    verify(Path(sys.argv[1]), Path(sys.argv[2]) if len(sys.argv) > 2 else DATA)
