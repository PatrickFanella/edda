#!/usr/bin/env python3
"""Read-only #330 proof. A pending checkpoint cannot pass publication validation."""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import subprocess
from urllib.parse import parse_qs, urlsplit

from sorcerer_contract_oracle import (
    COMMON, CONTRACTS, DEPENDENCIES, GRANTS, SOURCE_LISTS, FINITE_CONSUMED, execution_owner, level_fixture,
)

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / 'docs/investigations/issue-330-sorcerer-scope.json'
PIN = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
BASE = 'e7cd1ba8318ba296edd4fc5b6b84715ca11016f2'


def exact(a, b):
    """JSON equality must distinguish bool, integer and decimal operands."""
    if type(a) is not type(b):
        return False
    if isinstance(a, dict):
        return a.keys() == b.keys() and all(exact(a[k], b[k]) for k in a)
    if isinstance(a, list):
        return len(a) == len(b) and all(exact(x, y) for x, y in zip(a, b))
    return a == b


def acyclic(graph):
    done, active = set(), set()

    def visit(key):
        assert key not in active, 'dependency/containment cycle'
        if key in done:
            return
        active.add(key)
        for dep in graph.get(key, []):
            visit(dep)
        active.remove(key)
        done.add(key)

    for key in graph:
        visit(key)


def native_caps(graph, parents):
    acyclic(graph)
    acyclic({k: [v] for k, v in parents.items()})
    outgoing = Counter(d for deps in graph.values() for d in deps)
    children = Counter(parents.values())
    assert all(len(set(v)) == len(v) and len(v) <= 100 for v in graph.values()), 'incoming cap'
    assert max(outgoing.values(), default=0) <= 100, 'outgoing cap'
    assert max(children.values(), default=0) <= 100, 'native child cap'


def bounded_pages(fetch, endpoint, call_cap=20):
    """Exact page parsing; repeated payloads, ignored page or runaway APIs fail."""
    parsed = urlsplit(endpoint)
    query = parse_qs(parsed.query)
    assert query.get('per_page') == ['100'] and 'page' not in query, 'exact query'
    seen_pages, seen_ids, result = set(), set(), []
    for page in range(1, call_cap + 1):
        url = endpoint + '&page=' + str(page)
        data = fetch(url)
        assert isinstance(data, list) and len(data) <= 100, 'page shape'
        if not data:
            return result
        payload = json.dumps(data, sort_keys=True)
        assert payload not in seen_pages, 'repeated page'
        seen_pages.add(payload)
        for item in data:
            assert type(item['number']) is int and item['number'] not in seen_ids, 'repeated issue'
            seen_ids.add(item['number'])
            result.append(item)
        if len(data) < 100:
            return result
    raise AssertionError('bounded API call cap')


def prefix_check(body, prefix, append, marker):
    """Full released prefix plus exactly one complete append; pure later appends allowed."""
    assert body.startswith(prefix), 'modified or rolled-back prefix'
    expected = prefix + append
    assert body.startswith(expected), 'missing/partial/modified append'
    assert body.count(marker) == 1, 'duplicate append marker'


def expected_body(t):
    return ('Decision #330; one implementation PR; Task Size M.\n\n'
            'Exact execution scope and positive/independently violated negative controls:\n'
            + t['fixture'] + '\n\n' + COMMON
            + '\n\nExact native prerequisite numbers: ' + str(t['prerequisites'])
            + '; planned owned prerequisites: ' + str(t['depends'])
            + '.\n\nOwned source contexts:\n```json\n'
            + json.dumps(t['owned_contexts'], indent=2) + '\n```\n')


