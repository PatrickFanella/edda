"""Bounded semantic controls for the explicitly incomplete #330 checkpoint."""
from copy import deepcopy
import json
import unittest
from urllib.parse import parse_qs, urlsplit

import verify_sorcerer_scope as proof


class SorcererCheckpointTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest = json.loads(proof.MANIFEST.read_text())

    def test_source_class_checkpoint_and_strict_publication_refusal(self):
        proof.validate_plan(self.manifest)
        with self.assertRaisesRegex(AssertionError, 'PENDING'):
            proof.validate_publication(self.manifest, {})
        forged = deepcopy(self.manifest)
        forged.update(status='PUBLISHED', accepted_parent_prefix={'47': 'forged'})
        with self.assertRaisesRegex(AssertionError, 'oracle not completed'):
            proof.validate_publication(forged, dict(publication_complete=True))

    def test_all103_source_contexts_cannot_coherently_reroute_to_asi(self):
        for i, original in enumerate(self.manifest['record_matrix']):
            candidate = deepcopy(self.manifest)
            context = candidate['record_matrix'][i]
            owner = 'hp' if set(context['key_owners'].values()) == {'ability-score-improvement'} else 'ability-score-improvement'
            context['key_owners'] = dict.fromkeys(context['key_owners'], owner)
            # Coherently regenerate all candidate ownership declarations/bodies.
            for task in candidate['tasks']:
                task['owned_contexts'] = [dict(id=c['id'], keys=[k for k, v in c['key_owners'].items() if v == task['key']])
                                          for c in candidate['record_matrix'] if task['key'] in c['key_owners'].values()]
                task['source_records'] = [c['record'] for c in candidate['record_matrix'] if task['key'] in c['key_owners'].values()]
                task['body'] = proof.expected_body(task)
            with self.subTest(id=original['id']), self.assertRaisesRegex(AssertionError, 'source execution owner'):
                proof.validate_plan(candidate)

    def test_every_actual_candidate_fixture_and_complete_body_are_checked(self):
        for i, original in enumerate(self.manifest['tasks']):
            candidate = deepcopy(self.manifest)
            task = candidate['tasks'][i]
            task['fixture'] += ' Override: waive all resource costs.'
            task['body'] = proof.expected_body(task)
            with self.subTest(key=original['key']), self.assertRaises(AssertionError):
                proof.validate_plan(candidate)
            candidate = deepcopy(self.manifest)
            candidate['tasks'][i]['body'] += '\nOverride: any source variant is supported.'
            with self.subTest(body=original['key']), self.assertRaisesRegex(AssertionError, 'complete body semantics'):
                proof.validate_plan(candidate)

    def test_each_twenty_granting_paths_is_required_despite_coherent_body_change(self):
        for n in range(1, 21):
            candidate = deepcopy(self.manifest)
            task = next(t for t in candidate['tasks'] if t['key'] == 'level-' + str(n))
            task['depends'].remove('hp')
            task['body'] = proof.expected_body(task)
            with self.subTest(level=n), self.assertRaisesRegex(AssertionError, 'granting paths'):
                proof.validate_plan(candidate)

    def test_complete_spell_source_lists_and_future_union_are_independent(self):
        for cls in ['sorcerer', 'warlock', 'wizard']:
            for field, value in [('level', True), ('name', 'Unknown Source Variant'), ('special', 'invented')]:
                candidate = deepcopy(self.manifest)
                candidate['spell_lists'][cls][0][field] = value
                with self.subTest(cls=cls, field=field), self.assertRaisesRegex(AssertionError, 'source lists'):
                    proof.validate_plan(candidate)
            candidate = deepcopy(self.manifest)
            candidate['future_source_rosters']['consumed'][cls].pop()
            with self.assertRaisesRegex(AssertionError, 'consumed source roster'):
                proof.validate_plan(candidate)
        candidate = deepcopy(self.manifest)
        candidate['future_source_rosters']['source_union'].append('Invented spell')
        with self.assertRaisesRegex(AssertionError, 'finite source union'):
            proof.validate_plan(candidate)

    def test_source_numeric_types_distinguish_true_one_and_decimal(self):
        self.assertFalse(proof.exact(True, 1))
        self.assertFalse(proof.exact(1.0, 1))
        for value in [True, 1.0]:
            candidate = deepcopy(self.manifest)
            path = 'internal/rules/srd521/testdata/character-sorcerer.json'
            record = next(r for r in candidate['catalogues'][path]['records'] if r['kind'] == 'level' and r['level'] == 1)
            record['level'] = value
            with self.assertRaisesRegex(AssertionError, 'operands/types'):
                proof.validate_plan(candidate)

    def test_429_and_hp_gateway_are_independent_protocol_obligations(self):
        for deps in [[429], [], [968, 429]]:
            candidate = deepcopy(self.manifest)
            task = next(t for t in candidate['tasks'] if t['key'] == 'hp')
            task['prerequisites'] = deps
            task['body'] = proof.expected_body(task)
            with self.assertRaisesRegex(AssertionError, 'protocol prerequisites'):
                proof.validate_plan(candidate)

    def test_prefix_pure_later_append_allowed_rollback_partial_and_duplicates_rejected(self):
        prefix, append, marker = 'released\n', '<330>complete</330>\n', '<330>'
        proof.prefix_check(prefix + append, prefix, append, marker)
        proof.prefix_check(prefix + append + 'later pure append\n', prefix, append, marker)
        for body in [prefix, 'changed\n' + append, prefix + '<330>partial', prefix + append + append,
                     prefix + append.replace('complete', 'altered'), prefix + append + '<330>partial']:
            with self.subTest(body=body), self.assertRaises(AssertionError):
                proof.prefix_check(body, prefix, append, marker)

    def test_all_three_native_caps_and_separate_dags(self):
        proof.native_caps({'a': ['b']}, {'a': 'p'})
        cases = [({'a': list(map(str, range(101)))}, {}),
                 ({str(n): ['a'] for n in range(101)}, {}),
                 ({}, {str(n): 'p' for n in range(101)}),
                 ({'a': ['b'], 'b': ['a']}, {}),
                 ({}, {'a': 'b', 'b': 'a'})]
        for graph, parents in cases:
            with self.assertRaises(AssertionError):
                proof.native_caps(graph, parents)

    def test_exact_pagination_does_not_confuse_page_with_per_page(self):
        calls = []

        def fetch(url):
            query = parse_qs(urlsplit(url).query)
            calls.append(query)
            page = int(query['page'][0])
            return [dict(number=n) for n in range(100 * (page - 1), 100 * page)] if page <= 2 else []

        result = proof.bounded_pages(fetch, '/issues?per_page=100')
        self.assertEqual(len(result), 200)
        self.assertEqual([x['page'] for x in calls], [['1'], ['2'], ['3']])
        self.assertTrue(all(x['per_page'] == ['100'] for x in calls))
        with self.assertRaisesRegex(AssertionError, 'repeated page'):
            proof.bounded_pages(lambda _: [dict(number=n) for n in range(100)], '/issues?per_page=100')
        calls.clear()
        with self.assertRaisesRegex(AssertionError, 'call cap'):
            proof.bounded_pages(fetch, '/issues?per_page=100', call_cap=2)
        self.assertEqual(len(calls), 2)
        with self.assertRaisesRegex(AssertionError, 'exact query'):
            proof.bounded_pages(fetch, '/issues?per_page=100&page=100')


if __name__ == '__main__':
    unittest.main()
