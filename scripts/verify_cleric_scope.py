#!/usr/bin/env python3
"""Check #323 source ownership and exact native publication; --live is read only."""
import argparse
from copy import deepcopy
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import subprocess

from verify_barbarian_scope import acyclic, api, digest

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / 'docs/investigations/issue-323-cleric-scope.json'
RECEIPT = ROOT / 'docs/investigations/issue-323-link-verification.json'
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
REPO = 'repos/subculture-collective/edda'


def body(m, r):
    keys = sorted(k for k, v in m['source_key_owners'].items() if v == r['key'])
    records = [x['record'] for x in m['records'] if any(k in keys for k in x['key_owners'])]
    if r['key'].startswith('level-'):
        records = [x['record'] for x in m['records'] if x['record']['kind'] == 'level' and x['record']['level'] == int(r['key'][6:])]
    return (f"<!-- edda-323:{r['key']} -->\nParent: #{r['parent']}; decision #323; roadmap #47. Task Size {r['project_size']}.\n\n"
            f"One reviewable PR. Required native path: this terminal -> #{m['qualification_gate']['issue']} -> #337.\n\n"
            f"Source: pinned English SRD5.2.1 PDF36-40 and exact referenced pages below; SHA256 {PIN}. "
            "CC-BY-4.0 attribution/adaptation: docs/srd-5.2.1-conformance.md. Fixture and ownership decisions are Edda additions.\n\n"
            f"Scope and deterministic positive acceptance: {r['fixture']}\n\n"
            "Negative acceptance: independently violate every stated level, option, source, timing, number, target and resource predicate; reject unsupported variants and forged entitlement without mutation.\n\n"
            f"Required acceptance: {m['common_acceptance']}\n\n"
            "Feature resolvers own behavior with isolated test entitlements; transitions consume those contracts and apply grants once. "
            "#427/#428 own completed rests; #429 owns reusable XP/PB/CON transaction; #430 provenance; #503 list membership (not execution); "
            "#487/#488 conditions; #384-401 feat selection and effects. #502 remains Bard-specific. "
            "#279/#280 qualify spell/item effects; #181/#185 and #183 retain spell-created world lifecycle ownership. "
            "Exact chosen spell/feat/item terminal and its qualification path must exist before availability; planning splitters #273/#277/#278 are not execution evidence. "
            "#382 remains child of323, after351, directly gating279/280/337; isolated caster fixtures never grant production access. "
            "No runtime, grants, coverage promotion, production mutation or inference in this decision.\n\n"
            f"Native prerequisites: {', '.join('#'+str(n) for n in r['blocked_by_issues'])}.\n\n"
            'Owned source keys (one behavior owner; record context does not duplicate ownership):\n'+ '\n'.join('- '+k for k in keys)+
            '\n\nExact contextual source records, including nested choices, numeric parameters and table capacities:\n```json\n'+json.dumps(records,indent=2)+'\n```\n')


def pages(path):
    result, seen = [], set()
    for page in range(1, 101):
        batch = api(f'{path}?per_page=100&page={page}')
        ids = [x['id'] for x in batch]
        assert len(ids) == len(set(ids)) and not seen.intersection(ids), 'repeated API page'
        seen.update(ids)
        result += batch
        if len(batch) < 100:
            return result
    raise AssertionError('pagination call cap')


def graph_read():
    graph, cursor, seen = [], None, set()
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
        assert batch['pageInfo']['endCursor'] != cursor, 'repeated cursor'
        cursor = batch['pageInfo']['endCursor']
    raise AssertionError('graph call cap')


def live(m):
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


