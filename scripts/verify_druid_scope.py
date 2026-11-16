#!/usr/bin/env python3
"""Verify #324 source, owner matrix and native hierarchical qualification paths."""
import argparse
from copy import deepcopy
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import subprocess
from urllib.parse import parse_qs, parse_qsl, urlencode, urlsplit

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / 'docs/investigations/issue-324-druid-scope.json'
RECEIPT = ROOT / 'docs/investigations/issue-324-link-verification.json'
PDF_SHA = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
PROJECT = 'PVT_kwDODjtmGM4BrFJ2'


def digest(value):
    return hashlib.sha256(value.encode() if isinstance(value, str) else value).hexdigest()


def body(m, r):
    path = m['qualification_paths'][str(r['issue'])]
    return (f"<!-- edda-324:{r['key']} -->\nParent: #{r['parent']}; decision #324; roadmap #47. Task Size {r['project_size']}.\n\n"
            + 'One reviewable PR. Required native path: ' + ' -> '.join('#'+str(n) for n in path) + '.\n\n'
            + f"Source: pinned English SRD5.2.1 PDF41-46 and exact referenced clauses below; SHA256 {PDF_SHA}. "
            + 'CC-BY-4.0 attribution and adaptation: docs/srd-5.2.1-conformance.md. Fixture boundaries and ownership are Edda additions.\n\n'
            + f"Scope and deterministic positive fixture: {r['fixture']}\n\n"
            + 'Negative fixtures: violate each stated source/level/choice/target/resource/action/timing/numeric predicate independently; assert no mutation. Unknown variants remain unavailable.\n\n'
            + m['common_acceptance'] + '\n\n'
            + 'Feature behavior uses explicit isolated entitlements. Entry and level applications consume implemented contracts and apply grants once. Druid remains unavailable until independent qualification. '
            + '#58 sheet projection; #427/#428 completed rests; #429 shared XP/PB/CON transaction (Druid adapter distinct); #430 provenance; #503 class list membership distinct from spell execution. '
            + '#384/#389 selection/caps and all exact #334 effect children required. #181/#184/#185 retain familiar/creature lifecycle ownership and #183 world qualification. '
            + 'No runtime/grants/coverage/Full Rules promotion, inference or deployment in this decision.\n\n'
            + 'Native prerequisites: '+', '.join('#'+str(n) for n in r['blocked_by_issues'])+'.\n\n'
            + 'Owned mechanic keys:\n'+'\n'.join('- '+k for k in r['owned_keys'])
            + '\n\nExact contextual source records (references do not duplicate shared execution):\n```json\n'+json.dumps(r['source_records'],indent=2)+'\n```\n'
            + ('\nExact licensed source operands; implement only the bounded clause named above:\n\n'+r['source_excerpt']+'\n' if r.get('source_excerpt') else ''))


class Reader:
    def __init__(self):
        self.calls = 0

    def __call__(self, path, data=None):
        self.calls += 1
        assert self.calls <= 350, 'read-only API call cap exceeded'
        cmd = ['gh', 'api', path]
        if data is not None:
            assert path == 'graphql' and 'mutation' not in data['query'], 'read-only verifier'
            cmd += ['--input', '-']
        p = subprocess.run(cmd, input=json.dumps(data) if data is not None else None,
                           text=True, capture_output=True, check=True)
        value = json.loads(p.stdout)
        assert not (isinstance(value, dict) and value.get('errors')), value
        return value


def pages(api, endpoint):
    parts = urlsplit(endpoint)
    params = dict(parse_qsl(parts.query))
    assert 'page' not in params and 'per_page' not in params
    seen=set()
    for page in range(1, 21):
        batch = api(parts.path + '?' + urlencode({**params, 'per_page': 100, 'page': page}))
        assert isinstance(batch, list)
        ids=[x['id'] for x in batch]
        assert len(ids)==len(set(ids)) and not seen.intersection(ids), 'repeated API page'
        seen.update(ids)
        yield from batch
        if len(batch) < 100:
            return
    raise AssertionError('pagination cap exceeded')


def parent_path(graph, number):
    path = []
    while number is not None:
        assert number not in path, 'parent cycle'
        assert number in graph, 'missing parent graph node'
        path.append(number)
        number = graph[number]['parent']
    return path


