# Oath of Devotion source metadata (#314)

The batch imports exactly six identities from pinned SRD 5.2.1 PDF pp. 56–57:
Oath of Devotion; level 3 Oath of Devotion Spells and Sacred Weapon; level 7
Aura of Devotion; level 15 Smite of Protection; and level 20 Holy Nimbus.
Nineteen records retain these identities, level prerequisites, and canonical operands.

The five preparation rows retain both spells at Paladin levels 3, 5, 9, 13, and
17. Spell keys remain opaque references. Sacred Weapon retains its optional
Attack action timing, one held Melee weapon, one Channel Divinity use, ten-minute
duration, Charisma attack bonus with a +1 minimum, normal/Radiant choice on each
hit, 20-foot Bright Light plus 20 feet of Dim Light, and all three termination
conditions. Aura of Devotion retains self/allied Charmed immunity and suppression
for an already Charmed ally entering the aura. Smite of Protection retains the
Divine Smite casting trigger, self/allied Half Cover within the aura, and expiry
at the start of the Paladin's next turn. Holy Nimbus retains Bonus Action timing,
ten-minute duration, voluntary termination without an action, one use per Long
Rest, restoration with one level 5 slot without an action, saving throw Advantage
against Fiends/Undead, enemy start-of-turn Radiant damage equal to Charisma
modifier plus Proficiency Bonus, and Bright Light that is sunlight.

The external composition manifest loads canonical Paladin once and Devotion
once. Prerequisites reference the actual Channel Divinity resource, Spellcasting,
Paladin's Smite, Aura of Protection, Aura Expansion, and level 5 spell-slot owners.
Canonical Paladin supplies Channel Divinity capacity/recovery, spell preparation
rules, and the aura's scope, incapacitation condition, and level 18 expansion.
The subclass copies none of those definitions and imports no adjacent spells or
subclasses. The guarded character contract is unchanged. No runtime, creation,
advancement, coverage, or Full Rules support is granted; readiness is unavailable.

`catalogue_devotion_source_test.go` is a readable Go semantic graph oracle.
`scripts/verify_srd_devotion.py` independently transcribes the facts, reads the
pinned PDF directly, checks the preparation table and source clauses, and compares
every actual metadata field. It also runs the canonical Paladin comparator for
its full progression and operand owners. Neither proof relies on digests or
anchors alone. Both reject the same 128 structurally valid source contradictions,
including wrong existing operands, missing rows/conditions, altered costs,
recovery, numbers, ranges, durations, and invented scaling. Structural validation
must pass and readiness must fail before each contrary graph reaches the proofs.
Tests also cover duplicate canonical/batch composition, strict decoding,
source/schema failures, unknown mechanics, and fixed Fighter/Cleric and unavailable
Paladin guards.

The fixture retains CC-BY-4.0 attribution and an adaptation notice. See
[srd-5.2.1-conformance.md](srd-5.2.1-conformance.md).

```sh
rtk proxy python3 scripts/verify_srd_devotion.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

Independent review and hosted CI remain separate gates. The Go 1.27.3
`golangci-lint` shim has no installed binary; this change does not alter host tools.
