# Evoker source metadata (#319)

`internal/rules/srd521/testdata/character-evoker.json` records exactly six
inventory identities on pinned SRD 5.2.1 PDF page 82: Evoker; Evocation Savant
and Potent Cantrip at level 3; Sculpt Spells at 6; Empowered Evocation at 10;
and Overchannel at 14. Twenty records retain the nested choices and predicates.
There is no Evoker progression or spell table on that page. The adjacent Wizard
spell list is outside this batch.

The composition manifest imports canonical `character-wizard.json` once and
this batch once. Subclass selection and minimum Wizard levels use canonical
operands. Savant references the existing spellbook choices. Slot eligibility
references canonical Wizard and Spellcasting: Wizard owns the 20 level rows
and their table predicates; Spellcasting owns the spell list, Intelligence,
book and preparation rules. Prepared-spell count is not a slot-level operand.
The canonical Wizard fixture is neither copied nor changed. Duplicate class
or batch imports fail structural validation.

The adapted metadata retains these source distinctions:

- Savant chooses two Wizard Evocation spells no higher than level 2 for the
  spellbook for free, then one free Wizard Evocation spell whenever access to
  a new spell-slot level in this class is gained. The later spell can be at
  any level for which the Wizard has slots; it need not match the newly gained
  level. These additions do not change preparation counts or copy costs.
- Potent Cantrip applies to a cantrip cast at a creature when the attack misses
  or that target succeeds on its save. It deals half the cantrip damage, if
  any, with no additional cantrip effect. Both trigger alternatives remain.
- Sculpt Spells optionally chooses other affected creatures the caster can
  see when casting an Evocation spell. Their number equals 1 plus spell level.
  They automatically succeed on saves against that spell; no damage applies
  only if a successful save would normally deal half damage.
- Empowered Evocation optionally adds Intelligence modifier to one damage roll
  of a Wizard Evocation spell whenever it is cast.
- Overchannel optionally maximizes damage of a damaging Wizard spell cast
  with a level 1–5 spell slot, on the turn it is cast. First use has no adverse
  effect. Reuse before finishing a Long Rest deals 2d12 Necrotic damage per
  spell-slot level immediately after casting, ignoring Resistance and Immunity.
  Each further use before finishing a Long Rest increases the damage per spell
  level by 1d12. The metadata retains the source wording distinction between
  the initial slot-level multiplier and the later spell-level increase.

No runtime consumer, grants, derived statistics, creation, advancement,
coverage or Full Rules qualification changes. All catalogue mechanics remain
unavailable. The fixture retains the pinned artifact hash, CC-BY-4.0 attribution
and adaptation notice documented in `srd-5.2.1-conformance.md`.

`catalogue_evoker_source_test.go` is a readable, complete Go graph oracle.
`catalogue_evoker_test.go` checks composition, strict schema/source handling,
unknown mechanics, fixed Fighter/Cleric paths and unavailable Wizard/Evoker
creation. Structurally valid negative graphs remove alternatives, conditions
or predicates, substitute known wrong operands, alter every number and mechanic,
and invent scaling. Both Go and Python reject those graphs after structural
validation succeeds; readiness also stays unavailable. A positive graph is
accepted first, so an always-rejecting comparator cannot pass the controls.

`scripts/verify_srd_evoker.py` is a separate standalone comparator. It requires
no Go code or network: it verifies the full PDF digest, extracts the right
column of page 82 with Poppler, parses levels and numeric operands, confirms
the qualifying source clauses, and builds an independent expected graph.
It compares every actual metadata record and field plus the source envelope
and exact inventory identities. No fixture digest or source anchors alone
substitute for that semantic graph comparison.

Run from the repository root:

```sh
rtk proxy python3 scripts/verify_srd_evoker.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./internal/rules/srd521 -run TestEvoker -count=1
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The focused Go proofs run without the PDF. Set `EDDA_SRD_PDF` to additionally
run the standalone pinned-PDF comparison from the Go fixture.
