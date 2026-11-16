"""Independent read-only binding of #330's source oracle to ROOT's publication.

This deliberately reproduces the literal metadata renderer rather than importing
the coordinator renderer or accepting candidate-supplied expected body hashes.
"""
from collections import Counter
from functools import lru_cache
import gzip
import hashlib
import json
from pathlib import Path
import re

DIRECTORY = Path('/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication')
APPROVAL = '66b934f5cac58d21644822bd36098b0ee1cad535708a9e3f6b49bdf0a2f0fe41'
RENDERER = 'e466e1b41506c35f6f03b6c3cc8d2bbbfb57e168607fa3c63954291930247238'
VECTORS = '4ba7f97dd6f1702ff049056f10e46736cb1da27fe7274553b4bff9e54077bdee'

BUNDLE = Path(__file__).resolve().parents[1]/'docs/investigations/issue-330-caster-publication-evidence.json.gz'


@lru_cache(maxsize=1)
def bundled_evidence():
    with gzip.open(BUNDLE,'rb') as stream:
        raw=stream.read(64*1024*1024+1)
    assert len(raw)<=64*1024*1024, 'bounded offline evidence bundle'
    bundle=json.loads(raw)
    assert bundle['schema_version']==1 and bundle['approval_sha256']==APPROVAL
    return bundle['files']


def evidence(name, directory=DIRECTORY):
    path=directory/name
    if path.exists():
        return path.read_bytes()
    assert directory==DIRECTORY and BUNDLE.exists(), 'missing ROOT publication evidence: '+str(path)
    return bundled_evidence()[name].encode('utf-8')


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def pinned_inputs(directory=DIRECTORY):
    raw = evidence('approval.json',directory)
    assert digest(raw) == APPROVAL, 'independently pinned ROOT approval'
    approval = json.loads(raw)
    assert digest(evidence('native-renderer-contract.json',directory)) == RENDERER
    assert digest(evidence('native-renderer-vectors.json',directory)) == VECTORS
    for source in approval['inputs'].values():
        assert digest(evidence(Path(source['frozen_path']).name,directory)) == source['sha256'], 'frozen source bytes'
    baseline = evidence('native-baseline.json',directory)
    assert digest(baseline) == approval['native_baseline_sha256']
    return approval, json.loads(baseline)


def registry_check(registry, approval, baseline):
    assert set(registry) >= {'schema_version','approval_sha256','renderer_contract_sha256','status','nodes'}
    assert set(registry) <= {'schema_version','approval_sha256','renderer_contract_sha256','status','nodes','observed_at','updated_at'}
    assert type(registry['schema_version']) is int and registry['schema_version'] == 1
    assert registry['approval_sha256'] == APPROVAL and registry['renderer_contract_sha256'] == RENDERER
    assert registry['status'] in ['IDS_ALLOCATED','WIRING','VERIFIED']
    nodes = registry['nodes']
    assert set(nodes) == {n['alias'] for n in approval['approved_nodes']} and len(nodes) == 871
    assert Counter(a.split(':',1)[0] for a in nodes) == {'330':305,'331':281,'332':285}
    old = {n['number'] for n in baseline['graph']}
    for value in nodes.values():
        assert set(value) == {'number','id','node_id'}
        assert type(value['number']) is int and value['number'] > 0 and value['number'] not in old
        assert type(value['id']) is int and value['id'] > 0
        assert type(value['node_id']) is str and value['node_id']
    for key in ['number','id','node_id']:
        assert len({v[key] for v in nodes.values()}) == 871, 'unique native identities'
    return digest(json.dumps(nodes,sort_keys=True,separators=(',',':'),ensure_ascii=False).encode('utf-8'))


def number(ref, registry):
    assert type(ref) is str
    if re.fullmatch('[0-9]+',ref,flags=re.ASCII):
        return int(ref)
    return registry['nodes'][ref]['number']


def reproduce(node, registry):
    deps = [number(ref,registry) for ref in node['blocked_by_refs']]
    assert len(deps) == len(set(deps)), 'distinct resolved prerequisites'
    prerequisites = ', '.join('#'+str(n) for n in sorted(deps)) or 'none'
    return ('<!-- edda-caster-native:'+node['alias']+' -->\n'+node['approved_body']
            +'\n\n<!-- edda-caster-native-binding:start -->\n'
            +'Decision: #'+str(node['owner_decision'])+'\n'
            +'Source key: '+node['alias']+'\n'
            +'Native issue: #'+str(registry['nodes'][node['alias']]['number'])+'\n'
            +'Native parent: #'+str(number(node['parent_ref'],registry))+'\n'
            +'Native prerequisites: '+prerequisites+'\n'
            +'Root approval SHA256: '+APPROVAL+'\n'
            +'Renderer contract SHA256: '+RENDERER+'\n'
            +'Publication: source contract published; runtime unqualified.\n'
            +'<!-- edda-caster-native-binding:end -->\n')


