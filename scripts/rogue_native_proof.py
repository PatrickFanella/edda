"""Exact read-only Rogue native publication proof. FINAL prefixes are ROOT-released only."""
from collections import Counter
from copy import deepcopy
from datetime import datetime,timezone
import hashlib
import json
from pathlib import Path
from verify_fighter_scope import graph_read,api,acyclic
from publish_rogue_scope import rows,JOURNAL
from rogue_capacity import project

ROOT=Path(__file__).resolve().parents[1]
RECEIPT=ROOT/'docs/investigations/issue-329-link-verification.json'
PROJECT='PVT_kwDODjtmGM4BrFJ2'
OWNED_SEAL='eda80018e3e44db7813ccec745124718ce9f21ea04cf84ba6c06ac84295e6b80'
CONSUMED_SEAL='9a79ecaf85edeab10722e16df3c6602cf109c29b41f5d1e1c237cf58a493ef76'
# Never populate from current live bodies. These require the accepted ROOT release.
from freeze_rogue_parents import MERGE as ACCEPTED_POST_RANGER,SEALS as ACCEPTED_PREFIXES,validate_release

SUPPLEMENT_SEALS={'47': 'bf593511dbf09fa72139420f3dfcb74de84a8fc5884c2a0acd4f7b3eb11c3175', '54': 'ccc07cea6e80e01d472d553d2f40c8d36713bf3080e65d8756bee300e033d16d', '283': 'ccc07cea6e80e01d472d553d2f40c8d36713bf3080e65d8756bee300e033d16d', '337': 'ccc07cea6e80e01d472d553d2f40c8d36713bf3080e65d8756bee300e033d16d'}
ADDITIVE_QUALIFICATION_ROOTS=frozenset({279,280,337})

PUBLISHED_PARENT_SEALS={'47': 'b471db938c1d2357fee6bf70f03e19529f6d991422f8a978820874e3c8e2c842', '54': 'feda9427a6862e5d9b288d7e7017beb4fd3ed4965a02f9961b8a78d440da2169', '283': 'b4b5f948f6484924fb43d19586717fc982350e7ae1e63120ad6854c594c8464c', '337': '3a7bf09aa9eb6d48a81139902351f8bf2fcc6c5e96900425d6b52bbb3a197b07'}

def digest(x):return hashlib.sha256(json.dumps(x,sort_keys=True).encode()).hexdigest()

def caps(g):
    assert len(g)==len({x['number'] for x in g})
    graph={x['number']:x for x in g}
    for x in g:
        assert len(x['blocked_by'])==len(set(x['blocked_by']))<=100
        assert set(x['blocked_by'])<=set(graph)
        assert x['parent'] is None or x['parent'] in graph
    assert max(Counter(n for x in g for n in x['blocked_by']).values(),default=0)<=100
    assert max(Counter(x['parent'] for x in g if x['parent'] is not None).values(),default=0)<=100
    acyclic(graph);acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] is not None else []) for n,x in graph.items()})

def reaches(g,start,target):
    seen=set();todo=[start]
    while todo:
        n=todo.pop()
        if n==target:return True
        if n in seen:continue
        seen.add(n);todo.extend(g[n]['blocked_by'])
    return False

def read_graph(api_call=api,limit=20):
    from verify_rogue_scope import pages
    result=[];seen=set();cursor=None;cursors=set()
    for _ in range(limit):
        q='query($cursor:String){repository(owner:"subculture-collective",name:"edda"){issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{number title body state parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}}}}}'
        d=api_call('graphql',{'query':q,'variables':{'cursor':cursor}});assert not d.get('errors'),d.get('errors')
        batch=d['data']['repository']['issues'];assert len(batch['nodes'])<=100
        for x in batch['nodes']:
            assert x['number'] not in seen,'repeated issue page';seen.add(x['number'])
            deps=x['blockedBy']['nodes']
            if x['blockedBy']['pageInfo']['hasNextPage']:
                deps=pages(f'repos/subculture-collective/edda/issues/{x["number"]}/dependencies/blocked_by',api_call,limit=3)
            result.append(dict(number=x['number'],title=x['title'],body=x['body'],state=x['state'].lower(),parent=(x['parent'] or {}).get('number'),blocked_by=sorted(n['number'] for n in deps)))
        if not batch['pageInfo']['hasNextPage']:return result
        cursor=batch['pageInfo']['endCursor'];assert cursor and cursor not in cursors,'repeated cursor';cursors.add(cursor)
    raise AssertionError('graph call cap')

