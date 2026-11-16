# Read-only SRD encounter budgets

Issue #149 adds `srd521.CalculateEncounterBudget(party, monsters)` in
`internal/rules/srd521/encounter_budget.go`. It calculates XP spending budgets
and estimates difficulty for explicitly selected creatures. It has no database,
engine, handler, UI, model, dice, combat-start or reward dependency. Existing
legacy XP formulas and the equipment mutation lane are unchanged.

Pass one `EncounterPartyGroup{Level, Count}` per character level in the party.
Levels 1–20 and positive counts are required. Mixed-level parties sum each
level's budget; this is an explicit adaptation of the per-character table.
Pass `EncounterMonsterGroup{CR, XP, Count}` for each selected creature group.
`XP` is a pointer so missing/null XP cannot become CR 0's explicit zero.
Canonical CR strings are `0`, `1/8`, `1/4`, `1/2`, and integers `1` through `30`.
XP must match the pinned CR table, including CR 0's choice of 0 or 10.
Missing/null level or count becomes zero and is rejected. Negative XP,
nonpositive counts, unsupported levels/CR, and multiplication or accumulation
beyond int64 all return an error and a zero result. Standard Go JSON decoding
also rejects fractional, string and overflowing integer inputs.

`EncounterMonsterFromRecord(monster, count)` copies CR/XP from an existing
validated typed `Monster` and checks the pair against the XP table. Its stable
monster ID is retained in the derivation. This does not call `Available`, grant
playable support, drop unsupported mechanics, or execute the monster. The
Bandit catalogue fixture remains unavailable for combat.

The result contains per-level table rows, multiplied party subtotals, per-group
monster XP subtotals, total XP, three party spending ceilings, source IDs and
pages, the original artifact pin, attribution and adaptation notice. No input
slices or XP pointers are retained in the result.

## Difficulty interpretation

The source specifies spending budgets and says to spend near the budget without
going over. Its examples classify 150 XP as low for four level 1 characters
(low budget 200), and 46,000 XP as high for six level 15 characters (high budget
46,800). The calculator therefore uses the smallest ceiling containing the XP:

- `low`: total XP is at most the low budget, including zero.
- `moderate`: total exceeds low and is at most moderate.
- `high`: total exceeds moderate and is at most high.
- `above-high`: total exceeds high; this is a local out-of-budget label.

These comparisons are an adaptation, not a new SRD difficulty table. Empty
monster selections yield zero XP and `low`; empty parties fail. No monster-count
multiplier from older rules is applied. Terrain, unusual abilities, powerful
creatures, many creatures and CR 0 risks from p. 203 can change actual difficulty;
this numerical estimate does not adjudicate those risks or recommend encounters.
It never starts combat, awards XP, generates encounters or automates loot.

## Source and offline proof

The pinned English [SRD 5.2.1 PDF](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf)
was inspected on one-based PDF pp. 202–203 and 256. Its SHA-256 is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The existing repository pin is documented in `docs/srd-5.2.1-conformance.md`.
Stable local source IDs are `srd-5.2.1:toolbox/encounter-budget` (p. 202)
and `srd-5.2.1:monsters/experience-points-by-challenge-rating` (p. 256).

`testdata/encounter_budgets.json` transcribes every budget row and every CR/XP
pair from those pages. `encounter_budget_test.go` checks all rows, every CR
(including both CR 0 values and all fractional values), the three source
examples, mixed-level parties, every difficulty ceiling exactly and ±5 XP
(the nearest reachable XP increment), empty selections, invalid and absent
numeric data, overflow, derivations, source metadata and typed-record isolation.
These fixtures run offline without the PDF, models or production data.

Run `go test -race ./internal/rules/srd521 -run Encounter -count=1`,
`go test -short -race ./...`, `go vet ./...`, and `git diff --check`.
This slice has no durable effects or UI changes; database integration and browser
checks do not apply. Hosted CI and reviews are separate from these local checks.

This work includes material from the System Reference Document 5.2.1 (“SRD 5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd. The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International License, available at https://creativecommons.org/licenses/by/4.0/legalcode.

Adaptation: the source tables are represented as typed numeric rows and source-
linked derivations. Mixed-level budgets and the smallest-fitting-ceiling labels
are the local interpretations described above. No qualitative troubleshooting
advice is converted into an automatic mechanic.