def bind_own_rows(manifest, approval, registry):
    """Every own semantic row must agree with frozen source and native operands."""
    rows = manifest['tasks']+manifest['shared_split_proposal']+manifest['gates']
    own = {n['key']:n for n in approval['approved_nodes'] if n['owner_decision']==330}
    assert len(rows) == len(own) == 305 and {r['key'] for r in rows} == set(own)
    foreign = {'warlock-shared-output':'331:shared-output','wizard-shared-output':'332:shared-output'}
    def alias(ref):
        if type(ref) is int:
            return str(ref)
        return foreign.get(ref,'330:'+ref)
    for row in rows:
        frozen = own[row['key']]
        assert row['body'] == frozen['approved_body'] and row['title'] == frozen['title'], 'own frozen semantic envelope'
        assert alias(row['parent']) == frozen['parent_ref']
        assert sorted([str(n) for n in row['prerequisites']]+[alias(d) for d in row['depends']]) == sorted(frozen['blocked_by_refs']), 'own exact semantic native operands'
        assert type(row['issue']) is int and row['issue'] == registry['nodes'][frozen['alias']]['number']
    return own


def validate_native(manifest, receipt, directory=DIRECTORY, parent_supplements=False):
    from verify_sorcerer_scope import native_caps, typed_native_graph
    approval, baseline = pinned_inputs(directory)
    registry = json.loads(evidence('registry.json',directory))
    map_sha = registry_check(registry, approval, baseline)
    assert registry['status'] == 'VERIFIED', 'global mechanical readback incomplete'
    milestone = json.loads(evidence('ids-allocated-milestone.json',directory))
    assert milestone['nodes_map_sha256'] == map_sha, 'identity assignment changed after allocation'
    assert milestone['approval_sha256'] == APPROVAL and milestone['renderer_contract_sha256'] == RENDERER
    return validate_evidence(manifest, receipt, approval, baseline, registry, parent_supplements)


def validate_evidence(manifest, receipt, approval, baseline, registry, parent_supplements=False):
    from verify_sorcerer_scope import native_caps, typed_native_graph
    map_sha = registry_check(registry, approval, baseline)
    bind_own_rows(manifest, approval, registry)
    graph = typed_native_graph(receipt)
    assert len(graph) >= 1888
    old_by_number = {n['number']:n for n in baseline['graph']}
    new_by_number = {v['number']:a for a,v in registry['nodes'].items()}
    assert set(old_by_number)|set(new_by_number) <= set(graph), 'complete frozen actual graph identities'
    for old in baseline['graph']:
        actual = graph[old['number']]
        for field in ['title','body','parent','state']:
            if field == 'state' and old['number'] in [330,331,332]:
                assert actual[field] in ['open','closed'], 'decision state; mandatory gates remain below'
                continue
            if field == 'body' and parent_supplements and str(old['number']) in ['47','54','283','337']:
                from verify_sorcerer_scope import prefix_check, ACCEPTED_PREFIX_SHA256
                key = str(old['number'])
                seal = manifest['accepted_parent_prefix'][key]
                assert seal['prefix'] == old['body'] and digest(seal['prefix'].encode()) == ACCEPTED_PREFIX_SHA256[key]
                section = manifest['parent_appends'][key]
                prefix_check(actual['body'],seal['prefix'],section['text'],section['marker'])
                continue
            assert actual[field] == old[field], 'old contract preservation'
        extra = [number(e['blocker'],registry) for e in approval['allowed_old_root_edges'] if int(e['issue'])==old['number']]
        required = sorted(old['blocked_by']+extra)
        if old['number'] in [279,280,337]:
            assert set(required) <= set(actual['blocked_by']), 'every frozen own/historical root obligation'
        else:
            assert actual['blocked_by'] == required, 'no other old prerequisite mutation, including183'
    for n in approval['approved_nodes']:
        identity = registry['nodes'][n['alias']]; actual = graph[identity['number']]
        assert type(actual['id']) is int and type(actual['node_id']) is str, 'exact actual native identity types'
        assert actual['id'] == identity['id'] and actual['node_id'] == identity['node_id'], 'actual native ID binding'
        assert actual['state'] == 'open' and actual['title'] == n['title']
        assert actual['body'] == reproduce(n, registry), 'independent full native body reproduction'
        assert actual['parent'] == number(n['parent_ref'],registry)
        assert actual['blocked_by'] == sorted(number(d,registry) for d in n['blocked_by_refs'])
        assert receipt['project'][str(identity['number'])] == dict(type='Task',size='M')
    for n, expected in baseline['project'].items():
        assert receipt['project'][n] == expected, 'old Task/Project metadata preservation'
    assert all(d in graph for t in graph.values() for d in t['blocked_by']), 'complete actual dependency graph'
    assert all(t['parent'] is None or t['parent'] in graph for t in graph.values()), 'complete actual containment graph'
    native_caps({n:t['blocked_by'] for n,t in graph.items()},{n:t['parent'] for n,t in graph.items() if t['parent']})
    for decision in [330,331,332]:
        shared = registry['nodes'][str(decision)+':shared-output']['number']
        consuming = registry['nodes'][str(decision)+':class-gate']['number']
        assert consuming not in graph[shared]['blocked_by']
    assert graph[183]['blocked_by'] == old_by_number[183]['blocked_by']
    return dict(approval_sha256=APPROVAL, renderer_contract_sha256=RENDERER, nodes_map_sha256=map_sha,
                global_nodes=len(graph), own_nodes=305, native='PASS; all871 exact source contracts Task/M OPEN',
                future_policy='Frozen nonempty source heads/old requirements; only separately ROOT-approved source-bound monotonic extensions with manifest/head/native receipt. Publication may precede decision acceptance; no incomplete head qualifies.')
