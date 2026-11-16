#!/usr/bin/env python3
"""Single-writer, resumable mechanical publication of ROOT's frozen caster plan.

No source generation, issue closure, old contract edits or parent supplements.
Every mutation has a durable intent and a read-before-write recovery boundary.
"""
import argparse
from collections import Counter, deque
from datetime import datetime, timezone
import fcntl
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys

APPROVAL_SHA = '66b934f5cac58d21644822bd36098b0ee1cad535708a9e3f6b49bdf0a2f0fe41'
RENDERER_SHA = 'e466e1b41506c35f6f03b6c3cc8d2bbbfb57e168607fa3c63954291930247238'
DEFAULT = Path('/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication')
REPO = 'R_kgDOSicKgw'
PROJECT = 'PVT_kwDODjtmGM4BrFJ2'
TASK = 'IT_kwDODjtmGM4BwPAl'
SIZE_FIELD = 'PVTSSF_lADODjtmGM4BrFJ2zhpZbzc'
SIZE_M = '51f0db86'


def sha(value):
    return hashlib.sha256(value.encode() if isinstance(value, str) else value).hexdigest()


def now():
    return datetime.now(timezone.utc).isoformat()


def atomic(path, value):
    temporary = path.with_suffix(path.suffix + '.tmp')
    with temporary.open('w') as stream:
        json.dump(value, stream, indent=2, ensure_ascii=False)
        stream.write('\n')
        stream.flush()
        os.fsync(stream.fileno())
    os.replace(temporary, path)
    descriptor = os.open(path.parent, os.O_RDONLY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)


def resolve(ref, registry):
    assert isinstance(ref, str)
    return int(ref) if re.fullmatch(r'[0-9]+', ref, flags=re.ASCII) else registry['nodes'][ref]['number']


def allocation_body(node, contract):
    return contract['allocation_prefix_template'].format(alias=node['alias']) + node['approved_body']


def final_body(node, registry, contract):
    deps = [resolve(ref, registry) for ref in node['blocked_by_refs']]
    assert len(deps) == len(set(deps)), 'duplicate resolved prerequisite'
    return allocation_body(node, contract) + contract['final_suffix_template'].format(
        owner_decision=node['owner_decision'], alias=node['alias'],
        number=registry['nodes'][node['alias']]['number'],
        parent_number=resolve(node['parent_ref'], registry),
        prerequisites=', '.join('#' + str(n) for n in sorted(deps)) or 'none',
        approval_sha256=APPROVAL_SHA, renderer_contract_sha256=RENDERER_SHA)


def check_registry(registry, nodes, baseline, complete):
    assert set(registry) <= {'schema_version', 'approval_sha256', 'renderer_contract_sha256', 'status', 'nodes', 'updated_at', 'observed_at'}
    assert type(registry['schema_version']) is int and registry['schema_version'] == 1
    assert registry['approval_sha256'] == APPROVAL_SHA and registry['renderer_contract_sha256'] == RENDERER_SHA
    assert registry['status'] in ['ALLOCATING', 'IDS_ALLOCATED', 'WIRING', 'VERIFIED']
    aliases = {n['alias'] for n in nodes}
    assert set(registry['nodes']) <= aliases
    if complete:
        assert set(registry['nodes']) == aliases and len(aliases) == 871, 'complete 871 identities required'
    old = {n['number'] for n in baseline['graph']}
    for value in registry['nodes'].values():
        assert set(value) == {'number', 'id', 'node_id'}
        assert all(type(value[k]) is int and value[k] > 0 for k in ['number', 'id']), 'exact positive integer identity'
        assert isinstance(value['node_id'], str) and value['node_id']
        assert value['number'] not in old
    for key in ['number', 'id', 'node_id']:
        values = [v[key] for v in registry['nodes'].values()]
        assert len(set(values)) == len(values), 'duplicate registry ' + key


def uncertain_creation(journal, alias):
    # Called only when neither the registry nor the complete marker readback has
    # this alias. Even an ACK can precede a crash or contain issue:null. A prior
    # creation intent is never permission for a second blind create.
    return any(entry['operation']=='create '+alias for entry in journal['operations'])


