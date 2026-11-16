"""Direct pinned PDF list/header reconciliation; never a runtime admission registry."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
from sorcerer_contract_oracle import SOURCE_LISTS
from sorcerer_spell_oracle import CONFLICT_POLICY

PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
ROOT = Path(__file__).resolve().parents[1]


def column(pdf, page, x):
    return subprocess.run(['pdftotext', '-f', str(page), '-l', str(page), '-x', str(x), '-y', '0', '-W', '297', '-H', '783', '-layout', str(pdf), '-'], check=True, capture_output=True, text=True).stdout


def reconcile(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN
    lists = {}
    school = r'(?:Abjuration|Conjuration|Divination|Enchantment|Evocation|Illusion|Necromancy|Transmutation)'
    for cls, pages in [('sorcerer', range(67,70)), ('warlock', range(75,77)), ('wizard', range(79,83))]:
        rows, level = [], 0
        for page in pages:
            for x in [0,297]:
                for line in column(pdf,page,x).splitlines():
                    h = re.search(r'Level (\d) '+cls.title()+r' Spells',line)
                    if h:
                        level=int(h[1])
                    if 'Cantrips (Level 0' in line:
                        level=0
                    match=re.search(r'^\s*([A-Za-z][A-Za-z’\x27 /-]*?)\s{2,}('+school+r')\s*(.*)',line)
                    if match:
                        rows.append(dict(name=match[1].strip(),level=level,page=page,school=match[2],special=match[3].strip()))
        assert rows == SOURCE_LISTS[cls], 'actual complete printed '+cls+' list'
        lists[cls]={r['name'] for r in rows}
    inventory=json.loads((ROOT/'docs/investigations/issue-80-spell-item-inventory.json').read_text())['records']
    names={r['name'] for r in inventory if r['family']=='spells'}
    lines=[]
    for page in range(107,176):
        for x in [0,297]:
            lines += [(page,' '.join(s.replace('\u00ad','').split())) for s in column(pdf,page,x).splitlines() if s.strip()]
    headers={}
    for i,(page,line) in enumerate(lines):
        if line not in names or i+1>=len(lines) or not re.match(r'(?:Level [1-9] |\w+ Cantrip)',lines[i+1][1]):
            continue
        words=[]
        for _,s in lines[i+1:i+8]:
            if 'Casting Time:' in s:
                break
            words.append(s)
        headers[line]=dict(page=page,text=' '.join(words))
    assert len(headers)==339, 'all spell headers'
    differences=[]
    for cls,printed in lists.items():
        headed={name for name,h in headers.items() if re.search(r'\b'+cls.title()+r'\b',h['text'])}
        assert not printed-headed, 'unrecorded printed-only conflict'
        differences.extend(dict(**p) for p in CONFLICT_POLICY.values() if p['class']==cls)
        assert headed-printed == {p['spell'] for p in differences if p['class']==cls}, 'unrecorded header-only conflict'
    for conflict in differences:
        assert headers[conflict['spell']]['page']==conflict['header_page']
        assert conflict['spell'] not in lists[conflict['class']]
    from sorcerer_spell_oracle import SOURCE_PAGES
    for name,page in SOURCE_PAGES.items():
        assert name in headers and headers[name]['page']==page, 'every consumed source header/page'
    return dict(printed_counts={k:len(v) for k,v in lists.items()},header_counts={k:sum(bool(re.search(r'\b'+k.title()+r'\b',h['text'])) for h in headers.values()) for k in lists},conflicts=differences,admission_policy='printed list plus separately proven explicit feature grants; disputed list-based admission unavailable')
