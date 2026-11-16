#!/usr/bin/env python3
"""Strict Rogue source/native proof; FINAL prefix seal is a separate ROOT release gate."""
import argparse
from collections import Counter
import hashlib
import json
import re
from pathlib import Path
import subprocess
from verify_fighter_scope import graph_read,acyclic
from rogue_capacity import project

ROOT=Path(__file__).resolve().parents[1]
PATH=ROOT/'docs/investigations/issue-329-rogue-scope.json'
RECEIPT=ROOT/'docs/investigations/issue-329-link-verification.json'
PIN='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
AUDIT_FIELDS=['catalogues','records','issues','grants','evidence','common_acceptance','dispatch_inventory','unqualified_dispatch_ids','dispatch_bindings','reservation']
AUDIT_SEAL='34dfe34f09e59954d23b0ed9e1d5b5fc7178f77514e2f9dffadaf0d0126daec2'

def digest(x):return hashlib.sha256(json.dumps(x,sort_keys=True).encode()).hexdigest()

def pages(path,api,limit=3):
    assert '?' not in path
    result=[];seen=set()
    for page in range(1,limit+1):
        batch=api(f'{path}?per_page=100&page={page}');assert isinstance(batch,list) and len(batch)<=100
        ids=[x['id'] for x in batch];assert len(set(ids))==len(ids) and not seen.intersection(ids),'repeated page'
        seen.update(ids);result.extend(batch)
        if len(batch)<100:return result
    raise AssertionError('pagination call cap')

def parent_append(prefix,section,actual):
    assert actual.startswith(prefix+section),'original-only rollback or changed prefix/section'
    assert actual.count('<!-- edda-329:start -->')==actual.count('<!-- edda-329:end -->')==1,'altered duplicate marker'
    markers=re.findall(r'<!--\s*edda-329\s*:\s*(start|end)\s*-->',actual,re.I)
    assert [x.lower() for x in markers].count('start')==[x.lower() for x in markers].count('end')==1,'altered duplicate marker'
    assert len(actual)<=65536

def caps(graph):
    from rogue_native_proof import caps as check
    check(graph)

def multiclass_ownership(m):
    # The reduced multiclass grant includes proficiencies and level-1 features;
    # its Hit Die is supplied by the separate HP adapter, not its grant owner.
    composite='srd-5.2.1:mechanic/rogue/multiclass/hit-die-one-class-skill-thieves-tools-light-armor-and-level-1-features'
    hit_die='srd-5.2.1:mechanic/rogue/hit-die-sides'
    record=next(x for x in m['records'] if x['record']['id']=='srd-5.2.1:character/class/rogue')
    assert record['key_owners'][composite]==['multiclass'],'composite multiclass grant owner'
    assert record['key_owners'][hit_die]==['hp'],'standalone Hit Die owner'
    rows={x['key']:x for x in m['issues']}
    assert composite in rows['multiclass']['source_keys'] and composite not in rows['hp']['source_keys'],'composite multiclass source keys'
    assert hit_die in rows['hp']['source_keys'] and hit_die not in rows['multiclass']['source_keys'],'standalone Hit Die source keys'
    assert 'hp' in rows['multiclass']['depends'],'multiclass HP dependency'
    if m.get('published_issues'):
        published={x['key']:x for x in m['published_issues']}
        assert composite in published['multiclass']['source_keys'] and composite not in published['hp']['source_keys'],'published composite multiclass source keys'
        assert published['hp']['issue'] in published['multiclass']['blocked_by_issues'],'published multiclass HP dependency'