def live(m):
    api = Reader()
    graph, cursor, seen = [], None, set()
    for _ in range(20):
        q = '''query($cursor:String){repository(owner:"subculture-collective",name:"edda"){
        issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{
        number title body state parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}}}}}'''
        b = api('graphql', {'query': q, 'variables': {'cursor': cursor}})['data']['repository']['issues']
        for node in b['nodes']:
            assert node['number'] not in seen, 'repeated issue page'
            seen.add(node['number'])
            assert not node['blockedBy']['pageInfo']['hasNextPage'], 'blocker cap exceeded'
            graph.append(dict(number=node['number'],title=node['title'],body=node['body'] or '',state=node['state'].lower(),parent=(node['parent'] or {}).get('number'),blocked_by=sorted(x['number'] for x in node['blockedBy']['nodes'])))
        if not b['pageInfo']['hasNextPage']:break
        assert b['pageInfo']['endCursor'] != cursor, 'repeated cursor'
        cursor=b['pageInfo']['endCursor']
    else:raise AssertionError('graph page cap')
    project={}
    nums=sorted({r['issue'] for r in m['issues']+m['gates']}|set(m['reused_terminals']))
    for start in range(0,len(nums),30):
        batch=nums[start:start+30]
        fields=' '.join('i%d:issue(number:%d){issueType{name} projectItems(first:100){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}'%(n,n) for n in batch)
        data=api('graphql',{'query':'{repository(owner:"subculture-collective",name:"edda"){'+fields+'}}'})['data']['repository']
        for n in batch:
            node=data['i'+str(n)];assert not node['projectItems']['pageInfo']['hasNextPage']
            items=[x for x in node['projectItems']['nodes'] if x['project']['id']==PROJECT];assert len(items)==1
            project[str(n)]=dict(type=node['issueType']['name'],size=items[0]['fieldValueByName']['name'])
    indexed={x['number']:x for x in graph}
    old=json.loads(RECEIPT.read_text()) if RECEIPT.exists() else {}
    for n,seal in old.get('body_seals',{}).items():
        assert indexed[int(n)]['body'].startswith(seal['prefix']), 'sealed parent prefix rollback'
    seals={str(x['issue']):dict(prefix=indexed[x['issue']]['body'],sha256=digest(indexed[x['issue']]['body'])) for x in m['supplements']}
    return dict(observed_at=datetime.now(timezone.utc).isoformat(),graph=graph,project=project,api_calls=api.calls,body_seals=seals)


