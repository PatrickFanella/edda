#!/usr/bin/env python3
"""Offline regression probes for complete terminal-gate readback."""
import unittest
from unittest.mock import patch
from urllib.parse import parse_qs, urlsplit

import verify_effect_lifecycle_scope as scope


class BlockerPagination(unittest.TestCase):
    def run_fixture(self, repeat):
        calls = []

        def api(path, data=None):
            calls.append(path)
            # A broken fixture must fail promptly instead of looping.
            self.assertLessEqual(len(calls), 3)
            if path == 'graphql':
                return {'data': {'repository': {'issues': {
                    'pageInfo': {'hasNextPage': False},
                    'nodes': [{'number': 337, 'parent': {'number': 283},
                               'blockedBy': {'pageInfo': {'hasNextPage': True},
                                             'nodes': []}}]}}}}
            page = int(parse_qs(urlsplit(path).query)['page'][0])
            if page == 1 or repeat:
                return [{'id': n, 'number': n} for n in range(1, 101)]
            self.assertEqual(page, 2)
            return [{'id': 101, 'number': 101}]

        with patch.object(scope, 'api', api):
            return scope.live({'issues': []})

    def test_retains_terminal_after_first_hundred(self):
        receipt = self.run_fixture(False)
        self.assertEqual(receipt['graph'][0]['blocked_by'], list(range(1, 102)))

    def test_rejects_repeated_page(self):
        with self.assertRaisesRegex(AssertionError, 'repeated blocker page'):
            self.run_fixture(True)


if __name__ == '__main__':
    unittest.main()
