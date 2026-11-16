#!/usr/bin/env python3
"""Independent read-only publication proof for the frozen #332 reservation.

No issue writes, coordinator imports, runtime qualification or prefix refresh.
Source semantics remain checked separately against literal/PDF comparators.
"""
import hashlib
import gzip
import json
from pathlib import Path
from collections import Counter
from wizard_native_checkpoint import acyclic

APPROVAL_SHA256='66b934f5cac58d21644822bd36098b0ee1cad535708a9e3f6b49bdf0a2f0fe41'
PROJECTION_SHA256='8c8b33d7db488a98ba2543a44530dc1fe60a9cb3db0c497523a58a7ef784fb0f'
APPROVAL_DIR=Path(__file__).resolve().parents[1]/'docs/investigations/issue-332-approved-publication'
PROJECT='PVT_kwDODjtmGM4BrFJ2'
RENDERER_SHA256='e466e1b41506c35f6f03b6c3cc8d2bbbfb57e168607fa3c63954291930247238'
VECTORS_SHA256='4ba7f97dd6f1702ff049056f10e46736cb1da27fe7274553b4bff9e54077bdee'

def digest(data):return hashlib.sha256(data).hexdigest()
def read_pinned(path,expected):
    path=Path(path)
    data=path.read_bytes() if path.exists() else gzip.decompress(path.with_suffix(path.suffix+'.gz').read_bytes())
    assert digest(data)==expected,('frozen artifact changed',str(path))
    return json.loads(data)

def trusted_inputs(directory=APPROVAL_DIR):
    approval=read_pinned(directory/'approval.json',APPROVAL_SHA256)
    assert approval['status']=='APPROVED_FINITE_NATIVE_PUBLICATION'
    assert approval['counts']=={'330':305,'331':281,'332':285} and approval['new_nodes']==871
    frozen={n:read_pinned(directory/f'{n}-approved-proposal.json',row['sha256']) for n,row in approval['inputs'].items()}
    baseline=read_pinned(directory/'native-baseline.json',approval['native_baseline_sha256'])
    projection=read_pinned(directory/'approved-projection.json',PROJECTION_SHA256)
    nodes={n['alias']:n for n in approval['approved_nodes']}
    assert len(nodes)==871
    packs={'330':frozen['330']['tasks']+frozen['330']['shared_split_proposal']+frozen['330']['gates'],
           '331':frozen['331']['tasks']+frozen['331']['gates'],
           '332':frozen['332']['tasks']+frozen['332']['gates']}
    for owner,rows in packs.items():
        assert len(rows)==approval['counts'][owner]
        for row in rows:
            alias=owner+':'+row['key'];a=nodes[alias]
            # The approval itself pins the exact raw proposal. Body/graph facts
            # are also compared semantically, not accepted on digest agreement.
            assert a['key']==row['key'] and a['title']==row['title']
            assert a['type']=='Task' and a['size']=='M'
            assert digest(a['approved_body'].encode())==a['approved_body_sha256']
            assert a['approved_body']==row.get('native_body',row['body'])
            assert set(a['blocked_by_refs'])==set(projection['graph'][alias])
            assert a['parent_ref']==projection['parents'][alias]
    # Derive the entire graph independently from approval nodes, old baseline
    # and exactly five reviewed root additions. Never trust a receipt's graph.
    derived={str(r['number']):list(map(str,r['blocked_by'])) for r in baseline['graph']}
    parents={str(r['number']):str(r['parent']) for r in baseline['graph'] if r['parent']}
    for alias,a in nodes.items():derived[alias]=a['blocked_by_refs'];parents[alias]=a['parent_ref']
    assert {(e['issue'],e['blocker']) for e in approval['allowed_old_root_edges']}=={
        ('337','330:class-gate'),('337','331:class-gate'),('337','332:class-gate'),
        ('279','330:shared-aggregate'),('280','330:shared-aggregate')}
    for edge in approval['allowed_old_root_edges']:derived[edge['issue']].append(edge['blocker'])
    assert {k:sorted(v) for k,v in derived.items()}=={k:sorted(v) for k,v in projection['graph'].items()},'independent whole approved graph'
    assert parents==projection['parents'],'independent whole approved containment'
    return approval,frozen,baseline,projection,nodes

def validate_candidate(candidate,frozen):
    assert candidate==frozen,'own semantic scope/counts/source ownership must stay exactly frozen; bindings use sidecar only'


