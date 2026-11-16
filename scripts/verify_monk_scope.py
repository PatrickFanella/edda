#!/usr/bin/env python3
"""Strict offline receipt or read-only live audit for Monk decision326."""
import argparse
from copy import deepcopy
import hashlib
import json
from pathlib import Path
import subprocess
from verify_druid_scope import Reader, pages
from datetime import datetime, timezone
PROJECT='PVT_kwDODjtmGM4BrFJ2'
from verify_barbarian_scope import acyclic

ROOT=Path(__file__).resolve().parents[1]
MANIFEST=ROOT/'docs/investigations/issue-326-monk-scope.json'
RECEIPT=ROOT/'docs/investigations/issue-326-link-verification.json'
PIN='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'
START='<!-- edda-326:start -->'
END='<!-- edda-326:end -->'
def sha(b):return hashlib.sha256(b.encode() if isinstance(b,str) else b).hexdigest()
def next_cursor(info,seen):
 cursor=info['endCursor']
 assert cursor and cursor not in seen, 'repeated or missing cursor'
 seen.add(cursor)
 return cursor

def read_live(m):
    api = Reader()
    graph, cursor, seen = [], None, set()
    cursors=set()
    for _ in range(20):
        q = '''query($cursor:String){repository(owner:"subculture-collective",name:"edda"){
        issues(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{
        number title body state parent{number} blockedBy(first:100){pageInfo{hasNextPage}nodes{number}}}}}}'''
        b = api('graphql', {'query': q, 'variables': {'cursor': cursor}})['data']['repository']['issues']
        for node in b['nodes']:
            assert node['number'] not in seen, 'repeated issue page'
            seen.add(node['number'])
            assert not node['blockedBy']['pageInfo']['hasNextPage'], 'blocker cap exceeded'
            graph.append(dict(number=node['number'],title=node['title'],body=node['body'] or '',state=node['state'].lower(),parent=(node['parent'] or {}).get('number'),blocked_by=sorted(x['number'] for x in node['blockedBy']['nodes'])))
        if not b['pageInfo']['hasNextPage']:break
        cursor=next_cursor(b['pageInfo'],cursors)
    else:raise AssertionError('graph page cap')
    project={}
    nums=sorted({r['issue'] for r in m['issues']+m['gates']}|set(m['reused_terminals']))
    for start in range(0,len(nums),30):
        batch=nums[start:start+30]
        fields=' '.join('i%d:issue(number:%d){issueType{name} projectItems(first:100){pageInfo{hasNextPage} nodes{project{id} fieldValueByName(name:"Size"){... on ProjectV2ItemFieldSingleSelectValue{name}}}}}'%(n,n) for n in batch)
        data=api('graphql',{'query':'{repository(owner:"subculture-collective",name:"edda"){'+fields+'}}'})['data']['repository']
        for n in batch:
            node=data['i'+str(n)];assert not node['projectItems']['pageInfo']['hasNextPage']
            items=[x for x in node['projectItems']['nodes'] if x['project']['id']==PROJECT];assert len(items)==1
            project[str(n)]=dict(type=node['issueType']['name'],size=items[0]['fieldValueByName']['name'])
    indexed={x['number']:x for x in graph}
    old=json.loads(RECEIPT.read_text()) if RECEIPT.exists() else {}
    for n,seal in old.get('body_seals',{}).items():
        assert indexed[int(n)]['body'].startswith(seal['prefix']), 'sealed parent prefix rollback'
    seals={str(x['issue']):dict(prefix=indexed[x['issue']]['body'],sha256=sha(indexed[x['issue']]['body'])) for x in m['supplements']}
    return dict(observed_at=datetime.now(timezone.utc).isoformat(),graph=graph,project=project,api_calls=api.calls,body_seals=seals)


