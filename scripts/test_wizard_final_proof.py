"""Contrary controls only: synthetic appends never prove ROOT publication."""
import copy
import gzip
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
import unittest

from wizard_final_proof import (ROOT, DIRECTORY, MARKER, load_policy,
                                validate_parent_extensions, validate_final)
from wizard_publication_proof import trusted_inputs, validate_live


class WizardFinalProofTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.inputs = trusted_inputs()
        cls.release, cls.expected = load_policy()
        cls.original = json.loads(gzip.decompress((ROOT / 'docs/investigations/issue-332-native-proof/wizard-native-receipt.json.gz').read_bytes()))
        cls.registry = json.loads((ROOT / 'docs/investigations/issue-332-native-proof/VERIFIED.registry.json').read_bytes())
        cls.fixture = copy.deepcopy(cls.original)
        cls.fixture['status'] = 'SYNTHETIC FINAL APPENDS TEST ONLY; not collected native evidence'
        cls.fixture['observed_at'] = '2026-11-17T07:08:00+00:00'
        for row in cls.fixture['graph']:
            if row['number'] in cls.expected:
                row['body'] = cls.expected[row['number']]
            if row['number'] in (330, 331):
                row['state'] = 'CLOSED'

    def altered(self, number, **changes):
        receipt = copy.deepcopy(self.fixture)
        next(r for r in receipt['graph'] if r['number'] == number).update(changes)
        return receipt

    def test_package_is_pending_with_unicode_capacity_checked(self):
        release, expected = load_policy()
        self.assertEqual(set(expected), {47, 54, 283, 337})
        self.assertEqual(len(expected[47]), 65184)
        self.assertEqual(len(expected[47].encode()), 65278)
        package = json.loads((DIRECTORY / 'root-append-package.json').read_bytes())
        self.assertTrue(package['status'].startswith('PENDING'))
        self.assertIsNone(package['accepted_parent_prefix'])

    def test_exact_synthetic_appends_pass_all_native_checks_without_runtime_admission(self):
        proof = validate_final(self.fixture, self.registry, self.inputs)
        self.assertTrue(proof['final_prefix_sealed'])
        self.assertEqual(proof['runtime_implementation_status'], 'PAUSED')
        self.assertTrue(proof['runtime_status'].startswith('UNQUALIFIED'))
        self.assertEqual(len(proof['source_bindings']), 6)

    def test_unrelated_later_pure_appends_pass(self):
        receipt = copy.deepcopy(self.fixture)
        for row in receipt['graph']:
            if row['number'] in self.expected:
                row['body'] += '\n<!-- unrelated-later -->\nLater note.\n'
        validate_final(receipt, self.registry, self.inputs)

    def test_real_saved_unsealed_evidence_cannot_pass_final(self):
        with self.assertRaises(AssertionError):
            validate_final(self.original, self.registry, self.inputs)

    def test_rollback_or_partial_four_appends_reject(self):
        for number in self.expected:
            with self.subTest(number=number):
                body = self.release['body_seals'][str(number)]['body']
                for changed in [body, self.expected[number][:-1], self.expected[number][1:]]:
                    with self.assertRaises(AssertionError):
                        validate_final(self.altered(number, body=changed), self.registry, self.inputs)

    def test_changed_prefix_or_section_reject_even_with_forged_receipt_hashes(self):
        for number in self.expected:
            for changed in [self.expected[number].replace('source', 'forged', 1),
                            self.expected[number].replace('execution gated', 'execution admitted').replace('playable admission', 'playable permission')]:
                receipt = self.altered(number, body=changed)
                receipt['expected_body_sha256'] = hashlib.sha256(changed.encode()).hexdigest()
                receipt['accepted_parent_prefix'] = 'forged-and-rehashed'
                with self.assertRaises(AssertionError):
                    validate_final(receipt, self.registry, self.inputs)

    def test_duplicate_and_altered_duplicate_owned_sections_reject(self):
        for number in self.expected:
            prefix = self.release['body_seals'][str(number)]['body']
            section = self.expected[number][len(prefix):]
            for duplicate in [section, section.replace('execution gated', 'execution admitted'),
                              section.replace(MARKER, '<!-- edda-332-copy -->'),
                              section.replace(MARKER, '<!-- forged-marker -->')]:
                with self.assertRaises(AssertionError):
                    validate_parent_extensions(self.altered(number, body=self.expected[number] + duplicate))

    def test_unicode_bytes_overflow_rejects_while_character_count_fits(self):
        body = self.expected[47] + 'é' * 130
        self.assertLess(len(body), 65536)
        self.assertGreater(len(body.encode()), 65536)
        with self.assertRaisesRegex(AssertionError, 'Unicode'):
            validate_parent_extensions(self.altered(47, body=body))

    def test_coherent_manifest_receipt_and_marker_rewrite_cannot_authorize_policy(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp) / 'policy'
            shutil.copytree(DIRECTORY, directory)
            package = json.loads((directory / 'root-append-package.json').read_bytes())
            package['marker'] = '<!-- rewritten332 -->'
            for number in self.expected:
                path = directory / f'append-{number}.md'
                raw = path.read_bytes().replace(MARKER.encode(), package['marker'].encode())
                path.write_bytes(raw)
                package['appends'][str(number)]['section_sha256'] = hashlib.sha256(raw).hexdigest()
            (directory / 'root-append-package.json').write_text(json.dumps(package))
            receipt = copy.deepcopy(self.fixture)
            for row in receipt['graph']:
                row['body'] = row['body'].replace(MARKER, package['marker'])
            receipt['policy_sha256'] = hashlib.sha256((directory / 'root-append-package.json').read_bytes()).hexdigest()
            with self.assertRaisesRegex(AssertionError, 'pinned ROOT-only append package'):
                validate_parent_extensions(receipt, directory)

    def test_rewritten_release_and_self_consistent_body_hashes_reject(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp) / 'policy'
            shutil.copytree(DIRECTORY, directory)
            release = copy.deepcopy(self.release)
            release['body_seals']['47']['body'] += 'forged'
            release['body_seals']['47']['sha256'] = hashlib.sha256(release['body_seals']['47']['body'].encode()).hexdigest()
            (directory / 'accepted-post-warlock-parent-baselines.json.gz').write_bytes(gzip.compress(json.dumps(release).encode()))
            with self.assertRaisesRegex(AssertionError, 'pinned accepted release'):
                load_policy(directory)

    def test_changed_section_file_or_literal_marker_reject_with_unchanged_package(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp) / 'policy'
            shutil.copytree(DIRECTORY, directory)
            path = directory / 'append-47.md'
            path.write_bytes(path.read_bytes().replace(b'edda-332', b'edda-333'))
            with self.assertRaisesRegex(AssertionError, 'literal owned section'):
                load_policy(directory)

    def test_no_general_exception_for_old_bodies_titles_or_identity(self):
        for number, change in [(779, {'body': 'changed old owner'}),
                               (47, {'title': 'changed parent title'}),
                               (331, {'node_id': 'changed accepted predecessor identity'})]:
            with self.assertRaises(AssertionError):
                validate_final(self.altered(number, **change), self.registry, self.inputs)

    def test_new_body_metadata_or_identity_changes_reject(self):
        for change in [dict(body='rewritten owned body'), dict(type='Feature'),
                       dict(id=1), dict(node_id='forged'), dict(state='CLOSED')]:
            with self.assertRaises(AssertionError):
                validate_final(self.altered(1759, **change), self.registry, self.inputs)
        receipt = self.altered(1759)
        next(r for r in receipt['graph'] if r['number'] == 1759)['projects'][0]['fieldValueByName']['name'] = 'L'
        with self.assertRaises(AssertionError):
            validate_final(receipt, self.registry, self.inputs)

    def test_parent_prerequisites_mandatory_paths_and_caps_remain_required(self):
        for number, changes in [(1759, dict(parent=332)), (1759, dict(blocked_by=[])),
                                (337, dict(blocked_by=[])), (279, dict(blocked_by=list(range(1, 102))))]:
            with self.assertRaises(AssertionError):
                validate_final(self.altered(number, **changes), self.registry, self.inputs)

    def test_default_unsealed_proof_does_not_accept_new_parent_exception(self):
        with self.assertRaises(AssertionError):
            validate_live(self.fixture, self.registry, self.inputs)

    def test_predecessor_reopening_or_stale_receipt_reject(self):
        for number in [330, 331]:
            with self.assertRaises(AssertionError):
                validate_final(self.altered(number, state='OPEN'), self.registry, self.inputs)
        receipt = copy.deepcopy(self.fixture)
        receipt['observed_at'] = self.original['observed_at']
        with self.assertRaisesRegex(AssertionError, 'predates'):
            validate_final(receipt, self.registry, self.inputs)


if __name__ == '__main__':
    unittest.main()
