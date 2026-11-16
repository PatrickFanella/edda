# Frozen caster publication coordinator

ROOT authorized one serialized writer for 871 finite source-contract Tasks: 305
for Sorcerer #330, 281 for Warlock #331 and 285 for Wizard #332. This PR closes
#330 only. The other authors retain their source decisions and acceptance proofs.
These Tasks describe future implementation PRs; publishing them grants no runtime
behavior or class admission, and every new Task remains open.

The durable approval is
`~/.local/state/edda-overnight-219965e8/caster-publication/approval.json`, pinned
independently to SHA256
`66b934f5cac58d21644822bd36098b0ee1cad535708a9e3f6b49bdf0a2f0fe41`.
The frozen renderer contract has SHA256
`e466e1b41506c35f6f03b6c3cc8d2bbbfb57e168607fa3c63954291930247238`.
All three frozen proposal hashes, 871 approved body templates and the complete
dependency/containment projection are checked before publication. The approved
source predicates, operands, costs, choices, ownership and counts stay exact.

`scripts/caster_native_publication.py` owns the mechanical stream. It records a
durable intent before each mutation and writes the registry atomically. Allocation
uses the exact marker `<!-- edda-caster-native:{alias} -->` and the unmodified
approved body. A full marker scan precedes creation; uncertain failures stop for
read-before-write reconciliation. Project auto-add can race a read: an explicit
"already exists" response permits one recovery read and an update of that item,
without another add or a polling loop.

After all identities and Task/M metadata pass readback, the coordinator publishes
`IDS_ALLOCATED`. An immutable snapshot retains its exact bytes before the live
registry advances. The milestone sidecar records both the file hash and the
canonical nodes-map hash. Wiring then resolves every approved parent and blocker
through that complete map. The final body adds only the pinned renderer's exact
metadata suffix. The coordinator never regenerates frozen source contracts.

Only five old-root additions are allowed: the three consuming class gates block
#337, and the finite three-class shared aggregate blocks #279 and #280. The world
cohort is mandatory through Sorcerer's own shared output; #183 remains unchanged.
No new direct #429 consumer or #283 child is allowed. Every historical body,
title, state, parent, prerequisite and captured Project field is preserved.

Final readback must collect all 1,888 actual nodes and all 871 new Task/M records,
reproduce every full body, bind issue/database/GraphQL identities, compare exact
parents and prerequisites, and check separate dependency and containment DAGs.
The expected root incoming counts are #279=100, #280=95 and #337=59; #429 retains
100 outgoing dependencies and #283 retains 95 children. All three caps remain
100. `VERIFIED` retains immutable registry and actual receipt bytes, with the same
nodes-map hash delivered at allocation.

`scripts/finalize_sorcerer_native.py` consumes the retained VERIFIED evidence and
binds only real identity metadata into Sorcerer's frozen 305-row source plan. It
checks the immutable delivery hashes before writing local artifacts, retains a
compressed offline evidence bundle with the original JSON bytes, and prepares
four compact supplements for ROOT. It makes no GitHub calls. A partial registry
is rejected before any local source-plan write. ROOT alone publishes the
supplements; the final live proof still requires their exact sections against the
accepted POST-ROGUE prefixes.

`scripts/sorcerer_native_binding.py` provides a separate read-only reproduction
of the literal renderer and full frozen native obligations. It does not trust
candidate body hashes, claimed publication status or a generated expected receipt
as live evidence. Its controls reject coherent wrong DC/body claims, missing
protocol/root edges, old-owner mutations, wrong metadata, unknown aliases,
duplicate identities and JSON booleans/decimals used as numeric identities.

At this checkpoint allocation is held at 714 of 871 identities (305/281/128).
GitHub returned `createIssue.issue:null` for
`332:spell-guards-and-wards-dancing-lights`, without an error or identity. A
complete 1,731-node readback found no matching marker; the exact response and
reconciliation remain in the durable journal. All 714 allocated issues were read
back as OPEN Task/M with exact allocation bodies and no final wiring. ROOT holds
retry and WIRING pending its recorded reconciliation and bounded recovery
authorization. Ordinary allocation restart cannot cross the creation fence.

Native wiring, full final readback,
ROOT-only parent supplements, Sorcerer's final prefix seal and independent PR/CI
acceptance remain pending. Existing source conflicts and unbound forms remain
unavailable. The earlier full Go race/vet results are retained; this mechanical
work uses focused Python/source checks and does not repeat those full suites.

ROOT subsequently authorized one reconciled create attempt for that exact alias,
pinned to `a0f47a7feae2d058813822cca8b4d3b7f9330d63a77ac9c6155cfd467bf82c6f`.
`recover_caster_null_create.py` verifies the archived714 map, original intent,
exact source body and fresh full/recent scans, then consumes an exclusive durable
authorization record before at most one create. Sanitized status/request/rate
headers and the exact GraphQL result are retained. A second ambiguity installs
a native-write halt; ordinary allocation cannot cross the original fence. This
operational release changes no source scope or registry schema.

The authorized attempt ran at23:04:49Z and again returned HTTP200 with
`{"data":{"createIssue":{"issue":null}}}`. Request ID:
`670A:11DBB0:2FF46FF:96603EB:6AFB8C11`; remaining primary quota2248, no Retry-After.
The authorization is consumed and `native-write-halt.json` now blocks normal
coordinator mutations. Registry remains714; no IDS milestone, wiring, root edges
or parent supplements were published. The original response and the separate
recovery response are retained. This is a true external API blocker, not final
publication proof. See `issue-330-publication-recovery.json`.
