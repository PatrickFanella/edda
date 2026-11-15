# Cleric concentration and recovery fixture

Issue #77 adds Shield of Faith on the conscious level-one Cleric's self target.
The existing sheet prepare/confirm path previews the slot, Bonus Action,
components, +2 AC and ten-minute concentration limit. Preview never rolls or
writes. Confirmation saves the slot cost, effect identity, start/expiry times
and history atomically. Cure Wounds still uses a Magic action and does not end
concentration. The fixed Human Acolyte Cleric from #72 prepares both spells.

`stats.srd521.casting` now also carries `shield_prepared`, `focus`, optional
`concentration` and `last_long_rest`. `focus` is the trusted scene ruling that
the Holy Symbol can replace the spell's noncostly material component; verbal
and somatic availability remain explicit. A stored concentration has the
confirming action ID, spell ID and absolute campaign start/expiry seconds.
Only this single self effect is supported. Recasting replaces its exact ID;
ending it removes the AC term once without spending an action or recovering a
slot. Reload and receipt replay do not restart the timer or spend again.

Migration `00042` adds a seconds column to the existing campaign clock, default
zero for existing saves. Quiet downtime advances one minute; ordinary encounter
rounds advance six seconds, without rounding up a minute per action. Only actual
campaign time advances duration: reading, previewing and retrying are inert.
Expiry clears the saved effect and removes its derived AC bonus in the same
transaction as the time/history change. The character sheet and saved receipts
show campaign day and time derived from these absolute seconds, so the expiry remains inspectable after reload.

The Cleric can enter the same explicitly prepared, player-known ordinary
two-actor encounter as the Fighter without inventing a weapon attack. Shield of
Faith consumes its Bonus Action; Cure Wounds consumes its action. A durable
turn-budget flag enforces one slotted spell per turn even when both action types
remain available. Resolved NPC
damage triggers a Constitution save at `min(30, max(10, floor(damage/2)))`, saved
with its d20, modifier, public DC and outcome. Failure removes concentration;
zero HP ends it without a roll. No damage requires no check. The damage, check,
effect change, turn budget, clock and history share the transaction. Replaying
the action reads that result without rolling a second save.

Outside active combat, the player can confirm these bounded downtime fixtures:

- One quiet minute: advances time and expires effects, with no recovery.
- One uninterrupted hour of light activity: a Short Rest, optionally spending
  the one d8 Hit Die. Healing is d8 + Constitution, at least zero and capped at
  maximum HP. It recovers no Cleric spell slots and cannot spend a missing die.
- Eight uninterrupted hours including six hours of sleep: a Long Rest. Sleep
  ends concentration; completion restores HP, the d8 Hit Die, both level-one
  Cleric slots and the recorded Magic Initiate free use. Starting another Long
  Rest requires at least sixteen hours since the previous rest finished.

These are explicit confirmations that the stated uninterrupted downtime was
available. Random encounters, partial/interrupted rest resumption, Exhaustion,
maximum-HP reductions, other class recovery, arbitrary condition transitions,
other concentration spells and feat spell execution remain outside this fixture.
Legacy model tools cannot bypass the versioned action path. Recorded spell
choices still distinguish playable effects from unavailable effects.

No production conversion or mutation is needed. Apply the migration with the
normal release process before running the new binary. Its down migration drops
only the subminute clock column; retain a database backup before a deployment
rollback because old code cannot adjudicate saved Shield of Faith effects.

## Source and verification

This is an adaptation of the pinned
[SRD 5.2.1](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf):
rounds p. 13, Cleric recovery p. 37, components and one-slot-per-turn p. 105, Shield of Faith p. 162,
Concentration p. 179, Long Rest p. 185 and Short Rest p. 187. See
[the pin, license and attribution](srd-5.2.1-conformance.md).

Rules tests cover replacement, ending, exact expiry, recovery boundaries,
Constitution-save success/failure, DC rounding/cap and unconsciousness. Isolated
PostgreSQL tests cover durable replay, reload, stale previews, a history-write
failure with complete rollback, unchanged unrelated stats, rest limits, a saved
NPC damage/save and the Bonus Action budget. Desktop and phone Chromium
fixtures cover explicit target/slot selection, no early confirmation, reload
with the same expiry, rest confirmation and refreshed resources. These checks
do not qualify production deployment, a visible device keyboard or the full SRD.
