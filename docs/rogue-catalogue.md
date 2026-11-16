# Rogue class catalogue (#294)

The offline fixture `internal/rules/srd521/testdata/character-rogue.json`
adapts exactly the 19 source identities in #294 from pinned SRD 5.2.1 PDF
pp. 61–63. Its 79 records retain the Rogue class, 18 feature headings, all
20 progression rows, and nested choices, prerequisites and resource metadata.
Each nested record reuses its owning source heading. The fixture retains the
CC-BY-4.0 attribution and adaptation notice; the PDF stays outside the repository.

The p62 progression table retains every proficiency bonus, Sneak Attack die
count and die size, repeated Expertise and ASI grant, and subclass slot at
9/13/17. Core traits distinguish the level-one and multiclass skill domains
and retain both starting-equipment alternatives with exact quantities. Expertise
selects two existing skill proficiencies at levels 1 and 6; tools are outside
that domain. Weapon Mastery selects two proficient weapon kinds and permits
replacement on a Long Rest. Thief, language tables, equipment, feats and other
referenced definitions stay opaque.

Sneak Attack retains its once-per-turn creature hit, Finesse/Ranged weapon,
weapon damage type, Advantage route and alternative ally route (within five
feet, ally not Incapacitated, no Disadvantage). Steady Aim retains the no-prior-
movement prerequisite, Bonus Action, next attack on the current turn, and
Speed zero through the turn. Cunning Strike and Devious Strikes retain all six
alternatives and their separate die costs, save abilities, durations, repeat
saves, termination conditions and restrictions. Improved Cunning Strike permits
up to two effects and pays each cost. Reliable Talent retains the skill-or-tool
ability-check domain and the 9-or-lower to 10 replacement. Stroke of Luck
retains failed D20 Test eligibility, replacement with 20, one use, and Short
or Long Rest recovery. Uncanny Dodge, Evasion, Slippery Mind and Elusive retain
their source triggers, limits, saves and conditions.

The existing guarded character contract suffices. Conditional alternative
predicates use required mechanic keys and record operands; the species-specific
selected-option field is not repurposed. No schema, runtime consumer, grants,
derived statistics, creation, advancement, coverage or Full Rules support changes
are included. All required mechanics remain unavailable.

## Offline checks

```sh
rtk proxy env GOMAXPROCS=2 go test -p 2 ./internal/rules/srd521 -run Rogue -count=1
rtk proxy python3 scripts/verify_srd_rogue.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The Go source oracle is a readable independent transcription of the complete
graph. It compares all record fields, normalizing only the required-mechanic
set order. The standalone Python comparator reads the pinned PDF, extracts all
20 progression rows and actual feature cells, and maps source clauses to the
expected graph. It separately extracts each strike variant block, die cost,
save ability and applicable numeric duration. It compares complete semantic
records and source metadata, including unexpected fields. An optional second
argument selects a candidate catalogue for independent comparison.

Both oracles reject the same 322 contrary-source controls. Go first requires
each control to pass structural validation, then requires source rejection and
unavailable readiness. The controls include existing but wrong record operands,
missing variants, omitted or contrary conditions, changed numbers/costs,
changed recovery, altered proficiency/Sneak Attack table cells, and invented
scaling. Additional checks reject unknown JSON fields and trailing data,
schema/source mismatches, malformed links, duplicate imports, and unknown-key
readiness. Canonical class/subclass batches compose once each. Existing fixed
Fighter/Cleric creation and projection assertions run alongside the new tests.

Local tests establish metadata fidelity and boundaries; hosted CI and independent
review remain separate gates. The Go 1.27.3 golangci-lint shim has no installed
binary on this host; host tooling is unchanged.