def validate(m,r,require=False):
 assert m['source']['sha256']==PIN
 assert len(m['records'])==94
 assert len(m['record_matrix'])==94
 for rec,entry in zip(m['records'],m['record_matrix']):
  assert entry['id']==rec['id'] and entry['state']=='metadata only; Monk unavailable'
  assert entry['key_owners']=={k:m['source_key_owners'][k] for k in rec['required_mechanics']+rec['option_keys']}
  assert all(entry[k] for k in ['engine','database','web','tui'])
 for path,c in m['catalogues'].items():assert json.loads((ROOT/path).read_text())==c,'canonical source/numeric/operand drift'
 records=sum([c['records'] for p,c in m['catalogues'].items() if p.endswith(('character-monk.json','character-open-hand.json'))],[])
 assert records==m['records']
 keys={k for x in records for k in x['required_mechanics']+x['option_keys']}
 assert len(keys)==208 and keys==set(m['source_key_owners'])
 rows={x['key']:x for x in m['issues']}
 assert len(rows)==66 and len(m['reused_terminals'])==32
 assert sha(json.dumps(m['source_key_owners'],sort_keys=True))==m['ownership_sha256'],'owner drift'
 assert sha(json.dumps([{k:x[k] for k in ['key','fixture','depends','prerequisites','owned_keys','source_records']} for x in m['issues']],sort_keys=True))==m['contracts_sha256'],'contract drift'
 for k,v in m['source_key_owners'].items():assert (v in rows and k in rows[v]['owned_keys']) or v in {'external:384','external:389','external:488','external:727'}
 for x in rows.values():assert x['owned_keys']==sorted(k for k,v in m['source_key_owners'].items() if v==x['key'])
 for p,h in m['evidence_hashes'].items():assert sha((ROOT/p).read_bytes())==h,'runtime evidence drift'
 palm=next(x for x in records if x['kind']=='feature' and x['id'].endswith('/quivering-palm'))
 params={k.rsplit('/',1)[-1]:v for k,v in palm['parameters'].items()}
 witness=m['deterministic_witnesses']['quivering-palm']
 assert len(witness['faces'])==params['damage-dice'] and all(1<=n<=params['damage-die-sides'] for n in witness['faces']), 'Palm source dice count/sides'
 assert sum(witness['faces'])==witness['failed_damage'] and witness['successful_damage']==witness['failed_damage']//2
 assert witness['setup_focus']==params['focus-cost'] and witness['termination_focus']==0
 assert 'ten fixed D12 faces '+str(witness['faces']).replace(', ', ',') in rows['palm-end']['fixture']
 martial=m['deterministic_witnesses']['martial-arts'];modifier=(martial['strength']-10)//2
 level=next(x for x in records if x['kind']=='level' and x['level']==martial['monk_level'])
 sides=level['parameters']['srd-5.2.1:mechanic/monk/table/martial-arts-die-sides']
 assert 1<=martial['monk_face']<=sides and martial['monk_damage']==martial['monk_face']+modifier
 assert martial['unarmed_normal']==1+modifier
 assert martial['spear_normal_sides']==8 and 1<=martial['spear_normal_face']<=8 and martial['spear_normal_damage']==martial['spear_normal_face']+modifier
 graph={x['number']:x for x in r['graph']};assert len(graph)==len(r['graph'])
 acyclic(graph);acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] else []) for n,x in graph.items()})
 counts={}
 outgoing={}
 for x in graph.values():
  for n in x['blocked_by']:outgoing[n]=outgoing.get(n,0)+1
  assert len(x['blocked_by'])<=100 and len(set(x['blocked_by']))==len(x['blocked_by']), 'incoming blocker capacity'
  counts[x['parent']]=counts.get(x['parent'],0)+1
 assert all(v<=100 for k,v in counts.items() if k is not None), 'native child capacity'
 assert all(v<=100 for v in outgoing.values()), 'blocking relation capacity'
 assert graph[326]['parent']==283 and set(graph[326]['blocked_by'])=={285,291,313}
 assert all(graph[n]['state']=='closed' for n in [285,291,313,321,322,323,324,334])
 g=m['gates'][0]; assert g['parent']==326 and g['issue'] in graph[337]['blocked_by']
 terminals={x['issue'] for x in m['issues']}|set(m['reused_terminals'])
 assert len(terminals)==98 and set(g['blocked_by_issues'])==terminals|{326}
 assert set(m['qualification_paths'])=={str(n) for n in terminals}
 for n in terminals:
  assert m['qualification_paths'][str(n)]==[n,g['issue'],337]
  assert n in graph[g['issue']]['blocked_by']
 for x in m['issues']+m['gates']:
  a=graph[x['issue']]
  assert a['title']==x['title'] and a['body']==x['body'] and a['parent']==x['parent'],'owned body/title/parent'
  assert set(a['blocked_by'])==set(x['blocked_by_issues']),'owned prerequisite'
  assert r['project'][str(x['issue'])]==dict(type='Task',size='M'),'Task/Project Size'
  if x in m['issues']:
   assert set(x['blocked_by_issues'])==set(x['prerequisites'])|{rows[k]['issue'] for k in x['depends']}
   if x['parent'] in [50,270]:assert all(x['issue'] in graph[n]['blocked_by'] for n in [279,280])
 for n,edges in m['historical_edges'].items():assert set(edges)<=set(graph[int(n)]['blocked_by']), 'historical class gate edge'
 for n,old in m['preserved'].items():
  a=graph[int(n)]
  assert a['title']==old['title'] and a['parent']==old['parent'],'historical identity'
  assert set(old['blocked_by'])<=set(a['blocked_by']),'historical edge'
  s=next((s for s in m['supplements'] if str(s['issue'])==n),None)
  if not s:assert a['body']==old['body'],'consumed owner body'
  else:
   assert sha(old['body'])==s['previous_sha256']
   assert (ROOT/s['path']).read_text()==s['append']
   b=a['body'];seal=r['body_seals'][n]['prefix']
   assert b.startswith(seal) and sha(seal)==r['body_seals'][n]['sha256'],'published prefix rollback'
   assert seal.startswith(old['body']) and b.startswith(old['body']),'original prefix'
   present=s['append'] in b
   assert b.count(START)==b.count(END)==b.count(s['append'])<=1,'duplicate/changed/partial section'
   if require:assert present,'pending parent supplement'
  # Source and consumed terminal prerequisite scopes are exact. Shared qualification
  # roots admit later additive class edges; these do not qualify another class.
  if int(n) not in [47,54,283,337,279,280]:assert a['blocked_by']==old['blocked_by'],'source/consumed prerequisite drift'
 for n in m['reused_terminals']:assert r['project'][str(n)]==m['consumed_project'][str(n)]
 assert 334 in graph[337]['blocked_by'] and {298,299,300,301}==set(graph[334]['blocked_by'])
 assert len(next(s['append'] for s in m['supplements'] if s['issue']==47))<=100
 for s in m['supplements']:assert len(m['preserved'][str(s['issue'])]['body']+s['append'])<=65536
 for path in m['auxiliary_paths'].values():
  for a,b in zip(path,path[1:]):assert a in graph[b]['blocked_by']
 assert rows['d8']['issue'] in graph[rows['later']['issue']]['blocked_by'] and 429 not in graph[rows['later']['issue']]['blocked_by']
 for level in range(1,21):
  x=rows['level-'+str(level)]
  expected=[y for y in records if y['kind']=='level' and y['level']==level]+[y for y in m['catalogues']['internal/rules/srd521/testdata/character-advancement.json']['records'] if y['level']==level]
  assert x['source_records']==expected and 'd8' in x['depends']
  assert not {285,336,429}.intersection(x['prerequisites']), 'redundant cross-class edge'
  assert rows['d8']['issue'] in graph[x['issue']]['blocked_by'] and 429 in graph[rows['d8']['issue']]['blocked_by']
  assert 336 in graph[429]['blocked_by'] and 285 in graph[429]['blocked_by'] and 285 in graph[336]['blocked_by']
  if level>1:assert 'level-'+str(level-1) in x['depends']
 factor=json.loads((MANIFEST.parent/'issue-326-advancement-factorization.json').read_text())
 assert factor['unchanged_flattened_gate_blockers']==g['blocked_by_issues']
 assert factor['older_owner_changes']==[] and factor['native_removed_owned_edges']=={'822':[429],'823':[285,336,429],'824':[285,336]}
 assert set(factor['native_paths'])=={str(x['issue']) for x in m['issues'] if x['key'].startswith('level-') or x['key']=='later'}
 for paths in factor['native_paths'].values():
  for path in paths:
   for a,b in zip(path,path[1:]):assert b in graph[a]['blocked_by'], 'factored source transaction path'
 return len(terminals),len(graph)
