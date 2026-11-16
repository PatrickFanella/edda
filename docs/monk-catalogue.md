# Monk source metadata (#291)

`internal/rules/srd521/testdata/character-monk.json` adapts exactly the 23
Monk identities named in #291 from SRD 5.2.1 PDF pp. 49–52 into 77 metadata
records. The 20 level records retain every proficiency, feature, Martial Arts
die, Focus Point and Unarmored Movement cell. The level 1 em dashes remain
explicit absence keys, rather than invented resource capacities. The progression
table is on PDF p. 50; level records use the inventory's class anchor on p. 49.

Nested records retain skill and tool choices, equipment quantities and their
proficiency dependency, Focus actions and free/paid alternatives, redirect targets,
feat eligibility, condition choices, resources and recovery. Heightened Focus's
heading is on p. 51 and its continuation on p. 52 is retained explicitly. Numeric
action facts live on their owning Focus feature because the guarded schema does
not allow numeric parameters on choices. Predicate keys distinguish conditional
paid variants; operands identify the required resource or feature. The named
Open Hand subclass remains an opaque option, with no imported definition.

This is unavailable source metadata. Validation does not register rules, unlock
creation, grant effects, alter coverage or qualify Full Rules. The shared contract
and runtime consumers are unchanged. Pin, license and adaptation notice accompany
the data; attribution and the original artifact are documented in
[srd-5.2.1-conformance.md](srd-5.2.1-conformance.md).

The readable Go graph in `catalogue_monk_source_test.go` compares every serialized
record field. `scripts/verify_srd_monk.py` is a standalone offline comparator with
its own source-fact model. It reads the pinned PDF directly, checks source clauses,
compares complete actual metadata semantics, and extracts all 20 actual PDF table
rows to compare feature names and numeric cells. Neither proof uses a catalogue
digest. These transcriptions still require human source review; passing tests do
not establish playable rules support.

`TestMonkContrarySourceBothProofs` passes each mutated graph through structural
validation, rejects it against the Go graph, requires unavailable readiness, and
submits the same graphs to the independent Python comparator. Controls change
mechanic keys, numeric facts, prerequisite operands and numbers, choice counts and
domains, equipment quantities, resource scaling and proficiency cells. Canonical
composition loads existing class/subclass batches exactly once and rejects a
second Monk copy. Fixed Fighter and Cleric creation, projection, attack and healing
regressions remain asserted.

Run from the repository root with the pinned artifact available locally:

```sh
rtk proxy python3 scripts/verify_srd_monk.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The known golangci-lint Go 1.27.3 shim failure is an environment limitation;
no host tooling changes are part of this data PR. Independent review and hosted
CI are separate from local source checks.
