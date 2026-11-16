#!/usr/bin/env python3
"""Collect an UNSEALED exact native checkpoint; never publishes or final-seals."""
import json
from datetime import datetime,timezone
from pathlib import Path
import subprocess
from collections import Counter

def acyclic(graph):
    visiting=set();done=set()
    def visit(n):
        assert n not in visiting,('cycle',n)
        if n in done:return
        visiting.add(n)
        for x in graph.get(n,[]):visit(x)
        visiting.remove(n);done.add(n)
    for n in graph:visit(n)

def collect():
    query='''query($cursor:String){repository(owner:"subculture-collective",name:"edda"){issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{number title body state issueType{name} parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}} projectItems(first:20){pageInfo{hasNextPage}nodes{project{id} fieldValueByName(name:"Size"){...on ProjectV2ItemFieldSingleSelectValue{name}}}}}}}}'''
    graph=[];seen=set();cursors=set();cursor=None
    for call in range(20):
        payload=dict(query=query,variables=dict(cursor=cursor))
        r=subprocess.run(['gh','api','graphql','--input','-'],input=json.dumps(payload),text=True,capture_output=True,check=True)
        data=json.loads(r.stdout);assert not data.get('errors'),data.get('errors')
        page=data['data']['repository']['issues']
        for n in page['nodes']:
            assert n['number'] not in seen,'repeated issue/page';seen.add(n['number'])
            assert not n['blockedBy']['pageInfo']['hasNextPage'],'incoming cap exceeded'
            assert not n['projectItems']['pageInfo']['hasNextPage'],'project pagination needs expansion'
            graph.append(dict(number=n['number'],title=n['title'],body=n['body'] or '',state=n['state'],type=(n['issueType'] or {}).get('name'),parent=(n['parent'] or {}).get('number'),blocked_by=sorted(x['number'] for x in n['blockedBy']['nodes']),projects=n['projectItems']['nodes']))
        if not page['pageInfo']['hasNextPage']:break
        cursor=page['pageInfo']['endCursor'];assert cursor and cursor not in cursors,'repeated/empty cursor';cursors.add(cursor)
    else:raise AssertionError('bounded native page limit')
    deps={n['number']:n['blocked_by'] for n in graph};parents={n['number']:[n['parent']] if n['parent'] else [] for n in graph}
    assert all(x in seen for ds in deps.values() for x in ds),'incomplete foreign dependency graph'
    assert all(x in seen for ps in parents.values() for x in ps),'incomplete foreign containment graph'
    outgoing=Counter(x for ds in deps.values() for x in ds);children=Counter(n['parent'] for n in graph if n['parent'])
    assert max(outgoing.values(),default=0)<=100 and max(children.values(),default=0)<=100
    acyclic(deps);acyclic(parents)
    return dict(status='UNSEALED LIVE CHECKPOINT ONLY; no approved scope publication or POST-WARLOCK prefix release',observed_at=datetime.now(timezone.utc).isoformat(),api_calls=call+1,graph=graph,caps=dict(max_incoming=max(map(len,deps.values()),default=0),max_outgoing=max(outgoing.values(),default=0),max_children=max(children.values(),default=0)),accepted_parent_prefix=None)
if __name__=='__main__':
    import argparse
    p=argparse.ArgumentParser();p.add_argument('output',type=Path);a=p.parse_args();r=collect();a.output.write_text(json.dumps(r,indent=2)+'\n');print(len(r['graph']),'issues; caps',r['caps'],'both native DAGs PASS; UNSEALED')
