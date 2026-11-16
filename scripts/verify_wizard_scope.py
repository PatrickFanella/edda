#!/usr/bin/env python3
"""Verify actual #332 candidate semantics; checkpoint never implies publication."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
from wizard_contract_oracle import COMMON, CONTRACTS, expected_context_owners, expected_level_fixture
from wizard_spell_parts import PARTS, REUSED
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'docs/investigations'
PIN='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87'

def load(p):return json.loads(Path(p).read_text())
def keys(r):return set(r['required_mechanics']+r['option_keys'])|({r['requirement']['key']} if r.get('requirement') else set())
def check_plan(m):
    records=[]
    for p,batch in m['catalogues'].items():
        assert batch==load(ROOT/p),'actual candidate catalogue'
        records+=batch['records']
    assert records==m['records'] and len(records)==106,'all106 source contexts'
    oracle=expected_context_owners(records)
    assert len(m['record_matrix'])==106
    assert {c['record']['id']:c['record'] for c in m['record_matrix']}=={r['id']:r for r in records}
    assert {c['record']['id']:c['key_owners'] for c in m['record_matrix']}==oracle,'independent complete contextual ownership'
    tasks={t['key']:t for t in m['tasks']};assert len(tasks)==len(m['tasks'])
    for key,fixture in CONTRACTS.items():
        t=tasks[key];assert t['fixture']==fixture,('actual class fixture',key)
        assert t['body'].startswith('Decision #332; one implementation PR; Task Size M.\n\n'+fixture+'\n\n'+COMMON+'\n'),('actual class body',key)
    assert tasks['hp']['prerequisites']==[968] and 'hp' in tasks['later-entry']['depends'],'composite multiclass HP dependency'
    assert tasks['asi']['prerequisites']==list(range(384,402)), 'actual feat owners only'
    assert tasks['epic-boon']['prerequisites']==list(range(384,402))
    for c in m['record_matrix']:
        for k,o in c['key_owners'].items():assert o in tasks,('missing execution owner',o)
    # Ensure every required alternative reaches the granting level, even when
    # manifest, declarations and recomputed body digests are weakened together.
    def path(a,b,seen=None):
        if a==b:return True
        seen=set() if seen is None else seen
        if a in seen:return False
        seen.add(a)
        return any(path(x,b,seen) for x in tasks.get(a,{}).get('depends',[]) if not x.startswith(('330:','331:')))
    required={1:['core','skills','equipment','later-entry','casting','cantrips','book-initial','ritual','arcane-recovery'],2:['scholar'],3:['subclass','savant-initial','savant-new-slots','potent-attack','potent-save'],4:['asi','cantrips'],5:['memorize','savant-new-slots'],6:['sculpt'],7:['savant-new-slots'],8:['asi'],9:['savant-new-slots'],10:['empowered','cantrips'],11:['savant-new-slots'],12:['asi'],13:['savant-new-slots'],14:['overchannel-damage','overchannel-reuse'],15:['savant-new-slots'],16:['asi'],17:['savant-new-slots'],18:['mastery-select','mastery-cast','mastery-replace'],19:['epic-boon'],20:['signature-select','signature-cast','signature-recovery']}
    for n in range(1,21):
        t=tasks['level-'+str(n)]
        canonical=next(r for r in records if r['kind']=='level' and r['level']==n)
        assert t['fixture']==expected_level_fixture(n,canonical),('actual complete numeric level fixture',n)
        assert t['prerequisites']==[] and 'source-advancement-output' in t['depends']
        for k in required[n]+['hp','slots','preparation']:assert path(t['key'],k),('granting level path',n,k)
        if n>1:assert 'level-'+str(n-1) in t['depends'],'adjacent transition'
        assert 'TOTAL canonical level' in t['fixture'] and 'Wizard2 total6' in t['fixture'] and 'bool/fraction' in t['fixture'] and 'total21' in t['fixture']
    source=load(OUT/'issue-332-residual-source-oracle.json')
    expected={}
    for name,parts in PARTS.items():
        text=source[name]['text'];starts=[text.index('\n'+s)+1 for _,s in parts]
        assert sorted(starts)==starts
        header=text[:starts[0]]
        for j,(part,_) in enumerate(parts):
            slug=re.sub(r'[^a-z0-9]+','-',name.lower()).strip('-')
            key='spell-'+slug+'-'+part;clause=text[starts[j]:starts[j+1] if j+1<len(starts) else len(text)]
            expected[key]=(name,part,source[name]['page'],header,clause)
    assert len(expected)==179
    assert {k for k,t in tasks.items() if t.get('spell')}==expected.keys(),'complete179 exact parts'
    for k,(name,part,page,header,clause) in expected.items():
        t=tasks[k]
        assert (t['spell'],t['part'],t['source_page'])==(name,part,page)
        fixture=f'PDF{page} {name}, bounded {part}. Casting context:\n{header}\nExact source contract:\n{clause}\nPositive fixture executes these exact operands, alternatives, timing and outcomes. Independently violate each numeric/target/action/component/resource/timing predicate; assert source-correct outcome and no private disclosure. Other named parts retain separate execution owners.'
        assert t['fixture']==fixture,('actual numeric/alternative/timing spell fixture',k)
        assert t['body'].startswith('Decision #332; one implementation PR; Task Size M.\n\n'+fixture+'\n\n'+COMMON+'\n'),('actual spell body',k)
    for t in tasks.values():
        contexts=[dict(id=c['record']['id'],keys=[k for k,o in c['key_owners'].items() if o==t['key']]) for c in m['record_matrix'] if t['key'] in c['key_owners'].values()]
        assert t['owned_contexts']==contexts
        assert t['source_records']==[c['record'] for c in m['record_matrix'] if t['key'] in c['key_owners'].values()]
        expected_body=('Decision #332; one implementation PR; Task Size M.\n\n'+t['fixture']+'\n\n'+COMMON+'\n\nExact prerequisites: '+json.dumps(t['prerequisites'])+'; finite output dependencies: '+json.dumps(t['depends'])+'.\n'+'\nComplete source contexts:\n```json\n'+json.dumps(t['source_records'],indent=2)+'\n```\n')
        assert t['body']==expected_body,('whole candidate implementation body',t['key'])
        if t.get('spell'):assert any(d.startswith('casting-consumers-') for d in t['depends']),'real shared casting seam'
    assert len(m['spell_list'])==217 and len({x['name'] for x in m['spell_list']})==217
    from wizard_form_contracts import FORMS
    for key,(creature,page,scope) in FORMS.items():
        assert tasks[key]['fixture']==f'PDF{page}: '+scope,('actual form semantics',key)
        assert tasks[key]['creature']==creature and tasks[key]['source_page']==page
    if 'gates' in m:
        literal=load(ROOT/'scripts/wizard_qualification_oracle.json')
        assert m['gates']==literal['gates'],'complete literal qualification bodies, owners, rosters and protocol edges'
        assert m['spell_owners']==literal['spell_owners'],'all217 actual whole output owners'
        assert m['finite_protocol_rosters']==literal['finite_protocol_rosters'],'finite protocol consumers'
        rows={t['key']:t for t in m['tasks']+m['gates']}
        assert len(rows)==len(m['tasks'])+len(m['gates'])
        assert rows['shared-output']['parent']==332 and rows['class-gate']['parent']==332
        assert 'class-gate' not in rows['shared-output']['depends']
        for n in range(10):
            g=rows['printed-level-'+str(n)];assert g['parent']=='shared-output'
            assert all(r['name'] in g['body'] for r in m['spell_list'] if r['level']==n)
        for t in rows.values():assert len(set(t['depends'])|set(t['prerequisites']))<=100,'candidate incoming cap'
    return tasks

def column(pdf,p,x):
    return subprocess.check_output(['pdftotext','-f',str(p),'-l',str(p),'-x',str(x),'-y','0','-W','297','-H','783','-layout',str(pdf),'-'],text=True)

def source_check(m,pdf):
    assert hashlib.sha256(pdf.read_bytes()).hexdigest()==PIN
    import verify_srd_wizard as wizard
    import verify_srd_evoker as evoker
    wizard.compare(m['catalogues']['internal/rules/srd521/testdata/character-wizard.json']['records'],wizard.source_records(pdf))
    evoker.verify(pdf)
    lists={}
    for cls,pages in [('Wizard',range(79,83)),('Sorcerer',range(67,70)),('Warlock',range(75,77))]:
        rows=[];level=0
        for p in pages:
            for x in [0,297]:
                for line in column(pdf,p,x).splitlines():
                    match=re.search('Level (\\d) '+cls+' Spells',line)
                    if match:level=int(match[1])
                    if 'Cantrips (Level 0' in line:level=0
                    match=re.match(r'^\s*(.+?)\s{2,}(Abjuration|Conjuration|Divination|Enchantment|Evocation|Illusion|Necromancy|Transmutation)\s+(.+?)\s*$',line)
                    if match:rows.append(dict(name=match[1].strip(),school=match[2],special=match[3],level=level,page=p))
        lists[cls]=rows
    assert lists['Wizard']==m['spell_list'],'actual complete217 PDF list'
    names={r['name'] for r in load(ROOT/'docs/investigations/issue-80-spell-item-inventory.json')['records'] if r['family']=='spells'}
    lines=[]
    for p in range(107,176):
        for x in [0,297]:
            lines +=[(p,' '.join(t.replace('\u00ad','').split())) for t in column(pdf,p,x).splitlines() if t.strip() and 'System Reference Document' not in t]
    heads=[i for i,(_,t) in enumerate(lines[:-1]) if t in names and re.match(r'(?:Level [1-9] |\w+ Cantrip)',lines[i+1][1])]
    assert len(heads)==339
    sources={lines[i][1]:dict(page=lines[i][0],text=re.sub(r'(\w)-\n(\w)',r'\1\2','\n'.join(t for _,t in lines[i:heads[j+1] if j+1<len(heads) else len(lines)]))) for j,i in enumerate(heads)}
    actual=load(OUT/'issue-332-residual-source-oracle.json')
    assert {n:sources[n] for n in actual}==actual,'every actual source-oracle clause compared with PDF'
    conflicts=[]
    for cls,rows in lists.items():
        printed={x['name'] for x in rows};headed=set()
        for n,s in sources.items():
            header=s['text'].split('Casting Time:')[0]
            if re.search(r'\b'+cls+r'\b',header):headed.add(n)
        assert not printed-headed
        conflicts +=[(cls,n,sources[n]['page']) for n in sorted(headed-printed)]
    assert set(conflicts)=={('Sorcerer','Mind Spike',149),('Sorcerer','Phantasmal Force',151),('Wizard','Phantasmal Force',151)},'exact three source conflict pairs'
    from wizard_form_source import inventory
    actual_dispatch=load(OUT/'issue-332-form-dispatch.json')
    expected_dispatch=inventory(pdf)
    assert len(actual_dispatch)==len(expected_dispatch)==235
    for a,e in zip(actual_dispatch,expected_dispatch):
        assert {k:a[k] for k in e if k!='execution_status'}=={k:v for k,v in e.items() if k!='execution_status'},'complete actual typed dispatch eligibility'
        assert bool(a['proposed_output_bindings'])==(a['name'] in ['Baboon','Badger','Bat','Cat','Crab','Deer','Eagle','Frog','Giant Fire Beetle','Goat','Hawk','Hyena','Jackal','Lizard','Octopus','Owl','Piranha','Rat','Raven','Scorpion','Seahorse','Spider','Vulture','Weasel','Riding Horse','Commoner']),'bounded actual form output binding'
    return conflicts

def main():
    p=argparse.ArgumentParser();p.add_argument('--manifest',type=Path,default=OUT/'issue-332-wizard-scope.json');p.add_argument('--pdf',type=Path);p.add_argument('--final',action='store_true');a=p.parse_args()
    m=load(a.manifest);check_plan(m)
    if a.pdf:print('PDF/source conflicts:',source_check(m,a.pdf))
    if a.final:
        assert m['native_publication']=='PUBLISHED' and not m.get('unresolved_execution_seams'),'publication incomplete'
        assert m['accepted_parent_prefix'] is not None,'accepted POST-WARLOCK prefix release required'
        raise AssertionError('final native seal verifier must be completed after approved publication; no synthetic final proof')
    print('PASS checkpoint: actual source contracts/contexts; native publication and runtime qualification remain PENDING')
if __name__=='__main__':main()
