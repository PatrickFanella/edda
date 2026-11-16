#!/usr/bin/env python3
"""Build a local review candidate. This command never publishes native issues."""
import copy
import hashlib
import json
from pathlib import Path
import re
from wizard_contract_oracle import CONTRACTS, COMMON
from wizard_spell_parts import PARTS, REUSED
from wizard_form_contracts import FORMS
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'docs/investigations'

def load(p): return json.loads(Path(p).read_text())
def slug(n): return re.sub(r'[^a-z0-9]+','-',n.lower()).strip('-')
def body(t):
 return (f"Decision #332; one {'qualification' if t['role']=='qualification' else 'implementation'} PR; Task Size M.\n\n"
         +t['fixture']+'\n\n'+COMMON+'\n\nExact prerequisites: '+json.dumps(t['prerequisites'])+
         '; finite output dependencies: '+json.dumps(t['depends'])+'.\n')

def execution_owner(r,k):
 ident=r['id'].split(':character/')[1];tail=ident.rsplit('/',1)[-1];word=k.split(':mechanic/')[-1]
 if r['kind']=='level':return 'level-'+str(r['level'])
 if tail.startswith('table-'):return 'slots'
 if r['kind']=='class':
  if word.endswith('/hit-die-sides'):return 'hp'
  if '/multiclass/slots-' in word:return 'multiclass-slots'
  if '/multiclass/' in word:return 'later-entry'
  return 'core'
 if ident=='feature/wizard/spellcasting':
  if '/spellbook/' in word:return 'book-identity'
  if word.endswith('wizard-spell-list'):return 'membership'
  if 'cantrip' in word:return 'cantrips'
  if word.endswith('regain-all-slots-on-long-rest'):return 'slots'
  if word.endswith('change-any-prepared-from-book-on-long-rest'):return 'prepare-rest'
  if 'prepared' in word or 'prepare-from' in word:return 'preparation'
  return 'casting'
 if ident=='feature/wizard/spell-mastery':
  return 'mastery-replace' if ('replace' in word or 'replacement-count' in word) else 'mastery-select' if word.endswith('always-prepared') else 'mastery-cast'
 if ident=='feature/wizard/signature-spells':return 'signature-select' if word.endswith('always-prepared') else 'signature-cast'
 if ident=='feature/wizard/evoker/overchannel':return 'overchannel-reuse' if any(w in word for w in ['reuse','adverse','immediately','ignores','further','die-sides']) else 'overchannel-damage'
 if '/evoker/' in ident:
  if tail.startswith('identity'):return 'evoker-identity'
  if 'new-slots' in tail:return 'savant-new-slots'
  if 'savant' in tail:return 'savant-initial'
  if 'potent' in tail:
   return 'potent-save' if word.endswith('target-succeeds-on-save-against-cantrip') else 'potent-attack'
  if 'sculpt' in tail:return 'sculpt'
  if 'empowered' in tail:return 'empowered'
  if 'overchannel' in tail:return 'overchannel-damage'
 parent=r['parent_id'].rsplit('/',1)[-1]
 aliases={'starting-equipment':'equipment','spellbook-initial':'book-initial','spellbook-level-gain':'book-learning','prepared-spells':'preparation','prepared-slot-level':'preparation','copy-found-spell':'copy-found','copy-own-book':'copy-own','replace-lost-book':'lost-book','ritual-adept':'ritual','ritual-eligibility':'ritual','arcane-recovery-budget':'arcane-recovery','scholar-proficiency':'scholar','wizard-subclass':'subclass','ability-score-improvement':'asi','ability-score-improvement-qualification':'asi','epic-boon-qualification':'epic-boon','memorize-spell':'memorize','memorize-replacement':'memorize','mastery-level-1':'mastery-select','mastery-level-2':'mastery-select','signature-spells':'signature-select','signature-eligibility':'signature-select','signature-first':'signature-cast','signature-second':'signature-cast','signature-first-binding':'signature-cast','signature-second-binding':'signature-cast'}
 return aliases.get(tail,tail)

