# Potion of Healing self-use

Issue #79 supports one consumable through the shared SRD action preview and
confirmation path: a conscious, active level-one Fighter can drink an owned
Potion of Healing outside active legacy combat. Items require
`properties.srd521_id = "potion-of-healing"`, type `consumable` and positive
quantity. Legacy free-form items are not converted by name.

The character sheet offers the item with its quantity. Preview shows self as the
only applicable target, the known 2d4 + 2 healing effect, current/maximum HP, one
item cost and one Bonus Action cost. Confirmation rolls two d4, caps healing at
maximum HP, consumes one item and records the receipt/history in one transaction.
A full-HP character can intentionally use a potion; it still consumes one item.

The revision includes actor, campaign, item properties, timestamp and quantity.
A stale preview, empty inventory, foreign item, other target or unsupported
build cannot consume or heal. The selected item and actor are locked during
confirmation. Reusing the same action ID reads the committed receipt before
revalidating inventory, so a retry after the last item is consumed cannot consume
again. The sheet and inventory refresh after confirmation; transcript receipts
retain individual dice, rolled/applied healing, HP and quantity before/after.

This slice does not administer potions to other creatures, revive unconscious
characters, implement improved potions or arbitrary effects, or track Bonus
Action availability within initiative. Active legacy combat remains rejected.
The wider magic-item catalogue stays explicitly unsupported.

The pinned [SRD 5.2.1](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf)
specifies this item's Bonus Action and healing (p. 99), with HP-capped healing
on p. 17. These are adapted mechanical fixtures; see
[the source pin and required attribution](srd-5.2.1-conformance.md).

`internal/rules/srd521/consumable_test.go` checks ordinary/capped/full-HP healing
and invalid inputs. `internal/handlers/consumable_action_integration_test.go`
checks stale quantity, self targeting, zero inventory, durable replay and an
injected history failure that must roll back both HP and item consumption.
Frontend unit and desktop/phone browser checks cover preview, explicit
confirmation, refreshed inventory and transcript agreement.
