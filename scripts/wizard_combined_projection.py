"""Complete proposal projection over actual foreign nodes; no empty sentinels."""
import json,hashlib
from pathlib import Path
from collections import Counter
from wizard_native_checkpoint import acyclic

def project(base,s,w,z):
    graph={str(t['number']):[str(x) for x in t['blocked_by']] for t in base['graph']}
    parents={str(t['number']):str(t['parent']) for t in base['graph'] if t['parent']}
    packs={330:s['tasks']+s['shared_split_proposal']+s['gates'],331:w['tasks']+w['gates'],332:z['tasks']+z['gates']}
    def ref(v,n):
        v=str(v)
        if v.isdigit():return v
        if v=='warlock-shared-output':return '331:shared-output'
        if v=='wizard-shared-output':return '332:shared-output'
        if ':' in v:return v
        return str(n)+':'+v
    for n,ts in packs.items():
        for t in ts:
            ident=ref(t['key'],n);assert ident not in graph,('duplicate proposal key',ident)
            graph[ident]=sorted({str(x) for x in t['prerequisites']}|{ref(x,n) for x in t['depends']})
            parents[ident]=ref(t['parent'],n)
    for n in [330,331,332]:graph['337'].append(f'{n}:class-gate')
    for n in [279,280]:graph[str(n)].append('330:shared-aggregate')
    assert '330:world-cohort' in graph['330:shared-output'], 'mandatory world cohort through shared output; old183 unchanged'
    unknown=sorted({x for ds in graph.values() for x in ds if x not in graph}|{p for p in parents.values() if p not in graph})
    assert not unknown,('unbound foreign proposal references',unknown)
    incoming={k:len(set(v)) for k,v in graph.items()};outgoing=Counter(x for ds in graph.values() for x in set(ds));children=Counter(parents.values())
    errors={label:{k:v for k,v in counts.items() if v>100} for label,counts in [('incoming',incoming),('outgoing',outgoing),('children',children)]}
    assert not any(errors.values()), ('native caps',errors)
    acyclic(graph);acyclic({k:[p] for k,p in parents.items()})
    return dict(status='PROPOSAL ONLY; not published/native final seal',nodes=len(graph),counts={str(n):len(ts) for n,ts in packs.items()},errors=errors,maxima=dict(incoming=max(incoming.values()),outgoing=max(outgoing.values()),children=max(children.values())),old285_outgoing=outgoing['285'],old67_outgoing=outgoing['67'],parent279_incoming=incoming['279'],parent280_incoming=incoming['280'],parent283_children=children['283'],old429_outgoing=outgoing['429'],graph=graph,parents=parents)
if __name__=='__main__':
 import sys
 paths=[Path(p) for p in sys.argv[1:5]];r=project(*(json.loads(p.read_text()) for p in paths));r['input_sha256']={str(p):hashlib.sha256(p.read_bytes()).hexdigest() for p in paths};Path(sys.argv[5]).write_text(json.dumps(r,indent=2)+'\n');print({k:v for k,v in r.items() if k not in ['graph','parents','input_sha256']})
