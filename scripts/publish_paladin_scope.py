#!/usr/bin/env python3
"""Publish the authorized finite #327 graph. Default is a deterministic local preflight."""
import argparse
from collections import Counter
import importlib.util
import json
from pathlib import Path
import subprocess
from verify_fighter_scope import graph_read, acyclic

ROOT=Path(__file__).resolve().parents[1]
PATH=ROOT/'docs/investigations/issue-327-paladin-scope.json'
JOURNAL=Path('/tmp/edda-paladin-publication-journal.json')


def rows(m):
    extra=[
        ('blind','Admit shared Blinded with source-bound sight and attack consequences',268,[61,65,67,268],[],
         'PDF177. Existing authoritative Blinded instance prevents sight and auto-fails sight-required checks; own attacks Disadvantage,incoming Advantage. No invented duration or source. Positive:sight check auto-fails. Negative:hearing-only check,source-free condition flag,extra stacked Advantage. Preserve independent source instances and separately qualified senses; no new Blindsight grant.'),
        ('deaf','Admit shared Deafened with source-bound hearing consequences',268,[61,67,268],[],
         'PDF181. Existing authoritative Deafened instance prevents hearing and auto-fails hearing-required ability checks ONLY. Positive:hearing-required Perception auto-fails. Negative:all Perception checks fail,sight-only test,source-free flag,hearing magically restored by ending unrelated source. No invented duration or saving-throw penalty.'),
        ('invisible','Admit shared Invisible with observer-specific sight and attack consequences',268,[61,65,66,67,268],[],
         'PDF184. Invisible when Initiative rolled grants Advantage. Seen-target effects cannot target unless creator can see recipient;carried/worn equipment concealed. Incoming attacks Disadvantage,own attacks Advantage UNLESS observer can see Invisible creature. Positive:observer-specific seeing negates attack benefit. Negative:automatic Hide,undetectable position,all observers treated identically,no sight requirement,private location leaked. Suppression preserves underlying source and resumes only while its duration remains.'),
        ('blind-end','End exact trusted Blinded instances without clearing other sources',193,[64,269],['blind'],
         'PDF177/179;RestoringTouch55/LesserRestoration144 trusted source. Consume admitted Blinded store and canonical trusted receipt transaction;remove only source-authorized selected instances,restore derived sight only when no active contribution remains. No action/resource owned here;caller pays. Positive:trusted paid cure ends targeted instance. Negative:client cure flag,foreign instance,unrelated contribution deleted,double event,rollback leaves partial state. Do not expand504.'),
        ('deaf-end','End exact trusted Deafened instances without clearing other sources',193,[64,269],['deaf'],
         'PDF181/179;RestoringTouch55/LesserRestoration144 trusted source. End only authorized targeted Deafened instances;derive hearing from remaining sources. Caller owns target,action and payment,not generic cure here. Positive:paid trusted cure ends hearing penalty. Negative:forged request,foreign instance,other condition wiped,independent source removed,no-cost duplicate. Do not expand504.'),
        ('dispel-check','Resolve Dispel Magic per-spell checks and upcast threshold',193,[203,355,359,61],[],
         'PDF124. Action,VS,120ft,instant. Existing355 owns creature spells<=3. For EACH ongoing spell>=4 roll casting-ability check DC10+spelllevel;success ends only its qualifying effect. Slot>=spelllevel automatically ends,no roll. Positive:slot3 versuslevel4 DC14,slot4 automatically endslevel4;two spells distinct checks. Negative:one shared check,DCspelllevel,level3check,hidden level disclosed,nonspell curse dispelled. Stable per-spell order and owned dice survive replay;no class eligibility borrowed from Cleric fixture.'),
        ('dispel-object','Resolve Dispel Magic object and magical-effect target removal',193,[203,355,359,67],['dispel-check'],
         'PDF124. One creature/object/magical effect in120ft,VS Action,one selected source-qualified slot. Extend355 only for object/effect identity and exact recipient removal extent;consume per-spell checks. Positive:end one target spell link without ending unrelated caster recipients. Negative:whole-campaign wipe,nonspell magic-item destruction,instantaneous consequence undone,unknown effect ownership,121ft. Target facts and authorized scope must be durable authoritative records;no inferred area ownership.')]
    own=[dict(x,category='class') for x in m['issues']]
    shared=[dict(x,category='spell') for x in m['shared_issues']]
    shared += [dict(key=k,title=t,parent=p,prerequisites=req,depends=dep,fixture=f,category='other',project_size='M',issue=None) for k,t,p,req,dep,f in extra]
    for x in own+shared:
        x['depends']=list(x['depends']);x['prerequisites']=list(x['prerequisites'])
    by={x['key']:x for x in own+shared}
    by['shining-smite']['depends'].append('invisible')
    by['restoring']['depends'] += ['blind-end','deaf-end']
    by['oath-spells']['depends'] += ['zone-truth','beacon','dispel-check','dispel-object','guardian-damage','commune','flame-strike','protection-evil']
    by['smite']['depends'].append('divine-smite')
    by['steed']['depends'] += ['steed-summon','steed-stats','steed-control','steed-bonus']
    by['blessed-warrior']['depends'] += ['guidance','mending','resistance','sacred-flame','spare-dying','thaumaturgy']
    by['blessed-warrior']['prerequisites'].append(264)
    by['blessed-warrior']['depends'].append('light-extension')
    by['preparation']['depends'] += [x['key'] for x in shared if x['category']=='spell']+['dispel-check','dispel-object']
    # Complete list execution includes existing exact effects, not just the newly split spells.
    by['preparation']['prerequisites'] += [263,264,351,352,615,616,618,619,620,621,623,624,625,626,627]
    # No feature execution depends on an application/admission leaf. Applications consume them.
    ids={x['key']:i for i,x in enumerate(own+shared)}
    assert len(ids)==len(own+shared)==115
    acyclic({ids[x['key']]:dict(blocked_by=[ids[d] for d in x['depends']]) for x in own+shared})
    assert [x['key'] for x in own+shared if 429 in x['prerequisites']]==['hp']
    return own,shared