def semantic(m):
    multiclass_ownership(m)
    # Reconstruct complete source/owner contracts independently of the native receipt.
    from build_rogue_checkpoint import audit
    expected=json.loads(json.dumps(audit(write=False)))
    for k in AUDIT_FIELDS:
        assert json.dumps(m[k],sort_keys=True)==json.dumps(expected[k],sort_keys=True),f'complete typed actual source/contract:{k}'
    rows={x['key']:x for x in m['issues']};assert len(rows)==len(m['issues'])==72
    assert m['source_sha256']==PIN and m['base']=='952fab917329e00cceebf46a8bacd59cb1c30e99'
    for n in range(1,21):
        row=rows[f'level-{n}'];assert {'hp','multiclass'}<=set(row['depends'])
        if n>1:
            assert f'level-{n-1}' in row['depends']
            assert f'exact adjacent Rogue{n-1}->{n} ONLY' in row['fixture']
            assert 'first-character1 or adjacent' not in row['fixture'] and 'first1 consumes' not in row['fixture']
        else:assert 'first-character Rogue1 creation ONLY' in row['fixture']
        # Grant owners are source identities, never assigned to an unrelated ASI row.
        canonical=next(r['record'] for r in m['records'] if r['record']['kind']=='level' and r['record']['level']==n)
        for ident in canonical['progression']['feature_refs']:
            feature=next(r for r in m['records'] if r['record']['id']==ident)
            assert set(feature['execution_owners'])<=set(row['depends']),f'granting-level owner{n}:{ident}'
    for x in m['records']:
        r=x['record'];assert set(x['key_owners'])==set(r['required_mechanics']+r['option_keys'])
        assert all(set(v)<=set(rows) and v for v in x['key_owners'].values())
        if r['kind']=='feature' and not r['id'].endswith('/ability-score-improvement'):assert 'improvement' not in x['execution_owners']
    assert {'evasion','reliable-talent'}<=set(rows['level-7']['depends'])
    assert {'hp','multiclass'}<=set(rows['later-1']['depends'])
    assert rows['hp']['prerequisites']==[968,58] and not any(429 in x['prerequisites'] for x in rows.values())
    for p,x in m['evidence'].items():assert hashlib.sha256((ROOT/p).read_bytes()).hexdigest()==x['sha256']
    d=m['dispatch_inventory'];assert len(d)==597 and Counter(x['record']['family'] for x in d)=={'spells':339,'magic-items':258}
    assert len({x['record']['id'] for x in d})==597
    assert m['unqualified_dispatch_ids']==[x['record']['id'] for x in d if x['effect_owner'] is None]
    for n in [47,54,283,337]:
        section=(ROOT/f'docs/investigations/issue-329-parent-{n}-PENDING.md').read_text()
        assert len(section)<=(90 if n==47 else 750);parent_append('',section,section)
        from rogue_native_proof import SUPPLEMENT_SEALS
        assert hashlib.sha256(section.encode()).hexdigest()==SUPPLEMENT_SEALS[str(n)]

def validate(m):
    assert digest({k:m[k] for k in AUDIT_FIELDS})==AUDIT_SEAL,'independent audited source/contract seal'
    semantic(m);return m

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--pdf',required=True);p.add_argument('--live',action='store_true');p.add_argument('--publication',action='store_true');p.add_argument('--sealed',action='store_true');p.add_argument('--write-receipt',action='store_true');a=p.parse_args()
    assert hashlib.sha256(Path(a.pdf).read_bytes()).hexdigest()==PIN
    m=validate(json.loads(PATH.read_text()))
    for name in ['rogue','thief','advancement','multiclass','origin_feats','general_feats','fighting_styles','epic_boon_feats']:
        subprocess.run(['rtk','proxy','python3',str(ROOT/f'scripts/verify_srd_{name}.py'),a.pdf],check=True,stdout=subprocess.DEVNULL)
    from verify_rogue_consumed_sources import verify
    verify(Path(a.pdf),m)
    from rogue_native_proof import collect,validate as native
    if a.live:
        r=collect(m)
        if a.write_receipt:RECEIPT.with_suffix('.candidate.json').write_text(json.dumps(r,indent=2)+'\n')
        result=native(m,r,sealed=a.sealed)
        if a.write_receipt:
            RECEIPT.with_suffix('.candidate.json').replace(RECEIPT)
    else:
        assert not a.write_receipt,'live observation required'
        r=json.loads(RECEIPT.read_text());result=native(m,r,sealed=a.sealed)
    print('Rogue exact publication proof PASS',json.dumps(result),';runtime remains unavailable')

if __name__=='__main__':main()
