"""Offline final-parent controls; positive bodies are synthetic, never native evidence."""
import copy
import tempfile
from pathlib import Path
import unittest

import warlock_parent_seal as seal
import warlock_native_proof as native


class ParentSeal(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.prefixes = seal.load_prefixes()
        cls.sections = seal.load_sections()
        cls.graph = {n: dict(body=cls.prefixes[n] + cls.sections[n]) for n in cls.prefixes}

    def test_exact_and_later_unrelated_append(self):
        self.assertTrue(seal.validate_parents(self.graph))
        graph = {n: dict(body=r['body'] + '\nLater unrelated owner note.\n') for n, r in self.graph.items()}
        self.assertTrue(seal.validate_parents(graph))
        package = seal.package()
        self.assertEqual(package['status'], 'PENDING_ROOT_ONLY_APPENDS')
        self.assertFalse(package['final_prefix_sealed'])
        row = package['artifacts']['47']
        self.assertEqual(row['append_characters'], 70)
        self.assertEqual(row['append_utf8_bytes'], 70)
        self.assertLessEqual(row['expected_body_characters'], 65536)
        self.assertLessEqual(row['expected_body_utf8_bytes'], 65536)

    def test_each_prefix_rollback_change_partial_duplicate_and_altered_duplicate(self):
        _, _, baseline, _, _, _ = seal.a.load_frozen()
        original = {n['number']: n['body'] for n in baseline['graph']}
        for n in self.graph:
            prefix, section = self.prefixes[n], self.sections[n]
            cases = [original[n] + section, prefix, prefix[1:] + section,
                     prefix + section[:-1], prefix + section.replace('Warlock', 'forged'),
                     prefix + section * 2,
                     prefix + section + section.replace(seal.MARKER, '<!-- changed marker -->'),
                     prefix + section + section.replace(seal.MARKER, ''),
                     prefix + section + '\n<!-- EDDA 331 --> altered duplicate\n']
            for body in cases:
                with self.subTest(number=n, case=cases.index(body)):
                    graph = copy.deepcopy(self.graph)
                    graph[n]['body'] = body
                    with self.assertRaises(AssertionError):
                        seal.validate_parents(graph)

    def test_coherent_candidate_and_receipt_section_or_marker_forgery(self):
        for rewrite in [lambda s: s.replace('Warlock', 'Counterfeit'),
                        lambda s: s.replace(seal.MARKER, '<!-- forged331 -->')]:
            with tempfile.TemporaryDirectory() as tmp:
                directory = Path(tmp)
                (directory / seal.RELEASE_FILE).write_bytes((seal.a.DIRECTORY / seal.RELEASE_FILE).read_bytes())
                graph = copy.deepcopy(self.graph)
                for n in self.graph:
                    forged = rewrite(self.sections[n])
                    (directory / f'issue-331-append-{n}.md').write_text(forged)
                    graph[n]['body'] = self.prefixes[n] + forged
                # Even a matching candidate+receipt and recomputed candidate hash cannot set policy.
                with self.assertRaisesRegex(AssertionError, 'independent own section pin'):
                    seal.validate_parents(graph, directory)

    def test_coherent_predecessor_copy_and_hash_forgery(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            release = seal.a.strict_json((seal.a.DIRECTORY / seal.RELEASE_FILE).read_bytes())
            release['body_seals']['47']['body'] += 'forged'
            release['body_seals']['47']['sha256'] = seal.a.digest(release['body_seals']['47']['body'].encode())
            import json
            (directory / seal.RELEASE_FILE).write_text(json.dumps(release))
            with self.assertRaisesRegex(AssertionError, 'independent accepted predecessor delivery pin'):
                seal.load_prefixes(directory)

    def test_synthetic_complete_final_scaffolding_accepts_only_pinned_extensions(self):
        evidence = seal.a.strict_json((seal.a.DIRECTORY / 'issue-331-native-readback.json').read_bytes())
        snapshot = copy.deepcopy(evidence['snapshot'])
        for row in snapshot['graph']:
            if row['number'] in self.graph:
                row['body'] = self.graph[row['number']]['body'] + '\nUnrelated later note.\n'
        _, proposals, _, _, _, _ = seal.a.load_frozen()
        result = native.validate_native(None, proposals[331], evidence, snapshot)
        self.assertEqual(result['status'], 'FINAL_NATIVE_AND_PREFIX_PASS')
        self.assertTrue(result['final_prefix_sealed'])
        self.assertFalse(result['runtime_qualified'])
        # This is a synthetic offline positive control, never written as native evidence.
        next(r for r in snapshot['graph'] if r['number'] == 779)['body'] += 'drift'
        with self.assertRaises(AssertionError):
            native.validate_native(None, proposals[331], evidence, snapshot)

    def test_default_final_native_proof_rejects_unapplied_sections(self):
        evidence = seal.a.strict_json((seal.a.DIRECTORY / 'issue-331-native-readback.json').read_bytes())
        _, proposals, _, _, _, _ = seal.a.load_frozen()
        with self.assertRaisesRegex(AssertionError, 'exact accepted prefix and pinned own section pending'):
            native.validate_native(None, proposals[331], evidence, evidence['snapshot'])


if __name__ == '__main__':
    unittest.main()
