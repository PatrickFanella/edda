#!/usr/bin/env python3
"""ROOT's single reconciled recovery attempt; never a general retry switch."""
import fcntl
import json
import os
from pathlib import Path
import re

from caster_native_publication import (
    APPROVAL_SHA, RENDERER_SHA, DEFAULT, Coordinator, REPO, TASK, PROJECT,
    allocation_body, atomic, check_registry, now, sha,
)

AUTH_SHA='a0f47a7feae2d058813822cca8b4d3b7f9330d63a77ac9c6155cfd467bf82c6f'
ALIAS='332:spell-guards-and-wards-dancing-lights'
INTENT='463189eda52ff60234ecb2b3557beaba52a0c0a3d887af38c58cf2d60b7d95fe'
RECOVERY=DEFAULT/'recovery-20261116-null-create'
QUERY='mutation($input:CreateIssueInput!){createIssue(input:$input){issue{number databaseId id}}}'
from caster_mutation_guard import parse_http, http_graphql, MutationGuard

def consume(path,record):
    """O_EXCL and fsync fence replay before the one external mutation."""
    descriptor=os.open(path,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
    with os.fdopen(descriptor,'w') as stream:
        json.dump(record,stream,indent=2);stream.write('\n');stream.flush();os.fsync(stream.fileno())
    descriptor=os.open(path.parent,os.O_RDONLY)
    try:os.fsync(descriptor)
    finally:os.close(descriptor)


def identity(data):
    assert isinstance(data,dict) and type(data.get('number')) is int and data['number']>0
    assert type(data.get('databaseId')) is int and data['databaseId']>0
    assert type(data.get('id')) is str and data['id']
    return dict(number=data['number'],id=data['databaseId'],node_id=data['id'])


def load_authorization(c):
    raw=(RECOVERY/'root-one-attempt-authorization.json').read_bytes()
    assert sha(raw)==AUTH_SHA, 'explicit immutable ROOT recovery authorization'
    a=json.loads(raw)
    assert a['approval_sha256']==APPROVAL_SHA and a['renderer_contract_sha256']==RENDERER_SHA
    assert a['alias']==ALIAS and a['original_intent_request_sha256']==INTENT
    assert type(a['allowed_attempts']) is int and a['allowed_attempts']==1
    assert a['status']=='ROOT_AUTHORIZED_ONE_RECONCILED_ATTEMPT' and a['not_a_milestone'] is True
    for key in ['retained_registry','retained_journal','independent_readonly_evidence']:
        assert sha(Path(a[key]['path']).read_bytes())==a[key]['sha256'], 'ROOT recovery evidence changed'
    retained=json.loads(Path(a['retained_registry']['path']).read_text())
    check_registry(retained,c.nodes,c.baseline,False)
    assert len(retained['nodes'])==714 and c.registry['nodes']==retained['nodes'], 'retained714 identities changed'
    originals=json.loads(Path(a['retained_journal']['path']).read_text())['operations']
    original=next(e for e in originals if e['operation']=='create '+ALIAS)
    current=next(e for e in c.journal['operations'] if e['operation']=='create '+ALIAS)
    assert original==current and original['request_sha256']==INTENT
    assert original['started_at']==a['original_intent_started_at']
    assert original['response']=={'data':{'createIssue':{'issue':None}}} and original['status']=='RECONCILE_REQUIRED'
    node=next(n for n in c.nodes if n['alias']==ALIAS)
    assert sha(node['approved_body'])==a['approved_body_sha256']
    assert sha(allocation_body(node,c.contract))==a['allocation_body_sha256']
    return a,node


def matching(nodes,body,title):
    found=[n for n in nodes if '<!-- edda-caster-native:'+ALIAS+' -->' in (n['body'] or '')]
    assert len(found)<=1, 'duplicate recovery marker'
    for n in found:
        assert n['body']==body and n['title']==title and n['state'].lower()=='open'
    return found


def reconcile(c):
    """Complete allocation-only graph/metadata comparison, including old graph."""
    from caster_native_publication import old_contract_matches
    c.verify_inputs()
    graph=c.full_graph();by={n['number']:n for n in graph}
    for old in c.baseline['graph']:
        assert old_contract_matches(by[old['number']],old), ('old native contract drift',old['number'])
    approved={n['alias']:n for n in c.nodes}
    metadata=c.metadata(sorted(by))
    assert set(metadata)==set(by), 'incomplete recovery metadata'
    for number,row in by.items():
        meta=metadata[number]
        observed=dict(number=meta['number'],id=meta['databaseId'],node_id=meta['id'],
                      title=meta['title'],body=meta['body'] or '',state=meta['state'].lower(),
                      parent=(meta['parent'] or {}).get('number'),
                      blocked_by=sorted(n['number'] for n in meta['blockedBy']['nodes']))
        assert observed==row, ('graph/metadata contract disagreement',number)
    for number,expected in c.baseline['project'].items():
        actual=metadata[int(number)]
        assert actual['issueType']['name']==expected['type']
        assert actual['project_item']['fieldValueByName']['name']==expected['size']
    for alias,ident in c.registry['nodes'].items():
        actual=by[ident['number']];meta=metadata[ident['number']]
        assert {k:actual[k] for k in ident}==ident
        assert actual['body']==allocation_body(approved[alias],c.contract) and actual['title']==approved[alias]['title']
        assert actual['state']=='open' and actual['parent'] is None and actual['blocked_by']==[]
        assert meta['issueType']['name']=='Task' and meta['project_item']['fieldValueByName']['name']=='M'
    node=approved[ALIAS];found=matching(graph,allocation_body(node,c.contract),node['title'])
    expected={n['number'] for n in c.baseline['graph']}|{i['number'] for i in c.registry['nodes'].values()}|{n['number'] for n in found}
    assert set(by)==expected, 'unexpected native identities'
    if found:
        actual=found[0];meta=metadata[actual['number']]
        assert actual['parent'] is None and actual['blocked_by']==[]
        assert meta['issueType']['name']=='Task' and meta['project_item']['fieldValueByName']['name']=='M'
    return graph,metadata,found


def recover(c, authorization_path, authorization_sha):
    # Keep the prior authorization, consumed record, both nulls and halt intact.
    load_authorization(c)
    assert (RECOVERY/'one-attempt-consumed.json').exists(), 'prior consumed fence missing'
    assert authorization_sha!=AUTH_SHA and re.fullmatch('[0-9a-f]{64}',authorization_sha)
    raw=authorization_path.read_bytes();assert sha(raw)==authorization_sha
    a=json.loads(raw)
    assert a['status']=='ROOT_AUTHORIZED_ONE_RECONCILED_ATTEMPT'
    assert type(a['allowed_attempts']) is int and a['allowed_attempts']==1
    assert a['approval_sha256']==APPROVAL_SHA and a['renderer_contract_sha256']==RENDERER_SHA
    assert a['alias']==ALIAS and a['original_intent_request_sha256']==INTENT
    assert a['not_a_milestone'] is True and a['preserve_active_halt'] is True
    # Bind exact reviewed code and every retained authority/evidence file.
    for key,path in {'coordinator':'caster_native_publication.py','recovery':'recover_caster_null_create.py',
                     'guard':'caster_mutation_guard.py'}.items():
        assert a['code_sha256'][key]==sha(Path(__file__).with_name(path).read_bytes())
    for name in ['native-write-halt.json','registry.json','publication-journal.json',
                 'recovery-20261116-null-create/one-attempt-consumed.json',
                 'recovery-20261116-null-create/one-attempt-result.json']:
        assert a['retained_sha256'][name]==sha((c.directory/name).read_bytes()), ('retained recovery evidence changed',name)
    halt_raw=(c.directory/'native-write-halt.json').read_bytes()
    consumed=c.directory/('recovery-'+authorization_sha+'.consumed.json')
    assert not consumed.exists(), 'new ROOT authorization already consumed'
    MutationGuard(c.journal).wait()
    graph,metadata,found=reconcile(c)
    node=next(n for n in c.nodes if n['alias']==ALIAS);body=allocation_body(node,c.contract)
    event=dict(authorization_sha256=authorization_sha,alias=ALIAS,observed_at=now(),attempts=0)
    atomic(c.directory/('recovery-'+authorization_sha+'.preflight.json'),dict(observed_at=now(),graph=graph,metadata=metadata))
    recent_response=http_graphql('{repository(owner:"subculture-collective",name:"edda"){issues(first:100,orderBy:{field:CREATED_AT,direction:DESC}){totalCount nodes{number databaseId id title body state}}}}',{})
    atomic(c.directory/('recovery-'+authorization_sha+'.recent.json'),recent_response)
    assert recent_response['exit_code']==0 and recent_response['status']==200
    recent=json.loads(recent_response['body']);assert not recent.get('errors')
    page=recent['data']['repository']['issues']
    assert page['totalCount']==len(graph), 'native graph changed since complete reconciliation; re-halt'
    recent_found=matching(page['nodes'],body,node['title'])
    assert [n['number'] for n in recent_found]==[n['number'] for n in found], 'marker changed since full reconciliation'
    MutationGuard(c.journal).wait([(recent_response, __import__('time').time())])
    c.verify_inputs()
    assert (c.directory/'native-write-halt.json').read_bytes()==halt_raw
    # Consumption fences adoption too: restart never repeats this operation.
    consume(consumed,event|dict(status='CONSUMED_BEFORE_EXTERNAL_ACTION'))
    entry=None
    try:
        if found:
            actual=metadata[found[0]['number']];ident=identity(actual)
            event['status']='ADOPTED_EXISTING_WITHOUT_CREATE'
        else:
            variables={'input':dict(repositoryId=REPO,title=node['title'],body=body,issueTypeId=TASK,projectV2Ids=[PROJECT],clientMutationId=ALIAS)}
            assert sha(json.dumps(dict(query=QUERY,variables=variables),sort_keys=True))==INTENT
            entry=dict(operation='bounded-root-recovery create '+ALIAS,authorization_sha256=authorization_sha,
                       request_sha256=INTENT,request=dict(query=QUERY,variables=variables),status='INTENT',started_at=now())
            c.journal['operations'].append(entry);atomic(c.journal_path,c.journal)
            event['attempts']=1
            response=http_graphql(QUERY,variables);entry['sanitized_http_response']=response
            assert response['exit_code']==0 and 200<=response['status']<300
            value=json.loads(response['body']);assert not value.get('errors')
            ident=identity((value.get('data') or {}).get('createIssue',{}).get('issue'))
        actual=c.metadata([ident['number']])[ident['number']]
        assert identity(actual)==ident and actual['title']==node['title'] and actual['body']==body
        assert actual['state']=='OPEN' and actual['issueType']['name']=='Task'
        assert actual['parent'] is None and actual['blockedBy']['nodes']==[]
        assert actual['project_item'] is not None, 'project membership unreadable'
        # Size M may still require the normal paced allocator after a create.
        if found:assert actual['project_item']['fieldValueByName']['name']=='M'
        c.registry['nodes'][ALIAS]=ident;c.save_registry('ALLOCATING')
        if entry:entry.update(status='CONFIRMED_TYPED_IDENTITY_AND_BODY',finished_at=now());atomic(c.journal_path,c.journal)
        event.update(status='CONFIRMED; HALT RETAINED; ROOT review required before allocation',identity=ident)
    except Exception as error:
        if entry:entry.update(status='RECONCILE_REQUIRED',finished_at=now());atomic(c.journal_path,c.journal)
        event.update(status='AMBIGUOUS; HALT RETAINED',error=type(error).__name__+': '+str(error))
        raise
    finally:
        atomic(c.directory/('recovery-'+authorization_sha+'.result.json'),event)
        assert (c.directory/'native-write-halt.json').read_bytes()==halt_raw
    return event


def main():
    import argparse
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--authorization',type=Path,required=True)
    parser.add_argument('--authorization-sha256',required=True)
    args=parser.parse_args()
    with (DEFAULT/'coordinator.lock').open('a') as lock:
        fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
        print(json.dumps(recover(Coordinator(DEFAULT),args.authorization,args.authorization_sha256),indent=2))


if __name__=='__main__':main()
