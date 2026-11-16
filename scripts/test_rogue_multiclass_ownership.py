"""The complete multiclass grant stays separate from its HP-only dependency."""
from copy import deepcopy
import json
import unittest
from unittest.mock import patch

import build_rogue_checkpoint as generator
import rogue_native_proof as native
import verify_rogue_scope as entry


COMPOSITE = 'srd-5.2.1:mechanic/rogue/multiclass/hit-die-one-class-skill-thieves-tools-light-armor-and-level-1-features'


class MulticlassOwnershipTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest = json.loads(entry.PATH.read_text())
        cls.receipt = json.loads(native.RECEIPT.read_text())

    def test_literal_ownership_in_generated_and_published_contracts(self):
        entry.multiclass_ownership(generator.audit(write=False))
        entry.multiclass_ownership(self.manifest)
        rows = {x['key']: x for x in self.manifest['published_issues']}
        self.assertEqual(rows['hp']['issue'], 1098)
        self.assertEqual(rows['multiclass']['issue'], 1102)
        self.assertIn(1098, rows['multiclass']['blocked_by_issues'])

    def test_coherent_reroute_rejects_even_if_generator_agrees(self):
        candidate = deepcopy(self.manifest)
        record = next(x for x in candidate['records'] if x['record']['kind'] == 'class')
        record['key_owners'][COMPOSITE] = ['hp']
        for field in ['issues', 'published_issues']:
            rows = {x['key']: x for x in candidate[field]}
            rows['multiclass']['source_keys'].remove(COMPOSITE)
            rows['hp']['source_keys'].append(COMPOSITE)
            rows['hp']['source_keys'].sort()
        with patch.object(generator, 'audit', return_value=candidate):
            with self.assertRaisesRegex(AssertionError, 'composite multiclass grant owner'):
                entry.semantic(candidate)
        with self.assertRaisesRegex(AssertionError, 'composite multiclass grant owner'):
            native.validate(candidate, self.receipt, True)

    def test_hp_dependency_cannot_be_erased_from_either_contract(self):
        candidate = deepcopy(self.manifest)
        next(x for x in candidate['issues'] if x['key'] == 'multiclass')['depends'].remove('hp')
        with patch.object(generator, 'audit', return_value=candidate):
            with self.assertRaisesRegex(AssertionError, 'multiclass HP dependency'):
                entry.semantic(candidate)
        candidate = deepcopy(self.manifest)
        next(x for x in candidate['published_issues'] if x['key'] == 'multiclass')['blocked_by_issues'].remove(1098)
        with self.assertRaisesRegex(AssertionError, 'published multiclass HP dependency'):
            native.validate(candidate, self.receipt, True)


if __name__ == '__main__':
    unittest.main()
