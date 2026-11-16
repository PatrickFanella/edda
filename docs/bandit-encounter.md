# Explicit Bandit encounter fixture (#148)

This fixture executes one selected Bandit weapon action in the existing two-actor
encounter. Its source is the Bandit block on PDF page 261 of the
[pinned SRD 5.2.1](srd-5.2.1-conformance.md), SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The existing typed transcription retains the source attribution and CC-BY-4.0
notice. This runtime adapts those mechanical fields into attack resolution.

## Explicit instance contract

An NPC participates only through `properties.srd521_target`, with profile
`bandit-v1`, the controlling player's `actor_id`, exact `distance_feet`, explicit
`line_of_sight` and `cover`, and a `bandit` object. The object has `version`, the
content revision returned by `srd521.BanditRecord()`, and `action_id`, either
`scimitar` or `light-crossbow`. The revision hashes the entire validated block,
including required mechanics. A name such as Bandit never selects this profile.
No production data is seeded or migrated.

The existing campaign, player, location, prior encounter, hidden-NPC, ownership,
turn and spatial checks still apply. The instance retains its stored HP; public
previews omit NPC HP and private tactics. AC 12, initiative +1, walk speed 30,
and both named weapons come from the pinned record. The NPC's selected action
is explicit configuration, rather than an AI weapon-selection policy. Preview
lists the named actions with their current geometry eligibility; End turn
previews the configured attack. Changes to its profile require adjudication
rather than silently changing an active encounter's rules.

Scimitar requires visible melee reach within 5 feet. Light Crossbow requires a
visible target beyond 5 feet and within its normal 80-foot range. Half and
three-quarters cover use the existing symmetric AC bonus. Unseen targets, total
cover, unresolved facts, close ranged attacks and long range fail closed; this
fixture does not resolve their advantage/disadvantage mechanics. It creates no
path or movement ruling. The existing explicitly configured withdrawal/reaction
path remains separate from weapon selection.

## Persistence and presentation

Confirmation first checks the durable action receipt. New confirmations
revalidate the snapshot before rolling. Weapon damage, concentration effects,
turn budgets, encounter cursor, clock, receipt and history remain under the
existing transaction. A failed die or history write commits none of them;
retrying a failed attempt can roll again, while replaying a completed identity
returns its saved result without rolling.

Both weapons produce the existing typed AttackReceipt with their name, +3
modifier, exact damage dice and modifier, damage type and Bandit source ID.
Ordinary hits roll 1d6+1 slashing or 1d8+1 piercing; critical hits double only
the dice, and misses deal zero. Shared web resolution events and the TUI's
engine narrative use those receipts. The web receipt displays actual rolled
values without assuming every die is a d8.

## Verification and limits

Deterministic rule and geometry tests cover both weapons, ordinary/critical/miss
outcomes, invalid dice, stale revisions, unsupported actions and unsupported
geometry. PostgreSQL fixtures cover HP, AC, initiative, budgets, failed dice,
history rollback and retry, completed receipt replay/reload, hidden or relocated
NPCs and changed ownership/profile. Existing generic encounter and reaction
fixtures remain regression checks. Desktop and phone Chromium fixtures exercise
explicit confirmation, a lost response, receipt reload and socket reconnect for
both weapons. Browser fixtures use mocked API contracts; durable effects are
proved separately against isolated PostgreSQL.

This is not arbitrary monster execution, Multiattack, adjacent species, general
ranged combat, monster AI or full catalogue qualification. `Monster.Available()`
retains its conservative catalogue boundary; only the explicit Bandit fixture
uses the bounded runtime. The monster family coverage register is unchanged.