def validate_registry(registry,nodes,baseline):
    """Canonical registry input: exact alias -> {number,id,node_id} mapping."""
    assert isinstance(registry,dict),'registry must be an object'
    assert set(registry)>={'schema_version','approval_sha256','renderer_contract_sha256','status','nodes'}
    assert set(registry)<={'schema_version','approval_sha256','renderer_contract_sha256','status','nodes','updated_at','observed_at'},'unexpected registry top field'
    assert type(registry['schema_version']) is int and registry['schema_version']==1
    assert registry['approval_sha256']==APPROVAL_SHA256 and registry['renderer_contract_sha256']==RENDERER_SHA256,'independent registry hash pins'
    assert registry['status'] in ['IDS_ALLOCATED','WIRING','VERIFIED'],'incomplete registry phase'
    registry=registry['nodes']
    assert isinstance(registry,dict),'registry.nodes must be an object'
    assert set(registry)==set(nodes),'complete global871 alias registry required'
    old={r['number'] for r in baseline['graph']};numbers=[];ids=[];node_ids=[]
    for alias,row in registry.items():
        assert isinstance(row,dict),'registry entry must be an object'
        assert set(row)=={'number','id','node_id'},'exact registry entry fields'
        assert type(row['number']) is int and row['number']>0
        assert type(row['id']) is int and row['id']>0
        assert isinstance(row['node_id'],str) and row['node_id']
        assert row['number'] not in old,'registry reuses old issue'
        numbers.append(row['number']);ids.append(row['id']);node_ids.append(row['node_id'])
    assert len(set(numbers))==len(set(ids))==len(set(node_ids))==871,'noninjective alias registry'


def resolved(ref,registry):
    if isinstance(ref,str) and ref.isascii() and ref.isdecimal():return int(ref)
    registry=registry['nodes']
    assert ref in registry,('unbound approved output',ref)
    return registry[ref]['number']


def expected_native(projection,nodes,registry):
    graph={resolved(k,registry):sorted(resolved(x,registry) for x in v) for k,v in projection['graph'].items()}
    parents={resolved(k,registry):resolved(v,registry) for k,v in projection['parents'].items()}
    assert len(graph)==1888
    return graph,parents


def metadata(row):
    entries=[p for p in row['projects'] if p['project']['id']==PROJECT]
    assert len(entries)==1,('exact project membership',row['number'])
    value=entries[0]['fieldValueByName'];assert value is not None,('missing Project Size',row['number'])
    return {'type':row['type'],'size':value['name']}


def has_path(graph,start,end):
    seen=set();todo=[start]
    while todo:
        n=todo.pop()
        if n==end:return True
        if n in seen:continue
        seen.add(n);todo.extend(graph.get(n,()))
    return False


def approved_paths(projection,registry):
    """Resolve six formerly unresolved source obligations, without executing any."""
    groups={
        'animate-dead':['332:spell-animate-dead-creation','332:spell-animate-dead-command','332:spell-animate-dead-control-renewal','332:spell-animate-dead-upcast','331:creature-skeleton-statistics','331:creature-skeleton-attacks','330:zombie-stat-actions','330:zombie-undead-fortitude'],
        'phantom-steed':['332:spell-phantom-steed-create','332:spell-phantom-steed-ride-ending','332:form-riding-horse-statistics','332:form-riding-horse-hooves'],
        'shapechange':['332:spell-shapechange-form-admission','332:spell-shapechange-statistics','332:spell-shapechange-temporary-hp','332:spell-shapechange-equipment','332:qualified-form-output'],
        'simulacrum':['332:spell-simulacrum-create-statistics','332:spell-simulacrum-command-limitations','332:spell-simulacrum-repair','332:spell-simulacrum-ending','332:form-qualified-snapshot','332:form-commoner-statistics','332:form-commoner-training','332:form-commoner-club'],
        'named-interoperability':['332:spell-glyph-of-warding-stored-spell','332:spell-contingency-store','332:spell-guards-and-wards-doors','332:spell-guards-and-wards-stairs','332:spell-prismatic-wall-green','332:spell-resilient-sphere-disintegrate','332:spell-magic-jar-possession','332:spell-modify-memory-restore'],
        'complete-cross-class-output-graph':['330:shared-output','331:shared-output','332:shared-output','330:shared-aggregate','332:class-gate','332:source-advancement-output','332:casting-output','332:execution-output'],
    }
    return {name:dict(status='MANDATORY APPROVED OUTPUT PATHS; runtime execution UNQUALIFIED',outputs={alias:resolved(alias,registry) for alias in aliases}) for name,aliases in groups.items()}


def allocation_body(node):
    return '<!-- edda-caster-native:'+node['alias']+' -->\n'+node['approved_body']


