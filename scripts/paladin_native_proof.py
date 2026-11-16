"""Bounded read-only native publication proof for the finite Paladin graph."""
from collections import Counter
from copy import deepcopy
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
from verify_fighter_scope import graph_read, api, acyclic
from publish_paladin_scope import rows, body

ROOT=Path(__file__).resolve().parents[1]
RECEIPT=ROOT/'docs/investigations/issue-327-link-verification.json'
PROJECT='PVT_kwDODjtmGM4BrFJ2'
OWNED_NATIVE_SEAL='acb39e3741d86ffec6bc2ecede112df0eb5cbce9448e1d854d4a68be2e9975d3'
# Filled only from the root-released accepted Monk merge/bookkeeping receipt.
# An unaccepted prefix cannot pass sealed mode merely by being copied into the manifest.
ACCEPTED_MONK_MERGE='3042172c935bdefa606c008c8ddda277b7548fd8'
ACCEPTED_PARENT_SEALS={
    '47':'7f45098b2071c4037f80cc12ce24022be73a9fac519d419eae4de2f27ce7d16a',
    '54':'59f0852d5dc077e6206a7b16a0ef444cde3abed19044e50f7bb08c06c592570b',
    '283':'1b70b7140f557648e249bd2f3203a1f8277b31f1cc3539824eba5e3a951fb6d0',
    '337':'64e7dad70c20be9dc09b2e61c150a783aacc8c93f26330fa6dacf133d712b882',
}
PUBLISHED_PARENT_SEALS={
    '47':'c1b834b560b3f2bb02f6dc00b5fe4af0dc420da03d4b48ec6f609cf1d4b0abfd',
    '54':'4518e86ba4b4bf00f81dcb25629f458e9a26584832fdc8268e13bd64100f6c24',
    '283':'588045b9ddcf5cd3cb4b0295434419f50a745f0fa51a2621a2355df93a9059c3',
    '337':'8ec09dd27418755ce3828ace6c1b5bd1665c1059aa88dc0267f4884e6a355aae',
}


def validate_parent(prefix,section,actual):
    assert actual.startswith(prefix+section),'immutable parent prefix or owned section changed'
    assert len(prefix+section)<=65536
    assert actual.count('<!-- edda-327:start -->')==actual.count('<!-- edda-327:end -->')==1,'duplicate owned marker'


def live(m):
    receipt=dict(observed_at=datetime.now(timezone.utc).isoformat(),graph=graph_read(),project={})
    numbers=sorted({x['issue'] for x in m['published_issues']}|set(map(int,m['consumed_contracts'])))
    for offset in range(0,len(numbers),25):
        batch=numbers[offset:offset+25]
        fields=[f'i{n}:issue(number:{n})'+'{issueType{name} projectItems(first:100){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}' for n in batch]
        d=api('graphql',{'query':'{repository(owner:"subculture-collective",name:"edda"){'+' '.join(fields)+'}}'})
        assert not d.get('errors'),d.get('errors')
        for n in batch:
            node=d['data']['repository'][f'i{n}'];assert not node['projectItems']['pageInfo']['hasNextPage']
            projects=[x for x in node['projectItems']['nodes'] if x['project']['id']==PROJECT]
            assert len(projects)==1,f'Project membership #{n}'
            receipt['project'][str(n)]=dict(type=(node['issueType'] or {}).get('name'),size=(projects[0]['fieldValueByName'] or {}).get('name'))
    return receipt


def reachable(graph,start,target):
    todo=[start];seen=set()
    while todo:
        n=todo.pop()
        if n==target:return True
        if n in seen:continue
        seen.add(n);todo.extend(graph[n]['blocked_by'])
    return False


