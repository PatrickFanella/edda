#!/usr/bin/env python3
"""Combine actual three-class proposal structures; unresolved sources never qualify."""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
from verify_sorcerer_scope import BASELINE, MANIFEST, native_caps, validate_plan, acyclic


def combined(own, warlock, wizard, strict=True):
    validate_plan(own)
    baseline=json.loads(BASELINE.read_text())
    graph={str(x['number']):[str(n) for n in x['blocked_by']] for x in baseline['graph']}
    parents={str(x['number']):str(x['parent']) for x in baseline['graph'] if x['parent']}
    proposals=[(330,own['tasks']+own['shared_split_proposal']+own['gates']),
               (331,warlock['tasks']+warlock['gates']),(332,wizard['tasks']+wizard['gates'])]

    def ref(value, owner):
        if type(value) is int:
            assert value>0
            return str(value)
        assert isinstance(value,str)
        if value.isdigit():
            return value
        if value=='warlock-shared-output':
            return '331:shared-output'
        if value=='wizard-shared-output':
            return '332:shared-output'
        if value.startswith('330:'):
            return value[4:]
        if value.startswith(('331:','332:')):
            return value
        return value if owner==330 else str(owner)+':'+value

    metadata={}
    for owner,rows in proposals:
        for t in rows:
            key=ref(t['key'],owner)
            assert key not in graph, 'duplicate owner/key'
            assert t['type']=='Task' and t['size']=='M' and t['body'], 'concrete typed body missing'
            assert all(type(n) is int for n in t['prerequisites']), 'numeric old prerequisites'
            graph[key]=[ref(d,owner) for d in t['depends']]+[ref(n,owner) for n in t['prerequisites']]
            parents[key]=ref(t['parent'],owner)
            metadata[key]=dict(title=t['title'],body=t['body'],parent=parents[key],type='Task',size='M')
    for root,dep in [('279','shared-aggregate'),('280','shared-aggregate')]:
        graph[root].append(dep)
    for owner in [330,331,332]:
        graph['337'].append(ref('class-gate',owner))
    original183=next(x['blocked_by'] for x in baseline['graph'] if x['number']==183)
    assert graph['183']==[str(n) for n in original183], 'no183prerequisite mutation'
    assert 'world-cohort' in graph['shared-output'], 'mandatory world cohort via shared output'
    missing=sorted({d for ds in graph.values() for d in ds if d not in graph})
    missing_parents=sorted({p for p in parents.values() if p not in graph})
    acyclic(graph)
    acyclic({k:[v] for k,v in parents.items()})
    outgoing=Counter(d for ds in graph.values() for d in ds)
    childcounts=Counter(parents.values())
    violations=([dict(kind='incoming',node=k,count=len(ds)) for k,ds in graph.items() if len(ds)>100]
                +[dict(kind='outgoing',node=k,count=n) for k,n in outgoing.items() if n>100]
                +[dict(kind='children',node=k,count=n) for k,n in childcounts.items() if n>100])
    if strict:
        native_caps(graph,parents)
    assert graph['shared-aggregate']==['shared-output','331:shared-output','332:shared-output'], 'exact three shared heads'
    # No shared output may reach its consuming class qualification through any
    # indirect path. This check is independent of candidate path declarations.
    def reaches(start,target,seen=None):
        seen=set() if seen is None else seen
        if start==target:
            return True
        if start in seen:
            return False
        seen.add(start)
        return any(reaches(d,target,seen) for d in graph.get(start,[]))
    for owner in [330,331,332]:
        assert not reaches(ref('shared-output',owner),ref('class-gate',owner)), 'consumer gate backlink'
    out=Counter(d for ds in graph.values() for d in ds)
    children=Counter(parents.values())
    unresolved=list(wizard.get('unresolved_execution_seams',[]))
    unresolved+=warlock.get('unresolved_execution_seams',[])
    return dict(status='PENDING root acceptance and unresolved execution seams',native_publication_approved=False,
                class_node_counts={str(n):len(rows) for n,rows in proposals},new_nodes=len(metadata),
                incoming279=len(graph['279']),incoming280=len(graph['280']),incoming337=len(graph['337']),
                children283=children['283'],children330=children['330'],children331=children['331'],children332=children['332'],
                outgoing429=out['429'],outgoing487=out['487'],outgoing67=out['67'],
                max_incoming=max(map(len,graph.values())),max_outgoing=max(out.values()),max_children=max(children.values()),
                cap_violations=violations,unresolved_references=missing,unresolved_parents=missing_parents,unresolved_execution_seams=unresolved,
                dependency_dag='acyclic for actual proposal edges; unresolved references explicitly listed',containment_dag='acyclic for actual proposal edges',
                graph=graph,parents=parents,concrete_contracts=metadata)


def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--warlock',required=True,type=Path)
    p.add_argument('--wizard',required=True,type=Path)
    p.add_argument('--output',required=True,type=Path)
    args=p.parse_args()
    payloads={330:MANIFEST.read_bytes(),331:args.warlock.read_bytes(),332:args.wizard.read_bytes()}
    result=combined(*(json.loads(payloads[n]) for n in [330,331,332]),strict=False)
    result['provenance']={str(n):dict(path=str(path.resolve()),sha256=hashlib.sha256(payloads[n]).hexdigest()) for n,path in [(330,MANIFEST),(331,args.warlock),(332,args.wizard)]}
    args.output.write_text(json.dumps(result,indent=2,ensure_ascii=False)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k not in ['graph','parents','concrete_contracts','provenance']},indent=2))
    if result['cap_violations'] or result['unresolved_references'] or result['unresolved_parents']:
        raise SystemExit('Combined proof FAILED; exact diagnostic written, no publication permitted')


if __name__=='__main__':
    main()
