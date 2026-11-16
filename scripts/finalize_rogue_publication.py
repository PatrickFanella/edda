#!/usr/bin/env python3
"""Apply parent-authorized own-contract corrections before the single publication readback."""
import json
import hashlib
from pathlib import Path
from datetime import datetime,timezone
from collections import Counter
from build_rogue_checkpoint import audit
from publish_rogue_scope import rows,PATH,JOURNAL,call
from rogue_capacity import project
from verify_fighter_scope import graph_read
from render_rogue_decision import render

def main():
    j=json.loads(JOURNAL.read_text());assert j['operations'].get('project-metadata'),'initial journal must finish first'
    previous=json.loads(PATH.read_text());assert previous.get('final_seal')!='PASS','Final publication is immutable; verify only'
    m=json.loads(json.dumps(audit(write=False)))
    graph=graph_read();ours={x['issue'] for x in j['issues'].values()}
    base=[dict(x,blocked_by=[n for n in x['blocked_by'] if n not in ours]) for x in graph if x['number'] not in ours]
    projection=project(m,base);assert projection['new_tasks']==77
    print('Fresh before-correction capacity/BOTH DAGs PASS',projection['maxima'],flush=True)
    owned=rows(m,j['before_graph'],j['issues']);old={x['key']:x for x in previous['published_issues']}
    def save():
        temp=JOURNAL.with_suffix('.tmp');temp.write_text(json.dumps(j,indent=2)+'\n');temp.replace(JOURNAL)
    for x in owned:
        if x['body']!=old[x['key']]['body']:
            key='corrected-body:'+str(x['issue'])+':'+hashlib.sha256(x['body'].encode()).hexdigest()
            if key not in j['operations']:
                call('PATCH',f"repos/subculture-collective/edda/issues/{x['issue']}",dict(title=x['title'],body=x['body'],type='Task'));j['operations'][key]=True;save()
        for dep in sorted(set(x['blocked_by_issues'])-set(old[x['key']]['blocked_by_issues'])):
            key=f'edge:{x["issue"]}:{dep}'
            if key in j['operations']:continue
            remote=call('GET',f'repos/subculture-collective/edda/issues/{dep}')
            call('POST',f'repos/subculture-collective/edda/issues/{x["issue"]}/dependencies/blocked_by',dict(issue_id=remote['id']));j['operations'][key]=True;save()
    m['published_issues']=owned
    # Complete actual old source/shared contracts, not abbreviated headings or excerpts.
    mandatory={285,294,316,429,783,851,779,774,512,906,967,1064,614}|{n for x in owned for n in x['blocked_by_issues'] if n not in ours and n not in [279,280,329]}
    before={x['number']:x for x in j['before_graph']}
    todo=list(mandatory)
    while todo:
        n=todo.pop()
        for dep in before[n]['blocked_by']:
            if dep not in mandatory:mandatory.add(dep);todo.append(dep)
    m['consumed_contracts']={str(n):before[n] for n in sorted(mandatory)}
    m['publication']='NATIVE77Task publication complete; final parent prefix seal PENDING accepted POST-RANGER release/ROOT appends'
    m['publication_at']=datetime.now(timezone.utc).isoformat()
    PATH.write_text(json.dumps(m,indent=2)+'\n')
    (PATH.parent/'issue-329-capacity-proposal.json').write_text(json.dumps(projection,indent=2)+'\n')
    for n in [47,54,283,337]:
        content='Rogue #329: #1093; FINAL pending.'
        if n!=47:content='Rogue #329:77Tasks #1093–1169 (72 implementation leaves,five gates). Qualification #1093 consumes feature #1094/level #1095 cohorts and279/280. Shared #1096 also gates279/280; finite execution gateway #1097 owns qualification only. Universal339spell identity/typed dispatch differs from conditional actual effect execution. All owned paths remain mandatory after329closure. No runtime/FullRules promotion. FINAL prefix seal PENDING accepted POST-RANGER release and ROOT serial append.'
        section=f'\n<!-- edda-329:start -->\n{content}\n<!-- edda-329:end -->\n';assert len(section)<=(90 if n==47 else 750)
        (PATH.parent/f'issue-329-parent-{n}-PENDING.md').write_text(section)
    render(m)
    print('Finalized',len(owned),'Tasks',sum(len(x['blocked_by_issues']) for x in owned),'owned edges+3root; FINAL seal HELD',flush=True)

if __name__=='__main__':main()
