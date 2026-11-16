"""Offline contrary controls for the frozen mechanical publication boundary."""
import copy
import json
import unittest

from caster_native_publication import (
    APPROVAL_SHA, RENDERER_SHA, DEFAULT, Coordinator, allocation_body,
    final_body, check_registry, caps,
)


class PublicationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.coordinator = Coordinator(DEFAULT)
        cls.vectors = json.loads((DEFAULT / 'native-renderer-vectors.json').read_text())
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
