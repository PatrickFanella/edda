"""Independent read-only binding to ROOT's finite publication approval.

This module never creates issues or adopts a live plan. Approval identifies the
exact frozen templates; the source oracle still checks the actual Warlock data.
Registry adaptation and final-body rendering are kept separate from approval.
"""
import hashlib
import json
from pathlib import Path

from verify_warlock_scope import ROOT, exact, caps

APPROVAL_SHA256 = '66b934f5cac58d21644822bd36098b0ee1cad535708a9e3f6b49bdf0a2f0fe41'
OWN_PROPOSAL_SHA256 = '8d4a8ea724ee8e997093d0e3877054a1c4a499f1f00a059bcbb9d2d51bc4268a'
RENDERER_SHA256 = 'e466e1b41506c35f6f03b6c3cc8d2bbbfb57e168607fa3c63954291930247238'
VECTORS_SHA256 = '4ba7f97dd6f1702ff049056f10e46736cb1da27fe7274553b4bff9e54077bdee'
MILESTONE_SHA256 = '58c5aacb2a3e73b759dff6b98316661fb6bd9870b95e5c7fa9a0d39e7ee736ba'
DIRECTORY = ROOT / 'docs/investigations'
FILES = {330: 'issue-331-sorcerer-boundary.json', 331: 'issue-331-native-proposal.json',
         332: 'issue-331-wizard-boundary.json'}


def digest(raw):
    return hashlib.sha256(raw).hexdigest()


def strict_json(raw):
    def unique_pairs(pairs):
        result = {}
        for key, value in pairs:
            assert key not in result, ('duplicate JSON field', key)
            result[key] = value
        return result
    def reject_constant(value):
        raise AssertionError(('nonfinite JSON value', value))
    return json.loads(raw, object_pairs_hook=unique_pairs, parse_constant=reject_constant)


def reference(value, owner):
    if type(value) is int:
        assert value > 0, 'positive native reference'
        return str(value)
    assert type(value) is str and value, 'typed reference'
    if value.isascii() and value.isdecimal():
        assert str(int(value)) == value and int(value) > 0, 'canonical native reference'
        return value
    aliases = {'warlock-shared-output': '331:shared-output',
               'wizard-shared-output': '332:shared-output'}
    return aliases.get(value, value if ':' in value else f'{owner}:{value}')


def frozen_rows(proposals):
    result = {}
    for owner, proposal in proposals.items():
        rows = proposal['tasks'] + proposal['gates']
        if owner == 330:
            rows += proposal['shared_split_proposal']
        for row in rows:
            alias = f"{owner}:{row['key']}"
            assert alias not in result, 'duplicate frozen alias'
            deps = [reference(n, owner) for n in row['prerequisites']]
            deps += [reference(d, owner) for d in row['depends']]
            assert len(deps) == len(set(deps)), ('duplicate prerequisite', alias)
            result[alias] = dict(alias=alias, owner_decision=owner, key=row['key'],
                title=row['title'], type=row['type'], size=row['size'],
                parent_ref=reference(row['parent'], owner), blocked_by_refs=sorted(deps),
                approved_body=row.get('native_body', row['body']))
    return result


