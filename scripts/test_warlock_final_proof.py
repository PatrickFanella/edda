"""Offline replay and contrary controls for the actual sealed native receipt."""
import copy
import unittest
from unittest.mock import patch

import warlock_final_proof as proof


class FinalPublication(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.evidence = proof.a.strict_json(proof.EVIDENCE.read_bytes())

    def test_actual_sealed_native_proof_passes_without_runtime_claim(self):
        with patch.object(proof.native, 'read_live', side_effect=AssertionError('offline replay only')):
            result = proof.verify(self.evidence)
        self.assertEqual(result['status'], 'FINAL_NATIVE_AND_PREFIX_PASS')
        self.assertEqual(result['source_publication'], 'PASS')
        self.assertTrue(result['final_prefix_sealed'])
        self.assertFalse(result['runtime_qualified'])
        self.assertEqual(result['runtime_implementation'], 'PAUSED')
        self.assertEqual(result['decision_acceptance'], 'PENDING_ROOT_REVIEW_CI_MERGE')

    def test_coherent_root_provenance_or_native_parent_rewrite_fails(self):
        evidence = copy.deepcopy(self.evidence)
        name = next(iter(evidence['root_publication']))
        evidence['root_publication'][name]['sha256'] = 'f' * 64
        with self.assertRaisesRegex(AssertionError, 'exact sealed evidence provenance'):
            proof.verify(evidence)
        for n in (47, 54, 283, 337):
            evidence = copy.deepcopy(self.evidence)
            row = next(r for r in evidence['snapshot']['graph'] if r['number'] == n)
            row['body'] = row['body'].replace('Warlock decision', 'Counterfeit decision') if n == 47 else row['body'][:-1]
            with self.assertRaises(AssertionError):
                proof.verify(evidence)

    def test_native_source_fields_and_old_protections_remain_required(self):
        number = self.evidence['registry']['nodes']['331:hp']['number']
        for n, field, value in [(number, 'body', 'changed source contract'),
                                 (number, 'parent', 331), (number, 'blocked_by', [774]),
                                 (779, 'body', 'old owner changed'), (331, 'parent', 47)]:
            evidence = copy.deepcopy(self.evidence)
            next(r for r in evidence['snapshot']['graph'] if r['number'] == n)[field] = value
            with self.subTest(number=n, field=field), self.assertRaises(AssertionError):
                proof.verify(evidence)
        evidence = copy.deepcopy(self.evidence)
        evidence['snapshot']['project'][str(number)]['size'] = 'S'
        with self.assertRaises(AssertionError):
            proof.verify(evidence)


if __name__ == '__main__':
    unittest.main()
