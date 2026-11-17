"""Pinned POST-WARLOCK prefix policy; no collection, writes or fabricated seal."""
import gzip
import hashlib
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DIRECTORY = ROOT / 'docs/investigations/issue-332-final-preparation'
RELEASE_SHA = '2a14082c35d68c6859490194738c74a6b820f81737ff6b81cd2ddf57fbbc4524'
PACKAGE_SHA = '5d80a055a6ed236b7dee4b50299d9617438f20702a79837b858914fdc51990b1'
ACCEPTED_HEAD = '11f18dc15b1e214e6fe829a28177a32ebd0c39a1'
ACCEPTED_MERGE = '4b703526c2a0aa05d68dd4bfe721430dc686625e'
NODES_SHA = '866246ec119bba3198ee6f6809141d7afc8efba596ad1bcadd3a8ab2194afa4c'
SOURCE_SHA = 'ef1f0ea92dd01f8ced8195fb1ef61b082f7cad35a50a8d44bc05bb43fe4b4634'
PRIOR_RECEIPT_SHA = '05253922cd8b93e72e1feaa5d2b24c98cf316cd5eca8f147dadf779324a3a6df'
MARKER = '<!-- edda-332 -->'
SHORT = '\n\n<!-- edda-332 -->\nWizard #332: source decision; execution gated.\n'
LONG = ('\n\n<!-- edda-332 -->\nWizard/Evoker #332 reconciles 106 source contexts, 20 class rows and the 217 printed spell identities. '
        'Source records remain metadata. Every actual feature, level application, spell effect and lifecycle output must qualify '
        'through the published bounded children to the Wizard class gate and #337. Shared outputs route through the reserved '
        'three-class aggregate to #279/#280. No catalogue, decision closure or generic confirmation label supplies playable admission.\n')
# Independent literal pins, never expectations supplied by a candidate or receipt.
PREFIXES = {
    47: '2a519871a0cabc51b76d2bf6862f092a8c85b6fb8f9e6d28e0072ee1dff80fb9',
    54: '84dfad7c3b64fbb6432fb86503b00bd0c610d0bb7462ba6a2836104802333925',
    283: '8cdcdb3a307539e7be9f5d3d1cb3c31c71bf1f5efc52647bae3c8defa411835b',
    337: 'e762e12de06dc3af6479c948c5a89bb40fc00df7435208b9210c07c2f15972f8',
}
SECTIONS = {47: 'd38fa5ff4e509f940e0d9face47f45c4d4627bc0254ee7da40bad58a95fd513b',
            54: '7a9125eb0853fe27486c2ad299ced0dd24d40f9e658932220fe2eb36eb9a3944',
            283: '7a9125eb0853fe27486c2ad299ced0dd24d40f9e658932220fe2eb36eb9a3944',
            337: '7a9125eb0853fe27486c2ad299ced0dd24d40f9e658932220fe2eb36eb9a3944'}
RESULTS = {
    47: 'efe58e4675d673d047c39efbd4d2b288e13dcf7c98abb0f372bb6015295330d4',
    54: 'cdba8e04845f6161749dd5aaf819aa36c2cc02c912f960868adad6d567a635b3',
    283: 'dc7c29eaf9073b3954c0f6b7f060b5d79f1cf768889f0289407bdaad3a2529de',
    337: '7b73e50fc1c12be4d7a8175f7ddb9b22fec60194d222402e6d7d0bbf59c9a242',
}


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def load_policy(directory=DIRECTORY):
    raw = gzip.decompress((directory / 'accepted-post-warlock-parent-baselines.json.gz').read_bytes())
    assert sha(raw) == RELEASE_SHA, 'pinned accepted release bytes'
    release = json.loads(raw)
    assert release['status'] == 'ACCEPTED_POST_WARLOCK_PARENT_PREFIX_RELEASE'
    assert release['accepted_head'] == ACCEPTED_HEAD and release['accepted_merge'] == ACCEPTED_MERGE
    assert release['immutable_verified_nodes_map_sha256'] == NODES_SHA
    package_raw = (directory / 'root-append-package.json').read_bytes()
    assert sha(package_raw) == PACKAGE_SHA, 'pinned ROOT-only append package bytes'
    package = json.loads(package_raw)
    assert package['release_raw_sha256'] == RELEASE_SHA and package['marker'] == MARKER
    assert package['accepted_head'] == ACCEPTED_HEAD and package['accepted_merge'] == ACCEPTED_MERGE
    assert set(package['appends']) == set(map(str, PREFIXES))
    bodies = {}
    for number, pin in PREFIXES.items():
        row = release['body_seals'][str(number)]
        prefix = row['body']
        assert sha(prefix.encode()) == row['sha256'] == pin, 'independent predecessor body pin'
        assert len(prefix) == row['characters'] and len(prefix.encode()) == row['utf8_bytes']
        section = SHORT if number == 47 else LONG
        raw_section = (directory / f'append-{number}.md').read_bytes()
        assert raw_section == section.encode(), 'literal owned section and marker'
        assert sha(raw_section) == SECTIONS[number], 'independent owned section pin'
        assert prefix.count(MARKER) == 0 and section.count(MARKER) == 1
        assert len(section) <= (100 if number == 47 else 749)
        expected = prefix + section
        assert sha(expected.encode()) == RESULTS[number], 'independent complete result pin'
        assert len(expected) <= 65536 and len(expected.encode()) <= 65536, 'character and Unicode byte caps'
        append = package['appends'][str(number)]
        assert append['prefix_sha256'] == pin and append['section_sha256'] == SECTIONS[number]
        assert append['expected_result_sha256'] == RESULTS[number]
        bodies[number] = expected
    return release, bodies


