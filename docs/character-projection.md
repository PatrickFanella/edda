# Read-only SRD character projection (#58)

GET /api/v1/campaigns/{id}/character keeps the legacy character fields and adds
rules_version. Only SRD campaigns attempt a server-derived sheet. Missing or
unsupported builds return sheet_unavailable; they never use the legacy XP curve.
The TUI can consume the same srd521.Project function without web dependencies.

This slice projects an explicit level-one Fighter with chain mail, the Defense
fighting style, an optional shield, six scores, recorded skill proficiencies,
and current Second Wind/Hit Dice counters. It is not a character-creation validator,
equipment catalogue, or action resolver. The martial creation validator is described below; caster creation remains #72.

## Supported creation slice

The guided Crunch creation flow now creates one catalogue-backed SRD 5.2.1
character: a Human Soldier Fighter at level 1. It uses the standard array with
the Soldier background's +1 Strength, +1 Dexterity, and +1 Constitution
adjustment, resulting in Strength 16, Dexterity 15, Constitution 14,
Intelligence 8, Wisdom 10, and Charisma 12. The recorded choices are Soldier's
Savage Attacker feat, Athletics and Intimidation, Dice Set proficiency (the gold equipment option grants no physical set), the Human
Skillful skill (Insight), the Human Versatile feat (Skilled: History, Medicine,
and Nature), Fighter Perception and Survival, Defense fighting style, and
Longsword/Greatsword/Javelin weapon mastery.

The persisted equipment is Fighter package A plus the Soldier 50 GP option.
It records chain mail, greatsword, flail, eight javelins, a Dungeoneer's Pack,
a longsword, shield, and the remaining 29 GP. The longsword,
chain mail, and shield are marked equipped so the sheet's AC and action fixture
have matching provenance. Feat effects, tool use, Human resource, mastery
effects, and all other origin/class choices remain unimplemented. A request for `srd-5.2.1` with any other guided origin,
class, or background is rejected before world generation, so it cannot report
an incomplete build as playable. Legacy campaign creation remains separate.

These choices are derived from the pinned SRD 5.2.1 Character Creation pages
19–21, Fighter page 47, and Character Origins pages 83–86. The validation and
persistence evidence is `internal/rules/srd521/creation_test.go` and
`internal/handlers/world_creation_integration_test.go`; it does not establish
broader character-creation coverage.
The deterministic fixture is internal/rules/srd521/testdata/fighter1.json; no
production character or campaign is converted or seeded by this change.

The stored build is beneath stats.srd521. Current/max HP, level and XP remain
canonical character columns. Invalid vitals and unsupported levels/classes fail
closed. The sheet exposes all six ability modifiers/saves, all 18 skills,
proficiency, HP, AC/initiative, resources, capabilities and the server XP threshold.
Every derived value carries its calculation terms and stable source identifiers.
Capabilities are explicitly read only; no rolls, costs or class features execute.
The separate [check resolver and receipt fixture](srd-checks.md) consumes this
projection and supports optional recorded skill expertise. It does not execute
these capabilities or enable confirmed character-sheet actions.
[Action previews](action-previews.md) add read-only preparation for supported
checks and saves without performing them.

Source identifiers use srd-5.2.1:content/<family>/<slug>. The source map is:

| IDs | Pinned SRD PDF pages |
| --- | --- |
| character/ability-modifiers | 5–6 |
| character/proficiency, character/skill-proficiencies | 8–9 |
| character/hit-points, character/hit-dice | 16–18, 22–23 |
| character/level-advancement | 23–24 |
| fighter/core-traits, fighter/second-wind, fighter/weapon-mastery | 47–48 |
| armor/chain-mail, armor/shield | 92 |
| feats/defense | 87 |
| combat/initiative | 13 |
| actions/attack | 9 |

Reference: [pinned English SRD 5.2.1](https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf).
The attribution and artifact digest remain in docs/srd-5.2.1-conformance.md.
These calculations are Edda adaptations. This narrow projection does not promote
a whole class, character level, action family or catalogue to implemented coverage.

SRD starting inventory comes from the validated build. Legacy starter-package item grants are ignored for this version; narrative starting facts and relationships remain available. The selected gold option leaves 29 GP after buying the longsword and shield; no physical gaming set is purchased.
