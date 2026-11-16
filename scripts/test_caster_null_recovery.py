"""Offline controls for ROOT's finite one-use recovery authorization."""
import json
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import Mock, patch

from caster_native_publication import Coordinator
from recover_caster_null_create import consume, identity, parse_http, matching, ALIAS


class RecoveryTests(unittest.TestCase):
    def test_sanitized_headers_and_exact_response_body(self):
        body='{"data":{"createIssue":{"issue":null}}}\n'
        raw='HTTP/2.0 200 OK\r\nX-GitHub-Request-Id: ABCD\r\nX-RateLimit-Remaining: 2100\r\nRetry-After: 60\r\nSet-Cookie: SECRET\r\nAuthorization: SECRET\r\n\r\n'+body
        result=parse_http(raw)
        self.assertEqual(result['status'],200)
        self.assertEqual(result['body'],body)
        self.assertEqual(result['headers'],{'x-github-request-id':'ABCD','x-ratelimit-remaining':'2100','retry-after':'60'})
        self.assertNotIn('SECRET',json.dumps(result))

    def test_authorization_is_consumed_before_any_replay(self):
        with TemporaryDirectory() as directory:
            path=Path(directory)/'consumed.json';event=dict(attempts=1,status='CONSUMED')
            consume(path,event);before=path.read_bytes()
            with self.assertRaises(FileExistsError):consume(path,dict(attempts=2))
            self.assertEqual(path.read_bytes(),before)

    def test_typed_identity_never_accepts_null_bool_or_decimal(self):
        self.assertEqual(identity(dict(number=2101,databaseId=90001,id='I_TYPED')),dict(number=2101,id=90001,node_id='I_TYPED'))
        for bad in [None,{},dict(number=True,databaseId=90001,id='I_TYPED'),dict(number=2101,databaseId=1.5,id='I_TYPED'),dict(number=2101,databaseId=90001,id='')]:
            with self.subTest(bad=bad),self.assertRaises(AssertionError):identity(bad)

    def test_existing_marker_requires_exact_full_body_and_unique_identity(self):
        body='<!-- edda-caster-native:'+ALIAS+' -->\nSource\n'
        node=dict(body=body,title='Exact',state='OPEN')
        self.assertEqual(matching([node],body,'Exact'),[node])
        for bad in [[node,node],[node|dict(body=body+'changed')],[node|dict(title='changed')],[node|dict(state='CLOSED')]]:
            with self.assertRaises(AssertionError):matching(bad,body,'Exact')

    def test_second_ambiguity_halts_all_normal_native_mutation_paths(self):
        with TemporaryDirectory() as directory:
            c=Coordinator.__new__(Coordinator);c.directory=Path(directory);c.calls=0
            (c.directory/'native-write-halt.json').write_text('{}')
            with patch('caster_native_publication.subprocess.run') as execute:
                with self.assertRaisesRegex(AssertionError,'native-write halt'):
                    c.api('mutation($input:CreateIssueInput!){createIssue(input:$input){issue{id}}}',{},'create')
                execute.assert_not_called()


if __name__=='__main__':unittest.main()