def validate_approval(approval, proposals, baseline):
    assert approval['status'] == 'APPROVED_FINITE_NATIVE_PUBLICATION'
    assert exact(approval['counts'], {'330': 305, '331': 281, '332': 285})
    assert type(approval['new_nodes']) is int and approval['new_nodes'] == 871
    assert approval['coordinator'] == 330 and approval['source_pdf_sha256'] == '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
    expected = frozen_rows(proposals)
    actual = {n['alias']: n for n in approval['approved_nodes']}
    assert len(actual) == len(approval['approved_nodes']) == 871
    assert expected.keys() == actual.keys(), 'complete finite approved roster'
    for alias, row in expected.items():
        approved = actual[alias]
        for field, value in row.items():
            assert exact(approved[field], value), ('independent frozen field', alias, field)
        assert approved['approved_body_sha256'] == digest(row['approved_body'].encode()), ('complete approved body', alias)
        assert row['type'] == 'Task' and row['size'] == 'M'
    old = {str(n['number']): n for n in baseline['graph']}
    assert len(old) == len(baseline['graph']) == 1017
    graph = {key: [str(d) for d in n['blocked_by']] for key, n in old.items()}
    parents = {key: str(n['parent']) for key, n in old.items() if n['parent'] is not None}
    for alias, row in expected.items():
        assert alias not in graph
        graph[alias] = row['blocked_by_refs']
        parents[alias] = row['parent_ref']
    additions = {(e['issue'], e['blocker'], e['writer']) for e in approval['allowed_old_root_edges']}
    required = {('337', f'{n}:class-gate', 330) for n in (330, 331, 332)} | {
        ('279', '330:shared-aggregate', 330), ('280', '330:shared-aggregate', 330)}
    assert additions == required and len(additions) == len(approval['allowed_old_root_edges']) == 5
    for issue, blocker, _ in additions:
        assert blocker not in graph[issue], 'old edge must not be silently replaced'
        graph[issue] = graph[issue] + [blocker]
    assert all(d in graph for deps in graph.values() for d in deps), 'unbound dependency'
    assert all(p in graph for p in parents.values()), 'unbound parent'
    caps(graph, parents)
    assert len(graph) == 1888 and graph['183'] == [str(d) for d in old['183']['blocked_by']]
    assert parents['330:shared-aggregate'] == '191'
    assert set(graph['330:shared-aggregate']) == {f'{n}:shared-output' for n in (330, 331, 332)}
    for n in (330, 331, 332):
        for target in (f'{n}:class-gate', f'{n}:shared-output'):
            assert target in graph
        assert not any(f'{c}:class-gate' in reachable(graph, f'{n}:shared-output') for c in (330, 331, 332)), 'shared output consumes admission gate'
    own = {key for key in expected if key.startswith('331:')}
    assert own <= reachable(graph, '337'), 'all281 mandatory after decision closure'
    shared = {key for key, row in expected.items() if key.startswith('331:') and row['parent_ref'].startswith('331:shared-')}
    for root in ('279', '280', '337'):
        assert shared <= reachable(graph, root), ('all actual shared leaves mandatory', root)
    assert sum(p == '283' for p in parents.values()) == 95
    assert sum('429' in deps for deps in graph.values()) == 100
    assert all('429' not in row['blocked_by_refs'] and row['parent_ref'] != '283' for row in expected.values() if row['owner_decision'] == 331)
    return expected, graph, parents


def reachable(graph, start):
    seen, todo = set(), [start]
    while todo:
        key = todo.pop()
        if key not in seen:
            seen.add(key)
            todo.extend(graph.get(key, []))
    return seen


def load_frozen(directory=DIRECTORY):
    directory = Path(directory)
    raw = (directory / 'issue-331-publication-approval.json').read_bytes()
    assert digest(raw) == APPROVAL_SHA256, 'exact ROOT approval pin; arbitrary approval hashes forbidden'
    approval = json.loads(raw)
    proposals = {}
    for owner, filename in FILES.items():
        data = (directory / filename).read_bytes()
        assert digest(data) == approval['inputs'][str(owner)]['sha256'], ('exact approved input', owner)
        proposals[owner] = json.loads(data)
    assert approval['inputs']['331']['sha256'] == OWN_PROPOSAL_SHA256
    raw_baseline = (directory / 'issue-331-native-baseline.json').read_bytes()
    assert digest(raw_baseline) == approval['native_baseline_sha256'], 'exact approved baseline; not a final prefix release'
    baseline = json.loads(raw_baseline)
    expected, graph, parents = validate_approval(approval, proposals, baseline)
    return approval, proposals, baseline, expected, graph, parents


def validate_identities(identities, expected, baseline):
    """Validate normalized registry identities; never allocate or write them."""
    assert identities.keys() == expected.keys(), 'all871 registry aliases required'
    old_numbers = {n['number'] for n in baseline['graph']}
    for alias, identity in identities.items():
        assert set(identity) == {'number', 'id', 'node_id'}, ('registry identity fields', alias)
        assert type(identity['number']) is int and identity['number'] > 0
        assert type(identity['id']) is int and identity['id'] > 0
        assert type(identity['node_id']) is str and identity['node_id'], 'nonempty node identity'
        assert identity['number'] not in old_numbers, 'registry points at old owner'
    for field in ('number', 'id', 'node_id'):
        values = [x[field] for x in identities.values()]
        assert len(set(values)) == 871, ('duplicate identity', field)
    return {alias: value['number'] for alias, value in identities.items()}


def resolved_structure(expected, numbers):
    def number(ref):
        return int(ref) if ref.isascii() and ref.isdecimal() else numbers[ref]
    return {numbers[alias]: dict(parent=number(row['parent_ref']),
        blocked_by=sorted(number(d) for d in row['blocked_by_refs'])) for alias, row in expected.items()}


