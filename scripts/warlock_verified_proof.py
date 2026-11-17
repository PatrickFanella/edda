"""Pinned delivered VERIFIED evidence; collect once explicitly, recheck offline."""
import argparse
import json
from pathlib import Path

import warlock_approved_native as a
import warlock_allocation_binding as allocation

SOURCE = Path('/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication')
COPIES = {
    'issue-331-verified-registry.json': ('milestones/VERIFIED.registry.json', 'e1e879b87b24914f08007677d30aad713c9e148363c404153bfd6be3f9564753'),
    'issue-331-verified-milestone.json': ('verified-milestone.json', 'fe8969fcdb5b11f81c00e47c8be1a2496472a4b7bee2b32c4cc932f4ad806305'),
    'issue-331-coordinator-verified-receipt.json': ('milestones/VERIFIED.native-receipt.json', '087b2521ad5b9f8c40b278f3b23d8eea4b1b3e025ed2069285a9c3947db11d54'),
}
EVIDENCE = a.DIRECTORY / 'issue-331-native-readback.json'


def inputs():
    raw = {}
    for name, (_, sha) in COPIES.items():
        raw[name] = (a.DIRECTORY / name).read_bytes()
        assert a.digest(raw[name]) == sha, ('exact delivered VERIFIED bytes', name)
        a.strict_json(raw[name])
    prior_raw = (a.DIRECTORY / allocation.REGISTRY_COPY).read_bytes()
    prior_delivery_raw = (a.DIRECTORY / allocation.DELIVERY_COPY).read_bytes()
    allocation.bind(prior_raw, prior_delivery_raw)
    approval, _, baseline, expected, _, _ = a.load_frozen()
    delivery = a.strict_json(raw['issue-331-verified-milestone.json'])
    prior = a.strict_json(prior_delivery_raw)
    assert delivery['registry_path'] == str(SOURCE / COPIES['issue-331-verified-registry.json'][0])
    assert delivery['verified_receipt_path'] == str(SOURCE / COPIES['issue-331-coordinator-verified-receipt.json'][0])
    receipt_raw = raw['issue-331-coordinator-verified-receipt.json']
    registry = a.validate_milestone(raw['issue-331-verified-registry.json'], delivery, expected, baseline,
                                    prior, receipt_raw, prior_raw)
    coordinator = a.strict_json(receipt_raw)
    assert coordinator['observed_at'] == delivery['verified_receipt_observed_at']
    # Check actual collected fields, never the producer's expected hashes/report.
    assert all(n['state'] in ('open', 'closed') for n in coordinator['graph']), 'exact REST state vocabulary'
    normalized = {**coordinator, 'graph': [{**n, 'state': n['state'].upper()} for n in coordinator['graph']]}
    a.validate_wiring(approval, expected, baseline, registry, normalized)
    return approval, baseline, expected, registry, delivery, prior


def verify(snapshot, approval, baseline, expected, registry):
    result = a.validate_wiring(approval, expected, baseline, registry, snapshot)
    graph = {n['number']: n for n in snapshot['graph']}
    assert len(graph) == 1888, 'delivered finite graph extent'
    counts = {str(n): len(graph[n]['blocked_by']) for n in (279, 280, 337)}
    counts['children283'] = sum(n['parent'] == 283 for n in graph.values())
    counts['outgoing429'] = sum(429 in n['blocked_by'] for n in graph.values())
    assert counts == {'279': 100, '280': 95, '337': 59, 'children283': 95, 'outgoing429': 100}
    return {**result, 'native_readback': 'PASS', 'final_acceptance': 'UNQUALIFIED', 'caps': counts}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--collect-once', action='store_true', help='One authorized read-only native collection; exclusive evidence creation')
    args = parser.parse_args()
    if args.collect_once:
        assert not EVIDENCE.exists(), 'retain existing native evidence; no recollection'
        # Re-read preserved immutable originals and retain exact portable copies.
        for path, sha in [(Path(allocation.REGISTRY_PATH), allocation.REGISTRY_SHA256),
                          (Path(allocation.DELIVERY_PATH), allocation.DELIVERY_SHA256)]:
            assert a.digest(path.read_bytes()) == sha, 'preserved original allocation bytes'
        for name, (source, sha) in COPIES.items():
            raw = (SOURCE / source).read_bytes()
            assert a.digest(raw) == sha, ('original VERIFIED pin', source)
            target = a.DIRECTORY / name
            if target.exists():
                assert target.read_bytes() == raw, 'retain exact previously copied delivery'
            else:
                with target.open('xb') as output:
                    output.write(raw)
    approval, baseline, expected, registry, delivery, prior = inputs()
    if args.collect_once:
        from warlock_native_proof import read_live
        snapshot = read_live(a.validate_registry(registry, expected, baseline), baseline['project'])
        # Save the actual collection before validation so a failed proof cannot cause a second read.
        evidence = dict(registry=registry, delivery=delivery, allocation_delivery=prior,
                        snapshot=snapshot, copied_artifacts={name: dict(source=str(SOURCE / source), sha256=sha)
                            for name, (source, sha) in COPIES.items()})
        with EVIDENCE.open('x', encoding='utf-8') as output:
            output.write(json.dumps(evidence, indent=2) + '\n')
    evidence = a.strict_json(EVIDENCE.read_bytes())
    assert a.exact(evidence['registry'], registry) and a.exact(evidence['delivery'], delivery)
    assert a.exact(evidence['allocation_delivery'], prior)
    result = verify(evidence['snapshot'], approval, baseline, expected, registry)
    print(json.dumps(result, sort_keys=True))


if __name__ == '__main__':
    main()
