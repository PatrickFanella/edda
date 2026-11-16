# Wave 14 backend qualification (#89)

Observed 2026-11-15 against the Wave 13 merged baseline. This is isolated,
PostgreSQL-backed evidence for the declared SRD 5.2.1 partial slice. It is not
evidence for Full Rules, production, a paid model provider, a physical device,
or a human playtest.

## Result

**Typed mechanics: PASS. Free-text campaign action: FAIL for #89.**

`TestIntegrationDeclaredSliceQualification` creates one owned campaign per
supported build. The Fighter performs a deterministic typed Athletics preview
and confirmation; the confirmation spends its single deterministic roll once,
persists a completed receipt, and replays identically through a fresh engine.
Its stored sheet is projected again after that action. The Cleric's supported
sheet is likewise reloaded from its own campaign.

| Build | Reloaded supported scope | Result |
| --- | --- | --- |
| Human Soldier Fighter | `fighter-1-chain-mail-defense-read-only` | PASS |
| Human Acolyte Cleric | `cleric-1-cure-wounds-self-fixture` | PASS |

Both projections continue to report partial coverage, not Full Rules: 344
character/advancement, 335 world, and 753 spell/item scopes remain planned.

## Free-text blocker

In that same persisted Fighter campaign, the real engine receives ordinary
player input: `Inspect the gate`. Both `ProcessTurn` and the streaming path
used by WebSocket/TUI `ProcessTurnStream` return
`domain.ErrLegacyRulesMutation`. The guard is in
`internal/engine/stage_gather.go`, which calls
`domain.RequireLegacyRules(state.Campaign.RulesVersion)` before model work.

The test proves this failure is side-effect free after the valid typed action:
the injected provider has zero calls, injected dice has zero rolls, the stored
character JSON is unchanged, and the combined `session_logs` plus
`turn_attempts` count is unchanged. This is a passing fail-closed guard test,
but it means #89's required ordinary free-text action is not available for an
SRD campaign. The bounded follow-up is
[#342](https://github.com/subculture-collective/edda/issues/342). No gameplay
guard was removed and no free-text parser was added.

`TestIntegrationRulesVersionCompatibility` separately verifies the same SRD
guard through campaign HTTP mutation routes and engine stream entry points, and
also verifies that a legacy campaign remains usable. It is compatibility
evidence, not evidence that SRD free text works.

## Focused evidence

The following commands passed locally against Testcontainers' disposable
pgvector PostgreSQL database:

```sh
rtk proxy go test -tags integration -run TestIntegrationDeclaredSliceQualification -count=1 -p 2 -timeout 5m ./internal/handlers
rtk proxy go test -tags integration -run TestIntegrationRulesVersionCompatibility -count=1 -p 2 -timeout 5m ./internal/handlers
rtk proxy go test ./tui/...
```

Existing focused database suites cover encounter resolution, spell and
consumable costs, action receipts, committed-turn recovery, chapters and
exports. They are useful component evidence, but these commands do not prove a
continuous single-campaign end-to-end playthrough.

## Limits and remaining work

Keep #89 open for its missing SRD free-text action and its separate required
human/visible-phone checks. SRD Start Over and save writes are also unavailable
because the compatibility guard rejects those mutation routes; the legacy
start-over test cannot establish SRD reset support.

Full Rules remains unavailable until its separate expansion gates finish. That
is distinct from #89's declared partial-slice qualification.
