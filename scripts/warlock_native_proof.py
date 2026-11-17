"""Read-only native proof with independently pinned accepted predecessor bodies."""
import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path
import subprocess

from verify_warlock_scope import ROOT, exact, caps
from warlock_native_plan import build, projection, PROJECT

# Independent raw/body pins bind ROOT's accepted predecessor release.
# This release permits preparation; actual own sections must still be read back.
from warlock_parent_seal import load_prefixes, validate_parents
ACCEPTED_POST_SORCERER_PREFIXES = load_prefixes()
APPROVED_FINITE_EXTENSIONS = ()
BASELINE = ROOT / 'docs/investigations/issue-331-native-baseline.json'
FOREIGN = ROOT / 'docs/investigations/issue-331-sorcerer-boundary.json'
PROPOSAL = ROOT / 'docs/investigations/issue-331-native-proposal.json'


def validate_proposal(m, actual, baseline, foreign):
    from warlock_approved_native import load_frozen
    _, frozen, approved_baseline, _, _, _ = load_frozen()
    assert exact(actual, frozen[331]), 'unchanged281-node approved proposal'
    assert exact(foreign, frozen[330]) and exact(baseline, approved_baseline), 'explicit frozen input refresh only'
    expected = build(m, foreign, baseline['graph'])
    assert exact(actual, expected), 'complete actual candidate native bodies/parents/prerequisites/TaskSize'
    graph = {x['number']: x for x in baseline['graph']}
    assert len(graph) == 1017, 'observed baseline extent, not a final prefix'
    assert hashlib.sha256(graph[779]['body'].encode()).hexdigest() == '556d1618d1cd2bd9c956fe38861d89e94ee14ff61a505b5d2b6a1fa46d19d150'
    assert graph[779]['parent'] == 50 and graph[779]['blocked_by'] == [61, 64, 65, 66, 67, 488, 778]
    assert graph[968]['parent'] == 336 and graph[968]['blocked_by'] == [429]
    assert [x['number'] for x in m['source_prerequisites']] == [285, 296, 302, 303, 304, 318]
    for source in m['source_prerequisites']:
        old = graph[source['number']]
        assert source['state'] == 'CLOSED' and old['state'] in ('CLOSED', 'closed')
        assert source['body'] == old['body'] and source['title'] == old['title']
    for entry in m['spell_overlap_audit']:
        for owner in entry['candidate_owners']:
            old = graph[owner['issue']]
            assert owner['body'] == old['body'] and owner['title'] == old['title'], 'whole existing contract reuse'
    wizard = json.loads((ROOT / 'docs/investigations/issue-331-wizard-boundary.json').read_text())
    r = projection(actual, baseline['graph'], foreign, wizard)
    caps(r['graph'], r['parents'])
    assert not r['unresolved_foreign'], 'no empty/dangling foreign heads'
    assert all(parent in r['graph'] for parent in r['parents'].values()), 'complete containment graph'
    assert len(actual['tasks']) == 254 and len(actual['gates']) == 27, 'finite281 actual owned nodes'
    for old in baseline['graph']:
        if old['number'] not in (279, 280, 337):
            assert r['graph'][str(old['number'])] == [str(x) for x in old['blocked_by']], 'no unapproved old prerequisite mutation'
    assert r['children283'] == 95 and r['outgoing429'] == 100
    assert r['root279'] == 100 and r['root280'] == 95
    assert all(429 not in t['prerequisites'] for t in actual['tasks'] + actual['gates'])
    assert next(t for t in actual['tasks'] if t['key'] == 'hp')['prerequisites'] == [774, 968]
    return r


class Reader:
    """Bounded real GitHub queries with no repeated cursors or silent API hangs."""
    def __init__(self):
        self.calls = 0

    def __call__(self, query, variables=None):
        self.calls += 1
        assert self.calls <= 80, 'native read call cap'
        assert 'mutation' not in query, 'read-only proof'
        p = subprocess.run(['gh', 'api', 'graphql', '--input', '-'],
            input=json.dumps({'query': query, 'variables': variables or {}}),
            capture_output=True, text=True, check=True, timeout=45)
        result = json.loads(p.stdout)
        assert not result.get('errors'), 'GraphQL error'
        return result['data']['repository']