def caps(graph, parents):
    assert all(len(v) == len(set(v)) and len(v) <= 100 for v in graph.values())
    assert all(d in graph for deps in graph.values() for d in deps), 'missing dependency'
    assert all(n in graph and p in graph for n, p in parents.items()), 'missing parent'
    outgoing = Counter(d for deps in graph.values() for d in deps)
    children = Counter(parents.values())
    assert max(outgoing.values(), default=0) <= 100 and max(children.values(), default=0) <= 100
    for label, edges in [('dependency', graph), ('containment', {n: [parents[n]] if n in parents else [] for n in graph})]:
        incoming = {n: len(ds) for n, ds in edges.items()}
        reverse = {n: [] for n in edges}
        for n, ds in edges.items():
            for d in ds:
                reverse[d].append(n)
        queue = deque(n for n, count in incoming.items() if not count)
        count = 0
        while queue:
            n = queue.popleft(); count += 1
            for dependent in reverse[n]:
                incoming[dependent] -= 1
                if not incoming[dependent]:
                    queue.append(dependent)
        assert count == len(edges), label + ' DAG cycle'
    return dict(nodes=len(graph), max_incoming=max(map(len, graph.values())),
                max_outgoing=max(outgoing.values(), default=0), max_children=max(children.values(), default=0),
                incoming279=len(graph['279']), incoming280=len(graph['280']), incoming337=len(graph['337']),
                outgoing429=outgoing['429'], outgoing285=outgoing['285'], children283=children['283'])


