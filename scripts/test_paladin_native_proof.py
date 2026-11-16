"""Parent ownership cannot be rolled back or forged by appending a second section."""
import unittest
import json
from types import SimpleNamespace
from unittest.mock import patch
import verify_paladin_scope as entrypoint
import paladin_native_proof as native
from paladin_native_proof import validate_parent


class ParentPrefixTests(unittest.TestCase):
    prefix='Accepted immutable pre-Paladin body.\n'
    section='\n<!-- edda-327:start -->\nPaladin #845.\n<!-- edda-327:end -->'

    def test_exact_owned_section_and_future_unrelated_append(self):
        validate_parent(self.prefix,self.section,self.prefix+self.section)
        validate_parent(self.prefix,self.section,self.prefix+self.section+'\nUnrelated later class section.')

    def test_rollback_duplicate_marker_and_altered_second_section(self):
        for actual in [self.prefix,self.prefix+self.section+self.section,self.prefix+self.section+'\n<!-- edda-327:start -->altered<!-- edda-327:end -->',self.prefix+self.section.replace('edda-327:end','edda-327:changed'),'Changed '+self.prefix+self.section]:
            with self.subTest(actual=actual),self.assertRaises(AssertionError):
                validate_parent(self.prefix,self.section,actual)

    def test_default_entrypoint_rejects_stripping_published_sections(self):
        m=json.loads(entrypoint.MANIFEST.read_text())
        r=json.loads(native.RECEIPT.read_text())
        m['supplement_publication']='sealed'
        for node in r['graph']:
            if str(node['number']) in m['accepted_parent_baseline']:
                node['body']=m['accepted_parent_baseline'][str(node['number'])]['body']
        with patch.object(entrypoint,'MANIFEST',SimpleNamespace(read_text=lambda:json.dumps(m))),patch.object(native,'RECEIPT',SimpleNamespace(read_text=lambda:json.dumps(r))),patch('sys.argv',['verify_paladin_scope.py']):
            with self.assertRaisesRegex(AssertionError,'immutable parent prefix'):
                entrypoint.main()

    def test_owned_edge_cannot_be_weakened_even_with_matching_manifest_and_receipt(self):
        m=json.loads(entrypoint.MANIFEST.read_text());r=json.loads(native.RECEIPT.read_text())
        target=next(x for x in m['published_issues'] if x['key']=='level-7')
        target['blocked_by_issues'].pop()
        next(x for x in r['graph'] if x['number']==target['issue'])['blocked_by']=target['blocked_by_issues']
        with self.assertRaisesRegex(AssertionError,'immutable owned native contract drift'):
            native.validate(m,r,True)


if __name__=='__main__':unittest.main()
