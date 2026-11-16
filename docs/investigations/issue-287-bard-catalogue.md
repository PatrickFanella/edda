# Bard source catalogue (#287)

`internal/rules/srd521/testdata/character-bard.json` is the canonical offline
Bard batch. It records Bard and the twelve named class features from SRD 5.2.1
PDF pages 31–33. Source attribution, CC-BY-4.0 license and an adaptation notice
travel with the data. The pinned artifact SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.

The 20 class rows preserve proficiency, feature references, the Bardic die and
all table-only improvements. Cantrip/preparation counts and all nine slot
columns are separate level-count records. Absent slots have no positive count
at that level; the tests and PDF comparator assert their zero interpretation.
Level 6 and 14 subclass entries remain opaque mechanic keys. Expertise at
level 9 reuses its canonical feature and increases cumulative skill selections
to four. No College feature definitions are included.

Core and multiclass acquisition paths, skills, instruments, alternative starting
equipment and the instrument choice nested inside equipment A remain structured.
Prerequisites retain unevaluated keys, record operands and numeric parameters.
Spell lists, recommended spells/feats and always-prepared Power Word references
are opaque; none resolves to executable spell or feat definitions here.
Magical Secrets records its four list domains, new-preparation/replacement
conditions and Bard classification without imposing an invented fixed grant.

Bardic Inspiration uses `capacity_mechanic` with
`bard/inspiration/charisma-modifier-minimum-one`. Its numeric capacity stays zero
and it has no `values_by_level`: the source gives a formula, not a fixed or
class-level-scaled maximum. The schema adds only this formula to its existing
allowlist. Recovery keys distinguish the initial Long Rest recovery from level
5 Short Rest recovery and slot conversion, and level 18 initiative recovery to
two uses. These keys describe unavailable behavior; they do not spend, recover,
grant or project resources.

## Criterion to proof

| Requirement | Proof |
| --- | --- |
| Exactly 13 source identities and pages | `TestBardIdentitiesRowsAndSymbolicResource`; comparator cross-checks the pinned inventory |
| All 20 rows, features and all slot columns | Independent numeric rows in that test; comparator parses the PDF table directly |
| Nested options, exact numbers, predicates and operands | `TestBardExactNestedChoicesNumbersAndPredicates` plus the complete source digest in `TestBardExactSourceAndContraryControls` |
| Complete required-mechanic lists, no invented scaling | Whole-record source pin, deletion controls for every required key including prerequisites, and contrary capacity/recovery/target/selection controls |
| Guarded contract and unknown behavior | Formula, future-feature, missing-owner, bad-page, unknown-operand and unknown-field controls; readiness rejects unknown keys |
| Canonical composition without duplicates | `TestBardCompositionAndUnavailable` combines each canonical batch once, validates it and rejects repeated Bard records |
| No playable or Full Rules promotion | Composition readiness and FullSupportEligible remain false; existing fixed-build, creation, advancement and source regressions run with the package/full suite |

## Offline verification

Run from the repository root, using an existing copy of the PDF:

```sh
rtk proxy python scripts/verify_srd_bard.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./internal/rules/srd521 -count=1
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
rtk git diff --check
```

The comparator hashes the supplied PDF before extracting pages. It checks every
progression cell, canonical identities/pages and prose anchors for core traits,
choices, feature triggers, conditions, quantities and recovery. Go tests freeze
the complete imported source metadata and exercise schema and readiness guards.
Neither check downloads a source or calls an inference provider.

No runtime consumer, creation/advancement enablement, grants, derived bonuses or
coverage promotion changes are included. Local checks do not establish hosted
qualification, which belongs to the parent review. The installed `golangci-lint`
mise shim does not provide a binary for active Go 1.27.3; its attempted invocation
fails before analysis. No host/toolchain changes were made for that limitation.
