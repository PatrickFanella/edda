#!/usr/bin/env python3
"""Bind #330's frozen source plan to VERIFIED native evidence, locally only.

Never publishes parent supplements, edits GitHub, regenerates source contracts,
or claims that an open implementation Task grants executable behavior.
"""
import gzip
import hashlib
import json
from pathlib import Path

from sorcerer_native_binding import (
    APPROVAL, RENDERER, DIRECTORY, pinned_inputs, registry_check, validate_evidence,
)
from verify_sorcerer_scope import ROOT, MANIFEST, ACCEPTED_PREFIX_SHA256, validate_plan


def write(path, data):
    path.write_text(json.dumps(data, indent=2, ensure_ascii=False)+'\n')


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def main():
    approval, baseline = pinned_inputs()
    registry = json.loads(Path(approval['registry']).read_text())
    assert registry['status']=='VERIFIED'
    map_sha = registry_check(registry,approval,baseline)
    verified = json.loads((DIRECTORY/'verified-milestone.json').read_text())
    allocated = json.loads((DIRECTORY/'ids-allocated-milestone.json').read_text())
    assert verified['nodes_map_sha256']==allocated['nodes_map_sha256']==map_sha
    assert verified['approval_sha256']==APPROVAL and verified['renderer_contract_sha256']==RENDERER
    for delivery in [allocated, verified]:
        raw=Path(delivery['registry_path']).read_bytes()
        assert sha(raw)==delivery['registry_file_sha256']
        archived=json.loads(raw)
        assert registry_check(archived,approval,baseline)==map_sha
    raw_receipt=Path(verified['verified_receipt_path']).read_bytes()
    assert sha(raw_receipt)==verified['verified_receipt_sha256']
    receipt=json.loads(raw_receipt)
    assert receipt['observed_at']==verified['verified_receipt_observed_at']
    for key in ['all_new_tasks_open','native_body_parent_prerequisite_Task_Size_validation','both_DAGs','all_three_caps','old_protections']:
        assert verified[key] is True
    manifest=json.loads(Path(approval['inputs']['330']['frozen_path']).read_text())
    for row in manifest['tasks']+manifest['shared_split_proposal']+manifest['gates']:
        row['issue']=registry['nodes']['330:'+row['key']]['number']
    validate_plan(manifest)
    result=validate_evidence(manifest,receipt,approval,baseline,registry)
    release=json.loads(Path('/tmp/edda-sorcerer-accepted-parent-baselines.json').read_text())
    prefixes={}
    for key, expected in ACCEPTED_PREFIX_SHA256.items():
        body=release['body_seals'][key]['body']
        assert sha(body.encode())==expected
        prefixes[key]=dict(prefix=body,sha256=expected)
    sections={
        '47': ('Sorcerer #330: 305 Tasks; runtime unavailable.', '<!-- edda-330 -->',70),
        '54': ('Sorcerer #330 reconciles 103 source records, 278 mechanic/option contexts and all20 class rows. Its305 Task/M contracts (67 class,211 shared,27 qualification) are natively published and remain OPEN. One frozen coordinator also published331/332 contracts through the exact three-output aggregate to279/280. Source-only decisions and Task closure confer no runtime support. Printed-list/header conflicts and unbound variants remain unavailable.', '<!-- edda-330 -->',750),
        '283': ('Sorcerer #330 keeps qualification under330; no new283 child. Separate predicates, creation/proficiencies/equipment/choices, D6/fixed4 HP via968, multiclass/slots, subclass, repeated improvements, resources/recovery and all20 adjacent applications retain distinct owners. Exact source285 and487 outputs use finite qualification-only gateways. All305 Tasks are OPEN; class/runtime support remains unavailable until actual implementation outputs qualify.', '<!-- edda-330 -->',750),
        '337': ('Sorcerer #330 decision closure cannot bypass implementation gates. Every exact class/level output reaches own qualification; every shared output also reaches finite cohorts, its class-shared head and the330/331/332 aggregate to279/280. Frozen source bodies, real IDs/parents/prerequisites/Task/M, both DAGs, all3 caps and historical paths passed native readback. No183 mutation, new429 consumer or283 child. Runtime, production grants, deploy and FullRules remain unavailable.', '<!-- edda-330 -->',750),
    }
    appends={}
    for key,(text,marker,limit) in sections.items():
        section='\n\n'+marker+'\n'+text+'\n'
        assert len(section)<=limit
        appends[key]=dict(status='PENDING ROOT-only serial publication',limit=limit,marker=marker,text=section,sha256=sha(section.encode()))
    manifest.update(status='PUBLISHED',native_plan_status='All305 source-contract Tasks published; runtime UNQUALIFIED; ROOT parent supplements/final seal PENDING',
                    accepted_parent_prefix=prefixes,parent_appends=appends,
                    native_publication=dict(**result,global_milestone=verified,final_parent_seal='PENDING ROOT-only supplements'))
    manifest['root_prefix_release'].update(native_publication_approval=True,final_seal=False)
    write(MANIFEST,manifest)
    out=ROOT/'docs/investigations'
    write(out/'issue-330-parent-appends.json',appends)
    (out/'issue-330-link-verification.json').write_bytes(raw_receipt)
    write(out/'issue-330-native-publication.json',dict(status='VERIFIED mechanical publication; parent seal PENDING',
          own_issues={a:v for a,v in registry['nodes'].items() if a.startswith('330:')},
          global_nodes_map_sha256=map_sha,ids_allocated=allocated,verified=verified,independent_own_validation=result))
    names=['approval.json','330-approved-proposal.json','331-approved-proposal.json','332-approved-proposal.json',
           'native-baseline.json','approved-projection.json','independent-review-receipt.json','native-renderer-contract.json',
           'native-renderer-vectors.json','milestone-provenance-contract.json','ids-allocated-milestone.json','verified-milestone.json']
    files={name:(DIRECTORY/name).read_text() for name in names}
    files['registry.json']=Path(verified['registry_path']).read_text()
    files['IDS_ALLOCATED.registry.json']=Path(allocated['registry_path']).read_text()
    files['VERIFIED.registry.json']=Path(verified['registry_path']).read_text()
    files['VERIFIED.native-receipt.json']=raw_receipt.decode()
    bundle=dict(schema_version=1,approval_sha256=APPROVAL,files=files)
    raw=json.dumps(bundle,sort_keys=True,separators=(',',':'),ensure_ascii=False).encode()
    assert len(raw)<64*1024*1024
    (out/'issue-330-caster-publication-evidence.json.gz').write_bytes(gzip.compress(raw,mtime=0))
    validation=json.loads((out/'issue-330-validation.json').read_text())
    validation.update(status='Native mechanical publication PASS; ROOT parent seal PENDING',
                      native_publication=verified,independent_native_binding=result)
    validation['not_verified']=['ROOT parent supplement publication/final seal','runtime qualification','PostgreSQL integration','browser/device','independent final PR acceptance/CI/merge']
    validation['shared_source_semantics']['native_publication_complete']=True
    write(out/'issue-330-validation.json',validation)
    print(json.dumps(dict(own_nodes=305,global_nodes=1888,source_plan='PASS',native=result,parent_appends='PENDING ROOT-only',bundle_bytes=(out/'issue-330-caster-publication-evidence.json.gz').stat().st_size),indent=2))


if __name__=='__main__':
    main()