def _render_body(node,registry):
    """Pure format function for golden vectors; not registry/final validation."""
    refs=node['blocked_by_refs'];assert len(refs)==len(set(refs)),'duplicate prerequisite reference'
    numbers=[resolved(r,registry) for r in refs]
    assert len(numbers)==len(set(numbers)),'duplicate prerequisite identity'
    prerequisites=', '.join('#'+str(n) for n in sorted(numbers)) if numbers else 'none'
    return (allocation_body(node)+'\n\n<!-- edda-caster-native-binding:start -->\n'
            f"Decision: #{node['owner_decision']}\nSource key: {node['alias']}\n"
            f"Native issue: #{resolved(node['alias'],registry)}\nNative parent: #{resolved(node['parent_ref'],registry)}\n"
            f'Native prerequisites: {prerequisites}\nRoot approval SHA256: {APPROVAL_SHA256}\n'
            f'Renderer contract SHA256: {RENDERER_SHA256}\n'
            'Publication: source contract published; runtime unqualified.\n<!-- edda-caster-native-binding:end -->\n')


def renderer_contract():
    contract=read_pinned(APPROVAL_DIR/'native-renderer-contract.json',RENDERER_SHA256)
    vectors=read_pinned(APPROVAL_DIR/'native-renderer-vectors.json',VECTORS_SHA256)
    assert contract['approval_sha256']==APPROVAL_SHA256
    for case in vectors['cases']:
        assert allocation_body(case['approved_node'])==case['allocation_body']
        assert _render_body(case['approved_node'],vectors['example_registry'])==case['final_body']
    return contract,vectors


def validate_live(receipt,registry,inputs):
    approval,frozen,baseline,projection,nodes=inputs
    validate_registry(registry,nodes,baseline)
    assert registry['status']=='VERIFIED','ROOT verified wiring boundary required for native proof'
    renderer_contract()
    expected,parents=expected_native(projection,nodes,registry)
    live={r['number']:r for r in receipt['graph']}
    assert len(live)==len(receipt['graph']),'duplicate live issue'
    assert set(expected)<=set(live),'incomplete approved/foreign native graph'
    graph={n:r['blocked_by'] for n,r in live.items()}
    assert all(len(ds)==len(set(ds))<=100 for ds in graph.values()),'incoming/duplicate dependency cap'
    assert all(x in live for ds in graph.values() for x in ds),'missing foreign dependency node'
    outgoing=Counter(x for ds in graph.values() for x in ds);children=Counter(r['parent'] for r in live.values() if r['parent'])
    assert max(outgoing.values(),default=0)<=100 and max(children.values(),default=0)<=100,'complete foreign outgoing/child cap'
    acyclic(graph);acyclic({n:[r['parent']] if r['parent'] else [] for n,r in live.items()})
    for n,deps in expected.items():
        assert set(graph[n])==set(deps),('exact approved native prerequisites',n)
        assert live[n]['parent']==parents.get(n),('exact native containment',n)
    for alias,a in nodes.items():
        n=resolved(alias,registry);r=live[n]
        assert r['title']==a['title'] and r['body']==_render_body(a,registry),('whole approved native title/body',alias)
        identity=registry['nodes'][alias]
        assert type(r['id']) is int and r['id']==identity['id'] and r['node_id']==identity['node_id'] and r['typename']=='Issue',('actual GitHub issue identity',alias)
        assert metadata(r)=={'type':'Task','size':'M'},('actual Task/Project Size',alias)
        assert r['state'].upper()=='OPEN',('implementation/qualification task closed prematurely',alias)
    for old in baseline['graph']:
        now=live[old['number']]
        assert now['title']==old['title'] and now['body']==old['body'],('old body/title changed before root prefix release',old['number'])
        # Closed decisions may close after their accepted PR. They still remain
        # mandatory nodes/edges: never filter the graph by issue state.
        if str(old['number']) in baseline['project']:
            assert metadata(now)==baseline['project'][str(old['number'])],('old Task/Project metadata changed',old['number'])
    assert live[183]['blocked_by']==next(r['blocked_by'] for r in baseline['graph'] if r['number']==183),'old183 unchanged'
    q=resolved('332:class-gate',registry);shared=resolved('332:shared-output',registry);aggregate=resolved('330:shared-aggregate',registry)
    assert q in graph[337] and aggregate in graph[279] and aggregate in graph[280]
    assert not has_path(graph,shared,q),'shared output depends on consuming class qualifier'
    for alias,a in nodes.items():
        if not alias.startswith('332:'):continue
        n=resolved(alias,registry)
        assert has_path(graph,337,n),('mandatory Wizard terminal path',alias)
        # Shared runtime parts (and bounded generic forms) must also be reachable
        # from both actual shared roots through the aggregate/shared output.
        key=a['key']
        if key.startswith('spell-') and not key.startswith('spell-mastery') or key.startswith('form-'):
            assert has_path(graph,shared,n) and has_path(graph,279,n) and has_path(graph,280,n),('both shared native rootpaths',alias)
    return dict(status='PUBLISHED_NATIVE_PROOF_PASS_UNSEALED',approval_sha256=APPROVAL_SHA256,own_nodes=285,global_new_nodes=871,approved_graph_nodes=1888,observed_nodes=len(live),caps={'incoming':max(map(len,graph.values())),'outgoing':max(outgoing.values()),'children':max(children.values())},accepted_parent_prefix=None,source_bindings=approved_paths(projection,registry),runtime_status='UNQUALIFIED; source/native decision proof only')


