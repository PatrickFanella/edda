#!/usr/bin/env python3
"""Publish only parent-reserved Rogue tasks; durable journal, no old body/parent edits."""
import argparse
import importlib.util
import json
from pathlib import Path
import subprocess
from verify_fighter_scope import graph_read
from rogue_capacity import project
from publish_paladin_scope import call

ROOT=Path(__file__).resolve().parents[1]
PATH=ROOT/'docs/investigations/issue-329-rogue-scope.json'
JOURNAL=ROOT/'docs/investigations/issue-329-publication-journal.json'
GATES=['qualification','features','levels','shared','execution']

def body(m,x):
    source=x.get('source_context',[])
    return (f"<!-- edda-329:{x['key']} -->\nDecision #329; roadmap #47. One reviewable implementation/qualification PR; Task Size M.\n\n"
      f"Source: English SRD5.2.1,exact PDF locators below; SHA256 {m['source_sha256']}. CC-BY-4.0 attribution/adaptation: docs/srd-5.2.1-conformance.md; fixture/ownership decisions are Edda additions.\n\n"
      f"Exact scope and deterministic positive/negative controls:\n{x['fixture']}\n\n{m['common_acceptance']}\n\n"
      "Run GOMAXPROCS=2 go test -p 2 -short -race ./... and GOMAXPROCS=2 go vet -p 2 ./...; isolated PostgreSQL handler/state rollback/replay/reload/concurrency fixtures are separately mandatory. Changed web: build,focused lint,desktop/phone confirmation/reconnect; trusted TUI availability/confirmation fixtures.\n\n"
      "Ownership: each feature test uses isolated entitlements; source import,Task closure and planning277/278/273 never implement effects. Actual qualified source-bound output required before availability. Generic canonical identity/eligibility/typed-dispatch permission does NOT imply all-spell/item execution. Ordinary feat selection adds no blanket runtime Spellcasting/Concentration/encounter prerequisite. SpellRecall399 needs authoritative actual qualified Spellcasting; pureRogue,MagicInitiate386 and UMD alone do not grant it. HeroicInspiration774 differs from Bardic512. Existing rest427/428,transaction429,provenance430,conditions487/488/504/727/777/778,playerUnarmed779,Dash780/Disengage781/Dodge782 and feats385–401 keep exact scopes. No old bodies/parents/prerequisites are widened.\n\n"
      f"Native parent #{x['parent']}; exact blockers: "+', '.join('#'+str(n) for n in x['blocked_by_issues'])+".\n"
      "Mandatory path: this terminal -> finite cohort/qualification -> #337,including after #329 closes. Shared cohort also blocks #279/#280; full production class qualification consumes BOTH279/280. Narrow fixtures and universal eligibility/dispatcher tests are separate; broad Rogue/FullRules remain unavailable.\n\n"
      'Exact source/mechanic keys:\n'+'\n'.join('- '+k for k in x.get('source_keys',[]))+
      '\n\nComplete owned canonical source records (references are context,not duplicated runtime):\n```json\n'+json.dumps(source,indent=2)+'\n```\n')

