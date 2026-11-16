#!/usr/bin/env python3
"""Verify #322 source, owner matrix and native hierarchical qualification paths."""
import argparse
from copy import deepcopy
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import subprocess
from urllib.parse import parse_qs, parse_qsl, urlencode, urlsplit

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / 'docs/investigations/issue-322-bard-scope.json'
RECEIPT = ROOT / 'docs/investigations/issue-322-link-verification.json'
PDF_SHA = '8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
PROJECT = 'PVT_kwDODjtmGM4BrFJ2'


def digest(value):
    return hashlib.sha256(value.encode() if isinstance(value, str) else value).hexdigest()


def body(m, r):
    return (f"<!-- edda-322:{r['key']} -->\nParent: #{r['parent']}; decision #322; roadmap #47. Task Size {r['project_size']}.\n\n"
            f"One reviewable PR. Required path: this terminal -> Bard gate #{m['gate_issue']} -> #337. "
            f"Additional direct qualification gates: {r['additional_gates']}.\n\n"
            f"Source: pinned English SRD5.2.1 PDF31-35; exact referenced pages in scope below. SHA256 {PDF_SHA}. "
            "Attribution/license/adaptation: docs/srd-5.2.1-conformance.md; CC-BY-4.0. Source facts paraphrased; fixture/ownership choices are Edda additions.\n\n"
            f"Scope: {r['fixture']}\n\nPositive deterministic fixture: {r['positive']}\n\nNegative deterministic fixture: {r['negative']}\n\n"
            f"Required acceptance: {m['common_acceptance']}\n\n"
            "Feature tests use explicit isolated entitlements, never production grants. Transitions consume exact implemented contracts and apply grants once. "
            "Unknown options/effects/origins/multiclass configurations fail closed. #272/#427/#428 own completed rests; #273 casting modes, #58 sheet projection, "
            "#333/#430 proficiency provenance, #336/#429 advancement transaction, #335 multiclass contracts, #384 selection/#389 caps. "
            "#277 spells/#278 items/#269/#270 conditions retain shared effects. Chosen feat/spell effects require their exact terminal and qualification path "
            "before availability; source import or planning closure never supplies execution. No runtime/grants/coverage promotion or deployment in this decision.\n\n"
            f"Native prerequisites: {', '.join('#'+str(n) for n in r['blocked_by_issues'])}.\n\n"
            "Owned mechanic keys:\n" + '\n'.join('- '+k for k in r['owned_keys']) +
            '\n\nExact contextual source records (shared references do not duplicate effect ownership):\n```json\n' +
            json.dumps(r['source_records'], indent=2) + '\n```\n')


class Reader:
    def __init__(self):
        self.calls = 0

    def __call__(self, path, data=None):
        self.calls += 1
        assert self.calls <= 350, 'read-only API call cap exceeded'
        cmd = ['gh', 'api', path]
        if data is not None:
            assert path == 'graphql' and 'mutation' not in data['query'], 'read-only verifier'
            cmd += ['--input', '-']
        p = subprocess.run(cmd, input=json.dumps(data) if data is not None else None,
                           text=True, capture_output=True, check=True)
        value = json.loads(p.stdout)
        assert not (isinstance(value, dict) and value.get('errors')), value
        return value


def pages(api, endpoint):
    parts = urlsplit(endpoint)
    params = dict(parse_qsl(parts.query))
    assert 'page' not in params and 'per_page' not in params
    for page in range(1, 21):
        batch = api(parts.path + '?' + urlencode({**params, 'per_page': 100, 'page': page}))
        assert isinstance(batch, list)
        yield from batch
        if len(batch) < 100:
            return
    raise AssertionError('pagination cap exceeded')


def parent_path(graph, number):
    path = []
    while number is not None:
        assert number not in path, 'parent cycle'
        assert number in graph, 'missing parent graph node'
        path.append(number)
        number = graph[number]['parent']
    return path


