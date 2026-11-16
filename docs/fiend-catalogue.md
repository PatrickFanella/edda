# Fiend Patron source metadata (#318)

This batch adapts exactly six identities from pinned SRD 5.2.1 PDF p. 76:
Fiend Patron; level 3 Dark One’s Blessing and Fiend Spells; level 6 Dark One’s
Own Luck; level 10 Fiendish Resilience; and level 14 Hurl Through Hell.
The 17 records include prerequisites and a damage-type choice.

Dark One’s Blessing retains both enemy-to-zero-HP triggers, the 10-foot range
for someone else’s kill, and self Temporary Hit Points equal to Charisma modifier
plus Warlock level, minimum one. Fiend Spells retains all ten spell references
in four always-prepared rows at Warlock levels 3, 5, 7, and 9. Dark One’s Own
Luck retains ability check or saving throw timing, adding 1d10 after seeing the
roll but before any effects, Charisma modifier uses with a one-use minimum,
once per roll, and full recovery on finishing a Long Rest.

Fiendish Resilience retains a choice of one damage type other than Force on
finishing a Short or Long Rest, self Resistance, and termination when a different
type is chosen with this feature. Its twelve permitted types are choice metadata.
Hurl Through Hell retains once-per-turn optional attack-roll-hit timing, the
creature target’s Charisma save against the Warlock’s spell save DC, disappearance
on failure, 8d10 Psychic damage only for a non-Fiend, Incapacitated until the end
of the Warlock’s next turn, and return then to the prior or nearest unoccupied
space. It retains one use, Long Rest recovery, and restoration by expending one
Pact Magic slot without an action.

The external composition manifest loads canonical Warlock and Fiend once each.
Predicates reference the actual Warlock class, Pact Magic feature, prepared-spell
choice, and Pact slot resource. Canonical Warlock supplies its twenty progression
rows, Charisma spellcasting, preparation/count rules, and Pact slot capacity,
level, and recovery. Those definitions remain in their original owner; the batch
imports no Invocation, spell, item, or adjacent subclass definitions.

The guarded character contract is unchanged. Zero HP is explicit in the trigger
keys; it does not require relaxing the positive-only numeric parameter guard.
All feature levels use predicates with canonical class operands. No runtime,
grants, creation, advancement, coverage, or Full Rules support is added.
Readiness remains unavailable.

The readable Go graph oracle and standalone Python comparator independently
transcribe the complete metadata. The comparator reads the SHA-256-pinned PDF,
checks the source clauses and full preparation table against the actual records,
and runs the canonical Warlock source comparator. Both proofs reject the same
142 structurally valid contrary graphs: wrong existing operands, missing options,
conditions, rows or spells, changed formulas, costs, recovery, levels, numbers,
range, duration or termination, forbidden Force, and invented scaling. Each
control must pass structural validation and fail readiness before comparison.
Strict decoding, source/schema errors, nested unknown keys, unknown mechanics,
duplicate canonical/batch composition, fixed Fighter/Cleric behavior and unavailable
Warlock creation are also checked.

Source data retains CC-BY-4.0 attribution and an adaptation notice. See
[srd-5.2.1-conformance.md](srd-5.2.1-conformance.md).

```sh
rtk proxy python3 scripts/verify_srd_fiend.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
```

Independent parent review and hosted checks are separate gates. The known Go
1.27.3 golangci-lint shim lacks its binary; host tools are unchanged.