def validate(m,r,sealed=False):
    g={x['number']:x for x in r['graph']};assert len(g)==len(r['graph'])
    published=m['published_issues'];by={x['key']:x for x in published}
    assert len(published)==len(by)==121 and len({x['issue'] for x in published})==121
    frozen={x['key']:{k:x[k] for k in ['issue','title','parent','blocked_by_issues','body','project_size']} for x in published}
    assert hashlib.sha256(json.dumps(frozen,sort_keys=True).encode()).hexdigest()==OWNED_NATIVE_SEAL,'immutable owned native contract drift'
    source_own,source_shared=rows(m);source_by={x['key']:x for x in source_own+source_shared}
    for x in published:
        n=x['issue'];actual=g[n]
        assert actual['parent']==x['parent'] and actual['title']==x['title'] and actual['body']==x['body'],f'owned contract #{n}'
        assert actual['blocked_by']==x['blocked_by_issues'],f'owned prerequisites #{n}'
        assert r['project'][str(n)]==dict(type='Task',size=x['project_size']),f'Task/Size #{n}'
        if x['key'] in source_by:
            expected=source_by[x['key']]
            for k in ['fixture','title','prerequisites','depends','project_size']:assert x[k]==expected[k],f'audited leaf {x["key"]}:{k}'
            assert x['blocked_by_issues']==sorted(set(x['prerequisites'])|{by[k]['issue'] for k in x['depends']})
            assert x['body']==body(m,x),'generated leaf body'
    assert g[327]['parent']==283 and g[327]['blocked_by']==[285,292,314]
    assert all(g[n]['state']=='closed' for n in [285,292,314])
    assert g[327]['state'] in ['open','closed']
    main=by['qualification']['issue'];classgate=by['class-gate']['issue'];sharedgate=by['shared-gate']['issue']
    assert g[main]['parent']==327 and g[main]['blocked_by']==sorted([327,classgate,sharedgate])
    assert main in g[337]['blocked_by']
    assert set(g[classgate]['blocked_by'])=={by[x['key']]['issue'] for x in source_own}
    assert all(g[by[x['key']]['issue']]['parent']==classgate for x in source_own)
    for c in [1,2,3]:
        cohort=by[f'cohort-{c}']['issue'];assert g[cohort]['parent']==277
        expected={by[x['key']]['issue'] for x in source_shared if x.get('cohort')==c}
        assert len(expected)==[20,20,14][c-1] and set(g[cohort]['blocked_by'])==expected
        assert all(g[n]['parent']==cohort for n in expected)
        assert cohort in g[279]['blocked_by'] and cohort in g[280]['blocked_by']
    for key,parent in [('blind',268),('deaf',268),('invisible',268),('blind-end',193),('deaf-end',193),('dispel-check',193),('dispel-object',193)]:
        n=by[key]['issue'];assert g[n]['parent']==parent
        assert n in g[279]['blocked_by'] and n in g[280]['blocked_by']
    assert 962 in g[by['shining-smite']['issue']]['blocked_by'],'actual Invisible terminal'
    ownids={x['issue'] for x in published}
    assert [x['key'] for x in published if 429 in x['blocked_by_issues']]==['hp'],'sole429consumer'
    assert hashlib.sha256(m['consumed_contracts']['779']['body'].encode()).hexdigest()=='556d1618d1cd2bd9c956fe38861d89e94ee14ff61a505b5d2b6a1fa46d19d150'
    assert m['consumed_contracts']['779']['blocked_by']==[61,64,65,66,67,488,778]
    for key in [f'level-{n}' for n in range(1,21)]+['later-1']:
        assert by['hp']['issue'] in g[by[key]['issue']]['blocked_by']
        assert reachable(g,by[key]['issue'],429)
    for x in published:
        if x['key'] in source_by:assert x['issue'] in m['flattened_mandatory_owners']
    for leaf in m['flattened_mandatory_owners']:
        assert reachable(g,main,leaf),f'mandatory leaf path #{leaf}'
    assert set(m['new_implementation_leaves'])=={by[x['key']]['issue'] for x in source_own+source_shared}
    todo=[main];seen=set()
    while todo:
        n=todo.pop()
        if n in seen:continue
        seen.add(n);todo.extend(g[n]['blocked_by'])
    assert seen==set(m['mandatory_transitive_ownership']),'complete flattened mandatory ownership drift'
    # Replacing a decision with closed state does not remove the immutable prerequisite path.
    closed=deepcopy(g);closed[327]['state']='closed'
    for leaf in m['flattened_mandatory_owners']:assert reachable(closed,337,leaf)
    for x in g.values():assert len(x['blocked_by'])==len(set(x['blocked_by']))<=100,'incoming cap'
    assert max(Counter(x['parent'] for x in g.values() if x['parent'] is not None).values())<=100,'child cap'
    assert max(Counter(n for x in g.values() for n in x['blocked_by']).values())<=100,'outgoing cap'
    acyclic(g);acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] else []) for n,x in g.items()})
    for n,old in m['consumed_contracts'].items():
        current=g[int(n)]
        assert current['parent']==old['parent'] and current['title']==old['title'] and current['body']==old['body'],f'consumed contract #{n}'
        if int(n) in [279,280]:assert set(old['blocked_by'])<=set(current['blocked_by'])
        else:assert current['blocked_by']==old['blocked_by'],f'consumed prerequisites #{n}'
        assert r['project'][n]==m['consumed_project'][n],f'consumed metadata #{n}'
    for n,old in m['historical_graph'].items():
        assert g[int(n)]['parent']==old['parent'],f'historical parent #{n}'
        assert set(old['blocked_by'])<=set(g[int(n)]['blocked_by']),f'historical edge #{n}'
    for target,prerequisite in m['preserved_shared_edges']:
        assert prerequisite in g[target]['blocked_by'],f'preserved shared edge {target}->{prerequisite}'
    for n,parent in m['preserved_parents'].items():assert g[int(n)]['parent']==parent,f'preserved old owner parent #{n}'
    assert len(m['legacy_root_edges'])>=45 and set(m['legacy_root_edges'])<=set(g[337]['blocked_by'])
    supplements=json.loads((ROOT/'docs/investigations/issue-327-parent-supplements.json').read_text())
    assert supplements==m['supplements'],'owned supplement artifact drift'
    for s in supplements:
        assert len(s['append'])==s['append_chars']<=(100 if s['issue']==47 else 750)
        if sealed:
            assert ACCEPTED_MONK_MERGE and m['accepted_monk_merge']==ACCEPTED_MONK_MERGE,'accepted Monk merge not released'
            assert len(ACCEPTED_PARENT_SEALS)==4,'accepted prefix seals not released'
            seal=m['accepted_parent_baseline'][str(s['issue'])];prefix=seal['body']
            assert hashlib.sha256(prefix.encode()).hexdigest()==seal['sha256']==ACCEPTED_PARENT_SEALS[str(s['issue'])]
            actual=g[s['issue']]['body'];section=s['append']
            assert hashlib.sha256((prefix+section).encode()).hexdigest()==s['published_sha256']==PUBLISHED_PARENT_SEALS[str(s['issue'])],'published section seal drift'
            validate_parent(prefix,section,actual)
        else:
            assert '<!-- edda-327:start -->' not in g[s['issue']]['body'],'supplements published without accepted seal'
            if 'accepted_parent_baseline' in m:assert g[s['issue']]['body']==m['accepted_parent_baseline'][str(s['issue'])]['body'],'pending accepted prefix drift'
    return len(g)


