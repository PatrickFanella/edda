#!/usr/bin/env python3
"""Bounded API pagination fixtures and complete receipt corruption probes."""
import json
import unittest
from copy import deepcopy
from unittest.mock import patch
from urllib.parse import parse_qs, urlsplit

import verify_cleric_scope as scope


class Pagination(unittest.TestCase):
    def fixture(self, repeat=False, forever=False):
        calls=[]
        def api(path, data=None):
            calls.append(path)
            self.assertLessEqual(len(calls),100, 'fixture call cap')
            q=parse_qs(urlsplit(path).query)
            self.assertEqual(q['per_page'],['100'])
            page=int(q['page'][0])
            if repeat:page=1
            if forever or page==1:
                return [{'id':(page-1)*100+i,'number':(page-1)*100+i} for i in range(1,101)]
            self.assertEqual(page,2)
            return [{'id':101,'number':101}]
        with patch.object(scope,'api',api):return scope.pages('test')

    def test_complete(self):
        self.assertEqual([x['number'] for x in self.fixture()],list(range(1,102)))

    def test_repeat(self):
        with self.assertRaisesRegex(AssertionError,'repeated API page'):self.fixture(repeat=True)

    def test_call_cap(self):
        with self.assertRaisesRegex(AssertionError,'pagination call cap'):self.fixture(forever=True)


class Receipt(unittest.TestCase):
    def test_readable_progression_matches_actual_source_cells(self):
        m=json.loads(scope.MANIFEST.read_text())
        records=[x['record'] for x in m['records']]
        doc=(scope.ROOT/'docs/investigations/issue-323-cleric-decision.md').read_text()
        owners={r['key']:r['issue'] for r in m['issues']}
        def value(suffix,level):
            r=next(x for x in records if x['id'].endswith(suffix))
            return next((z['value'] for z in r['values_by_level'] if z['level']==level),0)
        for level in range(1,21):
            row=next(x for x in records if x['kind']=='level' and x['level']==level)
            values=[value(n,level) for n in ['choice/cleric/cantrips','choice/cleric/prepared-spells','resource/cleric/channel-divinity']]
            slots=[value('resource/cleric/spell-slot-'+str(i),level) for i in range(1,10)]
            text=f"|{level}|+{row['progression']['proficiency_bonus']}|"+'|'.join(map(str,values))+'|'+'/'.join(map(str,slots))+f"|#{owners['level-'+str(level)]}|"
            self.assertIn(text,doc)

    def test_decision_closure_keeps_terminal_gates_required(self):
        m=json.loads(scope.MANIFEST.read_text());r=json.loads(scope.RECEIPT.read_text())
        graph={x['number']:x for x in r['graph']}
        graph[323]['state']='closed'
        scope.validate(m,r)
        gate=m['qualification_gate']['issue']
        graph[gate]['blocked_by'].remove(m['issues'][0]['issue'])
        with self.assertRaisesRegex(AssertionError,'terminal gate edge'):
            scope.validate(m,r)

    def test_sealed_shared_supplements_survive_later_appends(self):
        m=json.loads(scope.MANIFEST.read_text());r=json.loads(scope.RECEIPT.read_text())
        later='\n\n<!-- later-class:start -->\nA later bounded decision.\n<!-- later-class:end -->\n'
        for supplement in m['supplements']:
            n=supplement['issue'];old=m['protected'][str(n)]['body'];append=supplement['append']
            with self.subTest(issue=n,case='later append'):
                candidate=deepcopy(r)
                next(x for x in candidate['graph'] if x['number']==n)['body']=old+append+later
                scope.validate(m,candidate)
            cases={'removed':old,'changed supplement':old+append.replace('Cleric decision','Changed decision',1),
                   'changed original prefix':'Changed prefix'+old+append,'duplicate supplement':old+append+append}
            for case,value in cases.items():
                with self.subTest(issue=n,case=case):
                    candidate=deepcopy(r)
                    next(x for x in candidate['graph'] if x['number']==n)['body']=value
                    with self.assertRaisesRegex(AssertionError,'legacy body'):
                        scope.validate(m,candidate)

    def test_full_source_body_edge_and_separate_cycle_guards(self):
        m=json.loads(scope.MANIFEST.read_text());r=json.loads(scope.RECEIPT.read_text())
        scope.validate(m,r)
        scope.probes(m,r)


if __name__=='__main__':unittest.main()
