#!/usr/bin/env python3
"""Strict Fighter source/ownership/native proof. Live reads only; supplements pending until sealed."""
import argparse
from copy import deepcopy
import hashlib
import json
import re
from pathlib import Path
from datetime import datetime, timezone
import subprocess
from verify_barbarian_scope import acyclic, api as github_api, digest

API_CALL_CAP=500
_api_calls=0
def api(path,data=None):
    global _api_calls
    _api_calls+=1
    assert _api_calls<=API_CALL_CAP,"total API call cap"
    return github_api(path,data)
REPO="repos/subculture-collective/edda"

ROOT=Path(__file__).resolve().parents[1]
MANIFEST=ROOT/'docs/investigations/issue-325-fighter-scope.json'
RECEIPT=ROOT/'docs/investigations/issue-325-link-verification.json'
PIN='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
ACCEPTED_DRUID_MERGE='3edec0379e876d30aa611bffa50674f731f61786'
ACCEPTED_PARENT_SEALS={
    '47':'bbd8bd860ccc5e27cb61251e76d1138abace0efe66463acecb7e003cb86ff1b4',
    '54':'5acc32c577df7e016e0c589a7b95dc710a2a4147ae599fd7b97fed1129532c86',
    '283':'c1d03df3d4aeda65d2869c2cb0de3bb42a8695a15fa3e5a9397fc67c69b1816a',
    '337':'7aae57ec2c0f8815c7ef8ef7284276694e31f7a855e452f8e22af4f460d22194',
}


def body(m,r):
    keys=sorted(k for k,v in m['source_key_owners'].items() if v==r['key'])
    records=[x['record'] for x in m['records'] if set(keys)&set(x['key_owners'])]
    for key,contract in m.get('additional_source_contracts',{}).items():
        if contract['owner']==r['key']:
            keys.append(key);records.append(contract)
    if r['key'].startswith('level-'):
        level=int(r['key'][6:])
        records=[x['record'] for x in m['records'] if (x['record']['kind']=='level' and x['record']['level']==level) or x['record']['id'].endswith(('resource/fighter/second-wind','resource/fighter/action-surge','resource/fighter/indomitable','choice/fighter/weapon-mastery','choice/fighter/fighting-style'))]
    return (f"<!-- edda-325:{r['key']} -->\nParent #{r['parent']}; decision #325; roadmap #47. One PR. Task Size {r['project_size']}.\n\n"
            f"Required native path: this terminal -> #{m['qualification_gate']['issue']} -> #337.\n\n"
            f"Pinned SRD5.2.1 PDF47–49 and explicitly cited consumed rules. SHA256 {PIN}. CC-BY-4.0 attribution/adaptation: docs/srd-5.2.1-conformance.md. Fixture/ownership decisions are Edda additions.\n\n"
            f"Scope and deterministic positive/negative fixtures: {r['fixture']}\n\n{m['common_acceptance']}\n\n"
            "Ownership: execution and admission are separate contracts. #427/#428 completed rests under272; #429 reusable XP/PB/CON under336 (Barbarian adapter remains D12/fixed7); Fighter D10/fixed6 separately owned here. #430 under333 provenance; #333 origins/#335 multiclass retain policy. #384 durable choices/#389 capped ability application; #385–401 exact feat effects; #487 Incapacitated/Unconscious, #488 Charmed/Frightened/Prone; #489–496 eight separate masteries under278; #58 sheet. #273 casting modes is not sheet ownership. #351–353 and #382 retain exact effect/concentration/Cleric eligibility paths; this Fighter decision grants none of their access. No generic spell implementation.\n\n"
            f"Exact native prerequisites: {', '.join('#'+str(n) for n in r['blocked_by_issues'])}.\n\n"
            'Owned mechanic keys:\n'+'\n'.join('- '+k for k in keys)+
            '\n\nContext records (not duplicate mechanic ownership):\n```json\n'+json.dumps(records,indent=2)+'\n```\n')


