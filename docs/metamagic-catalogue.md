# Metamagic source metadata (#305)

`internal/rules/srd521/testdata/character-metamagic.json` records exactly the ten
named options in the pinned SRD 5.2.1: Careful Spell on PDF p66 and Distant,
Empowered, Extended, Heightened, Quickened, Seeking, Subtle, Transmuted and
Twinned Spell on PDF p67. The 22 graph records retain the pin, attribution,
CC-BY-4.0 license and adaptation notice. No spell definitions or rule prose
are imported.

The existing `feature` kind represents each option at
`srd-5.2.1:character/feature/sorcerer/metamagic/<source-slug>`. Validation permits
this source family only with one matching inventory source, its exact source
slug, the canonical `choice/sorcerer/metamagic` parent, that choice's canonical
`feature/sorcerer/metamagic` parent, and Sorcerer class ancestry. It does not
permit arbitrary choice-owned features or introduce another option kind. The
per-source exception also requires exactly one source, so an ordinary Sorcerer
feature cannot append a secondary Metamagic source to bypass the option guards.

`character-metamagic-composition.json` composes canonical Sorcerer once and this
batch once. Every option has a predicate referencing the actual Metamagic
choice, Sorcery Points resource, Spellcasting feature and prepared-spell choice.
Twinned also references all nine actual spell-slot resources. No duplicate
resource, rest recovery, progression, preparation or casting owner is created.
Sorcerer's existing metadata retains option selection/replacement counts,
Charisma casting, Sorcery Point capacity/recovery, the ordinary one-option
limit, Sorcery Incarnate and Arcane Apotheosis exceptions.

The batch retains these option-specific conditions:

- Careful costs 1 point; on casting a spell forcing other creatures' saves,
  select up to Charisma modifier creatures (minimum one). They succeed and
  take no damage only where a successful save normally halves damage.
- Distant costs 1 point; choose doubling a range of at least 5 feet or changing
  Touch to 30 feet.
- Empowered costs 1 point when rolling spell damage; reroll up to Charisma
  modifier dice (minimum one), use the new rolls, and allow prior use of a
  different Metamagic option during casting.
- Extended costs 1 point for duration at least 1 minute; double it up to
  24 hours and grant Advantage on saves to maintain that spell's Concentration.
- Heightened costs 2 points; one target has Disadvantage on saves against the
  spell, with no invented first-save restriction.
- Quickened costs 2 points; change an action casting time to a Bonus Action
  for this casting. A level 1+ spell already cast this turn bars use; using it
  bars casting a level 1+ spell afterward on the same turn.
- Seeking costs 1 point after a missed spell attack; reroll the d20, use the
  new roll, and allow prior use of a different option during casting.
- Subtle costs 1 point; omit Verbal, Somatic and Material components except
  consumed Material components and those with a specified cost.
- Transmuted costs 1 point for a spell dealing a listed type; change that type
  to a different listed type: Acid, Cold, Fire, Lightning, Poison or Thunder.
- Twinned costs 1 point for a spell that can use a higher slot to target an
  additional creature; increase its effective level by one. This does not
  assert a changed slot expenditure or import an example spell.

All conditions and numeric fields remain uninterpreted metadata. No runtime,
grants, creation, advancement, coverage or Full Rules qualification changes.
Fixed Fighter/Cleric paths remain separate; Sorcerer remains unavailable.

## Offline evidence

The Go oracle independently transcribes the complete graph without reading
candidate expectations or running the Python builder. The standalone Python
proof reads the actual candidate JSON, checks the full PDF SHA-256, extracts
pp66–67 in column order, parses costs and numeric facts, verifies conditional
source clauses and compares complete semantic records. It also independently
checks canonical Sorcerer against pp64–66 and enforces the composition manifest.
Neither a digest of metadata nor text anchors alone constitute the comparison.

Structurally valid contrary graphs must fail both independent semantic proofs
and readiness: changed existing operands, omitted conditions or predicates,
wrong costs/numbers, missing or altered alternatives, timing and exception
changes, and invented recovery/scaling. Structural negatives separately cover
source/slug/ancestry, duplicate canonical composition, schema/pin changes,
unknown fields and unknown mechanics. Tests retain fixed Fighter/Cleric and
unavailable Sorcerer boundaries.

```sh
rtk proxy python3 scripts/verify_srd_metamagic.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf GOMAXPROCS=2 go test -p 2 -short -race ./internal/rules/srd521 -run 'Metamagic|Sorcerer|CharacterCatalogue' -count=1
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The ordinary suite needs no network, PDF or Poppler; Python performs the separate
semantic control checks. Setting `EDDA_SRD_PDF` additionally runs the real PDF
comparison. These are transcription and boundary checks, not playable support
qualification or hosted acceptance.