def validate_plan(m):
    assert exact(m['spell_lists'], SOURCE_LISTS), 'complete class source lists'
    assert exact(m['future_source_rosters']['consumed'], FINITE_CONSUMED), 'finite future consumed source roster'
    assert m['future_source_rosters']['source_union'] == sorted(set.union(*(set(v) for v in FINITE_CONSUMED.values()))), 'complete finite source union'
    assert m['decision'] == 330 and m['base'] == BASE and m['source']['sha256'] == PIN
    actual = []
    for name, count in [('sorcerer', 63), ('metamagic', 22), ('draconic', 18)]:
        path = 'internal/rules/srd521/testdata/character-' + name + '.json'
        data = json.loads((ROOT / path).read_text())
        assert exact(m['catalogues'][path], data), 'source operands/types'
        assert len(data['records']) == count
        actual.extend(data['records'])
    assert exact(actual, m['records']) and len(actual) == 103, 'complete source records'
    assert len(m['record_matrix']) == 103, 'complete contextual matrix'
    tasks = {t['key']: t for t in m['tasks']}
    assert len(tasks) == len(m['tasks']) == 66
    assert set(tasks) == set(CONTRACTS) | {'level-' + str(n) for n in range(1, 21)}, 'finite contracts'
    for record, context in zip(actual, m['record_matrix']):
        assert exact(context['record'], record) and context['id'] == record['id']
        expected = {k: execution_owner(record, k) for k in record['required_mechanics'] + record['option_keys']}
        assert context['key_owners'] == expected, 'independent source execution owner'
        assert context['status'] == 'unavailable'
        assert all(context[k] for k in ['rules', 'game', 'engine', 'combat', 'handlers', 'database', 'web', 'tui'])
    for key, t in tasks.items():
        assert t['type'] == 'Task' and t['size'] == 'M' and t['parent'] == 330
        if key in CONTRACTS:
            title, fixture, pre = CONTRACTS[key]
            deps = DEPENDENCIES.get(key, [])
            assert t['title'] == title and t['fixture'] == fixture, 'readable execution semantics'
            assert t['prerequisites'] == pre and t['depends'] == deps, 'exact protocol prerequisites'
        else:
            n = int(key.removeprefix('level-'))
            assert t['fixture'] == level_fixture(n), 'level selectors and source guards'
            mandatory = {'hp', 'predicates', 'later-entry', 'slots', 'preparation', 'cantrips'} | set(GRANTS.get(n, []))
            if n == 1:
                mandatory.add('creation')
            if n >= 3:
                mandatory.add('draconic-hp')
            assert set(t['depends']) == mandatory and t['prerequisites'] == [285], 'all20 independent granting paths'
        contexts = [dict(id=c['id'], keys=[k for k, v in c['key_owners'].items() if v == key])
                    for c in m['record_matrix'] if key in c['key_owners'].values()]
        assert exact(t['owned_contexts'], contexts), 'full owned contexts'
        records = [c['record'] for c in m['record_matrix'] if key in c['key_owners'].values()]
        assert exact(t['source_records'], records), 'full source operands'
        assert t['body'] == expected_body(t), 'complete body semantics'
    acyclic({key: t['depends'] for key, t in tasks.items()})
    assert tasks['hp']['prerequisites'] == [968]
    assert not any(429 in t['prerequisites'] for t in tasks.values()), '429 full'
    assert m['overlap']['title_candidates']['Light'] == [967], 'Light false overlap'
    assert 'Slow' not in m['overlap']['title_candidates'], 'Slow mastery false overlap'
    assert m['overlap']['title_candidates']['Polymorph'] == [698, 724, 725]
    return tasks


def validate_publication(m, receipt):
    validate_plan(m)
    assert m['status'] == 'PUBLISHED', 'PENDING checkpoint cannot qualify publication'
    assert m['accepted_parent_prefix'], 'root POST-ROGUE release required'
    assert receipt['publication_complete'] is True, 'native readback incomplete'
    # Deliberately unavailable until all shared semantic contracts and the root seal
    # are independently implemented and reviewed. Never accept a boolean assertion
    # or coherent manifest/receipt rewrite as publication proof.
    raise AssertionError('final shared/native semantic oracle not completed')


def source_check(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN
    for name in ['sorcerer', 'metamagic', 'draconic', 'advancement', 'multiclass',
                 'origin_feats', 'general_feats', 'fighting_styles', 'epic_boon_feats']:
        subprocess.run(['python3', str(ROOT / 'scripts' / ('verify_srd_' + name + '.py')), str(pdf)], check=True)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--allow-pending', action='store_true', help='Audit source/class checkpoint only; never publication proof')
    p.add_argument('--pdf', type=Path)
    args = p.parse_args()
    m = json.loads(MANIFEST.read_text())
    validate_plan(m)
    if args.pdf:
        source_check(args.pdf)
    if not args.allow_pending:
        validate_publication(m, {})
    print('PENDING source/class checkpoint: 103 records, 278 contextual keys, 66 class/level contracts; NOT publication proof')


if __name__ == '__main__':
    main()
