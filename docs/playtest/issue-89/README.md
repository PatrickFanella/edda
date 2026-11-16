# Wave 14 declared-slice qualification (#89)

The independently executable qualification work is recorded here. **The full
#89 qualification is incomplete:** SRD free-text play remains unavailable,
and the required phone-keyboard and final human checkpoint has not passed.
Keep #89 open. This candidate does not qualify Full Rules or production.

Application baseline: `c0383a255bc92726aa047b494a5cb9b2efa90229`, with the
bounded client retry fix, qualification tests and evidence in this PR. Observed November 15, 2026 UTC.
Related: [#89](https://github.com/subculture-collective/edda/issues/89),
[roadmap #47](https://github.com/subculture-collective/edda/issues/47),
[coverage epic #54](https://github.com/subculture-collective/edda/issues/54).

## Declared boundary

This run covers the fixed level-one Human Soldier Fighter and Human Acolyte
Cleric, the narrow Fighter 1→2 HP transition, inspectable checks/saves, ordinary
one-handed longsword, conscious-self spell/potion fixtures, explicit encounter
and reaction fixtures, and the recorded rest/concentration paths. Action Surge
and Tactical Mind remain read-only. Broader builds, targets, feature effects and
source families remain unavailable as stated by the server coverage report.

All broad register scopes remain unqualified. The 344 character/advancement,
335 world and 753 spell/item planning identities are separate source inventories,
not evidence of playable effects or a percentage of complete SRD support.
#89 qualifies a partial slice; #54 and its full-domain gates remain independent.

## Acceptance result

| #89 criterion | Result | Evidence and limit |
| --- | --- | --- |
| Create supported martial/caster fixtures and confirm sheet actions | PASS for the named deterministic paths | [Backend/TUI evidence](backend.md) and the new joined projection/action test; [browser evidence](browser.md) separates mocked client contracts from real PostgreSQL persistence |
| Act from free text in a versioned SRD campaign | FAIL | The engine's gather stage refuses SRD campaigns through `RequireLegacyRules`; legacy free-text and typed SRD confirmation passing independently do not satisfy this combined requirement |
| Spend once, resolve combat, recover failures, replay stable results, read/export chapters | PASS for the named fixtures | Existing database-backed action/encounter/reaction/recovery/chapter/export suites rerun against isolated PostgreSQL; this is a matrix of vertical fixtures, not one continuous real-browser/database campaign |
| Viewport fit, scroll ownership, touch dispatch and enlarged text | PASS for automated Chromium cases | Full current browser suite includes #78's six-size matrix, clipping/hit-testing negative controls, touch emulation and root-text reflow |
| Real phone keyboard, browser chrome and final human experience | INCONCLUSIVE | [Fresh device attempt and remaining checks](device-and-human.md); another session owns the only Android target |
| Legacy/TUI remain readable; coverage stays partial | PASS for automated compatibility/TUI fixtures | Legacy campaigns retain their separate path; no conversion or broad coverage promotion is exercised |

## Material finding: SRD free-text boundary (#342)

The web composer sends ordinary input through the WebSocket handler to
`Engine.ProcessTurnStream`. `internal/engine/stage_gather.go` refuses every
versioned campaign before the legacy narrative/tool pipeline runs. The prepared
sheet path uses `ConfirmAction` and succeeds independently. The existing
`TestIntegrationRulesVersionCompatibility` already preserves this intentional
safety boundary; the added qualification test joins it with a real supported
prepared action and reloaded state.

This is a safe refusal, but it prevents #89's explicit free-text acceptance from
passing. A fixture socket that fabricates a successful SRD narration would hide
the gap. Legacy free-text client dispatch is therefore tested separately from
SRD sheet confirmation. The fix must use an authoritative versioned action path;
simply removing the guard would expose legacy mutation tools to SRD data.

[Follow-up #342](https://github.com/subculture-collective/edda/issues/342) is a
bounded text-to-preparation implementation for the existing Athletics check.
It has native prerequisites #64/#88, belongs to #50 and blocks #89. It must
join actual server preparation/confirmation, deterministic sheet/text equivalence
and retry/reload evidence; it cannot close from a mocked success or by enabling
legacy tools. Broader free-text adjudication remains outside that exact slice.

Versioned legacy save/reset writes are also guarded. The legacy atomic Start
Over tests establish their named legacy behavior, not SRD reset support. The
report does not use those tests to claim a complete versioned campaign lifecycle.

## Bounded recovery fix

The qualification exposed a client recovery gap after a server rejection:
the composer had cleared its submitted value, but no explicit control exposed
the transport's retained action identity. The candidate adds **Retry previous
action**. It sends the stored input with the same action ID only after a click,
waits for reconnection, clears pending state on campaign changes and hides the
control after completion. It does not overwrite a newer draft or replay on its
own. A retry-specific narrative transition reuses the original optimistic player
entry so a later history response cannot add a duplicate line for that turn.

Hook, reducer and browser cases cover the retained identity, campaign round trip,
error/reconnect/retry and delayed history boundary. This is client recovery
support; it does not make an unsupported SRD free-text request executable.

## Reproduction and qualification limits

The commands and outcomes for the real database/TUI and isolated frontend
runs are in [backend.md](backend.md) and [browser.md](browser.md). The deterministic
checks use injectable dice and local doubles; they require no production
account, paid inference or production data. Testcontainers owns disposable
PostgreSQL. Playwright owns its temporary frontend server and fresh contexts.

A passing test run records that the declared current boundary behaves as
expected, including explicit refusals. It does not turn the failed product
acceptance above into a pass. Hosted results belong to this PR's exact head and
are separate from these local results. No deployment was performed.

The human checkpoint remains in [device-and-human.md](device-and-human.md).
Once functional blockers and those human checks are satisfied, #89 may qualify
the declared partial slice without waiting for every unrelated SRD expansion.
Full Rules remains gated on the complete register and all required domain
qualification evidence.
