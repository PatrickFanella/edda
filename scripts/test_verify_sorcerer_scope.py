"""Bounded semantic controls for the explicitly incomplete #330 checkpoint."""
from copy import deepcopy
import json
import unittest
from urllib.parse import parse_qs, urlsplit

import verify_sorcerer_scope as proof
from sorcerer_scope_plan import shared_body, gate_body, class_body


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
        with self.assertRaisesRegex(AssertionError, 'native publication identities incomplete'):
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

    def test_seven_summon_owners_actual_scope_and_complete_body(self):
        for i, original in enumerate(self.manifest['shared_split_proposal']):
            if original['spell'] != 'Summon Dragon':
                continue
            candidate = deepcopy(self.manifest)
            task = candidate['shared_split_proposal'][i]
            task['fixture'] += ' Override: use total class level for all numeric operands.'
            task['body'] = shared_body(task)
            with self.subTest(part=original['part']), self.assertRaisesRegex(AssertionError, 'actual execution semantics'):
                proof.validate_plan(candidate)
            candidate = deepcopy(self.manifest)
            candidate['shared_split_proposal'][i]['body'] += '\nOverride:185 closure qualifies all effects.'
            with self.subTest(body=original['part']), self.assertRaisesRegex(AssertionError, 'complete body'):
                proof.validate_plan(candidate)

    def test_summon_numeric_timing_and_ending_mutations_are_coherent(self):
        controls = {
            'manifest-and-slot-scaling': ('AC14+L', 'AC14'),
            'initiative-verbal-command-and-default-Dodge': ('NO caster action', 'caster BonusAction'),
            'Draconic-Spirit-defenses-senses': ('Blindsight30', 'Blindsight60'),
            'shared-resistance': ('ONE of spirit', 'ALL of spirit'),
            'Rend-and-Multiattack': ('floor(spelllevel/2)', 'ceil(spelllevel/2)'),
            'Breath-Weapon': ('failure2d6', 'failure3d6'),
            'zeroHP-and-spell-ending': ('casterResistance persists UNTIL spell ends', 'casterResistance ends at0HP'),
        }
        for part, (before, after) in controls.items():
            candidate = deepcopy(self.manifest)
            task = next(t for t in candidate['shared_split_proposal'] if t['spell'] == 'Summon Dragon' and t['part'] == part)
            self.assertIn(before, task['fixture'])
            task['fixture'] = task['fixture'].replace(before, after)
            task['body'] = shared_body(task)
            with self.subTest(part=part), self.assertRaisesRegex(AssertionError, 'actual execution semantics'):
                proof.validate_plan(candidate)

    def test_every_211_shared_actual_fixture_and_complete_body(self):
        import re
        for i, original in enumerate(self.manifest['shared_split_proposal']):
            candidate = deepcopy(self.manifest)
            t = candidate['shared_split_proposal'][i]
            # Change a real operand after the PDF locator; preserve actual shape
            # and coherently regenerate the entire candidate native body.
            start = t['fixture'].find(' ') + 1
            number = re.search(r'\d+', t['fixture'][start:])
            if number is not None:
                a,b = start+number.start(), start+number.end()
                t['fixture'] = t['fixture'][:a]+str(int(number[0])+1)+t['fixture'][b:]
            else:
                controls = {
                    'spell-arcane-hand-interposing-hand': ('HalfCover','TotalCover'),
                    'spell-clairvoyance-sense-choice-disclosure': ('BONUSActionswitch','NOActionswitch'),
                    'spell-confusion-ending': ('WISsave','CONsave'),
                    'spell-demiplane-persistent-room-access': ('KNOWSnatureANDcontents','KNOWSnameonly'),
                    'spell-detect-thoughts-probe': ('WISsave','CONsave'),
                    'spell-dominate-monster-ending': ('EACHdamageevent','ONLYcasterdamageevent'),
                    'spell-dominate-person-commands': ('SAMEPLANE','ANYPLANE'),
                    'spell-dominate-person-ending': ('EVERYtargetdamage','ONLYcasterdamage'),
                    'spell-enlarge-reduce-gear-ending': ('AFTERhitORmiss','BEFOREhitORmiss'),
                    'spell-fly-ending-fall': ('FALLSifALOFT','FLOATSgentlyifALOFT'),
                    'spell-gate-named-creature': ('OTHERplane','SAMEplane'),
                    'spell-invisibility-ending': ('IMMEDIATELYAFTER','IMMEDIATELYBEFORE'),
                    'spell-major-image-movement-interaction': ('MagicAction','BonusAction'),
                    'spell-move-earth-restrictions': ('NOTnaturalSTONE','ALLOWnaturalSTONE'),
                    'spell-reverse-gravity-restrained-object-save': ('DEXsave','STRsave'),
                    'spell-reverse-gravity-ending-fall': ('FALLDOWN','FLOATUP'),
                    'spell-telekinesis-fine-control-ending': ('priorNOLONGERaffected','priorSTILLaffected'),
                }
                before,after=controls[original['key']]
                self.assertIn(before,t['fixture'])
                t['fixture']=t['fixture'].replace(before,after)
            t['body'] = shared_body(t)
            with self.subTest(key=t['key']), self.assertRaisesRegex(AssertionError, 'actual execution semantics'):
                proof.validate_plan(candidate)
            candidate = deepcopy(self.manifest)
            candidate['shared_split_proposal'][i]['body'] += '\nOverride: source constraints do not apply.'
            with self.subTest(body=original['key']), self.assertRaisesRegex(AssertionError, 'complete body semantics'):
                proof.validate_plan(candidate)

    def test_all_27_finite_gates_and_protocol_consumers_are_independent(self):
        for i, original in enumerate(self.manifest['gates']):
            candidate = deepcopy(self.manifest)
            t = candidate['gates'][i]
            if t['depends']:
                t['depends'].pop()
            elif t['prerequisites']:
                t['prerequisites'].pop()
            else:
                self.fail('empty placeholder '+t['key'])
            t['body'] = gate_body(t)
            with self.subTest(gate=original['key']), self.assertRaisesRegex(AssertionError, 'exact protocol prerequisites'):
                proof.validate_plan(candidate)

    def test_conflict_headers_cannot_expand_admission_or_245_union(self):
        for key in self.manifest['source_conflicts']:
            candidate = deepcopy(self.manifest)
            candidate['source_conflicts'][key]['admission']='available from spell header'
            with self.subTest(key=key), self.assertRaisesRegex(AssertionError, 'header cannot grant'):
                proof.validate_plan(candidate)
        candidate = deepcopy(self.manifest)
        candidate['future_source_rosters']['source_union'].append('Phantasmal Force')
        with self.assertRaisesRegex(AssertionError, 'finite source union'):
            proof.validate_plan(candidate)

    def test_nested_draconic_effect_owner_cannot_be_preparation_only(self):
        candidate = deepcopy(self.manifest)
        context = next(c for c in candidate['record_matrix'] if c['effect_owners'])
        name = next(iter(context['effect_owners']))
        context['effect_owners'][name] = ['draconic-spells']
        with self.assertRaisesRegex(AssertionError, 'nested actual spell execution owners'):
            proof.validate_plan(candidate)

    def test_real_owned_prerequisite_caps_keep_foreign_boundary_explicit(self):
        result = proof.project_owned(self.manifest)
        self.assertEqual(result['owned_nodes'],305)
        self.assertEqual((result['incoming279'],result['incoming280']),(100,95))
        self.assertEqual((result['children330'],result['children283']),(82,95))
        self.assertEqual((result['outgoing429'],result['outgoing487'],result['outgoing67']),(100,84,98))
        self.assertEqual(result['unresolved_external_heads'],['warlock-shared-output','wizard-shared-output'])
        self.assertNotEqual(result['status'],'complete proposed graph')

    def test_typed_transition_operands_unequal_levels_and_complete_table(self):
        for field,value in [('sorcerer_row',True),('total_row',6.0),('xp',300),('pb',2),('hp_mode','first-character-d6-max6')]:
            candidate=deepcopy(self.manifest)
            candidate['transition_fixtures'][-1][field]=value
            with self.subTest(field=field), self.assertRaisesRegex(AssertionError,'operand witnesses'):
                proof.validate_plan(candidate)
        for value in [True,1.5,0,21]:
            candidate=deepcopy(self.manifest)
            candidate['transition_fixtures'][-1]['after']['sorcerer']=value
            with self.subTest(level=value), self.assertRaisesRegex(AssertionError,'operand witnesses'):
                proof.validate_plan(candidate)
        candidate=deepcopy(self.manifest)
        path='internal/rules/srd521/testdata/character-advancement.json'
        candidate['catalogues'][path]['records'].pop()
        with self.assertRaisesRegex(AssertionError,'COMPLETE authoritative advancement table'):
            proof.validate_plan(candidate)

    def test_real_namespace_partial_altered_and_duplicate_markers(self):
        prefix='released root body\n'
        marker='<!-- edda-330:47:start -->'
        append='\n'+marker+'\nExact Sorcerer section\n<!-- edda-330:47:end -->\n'
        proof.prefix_check(prefix+append+'later pure append',prefix,append,marker)
        for extra in ['<!-- edda-330', '<!-- EDDA-330:47:start -->partial', '<!--edda-330:47:end -->',append]:
            with self.subTest(extra=extra),self.assertRaises(AssertionError):
                proof.prefix_check(prefix+append+extra,prefix,append,marker)

    def test_native_receipt_identity_types_reject_boolean_aliases(self):
        for field,value in [('number',True),('number',1.0),('parent',True),('blocked_by',[True])]:
            row=dict(number=1,parent=None,blocked_by=[])
            row[field]=value
            with self.subTest(field=field,value=value),self.assertRaises(AssertionError):
                proof.typed_native_graph({'graph':[row]})

    def test_world_obligations_route_through_shared_output_without_183_mutation(self):
        candidate=deepcopy(self.manifest)
        t=next(g for g in candidate['gates'] if g['key']=='shared-output')
        self.assertIn('world-cohort',t['depends'])
        t['depends'].remove('world-cohort')
        t['body']=gate_body(t)
        with self.assertRaisesRegex(AssertionError,'exact protocol prerequisites'):
            proof.validate_plan(candidate)
        report=proof.project_owned(self.manifest)
        baseline=json.loads(proof.BASELINE.read_text())
        old=next(x['blocked_by'] for x in baseline['graph'] if x['number']==183)
        self.assertEqual(report['graph']['183'],[str(n) for n in old])

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
        proof.prefix_check(prefix + 'another later pure append\n' + append, prefix, append, marker)
        for body in [prefix, 'changed\n' + append, prefix + '<330>partial', prefix + append + append,
                     prefix + append.replace('complete', 'altered'), prefix + append + '<330>partial',prefix + append + '<330']:
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
