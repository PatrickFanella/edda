#!/usr/bin/env python3
"""Root-only serial publication for decision328; default is read-only preflight."""
import argparse
import json
from pathlib import Path
import subprocess
from verify_ranger_scope import ACCEPTED_PREFIX_SHA256, MANIFEST, ROOT, sha, supplement_plan


def api(number, body=None):
    command=['gh','api',f'repos/subculture-collective/edda/issues/{number}']
    if body is not None:command+=['-X','PATCH','--input','-']
    result=subprocess.run(command,input=json.dumps({'body':body}) if body is not None else None,
                          text=True,capture_output=True,check=True)
    return json.loads(result.stdout)['body']


def plan(manifest, supplement, current):
    n=str(supplement['issue'])
    baseline=manifest['accepted_parent_prefix'][n]
    assert baseline['sha256']==ACCEPTED_PREFIX_SHA256[n]
    assert (ROOT/supplement['path']).read_text()==supplement['append']
    return supplement_plan(current,supplement,baseline['prefix'],ACCEPTED_PREFIX_SHA256[n])


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--apply',action='store_true',help='For authorized parent serial publication only; writes47/54/283/337')
    args=parser.parse_args()
    manifest=json.loads(MANIFEST.read_text())
    assert manifest['accepted_parent_prefix'] is not None,'root accepted prefix release pending'
    assert [s['issue'] for s in manifest['supplements']]==[47,54,283,337]
    # Preflight every body before the first mutation. No parent append is applied
    # by the Ranger child; the root invokes --apply after its own review.
    for item in manifest['supplements']:
        current=api(item['issue']);desired=plan(manifest,item,current)
        print(f"#{item['issue']}: {len(current)} -> {len(desired)} characters; "
              +('already published' if current==desired else 'root append pending')
              +f"; current SHA256 {sha(current)}",flush=True)
    if not args.apply:return
    for item in manifest['supplements']:
        current=api(item['issue']);desired=plan(manifest,item,current)
        if desired!=current:
            assert api(item['issue'])==current,'concurrent parent body edit; stop for fresh preflight'
            api(item['issue'],desired)
        actual=api(item['issue'])
        assert actual.startswith(desired) and plan(manifest,item,actual)==actual,'parent publication readback'
        print(f"#{item['issue']}: published SHA256 {sha(actual)}",flush=True)


if __name__=='__main__':main()
