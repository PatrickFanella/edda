# Multiclass source metadata (#320)

This batch adapts the pinned English SRD 5.2.1 PDF pp.24–26 into metadata.
It adds no executable multiclass combination, grant, creation, advancement,
action, consumer, coverage promotion or Full Rules enablement.

The source and CC-BY-4.0 attribution remain in
[the conformance record](../srd-5.2.1-conformance.md). The fixture carries the
same pin, attribution and an adaptation notice. Its PDF SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.

## Source binding

The 344 named planning identities remain unchanged. The inventory decision
explicitly assigns non-indexed section requirements to #320; these are not new
named inventory identities. `MulticlassCatalogue` is a separate, narrow schema
rather than an exception to `CharacterSource` validation. It accepts only schema
1, issue 320, section `multiclassing`, inclusive PDF pages 24–26. Every requirement
has a unique versioned mechanic key with an exact allowlisted page range:

| Requirement | PDF pages |
| --- | --- |
| Prerequisites | 24–25 |
| Experience and proficiency | 25 |
| Hit Points and Hit Dice | 25 |
| Proficiencies; class features | 25 |
| Alternate Armor Class; Extra Attack | 25 |
| Spellcasting applicability; cantrips | 25 |
| Prepared spells; spell slots | 25–26 |
| Pact Magic | 26 |

The complete 20 × 9 shared slot table is on p.26. Zero represents an em dash,
not an available slot. The separate seven-term formula retains each canonical
class-level operand, its divisor, and upward rounding for Paladin and Ranger.
Warlock is not a shared-slot formula operand. Pact Magic remains its own owner;
its bidirectional use with prepared Spellcasting spells is an exception, not a
new slot pool, prepared list, grant or recovery rule.

Class prerequisite and restricted proficiency facts use the existing named
class source links. The actual core-trait and “As a Multiclass Character” sections
are on pp.28, 31, 36, 41, 47, 49, 53, 57, 61, 64, 70 and 77. Fighter uses
Strength **or** Dexterity; Monk, Paladin and Ranger require both listed abilities.
Every new and current class must qualify. Nested Bard, Ranger and Rogue
multiclass choices reference their canonical owners without copying choices.

The ordered `catalogues` list composes each of the twelve canonical class
fixtures and the global advancement fixture exactly once. All are strictly
decoded and validated as one canonical graph before references are accepted.
Duplicate/missing catalogue paths, classes, requirements, operands and source
links fail validation. No alternate path, arbitrary section, unknown source,
wrong page, extra JSON field or trailing object is accepted. Existing named-page
and unknown-key validation is unchanged.

## Metadata and proofs

`primary_any` is OR of AND groups at `minimum_score`. Restricted proficiencies
are requirements referring to canonical owners, not automatic grants.
Requirement `clauses` retain all conditions, `record_refs` retain canonical
operands, and `parameters` retain explicit values. Global advancement values and
class Hit Dice stay with their existing owners rather than being redefined.
The AC example's Draconic Resilience and Thirsting Blade exception remain named
unresolved conditions; their subclass/invocation implementations are not imported
or enabled here. Fighter's Two Extra Attacks operand is the canonical feature.

Structural validation checks shape, source and graph integrity. It deliberately
cannot establish source truth: a known but incorrect ability, class operand or
plausible table cell may be well formed. `ValidateReadiness` always refuses every
combination, including structurally valid metadata with unknown clauses. There
is no executable interpreter or mechanic registration for these records.

Two independent proofs check source truth:

- Go expectations transcribe every class prerequisite/proficiency, exact source
  link, ownership reference, formula term, requirement clause/operand/value and
  all table cells in readable source tables.
- `scripts/verify_srd_multiclass.py` reads the actual pinned PDF with Poppler,
  extracts the twelve primary ability expressions and restricted proficiency
  paragraphs and all 180 table cells, and compares the actual candidate JSON.
  A separate transcription checks every exception field, source, scope and
  formula. PDF paragraph assertions supplement those comparisons.

The Go contradiction controls stay structurally valid and unavailable while
both proofs reject wrong known abilities, AND/OR, proficiency subsets, class
operands, rounding/divisors, mixed-die pooling, Pact exceptions, prepared-list
ownership, attack/AC stacking, table cells and invented grants/scaling. Set
`EDDA_SRD_PDF` to run the standalone comparator against each control. Ordinary
Go checks require no PDF and still run the independent semantic expectations.

```sh
GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short ./internal/rules/srd521 -run Multiclass -count=1
python3 scripts/verify_srd_multiclass.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
GOMAXPROCS=2 go test -p 2 -short -race ./...
GOMAXPROCS=2 go vet -p 2 ./...
```

The batch is one bounded source section: twelve references, twelve requirement
groups, seven formula operands and twenty slot rows. It does not expand the
named inventory or import adjacent starting equipment, higher-level creation,
Trinkets, subclass definitions or invocation option catalogues.
