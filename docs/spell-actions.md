# Cure Wounds casting fixture

Issue #68 adds a level-one, unarmored Cleric fixture to the shared sheet
prepare/confirm path. The original slice supports Cure Wounds on the conscious caster,
using one level-one spell slot and a Magic action. This is a bounded mechanical
fixture, not a complete Cleric implementation. Issue #72 adds a valid fixed creation build.

The sheet exposes Wisdom, known/prepared status, verbal and somatic components,
component availability and current/maximum level-one slots. Choose Cure Wounds,
inspect the supported slot level, preview its healing and cost, explicitly select
a prepared target, then confirm. Preparation never rolls or changes state. Confirmation rolls
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

Missing preparation, unavailable components, zero slots, unqualified targets, higher
slots and unsupported builds cannot cast. The revision hashes the campaign and
entire actor snapshot, including casting resources; confirmation reloads and
locks them. A failed history write rolls back both HP and slots. Supported
encounters enforce their action budget; active legacy combat is rejected.
The current Cleric fixture does not start the Fighter's longsword encounter.
Concentration, duration and rest recovery use the [bounded lifecycle fixture](caster-lifecycle.md) from #77; the wider catalogue and
further Cure Wounds targets/levels remain roadmap work; the bounded ally slice is described below. The full spell entry stays
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
Cure Wounds on self has an executable spell action from #68; #77 adds Shield of Faith on self. Feat spells
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

## Willing adjacent ally (#263)

Cure Wounds also accepts one player-known allied NPC in the caster's current
location, outside the two-actor encounter fixture. The public preview offers
only qualified targets; selecting a target and inspecting the level-one slot
cost precede confirmation. Self-casting and legacy reads retain their existing
paths. The spell catalogue's self fixture and wider coverage status are unchanged;
this additional qualified slice does not enable arbitrary spell records.

Eligibility is a trusted fixture ruling stored in NPC `properties.srd521_ally`:

```json
{
  "profile": "cure-wounds-ally-v1",
  "actor_id": "8b2d58b7-9952-40bf-8641-af62ed37334e",
  "location_id": "5d6c90cf-55d1-4416-b2cc-248f586d640e",
  "allied": true,
  "willing": true,
  "conscious": true,
  "hidden": false,
  "conditions": "ordinary",
  "max_hp": 10,
  "distance_feet": 5,
  "line_of_sight": true,
  "cover": "none"
}
```

The example actor UUID identifies the fixture caster; the actual ruling must
name the campaign's caster. The location UUID binds the touch ruling to the
location where it was adjudicated. Both participants must currently be there;
moving them to another location does not transfer the old ruling, even after a
new preview. A trusted ruling for the new location is required. The existing
NPC type has current HP but no maximum HP field, so this bounded fixture explicitly supplies `max_hp`. All booleans,
ordinary conditions, maximum HP and spatial facts are required. Unknown fields
also fail closed, including case aliases and an input `revision` property. The
fixture exposes only distance, sight and cover from the spatial contract; its
revision is derived by the server. Positive NPC HP and alive state are
independently required; positive HP alone does not establish consciousness. Existing session history
must identify the NPC in `npcs_involved`. Disposition, faction, descriptions and
model output never establish alliance or willingness. `create_npc` rejects the
reserved `srd521_ally` key; `update_npc` does not accept properties. No fixture
installer, model inference, migration or production seed is added.

The existing `combat.Space` validates exact distance (0–5 feet), sight and cover;
this slice additionally requires no cover. These visibility, consent, ordinary
condition and five-foot restrictions bound Edda's supported fixture; they do not
claim to be additional SRD Cure Wounds requirements. Unconscious targets,
unresolved facts, hidden NPCs, other slots, conditions and encounter allies stay
unavailable. Hidden statistics/properties are never copied into public previews.

Confirmation locks and reloads the NPC rows in the same transaction as the
caster. The opaque revision includes qualified target HP, maximum HP, spatial
facts and update timestamp. Changed membership or eligibility invalidates a
prepared revision before dice. Healing changes only the chosen NPC's HP;
slot cost changes only the caster. Both changes, the dice/target receipt and
history commit together. Receipt lookup precedes revalidation, so replay after
reload or slot exhaustion returns the original result without another roll.

Additional evidence lives in `internal/handlers/ally_spell_integration_test.go`:
real PostgreSQL scene mutations, HTTP ownership and public preview checks,
failed-history rollback, deterministic 2d8 plus Wisdom, capping, new-engine
replay, receipt reload and zero-slot rejection. Existing receipt projection
fixtures exercise the shared web/TUI representation. Frontend browser fixtures
at 1440 and 390 pixels exercise ally selection, explicit confirmation, lost
response retry and reload reconciliation using API doubles. Those doubles do
not establish database atomicity, and the phone width is not a physical phone.

This adaptation uses the pinned SRD 5.2.1 PDF pp. 14–17, 105–106 and 121, verified
locally with SHA-256
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The [CC-BY-4.0 attribution and adaptation notice](srd-5.2.1-conformance.md)
continue to apply.
