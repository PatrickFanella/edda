"""Future qualification outputs can be added without erasing prior obligations."""
from copy import deepcopy
import json
import unittest

import rogue_native_proof as native
import verify_rogue_scope as entry


class SharedRootCompatibilityTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest = json.loads(entry.PATH.read_text())
        cls.receipt = json.loads(native.RECEIPT.read_text())

    def add_output(self, receipt, target):
        number = 990000 + target
        receipt['graph'].append(dict(
            number=number, title='Future finite qualification output', body='',
            state='open', parent=277, blocked_by=[],
        ))
        node = next(x for x in receipt['graph'] if x['number'] == target)
        node['blocked_by'] = sorted(node['blocked_by'] + [number])
        return node

    def test_baseline_and_future_additions_to_each_shared_root(self):
        native.validate(self.manifest, self.receipt, True)
        for target in [279, 280, 337]:
            receipt = deepcopy(self.receipt)
            self.add_output(receipt, target)
            with self.subTest(target=target):
                native.validate(self.manifest, receipt, True)

    def test_combined_future_roots_and_closed_decision(self):
        receipt = deepcopy(self.receipt)
        for target in [279, 280, 337]:
            self.add_output(receipt, target)
        next(x for x in receipt['graph'] if x['number'] == 329)['state'] = 'closed'
        native.validate(self.manifest, receipt, True)

    def test_every_existing_root_requirement_survives_decision_closure(self):
        for target in [279, 280, 337]:
            original = next(x for x in self.receipt['graph'] if x['number'] == target)
            for edge in original['blocked_by']:
                receipt = deepcopy(self.receipt)
                next(x for x in receipt['graph'] if x['number'] == 329)['state'] = 'closed'
                next(x for x in receipt['graph'] if x['number'] == target)['blocked_by'].remove(edge)
                with self.subTest(target=target, edge=edge), self.assertRaises(AssertionError):
                    native.validate(self.manifest, receipt, True)

    def test_addition_does_not_authorize_shared_body_or_title_changes(self):
        for target in [279, 280]:
            receipt = deepcopy(self.receipt)
            self.add_output(receipt, target)['body'] += ' changed'
            with self.subTest(target=target), self.assertRaisesRegex(AssertionError, 'shared qualification body'):
                native.validate(self.manifest, receipt, True)
        for target in [279, 280, 337]:
            receipt = deepcopy(self.receipt)
            self.add_output(receipt, target)['title'] += ' changed'
            with self.subTest(target=target), self.assertRaisesRegex(AssertionError, 'shared qualification title'):
                native.validate(self.manifest, receipt, True)

    def test_other_protected_nodes_cannot_acquire_arbitrary_prerequisites(self):
        receipt = deepcopy(self.receipt)
        self.add_output(receipt, 334)
        with self.assertRaises(AssertionError):
            native.validate(self.manifest, receipt, True)


if __name__ == '__main__':
    unittest.main()
