# SRD 5.2.1 species catalogue

Issue #307 imports nine species into offline character metadata: Dragonborn,
Dwarf and Elf (PDF p84); Gnome and Goliath (p85); Halfling, Human, Orc and Tiefling
(p86). Elf continues on p85 and Goliath continues on p86. References identify the
inventory heading's starting page. The creation rules on pp19–22 and the species
contract on pp83–84 establish selection, creature type, size, speed and automatic
traits. This batch adds no background ability increases or language choices to
species; those are separate creation rules.

The canonical data is `internal/rules/srd521/testdata/character-species.json`.
Composition appends its decoded records to other canonical batches in memory;
there is no copied composition manifest. The exact-record transcript in
`species-source-facts.txt` is a test expectation, never a catalogue input.

## Representation

- Nine `origin` records own every choice, resource and prerequisite. Ancestries
  and lineages are alternatives within these identities.
- `species` retains category, creature type, speed, approximate height ranges,
  size alternatives and typed traits. Fixed traits and spell grants are not
  fabricated feature choices.
- Traits carry required opaque mechanic keys, exact numeric parameters and
  conjunctive prerequisite references. Selected-option predicates retain an exact
  choice operand and option key. Nested choices inherit their owning selection's
  scope; explicit predicates retain conditional branches such as High Elf's
  cantrip replacement and Rock Gnome's device effects.
- Spell grants retain character levels 1, 3 and 5, fixed versus selected status,
  preparation, casting ability selection, appropriate-slot casting, free uses
  and recovery. They import no spell definitions. Human's Origin feat selection
  and recommendation similarly import no feat definitions.
- Dragonborn's damage rows use `character_levels` at 5, 11 and 17. Other level
  gates explicitly say `minimum-character-level`. Neither field uses class levels
  or imports the global advancement table.
- Proficiency Bonus resource maxima use `capacity_mechanic` with zero in the
  unused fixed-capacity field. The only allowed symbolic maximum is the required
  Proficiency Bonus mechanic. Fixed pools retain positive integer capacities.
- Gnome device choices reference the opaque Prestidigitation effect domain and
  its selected effect's option domain. The source's explicit ignite/extinguish
  example retains both alternatives under a predicate on the chosen effect.
  This does not import the spell's definition or claim its effect domain is
  available. Device duration, AC, HP, activation, dismantling and simultaneous
  limit remain typed numbers and required keys.

## Validation and limits

`TestSpeciesExactSourceFacts` compares every decoded record against the reviewed
source transcript, including full option sets, predicates, operands, numerical
parameters, character-level rows, resource references and ownership. Schema
negative cases cover source corruption, wrong ownership, invalid types and
numbers, invalid option predicates, resource expressions and unknown JSON fields.
Readiness rejects unknown mechanics; catalogue validation grants no availability.

Run the pinned-PDF comparator from the repository root:

```sh
rtk proxy python scripts/verify_srd_species.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy go test -short -race ./...
rtk proxy go vet ./...
```

The comparator checks the PDF hash, inventory identities/pages, complete numeric
metadata, lineage spell tables, ancestry damage mappings and character-level
rows, plus edition-specific clauses on all three pages. The Go transcript test
covers the full metadata, including symbolic trigger/recovery and nested choices.
Fixed Human Fighter/Soldier and Human Cleric/Acolyte regression fixtures remain
part of the suite. Species validation has no runtime consumer and does not enable
creation, grants, derived statistics, actions, starting inventory, coverage or
Full Rules support. Unimplemented mechanics and referenced definitions remain
unavailable.

## Source and attribution

Pinned artifact SHA-256:
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
See [the conformance source pin](srd-5.2.1-conformance.md).

This work includes material from the System Reference Document 5.2.1 (“SRD
5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd.
The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.
Species facts have been adapted into structured metadata; keys, predicates,
validation and test representations are Edda additions.
