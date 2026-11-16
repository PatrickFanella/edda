# Frozen Wizard publication authority

These gzip files contain byte-exact copies of the root's approved finite
publication inputs. Decompression preserves the original JSON bytes and hashes.
They make the offline proof portable; they are not live native receipts.

The authority is `approval.json`, raw SHA-256
`66b934f5cac58d21644822bd36098b0ee1cad535708a9e3f6b49bdf0a2f0fe41`.
The independent reader pins this value in code. It independently reconstructs the
complete graph from the frozen old baseline, 871 approved node templates and the
five allowed root-edge additions, then compares the frozen projection. All 871
source templates and the own 285-node scope remain exact. Recorded candidate or
receipt hashes cannot replace these expectations.

The renderer contract's raw SHA-256 is
`e466e1b41506c35f6f03b6c3cc8d2bbbfb57e168607fa3c63954291930247238`.
The golden vectors' raw SHA-256 is
`4ba7f97dd6f1702ff049056f10e46736cb1da27fe7274553b4bff9e54077bdee`.
The vectors are partial synthetic format tests. They must fail complete registry
validation and cannot establish allocation or publication. The reader implements
the marker, byte-exact source body and suffix itself; it imports no publisher code.

Only the Sorcerer author writes GitHub issues, native relations, Project metadata,
the global registry or journal. Wizard writes local code and proofs. All actual
implementation outputs remain UNQUALIFIED. Native IDs identify future work and
supply no runtime form, spell, class or Full Rules admission.

Run the offline frozen approval and renderer check now:

```sh
rtk proxy python scripts/wizard_publication_proof.py
```

After the root delivers `IDS_ALLOCATED`, use its immutable registry snapshot and
milestone sidecar, with the exact delivered sidecar byte hash. The binder requires
`--registry`, `--ids-milestone` and `--ids-milestone-sha256`; it writes only the own
`--binding-output`. Never substitute mutable `registry.json` after a status change.
The actual paths and hashes will be recorded when delivered; no command is supplied
before that evidence exists.

After the root delivers `VERIFIED`, retain the allocated snapshot with
`--allocated-registry`, and supply the immutable verified snapshot through
`--registry`. Supply the root's `--verified-milestone`,
`--verified-milestone-sha256` and exact `--coordinator-receipt`. The verifier pins
both milestone files, independently recomputes both registry byte hashes and the
stable nodes-map digest, checks unchanged identity assignments and the collected
evidence's byte hash/observation time. It does not add registry fields. The pinned
milestone contract SHA-256 is
`58c5aacb2a3e73b759dff6b98316661fb6bd9870b95e5c7fa9a0d39e7ee736ba`.

Only then run `--collect-native` once, saving the actual collection through
`--native-receipt` and the independent result through `--proof-output`. The proof
checks all own 285 nodes and complete foreign dependencies/roots, all three caps,
both DAGs and old body/title/parent/prerequisite/Task/Size protections. It binds
actual database and GraphQL IDs to real Issues, rejects PR substitutions, partial
or duplicate identities and altered annotations, and preserves the collected raw
receipt even if validation fails. `--provenance-output` records the separate
milestone checks; status/hash agreement alone never qualifies the native graph.

That proof remains explicitly UNSEALED. The root's accepted POST-WARLOCK prefix
release and real four serial appends are still required for a final prefix seal.
`--final` refuses while that release is held. Do not rerun the collection to poll
publication. Rechecking a saved native receipt performs no network calls.