def call(method,path,data=None):
    cmd=['rtk','proxy','gh','api','-X',method,path]
    if data is not None:cmd+=['--input','-']
    p=subprocess.run(cmd,input=None if data is None else json.dumps(data),capture_output=True,text=True,check=True)
    return json.loads(p.stdout) if p.stdout.strip() else None


def body(m,x):
    return (f"<!-- edda-327:{x['key']} -->\nDecision #327; roadmap #47. One implementation PR; Task Size {x['project_size']}.\n\n"
            f"Source: pinned SRD5.2.1 PDF, exact pages in scope; SHA256 {m['source_sha256']}. CC-BY-4.0 attribution and adaptation notice: docs/srd-5.2.1-conformance.md. Fixture boundaries are Edda additions.\n\n"
            f"Exact scope, positive and independently violated negative fixtures:\n{x['fixture']}\n\n{m['common_acceptance']}\n\n"
            'Run GOMAXPROCS=2 go test -p 2 -short -race ./... and GOMAXPROCS=2 go vet -p 2 ./...; additionally isolated PostgreSQL integration tests with injectable dice,atomic rollback,durable replay/reload and concurrent confirmation. Short tests do not run integration-tag tests. No production access,inference,deployment or coverage promotion.\n\n'
            'Ownership: feature execution is separate from creation,later-class eligibility and every adjacent level application. Generic planning parents do not implement effects. Source catalogues do not unlock actions. Consume qualified named terminals; never copy spell/item/condition effects or another class eligibility. Unknown authoritative world/possession/plane/item facts stay unavailable. For automatic benefits,use triggering confirmed event; no invented optional activation.\n\n'
            f"Native parent #{x['parent']}. Exact native prerequisites: "+', '.join('#'+str(n) for n in x['blocked_by_issues'])+'.\n\n'
            'Mandatory native qualification: this leaf -> its declared cohort/class/shared gate -> Paladin qualification -> #337. Shared spell cohorts also block #279 and #280. The five condition and two Dispel leaves directly block #279/#280. Closing327 never removes these edges.\n\n'
            'Exact source keys:\n'+'\n'.join('- '+k for k in x.get('source_keys',[]))+
            '\n\nComplete numeric/source context (opaque references do not duplicate execution):\n```json\n'+json.dumps(x.get('numeric_context',[]),indent=2)+'\n```\n')


