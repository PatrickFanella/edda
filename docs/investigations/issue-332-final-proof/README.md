# Wizard #332 final source-publication proof

This checkpoint binds ROOT's actual four-parent publication receipt and journal
to one subsequent independent complete native/Project metadata collection. The
accepted POST-WARLOCK release and exact four sections/results are pinned in code;
candidate-controlled hashes cannot change them. Original ROOT provenance paths
and exact raw bytes remain in the copied receipt and journal. No producer code
or source regeneration is used.

`wizard-final-native-receipt.json.gz` contains the single fresh independent
collection, saved exclusively before proof assertions. Its collection-level
status text remains the collector's original unsealed wording; the separate
`wizard-final-native-proof.json` applies the pinned approval, identity, native,
ROOT-publication and final-prefix checks to those exact saved bytes. The receipt
is never rewritten to announce validation success.

The final proof checks all 871 literal rendered source bodies, actual native
identities, titles, Task/Project Size M and OPEN states; exact approved containment
and prerequisites; old title/body/state/metadata and actual identity protections;
both DAGs, all caps and Wizard's mandatory output paths. Only the four exact
accepted-prefix-plus-owned-section extensions are permitted. Both accepted
predecessors remain CLOSED and #332 remains OPEN for ROOT review and merge.

`wizard-final-bindings.json` retains all 285 original binding rows and six maps,
with final source-publication status and actual evidence provenance. It does not
qualify runtime execution. Runtime qualification is false and runtime roadmap
implementation remains PAUSED. Historical allocation, unsealed and preparation
checkpoints remain intact as phase evidence, with this directory providing the
current final source-publication result.

ROOT receipt raw SHA-256:
`3f429d86cb00e7b457aeece26918821e494bb9c2536266b9dc95a740de9d039d`.
ROOT journal raw SHA-256:
`b5fdd3dc0ef02ceb07df4fba891a638dfb7c993248d46879af3d998905b48e47`.
`provenance.json` records original paths and exact hashes of the saved native
receipt, copied ROOT publication records, milestone proof and final results.

Offline replay from the repository root:

```sh
rtk proxy python - <<'PYCODE'
import gzip, hashlib, json, sys
from pathlib import Path
sys.path.insert(0, 'scripts')
from wizard_final_proof import validate_final
from wizard_publication_proof import trusted_inputs
p = Path('docs/investigations/issue-332-final-proof')
m = json.loads((p / 'provenance.json').read_bytes())
for name, sha in m['artifacts'].items():
    assert hashlib.sha256((p / name).read_bytes()).hexdigest() == sha
raw = gzip.decompress((p / 'wizard-final-native-receipt.json.gz').read_bytes())
assert hashlib.sha256(raw).hexdigest() == m['native_receipt_raw_sha256']
receipt = json.loads(raw)
registry = json.loads(Path('docs/investigations/issue-332-native-proof/VERIFIED.registry.json').read_bytes())
proof = json.loads(json.dumps(validate_final(receipt, registry, trusted_inputs())))
saved = json.loads((p / 'wizard-final-native-proof.json').read_bytes())
assert all(saved[key] == value for key, value in proof.items())
assert saved['observed_parent_body_sha256'] == saved['expected_parent_result_sha256']
assert saved['runtime_qualified'] is False
print(saved['status'], saved['observed_at'], saved['caps'])
PYCODE
```

Focused contrary tests continue to reject rollback, partial/changed sections,
duplicates, altered duplicates, coherent candidate manifest/receipt/marker
rewrites, stale snapshots and native drift. Added ROOT-provenance controls reject
rewritten publication receipts/journals and snapshots predating actual appends.
Synthetic test success is separate from this actual saved-evidence result.

ROOT alone owns fresh independent acceptance, exact-head CI and merge. The PR
remains draft. No native writes, owner-contract edits, new tasks, runtime work,
Go/vet repeats, broad source audit, watches, merges, closures or deployment are
performed by this checkpoint.