def pages(path,call_cap=100):
    result=[];seen=set()
    for page in range(1,call_cap+1):
        batch=api(f'{path}?per_page=100&page={page}')
        ids=[x['id'] for x in batch]
        assert len(ids)==len(set(ids)) and not seen.intersection(ids),'repeated API page'
        seen.update(ids);result+=batch
        if len(batch)<100:return result
    raise AssertionError('pagination call cap')


def graph_read():
    graph, cursor, seen = [], None, set()
    cursors=set()
    for _ in range(100):
        q = '''query($cursor:String){repository(owner:"subculture-collective",name:"edda"){
        issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{number title body state parent{number}
        blockedBy(first:100){pageInfo{hasNextPage} nodes{number}}}}}}'''
        data = api('graphql', {'query': q, 'variables': {'cursor': cursor}})
        assert not data.get('errors'), data.get('errors')
        batch = data['data']['repository']['issues']
        for node in batch['nodes']:
            assert node['number'] not in seen, 'repeated issue page'
            seen.add(node['number'])
            graph.append(dict(number=node['number'], parent=(node['parent'] or {}).get('number'),
                              blocked_by=sorted(x['number'] for x in (pages(f"{REPO}/issues/{node['number']}/dependencies/blocked_by") if node['blockedBy']['pageInfo']['hasNextPage'] else node['blockedBy']['nodes'])),
                              title=node['title'], body=node['body'], state=node['state'].lower()))
        if not batch['pageInfo']['hasNextPage']:
            return graph
        assert batch['pageInfo']['endCursor'] and batch['pageInfo']['endCursor'] not in cursors, 'repeated cursor'
        cursors.add(batch['pageInfo']['endCursor'])
        cursor = batch['pageInfo']['endCursor']
    raise AssertionError('graph call cap')


def live(m):
    global _api_calls
    _api_calls=0
    receipt = {'observed_at': datetime.now(timezone.utc).isoformat(), 'graph': graph_read(), 'project': {}}
    owned = m['issues']+[m['qualification_gate']]+m['subordinate_gates']
    owned += [{'issue':int(n)} for n in m['consumed_contracts']]
    for start in range(0, len(owned), 30):
        fields = [f'i{r["issue"]}:issue(number:{r["issue"]})'+'{issueType{name} projectItems(first:100){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}' for r in owned[start:start+30]]
        data = api('graphql', {'query':'{repository(owner:"subculture-collective",name:"edda"){'+' '.join(fields)+'}}'})
        assert not data.get('errors'), data.get('errors')
        for r in owned[start:start+30]:
            node = data['data']['repository'][f'i{r["issue"]}']
            assert not node['projectItems']['pageInfo']['hasNextPage']
            items = [x for x in node['projectItems']['nodes'] if x['project']['id']=='PVT_kwDODjtmGM4BrFJ2']
            assert len(items)==1
            receipt['project'][str(r['issue'])]={'type':(node['issueType'] or {}).get('name'),'size':(items[0]['fieldValueByName'] or {}).get('name')}
    return receipt