class Coordinator:
    def __init__(self, directory):
        self.directory = directory
        raw = (directory / 'approval.json').read_bytes()
        assert sha(raw) == APPROVAL_SHA, 'root approval pin'
        self.approval = json.loads(raw)
        raw = (directory / 'native-renderer-contract.json').read_bytes()
        assert sha(raw) == RENDERER_SHA, 'root renderer pin'
        self.contract = json.loads(raw)
        self.nodes = self.approval['approved_nodes']
        assert len(self.nodes) == 871 and len({n['alias'] for n in self.nodes}) == 871
        assert Counter(str(n['owner_decision']) for n in self.nodes) == {'330': 305, '331': 281, '332': 285}
        self.verify_inputs()
        self.baseline = json.loads(Path(self.approval['native_baseline']).read_text())
        assert len(self.baseline['graph']) == 1017
        self.registry_path = Path(self.approval['registry'])
        self.journal_path = Path(self.approval['journal'])
        self.registry = (json.loads(self.registry_path.read_text()) if self.registry_path.exists() else
                         dict(schema_version=1, approval_sha256=APPROVAL_SHA, renderer_contract_sha256=RENDERER_SHA,
                              status='ALLOCATING', nodes={}))
        self.journal = (json.loads(self.journal_path.read_text()) if self.journal_path.exists() else
                        dict(schema_version=1, approval_sha256=APPROVAL_SHA, operations=[]))
        assert self.journal['approval_sha256'] == APPROVAL_SHA
        check_registry(self.registry, self.nodes, self.baseline, False)
        self.calls = 0

    def verify_inputs(self):
        for value in self.approval['inputs'].values():
            assert sha(Path(value['frozen_path']).read_bytes()) == value['sha256'], 'frozen source input drift'
        assert sha(Path(self.approval['native_baseline']).read_bytes()) == self.approval['native_baseline_sha256']
        projection = json.loads(Path(self.approval['projection']).read_text())
        assert Counter(projection['input_sha256'].values()) == Counter(
            [v['sha256'] for v in self.approval['inputs'].values()] + [self.approval['native_baseline_sha256']])
        for n in self.nodes:
            assert sha(n['approved_body']) == n['approved_body_sha256']
            assert n['type'] == 'Task' and n['size'] == 'M'
        graph = {str(n['number']): list(map(str, n['blocked_by'])) for n in json.loads(Path(self.approval['native_baseline']).read_text())['graph']}
        parents = {str(n['number']): str(n['parent']) for n in json.loads(Path(self.approval['native_baseline']).read_text())['graph'] if n['parent']}
        for n in self.nodes:
            graph[n['alias']] = n['blocked_by_refs']; parents[n['alias']] = n['parent_ref']
        for e in self.approval['allowed_old_root_edges']:
            graph[e['issue']].append(e['blocker'])
        assert {k: sorted(v) for k, v in graph.items()} == {k: sorted(v) for k, v in projection['graph'].items()}, 'approved graph reproduction'
        assert parents == projection['parents'], 'approved containment reproduction'
        caps(graph, parents)

    def api(self, query, variables=None, operation=None):
        if re.match(r'\s*mutation\b',query):
            assert operation, 'native mutation must have a durable operation'
            assert not (self.directory/'native-write-halt.json').exists(), 'ROOT native-write halt; no automatic continuation'
        self.calls += 1
        assert self.calls <= 15000, 'finite API call cap'
        payload = dict(query=query, variables=variables or {})
        entry = None
        if operation:
            entry = dict(operation=operation, started_at=now(), request_sha256=sha(json.dumps(payload, sort_keys=True)), status='INTENT')
            self.journal['operations'].append(entry)
            atomic(self.journal_path, self.journal)
        result = subprocess.run(['gh', 'api', 'graphql', '--input', '-'], input=json.dumps(payload), text=True,
                                capture_output=True, timeout=90)
        try:
            value = json.loads(result.stdout)
        except ValueError:
            value = None
        if entry is not None:
            entry.update(finished_at=now(), status='ACKNOWLEDGED' if result.returncode == 0 and value and not value.get('errors') else 'RECONCILE_REQUIRED',
                         response=value, stderr=result.stderr[-4000:])
            atomic(self.journal_path, self.journal)
        if result.returncode or not value or value.get('errors'):
            raise RuntimeError('API operation failed; read-before-write reconciliation required: ' + str(value or result.stderr[-4000:]))
        return value['data']

    def save_registry(self, status):
        self.registry['status'] = status; self.registry['updated_at'] = now()
        check_registry(self.registry, self.nodes, self.baseline, status != 'ALLOCATING')
        atomic(self.registry_path, self.registry)

    def milestone(self, status):
        """Retain delivered bytes permanently before the mutable status advances."""
        assert status in ['IDS_ALLOCATED', 'VERIFIED'] and self.registry['status'] == status
        check_registry(self.registry, self.nodes, self.baseline, True)
        directory = self.directory / 'milestones'
        directory.mkdir(exist_ok=True)
        raw = self.registry_path.read_bytes()
        snapshot = directory / (status + '.registry.json')
        self.immutable(snapshot, raw)
        map_sha = sha(json.dumps(self.registry['nodes'], sort_keys=True, separators=(',', ':'), ensure_ascii=False))
        sidecar = dict(milestone=status, observed_at=now(), approval_sha256=APPROVAL_SHA,
                       renderer_contract_sha256=RENDERER_SHA, registry_path=str(snapshot), live_registry_path=str(self.registry_path),
                       registry_file_sha256=sha(raw), nodes_map_sha256=map_sha, node_count=871,
                       class_counts={'330':305, '331':281, '332':285})
        if status == 'VERIFIED':
            prior = json.loads((self.directory / 'ids-allocated-milestone.json').read_text())
            assert prior['nodes_map_sha256'] == map_sha
            receipt_path = self.directory / 'verified-native-receipt.json'
            receipt_raw = receipt_path.read_bytes(); receipt = json.loads(receipt_raw)
            retained = directory / 'VERIFIED.native-receipt.json'
            self.immutable(retained, receipt_raw)
            sidecar.update(verified_receipt_path=str(retained), verified_receipt_sha256=sha(receipt_raw),
                           verified_receipt_observed_at=receipt['observed_at'], graph_node_count=1888, all_new_tasks_open=True,
                           native_body_parent_prerequisite_Task_Size_validation=True,
                           both_DAGs=True, all_three_caps=True, old_protections=True,
                           validation_details=dict(
                               native_body_parent_prerequisite_Task_Size_validation='PASS; exact871 frozen contracts',
                               both_DAGs='PASS; independently checked dependency and containment DAGs',
                               all_three_caps=receipt['report'],
                               old_protections='PASS; all1017 old bodies/titles/states/parents/Project metadata and all old edges; only five approved root additions'))
        path = self.directory / ('ids-allocated-milestone.json' if status == 'IDS_ALLOCATED' else 'verified-milestone.json')
        if path.exists():
            existing = json.loads(path.read_text())
            for field in ['registry_file_sha256', 'nodes_map_sha256', 'milestone']:
                assert existing[field] == sidecar[field], 'delivered milestone cannot change'
            return existing
        atomic(path, sidecar)
        self.journal['operations'].append(dict(operation='milestone ' + status, status='RETAINED', receipt=sidecar))
        atomic(self.journal_path, self.journal)
        print(json.dumps(sidecar, indent=2), flush=True)
        return sidecar

    @staticmethod
    def immutable(path, raw):
        if path.exists():
            assert path.read_bytes() == raw, 'immutable delivered evidence differs'
            return
        with path.open('xb') as stream:
            stream.write(raw); stream.flush(); os.fsync(stream.fileno())

    def full_graph(self):
        graph = []; cursor = None; seen = set(); cursors = set()
        query = '''query($cursor:String){repository(owner:"subculture-collective",name:"edda"){
          issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{
          number databaseId id title body state parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}}}}}'''
        for _ in range(25):
            page = self.api(query, {'cursor': cursor})['repository']['issues']
            for n in page['nodes']:
                assert type(n['number']) is int and n['number'] not in seen and not n['blockedBy']['pageInfo']['hasNextPage']
                seen.add(n['number'])
                graph.append(dict(number=n['number'], id=n['databaseId'], node_id=n['id'], title=n['title'], body=n['body'] or '',
                                  state=n['state'].lower(), parent=(n['parent'] or {}).get('number'), blocked_by=sorted(d['number'] for d in n['blockedBy']['nodes'])))
            if not page['pageInfo']['hasNextPage']:
                return graph
            cursor = page['pageInfo']['endCursor']; assert cursor and cursor not in cursors; cursors.add(cursor)
        raise AssertionError('native graph page cap')

    def metadata(self, numbers):
        result = {}
        for offset in range(0, len(numbers), 25):
            batch = numbers[offset:offset + 25]
            fields = ' '.join('i%d:issue(number:%d){id databaseId number title body state parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}} issueType{id name} projectItems(first:100){pageInfo{hasNextPage} nodes{id project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}' % (n, n) for n in batch)
            data = self.api('{repository(owner:"subculture-collective",name:"edda"){' + fields + '}}')['repository']
            for n in batch:
                item = data['i' + str(n)]
                assert item and not item['projectItems']['pageInfo']['hasNextPage'] and not item['blockedBy']['pageInfo']['hasNextPage']
                items = [i for i in item['projectItems']['nodes'] if i['project']['id'] == PROJECT]
                assert len(items) <= 1, 'duplicate project items'
                result[n] = item | {'project_item': items[0] if items else None}
        return result

    def preflight(self):
        graph = self.full_graph(); live = {n['number']: n for n in graph}
        for old in self.baseline['graph']:
            assert all(live[old['number']][k] == v for k, v in old.items()), 'pre-publication baseline drift #' + str(old['number'])
        assert len(live) == 1017, 'unexpected issue identities before allocation'
        meta = self.metadata(sorted(map(int, self.baseline['project'])))
        for number, expected in self.baseline['project'].items():
            actual = meta[int(number)]
            assert actual['issueType']['name'] == expected['type'] and actual['project_item']['fieldValueByName']['name'] == expected['size']
        report = caps({str(n['number']): list(map(str, n['blocked_by'])) for n in graph}, {str(n['number']): str(n['parent']) for n in graph if n['parent']})
        atomic(self.directory / 'preflight-live.json', dict(observed_at=now(), graph=graph, project=self.baseline['project'], report=report))
        print('PREFLIGHT PASS', report, flush=True)

    def allocate(self):
        self.verify_inputs()
        graph = self.full_graph()
        live = {n['number']:n for n in graph}
        for old in self.baseline['graph']:
            assert all(live[old['number']][k] == v for k, v in old.items()), 'old contract changed before allocation'
        by_marker = {}
        for item in graph:
            for alias in re.findall(r'<!-- edda-caster-native:([^\s]+) -->', item['body']):
                assert alias not in by_marker, 'duplicate native marker'
                by_marker[alias] = item
        assert set(by_marker) <= {n['alias'] for n in self.nodes}, 'unapproved marker'
        assert set(live) == {n['number'] for n in self.baseline['graph']} | {n['number'] for n in by_marker.values()}, 'unapproved new identity'
        existing_metadata = self.metadata(sorted(value['number'] for value in self.registry['nodes'].values()))
        for index, n in enumerate(self.nodes, 1):
            alias = n['alias']; existing = by_marker.get(alias)
            if alias in self.registry['nodes']:
                assert existing and self.registry['nodes'][alias] == {k: existing[k] for k in ['number', 'id', 'node_id']}
            elif existing:
                assert existing['title'] == n['title'] and existing['body'] == allocation_body(n, self.contract) and existing['state'] == 'open'
                self.registry['nodes'][alias] = {k: existing[k] for k in ['number', 'id', 'node_id']}
                self.save_registry('ALLOCATING')
            else:
                # Full marker scan above is the read-before-create boundary. On any
                # uncertain API failure the process stops; resume scans again.
                assert not uncertain_creation(self.journal, alias), 'ambiguous prior creation has no readable marker; do not create a duplicate: ' + alias
                payload = self.api('''mutation($input:CreateIssueInput!){createIssue(input:$input){issue{number databaseId id}}}''',
                                {'input': dict(repositoryId=REPO, title=n['title'], body=allocation_body(n, self.contract), issueTypeId=TASK,
                                               projectV2Ids=[PROJECT], clientMutationId=alias)}, 'create ' + alias)
                created = payload.get('createIssue') if isinstance(payload,dict) else None
                data = created.get('issue') if isinstance(created,dict) else None
                valid = (isinstance(data,dict) and type(data.get('number')) is int and data['number']>0
                         and type(data.get('databaseId')) is int and data['databaseId']>0
                         and type(data.get('id')) is str and bool(data['id']))
                if not valid:
                    self.journal['operations'][-1]['status']='RECONCILE_REQUIRED'
                    atomic(self.journal_path,self.journal)
                    raise RuntimeError('Ambiguous createIssue returned no complete typed identity; reconcile exact marker before any retry: '+alias)
                self.registry['nodes'][alias] = dict(number=data['number'], id=data['databaseId'], node_id=data['id'])
                self.save_registry('ALLOCATING')
            identity = self.registry['nodes'][alias]
            current = existing_metadata.get(identity['number'])
            if current is None:
                current = self.metadata([identity['number']])[identity['number']]
            assert current['id'] == identity['node_id'] and current['databaseId'] == identity['id'] and current['state'] == 'OPEN'
            assert current['title'] == n['title'] and current['body'] == allocation_body(n, self.contract)
            assert current['issueType']['name'] == 'Task'
            item = current['project_item']
            if item is None:
                try:
                    item = self.api('mutation($input:AddProjectV2ItemByIdInput!){addProjectV2ItemById(input:$input){item{id}}}',
                                    {'input': dict(projectId=PROJECT, contentId=identity['node_id'])}, 'project-add ' + alias)['addProjectV2ItemById']['item']
                except RuntimeError as error:
                    if 'Content already exists in this project' not in str(error):
                        raise
                    # Auto-add raced the initial read. One bounded recovery read,
                    # never another add or a polling loop.
                    item = self.metadata([identity['number']])[identity['number']]['project_item']
                    assert item is not None, 'auto-added item not yet readable; resume later'
            if not item.get('fieldValueByName') or item['fieldValueByName']['name'] != 'M':
                self.api('mutation($input:UpdateProjectV2ItemFieldValueInput!){updateProjectV2ItemFieldValue(input:$input){projectV2Item{id}}}',
                         {'input': dict(projectId=PROJECT, itemId=item['id'], fieldId=SIZE_FIELD, value=dict(singleSelectOptionId=SIZE_M))}, 'size-M ' + alias)
            if index % 10 == 0 or index == 871:
                print('ALLOCATED', index, '/871', alias, '#' + str(identity['number']), flush=True)
        self.verify_inputs()
        metadata = self.metadata(sorted(v['number'] for v in self.registry['nodes'].values()))
        for n in self.nodes:
            ident = self.registry['nodes'][n['alias']]; actual = metadata[ident['number']]
            assert (actual['id'], actual['databaseId']) == (ident['node_id'], ident['id'])
            assert actual['state'] == 'OPEN' and actual['issueType']['name'] == 'Task' and actual['project_item']['fieldValueByName']['name'] == 'M'
            assert actual['title'] == n['title'] and actual['body'] == allocation_body(n, self.contract)
        self.save_registry('IDS_ALLOCATED')
        print('IDS_ALLOCATED', self.registry_path, flush=True)

    def wire(self):
        self.verify_inputs(); check_registry(self.registry, self.nodes, self.baseline, True)
        delivered = json.loads((self.directory / 'ids-allocated-milestone.json').read_text())
        snapshot = Path(delivered['registry_path']).read_bytes()
        assert sha(snapshot) == delivered['registry_file_sha256']
        assert json.loads(snapshot)['status'] == 'IDS_ALLOCATED'
        assert delivered['nodes_map_sha256'] == sha(json.dumps(self.registry['nodes'], sort_keys=True, separators=(',', ':'), ensure_ascii=False))
        self.save_registry('WIRING')
        graph = self.full_graph(); identities = {n['number']: n['node_id'] for n in graph}
        for index, n in enumerate(self.nodes, 1):
            alias = n['alias']; identity = self.registry['nodes'][alias]; number = identity['number']
            actual = self.metadata([number])[number]
            assert actual['title'] == n['title'] and actual['state'] == 'OPEN'
            body = final_body(n, self.registry, self.contract)
            assert actual['body'] in [allocation_body(n, self.contract), body], 'unexpected owned body drift'
            parent = resolve(n['parent_ref'], self.registry)
            deps = sorted(resolve(ref, self.registry) for ref in n['blocked_by_refs'])
            prior = sorted(d['number'] for d in actual['blockedBy']['nodes'])
            assert set(prior) <= set(deps), 'unapproved owned prerequisite'
            assert actual['parent'] is None or actual['parent']['number'] == parent
            mutations = []
            if actual['body'] != body:
                mutations.append(('body', 'updateIssue', dict(id=identity['node_id'], body=body)))
            if actual['parent'] is None:
                mutations.append(('parent', 'addSubIssue', dict(issueId=identities[parent], subIssueId=identity['node_id'])))
            for i, dep in enumerate(d for d in deps if d not in prior):
                mutations.append(('d'+str(i), 'addBlockedBy', dict(issueId=identity['node_id'], blockingIssueId=identities[dep])))
            if mutations:
                # GraphQL mutation fields execute serially. Variables keep the
                # source bodies and identifiers out of executable query text.
                inputs = {}; fields = []; declarations = []
                for i, (field, operation, value) in enumerate(mutations):
                    inputs['v' + str(i)] = value
                    declarations.append('$v%d:%sInput!' % (i, operation[0].upper() + operation[1:]))
                    fields.append(field + ':' + operation + '(input:$v' + str(i) + '){issue{id}}')
                self.api('mutation(' + ','.join(declarations) + '){' + ' '.join(fields) + '}', inputs, 'wire ' + alias)
            if index % 10 == 0 or index == 871:
                print('WIRED', index, '/871', alias, '#' + str(number), flush=True)
        for edge in self.approval['allowed_old_root_edges']:
            number = int(edge['issue']); dep = resolve(edge['blocker'], self.registry)
            actual = self.metadata([number])[number]
            old = next(n for n in self.baseline['graph'] if n['number'] == number)
            assert actual['title'] == old['title'] and actual['body'] == old['body'] and (actual['parent'] or {}).get('number') == old['parent']
            allowed = set(old['blocked_by']) | {resolve(e['blocker'], self.registry) for e in self.approval['allowed_old_root_edges'] if e['issue'] == str(number)}
            prior = {d['number'] for d in actual['blockedBy']['nodes']}
            assert set(old['blocked_by']) <= prior <= allowed
            if dep not in prior:
                self.api('mutation($input:AddBlockedByInput!){addBlockedBy(input:$input){issue{id}}}',
                         {'input': dict(issueId=identities[number], blockingIssueId=identities[dep])}, 'old-root-edge ' + str(number) + ' ' + edge['blocker'])
        print('WIRING COMPLETE; verification required', flush=True)

    def verify(self):
        self.verify_inputs(); check_registry(self.registry, self.nodes, self.baseline, True)
        graph = self.full_graph(); live = {n['number']: n for n in graph}
        assert len(live) == 1888
        old_metadata = sorted(map(int, self.baseline['project']))
        metadata = self.metadata(sorted(set(old_metadata) | {v['number'] for v in self.registry['nodes'].values()}))
        for old in self.baseline['graph']:
            actual = live[old['number']]
            for key, value in old.items():
                if key != 'blocked_by':
                    assert actual[key] == value, 'old contract changed #' + str(old['number']) + ' ' + key
            expected = sorted(old['blocked_by'] + [resolve(e['blocker'], self.registry) for e in self.approval['allowed_old_root_edges'] if int(e['issue']) == old['number']])
            assert actual['blocked_by'] == expected, 'old edges changed #' + str(old['number'])
        for number, expected in self.baseline['project'].items():
            actual = metadata[int(number)]
            assert actual['issueType']['name'] == expected['type'] and actual['project_item']['fieldValueByName']['name'] == expected['size']
        hashes = {}
        for n in self.nodes:
            identity = self.registry['nodes'][n['alias']]; actual = live[identity['number']]
            assert {k: actual[k] for k in identity} == identity
            body = final_body(n, self.registry, self.contract)
            assert actual['body'] == body and actual['title'] == n['title'] and actual['state'] == 'open'
            assert actual['parent'] == resolve(n['parent_ref'], self.registry)
            assert actual['blocked_by'] == sorted(resolve(d, self.registry) for d in n['blocked_by_refs'])
            meta = metadata[identity['number']]
            assert meta['issueType']['name'] == 'Task' and meta['project_item']['fieldValueByName']['name'] == 'M'
            hashes[n['alias']] = sha(body)
        report = caps({str(n['number']): list(map(str, n['blocked_by'])) for n in graph}, {str(n['number']): str(n['parent']) for n in graph if n['parent']})
        assert (report['incoming279'], report['incoming280'], report['incoming337'], report['children283'], report['outgoing429']) == (100, 95, 59, 95, 100)
        receipt = dict(schema_version=1, status='VERIFIED', approval_sha256=APPROVAL_SHA, renderer_contract_sha256=RENDERER_SHA,
                       coordinator_code_sha256=sha(Path(__file__).read_bytes()),
                       observed_at=now(), graph=graph, project={str(n): dict(type=m['issueType']['name'], size=m['project_item']['fieldValueByName']['name']) for n, m in metadata.items()},
                       expected_final_body_sha256=hashes, report=report,
                       limits='Mechanical source-contract publication only; all871 Task/M OPEN; no runtime qualification, parent supplements, class acceptance, CI or merge.')
        receipt['accepted_rogue_guard'] = self.rogue_guard(receipt)
        atomic(self.directory / 'verified-native-receipt.json', receipt)
        self.save_registry('VERIFIED')
        print('VERIFIED', report, flush=True)

    def rogue_guard(self, receipt):
        """Run the actual accepted earlier proof without weakening its guards."""
        common = Path(subprocess.check_output(['git','rev-parse','--git-common-dir'], text=True).strip())
        root = common.parent
        for name in ['rogue_native_proof.py','verify_rogue_scope.py','freeze_rogue_parents.py',
                     'publish_rogue_scope.py','rogue_capacity.py','verify_fighter_scope.py','publish_paladin_scope.py']:
            expected = subprocess.check_output(['git','show',self.approval['root_main']+':scripts/'+name])
            assert (root/'scripts'/name).read_bytes() == expected, 'accepted Rogue proof code changed'
        sys.path.insert(0, str(root/'scripts'))
        from rogue_native_proof import validate
        manifest = json.loads((root/'docs/investigations/issue-329-rogue-scope.json').read_text())
        report = validate(manifest, receipt, sealed=True)
        assert report == dict(nodes=1888, owned_tasks=77, owned_edges=391, root_edges=3, terminals=72, final_seal='PASS')
        return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase', choices=['preflight', 'allocate', 'wire', 'verify', 'snapshot-ids', 'snapshot-verified'])
    parser.add_argument('--directory', type=Path, default=DEFAULT)
    args = parser.parse_args()
    with (args.directory / 'coordinator.lock').open('a') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        coordinator = Coordinator(args.directory)
        if args.phase.startswith('snapshot-'):
            coordinator.milestone('IDS_ALLOCATED' if args.phase == 'snapshot-ids' else 'VERIFIED')
        else:
            getattr(coordinator, args.phase)()


if __name__ == '__main__':
    main()
