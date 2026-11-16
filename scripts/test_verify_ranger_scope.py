"""Independent contrary ownership, source, operand and pagination controls."""
from copy import deepcopy
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import verify_ranger_scope as v
from verify_druid_scope import pages, Reader

class Controls(unittest.TestCase):
 def setUp(self):self.m=json.loads(v.MANIFEST.read_text())
 def reject(self,change):
  m=deepcopy(self.m);change(m)
  with self.assertRaises((AssertionError,KeyError)):v.validate_plan(m)
 def test_canonical(self):v.validate_plan(self.m)
 def test_rehashed_feature_owner_and_context(self):
  # Coherent reassignment also changes source context, not just a checksum.
  for slug in ['extra-attack','roving','feral-senses','precise-hunter','relentless-hunter','foe-slayer','natures-veil','hunters-lore','superior-hunters-defense']:
   def mutate(m):
    c=next(c for c in m['record_matrix'] if c['record']['kind']=='feature' and c['id'].endswith('/'+slug))
    c['key_owners']={k:'ability-score-improvement' for k in c['key_owners']}
    old=next(x for x in m['tasks'] if x['key']==slug)
    asi=next(x for x in m['tasks'] if x['key']=='ability-score-improvement')
    asi['source_records']+=old['source_records'];old['source_records']=[]
    m['record_owners'][c['id']]='ability-score-improvement'
    m['ownership_sha256']=v.sha(json.dumps(m['record_owners'],sort_keys=True))
   self.reject(mutate)
 def test_every_feature_native_path(self):
  for n in range(1,21):
   def mutate(m):
    row=next(x for x in m['tasks'] if x['key']=='level-'+str(n));row['depends']=[]
   self.reject(mutate)
 def test_later_admission_and_hp_mandatory(self):
  for dep in ['later-entry','hp']:
   self.reject(lambda m:next(x for x in m['tasks'] if x['key']=='level-2')['depends'].remove(dep))
 def test_total_operand(self):
  self.reject(lambda m:next(x for x in m['tasks'] if x['key']=='level-2').update(fixture='Class2 XP300 PB2 from sourceclassrow'))
 def test_unequal_class_total_controls(self):
  for field,value in [('ranger_row',6),('total_row',2),('xp',300),('pb',2),('hp_mode','first10plusCON')]:
   self.reject(lambda m:m['operand_witnesses'][1].update({field:value}))
 def test_integer_admitted_levels(self):
  for value in [True,1.0,1.5,0,21]:
   self.reject(lambda m:m['operand_witnesses'][0]['after'].update(ranger=value))
  for field in ['ranger_row','total_row']:
   self.reject(lambda m:m['operand_witnesses'][0].update({field:True}))
 def test_source_numeric(self):
  def mutate(m):
   p=next(p for p in m['catalogues'] if p.endswith('character-ranger.json'))
   m['catalogues'][p]['records'][0]['parameters']['srd-5.2.1:mechanic/ranger/hit-die-sides']=12
  self.reject(mutate)
 def test_dependency_cycle(self):
  self.reject(lambda m:next(x for x in m['gates'] if x['key']=='execution-dependencies')['depends'].append('level-2'))
 def test_unknown_owner(self):
  self.reject(lambda m:m['record_matrix'][0]['key_owners'].update({'invented':'invented'}))
 def test_live_refuses_plan(self):
  for x in self.m['tasks']+self.m['gates']:x['issue']=None
  with self.assertRaises(AssertionError):v.validate_live(self.m,{'graph':[]})
 def test_wrong_pdf(self):
  with tempfile.TemporaryDirectory() as tmp:
   p=Path(tmp)/'wrong.pdf';p.write_bytes(b'wrong source')
   with self.assertRaisesRegex(AssertionError,'wrong PDF'):v.source_check(self.m,p)
 def test_exact_rest_query_pages(self):
  calls=[]
  def api(path):
   calls.append(path)
   return [{'id':i} for i in range(100)] if len(calls)==1 else [{'id':100}]
  self.assertEqual(len(list(pages(api,'repos/o/r/issues?state=all'))),101)
  self.assertEqual(calls,['repos/o/r/issues?state=all&per_page=100&page=1','repos/o/r/issues?state=all&per_page=100&page=2'])
 def test_repeated_rest_page(self):
  with self.assertRaisesRegex(AssertionError,'repeated API page'):list(pages(lambda _: [{'id':i} for i in range(100)],'repos/o/r/issues'))
 def test_cursor_missing_repeat(self):
  for cursor,seen in [(None,set()),('same',{'same'})]:
   with self.assertRaises(AssertionError):v.next_cursor({'endCursor':cursor},seen)
 def test_parent_prefix_and_later_appends(self):
  prefix='accepted Paladin prefix';section='\n<!-- edda-328:start -->\nsection\n<!-- edda-328:end -->\n'
  item=dict(append=section,previous_sha256=v.sha(prefix));h=v.sha(prefix)
  self.assertEqual(v.supplement_plan(prefix,item,prefix,h),prefix+section)
  self.assertEqual(v.supplement_plan(prefix+section+'later append',item,prefix,h),prefix+section+'later append')
  for current in [prefix[:-1],prefix+section+section,prefix+section.replace('section','altered'),prefix+'<!-- edda-328:start -->']:
   with self.assertRaises(AssertionError):v.supplement_plan(current,item,prefix,h)
 def test_native_synthetic_positive_and_every_owned_edge(self):
  from ranger_scope_test_support import synthetic_receipt
  m,r=synthetic_receipt(self.m);v.validate_live(m,r)
  # Closing the decision never substitutes for any implementation/gate edge.
  next(x for x in r['graph'] if x['number']==328)['state']='closed'
  v.validate_live(m,r)
  # Every owned edge is independently required even after decision closure.
  for x in m['tasks']+m['gates']:
   for edge in set(x['prerequisites'])|{v.rows(m)[d]['issue'] for d in x['depends']}:
    broken=deepcopy(r);node=next(a for a in broken['graph'] if a['number']==x['issue']);node['blocked_by'].remove(edge)
    with self.assertRaises(AssertionError):v.validate_live(m,broken)
 def test_native_metadata_body_parent_and_root(self):
  from ranger_scope_test_support import synthetic_receipt
  m,r=synthetic_receipt(self.m);owned=m['tasks'][0]['issue'];g=lambda receipt,n:next(x for x in receipt['graph'] if x['number']==n)
  changes=[lambda q:g(q,owned).update(body='altered'),lambda q:g(q,owned).update(parent=328),lambda q:q['project'][str(owned)].update(size='L'),lambda q:q['project'][str(owned)].update(type='Feature'),lambda q:g(q,429).update(body='altered'),lambda q:g(q,279)['blocked_by'].remove(v.rows(m)['shared-gate']['issue'])]
  for change in changes:
   broken=deepcopy(r);change(broken)
   with self.assertRaises(AssertionError):v.validate_live(m,broken)
 def test_native_caps_and_both_cycles(self):
  from ranger_scope_test_support import synthetic_receipt
  m,r=synthetic_receipt(self.m)
  for mode in ['incoming','outgoing','children','dependency-cycle','parent-cycle']:
   q=deepcopy(r);node=next(x for x in q['graph'] if x['number']==m['tasks'][0]['issue'])
   if mode=='incoming':
    node['blocked_by']=list(range(800000,800101));q['graph'] += [dict(number=n,parent=None,blocked_by=[]) for n in node['blocked_by']]
   if mode=='outgoing':q['graph'].append(dict(number=800000,parent=None,blocked_by=[429]))
   if mode=='children':q['graph'] += [dict(number=810000+i,parent=328,blocked_by=[]) for i in range(101)]
   if mode=='dependency-cycle':node['blocked_by'].append(node['number'])
   if mode=='parent-cycle':node['parent']=node['number']
   expected={'incoming':'incoming cap','outgoing':'outgoing cap','children':'child cap','dependency-cycle':'cycle','parent-cycle':'cycle'}[mode]
   with self.assertRaisesRegex(AssertionError,expected):v.validate_live(m,q)
 def test_final_release_seal_and_required_supplements(self):
  from ranger_scope_test_support import synthetic_receipt
  m,r=synthetic_receipt(self.m);m['accepted_parent_prefix']={};r['parent_body_seals']={};pins={}
  for item in m['supplements']:
   n=str(item['issue']);prefix='accepted root '+n;h=v.sha(prefix);pins[n]=h
   item['previous_sha256']=h;m['accepted_parent_prefix'][n]=dict(prefix=prefix,sha256=h)
   node=next((x for x in r['graph'] if x['number']==int(n)),None)
   if node is None:node=dict(number=int(n));r['graph'].append(node)
   node['body']=prefix;r['parent_body_seals'][n]=dict(prefix=prefix,sha256=h)
  with patch.object(v,'ACCEPTED_PREFIX_SHA256',pins):
   v.validate_seal(m,r)
   with self.assertRaisesRegex(AssertionError,'supplement pending'):v.validate_seal(m,r,True)
   q=deepcopy(r);next(x for x in q['graph'] if x['number']==47)['body']='rollback'
   with self.assertRaises(AssertionError):v.validate_seal(m,q)
   forged=deepcopy(m);forged['accepted_parent_prefix']['47']['prefix']='forged';forged['accepted_parent_prefix']['47']['sha256']=v.sha('forged')
   with self.assertRaisesRegex(AssertionError,'immutable'):v.validate_seal(forged,r)
   for item in m['supplements']:next(x for x in r['graph'] if x['number']==item['issue'])['body']+=item['append']+'later pure append'
   v.validate_seal(m,r,True)
 def test_api_readonly_and_callcap(self):
  r=Reader();r.calls=350
  with self.assertRaisesRegex(AssertionError,'call cap'):r('graphql',{'query':'{}'})
  with self.assertRaisesRegex(AssertionError,'read-only'):Reader()('graphql',{'query':'mutation{bad}'})
if __name__=='__main__':unittest.main()