def rows(m,g,ids):
    planned,virtual,consumers=project(m,g,True)
    reverse={v:k for k,v in virtual.items()}
    real=lambda n:ids[reverse[n]]['issue'] if n<0 else n
    result=[]
    for t in m['issues']:
        x=dict(t,issue=ids[t['key']]['issue'],publication='PUBLISHED contract; runtime UNQUALIFIED')
        p=planned[virtual[t['key']]]
        x.update(parent=real(p['parent']),blocked_by_issues=sorted(real(n) for n in p['blocked_by']))
        x['body']=body(m,x);result.append(x)
    for key in GATES:
        p=planned[virtual[key]]
        x=dict(key=key,issue=ids[key]['issue'],title='Qualify finite Rogue '+key+' contracts',project_size='M',parent=real(p['parent']),blocked_by_issues=sorted(real(n) for n in p['blocked_by']))
        roster=consumers if key=='execution' else [t['key'] for t in result if t['issue'] in x['blocked_by_issues']]
        x['fixture']='Finite prerequisite qualification ONLY; no new runtime implementation or eligibility requirement. Every listed blocker remains mandatory after329closure. Exact consumer roster: '+', '.join(roster)+'.'
        if key=='execution':
            x['fixture']+=' ONLY61check/save receipts,64inertprepare/durableconfirm/replay,65attackhit/damage,66actor/turn/actioncost events,67trustedtarget/range/sight/cover;273actual qualified casting-mode outputs (planning closure supplies NONE);353source-bound concentration start/maintain/end events ONLY when selected spell/mode requires them. Item/scroll routers and exact qualified effect owners must prove PDF206/244 modes: normal casting time,no slot,component exceptions,source access instead of class prepared eligibility; selected spell concentration remains with user. Do not import906class prepared/material requirements into item/scroll use. Preserve Spell Scroll fixed DC/attack table. Qualify actual outputs independently in each named consumer. No consumer can block this gateway; no effect/class/transaction/casting implementation owned here. No activeencounter/Spellcasting/Concentration prerequisite introduced.'
        elif key=='qualification':x['fixture']+=' Full production class qualification also requires actual279spell/280item qualifications; universal scroll identity permission alone never qualifies all effects. Keep own feature and level cohorts mandatory.'
        x['source_keys']=['srd-5.2.1:mechanic/rogue-qualification/'+key];x['source_context']=[];x['body']=body(m,x);result.append(x)
    assert len(result)==77 and all(len(x['body'])<=65536 and len(x['blocked_by_issues'])<=100 for x in result)
    return result

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--apply',action='store_true');a=p.parse_args()
    m=json.loads(PATH.read_text())
    assert not(a.apply and m.get('final_seal')=='PASS'),'Final publication is immutable; verify only'
    graph=graph_read()
    if JOURNAL.exists():
        j=json.loads(JOURNAL.read_text());ours={x['issue'] for x in j['issues'].values()}
        base=[dict(x,blocked_by=[n for n in x['blocked_by'] if n not in ours]) for x in graph if x['number'] not in ours]
    else:
        base=graph;j=dict(issues={},operations={},before_graph=base)
    projection=project(m,base);assert projection['new_tasks']==77
    print('Fresh prewrite caps and BOTH DAGs PASS',projection['maxima'],flush=True)
    rows(m,base,{key:dict(issue=200000+i) for i,key in enumerate(GATES+[t['key'] for t in m['issues']])})
    if not a.apply:return
    def save():
        temp=JOURNAL.with_suffix('.tmp');temp.write_text(json.dumps(j,indent=2)+'\n');temp.replace(JOURNAL)
    save()
    for key,title in [(k,'Qualify finite Rogue '+k+' contracts') for k in GATES]+[(t['key'],t['title']) for t in m['issues']]:
        if key in j['issues']:continue
        # Markers are unique across the complete readback before creating anything.
        assert not any(f'<!-- edda-329:{key} -->' in x.get('body','') for x in graph),'untracked existing owned issue'
        d=call('POST','repos/subculture-collective/edda/issues',dict(title=title,body=f'<!-- edda-329:{key} -->\nReserved publication in progress; no executable qualification.',type='Task'))
        j['issues'][key]={k:d[k] for k in ['id','node_id']};j['issues'][key]['issue']=d['number'];save();print('created',key,d['number'],flush=True)
    owned=rows(m,base,j['issues']);m['published_issues']=owned;m['publication']='NATIVE publication in progress; FINAL parent seal PENDING';PATH.write_text(json.dumps(m,indent=2)+'\n')
    def once(key,fn):
        if key in j['operations']:return
        fn();j['operations'][key]=True;save()
    for x in owned:
        once(f"body:{x['issue']}",lambda x=x:call('PATCH',f"repos/subculture-collective/edda/issues/{x['issue']}",dict(title=x['title'],body=x['body'],type='Task')))
        once(f"parent:{x['issue']}",lambda x=x:call('POST',f"repos/subculture-collective/edda/issues/{x['parent']}/sub_issues",dict(sub_issue_id=j['issues'][x['key']]['id'])))
    idmap={v['issue']:v['id'] for v in j['issues'].values()}
    relations=[(x['issue'],n) for x in owned for n in x['blocked_by_issues']]+[(337,j['issues']['qualification']['issue']),(279,j['issues']['shared']['issue']),(280,j['issues']['shared']['issue'])]
    for n,dep in relations:
        if dep not in idmap:idmap[dep]=call('GET',f'repos/subculture-collective/edda/issues/{dep}')['id']
        once(f'edge:{n}:{dep}',lambda n=n,dep=dep:call('POST',f'repos/subculture-collective/edda/issues/{n}/dependencies/blocked_by',dict(issue_id=idmap[dep])))
    spec=importlib.util.spec_from_file_location('roadmap',Path('/home/onnwee/.codex/skills/manage-roadmaps-and-issues/scripts/roadmap.py'))
    module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
    api=module.GitHub('subculture-collective/edda',runner=lambda args,**kw:subprocess.run(['rtk','proxy']+args,**kw))
    once('project-metadata',lambda:api.add_to_project({x['key']:x['issue'] for x in owned},{x['key']:'M' for x in owned},True,print))
    m['publication']='NATIVE publication complete; exact readback and FINAL parent seal PENDING';PATH.write_text(json.dumps(m,indent=2)+'\n');print('77 Tasks published;',len(relations),'edges;sealed readback REQUIRED',flush=True)

if __name__=='__main__':main()
