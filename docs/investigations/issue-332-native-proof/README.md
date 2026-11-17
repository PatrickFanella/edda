# Wizard #332 actual native proof, unsealed

This checkpoint follows the preserved allocation checkpoint. ROOT delivered the
immutable VERIFIED registry, milestone and coordinator receipt. Their exact raw
bytes and the preserved IDS bytes passed the pinned independent milestone
validator, including all five literal true claims, observation ordering and the
unchanged 871-node identity map. Original delivery paths remain provenance facts;
no mutable registry, journal or progress file was inspected.

The coordinator receipt has a different schema from the independent collector:
it lacks per-row `typename` and Project metadata. Its byte hash and milestone
provenance were validated without inventing those fields. Its original JSON bytes
are retained through lossless gzip compression. It is not substituted for the
fresh independent receipt.

`wizard-native-receipt.json.gz` is the result of one complete read-only native
collection with the existing independent collector. Evidence files were created
exclusively, and the fresh receipt was saved before native proof assertions.
`wizard-native-proof.json` records the independent proof result.
`wizard-verified-bindings.json` retains the same 285 identity/expected-body/parent/
prerequisite bindings and six source-obligation maps as the allocation checkpoint,
now linked to the independently checked actual receipt. None of these outputs
regenerates or changes approved source semantics.

The pinned validator compares literal rendered bodies, titles, actual issue IDs,
Task/Project Size M and OPEN state for all 871 new nodes; exact parents and
prerequisites for the complete 1,888-node approved graph; old bodies, titles,
states and Project metadata; both DAGs and all three caps; and mandatory Wizard
terminal/shared-root paths. The result remains unsealed. No accepted POST-WARLOCK
prefix release has been supplied. All four parent supplements remain ROOT-only,
with serial acceptance ordered 330 then 331 then 332. Runtime is UNQUALIFIED and
runtime roadmap implementation is PAUSED.

`provenance.json` records original delivery paths/raw hashes, portable-copy
provenance and exact byte hashes for evidence and proof artifacts. The IDS bytes
remain in `../issue-332-allocation-binding/`. For compressed receipts, the raw
JSON hash is distinct from the committed gzip file hash.

Replaying saved evidence is offline. From the repository root:

```sh
rtk proxy python - <<'PYCODE'
import gzip, hashlib, json, sys
from pathlib import Path
sys.path.insert(0, 'scripts')
from wizard_milestone_proof import validate_transition
from wizard_publication_proof import trusted_inputs, validate_live, validate_candidate
root = Path('docs/investigations')
p = root / 'issue-332-native-proof'
a = root / 'issue-332-allocation-binding'
m = json.loads((p / 'provenance.json').read_bytes())
for name, sha in m['artifacts'].items():
    assert hashlib.sha256((p / name).read_bytes()).hexdigest() == sha
inputs = trusted_inputs()
validate_candidate(json.loads((root / 'issue-332-wizard-scope.json').read_bytes()), inputs[1]['332'])
coordinator = gzip.decompress((p / 'VERIFIED.native-receipt.json.gz').read_bytes())
transition = validate_transition(
    json.loads((a / 'ids-allocated-milestone.json').read_bytes()),
    (a / 'IDS_ALLOCATED.registry.json').read_bytes(),
    json.loads((p / 'verified-milestone.json').read_bytes()),
    (p / 'VERIFIED.registry.json').read_bytes(), coordinator, inputs)
assert transition == json.loads((p / 'milestone-provenance-proof.json').read_bytes())
receipt = json.loads(gzip.decompress((p / 'wizard-native-receipt.json.gz').read_bytes()))
proof = validate_live(receipt, json.loads((p / 'VERIFIED.registry.json').read_bytes()), inputs)
saved = json.loads((p / 'wizard-native-proof.json').read_bytes())
assert all(saved[key] == value for key, value in proof.items())
print(proof['status'], receipt['observed_at'], proof['caps'])
PYCODE
```

No producer imports, source regeneration, Go/vet reruns, native writes, owner
issue edits, closures, deployments or hosted polling belong to this checkpoint.
ROOT owns independent acceptance, supplements, CI and merge. The final hosted
snapshot is deferred until the actual final-prefix update.

Actual collection and validation passed at
`2026-11-17T06:27:03.985841+00:00`, with 1,888 nodes and 19 API calls.
All 871 new nodes and 1,017 old-node protections passed. Specific counts are
#279 incoming 100, #280 incoming 95, #337 incoming 59, #283 children 95 and
#429 outgoing 100. Overall maxima are incoming 100, outgoing 100, children 95.

Local validation: 20 existing publication/milestone proof tests passed in
17.067 seconds; five allocation checkpoint tests passed in 0.164 seconds.
The documented portable offline replay passed against the saved actual receipt
and exact artifact hashes. The allocation binding rows and six source-obligation
maps match the new verified sidecar exactly; only proof provenance/status advance.
The final-prefix refusal remains covered by the passing allocation tests.
