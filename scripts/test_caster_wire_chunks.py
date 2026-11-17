"""Offline wire tests using all frozen nodes and the real durable API boundary."""
import copy
import json
import re
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import Mock, patch

from caster_native_publication import Coordinator, allocation_body, final_body, resolve, sha
import test_caster_native_publication as publication_tests


class WireChunksTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        publication_tests.PublicationTest.setUpClass()
        cls.frozen = publication_tests.PublicationTest.coordinator
        cls.registry = publication_tests.PublicationTest.registry
        cls.largest = max(cls.frozen.nodes, key=lambda n: len(n['blocked_by_refs']))
        assert len(cls.largest['blocked_by_refs']) == 80

    def coordinator(self, directory, partial=0):
        c = Coordinator.__new__(Coordinator)
        c.directory = Path(directory)
        c.nodes = self.frozen.nodes
        c.baseline = self.frozen.baseline
        c.contract = self.frozen.contract
        c.registry = copy.deepcopy(self.registry)
        from sorcerer_native_binding import pinned_inputs
        c.approval = pinned_inputs()[0]
        c.verify_inputs = Mock()
        c.registry_path = c.directory/'registry.json'
        raw = json.dumps(c.registry).encode()
        snapshot = c.directory/'ids.json'; snapshot.write_bytes(raw)
        (c.directory/'ids-allocated-milestone.json').write_text(json.dumps(dict(
            registry_path=str(snapshot), registry_file_sha256=sha(raw),
            nodes_map_sha256=sha(json.dumps(c.registry['nodes'], sort_keys=True, separators=(',', ':'), ensure_ascii=False)))))
        c.journal = dict(operations=[])
        c.journal_path = c.directory/'journal.json'
        c.calls = 0
        self.rows = {}
        for old in c.baseline['graph']:
            self.rows[old['number']] = dict(title=old['title'], body=old['body'], parent=dict(number=old['parent']) if old['parent'] else None,
                                          blockedBy=dict(nodes=[dict(number=n) for n in old['blocked_by']]))
        for e in c.approval['allowed_old_root_edges']:
            self.rows[int(e['issue'])]['blockedBy']['nodes'].append(dict(number=resolve(e['blocker'], c.registry)))
        for n in c.nodes:
            number = c.registry['nodes'][n['alias']]['number']
            deps = sorted(resolve(d, c.registry) for d in n['blocked_by_refs'])
            self.rows[number] = dict(title=n['title'], body=final_body(n,c.registry,c.contract), state='OPEN',
                                     parent=dict(number=resolve(n['parent_ref'],c.registry)), blockedBy=dict(nodes=[dict(number=d) for d in deps]))
        target = self.rows[c.registry['nodes'][self.largest['alias']]['number']]
        target['body'] = allocation_body(self.largest,c.contract)
        target['parent'] = None
        target['blockedBy']['nodes'] = target['blockedBy']['nodes'][:partial]
        self.identities = {n['number']: 'OLD_'+str(n['number']) for n in c.baseline['graph']}
        self.identities.update({n['number']: n['node_id'] for n in c.registry['nodes'].values()})
        self.by_id = {v:k for k,v in self.identities.items()}
        c.full_graph = Mock(return_value=[dict(number=k,node_id=v) for k,v in self.identities.items()])
        c.metadata = Mock(side_effect=lambda numbers: {n:copy.deepcopy(self.rows[n]) for n in numbers})
        return c

    def transport(self, query, variables):
        fields = re.findall(r'(\w+):(updateIssue|addSubIssue|addBlockedBy)\(input:\$(v\d+)\)',query)
        self.assertGreater(len(fields),0); self.assertLessEqual(len(fields),4)
        data = {}
        for field,operation,var in fields:
            value = variables[var]
            if operation == 'updateIssue':
                self.rows[self.by_id[value['id']]]['body'] = value['body']
            elif operation == 'addSubIssue':
                self.rows[self.by_id[value['subIssueId']]]['parent'] = dict(number=self.by_id[value['issueId']])
            else:
                self.rows[self.by_id[value['issueId']]]['blockedBy']['nodes'].append(dict(number=self.by_id[value['blockingIssueId']]))
            data[field] = dict(issue=dict(id='I_RESULT'))
        return dict(status=200,exit_code=0,headers={},body=json.dumps(dict(data=data)))

    def run_wire(self,c,transport):
        with patch('caster_mutation_guard.MutationGuard.wait') as pace, patch('caster_mutation_guard.http_graphql',side_effect=transport) as http, patch('builtins.print'):
            c.wire()
        return pace,http

    def test_largest_node_order_coverage_labels_and_no_source_changes(self):
        with TemporaryDirectory() as directory:
            c = self.coordinator(directory)
            old = copy.deepcopy({n['number']:self.rows[n['number']] for n in c.baseline['graph']})
            pace,http = self.run_wire(c,self.transport)
            self.assertEqual(http.call_count,21); self.assertEqual(pace.call_count,21)
            saved = json.loads(c.journal_path.read_text())['operations']
            self.assertEqual(len(saved),21)
            labels=[]; values=[]
            for e in saved:
                self.assertEqual(e['status'],'ACKNOWLEDGED')
                fields = re.findall(r'(\w+):(?:updateIssue|addSubIssue|addBlockedBy)\(',e['request']['query'])
                self.assertEqual(e['operation'],'wire '+self.largest['alias']+' fields '+','.join(fields))
                labels += fields; values += list(e['request']['variables'].values())
            self.assertEqual(labels,['body','parent']+['d'+str(i) for i in range(80)])
            self.assertEqual(values[0]['body'],final_body(self.largest,c.registry,c.contract))
            self.assertEqual(values[1]['issueId'],self.identities[resolve(self.largest['parent_ref'],c.registry)])
            self.assertEqual([self.by_id[v['blockingIssueId']] for v in values[2:]], sorted(resolve(d,c.registry) for d in self.largest['blocked_by_refs']))
            self.assertEqual(old,{n:self.rows[n] for n in old})
            # A new coordinator re-reads native partial/final state, not ACK assumptions.
            restarted = self.coordinator(directory)
            # Restore the transport's final state after fixture construction.
            self.rows[c.registry['nodes'][self.largest['alias']]['number']] = dict(title=self.largest['title'],state='OPEN',body=values[0]['body'],parent=dict(number=resolve(self.largest['parent_ref'],c.registry)),blockedBy=dict(nodes=[dict(number=self.by_id[v['blockingIssueId']]) for v in values[2:]]))
            _,http = self.run_wire(restarted,self.transport)
            http.assert_not_called()

    def test_partial_reload_sends_only_remaining_exact_edges(self):
        with TemporaryDirectory() as directory:
            c=self.coordinator(directory,partial=7)
            target=self.rows[c.registry['nodes'][self.largest['alias']]['number']]
            target.update(body=final_body(self.largest,c.registry,c.contract),parent=dict(number=resolve(self.largest['parent_ref'],c.registry)))
            known=copy.deepcopy(target['blockedBy']['nodes'])
            _,http=self.run_wire(c,self.transport)
            sent=[v for call in http.call_args_list for v in call.args[1].values()]
            self.assertEqual(len(sent),73)
            self.assertTrue(all(set(v)=={'issueId','blockingIssueId'} for v in sent))
            self.assertFalse({self.by_id[v['blockingIssueId']] for v in sent}&{n['number'] for n in known})

    def test_second_batch_502_stops_and_retains_first_ack(self):
        with TemporaryDirectory() as directory:
            c=self.coordinator(directory); calls=[]
            def http(query,variables):
                calls.append(query)
                if len(calls)==2: return dict(status=502,exit_code=1,headers={},body='<html>Bad Gateway</html>')
                return self.transport(query,variables)
            with self.assertRaises(json.JSONDecodeError): self.run_wire(c,http)
            self.assertEqual(len(calls),2)
            saved=json.loads(c.journal_path.read_text())['operations']
            self.assertEqual([e['status'] for e in saved],['ACKNOWLEDGED','RECONCILE_REQUIRED'])
            self.assertEqual(saved[-1]['sanitized_http_response']['status'],502)
            halt=c.directory/'native-write-halt.json'; original=halt.read_bytes()
            with self.assertRaisesRegex(AssertionError,'native-write halt'): self.run_wire(c,http)
            self.assertEqual(len(calls),2); self.assertEqual(halt.read_bytes(),original)

    def test_every_chunk_requires_nested_identity_and_durable_label(self):
        for failed in range(1,22):
            with self.subTest(chunk=failed), TemporaryDirectory() as directory:
                c=self.coordinator(directory); count=[0]
                def malformed(query,variables):
                    count[0]+=1
                    response=self.transport(query,variables)
                    if count[0]==failed:
                        data=json.loads(response['body'])
                        data['data'][next(iter(data['data']))]['issue']=None
                        response['body']=json.dumps(data)
                    return response
                with self.assertRaisesRegex(AssertionError,'missing typed mutation identity'): self.run_wire(c,malformed)
                self.assertEqual(count[0],failed)
                saved=json.loads(c.journal_path.read_text())['operations']
                self.assertEqual(saved[-1]['status'],'RECONCILE_REQUIRED')
                self.assertIn(' fields ',saved[-1]['operation'])

    def test_old_root_body_drift_still_refuses_without_writes(self):
        with TemporaryDirectory() as directory:
            c=self.coordinator(directory)
            n=self.largest; target=self.rows[c.registry['nodes'][n['alias']]['number']]
            target.update(body=final_body(n,c.registry,c.contract),parent=dict(number=resolve(n['parent_ref'],c.registry)),blockedBy=dict(nodes=[dict(number=resolve(d,c.registry)) for d in n['blocked_by_refs']]))
            root=int(c.approval['allowed_old_root_edges'][0]['issue'])
            self.rows[root]['body']+='drift'
            with patch('caster_mutation_guard.http_graphql') as http, patch('builtins.print'), self.assertRaises(AssertionError): c.wire()
            http.assert_not_called()


if __name__=='__main__': unittest.main()
