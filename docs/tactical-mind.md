# Tactical Mind: one persisted failed-check window (#339)

The supported level-two Human Soldier Fighter can explicitly use or decline
Tactical Mind for one server-adjudicated ordinary ability check. Attacks, saves,
advantage/disadvantage checks, unknown outcomes, other characters and later
levels do not qualify. The player can select
**Edda fixture · Easy Strength (Athletics) · DC 10**
from the web action menu or use `/easy-athletics` in the TUI. General scene
adjudication remains unavailable. This fixed, disclosed difficulty is Edda's
bounded adjudication fixture, not an inference about a narrated task.

The pinned English SRD 5.2.1 PDF, printed p.48, says to add one d10 to the
failed ability check instead of healing. Second Wind is expended only if the
new total succeeds. This feature does not specify an Action, Bonus Action or
Reaction cost. It does not repeat the base check, heal, repeat narration or
advance the scene. The original receipt and history remain intact; a new
Tactical Mind receipt records the d10, resulting total/outcome and exact
Second Wind before/after values. Decline closes the window without a die or
resource charge. A still-failed attempt also closes it; this bounded slice
never chains another attempt on the same check.

Source review included D20 Tests and ability checks (p.6), Actions and Bonus
Actions (pp.9–10), Second Wind and Tactical Mind (p.48), and the glossary's
Long Rest and Short Rest rules. Second Wind recovers one expended use on a
Short Rest and all uses on a Long Rest; recovery remains owned by the existing
Fighter rest implementation. Tactical Mind does not award rest benefits or
spend an encounter action. Source artifact:
`/tmp/edda-wave1/SRD_CC_v5.2.1.pdf`, SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
These mechanics are Edda adaptations of the CC-BY-4.0 source; license,
attribution, source URL and pin remain in [the conformance document](srd-5.2.1-conformance.md).
Coverage remains partial.

## Trusted window and persistence

Migration 00043 adds only `tactical_mind_windows`, with cascading actor/campaign
references, original check ID and public receipt, private difficulty, owner,
actor revision, last history turn, original-history fingerprint and consumed
flag. It has no player/model write endpoint. No production migration was run.

The distinct `easy-strength-athletics-dc10` entry uses the existing ability-check
prepare/explicit-confirm boundary, limited by the recorded Human/Soldier
Fighter 2 fixture guard and actual ownership. It accepts no ability, skill,
slot, DC or outcome overrides. Preparation discloses fixed DC 10 and rolls
nothing. Confirmation derives Strength (Athletics), rolls one ordinary d20,
and records its public check receipt, known success/failure and disclosed DC.
It makes no world-fiction claim, spends no Second Wind and rolls no d10.
The SRD p.6 Typical Difficulty Classes table defines Easy as DC 10; selecting
this fixture is the explicit Edda adjudication choice.

On failure, `engine.ConfirmAction` calls `game.OpenTacticalMindWindow` after
all final history/campaign writes, with a fresh locked current snapshot in
the same transaction. The original action ID identifies the saved base check.
A successful check opens no eligible window. History, receipt and window
creation roll back together on either history or window-write failure.
The ordinary sheet/free-text checks still return unknown outcomes and do
not open a window. There is no general player/model DC or GM endpoint.

The window is bound to the character owner as well as the campaign owner.
Fresh authenticated preparation rejects a character reassigned to another user;
transferring both character and campaign cannot transfer the old owner's window.
The executable eligibility check requires the recorded Human origin and Soldier
background during both window creation and consumption. Revision and history
checks reject changed actor state, later actions, and edited or replaced original
history records even if their turn number stays the same. Confirmation locks the
campaign, actor and window, retains shared locks on the original history rows,
and reloads those facts. The existing transaction atomically closes the window,
updates only the Second Wind JSON leaf, and writes history and a durable action receipt.
Unrelated stats and legacy reads remain intact. Same-ID retry/reload returns
the original durable receipt without another die; a different ID cannot reuse
the consumed window. An intervening rest or other action invalidates the window.

## Adapters and verification limits

The web action menu has the selectable base fixture and use/decline selections
for Fighter 2. All prepare through `game.PrepareStoredAction` and require the existing explicit Confirm
button. The typed public receipt has no private difficulty field. A known DC
in the original check remains known; hidden DCs remain omitted.

The shared Go preparation/confirmation boundary is available to Go consumers.
The TUI follows the merged Action Surge command pattern: `/easy-athletics`
prepares the fixed DC 10 base check and `/confirm-easy-athletics` explicitly
confirms it. After failure, `/tactical-mind`
previews use, `/decline-tactical-mind` previews decline, and
`/confirm-tactical-mind` confirms the pending choice. `/cancel-tactical-mind`
dismisses local confirmation without consuming the server window. Routing
tests exercise preview, explicit confirmation, same-ID lost-response retry and
decline selection without entering the inference path. Pending TUI identity is
in memory; process-restart recovery is not claimed.

Rules tests cover d10 1/10, equality success, still-failed cost, decline,
invalid dice, unavailable resources, attacks/saves and nonordinary checks.
Isolated PostgreSQL integration exercises injectable dice, history-failure
rollback, reload, durable replay, concurrent confirmations, stale actor/history,
empty resources, and fresh authenticated ownership negatives. These use a real
disposable PostgreSQL container, not a database double. The authenticated HTTP
product-path suite uses no directly opened/seeded window:
it previews and confirms the base check, then prepares and confirms Tactical
Mind use/decline, with exact d20/d10/cost, base success/no-window, replay/reload,
forged IDs/fields, unsupported origins/backgrounds, same-turn history changes,
and history/window failure rollback. Browser fixtures use API doubles and drive
the full selectable base-check-to-Tactical flow, including use/decline, base
success/no-window, explicit confirmation and lost-response/reload recovery
at 1440×844 and 390×844. Phone viewport emulation does not establish physical
phone keyboard behavior or end-to-end browser-to-PostgreSQL integration.
