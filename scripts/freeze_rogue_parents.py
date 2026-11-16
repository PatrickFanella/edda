#!/usr/bin/env python3
"""Freeze only ROOT's explicit accepted POST-RANGER release after exact fresh match."""
import hashlib
import json
from pathlib import Path
from verify_fighter_scope import api
ROOT=Path(__file__).resolve().parents[1]
MERGE='e7cd1ba8318ba296edd4fc5b6b84715ca11016f2'
HEAD='260c304f776d8b1dfe7b8ca1fd53e47842b8eec1'
# Exact user-released hashes,never derived from live bodies.
SEALS={'47': '2fca8c5b3b4a17570c860687e106b9d38682df80f37c76583d4b22bbbff59b03', '54': '8ccbaaf02ab940e918e7b427ec1f50dc9609b0a90c4eaf510e2c0d370c0ea08e', '283': 'f5063c9f5d6c19478d842c885af45901d6d6215d9e46ac6492fc40c048824949', '337': 'e81104c8fbb0fb607225271d765192a4eee3e2acc13aff881464408776062f85'}

def validate_release(r):
    assert r['accepted_ranger_merge']==MERGE and r['accepted_head']==HEAD
    assert set(r['body_seals'])==set(SEALS)
    for n,h in SEALS.items():
        x=r['body_seals'][n]
        assert x['sha256']==h==hashlib.sha256(x['body'].encode()).hexdigest()
        assert len(x['body'])==x['characters']
        assert '<!-- edda-329:start -->' not in x['body']

def main():
    r=json.loads(Path('/tmp/edda-rogue-accepted-parent-baselines.json').read_text());validate_release(r)
    q='{repository(owner:"subculture-collective",name:"edda"){'+''.join(f'i{n}:issue(number:{n})'+'{body}' for n in SEALS)+'}}'
    d=api('graphql',{'query':q});assert not d.get('errors'),d.get('errors')
    for n in SEALS:assert d['data']['repository'][f'i{n}']['body']==r['body_seals'][n]['body'],f'fresh released prefix match#{n}'
    p=ROOT/'docs/investigations/issue-329-accepted-parent-baselines.json';p.write_text(json.dumps(r,indent=2)+'\n')
    p=ROOT/'docs/investigations/issue-329-rogue-scope.json';m=json.loads(p.read_text());m['accepted_post_ranger']=MERGE;m['accepted_parent_baseline']=r
    m['publication']='NATIVE77Task publication verified; accepted POST-RANGER prefix frozen; ROOT exact appends/FINAL sealed proof PENDING'
    p.write_text(json.dumps(m,indent=2)+'\n')
    print('Four released full parent prefixes match fresh live bodies; frozen; NO parent writes')

if __name__=='__main__':main()
