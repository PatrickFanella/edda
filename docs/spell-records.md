# Offline spell record contract (schema version 1)

Issue #194 defines `SpellRecords`, decoded by `DecodeSpellRecords(io.Reader)` in
`internal/rules/srd521/spell_record.go`. This is offline data validation. It has no
runtime consumers, action registration, database writes or coverage promotion.
`testdata/spell_records.json` contains one complete new source record, Acid Arrow,
and adaptations of the two existing bounded fixtures. Subsequent catalogue
imports are separately owned by #196 and #202; the #218 batch is described below.

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

The initial version covers the shapes needed by these three fixtures. The #218
extensions below add reaction triggers, long casting times and area shapes.
Cantrip upgrades and other new spell-specific payloads still need explicit
contract extensions and offline fixtures before import.
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

## Issue #218: Scorching Ray through Shield of Faith

`testdata/spells_218.json` contains exactly twelve complete source families from
pinned PDF pages 159–162. `spells_218_test.go` checks their names, initial source
pages, headers, materials, effects, scaling, mechanic dependencies and nested
choices. Seeming continues onto page 161; Shield continues onto page 162. The
source PDF hash was checked against the pin above. No PDF or extracted text is
needed by the offline tests.

This batch extends schema version 1 with optional, typed `source_details` on
complete source records. It adds only shapes present in this batch: a ten-minute
cast, conditional Bonus Action and Reaction triggers, unlimited range, selected
creatures, a Sphere, rays, alternative targets, indefinite duration and per-slot
ray count scaling. `scaling.kind = slot-count` uses zero dice fields and the typed
`rays.additional_per_slot` count; its effect reference names the additional ray.
The existing `slot` shape still represents damage dice additions. A timed round
is recorded as six seconds, with Shield's exact start-of-next-turn boundary
retained separately in `end_timing`.

Each selected detail payload requires every field, including explicit false,
zero and list values. There is no catch-all prose or arbitrary parameter map.
The typed payloads retain:

- Scorching Ray's three attacks, one/several-target choice and extra ray per slot.
- Scrying's creature and seen-location modes, all three knowledge and three
  physical-connection modifiers, success lockout and sensor behavior.
- Searing Smite's melee weapon/Unarmed Strike triggers, initial and recurring
  damage, after-damage Constitution save and scaling of both damage effects.
- Secret Chest's separate 5,000 GP and 50 GP component minima, dimensions,
  capacity, Ethereal storage, Magic actions, daily cumulative decay and endings.
  The common material field records the combined 5,050 GP minimum; the typed
  fields retain both independently required objects. Neither is consumed.
- See Invisibility's creature/object perception and ghostly Ethereal view.
- Seeming's per-target appearance choices, height/weight/limb constraints,
  unwilling-target save, physical inspection and Study discovery check.
- Sending's two addressing choices, 25-word message and immediate like reply,
  cross-plane delivery failure, sender feedback and recipient blocking.
- Sequester's object/willing-creature choices, concealment, creature suspension,
  waived sustenance and aging, chosen early-end condition and damage ending.
- Shapechange's seen-form and CR restrictions, first-form temporary Hit Points,
  full retained-statistics list, retained Spellcasting and equipment choices.
- Shatter's ten-foot Sphere, Constitution save/half damage, Construct
  Disadvantage, eligible nonmagical objects and per-slot damage scaling.
- Shield's two reaction triggers, +5 AC including the triggering attack,
  Magic Missile protection and exact next-turn expiration.
- Shield of Faith's chosen creature within 60 feet, +2 AC and ten-minute
  concentration, with all source slot levels retained as unavailable metadata.

The full Shield of Faith source record is separate from the existing bounded
fixture. `ComposeSpellSourceCatalogue` validates its inputs, excludes bounded
compatibility fixtures and rejects duplicate complete source IDs. Composing the
#194 and #218 fixtures yields thirteen distinct source records: Acid Arrow plus
this batch. The bounded fixtures keep their existing IDs and exact availability
checks; added detail payloads cannot make those checks pass.

All twelve source records remain unavailable regardless of their dependency
keys. This batch supplies data for later effect-instance work (#351); it adds
no resolver, action, entitlement, tool, inventory or coverage claim. Negative
fixtures cover omitted detail fields, missing dependencies, invalid variants,
costs, chances, triggers and scaling references, unknown dependencies and
attempts to attach source metadata to bounded compatibility records.

The source-derived JSON retains the pinned source, CC-BY-4.0 attribution and an
adaptation notice. It is an adaptation into structured metadata under the source
license described above; the repository software license does not replace it.

## Issue #198: Barkskin through Burning Hands

`testdata/spells_198.json` imports the exact twelve families on PDF pp.112–114.
Bestow Curse retains its p.113 continuation and Blink its p.114 continuation in
explicit `source_pages` lists. All records retain the pinned source attribution
and adaptation notice. `spells_198_test.go` supplies literal source-value fixtures
for every header, choice, effect, dependency and scaling value.

The typed extensions retain Barkskin's AC 17 and willing target; Beacon of Hope's
Wisdom/Death save Advantage and maximum healing; Befuddlement's damage, Magic
restrictions, 30-day repeat save and three restoration alternatives; all four
Bestow Curse choices, six ability options and five slot-duration bands; Black
Tentacles' Strength saves, once-per-turn triggers and Athletics escape; both Blade
Barrier forms, dimensions, cover and Force damage; Bless's unconsumed Holy Symbol
worth at least 5 GP and target scaling; both Blight target modes; both
Blindness/Deafness conditions and repeat saves; Blink's d6, Ethereal interaction
and return cases; Blur's perception exceptions; and Burning Hands' Cone damage
and eligible object ignition.

`slot-targets` references the typed additional-target count for Bless and
Blindness/Deafness. `slot-duration` references Bestow Curse's slot table, including
concentration removal at slot 5 and indefinite duration at slot 9. Zero seconds
with `until_dispelled=true` is explicit source data. Curse choices use `none` or
`not-applicable` for fields outside the selected choice rather than omitting
required values. These fields are metadata; no effect resolver reads them.

Every new payload field is required recursively. Null, omitted values, blank
labels, empty required lists (including `[ ]`), mixed entries, duplicate keys,
case aliases and unknown fields fail offline decoding. Required common numeric
fields also distinguish omission/null from explicit zero. Negative Scrying
modifiers in the earlier batch remain supported. Composition of #194, #218 and
#198 yields 25 unique complete source records and excludes both bounded fixtures.
Their exact availability checks are unchanged; every imported record remains
unavailable. No runtime, inventory, tool, action, coverage or Full Rules behavior
is added.

To recheck the PDF pin, inventory identities/pages and edition-specific source
anchors, run:

```sh
rtk proxy python3 scripts/verify_srd_spells_198.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
```

The source check supplements the full offline literal fixtures; it is not an
automatic conformance proof. The PDF and extracted text are not vendored. No
network, inference, UI or database tests are needed for this data-only batch.
