# Warlock publication proof before final prefix acceptance

ROOT approved exactly 871 future Tasks: Sorcerer 305, Warlock 281 and Wizard 285. The 1017-node baseline and those Tasks form the approved 1888-node projection. Sorcerer is the sole native writer. This checkout contains read-only proof code and frozen input copies; it has no native mutation implementation.

The Warlock proposal remains byte-for-byte SHA256 `8d4a8ea724ee8e997093d0e3877054a1c4a499f1f00a059bcbb9d2d51bc4268a`. The source predicates, ownership, counts, bodies and dependency aliases are unchanged from ROOT's accepted bounded delta review. The Sorcerer/Wizard boundary snapshots and native baseline were refreshed only from ROOT's frozen approved files. `issue-331-approval-provenance.json` records their original paths, approved heads and hashes.

| Independent pin | SHA256 |
| --- | --- |
| Finite publication approval | `66b934f5cac58d21644822bd36098b0ee1cad535708a9e3f6b49bdf0a2f0fe41` |
| Native renderer contract | `e466e1b41506c35f6f03b6c3cc8d2bbbfb57e168607fa3c63954291930247238` |
| Synthetic renderer vectors | `4ba7f97dd6f1702ff049056f10e46736cb1da27fe7274553b4bff9e54077bdee` |
| Milestone provenance contract | `58c5aacb2a3e73b759dff6b98316661fb6bd9870b95e5c7fa9a0d39e7ee736ba` |

`warlock_approved_native.py` independently normalizes every frozen approved alias and compares all 871 complete bodies, titles, parents, prerequisites and Task/M expectations to the pinned approval. It implements the renderer with independent literal templates and tests those literals against the pinned golden vectors. It imports no coordinator renderer. The partial synthetic vectors cannot satisfy the complete 871-alias registry requirement.

Every registry identity must have exactly `number`, `id` and `node_id`. Numeric identities must be positive JSON integers; booleans and floats fail. All three fields must be unique across the 871 aliases, and issue numbers cannot overlap the baseline. Duplicate JSON fields and nonfinite values fail. The final body preserves all approved body text and trailing newlines, then appends only the exact renderer suffix with actual issue, parent and numerically sorted prerequisite numbers.

## Unsealed actual-native proof

ROOT's immutable VERIFIED delivery has now been validated independently. A single complete read-only collection from `2026-11-17T06:25:56.471901+00:00` to `2026-11-17T06:26:52.428367+00:00` used 63 bounded GraphQL queries and recorded 1888 graph nodes plus 1096 Project metadata rows. The existing independent `validate_wiring` function passed all 871 complete rendered bodies, titles, database/node identities, OPEN states, exact native parents/prerequisites and Task/M fields; all 1017 baseline protections; dependency/containment DAGs and caps; and the mandatory Warlock paths through 337 and shared roots 279/280.

The actual result is **NATIVE_WIRING_PASS_PREFIX_PENDING**, with native readback **PASS**. Final acceptance remains **UNQUALIFIED**; `final_prefix_sealed` and `runtime_qualified` remain false. Native counts are 279 incoming 100, 280 incoming 95, 337 incoming 59, 283 children 95 and 429 consumers 100. This supersedes the earlier allocation-only boundary below, while retaining that checkpoint's original bytes and claims.

| Committed actual evidence | SHA256 |
| --- | --- |
| `issue-331-verified-registry.json` | `e1e879b87b24914f08007677d30aad713c9e148363c404153bfd6be3f9564753` |
| `issue-331-verified-milestone.json` | `fe8969fcdb5b11f81c00e47c8be1a2496472a4b7bee2b32c4cc932f4ad806305` |
| `issue-331-coordinator-verified-receipt.json` | `087b2521ad5b9f8c40b278f3b23d8eea4b1b3e025ed2069285a9c3947db11d54` |
| `issue-331-native-readback.json` | `e2f47080079473c2c08a0f837b811ca0ca24225e8116866f0ba87d7e4ff1abdc` |

