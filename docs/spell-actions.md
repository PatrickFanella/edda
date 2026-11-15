# Cure Wounds casting fixture

Issue #68 adds a level-one, unarmored Cleric fixture to the shared sheet
prepare/confirm path. It supports Cure Wounds on the conscious caster only,
using one level-one spell slot and a Magic action. This is a bounded mechanical
fixture, not a complete Cleric build or a new character creation option (#72).

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
