"""Bounded offline replay of the actual unsealed native evidence."""
import copy
import unittest
from unittest.mock import patch

import warlock_verified_proof as proof
import warlock_native_proof as native


class VerifiedNativeEvidence(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.approval, cls.baseline, cls.expected, cls.registry, _, _ = proof.inputs()
        cls.evidence = proof.a.strict_json(proof.EVIDENCE.read_bytes())

    def verify(self, snapshot):
        return proof.verify(snapshot, self.approval, self.baseline, self.expected, self.registry)

    def test_actual_complete_snapshot_passes_without_final_seal(self):
        with patch.object(native, 'read_live', side_effect=AssertionError('offline replay only')):
            result = self.verify(self.evidence['snapshot'])
        self.assertEqual(result['status'], 'NATIVE_WIRING_PASS_PREFIX_PENDING')
        self.assertEqual(result['native_readback'], 'PASS')
        self.assertEqual(result['final_acceptance'], 'UNQUALIFIED')
        self.assertFalse(result['final_prefix_sealed'])
        self.assertFalse(result['runtime_qualified'])
        self.assertIsNone(native.ACCEPTED_POST_SORCERER_PREFIXES)
        with self.assertRaisesRegex(AssertionError, 'accepted POST-SORCERER prefix release pending'):
            native.validate_native(None, None, None, None)

    def test_actual_owned_body_parent_edge_and_metadata_drift_fail(self):
        number = self.registry['nodes']['331:hp']['number']
        for field, value in [('body', 'altered native body'), ('parent', 331), ('blocked_by', [774])]:
            with self.subTest(field=field):
                snapshot = copy.deepcopy(self.evidence['snapshot'])
                row = next(n for n in snapshot['graph'] if n['number'] == number)
                row[field] = value
                with self.assertRaises(AssertionError):
                    self.verify(snapshot)
        snapshot = copy.deepcopy(self.evidence['snapshot'])
        snapshot['project'][str(number)]['size'] = 'S'
        with self.assertRaises(AssertionError):
            self.verify(snapshot)

    def test_old_owner_drift_and_allocation_status_fail(self):
        snapshot = copy.deepcopy(self.evidence['snapshot'])
        next(n for n in snapshot['graph'] if n['number'] == 779)['body'] += 'changed'
        with self.assertRaises(AssertionError):
            self.verify(snapshot)
        registry = {**self.registry, 'status': 'IDS_ALLOCATED'}
        with self.assertRaisesRegex(AssertionError, 'requires coordinator VERIFIED'):
            proof.verify(self.evidence['snapshot'], self.approval, self.baseline, self.expected, registry)


if __name__ == '__main__':
    unittest.main()
