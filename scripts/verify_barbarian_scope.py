#!/usr/bin/env python3
"""Verify #321 source/ownership/publication offline; --live only reads GitHub."""
import argparse
from copy import deepcopy
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / 'docs/investigations/issue-321-barbarian-scope.json'
RECEIPT = ROOT / 'docs/investigations/issue-321-link-verification.json'
MIGRATION = ROOT / 'docs/investigations/issue-321-gate-migration.json'
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
REPO = 'repos/subculture-collective/edda'


def digest(value):
    return hashlib.sha256(value.encode()).hexdigest()


def api(path, data=None):
    command = ['gh', 'api', path]
    if data is not None:
        command += ['--input', '-']
    return json.loads(subprocess.run(command, input=json.dumps(data) if data is not None else None,
                                    capture_output=True, text=True, check=True).stdout)


def body(m, r):
    gate = m['qualification_gate']['issue']
    keys = sorted(k for k, v in m['source_key_owners'].items() if v == r['key'])
    source_records = [x['record'] for x in m['records'] if any(k in keys for k in x['key_owners'])]
    if 'level_record' in r:
        source_records = [x['record'] for x in m['records'] if x['record']['id'] == r['level_record']]
    return (f"<!-- edda-321:{r['key']} -->\nParent: #{r['parent']}; decision #321; roadmap #47; blocks #{gate}; #{gate} directly blocks #337.\n\n"
            f"Delivery: one reviewable PR, Task Size {r['project_size']}.\n\n"
            f"Source: pinned English SRD5.2.1 pp28-30 and explicitly named referenced pages below; SHA256 {PIN}. "
            "CC-BY-4.0 attribution and adaptation: docs/srd-5.2.1-conformance.md.\n\n"
            f"Scope and deterministic positive fixture: {r['fixture']}\n\n"
            "Negative fixtures: violate each stated eligibility/number/target/action/resource/timing predicate independently; "
            "reject unsupported variants and forged levels. Preview and cancellation never mutate.\n\n"
            f"Required acceptance: {m['common_acceptance']}\n\n"
            "Ownership: feature resolvers own behavior; level/entry applications consume them and apply grants once. "
            "Feature tests use explicit isolated entitlements, never grant production class admission. Data and earlier levels never unlock broader support. "
            "Shared #272 rests, #273 casting, #333 origins, #335 multiclass, #336 advancement, #384 selection/#389 ability updater, "
            f"#269/#270 conditions and #278 equipment retain their scopes. Exact shared terminals in this matrix must block #{gate} natively; #{gate} blocks #337. Planning closure is insufficient. "
            "Selected feat effect must have its own terminal gate before admission; never implement spell/item effects here. "
            "No child implementation or qualification is supplied by this decision.\n\n"
            f"Native blocked-by: {', '.join('#'+str(n) for n in r['blocked_by_issues'])}.\n\n"
            "Exact owned source keys (shared keys have one owner, consumers do not implement them):\n" +
            ('\n'.join('- '+k for k in keys) or '- Application/shared referenced-rule scope above; no duplicate mechanic ownership.') +
            '\n\nExact source records including nested alternatives, conditions and numbers (context, not duplicate ownership):\n```json\n' +
            json.dumps(source_records, indent=2) + '\n```\n')


def acyclic(graph):
    active, done = set(), set()
    def visit(n):
        assert n in graph, f'missing node {n}'
        assert n not in active, f'cycle {n}'
        if n in done:
            return
        active.add(n)
        for b in graph[n]['blocked_by']:
            visit(b)
        active.remove(n)
        done.add(n)
    for n in graph:
        visit(n)


