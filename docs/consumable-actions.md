# Potion of Healing use and administration

Issues #79 and #266 support one consumable through the shared SRD action preview and
confirmation path: a conscious actor with a supported Fighter or Cleric sheet can drink an owned
Potion of Healing, or administer it to one qualified ally outside encounters. Items require
`properties.srd521_id = "potion-of-healing"`, type `consumable` and positive
quantity. Legacy free-form items are not converted by name.

The character sheet offers the item with its quantity. Preview requires explicit selection of self or one qualified ally and shows
the known 2d4 + 2 healing effect, one item cost and one Bonus Action cost. Confirmation rolls two d4, caps healing at
maximum HP, consumes one item and records the receipt/history in one transaction.
A full-HP character can intentionally use a potion; it still consumes one item.

The revision includes actor, campaign, item properties, timestamp and quantity.
A stale preview, empty inventory, foreign item, unqualified target or unsupported
build cannot consume or heal. The selected item, actor and eligible allies are locked during
confirmation. Reusing the same action ID reads the committed receipt before
revalidating inventory, so a retry after the last item is consumed cannot consume
again. The sheet and inventory refresh after confirmation; transcript receipts
retain individual dice, rolled/applied healing, HP and quantity before/after.

Ally administration requires the separate trusted `srd521_ally` profile
`potion-healing-ally-v1`; the Cure Wounds profile remains spell-specific.
The shared eligibility validator requires exactly these non-null keys:
`profile`, `actor_id`, `location_id`, `allied`, `willing`, `conscious`, `hidden`,
`conditions`, `max_hp`, `distance_feet`, `line_of_sight`, `cover`. The actor and
adjudicated location must match the current scene; the living NPC must have
positive current HP, a valid maximum and known campaign history. Allied,
willing and conscious must be true, hidden false, conditions ordinary, cover
none, sight true and explicit touch distance at most 5 feet. Missing, extra or
case-alias fields fail closed. The revision includes target HP, timestamp and
facts; confirmation queries these facts again under row locks. Moving both
participants to another scene requires a fresh trusted location/touch ruling.
Model NPC creation rejects the entire reserved namespace; update/dialogue tools
cannot author it. Disposition and narrative prose never provide consent.

Unconscious/unwilling targets, thrown effects and stronger variants remain
unsupported. The two-actor encounter offers only self-use; it has no allied
participant support. Active legacy combat remains rejected.
The wider magic-item catalogue stays explicitly unsupported.

The pinned [SRD 5.2.1](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf)
specifies this item's Bonus Action, five-foot administration and healing
(p. 99); p. 236 distinguishes the basic 2d4 + 2 potion from stronger variants.
Bonus Action restrictions are on p. 10 and HP-capped healing on p. 17. These are adapted mechanical fixtures; see
[the source pin and required attribution](srd-5.2.1-conformance.md).

`internal/rules/srd521/consumable_test.go` checks ordinary/capped/full-HP healing
and invalid inputs. `internal/handlers/consumable_action_integration_test.go`
checks stale quantity, self targeting, zero inventory, durable replay and an
injected history failure that must roll back both HP and item consumption.
Frontend unit and desktop/phone browser checks cover preview, explicit
confirmation, refreshed inventory and transcript agreement.

`internal/handlers/ally_potion_integration_test.go` adds strict fact/profile and
scene negatives, current-target revalidation at the locking query, inventory
ownership/identity/quantity changes, HP/quantity/receipt/history/clock rollback,
HTTP projection and slot independence. Channel-gated dice hold a real transaction
while a second engine confirms the same ID or contends with a distinct ID for
the last potion. The contender retries after the first commit; only one heal,
cost, history row and receipt commit. Recreated engines replay and reload the
same receipt without rolling. These tests run on disposable PostgreSQL.

The browser journeys at 1440 and 390 pixels use API doubles to check explicit
ally/self selection, costs, lost responses, retry, reload reconciliation,
inventory refresh and transcript receipts. They prove UI behavior separately
from database durability; 390 pixels is a viewport, not a physical phone.
The existing shared Go/API result remains available to TUI consumers; this change
does not add a terminal target-selection screen.