def read_live(issue_map, protected_project_numbers=()):
    started_at = datetime.now(timezone.utc).isoformat()
    api, cursor, cursors, seen, graph = Reader(), None, set(), set(), []
    query = 'query($cursor:String){repository(owner:"subculture-collective",name:"edda"){issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{id databaseId number title body state parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}}}}}'
    for _ in range(20):
        page = api(query, {'cursor': cursor})['issues']
        for row in page['nodes']:
            assert row['number'] not in seen, 'duplicate issue/page'
            seen.add(row['number'])
            assert not row['blockedBy']['pageInfo']['hasNextPage'], 'incoming cap'
            graph.append(dict(number=row['number'], id=row['databaseId'], node_id=row['id'], title=row['title'], body=row['body'] or '',
                state=row['state'], parent=(row['parent'] or {}).get('number'),
                blocked_by=sorted(n['number'] for n in row['blockedBy']['nodes'])))
        if not page['pageInfo']['hasNextPage']:
            break
        cursor = page['pageInfo']['endCursor']
        assert cursor and cursor not in cursors, 'missing/repeated cursor'
        cursors.add(cursor)
    else:
        raise AssertionError('native graph page cap')
    project = {}
    nums = sorted(set(issue_map.values()) | {779, 968} | {int(n) for n in protected_project_numbers})
    for start in range(0, len(nums), 25):
        batch = nums[start:start+25]
        fields = ' '.join(f'i{n}:issue(number:{n})' + '{issueType{name} projectItems(first:100){pageInfo{hasNextPage}nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}' for n in batch)
        result = api('{repository(owner:"subculture-collective",name:"edda"){' + fields + '}}')
        for n in batch:
            row = result['i' + str(n)]
            assert not row['projectItems']['pageInfo']['hasNextPage'], 'project page bound'
            entries = [x for x in row['projectItems']['nodes'] if x['project']['id'] == PROJECT]
            assert len(entries) == 1, 'exact Project membership'
            project[str(n)] = dict(type=(row['issueType'] or {}).get('name'), size=(entries[0]['fieldValueByName'] or {}).get('name'))
    return dict(graph=graph, project=project, api_calls=api.calls,
                observed_from=started_at, observed_at=datetime.now(timezone.utc).isoformat())


def validate_native(m, plan, receipt, snapshot):
    """Final proof additionally requires ROOT's real accepted prefix release."""
    assert ACCEPTED_POST_SORCERER_PREFIXES is not None, 'ROOT accepted POST-SORCERER prefix release pending'
    assert set(ACCEPTED_POST_SORCERER_PREFIXES) == {47, 54, 283, 337}, 'complete ROOT release'
    from warlock_approved_native import load_frozen, validate_wiring
    approval, proposals, baseline, expected, _, _ = load_frozen()
    assert exact(plan, proposals[331]), 'exact approved own proposal'
    graph = {t['number']: t for t in snapshot['graph']}
    validate_parents(graph)
    result = validate_wiring(approval, expected, baseline, receipt['registry'], snapshot)
    return {**result, 'status': 'FINAL_NATIVE_AND_PREFIX_PASS', 'final_prefix_sealed': True}


def prove_registry(m, registry_raw, delivery, stage, prior_allocation=None, verified_receipt_raw=None, prior_allocation_raw=None):
    """Run only at a ROOT-delivered allocation/wiring boundary; never poll."""
    from warlock_approved_native import (load_frozen, load_renderer, load_milestone_contract,
        validate_milestone, validate_registry, validate_allocation, validate_wiring)
    approval, proposals, baseline, expected, _, _ = load_frozen()
    load_renderer()
    load_milestone_contract()
    validate_proposal(m, proposals[331], baseline, proposals[330])
    assert stage in ('allocation', 'wiring'), 'exact milestone proof mode'
    assert delivery['milestone'] == ('IDS_ALLOCATED' if stage == 'allocation' else 'VERIFIED')
    registry = validate_milestone(registry_raw, delivery, expected, baseline, prior_allocation, verified_receipt_raw, prior_allocation_raw)
    numbers = validate_registry(registry, expected, baseline)
    snapshot = read_live(numbers, baseline['project'])
    check = validate_allocation if stage == 'allocation' else validate_wiring
    result = check(approval, expected, baseline, registry, snapshot)
    return dict(registry=registry, delivery=delivery, allocation_delivery=prior_allocation,
                snapshot=snapshot, result=result)
