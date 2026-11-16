# Offline spell record contract (schema version 1)

Issue #194 defines `SpellRecords`, decoded by `DecodeSpellRecords(io.Reader)` in
`internal/rules/srd521/spell_record.go`. This is offline data validation. It has no
runtime consumers, action registration, database writes or coverage promotion.
`testdata/spell_records.json` contains one complete new source record, Acid Arrow,
and adaptations of the two existing bounded fixtures. Subsequent catalogue
imports remain separately owned by #196, #202 and #218.

The envelope retains the exact `Source` pin, license, attribution and adaptation
notice checked by the conformance register. Each record has its own stable
`srd-5.2.1:content/spells/<slug>` ID, matching source ID and one-based PDF page.
The existing issue-80 inventory supplies these identities and pages. The source
range and class lists describe source facts; they do not grant playable targets
or casting abilities. See [the source pin](srd-5.2.1-conformance.md) and
[the expansion decision](investigations/issue-80-spell-item-expansion.md).

| Field | Contract |
| --- | --- |
| `scope` | `complete-source-record` or `bounded-self-level-1` |
| `level`, `school`, `lists` | Spell level 0–9, school and unique class lists |
| `casting_time`, `range` | Casting time category; feet, touch or self range |
| `target` | Single target shape, count and explicit eligibility |
| `components` | Explicit verbal/somatic flags; material description, cost in copper pieces and explicit consumption flag, or null material |
| `ritual`, `concentration`, `duration`, `duration_seconds` | Explicit flags and instantaneous or positive timed duration |
| `scaling` | None, or per-slot dice increments with base level and named effect references |
| `effects` | Named mechanic dependency, timing, dice, damage type, bonus, casting modifier and miss behavior |
| `required_mechanics` | Nonempty unique dependencies; names never register resolvers |
| `slot_levels` | Ordered unique data slot levels; bounded fixtures require exactly `[1]` |

Version 1 covers the shapes needed by these three fixtures. Reaction triggers,
long casting times, area shapes, cantrip upgrades and other spell-specific
payloads need explicit contract extensions and offline fixtures before import.
A structurally valid record is not proof of complete source transcription.
Reviewers must compare the actual source and preserve every required mechanic.
Unknown JSON fields, invalid shapes, duplicate IDs, missing source links and
invalid dependency references fail validation. Unknown mechanic names can be
retained for future planning, but remain unavailable.

Acid Arrow records the Wizard level-two Evocation, Action, 90-foot range,
V/S/powdered rhubarb leaf, instantaneous duration and single target. It retains
4d4 initial Acid damage, half initial damage only on a miss, and 2d4 damage at
the end of the target's next turn on a hit. Both damage components gain 1d4 per
slot above level two. Its dependencies remain unavailable, including ranged
spell attacks, delayed damage, miss damage, material handling and scaling.
No Acid Arrow action is added.

Cure Wounds retains source page 121, Touch range and source class lists while
restricting the fixture to conscious self and one level-one slot. Its 2d8 plus
casting modifier healing is checked against `ResolveCureWounds`. Higher-slot
healing and other targets belong to later work. Shield of Faith retains page
162, 60-foot source range, V/S/prayer scroll, Bonus Action, +2 AC and ten-minute
concentration; its fixture also stays self-only and level-one-slot-only. The
existing Cleric resolver supplies the Holy Symbol focus substitution. These
bounded records do not claim complete catalogue coverage for either spell.

`MechanicsAvailable()` checks only the explicit dependencies of those two
existing bounded fixtures. Complete source records always return false. Missing,
unknown or unsupported required keys also return false. This method is a data
inspection result; the existing resolvers still enforce preparation, components,
slots, target eligibility and lifecycle rules. No production path calls it.

Offline tests cover the exact Acid Arrow payload, bounded existing fixtures,
unavailable dependencies and unchanged partial coverage. The negative fixture
matrix in `testdata/spell_record_invalid.json` exercises malformed fields,
source/attribution/pin loss, duplicate mechanic keys and attempted target/slot
widening; tests also reject duplicate records and trailing JSON. Run:

```sh
rtk proxy go test -short -race -count=1 ./internal/rules/srd521
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
```

Source verification used the existing local PDF with SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`:
spell entries on pages 107, 121 and 162, with component rules on pages 105–106.
CI and offline tests require no PDF or network. The PDF and extracted text are
not vendored.

This work includes material from the System Reference Document 5.2.1 (“SRD
5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd.
The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.
Spell facts are adapted into structured Edda metadata; Cure Wounds and Shield of
Faith are deliberately restricted to existing fixtures. The software license
does not replace CC-BY-4.0 for the source-derived material.