def validate(m, receipt):
    assert m['source']['sha256'] == PIN
    records=[]
    for path, snapshot in m['catalogues'].items():
        current=json.loads((ROOT/path).read_text())
        assert current==snapshot and current['source']['sha256']==PIN, f'source drift {path}'
        records+=current['records']
    assert len(records)==83 and records==[x['record'] for x in m['records']]
    assert [r['level'] for r in records if r['kind']=='level']==list(range(1,21))
    keys={k for r in records for k in r['required_mechanics']+r['option_keys']}
    assert set(m['source_key_owners'])==keys and len(keys)==193
    rows={r['key']:r for r in m['issues']}
    assert len(rows)==len(m['issues'])==75 and len({r['issue'] for r in rows.values()})==len(rows)
    assert set(m['source_key_owners'].values())<=set(rows)|{'external:352','external:384','external:503'}
    for x in m['records']:
        assert x['key_owners']=={k:m['source_key_owners'][k] for k in x['record']['required_mechanics']+x['record']['option_keys']}
        assert x['state']=='metadata only; broad Cleric/Life Domain unqualified'
        assert x['evidence']==list(m['evidence_hashes'])
        assert set(x['support_evidence'])=={'engine','web','tui','qualification'}
        assert all(x['support_evidence'].values())
        if x['record']['kind']=='level':assert x['application_owner']=='level-'+str(x['record']['level'])
    for path, expected in m['evidence_hashes'].items():
        assert hashlib.sha256((ROOT/path).read_bytes()).hexdigest()==expected, f'evidence drift {path}'
    graph={r['number']:r for r in receipt['graph']}
    assert len(graph)==len(receipt['graph'])
    acyclic(graph)
    acyclic({n:{'blocked_by':[r['parent']] if r['parent'] else []} for n,r in graph.items()})
    assert all(len(r['blocked_by'])<=100 for r in graph.values())
    counts={}
    for r in graph.values():counts[r['parent']]=counts.get(r['parent'],0)+1
    assert all(v<=100 for k,v in counts.items() if k is not None)
    g=m['qualification_gate']; gate=g['issue']
    assert graph[323]['state'] in {'open','closed'} and graph[323]['parent']==283
    assert {285,288,310}<=set(graph[323]['blocked_by'])
    assert all(graph[n]['state']=='closed' for n in [285,288,310])
    assert graph[gate]['parent']==g['parent']==323 and graph[gate]['body']==g['body'] and graph[gate]['title']==g['title']
    sub=m['subordinate_gates']
    assert len(sub)==1
    sg=sub[0]
    grouped={r['issue'] for r in rows.values() if r['key'].startswith(('spell-','restore-'))}
    assert len(grouped)==14
    assert set(sg['blocked_by_issues'])=={323}|grouped
    assert graph[sg['issue']]['parent']==sg['parent']==gate
    assert graph[sg['issue']]['title']==sg['title'] and graph[sg['issue']]['body']==sg['body']
    assert set(graph[sg['issue']]['blocked_by'])==set(sg['blocked_by_issues'])
    assert set(range(384,402))<=set(m['reused_terminals'])
    bounding=m['gate_bounding']
    assert bounding['flattened_requirement_count']==111 and bounding['direct_count']==98
    assert bounding['subordinate']==sg['issue'] and set(bounding['mandatory_group'])==grouped
    assert bounding['subordinate_edges_read_back_before_direct_removal']
    required=({323}|{r['issue'] for r in rows.values()}|set(m['reused_terminals'])|set(m['qualification_dependencies'])|{sg['issue']})-grouped
    assert set(g['blocked_by_issues'])==required==set(graph[gate]['blocked_by']), 'terminal gate edge'
    assert gate in graph[337]['blocked_by'], 'gate to337'
    for r in rows.values():
        n=r['issue']
        assert graph[n]['parent']==r['parent'], f'parent{n}'
        assert set(r['blocked_by_issues'])==set(r['prerequisites'])|{rows[k]['issue'] for k in r['depends']}
        assert set(graph[n]['blocked_by'])==set(r['blocked_by_issues']), f'prerequisites{n}'
        assert graph[n]['title']==r['title'] and graph[n]['body']==body(m,r), f'body{n}'
    for r in list(rows.values())+[g,sg]:
        assert receipt['project'][str(r['issue'])]=={'type':'Task','size':r['project_size']}
    for n, old in {**m['protected'], **m['consumed_contracts']}.items():
        actual=graph[int(n)]
        assert actual['parent']==old['parent'] and set(old['blocked_by'])<=set(actual['blocked_by']), f'legacy edge{n}'
        body_valid=actual['body']==old['body']
        if int(n) in [47,54,283,337]:
            append=next(x['append'] for x in m['supplements'] if x['issue']==int(n))
            sealed=old['body']+append
            body_valid=(actual['body'].startswith(sealed) and actual['body'].count(append)==1
                        and actual['body'].count('<!-- edda-323-followups:start -->')==1
                        and actual['body'].count('<!-- edda-323-followups:end -->')==1)
        assert body_valid and actual['title']==old['title'], f'legacy body{n}'
        if int(n) in [334,382,498,501] or n in m['consumed_contracts']:assert actual['blocked_by']==old['blocked_by']
    assert graph[382]['parent']==323 and 351 in graph[382]['blocked_by']
    assert all(382 in graph[n]['blocked_by'] for n in [279,280,337])
    assert 183 in m['qualification_dependencies']
    for key in ['wish-duplicate','wish-object','wish-redo','wish-reality']:
        assert rows[key]['issue'] in graph[183]['blocked_by']
    assert {58,429}<=set(rows['restore-ability']['blocked_by_issues'])
    for n, expected in m['consumed_project'].items():
        assert receipt['project'][n] == expected, f'consumed Task/size {n}'
    for r in rows.values():
        if r['parent'] in [191,193]:
            assert all(r['issue'] in graph[n]['blocked_by'] for n in [279,280]), 'shared qualification edge'
    for level in range(1,21):
        row=rows['level-'+str(level)]
        assert row['parent']==gate
        if level>1:assert 'level-'+str(level-1) in row['depends']
        assert {'d8','preparation','slots'}<=set(row['depends'])
    for level in [4,8,12,16]:assert 'asi' in rows['level-'+str(level)]['depends']
    supplement_path=ROOT/'docs/investigations/issue-323-parent-supplements.json'
    assert json.loads(supplement_path.read_text())==m['supplements'], 'supplement drift'
    for s in m['supplements']:
        assert s['append'].startswith('\n\n<!-- edda-323-followups:start -->') and s['append'].endswith('<!-- edda-323-followups:end -->\n')
        assert digest(m['protected'][str(s['issue'])]['body'])==s['previous_sha256']
    return len(keys),len(graph)


