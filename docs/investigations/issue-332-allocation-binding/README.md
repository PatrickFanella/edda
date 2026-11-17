# Wizard #332 allocation binding checkpoint

Status: **IDS_BOUND_NATIVE_READBACK_PENDING**. These sidecars bind the approved
Wizard source obligations to the actual identities in ROOT's delivered
`IDS_ALLOCATED` milestone, observed at `2026-11-17T01:44:25.687122+00:00`.
They establish no completed native wiring or runtime behavior.

`IDS_ALLOCATED.registry.json` and `ids-allocated-milestone.json` are byte-exact
portable copies. The delivery's original `registry_path` and `live_registry_path`
remain unchanged as provenance facts. Do not follow the mutable live path.
`provenance.json` records original paths, copy hashes, the stable identity map
hash, the binding file hash and pending boundaries. Offline tests read only these
copies and the committed frozen approval bundle; the original host paths need not
exist. The existing publication CLI's original-path guard remains unchanged.

`wizard-bindings.json` is the existing independent `build_binding` output for
all 285 Wizard nodes. Each entry records actual issue number, database ID and
GraphQL node ID, plus expected title, Task/Size metadata, parent, prerequisites
and final rendered body SHA-256. Parent/prerequisite/body/metadata fields are
expectations for later native readback. The registry hash inside that file uses
the existing builder's normalized JSON encoding; `provenance.json` separately
records the original delivered registry byte hash.

The six approved execution maps cover Animate Dead, Phantom Steed, Shapechange,
Simulacrum, named interoperability and the complete cross-class output graph.
Their numbers identify mandatory implementation outputs; they do not admit
runtime forms, spells, classes or Full Rules behavior. For example, Animate Dead
creation is #1820 and its Skeleton statistics and Undead Fortitude dependencies
include #1684 and #1446. Wizard's class gate is #2036.

The source manifest and approved publication bundle remain byte-exact frozen.
The first WIRING request (`wire330:creation`) reportedly returned HTTP 502 and
may have partially committed. This checkpoint records ROOT's halt hash
`b5c5cbe912a3c2afae2d5bb66bf5026d29bd5c83e1fb2577aeb051f552a5cc50`;
it neither inspects nor resolves that incident. No mutable registry, journal or
progress file was read, no native collection or GraphQL request was made, and no
native issue, relation or Project metadata was changed by this checkpoint.

No VERIFIED milestone or accepted POST-WARLOCK parent prefix was supplied.
Wiring is UNQUALIFIED, native proof and final prefix proof are PENDING, all four
parent supplements (#47, #54, #283, #337) remain PENDING and ROOT-only, and runtime
remains UNQUALIFIED. Existing `--final` refuses missing final criteria. A draft
PR for this checkpoint does not close #332.

Run the bounded offline checks from the repository root:

```sh
rtk proxy python -m unittest discover -s scripts -p 'test_wizard_*proof.py' -v
rtk proxy python -m unittest discover -s scripts -p 'test_wizard_allocation_binding.py' -v
```

The existing proof suites use synthetic fixtures for contrary controls. The new
allocation tests pin the actual delivered bytes, independently recompute the
stable map and full binding, retain frozen source/renderer checks, reject
identity replacement even after a byte rehash, and reject allocation-only native
proof, VERIFIED claims and final sealing without the parent-prefix release.
No Go/race/vet or full source/PDF audit was repeated.

Checkpoint validation passed: 20 existing publication/milestone proof tests in
16.592 seconds and five actual allocation-binding tests in 0.157 seconds.
`git diff --check` passed. These are local offline results; hosted checks and
native/runtime qualification are separate.