def validate(m,r,sealed=False):
    sealed=sealed or m.get("supplement_publication")=="sealed"
    assert m['source']['sha256']==PIN
    assert m['accepted_parent_baseline']['accepted_druid_merge']==ACCEPTED_DRUID_MERGE,'accepted merge drift'
    assert m['accepted_parent_baseline']['body_sha256']==ACCEPTED_PARENT_SEALS,'accepted seal drift'
    for n,h in ACCEPTED_PARENT_SEALS.items():
        assert digest(m['protected'][n]['body'])==h,'accepted prefix drift'
    records=[]
    for p,c in m['catalogues'].items():
        assert json.loads((ROOT/p).read_text())==c and c['source']['sha256']==PIN,'source drift'
        records+=c['records']
    assert len(records)==67 and records==[x['record'] for x in m['records']]
    assert [x['level'] for x in records if x['kind']=='level']==list(range(1,21))
    keys={k for x in records for k in x['required_mechanics']+x['option_keys']}
    assert set(m['source_key_owners'])==keys and len(keys)==134,'source keys'
    assert digest(json.dumps(m['source_key_owners'],sort_keys=True))=='7d03be396a3a7c451162f0c647edf8f5413a16dd5ffbf39ed7e4b83b5a87facb','mechanic ownership drift'
    rows={x['key']:x for x in m['issues']}
    assert len(rows)==len(m['issues'])==45 and len({x['issue'] for x in rows.values()})==45
    assert set(m['reused_terminals'])==set([58,333,335,336,338,339,384]+list(range(385,402))+[427,428,429,430,487,488]+list(range(489,497))),'reused owner drift'
    assert digest(json.dumps({x['key']:{k:x[k] for k in ['title','fixture','prerequisites','depends']} for x in m['issues']},sort_keys=True))=='fee2f7016b04e0c4830b9ed85f888d98e5a0681d21ab6a5b408e97d09a9b300f','audited terminal scope drift'
    operands={'d10': ['D10', 'fixed6', 'min1', '10+CON', '429'], 'equipment': ['8 Javelins', '4GP', '20 Arrows', '11GP', 'C155GP'], 'second-wind': ['1d10+Fighter level', 'Bonus Action', 'ONE', 'ALL', '2 at1,3 at4,4 at10'], 'surge': ['EXCEPT Magic', 'one use2–16,two17–20', 'once per turn'], 'mind': ['1d10', 'only if new total succeeds', 'unknown outcome'], 'improvement': ['4,6,8,12,14,16', 'max20', 'Skilled', 'Spellcasting', 'STR13 OR DEX13'], 'attacks': ['2 attacks atFighter5,3 at11,4 at20', 'do not stack'], 'shift': ['half current Speed', 'Second Wind', 'Not Tactical Mind'], 'indomitable': ['MUST use new roll', '1 at9,2 at13,3 at17', 'Long Rest only'], 'critical': ['19–20', 'at15 threshold18–20', 'Unarmed Strikes'], 'defy': ['18–20', 'regain1HP'], 'rally': ['5+CON', 'HP<=half max', 'at least1HP'], 'tactical-master': ['Push,Sap,or Slow', 'FOR THAT ATTACK'], 'mastery': ['Three', 'four4,five10,six16', 'ONE'], 'studied': ['NEXT attack roll', 'THAT creature', 'end of your next turn'], 'heroic': ['SELF', 'starting own turn WITHOUT it']}
    for key,values in operands.items():
        assert all(v in rows[key]['fixture'] for v in values),f'source operand {key}'
    assert set(m['source_key_owners'].values())<=set(rows)|{'external:389','external:492','external:493','external:494'}
    for x in m['records']:
        assert x['key_owners']=={k:m['source_key_owners'][k] for k in x['record']['required_mechanics']+x['record']['option_keys']}
        assert x['state']=='metadata only; broad Fighter/Champion unqualified'
        assert set(x['support_evidence'])=={'engine','web','tui','qualification'} and all(x['support_evidence'].values())
    for p,h in m['evidence_hashes'].items():assert hashlib.sha256((ROOT/p).read_bytes()).hexdigest()==h,'evidence drift'
    graph={x['number']:x for x in r['graph']}
    assert len(graph)==len(r['graph'])
    shared=rows['heroic-inspiration']
    assert shared['prerequisites']==[61,64,65] and shared['depends']==[]
    assert m['additional_source_contracts']=={'srd-5.2.1:mechanic/rules-glossary/heroic-inspiration':{'owner':'heroic-inspiration','issue':shared['issue'],'pdf_page':183,'source_sha256':PIN,'prerequisites':[61,64,65],'parent':50,'reservation':'Parent confirmed native50, exact61/64/65, before publication'}}
    for target in [279,280]:assert shared['issue'] in graph[target]['blocked_by'],'shared qualification edge'
    assert 'heroic-inspiration' in rows['heroic']['depends']
    for x in graph.values():assert len(x['blocked_by'])==len(set(x['blocked_by']))<=100,'duplicate/cap'
    acyclic(graph)
    acyclic({n:{'blocked_by':[x['parent']] if x['parent'] else []} for n,x in graph.items()})
    counts={}
    for x in graph.values():counts[x['parent']]=counts.get(x['parent'],0)+1
    assert all(v<=100 for n,v in counts.items() if n is not None)
    assert graph[325]['parent']==283 and graph[325]['state'] in ['open','closed']
    assert graph[325]['blocked_by']==[285,290,312]
    assert all(graph[n]['state']=='closed' for n in [285,290,312,338,339])
    g=m['qualification_gate'];gate=g['issue']
    required={325}|{x['issue'] for x in rows.values()}|set(m['reused_terminals'])
    assert len(required)==84 and set(g['blocked_by_issues'])==required==set(graph[gate]['blocked_by']),'terminal gate edge'
    assert graph[gate]['parent']==325 and graph[gate]['title']==g['title'] and graph[gate]['body']==g['body']
    listed=[int(n) for n in re.findall(r'#(\d+)',g['body'].split('Mandatory native blockers (84): ')[1].split('.\n\n')[0])]
    assert listed==g['blocked_by_issues'],'gate body inventory'
    assert gate in graph[337]['blocked_by'],'gate to337'
    for x in rows.values():
        n=x['issue'];actual=graph[n]
        assert actual['title']==x['title'] and actual['body']==body(m,x),'terminal body'
        assert actual['parent']==x['parent']==(50 if x['key']=='heroic-inspiration' else gate),'terminal parent'
        expected=set(x['prerequisites'])|{rows[k]['issue'] for k in x['depends']}
        assert set(x['blocked_by_issues'])==set(actual['blocked_by'])==expected,'prerequisite'
    for x in list(rows.values())+[g]:assert r['project'][str(x['issue'])]=={'type':'Task','size':x['project_size']},'Task/Size'
    for n,old in {**m['protected'],**m['consumed_contracts']}.items():
        actual=graph[int(n)]
        assert actual['parent']==old['parent'] and actual['title']==old['title'],'legacy metadata'
        assert set(old['blocked_by'])<=set(actual['blocked_by']),'legacy edge'
        if int(n) not in [47,54,283,337]:
            assert actual['body']==old['body'] and actual['blocked_by']==old['blocked_by'],'consumed body/edge'
        else:
            s=next(x for x in m['supplements'] if x['issue']==int(n));prefix=old['body'];section=s['append']
            assert digest(prefix)==s['previous_sha256']
            if sealed:
                assert actual['body'].startswith(prefix+section) and actual['body'].count(section)==1,'sealed body'
                assert actual['body'].count('<!-- edda-325:start -->')==actual['body'].count('<!-- edda-325:end -->')==1,'duplicate section'
            else:
                assert actual['body']==prefix,'pending baseline drift'
    for n,expected in m['consumed_project'].items():assert r['project'][n]==expected,'consumed Task/Size'
    assert len(m['legacy_root_edges'])==45 and set(m['legacy_root_edges'])<=set(graph[337]['blocked_by'])
    for n in [498,501,551,630]:assert n in graph[337]['blocked_by'],'class gate'
    assert 628 in graph[551]['blocked_by'],'Cleric subordinate gate'
    for n in range(1,21):
        x=rows['level-'+str(n)]
        assert {'d10','style','mastery'}<=set(x['depends'])
        if n>1:assert 'level-'+str(n-1) in x['depends']
        if n in [4,6,8,12,14,16]:assert 'improvement' in x['depends']
    assert {492,493,494}<=set(rows['tactical-master']['prerequisites'])
    assert {335,430}<=set(rows['multiclass']['prerequisites']) and 'd10' in rows['multiclass']['depends']
    assert {336,429,58}<=set(rows['d10']['prerequisites'])
    assert set(range(385,402))<=set(rows['improvement']['prerequisites'])
    assert len(next(x['append'] for x in m['supplements'] if x['issue']==47))<=200
    for supplement in m['supplements']:
        n=str(supplement['issue']);prefix=m['protected'][n]['body'];append=supplement['append']
        assert supplement['previous_sha256']==ACCEPTED_PARENT_SEALS[n]
        assert supplement['previous_chars']==len(prefix) and supplement['append_chars']==len(append),'capacity metadata drift'
        assert supplement['result_chars']==len(prefix)+len(append)<=supplement['maximum_chars']==65536,'body capacity'
    assert json.loads((ROOT/'docs/investigations/issue-325-parent-supplements.json').read_text())==m['supplements'],'supplement file drift'
    return len(keys),len(graph)


