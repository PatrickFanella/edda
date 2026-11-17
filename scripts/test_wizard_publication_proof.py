"""Full frozen reservation contrary controls; synthetic IDs never prove publication."""
import copy
import json
from pathlib import Path
import unittest
from wizard_publication_proof import (trusted_inputs,validate_candidate,validate_registry,
 expected_native,validate_live,PROJECT,APPROVAL_DIR,APPROVAL_SHA256,RENDERER_SHA256,renderer_contract,_render_body)

class PublicationProofTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.inputs=trusted_inputs();a,f,b,p,n=cls.inputs
        cls.registry={alias:dict(number=2000+i,id=900000+i,node_id='test-node-'+str(i)) for i,alias in enumerate(n)}
        cls.registry=dict(schema_version=1,approval_sha256=APPROVAL_SHA256,renderer_contract_sha256=RENDERER_SHA256,status='VERIFIED',nodes=cls.registry)
        expected,parents=expected_native(p,n,cls.registry)
        old={r['number']:r for r in b['graph']}
        aliases={r['number']:alias for alias,r in cls.registry['nodes'].items()}
        rows=[]
        for number,deps in expected.items():
            if number in old:
                raw=old[number];body=raw['body'];title=raw['title'];md=b['project'].get(str(number),dict(type=None,size=None));state=raw['state']
            else:
                raw=n[aliases[number]];body=_render_body(raw,cls.registry);title=raw['title'];md=dict(type='Task',size='M');state='OPEN'
            identity=cls.registry['nodes'].get(aliases.get(number,''),dict(id=number+100,node_id='old-'+str(number)))
            rows.append(dict(number=number,id=identity['id'],node_id=identity['node_id'],typename='Issue',title=title,body=body,state=state,parent=parents.get(number),blocked_by=deps,type=md['type'],projects=[dict(project=dict(id=PROJECT),fieldValueByName=dict(name=md['size']))]))
        cls.receipt=dict(status='SYNTHETIC TEST FIXTURE ONLY',graph=rows)
    def test_closed_decisions_still_require_all_source_and_terminal_edges(self):
        receipt=copy.deepcopy(self.receipt)
        for row in receipt['graph']:
            if row['number'] in (330,331,332):row['state']='CLOSED'
        validate_live(receipt,self.registry,self.inputs)
        for alias in ['330:shared-aggregate','331:shared-output','332:class-gate']:
            changed=copy.deepcopy(receipt)
            number=self.registry['nodes'][alias]['number']
            row=next(r for r in changed['graph'] if r['number']==number)
            row['blocked_by'].pop()
            with self.assertRaises(AssertionError):validate_live(changed,self.registry,self.inputs)
        changed=copy.deepcopy(receipt)
        old=next(r for r in changed['graph'] if r['number'] not in (330,331,332))
        old['state']='CLOSED' if old['state'].upper()=='OPEN' else 'OPEN'
        with self.assertRaises(AssertionError):validate_live(changed,self.registry,self.inputs)

    def test_frozen_sources_and_approved_template_semantics(self):
        _,f,_,_,_=self.inputs
        actual=json.loads((Path(__file__).resolve().parents[1]/'docs/investigations/issue-332-wizard-scope.json').read_text())
        validate_candidate(actual,f['332'])
    def test_whole_complete_graph_positive_is_unsealed_and_unqualified(self):
        r=validate_live(self.receipt,self.registry,self.inputs)
        self.assertEqual(r['own_nodes'],285);self.assertIsNone(r['accepted_parent_prefix'])
        self.assertTrue(r['runtime_status'].startswith('UNQUALIFIED'))
    def reject(self,change,registry=False):
        data=copy.deepcopy(self.registry if registry else self.receipt);change(data['nodes'] if registry else data)
        with self.assertRaises(AssertionError):
            validate_live(self.receipt if registry else data,data if registry else self.registry,self.inputs)
    def test_bound_source_paths_do_not_qualify_runtime_or_prefix(self):
        from wizard_publication_proof import build_binding
        binding=build_binding(self.registry,self.inputs)
        self.assertEqual(len(binding['own_nodes']),285)
        self.assertEqual(len(binding['approved_execution_paths']),6)
        self.assertEqual(binding['runtime_status'],'UNQUALIFIED')
        self.assertIsNone(binding['accepted_parent_prefix'])
        self.assertEqual(binding['own_nodes']['332:hp']['blocked_by'],[968])
    def test_large_database_id_is_preserved_without_float_conversion(self):
        from wizard_native_checkpoint import database_id
        self.assertEqual(database_id(dict(fullDatabaseId='9876543210987654321')),9876543210987654321)
        self.assertEqual(database_id(dict(databaseId=123)),123)
        for value in [True,1.5,'1.5','',0,'０']:
            self.assertIsNone(database_id(dict(fullDatabaseId=value)))
    def test_renderer_vectors_remain_partial_only(self):
        _,vectors=renderer_contract()
        with self.assertRaises(AssertionError):validate_registry(vectors['example_registry'],self.inputs[-1],self.inputs[2])
    def test_registry_exact_top_fields_and_phase(self):
        for field,value in [('approval_sha256','forged'),('renderer_contract_sha256','forged'),('status','ALLOCATING'),('schema_version',True),('unknown','value')]:
            bad=copy.deepcopy(self.registry);bad[field]=value
            with self.assertRaises(AssertionError):validate_registry(bad,self.inputs[-1],self.inputs[2])
    def test_actual_identity_must_match_real_issue(self):
        n=self.registry['nodes']['332:hp']['number']
        for field,value in [('id',True),('node_id','another-node'),('typename','PullRequest')]:
            self.reject(lambda r,n=n,field=field,value=value:next(x for x in r['graph'] if x['number']==n).__setitem__(field,value))
    def test_complete_registry_and_number_id_node_injectivity(self):
        self.reject(lambda r:r.pop('331:shared-output'),True)
        self.reject(lambda r:r['332:hp'].update(number=r['330:hp']['number']),True)
        self.reject(lambda r:r['332:hp'].update(id=r['330:hp']['id']),True)
        self.reject(lambda r:r['332:hp'].update(node_id=r['330:hp']['node_id']),True)
        self.reject(lambda r:r['332:hp'].update(number=968),True)
        self.reject(lambda r:r['332:hp'].update(number=True),True)
    def test_each_own285_body_and_task_metadata_checked(self):
        own={r['number'] for a,r in self.registry['nodes'].items() if a.startswith('332:')}
        for row in self.receipt['graph']:
            if row['number'] not in own:continue
            n=row['number']
            for field,value in [('body',row['body']+'\nOverride: metadata executes.'),('title','wrong title'),('type','Feature')]:
                self.reject(lambda r,n=n,field=field,value=value:next(x for x in r['graph'] if x['number']==n).__setitem__(field,value))
    def test_changed_qualification_edge_and_closed_decision_bypass(self):
        n=self.registry['nodes']['332:class-gate']['number']
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==n)['blocked_by'].remove(332))
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==337)['blocked_by'].remove(n))
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==183)['blocked_by'].append(self.registry['nodes']['330:world-cohort']['number']))
    def test_all_old_body_title_parent_and_project_protections(self):
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==429).__setitem__('body','replacement'))
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==968).__setitem__('parent',332))
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==285).__setitem__('title','new title'))
    def test_actual_foreign_structure_and_cap_weakening(self):
        n=self.registry['nodes']['331:shared-output']['number']
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==n).__setitem__('blocked_by',[]))
        self.reject(lambda r:r['graph'].append(copy.deepcopy(r['graph'][0])))
        self.reject(lambda r:next(x for x in r['graph'] if x['number']==n).__setitem__('blocked_by',list(range(1,102))))
    def test_frozen_candidate_cannot_change_scopes_with_new_digests(self):
        _,f,_,_,_=self.inputs;m=copy.deepcopy(f['332'])
        t=next(t for t in m['tasks'] if t['key']=='ritual');t['fixture']=t['fixture'].replace('Detect Magic10minutes+Action','Detect Magic11minutes');t['body']=t['body'].replace('Detect Magic10minutes+Action','Detect Magic11minutes')
        with self.assertRaises(AssertionError):validate_candidate(m,f['332'])

if __name__=='__main__':unittest.main()