def live(m):
    api = Reader()
    graph, cursor = [], None
    for _ in range(20):
        q = '''query($cursor:String){repository(owner:"subculture-collective",name:"edda"){
        issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{
        number parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}}}}}'''
        b = api('graphql', {'query': q, 'variables': {'cursor': cursor}})['data']['repository']['issues']
        for node in b['nodes']:
            assert not node['blockedBy']['pageInfo']['hasNextPage'], 'unexpected blocker pagination/cap'
            graph.append(dict(number=node['number'], parent=(node['parent'] or {}).get('number'),
                              blocked_by=sorted(x['number'] for x in node['blockedBy']['nodes'])))
        if not b['pageInfo']['hasNextPage']:
            break
        cursor = b['pageInfo']['endCursor']
    else:
        raise AssertionError('graph pagination cap')
    bodies, project = {}, {}
    rows = m['issues'] + [dict(issue=m['gate_issue'], title=m['gate_title'], project_size='M')]
    for r in rows:
        n = r['issue']
        x = api(f'repos/subculture-collective/edda/issues/{n}')
        want = m['gate_body'] if n == m['gate_issue'] else body(m, r)
        assert x['title'] == r['title'] and x['body'] == want, f'body/title drift #{n}'
        bodies[str(n)] = dict(title=x['title'], body=x['body'], sha256=digest(x['body']), updated_at=x['updated_at'])
    metadata_rows = rows + [dict(issue=n) for n in m['reused_terminals']]
    for start in range(0, len(metadata_rows), 40):
        batch = metadata_rows[start:start+40]
        fields = ' '.join('i%d:issue(number:%d){issueType{name} projectItems(first:20){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}' % (r['issue'], r['issue']) for r in batch)
        data = api('graphql', {'query': '{repository(owner:"subculture-collective",name:"edda"){' + fields + '}}'})['data']['repository']
        for r in batch:
            n = r['issue']
            node = data[f'i{n}']
            assert not node['projectItems']['pageInfo']['hasNextPage']
            items = [i for i in node['projectItems']['nodes'] if i['project']['id'] == PROJECT]
            assert len(items) <= 1
            project[str(n)] = dict(type=(node['issueType'] or {}).get('name'), size=(items[0]['fieldValueByName'] or {}).get('name') if items else None)
    existing = {}
    for n in m['preserved_issues']:
        x = api(f'repos/subculture-collective/edda/issues/{n}')
        existing[str(n)] = dict(title=x['title'], sha256=digest(x['body'] or ''), state=x['state'])
    supplements = {}
    for s in m['supplements']:
        x = api(f"repos/subculture-collective/edda/issues/{s['issue']}")
        count = (x['body'] or '').count(s['append'])
        assert count <= 1, 'duplicate marked parent supplement'
        supplements[str(s['issue'])] = dict(status='published' if count == 1 else 'parent-body pending', append_sha256=digest(s['append']), observed_body_sha256=digest(x['body'] or ''))
    indexed = {r['number']:r for r in graph}
    terminals = sorted({r['issue'] for r in m['issues']} | set(m['reused_terminals']))
    return dict(schema_version=1, observed_at=datetime.now(timezone.utc).isoformat(), method='Read-only GraphQL full repository parents/dependencies and REST exact bodies; Project fields', graph=sorted(graph, key=lambda x:x['number']), bodies=bodies, project=project, existing=existing, supplements=supplements, api_calls=api.calls,
                parent_paths={str(n):parent_path(indexed, n) for n in terminals + [m['gate_issue'], 322]},
                qualification_paths={str(n):[n, m['gate_issue'], 337] for n in terminals})


