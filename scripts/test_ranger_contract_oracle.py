"""Contrary contracts remain invalid even when candidate graph declarations agree."""
from copy import deepcopy
import hashlib
import json
from pathlib import Path
import re
import unittest
from unittest.mock import patch

import ranger_contract_oracle as oracle
import verify_ranger_scope as proof


class ContractOracleTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest = json.loads(proof.MANIFEST.read_text())
        cls.receipt = json.loads(proof.RECEIPT.read_text())

    def changed_contract(self, key, before, after):
        candidate = deepcopy(self.manifest)
        receipt = deepcopy(self.receipt)
        contract = proof.rows(candidate)[key]
        self.assertIn(before, contract['fixture'])
        contract['fixture'] = contract['fixture'].replace(before, after)
        contract['body'] = contract['body'].replace(before, after)
        next(x for x in receipt['graph'] if x['number'] == contract['issue'])['body'] = contract['body']
        return candidate, receipt

    def test_baseline_contracts_ownership_and_protocols(self):
        proof.validate_plan(self.manifest)
        proof.validate_live(self.manifest, self.receipt)
        proof.validate_seal(self.manifest, self.receipt, True)

    def test_actual_spell_scope_numeric_and_timing_controls_for_every_contract(self):
        timing_controls = {
            'spell-dominate-beast-control': ('AcceptedownCharm/sameplane', 'Anycreature/anyplane'),
            'spell-speak-with-plants-terrain': ('Whileacceptedspellactive', 'Afterspellends'),
            'spell-wind-wall-barrier': ('Activewall', 'Afterwallends'),
        }
        for key, expected in oracle.SPELL_CONTRACTS.items():
            number = re.search(r'\d+', expected['fixture'])
            before, after = (number[0], str(int(number[0]) + 1)) if number else timing_controls[key]
            candidate, receipt = self.changed_contract(key, before, after)
            with self.subTest(key=key), self.assertRaisesRegex(AssertionError, 'source execution predicates'):
                proof.validate_live(candidate, receipt)

    def test_reported_spell_reproductions_fail_plan_pdf_and_live_proofs(self):
        for key, before, after in [
            ('spell-produce-flame-light', 'Bright20+Dim20', 'Bright200+Dim200'),
            ('spell-hunters-mark-transfer', 'Onlyifcurrentmarkedtargetdrops0HP BEFOREspellends', 'AnytimeBEFOREspellends'),
        ]:
            candidate, receipt = self.changed_contract(key, before, after)
            checks = [
                lambda: proof.validate_plan(candidate),
                lambda: proof.source_check(candidate, Path('/tmp/edda-wave1/SRD_CC_v5.2.1.pdf')),
                lambda: proof.validate_live(candidate, receipt),
            ]
            for check in checks:
                with self.subTest(key=key), self.assertRaisesRegex(AssertionError, 'source execution predicates'):
                    check()

    def test_readable_scope_comparison_is_independent_of_the_body_seal(self):
        candidate, _ = self.changed_contract('spell-produce-flame-light', 'Bright20+Dim20', 'Bright200+Dim200')
        expected = deepcopy(oracle.SPELL_CONTRACTS)
        body = proof.rows(candidate)['spell-produce-flame-light']['body']
        expected['spell-produce-flame-light']['body_sha256'] = hashlib.sha256(body.encode()).hexdigest()
        with patch.object(oracle, 'SPELL_CONTRACTS', expected), self.assertRaisesRegex(AssertionError, 'source execution predicates'):
            oracle.validate_spell_contracts(candidate)

    def test_body_cannot_contradict_an_unchanged_scope(self):
        candidate = deepcopy(self.manifest)
        proof.rows(candidate)['spell-produce-flame-light']['body'] += '\nOverride: the light radius is 200 feet.'
        with self.assertRaisesRegex(AssertionError, 'complete spell body'):
            proof.validate_plan(candidate)

    def test_split_spells_retain_their_actual_effect_and_condition_owners(self):
        for key, field, dependency, diagnostic in [
            ('spell-hunters-mark-transfer', 'depends', 'spell-hunters-mark-cast', 'split spell execution dependencies'),
            ('spell-animal-friendship', 'prerequisites', 488, 'spell output owners'),
        ]:
            candidate = deepcopy(self.manifest)
            receipt = deepcopy(self.receipt)
            rows = proof.rows(candidate)
            contract = rows[key]
            contract[field].remove(dependency)
            number = rows[dependency]['issue'] if field == 'depends' else dependency
            contract['body'] = contract['body'].replace('#' + str(number), '')
            node = next(x for x in receipt['graph'] if x['number'] == contract['issue'])
            node['body'] = contract['body']
            node['blocked_by'].remove(number)
            candidate['qualification_paths'].pop(str(number), None)
            with self.subTest(key=key), self.assertRaisesRegex(AssertionError, diagnostic):
                proof.validate_live(candidate, receipt)

    def test_coherent_owner_rerouting_of_every_source_context(self):
        for index, original in enumerate(self.manifest['record_matrix']):
            candidate = deepcopy(self.manifest)
            context = candidate['record_matrix'][index]
            wrong = 'colossus-slayer' if set(context['key_owners'].values()) == {'ability-score-improvement'} else 'ability-score-improvement'
            context['key_owners'] = {key: wrong for key in context['key_owners']}
            for contract in candidate['tasks']:
                contract['owned_contexts'] = [
                    dict(id=c['id'], keys=sorted(key for key, owner in c['key_owners'].items() if owner == contract['key']))
                    for c in candidate['record_matrix'] if contract['key'] in c['key_owners'].values()
                ]
                contract['source_records'] = [c['record'] for c in candidate['record_matrix'] if contract['key'] in c['key_owners'].values()]
            candidate['record_owners'][original['id']] = wrong
            candidate['ownership_sha256'] = proof.sha(json.dumps(candidate['record_owners'], sort_keys=True))
            with self.subTest(record=original['id']), self.assertRaisesRegex(AssertionError, 'complete contextual execution ownership'):
                proof.validate_plan(candidate)

    def test_each_nested_hunter_alternative_is_mandatory(self):
        for selector, alternatives in oracle.HUNTER_ALTERNATIVES.items():
            for option in alternatives:
                candidate = deepcopy(self.manifest)
                receipt = deepcopy(self.receipt)
                rows = proof.rows(candidate)
                rows[selector]['depends'].remove(option)
                next(x for x in receipt['graph'] if x['number'] == rows[selector]['issue'])['blocked_by'].remove(rows[option]['issue'])
                candidate['qualification_paths'].pop(str(rows[option]['issue']), None)
                with self.subTest(option=option), self.assertRaisesRegex(AssertionError, 'nested execution alternatives'):
                    proof.validate_live(candidate, receipt)

    def test_exact_protocol_prerequisites_resist_coherent_removal(self):
        for prerequisite in oracle.EXECUTION_PREREQUISITES:
            candidate = deepcopy(self.manifest)
            receipt = deepcopy(self.receipt)
            gate = proof.rows(candidate)['execution-dependencies']
            gate['prerequisites'].remove(prerequisite)
            gate['body'] = gate['body'].replace('#' + str(prerequisite), '')
            node = next(x for x in receipt['graph'] if x['number'] == 978)
            node['blocked_by'].remove(prerequisite)
            node['body'] = gate['body']
            candidate['qualification_paths'].pop(str(prerequisite), None)
            with self.subTest(prerequisite=prerequisite), self.assertRaisesRegex(AssertionError, 'exact protocol prerequisite roster'):
                proof.validate_live(candidate, receipt)

    def test_declared_path_removal_does_not_remove_protocol_obligation(self):
        for prerequisite in oracle.EXECUTION_PREREQUISITES:
            candidate = deepcopy(self.manifest)
            candidate['qualification_paths'].pop(str(prerequisite))
            with self.subTest(prerequisite=prerequisite), self.assertRaisesRegex(AssertionError, 'mandatory protocol path declarations'):
                proof.validate_plan(candidate)

    def test_every_finite_consumer_requires_actual_protocol_output(self):
        consumers = set(oracle.EXECUTION_CLASS_CONSUMERS) | set(oracle.SPELL_CONTRACTS)
        self.assertEqual(len(consumers), 72)
        for key in consumers:
            candidate = deepcopy(self.manifest)
            rows = proof.rows(candidate)
            rows[key]['depends'].remove('execution-dependencies')
            with self.subTest(consumer=key), self.assertRaisesRegex(AssertionError, 'finite protocol consumer roster'):
                oracle.validate_protocols(candidate, rows)

    def test_omitting_declared_paths_cannot_hide_native_root_removal(self):
        candidate = deepcopy(self.manifest)
        receipt = deepcopy(self.receipt)
        candidate['qualification_paths'] = {}
        graph = {x['number']: x for x in receipt['graph']}
        graph[337]['blocked_by'].remove(970)
        with self.assertRaisesRegex(AssertionError, 'mandatory execution path'):
            oracle.validate_native_obligations(candidate, graph, proof.rows(candidate))


if __name__ == '__main__':
    unittest.main()
