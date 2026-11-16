# Shield of Faith recipient fixture (#351)

This adaptation implements two explicitly prepared level-one Clerics casting on
one known NPC in the same campaign and scene. It extends the existing conscious
self fixture, rather than granting general Cleric or spell support. Production
eligibility and discovery for other recipients remain unavailable pending
[#382](https://github.com/subculture-collective/edda/issues/382). Native readback confirms #382 is a child of #323,
requires #351, and directly blocks #279, #280 and #337. The class
investigation #323 and imported spell records do not grant that eligibility.

The source is the pinned SRD 5.2.1, PDF/printed pages 105, 106, 162 and 179;
SHA-256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The source rules are adapted to one fixed +2 AC spell, one recipient, a level-one
slot, Bonus Action, V/S and an available noncostly Holy Symbol focus. The existing
encounter budget enforces one slot per turn. Ending concentration costs no action
and never refunds a slot. No Bless, arbitrary potency, multiple recipients per
cast, new character creation, item activation or Full Rules qualification is added.

## Entry and authority

The web's existing spell selector calls the shared stored-action preparation
endpoint. It requires explicit recipient selection before confirmation. The
prepared isolated test caster must additionally have
`stats.srd521.casting.isolated_recipient_effects = true`; there is no public API
that grants this test entitlement. Existing self-only casters remain self-only.

An NPC must be alive, previously encountered, and in the caster's current scene.
The model-facing NPC writer reserves `srd521_faith` and its case aliases, alongside
the allied-healing namespace. Its authoritative fixture has exactly seven fields:
`profile` (`shield-of-faith-isolated-v1`), `actor_ids` (explicit permitted casters),
`location_id`, `hidden` (false), `distance_feet` (0 through 60), `line_of_sight`
(true), and `cover` (a recognized value other than total). Missing, null, unknown,
foreign-scene and hidden facts do not produce recipient options. These records
are deliberately installed only in disposable tests; discovery from prose,
catalogue presence or a client-provided ruling is unsupported.

The local TUI accepts `/faith` followed by explicit caster and recipient UUIDs,
or `/end-faith` followed by the caster UUID. Both only prepare. `/confirm-faith`
commits the saved request; `/cancel-faith` dismisses it. An uncertain delivery
retains the same action ID for retry. The TUI checks caster ownership against the
campaign owner and uses the same preparation and confirmation engine as the API.
TUI request retention is in memory; browser request retention also survives page
reload through session storage and receipt reconciliation.

## Persistence and AC

Migration 00044 creates separate `concentration_casts` identities and
`concentration_recipients` links. One caster's replacement/ending deletes only
that caster's superseded links; cast identities remain for history. The existing
caster concentration field remains the lifecycle authority. A nonself cast stores
its recipient ID there so pure sheet projection cannot accidentally grant its
caster self AC. Old self-only saves, including opaque legacy effect IDs, remain
readable. Ordinary equipment and unrelated stats/casting JSON retain their
existing paths.

Casts have absolute start/expiry seconds and exactly 600 seconds maximum. Reads
require `start <= now < expiry` and a live matching caster concentration. For the
fixed equal +2 bonus, the latest confirmed cast wins. A database-generated
`cast_order` records confirmation order under the campaign lock, including casts
in the same campaign second. Failed transactions may leave sequence gaps without
persisting a cast. This implements source precedence without a new SRD
simultaneity rule. Suppressed casts retain their original expiry. Applying the
projection repeatedly still adds only one source-attributed spell term. NPC attack
AC and public character AC use the same live selection; other AC terms remain
separate.

Confirmation locks the campaign first, then actor/effect rows in stable order,
and revalidates actor, recipient, scene, effect and clock fingerprints. Costs,
cast identity, links, lifecycle/turn/clock, history and durable receipt share the
existing transaction and campaign advisory exclusion. Ordinary check revisions
retain their existing contract. Same-ID replay returns before stale-revision
validation and never charges, rolls or writes again.

## Verification and limits

`TestIntegrationFaithRecipients` uses a disposable pgvector PostgreSQL 16
container with real migrations, authorized HTTP preparation, forged cost/timing
rejection, real TUI engine preparation, and fresh engine confirmations. It checks
independent casters/targets, latest/resumption/original expiry, same-caster
replacement, later confirmation with a lower UUID in the same second, exact expiry,
stale actor/recipient/effect snapshots,
range/hidden/cover negatives, no doubled AC, no caster AC for a nonself cast,
explicit entitlement, complete history-failure rollback, saved receipt replay,
concurrent confirmations, unrelated JSON and public disclosure. Dice are injected;
this spell and voluntary ending consume no dice. Existing caster lifecycle and
handler integration suites cover encounter budgets, damage and legacy behavior.

The browser fixtures run at 1440px and 390px with API doubles. They check explicit
recipient confirmation, lost response, receipt reconciliation after reload and
original expiry display. They do not prove a physical phone or live backend
browser session. TUI fixtures drive the actual Bubble Tea update boundary with
an engine double for preview/confirmation/retry/cancel; the PostgreSQL test also
calls the real TUI engine preparation method. No production data, migration
execution, deployment, inference or fleet configuration is involved.

This work includes material from the System Reference Document 5.2.1 (“SRD
5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd.
The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.
Changes: paraphrased rules, bounded fixture contracts and deterministic ordering.
See [the source pin, attribution and exclusions](srd-5.2.1-conformance.md).