ALLOCATION_PREFIX = '<!-- edda-caster-native:{alias} -->\n'
FINAL_SUFFIX = ('\n\n<!-- edda-caster-native-binding:start -->\n'
    'Decision: #{owner_decision}\nSource key: {alias}\nNative issue: #{number}\n'
    'Native parent: #{parent_number}\nNative prerequisites: {prerequisites}\n'
    'Root approval SHA256: {approval_sha256}\n'
    'Renderer contract SHA256: {renderer_contract_sha256}\n'
    'Publication: source contract published; runtime unqualified.\n'
    '<!-- edda-caster-native-binding:end -->\n')


def load_renderer(directory=DIRECTORY):
    raw = (Path(directory) / 'issue-331-renderer-contract.json').read_bytes()
    assert digest(raw) == RENDERER_SHA256, 'exact ROOT renderer pin'
    contract = json.loads(raw)
    assert contract['approval_sha256'] == APPROVAL_SHA256
    assert contract['allocation_prefix_template'] == ALLOCATION_PREFIX
    assert contract['final_suffix_template'] == FINAL_SUFFIX, 'independent literal renderer contract'
    return contract


def load_milestone_contract(directory=DIRECTORY):
    raw = (Path(directory) / 'issue-331-milestone-contract.json').read_bytes()
    assert digest(raw) == MILESTONE_SHA256
    contract = json.loads(raw)
    assert contract['approval_sha256'] == APPROVAL_SHA256
    assert contract['renderer_contract_sha256'] == RENDERER_SHA256 and contract['golden_vectors_sha256'] == VECTORS_SHA256
    return contract


def validate_registry(registry, expected, baseline):
    required = {'schema_version', 'approval_sha256', 'renderer_contract_sha256', 'status', 'nodes'}
    optional = {'updated_at', 'observed_at'}
    assert required <= registry.keys() <= required | optional, 'exact registry envelope'
    assert type(registry['schema_version']) is int and registry['schema_version'] == 1
    assert registry['approval_sha256'] == APPROVAL_SHA256 and registry['renderer_contract_sha256'] == RENDERER_SHA256
    assert registry['status'] in ('IDS_ALLOCATED', 'WIRING', 'VERIFIED'), 'partial/ALLOCATING registry cannot qualify'
    assert type(registry['nodes']) is dict
    return validate_identities(registry['nodes'], expected, baseline)


def allocation_body(node):
    return ALLOCATION_PREFIX.format(alias=node['alias']) + node['approved_body']


def _format_final_body(node, numbers):
    """Pure formatting helper; partial golden vectors test only this function."""
    def number(ref):
        assert type(ref) is str and ref
        return int(ref) if ref.isascii() and ref.isdecimal() else numbers[ref]
    deps = [number(d) for d in node['blocked_by_refs']]
    assert len(deps) == len(set(deps)), 'duplicate resolved prerequisite'
    suffix = FINAL_SUFFIX.format(owner_decision=node['owner_decision'], alias=node['alias'],
        number=numbers[node['alias']], parent_number=number(node['parent_ref']),
        prerequisites=', '.join('#' + str(d) for d in sorted(deps)) if deps else 'none',
        approval_sha256=APPROVAL_SHA256, renderer_contract_sha256=RENDERER_SHA256)
    return allocation_body(node) + suffix


def final_bodies(registry, expected, baseline):
    numbers = validate_registry(registry, expected, baseline)
    return {alias: _format_final_body(node, numbers) for alias, node in expected.items()}


def validate_allocation(approval, expected, baseline, registry, snapshot):
    """Bind real IDs during the writer's transition; never claim wiring PASS."""
    numbers = validate_registry(registry, expected, baseline)
    structure = resolved_structure(expected, numbers)
    bodies = final_bodies(registry, expected, baseline)
    graph = {n['number']: n for n in snapshot['graph']}
    assert len(graph) == len(snapshot['graph'])
    _graph_types(graph)
    caps({n: t['blocked_by'] for n, t in graph.items()},
         {n: t['parent'] for n, t in graph.items() if t['parent'] is not None})
    for alias, node in expected.items():
        actual = graph[numbers[alias]]
        identity = registry['nodes'][alias]
        assert type(actual['id']) is int and type(actual['node_id']) is str
        assert actual['id'] == identity['id'] and actual['node_id'] == identity['node_id']
        assert actual['title'] == node['title'] and actual['state'] == 'OPEN'
        assert actual['body'] in (allocation_body(node), bodies[alias]), ('exact transitional body', alias)
        assert actual['parent'] in (None, structure[numbers[alias]]['parent'])
        assert len(actual['blocked_by']) == len(set(actual['blocked_by']))
        assert set(actual['blocked_by']) <= set(structure[numbers[alias]]['blocked_by'])
        assert snapshot['project'][str(numbers[alias])] == dict(type='Task', size='M')
    _old_protection(approval, baseline, graph, snapshot, numbers)
    return dict(status='IDS_BOUND_WIRING_NOT_QUALIFIED', own_nodes=281, allocated_nodes=871,
                wiring_verified=False, final_prefix_sealed=False, runtime_qualified=False)


