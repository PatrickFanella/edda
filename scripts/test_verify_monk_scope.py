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
