"""Offline binding of ROOT's delivered IDS_ALLOCATED bytes; no native readback."""
import argparse
import json

import warlock_approved_native as a

REGISTRY_PATH = '/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/milestones/IDS_ALLOCATED.registry.json'
DELIVERY_PATH = '/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/ids-allocated-milestone.json'
REGISTRY_SHA256 = '73d9755cd7181155544f5d79a944bdaf653028425a75b5131b9adef1faf76392'
DELIVERY_SHA256 = '4ada8b216c27369df0cca9304d4ec5f4f556035362ee0c3d26ebef8d21fdf9a6'
NODES_SHA256 = '866246ec119bba3198ee6f6809141d7afc8efba596ad1bcadd3a8ab2194afa4c'
REGISTRY_COPY = 'issue-331-allocated-registry.json'
DELIVERY_COPY = 'issue-331-allocated-milestone.json'
BINDING = 'issue-331-allocation-bindings.json'


def bind(registry_raw, delivery_raw, directory=a.DIRECTORY):
    assert a.digest(registry_raw) == REGISTRY_SHA256, 'exact delivered allocation bytes'
    assert a.digest(delivery_raw) == DELIVERY_SHA256, 'exact delivered sidecar bytes'
    # Parse every frozen input strictly before the existing pinned loaders use it.
    for filename in [*a.FILES.values(), 'issue-331-publication-approval.json',
                     'issue-331-native-baseline.json', 'issue-331-renderer-contract.json',
                     'issue-331-milestone-contract.json']:
        a.strict_json((directory / filename).read_bytes())
    _, _, baseline, expected, _, _ = a.load_frozen(directory)
    a.load_renderer(directory)
    a.load_milestone_contract(directory)
    delivery = a.strict_json(delivery_raw)
    assert delivery['milestone'] == 'IDS_ALLOCATED'
    assert delivery['registry_path'] == REGISTRY_PATH
    assert delivery['nodes_map_sha256'] == NODES_SHA256
    assert delivery['observed_at'] == '2026-11-17T01:44:25.687122+00:00'
    registry = a.validate_milestone(registry_raw, delivery, expected, baseline)
    numbers = a.validate_registry(registry, expected, baseline)
    structure = a.resolved_structure(expected, numbers)
    bodies = a.final_bodies(registry, expected, baseline)
    own = {alias: dict(**registry['nodes'][alias],
               expected_parent_number=structure[numbers[alias]]['parent'],
               expected_blocked_by_numbers=structure[numbers[alias]]['blocked_by'],
               expected_rendered_body_sha256=a.digest(bodies[alias].encode('utf-8')))
           for alias in sorted(expected) if expected[alias]['owner_decision'] == 331}
    assert len(own) == 281
    return dict(status='ALLOCATION IDS_BOUND', own_nodes=281, allocated_nodes=871,
        approval_sha256=a.APPROVAL_SHA256, own_proposal_sha256=a.OWN_PROPOSAL_SHA256,
        renderer_contract_sha256=a.RENDERER_SHA256, milestone_contract_sha256=a.MILESTONE_SHA256,
        observed_at=delivery['observed_at'], nodes_map_sha256=NODES_SHA256,
        copied_artifacts={REGISTRY_COPY: dict(source=REGISTRY_PATH, sha256=REGISTRY_SHA256),
                          DELIVERY_COPY: dict(source=DELIVERY_PATH, sha256=DELIVERY_SHA256)},
        native_wiring='UNQUALIFIED', native_readback='UNQUALIFIED', final_acceptance='UNQUALIFIED',
        final_prefix_sealed=False, runtime_qualified=False, bindings=own)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--write', action='store_true', help='Write derived binding sidecar from committed exact-byte copies')
    args = parser.parse_args()
    result = bind((a.DIRECTORY / REGISTRY_COPY).read_bytes(), (a.DIRECTORY / DELIVERY_COPY).read_bytes())
    path = a.DIRECTORY / BINDING
    raw = (json.dumps(result, indent=2, ensure_ascii=False) + '\n').encode('utf-8')
    if args.write:
        path.write_bytes(raw)
    else:
        assert path.read_bytes() == raw, 'committed binding differs from offline derivation'
    print('PASS ALLOCATION IDS_BOUND: 281 own / 871 allocated; wiring/readback/final acceptance UNQUALIFIED')


if __name__ == '__main__':
    main()
