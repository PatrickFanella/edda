"""Construct exact finite output rosters; never writes to GitHub."""
import json
from pathlib import Path
from build_wizard_scope import body
from wizard_spell_parts import REUSED

CAST_SEAMS=('cast-components','cast-cost-privacy','cast-long-ritual','effect-time-span-stack')

def build(m,s,w):
    own=m['tasks'];gates=[]
    def gate(key,parent,depends,prerequisites,fixture):
        t=dict(key=key,title='Qualify Wizard '+key,role='qualification',type='Task',size='M',parent=parent,depends=sorted(set(depends)),prerequisites=sorted(set(prerequisites)),fixture=fixture,issue=None)
        t['body']=body(t)+'\nOwn qualification only. No effect implementation or class admission. Preserve every prior prerequisite and exact source boundary; root-approved source-bound additive obligations require semantic proof and native readback. Decision closure does not qualify these outputs.\n'
        gates.append(t);return t
    # Exact foreign statblocks are shared outputs, not familiar entitlement.
    rows={t['key']:t for t in own}
    for key in ['spell-animate-dead-creation','spell-animate-dead-command']:
        rows[key]['depends']+=['331:creature-skeleton-statistics','331:creature-skeleton-attacks','330:zombie-stat-actions','330:zombie-undead-fortitude']
        rows[key]['depends']=sorted(set(rows[key]['depends']))
        t=rows[key]
        t['body']=body(t)+'\nComplete source contexts:\n```json\n'+json.dumps(t['source_records'],indent=2)+'\n```\n'
    gate('source-advancement-output',332,[],[285],
         'Qualification ONLY of WHOLE canonical source285: all20 total-level XP/proficiency rows including complete multilevel admitted configuration. Exact22 consumers: '+json.dumps(['slots','multiclass-slots']+['level-'+str(n) for n in range(1,21)])+'. Wizard class rows/HP/features remain separate. Wizard1/total5 PB3, Wizard2/total6 PB3; integers1-20 not bool/fraction, total21 invalid. Own no runtime transaction, HP, die, class eligibility or widened285/968 scope.')
    for key in ['slots','multiclass-slots']+['level-'+str(n) for n in range(1,21)]:
        rows[key]['prerequisites'].remove(285);rows[key]['depends'].append('source-advancement-output')
    for spellkey in ['spell-phantom-steed-create','spell-phantom-steed-ride-ending']:
        rows[spellkey]['depends']+=['form-riding-horse-statistics','form-riding-horse-hooves']
    normal=w['familiar_source_roster']['normal']
    form_refs=['331:'+t['key'] for t in w['creature_tasks'] if t.get('creature') in normal]
    gate('qualified-form-output',332,form_refs+list(k for k in rows if k.startswith('form-') and k!='form-qualified-snapshot'),[],
         'Exact named generic Beast outputs: '+json.dumps(normal+['Riding Horse'])+'; Humanoid Commoner. Preserve whole source outputs; no familiar Celestial/Fey/Fiend transformation, no familiar control/turn/attack restrictions. Shapechange retains seen/CR/type eligibility; unbound canonical typed output IDs remain unavailable, never redefine source permission. Simulacrum retains both Beast and Humanoid source alternatives; only actual qualified snapshot outputs execute.')
    for key,t in rows.items():
        if t.get('spell') in ['Shapechange','Simulacrum']:t['depends'].append('qualified-form-output')
        if t.get('spell')=='Simulacrum':t['depends'].append('form-qualified-snapshot')
    rows['form-qualified-snapshot']['depends']=['331:'+t['key'] for t in w['creature_tasks'] if t.get('creature')=='Cat']+['form-commoner-statistics','form-commoner-training','form-commoner-club']
    gate('form-cohort',277,[k for k in rows if k.startswith('form-')],[], 'Exact6 generic stat/action/snapshot outputs; no generic all-form execution. All source-eligible unbound dispatch IDs stay unavailable.')
    owners={}
    for row in m['spell_list']:
        name=row['name'];refs=[]
        if name in REUSED:refs+=[str(n) for n in REUSED[name]]
        refs+=['332:'+t['key'] for t in own if t.get('spell')==name]
        if not refs:
            refs=['330:'+t['key'] for t in s['shared_split_proposal'] if t.get('spell')==name]
            if not refs:
                refs=['331:'+t['key'] for t in w['shared_effect_tasks'] if t.get('spell')==name]
            if not refs:
                # Full source bodies were collected by earlier audits; retain
                # exact snapshots, not titles, as evidence for reuse review.
                entries=s.get('overlap_audit',{}).get(name,[])
                refs=[str(e['issue']) for e in entries]
            if not refs:
                entries=next((x for x in w['spell_overlap_audit'] if x['name']==name),None)
                if entries:refs=[str(e['issue']) for e in entries['candidate_owners']]
        assert refs,('missing effect owners',name)
        owners[name]=sorted(set(refs))
    # Find Familiar requires actual normal statblocks, actions and lifecycle,
    # never only the class-specific Chain cast adapter or planning274.
    familiar=w['familiar_source_roster']['normal']+['Find Familiar']
    owners['Find Familiar']=sorted(set(owners['Find Familiar'])|{'331:'+t['key'] for t in w['creature_tasks'] if t.get('creature') in familiar})
    gate('casting-output',332,['330:'+x for x in CAST_SEAMS],[],
         'Only source-qualified modes for exact217 printed Wizard identities; require real four finite245 casting/component/cost/privacy/long-ritual/stack outputs from330. No273/277 planning closure, arbitrary future effect, class qualifier dependency or inference.')
    for key in sorted({t['parent'] for t in own if t.get('spell')}):
        deps=[t['key'] for t in own if t['parent']==key]
        gate(key,277,deps,[], 'Exact bounded Wizard residual runtime parts: '+json.dumps(deps)+'. Each source paragraph, numeric operand, alternative and timing predicate must execute against actual named outputs. This cohort owns qualification only.')
    cohorts=[]
    for level in range(10):
        spells=[r for r in m['spell_list'] if r['level']==level]
        refs=sorted({ref for r in spells for ref in owners[r['name']]})
        slices=[]
        for i in range(0,len(refs),80):
            key=f'printed-level-{level}-outputs-{i//80+1}'
            gate(key,'printed-level-'+str(level),refs[i:i+80],[],
                 f'Printed Wizard spell level{level}, mandatory finite outputs slice{i//80+1}: '+json.dumps(refs[i:i+80])+'. Every listed output required; metadata and planned heads supply no execution.')
            slices.append(key)
        key='printed-level-'+str(level);cohorts.append(key)
        gate(key,'shared-output',slices,[],
             f'Exact printed Wizard spell level{level} source identities, school, special flags and page: '+json.dumps(spells,ensure_ascii=False)+
             '. Whole source-specific output roster: '+json.dumps({r['name']:owners[r['name']] for r in spells},ensure_ascii=False)+
             '. No class preparation/eligibility granted by this qualification. Preserve all named outputs, variants and split parts; no duplicate effects. Source conflict ledger retains Wizard Phantasmal Force outside list.')
    gate('shared-output',332,cohorts+['casting-output','qualified-form-output','form-cohort']+sorted({t['parent'] for t in own if t.get('spell')}),[],
         'Exact217 printed Wizard identities ONLY: '+json.dumps(sorted(owners),ensure_ascii=False)+
         '. Sole Wizard input to reserved330/331/332 shared aggregate under191; path through aggregate to279/280, and through own class-gate to337. Never depends on class qualifier. All actual effect, lifecycle and named-form outputs required; unknown effect variants unavailable.')
    for key in ['feature-cohort','level-cohort']:
        deps=[t['key'] for t in own if t['parent']==key]
        gate(key,332,deps,[], 'Exact Wizard execution roster: '+json.dumps(deps)+'. Every granting class level reaches all required execution alternatives.')
    gate('class-gate',332,['feature-cohort','level-cohort','shared-output'],[332],
         'Qualify all106 source contexts,20 individual Wizard admissions and adjacent transitions, first-character and later-entry, exact Evoker alternatives and shared outputs. Mandatory blocker of337 only. Closed332 remains a source decision, never implementation.')
    # Exact named effect consumers. Broad stored-spell permission dispatches only
    # to qualified outputs; an unbound ID remains unavailable without narrowing
    # the source's eligibility rule to this class reservation.
    def effect_deps(key,*names):
        rows[key]['depends']+=list({ref for name in names for ref in owners[name]})
    for key,names in {
        'spell-guards-and-wards-doors':['Arcane Lock'],
        'spell-guards-and-wards-stairs':['Web'],
        'spell-guards-and-wards-dancing-lights':['Dancing Lights'],
        'spell-guards-and-wards-magic-mouth':['Magic Mouth'],
        'spell-guards-and-wards-stinking-cloud':['Stinking Cloud'],
        'spell-guards-and-wards-gust':['Gust of Wind'],
        'spell-guards-and-wards-suggestion':['Suggestion'],
        'spell-prismatic-wall-green':['Passwall'],
        'spell-prismatic-wall-orange':['Gust of Wind'],
        'spell-prismatic-wall-indigo':['Daylight'],
        'spell-prismatic-wall-violet':['Dispel Magic'],
        'spell-resilient-sphere-disintegrate':['Disintegrate'],
        'spell-wall-of-force-barrier':['Disintegrate'],
        'spell-magic-jar-possession':['Protection from Evil and Good','Magic Circle'],
        'spell-modify-memory-restore':['Remove Curse'],
    }.items():
        if all(n in owners for n in names):effect_deps(key,*names)
        else:
            # Daylight and Greater Restoration are not Wizard-list identities;
            # their actual existing source owners remain explicit prerequisites.
            if key=='spell-prismatic-wall-indigo':rows[key]['prerequisites'].append(934)
    rows['spell-modify-memory-restore']['prerequisites'].append(625)
    # Named conditional dispatch remains separate from complete source permission.
    for key in ['spell-contingency-store','spell-contingency-trigger','spell-glyph-of-warding-stored-spell']:
        rows[key]['depends'].append('casting-output')
    condition_consumers=sorted(k for k,t in rows.items() if t.get('spell') and any(word in t['fixture'] for word in ['Blinded','Charmed','Frightened','Restrained','Incapacitated','Unconscious','Invisible','Stunned','Petrified']))
    gate('condition-output',332,['330:condition-petrified'],[487,488,504,727,777,960,962,963],
         'Qualification only. WHOLE source-bound condition outputs, retained immunity/ending/source provenance and observer consequences. Exact finite consumers: '+json.dumps(condition_consumers)+'. Each own spell part still implements its actual trigger/save/timing, not generic condition duplicates; no runtime blanket condition prerequisite.')
    for key in condition_consumers:rows[key]['depends'].append('condition-output')
    # One old285 edge, and finite casting consumers without runtime restrictions.
    casting=sorted(k for k,t in rows.items() if 'casting-output' in t['depends'])
    for start in range(0,len(casting),80):
        members=casting[start:start+80];key='casting-consumers-'+str(start//80+1)
        gate(key,332,['casting-output'],[], 'Qualification only. Exact finite casting output consumer roster: '+json.dumps(members)+'. No runtime Spellcasting, active Concentration or active encounter requirement; each source-qualified casting mode retains its own rules.')
        for k in members:rows[k]['depends']=[key if d=='casting-output' else d for d in rows[k]['depends']]
    execution=sorted(k for k,t in rows.items() if t.get('spell') or t.get('creature') or k in ['casting','potent-attack','potent-save','sculpt','empowered','overchannel-damage','overchannel-reuse'])
    gate('execution-output',332,[],[61,64,65,66,67,353], 'Qualification only of actual whole transaction/dice/target/action/concentration/duration outputs for exact finite named Wizard consumers: '+json.dumps(execution)+'. No implementation, no runtime prerequisite added, no widening978/1097 Ranger/Rogue rosters.')
    for start in range(0,len(execution),80):
        members=execution[start:start+80];key='execution-consumers-'+str(start//80+1)
        gate(key,332,['execution-output'],[], 'Qualification only. Exact finite execution consumer roster: '+json.dumps(members)+'. Require each relevant actual primitive output in leaf tests, not gateway closure.')
        for k in members:rows[k]['depends'].append(key)
    for t in own:
        t['depends']=sorted(set(t['depends']));t['prerequisites']=sorted(set(t['prerequisites']))
        t['body']=body(t)+'\nComplete source contexts:\n```json\n'+json.dumps(t['source_records'],indent=2)+'\n```\n'
    m['gates']=gates;m['spell_owners']=owners
    m['finite_protocol_rosters']=dict(advancement=['slots','multiclass-slots']+['level-'+str(n) for n in range(1,21)],casting=casting,execution=execution)
    m['unresolved_execution_seams']=[
        'Animate Dead: Skeleton statistics/attacks consume331 and Zombie/Fortitude consume330; own creation/commands/control still require exact protocol outputs.',
        'Phantom Steed: actual Riding Horse statblock with source100ft/13mph exceptions.',
        'Shapechange: enumerate finite seen eligible SRD forms by CR, excluding Construct/Undead, and reuse exact whole statblock/action owners; no any-form implementation claim.',
        'Simulacrum: actual Beast/Humanoid authorized source snapshot, retained statistics/action outputs and construct/repair/command lifecycle.',
        'Named spell interoperability: root-approved finite cross-output dependency routing without runtime cycles, including Glyph/Contingency stored spells, Guards and Wards five alternatives, and Prismatic Wall destruction exceptions.',
        'Complete foreign330/331 graph, all native caps, both DAGs and exact primitive consumer roster before publication.',
    ]
    return m
if __name__=='__main__':
    import argparse
    p=argparse.ArgumentParser();p.add_argument('sorcerer');p.add_argument('warlock');a=p.parse_args()
    out=Path(__file__).resolve().parents[1]/'docs/investigations/issue-332-wizard-scope.json'
    m=build(json.loads(out.read_text()),json.loads(Path(a.sorcerer).read_text()),json.loads(Path(a.warlock).read_text()));out.write_text(json.dumps(m,indent=2,ensure_ascii=False)+'\n')
    print(len(m['tasks']),'tasks',len(m['gates']),'qualification nodes; unresolved execution seams remain explicit')
