"""Offline contrary controls for the frozen mechanical publication boundary."""
import copy
import json
from pathlib import Path
from tempfile import TemporaryDirectory
from types import SimpleNamespace
import unittest
from unittest.mock import Mock

from caster_native_publication import (
    APPROVAL_SHA, RENDERER_SHA, Coordinator, allocation_body,
    final_body, check_registry, caps, uncertain_creation,
)


class PublicationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        from sorcerer_native_binding import pinned_inputs, evidence
        approval, baseline = pinned_inputs()
        cls.coordinator = SimpleNamespace(nodes=approval['approved_nodes'],baseline=baseline,
                                          contract=json.loads(evidence('native-renderer-contract.json')))
        cls.vectors = json.loads(evidence('native-renderer-vectors.json'))
        cls.registry = dict(schema_version=1, approval_sha256=APPROVAL_SHA,
                            renderer_contract_sha256=RENDERER_SHA, status='IDS_ALLOCATED',
                            nodes={n['alias']: dict(number=10000+i, id=20000+i, node_id='I_SYNTHETIC_'+str(i))
                                   for i, n in enumerate(cls.coordinator.nodes)})

    def test_root_literal_renderer_vectors(self):
        for case in self.vectors['cases']:
            self.assertEqual(allocation_body(case['approved_node'], self.coordinator.contract), case['allocation_body'])
            actual = final_body(case['approved_node'], self.vectors['example_registry'], self.coordinator.contract)
            self.assertEqual(actual, case['final_body'])
            changed = copy.deepcopy(case['approved_node'])
            changed['approved_body'] = changed['approved_body'].rstrip()
            if changed != case['approved_node']:
                self.assertNotEqual(final_body(changed, self.vectors['example_registry'], self.coordinator.contract), actual)

    def test_full_registry_and_independently_invalid_identities(self):
        check_registry(self.registry, self.coordinator.nodes, self.coordinator.baseline, True)
        first, second = list(self.registry['nodes'])[:2]
        for field, value in [('number', True), ('number', 2101.0), ('number', 0), ('id', False), ('id', 1.5), ('node_id', '')]:
            with self.subTest(field=field, value=value):
                changed = copy.deepcopy(self.registry); changed['nodes'][first][field] = value
                with self.assertRaises(AssertionError):
                    check_registry(changed, self.coordinator.nodes, self.coordinator.baseline, True)
        for key in ['number', 'id', 'node_id']:
            changed = copy.deepcopy(self.registry); changed['nodes'][second][key] = changed['nodes'][first][key]
            with self.assertRaises(AssertionError):
                check_registry(changed, self.coordinator.nodes, self.coordinator.baseline, True)
        changed = copy.deepcopy(self.registry); changed['nodes'][first]['number'] = 330
        with self.assertRaises(AssertionError):
            check_registry(changed, self.coordinator.nodes, self.coordinator.baseline, True)

    def test_partial_vectors_and_coherent_wrong_pins_unavailable(self):
        with self.assertRaises(AssertionError):
            check_registry(self.vectors['example_registry'], self.coordinator.nodes, self.coordinator.baseline, True)
        for field in ['approval_sha256', 'renderer_contract_sha256']:
            changed = copy.deepcopy(self.registry); changed[field] = '0'*64
            with self.assertRaises(AssertionError):
                check_registry(changed, self.coordinator.nodes, self.coordinator.baseline, True)
        changed = copy.deepcopy(self.registry); changed['nodes'].pop(next(iter(changed['nodes'])))
        with self.assertRaises(AssertionError):
            check_registry(changed, self.coordinator.nodes, self.coordinator.baseline, True)

    def test_reference_resolution_no_unknown_or_duplicate_fallback(self):
        node = copy.deepcopy(self.vectors['cases'][0]['approved_node'])
        node['blocked_by_refs'] = ['unknown:future']
        with self.assertRaises(KeyError):
            final_body(node, self.vectors['example_registry'], self.coordinator.contract)
        node['blocked_by_refs'] = ['67', '067']
        with self.assertRaises(AssertionError):
            final_body(node, self.vectors['example_registry'], self.coordinator.contract)

    def test_ambiguous_creation_never_retries_without_readable_identity(self):
        for status in ['INTENT','RECONCILE_REQUIRED','ACKNOWLEDGED']:
            self.assertTrue(uncertain_creation({'operations':[dict(operation='create 330:hp',status=status)]},'330:hp'))
        self.assertFalse(uncertain_creation({'operations':[dict(operation='size-M 330:hp',status='RECONCILE_REQUIRED')]},'330:hp'))
        self.assertFalse(uncertain_creation({'operations':[]},'330:hp'))

    def fake_allocator(self):
        c=Coordinator.__new__(Coordinator)
        c.nodes=self.coordinator.nodes
        c.baseline=self.coordinator.baseline
        c.contract=self.coordinator.contract
        c.registry=dict(schema_version=1,approval_sha256=APPROVAL_SHA,renderer_contract_sha256=RENDERER_SHA,status='ALLOCATING',nodes={})
        c.journal=dict(operations=[])
        c.verify_inputs=Mock()
        c.full_graph=Mock(return_value=copy.deepcopy(c.baseline['graph']))
        c.metadata=Mock(return_value={})
        c.save_registry=Mock(side_effect=AssertionError('must not save an incomplete creation identity'))
        return c

    def test_ordinary_allocate_restart_cannot_cross_unresolved_creation_fence(self):
        c=self.fake_allocator()
        c.journal['operations']=[dict(operation='create '+c.nodes[0]['alias'],status='RECONCILE_REQUIRED')]
        c.api=Mock(side_effect=AssertionError('must not issue a second create'))
        with self.assertRaisesRegex(AssertionError,'ambiguous prior creation'):
            c.allocate()
        c.api.assert_not_called()
        self.assertEqual(c.registry['nodes'],{})

    def test_null_success_and_malformed_identity_are_durably_fenced(self):
        for issue in [None,{},dict(number=True,databaseId=123,id='I_TEST'),dict(number=2101,databaseId=1.5,id='I_TEST')]:
            with self.subTest(issue=issue),TemporaryDirectory() as directory:
                c=self.fake_allocator();c.directory=Path(directory);c.journal_path=Path(directory)/'journal.json'
                response=dict(createIssue=dict(issue=issue))
                def acknowledged(query,variables,operation):
                    c.journal['operations'].append(dict(operation=operation,status='ACKNOWLEDGED',response=dict(data=response)))
                    return response
                c.api=Mock(side_effect=acknowledged)
                with self.assertRaisesRegex(RuntimeError,'no complete typed identity'):
                    c.allocate()
                saved=json.loads(c.journal_path.read_text())
                self.assertEqual(saved['operations'][-1]['status'],'RECONCILE_REQUIRED')
                self.assertEqual(saved['operations'][-1]['response'],dict(data=response))
                self.assertEqual(c.registry['nodes'],{})
                with self.assertRaisesRegex(AssertionError,'ambiguous prior creation'):
                    c.allocate()
                self.assertEqual(c.api.call_count,1)

    def test_both_dags_and_three_independent_caps(self):
        graph = {str(i): [] for i in range(105)}
        graph.update({'279': [], '280': [], '337': [], '429': [], '285': [], '283': []})
        caps(graph, {})
        for change in ['incoming', 'outgoing', 'children', 'dependency-cycle', 'containment-cycle', 'missing']:
            g = copy.deepcopy(graph); parents = {}
            if change == 'incoming': g['0'] = [str(i) for i in range(1, 102)]
            if change == 'outgoing':
                for i in range(1, 102): g[str(i)] = ['0']
            if change == 'children': parents = {str(i): '0' for i in range(1, 102)}
            if change == 'dependency-cycle': g['0'] = ['1']; g['1'] = ['0']
            if change == 'containment-cycle': parents = {'0': '1', '1': '0'}
            if change == 'missing': g['0'] = ['missing']
            with self.subTest(change=change), self.assertRaises(AssertionError): caps(g, parents)


if __name__ == '__main__':
    unittest.main()
