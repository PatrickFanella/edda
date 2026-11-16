"""Project proposed graph without mutating native state."""
from collections import Counter
from copy import deepcopy
from verify_fighter_scope import acyclic

PRIMITIVES=[61,64,65,66,67,273,353]
def project(m,graph,return_graph=False):
    g={x['number']:deepcopy(x) for x in graph}
    tasks=m['issues']; ids={t['key']:-i-10 for i,t in enumerate(tasks)}
    ids.update(qualification=-1,features=-2,levels=-3,shared=-4,execution=-5)
    consumers=[]
    for t in tasks:
        level=t['key'].startswith('level-') or t['key']=='later-1'
        shared=t['category']!='class'
        deps=set(t['prerequisites'])|{ids[k] for k in t['depends']}
        if not level and t['key'] not in ['hp','traits','skills','equipment','multiclass','languages','mastery','improvement','boon']:
            deps.add(ids['execution']);consumers.append(t['key'])
        parent=ids['shared'] if shared else ids['levels'] if level else ids['features']
        if t['key']=='hide':parent=50
        g[ids[t['key']]]=dict(parent=parent,blocked_by=sorted(deps))
    g[-1]=dict(parent=329,blocked_by=[329,279,280,-2,-3,-4])
    g[-2]=dict(parent=-1,blocked_by=[ids[t['key']] for t in tasks if t['category']=='class' and not(t['key'].startswith('level-') or t['key']=='later-1')])
    g[-3]=dict(parent=-1,blocked_by=[ids[t['key']] for t in tasks if t['key'].startswith('level-') or t['key']=='later-1'])
    g[-4]=dict(parent=192,blocked_by=[ids[t['key']] for t in tasks if t['category']!='class'])
    g[-5]=dict(parent=329,blocked_by=PRIMITIVES)
    for n in [279,280]:g[n]['blocked_by'].append(-4)
    g[337]['blocked_by'].append(-1)
    acyclic(g);acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] is not None else []) for n,x in g.items()})
    incoming={n:len(x['blocked_by']) for n,x in g.items()};outgoing=Counter(k for x in g.values() for k in x['blocked_by']);children=Counter(x['parent'] for x in g.values() if x['parent'] is not None)
    for label,counts in [('incoming',incoming),('outgoing',outgoing),('children',children)]:
        assert max(counts.values())<=100,(label,max(counts.values()))
    if return_graph:return g,ids,consumers
    return dict(new_tasks=len(tasks)+5,terminals=len(tasks),gates=5,execution_consumers=consumers,maxima=dict(incoming=max(incoming.values()),outgoing=max(outgoing.values()),children=max(children.values())),projected={str(n):dict(incoming=incoming.get(n,0),outgoing=outgoing[n],children=children[n]) for n in [279,280,283,329,337,429,61,64,65,66,67,192,50,968,-1,-2,-3,-4,-5]})
