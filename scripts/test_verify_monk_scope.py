#!/usr/bin/env python3
"""Bounded pagination and native-receipt adversaries for decision326."""
import json
import unittest
from copy import deepcopy
from pathlib import Path
from urllib.parse import parse_qs,urlsplit
import verify_monk_scope as v

class MonkProof(unittest.TestCase):
 def test_cursor_repetition_nonadjacent_and_missing(self):
  seen=set()
  for cursor in ['a','b']:self.assertEqual(v.next_cursor(dict(endCursor=cursor),seen),cursor)
  for cursor in ['a','b',None,'']:
   with self.assertRaises(AssertionError):v.next_cursor(dict(endCursor=cursor),seen)
 def test_exact_query_parameters(self):
  calls=[]
  def api(url):
   q=parse_qs(urlsplit(url).query)
   self.assertEqual(q,dict(state=['all'],labels=['a b'],per_page=['100'],page=[str(len(calls)+1)]))
   calls.append(url)
   return [dict(id=i) for i in range(100)] if len(calls)==1 else []
  self.assertEqual(len(list(v.pages(api,'repos/o/r/issues?state=all&labels=a+b'))),100)
  self.assertEqual(len(calls),2)
 def test_page_cap(self):
  calls=[]
  def api(url):
   page=int(parse_qs(urlsplit(url).query)['page'][0]);calls.append(page)
   return [dict(id=page*100+i) for i in range(100)]
  with self.assertRaisesRegex(AssertionError,'cap'):list(v.pages(api,'repos/o/r/issues'))
  self.assertEqual(len(calls),20)
 def test_existing_page_parameter_rejected(self):
  for tail in ['page=1','per_page=2']:
   with self.assertRaises(AssertionError):list(v.pages(lambda _:[], 'repos/o/r/issues?'+tail))
 def test_repeated_page_rejected(self):
  calls=[]
  def api(url):calls.append(url);return [dict(id=i) for i in range(100)]
  with self.assertRaisesRegex(AssertionError,'repeated'):list(v.pages(api,'repos/o/r/issues'))
  self.assertEqual(len(calls),2)
 def test_default_read_only_reader(self):
  with self.assertRaises(AssertionError):v.Reader()('repos/o/r/issues',dict(title='forbidden write'))
 def test_receipt_and_all_corruptions(self):
  m=json.loads(v.MANIFEST.read_text());r=json.loads(v.RECEIPT.read_text())
  v.validate(m,r);v.probes(m,r)
 def test_native_capacity_limits(self):
  m=json.loads(v.MANIFEST.read_text());r=json.loads(v.RECEIPT.read_text())
  for mode,message in [('incoming','incoming blocker capacity'),('outgoing','blocking relation capacity'),('children','native child capacity')]:
   b=deepcopy(r)
   if mode=='incoming':b['graph'].append(dict(number=90000,title='cap probe',body='',state='open',parent=None,blocked_by=[x['number'] for x in b['graph'][:101]]))
   else:
    for i in range(101):b['graph'].append(dict(number=90000+i,title='cap probe',body='',state='open',parent=777 if mode=='children' else None,blocked_by=[777] if mode=='outgoing' else []))
   with self.assertRaisesRegex(AssertionError,message):v.validate(m,b)
 def test_every_level_and_later_entry_requires_adapter(self):
  m=json.loads(v.MANIFEST.read_text());r=json.loads(v.RECEIPT.read_text())
  targets=[x['issue'] for x in m['issues'] if x['key'].startswith('level-') or x['key']=='later']
  for n in targets+[783]:
   b=deepcopy(r);node=next(x for x in b['graph'] if x['number']==n)
   node['blocked_by'].remove(429 if n==783 else 783)
   with self.assertRaises(AssertionError):v.validate(m,b)
 def test_evasion_has_independent_execution_ownership_and_level7_path(self):
  m=json.loads(v.MANIFEST.read_text());r=json.loads(v.RECEIPT.read_text());rows={x['key']:x for x in m['issues']}
  self.assertEqual(set(rows['evasion']['owned_keys']),v.EVASION_KEYS)
  self.assertEqual({x['id'].rsplit('/',1)[-1] for x in rows['evasion']['source_records']},{'evasion','evasion-eligibility'})
  a=deepcopy(m)
  for key in v.EVASION_KEYS:a['source_key_owners'][key]='asi'
  for entry in a['record_matrix']:
   for key in v.EVASION_KEYS:
    if key in entry['key_owners']:entry['key_owners'][key]='asi'
  a['ownership_sha256']=v.sha(json.dumps(a['source_key_owners'],sort_keys=True))
  with self.assertRaisesRegex(AssertionError,'Evasion execution ownership'):v.validate(a,r)
  a,b=deepcopy(m),deepcopy(r);row=next(x for x in a['issues'] if x['key']=='level-7')
  row['depends']=sorted('asi' if k=='evasion' else k for k in row['depends']);row['blocked_by_issues']=sorted((set(row['blocked_by_issues'])-{804})|{799})
  row['body']=row['body'].replace('#804','#799');node=next(x for x in b['graph'] if x['number']==829);node['body']=row['body'];node['blocked_by']=row['blocked_by_issues']
  a['contracts_sha256']=v.sha(json.dumps([{k:x[k] for k in ['key','fixture','depends','prerequisites','owned_keys','source_records']} for x in a['issues']],sort_keys=True))
  with self.assertRaisesRegex(AssertionError,'level7 Evasion contract'):v.validate(a,b)
 def test_advancement_selects_total_level_independently_of_monk_level(self):
  m=json.loads(v.MANIFEST.read_text());r=json.loads(v.RECEIPT.read_text());w=m['deterministic_witnesses']['unequal-class-total-advancement']
  self.assertEqual(w['classes_before'],{'fighter':4,'monk':1});self.assertEqual(w['classes_after'],{'fighter':4,'monk':2})
  self.assertEqual((w['monk_level'],w['total_level'],w['experience_points'],w['proficiency_bonus']),(2,6,14000,3))
  self.assertEqual((w['martial_die_sides'],w['focus_points'],w['movement_feet']),(6,2,10))
  later=m['deterministic_witnesses']['later-class-total-advancement'];self.assertEqual((later['monk_level'],later['total_level'],later['experience_points'],later['proficiency_bonus']),(1,5,6500,3))
  for field,value,message in [('experience_points',300,'total-level XP/PB'),('proficiency_bonus',2,'total-level XP/PB'),('total_level',2,'unequal class/total witness')]:
   a=deepcopy(m);a['deterministic_witnesses']['unequal-class-total-advancement'][field]=value
   with self.assertRaisesRegex(AssertionError,message):v.validate(a,r)
  a=deepcopy(m);a['advancement_selection']['row_index']='Monk-class-level'
  with self.assertRaisesRegex(AssertionError,'class/total advancement selector'):v.validate(a,r)
  a=deepcopy(m);row=next(x for x in a['issues'] if x['key']=='level-2')
  row['source_records'].append(next(x for x in a['catalogues'][v.ADVANCEMENT_TABLE]['records'] if x['level']==2))
  a['contracts_sha256']=v.sha(json.dumps([{k:x[k] for k in ['key','fixture','depends','prerequisites','owned_keys','source_records']} for x in a['issues']],sort_keys=True))
  with self.assertRaisesRegex(AssertionError,'Monk class row only; advancement selected separately'):v.validate(a,r)
 def test_parent_publication_preserves_appends_and_rejects_corruption(self):
  from publish_monk_supplements import plan
  m=json.loads(v.MANIFEST.read_text())
  for s in m['supplements']:
   prefix=m['preserved'][str(s['issue'])]['body']
   self.assertEqual(plan(prefix,s,prefix),prefix+s['append'])
   later=prefix+'\nUnrelated later append.\n'+s['append']
   self.assertEqual(plan(later,s,prefix),later)
   for corrupt in [prefix+s['append']*2,prefix+v.START,prefix+v.END,prefix+s['append'].replace('#326','#325'),'prefix rollback']:
    with self.assertRaises(AssertionError):plan(corrupt,s,prefix)
 def test_read_call_cap(self):
  reader=v.Reader();reader.calls=350
  with self.assertRaisesRegex(AssertionError,'cap'):reader('graphql')
 def test_closed_decision_missing_terminal_still_fails(self):
  m=json.loads(v.MANIFEST.read_text());r=json.loads(v.RECEIPT.read_text())
  graph={x['number']:x for x in r['graph']};graph[326]['state']='closed'
  graph[m['gates'][0]['issue']]['blocked_by'].remove(m['issues'][0]['issue'])
  with self.assertRaises(AssertionError):v.validate(m,r)
 def test_closed_decision_missing_gate_still_fails(self):
  m=json.loads(v.MANIFEST.read_text());r=json.loads(v.RECEIPT.read_text())
  graph={x['number']:x for x in r['graph']};graph[326]['state']='closed'
  graph[337]['blocked_by'].remove(m['gates'][0]['issue'])
  with self.assertRaises(AssertionError):v.validate(m,r)
if __name__=='__main__':unittest.main()