def probes(m,r):
 def reject(label,fn):
  a,b=deepcopy(m),deepcopy(r);fn(a,b)
  try:validate(a,b)
  except (AssertionError,KeyError):return
  raise AssertionError('accepted '+label)
 def node(b,n):return next(x for x in b['graph'] if x['number']==n)
 tests=[
 ('source',lambda a,b:a['catalogues'][next(iter(a['catalogues']))]['records'][0]['parameters'].update(bad=9)),
 ('Palm extra die',lambda a,b:a['deterministic_witnesses']['quivering-palm']['faces'].append(12)),
 ('Palm half damage',lambda a,b:a['deterministic_witnesses']['quivering-palm'].update(successful_damage=33)),
 ('Martial ability modifier',lambda a,b:a['deterministic_witnesses']['martial-arts'].update(monk_damage=4)),
 ('numeric',lambda a,b:a['catalogues'][next(iter(a['catalogues']))]['records'][0]['parameters'].update({'srd-5.2.1:mechanic/monk/hit-die-sides':9})),
 ('pin',lambda a,b:a['source'].update(sha256='bad')),
 ('operand',lambda a,b:a['issues'][0].update(fixture='wrong clause')),
 ('owner',lambda a,b:a['source_key_owners'].pop(next(iter(a['source_key_owners']))))]
 owned={x['issue'] for x in m['issues']+m['gates']}
 for n in sorted(owned|{int(n) for n in m['preserved']}|{int(n) for n in m['historical_edges']}):
  required=set(node(r,n)['blocked_by']) if n in owned else set(m['preserved'][str(n)]['blocked_by'] if str(n) in m['preserved'] else m['historical_edges'][str(n)])
  if n in [279,280]:required|={x['issue'] for x in m['issues'] if x['parent'] in [50,270]}
  if n==337:required.add(m['gates'][0]['issue'])
  for e in required:tests.append((f'edge{n}/{e}',lambda a,b,n=n,e=e:node(b,n)['blocked_by'].remove(e)))
 for n in owned:
  for field,val in [('title','bad'),('body','bad'),('parent',283)]:
   tests.append((f'{field}{n}',lambda a,b,n=n,f=field,v=val:node(b,n).update({f:v})))
  tests.append((f'metadata{n}',lambda a,b,n=n:b['project'][str(n)].update(type='Feature')))
 for old_n in m['preserved']:
  if int(old_n) in [47,54,283,337]:continue
  for field,val in [('title','bad'),('body','bad'),('parent',None)]:
   if field=='parent' and node(r,int(old_n))['parent'] is None:val=283
   tests.append((f'preserved {field}{old_n}',lambda a,b,n=int(old_n),f=field,v=val:node(b,n).update({f:v})))
 for old_n in m['reused_terminals']:
  for field,val in [('type','Feature'),('size','XL')]:
   tests.append((f'consumed metadata{old_n}/{field}',lambda a,b,n=old_n,f=field,v=val:b['project'][str(n)].update({f:v})))
 for source_n in [285,291,313,321,322,323,324,334]:
  tests.append((f'closed source{source_n}',lambda a,b,n=source_n:node(b,n).update(state='open')))
 g=m['gates'][0]['issue'];n=m['issues'][0]['issue']
 tests += [('dependency cycle',lambda a,b:node(b,n)['blocked_by'].append(g)),('parent cycle',lambda a,b:node(b,326).update(parent=g))]
 for s in m['supplements']:
  n=s['issue']
  for suffix in [START,END,s['append']+s['append'],s['append'].replace('#326','#325')]:
   tests.append((f'section{n}',lambda a,b,n=n,t=suffix:node(b,n).update(body=node(b,n)['body']+t)))
 for label,fn in tests:reject(label,fn)
 # Closed planning decision cannot erase any terminal requirement.
 b=deepcopy(r);node(b,326)['state']='closed';validate(m,b)
 for s in m['supplements']:
  n=s['issue'];b=deepcopy(r);node(b,n)['body']+='\nLater unrelated append.\n';validate(m,b)
  b=deepcopy(r);prefix=m['preserved'][str(n)]['body']+s['append'];b['body_seals'][str(n)]=dict(prefix=prefix,sha256=sha(prefix));node(b,n)['body']=m['preserved'][str(n)]['body']
  try:validate(m,b)
  except AssertionError:pass
  else:raise AssertionError('sealed rollback accepted')
 # Additive unowned class gate edge and body append must remain admissible.
 b=deepcopy(r);b['graph'].append(dict(number=99999,title='Unowned class gate',body='not qualification evidence',state='open',parent=None,blocked_by=[]));node(b,337)['blocked_by'].append(99999);validate(m,b)
 calls=[]
 from urllib.parse import parse_qs,urlsplit
 def page(url):
  q=parse_qs(urlsplit(url).query);assert q==dict(state=['all'],per_page=['100'],page=q['page'])
  n=int(q['page'][0]);calls.append(n)
  return [dict(id=n*100+i) for i in range(100 if n==1 else 1)]
 assert len(list(pages(page,'repos/o/r/issues?state=all')))==101 and calls==[1,2]
 for callback in [lambda u:[dict(id=i) for i in range(100)],lambda u:[dict(id=int(parse_qs(urlsplit(u).query)['page'][0])*100+i) for i in range(100)]]:
  try:list(pages(callback,'repos/o/r/issues'))
  except AssertionError:pass
  else:raise AssertionError('pagination repetition/cap')
 print('PASS',len(tests),'bounded adversarial guards; pending/sealed/closed/additive controls')
