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
    def test_accepted_prefix_and_capacity_seals(self):
        candidate=deepcopy(self.m);candidate['accepted_parent_baseline']['body_sha256']['47']='bad'
        with self.assertRaisesRegex(AssertionError,'accepted seal drift'):scope.validate(candidate,self.r)
        candidate=deepcopy(self.m);candidate['protected']['47']['body']+='Changed prefix'
        with self.assertRaisesRegex(AssertionError,'accepted prefix drift'):scope.validate(candidate,self.r)
        candidate=deepcopy(self.m);candidate['supplements'][0]['append_chars']+=1
        with self.assertRaisesRegex(AssertionError,'capacity metadata drift'):scope.validate(candidate,self.r)

    def test_full_guards(self):scope.validate(self.m,self.r);scope.probes(self.m,self.r)
    def test_closed_decision_retains_terminal_gate(self):
        next(x for x in self.r['graph'] if x['number']==325)['state']='closed';scope.validate(self.m,self.r)
        next(x for x in self.r['graph'] if x['number']==self.m['qualification_gate']['issue'])['blocked_by'].pop()
        with self.assertRaisesRegex(AssertionError,'terminal gate edge'):scope.validate(self.m,self.r)
    def test_sealed_append_contract(self):
        sealed=deepcopy(self.r)
        for s in self.m['supplements']:
            next(x for x in sealed['graph'] if x['number']==s['issue'])['body']=self.m['protected'][str(s['issue'])]['body']+s['append']
        scope.validate(self.m,sealed,sealed=True)
        for s in self.m['supplements']:
            n=s['issue'];original=self.m['protected'][str(n)]['body'];owned=s['append']
            candidate=deepcopy(sealed);next(x for x in candidate['graph'] if x['number']==n)['body']+='\nLater pure append.\n';scope.validate(self.m,candidate,sealed=True)
            for value in [original,'changed'+original+owned,original+owned.replace('Fighter','Changed',1),original+owned+owned,original+owned+'\n<!-- edda-325:start -->',original+owned+'\n<!-- edda-325:end -->']:
                candidate=deepcopy(sealed);next(x for x in candidate['graph'] if x['number']==n)['body']=value
                with self.assertRaises(AssertionError):scope.validate(self.m,candidate,sealed=True)
    def test_future_root_edge_is_not_a_fighter_requirement(self):
        candidate=deepcopy(self.r);n=max(x['number'] for x in candidate['graph'])+1
        candidate['graph'].append({'number':n,'title':'Future unrelated gate','body':'','state':'open','parent':None,'blocked_by':[]})
        next(x for x in candidate['graph'] if x['number']==337)['blocked_by'].append(n)
        scope.validate(self.m,candidate)
        scope.probes(self.m,candidate)

    def test_shared_gates_allow_additions_but_retain_every_required_edge(self):
        for n in [279,280]:
            candidate=deepcopy(self.r);node=next(x for x in candidate['graph'] if x['number']==n)
            extra=next(x['number'] for x in candidate['graph'] if x['number'] not in node['blocked_by'] and x['parent'] is None and not x['blocked_by'] and x['number']!=n)
            node['blocked_by'].append(extra);node['blocked_by'].sort()
            scope.validate(self.m,candidate)
            for required in self.m['protected'][str(n)]['blocked_by']:
                damaged=deepcopy(candidate);next(x for x in damaged['graph'] if x['number']==n)['blocked_by'].remove(required)
                with self.assertRaisesRegex(AssertionError,'legacy edge|shared qualification edge'):scope.validate(self.m,damaged)
            node['body']+='Changed shared contract'
            with self.assertRaisesRegex(AssertionError,f'consumed body #{n}'):scope.validate(self.m,candidate)

    def test_player_unarmed_owner_is_required_by_champion_criticals(self):
        candidate=deepcopy(self.r);node=next(x for x in candidate['graph'] if x['number']==747)
        self.assertIn(779,node['blocked_by']);node['blocked_by'].remove(779)
        with self.assertRaisesRegex(AssertionError,'prerequisite'):scope.validate(self.m,candidate)
        candidate=deepcopy(self.r);next(x for x in candidate['graph'] if x['number']==779)['body']+='Changed shared action'
        with self.assertRaisesRegex(AssertionError,'consumed body #779'):scope.validate(self.m,candidate)

    def test_heroic_inspiration_keeps_unknown_outcome_die_window(self):
        common=self.m['common_acceptance']
        self.assertNotIn('Unknown outcomes never imply success/failure or open reactive windows.',common)
        self.assertIn('Outcome-dependent reactions such as Tactical Mind and Indomitable require a disclosed qualifying outcome.',common)
        self.assertIn('Heroic Inspiration remains available immediately after an owned known die even when the DC and outcome are undisclosed.',common)
        heroic=next(x for x in self.m['issues'] if x['key']=='heroic-inspiration')
        self.assertIn('known owned ability-check die with secret DC and undisclosed outcome',heroic['fixture'])
        candidate=deepcopy(self.m);candidate['common_acceptance']=common.replace('Heroic Inspiration remains available immediately after an owned known die even when the DC and outcome are undisclosed.','')
        with self.assertRaisesRegex(AssertionError,'outcome-independent window contract'):scope.validate(candidate,self.r)
        candidate=deepcopy(self.m);candidate['common_acceptance']+=' Unknown outcomes never imply success/failure or open reactive windows.'
        with self.assertRaisesRegex(AssertionError,'contradictory outcome window contract'):scope.validate(candidate,self.r)

    def test_other_consumed_prerequisites_remain_exact(self):
        for n in [334,779]:
            candidate=deepcopy(self.r);next(x for x in candidate['graph'] if x['number']==n)['blocked_by'].append(777)
            with self.assertRaisesRegex(AssertionError,f'consumed edge #{n}'):scope.validate(self.m,candidate)

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
