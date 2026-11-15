# Narrative combat turns: supported fixture

Issue #66 adds a two-actor ordinary melee encounter to the shared prepared-action
engine. It is explicitly partial SRD 5.2.1 support. Legacy narrative combat keeps
its existing path.

From Character, select **Start ordinary melee encounter** with an equipped,
identified longsword, preview, select a known fixture target, and confirm. The
existing target contract still requires a previously encountered, colocated NPC
with the exact `unarmored-fixture-v1` property assigned to this actor at five feet.
Arbitrary NPCs are unavailable. This change does not seed an encounter into a
production world or expose hidden actors.

The fixture declares the NPC's ordinary Unarmed Strike (+2, 1 bludgeoning on a
hit), AC 10, no conditions, and a 30-foot movement allowance for each actor.
Initiative rolls once: player d20 + projected Dexterity modifier, NPC d20 + 0.
Equal totals put the player first, the fixture's explicit tie choice. A leading
NPC acts immediately; ending the player's turn resolves the next NPC turn and
returns to the player. NPC attacks use the same ordinary attack-roll resolver as
the supported longsword, including natural 1 and 20. Fixed unarmed damage has no
critical damage dice to double.

The saved encounter contains initiative dice/order, round, active actor, HP and
remaining action, bonus action and movement. Longsword Attack consumes one action;
Potion of Healing self-use consumes one bonus action and one owned potion. Dash
consumes an action and adds 30 feet. Spend 5 feet consumes movement only; it does
not claim to change a scene position or target range. Ending a turn resets the
next actor's budget; wrapping initiative increments the round. At zero HP the
fixture completes, without claiming death-save or encounter reward support.

Every confirmation reloads the actor, inventory, legal targets and latest saved
encounter under the campaign turn exclusion. Its fingerprint invalidates stale
previews. Mechanics, NPC dice receipts, HP, turn budget and history commit in one
transaction. Retrying the same action ID returns the saved result without rolling
or charging again. Reload reads the latest encounter from full campaign history,
without the recent-context cutoff. Existing save/reset operations include that
history; no migration is needed.

Play uses the saved active encounter to show the combat surface after a result
or reload. Initiative and budget stay visible above the internally scrolling
story/action panel. Choose combat action opens the preview/target/confirm controls
and saved dice results; Read combat story returns to the transcript. Character
exposes the same preparation path. A completed encounter returns to normal play.
Unsupported checks during an active fixture require a different supported action.
SRD campaigns retain the existing guard against legacy model-driven mechanics.
Other reactions, mastery effects, general monster statistics, broader spatial adjudication,
conditions, spellcasting and the wider catalogue remain later roadmap work.

Verification includes deterministic economy/initiative tests, an isolated
PostgreSQL journey for initiative, attacks, potion cost, movement exhaustion,
round advancement, reload and duplicate confirmation, frontend interaction tests,
and isolated Chromium layout/reload checks at 1366×768, 360×800 and 390×844.
These fixtures do not qualify a live provider or a physical phone keyboard.


## Exact spatial rulings (#67)

The `ordinary-spatial-fixture-v1` target profile extends the same two-actor
fixture. Its server-owned `srd521_target` properties declare `actor_id`, exact
`distance_feet`, `line_of_sight` and `cover` (`none`, `half`, `three-quarters`, or
`total`). Missing or invalid facts require clarification in those scene properties
before a roll. The original `unarmored-fixture-v1` declaration at exactly five
feet retains its ordinary clear-sight/no-cover contract. Other distances on that
old profile receive no compatibility defaults.

Previews show player-known targets with distance, five-foot reach and a reason
when an attack is unavailable. Confirmation rechecks current campaign membership,
location, encounter membership, visibility and spatial rulings. Unencountered
NPCs, `hidden: true` targets and private properties are excluded, including from
encounter summaries and accessibility text. Hidden or relocated encounter members
pause the fixture until the scene membership is clarified.

Starting an encounter records the declared spatial ruling with its source hash.
Approach and Retreat each spend five movement feet and change the saved distance
by five feet on the declared unobstructed, visible, no-cover path. The old Spend
5 feet action only spends budget. Distance survives turns and reloads. A changed
server ruling replaces the prior spatial snapshot on the next committed action
and invalidates old previews. No model or client-supplied modifier decides reach.

Ordinary attacks require sight, no total cover and at most five feet of reach.
Half and three-quarters cover add +2 and +5 AC respectively, symmetrically for
this fixture. NPCs hold position and end their turn when a resolved distance,
sight or cover ruling blocks their attack. They do not teleport or invent a path.
Unknown spatial facts pause NPC resolution without rolling. General unseen
attacks, obstacles, asymmetric cover and arbitrary movement remain unsupported.


## Explicit opportunity attacks (#81)

A spatial fixture target can declare the server-owned `tactic: "withdraw"`.
After its ordinary attack, that NPC attempts one voluntary five-foot retreat on
the fixture's clear path. When that step would leave five-foot reach, the player
is conscious, has a reaction and still has an equipped supported longsword, the
engine saves a pending opportunity-attack trigger **before movement**. The NPC's
already-resolved attack is committed with that trigger. No player roll or reaction
cost happens automatically. Other tactics retain the original stationary behavior.

Play offers **Preview opportunity attack** and **Preview decline** in the current
combat frame. Both open the existing read-only preparation and explicit confirmation
flow. The trigger chooses the known target and supported weapon; there is no
additional dice prompt. The character panel also offers the pending responses after
a preview reads the trigger. Preview and confirmation reject other actions until
the choice is answered. Current NPC HP and unplayed tactics stay out of previews.

Confirming uses the existing longsword resolver and records exactly one reaction
cost, damage and dice receipt. Declining rolls nothing and spends no reaction.
Both resume the suspended movement once and return to the player, whose reaction
recovers at the start of that turn. If a new ruling blocks the retreat path, decline
cancels that step and ends the NPC turn instead of inventing movement. An attack
that defeats the NPC completes the encounter without taking the suspended step.

Trigger identity, eligibility, equipment and the spatial fingerprint are checked
again on confirmation. Stale responses require a fresh preview; a changed weapon
can still be declined. Same-ID confirmation retries and fresh-engine reconnects
return the saved result without repeated attacks, resource use or movement. Legacy
model tools remain blocked from mutating versioned campaigns by the existing
production-tool rules guard; the model cannot consent to a player reaction.

This is one ordinary opportunity-attack fixture, not general NPC movement policy,
Disengage, forced movement, reach weapons, condition handling or other reaction
features. Tests cover saved triggers, acceptance/decline, stale rulings, changed
gear, duplicate/reconnect responses, resource recovery and desktop/phone controls.
