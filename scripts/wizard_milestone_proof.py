"""Read-only root-delivered publication provenance; never polls or writes registry."""
import hashlib
import json
from pathlib import Path
from datetime import datetime
from wizard_publication_proof import (APPROVAL_DIR,APPROVAL_SHA256,RENDERER_SHA256,
 read_pinned,validate_registry,decode_registry)

MILESTONE_CONTRACT_SHA256='58c5aacb2a3e73b759dff6b98316661fb6bd9870b95e5c7fa9a0d39e7ee736ba'
COUNTS={'330':305,'331':281,'332':285}


def nodes_digest(registry,inputs):
    validate_registry(registry,inputs[-1],inputs[2])
    raw=json.dumps(registry['nodes'],sort_keys=True,separators=(',',':'),ensure_ascii=False).encode('utf-8')
    return hashlib.sha256(raw).hexdigest()


def timestamp(value):
    assert isinstance(value,str) and value,'missing observation timestamp'
    parsed=datetime.fromisoformat(value.replace('Z','+00:00'))
    assert parsed.tzinfo is not None,'observation timestamp must include timezone'
    return parsed


def validate_milestone(delivery,registry_bytes,inputs,phase):
    contract=read_pinned(APPROVAL_DIR/'milestone-provenance-contract.json',MILESTONE_CONTRACT_SHA256)
    fields=contract['required_ids_allocated_fields']
    if phase=='VERIFIED':fields=fields+contract['required_verified_additional_fields']
    assert set(fields)<=set(delivery),'missing delivered milestone provenance'
    assert delivery['milestone']==phase and phase in ['IDS_ALLOCATED','VERIFIED']
    assert delivery['approval_sha256']==APPROVAL_SHA256
    assert delivery['renderer_contract_sha256']==RENDERER_SHA256
    assert type(delivery['node_count']) is int and delivery['node_count']==871
    assert delivery['class_counts']==COUNTS and all(type(n) is int for n in delivery['class_counts'].values())
    timestamp(delivery['observed_at'])
    assert hashlib.sha256(registry_bytes).hexdigest()==delivery['registry_file_sha256'],'actual milestone registry byte hash'
    registry=decode_registry(registry_bytes);assert registry['status']==phase,'phase snapshot must retain exact status'
    assert nodes_digest(registry,inputs)==delivery['nodes_map_sha256'],'independent stable identity map hash'
    # A delivery may reference an immutable phase snapshot of the root registry.
    # The path is provenance, not authority to read arbitrary data or mutate it.
    assert isinstance(delivery['registry_path'],str) and delivery['registry_path']
    return registry


def validate_transition(ids_delivery,ids_bytes,verified_delivery,verified_bytes,evidence_bytes,inputs):
    ids=validate_milestone(ids_delivery,ids_bytes,inputs,'IDS_ALLOCATED')
    verified=validate_milestone(verified_delivery,verified_bytes,inputs,'VERIFIED')
    assert ids_delivery['nodes_map_sha256']==verified_delivery['nodes_map_sha256'],'identity reassignment across phases'
    assert ids['nodes']==verified['nodes'],'full stable registry identities'
    assert timestamp(ids_delivery['observed_at'])<=timestamp(verified_delivery['observed_at']),'milestone time reversal'
    assert hashlib.sha256(evidence_bytes).hexdigest()==verified_delivery['verified_receipt_sha256'],'actual verified collected evidence byte hash'
    evidence=json.loads(evidence_bytes)
    assert 'graph' in evidence and 'observed_at' in evidence,'VERIFIED must reference collected native evidence, not projection'
    assert evidence['observed_at']==verified_delivery['verified_receipt_observed_at'],'collected evidence observation time'
    assert timestamp(ids_delivery['observed_at'])<=timestamp(evidence['observed_at'])<=timestamp(verified_delivery['observed_at']),'collected evidence time outside delivered publication boundaries'
    count=verified_delivery['graph_node_count']
    assert type(count) is int and count>=1888 and len(evidence['graph'])==count
    for name in ['all_new_tasks_open','native_body_parent_prerequisite_Task_Size_validation','both_DAGs','all_three_caps','old_protections']:
        assert verified_delivery[name] is True,('coordinator did not qualify boundary',name)
    return dict(status='DELIVERED_MILESTONE_PROVENANCE_PASS; independent live proof still required',
                approval_sha256=APPROVAL_SHA256,renderer_contract_sha256=RENDERER_SHA256,
                milestone_contract_sha256=MILESTONE_CONTRACT_SHA256,nodes_map_sha256=nodes_digest(verified,inputs),
                allocated_registry_sha256=hashlib.sha256(ids_bytes).hexdigest(),verified_registry_sha256=hashlib.sha256(verified_bytes).hexdigest(),
                coordinator_evidence_sha256=hashlib.sha256(evidence_bytes).hexdigest(),
                native_live_proof=False,accepted_parent_prefix=None,runtime_status='UNQUALIFIED')
