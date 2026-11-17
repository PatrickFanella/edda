# Wizard #332 final append preparation — PENDING

ROOT accepted the POST-WARLOCK release after PR #2046 head
`11f18dc15b1e214e6fe829a28177a32ebd0c39a1`, merged as
`4b703526c2a0aa05d68dd4bfe721430dc686625e`. The original release path is
`/home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/accepted-post-warlock-parent-baselines.json`.
The committed gzip copy decompresses to its exact raw SHA-256
`2a14082c35d68c6859490194738c74a6b820f81737ff6b81cd2ddf57fbbc4524`.
No main merge or source regeneration was needed for this sidecar policy.

This package is ready for ROOT's review and exclusive native application. It is
not a receipt of applied appends or final-seal success. No native requests or
writes were made during this preparation. All 285 Wizard source semantics,
allocation identities and six obligation maps remain unchanged. Runtime
implementation is PAUSED and runtime qualification remains UNQUALIFIED.

`root-append-package.json` raw SHA-256:
`5d80a055a6ed236b7dee4b50299d9617438f20702a79837b858914fdc51990b1`.
It records the exact predecessor body seals, section/result hashes, character
counts and UTF-8 byte capacities. `append-47.md`, `append-54.md`, `append-283.md`
and `append-337.md` are exact copies of Wizard's existing approved sections,
including their leading blank lines and final newline. ROOT must concatenate
each entire section to the corresponding exact released body, with no newline
normalization, body replacement or regeneration. Native application remains
ROOT-only, after ROOT review; the author does not edit issues or owner contracts.

| Parent | Prefix chars / bytes | Section chars / bytes | Result chars / bytes | Remaining chars / bytes |
| --- | --- | --- | --- | --- |
| #47 | 65,117 / 65,211 | 67 / 67 | 65,184 / 65,278 | 352 / 258 |
| #54 | 44,263 / 44,307 | 485 / 485 | 44,748 / 44,792 | 20,788 / 20,744 |
| #283 | 32,930 / 32,970 | 485 / 485 | 33,415 / 33,455 | 32,121 / 32,081 |
| #337 | 32,012 / 32,051 | 485 / 485 | 32,497 / 32,536 | 33,039 / 33,000 |

Capacities use 65,536 for both character and UTF-8 byte checks. #47's 67-character
append is below the 100-character limit; each other section is below 750.

`wizard_final_proof.py` independently pins the raw release, accepted head/merge,
all four predecessor body hashes, the append package, literal owned marker and
section text, section hashes, exact result hashes, frozen Wizard source bytes,
stable 871-node map and prior actual native identity receipt. Candidate or receipt
hash claims cannot update those expectations. Final mode requires each live
parent body to start with its exact released prefix plus exact Wizard section,
once, in that order. Unrelated later pure appends are allowed within both size
caps; duplicate or altered copies of the owned section are rejected.

The publication validator's ordinary mode retains its original strict old-body
checks. Only final mode admits the four validated body extensions. Titles,
parents, prerequisites, Task/Size, all 871 OPEN implementation nodes, old native
identities, both DAGs, caps and mandatory output paths remain checked. Accepted
330/331 decisions must remain CLOSED; #332 remains OPEN for ROOT review and merge.
No other old body exception is permitted.

The new tests use explicitly synthetic appends layered on the previously saved
actual collection. Positive synthetic tests are not publication evidence. Tests
reject rollback, missing/partial appends, changed predecessor or owned text,
duplicate and altered duplicate sections, coherent manifest/receipt/marker
rewrites, release rewrites, Unicode byte overflow, stale receipts, reopened
predecessors, old identities and owner-body drift, altered new bodies/metadata,
parent/prerequisite changes and cap weakening. Exact sections and unrelated later
pure appends pass. The saved actual unsealed receipt still fails final validation.

Run the bounded offline checks:

```sh
rtk proxy python -m unittest discover -s scripts -p 'test_wizard_*proof.py' -v
rtk proxy python -m unittest discover -s scripts -p 'test_wizard_allocation_binding.py' -v
rtk proxy python -m py_compile scripts/wizard_final_proof.py scripts/wizard_publication_proof.py scripts/test_wizard_final_proof.py
rtk git diff --check
```

Only after ROOT supplies its actual append receipt will the author collect ONE
fresh complete read-only native/metadata receipt, save it exclusively before
assertions and invoke the existing publication CLI with `--final` and the pinned
IDS/VERIFIED provenance. The CLI now refuses to overwrite collected evidence.
Final proof requires both the accepted release and actual matching native bodies;
release acceptance or package preparation alone cannot pass it. No final native
receipt or final success artifact is produced in this checkpoint.

The PR remains draft. The final hosted snapshot is reserved for the actual
sealed head. ROOT owns review, serial publication, CI and merge; no watches,
closures, deployment or runtime work are authorized here.