def validate(m, receipt):
    assert m['source']['sha256'] == PDF_SHA, 'source pin mismatch'
    records = []
    for path, c in m['catalogues'].items():
        assert json.loads((ROOT/path).read_text()) == c, 'source record mismatch'
        assert c['source']['sha256'] == PDF_SHA
        if 'advancement' not in path:
            records += c['records']
        else:
            assert len(c['records']) == 20
    assert len(records) == 84
    assert {r['level'] for r in records if r['kind'] == 'level'} == set(range(1, 21))
    assert set(m['record_evidence']) == {r['id'] for r in records}, 'missing source record'
    keys = {k for r in records for k in r['required_mechanics']}
    assert set(m['source_key_owners']) == keys, 'missing source key'
    rows = {r['key']:r for r in m['issues']}
    assert len(rows) == len(m['issues']) == 48
    assert set(m['source_key_owners'].values()) <= rows.keys()
    feats = json.loads((ROOT/'docs/investigations/issue-334-feat-scope.json').read_text())
    feat_rows = {r['key']:r['issue'] for r in feats['issues']}
    assert m['feat_effect_choices'] == [dict(id=f['id'], category=f['category'], effect_issue=feat_rows[f['effect_owner']], source_records=f['records'], eligibility='Only with the actual source-linked Bard grant and exact #384 prerequisites; unsupported origin/multiclass/feat variant remains unavailable.') for f in feats['feats']], 'feat source/owner drift'
    for key in ['improvement', 'boon']:
        assert {f['effect_issue'] for f in m['feat_effect_choices']} <= set(rows[key]['prerequisites']), 'missing feat terminal'
    assert {f['effect_issue'] for f in m['feat_effect_choices']} <= set(m['reused_terminals'])
    canonical = {r['id']:r for c in m['catalogues'].values() for r in c['records']}
    for row in rows.values():
        assert all(canonical[r['id']] == r for r in row['source_records']), 'contextual source drift'
        if row['key'].startswith('level-'):
            level = int(row['key'].split('-')[1])
            assert row['source_records'] == [canonical[f'srd-5.2.1:character/level/bard/{level}'], canonical[f'srd-5.2.1:character/advancement/{level}']]
    for record in records:
        e = m['record_evidence'][record['id']]
        assert e['owners'] == sorted({m['source_key_owners'][k] for k in record['required_mechanics']})
        assert e['state'].startswith('unavailable') and all((ROOT/p).is_file() for p in e['proof'])
    assert m['option_key_consumers'] == {
        r['id']: {k:m['record_evidence'][r['id']]['owners'] for k in r['option_keys']}
        for r in records if r['option_keys']
    }, 'missing nested option consumer'
    for p, h in m['evidence_hashes'].items():
        assert digest((ROOT/p).read_bytes()) == h, f'evidence drift {p}'
    g = {r['number']:r for r in receipt['graph']}
    active, done = set(), set()
    def visit(n):
        assert n not in active, f'cycle at #{n}'
        if n in done:
            return
        assert n in g, 'missing graph node'
        active.add(n)
        for b in g[n]['blocked_by']:
            visit(b)
        active.remove(n)
        done.add(n)
    for n in g:
        visit(n)
    for n in g:
        chain, current = set(), n
        while current in g and g[current]['parent'] is not None:
            assert current not in chain, 'parent cycle'
            chain.add(current)
            current = g[current]['parent']
    assert g[322]['parent'] == 283 and {285, 287, 309} <= set(g[322]['blocked_by']), 'missing #322 source prerequisite'
    assert all(receipt['existing'][str(n)]['state'] == 'closed' for n in [285, 287, 309])
    assert g[337]['parent'] == 283 and 322 in g[337]['blocked_by']
    assert g[334]['parent'] == 283 and 334 in g[337]['blocked_by'], 'historical feat decision gate weakened'
    assert {298, 299, 300, 301} <= set(g[334]['blocked_by']), 'historical feat source edges weakened'
    gate = m['gate_issue']
    assert g[gate]['parent'] == 283 and gate in g[337]['blocked_by'], 'missing Bard gate path'
    terminals = {r['issue'] for r in rows.values()} | set(m['reused_terminals'])
    assert len(terminals) < 100 and set(g[gate]['blocked_by']) == terminals | {322}, 'missing/extra Bard gate terminal'
    assert receipt['parent_paths'] == {str(n):parent_path(g, n) for n in terminals | {gate, 322}}, 'parent path drift'
    assert receipt['qualification_paths'] == {str(n):[n, gate, 337] for n in terminals}, 'qualification path drift'
    assert set(m['preserved_direct_337']) <= set(g[337]['blocked_by']), 'historical feat gate weakened'
    for r in rows.values():
        n = r['issue']
        assert g[n]['parent'] == r['parent'], 'wrong native parent'
        assert set(r['blocked_by_issues']) == set(r['prerequisites']) | {rows[k]['issue'] for k in r['depends']}
        assert set(g[n]['blocked_by']) == set(r['blocked_by_issues']), 'missing/extra child prerequisite'
        assert r['owned_keys'] == sorted(k for k, o in m['source_key_owners'].items() if o == r['key'])
        for other in r['additional_gates']:
            assert n in g[other]['blocked_by'], 'missing cross-domain gate'
        saved = receipt['bodies'][str(n)]
        assert saved['title'] == r['title'] and saved['body'] == body(m, r) and saved['sha256'] == digest(body(m, r)), 'body hash mismatch'
        assert receipt['project'][str(n)] == dict(type='Task', size=r['project_size'])
    assert receipt['bodies'][str(gate)]['body'] == m['gate_body'] and receipt['bodies'][str(gate)]['sha256'] == digest(m['gate_body'])
    assert receipt['project'][str(gate)] == dict(type='Task', size='M')
    assert set(receipt['project']) == {str(n) for n in terminals | {gate}}, 'missing Project metadata readback'
    for n, expected in m['preserved_issues'].items():
        assert g[int(n)]['parent'] == expected['parent'], 'existing parent changed'
        assert set(g[int(n)]['blocked_by']) == set(expected['blocked_by']), 'existing prerequisites changed'
        assert receipt['existing'][n]['sha256'] == expected['sha256'], 'existing body changed since evidence seal'
    for s in m['supplements']:
        assert (ROOT/s['path']).read_text() == s['append']
        assert receipt['supplements'][str(s['issue'])]['append_sha256'] == digest(s['append'])
        assert receipt['supplements'][str(s['issue'])]['status'] in ['parent-body pending', 'published']
    return len(terminals), len(keys), len(g)


