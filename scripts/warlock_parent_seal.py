"""Independent predecessor and own-section pins for ROOT-only Warlock appends."""
import json
import re

import warlock_approved_native as a

RELEASE_FILE = 'issue-331-accepted-post-sorcerer.json'
RELEASE_PATH = '/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/accepted-post-sorcerer-parent-baselines.json'
RELEASE_SHA256 = '85f9f6cdcfedb453d3ec7abbaa52042c3cea2f1fb62d57e802b70efa3fbae92d'
MARKER = '<!-- edda331 -->'
# These literals are verifier policy, never read from a candidate manifest/receipt.
PREFIX_PINS = {
    47: ('a4b56fe0bf38237499bec40387204ee72835f16ac77339d6574ef9c95b049243', 65047, 65141),
    54: ('1255b663f2161b564deea9f7597e1661b3521741da8f04aa53f23aaff902ad2e', 44005, 44049),
    283: ('110b18f6c7629e1286f7a6d8c1d1d705579b7a71dc48274bd024ea18d3213ba8', 32551, 32591),
    337: ('0defeea1fa26056e4a8126fa3440ccd55919e58633b38d37afa92f20378f7a25', 31594, 31633),
}
SECTION_PINS = {
    47: '3ed33bbc6f54d90ecd4c6309500772da424e0bb3245c753ca755623910745d39',
    54: '4b26cef7c1223c2fcbae56739793ac485c2c7c7ddd6f2ac100c4b434248d536d',
    283: 'd860127155f529fbb1258ab168a392490e41218f08cb1cf80002f44c259bf81e',
    337: 'fc12aed1b22a3a0a252f6866d0ca3e44fa59b6d02171a65ea855e4861986d0a1',
}


def load_prefixes(directory=a.DIRECTORY):
    raw = (directory / RELEASE_FILE).read_bytes()
    assert a.digest(raw) == RELEASE_SHA256, 'independent accepted predecessor delivery pin'
    release = a.strict_json(raw)
    assert release['status'] == 'ACCEPTED_POST_SORCERER_PARENT_PREFIX_RELEASE'
    assert release['accepted_head'] == '1d1c9d44a1a0a223948084ed290b585b01144ac1'
    assert release['accepted_merge'] == '6936b36af81ca9eec036395c4851c916ae25a8e1'
    assert set(release['body_seals']) == {str(n) for n in PREFIX_PINS}
    prefixes = {}
    for n, (sha, characters, size) in PREFIX_PINS.items():
        seal = release['body_seals'][str(n)]
        body = seal['body']
        assert a.digest(body.encode('utf-8')) == sha == seal['sha256'], 'independent predecessor body pin'
        assert len(body) == characters == seal['characters']
        assert len(body.encode('utf-8')) == size == seal['utf8_bytes']
        assert MARKER not in body
        prefixes[n] = body
    return prefixes


def load_sections(directory=a.DIRECTORY):
    sections = {}
    for n, sha in SECTION_PINS.items():
        raw = (directory / f'issue-331-append-{n}.md').read_bytes()
        assert a.digest(raw) == sha, 'independent own section pin'
        section = raw.decode('utf-8')
        assert section.count(MARKER) == 1
        assert len(section) <= (100 if n == 47 else 750)
        sections[n] = section
    return sections


def validate_parents(graph, directory=a.DIRECTORY):
    prefixes, sections = load_prefixes(directory), load_sections(directory)
    for n, prefix in prefixes.items():
        actual, section = graph[n]['body'], sections[n]
        assert actual.startswith(prefix + section), ('exact accepted prefix and pinned own section pending', n)
        assert actual.count(MARKER) == 1, 'duplicate own marker'
        tail = actual[len(prefix + section):]
        # Reject a duplicate own section even if its marker was removed/altered.
        payload = section.replace(MARKER, '').strip()
        assert payload not in tail, 'altered duplicate own section'
        assert not re.search(r'edda\s*[-_:]?\s*331', tail, re.I), 'altered duplicate own marker'
    return True


def package(directory=a.DIRECTORY):
    prefixes, sections = load_prefixes(directory), load_sections(directory)
    artifacts = {}
    for n in prefixes:
        body = prefixes[n] + sections[n]
        # Conservative check in both representations, including the full roadmap body.
        assert len(body) <= 65536 and len(body.encode('utf-8')) <= 65536, 'parent body capacity'
        artifacts[str(n)] = dict(path=str(directory / f'issue-331-append-{n}.md'),
            prefix_sha256=PREFIX_PINS[n][0], section_sha256=SECTION_PINS[n], marker=MARKER,
            append_characters=len(sections[n]), append_utf8_bytes=len(sections[n].encode('utf-8')),
            expected_body_sha256=a.digest(body.encode('utf-8')), expected_body_characters=len(body),
            expected_body_utf8_bytes=len(body.encode('utf-8')))
    return dict(status='PENDING_ROOT_ONLY_APPENDS', release_source=RELEASE_PATH,
        release_sha256=RELEASE_SHA256, final_prefix_sealed=False, runtime_qualified=False, artifacts=artifacts)


if __name__ == '__main__':
    print(json.dumps(package(), indent=2))
