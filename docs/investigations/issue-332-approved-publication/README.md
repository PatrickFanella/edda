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

After the root delivers `IDS_ALLOCATED`, bind the complete real registry without
performing any GitHub mutation:

```sh
rtk proxy python scripts/wizard_publication_proof.py \
  --registry /home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/registry.json \
  --binding-output docs/investigations/issue-332-native-binding.json
```

After the root delivers the `VERIFIED` wiring boundary, collect one complete fresh
native readback. This verifies the own 285 nodes, complete foreign dependencies
and roots, all three caps, both DAGs and the old body/title/parent/prerequisite/
Task/Size protections. It binds database IDs and GraphQL IDs to real Issues,
rejecting PR substitutions, partial/duplicate identities and altered annotations.

```sh
rtk proxy python scripts/wizard_publication_proof.py \
  --registry /home/onnwee/.local/state/edda-overnight-219965e8/caster-publication/registry.json \
  --binding-output docs/investigations/issue-332-native-binding.json \
  --collect-native \
  --native-receipt docs/investigations/issue-332-native-readback.json.gz \
  --proof-output docs/investigations/issue-332-published-native-proof.json
```

That proof remains explicitly UNSEALED. The root's accepted POST-WARLOCK prefix
release and real four serial appends are still required for a final prefix seal.
`--final` refuses while that release is held. Do not rerun the collection to poll
publication. Rechecking a saved native receipt performs no network calls.
