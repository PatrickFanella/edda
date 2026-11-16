# Warlock publication proof at the approved reservation

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

The three proof stages have different claims:

1. **Allocation readback:** accepts only a ROOT-delivered immutable IDS_ALLOCATED registry snapshot and sidecar. It validates the complete typed identity map and independently reads real GitHub Issues, full allocation/final bodies, Task/M metadata and permitted partial wiring. Its result explicitly leaves wiring, final prefix and runtime qualification false.
2. **Wiring readback:** requires the delivered immutable VERIFIED registry snapshot, actual coordinator receipt bytes and the retained allocation delivery. It rereads and hashes both registry snapshots, recomputes the canonical stable nodes-map digest and requires identical identity maps across milestones. It then independently reads all actual native Issues and protected Project metadata. All 871 final bodies, actual database/node identities, titles, parents, exact prerequisites, OPEN states and Task/M fields must match. Every own Task remains mandatory through 337 after decision 331 closes; shared leaves also retain their 279/280 paths. This stage leaves the final prefix and runtime qualification false.
3. **Final acceptance proof:** additionally requires ROOT's accepted POST-SORCERER prefix release and the four exact ROOT-applied compact supplements. The current prefix constant is deliberately unset. No baseline, status flag, recorded hash, approval reservation or wiring result can fill that release.

The native reader has a fixed 80-call budget, at most 20 issue pages, 100 prerequisites per issue and 25 metadata issues per query. It rejects repeated issue identities or cursors, missing endpoints, truncated prerequisite/Project connections and GraphQL errors. Every subprocess has a 45-second limit. Incoming dependencies, outgoing consumers and native children each retain their 100 cap; dependency and containment cycles are checked separately. No new 429 consumer or 283 child is allowed. All old bodies, titles, parents, prerequisites and protected Project fields remain guarded. Roots 279/280/337 may receive unrelated additive prerequisites while preserving their old edges and the exact five approved owned additions; 183 remains unchanged. Historical parent-body prefixes are preserved during wiring, independently of the later final seal.

The source-only command remains:

```sh
rtk proxy python3 scripts/verify_warlock_scope.py --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf --source-class-checkpoint
```

After ROOT delivers each milestone, use `--allocated-registry` or `--wired-registry` with that delivery's exact immutable `registry_path`, `--milestone` with its sidecar and `--live`. Wiring also requires `--allocation-evidence` pointing to the retained own allocation readback. These modes refuse to overwrite existing milestone evidence. They do not poll or discover mutable registry progress. The default strict mode requires `--live`, actual wiring evidence and ROOT's final release.

The offline controls use synthetic native IDs only. They check coherent body/hash/receipt weakening, complete alias coverage, numeric types, duplicate identities, renderer whitespace/suffix changes, parent/prerequisite/Task/Size mutations, old-owner changes, lost mandatory paths, byte snapshot replacement and reassignment between milestones. Synthetic success is never recorded as actual native publication. Runtime spell effects, class admission, database behavior, browser/device behavior and FullRules remain unqualified.
