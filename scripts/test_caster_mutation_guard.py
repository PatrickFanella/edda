"""Offline clock/API/journal controls; never a native mutation probe."""
import copy
import json
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import Mock, patch

from caster_mutation_guard import MutationGuard, NOT_BEFORE, timestamp, EXTERNAL_STARTS
from caster_native_publication import Coordinator, old_contract_matches, sha, atomic, APPROVAL_SHA, RENDERER_SHA
import recover_caster_null_create as recovery


class GuardTests(unittest.TestCase):
    def test_cooldown_and_known_external_budget(self):
        guard=MutationGuard(dict(operations=[]),clock=lambda:NOT_BEFORE-61)
        with self.assertRaisesRegex(AssertionError,'cooldown'):guard.wait()
        self.assertGreaterEqual(guard.deadline(),max(EXTERNAL_STARTS)+7.5)

    def test_pacing_restart_and_ambiguous_full_hour(self):
        with TemporaryDirectory() as directory:
            path=Path(directory)/'journal.json'
            operations=[dict(started_at='2026-11-17T01:00:00Z',status='ACKNOWLEDGED')]
            atomic(path,dict(operations=operations))
            clock=[timestamp('2026-11-17T01:00:01Z')];sleeps=[]
            def sleep(delay):sleeps.append(delay);clock[0]+=delay
            MutationGuard(json.loads(path.read_text()),lambda:clock[0],sleep).wait()
            self.assertEqual(sleeps,[6.5])
            operations[0].update(status='RECONCILE_REQUIRED',finished_at='2026-11-17T01:00:03Z')
            self.assertEqual(MutationGuard(dict(operations=operations)).deadline(),timestamp('2026-11-17T02:00:03Z'))
            operations[0]['status']='INTENT'
            self.assertGreaterEqual(MutationGuard(dict(operations=operations)).deadline(),timestamp('2026-11-17T02:00:00Z'))

    def test_rolling_budget_counts_external_patches(self):
        # 477 other content writes + ROOT's three PATCHes exhaust 480.
        entries=[dict(started_at='2026-11-17T00:00:10Z',status='ACKNOWLEDGED') for _ in range(477)]
        self.assertGreater(MutationGuard(dict(operations=entries)).deadline(),EXTERNAL_STARTS[0]+3600)

    def test_retry_after_and_primary_reset_survive_restart(self):
        start=timestamp('2026-11-17T01:00:00Z')
        response=dict(headers={'retry-after':'120','x-ratelimit-remaining':'0','x-ratelimit-reset':str(start+300)})
        journal=dict(operations=[dict(started_at='2026-11-17T01:00:00Z',finished_at='2026-11-17T01:00:02Z',status='ACKNOWLEDGED',sanitized_http_response=response)])
        with TemporaryDirectory() as directory:
            path=Path(directory)/'journal';atomic(path,journal)
            self.assertEqual(MutationGuard(json.loads(path.read_text())).deadline(),start+301)
        response['headers']={'retry-after':'Tue, 17 Nov 2026 02:00:00 GMT'}
        self.assertEqual(MutationGuard(journal).deadline(),start+3600)
        response['headers']={'x-ratelimit-remaining':'0'}
        with self.assertRaises(AssertionError):MutationGuard(journal).deadline()

    def test_only_three_decisions_may_close_no_body_edge_or_parent_bypass(self):
        old=dict(number=330,state='open',title='Frozen',body='Source',parent=283,blocked_by=[1])
        self.assertTrue(old_contract_matches(old|dict(state='closed'),old))
        for field,value in [('body','drift'),('parent',54),('blocked_by',[])]:
            self.assertFalse(old_contract_matches(old|dict(state='closed')|{field:value},old))
        old['number']=329
        self.assertFalse(old_contract_matches(old|dict(state='closed'),old))

    def test_transport_failure_retains_sanitized_response_and_halts_restart(self):
        with TemporaryDirectory() as directory:
            c=Coordinator.__new__(Coordinator);c.directory=Path(directory);c.calls=0
            c.journal=dict(operations=[]);c.journal_path=c.directory/'journal.json'
            response=dict(status=200,exit_code=0,headers={'x-github-request-id':'TEST'},body='{"data":{"createIssue":null}}')
            with patch('caster_mutation_guard.MutationGuard.wait'),patch('caster_mutation_guard.http_graphql',return_value=response) as api:
                with self.assertRaises(AssertionError):c.api('mutation{createIssue{issue{id}}}',{},'create test')
                saved=json.loads(c.journal_path.read_text())
                self.assertEqual(saved['operations'][0]['sanitized_http_response'],response)
                self.assertEqual(saved['operations'][0]['status'],'RECONCILE_REQUIRED')
                with self.assertRaisesRegex(AssertionError,'native-write halt'):c.api('mutation{createIssue{issue{id}}}',{},'create test')
                self.assertEqual(api.call_count,1)

    def test_new_bounded_recovery_null_consumes_once_preserves_halt(self):
        self.run_recovery(existing=False)

    def test_new_bounded_recovery_adopts_exact_existing_without_create(self):
        self.run_recovery(existing=True)

    def run_recovery(self,existing):
        with TemporaryDirectory() as directory:
            root=Path(directory);retained=root/'recovery-20261116-null-create';retained.mkdir()
            for name in ['native-write-halt.json','registry.json','publication-journal.json',
                         'recovery-20261116-null-create/one-attempt-consumed.json',
                         'recovery-20261116-null-create/one-attempt-result.json']:
                (root/name).write_text('{}')
            halt=(root/'native-write-halt.json').read_bytes()
            c=Mock(directory=root,journal=dict(operations=[]),journal_path=root/'publication-journal.json',
                   nodes=[dict(alias=recovery.ALIAS,title='Frozen')],contract={},registry=dict(nodes={}))
            ident=dict(number=2000,databaseId=9000,id='I_NEW',title='Frozen',body='<!-- edda-caster-native:'+recovery.ALIAS+' -->Body',state='OPEN',
                       issueType=dict(name='Task'),parent=None,blockedBy=dict(nodes=[]),project_item=dict(fieldValueByName=dict(name='M')))
            c.metadata.return_value={2000:ident}
            auth=dict(status='ROOT_AUTHORIZED_ONE_RECONCILED_ATTEMPT',allowed_attempts=1,
                      approval_sha256=APPROVAL_SHA,renderer_contract_sha256=RENDERER_SHA,alias=recovery.ALIAS,
                      original_intent_request_sha256=recovery.INTENT,not_a_milestone=True,preserve_active_halt=True,
                      code_sha256={k:sha(Path(recovery.__file__).with_name(v).read_bytes()) for k,v in
                          dict(coordinator='caster_native_publication.py',recovery='recover_caster_null_create.py',guard='caster_mutation_guard.py').items()},
                      retained_sha256={name:sha(p.read_bytes()) for name in ['native-write-halt.json','registry.json','publication-journal.json',
                          'recovery-20261116-null-create/one-attempt-consumed.json','recovery-20261116-null-create/one-attempt-result.json'] for p in [root/name]})
            path=root/'auth.json';atomic(path,auth);digest=sha(path.read_bytes())
            found=[ident] if existing else []
            recent=dict(exit_code=0,status=200,headers={},body=json.dumps(dict(data=dict(repository=dict(issues=dict(totalCount=1,nodes=found))))))
            null=dict(exit_code=0,status=200,headers={},body='{"data":{"createIssue":{"issue":null}}}')
            responses=Mock(side_effect=[recent,null])
            original_sha=sha
            def payload_sha(value):
                if isinstance(value,str) and 'CreateIssueInput' in value:return recovery.INTENT
                return original_sha(value)
            with patch.object(recovery,'RECOVERY',retained),patch.object(recovery,'load_authorization'),patch.object(recovery,'reconcile',return_value=([{}],{2000:ident},found)),patch.object(recovery,'allocation_body',return_value=ident['body']),patch.object(recovery,'MutationGuard') as guard,patch.object(recovery,'http_graphql',responses),patch.object(recovery,'sha',side_effect=payload_sha):
                if existing:
                    result=recovery.recover(c,path,digest)
                    self.assertEqual(result['attempts'],0);c.save_registry.assert_called_once_with('ALLOCATING')
                else:
                    with self.assertRaises(AssertionError):recovery.recover(c,path,digest)
                    self.assertEqual(c.journal['operations'][-1]['status'],'RECONCILE_REQUIRED')
                    c.save_registry.assert_not_called()
                self.assertEqual(responses.call_count,1 if existing else 2)
                # A restart cannot reuse either the original retained hash or consumed token.
                with self.assertRaises(AssertionError):recovery.recover(c,path,digest)
                self.assertEqual(responses.call_count,1 if existing else 2)
            self.assertEqual((root/'native-write-halt.json').read_bytes(),halt)
            self.assertTrue((root/('recovery-'+digest+'.consumed.json')).exists())


if __name__=='__main__':unittest.main()