def build():
 records=[];catalogues={}
 for n in ['wizard','evoker']:
  p=f'internal/rules/srd521/testdata/character-{n}.json';catalogues[p]=load(ROOT/p);records+=catalogues[p]['records']
 matrix=[]
 for r in records:
  keys=set(r['required_mechanics']+r['option_keys'])
  if r.get('requirement'):keys.add(r['requirement']['key'])
  matrix.append(dict(record=copy.deepcopy(r),key_owners={k:execution_owner(r,k) for k in sorted(keys)}))
 tasks=[]
 for k,fixture in CONTRACTS.items():
  pre={'hp':[968],'core':[430],'skills':[430],'later-entry':[430],'slots':[285,428],'multiclass-slots':[285,428],'asi':list(range(384,402)),'epic-boon':list(range(384,402))}.get(k,[])
  depends={'later-entry':['hp','casting','ritual','arcane-recovery','book-initial','cantrips','preparation','slots','multiclass-slots'], 'casting':['membership','preparation','slots','casting-output'], 'book-initial':['membership','book-identity'], 'book-learning':['membership','slots','book-identity'], 'copy-found':['membership','preparation','book-identity'], 'copy-own':['book-identity'], 'lost-book':['copy-own','preparation'], 'prepare-rest':['preparation'], 'preparation':['membership','book-identity'], 'ritual':['book-identity','casting-output'], 'arcane-recovery':['slots','book-identity'], 'subclass':['evoker-identity','savant-initial','savant-new-slots','potent-attack','potent-save'], 'mastery-select':['book-identity','preparation'], 'mastery-cast':['mastery-select','casting-output'], 'mastery-replace':['mastery-select'], 'signature-select':['book-identity'], 'signature-cast':['signature-select','casting-output'], 'signature-recovery':['signature-cast'], 'savant-initial':['membership','book-identity'], 'savant-new-slots':['membership','book-identity','slots'], 'sculpt':['evoker-identity','shared-output'], 'empowered':['evoker-identity','shared-output'], 'potent-attack':['evoker-identity','shared-output'], 'potent-save':['evoker-identity','shared-output'], 'overchannel-damage':['slots','evoker-identity','shared-output'], 'overchannel-reuse':['overchannel-damage']}.get(k,[])
  tasks.append(dict(key=k,title='Implement Wizard '+k,fixture=fixture,role='implementation',parent='feature-cohort',prerequisites=pre,depends=depends,issue=None,type='Task',size='M'))
 rows={r['level']:r for r in records if r['kind']=='level'}
 grants={1:['core','skills','equipment','later-entry','casting','cantrips','book-identity','book-initial','preparation','slots','multiclass-slots','ritual','arcane-recovery'],2:['scholar'],3:['subclass','savant-initial','savant-new-slots','potent-attack','potent-save'],4:['asi','cantrips'],5:['memorize'],6:['sculpt'],8:['asi'],10:['cantrips','empowered'],12:['asi'],14:['overchannel-damage','overchannel-reuse'],16:['asi'],18:['mastery-select','mastery-cast','mastery-replace'],19:['epic-boon'],20:['signature-select','signature-cast','signature-recovery']}
 for n,r in rows.items():
  deps=['hp','slots','book-learning','preparation']+grants.get(n,[])
  if n>1:deps+=['level-'+str(n-1)]
  if n in [5,7,9,11,13,15,17]:deps+=['savant-new-slots']
  fixture=f'PDF23–26/77–82 Wizard class{n}; exact source row '+json.dumps(r,sort_keys=True)+'. Require adjacent prior Wizard level and qualified whole admitted class configuration; XP/livePB select TOTAL canonical level. Class/total integers1–20, never bool/fraction; class<=total. Witness Wizard1 total5 PB3 slots2 and Wizard2 total6 PB3 slots3, preparation5, learned spells level1. Reject total21, unknown other class, skipped level, duplicate grant or using total level for Wizard HP/features/slots. First-character1 and later-class1 are separate. Grant all required feature alternatives atomically.'
  tasks.append(dict(key='level-'+str(n),title=f'Apply Wizard class level {n}',fixture=fixture,role='implementation',parent='level-cohort',prerequisites=[285],depends=sorted(set(deps)),issue=None,type='Task',size='M'))
 source=load(OUT/'issue-332-residual-source-oracle.json')
 for name,parts in PARTS.items():
  text=source[name]['text'];positions=[text.index('\n'+m)+1 for _,m in parts]
  context=text[:positions[0]]
  for i,(part,marker) in enumerate(parts):
   clause=text[positions[i]:positions[i+1] if i+1<len(parts) else len(text)]
   key='spell-'+slug(name)+'-'+part
   fixture=f"PDF{source[name]['page']} {name}, bounded {part}. Casting context:\n{context}\nExact source contract:\n{clause}\nPositive fixture executes these exact operands, alternatives, timing and outcomes. Independently violate each numeric/target/action/component/resource/timing predicate; assert source-correct outcome and no private disclosure. Other named parts retain separate execution owners."
   tasks.append(dict(key=key,title=f'Resolve {name} {part}',spell=name,part=part,fixture=fixture,source_page=source[name]['page'],role='implementation',parent='new-effect-cohort-'+str((len([t for t in tasks if t.get('spell')]))//50+1),prerequisites=REUSED.get(name,[]),depends=['casting-output'],issue=None,type='Task',size='M'))
 for k,(creature,page,scope) in FORMS.items():
  tasks.append(dict(key=k,title='Resolve Wizard consumed '+creature+' '+k,creature=creature,source_page=page,fixture=f'PDF{page}: '+scope,role='implementation',parent='form-cohort',prerequisites=[],depends=[],issue=None,type='Task',size='M'))
 for t in tasks:
  t['owned_contexts']=[dict(id=c['record']['id'],keys=[k for k,o in c['key_owners'].items() if o==t['key']]) for c in matrix if t['key'] in c['key_owners'].values()]
  t['source_records']=[c['record'] for c in matrix if t['key'] in c['key_owners'].values()]
  t['body']=body(t)+'\nComplete source contexts:\n```json\n'+json.dumps(t['source_records'],indent=2)+'\n```\n'
 return dict(decision=332,status='PENDING root combined finite native publication approval; not final proof',base='e7cd1ba8318ba296edd4fc5b6b84715ca11016f2',source=dict(sha256='8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87',pages=[23,24,25,26,77,78,79,80,81,82,104,105,106],license='CC-BY-4.0'),catalogues=catalogues,records=records,record_matrix=matrix,tasks=tasks,spell_list=load(OUT/'issue-332-wizard-spell-list.json'),accepted_parent_prefix=None,native_publication='PENDING',supplements_status='PENDING ROOT-only serial application')
if __name__=='__main__':
 m=build();(OUT/'issue-332-wizard-scope.json').write_text(json.dumps(m,indent=2)+'\n');print(len(m['tasks']),'candidate tasks;',len(m['record_matrix']),'complete contexts')
