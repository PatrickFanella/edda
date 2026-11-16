# Level-two Fighter Action Surge

Issue #338 enables Action Surge only for the existing created level-two Human
Soldier Fighter, during its active ordinary two-actor encounter turn. The
server validates the stored creation provenance, supported projection, living
active actor, campaign, encounter membership, turn, pending reaction and use.
Other levels and builds remain unavailable. This is partial SRD support.

The prepared-action API provides read-only preview, then explicit confirmation.
The web character action selector offers Action Surge and Fighter recovery;
encounter summaries distinguish the ordinary Action and the extra Action.
The TUI narrative input accepts `/action-surge` for preview,
`/confirm-action-surge` for explicit confirmation or retry of the same ID,
and `/cancel-action-surge` to dismiss the local pending request. These commands
call the shared rules engine without a model. The TUI character sheet remains
an existing placeholder; this change does not claim a general TUI action menu.
TUI pending confirmation identity lasts for that app session. Web pending
identity is stored in sessionStorage and reconciled with the durable receipt
after reload; the database receipt itself survives engine reconstruction.

Activation consumes one use and sets a separate `surge_action` grant and
`surge_used` turn flag. It never replaces the ordinary Action or resets Bonus
Action, Reaction, movement or the spell-slot-per-turn flag. The ordinary Action
is consumed first. The extra grant can fund only the existing longsword Attack
or Dash; Magic cannot consume it. Unsupported intents fail before spending.
There is no broader action catalogue. End turn explicitly relinquishes unused
resources; the existing engine runs the NPC only after player turn advancement.
Turn advancement resets both Surge fields. They default false in historical
serialized budgets. Historical level-two builds without `action_surge_spent`
start with the one use previously displayed by the sheet.

Short Rest requires one confirmed uninterrupted hour of light activity and
restores Action Surge and one Second Wind use. Long Rest requires eight hours,
including six hours sleep, restores Action Surge, existing HP, both Hit Dice and
Second Wind uses, and requires sixteen hours after the prior Long Rest before
another starts. Rest is unavailable in an active encounter. This slice does
not simulate interruptions or add Hit Die spending for Fighters.

Confirmation uses the existing receipt-first transaction and campaign exclusion.
It locks campaign/actor and relevant scene membership, fingerprints campaign
and actor ownership again, re-prepares the action,
and checks the revision before mutations. Actor resource, encounter budget and
cursor, clock, effects, typed receipt and history commit together. A failed
history insert rolls back the resource and grant. The same action ID and input
replays the committed result before checking its now-stale preview, without dice
or a second mutation. Public receipt conversion selects typed fields and excludes
private NPC state.

## Source and adaptation

Read directly from `/tmp/edda-wave1/SRD_CC_v5.2.1.pdf`: PDF p.48 (Action Surge
and Second Wind), pp.9–10 (actions, Bonus Actions and Reactions), p.185 (Long
Rest and Magic) and p.187 (Short Rest). SHA-256:
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The repository's original pin and license are in `srd-5.2.1-conformance.md`.
Stable source ID: `srd-5.2.1:content/fighter/action-surge`.

Adaptation: the source feature is represented by a spent flag, a persisted
restricted grant, explicit preview/confirmation and fixed-fixture action choices.
Rest duration and recovery apply only to declared uninterrupted downtime. No
new level, class, subclass, general action catalogue or Full Rules claim is added.

This work includes material from the System Reference Document 5.2.1 (“SRD 5.2.1”)
by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd. The SRD
5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.

## Verification

- Rules/turn-budget tests cover activation before and after the ordinary Action,
  exactly two Actions, duplicate use, Magic/unsupported rejection without budget
  changes, unrelated resource retention, legacy JSON and reset on advancement.
- Isolated PostgreSQL tests cover preview purity, ownership and stale actor,
  resource, encounter and turn guards, history rollback, fresh-engine receipt
  replay/reload, two legal longsword attacks, no third Action, no automatic NPC
  advancement while the grant remains, and rest recovery outside combat.
- TUI tests exercise actual input routing, preview, explicit confirmation,
  uncertain-delivery retry identity and cancellation without narrative inference.
- Browser fixtures exercise the real application with API doubles at 1440 and
  390 CSS pixels: preview, confirmation, injected lost response, reload receipt
  reconciliation, exhausted availability, socket reconnect and overflow checks.
  These prove UI behavior, not database durability or physical-phone behavior.

No production mutation, deployment or paid inference was performed. Hosted CI
and independent review belong to the parent; this is a draft PR.

Local qualification on Kvant (2026-11-16 UTC): `go test -short -race ./...`
passed 3,160 tests across 41 packages; `go vet ./...` passed;
`go test -tags integration -race -run Integration -count=1 -timeout 10m ./...`
passed 244 tests across 41 packages using disposable Testcontainers PostgreSQL.
Frontend production build, all 66 Vitest behavior tests and changed-file ESLint
passed. Ten Playwright journeys covering Surge, encounter, advancement and
caster recovery passed with two workers. The shared preview browser separately
confirmed the production ActionPreviewPanel component with API doubles at 390
and 1440 CSS pixels, including lost response and receipt recovery. No physical
phone was used. The installed golangci-lint shim could not execute under active
Go 1.27.3 (it is installed for Go 1.25.13/1.27.0); no host tooling was changed.
The frontend build retained its existing large-chunk warning.
