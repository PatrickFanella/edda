# Martial attack fixture

Issue #65 adds one explicit SRD 5.2.1 fixture to the shared prepare/confirm path:
a level-one Fighter's ordinary, one-handed longsword attack against a known,
colocated unarmored fixture target. Legacy actions retain their existing path.
The broad combat, equipment and Fighter coverage entries remain unsupported.

In the sheet, select an equipped longsword, preview its Strength plus proficiency
modifier, inspect damage and the Attack action cost, select a known target, and
confirm. Preview never rolls. Confirmation rolls once, applies damage, and saves
its typed receipt and transcript in the existing campaign transaction. Retrying
the same action ID returns the saved result, including after an engine restart.
A changed item, actor, target or campaign revision requires a fresh preview.

## Explicit fixture setup

This is development fixture content, not a conversion of arbitrary old equipment
or NPCs. No production data is changed or automatically annotated.

- The existing supported SRD character build is required.
- An owned, equipped weapon item with positive quantity has
  `properties.srd521_id = "longsword"`.
- A living NPC with positive HP at the actor's current location has exactly
  `properties.srd521_target = {"profile":"unarmored-fixture-v1",
  "actor_id":"<the fixture actor UUID>","distance_feet":5}`. The integration
  test constructs that UUID from its isolated character; this is a data contract.
- The target must already appear in this campaign's `session_logs.npcs_involved`.
  The fixture profile declares AC 10, ordinary circumstances, no conditions,
  cover, resistance, immunity, vulnerability or mastery effects.

Only targets satisfying that entire contract are offered. The profile does not
implement general target/range adjudication. Active legacy combat is rejected;
initiative, turn ownership, action economy, mastery and reactions remain later
roadmap work. The Attack action cost is displayed and recorded, with no claim
that a combat-round budget is enforced.

## Rules and evidence

The pinned [SRD 5.2.1](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf)
provides attack rolls (p. 7), melee attacks and damage/critical hits (pp. 15–16),
Fighter weapon proficiency (p. 47), and the longsword (p. 91). The fixture uses
Strength plus proficiency to hit, 1d8 plus Strength slashing on a hit, a natural
1 miss, and a natural 20 hit with two damage dice and one modifier. Damage cannot
be negative. See [the source pin and attribution](srd-5.2.1-conformance.md).
These are adapted mechanical fixtures; no source prose is reproduced.

`internal/rules/srd521/attack_test.go` covers ordinary hits/misses, natural 1/20,
critical damage and invalid dice. `internal/handlers/martial_action_integration_test.go`
covers hidden/invalid targets, stale equipment, exact damage, history and durable
replay. Frontend unit and desktop/phone browser fixtures cover explicit target
selection and confirmation without early submission.
