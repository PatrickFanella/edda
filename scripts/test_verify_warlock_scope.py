"""Contrary probes against the actual Warlock candidate, not a synthetic receipt."""
import copy
import json
from pathlib import Path
import unittest
from unittest.mock import patch

import verify_warlock_scope as v


class WarlockCandidateProof(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.candidate = json.loads(v.MANIFEST.read_text())
        cls.expected = v.source_records(Path('/tmp/edda-wave1/SRD_CC_v5.2.1.pdf'))

    def reject(self, mutation):
        candidate = copy.deepcopy(self.candidate)
        mutation(candidate)
        with self.assertRaises(AssertionError):
            v.validate_candidate(candidate, self.expected)

    def test_actual_candidate(self):
        v.validate_candidate(self.candidate, self.expected)

    def test_every_record_and_context_required(self):
        for n in range(191):
            self.reject(lambda m, n=n: m['records'].pop(n))
            self.reject(lambda m, n=n: m['record_matrix'].pop(n))

    def test_every_context_owner_independently_protected(self):
        for n, row in enumerate(self.candidate['record_matrix']):
            for key in row['key_owners']:
                self.reject(lambda m, n=n, key=key: m['record_matrix'][n]['key_owners'].__setitem__(key, 'contrary-owner'))

    def test_numeric_json_is_typed(self):
        for n, record in enumerate(self.candidate['records']):
            for key, value in record.get('parameters', {}).items():
                if type(value) is int:
                    for bad in [True, float(value)]:
                        self.reject(lambda m, n=n, key=key, bad=bad: m['records'][n]['parameters'].__setitem__(key, bad))

    def test_all_independent_grants(self):
        for n, task in enumerate(self.candidate['tasks']):
            if task['key'].startswith('level-'):
                for dep in task['depends']:
                    if dep in ('hp', 'preparation', 'eldritch-invocations', 'ability-score-improvement', 'mystic-arcanum', 'warlock-subclass'):
                        self.reject(lambda m, n=n, dep=dep: m['tasks'][n]['depends'].remove(dep))

    def test_coherent_body_fixture_and_fake_receipt_weakening(self):
        for n in range(len(self.candidate['tasks'])):
            def weaken(m, n=n):
                m['tasks'][n]['fixture'] = 'Accept any known spell with no target or resource validation.'
                m['tasks'][n]['body'] = v.body(m['tasks'][n]['key'], m['tasks'][n]['fixture'])
                m['native_receipt'] = {'body': m['tasks'][n]['body'], 'state': 'PUBLISHED'}
            self.reject(weaken)

    def test_every_actual_shared_and_creature_contract(self):
        for field in ('shared_effect_tasks', 'creature_tasks', 'gear_tasks'):
            for n in range(len(self.candidate[field])):
                def weaken(m, n=n, field=field):
                    t = m[field][n]
                    t['fixture'] = 'Generic unsupported effect with no bounds.'
                    t['body'] = t['fixture']
                    m['native_receipt'] = {'state': 'PUBLISHED', 'body': t['body']}
                self.reject(weaken)

    def test_mind_spike_does_not_grant_sorcerer_membership(self):
        self.reject(lambda m: m['source_conflicts'].__setitem__('Sorcerer–Mind Spike', 'Header automatically grants Sorcerer list membership.'))
        self.reject(lambda m: m['familiar_source_roster']['normal'].append('Venomous Snake'))
        self.reject(lambda m: m['familiar_source_roster']['chain'].remove('Sphinx of Wonder'))

    def test_whole_owner_reuse_rejects_coherent_name_only_matches(self):
        for name, issue in [('Gate', 179), ('Light', 1076), ('Message', 640), ('Resistance', 586), ('Find Familiar', 181)]:
            def fake(m, name=name, issue=issue):
                row = next(x for x in m['spell_overlap_audit'] if x['name'] == name)
                row['candidate_owners'] = [dict(issue=issue, title='forged matching title', body='forged matching body')]
                m['native_receipt'] = {'state': 'PUBLISHED', 'owner': issue}
            self.reject(fake)

    def test_unequal_actual_candidate_class_total_operands(self):
        for n in range(2):
            for field, bad in [('warlock_row', True), ('total_row', 5.0), ('experience_points', 300), ('proficiency_bonus', 2), ('hp_mode', 'first-character-D8')]:
                self.reject(lambda m, n=n, field=field, bad=bad: m['transition_witnesses'][n].__setitem__(field, bad))

    def test_complete_actual_native_proposal_and_coherent_weakening(self):
        import warlock_native_proof as native
        proposal = json.loads(native.PROPOSAL.read_text())
        baseline = json.loads(native.BASELINE.read_text())
        foreign = json.loads(native.FOREIGN.read_text())
        result = native.validate_proposal(self.candidate, proposal, baseline, foreign)
        self.assertFalse(result['unresolved_foreign'])
        for section, n, field, bad in [('tasks', 0, 'native_body', 'weakened complete body'),
                                       ('gates', 0, 'prerequisites', []),
                                       ('gates', 0, 'parent', 283)]:
            wrong = copy.deepcopy(proposal)
            wrong[section][n][field] = bad
            with self.assertRaises(AssertionError):
                native.validate_proposal(self.candidate, wrong, baseline, foreign)

    def test_graphql_pages_never_hang_on_forged_api(self):
        import warlock_native_proof as native
        class Repeated:
            calls = 0
            def __call__(self, query, variables):
                self.calls += 1
                return {'issues': {'nodes':[dict(number=1, id='I_old', databaseId=1, title='x', body='', state='OPEN', parent=None,
                    blockedBy={'pageInfo': {'hasNextPage':False}, 'nodes':[]})],
                    'pageInfo': {'hasNextPage':True, 'endCursor':'same'}}}
        fake = Repeated()
        with patch.object(native, 'Reader', return_value=fake):
            with self.assertRaises(AssertionError):
                native.read_live({})
        self.assertEqual(fake.calls, 2)

    def test_observed_baseline_never_becomes_future_root_release(self):
        import warlock_native_proof as native
        self.assertIsNone(native.ACCEPTED_POST_SORCERER_PREFIXES)
        with self.assertRaises(AssertionError):
            native.validate_native(self.candidate, {}, {'state':'PUBLISHED', 'root_released':True}, {})

    def test_actual_native_flags_cannot_forge_publication(self):
        for live in [False, True]:
            receipt = {'state': 'PUBLISHED', 'root_accepted_post_sorcerer_release': True,
                       'root_serial_appends': [47, 54, 283, 337]}
            with self.assertRaises(AssertionError):
                v.validate_publication(self.candidate, receipt, live=live)

    def test_hp_adapter_and_metadata(self):
        n = next(n for n, t in enumerate(self.candidate['tasks']) if t['key'] == 'hp')
        self.reject(lambda m: m['tasks'][n].__setitem__('prerequisites', [429]))
        for field, bad in [('type', 'Feature'), ('size', 'L'), ('parent', 283)]:
            self.reject(lambda m, field=field, bad=bad: m['tasks'][n].__setitem__(field, bad))

    def test_dependency_and_containment_separate_caps(self):
        v.caps({'a': ['b']}, {'b': 'a'})
        for graph, parents in [({'a': ['b'], 'b': ['a']}, {}), ({}, {'a': 'b', 'b': 'a'}),
                               ({'a': list(range(101))}, {}),
                               ({n: ['root'] for n in range(101)}, {}),
                               ({}, {n: 'parent' for n in range(101)})]:
            with self.assertRaises(AssertionError):
                v.caps(graph, parents)

    def test_bounded_actual_queries(self):
        endpoint = '/issues?state=all&per_page=100'
        calls = []
        def fetch(url):
            calls.append(url)
            return [{'number': n} for n in range(100)]
        with self.assertRaises(AssertionError):
            v.bounded_pages(fetch, endpoint)
        self.assertEqual(calls, [endpoint + '&page=1', endpoint + '&page=2'])
        calls.clear()
        def endless(url):
            calls.append(url)
            return [{'number': len(calls) * 100 + n} for n in range(100)]
        with self.assertRaises(AssertionError):
            v.bounded_pages(endless, endpoint, limit=3)
        self.assertEqual(len(calls), 3)

    def test_exact_prefix_and_permitted_later_append(self):
        prefix, append = 'old owner body\n', '<!-- edda331 --> exact root supplement\n'
        v.prefix_check(prefix + append + 'later legitimate append\n', prefix, append, '<!-- edda331 -->')
        for actual in [append, prefix, prefix + append.replace('exact', 'weak'), prefix + append * 2]:
            with self.assertRaises(AssertionError):
                v.prefix_check(actual, prefix, append, '<!-- edda331 -->')


if __name__ == '__main__':
    unittest.main()
