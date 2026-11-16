# Druid class catalogue (#289)

`internal/rules/srd521/testdata/character-druid.json` adapts exactly the 14
Druid source identities in #289: Druid and Spellcasting (PDF p41); Druidic,
Primal Order and Wild Shape (p42); Wild Companion, Druid Subclass, Ability
Score Improvement, Wild Resurgence, Elemental Fury, Improved Elemental Fury,
Beast Spells, Epic Boon and Archdruid (p43). Its 76 records include nested
metadata under those identities. No adjacent definitions are imported.

All 20 p41 progression rows retain proficiency, feature grants and repeated
ASI grants, subclass slots at 6/10/14, Wild Shape uses, cantrips, preparations
and all nine spell-slot columns. An absent slot has no positive capacity entry.
The p42 Beast Shapes rows retain known-form counts, rational CR numerators and
denominators, and the level-eight flight restriction. Magician’s extra cantrip
is separate from the class-table count. Core traits, starting-equipment bundles,
recommended spells/forms/feat, subclass and spell-list domains, Wild Shape
statistics/equipment rules, recovery, conversions, and conditional improvements
remain required unavailable mechanic keys, numbers and predicate operands.

The existing guarded character contract suffices. There are no schema changes,
runtime consumers, grants, bonuses, creation or advancement changes, coverage
promotions, Circle/Beast/spell/feat definitions or Full Rules support claims.
The catalogue retains the pinned SRD attribution, CC-BY-4.0 license and an
adaptation notice. The source PDF is not copied into the repository.

## Offline proof

```sh
rtk proxy env GOMAXPROCS=2 go test -p 2 ./internal/rules/srd521 -run Druid -count=1
rtk proxy python3 scripts/verify_srd_druid.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -short -race -p 2 ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The Go proof freezes explicit expected values for every serialized record field,
including all numeric rows, option domains, predicates/operands and links. It
compares fields with semantic diagnostics rather than comparing record hashes.
It also checks the exact inventory/page set, strict decoding, malformed links,
unknown-key readiness, composition with existing class catalogues, and the
existing fixed Fighter/Cleric creation and projection assertions.

The standalone comparator checks the PDF SHA-256, extracts every progression
cell and Beast Shapes row, and builds the expected graph through explicit
source-clause-to-mechanic mappings. It independently extracts prose numbers
such as DC, dice, ranges and spell levels per use. It compares every graph
field, rather than stopping at PDF anchor searches. It accepts an optional
catalogue path for inspecting a separate candidate file.

Both paths reject substantive contrary-source controls for a changed range,
redirected selection operand, changed damage domain and altered Short Rest
recovery. Go additionally rejects source-inaccurate slot capacity and CR values.
These controls pass structural catalogue validation in Go; they demonstrate the
difference between valid metadata and faithful source metadata. Neither proof
executes the recorded mechanics or claims playable Druid support.
