"""Bounded controls for checkpoint claims; publication is intentionally unavailable."""
from copy import deepcopy
import json
from pathlib import Path
import unittest
from urllib.parse import parse_qs,urlparse
import verify_rogue_scope as v

class CheckpointTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.m=json.loads(v.PATH.read_text());cls.r=json.loads(v.RECEIPT.read_text()) if v.RECEIPT.exists() else {}
    def test_numeric_semantics_beyond_seal(self):
        for change in ['dc','source','grant','feat','gateway','boolean-level','fractional-level']:
            m=deepcopy(self.m)
            if change=='dc':
                x=next(x for x in m['issues'] if x['key']=='scroll-check');x['fixture']=x['fixture'].replace('DC12/13','DC11/13')
            elif change=='source':m['catalogues']['rogue']['records'][0]['parameters']['srd-5.2.1:mechanic/rogue/hit-die-sides']=10
            elif change=='grant':next(x for x in m['issues'] if x['key']=='level-7')['depends'].remove('evasion')
            elif change=='feat':next(x for x in m['records'] if x['record']['name']=='Evasion')['execution_owners']=['improvement']
            elif change=='gateway':next(x for x in m['issues'] if x['key']=='scroll-item')['fixture']='Light and Detect Magic ONLY'
            else:
                row=next(x for x in m['catalogues']['rogue']['records'] if x['kind']=='level' and x['level']==1)
                row['level']=True if change=='boolean-level' else 1.0
            with self.assertRaises(AssertionError):v.semantic(m)
    def test_complete_consumed_contract_not_excerpt(self):
        import tempfile
        from unittest.mock import patch
        import verify_rogue_consumed_sources as source
        original=json.loads(source.PATH.read_text())
        for change in ['item','inherited','table']:
            bad=deepcopy(original)
            if change=='item':bad['items']['Wand of Magic Detection']['complete_contract']+=' Requires Attunement.'
            elif change=='inherited':bad['casting']['106']=bad['casting']['106'].replace('no action required','Action required')
            else:bad['items']['Spell Scroll']['complete_contract']=bad['items']['Spell Scroll']['complete_contract'].replace('19 +11','18 +10')
            self.assertNotEqual(original,bad)
            with tempfile.TemporaryDirectory() as tmp:
                candidate=Path(tmp)/'source.json';candidate.write_text(json.dumps(bad))
                with patch.object(source,'PATH',candidate):
                    with self.assertRaises(AssertionError):source.verify(Path('/tmp/edda-wave1/SRD_CC_v5.2.1.pdf'),self.m)
    def test_every_record_and_application_guard(self):
        for index in range(107):
            m=deepcopy(self.m);record=m['records'][index]
            if record['key_owners']:
                key=next(iter(record['key_owners']));record['key_owners'][key]=['improvement'] if record['key_owners'][key]!=['improvement'] else ['evasion']
            else:record['record']['name']='coherently changed source'
            with self.assertRaises(AssertionError):v.semantic(m)
        for level in range(1,21):
            m=deepcopy(self.m);next(x for x in m['issues'] if x['key']==f'level-{level}')['depends'].remove('hp')
            with self.assertRaises(AssertionError):v.semantic(m)
    def test_native_metadata_and_owner_binding(self):
        import rogue_native_proof as n
        first=self.m['published_issues'][0]['issue']
        for mode in ['type','size','body','source']:
            r=deepcopy(self.r)
            if mode=='type':r['project'][str(first)]['type']='Feature'
            elif mode=='size':r['project'][str(first)]['size']='L'
            elif mode=='body':next(x for x in r['graph'] if x['number']==first)['body']+=' altered'
            else:next(x for x in r['graph'] if x['number']==779)['body']+=' altered source contract'
            with self.assertRaises(AssertionError):n.validate(self.m,r)
    def test_native_coherent_gateway_weakening(self):
        from unittest.mock import patch
        import rogue_native_proof as n
        m=deepcopy(self.m);r=deepcopy(self.r)
        row=next(x for x in m['published_issues'] if x['key']=='execution')
        row['blocked_by_issues'].remove(67)
        actual=next(x for x in r['graph'] if x['number']==row['issue']);actual['blocked_by'].remove(67)
        # Even if a caller recomputes its own matrix/receipt seal, reconstructed roster rejects it.
        with patch.object(n,'OWNED_SEAL',n.digest(m['published_issues'])):
            with self.assertRaises(AssertionError):n.validate(m,r)
    def test_closure_and_legacy_paths(self):
        import rogue_native_proof as n
        closed=deepcopy(self.r);next(x for x in closed['graph'] if x['number']==329)['state']='closed'
        n.validate(self.m,closed)
        g={x['number']:x for x in closed['graph']}
        for x in self.m['published_issues']:self.assertTrue(n.reaches(g,337,x['issue']))
        # Deleting an unrelated historical required path still fails native proof.
        bad=deepcopy(closed);next(x for x in bad['graph'] if x['number']==783)['blocked_by'].remove(429)
        with self.assertRaises(AssertionError):n.validate(self.m,bad)
    def test_final_original_rollback_and_later_appends(self):
        import rogue_native_proof as n
        sealed=deepcopy(self.r)
        for x in sealed['graph']:
            if str(x['number']) in n.ACCEPTED_PREFIXES:
                prefix=self.m['accepted_parent_baseline']['body_seals'][str(x['number'])]['body']
                section=Path(f"docs/investigations/issue-329-parent-{x['number']}-PENDING.md").read_text()
                x['body']=prefix+section
        n.validate(self.m,sealed,sealed=True)
        rollback=deepcopy(sealed)
        next(x for x in rollback['graph'] if x['number']==47)['body']=self.m['accepted_parent_baseline']['body_seals']['47']['body']
        with self.assertRaises(AssertionError):n.validate(self.m,rollback,sealed=True)
        for x in sealed['graph']:
            if x['number'] in [47,54,283,337]:x['body']+='\nlegitimate later parent append'
        n.validate(self.m,sealed,sealed=True)
        wrong=deepcopy(self.m);wrong['accepted_parent_baseline']['body_seals']['47']['body']+='coherent changed prefix'
        with self.assertRaises(AssertionError):n.validate(wrong,sealed,sealed=True)
    def test_marker_counts_and_append(self):
        prefix='accepted prefix';section=Path('docs/investigations/issue-329-parent-47-PENDING.md').read_text()
        v.parent_append(prefix,section,prefix+section+'\nlegitimate later parent append')
        for actual in [prefix,prefix+section+section,prefix+section+'<!-- edda-329:start --> altered duplicate <!-- edda-329:end -->',prefix+section+'<!-- edda-329 : START --> altered duplicate <!-- edda-329 : END -->','changed'+section]:
            with self.assertRaises(AssertionError):v.parent_append(prefix,section,actual)
    def test_pagination_100_101(self):
        for count,expected_calls in [(100,2),(101,2),(99,1)]:
            calls=[]
            def fake(url):
                calls.append(url);assert len(calls)<=3
                q=parse_qs(urlparse(url).query,strict_parsing=True)
                assert set(q)=={'page','per_page'} and q['per_page']==['100']
                page=int(q['page'][0]);assert 1<=page<=3
                return [dict(id=n) for n in range((page-1)*100,min(count,page*100))]
            self.assertEqual(len(v.pages('exact/path',fake)),count);self.assertEqual(len(calls),expected_calls)
    def test_production_graph_pagination_100_101(self):
        import rogue_native_proof as n
        for count in [100,101]:
            calls=[]
            def fake(path,data=None):
                calls.append(path);assert len(calls)<=6,'fake call cap'
                if path=='graphql':
                    assert 'issues(first:100,after:$cursor)' in data['query']
                    page=0 if data['variables']['cursor'] is None else 1
                    selected=list(range(page*100,min(count+1,(page+1)*100)))
                    nodes=[]
                    for number in selected:
                        deps=[dict(number=i) for i in range(1,min(count,100)+1)] if number==0 else []
                        nodes.append(dict(number=number,title='fixture',body='',state='OPEN',parent=None,blockedBy=dict(nodes=deps,pageInfo=dict(hasNextPage=number==0 and count>100))))
                    return dict(data=dict(repository=dict(issues=dict(nodes=nodes,pageInfo=dict(hasNextPage=page==0,endCursor='second')))))
                u=urlparse(path);self.assertEqual(u.path,'repos/subculture-collective/edda/issues/0/dependencies/blocked_by')
                q=parse_qs(u.query,strict_parsing=True);self.assertEqual(set(q),{'per_page','page'});self.assertEqual(q['per_page'],['100'])
                page=int(q['page'][0]);self.assertIn(page,[1,2])
                return [dict(id=i,number=i) for i in range(1+(page-1)*100,min(count+1,1+page*100))]
            result=n.read_graph(fake,limit=3);self.assertEqual(len(result),count+1)
            self.assertEqual(len(calls),2 if count==100 else 4)
            if count==100:n.caps(result)
            else:
                with self.assertRaises(AssertionError):n.caps(result)
    def test_repeat_and_bound(self):
        calls=[]
        def fake(url):
            calls.append(url);assert len(calls)<=3
            return [dict(id=n) for n in range(100)]
        with self.assertRaisesRegex(AssertionError,'repeated'):v.pages('path',fake)
        calls=[]
        def endless(url):
            calls.append(url);assert len(calls)<=3
            page=int(parse_qs(urlparse(url).query)['page'][0]);return [dict(id=page*100+n) for n in range(100)]
        with self.assertRaisesRegex(AssertionError,'call cap'):v.pages('path',endless)
        self.assertEqual(len(calls),3)
    def test_caps_each_direction(self):
        for count in [100,101]:
            base=[dict(number=i,parent=None,blocked_by=[],state='open') for i in range(count+1)]
            for mode in ['incoming','outgoing','children']:
                g=deepcopy(base)
                if mode=='incoming':g[0]['blocked_by']=list(range(1,count+1))
                elif mode=='outgoing':
                    for x in g[1:]:x['blocked_by']=[0]
                else:
                    for x in g[1:]:x['parent']=0
                if count==100:v.caps(g)
                else:
                    with self.assertRaises(AssertionError):v.caps(g)
    def test_both_cycles(self):
        g=[dict(number=1,parent=None,blocked_by=[2]),dict(number=2,parent=None,blocked_by=[1])]
        with self.assertRaises(AssertionError):v.caps(g)
        for x in g:x['blocked_by']=[]
        g[0]['parent']=2;g[1]['parent']=1
        with self.assertRaises(AssertionError):v.caps(g)

if __name__=='__main__':unittest.main()
