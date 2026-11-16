#!/usr/bin/env python3
"""Bounded pagination and full Fighter ownership/publication corruption probes."""
import json
import unittest
from copy import deepcopy
from unittest.mock import patch
from urllib.parse import parse_qs,urlsplit
import verify_fighter_scope as scope


class Pagination(unittest.TestCase):
    def fixture(self,repeat=False,full=False):
        calls=[]
        def api(path,data=None):
            calls.append(path);self.assertLessEqual(len(calls),3)
            q=parse_qs(urlsplit(path).query);self.assertEqual(q['per_page'],['100']);self.assertEqual(set(q),{'page','per_page'})
            page=int(q['page'][0]);self.assertEqual(q['page'],[str(len(calls))])
            if repeat:page=1
            if full or page==1:return [{'id':(page-1)*100+n} for n in range(1,101)]
            return [{'id':101}]
        with patch.object(scope,'api',api):return scope.pages('test',call_cap=3)

    def test_complete(self):self.assertEqual(len(self.fixture()),101)
    def test_repeat(self):
        with self.assertRaisesRegex(AssertionError,'repeated API page'):self.fixture(repeat=True)
    def test_cap(self):
        with self.assertRaisesRegex(AssertionError,'pagination call cap'):self.fixture(full=True)
    def test_total_call_cap(self):
        with patch.object(scope,'API_CALL_CAP',2),patch.object(scope,'_api_calls',0),patch.object(scope,'github_api',return_value=[]):
            scope.api('one');scope.api('two')
            with self.assertRaisesRegex(AssertionError,'total API call cap'):scope.api('three')

    def test_graph_call_cap(self):
        calls=[]
        def api(path,data=None):
            calls.append(path);self.assertLessEqual(len(calls),100)
            return {'data':{'repository':{'issues':{'nodes':[],'pageInfo':{'hasNextPage':True,'endCursor':str(len(calls))}}}}}
        with patch.object(scope,'api',api):
            with self.assertRaisesRegex(AssertionError,'graph call cap'):scope.graph_read()
        self.assertEqual(len(calls),100)

    def test_repeated_cursor(self):
        calls=[]
        def api(path,data=None):
            calls.append(path);self.assertLessEqual(len(calls),2)
            return {'data':{'repository':{'issues':{'nodes':[],'pageInfo':{'hasNextPage':True,'endCursor':'same'}}}}}
        with patch.object(scope,'api',api):
            with self.assertRaisesRegex(AssertionError,'repeated cursor'):scope.graph_read()


class Receipt(unittest.TestCase):
    def setUp(self):
        self.m=json.loads(scope.MANIFEST.read_text());self.r=json.loads(scope.RECEIPT.read_text())
    def test_full_guards(self):scope.validate(self.m,self.r);scope.probes(self.m,self.r)
    def test_closed_decision_retains_terminal_gate(self):
        next(x for x in self.r['graph'] if x['number']==325)['state']='closed';scope.validate(self.m,self.r)
        next(x for x in self.r['graph'] if x['number']==self.m['qualification_gate']['issue'])['blocked_by'].pop()
        with self.assertRaisesRegex(AssertionError,'terminal gate edge'):scope.validate(self.m,self.r)
    def test_sealed_append_contract(self):
        sealed=deepcopy(self.r)
        for s in self.m['supplements']:
            next(x for x in sealed['graph'] if x['number']==s['issue'])['body']+=s['append']
        scope.validate(self.m,sealed,sealed=True)
        for s in self.m['supplements']:
            n=s['issue'];original=self.m['protected'][str(n)]['body'];owned=s['append']
            candidate=deepcopy(sealed);next(x for x in candidate['graph'] if x['number']==n)['body']+='\nLater pure append.\n';scope.validate(self.m,candidate,sealed=True)
            for value in [original,'changed'+original+owned,original+owned.replace('Fighter','Changed',1),original+owned+owned]:
                candidate=deepcopy(sealed);next(x for x in candidate['graph'] if x['number']==n)['body']=value
                with self.assertRaises(AssertionError):scope.validate(self.m,candidate,sealed=True)
    def test_all_twenty_readable_rows(self):
        records=[x['record'] for x in self.m['records']];owners={x['key']:x['issue'] for x in self.m['issues']}
        doc=(scope.ROOT/'docs/investigations/issue-325-fighter-decision.md').read_text()
        for n in range(1,21):
            row=next(x for x in records if x['kind']=='level' and x['level']==n)
            values=[]
            for slug in ['resource/fighter/second-wind','choice/fighter/weapon-mastery','resource/fighter/action-surge','resource/fighter/indomitable']:
                r=next(x for x in records if x['id'].endswith(slug));values.append(next((v['value'] for v in r['values_by_level'] if v['level']==n),0))
            prefix=f"|{n}|+{row['progression']['proficiency_bonus']}|"+'|'.join(map(str,values))+'|'
            self.assertTrue(any(line.startswith(prefix) and line.endswith(f"|#{owners['level-'+str(n)]}|") for line in doc.splitlines()))


if __name__=='__main__':unittest.main()