`issue-331-unsealed-native-validation.json` records the observation interval, API budget, original delivered paths, exact artifact hashes, caps and qualification boundary. Delivered files were copied byte-for-byte with exclusive creation; existing allocation copies and originals were reread and hashed, never rewritten. Both milestones retain canonical nodes-map digest `866246ec119bba3198ee6f6809141d7afc8efba596ad1bcadd3a8ab2194afa4c`. The coordinator's actual collected fields also pass the independent wiring validator after an in-memory REST `open`/`closed` to GraphQL `OPEN`/`CLOSED` conversion; its original receipt bytes remain exact. Producer reports and recorded expected-body hashes do not substitute for comparing actual fields.

Recheck the saved actual evidence offline, without collecting again:

```sh
rtk proxy python3 scripts/warlock_verified_proof.py
rtk proxy python3 -m unittest discover -s scripts -p 'test_warlock_verified_proof.py' -v
```

The three actual-evidence replay controls pass, including rejection of body, parent, prerequisite, Task/M and old-owner drift; allocation status cannot qualify wiring. The existing 14 approved-native controls and four allocation-binding controls also pass. Final proof still rejects the absent accepted prefix release. `--collect-once` is an explicit live mode and refuses to replace existing native evidence; do not run it for replay.

No frozen semantic manifest or proposal changed, no source generation or full source/Go/vet audit ran, and no native mutation occurred. ROOT owns acceptance in order 330 then 331 then 332, supplements, final hosted CI and merge. The accepted POST-SORCERER release remains HELD; all four appends remain root-only pending. Runtime roadmap implementation remains PAUSED. Await explicit accepted predecessor-prefix delivery before final sealing.

## Delivered allocation binding checkpoint

The immutable ROOT delivery observed at `2026-11-17T01:44:25.687122+00:00` has been independently validated offline. Its 871 typed identities cover Sorcerer 305, Warlock 281 and Wizard 285. This checkpoint is **ALLOCATION IDS_BOUND**. Native wiring, native readback and final acceptance remain **UNQUALIFIED**.

`issue-331-allocated-registry.json` and `issue-331-allocated-milestone.json` are exact-byte portable copies of the delivered files. `issue-331-allocation-bindings.json` records their original absolute paths and hashes, the canonical nodes-map digest, unchanged proposal/approval/renderer/contract pins, and all 281 own identities. Each own entry includes the expected native parent, sorted native prerequisites and SHA256 of the expected complete rendered body. These are expectations derived from frozen approval and allocated IDs; they are not observations of native fields.

| Delivered artifact | SHA256 |
| --- | --- |
| Immutable IDS_ALLOCATED registry | `73d9755cd7181155544f5d79a944bdaf653028425a75b5131b9adef1faf76392` |
| Milestone sidecar | `4ada8b216c27369df0cca9304d4ec5f4f556035362ee0c3d26ebef8d21fdf9a6` |
| Canonical complete nodes map | `866246ec119bba3198ee6f6809141d7afc8efba596ad1bcadd3a8ab2194afa4c` |

Reproduce the binding from the committed copies without accessing the original paths or network:

```sh
rtk proxy python3 scripts/warlock_allocation_binding.py
rtk proxy python3 -m unittest discover -s scripts -p 'test_warlock_allocation_binding.py' -v
rtk proxy python3 -m unittest discover -s scripts -p 'test_warlock_approved_native.py' -v
```

The binder validates strict JSON, complete exact typed identities, frozen approval/renderer/contract hashes, raw delivery hashes and stable map continuity using the existing pure functions. It compares the committed binding against deterministic derivation by default; `--write` explicitly rewrites only that derived sidecar. It never imports the native reader or follows the sidecar's live-registry path. Four actual-delivery binding controls and 14 existing synthetic approved-native controls passed. Changed bytes, coherent reassignment and altered frozen renderer inputs are rejected. Tests also confirm that IDS_ALLOCATED cannot satisfy wiring and that the default strict proof and final native proof still reject the absent accepted prefix release.

