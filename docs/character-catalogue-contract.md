# Character metadata contract

Issue #284 adds an offline contract in `internal/rules/srd521/catalogue.go`.
It does not load a runtime catalogue. The initial source-derived fixture is
`internal/rules/srd521/testdata/character-fighter1.json`: one Fighter class,
one level-one row, three feature headings, two choice domains and one resource.
It is a schema example, not a complete Fighter import or playable build.

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
This PR imports no prerequisite predicates or adjacent origin/feat/subclass data.

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

`TestCharacterCatalogueFighterOne` checks the sole fixture and refuses readiness.
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
