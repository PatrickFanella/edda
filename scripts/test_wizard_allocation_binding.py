"""Offline regression checks against the actual ROOT-delivered allocation bytes."""
import copy
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import unittest

from wizard_milestone_proof import nodes_digest, validate_milestone
from wizard_publication_proof import (build_binding, trusted_inputs, validate_candidate,
                                      validate_live, renderer_contract)

ROOT = Path(__file__).resolve().parents[1]
CHECKPOINT = ROOT / 'docs/investigations/issue-332-allocation-binding'
REGISTRY_SHA = '73d9755cd7181155544f5d79a944bdaf653028425a75b5131b9adef1faf76392'
DELIVERY_SHA = '4ada8b216c27369df0cca9304d4ec5f4f556035362ee0c3d26ebef8d21fdf9a6'
MAP_SHA = '866246ec119bba3198ee6f6809141d7afc8efba596ad1bcadd3a8ab2194afa4c'
BINDING_SHA = '2853d05cf5bd61c50ab604a07fdc3a7f6d7f366cd8fd86e5c544b0fa96a3417c'


class DeliveredAllocationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.inputs = trusted_inputs()
        cls.raw = (CHECKPOINT / 'IDS_ALLOCATED.registry.json').read_bytes()
        cls.delivery = json.loads((CHECKPOINT / 'ids-allocated-milestone.json').read_bytes())
        cls.registry = validate_milestone(cls.delivery, cls.raw, cls.inputs, 'IDS_ALLOCATED')
        cls.binding = json.loads((CHECKPOINT / 'wizard-bindings.json').read_bytes())

    def test_original_bytes_and_portable_provenance(self):
        provenance = json.loads((CHECKPOINT / 'provenance.json').read_bytes())
        for key, expected in [('registry_copy', REGISTRY_SHA), ('milestone_copy', DELIVERY_SHA),
                              ('binding', BINDING_SHA)]:
            row = provenance[key]
            self.assertEqual(hashlib.sha256((CHECKPOINT / row['path']).read_bytes()).hexdigest(), expected)
            self.assertEqual(row['sha256'], expected)
        self.assertEqual(provenance['registry_copy']['original_path'], self.delivery['registry_path'])
        self.assertEqual(provenance['milestone_copy']['original_path'],
                         '/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/ids-allocated-milestone.json')
        self.assertEqual(self.delivery['registry_path'],
                         '/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/milestones/IDS_ALLOCATED.registry.json')
        self.assertEqual(nodes_digest(self.registry, self.inputs), MAP_SHA)
        self.assertEqual(provenance['nodes_map_sha256'], MAP_SHA)
        self.assertEqual(provenance['observed_at'], self.delivery['observed_at'])

    def test_complete_committed_bindings_match_frozen_renderer_and_actual_ids(self):
        renderer_contract()
        validate_candidate(json.loads((ROOT / 'docs/investigations/issue-332-wizard-scope.json').read_bytes()),
                           self.inputs[1]['332'])
        self.assertEqual(self.binding, build_binding(self.registry, self.inputs))
        self.assertEqual(len(self.binding['own_nodes']), 285)
        hp = self.binding['own_nodes']['332:hp']
        self.assertEqual((hp['number'], hp['id'], hp['node_id'], hp['parent'], hp['blocked_by']),
                         (1759, 6432432709, 'I_kwDOSicKg88AAAABf2ciRQ', 2034, [968]))
        paths = self.binding['approved_execution_paths']
        self.assertEqual(set(paths), {'animate-dead', 'phantom-steed', 'shapechange', 'simulacrum',
                                     'named-interoperability', 'complete-cross-class-output-graph'})
        self.assertEqual(paths['animate-dead']['outputs']['331:creature-skeleton-statistics'], 1684)
        self.assertEqual(paths['animate-dead']['outputs']['330:zombie-undead-fortitude'], 1446)
        self.assertEqual(paths['complete-cross-class-output-graph']['outputs']['332:class-gate'], 2036)

    def test_delivered_bytes_and_stable_map_cannot_be_replaced(self):
        changed = copy.deepcopy(self.registry)
        changed['nodes']['332:hp']['id'] += 1
        raw = json.dumps(changed).encode()
        with self.assertRaisesRegex(AssertionError, 'byte hash'):
            validate_milestone(self.delivery, raw, self.inputs, 'IDS_ALLOCATED')
        delivery = copy.deepcopy(self.delivery)
        delivery['registry_file_sha256'] = hashlib.sha256(raw).hexdigest()
        with self.assertRaisesRegex(AssertionError, 'stable identity map'):
            validate_milestone(delivery, raw, self.inputs, 'IDS_ALLOCATED')

    def test_allocated_snapshot_cannot_prove_verified_or_native_state(self):
        with self.assertRaises(AssertionError):
            validate_milestone(self.delivery, self.raw, self.inputs, 'VERIFIED')
        with self.assertRaisesRegex(AssertionError, 'verified wiring boundary'):
            validate_live({}, self.registry, self.inputs)
        self.assertEqual(self.binding['status'], 'IDS_BOUND_NATIVE_READBACK_PENDING')
        self.assertIsNone(self.binding['accepted_parent_prefix'])
        self.assertEqual(self.binding['runtime_status'], 'UNQUALIFIED')
        provenance = json.loads((CHECKPOINT / 'provenance.json').read_bytes())
        self.assertFalse(provenance['actual_native_collected'])
        self.assertIsNone(provenance['verified_milestone'])
        self.assertIsNone(provenance['accepted_parent_prefix'])

    def test_final_cli_still_rejects_missing_prefix_release(self):
        result = subprocess.run([sys.executable, str(ROOT / 'scripts/wizard_publication_proof.py'), '--final'],
                                capture_output=True, text=True, timeout=15)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('accepted POST-WARLOCK root prefix release', result.stderr)
        self.assertIn('no final seal may be fabricated', result.stderr)


if __name__ == '__main__':
    unittest.main()
