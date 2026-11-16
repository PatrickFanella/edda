# Draconic Sorcery source metadata (#317)

This batch imports exactly six identities from pinned SRD 5.2.1 PDF pp. 69–70:
Draconic Sorcery; Draconic Resilience (3); Draconic Spells (3); Elemental
Affinity (6); Dragon Wings (14); and Dragon Companion (18). Eighteen records
retain these identities, their Sorcerer-level prerequisites, choices, and operands.

Draconic Resilience records +3 maximum HP initially and +1 for each additional
Sorcerer level, plus the unarmored base AC formula of 10 + Dexterity modifier +
Charisma modifier. These formulas are unavailable metadata. Draconic Spells
retains all four preparation rows at Sorcerer levels 3, 5, 7, and 9, including
all four spells at level 3. Spell identities remain opaque mechanic references.
Elemental Affinity records exactly one choice among Acid, Cold, Fire, Lightning,
and Poison; that same choice controls Resistance and the optional Charisma
modifier bonus to one damage roll when casting a spell dealing that type.

Dragon Wings records Bonus Action manifestation on the caster, 60-foot Fly
Speed for one hour, dismissal without an action, one use recovered on a Long
Rest, and restoration by spending 3 Sorcery Points without an action. Dragon
Companion records Summon Dragon without a Material component, one slot-free
casting per Long Rest, and the separate optional modification chosen whenever
casting starts: no Concentration, with a one-minute duration for that casting
only. The spell effect and dragon stat block are outside this batch.

The external composition manifest loads canonical Sorcerer once and this batch
once. Prerequisites point to the actual Spellcasting, prepared-spell choice,
level 5 spell-slot, Sorcery Point, and Innate Sorcery owners. Canonical Sorcerer
supplies preparation exclusions, spell classification, slot and point capacity
and recovery, and Innate Sorcery rules; this batch copies none of those records.
The guarded character contract is unchanged. No runtime, grants, AC, HP, flight,
summoning, creation, advancement, coverage, or Full Rules support is promoted.
Readiness remains unavailable, including for unknown mechanics.

`catalogue_draconic_source_test.go` is the readable Go semantic graph oracle.
`scripts/verify_srd_draconic.py` independently checks the pinned PDF clauses and
complete preparation table against every actual metadata field. It also runs
the canonical Sorcerer comparator, including its 20 progression rows, nine slot
columns, conversion table, and semantic controls. Both subclass proofs reject
structurally valid contradictory operands, omitted conditions/options/table rows,
wrong costs/recovery/formulas/numbers/levels/durations, and invented scaling.
Each negative graph must validate structurally and fail readiness before source
comparison. Tests retain strict decoding, source/schema guards, duplicate
canonical/batch rejection, and fixed Fighter/Cleric plus unavailable Sorcerer
boundaries. No existing tests or default time limits are changed.

The fixture retains CC-BY-4.0 attribution and an adaptation notice. See
[srd-5.2.1-conformance.md](srd-5.2.1-conformance.md).

```sh
rtk proxy python3 scripts/verify_srd_draconic.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

Independent review and hosted CI remain parent-owned gates. The Go 1.27.3
`golangci-lint` shim has no installed binary; this change does not alter host tools.
