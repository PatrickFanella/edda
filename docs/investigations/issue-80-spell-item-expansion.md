# Spell, item and effect expansion decision (#80)

Observed 2026-11-15 against Wave 12 main `077d008` (#189/#190), following
Wave 11 `892fc87`. This decision extends #47/#54 with source-linked work. It does not
implement the remaining catalogue or qualify full SRD coverage.

## Source and inventory

The English SRD 5.2.1 PDF matches the repository pin:
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The [official PDF](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf)
contains these indexed source families:

| Family | Named entries | Source pages |
| --- | ---: | --- |
| Spells | 339 | 107–175 |
| Magic-item headings | 258 | 209–253 |
| Weapons | 37 | 91 |
| Armor and shield | 13 | 92 |
| Tools | 25 | 93–94 |
| Adventuring gear | 81 | 94–100 |
| Total | 753 | |

These are source families, not a count of executable effects or every variant.
For example, Potions of Healing is one magic-item family with several strengths;
weapon +1/+2/+3 and other source variants must be explicit nested records during
import. Potion of Healing and Spell Scroll also appear in mundane gear; retain
both source presentations and deduplicate their actual runtime identity.
Packages, ammunition, foci, musical instruments and gaming sets also contain
nested variants or constituents. Data batches must assert those are complete.

The companion [inventory](issue-80-spell-item-inventory.json) records title, page,
stable planning ID, current unsupported/partial
boundary and one exact owning data issue. No spell prose, item descriptions or
mechanical values are transcribed here. `scripts/inventory_srd_expansion.py`
reproduces the title/page inventory offline with Poppler after checking the PDF
hash. It joins wrapped headings and the small-cap Acid Splash title, excludes
summoned-creature Actions/Traits, and reads weapon/armor tables separately.

Mount/vehicle tables, lifestyle/services, currency, crafting and broader rules
sections are not counted among these 753 named families. Separate bounded
investigations own those exact source sections; they cannot be silently omitted
from the equipment qualification gate. Creature stat blocks remain the #82
inventory and #184; spell ownership/control is coordinated with #181/#185.

## Actual support and remaining boundaries

| Current evidence | Supported boundary | Still unavailable or unqualified |
| --- | --- | --- |
| `spell.go`, `spell_preview.go`, `spell_action_integration_test.go` | Level-one Cleric Cure Wounds, conscious self, one level-one slot, explicit components, capped healing and committed receipt replay | Ally/enemy eligibility, touch/range adjudication, higher slots, general spell attack/save/area/ongoing effects, spell-list breadth |
| `lifecycle.go`, `caster_state.go`, `lifecycle_integration_test.go` | Shield of Faith on self, +2 AC, one concentration identity, absolute ten-minute expiry, damage Constitution receipts, replace/end and ordinary six-second rounds | Multiple recipients/casters, arbitrary stacking, dispelling/countering, periodic triggers, concentration exceptions, other duration/condition effects |
| Same lifecycle fixtures | Explicit uninterrupted Cleric Short/Long Rest, one optional d8, source-limited slot/Hit Die/free-use recovery | Interrupted rests, other classes/species/features, exhaustion, maximum/temporary HP exceptions, item dawn recharge and destruction |
| `consumable.go`, `consumable_action_integration_test.go` | Basic Potion of Healing on conscious self; committed quantity/cost/healing receipts and ordinary encounter Bonus Action | Administering to allies, unconscious/unwilling targets, stronger potions and other activated items |
| Fixed creation/projected build | Chain-mail/shield Fighter and unarmored Cleric; recorded equipment, prepared spells, feats and source-derived AC | Equipping/don/doff flow, general weapon/armor/tool/gear interactions, attunement, charges, curses, sentience and magic-item activation |
| `coverage.json`, legacy inventory/tools | Explicit partial family register; legacy free-form tools stay guarded for versioned campaigns | A source-complete validated spell/item catalogue; catalogue presence cannot enable unsupported mechanics or Full Rules |

The Human Cleric records Guidance, Sacred Flame, Thaumaturgy, Bless, Guiding
Bolt, Light, Mending and Detect Magic without making their effects executable.
Magic Initiate's free use is stored/recovered, but that does not implement ritual
casting, feat spells or item-casting rules. The rule paths remain shared by web
and TUI; this decision does not create another combat or spell engine.

## Batch decision

Use three tracking groups beneath #54: #191 for spells, #192 for equipment and
magic items, and #193 for effect interactions. Data-only work is capped at 12 named
families per PR: 29 spell batches, 22 magic-item batches and 16 mundane batches.
Each must preserve every source-listed variant and required mechanic key. A
record is playable only when every required key has source-linked resolver
fixtures. Unknown keys fail closed.

Two initial contracts define spell and item/variant records, without granting
new playable actions. Five runtime slices cover one adjacent ally's Cure Wounds,
Light on one known object, Detect Magic from the recorded feat use, one allied
basic potion administration and chain-mail/shield equip changes. Each retains
explicit targets/costs/timing, stale-preview checks, atomic history and committed
receipt replay. Catalogue imports remain independently reviewable from effects.

Eleven bounded investigations own the three named condition groups, stacking
and duration, interrupted recovery/recharge, ritual/component/item casting,
mounts/vehicles, prices/services/lifestyle, crafting and residual spell/item
mechanic-key reconciliation. Reconciliation waits for all relevant data batches
so every record and key has an implementation owner. Decision closure alone
never supplies evidence for its resulting mechanics. Before closing, each
splitter must publish exact one-PR children and add their terminal implementation
blockers to the full spell/item qualification gates.

The reconciliation prerequisite sets are intentional: 29 spell imports and
38 item imports are independent inputs. They cannot be replaced by a serial
chain merely to reduce fan-in. This is the full existing coverage roadmap, not
an added release promise. The graph checker warns about size and fan-in; the
bounded data ownership and final gates are the reason to accept those warnings.

Find Familiar uses existing creature control decision #181. Animate Objects,
Find Steed, Giant Insect and Summon Dragon use #185, with creature data #184.
Their spell prerequisites must be explicit, and their implementation children
must block both the relevant spell gate and #183 world qualification. Do not
create duplicate catalogue/stat-block owners or count summons as playable from
a parsed block.

The two final gates, #279 for spells and #280 for items, distinguish complete
data, implemented mechanics and qualification. #88 reports actual partial coverage while this work remains.
#89 qualifies only its declared playable slice. #54 stays open until all required
implementation and conformance work is complete or the owner explicitly narrows
that target. Closing #80 is only the batch decision.

## Published execution order

All 87 work issues are #194–#280, with source families and exact blocker sets
recorded in the companion inventory. Tracking groups #191–#193 are children of
#54 and hold those work issues natively.

After #80 closes, #194/#195 establish the spell/item contracts. The nine named
condition, timing/recovery, casting, mount/economy/crafting decisions #268–#276
can also start. Data imports #196–#224 follow #194; #225–#262 follow #195.
Runtime fixtures #263–#267 name their cross-domain prerequisites explicitly.
Residual mechanic reconciliation #277/#278 waits for every respective data
batch. Qualification #279/#280 waits for those decisions, their implementation
children and the named runtime fixtures.

The active-source prerequisites added to #181/#185 use exact spell data owners
from this inventory. The future implementation blockers required by their
acceptance criteria remain the responsibility of those bounded investigations.

## Verification and provenance

Run the offline reproduction with a locally supplied pinned PDF:

```sh
python3 scripts/inventory_srd_expansion.py --pdf /tmp/edda-srd521-wave11.pdf --check docs/investigations/issue-80-spell-item-inventory.json
```

The investigation also checks unique IDs, exactly one data owner per indexed
family, batches of at most 12, acyclic native blockers, issue body/parent readback
and Edda Project Size M. Full Go and PostgreSQL/browser checks qualify the merged
application changes separately; this inventory adds no production runtime code.

This work includes material from the System Reference Document 5.2.1 (“SRD
5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd.
The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.
Names/page references and source families are adapted into Edda planning metadata;
no rules text or numeric item/spell catalogue is imported. See the existing
[pin, license and attribution](../srd-5.2.1-conformance.md).