def collect(m):
    r=dict(observed_at=datetime.now(timezone.utc).isoformat(),limits='Single complete GitHub native graph/body readback and paginated exact Task/Project Size metadata; no runtime/integration/device proof. FINAL prefixes not accepted.',graph=read_graph(),project={})
    nums=sorted({x['issue'] for x in m['published_issues']}|set(map(int,m['consumed_contracts'])))
    for offset in range(0,len(nums),25):
        batch=nums[offset:offset+25]
        q='{repository(owner:"subculture-collective",name:"edda"){'+''.join(f'i{n}:issue(number:{n})'+'{issueType{name} projectItems(first:100){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}' for n in batch)+'}}'
        d=api('graphql',{'query':q});assert not d.get('errors'),d.get('errors')
        for n in batch:
            node=d['data']['repository'][f'i{n}'];assert not node['projectItems']['pageInfo']['hasNextPage']
            items=[x for x in node['projectItems']['nodes'] if x['project']['id']==PROJECT];assert len(items)==1
            r['project'][str(n)]=dict(type=(node['issueType'] or {}).get('name'),size=(items[0]['fieldValueByName'] or {}).get('name'))
    return r

def validate(m,r,sealed=False):
    from verify_rogue_scope import multiclass_ownership
    multiclass_ownership(m)
    sealed=sealed or m.get('final_seal')=='PASS'
    caps(r['graph']);g={x['number']:x for x in r['graph']};owned=m['published_issues'];by={x['key']:x for x in owned}
    assert len(owned)==len(by)==77 and len({x['issue'] for x in owned})==77
    assert digest(owned)==OWNED_SEAL,'independent complete owned contract seal'
    assert digest(m['consumed_contracts'])==CONSUMED_SEAL,'complete consumed contracts seal'
    journal=json.loads(JOURNAL.read_text())
    expected=rows(m,journal['before_graph'],journal['issues'])
    assert owned==expected,'source/body/key/gateway roster reconstruction'
    for x in owned:
        actual=g[x['issue']]
        for k in ['title','body','parent']:assert actual[k]==x[k],f'owned#{x["issue"]}:{k}'
        assert actual['blocked_by']==x['blocked_by_issues'],f'owned#{x["issue"]} prerequisites'
        assert r['project'][str(x['issue'])]==dict(type='Task',size='M')
    for n,expected in m['consumed_contracts'].items():
        actual=g[int(n)]
        for k in ['title','body','parent','blocked_by']:assert actual[k]==expected[k],f'consumed#{n}:{k}'
    # Preserve EVERY preexisting dependency/containment path; unrelated additive edges allowed.
    for old in journal['before_graph']:
        actual=g[old['number']]
        assert old['parent']==actual['parent'],f'old containment#{old["number"]}'
        assert set(old['blocked_by'])<=set(actual['blocked_by']),f'old required path#{old["number"]}'
    for x in owned:
        assert reaches(g,337,x['issue']),f'mandatory337 path#{x["issue"]}'
    closed=deepcopy(g);closed[329]['state']='closed'
    for x in owned:assert reaches(closed,337,x['issue'])
    for n in [285,294,316]:assert g[n]['state']=='closed'
    assert g[329]['parent']==283 and g[329]['blocked_by']==[285,294,316] and g[329]['state'] in ['open','closed']
    assert g[968]['parent']==336 and g[968]['blocked_by']==[429]
    assert hashlib.sha256(g[779]['body'].encode()).hexdigest()=='556d1618d1cd2bd9c956fe38861d89e94ee14ff61a505b5d2b6a1fa46d19d150','generic PLAYER Unarmed contract'
    assert g[779]['parent']==50 and g[779]['blocked_by']==[61,64,65,66,67,488,778]
    assert Counter(n for x in g.values() for n in x['blocked_by'])[429]==100
    assert not any(429 in x['blocked_by_issues'] for x in owned)
    assert set(g[by['qualification']['issue']]['blocked_by'])=={329,279,280,by['features']['issue'],by['levels']['issue'],by['shared']['issue']}
    assert by['qualification']['issue'] in g[337]['blocked_by']
    assert all(by['shared']['issue'] in g[n]['blocked_by'] for n in [279,280])
    assert g[by['shared']['issue']]['parent']==192 and g[by['hide']['issue']]['parent']==50
    assert all(x['parent']!=283 for x in owned)
    # Every immutable historical edge and containment survives; only new owned/root edges added.
    old={x['number']:x for x in journal['before_graph']}
    for n,x in old.items():
        assert x['parent']==g[n]['parent']
        expected=set(x['blocked_by'])
        if n==337:expected.add(by['qualification']['issue'])
        if n in [279,280]:expected.add(by['shared']['issue'])
        if n in ADDITIVE_QUALIFICATION_ROOTS:
            # Later classes can add required outputs to these shared roots;
            # their existing requirements and our owned paths remain mandatory.
            assert expected<=set(g[n]['blocked_by']),f'old required root edges#{n}'
            assert x['title']==g[n]['title'],f'shared qualification title#{n}'
            if n in [279,280]:assert x['body']==g[n]['body'],f'shared qualification body#{n}'
        else:
            assert set(g[n]['blocked_by'])==expected,f'old edges#{n}'
    for x in m['issues']:
        assert reaches(g,337,by[x['key']]['issue'])
    for level in range(1,21):
        n=by[f'level-{level}']['issue'];assert by['hp']['issue'] in g[n]['blocked_by'];assert reaches(g,n,429)
        assert by['multiclass']['issue'] in g[n]['blocked_by']
    assert by['hp']['issue'] in g[by['later-1']['issue']]['blocked_by']
    closed=deepcopy(g);closed[329]['state']='closed'
    for x in m['issues']:assert reaches(closed,337,by[x['key']]['issue'])
    assert g[by['execution']['issue']]['blocked_by']==[61,64,65,66,67,273,353]
    consumers=project(m,journal['before_graph'])['execution_consumers']
    assert {x['key'] for x in owned if by['execution']['issue'] in x['blocked_by_issues']}==set(consumers)
    assert not any(by[k]['issue'] in g[by['execution']['issue']]['blocked_by'] for k in consumers)
    # Item integrations preserve exact separate effect and recovery owners.
    for key,dep in [('scroll-aid',614),('scroll-light',967),('scroll-detect',906),('ring-jumping',1064)]:assert dep in g[by[key]['issue']]['blocked_by']
    assert set(range(385,402))|{384,389,487,427,428,778,779,962}<=set(g[by['improvement']['issue']]['blocked_by'])
    release=m['accepted_parent_baseline'];validate_release(release)
    assert m['accepted_post_ranger']==ACCEPTED_POST_RANGER
    assert release==json.loads((ROOT/'docs/investigations/issue-329-accepted-parent-baselines.json').read_text()),'complete accepted prefix release drift'
    if sealed:
        from verify_rogue_scope import parent_append
        for n,h in ACCEPTED_PREFIXES.items():
            prefix=release['body_seals'][n]['body']
            assert hashlib.sha256(prefix.encode()).hexdigest()==h,'independent released prefix seal'
            section=(ROOT/f'docs/investigations/issue-329-parent-{n}-PENDING.md').read_text()
            assert hashlib.sha256(section.encode()).hexdigest()==SUPPLEMENT_SEALS[n],'independent exact supplement seal'
            assert hashlib.sha256((prefix+section).encode()).hexdigest()==PUBLISHED_PARENT_SEALS[n],'ROOT exact resulting publication seal'
            parent_append(prefix,section,g[int(n)]['body'])
    return dict(nodes=len(g),owned_tasks=len(owned),owned_edges=sum(len(x['blocked_by_issues']) for x in owned),root_edges=3,terminals=72,final_seal='PENDING' if not sealed else 'PASS')
