# Read-only SRD character projection (#58)

GET /api/v1/campaigns/{id}/character keeps the legacy character fields and adds
rules_version. Only SRD campaigns attempt a server-derived sheet. Missing or
unsupported builds return sheet_unavailable; they never use the legacy XP curve.
The TUI can consume the same srd521.Project function without web dependencies.

This slice projects an explicit level-one Fighter with chain mail, the Defense
fighting style, an optional shield, six scores, recorded skill proficiencies,
and current Second Wind/Hit Dice counters. It is not a character-creation validator,
equipment catalogue, or action resolver. Creation validation remains #69/#72.
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
these capabilities or enable live character-sheet actions.

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