def validate(m, receipt):
    assert m['source']['sha256'] == PIN
    records = []
    for path, snapshot in m['catalogues'].items():
        current = json.loads((ROOT / path).read_text())
        assert current == snapshot, f'catalogue drift {path}'
        assert current['source']['sha256'] == PIN
        records += current['records']
    assert len(records) == 80 and [x['record'] for x in m['records']] == records
    levels = [r for r in records if r['kind'] == 'level']
    assert [r['level'] for r in levels] == list(range(1, 21))
    keys = {k for r in records for k in r['required_mechanics'] + r['option_keys']}
    assert set(m['source_key_owners']) == keys
    rows = {r['key']: r for r in m['issues']}
    assert len(rows) == len(m['issues']) == 71
    assert len({r['issue'] for r in rows.values()}) == 71
    assert set(m['source_key_owners'].values()) <= set(rows) | {'external:384', 'external:389'}
    for x in m['records']:
        assert x['key_owners'] == {k: m['source_key_owners'][k] for k in x['record']['required_mechanics'] + x['record']['option_keys']}
        assert x['state'] == 'unqualified metadata; no executable Barbarian fixture'
        assert x['evidence'] == list(m['evidence_hashes'])
        if x['record']['kind'] == 'level':
            assert x['application_owner'] == f"level-{x['record']['level']}"
            assert rows[x['application_owner']]['level_record'] == x['record']['id']
    for path, expected in m['evidence_hashes'].items():
        assert hashlib.sha256((ROOT / path).read_bytes()).hexdigest() == expected, f'evidence drift {path}'
    graph = {r['number']: r for r in receipt['graph']}
    assert len(graph) == len(receipt['graph'])
    acyclic(graph)
    acyclic({n: {"blocked_by": [r["parent"]] if r["parent"] else []} for n, r in graph.items()})
    assert graph[321]['parent'] == graph[337]['parent'] == 283
    assert {285, 286, 308} <= set(graph[321]['blocked_by']), 'missing source edge'
    assert 321 in graph[337]['blocked_by']
    for n in [285, 286, 308]:
        assert receipt['source_prerequisites'][str(n)] == 'closed'
    gate = m['qualification_gate']
    g = gate['issue']
    assert set(gate['blocked_by_issues']) == {321} | {r['issue'] for r in rows.values()}
    assert graph[g]['parent'] == gate['parent'] == 283
    assert set(graph[g]['blocked_by']) == {321} | {r['issue'] for r in rows.values()}, 'missing class gate prerequisites'
    assert g in graph[337]['blocked_by'], 'missing class gate to #337'
    assert receipt['bodies'][str(g)] == {'title': gate['title'], 'sha256': digest(gate['body'])}
    assert receipt['project'][str(g)] == {'type': 'Task', 'size': gate['project_size']}
    migration = json.loads(MIGRATION.read_text())
    assert migration['gate'] == g
    assert migration['preconditions']['states'] == {'321': 'open', '337': 'open'}
    assert migration['preconditions']['337_blocked_by_321']
    assert migration['all_paths_verified_before_remaining_removal'] and migration['all_71_terminal_paths_verified']
    owned = {r['issue'] for r in rows.values()}
    assert set(migration['removed_owned']) == set(migration['before_337']) & owned
    assert set(migration['protected_legacy']) == set(migration['before_337']) - owned
    assert set(migration['historical_334_snapshot_edges']) <= set(migration['protected_legacy'])
    assert set(migration['protected_legacy']) <= set(graph[337]['blocked_by']), 'removed historical #337 edge'
    assert set(migration['after_337']) == set(migration['protected_legacy']) | {g}
    for r in rows.values():
        n = r['issue']
        assert set(r['blocked_by_issues']) == set(r['prerequisites']) | {rows[k]['issue'] for k in r['depends']}
        assert graph[n]['parent'] == r['parent'], f'parent {n}'
        assert set(graph[n]['blocked_by']) == set(r['blocked_by_issues']), f'blockers {n}'
        assert n in graph[g]['blocked_by'], f'missing class gate {n}'
        assert receipt['bodies'][str(n)] == {'title': r['title'], 'sha256': digest(body(m, r))}, f'body {n}'
        assert receipt['project'][str(n)] == {'type': 'Task', 'size': r['project_size']}
    for n in m['reused_terminals']:
        assert n in graph[337]['blocked_by'], f'missing reused gate {n}'
    assert set(m['shared_terminal_keys']) == {r['key'] for r in rows.values() if r['parent'] != 283}
    for s in m['supplements']:
        saved = receipt['supplements'][str(s['issue'])]
        assert saved['append_sha256'] == digest(s['append'])
        assert saved['publication_prefix_sha256'] == s['previous_sha256']
    return len(keys), len(graph)


