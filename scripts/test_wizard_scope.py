"""Contrary controls mutate actual candidate contracts, not reported digests."""
import copy
import json
from pathlib import Path
import unittest
from verify_wizard_scope import check_plan
ROOT=Path(__file__).resolve().parents[1]

class WizardScopeTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.base=json.loads((ROOT/'docs/investigations/issue-332-wizard-scope.json').read_text())
    def reject(self,change):
        m=copy.deepcopy(self.base);change(m)
        with self.assertRaises(AssertionError):check_plan(m)
    def test_checkpoint(self):check_plan(self.base)
    def test_each106_complete_contexts_rejects_coherent_reroute(self):
        for i in range(106):
            def change(m,i=i):
                c=m['record_matrix'][i]
                c['key_owners']={k:'asi' for k in c['key_owners']}
            # ASI's own contexts already correctly own ASI; use HP there.
            if set(self.base['record_matrix'][i]['key_owners'].values())=={'asi'}:
                def change(m,i=i):m['record_matrix'][i]['key_owners'].update({k:'hp' for k in m['record_matrix'][i]['key_owners']})
            self.reject(change)
    def test_all_actual239_fixtures_reject_body_and_scope_weakening(self):
        for i in range(len(self.base['tasks'])):
            def change(m,i=i):
                old=m['tasks'][i]['fixture'];new='Metadata alone qualifies this behavior.'
                m['tasks'][i]['fixture']=new;m['tasks'][i]['body']=m['tasks'][i]['body'].replace(old,new)
            self.reject(change)
    def test_multiclass_composite_not_hp_classifier(self):
        def change(m):
            c=next(c for c in m['record_matrix'] if c['record']['kind']=='class')
            k='srd-5.2.1:mechanic/wizard/multiclass/hit-die-and-level-1-features';c['key_owners'][k]='hp'
        self.reject(change)
    def test_multiclass_hp_dependency(self):
        self.reject(lambda m:next(t for t in m['tasks'] if t['key']=='later-entry')['depends'].remove('hp'))
    def test_each_grant_alternative_is_required(self):
        for level,owner in [(3,'potent-save'),(6,'sculpt'),(10,'empowered'),(14,'overchannel-reuse'),(18,'mastery-cast'),(20,'signature-recovery')]:
            def change(m,level=level,owner=owner):
                for t in m['tasks']:t['depends']=[d for d in t['depends'] if d!=owner]
            self.reject(change)
    def test_each20_adjacent_transition(self):
        for n in range(2,21):
            self.reject(lambda m,n=n:next(t for t in m['tasks'] if t['key']=='level-'+str(n))['depends'].remove('level-'+str(n-1)))
    def test_every_numeric_spell_operand(self):
        import re
        for i,t in enumerate(self.base['tasks']):
            if not t.get('spell'):continue
            for value in set(re.findall(r'\d+',t['fixture'])):
                def change(m,i=i,value=value):
                    t=m['tasks'][i];old=t['fixture'];new=old.replace(value,str(int(value)+1));t['fixture']=new;t['body']=t['body'].replace(old,new)
                self.reject(change)
    def test_ordinary_feat_has_no_caster_gate(self):
        self.reject(lambda m:next(t for t in m['tasks'] if t['key']=='asi')['prerequisites'].append(428))

    def test_actual_qualification_bodies_and_owner_rosters_cannot_weaken_together(self):
        for i in range(len(self.base['gates'])):
            def change(m,i=i):
                t=m['gates'][i];t['fixture']='Source metadata qualifies execution.';t['body']='Source metadata qualifies execution.';t['depends']=[];t['prerequisites']=[]
            self.reject(change)
        self.reject(lambda m:m['spell_owners']['Acid Arrow'].remove('354'))
    def test_each6_actual_form_semantics(self):
        for k in ['form-riding-horse-statistics','form-riding-horse-hooves','form-commoner-statistics','form-commoner-training','form-commoner-club','form-qualified-snapshot']:
            def change(m,k=k):
                t=next(t for t in m['tasks'] if t['key']==k);old=t['fixture'];t['fixture']='Any metadata form executes';t['body']=t['body'].replace(old,t['fixture'])
            self.reject(change)
    def test_native_reader_repeated_page_and_missing_foreign_nodes(self):
        from unittest.mock import patch
        from types import SimpleNamespace
        from wizard_native_checkpoint import collect
        def page(number,blocked=(),more=True):
            issue=dict(number=number,title='Source output',body='exact',state='OPEN',issueType=dict(name='Task'),parent=None,blockedBy=dict(pageInfo=dict(hasNextPage=False),nodes=[dict(number=x) for x in blocked]),projectItems=dict(pageInfo=dict(hasNextPage=False),nodes=[]))
            return SimpleNamespace(stdout=json.dumps(dict(data=dict(repository=dict(issues=dict(pageInfo=dict(hasNextPage=more,endCursor='same-cursor'),nodes=[issue]))))))
        for pages in [[page(1),page(1)], [page(1),page(2)], [page(1,blocked=[2],more=False)]]:
            with patch('wizard_native_checkpoint.subprocess.run',side_effect=pages):
                with self.assertRaises(AssertionError):collect()
    def test_native_both_dags_reject_closed_dependency_cycles(self):
        from wizard_native_checkpoint import acyclic
        with self.assertRaises(AssertionError):acyclic({'closed-decision':['effect'],'effect':['closed-decision']})

if __name__=='__main__':unittest.main()
