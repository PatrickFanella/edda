# Character metadata contract

Issue #284 adds an offline contract in `internal/rules/srd521/catalogue.go`.
It does not load a runtime catalogue. The canonical Fighter data is now in
`internal/rules/srd521/testdata/character-fighter.json` (#290).
`character-fighter1.json` explicitly references its level-one subset: one class,
one level row, three features, five choices and one resource. The original eight
contract identities are reused; three core-trait choices complete that subset.
The reference file is a test composition manifest, not a `CharacterCatalogue`.
Neither fixture is a playable build.

`CharacterCatalogue` has schema version 1, a pinned `Source`, and tagged
`CharacterRecord` records. The supported kinds are `class`, `level`, `feature`,
`subclass`, `origin` (background or species), `feat`, `prerequisite`, `choice`,
`resource` and `advancement`. Stable IDs use `srd-5.2.1:character/<kind>/<owner>/<slug>`;
class IDs may end at the class slug. Do not rename or reuse IDs. A different
rules version needs its own namespace and source pin.

Each record carries its name, exact one-based PDF source page and planning
inventory ID, parent ownership, explicit choice/resource/prerequisite reference
arrays and required mechanic keys. Class levels are 1–20. Features belong to
class-level or subclass records; nested choices/resources/prerequisites belong
to their referring owner. A choice has a positive selection count and mechanic
keys naming source-defined option domains. A resource has a positive capacity;
recovery rules remain required mechanic keys. A prerequisite has a required
predicate key, stable record operands and named integer parameters. Parameters
are uninterpreted metadata until a separately qualified implementation exists.
Fighter choices retain opaque feat-qualification and usable-weapon-mastery
predicates. Adjacent origin, feat and subclass definitions remain separate.

`DecodeCharacterCatalogue` rejects unknown JSON fields and trailing objects.
`Validate(repo)` checks structure, pin, attribution, adaptation notice, exact
inventory/page links, typed ownership and reference integrity. It permits
well-formed unresolved mechanic names so later metadata batches can be reviewed
without claiming implementation. `ValidateReadiness(repo)` additionally rejects
unimplemented required feature, prerequisite and choice keys. The code-owned
qualified-key function currently returns false for every key, including keys
whose names resemble existing fixed builds. JSON cannot supply a support flag.
Passing structural validation never enables creation, advancement or actions,
and neither validator changes the coverage register or Full Rules eligibility.
Existing fixed Fighter/Cleric entry points and legacy rules remain separate.

## Source and adaptation

The fixture was checked against the English SRD 5.2.1 PDF, SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`,
and `docs/investigations/issue-74-character-inventory.json`. Fighter and its
level-one table row are on page 47. Fighting Style is on page 47; Second Wind
and Weapon Mastery are on page 48. The table supplies two Second Wind uses and
three Weapon Mastery selections; Fighting Style supplies one feat choice.
The fixture stores option domains rather than copying the separate feat or
weapon catalogues. Mechanic names, IDs and graph structure are Edda additions;
names, page references, selection counts and capacity are adapted metadata.
No source prose is imported. This does not prove complete class mechanics.

The source's CC-BY-4.0 license, exact attribution and adaptation notice are
retained in the fixture's `source`. The software license does not replace that
license. See [the source pin and attribution](srd-5.2.1-conformance.md).

## Offline proof

`TestCharacterCatalogueFighterOne` checks the composed level-one fixture and refuses readiness.
`TestCharacterCatalogueRejectsInvalidMetadata` covers duplicate IDs, invalid
levels, missing/mismatched source links, license/pin/notice changes, invalid
parent/choice/resource/prerequisite links and malformed required keys.
`TestCharacterCatalogueKeysStayUnavailable` rejects unknown and unimplemented
feature/choice keys. Strict-decoder fixtures reject data-supplied availability.
`TestCharacterCatalogueDoesNotPromoteFixedBuilds` checks the two fixed creation
builds, rejects wider builds and checks unchanged partial coverage. Existing
creation, advancement, actions and legacy regression tests run with the full
short suite. There are no handler, database, UI or production changes here.

Run from the repository root:

```sh
rtk proxy go test -short -race -count=1 ./internal/rules/srd521
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
```

Source validation is offline and trusts the committed inventory. It checks
identity/page correspondence; it cannot prove that arbitrary new numeric
parameters correctly transcribe the PDF. New batches still need source review
and mechanical evidence before any runtime qualification.


## Global advancement table (#285)

`testdata/character-advancement.json` imports exactly the 20 Character level
identities from the planning inventory, all on PDF page 23. Each uses the
`advancement` kind and stable ID `srd-5.2.1:character/advancement/<level>`.
This additive schema-version-1 extension leaves `level` records class-owned.
An advancement row has no parent and has explicit empty choice, resource,
prerequisite and option arrays: the Character Advancement table lists no nested
choices. Class selection, HP gains, new class features, feat choices and ability
changes in the neighboring Gaining a Level section remain separate work.

The `advancement` object requires integer `experience_points` and
`proficiency_bonus` values. Pointer fields distinguish absent/null values from
level one's valid zero XP. Validation rejects negative XP, nonzero first-level
XP, zero XP at later levels, bonuses outside 2–6, invalid levels, duplicate
levels, wrong inventory identities/pages and numeric fields on other kinds.
Exact source values and complete 1–20 coverage are asserted by the batch fixture;
structural validation can still accept partial batches for later composition.
The two required keys are `srd-5.2.1:mechanic/character/experience-threshold`
and `srd-5.2.1:mechanic/character/proficiency-bonus`. They remain unimplemented;
recording the values does not execute threshold advancement or bonus changes.
No class, feature, choice, playable registration or runtime consumer is added.
The fixture retains the pinned source's attribution and an adaptation notice.

`TestCharacterAdvancementSourceTable` asserts the exact IDs, names, pages and
20 numeric rows, rejects readiness for each row and checks unchanged runtime
registration and Full Rules ineligibility. Negative fixtures cover missing/null
numbers, invalid values/types, source/owner mismatches, required keys and
availability fields. Existing class/choice/source guard fixtures still run.

To recheck the source locally without network access:

```sh
rtk proxy python3 scripts/verify_srd_advancement.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy go test -short -race -count=1 ./internal/rules/srd521
```

The verifier requires Poppler's `pdftotext`, checks the full artifact SHA-256,
extracts the three table columns on page 23 and compares all 20 rows with the
fixture and pinned inventory. Normal Go fixtures use an independent numeric
transcription and need neither the PDF nor Poppler. These checks establish data
completeness, not runtime advancement qualification. No database, handler or UI
mutation is involved.


## Fighter catalogue (#290)

`character-fighter.json` records exactly Fighter and its 15 named class feature
identities, linked to the inventory on PDF pages 47–48. Its 20 class-owned
`level` rows preserve the p47 Fighter Features table. The `progression` object
records proficiency bonus, feature references and unresolved feature keys.
Repeated ASI, Action Surge and Indomitable grants reference their original
feature IDs. Subclass improvement slots at 7/10/15/18 require an unavailable key;
Champion is only a named subclass choice domain. No Champion feature is imported.
Global `advancement` records and their XP thresholds remain independent.

The additive schema-version-1 fields are metadata only:

- `values_by_level` records each applicable resource capacity or choice count.
  Second Wind is 2/3/4 uses at levels 1/4/10. Weapon Mastery is 3/4/5/6 selections
  at levels 1/4/10/16. Action Surge is 1/2 uses at levels 2/17; Indomitable is
  1/2/3 uses at levels 9/13/17. Every applicable level has an explicit value.
- `options` retains the A/B/C starting-equipment bundles and exact quantities.
  Core skills retain all nine options and a selection count of two. Fighting
  Style, mastery, subclass, ASI, Tactical Master and Epic Boon retain their
  source domains and selection counts. Recommendations and replacement rules
  have mechanic keys. Referenced feat and weapon catalogues are not copied.
- `parameters` retains the d10 Hit Point Die and the 2/3/4 attacks of the three
  Extra Attack features. Each numeric or quantity key must also appear in
  `required_mechanics`. These values never change a character or action.
- A `prerequisite` may now belong to a choice. Existing owner-reference,
  same-source, duplicate, predicate and parent-cycle checks still apply.
  Choice/resource parenting otherwise retains the original restrictions.

Progression grants must resolve to features of the same class at the current
or an earlier level. Counts require positive values, strictly increasing levels
within 1–20 and an initial value matching `select` or `capacity`. Equipment
alternatives must name a declared option key, retain positive quantities and
require every quantity key. Readiness continues to reject every mechanic key.

`catalogue_fighter_test.go` asserts exact identities/pages, all 20 progression
rows, repeated grants, resource and mastery counts, nested qualification,
equipment alternatives and canonical level-one composition. Negative fixtures
exercise malformed progression, future/missing grants, counts, parameters,
quantities, unavailable keys and JSON-supplied support. The original guard
fixtures and global advancement tests continue to run.

For direct source comparison, run:

```sh
rtk proxy python3 scripts/verify_srd_fighter.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
```

The verifier checks the artifact SHA-256, all 16 inventory identities/pages,
all 20 extracted table rows including feature labels and resource annotations,
and text anchors for nested source choices. Go fixtures independently check
exact choice domains, quantities and predicates. The source was also read
locally on pp47–48. These checks establish this data batch's completeness;
they do not evaluate feat eligibility, resolve adjacent catalogues or qualify
creation, advancement, actions, Champion mechanics or Full Rules. The catalogue
retains CC-BY-4.0 attribution and its adaptation notice.


## Champion catalogue (#312)

`testdata/character-champion.json` adds exactly seven source identities from PDF
page 49: Champion; Improved Critical and Remarkable Athlete at level 3;
Additional Fighting Style at 7; Heroic Warrior at 10; Superior Critical at 15;
and Survivor at 18. It uses the existing schema without changing validation.
`character-champion-composition.json` explicitly composes the canonical Fighter
fixture and this batch for offline tests. The batch requires that composition
to resolve its Fighter parent and class-level prerequisite operands. It copies
no Fighter records and does not connect Champion to runtime progression slots.
Global advancement remains separate.

Each feature has an opaque Fighter minimum-level predicate. The additional
Fighting Style choice selects one further distinct feat from the existing
`feats/fighting-style` domain; adjacent feat definitions are not imported.
Critical metadata retains weapon and Unarmed Strike applicability, inclusive
19–20 and 18–20 d20 ranges, and Superior Critical's replacement of the earlier
range. Remarkable Athlete retains Advantage on Initiative and Strength
(Athletics), plus optional movement up to half Speed immediately after a
Critical Hit without provoking Opportunity Attacks. **Page 49 has no jump
benefit.** The metadata follows this pinned edition.

Heroic Warrior records optional self-granted Heroic Inspiration during combat
at the start of the character's turn, only when absent. Its nested resource has
capacity one to describe that absence condition and no stacking; it is not a
rest-based use pool. Survivor records both named benefits through mechanic
keys: Defy Death gives Advantage on Death Saving Throws and the benefit of a
20 on an 18–20 roll; Heroic Rally restores 5 plus Constitution modifier at the
start of each own turn only while Bloodied and with at least 1 Hit Point.
These are uninterpreted parameters and conditions, never executable effects.
The batch retains the pin, CC-BY-4.0 attribution and its own adaptation notice.

`TestChampionSourceAndImprovements` checks exact identities/pages, all feature
levels, numeric details, choice domain/count, resource timing/capacity keys,
conditions and unchanged canonical Fighter records. `TestChampionMetadataGuards`
rejects malformed ownership, sources, references, counts, numeric keys and
predicates, checks unknown-key readiness and rejects JSON-supplied availability.
It also checks fixed Fighter/Cleric creation, unavailable Champion creation and
unchanged coverage. Existing parent, source, reference, count, readiness and
global advancement fixtures remain in the full suite.

Run the source comparison offline:

```sh
rtk proxy python3 scripts/verify_srd_champion.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
```

The verifier checks the full PDF hash, crops page 49 to the Champion column,
compares the seven inventory identities/pages, six extracted feature levels,
critical ranges, movement fraction, Survivor numbers and choice/condition text
anchors. Go fixtures independently assert exact mechanic keys. These checks
establish metadata transcription; they do not qualify playable Champion,
feat eligibility, creation, grants, advancement, combat or Full Rules.


## Background catalogue (#306)

`testdata/character-backgrounds.json` adds exactly four canonical `origin`
identities: Acolyte, Criminal, Sage and Soldier, all on PDF page 83. Fourteen
owned choice and prerequisite records retain their nested metadata. It uses the
existing schema with no validator or runtime changes. The source retains the
full PDF pin, CC-BY-4.0 attribution and adaptation notice.

Each background records its exact three ability scores through seven complete
alternatives: the six ordered +2/+1 distributions to distinct listed scores,
and +1 to all three. A choice selects one distribution. Quantities record the
increases; an owned prerequisite refers to that choice and records the maximum
resulting score of 20. These are the page 83 background rules, also confirmed
in the creation instructions on page 21. They do not impose a minimum score.

Required keys retain each specified feat, both skills and the fixed tool
proficiency. Magic Initiate retains its Cleric or Wizard reference. Soldier
instead owns a one-kind-of-Gaming-Set choice. A prerequisite on its equipment
choice records that package A uses that same selected kind. Choice records are
siblings owned by the origin; the predicate links their dependency. The Gaming
Set domain and the referenced feats, tools and items remain opaque keys. This
batch imports no adjacent definitions or feat-internal choices.

Equipment alternatives preserve the entire package A or 50 GP for package B.
The packages retain prayer/history book distinctions, parchment sheet counts,
two Criminal daggers and pouches, Soldier’s 20 arrows and all other item counts.
Package A currency is 8/16/8/14 GP for Acolyte/Criminal/Sage/Soldier respectively.
No starting inventory, currency, grants or fixed Human Soldier/Acolyte creation
behavior changes. Every catalogue mechanic remains unavailable; structural
validation does not promote coverage or qualify Full Rules.

`catalogue_backgrounds_test.go` independently transcribes the four IDs/pages,
all ability distributions, exact feat/skill/tool keys, both equipment bundles,
and both prerequisite predicates and operands. It composes advancement, Fighter,
Champion and backgrounds once each, validating duplicate IDs, ownership and
references across the complete canonical graph. The level-one Fighter subset
manifest is a reference, not another canonical catalogue to concatenate.
Negative cases cover source/pin/notice changes, malformed and unknown keys,
missing or wrong-owner references, quantities, counts, prerequisites and
JSON-supplied availability. Unknown well-formed mechanics pass structural
validation but fail readiness. Fixed creation snapshots, including existing
currency, and unchanged coverage are checked around background validation;
the existing creation tests additionally assert derived statistics and feats.

Run from the repository root:

```sh
rtk proxy python3 scripts/verify_srd_backgrounds.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
```

The comparator verifies the full PDF hash, reads the two page 83 columns and
compares the four inventory identities/pages, parsed ability/feat/skill/tool
labels, enumerated distributions, maximum-score rule, equipment quantities and
currency, and Soldier’s matching choice. Go fixtures require neither the PDF
nor Poppler. These checks establish metadata transcription and graph integrity;
independent source review and hosted qualification remain separate. Database,
browser and paid inference checks are outside this metadata-only scope.
