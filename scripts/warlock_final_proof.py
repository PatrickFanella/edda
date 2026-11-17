"""Final source-publication proof; collect once explicitly, replay offline by default."""
import argparse
import json

import warlock_verified_proof as verified
import warlock_parent_seal as seal
import warlock_native_proof as native

a = verified.a
COPIES = {
    'issue-331-root-parent-publication-receipt.json': ('root-warlock-parent-publication-receipt.json', '53e8b461f52d2010ef488612886b80c9ee31347991a2a09ba7eae6af9498811e'),
    'issue-331-root-parent-publication-journal.json': ('root-warlock-parent-publication.json', '4b403e2bfaa0bfe2e222f21054401dbf6afb468990492ea4385702d996f0c9e1'),
}
PACKAGE_SHA256 = 'a27066d3ea03181b3cdcff196b0a5c2c9ab6432db71f5230565d9fb0d93a82fe'
EVIDENCE = a.DIRECTORY / 'issue-331-sealed-native-readback.json'


def publication_inputs():
    objects = {}
    for name, (_, sha) in COPIES.items():
        raw = (a.DIRECTORY / name).read_bytes()
        assert a.digest(raw) == sha, ('exact ROOT publication pin', name)
        objects[name] = a.strict_json(raw)
    receipt, journal = [objects[name] for name in COPIES]
    assert receipt['status'] == 'ROOT_FOUR_WARLOCK_APPENDS_PUBLISHED_AND_EXACT_READBACK_PASS'
    assert receipt['accepted_predecessor_release_sha256'] == seal.RELEASE_SHA256
    assert receipt['package_sha256'] == journal['package_sha256'] == PACKAGE_SHA256
    assert a.digest((a.DIRECTORY / 'issue-331-pending-root-appends.json').read_bytes()) == PACKAGE_SHA256
    assert a.exact(receipt['operations'], journal['operations'])
    assert [r['number'] for r in receipt['operations']] == [47, 54, 283, 337]
    package = seal.package()
    for operation in receipt['operations']:
        n = operation['number']
        row = package['artifacts'][str(n)]
        assert operation['status'] == 'ACKNOWLEDGED_AND_EXACT_READBACK'
        assert operation['prior_sha256'] == seal.PREFIX_PINS[n][0]
        assert operation['append_sha256'] == seal.SECTION_PINS[n]
        assert operation['expected_sha256'] == row['expected_body_sha256']
        for field, value in row.items():
            if field != 'path':  # Original absolute artifact path is retained by the raw receipt pin.
                assert a.exact(receipt['expected_results'][str(n)][field], value)
    return {name: dict(source=str(verified.SOURCE / source), sha256=sha)
            for name, (source, sha) in COPIES.items()}


def verify(evidence):
    provenance = publication_inputs()
    approval, baseline, expected, registry, delivery, prior = verified.inputs()
    for key, value in [('registry', registry), ('delivery', delivery), ('allocation_delivery', prior),
                       ('root_publication', provenance)]:
        assert a.exact(evidence[key], value), ('exact sealed evidence provenance', key)
    snapshot = evidence['snapshot']
    wiring = verified.verify(snapshot, approval, baseline, expected, registry)
    _, proposals, _, _, _, _ = a.load_frozen()
    result = native.validate_native(None, proposals[331], evidence, snapshot)
    graph = {n['number']: n for n in snapshot['graph']}
    assert graph[330]['state'] == 'CLOSED' and graph[331]['state'] == graph[332]['state'] == 'OPEN'
    # Exact current full-body result check, in addition to the extensible pinned seal rule.
    for n, row in seal.package()['artifacts'].items():
        body = graph[int(n)]['body']
        assert a.digest(body.encode('utf-8')) == row['expected_body_sha256']
        assert len(body) == row['expected_body_characters']
        assert len(body.encode('utf-8')) == row['expected_body_utf8_bytes']
    return {**result, 'source_publication': 'PASS', 'native_readback': 'PASS',
            'caps': wiring['caps'], 'decision_acceptance': 'PENDING_ROOT_REVIEW_CI_MERGE',
            'runtime_implementation': 'PAUSED'}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--collect-once', action='store_true')
    args = parser.parse_args()
    if args.collect_once:
        assert not EVIDENCE.exists(), 'retain sealed native evidence; no recollection'
        for name, (source, sha) in COPIES.items():
            raw = (verified.SOURCE / source).read_bytes()
            assert a.digest(raw) == sha, 'original ROOT publication bytes'
            target = a.DIRECTORY / name
            if target.exists():
                assert target.read_bytes() == raw
            else:
                with target.open('xb') as output:
                    output.write(raw)
        provenance = publication_inputs()
        _, baseline, expected, registry, delivery, prior = verified.inputs()
        snapshot = native.read_live(a.validate_registry(registry, expected, baseline), baseline['project'])
        evidence = dict(registry=registry, delivery=delivery, allocation_delivery=prior,
                        root_publication=provenance, snapshot=snapshot)
        # Retain actual observed fields before any assertion about the collected graph.
        with EVIDENCE.open('x', encoding='utf-8') as output:
            output.write(json.dumps(evidence, indent=2) + '\n')
    evidence = a.strict_json(EVIDENCE.read_bytes())
    print(json.dumps(verify(evidence), sort_keys=True))


if __name__ == '__main__':
    main()
