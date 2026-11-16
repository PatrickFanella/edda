# Thief subclass catalogue (#316)

`internal/rules/srd521/testdata/character-thief.json` adapts exactly six
SRD 5.2.1 identities on PDF p. 64: Thief, Fast Hands (3), Second-Story
Work (3), Supreme Sneak (9), Use Magic Device (13), and Thief’s Reflexes
(17). Its 28 records retain the source headings, nested choices, conditions,
operands and numbers. The fixture retains the pinned source, CC-BY-4.0
attribution and adaptation notice. The PDF stays outside the repository.

The composition manifest loads canonical `character-rogue.json` once, then
this subclass. Rogue definitions are referenced without copying or editing
them. Selection references the actual Rogue subclass choice. Stealth Attack
extends the canonical Cunning Strike choice, uses its Sneak Attack dice cost
before rolling, and retains Hide’s Invisible condition only when this attack
meets the end-of-turn Three-Quarters Cover or Total Cover requirement.
Canonical strike options, limits, costs and Sneak Attack eligibility remain in
Rogue’s source graph.

Fast Hands retains the Bonus Action and both top-level options. Sleight of
Hand records the Dexterity check and the lock, trap-with-tools and pocket
alternatives; Use an Object retains Utilize or Magic to use an item requiring
that action. Second-Story Work retains Climb Speed equal to Speed and the
optional Dexterity substitution for Strength in jump distance. Use Magic
Device retains four simultaneous attunements, the 1d6 charge-conservation
roll and result of 6, Intelligence spellcasting for any Spell Scroll, reliable
cantrip/level-one casting, the higher-level Intelligence (Arcana) check at
DC 10 plus spell level, and success/failure outcomes. Thief’s Reflexes retains
two turns in the first round of any combat, at normal Initiative and
Initiative minus 10. No legacy surprise restriction is added.

The guarded contract is unchanged. Spell and item definitions remain opaque.
This batch changes no runtime, grants, inventory, activation, creation,
advancement, coverage or Full Rules support. All required mechanics remain
unavailable.

## Offline verification

```sh
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 ./internal/rules/srd521 -run Thief -count=1
rtk proxy python3 scripts/verify_srd_thief.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

The readable Go oracle independently transcribes every record field. The
standalone Python comparator reads the pinned PDF, extracts feature levels,
dice, costs, DC, initiative offset and word-valued limits, checks source
clauses, and compares the actual complete graph and source metadata. It
accepts an optional candidate catalogue path after the PDF path. Neither
proof relies on catalogue digests or heading anchors alone.

Both proofs reject the same 143 structurally valid contrary graphs: wrong existing
operands, missing predicates and options, altered conditions, numbers, costs,
levels, timing, limits, invented recovery/range and invented scaling. Every control must pass structural
validation and remain unavailable before source rejection counts. The same standalone comparator must also
accept the unmodified graph as a positive control. Separate
checks cover strict decoding, source/schema mismatches, unknown mechanics,
duplicate class/subclass imports and missing canonical Rogue. Existing fixed
Fighter/Cleric creation, projection and Cure Wounds assertions also run.

Local checks establish metadata fidelity and the tested boundaries. Hosted CI
and parent-owned independent review remain separate gates. This host’s Go
1.27.3 linter shim lacks its binary; no host tooling changes are included.
