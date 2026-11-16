"""Independent literal body/graph controls; synthetic evidence never publishes."""
from copy import deepcopy
import json
import unittest

import sorcerer_native_binding as binding


class NativeBindingTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.approval, cls.baseline = binding.pinned_inputs()
        cls.registry = dict(schema_version=1,approval_sha256=binding.APPROVAL,renderer_contract_sha256=binding.RENDERER,
                            status='VERIFIED',nodes={n['alias']:dict(number=10000+i,id=20000+i,node_id='I_SYNTHETIC_'+str(i)) for i,n in enumerate(cls.approval['approved_nodes'])})
        cls.manifest = json.loads(binding.evidence('330-approved-proposal.json'))
        for row in cls.manifest['tasks']+cls.manifest['shared_split_proposal']+cls.manifest['gates']:
            row['issue'] = cls.registry['nodes']['330:'+row['key']]['number']
        cls.receipt = dict(graph=deepcopy(cls.baseline['graph']),project=deepcopy(cls.baseline['project']))
        graph = {n['number']:n for n in cls.receipt['graph']}
        for edge in cls.approval['allowed_old_root_edges']:
            graph[int(edge['issue'])]['blocked_by'].append(binding.number(edge['blocker'],cls.registry))
            graph[int(edge['issue'])]['blocked_by'].sort()
        for n in cls.approval['approved_nodes']:
            identity=cls.registry['nodes'][n['alias']]
            cls.receipt['graph'].append(dict(**identity,title=n['title'],body=binding.reproduce(n,cls.registry),state='open',
                                             parent=binding.number(n['parent_ref'],cls.registry),blocked_by=sorted(binding.number(d,cls.registry) for d in n['blocked_by_refs'])))
            cls.receipt['project'][str(identity['number'])]=dict(type='Task',size='M')

    def validate(self, receipt):
        return binding.validate_evidence(self.manifest,receipt,self.approval,self.baseline,self.registry)

    def test_literal_renderer_vectors(self):
        vectors=json.loads(binding.evidence('native-renderer-vectors.json'))
        for case in vectors['cases']:
            self.assertEqual(binding.reproduce(case['approved_node'],vectors['example_registry']),case['final_body'])

    def test_complete_independent_frozen_semantic_graph(self):
        self.assertEqual(self.validate(self.receipt)['own_nodes'],305)

    def test_decision_closure_cannot_bypass_any_implementation_gate(self):
        candidate=deepcopy(self.receipt)
        for row in candidate['graph']:
            if row['number'] in [330,331,332]:row['state']='closed'
        self.assertEqual(self.validate(candidate)['global_nodes'],1888)
        aggregate=self.registry['nodes']['330:shared-aggregate']['number']
        row=next(n for n in candidate['graph'] if n['number']==aggregate)
        row['blocked_by'].pop()
        with self.assertRaises(AssertionError):self.validate(candidate)

    def test_unrelated_shared_root_addition_preserves_all_frozen_paths(self):
        candidate=deepcopy(self.receipt)
        candidate['graph'].append(dict(number=99999,title='Unrelated later qualification',body='',state='open',parent=None,blocked_by=[]))
        next(n for n in candidate['graph'] if n['number']==280)['blocked_by'].append(99999)
        self.assertEqual(self.validate(candidate)['global_nodes'],1889)
        row=next(n for n in candidate['graph'] if n['number']==280)
        row['blocked_by'].remove(self.registry['nodes']['330:shared-aggregate']['number'])
        with self.assertRaises(AssertionError):self.validate(candidate)

    def test_every_actual_body_and_protocol_edge_is_mandatory(self):
        # Representative source kinds, nested choices, foreign cohorts and exact
        # terminal obligations; all871 complete bodies are checked on each run.
        aliases=['330:innate-benefits','330:metamagic','330:shared-output','330:shared-aggregate','331:shared-output','332:shared-output']
        available={n['alias'] for n in self.approval['approved_nodes']}
        aliases=[a for a in aliases if a in available]
        for alias in aliases:
            number=self.registry['nodes'][alias]['number']
            for mutation in ['body','parent','state','project','id','prerequisite']:
                candidate=deepcopy(self.receipt);row=next(n for n in candidate['graph'] if n['number']==number)
                if mutation=='body':row['body']=row['body'].replace('unqualified','qualified')
                if mutation=='parent':row['parent']=283
                if mutation=='state':row['state']='closed'
                if mutation=='project':candidate['project'][str(number)]['size']='L'
                if mutation=='id':row['id']+=1
                if mutation=='prerequisite':
                    if row['blocked_by']:row['blocked_by'].pop()
                    else:row['blocked_by']=[285]
                with self.subTest(alias=alias,mutation=mutation),self.assertRaises(AssertionError):self.validate(candidate)

    def test_coherent_receipt_claims_cannot_override_actual_source_body(self):
        candidate=deepcopy(self.receipt)
        n=self.registry['nodes']['330:innate-benefits']['number']
        row=next(r for r in candidate['graph'] if r['number']==n)
        row['body']=row['body'].replace('DC14->15','DC15->16')
        candidate['expected_final_body_sha256']={'330:innate-benefits':binding.digest(row['body'].encode())}
        candidate['status']='VERIFIED'
        with self.assertRaises(AssertionError):self.validate(candidate)

    def test_all_five_root_edges_and_old_guards_remain_required(self):
        for edge in self.approval['allowed_old_root_edges']:
            candidate=deepcopy(self.receipt);row=next(n for n in candidate['graph'] if n['number']==int(edge['issue']))
            row['blocked_by'].remove(binding.number(edge['blocker'],self.registry))
            with self.assertRaises(AssertionError):self.validate(candidate)
        for number in [183,429,779]:
            candidate=deepcopy(self.receipt);row=next(n for n in candidate['graph'] if n['number']==number)
            row['blocked_by'].append(self.registry['nodes']['330:shared-output']['number'])
            with self.assertRaises(AssertionError):self.validate(candidate)


if __name__=='__main__':unittest.main()
