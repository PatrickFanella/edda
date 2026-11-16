# Wizard class catalogue (#297)

`internal/rules/srd521/testdata/character-wizard.json` adapts the 11 source
identities in #297 from pinned SRD 5.2.1 PDF pp. 77–79. Its 86 records retain
the class, ten feature headings, all 20 progression rows, and nested choices,
prerequisites and resources. Nested records reuse their owning source heading.
The fixture includes CC-BY-4.0 attribution and an adaptation notice. The PDF
stays outside the repository.

Every table row retains proficiency bonus, cantrip and prepared counts, all nine
spell-slot columns including explicit zeros, repeated ASIs and subclass slots
at 6/10/14. Table numbers use the existing prerequisite parameter contract,
which accepts zero; ordinary feature parameters require positive numbers.
Core traits distinguish level-one and multiclass entry and retain skill and
starting-equipment choices. Evoker, spell and feat definitions remain opaque.

Spellcasting retains the Wizard list, Intelligence, focuses, cantrip learning
and replacement, spellbook dimensions and reading eligibility, initial and
level-gain spell choices, slot-level eligibility, preparation counts and changes,
and the exclusion of always-prepared spells from the preparation count.
Separate copy predicates retain found-spell costs of 2 hours and 50 GP per spell
level, own-book costs of 1 hour and 10 GP per spell level, and lost-book recovery
from prepared spells only, with the remainder requiring newly found spells.
Ritual Adept requires a Ritual spell in the book and reading from it, without
requiring preparation.

Arcane Recovery retains the Short Rest trigger, spellbook study, expended-slot
selection, combined-level budget of half the Wizard level rounded up, maximum
slot level 5, one use and Long Rest recovery. Scholar selects one proficient
skill from its six-skill list, excluding Insight. Memorize Spell retains the
Short Rest study trigger and replacement of one level 1+ Wizard spell prepared
for Spellcasting with another level 1+ spell from the book.

Spell Mastery selects one level 1 and one level 2 book spell with an Action
casting time, always prepares them, permits slotless casting at their lowest
levels, requires slots for higher levels, and permits replacing one with an
eligible same-level book spell after Long Rest study. Signature Spells selects
two distinct level 3 book spells and always prepares them. Each has its own
one-use, level-3 slotless casting resource and recovers on a Short or Long Rest;
higher-level casting requires a slot. Qualification predicates retain feat
alternatives and the recommended Boon of Spell Recall.

No shared schema extension or runtime consumer changes are needed. These records
supply no grants, derived statistics, creation, advancement, coverage or Full
Rules promotion. Readiness remains unavailable.

## Verification

```sh
rtk proxy env GOMAXPROCS=2 go test -p 2 ./internal/rules/srd521 -run Wizard -count=1
rtk proxy python3 scripts/verify_srd_wizard.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race -timeout 30m ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The readable Go oracle compares every record field, normalizing only mechanic
set ordering. The standalone Python comparator reads the pinned PDF directly,
extracts actual table cells and wrapped feature headings, and maps source
clauses to a complete expected semantic graph. It compares source metadata and
all fields, including omitted fields; it invokes neither Go nor a network
service. An optional second argument selects a candidate JSON file.

Both oracles reject 525 contrary controls. Go first requires each control to
pass structural validation, then requires semantic rejection and unavailable
readiness. Controls alter existing record operands, variants, conditions,
equipment quantities, costs, recovery, numbers and table cells, and introduce
invented scaling. Strict decoding/source/schema/unknown-key checks and duplicate
batch rejection remain separate. Canonical class, subclass, origin and feat
batches compose once each; fixed Fighter and Cleric creation/projection checks
run alongside the new assertions.

Hosted CI and independent parent review are separate gates. The active Go
1.27.3 golangci-lint shim lacks a binary; host tooling is unchanged.

The first full race run on base `5a1e408` exceeded the default ten-minute
package timeout in `TestSorcererGuardsReadinessComposition`. The command above
keeps the required concurrency and race settings while allowing the full
package to finish. This changes no repository or host timeout setting.
