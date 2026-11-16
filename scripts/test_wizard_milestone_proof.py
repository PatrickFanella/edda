"""Milestone byte and stable-map controls; synthetic fixtures never prove live state."""
import copy
import hashlib
import json
import unittest
import test_wizard_publication_proof as fixtures
from wizard_milestone_proof import nodes_digest,validate_milestone,validate_transition,COUNTS
from wizard_publication_proof import APPROVAL_SHA256,RENDERER_SHA256

class MilestoneProofTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        fixtures.PublicationProofTests.setUpClass();cls.inputs=fixtures.PublicationProofTests.inputs
        cls.ids=copy.deepcopy(fixtures.PublicationProofTests.registry);cls.ids['status']='IDS_ALLOCATED'
        cls.verified=copy.deepcopy(fixtures.PublicationProofTests.registry)
        cls.evidence=dict(observed_at='2026-11-16T22:20:00+00:00',graph=fixtures.PublicationProofTests.receipt['graph'])
    def phase(self,registry,phase):
        b=json.dumps(registry).encode();d=dict(milestone=phase,observed_at='2026-11-16T22:10:00+00:00' if phase=='IDS_ALLOCATED' else '2026-11-16T22:21:00+00:00',approval_sha256=APPROVAL_SHA256,renderer_contract_sha256=RENDERER_SHA256,registry_path='/immutable/'+phase+'.json',registry_file_sha256=hashlib.sha256(b).hexdigest(),nodes_map_sha256=nodes_digest(registry,self.inputs),node_count=871,class_counts=COUNTS.copy())
        return d,b
    def transition(self,change=lambda d:None):
        ids,ib=self.phase(self.ids,'IDS_ALLOCATED');ver,vb=self.phase(self.verified,'VERIFIED');eb=json.dumps(self.evidence).encode()
        ver.update(verified_receipt_path='/immutable/readback.json',verified_receipt_sha256=hashlib.sha256(eb).hexdigest(),verified_receipt_observed_at=self.evidence['observed_at'],graph_node_count=1888,all_new_tasks_open=True,native_body_parent_prerequisite_Task_Size_validation=True,both_DAGs=True,all_three_caps=True,old_protections=True)
        change(ver)
        return validate_transition(ids,ib,ver,vb,eb,self.inputs)
    def test_bytehash_changes_with_progress_but_nodes_map_is_stable(self):
        a=nodes_digest(self.ids,self.inputs);b=nodes_digest(self.verified,self.inputs)
        self.assertEqual(a,b)
        self.assertNotEqual(hashlib.sha256(json.dumps(self.ids).encode()).hexdigest(),hashlib.sha256(json.dumps(self.verified).encode()).hexdigest())
        proof=self.transition();self.assertFalse(proof['native_live_proof']);self.assertEqual(proof['runtime_status'],'UNQUALIFIED')
    def test_duplicate_raw_registry_keys_reject(self):
        from wizard_publication_proof import decode_registry
        with self.assertRaisesRegex(AssertionError,'duplicate registry JSON key'):
            decode_registry('{"nodes":{"332:hp":{},"332:hp":{}}}')
    def test_exact_canonical_ensure_ascii_false_digest(self):
        expected=hashlib.sha256(json.dumps(self.ids['nodes'],sort_keys=True,separators=(',',':'),ensure_ascii=False).encode('utf-8')).hexdigest()
        self.assertEqual(nodes_digest(self.ids,self.inputs),expected)
    def test_milestone_bytes_time_counts_and_flags(self):
        for key,value in [('registry_file_sha256','wrong'),('nodes_map_sha256','wrong'),('verified_receipt_sha256','wrong'),('verified_receipt_observed_at','2026-11-16T22:19:00+00:00'),('node_count',True),('class_counts',{'330':305,'331':281,'332':284}),('both_DAGs',False),('old_protections','PASS')]:
            with self.subTest(key=key):
                with self.assertRaises(AssertionError):self.transition(lambda v,key=key,value=value:v.__setitem__(key,value))
    def test_coherent_identity_reassignment_still_rejects_across_phases(self):
        ids,ib=self.phase(self.ids,'IDS_ALLOCATED')
        registry=copy.deepcopy(self.verified);registry['nodes']['332:hp']['id']+=100000000
        ver,vb=self.phase(registry,'VERIFIED')
        eb=json.dumps(self.evidence).encode();ver.update(verified_receipt_path='/immutable/readback.json',verified_receipt_sha256=hashlib.sha256(eb).hexdigest(),verified_receipt_observed_at=self.evidence['observed_at'],graph_node_count=1888,all_new_tasks_open=True,native_body_parent_prerequisite_Task_Size_validation=True,both_DAGs=True,all_three_caps=True,old_protections=True)
        with self.assertRaisesRegex(AssertionError,'reassignment'):validate_transition(ids,ib,ver,vb,eb,self.inputs)
    def test_wrong_registry_snapshot_not_blindly_rehashed(self):
        d,b=self.phase(self.ids,'IDS_ALLOCATED');changed=copy.deepcopy(self.ids);changed['observed_at']='later'
        with self.assertRaisesRegex(AssertionError,'byte hash'):validate_milestone(d,json.dumps(changed).encode(),self.inputs,'IDS_ALLOCATED')

if __name__=='__main__':unittest.main()
