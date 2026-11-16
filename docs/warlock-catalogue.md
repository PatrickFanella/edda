# Warlock class catalogue (#296)

The offline fixture `internal/rules/srd521/testdata/character-warlock.json`
adapts exactly the 10 source identities in #296 from pinned SRD 5.2.1 PDF
pp. 70–72. Its 71 records retain the class, nine feature headings, all 20
progression rows, and nested choices, predicates and resources. Nested records
reuse their owning source heading. Source attribution, CC-BY-4.0 licensing and
the adaptation notice travel with the fixture; the PDF stays outside the repo.

Every table row retains proficiency bonus, Invocation count, cantrip count,
prepared-spell count, Pact slot count and slot level. Repeated ASIs at levels
8/12/16 and Mystic Arcanum grants at 13/15/17 reference their original source
identities; subclass slots at 6/10/14 remain opaque. Core metadata retains
Charisma, D8, saves, proficiencies, both starting-equipment alternatives and
their exact quantities, the seven-skill domain with two selections, and the
narrower multiclass grants and reference to multiclass slot rules.

Invocation metadata records the initial selection, table increases, all
prerequisites, one eligible replacement on gaining a Warlock level, the
prohibition on replacing a prerequisite for an owned Invocation, and the
no-duplicates rule with its explicit description exception. Pact of the Tome
is only a recommendation key. Actual Invocation definitions belong to
#302–#304 and are absent here. Fiend Patron, equipment, spells and feats
also remain opaque references.

Pact Magic retains the Warlock spell domains, Charisma and Arcane Focus,
cantrip and prepared-spell replacements on class level gain, table increases,
spell-level eligibility, and the exception for always-prepared Warlock spells.
Pact slots have their own resource, all at the table's slot level, expended to
cast level 1–5 spells at that slot level and recovered on a Short or Long Rest.
They are separate from ordinary Spellcasting and from Mystic Arcanum resources.
Magical Cunning takes one minute, restores expended Pact slots up to half their
maximum rounded up, and has one use per Long Rest. Eldritch Master changes
that recovery to all expended Pact slots when using Magical Cunning.

Contact Patron keeps Contact Other Plane always prepared and distinguishes
ordinary spellcasting from the free cast to contact the patron. The latter
uses its own one-per-Long-Rest resource, spends no slot and automatically
succeeds on the spell's save. Mystic Arcanum keeps four spell-level domains
(6/7/8/9), their acquisition levels (11/13/15/17), four separate free-use
resources, no slot expenditure for those uses, Long Rest recovery, and the
one-arcanum replacement on each Warlock level with a Warlock spell of the
same spell level. Feat alternatives retain qualification predicates and
Boon of Fate remains a recommendation.

The existing guarded contract suffices; no schema or production code changes
are needed. All mechanics remain unavailable, including unknown keys.
This batch enables no runtime execution, grants, derived statistics,
creation, advancement, coverage or Full Rules promotion.

## Verification

```sh
rtk proxy env GOMAXPROCS=2 go test -p 2 ./internal/rules/srd521 -run Warlock -count=1
rtk proxy python3 scripts/verify_srd_warlock.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The readable Go oracle independently transcribes the complete expected graph.
The standalone Python comparator checks the pinned PDF, extracts all 20 actual
table rows and feature cells, checks source clauses, and constructs its expected
semantic graph without calling Go or reading expectations from the fixture.
Both compare every field, including absent fields; only required-mechanic set
ordering is normalized. The comparator accepts an optional candidate JSON path.

The contrary controls change actual operands to other existing records, omit
alternatives and conditions, alter source numbers, equipment quantities,
resource recovery and table trajectories, or invent scaling. Go requires each
control to pass structural validation before both source rejection and
unavailable readiness. Strict decoding, schema/source/unknown-key tests,
canonical composition of each available batch once, duplicate rejection and
the existing fixed Fighter/Cleric creation/projection assertions are included.

Local checks establish metadata fidelity and the unavailable boundary.
Hosted CI and independent review are separate gates. The Go 1.27.3
golangci-lint shim has no installed binary on this host; host tooling is unchanged.
