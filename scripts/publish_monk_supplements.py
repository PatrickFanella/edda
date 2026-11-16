#!/usr/bin/env python3
"""Parent-only serial publication for #326. Default is read-only preflight."""
import argparse
import json
import subprocess
from pathlib import Path
from verify_monk_scope import ACCEPTED_PREFIX_SHA256, START, END, sha

ROOT=Path(__file__).resolve().parents[1]
ARTIFACT=ROOT/'docs/investigations/issue-326-parent-supplements.json'
REPO='repos/subculture-collective/edda'
def api(number,body=None):
 command=['gh','api',f'{REPO}/issues/{number}']
 if body is not None:command+=['-X','PATCH','--input','-']
 result=subprocess.run(command,input=json.dumps({'body':body}) if body is not None else None,text=True,capture_output=True,check=True)
 return json.loads(result.stdout)['body']
def plan(current,supplement,prefix):
 assert sha(prefix)==ACCEPTED_PREFIX_SHA256[str(supplement['issue'])]
 assert supplement['previous_sha256']==sha(prefix)
 assert current.startswith(prefix),'accepted prefix changed or rolled back'
 section=supplement['append']
 assert current.count(START)==current.count(END)==current.count(section)<=1,'partial, changed or duplicate owned section'
 desired=current if section in current else current+section
 assert len(desired)<=65536,'GitHub body capacity'
 assert desired.startswith(current) and desired.count(section)==1
 return desired
def main():
 parser=argparse.ArgumentParser(description=__doc__)
 parser.add_argument('--apply',action='store_true',help='Authorized root serial publication; writes only47/54/283/337')
 args=parser.parse_args()
 manifest=json.loads((ROOT/'docs/investigations/issue-326-monk-scope.json').read_text())
 supplements=json.loads(ARTIFACT.read_text())
 assert [s['issue'] for s in supplements]==[47,54,283,337]
 # Preflight every target before the first write; fetch each target again before
 # its serial PATCH so all unrelated appended prose is retained.
 for s in supplements:
  number=s['issue'];prefix=manifest['preserved'][str(number)]['body']
  assert (ROOT/s['path']).read_text()==s['append']
  current=api(number);desired=plan(current,s,prefix)
  print(f'#{number}: {len(current)} -> {len(desired)} characters; '+('already published' if desired==current else 'pending append'),flush=True)
 if not args.apply:return
 for s in supplements:
  number=s['issue'];prefix=manifest['preserved'][str(number)]['body']
  current=api(number);desired=plan(current,s,prefix)
  if desired!=current:
   assert api(number)==current,'concurrent body edit; stop for fresh serial preflight'
   api(number,desired)
  actual=api(number)
  assert actual.startswith(desired) and plan(actual,s,prefix)==actual,'publication readback'
  print(f'#{number}: published SHA256 {sha(actual)}',flush=True)
if __name__=='__main__':main()