def preflight(m,graph,own,shared):
    """Use disjoint temporary IDs to prove the full proposed graph before any write."""
    gatekeys=['qualification','class-gate','shared-gate','cohort-1','cohort-2','cohort-3']
    ids={k:-(i+1) for i,k in enumerate(gatekeys+[x['key'] for x in own+shared])}
    planned={n:dict(x,blocked_by=list(x['blocked_by'])) for n,x in graph.items()}
    reused=sorted({n for x in own+shared for n in x['prerequisites']} - {429,285,336})
    for x in own+shared:
        parent=ids['class-gate'] if x['category']=='class' else ids[f"cohort-{x['cohort']}"] if x['category']=='spell' else x['parent']
        deps=sorted(set(x['prerequisites'])|{ids[k] for k in x['depends']})
        trial=dict(x,parent=parent,blocked_by_issues=deps)
        assert len(body(m,trial))<=65536,'body size preflight'
        planned[ids[x['key']]]=dict(parent=parent,blocked_by=deps)
    other={ids[x['key']] for x in shared if x['category']=='other'}
    for key in gatekeys:
        if key=='qualification':parent=327;deps=[327,ids['class-gate'],ids['shared-gate']]
        elif key=='class-gate':parent=ids['qualification'];deps=[ids[x['key']] for x in own]
        elif key=='shared-gate':parent=ids['qualification'];deps=sorted(set(reused)|other|{ids[f'cohort-{c}'] for c in [1,2,3]})
        else:
            c=int(key[-1]);parent=277;deps=[ids[x['key']] for x in shared if x.get('cohort')==c]
        planned[ids[key]]=dict(parent=parent,blocked_by=deps)
    for target in [279,280]:planned[target]['blocked_by']+=sorted(other|{ids[f'cohort-{c}'] for c in [1,2,3]})
    planned[337]['blocked_by'].append(ids['qualification'])
    acyclic(planned);acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] else []) for n,x in planned.items()})
    assert max(Counter(x['parent'] for x in planned.values() if x['parent'] is not None).values())<=100
    out=Counter(n for x in planned.values() for n in x['blocked_by'])
    assert max(out.values())<=100,'outgoing capacity preflight'
    assert all(len(x['blocked_by'])==len(set(x['blocked_by']))<=100 for x in planned.values())
    assert 99<=out[429]<=100 and len(planned[279]['blocked_by'])==97 and len(planned[280]['blocked_by'])==92
    print(f'Full prewrite DAG/body/cap proof passed:429out{out[429]};279in97;280in92',flush=True)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--apply',action='store_true');a=p.parse_args()
    m=json.loads(PATH.read_text());own,shared=rows(m)
    print('Preflight:54class+54spell+7condition/Dispel leaves;6finite gates;121new Tasks;one429consumer')
    if not a.apply:return
    graph={x['number']:x for x in graph_read()}
    if JOURNAL.exists():
        previous=json.loads(JOURNAL.read_text())
        ours={x['issue'] for x in previous['issues'].values()}
        # Resume against the live external graph after subtracting only our known IDs.
        # No remote edge is deleted: this is a local preflight projection.
        graph={n:dict(x,blocked_by=[k for k in x['blocked_by'] if k not in ours]) for n,x in graph.items() if n not in ours}
    children=Counter(x['parent'] for x in graph.values());outgoing=Counter(n for x in graph.values() for n in x['blocked_by'])
    assert len(graph[279]['blocked_by'])<=87 and len(graph[280]['blocked_by'])<=82,'shared incoming capacity changed'
    assert outgoing[429]<=99 and children[191]==84 and children[277]==0,'reserved capacity changed'
    assert all(x['state']=='closed' for n,x in graph.items() if n in [285,292,314])
    preflight(m,graph,own,shared)
    if JOURNAL.exists():j=json.loads(JOURNAL.read_text())
    else:j=dict(issues={},before_graph=list(graph.values()));JOURNAL.write_text(json.dumps(j,indent=2)+'\n')
    def save():JOURNAL.write_text(json.dumps(j,indent=2)+'\n')
    def create(key,title):
        if key in j['issues']:return j['issues'][key]
        d=call('POST','repos/subculture-collective/edda/issues',dict(title=title,body=f'<!-- edda-327:{key} -->\nPublication in progress; no executable qualification. Full source-bound contract follows before readback.',type='Task'))
        j['issues'][key]=dict(issue=d['number'],id=d['id'],node_id=d['node_id']);save();print(key,d['number'],flush=True);return j['issues'][key]
    gatekeys=['qualification','class-gate','shared-gate','cohort-1','cohort-2','cohort-3']
    titles=['Qualify exact Paladin and Devotion evidence across all twenty levels','Qualify fifty-four Paladin feature and level applications','Qualify finite Paladin consumed effects and shared contracts']+[f'Qualify finite Paladin spell cohort {c}' for c in [1,2,3]]
    gates=[dict(key=k,title=t,issue=create(k,t)['issue'],project_size='M') for k,t in zip(gatekeys,titles)]
    for x in own+shared:x['issue']=create(x['key'],x['title'])['issue']
    by={x['key']:x for x in own+shared+gates}
    for x in own+shared:
        if x['category']=='class':x['parent']=by['class-gate']['issue']
        elif x['category']=='spell':x['parent']=by[f"cohort-{x['cohort']}"]['issue']
        x['blocked_by_issues']=sorted(set(x['prerequisites'])|{by[k]['issue'] for k in x['depends']})
        if x['category']=='spell':
            inv=json.loads((ROOT/'docs/investigations/issue-80-spell-item-inventory.json').read_text())
            record=next(r for r in inv['records'] if r['family']=='spells' and r['name']==x['spell'])
            x['source_keys']=[record['id'],f"srd-5.2.1:mechanic/paladin-consumed/{x['key']}"]
            x['numeric_context']=[dict(source_page=record['source_page'],scope=x['fixture'])]
        elif x['category']=='other':
            condition={'blind':'blinded','deaf':'deafened','invisible':'invisible','blind-end':'blinded','deaf-end':'deafened'}.get(x['key'])
            x['source_keys']=([f'srd-5.2.1:mechanic/conditions/{condition}'] if condition else ['srd-5.2.1:content/spells/dispel-magic'])+[f"srd-5.2.1:mechanic/paladin-consumed/{x['key']}"]
        x['body']=body(m,x)
    reused=sorted({n for x in own+shared for n in x['prerequisites']} - {429,285,336})
    by['qualification'].update(parent=327,blocked_by_issues=[327,by['class-gate']['issue'],by['shared-gate']['issue']])
    by['class-gate'].update(parent=by['qualification']['issue'],blocked_by_issues=sorted(x['issue'] for x in own))
    by['shared-gate'].update(parent=by['qualification']['issue'],blocked_by_issues=sorted(set(reused)|{by[f'cohort-{c}']['issue'] for c in [1,2,3]}|{x['issue'] for x in shared if x['category']=='other'}))
    for c in [1,2,3]:by[f'cohort-{c}'].update(parent=277,blocked_by_issues=sorted(x['issue'] for x in shared if x.get('cohort')==c))
    for g in gates:
        g['body']=f"<!-- edda-327:{g['key']} -->\nDecision327;Task Size M;roadmap47.\n\n"+g['title']+'. Planning closure never qualifies implementation. All finite mandatory blockers below must have source-linked execution,authoritative web AND trusted TUI confirmation,isolated PostgreSQL replay/reload/concurrency/atomic rollback and independent acceptance. Imported records and generic parent closure do not qualify effects. No production admission,grants,coverage promotion or deployment here.\n\nExact mandatory blockers: '+', '.join('#'+str(n) for n in g['blocked_by_issues'])+'.\n\nNative parent #'+str(g['parent'])+'. Required path: this gate -> Paladin qualification #'+str(by['qualification']['issue'])+' -> #337. Cohorts additionally directly block279/280. Frozen source/body/Task/Size/edges remain mandatory after327closure. No catch-all or future scope.\n'
    allrows=own+shared+gates
    assert all(len(x['body'])<=65536 and len(x['blocked_by_issues'])<=100 for x in allrows)
    planned={n:dict(x,blocked_by=list(x['blocked_by'])) for n,x in graph.items()}
    for x in allrows:planned[x['issue']]=dict(number=x['issue'],parent=x['parent'],blocked_by=x['blocked_by_issues'])
    planned[337]['blocked_by'].append(by['qualification']['issue'])
    extraedges=[by[f'cohort-{c}']['issue'] for c in [1,2,3]]+[x['issue'] for x in shared if x['category']=='other']
    for n in [279,280]:planned[n]['blocked_by']+=extraedges
    acyclic(planned);acyclic({n:dict(blocked_by=[x['parent']] if x['parent'] else []) for n,x in planned.items()})
    assert max(Counter(x['parent'] for x in planned.values() if x['parent'] is not None).values())<=100
    assert max(Counter(n for x in planned.values() for n in x['blocked_by']).values())<=100
    assert all(len(x['blocked_by'])<=100 for x in planned.values())
    m['published_issues']=allrows;m['flattened_mandatory_owners']=sorted({x['issue'] for x in own+shared}|set(reused)|{429,285,336})
    m['publication']='NATIVE PUBLICATION IN PROGRESS; parent supplements PENDING';PATH.write_text(json.dumps(m,indent=2)+'\n')
    # All operations resume by exact issue IDs; never replace an old native parent.
    for x in allrows:
        call('PATCH',f"repos/subculture-collective/edda/issues/{x['issue']}",dict(title=x['title'],body=x['body'],type='Task'))
        if f"parent:{x['issue']}" not in j:
            call('POST',f"repos/subculture-collective/edda/issues/{x['parent']}/sub_issues",dict(sub_issue_id=j['issues'][x['key']]['id']))
            j[f"parent:{x['issue']}"]=True;save()
    relations=[(x['issue'],n) for x in allrows for n in x['blocked_by_issues']]+[(337,by['qualification']['issue'])]+[(n,k) for n in [279,280] for k in extraedges]
    idmap={v['issue']:v['id'] for v in j['issues'].values()}
    for n,dep in relations:
        k=f'edge:{n}:{dep}'
        if k in j:continue
        if dep not in idmap:idmap[dep]=call('GET',f'repos/subculture-collective/edda/issues/{dep}')['id']
        call('POST',f'repos/subculture-collective/edda/issues/{n}/dependencies/blocked_by',dict(issue_id=idmap[dep]));j[k]=True;save()
    spec=importlib.util.spec_from_file_location('roadmap',Path('/home/onnwee/.codex/skills/manage-roadmaps-and-issues/scripts/roadmap.py'))
    module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
    api=module.GitHub('subculture-collective/edda',runner=lambda args,**kw:subprocess.run(['rtk','proxy']+args,**kw))
    api.add_to_project({x['key']:x['issue'] for x in allrows},{x['key']:x['project_size'] for x in allrows},True,print)
    print('Published; exact full graph/body/Task/Size readback REQUIRED before acceptance',flush=True)


if __name__=='__main__':main()
