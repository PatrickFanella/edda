#!/usr/bin/env python3
"""ROOT's single reconciled recovery attempt; never a general retry switch."""
import fcntl
import json
import os
from pathlib import Path
import re
import subprocess

from caster_native_publication import (
    APPROVAL_SHA, RENDERER_SHA, DEFAULT, Coordinator, REPO, TASK, PROJECT,
    allocation_body, atomic, check_registry, now, sha,
)

AUTH_SHA='a0f47a7feae2d058813822cca8b4d3b7f9330d63a77ac9c6155cfd467bf82c6f'
ALIAS='332:spell-guards-and-wards-dancing-lights'
INTENT='463189eda52ff60234ecb2b3557beaba52a0c0a3d887af38c58cf2d60b7d95fe'
RECOVERY=DEFAULT/'recovery-20261116-null-create'
QUERY='mutation($input:CreateIssueInput!){createIssue(input:$input){issue{number databaseId id}}}'
SAFE_HEADERS={'date','x-github-request-id','x-ratelimit-limit','x-ratelimit-remaining',
              'x-ratelimit-used','x-ratelimit-reset','x-ratelimit-resource','retry-after'}


def parse_http(raw):
    """Keep only response metadata; never retain unfiltered HTTP headers."""
    head,separator,body=raw.replace('\r\n','\n').partition('\n\n')
    assert separator and re.match(r'^HTTP/\S+ \d{3}',head), 'HTTP response envelope missing'
    lines=head.splitlines();status=int(lines[0].split()[1]);headers={}
    for line in lines[1:]:
        key,colon,value=line.partition(':')
        if colon and key.lower() in SAFE_HEADERS:
            headers[key.lower()]=value.strip()
    return dict(status=status,headers=headers,body=body)


def http_graphql(query,variables):
    result=subprocess.run(['gh','api','graphql','--include','--input','-'],
                          input=json.dumps(dict(query=query,variables=variables)),text=True,
                          capture_output=True,timeout=90)
    response=parse_http(result.stdout)
    response['exit_code']=result.returncode
    return response


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


def recover(c):
    a,node=load_authorization(c)
    consumed=RECOVERY/'one-attempt-consumed.json'
    assert not consumed.exists(), 'ROOT authorization already consumed; no second attempt'
    assert not (DEFAULT/'native-write-halt.json').exists(), 'native writes held'
    c.verify_inputs()
    graph=c.full_graph();by={n['number']:n for n in graph}
    for old in c.baseline['graph']:
        assert all(by[old['number']][k]==v for k,v in old.items()), 'old native contract drift'
    approved={n['alias']:n for n in c.nodes}
    for alias,ident in c.registry['nodes'].items():
        actual=by[ident['number']]
        assert {k:actual[k] for k in ident}==ident
        assert actual['body']==allocation_body(approved[alias],c.contract) and actual['title']==approved[alias]['title']
        assert actual['state']=='open' and actual['parent'] is None and actual['blocked_by']==[]
    body=allocation_body(node,c.contract)
    found=matching(graph,body,node['title'])
    assert len(graph)==1731+len(found), 'unexpected native identities before recovery'
    recent_query='{repository(owner:"subculture-collective",name:"edda"){issues(first:100,orderBy:{field:CREATED_AT,direction:DESC}){totalCount nodes{number databaseId id title body state}}}}'
    recent_response=http_graphql(recent_query,{})
    atomic(RECOVERY/'coordinator-prewrite-recent.json',dict(observed_at=now(),response=recent_response))
    assert recent_response['exit_code']==0 and recent_response['status']==200
    assert 'retry-after' not in recent_response['headers'], 'server Retry-After restriction'
    assert recent_response['headers'].get('x-ratelimit-remaining')!='0', 'primary rate reset restriction'
    recent=json.loads(recent_response['body']);assert not recent.get('errors')
    page=recent['data']['repository']['issues'];assert type(page['totalCount']) is int
    assert 0<=page['totalCount']-len(graph)<=100, 'recent scan cannot cover later native numbers'
    newer=[n for n in page['nodes'] if n['number']>max(by)]
    assert len(newer)==page['totalCount']-len(graph), 'recent scan does not cover later numbers'
    found_recent=matching(page['nodes'],body,node['title'])
    if found and found_recent:assert found[0]['number']==found_recent[0]['number']
    existing=(found_recent or found)
    event=dict(authorization_sha256=AUTH_SHA,alias=ALIAS,original_intent_sha256=INTENT,observed_at=now(),attempts=0)
    if existing:
        data=existing[0]
        ident=identity(dict(number=data['number'],databaseId=data.get('databaseId',data.get('id')),id=data.get('node_id',data.get('id'))))
        event['status']='ADOPTED_EXISTING_WITHOUT_CREATE'
    else:
        variables={'input':dict(repositoryId=REPO,title=node['title'],body=body,issueTypeId=TASK,projectV2Ids=[PROJECT],clientMutationId=ALIAS)}
        assert sha(json.dumps(dict(query=QUERY,variables=variables),sort_keys=True))==INTENT
        c.verify_inputs()
        event.update(status='AUTHORIZATION_CONSUMED_BEFORE_CREATE',attempts=1,request_sha256=INTENT)
        consume(consumed,event)
        entry=dict(operation='root-one-attempt create '+ALIAS,authorization_sha256=AUTH_SHA,request_sha256=INTENT,status='INTENT',started_at=now())
        c.journal['operations'].append(entry);atomic(c.journal_path,c.journal)
        try:
            response=http_graphql(QUERY,variables)
            event['response']=response;entry['sanitized_http_response']=response
            assert response['exit_code']==0 and 200<=response['status']<300
            result=json.loads(response['body']);assert not result.get('errors')
            ident=identity((result.get('data') or {}).get('createIssue',{}).get('issue'))
        except Exception as error:
            entry.update(status='RECONCILE_REQUIRED',finished_at=now())
            event.update(status='AMBIGUOUS; ALL NATIVE WRITES HELD',error=type(error).__name__+': '+str(error))
            atomic(c.journal_path,c.journal);atomic(RECOVERY/'one-attempt-result.json',event)
            atomic(DEFAULT/'native-write-halt.json',event)
            raise
    try:
        actual=c.metadata([ident['number']])[ident['number']]
        assert identity(actual)==ident and actual['body']==body and actual['title']==node['title']
        assert actual['state']=='OPEN' and actual['issueType']['name']=='Task'
    except Exception as error:
        event.update(status='READBACK UNCONFIRMED; ALL NATIVE WRITES HELD',error=type(error).__name__+': '+str(error))
        atomic(RECOVERY/'one-attempt-result.json',event);atomic(DEFAULT/'native-write-halt.json',event)
        raise
    c.registry['nodes'][ALIAS]=ident;c.save_registry('ALLOCATING')
    event.update(status='CONFIRMED_IDENTITY_AND_EXACT_BODY; original allocation may continue',identity=ident,finished_at=now(),project_size='Original allocation readback still required')
    if not existing:
        entry.update(status='CONFIRMED_TYPED_IDENTITY_AND_BODY',finished_at=now());atomic(c.journal_path,c.journal)
    atomic(RECOVERY/'one-attempt-result.json',event)
    print(json.dumps(event,indent=2))


def main():
    with (DEFAULT/'coordinator.lock').open('a') as lock:
        fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB)
        recover(Coordinator(DEFAULT))


if __name__=='__main__':main()
