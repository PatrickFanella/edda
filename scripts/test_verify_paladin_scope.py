"""Checkpoint proofs stay honest until native publication is complete."""
from copy import deepcopy
import json
import unittest
from pathlib import Path
import os
import verify_paladin_scope as proof
import verify_paladin_consumed_sources as consumed


class PaladinCheckpointTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest=json.loads(proof.MANIFEST.read_text())

    def test_complete_checkpoint_and_corruption_controls(self):
        self.assertEqual(proof.validate(self.manifest),87)
        proof.probes(self.manifest)

    def test_every_class_application_consumes_own_hp_adapter(self):
        rows={x['key']:x for x in self.manifest['issues']}
        self.assertEqual([x['key'] for x in rows.values() if 429 in x['prerequisites']],['hp'])
        for key in [f'level-{n}' for n in range(1,21)]+['later-1']:
            self.assertIn('hp',rows[key]['depends'])

    def test_spell_effects_are_not_membership(self):
        rows={x['key']:x for x in self.manifest['shared_issues']}
        self.assertIn(352,rows['bless-extension']['prerequisites'])
        self.assertIn(265,rows['detect-extension']['prerequisites'])
        self.assertIn(625,rows['curse-item']['prerequisites'])
        self.assertFalse(any(x['spell']=='Dispel Magic' for x in rows.values()))

    def test_unpublished_issue_id_cannot_be_advertised_as_receipt(self):
        m=deepcopy(self.manifest);m['issues'][0]['issue']=99999
        with self.assertRaises(AssertionError):proof.validate(m)

    def test_inherited_light_and_prepared_ritual_contracts(self):
        consumed.validate_inherited_contracts(self.manifest)
        controls={
            'light-extension':[
                ('NO ACTION','an Action'),('NOT Incapacitated','Incapacitated'),
                ('its own ongoing','any ongoing'),('does not end','ends'),
                ('even if willing','unless willing'),
                ('non-Incapacitated caster dismissal','standalone voluntary dismissal invented from264'),
            ],
            'detect-extension':[
                ('plus10minutes','plus1minute'),('no spell slot','one spell slot'),
                ('no extra Ritual Adept feature required','Ritual Adept feature required'),
                ('prepared Ritual-tag spell','unprepared Ritual-tag spell'),
                ('interruption fails without spending a slot','interruption spends a slot'),
            ],
        }
        for collection in ['shared_issues','published_issues']:
            for key,changes in controls.items():
                for old,new in changes:
                    m=deepcopy(self.manifest);x=next(x for x in m[collection] if x['key']==key)
                    self.assertIn(old,x['fixture']);x['fixture']=x['fixture'].replace(old,new)
                    with self.subTest(collection=collection,key=key,change=new),self.assertRaisesRegex(AssertionError,'inherited casting contract'):
                        consumed.validate_inherited_contracts(m)
        m=deepcopy(self.manifest)
        next(x for x in m['shared_issues'] if x['key']=='light-extension')['prerequisites'].remove(487)
        with self.assertRaisesRegex(AssertionError,'actual Incapacitated owner'):
            consumed.validate_inherited_contracts(m)

    @unittest.skipUnless(os.environ.get('EDDA_SRD_PDF'),'set EDDA_SRD_PDF for direct source controls')
    def test_general_casting_clauses_against_actual_pinned_pdf(self):
        blocks=consumed.source(Path(os.environ['EDDA_SRD_PDF']));consumed.validate(blocks)
        for clause in consumed.CASTING_CLAUSES:
            bad=dict(blocks);bad['__casting__']=bad['__casting__'].replace(clause,'REMOVED SOURCE RULE')
            with self.subTest(clause=clause),self.assertRaisesRegex(AssertionError,'general casting clause'):
                consumed.validate(bad)


if __name__=='__main__':unittest.main()