At the allocation checkpoint, ROOT reported that the first `330:creation` wiring attempt returned HTTP 502, possibly after partial native field changes. That checkpoint recorded native writes halted under `b5c5cbe912a3c2afae2d5bb66bf5026d29bd5c83e1fb2577aeb051f552a5cc50`; reconciliation belongs to ROOT. This checkpoint performs no live queries, allocation readback or native mutations, and makes no claim about those partial fields. No VERIFIED milestone had been delivered at that checkpoint; the later unsealed proof above records the subsequent delivery. The accepted POST-SORCERER prefix remains HELD and all four appends remain root-only pending. Issue #331 remains open; runtime and broader implementation qualification are unchanged.

The three proof stages have different claims:

1. **Allocation readback:** accepts only a ROOT-delivered immutable IDS_ALLOCATED registry snapshot and sidecar. It validates the complete typed identity map and independently reads real GitHub Issues, full allocation/final bodies, Task/M metadata and permitted partial wiring. Its result explicitly leaves wiring, final prefix and runtime qualification false.
2. **Wiring readback:** requires the delivered immutable VERIFIED registry snapshot, actual coordinator receipt bytes and the retained allocation delivery. It rereads and hashes both registry snapshots, recomputes the canonical stable nodes-map digest and requires identical identity maps across milestones. It then independently reads all actual native Issues and protected Project metadata. All 871 final bodies, actual database/node identities, titles, parents, exact prerequisites, OPEN states and Task/M fields must match. Every own Task remains mandatory through 337 after decision 331 closes; shared leaves also retain their 279/280 paths. This stage leaves the final prefix and runtime qualification false.
3. **Final acceptance proof:** additionally requires ROOT's accepted POST-SORCERER prefix release and the four exact ROOT-applied compact supplements. The current prefix constant is deliberately unset. No baseline, status flag, recorded hash, approval reservation or wiring result can fill that release.

The native reader has a fixed 80-call budget, at most 20 issue pages, 100 prerequisites per issue and 25 metadata issues per query. It rejects repeated issue identities or cursors, missing endpoints, truncated prerequisite/Project connections and GraphQL errors. Every subprocess has a 45-second limit. Incoming dependencies, outgoing consumers and native children each retain their 100 cap; dependency and containment cycles are checked separately. No new 429 consumer or 283 child is allowed. All old bodies, titles, parents, prerequisites and protected Project fields remain guarded. Roots 279/280/337 may receive unrelated additive prerequisites while preserving their old edges and the exact five approved owned additions; 183 remains unchanged. Historical parent-body prefixes are preserved during wiring, independently of the later final seal.

The source-only command remains:

```sh
rtk proxy python3 scripts/verify_warlock_scope.py --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf --source-class-checkpoint
```

Only after ROOT explicitly authorizes live collection at a delivered milestone, use `--allocated-registry` or `--wired-registry` with that delivery's exact immutable `registry_path`, `--milestone` with its sidecar and `--live`. Wiring also requires `--allocation-evidence` pointing to the retained own allocation readback. These modes refuse to overwrite existing milestone evidence. They do not poll or discover mutable registry progress. The default strict mode requires `--live`, actual wiring evidence and ROOT's final release.

The existing approved-native controls use synthetic native IDs; the allocation-binding controls above use the delivered immutable identities offline. They check coherent body/hash/receipt weakening, complete alias coverage, numeric types, duplicate identities, renderer whitespace/suffix changes, parent/prerequisite/Task/Size mutations, old-owner changes, lost mandatory paths, byte snapshot replacement and reassignment between milestones. Synthetic success is never recorded as actual native publication. Runtime spell effects, class admission, database behavior, browser/device behavior and FullRules remain unqualified.