def build_binding(registry,inputs):
    approval,frozen,baseline,projection,nodes=inputs
    validate_registry(registry,nodes,baseline)
    own={alias:dict(**registry['nodes'][alias],title=a['title'],parent=resolved(a['parent_ref'],registry),
                    blocked_by=sorted(resolved(x,registry) for x in a['blocked_by_refs']),
                    body_sha256=digest(_render_body(a,registry).encode()),type='Task',size='M')
         for alias,a in nodes.items() if a['owner_decision']==332}
    assert len(own)==285
    return dict(schema_version=1,status='IDS_BOUND_NATIVE_READBACK_PENDING',approval_sha256=APPROVAL_SHA256,
                renderer_contract_sha256=RENDERER_SHA256,registry_sha256=digest(json.dumps(registry,sort_keys=True).encode()),
                own_frozen_source_sha256=approval['inputs']['332']['sha256'],own_nodes=own,
                approved_execution_paths=approved_paths(projection,registry),accepted_parent_prefix=None,
                parent_supplements_status='PENDING ROOT-only serial application',runtime_status='UNQUALIFIED')


def main():
    import argparse
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--registry',type=Path)
    p.add_argument('--binding-output',type=Path)
    p.add_argument('--collect-native',action='store_true',help='Run ONCE only after ROOT VERIFIED wiring delivery')
    p.add_argument('--native-receipt',type=Path,help='Existing receipt, or output path with --collect-native')
    p.add_argument('--proof-output',type=Path)
    p.add_argument('--final',action='store_true')
    a=p.parse_args();inputs=trusted_inputs();renderer_contract()
    own=Path(__file__).resolve().parents[1]/'docs/investigations/issue-332-wizard-scope.json'
    validate_candidate(json.loads(own.read_text()),inputs[1]['332'])
    if a.final:
        raise AssertionError('accepted POST-WARLOCK root prefix release and real four appends are still HELD; no final seal may be fabricated')
    if not a.registry:
        assert not any([a.binding_output,a.collect_native,a.native_receipt,a.proof_output]),'registry required'
        print('PASS exact frozen871-node approval and renderer; IDs/native publication proof pending ROOT boundary delivery')
        return
    registry=json.loads(a.registry.read_text());binding=build_binding(registry,inputs)
    if a.binding_output:a.binding_output.write_text(json.dumps(binding,indent=2)+'\n')
    if a.collect_native:
        assert a.native_receipt is not None and registry['status']=='VERIFIED','ROOT VERIFIED receipt boundary required'
        from wizard_native_checkpoint import collect
        receipt=collect()
    elif a.native_receipt:
        raw=a.native_receipt.read_bytes()
        receipt=json.loads(gzip.decompress(raw) if a.native_receipt.suffix=='.gz' else raw)
    else:
        print('PASS complete871 registry and own285 mechanical bindings; fresh native proof remains pending')
        return
    if a.collect_native:
        encoded=json.dumps(receipt,indent=2).encode()
        a.native_receipt.write_bytes(gzip.compress(encoded,mtime=0) if a.native_receipt.suffix=='.gz' else encoded)
    proof=validate_live(receipt,registry,inputs)
    proof['native_receipt_sha256']=digest(a.native_receipt.read_bytes())
    if a.proof_output:a.proof_output.write_text(json.dumps(proof,indent=2)+'\n')
    print('PASS own285/global871 actual native identities/bodies/parents/edges/TaskSize; full foreign graph, both DAGs/caps/old protections; final prefixes HELD, runtime UNQUALIFIED')

if __name__=='__main__':main()
