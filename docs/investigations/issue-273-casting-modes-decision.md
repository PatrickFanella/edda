# Ritual, component and item-casting ownership (#273)

Most named casting-mode rules already have bounded published execution owners. Reuse them without changing their bodies or treating planning closure as execution. The remaining item-user ability/default-modifier boundary is assigned to **#2051**, one Hat of Disguise adapter consuming the existing Disguise Self effect. Its exact publication and verification are recorded in the companion scope and receipt.

## Source and existing owners

The English SRD 5.2.1 PDF matches SHA-256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`. Source facts and fixtures retain the CC-BY-4.0 attribution and adaptation notice in [the source register](../srd-5.2.1-conformance.md).

| Named scope | PDF | Reused implementation contracts and boundary |
|---|---|---|
| Ritual casting | 104–105 | #906 and #1449: prepared Ritual-tag spell, normal casting time plus ten minutes, no slot and no extra Ritual Adept requirement. #1449 owns each-turn Magic action, casting Concentration, interruption with no slot spent and restart. Wizard's own-book preparation exception remains class-owned. Ritual casting does not upcast. |
| Costly and consumed materials | 104–105 | #1447: actual available material with specified price/consumption; a focus replaces only eligible unpriced, unconsumed M. Preserve V/S and hand/armor requirements for normal casting. Item/scroll exceptions remain separate. Named effects retain their actual material costs and results. |
| Slot and free-use routing | 104–106 | #1448: qualified spell/slot level, one slot expenditure to cast per turn, private authoritative invalid-target costs versus unavailable unknown facts, and separate item/free channels. Feature free-use eligibility and recovery remain with the feature owner, including Magic Initiate #386. A slot-to-resource conversion is not a slot spent to cast. |
| Spell Scroll casting | 205–206, 244 | #1142: all 339 canonical identities have typed conditional dispatch, fixed scroll DC/attack values, ordinary class-list eligibility and the normal higher-level check. #1133 separately owns Thief's level-2–9 INT Arcana check and its destruction exception. Permission never supplies an unqualified effect resolver. |
| Item spells and Concentration | 206, 251 | #1140 casts Detect Magic from a Wand of Magic Detection through #906, with the item's no-component/no-slot channel and the spell's actual Concentration. #1148 owns this Wand's separate dawn recharge. These are neither a generic item catalogue nor a potion-casting rule. |
| Duration, dismissal and stacking | 106 | #1450: capable caster's no-action dismissal, source-bound expiry and strongest/latest tied same-spell effects. Incapacitated alone does not end a non-Concentration spell. The individual effect owns any explicit recast ending. |

The four #1447–1450 protocol contracts are limited to the accepted 245-identity finite domain. Their publication is not support for all 339 spells. Spell execution outside that domain remains with the existing, catalogue-blocked residual reconciliation #277; item effects outside the named fixtures remain with #278. Neither planning parent is an executable resolver.

## Bounded residual item adapter

PDF206 also defines item-user spellcasting ability selection and the +0 fallback when the user has no spellcasting ability. Whole-body inspection found no existing task that binds those operands together with an actual named item/effect. Hat of Disguise on PDF224 supplies a small level-one fixture: wearing and attunement permit Disguise Self; removal ends the spell. The existing #1293 owns the actual illusion. The new task owns only that Hat's attunement and casting adapter, including DC10 for no ability/PB2, and DC14 or DC15 for explicitly selected CHA3/WIS4 with PB3.

The new task does not widen #1141's four-item attunement domain, copy #1293 or require the four unrelated item effects. It consumes the canonical completed Short Rest #427 and lifecycle #1450. Normal attunement capacity remains three; a different source-qualified capacity stays a separate adapter. A Magic Initiate ability is not erased merely because the Spellcasting feature is absent.

The task #2051 is a child of #273 and directly required by item qualification #280. Its exact prerequisites are #195, #427, #1293 and #1450. It adds no spell effect: spell qualification #279 already requires the unchanged #1293 through the accepted shared aggregate. All historical body, parent and prerequisite contracts remain intact. The full spell gate is already at its 100-blocker cap; this item-only task needs no new edge there. No new direct primitive #67 or transaction #429 consumer is introduced.

## Evidence and limits

The source matrix records the named clauses, existing owner body hashes, actual dependency paths and the one residual task. Fresh publication readback at 2026-11-17T07:50:45Z checked its exact body, parent, blockers, OPEN state and Task/Size M, plus the mandatory #280 edge. All 1,888 prior graph contracts remain unchanged apart from adding #2051 to #280. Both DAGs and all three native caps pass across 1,889 issues; #279 has 100 blockers, #280 has 96, #429 has 100 consumers and #283 has 95 children. Decision closure leaves all implementation obligations mandatory.

Current gameplay callers remain bounded Cure Wounds/Shield of Faith fixtures. Catalogue records, imported spells, published tasks and this decision grant no general ritual, item or class execution. Runtime implementation remains paused. This change contains no runtime files, database migration, UI behavior, production mutation or Full Rules admission. Focused source/native checks and hosted CI apply to this decision; implementation tasks retain their own runtime, PostgreSQL and device requirements.
