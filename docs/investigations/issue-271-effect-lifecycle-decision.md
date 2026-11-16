# Concentration, duration and stacking decision (#271)

Observed November 16, 2026 UTC / November 15 Chicago against current
`origin/main` commit `5c4835886d07200431074a9e0fdb484a4be54162`.
This is a source-backed split under #193/#54/#47. It adds no runtime rules and
does not qualify Full Rules. Parent review and merge remain separate.

## Source and exact boundaries

The supplied `/tmp/edda-wave1/SRD_CC_v5.2.1.pdf` matches the repository pin:
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
The [official English PDF](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf)
was read locally after hashing. Page numbers below are both PDF and printed pages.
No secondary rule summaries or 2014 rules substitute for this edition.

| Named scope | Source finding | Chosen fixture boundary |
| --- | --- | --- |
| Same-named effects/different casters | p. 106: repeated castings of one spell do not combine; greatest potency applies, with latest casting winning equal potency. Different spells combine subject to their own rules. | Two Shield of Faith casters, one recipient, independent cast IDs and clocks; +2 AC once. A suppressed still-live effect resumes when the active one ends. General potency ordering remains unresolved for reconciliation. |
| Multiple recipients | Bless p. 113 targets up to three creatures within 30 feet, adds d4 to attacks/saves, and lasts concentration up to a minute. It requires a Holy Symbol worth at least 5 GP; this explicit priced material is not replaced by an arbitrary focus. Concentration belongs to the creator (p. 179). | One level-one cast and one cost own up to three explicit recipient links; caster ending removes those links. No upcasting or arbitrary roll effects. |
| Periodic/start/end triggers | Acid Arrow p. 107 is instantaneous but on a hit also damages at the end of the target’s next turn; a miss has only half initial damage. Heroism p. 140 refreshes temporary HP at start of the recipient’s turn. Spirit Guardians pp. 164–165 checks Emanation movement into a creature, creature entry and creature turn end, only once per turn. | One Acid Arrow delayed event is an implementation PR. Heroism and Spirit Guardians remain a named one-session spike because temporary HP and area/movement/phase ownership need separate prerequisites. Do not invent a generic start-turn tick for Spirit Guardians. |
| Dispel Magic | p. 124: one creature/object/magical effect within 120 feet; all ongoing spells of level three or lower on target end. Each higher-level spell needs a spellcasting ability check against 10 + spell level; upcasting raises the automatic threshold. | First implement level-three cast on one creature with only qualifying low-level spells. Higher-level checks and object/effect removal extent get a bounded spike. Removing one recipient link must not silently remove other recipients. |
| Counterspell | p. 120: reaction to a visible creature within 60 feet casting with V/S/M; Counterspell itself uses S. The target makes a Constitution save. Failure wastes casting time but does not spend the interrupted spell’s slot. | One durable reaction window around existing Cure Wounds with a trusted isolated NPC reactor. No 2014 level comparison or spellcasting ability check; no nested reaction priority or inferred NPC spell authority. |
| Damage/unconscious concentration endings | p. 179: Constitution DC is max(10, half damage rounded down), capped at 30. Concentration ends when another concentration cast starts, on incapacitation or death, or voluntarily without action. Incapacitated p. 184 breaks concentration; Unconscious p. 191 includes Incapacitated even at positive HP. | Trusted resolved-damage/condition events end only the correct caster’s cast and recipients. One check per separate damage source while concentrating. No check for zero damage or automatic ending; waking/healing never reinstates an ended cast. |

These choices paraphrase the source and define planning keys; they do not register
mechanic support. Unknown level, ownership, range, phase or disclosure facts must
keep the action unavailable. Whether Dispel Magic targets an entire area effect,
which effects count as equally potent beyond the fixed examples, simultaneous
start/end ordering, and cast interruption for non-slot resources remain explicit
unknowns assigned to the named spikes or existing #273/#277/#278 reconciliation.
Starting a concentration cast ends the old concentration before success is known;
a countered attempt cannot restore it by refunding a spell slot.

## Existing evidence and its limits

[`lifecycle.go`](../../internal/rules/srd521/lifecycle.go) defines one stored
`Concentration` (action ID, spell ID, absolute start/expiry) in a caster build.
Only conscious Cleric self Shield of Faith is enabled. It spends a level-one
slot, has a ten-minute absolute expiry, replaces the caster’s effect, and can end
voluntarily. `ConcentrationDamage` uses the source DC formula, clears at zero HP
without dice, and otherwise records a Constitution check. This is narrow fixture
support, not general damage/condition or multi-caster support.

[`caster_state.go`](../../internal/engine/caster_state.go) saves casting fields
and the seconds clock under the caller’s transaction, preserving unknown casting
JSON keys. [`encounter.go`](../../internal/engine/encounter.go) invokes the damage
check only from its resolved ordinary NPC attack. The
[combat encounter](../../internal/combat/encounter.go) owns a bounded two-actor
turn cursor, action/Bonus Action budget and one-slot-per-turn flag. Existing
opportunity reactions suspend NPC withdrawal; that does not implement an arbitrary
pending spell attempt or Counterspell. The
[lifecycle preview](../../internal/game/lifecycle_preview.go) previews supported
self actions and explicit uninterrupted rest; it supplies no general target graph.

