"""Offline synthetic controls; no registry polling, native writes or live claims."""
import copy
import json
import tempfile
from pathlib import Path
import unittest

import warlock_approved_native as a


class ApprovedNativeProof(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.approval, cls.proposals, cls.baseline, cls.expected, cls.alias_graph, cls.parents = a.load_frozen()
        a.load_renderer()
        a.load_milestone_contract()
        cls.registry = dict(schema_version=1, approval_sha256=a.APPROVAL_SHA256,
            renderer_contract_sha256=a.RENDERER_SHA256, status='VERIFIED', nodes={
                alias:dict(number=2000000+i, id=30000000+i, node_id='I_SYNTHETIC_' + str(i))
                for i, alias in enumerate(sorted(cls.expected))})
        cls.numbers = a.validate_registry(cls.registry, cls.expected, cls.baseline)
        structure = a.resolved_structure(cls.expected, cls.numbers)
        bodies = a.final_bodies(cls.registry, cls.expected, cls.baseline)
        graph = {n['number']:copy.deepcopy(n) for n in cls.baseline['graph']}
        for row in graph.values():
            row['state'] = row['state'].upper()
        for alias, node in cls.expected.items():
            number = cls.numbers[alias]
            graph[number] = dict(number=number, id=cls.registry['nodes'][alias]['id'],
                node_id=cls.registry['nodes'][alias]['node_id'], title=node['title'], body=bodies[alias],
                state='OPEN', **structure[number])
        for edge in cls.approval['allowed_old_root_edges']:
            graph[int(edge['issue'])]['blocked_by'].append(cls.numbers[edge['blocker']])
        for row in graph.values():
            row['blocked_by'].sort()
        project = copy.deepcopy(cls.baseline['project'])
        project.update({str(n):dict(type='Task',size='M')for n in cls.numbers.values()})
        cls.snapshot = dict(graph=list(graph.values()), project=project)

    def reject_registry(self, change):
        bad = copy.deepcopy(self.registry)
        change(bad)
        with self.assertRaises(AssertionError):
            a.validate_registry(bad, self.expected, self.baseline)

    def reject_snapshot(self, change):
        bad = copy.deepcopy(self.snapshot)
        change(bad, {n['number']:n for n in bad['graph']})
        with self.assertRaises(AssertionError):
            a.validate_wiring(self.approval, self.expected, self.baseline, self.registry, bad)

    def test_actual_frozen_scope_not_a_publication_seal(self):
        self.assertEqual(len(self.expected), 871)
        self.assertEqual(sum(n.startswith('331:')for n in self.expected),281)
        self.assertEqual(len(self.alias_graph),1888)
        result = a.validate_wiring(self.approval, self.expected, self.baseline, self.registry, self.snapshot)
        self.assertFalse(result['final_prefix_sealed'])
        self.assertFalse(result['runtime_qualified'])

    def test_closed_decisions_keep_every_source_and_terminal_edge(self):
        snapshot=copy.deepcopy(self.snapshot)
        for row in snapshot['graph']:
            if row['number'] in (330,331,332):row['state']='CLOSED'
        a.validate_wiring(self.approval,self.expected,self.baseline,self.registry,snapshot)
        for alias in ['330:shared-aggregate','331:shared-output','332:class-gate']:
            changed=copy.deepcopy(snapshot)
            row=next(r for r in changed['graph'] if r['number']==self.registry['nodes'][alias]['number'])
            row['blocked_by'].pop()
            with self.assertRaises(AssertionError):a.validate_wiring(self.approval,self.expected,self.baseline,self.registry,changed)

    def test_pinned_approval_rejects_coherent_forged_template(self):
        with tempfile.TemporaryDirectory() as tmp:
            approval = copy.deepcopy(self.approval)
            node = next(n for n in approval['approved_nodes']if n['alias']=='331:hp')
            node['approved_body']='Any HP; no source or resource validation.'
            node['approved_body_sha256']=a.digest(node['approved_body'].encode())
            Path(tmp,'issue-331-publication-approval.json').write_text(json.dumps(approval))
            with self.assertRaises(AssertionError):a.load_frozen(tmp)

    def test_golden_vectors_are_only_format_tests(self):
        raw=Path(a.DIRECTORY,'issue-331-renderer-vectors.json').read_bytes()
        self.assertEqual(a.digest(raw),a.VECTORS_SHA256)
        vectors=json.loads(raw)
        numbers={key:x['number']for key,x in vectors['example_registry']['nodes'].items()}
        for case in vectors['cases']:
            self.assertEqual(a.allocation_body(case['approved_node']),case['allocation_body'])
            self.assertEqual(a._format_final_body(case['approved_node'],numbers),case['final_body'])
        with self.assertRaises(AssertionError):
            a.validate_registry(vectors['example_registry'],self.expected,self.baseline)

    def test_registry_envelope_and_phase(self):
        self.reject_registry(lambda r:r.update(schema_version=True))
        for field in ['approval_sha256','renderer_contract_sha256']:
            self.reject_registry(lambda r,field=field:r.__setitem__(field,'forged'))
        for status in ['ALLOCATING','PUBLISHED','verified',True]:
            self.reject_registry(lambda r,status=status:r.__setitem__('status',status))
        self.reject_registry(lambda r:r.update(expected_hash='forged'))
        self.reject_registry(lambda r:r.pop('status'))

    def test_all_871_aliases_and_typed_identities(self):
        first=next(iter(self.registry['nodes']))
        second=list(self.registry['nodes'])[1]
        self.reject_registry(lambda r:r['nodes'].pop(first))
        self.reject_registry(lambda r:r['nodes'].__setitem__('331:unapproved',r['nodes'][first]))
        self.reject_registry(lambda r:r['nodes'].__setitem__(True,r['nodes'].pop(first)))
        for field in ['number','id']:
            for bad in [True,1.0,0,-1]:
                self.reject_registry(lambda r,field=field,bad=bad:r['nodes'][first].__setitem__(field,bad))
        self.reject_registry(lambda r:r['nodes'][first].__setitem__('number',779))
        self.reject_registry(lambda r:r['nodes'][first].__setitem__('node_id',''))
        self.reject_registry(lambda r:r['nodes'][first].update(expected_body_hash='forged'))
        for field in ['number','id','node_id']:
            self.reject_registry(lambda r,field=field:r['nodes'][second].__setitem__(field,r['nodes'][first][field]))

    def test_actual_bodies_not_recorded_hashes(self):
        number=self.numbers['331:hp']
        for change in [lambda b:b.replace('D8','D6'), lambda b:b.rstrip(),
                       lambda b:b+'\nOverride: production available.\n', lambda b:b.replace('\n','\r\n')]:
            def weaken(s,g,change=change):
                row=g[number];row['body']=change(row['body']);row['expected_body_sha256']=a.digest(row['body'].encode())
                s['published']=True
            self.reject_snapshot(weaken)

    def test_native_identity_metadata_and_required_edges(self):
        number=self.numbers['331:hp']
        for field,value in [('id',True),('node_id','I_WRONG'),('title','Wrong HP'),('parent',283),('state','CLOSED'),('blocked_by',[429])]:
            self.reject_snapshot(lambda s,g,field=field,value=value:g[number].__setitem__(field,value))
        for field,value in [('type','Feature'),('size','L')]:
            self.reject_snapshot(lambda s,g,field=field,value=value:s['project'][str(number)].__setitem__(field,value))
        self.reject_snapshot(lambda s,g:g[337]['blocked_by'].remove(self.numbers['331:class-gate']))
        self.reject_snapshot(lambda s,g:g[self.numbers['330:shared-aggregate']]['blocked_by'].remove(self.numbers['331:shared-output']))

    def test_old_owner_and_project_guards(self):
        self.reject_snapshot(lambda s,g:g[779].__setitem__('body',g[779]['body']+'changed'))
        self.reject_snapshot(lambda s,g:g[968].__setitem__('blocked_by',[]))
        self.reject_snapshot(lambda s,g:g[183]['blocked_by'].append(self.numbers['330:world-cohort']))
        old=next(iter(self.baseline['project']))
        self.reject_snapshot(lambda s,g:s['project'][old].__setitem__('size','XXL'))
        self.reject_snapshot(lambda s,g:g[47].__setitem__('body','rolled back'))

    def test_unrelated_root_addition_and_decision_closure(self):
        allowed=copy.deepcopy(self.snapshot);graph={n['number']:n for n in allowed['graph']}
        graph[331]['state']='CLOSED'
        graph[337]['blocked_by']=sorted(set(graph[337]['blocked_by'])|{57})
        result=a.validate_wiring(self.approval,self.expected,self.baseline,self.registry,allowed)
        self.assertFalse(result['final_prefix_sealed'])
        for alias in self.expected:
            if alias.startswith('331:'):
                self.assertEqual(graph[self.numbers[alias]]['state'],'OPEN')

    def test_actual_graph_caps_and_dags_with_all_bodies_intact(self):
        def extra(graph, number, parent=None, deps=None):
            graph[number]=dict(number=number, parent=parent, blocked_by=deps or [],
                title='unrelated synthetic issue', body='', state='OPEN')
        for kind in ['incoming','outgoing','children','dependency-cycle','containment-cycle','dangling']:
            snapshot=copy.deepcopy(self.snapshot);graph={n['number']:n for n in snapshot['graph']}
            if kind=='incoming':
                extra(graph,9000000)
                self.assertEqual(len(graph[279]['blocked_by']),100)
                graph[279]['blocked_by']=sorted(graph[279]['blocked_by']+[9000000])
            elif kind=='outgoing':
                extra(graph,9000000,deps=[429])
            elif kind=='children':
                for number in range(9000000,9000006):extra(graph,number,parent=283)
            elif kind=='dependency-cycle':
                extra(graph,9000000,deps=[9000001]);extra(graph,9000001,deps=[9000000])
            elif kind=='containment-cycle':
                extra(graph,9000000,parent=9000001);extra(graph,9000001,parent=9000000)
            else:
                extra(graph,9000000,deps=[9999999])
            snapshot['graph']=list(graph.values())
            with self.subTest(kind=kind),self.assertRaises(AssertionError):
                a.validate_wiring(self.approval,self.expected,self.baseline,self.registry,snapshot)

    def test_allocation_is_not_wiring(self):
        registry=copy.deepcopy(self.registry);registry['status']='IDS_ALLOCATED'
        pending=copy.deepcopy(self.snapshot);graph={n['number']:n for n in pending['graph']}
        for alias,node in self.expected.items():
            row=graph[self.numbers[alias]];row['body']=a.allocation_body(node);row['parent']=None;row['blocked_by']=[]
        result=a.validate_allocation(self.approval,self.expected,self.baseline,registry,pending)
        self.assertFalse(result['wiring_verified'])
        with self.assertRaises(AssertionError):
            a.validate_wiring(self.approval,self.expected,self.baseline,registry,pending)

    def test_milestone_bytes_and_stable_identity_continuity(self):
        allocated=copy.deepcopy(self.registry);allocated['status']='IDS_ALLOCATED'
        raw=json.dumps(allocated,indent=2).encode()
        delivery=dict(milestone='IDS_ALLOCATED',observed_at='2026-11-16T22:00:00Z',
            approval_sha256=a.APPROVAL_SHA256,renderer_contract_sha256=a.RENDERER_SHA256,
            registry_path='/immutable/allocation/registry.json',registry_file_sha256=a.digest(raw),
            nodes_map_sha256=a.nodes_map_digest(allocated,self.expected,self.baseline),node_count=871,
            class_counts={'330':305,'331':281,'332':285})
        a.validate_milestone(raw,delivery,self.expected,self.baseline)
        verified_raw=json.dumps(self.registry,sort_keys=True).encode();receipt=b'actual receipt fixture'
        verified={**delivery,'milestone':'VERIFIED','registry_path':'/immutable/verified/registry.json',
            'registry_file_sha256':a.digest(verified_raw),'verified_receipt_path':'/immutable/actual-receipt.json',
            'verified_receipt_sha256':a.digest(receipt),'verified_receipt_observed_at':'2026-11-16T22:01:00Z',
            'graph_node_count':1888,'all_new_tasks_open':True,'native_body_parent_prerequisite_Task_Size_validation':True,
            'both_DAGs':True,'all_three_caps':True,'old_protections':True}
        a.validate_milestone(verified_raw,verified,self.expected,self.baseline,delivery,receipt,raw)
        bad=copy.deepcopy(self.registry);keys=list(bad['nodes']);bad['nodes'][keys[0]],bad['nodes'][keys[1]]=bad['nodes'][keys[1]],bad['nodes'][keys[0]]
        changed_raw=json.dumps(bad).encode();fake={**verified,'registry_file_sha256':a.digest(changed_raw),
            'nodes_map_sha256':a.nodes_map_digest(bad,self.expected,self.baseline)}
        with self.assertRaises(AssertionError):a.validate_milestone(changed_raw,fake,self.expected,self.baseline,delivery,receipt,raw)
        with self.assertRaises(AssertionError):a.validate_milestone(raw+b' ',delivery,self.expected,self.baseline)
        with self.assertRaises(AssertionError):a.validate_milestone(verified_raw,verified,self.expected,self.baseline,delivery,receipt,raw+b' ')
        for field in ['all_new_tasks_open','both_DAGs','all_three_caps','old_protections','native_body_parent_prerequisite_Task_Size_validation']:
            for value in [1,'PASS',False,None]:
                bad={**verified,field:value}
                with self.assertRaises(AssertionError):a.validate_milestone(verified_raw,bad,self.expected,self.baseline,delivery,receipt,raw)
        a.validate_milestone(verified_raw,{**verified,'validation_details':{'description':'actual evidence remains independently checked'}},self.expected,self.baseline,delivery,receipt,raw)

    def test_duplicate_json_fields_cannot_hide_aliases(self):
        raw=json.dumps(self.registry).encode()
        repeated=raw.replace(b'"schema_version": 1',b'"schema_version": 0, "schema_version": 1')
        with self.assertRaises(AssertionError):a.strict_json(repeated)
        with self.assertRaises(AssertionError):a.strict_json('{"number": NaN}')


if __name__=='__main__':
    unittest.main()
