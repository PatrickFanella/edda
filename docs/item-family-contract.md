# Item family contract, schema 1

`internal/rules/srd521/items.go` defines descriptive item families for the pinned
English SRD 5.2.1. `LoadItemCatalogue` loads a fresh validated copy;
`DecodeItemCatalogue` validates supplied JSON without network or database access.
The sole initial fixture is Potion of Healing. This is a contract fixture for
#195, not a catalogue batch or a new gameplay implementation.

The root records `schema_version: 1`, `rules_version: srd-5.2.1` and the same
source pin, CC-BY-4.0 attribution and adaptation notice as the coverage register.
See [source provenance](srd-5.2.1-conformance.md). The fixture adapts names and
numeric rules into structured data; it does not reproduce descriptive prose.

Each family has a canonical ID, name, mundane/magic kind, source presentations
and nested variants. Presentations link the IDs/pages in the
[planning inventory](investigations/issue-80-spell-item-inventory.json).
Potion of Healing's mundane-gear presentation (p. 99) explicitly calls it a magic
item; its magic family presentation is Potions of Healing (p. 236). Both point
to one canonical family, never two inventory entitlements.

Every variant has its own stable ID (`<family-id>/variant/<slug>`), name, rarity,
attunement mode and restrictions, charge capacity/per-use/recharge description,
equip/use restrictions, cost, required mechanic keys and source healing dice.
Single-variant mundane families will also require an explicit variant rather
than inheriting a family identity. A data batch must extend the reviewed source
manifest and schema with its own source evidence; arbitrary families and new
properties are rejected by this initial validator.

The reviewed variant manifest lives in validation code, independently of incoming
JSON. Omitting both a variant and a claimed variant list cannot bypass it.
Duplicate IDs, duplicate JSON properties, unknown fields, unknown mechanic keys,
missing requirements, trailing JSON and invalid source/cost/charge values fail
closed. Schema changes require review; there is no free-form properties bag.

| Nested variant | Rarity | Healing data, p. 236 | Purchase cost |
| --- | --- | --- | --- |
| `basic` | Common | 2d4 + 2 | 50 GP, p. 99 |
| `greater` | Uncommon | 4d4 + 4 | Unspecified in family table |
| `superior` | Rare | 8d4 + 8 | Unspecified in family table |
| `supreme` | Very Rare | 10d4 + 20 | Unspecified in family table |

No price is inferred from rarity. `unspecified` with zero amount and `none`
currency means no listed price, never free. All four potions explicitly have no
attunement, equip requirement or charge pool/recharge. Consuming one item is
represented as a use restriction, not as a rechargeable resource. These records
contain no runtime quantity or current-charge balance.

The use keys describe a Bonus Action, consuming one potion and administering
within 5 feet (p. 99; general potion use p. 204). Required mechanic keys describe
those requirements plus healing dice (p. 236) and the normal maximum-HP cap
(p. 17). Requiring a key does not register or implement its resolver. In
particular, recording administration and stronger potions does not make them
playable.

Imports have no link to campaign state, resource balances, equipment projection,
AC calculation, creation entitlements or action registration. There is no
catalogue-to-action adapter. Unknown runtime grant properties are rejected in
schema fixtures. The existing `ResolveHealingPotion` still handles only basic
2d4+2 healing; its caller still requires an owned basic potion and conscious
self-use. `ResolveLongsword` still resolves the ordinary one-handed fixture.
Neither resolver reads this catalogue. Source data presence does not promote
coverage entries or Full Rules support.

Verification is offline: `TestItemCatalogueSourceAndVariants` joins source
presentations to the pinned inventory and checks all four variants and fresh
loads. Table-driven negative fixtures cover schema/version, duplicate/missing
IDs, source provenance, restrictions, unknown mechanics, price, charges/recharge
and invented healing. JSON fixtures also reject resource/AC/action/inventory
grants. Existing potion and longsword tests exercise their unchanged bounded
resolvers. No database or UI mutation is introduced by this change.

This work includes material from the System Reference Document 5.2.1 (“SRD
5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd.
The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.
Names, numeric values and requirements are adapted into Edda's structured item
contract; the software license does not replace this source license.
