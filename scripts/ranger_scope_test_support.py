"""Synthetic graph fixture; deliberately never represents hosted publication evidence."""
from copy import deepcopy
import verify_ranger_scope as v

def synthetic_receipt(manifest):
 m=deepcopy(manifest);rr=v.rows(m)
 # Use the published identities in this isolated graph so identity guards are
 # exercised too. Synthetic graph contents never constitute a hosted readback.
 assert all(type(x['issue']) is int for x in rr.values())
 g={}
 def node(n):
  return g.setdefault(n,dict(number=n,title='synthetic anchor',body='',state='open',parent=None,blocked_by=[]))
 for n,x in m['reused'].items():g[int(n)]=deepcopy(x)
 for n,x in m['protected_owners'].items():g[int(n)]=deepcopy(x)
 for n,edges in m['historical_edges'].items():node(int(n))['blocked_by']=deepcopy(edges)
 for n,consumers in m.get('historical_outgoing',{}).items():
  for consumer in consumers:
   if int(n) not in node(consumer)['blocked_by']:node(consumer)['blocked_by'].append(int(n))
 node(328).update(parent=283,blocked_by=[285,293,315],state='open')
 for n in [285,293,315]:node(n)['state']='closed'
 gateway=__import__('json').loads((v.ROOT/'docs/investigations/issue-328-gateway-receipt.json').read_text())['issue']
 node(968).update(parent=336,blocked_by=[429],title=gateway['title'],body=gateway['body'])
 m['qualification_paths']={}
 project={'968':dict(type='Task',size='M')}
 for x in rr.values():
  parent=x.get('parent')
  if x.get('parent_key'):parent=rr[x['parent_key']]['issue']
  x['parent']=parent
  deps=sorted(set(x['prerequisites'])|{rr[d]['issue'] for d in x['depends']})
  node(x['issue']).update(title=x['title'],body=x['body'],parent=parent,blocked_by=deps)
  project[str(x['issue'])]=dict(type='Task',size='M')
  for root in x.get('roots',[]):node(root)['blocked_by'].append(x['issue'])
 for x in m['tasks']:
  parent=rr[x['parent_key']]['issue']
  if x.get('spell'):paths=[[x['issue'],parent,rr['shared-gate']['issue'],r] for r in [279,280]]+[[x['issue'],parent,rr['shared-gate']['issue'],rr['class-gate']['issue'],337]]
  else:paths=[[x['issue'],parent,rr['class-gate']['issue'],337]]
  m['qualification_paths'][str(x['issue'])]=paths
 for primitive in [61,64,65,66,67,273,353]:
  casting=rr['spellcasting'];cohort=rr[casting['parent_key']]
  m['qualification_paths'][str(primitive)]=[[primitive,rr['execution-dependencies']['issue'],casting['issue'],cohort['issue'],rr['class-gate']['issue'],337]]
 for i in range(100-sum(429 in x['blocked_by'] for x in g.values())):node(700000+i)['blocked_by']=[429]
 # Add only missing anchors; no implied implementation for them.
 for n in [n for x in list(g.values()) for n in x['blocked_by']+([x['parent']] if x['parent'] else [])]:node(n)
 return m,dict(graph=list(g.values()),project=project)