def _old_protection(approval, baseline, graph, snapshot, numbers):
    new_ids = set(numbers.values())
    for old in baseline['graph']:
        actual = graph[old['number']]
        assert actual['title'] == old['title'] and actual['parent'] == old['parent'], 'old title/parent'
        if old['number'] in (330, 331, 332):
            assert actual['state'].upper() in ('OPEN', 'CLOSED'), 'decision state only'
        else:
            assert actual['state'].upper() == old['state'].upper(), 'old state drift'
        if old['number'] in (279, 280, 337):
            assert set(old['blocked_by']) <= set(actual['blocked_by']), 'old required root edge removed'
            permitted = {numbers[e['blocker']] for e in approval['allowed_old_root_edges'] if int(e['issue']) == old['number']}
            assert not ((set(actual['blocked_by']) - set(old['blocked_by']) - permitted) & new_ids), 'unapproved owned addition to old root'
        else:
            assert exact(actual['blocked_by'], old['blocked_by']), 'old owner prerequisite mutation'
        if old['number'] in (47, 54, 283, 337):
            assert actual['body'].startswith(old['body']), 'historical prefix rollback; not a final seal'
        else:
            assert actual['body'] == old['body'], 'old whole body mutation'
    for number, metadata in baseline['project'].items():
        assert exact(snapshot['project'][number], metadata), 'old Project/Task/Size mutation'


def _graph_types(graph):
    for number, row in graph.items():
        assert type(number) is int and number > 0, 'actual graph number type'
        assert row['parent'] is None or (type(row['parent']) is int and row['parent'] > 0), 'actual graph parent type'
        assert type(row['blocked_by']) is list
        assert all(type(n) is int and n > 0 for n in row['blocked_by']), 'actual graph prerequisite type'
        assert row['blocked_by'] == sorted(set(row['blocked_by'])), 'actual graph sorted unique prerequisites'
        assert row['parent'] is None or row['parent'] in graph, 'missing actual containment endpoint'
        assert all(n in graph for n in row['blocked_by']), 'missing actual prerequisite endpoint'


def validate_wiring(approval, expected, baseline, registry, snapshot):
    """Complete native publication proof, explicitly before any prefix seal."""
    numbers = validate_registry(registry, expected, baseline)
    assert registry['status'] == 'VERIFIED', 'wiring checkpoint requires coordinator VERIFIED boundary'
    bodies = final_bodies(registry, expected, baseline)
    structure = resolved_structure(expected, numbers)
    graph = {n['number']: n for n in snapshot['graph']}
    assert len(graph) == len(snapshot['graph']), 'duplicate actual native identity'
    _graph_types(graph)
    caps({n: t['blocked_by'] for n, t in graph.items()},
         {n: t['parent'] for n, t in graph.items() if t['parent'] is not None})
    for alias, node in expected.items():
        identity = registry['nodes'][alias]
        actual = graph[numbers[alias]]
        assert type(actual['id']) is int and type(actual['node_id']) is str
        assert actual['id'] == identity['id'] and actual['node_id'] == identity['node_id'], ('actual Issue identity', alias)
        assert actual['title'] == node['title'] and actual['body'] == bodies[alias], ('actual complete rendered body/title', alias)
        assert actual['parent'] == structure[numbers[alias]]['parent']
        assert exact(actual['blocked_by'], structure[numbers[alias]]['blocked_by']), ('exact new-node prerequisites', alias)
        assert actual['state'] == 'OPEN', 'all871 future Tasks must remain OPEN'
        assert snapshot['project'][str(numbers[alias])] == dict(type='Task', size='M')
    _old_protection(approval, baseline, graph, snapshot, numbers)
    for edge in approval['allowed_old_root_edges']:
        assert numbers[edge['blocker']] in graph[int(edge['issue'])]['blocked_by'], 'mandatory actual root path'
    deps = {n: t['blocked_by'] for n, t in graph.items()}
    for n in (330, 331, 332):
        shared = reachable(deps, numbers[f'{n}:shared-output'])
        assert all(numbers[f'{c}:class-gate'] not in shared for c in (330, 331, 332)), 'shared/admission cycle or backlink'
    own = {numbers[a] for a in expected if a.startswith('331:')}
    assert own <= reachable(deps, 337), 'all281 remain mandatory after331closure'
    for alias, node in expected.items():
        if alias.startswith('331:') and node['parent_ref'].startswith('331:shared-'):
            assert all(numbers[alias] in reachable(deps, root) for root in (279, 280, 337)), 'shared leaf mandatory paths'
    assert len(graph[numbers['330:shared-aggregate']]['blocked_by']) == 3
    assert sum(t['parent'] == 283 for t in graph.values()) == 95
    assert sum(429 in t['blocked_by'] for t in graph.values()) == 100
    assert graph[331]['state'] in ('OPEN', 'CLOSED'), 'decision closure is allowed, implementation closure is not'
    return dict(status='NATIVE_WIRING_PASS_PREFIX_PENDING', own_nodes=281, published_nodes=871,
                final_prefix_sealed=False, runtime_qualified=False)


