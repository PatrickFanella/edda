# SRD sheet checks and saving throws (#61)

`game.ResolveCharacterCheck` is the shared Go boundary for the supported SRD
character-sheet fixture. It checks explicit `srd-5.2.1` campaign metadata,
projects the stored character through `srd521.Project`, and delegates to
`srd521.ResolveCheck`. Dice are injected through `Intn(n int) int`, using
zero-based values like the existing legacy dice interface. A shared dice source
must support concurrent callers. Neither React nor a language model supplies
modifiers or resolves the total and outcome.

The supported projection remains the level-one Fighter described in
[character-projection.md](character-projection.md). The resolver supports its
six ability checks, six saving throws and 18 skills using their listed abilities.
Recorded skill proficiency is added once. Optional `stats.srd521.expertise`
lists distinct proficient skills and doubles their proficiency contribution
instead of adding a second proficiency bonus. Expertise never applies to a
save or an unrelated check. This field records an already adjudicated grant;
it does not validate class/level eligibility or award Expertise to a Fighter.
Synthetic expertise cases exercise the arithmetic only. Character creation
validation remains outside this slice.

An ordinary roll uses one d20. Advantage keeps the higher of two, disadvantage
keeps the lower, and both together cancel to one ordinary die. Equal dice keep
the first; both dice remain in the receipt. A total equal to the DC succeeds.
Natural 1 and 20 have no special success/failure rule for these checks and saves.
Attack rolls, death saves, damage, rerolls, alternative skill/ability pairings,
and automatic adjudication of armor or other situational effects are unsupported.
The caller supplies adjudicated advantage/disadvantage circumstances.

`CheckResolution.Success()` is private server adjudication. `PlayerReceipt()`
returns an independent public copy containing the individual dice, kept flags,
effective roll mode, derived modifier terms, total and stable source IDs.
`Difficulty.PlayerKnown` discloses the DC; `Difficulty.OutcomeKnown` independently
discloses success/failure. Both default to false. An unknown result is `unknown`,
never inferred as failure. No margin, reason, entity, narrative or world fact is
accepted in this receipt. JSON marshaling a resolution also emits only that
player-safe projection. Disclosure flags are trusted server context and must
not be accepted from a player or model request.

`tools.SRDCheckResult` wraps only the public receipt for the engine journal.
The `srd_check` conversion in `engineTurnResultToAPI` and history readback
whitelists that typed receipt and ignores raw arguments, sibling DC/margin
fields, and unknown receipt fields. It does not use the legacy `skill_check`
argument/result merge. Existing legacy arithmetic, registry and campaign guards
are unchanged. This slice does not register a new live tool or enable the SRD
campaign turn loop; preview and confirmation remain #63/#64. `pkg/api.CheckReceipt`
aliases the shared rules DTO so future web/TUI adapters need no arithmetic copy.

The Go-generated fixture at
`frontend/tests/fixtures/srd-check-sheet.json` pairs the sheet with four resolved
tests: an ordinary Athletics check, an advantaged Athletics check with a secret
DC, a disadvantaged Constitution save and a Perception check with unrevealed
outcome. `TestCheckSheetFixture` generates them through the game and tool
boundaries and compares the exact JSON. The frontend behavior test passes those
receipts to `RulesSheet`, opens the details and inspects dice, modifiers, sources
and disclosure text. `RulesSheet` displays supplied receipts without computing
them or executing an action. The existing campaign character endpoint remains
read only.

After reviewing an intentional rules change, regenerate the fixture with:

```sh
rtk proxy env EDDA_UPDATE_CHECK_FIXTURE=1 go test -short -run '^TestCheckSheetFixture$' ./internal/game
```

Source IDs extend the map in the projection document. They refer to the
[pinned English SRD 5.2.1 PDF](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf):

| ID suffix after `srd-5.2.1:content/` | PDF pages |
| --- | --- |
| `d20-tests/ability-checks` | 6 |
| `d20-tests/saving-throws` | 7 |
| `d20-tests/advantage-disadvantage` | 7–8 |
| `character/expertise` | 8, 182 |

The artifact hash and required attribution remain in
[srd-5.2.1-conformance.md](srd-5.2.1-conformance.md). These calculations and
receipt fixtures are Edda adaptations. Partial fixture coverage does not qualify
the entire D20 Tests family, a class, or a character level as implemented.
