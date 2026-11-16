#!/usr/bin/env python3
"""Read-only #330 proof. A pending checkpoint cannot pass publication validation."""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import subprocess
from urllib.parse import parse_qs, urlsplit

from sorcerer_contract_oracle import (
    COMMON, CONTRACTS, DEPENDENCIES, GRANTS, SOURCE_LISTS, FINITE_CONSUMED, SUMMON_DRAGON, execution_owner, level_fixture,
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
    """Exact released prefix/owned section; unrelated later pure appends allowed."""
    import re
    assert body.startswith(prefix), 'modified or rolled-back prefix'
    tail=body[len(prefix):]
    assert tail.count(append)==1, 'missing/partial/modified append'
    if marker.startswith('<!--'):
        namespace=re.search(r'edda-\d+',marker,re.I)
        assert namespace, 'owned marker namespace'
        pattern=r'<!--\s*'+re.escape(namespace[0])
        assert len(re.findall(pattern,tail,re.I))==len(re.findall(pattern,append,re.I)), 'duplicate or partial/altered owned marker'
    else:
        stem=marker[:-1] if marker.endswith('>') else marker
        assert tail.count(stem)==append.count(stem), 'duplicate or partial append marker'
    assert tail.count(marker)==1, 'duplicate append marker'


def expected_body(t):
    return ('Decision #330; one implementation PR; Task Size M.\n\n'
            'Exact execution scope and positive/independently violated negative controls:\n'
            + t['fixture'] + '\n\n' + COMMON
            + '\n\nExact native prerequisite numbers: ' + str(t['prerequisites'])
            + '; planned owned prerequisites: ' + str(t['depends'])
            + '.\n\nOwned source contexts:\n```json\n'
            + json.dumps(t['owned_contexts'], indent=2) + '\n```\n')


def summon_body(fixture):
    return ('Decision #330; shared Summon Dragon bounded execution owner; one Task/M PR.\n\n'
            + fixture + '\n\n' + COMMON
            + '\n\nRequired terminal routing: finite shared cohort->Sorcerer shared output->three-caster aggregate->279/280 and own class gate->337; exact world/statblock leaf qualification also required at183. No185/274planning substitution. Native publication PENDING.\n')


# Filled only from an explicit root release, never candidate receipts.
ACCEPTED_PREFIX_SHA256 = {'47': 'b471db938c1d2357fee6bf70f03e19529f6d991422f8a978820874e3c8e2c842', '54': 'bd76c98ef2c4f8710f359cae305bc462770c637309de90a560b96ba07bd96ae4', '283': 'a643f239df6a0eef4abaa685c0038a6c00b512c8cb6b93f307599bf392d13010', '337': '328223c535edca542515767b900738d33beb301ade6f22865d49782bffc77c44'}
APPROVED_EXTERNAL_HEADS = {}
BASELINE = ROOT / 'docs/investigations/issue-330-native-baseline.json'


def validate_plan(m):
    from sorcerer_scope_plan import complete_plan, transition_fixtures
    from sorcerer_spell_oracle import CONFLICT_POLICY, REUSED_OUTPUTS
    assert exact(m['spell_lists'], SOURCE_LISTS), 'complete class source lists'
    assert exact(m['future_source_rosters']['consumed'], FINITE_CONSUMED), 'finite future consumed source roster'
    assert m['future_source_rosters']['source_union'] == sorted(set.union(*(set(v) for v in FINITE_CONSUMED.values()))), 'complete finite source union'
    assert len(m['future_source_rosters']['source_union']) == 245
    assert m['decision'] == 330 and m['base'] == BASE and m['source']['sha256'] == PIN
    for path, digest in m['evidence_hashes'].items():
        assert hashlib.sha256((ROOT / path).read_bytes()).hexdigest() == digest, 'caller evidence changed'
    advancement='internal/rules/srd521/testdata/character-advancement.json'
    assert exact(m['catalogues'][advancement],json.loads((ROOT/advancement).read_text())), 'COMPLETE authoritative advancement table'
    assert exact(m['transition_fixtures'],transition_fixtures(ROOT)), 'typed class/total/XP/PB/HP operand witnesses'
    actual = []
    for name, count in [('sorcerer',63),('metamagic',22),('draconic',18)]:
        path = 'internal/rules/srd521/testdata/character-' + name + '.json'
        data = json.loads((ROOT/path).read_text())
        assert exact(m['catalogues'][path], data), 'source operands/types'
        assert len(data['records']) == count
        actual.extend(data['records'])
    assert exact(actual,m['records']) and len(actual)==103, 'complete source records'
    assert len(m['record_matrix'])==103
    innate=next(t for t in m['tasks'] if t['key']=='innate-benefits')
    witness=re.search(r'Positive:\s*CHA([+-]?\d+) PB(\d+) DC(\d+)->(\d+)',innate['fixture'])
    assert witness, 'Innate Sorcery DC witness'
    cha,pb,before,after=map(int,witness.groups())
    assert (before,after)==(8+cha+pb,9+cha+pb), 'Innate Sorcery DC formula'
    expected_tasks, expected_shared, expected_gates, routes = complete_plan(actual)
    for record, context in zip(actual,m['record_matrix']):
        assert exact(context['record'],record) and context['id']==record['id']
        expected={k:execution_owner(record,k) for k in record['required_mechanics']+record['option_keys']}
        assert context['key_owners']==expected, 'independent source execution owner'
        effects={name:owners for name,owners in routes.items() if any(k=='srd-5.2.1:mechanic/spells/'+name.lower().replace('’','').replace('/','-').replace(' ','-') for k in record['required_mechanics'])}
        assert exact(context['effect_owners'],effects), 'nested actual spell execution owners'
        assert context['status']=='unavailable'
        assert all(context[k] for k in ['rules','game','engine','combat','handlers','database','web','tui'])
    for field,expected in [('tasks',expected_tasks),('shared_split_proposal',expected_shared),('gates',expected_gates)]:
        candidates=m[field]
        assert len(candidates)==len(expected) and len({t['key'] for t in candidates})==len(expected), 'complete finite contract roster'
        for t,e in zip(candidates,expected):
            assert set(t)==set(e), 'complete contract fields'
            for key in e:
                if key=='issue':
                    assert t[key] is None or type(t[key]) is int and t[key]>0, 'exact native numeric identity'
                    continue
                diagnostic = 'complete body semantics' if key=='body' else 'actual execution semantics / exact protocol prerequisites / all20 granting paths'
                assert exact(t[key],e[key]), diagnostic+': '+t['key']+'/'+key
    assert exact(m['spell_execution_owners'],routes), 'complete source-to-execution routes'
    assert m['overlap']['title_candidates']==REUSED_OUTPUTS, 'exact reused outputs; no title aliases'
    baseline=json.loads(BASELINE.read_text())
    assert exact(m['overlap_audit'],baseline['overlap_audit']), 'whole reused contract body/title/parent/prerequisites'
    assert exact(m['protected_owners'],baseline['protected_owners']), 'protected source owner'
    assert m['protected_owners']['779']['body_sha256']=='556d1618d1cd2bd9c956fe38861d89e94ee14ff61a505b5d2b6a1fa46d19d150'
    assert m['protected_owners']['779']['blocked_by']==[61,64,65,66,67,488,778]
    assert set(m['source_conflicts'])==set(CONFLICT_POLICY), 'complete source conflict ledger'
    for key,expected in CONFLICT_POLICY.items():
        entry=m['source_conflicts'][key]
        assert all(exact(entry[k],v) for k,v in expected.items()), 'source conflict locators'
        assert entry['feature_exception']=='Only a separately source-qualified feature route with its own exact owner/proof may admit', 'exact separate feature-route policy'
        assert entry['effect_owner']==('Warlock331 single agreeing printed75/header149 effect owner reserved' if entry['spell']=='Mind Spike' else 'unallocated; outside245 reservation'), 'single effect owner reservation'
        assert entry['admission']=='unavailable for disputed list-based route', 'header cannot grant admission'
        assert entry['spell'] not in [x['name'] for x in SOURCE_LISTS[entry['class']]], 'disputed printed admission'
    assert exact(m['shared_counts'],dict(spell_names=98,spell_parts=204,seams=3,casting_protocols=4,shared_runtime_leaves=211,qualification_nodes=27,class_leaves=67)), 'exact source-derived counts/types'
    allrows=expected_tasks+expected_shared+expected_gates
    acyclic({t['key']:t['depends'] for t in allrows})
    return {t['key']:t for t in m['tasks']+m['shared_split_proposal']+m['gates']}


def project_owned(m):
    rr=validate_plan(m)
    baseline=json.loads(BASELINE.read_text())
    graph={str(t['number']):[str(n) for n in t['blocked_by']] for t in baseline['graph']}
    parents={str(t['number']):str(t['parent']) for t in baseline['graph'] if t['parent']}
    for key,t in rr.items():
        assert key not in graph
        graph[key]=list(t['depends'])+[str(n) for n in t['prerequisites']]
        parents[key]=str(t['parent'])
    graph['279'].append('shared-aggregate')
    graph['280'].append('shared-aggregate')
    graph['337'].append('class-gate')
    # The two foreign heads are explicit unresolved interfaces, not zero-input
    # publication gates. No final cap/publication proof may accept this boundary.
    unresolved=[]
    for key,parent in [('warlock-shared-output',331),('wizard-shared-output',332)]:
        if key not in APPROVED_EXTERNAL_HEADS:
            unresolved.append(key)
        else:
            head=APPROVED_EXTERNAL_HEADS[key]
            assert head['parent']==parent and head['blocked_by'], 'genuine nonempty external structure'
            graph[key]=[str(n) for n in head['blocked_by']]
            parents[key]=str(parent)
    native_caps(graph,parents)
    outgoing=Counter(n for ds in graph.values() for n in ds)
    children=Counter(parents.values())
    return dict(status='PENDING external complete input structures' if unresolved else 'complete proposed graph',
                owned_nodes=len(rr),class_leaves=len(m['tasks']),shared_leaves=len(m['shared_split_proposal']),qualification_nodes=len(m['gates']),
                incoming279=len(graph['279']),incoming280=len(graph['280']),children330=children['330'],children283=children['283'],
                outgoing429=outgoing['429'],outgoing487=outgoing['487'],outgoing67=outgoing['67'],
                max_incoming=max(map(len,graph.values())),max_outgoing=max(outgoing.values()),max_children=max(children.values()),
                unresolved_external_heads=unresolved,dependency_dag='acyclic at explicit external boundary',containment_dag='acyclic at explicit external boundary',
                graph=graph,parents=parents)


def native_body(t, identities):
    parent=identities[t['parent']] if isinstance(t['parent'],str) else t['parent']
    blockers=sorted(t['prerequisites']+[identities[d] for d in t['depends']])
    return (t['body']+'\nExact native containment parent: #'+str(parent)
            +'\nExact native blocked-by at publication: '+', '.join('#'+str(n) for n in blockers)+'\n')


def typed_native_graph(receipt):
    for t in receipt['graph']:
        assert type(t['number']) is int and t['number']>0, 'exact native number type'
        assert t['parent'] is None or type(t['parent']) is int and t['parent']>0, 'exact native parent type'
        assert all(type(n) is int and n>0 for n in t['blocked_by']), 'exact native prerequisite types'
    graph={t['number']:t for t in receipt['graph']}
    assert len(graph)==len(receipt['graph']), 'duplicate native node'
    return graph


def validate_publication(m,receipt,pdf=Path('/tmp/edda-wave1/SRD_CC_v5.2.1.pdf')):
    rr=validate_plan(m)
    assert m['status']=='PUBLISHED', 'PENDING checkpoint cannot qualify publication'
    assert all(type(t['issue']) is int for t in rr.values()), 'native publication identities incomplete'
    from sorcerer_native_binding import validate_native
    validate_native(m, receipt, parent_supplements=True)
    graph=typed_native_graph(receipt)
    assert graph[330]['parent']==283 and graph[330]['blocked_by']==[285,295,305,317]
    for n in [285,295,305,317]:
        assert graph[n]['state'].lower()=='closed'
    source_check(pdf)
    return True


def read_live(m):
    from datetime import datetime, timezone
    from verify_druid_scope import Reader
    api=Reader()
    graph=[]
    cursor=None
    cursors=set()
    numbers=set()
    for _ in range(25):
        query='query($cursor:String){repository(owner:"subculture-collective",name:"edda"){issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{number databaseId id title body state parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}}}}}'
        page=api('graphql',{'query':query,'variables':{'cursor':cursor}})['data']['repository']['issues']
        for node in page['nodes']:
            assert type(node['number']) is int and node['number'] not in numbers, 'repeated native graph page'
            numbers.add(node['number'])
            assert not node['blockedBy']['pageInfo']['hasNextPage'], 'incoming over100 or truncated'
            graph.append(dict(number=node['number'],id=node['databaseId'],node_id=node['id'],title=node['title'],body=node['body'] or '',state=node['state'].lower(),parent=(node['parent']or{}).get('number'),blocked_by=sorted(n['number'] for n in node['blockedBy']['nodes'])))
        info=page['pageInfo']
        if not info['hasNextPage']:
            break
        cursor=info['endCursor']
        assert cursor and cursor not in cursors, 'repeated native cursor'
        cursors.add(cursor)
    else:
        raise AssertionError('native graph call cap')
    nums=sorted({t['issue'] for t in m['tasks']+m['shared_split_proposal']+m['gates'] if t['issue'] is not None}|{int(n) for n in m['protected_owners']})
    if m['status']=='PUBLISHED':
        from sorcerer_native_binding import pinned_inputs, registry_check, evidence
        approval, baseline = pinned_inputs()
        registry=json.loads(evidence('registry.json'))
        registry_check(registry,approval,baseline)
        nums=sorted({v['number'] for v in registry['nodes'].values()} | set(map(int,baseline['project'])))
    project={}
    for start in range(0,len(nums),25):
        batch=nums[start:start+25]
        fields=' '.join('i%d:issue(number:%d){issueType{name} projectItems(first:100){pageInfo{hasNextPage}nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}'%(n,n) for n in batch)
        data=api('graphql',{'query':'{repository(owner:"subculture-collective",name:"edda"){'+fields+'}}'})['data']['repository']
        for n in batch:
            node=data['i'+str(n)]
            assert not node['projectItems']['pageInfo']['hasNextPage'], 'project pagination cap'
            items=[i for i in node['projectItems']['nodes'] if i['project']['id']=='PVT_kwDODjtmGM4BrFJ2']
            assert len(items)==1 and node['issueType'] and items[0]['fieldValueByName'], 'native type/Size missing'
            project[str(n)]=dict(type=node['issueType']['name'],size=items[0]['fieldValueByName']['name'])
    return dict(status='PENDING' if m['status']!='PUBLISHED' else 'readback',observed_at=datetime.now(timezone.utc).isoformat(),graph=graph,project=project,api_calls=api.calls,publication_complete=False,parent_body_seals={})


def validate_snapshot(m,receipt):
    validate_plan(m)
    graph=typed_native_graph(receipt)
    baseline=json.loads(BASELINE.read_text())
    for n,old in baseline['protected_owners'].items():
        live=graph[int(n)]
        assert all(live[k]==old[k] for k in ['title','body','parent','blocked_by']), 'live full source-owner contract drift: '+n
    for old in baseline['graph']:
        live=graph[old['number']]
        assert live['parent']==old['parent'] and set(old['blocked_by'])<=set(live['blocked_by']), 'old native path changed'
    assert graph[330]['parent']==283 and graph[330]['blocked_by']==[285,295,305,317]
    assert all(graph[n]['state']=='closed' for n in [285,295,305,317])
    native_caps({n:r['blocked_by'] for n,r in graph.items()},{n:r['parent'] for n,r in graph.items() if r['parent']})
    return True


def source_check(pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest() == PIN
    extracted = subprocess.run(['pdftotext', '-f', '166', '-l', '166', str(pdf), '-'], check=True, capture_output=True, text=True).stdout
    flat = ' '.join(extracted.replace('\u00ad', '').split())
    for clause in ['AC 14 + the spell’s level', '50 + 10 for each spell level above 5',
                   'PB equals your Proficiency Bonus', 'takes the Dodge action',
                   'equal to half the spell’s level (round down)',
                   '1d6 + 4 + the spell’s level', 'each creature in a 30-foot Cone',
                   'The creature disappears when it drops to 0 Hit Points or when the spell ends.',
                   'until the spell ends']:
        assert clause in flat, 'Summon Dragon source clause: ' + clause
    from sorcerer_source_reconciliation import reconcile
    reconciliation=reconcile(pdf)
    assert len(reconciliation['conflicts'])==3
    for name in ['sorcerer', 'metamagic', 'draconic', 'advancement', 'multiclass',
                 'origin_feats', 'general_feats', 'fighting_styles', 'epic_boon_feats']:
        subprocess.run(['python3', str(ROOT / 'scripts' / ('verify_srd_' + name + '.py')), str(pdf)], check=True)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--allow-pending', action='store_true', help='Audit source/class checkpoint only; never publication proof')
    p.add_argument('--pdf', type=Path)
    p.add_argument('--receipt', type=Path, default=ROOT/'docs/investigations/issue-330-link-verification.json')
    p.add_argument('--live', action='store_true', help='Bounded read-only native full-body/type/Size snapshot')
    p.add_argument('--write-receipt', type=Path, help='Save observed native evidence, never a publication assertion')
    args = p.parse_args()
    m = json.loads(MANIFEST.read_text())
    validate_plan(m)
    if args.pdf and args.allow_pending:
        source_check(args.pdf)
    receipt=json.loads(args.receipt.read_text()) if args.receipt.exists() else {}
    if args.live:
        receipt=read_live(m)
        validate_snapshot(m,receipt)
    if not args.allow_pending:
        validate_publication(m, receipt, args.pdf or Path('/tmp/edda-wave1/SRD_CC_v5.2.1.pdf'))
        receipt['publication_complete']=True
        receipt['status']='verified source/native publication; runtime remains unqualified'
    if args.write_receipt:
        args.write_receipt.write_text(json.dumps(receipt,indent=2,ensure_ascii=False)+'\n')
    print('PENDING source/contract proof: 103 records, 278 contexts, 67 class leaves, 211 shared leaves, 27 gates; NOT native publication proof' if args.allow_pending else 'PASS: complete source/native publication proof; no runtime or Full Rules qualification')


if __name__ == '__main__':
    main()
