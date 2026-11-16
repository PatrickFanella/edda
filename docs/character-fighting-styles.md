# Fighting Style feat catalogue (#299)

`internal/rules/srd521/testdata/character-fighting-styles.json` contains exactly
four canonical feat identities and four owned prerequisite records. Archery is
on PDF page 87; Defense, Great Weapon Fighting and Two-Weapon Fighting are on
page 88. Stable IDs use `srd-5.2.1:character/feat/<slug>` and
`srd-5.2.1:character/prerequisite/<slug>/fighting-style-feature`.

The source is the English SRD 5.2.1 PDF with SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The fixture retains the source pin, CC-BY-4.0 license, full attribution and an
adaptation notice. Names, pages, values and conditions are adapted source
metadata; IDs, predicate names and mechanic keys are Edda additions. See the
[source attribution and pin](srd-5.2.1-conformance.md).

Each feat retains the Fighting Style category through the required key
`feats/category/fighting-style` and the default maximum of one acquisition from
Parts of a Feat on page 87 through `feats/maximum-times-taken: 1`. None of these
four descriptions supplies a Repeatable exception. All mechanic names below
have the `srd-5.2.1:mechanic/` prefix.

Each owned prerequisite uses `prerequisite/has-fighting-style-feature`, with
explicit empty record operands and numeric parameters. This means the general
source condition, possession of the Fighting Style feature. It does not require
Fighter levels or refer to Fighter's particular feature record. An empty operand
list does not waive the opaque predicate. Referenced classes, features, weapon
categories and properties remain unavailable mechanic references.

| Feat | Retained source benefit and conditions |
| --- | --- |
| Archery | +2 to attack rolls made with Ranged weapons. |
| Defense | +1 Armor Class while wearing Light, Medium or Heavy armor. |
| Great Weapon Fighting | When rolling damage for one's own attack with a Melee weapon held with two hands, optionally treat any damage-die result of 1 or 2 as 3. The weapon must have the Two-Handed or Versatile property. This is the 5.2.1 die treatment, with no reroll. |
| Two-Weapon Fighting | Optionally add the ability modifier to damage of the extra attack resulting from using a Light-property weapon, only if the modifier is not already being added to that damage. |

Required keys retain all listed conditions, including the armor/property
alternatives, optional benefits and the extra-attack restriction. Parameters
retain the Archery and Defense bonuses, both endpoints of the Great Weapon
Fighting die-result range and its replacement value. No additional feat choice,
resource, turn limit, off-hand condition or weapon definition is invented.

The only contract change permits the existing positive numeric `parameters`
map on `feat` records, alongside `class` and `feature`. Every numeric key must
still appear in required mechanics. Other record kinds, source ownership,
reference integrity, explicit arrays, choice counts and readiness guards retain
their existing checks. No runtime consumer reads this fixture. Validation does
not grant feats, select styles, alter inventory or derived statistics, enable
actions, change coverage, or qualify Full Rules. Fixed Defense and the ordinary
one-hand longsword continue through their existing runtime paths.

## Verification

Run from the repository root:

```sh
rtk proxy python3 scripts/verify_srd_fighting_styles.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
rtk proxy git diff --check
```

The PDF comparator checks the entire artifact hash, extracts the relevant
columns on pages 87–88, compares the complete four feat descriptions, and checks
exact inventory identities/pages, category, prerequisite key and operands,
acquisition limit, numbers and all condition keys. Go tests independently
transcribe the eight records and run without PDF or Poppler. Contrary-source
mutations test wrong bonuses, armor alternatives, old reroll behavior, die
treatment, missing hand/property restrictions, duplicate modifier addition,
Fighter-only prerequisites and repeatability. Structural validation intentionally
permits well-formed future mechanics; the source oracle rejects these contrary
values even when they are structurally valid.

Invalid fixtures cover pin, attribution/adaptation, source pages/ownership,
duplicate IDs, missing/orphan/wrong-owner prerequisites, unknown operands,
unrequired keys, invalid numbers and unknown JSON fields. Well-formed unknown
mechanics remain unavailable. Composition includes advancement, Fighter,
Champion, Cleric and backgrounds exactly once, without concatenating the
level-one Fighter subset manifest. Runtime checks compare fixed Fighter/Cleric
creation before and after validation, assert existing Defense AC and raw
longsword damage results of 1 and 2, and check unchanged coverage and failed
Full Rules eligibility. The full suite retains existing count/choice/conditional
and fixed-runtime regressions.

These checks establish metadata transcription and local regressions. Independent
review and hosted qualification belong to the parent task. Database, browser,
paid inference, deployment and production changes are outside this batch.