def probes(m,r,sealed=False):
    validate(m,r,sealed)
    rejected=0
    def reject(rr):
        nonlocal rejected
        try:validate(m,rr,sealed)
        except (AssertionError,KeyError):rejected+=1;return
        raise AssertionError('accepted native corruption')
    for x in m['published_issues']:
        for field,value in [('parent',283),('body','changed'),('title','changed')]:
            rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==x['issue'])[field]=value;reject(rr)
        for edge in x['blocked_by_issues']:
            rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==x['issue'])['blocked_by'].remove(edge);reject(rr)
    for target,edge in [(337,845),(279,848),(280,848),(334,298)]:
        rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==target)['blocked_by'].remove(edge);reject(rr)
    by={x['key']:x for x in m['published_issues']}
    for target in [279,280]:
        for key in ['cohort-1','cohort-2','cohort-3','blind','deaf','invisible','blind-end','deaf-end','dispel-check','dispel-object']:
            rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==target)['blocked_by'].remove(by[key]['issue']);reject(rr)
    for target,edge in m['preserved_shared_edges']:
        rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==target)['blocked_by'].remove(edge);reject(rr)
    rr=deepcopy(r);rr['project']['851']['size']='L';reject(rr)
    rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==851)['blocked_by'].append(337);reject(rr)
    # The unrelated Ranger gateway968 is allowed to fill429's100th outgoing slot.
    # A101st unrelated consumer must fail solely at the platform capacity boundary.
    count=sum(429 in x['blocked_by'] for x in r['graph'])
    rr=deepcopy(r)
    for i in range(100-count):rr['graph'].append(dict(number=999900+i,parent=None,title='Capacity control',body='',state='open',blocked_by=[429]))
    validate(m,rr,sealed)
    rr['graph'].append(dict(number=999999,parent=None,title='Capacity control',body='',state='open',blocked_by=[429]));reject(rr)
    rr=deepcopy(r);target=next(z for z in rr['graph'] if z['number']==279)
    for i in range(101-len(target['blocked_by'])):
        n=999800+i;rr['graph'].append(dict(number=n,parent=None,title='Incoming capacity control',body='',state='open',blocked_by=[]));target['blocked_by'].append(n)
    reject(rr)
    rr=deepcopy(r);rr['graph'].append(dict(number=999700,parent=None,title='Containment capacity control',body='',state='open',blocked_by=[]))
    for i in range(101):rr['graph'].append(dict(number=999500+i,parent=999700,title='Child capacity control',body='',state='open',blocked_by=[]))
    reject(rr)
    # Future unrelated prerequisites on the two shared qualification gates are allowed.
    for target in [279,280]:
        rr=deepcopy(r);rr['graph'].append(dict(number=999998,parent=None,title='Future unrelated owner',body='',state='open',blocked_by=[]))
        next(z for z in rr['graph'] if z['number']==target)['blocked_by'].append(999998);validate(m,rr,sealed)
    if sealed:
        sections=json.loads((ROOT/'docs/investigations/issue-327-parent-supplements.json').read_text())
        for s in sections:
            n=s['issue'];prefix=m['accepted_parent_baseline'][str(n)]['body'];section=s['append']
            for text in [prefix,prefix+section+section,prefix+section.replace('edda-327:end','edda-327:changed'),prefix+section+'\n<!-- edda-327:start -->altered<!-- edda-327:end -->','changed'+prefix+section]:
                rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==n)['body']=text;reject(rr)
            rr=deepcopy(r);next(z for z in rr['graph'] if z['number']==n)['body']+='\nUnrelated future pure append.';validate(m,rr,True)
    return rejected