def nodes_map_digest(registry, expected, baseline):
    validate_registry(registry, expected, baseline)
    raw = json.dumps(registry['nodes'], sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode('utf-8')
    return digest(raw)


def validate_milestone(registry_raw, delivery, expected, baseline, prior_allocation=None, verified_receipt_raw=None, prior_allocation_raw=None):
    """Provenance is necessary but never sufficient for actual native proof."""
    registry = strict_json(registry_raw)
    stable = nodes_map_digest(registry, expected, baseline)
    required = {'milestone', 'observed_at', 'approval_sha256', 'renderer_contract_sha256',
                'registry_path', 'registry_file_sha256', 'nodes_map_sha256', 'node_count', 'class_counts'}
    assert required <= delivery.keys(), 'complete delivered milestone provenance'
    assert delivery['milestone'] in ('IDS_ALLOCATED', 'VERIFIED')
    assert delivery['approval_sha256'] == APPROVAL_SHA256 and delivery['renderer_contract_sha256'] == RENDERER_SHA256
    assert type(delivery['node_count']) is int and delivery['node_count'] == 871
    assert exact(delivery['class_counts'], {'330': 305, '331': 281, '332': 285})
    assert type(delivery['observed_at']) is str and delivery['observed_at']
    assert type(delivery['registry_path']) is str and delivery['registry_path']
    mutable = '/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/registry.json'
    assert Path(delivery['registry_path']).resolve() != Path(mutable).resolve(), 'immutable delivered snapshot required'
    assert delivery['registry_file_sha256'] == digest(registry_raw), 'independently recomputed exact registry bytes'
    assert delivery['nodes_map_sha256'] == stable, 'independently recomputed stable identities'
    if delivery['milestone'] == 'IDS_ALLOCATED':
        assert registry['status'] == 'IDS_ALLOCATED', 'preserved allocation snapshot status'
    else:
        additional = {'verified_receipt_path', 'verified_receipt_sha256', 'verified_receipt_observed_at',
                      'graph_node_count', 'all_new_tasks_open', 'native_body_parent_prerequisite_Task_Size_validation',
                      'both_DAGs', 'all_three_caps', 'old_protections'}
        assert additional <= delivery.keys(), 'complete actual collection provenance'
        assert registry['status'] == 'VERIFIED' and prior_allocation is not None
        assert prior_allocation_raw is not None, 'reread preserved allocation bytes, never trust only a sidecar'
        validate_milestone(prior_allocation_raw, prior_allocation, expected, baseline)
        assert prior_allocation['milestone'] == 'IDS_ALLOCATED'
        assert prior_allocation['nodes_map_sha256'] == stable, 'no reassignment across milestones'
        assert verified_receipt_raw is not None and delivery['verified_receipt_sha256'] == digest(verified_receipt_raw)
        assert type(delivery['verified_receipt_observed_at']) is str and delivery['verified_receipt_observed_at']
        assert type(delivery['verified_receipt_path']) is str and delivery['verified_receipt_path']
        assert type(delivery['graph_node_count']) is int and delivery['graph_node_count'] >= 1888
        for field in ('all_new_tasks_open', 'native_body_parent_prerequisite_Task_Size_validation',
                      'both_DAGs', 'all_three_caps', 'old_protections'):
            assert delivery[field] is True, ('coordinator milestone did not pass', field)
    return registry