def probes(m, receipt):
    def rejects(label, change):
        a, b = deepcopy(m), deepcopy(receipt)
        change(a, b)
        try:
            validate(a, b)
        except (AssertionError, KeyError):
            return label
        raise AssertionError('corruption accepted: ' + label)
    n, gate = m['issues'][0]['issue'], m['gate_issue']
    tests = [('source mismatch', lambda a,b:a['catalogues'][next(iter(a['catalogues']))]['records'][0].update(name='wrong')),
             ('missing source key', lambda a,b:a['source_key_owners'].pop(next(iter(a['source_key_owners'])))),
             ('missing child', lambda a,b:a['issues'].pop()),
             ('missing terminal gate', lambda a,b:next(x for x in b['graph'] if x['number']==gate)['blocked_by'].remove(n)),
             ('missing gate to337', lambda a,b:next(x for x in b['graph'] if x['number']==337)['blocked_by'].remove(gate)),
             ('wrong body hash', lambda a,b:b['bodies'][str(n)].update(sha256='wrong')),
             ('dependency cycle', lambda a,b:next(x for x in b['graph'] if x['number']==n)['blocked_by'].append(gate))]
    tests += [('parent cycle', lambda a,b:next(x for x in b['graph'] if x['number']==283).update(parent=gate)),
              ('missing child prerequisite', lambda a,b:next(x for x in b['graph'] if x['number']==n)['blocked_by'].pop()),
              ('missing cross-domain gate', lambda a,b:next(x for x in b['graph'] if x['number']==279)['blocked_by'].remove(n))]
    for source in [285, 287, 309]:
        tests.append((f'missing source edge {source}', lambda a,b,s=source:next(x for x in b['graph'] if x['number']==322)['blocked_by'].remove(s)))
    tests.append(('missing legacy feat source edge', lambda a,b:next(x for x in b['graph'] if x['number']==334)['blocked_by'].remove(298)))
    results = [rejects(label, change) for label, change in tests]
    calls = []
    def synthetic(url):
        q = parse_qs(urlsplit(url).query)
        page = int(q['page'][0])
        assert q['state'] == ['all'] and q['per_page'] == ['100']
        calls.append(page)
        return list(range(100)) if page < 10 else [1000]
    assert len(list(pages(synthetic, 'repos/o/r/issues?state=all'))) == 901
    assert calls == list(range(1, 11))
    capped = []
    def endless(url):
        capped.append(int(parse_qs(urlsplit(url).query)['page'][0]))
        return list(range(100))
    try:
        list(pages(endless, 'repos/o/r/issues?state=all'))
    except AssertionError as error:
        assert str(error) == 'pagination cap exceeded'
    else:
        raise AssertionError('unbounded pagination accepted')
    assert capped == list(range(1, 21))
    print('PASS corruption probes: ' + ', '.join(results) + '; bounded exact-query pagination')


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--live', action='store_true')
    p.add_argument('--pdf', type=Path)
    p.add_argument('--probes', action='store_true')
    p.add_argument('--require-supplements', action='store_true', help='Parent acceptance: require all four exact marked sections live')
    args = p.parse_args()
    m = json.loads(MANIFEST.read_text())
    if args.pdf:
        assert digest(args.pdf.read_bytes()) == PDF_SHA, 'PDF source mismatch'
    r = live(m) if args.live else json.loads(RECEIPT.read_text())
    counts = validate(m, r)
    if args.require_supplements:
        assert all(s['status'] == 'published' for s in r['supplements'].values()), 'parent-body publication pending'
    if args.probes:
        probes(m, r)
    if args.live:
        RECEIPT.write_text(json.dumps(r, indent=2) + '\n')
    print(f'PASS {counts[0]} mandatory terminals, {counts[1]} exact keys, {counts[2]} acyclic nodes; exact native bodies/parents/prerequisites/Project; parent-body status: ' + str({k:v['status'] for k,v in r['supplements'].items()}))


if __name__ == '__main__':
    main()
