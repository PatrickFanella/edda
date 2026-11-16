# Paladin source catalogue

`internal/rules/srd521/testdata/character-paladin.json` adapts the 18 identities
assigned to #292 from pinned SRD 5.2.1 PDF pages 53–55. Its 68 records include
all 20 progression rows, repeated Ability Score Improvement grants, opaque
subclass improvements, all five spell-slot columns, prepared-spell counts,
Channel Divinity counts and explicit Lay On Hands capacities at every level.
The existing guarded character contract needs no extension.

Nested metadata retains the six skill alternatives, both equipment bundles
and their quantities, proficient weapon mastery choices, Fighting Style or
Blessed Warrior and its Cleric cantrips, spell-slot eligibility, feat
qualification, and the six Restoring Touch condition alternatives. Restoring
Touch's mechanic explicitly permits one or more conditions at five pool HP
each; its choice record names the selection domain, not an active grant.
Smite and Steed retain always-prepared spell references and separate once per
Long Rest slot-free casts. Channel Divinity, Divine Sense, auras, Abjure Foes,
Radiant Strikes, costs, recovery, termination conditions and numerical limits
are recorded as opaque mechanics and parameters.

Spell lists, spells, feats, weapons and subclasses remain references. This
batch imports no Oath definition, runtime consumer, active grant, derived
bonus, creation or advancement enablement, coverage promotion or Full Rules
claim. `ValidateReadiness` continues to reject every unimplemented key.
Canonical composition tests include Fighter, Cleric, Barbarian, Bard, Druid,
College of Lore and Paladin exactly once, and reject duplicate definitions.
Fixed Fighter and Cleric creation and projection assertions remain explicit
regressions.

The source envelope retains the PDF pin, CC-BY-4.0 license, attribution and an
adaptation notice. See `docs/srd-5.2.1-conformance.md` for the project source
policy. The expected PDF SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.

The Go source graph asserts every serialized record field. The standalone
`scripts/verify_srd_paladin.py` comparator reads the pinned PDF directly,
extracts all progression table cells and numeric prose values, binds clauses
to the expected semantic graph, and compares actual metadata recursively.
It does not read Go expectations or accept a digest as semantic proof. Both
checks reject contrary existing operands, conditions, recovery, costs, ranges,
omitted alternatives or table cells, and invented scaling. Go controls also
prove that each altered catalogue remains structurally valid before rejecting
its semantics.

Run from the repository root:

```sh
rtk proxy python scripts/verify_srd_paladin.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```