def validate(m, r):
    from verify_barbarian_scope import acyclic
    assert m['source']['sha256']==PDF_SHA, 'source pin'
    source=[]
    for path,c in m['catalogues'].items():
        assert json.loads((ROOT/path).read_text())==c and c['source']['sha256']==PDF_SHA, 'canonical source drift'
        if path.endswith(('character-druid.json','character-circle-of-the-land.json')):source+=c['records']
    assert len(source)==103 and source==[x['record'] for x in m['records']], 'record identities or values'
    assert [x['level'] for x in source if x['kind']=='level']==list(range(1,21))
    keys={k for x in source for k in x['required_mechanics']+x['option_keys']}
    assert keys==set(m['source_key_owners']) and len(keys)==261, 'source keys/options'
    rows={x['key']:x for x in m['issues']};assert len(rows)==len(m['issues'])==95
    assert set(m['source_key_owners'].values())<=set(rows)|{'external:384','external:503'}
    for x in m['records']:
        assert x['key_owners']=={k:m['source_key_owners'][k] for k in x['record']['required_mechanics']+x['record']['option_keys']}
        assert x['state']=='metadata only; Druid unavailable'
        assert set(x['support_evidence'])=={'engine','web','tui','qualification'} and all(x['support_evidence'].values())
    for path,h in m['evidence_hashes'].items():assert digest((ROOT/path).read_bytes())==h, 'engine/evidence drift'
    for x in m['issues']:
        assert x['owned_keys']==sorted(k for k,v in m['source_key_owners'].items() if v==x['key'])
        if not x['key'].startswith('level-'):
            assert x['source_records']==[y['record'] for y in m['records'] if any(k in x['owned_keys'] for k in y['key_owners'])] or (not x['owned_keys'] and x['key'].startswith('spell-')), 'context record drift'
        if x.get('source_excerpt'):assert digest(x['source_excerpt'])==m['source_excerpt_sha256'][x['key']], 'spell operands'
    feats=json.loads((ROOT/'docs/investigations/issue-334-feat-scope.json').read_text())
    featrows={x['key']:x['issue'] for x in feats['issues']}
    expected=[dict(id=f['id'],category=f['category'],effect_issue=featrows[f['effect_owner']],source_records=f['records'],eligibility='Only with the actual source-linked Druid grant and exact #384 prerequisites; unsupported origin/multiclass/feat variant remains unavailable.') for f in feats['feats']]
    assert m['feat_effect_choices']==expected, 'feat source/caps/effects'
    for k in ['asi','boon']:assert set(range(384,402))<=set(rows[k]['prerequisites'])
    graph={x['number']:x for x in r['graph']};assert len(graph)==len(r['graph'])
    acyclic(graph)
    acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] else []) for n,x in graph.items()})
    assert all(len(x['blocked_by'])<=100 for x in graph.values()), 'blocker capacity'
    counts={}
    for x in graph.values():counts[x['parent']]=counts.get(x['parent'],0)+1
    assert all(v<=100 for k,v in counts.items() if k is not None), 'native child capacity'
    assert graph[324]['parent']==283 and {285,289,311}<=set(graph[324]['blocked_by']), 'source prerequisite edges'
    assert all(graph[n]['state']=='closed' for n in [285,289,311])
    gate=m['gates'][0]['issue']; assert graph[gate]['parent']==324 and gate in graph[337]['blocked_by']
    assert len(m['gates'])==3
    terminals={x['issue'] for x in m['issues']}|set(m['reused_terminals'])
    assert len(terminals)==141, 'mandatory leaf count'
    assert set(m['qualification_paths'])=={str(n) for n in terminals}, 'missing leaf path'
    for n in terminals:
        path=m['qualification_paths'][str(n)]
        assert path[0]==n and path[-2:]==[gate,337] and len(path)==4
        assert path[1] in {x['issue'] for x in m['gates'][1:]}
        for a,b in zip(path,path[1:]):assert a in graph[b]['blocked_by'], 'terminal qualification edge'
    feature={x['issue'] for x in m['issues'] if x['parent'] not in [191,193,270]}
    shared=terminals-feature
    assert set(m['gates'][1]['blocked_by_issues'])==feature
    assert set(m['gates'][2]['blocked_by_issues'])==shared
    assert set(m['gates'][0]['blocked_by_issues'])=={324,*[x['issue'] for x in m['gates'][1:]],*m['qualification_dependencies']}
    for x in m['issues']+m['gates']:
        n=x['issue']; actual=graph[n]
        assert actual['parent']==x['parent'] and actual['title']==x['title'], 'body/parent/title'
        assert actual['body']==(x['body'] if x in m['gates'] else body(m,x)), 'body drift'
        assert set(actual['blocked_by'])==set(x['blocked_by_issues']), 'native prerequisite drift'
        assert r['project'][str(n)]==dict(type='Task',size=x['project_size']), 'Task/Project Size'
        if x in m['issues']:
            assert set(x['blocked_by_issues'])==set(x['prerequisites'])|{rows[k]['issue'] for k in x['depends']}
            if x['parent'] in [191,193,270]:assert all(n in graph[g]['blocked_by'] for g in [279,280]), 'cross-domain gate'
    for n,old in m['preserved'].items():
        actual=graph[int(n)]
        assert actual['parent']==old['parent'] and actual['title']==old['title'], 'preserved parent/title'
        assert set(old['blocked_by'])<=set(actual['blocked_by']), 'historical prerequisite weakened'
        supplement=next((s['append'] for s in m['supplements'] if s['issue']==int(n)),None)
        if supplement:
            seal=r['body_seals'][n]
            assert digest(seal['prefix'])==seal['sha256'] and seal['prefix'].startswith(old['body']), 'parent prefix seal'
            assert actual['body'].startswith(seal['prefix']), 'preserved parent prefix rollback'
            assert actual['body'].count(supplement)<=1, 'duplicate supplement'
        else:
            assert actual['body']==old['body'], 'preserved body drift'
        if int(n) not in [337,279,280]:assert actual['blocked_by']==old['blocked_by']
    assert set(m['preserved_direct_337'])<=set(graph[337]['blocked_by']) and 334 in graph[337]['blocked_by']
    assert {298,299,300,301}<=set(graph[334]['blocked_by'])
    for n,metadata in m['consumed_project'].items():assert r['project'][n]==metadata, 'consumed metadata drift'
    for level in range(1,21):
        x=rows['level-'+str(level)]
        assert {'d8','preparation','slots'}<=set(x['depends'])
        if level>1:assert 'level-'+str(level-1) in x['depends']
        canonical=[y for y in source if y['kind']=='level' and y['level']==level]
        canonical += [y for y in m['catalogues']['internal/rules/srd521/testdata/character-advancement.json']['records'] if y['level']==level]
        assert x['source_records']==canonical, 'exact class/global level row'
    for level in [4,8,12,16]:assert 'asi' in rows['level-'+str(level)]['depends']
    for x in m['supplements']:
        assert (ROOT/x['path']).read_text()==x['append'], 'supplement artifact drift'
        assert x['previous_sha256']==digest(m['preserved'][str(x['issue'])]['body']), 'supplement previous prefix hash'
        assert len(m['preserved'][str(x['issue'])]['body']+x['append'])<=65536, 'body length cap'
    assert len(next(x['append'] for x in m['supplements'] if x['issue']==47))<=1000
    return len(terminals),len(keys),len(graph)