def probes(m,r):
    def reject(mm,rr):
        try:validate(mm,rr)
        except (AssertionError,KeyError):return
        raise AssertionError('accepted corruption')
    mm=deepcopy(m);mm['source']['sha256']='bad';reject(mm,r)
    mm=deepcopy(m);next(iter(mm['catalogues'].values()))['records'][0]['name']='bad';reject(mm,r)
    mm=deepcopy(m);mm['source_key_owners'].pop(next(iter(mm['source_key_owners'])));reject(mm,r)
    mm=deepcopy(m);mm['records'][0]['key_owners']={};reject(mm,r)
    mm=deepcopy(m);mm['evidence_hashes'][next(iter(mm['evidence_hashes']))]='bad';reject(mm,r)
    g=m['qualification_gate']['issue']
    for n in m['qualification_gate']['blocked_by_issues']:
        rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==g)['blocked_by'].remove(n);reject(m,rr)
    sg=m['subordinate_gates'][0]
    for n in sg['blocked_by_issues']:
        rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==sg['issue'])['blocked_by'].remove(n);reject(m,rr)
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==sg['issue'])['parent']=323;reject(m,rr)
    for target,edge in [(337,g),(337,382),(337,334),(323,288)]:
        rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==target)['blocked_by'].remove(edge);reject(m,rr)
    for field,value in [('body','bad'),('title','bad'),('parent',283)]:
        rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==m['issues'][0]['issue'])[field]=value;reject(m,rr)
    rr=deepcopy(r);rr['project'][str(g)]['size']='XL';reject(m,rr)
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==279)['blocked_by'].remove(next(x['issue'] for x in m['issues'] if x['parent']==191));reject(m,rr)
    mm=deepcopy(m);next(x for x in mm['issues'] if x['key']=='level-12')['depends'].remove('asi');reject(mm,r)
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==283)['parent']=g;reject(m,rr)
    rr=deepcopy(r);next(x for x in rr['graph'] if x['number']==m['issues'][0]['issue'])['blocked_by'].append(337);reject(m,rr)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--live',action='store_true');p.add_argument('--pdf',type=Path);a=p.parse_args()
    m=json.loads(MANIFEST.read_text())
    if a.pdf:assert hashlib.sha256(a.pdf.read_bytes()).hexdigest()==PIN
    r=live(m) if a.live else json.loads(RECEIPT.read_text())
    keys,nodes=validate(m,r);probes(m,r)
    if a.live:RECEIPT.write_text(json.dumps(r,indent=2)+'\n')
    print(f'PASS:83 records/20 rows/{keys} keys; {len(m["issues"])} terminals; exact bodies/parents/prerequisites/Task sizes; terminal->ClericGate->337; {nodes} nodes, separate dependency/containment DAGs; corruption probes rejected')


if __name__=='__main__':main()
