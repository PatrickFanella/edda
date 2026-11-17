"""Actual delivered-byte binding controls; deliberately no live collection."""
import copy
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import warlock_allocation_binding as b
import warlock_native_proof as native
import verify_warlock_scope as scope


class AllocationBinding(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.raw = (b.a.DIRECTORY / b.REGISTRY_COPY).read_bytes()
        cls.delivery_raw = (b.a.DIRECTORY / b.DELIVERY_COPY).read_bytes()
        cls.registry = b.a.strict_json(cls.raw)
        cls.delivery = b.a.strict_json(cls.delivery_raw)
        cls.result = b.bind(cls.raw, cls.delivery_raw)
        _, _, cls.baseline, cls.expected, _, _ = b.a.load_frozen()

    def test_actual_portable_binding_and_every_native_reference(self):
        committed = b.a.strict_json((b.a.DIRECTORY / b.BINDING).read_bytes())
        self.assertEqual(committed, self.result)
        self.assertEqual(len(committed['bindings']), 281)
        hp = committed['bindings']['331:hp']
        self.assertEqual(hp['number'], 1483)
        self.assertEqual(hp['expected_parent_number'], 1748)
        self.assertEqual(hp['expected_blocked_by_numbers'], [774, 968])
        self.assertEqual(hp['expected_rendered_body_sha256'],
                         'a2e9a2df95343ee9ea6f21e12cbdc460dc61b2fddd94dba063fad732c9258b9e')
        def resolve(ref):
            return int(ref) if ref.isdecimal() else self.registry['nodes'][ref]['number']
        for alias, bound in committed['bindings'].items():
            self.assertTrue(alias.startswith('331:'))
            for field, value in self.registry['nodes'][alias].items():
                self.assertEqual(bound[field], value)
            approved = self.expected[alias]
            self.assertEqual(bound['expected_parent_number'], resolve(approved['parent_ref']))
            self.assertEqual(bound['expected_blocked_by_numbers'], sorted(resolve(r) for r in approved['blocked_by_refs']))
        with patch.object(native, 'read_live', side_effect=AssertionError('forbidden live read')):
            self.assertEqual(b.bind(self.raw, self.delivery_raw), committed)
        self.assertEqual(committed['status'], 'ALLOCATION IDS_BOUND')
        for field in ('native_wiring', 'native_readback', 'final_acceptance'):
            self.assertEqual(committed[field], 'UNQUALIFIED')
        self.assertFalse(committed['final_prefix_sealed'])

    def test_exact_byte_pins_reject_whitespace_and_coherent_reassignment(self):
        for raw, delivery in [(self.raw + b' ', self.delivery_raw), (self.raw, self.delivery_raw + b' ')]:
            with self.assertRaises(AssertionError):
                b.bind(raw, delivery)
        changed = copy.deepcopy(self.registry)
        aliases = sorted(changed['nodes'])
        changed['nodes'][aliases[0]], changed['nodes'][aliases[1]] = changed['nodes'][aliases[1]], changed['nodes'][aliases[0]]
        raw = json.dumps(changed).encode()
        delivery = {**self.delivery, 'registry_file_sha256': b.a.digest(raw),
                    'nodes_map_sha256': b.a.nodes_map_digest(changed, self.expected, self.baseline)}
        with self.assertRaisesRegex(AssertionError, 'exact delivered allocation bytes'):
            b.bind(raw, json.dumps(delivery).encode())

    def test_altered_frozen_renderer_fails_before_binding(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            names = [*b.a.FILES.values(), 'issue-331-publication-approval.json',
                     'issue-331-native-baseline.json', 'issue-331-renderer-contract.json',
                     'issue-331-milestone-contract.json']
            for name in names:
                (directory / name).write_bytes((b.a.DIRECTORY / name).read_bytes())
            renderer = directory / 'issue-331-renderer-contract.json'
            renderer.write_bytes(renderer.read_bytes() + b' ')
            with self.assertRaisesRegex(AssertionError, 'exact ROOT renderer pin'):
                b.bind(self.raw, self.delivery_raw, directory)

    def test_allocation_cannot_qualify_wiring_or_final_proof(self):
        with self.assertRaisesRegex(AssertionError, 'requires coordinator VERIFIED'):
            b.a.validate_wiring(None, self.expected, self.baseline, self.registry, {})
        self.assertIsNone(native.ACCEPTED_POST_SORCERER_PREFIXES)
        with self.assertRaisesRegex(AssertionError, 'accepted POST-SORCERER prefix release pending'):
            native.validate_native(None, None, None, None)
        with self.assertRaisesRegex(AssertionError, 'strict proof requires live'):
            scope.validate_publication({}, self.result)
        with self.assertRaisesRegex(AssertionError, 'ROOT final prefix release pending'):
            scope.validate_publication({}, self.result, live=True)


if __name__ == '__main__':
    unittest.main()
