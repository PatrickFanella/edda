# Cure Wounds casting fixture

Issue #68 adds a level-one, unarmored Cleric fixture to the shared sheet
prepare/confirm path. It supports Cure Wounds on the conscious caster only,
using one level-one spell slot and a Magic action. This is a bounded mechanical
fixture, not a complete Cleric implementation. Issue #72 adds a valid fixed creation build.

The sheet exposes Wisdom, known/prepared status, verbal and somatic components,
component availability and current/maximum level-one slots. Choose Cure Wounds,
inspect the supported slot level, preview its healing and cost, explicitly select
self, then confirm. Preparation never rolls or changes state. Confirmation rolls
2d8 + Wisdom, caps healing at maximum HP, decrements one slot and saves the
receipt and history in the existing transaction. Reusing the same action ID,
including after an engine restart or consuming the last slot, reads the saved
result without rolling again.

The fixture is `internal/rules/srd521/testdata/cleric1.json`. Its explicit
`stats.srd521.casting` object records `prepared` (Cure Wounds only), `slots`
(0–2), and `verbal`/`somatic` availability. Those booleans are trusted scene
rulings, not client-selected overrides or a general condition system. Other
stored character stats are retained when HP and slots change. No schema change,
automatic legacy conversion, production seed or production mutation is needed.

Missing preparation, unavailable components, zero slots, other targets, higher
slots and unsupported builds cannot cast. The revision hashes the campaign and
entire actor snapshot, including casting resources; confirmation reloads and
locks them. A failed history write rolls back both HP and slots. Supported
encounters enforce their action budget; active legacy combat is rejected.
The current Cleric fixture does not start the Fighter's longsword encounter.
Concentration, duration and rest recovery remain #77; the wider catalogue and
additional Cure Wounds targets/levels remain #80. The full spell entry stays
unsupported in the coverage register.

The pinned [SRD 5.2.1](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf)
provides Cleric core traits and spellcasting (pp. 36–38), spell components and
self-targeting (pp. 105–106), Cure Wounds (p. 121) and capped healing (p. 17).
The mechanical fixture is an adaptation; see
[the pin, license and required attribution](srd-5.2.1-conformance.md).
No source prose is reproduced here.

## Verification

- `internal/rules/srd521/spell_test.go`: Cleric projection, ordinary/capped/full-HP
  healing, missing preparation/components/slots and invalid d8 values.
- `internal/handlers/spell_action_integration_test.go`: explicit target and slot
  validation, stale resource rejection, injected history failure with atomic
  rollback, exact healing/cost and durable replay after slot exhaustion.
- `internal/handlers/srd_item_receipt_test.go`: typed player-safe receipts agree
  between confirmation and reconstructed history, excluding private envelopes.
- Frontend interaction and desktop/phone Chromium fixtures: explicit slot/target
  selection, no early confirmation, refreshed HP/slots and transcript receipt.

These isolated tests do not qualify real device keyboards, all Cleric features,
live models, or production deployment.

## Guided caster creation (#72)

The existing setup flow also accepts Human Acolyte Cleric. The server validates
this combination before generation and persists the same casting contract used
by play. Its standard array is STR 10, DEX 12, CON 14, INT 8, WIS 15, CHA 13;
Acolyte adds +1 INT/WIS/CHA. The unarmored projection has 10 HP, AC 11, Wisdom
casting +3, two level-one slots and one d8 Hit Die.

Acolyte grants Insight, Religion, Calligrapher's Supplies and Magic Initiate
(Cleric; Wisdom): Light, Mending and Detect Magic with one recorded free use.
Class skills are Medicine and Persuasion; Divine Order is Protector. Human adds
Perception and Skilled (Athletics, Survival, Nature). Three class cantrips are
Guidance, Sacred Flame and Thaumaturgy; four prepared spells are Bless, Cure
Wounds, Guiding Bolt and Shield of Faith. These choices are recorded, but only
Cure Wounds on self has an executable spell action in this slice. Feat spells
and unimplemented class spells do not acquire actions merely by being recorded.

Cleric equipment option B contributes 110 GP. Acolyte option A contributes
Calligrapher's Supplies, a prayer book, Holy Symbol, ten parchment sheets, a robe
and 8 GP: 118 GP total, with no armor purchased or equipped. Client-authored
starter items are excluded, as for the Fighter. Generation, inventory and setup
remain one transaction. A database test creates this caster and confirms Cure
Wounds directly, without modifying its character data; desktop/phone creation
fixtures retain choices through a failed generation and retry.

Sources in the pinned SRD: character creation pp. 19–22, Cleric pp. 36–38,
Acolyte p. 83, Human p. 86, and Magic Initiate/Skilled p. 87. This is an adaptation
under the attribution linked above. Broader character choice and advancement
remain later roadmap work; no legacy data is converted.
