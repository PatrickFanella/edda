# Feat decision for #334

Observed 2026-11-16 at `0abaffbaaef23c07f6aa95b559680ccabaae2824`; evidence refreshed at
`61a80bdaa9bfa1ee5540023c7809d94244ed3c02` after the Shield of Faith merge.
This decision publishes work; it does not enable mechanics, grants, coverage or Full Rules.
#334 closed through accepted decision PR #403. Its implementation children remain open.

The [matrix](issue-334-feat-scope.json) preserves every record, nested choice, predicate,
numeric parameter and source key from the four canonical feat catalogues, joins all 17 exact
identities to current evidence, and assigns each unique required mechanic key one owner.
The [receipt](issue-334-link-verification.json) records full-repository native parent/blocker
readback, body hashes and observation time. `scripts/verify_feat_scope.py` checks this offline;
`--live` repeats read-only GitHub inspection and refreshes the receipt only after validation.
`--pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf` checks the exact PDF SHA.

The pinned English SRD 5.2.1 PDF SHA is
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
PDF pp87–88 were read directly, along with advancement/prerequisite/Constitution rules
pp23–26, Fighter Fighting Style/ASI/Epic Boon rules pp47–49, and Truesight/Unarmed
Strike/teleportation p190. Source attribution and CC-BY-4.0 adaptation notice remain in
`docs/srd-5.2.1-conformance.md` and the manifest. These are adapted source facts, not verbatim rules.

## Current evidence and exact variants

Four Origin, two General, four Fighting Style and seven Epic Boon definitions exist after
#298/#299/#300/#301 closed. Catalogue source-oracle tests prove metadata equality, not
execution. The manifest copies each owned record verbatim so prerequisite alternatives,
choice counts, repeatability and ASI caps remain machine checkable.

`creation.go` records Savage Attacker without implementing its optional weapon damage
rolls. Skilled has fixed skills in the two builds; it lacks general three-skill/tool and repeat
choices. Magic Initiate records Cleric/Wisdom, Light/Mending and Detect Magic, with one
free use; recorded choices do not execute every spell or implement general list/ability,
repeat and level-change choices. #264/#265 remain their exact Light/Detect Magic owners.
`character.go` implements Defense +1 AC only within the fixed supported Fighter/armor
boundary. Generic legacy `internal/rules/feats.go` Alert (+5/no-surprise) is a different source
variant and must never satisfy SRD Alert (optional PB and initiative swap). Legacy names
and `grant_feat` are not source-linked production eligibility.

ASI permits level4+, one +2 or two distinct +1, cap20, repeatable. Grappler requires
level4+ and STR or DEX13+, permits STR/DEX+1 cap20, and is not repeatable. Each boon
requires level19+, is not repeatable, and has +1 cap30: any ability except Irresistible Offense
(STR/DEX) and Spell Recall (INT/WIS/CHA, additionally requires Spellcasting feature).
Magic Initiate is repeatable only with a different one of Cleric/Druid/Wizard lists;
Skilled is repeatable; every other feat defaults to once. Feat category and prerequisite
exceptions must come from the actual source grant. Magic Initiate alone is not the
Spellcasting feature prerequisite for Spell Recall.
Irresistible Offense ignores Resistance to the character's Bludgeoning, Piercing and
Slashing damage, including spell damage of those types; other damage Resistance
and all Immunity remain. Magic Initiate natively depends on #272 for completed-rest
recovery and #336 for level-gain replacement, as well as casting/selection owners.

ASI application is shared in the ASI child, including boon/Grappler allowed alternatives,
Constitution maximum-HP changes, derived modifiers and caps. Boon and Grappler children
consume it rather than inventing another ability updater. Feat ASI is separate from background
ability assignment and ability-setting items; those owners retain their semantics. Fighter
level4/6/8/12/14/16 choices, level19 boon choice, Fighter-level style replacement, Champion
additional style, other class grants, and optional post20 XP boons retain their source owners.
A level19 requirement alone never creates a boon grant. No data import or isolated fixture
flag can advance a production character to obtain one.

## Single owners and cross-domain prerequisites

| Owner | Exact retained responsibility | Feat consumers |
| --- | --- | --- |
| #321–#332 class splitters | Class/subclass level grants, named feature prerequisites and replacement triggers; #325 Fighter/Champion, #327 Paladin and #328 Ranger retain Fighting Style grants | All feat acquisition from those features |
| #333 | Background/species creation choices and proficiency provenance | Origin feat grants; Skilled choices consume shared proficiency semantics |
| #335 | Multiclass eligibility, shared slots, nonstacking AC/Extra Attack | No feat child grants class levels or slots |
| #336 | Advancement levels, HP/Constitution, higher-level starts and optional post20 feat grants | Feat selection consumes a real grant and ASI requests recomputation through that transition |
| #264 / #265 | Exact Light object fixture / recorded Detect Magic free-use fixture | Magic Initiate never reimplements their spell effects |
| #273 / #277 | Casting modes, slot/free-use routing and all remaining spell effect keys, respectively | Magic Initiate selection and Spell Recall wrap the same authoritative cast receipt |
| #267 / #278, #247–#252 | Equipment state, remaining weapon/armor/item variants | Defense, Archery, two-handed/Light weapon predicates; no invented item properties |
| #268 / #269 / #270 | Invisible/senses, incapacitating conditions, Grappled/movement and other bodily conditions | Night Spirit/Truesight, Alert eligibility, Grappler |
| #272 | Source-correct completed/interrupted rests and recharge | Magic Initiate, Fate |
| #145 | Authoritative light/visibility facts | Night Spirit, Truesight, Dimensional Travel |
| #351 / #353 | Cast-owned effects and concentration event endings | No feat creates a competing concentration/effect lifecycle |
| #338 / #339 | Existing Action Surge / Tactical Mind concrete fixtures | Neither implements a feat or justifies higher-level grants |
| #382 | Existing Cleric Shield of Faith recipient production admission | Example of the separate admission boundary; no new feat entitlement |