The [rules fixtures](../../internal/rules/srd521/lifecycle_test.go) cover
replacement/ending/expiry, damage-save rounding/cap and the zero-HP case.
[`TestIntegrationCasterLifecycle`](../../internal/handlers/lifecycle_integration_test.go)
contains saved effect identity, replay/reload, stale preview, rollback, NPC damage
and encounter expiry assertions. It has `//go:build integration`, so those
PostgreSQL assertions were inspected, not executed by this decision’s short suite.
The existing browser fixture is historical test evidence, not a device run here.
See also [the existing bounded lifecycle description](../caster-lifecycle.md).

## Ownership, prerequisites and delivery

The [machine-readable mapping](issue-271-effect-lifecycle-scope.json) records
source pages, fixture, targets, costs, timing, proposed mechanic keys, exclusions,
exact native prerequisites, issue numbers and qualification gates for each child.
The [readback receipt](issue-271-link-verification.json) records native parents,
all repository issue dependencies and hashes of the fetched child bodies.

Search of the live issue list found no existing implementation owner for these
six fixtures. Existing #77 is the closed self-only baseline, #81 the closed
opportunity-reaction baseline, and #68 the closed Cure Wounds baseline; they are
reused rather than duplicated. #194 owns the spell contract, and catalogue owners
remain #196 Acid Arrow, #198 Bless, #201 Counterspell, #203 Dispel Magic,
#209 Heroism, #218 Shield of Faith and #220 Spirit Guardians. No new catalogue
issues are created. #269 owns condition acquisition and general consequences;
the new concentration adapter consumes trusted events without implementing those
conditions. #272 owns rest/recharge and temporary-HP recovery boundaries. #273
owns item/free-use casting and resources. #277/#278 retain residual mechanic-key
reconciliation, including arbitrary potency rules and source exceptions.

Higher-level Acid Arrow and Dispel Magic fixtures are isolated authorized engine
fixtures. Their issues require explicit entitlement evidence and keep production
discovery unavailable until exact #332 Wizard / #323 Cleric implementation children
supply eligibility. Closing those investigations alone cannot unlock them. The
Counterspell NPC is similarly an explicit isolated fixture, not general NPC AI.
This limits what each one-PR implementation can claim while keeping it independently
verifiable. No unknown class implementation is represented as a completed prerequisite.

Six implementation issues and two bounded spikes attach directly under #193.
Every one blocks both #279 and #280 natively, in addition to their existing
blockers; spikes must add their future terminal implementations directly to both
gates before closing. None depends on a gate. Published prerequisites encode
actual contracts/source inputs or the shared fixture being extended, rather than
serializing independent work. The open #271 decision is not added as a prerequisite
to its children: this PR’s review fixes the written decision before implementation.

| Published child | Type | Exact blockers |
| --- | --- | --- |
| [#351: Track two Shield of Faith casters with independent recipient effects](https://github.com/subculture-collective/edda/issues/351) | task | #194, #218 |
| [#352: Apply one level-one Bless cast to up to three explicit recipients](https://github.com/subculture-collective/edda/issues/352) | task | #198, #351 |
| [#353: End the exact concentration cast on resolved damage and incapacitation events](https://github.com/subculture-collective/edda/issues/353) | task | #351 |
| [#354: Resolve Acid Arrow initial damage and one durable next-turn end tick](https://github.com/subculture-collective/edda/issues/354) | task | #196 |
| [#355: Dispel level-three-or-lower spells on one explicit creature](https://github.com/subculture-collective/edda/issues/355) | task | #203, #352 |
| [#356: Persist one Counterspell reaction window around Cure Wounds casting](https://github.com/subculture-collective/edda/issues/356) | task | #201 |
| [#358: Split Heroism start-turn and Spirit Guardians entry/end timing fixtures](https://github.com/subculture-collective/edda/issues/358) | spike | Existing baseline only |
| [#359: Split Dispel Magic higher-level checks and object/effect targeting](https://github.com/subculture-collective/edda/issues/359) | spike | #203, #355 |

## Verification

Both requested commands passed on this decision branch:

- `go test -short -race ./...`
- `go vet ./...`

Reproduce documentation, mapping, pin and dependency verification:

```sh
python3 scripts/verify_effect_lifecycle_scope.py --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
python3 scripts/verify_effect_lifecycle_scope.py --live --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
git diff --check
```

Live verification requires an authenticated GitHub CLI (`gh`).
Offline verification checks complete unique scope ownership, child fields, exact
prerequisites, recorded parents, both gate links and acyclicity. `--live` fetches
every repository issue’s native graph and reads each child body back, rejecting
pagination omissions or cycles. The receipt is a dated snapshot; future tracker
changes require a fresh live check. No production database, migration, deployment,
frontend build, browser/device journey or tagged integration suite ran here.
Those remain required checks of the implementation children where applicable.

This work includes material from the System Reference Document 5.2.1 (“SRD
5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd.
The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International
License, available at https://creativecommons.org/licenses/by/4.0/legalcode.
Changes: rules are paraphrased into fixture decisions, page references and
planning metadata. See [pin, license and attribution](../srd-5.2.1-conformance.md).