def probes(m,r,sealed=False):
    sealed=sealed or m.get("supplement_publication")=="sealed"
    validate(m,r,sealed)
    def reject(mm,rr):
        try:validate(mm,rr,sealed)
        except (AssertionError,KeyError):return
        raise AssertionError('accepted corruption')
    mm=deepcopy(m);mm['source']['sha256']='bad';reject(mm,r)
    mm=deepcopy(m);next(iter(mm['catalogues'].values()))['records'][0]['name']='bad';reject(mm,r)
    mm=deepcopy(m);mm['source_key_owners'].pop(next(iter(mm['source_key_owners'])));reject(mm,r)
    mm=deepcopy(m);mm['records'][0]['key_owners']={};reject(mm,r)
    for n in m['qualification_gate']['blocked_by_issues']:
        rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==m['qualification_gate']['issue'])['blocked_by'].remove(n);reject(m,rr)
    for x in m['issues']:
        for field,val in [('body','changed'),('title','changed'),('parent',283),('blocked_by',[337])]:
            rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==x['issue'])[field]=val;reject(m,rr)
    for target in [279,280]:
        rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==target)['blocked_by'].remove(next(x['issue'] for x in m['issues'] if x['key']=='heroic-inspiration'));reject(m,rr)
    for target,edge in [(337,m['qualification_gate']['issue']),(334,298)]:
        rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==target)['blocked_by'].remove(edge);reject(m,rr)
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==283)['parent']=m['qualification_gate']['issue'];reject(m,rr)
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==m['issues'][0]['issue'])['blocked_by'].append(337);reject(m,rr)
    rr=deepcopy(r);rr['project'][str(m['issues'][0]['issue'])]['size']='XL';reject(m,rr)
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==m['qualification_gate']['issue'])['blocked_by'].append(325);reject(m,rr)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--live',action='store_true');p.add_argument('--sealed','--require-supplements',dest='sealed',action='store_true');p.add_argument('--pdf',type=Path);a=p.parse_args()
    m=json.loads(MANIFEST.read_text())
    if a.pdf:
        assert hashlib.sha256(a.pdf.read_bytes()).hexdigest()==PIN
        glossary=subprocess.run(['rtk','proxy','pdftotext','-f','183','-l','183','-x','0','-y','0','-W','310','-H','792','-layout',str(a.pdf),'-'],check=True,capture_output=True,text=True).stdout
        flat=' '.join(re.sub(r'-\s+','',glossary).split())
        for anchor in ['you (a player character) have Heroic Inspiration','reroll any die immediately after rolling it, and you must use the new roll','already have it, it’s lost unless you give it to a player character who lacks it']:
            assert anchor in flat,'Heroic Inspiration source operand'
        for name in ['fighter','champion','advancement','multiclass','origin_feats','general_feats','fighting_styles','epic_boon_feats']:
            subprocess.run(['rtk','proxy','python3',str(ROOT/f'scripts/verify_srd_{name}.py'),str(a.pdf)],check=True)
    r=live(m) if a.live else json.loads(RECEIPT.read_text())
    sealed=a.sealed or m.get('supplement_publication')=='sealed'
    keys,nodes=validate(m,r,sealed)
    probes(m,r,sealed)
    if a.live:RECEIPT.write_text(json.dumps(r,indent=2)+'\n')
    print(f'PASS:67 records/20 rows/{keys} keys;45 terminals (44 Fighter +1 shared);84 gate blockers;{nodes} nodes;exact bodies/parents/Task sizes;separate DAGs; supplements '+('sealed' if sealed else 'PENDING parent publication'))


if __name__=='__main__':main()
