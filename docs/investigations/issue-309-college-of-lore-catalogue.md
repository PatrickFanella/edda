# College of Lore source catalogue (#309)

The unavailable catalogue in `internal/rules/srd521/testdata/character-college-of-lore.json`
adapts SRD 5.2.1 PDF page 35 into the existing character contract. Its five source
identities are College of Lore, Bonus Proficiencies (Bard 3), Cutting Words
(Bard 3), Magical Discoveries (Bard 6) and Peerless Skill (Bard 14). The other ten
records are their nested choices and prerequisites. Each nested record retains
its containing heading's exact inventory ID and page. The source envelope retains
the pinned artifact, CC-BY-4.0 attribution and an explicit adaptation notice.

`character-college-of-lore-composition.json` composes the batch with canonical
`character-bard.json`. It copies no Bard definition. Level predicates use the
canonical Bard class, and Inspiration predicates use the canonical
`resource/bard/bardic-inspiration` operand. Resource references in the contract
are ownership links, so external resource dependencies belong in predicate
operands. Bard retains ownership of its capacity formula, die progression, rest
recovery, Font of Inspiration conversion and Superior Inspiration recovery.
The standalone comparator also runs the canonical Bard source comparison.

The p35 metadata preserves these edition-specific conditions:

- Bonus Proficiencies selects three skills from all eighteen skill keys.
- Cutting Words requires a visible creature within 60 feet, a damage roll or a
  successful ability check or attack roll, a Reaction and one Inspiration use.
  It rolls the canonical die and subtracts from the creature's roll, reducing
  damage or potentially turning success into failure. No hearing, charm immunity,
  saving throw or older-edition restriction is added.
- Magical Discoveries selects two spells from Cleric, Druid or Wizard lists in
  any combination. Each is a cantrip or a spell for which the Bard Features table
  gives slots. Both remain prepared; on gaining a Bard level one can be replaced
  under the same list and level conditions. List keys are opaque domains, not
  imported spell definitions. There is no fixed spell grant, new preparation
  count, Bard-list option or invented level scaling.
- Peerless Skill follows an own failed ability check or attack roll. It spends
  one Inspiration use, rolls the canonical die and adds it to the d20, potentially
  succeeding. If the result still fails, Inspiration is not expended. No Reaction
  cost, saving throw option or older-edition ability-check-only limit is added.

## Criterion proof

| Criterion | Evidence |
| --- | --- |
| Five named identities, exact pages and owners | Complete readable `loreSourceRecords` expectations; PDF comparator independently cross-checks the five inventory identities |
| Levels, numbers, mechanics, source conditions and operands | Whole-record semantic comparison in Go and explicit standalone `SOURCE_FACTS`, including absent fields; every mechanic, numeric value and operand has a contrary control |
| Choices and no invented scaling | Exact counts and option domains; structurally valid extra/missing option, changed count and scaling controls |
| Canonical composition and resource recovery relationship | Composition validates only with canonical Bard, retains Bard records unchanged and rejects duplicates; canonical Bard semantic checks and PDF comparator run too |
| Unknown fields/keys remain unavailable | Strict decoder and readiness tests; no production schema or runtime code changes |
| Fixed runtime, creation and coverage boundaries | Bard/Lore creation remains rejected; qualified fixed Cure Wounds regression remains unchanged; coverage and Full Rules eligibility checks |

The semantic controls first pass the generic structural validator. This matters:
a reference to an existing but wrong spellcasting feature, a coherent source-owner
substitution, or a changed positive range/count is valid metadata syntax but
contradicts p35. Source proof compares readable fields, not a catalogue digest.
The supplied PDF is SHA-256 checked before extraction; prose anchors establish
the source passage alongside the explicit metadata comparison.

## Local verification

```sh
rtk proxy python scripts/verify_srd_college_of_lore.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./internal/rules/srd521 -run CollegeOfLore -count=1
rtk proxy env GOMAXPROCS=2 EDDA_SRD_PDF=/tmp/edda-wave1/SRD_CC_v5.2.1.pdf go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
rtk git diff --check
```

The focused tests, standalone comparator, full short race suite and vet passed,
including 76 structurally valid contrary controls. Go formatting and diff
whitespace checks passed before commit. `golangci-lint run` could not analyze:
the installed mise shim does not provide a binary for active Go 1.27.3. No host
or toolchain changes were made.

No runtime grants, resource creation/spending/recovery, automatic effects,
creation/advancement enablement or coverage/Full Rules promotion is included.
No inference, database mutation, deployment or production check is required for
this data-only change. Hosted qualification and independent review belong to the
parent; local results do not establish those gates.