def validate_parent_extensions(receipt, directory=DIRECTORY):
    release, expected = load_policy(directory)
    rows = {r['number']: r for r in receipt['graph']}
    assert len(rows) == len(receipt['graph']), 'duplicate native issue'
    for number, prefix in expected.items():
        body = rows[number]['body']
        assert body.startswith(prefix), ('exact accepted prefix plus owned section required', number)
        assert body.count(MARKER) == 1, ('duplicate owned marker', number)
        tail = body[len(prefix):]
        assert not re.search(r'<!--\s*edda[-:]?332\b', tail, re.IGNORECASE), 'altered duplicate owned marker'
        assert 'Wizard #332: source decision' not in tail and 'Wizard/Evoker #332 reconciles' not in tail, 'altered duplicate owned section'
        assert len(body) <= 65536 and len(body.encode()) <= 65536, 'live Unicode byte/character caps'
    return set(expected)


def validate_final(receipt, registry, inputs):
    from wizard_milestone_proof import nodes_digest, timestamp
    from wizard_publication_proof import validate_live, validate_candidate
    # No candidate-controlled source or receipt hashes can authorize a new policy.
    source = (ROOT / 'docs/investigations/issue-332-wizard-scope.json').read_bytes()
    assert sha(source) == SOURCE_SHA, 'byte-exact own frozen source'
    validate_candidate(json.loads(source), inputs[1]['332'])
    assert nodes_digest(registry, inputs) == NODES_SHA
    release, _ = load_policy()
    assert timestamp(receipt['observed_at']) >= timestamp(release['observed_at']), 'receipt predates accepted release'
    # Retain actual old database/GraphQL identities, absent from the old source baseline.
    prior_raw = (ROOT / 'docs/investigations/issue-332-native-proof/wizard-native-receipt.json.gz').read_bytes()
    assert sha(prior_raw) == PRIOR_RECEIPT_SHA, 'pinned prior actual native identities'
    prior = {r['number']: r for r in json.loads(gzip.decompress(prior_raw))['graph']}
    rows = {r['number']: r for r in receipt['graph']}
    for old in inputs[2]['graph']:
        number = old['number']
        for field in ('id', 'node_id', 'typename'):
            assert rows[number][field] == prior[number][field], ('protected old identity', number, field)
    assert rows[330]['state'].upper() == rows[331]['state'].upper() == 'CLOSED', 'accepted predecessor decisions remain closed'
    assert rows[332]['state'].upper() == 'OPEN', 'Wizard remains open pending ROOT review and merge'
    proof = validate_live(receipt, registry, inputs, final_parent_policy=True)
    return dict(proof, status='FINAL_NATIVE_AND_PREFIX_PASS', final_prefix_sealed=True,
                accepted_parent_prefix=RELEASE_SHA, accepted_predecessor_head=ACCEPTED_HEAD,
                accepted_predecessor_merge=ACCEPTED_MERGE, owned_section_sha256=SECTIONS,
                expected_parent_result_sha256=RESULTS,
                observed_parent_body_sha256={n: sha(rows[n]['body'].encode()) for n in PREFIXES},
                observed_at=receipt['observed_at'], runtime_implementation_status='PAUSED')