def probes(m,r):
    def rejects(label,fn):
        mm,rr=deepcopy(m),deepcopy(r);fn(mm,rr)
        try:validate(mm,rr)
        except (AssertionError,KeyError):return
        raise AssertionError('accepted corruption '+label)
    tests=[('pin',lambda a,b:a['source'].update(sha256='bad')),
           ('record',lambda a,b:next(iter(a['catalogues'].values()))['records'][0].update(name='bad')),
           ('keys',lambda a,b:a['source_key_owners'].pop(next(iter(a['source_key_owners'])))),
           ('operand',lambda a,b:next(iter(a['catalogues'].values()))['records'][0]['parameters'].update(bad=1)),
           ('spell text',lambda a,b:next(x for x in a['issues'] if x.get('source_excerpt')).update(source_excerpt='bad')),
           ('metadata',lambda a,b:b['project'][str(a['gates'][0]['issue'])].update(type='Feature'))]
    for target in [324,334,337,279,280]+[x['issue'] for x in m['issues']+m['gates']]:
        for edge in next(x for x in r['graph'] if x['number']==target)['blocked_by']:
            tests.append((f'edge{target}/{edge}',lambda a,b,n=target,e=edge:next(x for x in b['graph'] if x['number']==n)['blocked_by'].remove(e)))
    n=m['issues'][0]['issue'];g=m['gates'][0]['issue']
    tests += [('body',lambda a,b:next(x for x in b['graph'] if x['number']==n).update(body='bad')),
              ('parent',lambda a,b:next(x for x in b['graph'] if x['number']==n).update(parent=283)),
              ('dependency cycle',lambda a,b:next(x for x in b['graph'] if x['number']==n)['blocked_by'].append(g)),
              ('containment cycle',lambda a,b:next(x for x in b['graph'] if x['number']==324).update(parent=g))]
    for label,fn in tests:rejects(label,fn)
    calls=[]
    def synthetic(url):
        q=parse_qs(urlsplit(url).query);assert q['state']==['all'] and q['per_page']==['100']
        page=int(q['page'][0]);calls.append(page)
        return [dict(id=page*100+i) for i in range(100 if page<10 else 1)]
    assert len(list(pages(synthetic,'repos/o/r/issues?state=all')))==901 and calls==list(range(1,11))
    try:list(pages(lambda url:[dict(id=i) for i in range(100)],'repos/o/r/issues'))
    except AssertionError as e:assert str(e)=='repeated API page'
    else:raise AssertionError('repeated page accepted')
    try:list(pages(lambda url:[dict(id=int(parse_qs(urlsplit(url).query)['page'][0])*100+i) for i in range(100)],'repos/o/r/issues'))
    except AssertionError as e:assert str(e)=='pagination cap exceeded'
    else:raise AssertionError('unbounded page accepted')
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==324)['state']='closed';validate(m,rr)
    for supplement in m['supplements']:
        rr=deepcopy(r);x=next(x for x in rr['graph'] if x['number']==supplement['issue']);x['body']+='\nLater authorized append-only note.\n';validate(m,rr)
        # A published seal must reject original-only rollback even without the require flag.
        rr=deepcopy(r);n=str(supplement['issue']);prefix=m['preserved'][n]['body']+supplement['append'];rr['body_seals'][n]=dict(prefix=prefix,sha256=digest(prefix))
        x=next(x for x in rr['graph'] if str(x['number'])==n);x['body']=m['preserved'][n]['body']
        try:validate(m,rr)
        except AssertionError:pass
        else:raise AssertionError('published parent seal downgraded')
    print(f'PASS {len(tests)} source/operand/body/metadata/edge/cycle corruption probes; exact query pages/repetition/cap')