def live(m):
    # REST pagination for both issues and edges avoids the 100-blocker truncation
    # when #337 accumulates terminal requirements from other decision families.
    def pages(path):
        result, seen, page = [], set(), 1
        while True:
            batch = api(f'{path}?per_page=100&page={page}')
            ids = [x['id'] for x in batch]
            assert len(ids) == len(set(ids)) and not seen.intersection(ids), 'repeated API page'
            seen.update(ids)
            result += batch
            if len(batch) < 100:
                return result
            page += 1
    graph, cursor = [], None
    while True:
        q = '''query($cursor:String){repository(owner:"subculture-collective",name:"edda"){
        issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{number parent{number} blockedBy(first:100){pageInfo{hasNextPage} nodes{number}}}}}}'''
        result = api('graphql', {'query': q, 'variables': {'cursor': cursor}})
        assert not result.get('errors'), result.get('errors')
        batch = result['data']['repository']['issues']
        for node in batch['nodes']:
            graph.append({'number': node['number'], 'parent': (node['parent'] or {}).get('number'),
                          'blocked_by': sorted(x['number'] for x in (pages(f"{REPO}/issues/{node['number']}/dependencies/blocked_by") if node['blockedBy']['pageInfo']['hasNextPage'] else node['blockedBy']['nodes']))})
        if not batch['pageInfo']['hasNextPage']:
            break
        cursor = batch['pageInfo']['endCursor']
    receipt = {'observed_at': datetime.now(timezone.utc).isoformat(), 'graph': graph, 'bodies': {}, 'project': {}, 'supplements': {}}
    for r in m['issues'] + [m['qualification_gate']]:
        x = api(f"{REPO}/issues/{r['issue']}")
        assert x['title'] == r['title'] and x['body'] == (r['body'] if 'body' in r else body(m, r)), f"live body drift {r['issue']}"
        receipt['bodies'][str(r['issue'])] = {'title': x['title'], 'sha256': digest(x['body'])}
    fields = [f'i{r["issue"]}:issue(number:{r["issue"]})' + '{issueType{name} projectItems(first:100){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}' for r in m['issues'] + [m['qualification_gate']]]
    result = api('graphql', {'query': '{repository(owner:"subculture-collective",name:"edda"){'+' '.join(fields)+'}}'})
    assert not result.get('errors'), result.get('errors')
    for r in m['issues'] + [m['qualification_gate']]:
        node = result['data']['repository'][f'i{r["issue"]}']
        assert not node['projectItems']['pageInfo']['hasNextPage']
        items = [x for x in node['projectItems']['nodes'] if x['project']['id'] == 'PVT_kwDODjtmGM4BrFJ2']
        assert len(items) == 1
        receipt['project'][str(r['issue'])] = {'type': node['issueType']['name'], 'size': items[0]['fieldValueByName']['name']}
    receipt['source_prerequisites'] = {str(n): api(f'{REPO}/issues/{n}')['state'] for n in [285, 286, 308]}
    for s in m['supplements']:
        x = api(f"{REPO}/issues/{s['issue']}")
        assert x['body'].count(s['append']) == 1, 'owned supplement changed'
        receipt['supplements'][str(s['issue'])] = {'append_sha256': digest(s['append']), 'publication_prefix_sha256': s['previous_sha256'], 'observed_prefix_sha256': digest(x['body'].replace(s['append'], '', 1))}
    return receipt


def probes(m, receipt):
    def reject(label, mm, rr):
        try:
            validate(mm, rr)
        except (AssertionError, KeyError):
            return
        raise AssertionError('accepted corruption: '+label)
    mm = deepcopy(m); mm['source']['sha256'] = 'wrong'; reject('wrong source', mm, receipt)
    mm = deepcopy(m); next(iter(mm['catalogues'].values()))['records'][0]['name'] = 'wrong'; reject('wrong nested source', mm, receipt)
    mm = deepcopy(m); mm['issues'].pop(); reject('missing child', mm, receipt)
    mm = deepcopy(m); mm['source_key_owners'].pop(next(iter(mm['source_key_owners']))); reject('missing key', mm, receipt)
    for n in [285, 286, 308]:
        rr = deepcopy(receipt); next(x for x in rr['graph'] if x['number'] == 321)['blocked_by'].remove(n); reject('source edge', m, rr)
    for r in m['issues']:
        rr = deepcopy(receipt); next(x for x in rr['graph'] if x['number'] == m['qualification_gate']['issue'])['blocked_by'].remove(r['issue']); reject('terminal gate', m, rr)
    rr = deepcopy(receipt); next(x for x in rr['graph'] if x['number'] == 337)['blocked_by'].remove(m['qualification_gate']['issue']); reject('class gate to root', m, rr)
    rr = deepcopy(receipt); next(x for x in rr['graph'] if x['number'] == m['qualification_gate']['issue'])['blocked_by'].remove(321); reject('planning prerequisite', m, rr)
    for legacy in [334, 384]:
        rr = deepcopy(receipt); next(x for x in rr['graph'] if x['number'] == 337)['blocked_by'].remove(legacy); reject('historical root edge', m, rr)
    rr = deepcopy(receipt); next(x for x in rr['graph'] if x['number'] == 283)['parent'] = m['qualification_gate']['issue']; reject('parent cycle', m, rr)
    rr = deepcopy(receipt); rr['bodies'][str(m['issues'][0]['issue'])]['sha256'] = 'wrong'; reject('body hash', m, rr)
    rr = deepcopy(receipt); next(x for x in rr['graph'] if x['number'] == m['issues'][0]['issue'])['blocked_by'].append(337); reject('cycle', m, rr)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--live', action='store_true')
    parser.add_argument('--pdf', type=Path)
    args = parser.parse_args()
    m = json.loads(MANIFEST.read_text())
    if args.pdf:
        assert hashlib.sha256(args.pdf.read_bytes()).hexdigest() == PIN
    receipt = live(m) if args.live else json.loads(RECEIPT.read_text())
    keys, nodes = validate(m, receipt)
    probes(m, receipt)
    if args.live:
        RECEIPT.write_text(json.dumps(receipt, indent=2)+'\n')
    print(f'PASS: 80 exact records, 20 rows, {keys} unique keys; 71 terminals plus class gate: exact bodies/parents/blockers/Task sizes; {nodes} acyclic nodes; corruption probes rejected')


if __name__ == '__main__':
    main()