def source_check(pdf):
 import re
 weapons=subprocess.run(['pdftotext','-f','91','-l','91','-layout',str(pdf),'-'],capture_output=True,text=True,check=True).stdout
 assert re.search(r'Spear\s+1d6 Piercing[^\n]+Versatile \(1d8\)',weapons), 'PDF91 Spear normal/Versatile source'
 assert sha(pdf.read_bytes())==PIN
 for s in ['monk','open_hand','advancement','multiclass']:
  subprocess.run(['python3',str(ROOT/'scripts'/('verify_srd_'+s+'.py')),str(pdf)],check=True)
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--live',action='store_true');p.add_argument('--pdf',type=Path);p.add_argument('--probes',action='store_true');p.add_argument('--require-supplements',action='store_true');a=p.parse_args()
 m=json.loads(MANIFEST.read_text())
 if a.pdf:source_check(a.pdf)
 r=read_live(m) if a.live else json.loads(RECEIPT.read_text())
 count=validate(m,r,a.require_supplements)
 if a.probes:probes(m,r)
 if a.live:RECEIPT.write_text(json.dumps(r,indent=2)+'\n')
 graph={x['number']:x for x in r['graph']}
 statuses={s['issue']:('published' if s['append'] in graph[s['issue']]['body'] else 'pending parent append') for s in m['supplements']}
 print('PASS mandatory leaves/nodes',count,'planning consistency only; Monk unavailable; supplements',statuses)
if __name__=='__main__':main()