def source_check(m,pdf):
    assert digest(pdf.read_bytes())==PDF_SHA
    for script in ['verify_srd_druid.py','verify_srd_circle_of_the_land.py','verify_srd_advancement.py','verify_srd_multiclass.py']:
        subprocess.run(['python3',str(ROOT/'scripts'/script),str(pdf)],check=True)
    import re
    raw=subprocess.run(['pdftotext','-f','107','-l','175','-raw',str(pdf),'-'],check=True,text=True,capture_output=True).stdout
    headings=list(re.finditer(r'^([^\n]+)\n(?:Level [0-9] |[A-Za-z]+ Cantrip)',raw,re.M))
    for row in m['issues']:
        if row['key'].startswith('spell-') and row.get('source_excerpt'):
            name=row['source_excerpt'].splitlines()[0].lower()
            found=[(i,x) for i,x in enumerate(headings) if x.group(1).lower()==name];assert len(found)==1
            i,x=found[0];text=raw[x.start():headings[i+1].start() if i+1<len(headings) else len(raw)]
            text=re.sub(r'\f?System Reference Document 5.2.1\n\d+\n','',text).strip()
            if name=='fireball':text=text.split('Otherworldly Steed')[0].strip()
            assert text==row['source_excerpt'], 'exact spell source operands '+name

    for first,last,clauses in [
        (184,184,['Inactive.', 'No Concentration.', 'Speechless.', 'Surprised.']),
        (186,186,['Paralyzed [Condition]','Poisoned [Condition]','Speed 0.','Strength and Dexterity saving throws.','within 5 feet of','Disadvantage on attack rolls and ability checks.']),
        (187,187,['Restrained [Condition]','Disadvantage on Dexterity saving throws.']),
        (105,106,['One Spell with a Spell Slot per Turn','doesn’t specify a cost','must hold it']),
        (185,185,['Interrupting the Rest.','6 hours','16 hours']),
        (187,187,['Short Rest','1-hour','An interrupted Short Rest confers no benefits.'])]:
        text=subprocess.run(['pdftotext','-f',str(first),'-l',str(last),'-raw',str(pdf),'-'],check=True,text=True,capture_output=True).stdout
        flat=' '.join(text.replace('\u00ad','').split())
        for clause in clauses:assert clause in flat, f'PDF{first} clause missing {clause}'


def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--live',action='store_true');p.add_argument('--probes',action='store_true');p.add_argument('--pdf',type=Path);p.add_argument('--require-supplements',action='store_true');a=p.parse_args()
    m=json.loads(MANIFEST.read_text())
    if a.pdf:source_check(m,a.pdf)
    r=live(m) if a.live else json.loads(RECEIPT.read_text())
    counts=validate(m,r)
    if a.probes:probes(m,r)
    graph={x['number']:x for x in r['graph']}
    statuses={s['issue']:'published' if s['append'] in graph[s['issue']]['body'] else 'parent-body pending' for s in m['supplements']}
    if a.require_supplements:assert all(v=='published' for v in statuses.values()), 'parent-body publication pending'
    if a.live:RECEIPT.write_text(json.dumps(r,indent=2)+'\n')
    print(f'PASS {counts[0]} mandatory leaves, {counts[1]} exact keys, {counts[2]} nodes; separate dependency/containment DAGs; supplements {statuses}')


if __name__=='__main__':main()
