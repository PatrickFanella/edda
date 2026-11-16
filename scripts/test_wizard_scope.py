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

if __name__=='__main__':unittest.main()