The new shared selection child owns category/prerequisite/repeatability validation and
atomic selection storage. Seventeen individual effect children own exact named feat behavior,
with positive and negative deterministic fixtures, shared ASI prerequisites where needed,
and explicit durability/disclosure/web/TUI acceptance. No additional research spike is needed:
the remaining decisions already have named class/spell/item/condition owners above.
Existing owner bodies are preserved. A splitter consuming this plan must publish its exact
terminal implementation children and attach them directly to #337 before claiming completion;
closing a source decision alone never passes the character gate.

The current native graph confirms #334 is a child of #283, blocked by #298–#301;
#337 is a child of #283 blocked by #334 and the other character splitters. This agrees with
#74 and #283's explicit source decision. Every new terminal directly blocks #337. Existing
concrete check/attack/initiative/target/visibility/Light/Detect Magic/equipment owners are
also directly linked to #337, preserving their native parent and bodies. Cross-domain
#279/#280/#183 remain independently required; no new dependency on a qualification gate
is introduced that could cycle back into #337.

## Verification and limits

Run from the checkout with `rtk proxy`:

```sh
python3 scripts/verify_feat_scope.py --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
python3 scripts/verify_feat_scope.py --live --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
python3 scripts/verify_srd_origin_feats.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
GOMAXPROCS=2 go test -p 2 -short -race ./...
GOMAXPROCS=2 go vet -p 2 ./...
git diff --check
```

The matrix verifier checks exact nested catalogue equality, all unique mechanic-key owners,
evidence file existence, 17 identities/four category counts, exact child body hashes and
native parents/prerequisites, all four #334 source edges, every terminal gate edge, and repository-wide dependency
acyclicity. The four existing source verifiers independently compare catalogue transcriptions
to the pinned PDF. These checks prove source/plan/tracker consistency, not playable children.

Decision-only diff: no runtime code, database/browser mutation, entitlement flags, grants,
coverage register changes, production deployment or paid inference. Future mutation children
must prove PostgreSQL rollback/replay/reload/concurrency and real client journeys in their
own PRs. Hosted checks and independent parent acceptance are separate from local checks.

## Published terminal map

| Issue | Exact owner | Native parent | Direct gate |
| --- | --- | --- | --- |
| #384 | selection | #283 | blocks #337 |
| #385 | alert | #283 | blocks #337 |
| #386 | magic-initiate | #283 | blocks #337 |
| #387 | savage-attacker | #283 | blocks #337 |
| #388 | skilled | #283 | blocks #337 |
| #389 | ability-score-improvement | #283 | blocks #337 |
| #390 | grappler | #283 | blocks #337 |
| #391 | archery | #283 | blocks #337 |
| #392 | defense | #283 | blocks #337 |
| #393 | great-weapon-fighting | #283 | blocks #337 |
| #394 | two-weapon-fighting | #283 | blocks #337 |
| #395 | boon-of-combat-prowess | #283 | blocks #337 |
| #396 | boon-of-dimensional-travel | #283 | blocks #337 |
| #397 | boon-of-fate | #283 | blocks #337 |
| #398 | boon-of-irresistible-offense | #283 | blocks #337 |
| #399 | boon-of-spell-recall | #283 | blocks #337 |
| #400 | boon-of-the-night-spirit | #283 | blocks #337 |
| #401 | boon-of-truesight | #283 | blocks #337 |

Publication readback confirmed each supplement retained its prior body byte for byte.
A subsequent concurrent edit changed #47's earlier text while retaining our exact section.
The live verifier permits external edits outside that section and records original/current
prefix hashes without overwriting them. Created child bodies still require exact equality.

Local verification for this decision: all four existing source comparators passed against
the pinned PDF (Fighting Style also rejected 255 contrary-source cases). Full Go short race
and vet passed with `GOMAXPROCS=2` and `-p 2`. Python compilation and diff whitespace
checks passed. New verifier corruption probes rejected changed source data, missing terminal
gate edge, changed child-body hash and a dependency cycle. The verifier also rejects
each missing #334 source edge and removal of all four, even when those source issues
remain closed. Live readback matched 18 exact
child bodies, native parents and prerequisites, Task types and Edda Project sizes (Magic
Initiate, Grappler and Night Spirit L; other children M), with 146 keys and 285 acyclic issue
nodes. No new child mechanics were implemented or qualified by these checks.

The evidence refresh checks the concurrent Shield of Faith projection change: a nonself
cast no longer adds its bonus to the caster's pure sheet projection. This changes no
feat grant, Defense prerequisite, or fixed-build feat behavior. Only the reviewed
`character.go` evidence hash and observation revision were refreshed; source records,
child scopes and the strict evidence comparison remain unchanged. The native receipt
was read back again after the roadmap checklist updates.
