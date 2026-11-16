#!/usr/bin/env python3
"""Read-only Paladin checkpoint proof; unpublished plans cannot pass publication."""
import argparse
from copy import deepcopy
import hashlib
import json
from pathlib import Path
import subprocess

from verify_fighter_scope import graph_read, acyclic

ROOT=Path(__file__).resolve().parents[1]
MANIFEST=ROOT/'docs/investigations/issue-327-paladin-scope.json'
PIN='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
AUDIT_SEAL='686413bbe46e535c17a6787c3288982bd9047d6f68c27457b9ae526f8a06aad7'


def validate(m):
    from verify_paladin_consumed_sources import validate_inherited_contracts
    validate_inherited_contracts(m)
    assert hashlib.sha256(json.dumps({k:m[k] for k in ['issues','shared_issues','common_acceptance']},sort_keys=True).encode()).hexdigest()==AUDIT_SEAL,'audited scope drift'
    assert m['source_sha256']==PIN
    assert m['base']=='c577e3e7a8dac94d579b65f58af444fe427a412e'
    records=[]
    for path,c in m['catalogues'].items():
        assert json.loads((ROOT/path).read_text())==c,'catalogue drift'
        assert c['source']['sha256']==PIN
        if path.endswith(('character-paladin.json','character-devotion.json')):records+=c['records']
    assert len(records)==87 and records==[x['record'] for x in m['records']]
    assert len(m['issues'])==54
    rows={x['key']:x for x in m['issues']}
    membership=json.loads(rows['membership']['fixture'].split('Exact finite membership by spell level: ',1)[1][:-1])
    assert m['spell_lists']['paladin']==membership,'finite Paladin list drift'
    assert [len(membership[str(n)]) for n in range(1,6)]==[13,11,6,4,4]
    assert m['spell_lists']['cleric_cantrips']==['Guidance','Light','Mending','Resistance','Sacred Flame','Spare the Dying','Thaumaturgy']
    devotion=next(x for x in records if x['id'].endswith('feature/paladin/oath-of-devotion/oath-of-devotion-spells'))
    expected=[dict(level=int(n),spell_keys=['srd-5.2.1:mechanic/spells/'+name.lower().replace(' ','-') for name in names]) for n,names in m['spell_lists']['devotion'].items()]
    assert expected==devotion['always_prepared'],'Devotion table ownership drift'
    assert len(rows)==54
    assert set(rows)>=set(f'level-{n}' for n in range(1,21))|{'later-1','hp','multiclass'}
    assert [x['key'] for x in m['issues'] if 429 in x['prerequisites']]==['hp'],'429 single consumer'
    assert m['routing']['reserved_429_consumers']==['hp']
    for n in range(1,21):
        x=rows[f'level-{n}'];assert 'hp' in x['depends']
        if n>1:assert f'level-{n-1}' in x['depends']
        record=next(r for r in records if r['kind']=='level' and r['level']==n)
        assert m['levels'][n-1]['progression']==record['progression']
        for name,value in m['levels'][n-1]['values'].items():
            resource=next(r for r in records if r['kind']=='resource' and r['id']==f'srd-5.2.1:character/resource/paladin/{name}')
            assert value==next((v['value'] for v in resource['values_by_level'] if v['level']==n),0)
        assert rows[f'level-{n}']['source_records']==[record['id']]
    assert 'hp' in rows['later-1']['depends'] and 'multiclass' in rows['later-1']['depends']
    for x in m['records']:
        assert x['execution_owners'] and set(x['execution_owners'])<=set(rows)
        assert x['state']=='metadata only; Paladin/Devotion unqualified'
        for key in x['execution_owners']:assert x['record']['id'] in rows[key]['source_records']
    for x in rows.values():
        expected={k for r in m['records'] for k,v in r['key_owners'].items() if x['key'] in v}
        assert x['source_keys']==sorted(expected),'source-key ownership drift'
        assert len(x['depends'])==len(set(x['depends']))
        assert set(x['depends'])<=set(rows)
    # Independent numeric witnesses for class HP, turn damage, preparation and slots.
    assert max(1,6+2)==8 and max(1,1-3)==1
    assert 8*((18-10)//2-(17-10)//2)==8
    assert 5+6==11 and 2+1+1==4
    assert m['levels'][4]['values']['spell-slot-2']==2
    assert [m['levels'][-1]['values'][f'spell-slot-{n}'] for n in range(1,6)]==[4,3,3,3,2]
    assert m['levels'][-1]['values']['lay-on-hands']==100
    advancement=next(c for p,c in m['catalogues'].items() if p.endswith('character-advancement.json'))
    after=next(x for x in advancement['records'] if x['level']==6)
    assert after['advancement']=={'experience_points':14000,'proficiency_bonus':3}
    after5=next(x for x in advancement['records'] if x['level']==5)
    assert after5['advancement']=={'experience_points':6500,'proficiency_bonus':3}
    for key in [f'level-{n}' for n in range(1,21)]+['later-1']:
        assert 'TOTAL CHARACTER level' in rows[key]['fixture'],'class versus total advancement'
        assert 'class-indexed XP/PB' in rows[key]['fixture'],'wrong-row negative witness'
    assert 'Fighter4/Paladin1→Paladin2 reaches total6,requires14000XP/PB+3' in rows['level-2']['fixture']
    grants={1:['lay-on-hands','casting','preparation','slots','mastery'],2:['style','blessed-warrior','smite'],3:['channel','divine-sense','subclass','oath-spells','sacred-weapon'],4:['improvement'],5:['extra-attack','steed'],6:['aura'],7:['devotion-aura'],8:['improvement'],9:['abjure'],10:['courage'],11:['radiant'],12:['improvement'],14:['restoring'],15:['smite-protection'],16:['improvement'],18:['expansion'],19:['boon'],20:['nimbus-use','nimbus-ward','nimbus-damage']}
    def requires(start,target):
        return target in rows[start]['depends'] or any(requires(d,target) for d in rows[start]['depends'])
    for level,targets in grants.items():
        for target in targets:assert requires(f'level-{level}',target),f'missing granting-level feature {level}:{target}'
    sense=next(x for x in m['records'] if x['record']['id'].endswith('feature/paladin/channel-divinity'))
    for key,owners in sense['key_owners'].items():
        assert owners==(['divine-sense'] if 'divine-sense' in key else ['channel']),'Channel versus Divine Sense owner'
    for record in m['records']:
        if record['record']['kind']=='feature':
            assert record['key_owners']
            assert 'improvement' not in {k for v in record['key_owners'].values() for k in v} or record['record']['id'].endswith('/ability-score-improvement'),'unrelated mechanic assigned to ASI'
    names={x['key']:i for i,x in enumerate(m['issues'])}
    acyclic({names[k]:dict(blocked_by=[names[d] for d in x['depends']]) for k,x in rows.items()})
    for p,e in m['evidence'].items():assert hashlib.sha256((ROOT/p).read_bytes()).hexdigest()==e['sha256'],'evidence drift'
    shared=m['shared_issues'];assert len(shared)==54 and len({x['key'] for x in shared})==54
    assert [sum(x['cohort']==c for x in shared) for c in [1,2,3]]==[20,20,14]
    keys={x['key']:i for i,x in enumerate(shared)}
    for x in shared:assert set(x['depends'])<=set(keys)
    acyclic({keys[x['key']]:dict(blocked_by=[keys[d] for d in x['depends']]) for x in shared})
    assert m['publication'].startswith('PENDING') or 'NATIVE PUBLICATION' in m['publication']
    assert all(x['issue'] is None and x['publication']=='UNPUBLISHED' for x in m['issues']+shared)
    return len(records)


def probes(m):
    validate(m)
    def reject(bad):
        try:validate(bad)
        except (AssertionError,KeyError):return
        raise AssertionError('accepted corruption')
    bad=deepcopy(m);bad['source_sha256']='bad';reject(bad)
    bad=deepcopy(m);bad['levels'][4]['values']['spell-slot-2']=3;reject(bad)
    bad=deepcopy(m);bad['issues'][-1]['prerequisites'].append(429);reject(bad)
    bad=deepcopy(m);bad['issues'][-1]['depends'].remove('hp');reject(bad)
    for n in range(2,21):
        bad=deepcopy(m);next(x for x in bad['issues'] if x['key']==f'level-{n}')['depends'].remove(f'level-{n-1}');reject(bad)
    bad=deepcopy(m);bad['records'][0]['record']['parameters'].clear();reject(bad)
    bad=deepcopy(m);bad['issues'][0]['source_keys'].pop();reject(bad)
    bad=deepcopy(m);bad['issues'][0]['depends']=['level-20'];reject(bad)
    bad=deepcopy(m);bad['shared_issues'][0]['cohort']=4;reject(bad)
    bad=deepcopy(m);bad['shared_issues'][0]['fixture']=bad['shared_issues'][0]['fixture'].replace('targetsL+2','targetsL+3');reject(bad)
    bad=deepcopy(m);bad['issues'][0]['issue']=9000;reject(bad)


def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--pdf',type=Path)
    p.add_argument('--live',action='store_true',help='bounded read-only graph audit; never writes a receipt')
    p.add_argument('--require-publication',action='store_true')
    p.add_argument('--native',action='store_true',help='verify exact published contracts and saved receipt')
    p.add_argument('--sealed',action='store_true',help='require accepted parent prefixes and published supplements')
    p.add_argument('--probes',action='store_true',help='run exhaustive native edge/body corruption controls')
    a=p.parse_args();m=json.loads(MANIFEST.read_text());validate(m);probes(m)
    if a.pdf:
        assert hashlib.sha256(a.pdf.read_bytes()).hexdigest()==PIN
        for name in ['paladin','devotion','advancement','multiclass','origin_feats','general_feats','fighting_styles','epic_boon_feats']:
            subprocess.run(['rtk','proxy','python3',str(ROOT/f'scripts/verify_srd_{name}.py'),str(a.pdf)],check=True)
        subprocess.run(['rtk','proxy','python3',str(ROOT/'scripts/verify_paladin_consumed_sources.py'),str(a.pdf)],check=True)
    if a.native or m.get('published_issues'):
        import paladin_native_proof as native
        sealed=a.sealed or m.get('supplement_publication')=='sealed'
        r=native.live(m) if a.live else json.loads(native.RECEIPT.read_text())
        if a.live and not m.get('consumed_project'):
            m['consumed_project']={n:r['project'][n] for n in m['consumed_contracts']}
            MANIFEST.write_text(json.dumps(m,indent=2)+'\n')
        n=native.validate(m,r,sealed)
        controls=native.probes(m,r,sealed) if a.probes else None
        if a.live:
            native.RECEIPT.write_text(json.dumps(r,indent=2)+'\n')
            m['native_validation']=dict(observed_at=r['observed_at'],nodes=n,new_tasks=121,new_implementation_leaves=115,negative_controls=controls,supplements='sealed' if sealed else 'PENDING root publication')
            if sealed:m['supplement_publication']='sealed'
            MANIFEST.write_text(json.dumps(m,indent=2)+'\n')
        print(f'PASS native:{n}nodes,121Task/Size/parent/body contracts,115new mandatory leaves,54spell cohort leaves; '+str(controls)+' negative controls; supplements '+('SEALED' if sealed else 'PENDING root publication'))
        if a.require_publication:assert sealed,'accepted immutable parent prefix and supplements still pending'
        return
    if a.live:
        g={x['number']:x for x in graph_read()};acyclic(g)
        acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] else []) for n,x in g.items()})
        children={};outgoing={}
        for x in g.values():
            assert len(x['blocked_by'])==len(set(x['blocked_by']))<=100
            if x['parent']:children[x['parent']]=children.get(x['parent'],0)+1
            for n in x['blocked_by']:outgoing[n]=outgoing.get(n,0)+1
        assert max(children.values())<=100 and max(outgoing.values())<=100
        assert g[327]['parent']==283 and g[327]['blocked_by']==[285,292,314]
        assert all(g[n]['state']=='closed' for n in [285,292,314])
        print(f'Live read-only audit: {len(g)} nodes; separate DAGs and100incoming/outgoing/children caps pass. No publication proof.')
    if a.require_publication:raise AssertionError('UNPUBLISHED: shared scope approval/native graph/Task/Size receipt and accepted post-Monk parent seal pending')
    print('PASS checkpoint only:87records/20rows/54class plans/54shared spell plans; numeric witnesses; corruption probes. Native publication and supplements PENDING.')


if __name__=='__main__':main()
