# Barbarian decision for #321

Observed against main `32252f64380eb49dd47f7d1f53171b6a8791624e` on 2026-11-16.

The implementation observation includes the merged Metamagic-family validator
guard. Its exact diff does not change Barbarian eligibility or mechanics; the
evidence hash is refreshed while source records, owners and gate requirements
remain unchanged.
This decision publishes implementation work. Barbarian remains unqualified.
#321 closed through [merged PR #500](https://github.com/subculture-collective/edda/pull/500)
after independent acceptance and hosted checks. The implementation terminals and
qualification gate #498 remain open.

The [matrix](issue-321-barbarian-scope.json) retains all 80 exact class/subclass
records, all 20 progression rows, nested alternatives, prerequisites, numeric parameters,
required mechanics and opaque option keys. Each of 189 unique keys has exactly one
owner. Each row has a separate application owner. The [native receipt](issue-321-link-verification.json)
checks every published title/body, parent, blocker and Edda Project Task/Size value,
and the complete repository dependency and parent graphs.

## Source and current support

The supplied English SRD 5.2.1 PDF independently hashes to
`8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.
Direct audit covered Barbarian pp28–30, Berserker p30, creation/progression/HP/multiclass
pp19–26, mastery/weapon/armor pp89–92, conditions/action/recovery pp174–191.
These facts are adaptations under CC-BY-4.0; retain the complete source attribution in
`docs/srd-5.2.1-conformance.md` and the matrix source block. Existing source comparators
independently check class rows, source headings, conditions, numbers and advancement.
Closed live prerequisites #285/#286/#308 supply metadata rather than executable features.

`catalogue.go` deliberately returns false for every `qualifiedCharacterMechanic`.
`creation.go` supports fixed Human Soldier Fighter and Human Acolyte Cleric choices.
`character.go` admits only its bounded Fighter1/2 and Cleric1 projections; it rejects
Barbarian rather than constructing a sheet. `advancement.go` only advances Fighter1→2
with 300 XP. Generic world-builder class names and older rules are not SRD admission.
Frontend `api/types.ts` carries partial coverage, source-linked capabilities and unavailable
sheet reasons; it supplies no Barbarian choices. `tui/character/character.go` is a placeholder,
so current web/TUI parity cannot be claimed. Every child requires its own shared action
path, availability/confirmation/reconnect and persistence fixtures. Evidence file hashes are
strictly checked; catalogue equality and passing tests cannot enable the class.

All 20 named Barbarian features, four Berserker features and the subclass identity remain
unimplemented. The matrix preserves the two separate Improved Brutal Strike identities,
all nested equipment/skill/mastery/subclass/feat/blow choices and conditional predicates.
No executable Barbarian fixture was found; supported generic attack/check/action/sheet
infrastructure is evidence for reuse, never evidence for an admitted Barbarian.

## Ownership and bounded applications

71 one-PR Task M issues separate behavior from application. Large Rage behavior is split
into entry/extension/ending, modifiers, casting restrictions and recovery. Brutal Strike
has a core attack choice, Forceful, Hamstring, expanded domain, Staggering, Sundering and
level17 upgrade. Intimidating Presence separates effect use from resource restoration.
Shared mastery has one terminal per property rather than an eight-effect omnibus PR.

First-class traits include STR/CON saves, Simple/Martial weapons, Light/Medium armor/Shields,
two of six named skills and either exact equipment bundle. Creation HP is 12+CON.
Later-class entry has D12 HP, Martial weapons and Shields plus level1 features; it grants
no first-class saves, skills, armor or equipment. #431/#497 require STR13 and every other
class prerequisite. The shared transaction remains distinct from that admission application.

Each level application consumes its named feature resolvers and applies only the exact
row once. Higher-level starts must compose individually supported transitions; neither
imported rows nor earlier levels grant later features. Shared advancement owns XP, D12/fixed7
HP minimum1, total-level PB, Hit Dice and CON-derived recomputation. Feature owners never
advance a character. Row application does not duplicate a resolver or refill spent resources.

Rage grants no spells/prepared list/slots: its casting restriction consumes the existing
cast boundary and Concentration lifecycle. Primal Knowledge's permanent skill selection
and optional per-check STR substitution are separate. Feat grants at4/8/12/16 and19 consume
#384 selection and #389 source-specific ability updates. Irresistible Offense is a recommendation,
not an automatic boon. Specific chosen feat effects must be independently supported and
required directly by #337 before admission. Primal Champion uses cap25, not ordinary ASI20
or boon30, and must update CON-derived HP once across all attained levels.

Shared source owners retain their scopes: #272 rests; #273 casting modes (live title confirms
this is not a sheet issue); #269/#270 conditions; #278 equipment/mastery; #333 origin provenance;
#335 multiclass; #336 advancement; #384 selection/#389 ability updates; #351/#353 effect lifecycle.
#247/#248/#249 are weapon catalogues, not property implementations. #267's chain-mail/Shield
fixture does not establish arbitrary armor readiness. Unsupported weapon/armor variants
remain unavailable until their exact item-owner terminal implementation is attached to #337;
this decision does not widen their scope or promote a partial selection to full class support.

Shared terminal children #427/#428/#429/#430/#431/#487/#488/#489–#496 attach under those
existing owners. Barbarian-specific consumers stay under #283, matching #334's hierarchy.
#429 is a reusable XP/PB/CON transaction with a D12 adapter; another class must consume
its shared contract with its own source dice rather than duplicate that transaction.
#430 stores generic proficiency provenance despite its Barbarian consumer title.
Every new terminal, including shared children and #497 multiclass entry, natively blocks
Barbarian qualification #498; #498 directly blocks #337 and also requires decision #321. Reused concrete check/attack/action/movement/equipment/effect/selection owners also
directly block it. Planning #272/#273/#333/#335/#336 remain prerequisites; their closure
cannot substitute for required consumed terminal implementations.

## Exact class rows

PB here is the printed single-class row; multiclass uses total character-level PB.
Rage damage is a bonus on eligible STR damage, not a second resource.

| Level | PB | Rage uses | Rage damage | Masteries | New feature grants | Application |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | +2 | 2 | +2 | 2 | Rage, Unarmored Defense, Weapon Mastery | #467 |
| 2 | +2 | 2 | +2 | 2 | Danger Sense, Reckless Attack | #468 |
| 3 | +2 | 3 | +2 | 2 | Barbarian Subclass, Primal Knowledge | #469 |
| 4 | +2 | 3 | +2 | 3 | Ability Score Improvement | #470 |
| 5 | +3 | 3 | +2 | 3 | Extra Attack, Fast Movement | #471 |
| 6 | +3 | 4 | +2 | 3 | Berserker Mindless Rage | #472 |
| 7 | +3 | 4 | +2 | 3 | Feral Instinct, Instinctive Pounce | #473 |
| 8 | +3 | 4 | +2 | 3 | Ability Score Improvement | #474 |
| 9 | +4 | 4 | +3 | 3 | Brutal Strike | #475 |
| 10 | +4 | 4 | +3 | 4 | Berserker Retaliation | #476 |
| 11 | +4 | 4 | +3 | 4 | Relentless Rage | #477 |
| 12 | +4 | 5 | +3 | 4 | Ability Score Improvement | #478 |
| 13 | +5 | 5 | +3 | 4 | Improved Brutal Strike | #479 |
| 14 | +5 | 5 | +3 | 4 | Berserker Intimidating Presence | #480 |
| 15 | +5 | 5 | +3 | 4 | Persistent Rage | #481 |
| 16 | +5 | 5 | +4 | 4 | Ability Score Improvement | #482 |
| 17 | +6 | 6 | +4 | 4 | Improved Brutal Strike | #483 |
| 18 | +6 | 6 | +4 | 4 | Indomitable Might | #484 |
| 19 | +6 | 6 | +4 | 4 | Epic Boon | #485 |
| 20 | +6 | 6 | +4 | 4 | Primal Champion | #486 |

## Published terminal map

Each linked child contains exact owned keys, source-record context including nested values,
positive/negative eligibility fixtures and the complete durability/client/privacy acceptance.
Shared children with no class catalogue key have an explicit referenced-rule scope instead
of inventing a second mechanic owner.

| Issue | Scope owner | Native parent | Gate |
| --- | --- | --- | --- |
| [#427](https://github.com/subculture-collective/edda/issues/427) | Complete source-correct Short Rest and publish recharge events | #272 | blocks #498 → #337 |
| [#428](https://github.com/subculture-collective/edda/issues/428) | Complete source-correct Long Rest with interruption and recovery receipt | #272 | blocks #498 → #337 |
| [#429](https://github.com/subculture-collective/edda/issues/429) | Confirm one source-linked D12 level and derived-stat transition | #336 | blocks #498 → #337 |
| [#430](https://github.com/subculture-collective/edda/issues/430) | Confirm creation proficiency provenance for Barbarian consumers | #333 | blocks #498 → #337 |
| [#431](https://github.com/subculture-collective/edda/issues/431) | Validate Barbarian multiclass entry and nonstacking feature admission | #335 | blocks #498 → #337 |
| [#432](https://github.com/subculture-collective/edda/issues/432) | Confirm Barbarian first-class traits and two skill choices | #283 | blocks #498 → #337 |
| [#433](https://github.com/subculture-collective/edda/issues/433) | Confirm either Barbarian starting equipment package | #283 | blocks #498 → #337 |
| [#434](https://github.com/subculture-collective/edda/issues/434) | Confirm Rage entry extension and source ending rules | #283 | blocks #498 → #337 |
| [#435](https://github.com/subculture-collective/edda/issues/435) | Apply Rage resistance Strength damage and advantage | #283 | blocks #498 → #337 |
| [#436](https://github.com/subculture-collective/edda/issues/436) | Enforce Rage casting and Concentration restrictions | #283 | blocks #498 → #337 |
| [#437](https://github.com/subculture-collective/edda/issues/437) | Recover exact Barbarian Rage resources on completed rests | #283 | blocks #498 → #337 |
| [#438](https://github.com/subculture-collective/edda/issues/438) | Project Barbarian Unarmored Defense with shield eligibility | #283 | blocks #498 → #337 |
| [#439](https://github.com/subculture-collective/edda/issues/439) | Confirm Barbarian melee mastery choices and rest replacement | #283 | blocks #498 → #337 |
| [#440](https://github.com/subculture-collective/edda/issues/440) | Resolve Danger Sense Dexterity saves | #283 | blocks #498 → #337 |
| [#441](https://github.com/subculture-collective/edda/issues/441) | Confirm Reckless Attack on first own-turn attack | #283 | blocks #498 → #337 |
| [#442](https://github.com/subculture-collective/edda/issues/442) | Confirm level3 Path of the Berserker selection | #283 | blocks #498 → #337 |
| [#443](https://github.com/subculture-collective/edda/issues/443) | Confirm Primal Knowledge additional Barbarian skill | #283 | blocks #498 → #337 |
| [#444](https://github.com/subculture-collective/edda/issues/444) | Confirm Primal Knowledge Strength substitution per check | #283 | blocks #498 → #337 |
| [#445](https://github.com/subculture-collective/edda/issues/445) | Offer qualified Barbarian feat grants at levels4 8 12 16 | #283 | blocks #498 → #337 |
| [#446](https://github.com/subculture-collective/edda/issues/446) | Resolve Barbarian two attacks per own-turn Attack action | #283 | blocks #498 → #337 |
| [#447](https://github.com/subculture-collective/edda/issues/447) | Project Barbarian Fast Movement without Heavy armor | #283 | blocks #498 → #337 |
| [#448](https://github.com/subculture-collective/edda/issues/448) | Resolve Feral Instinct initiative advantage | #283 | blocks #498 → #337 |
| [#449](https://github.com/subculture-collective/edda/issues/449) | Confirm Instinctive Pounce during Rage entry | #283 | blocks #498 → #337 |
| [#450](https://github.com/subculture-collective/edda/issues/450) | Confirm Brutal Strike attack damage and one effect selection | #283 | blocks #498 → #337 |
| [#451](https://github.com/subculture-collective/edda/issues/451) | Resolve Brutal Strike Forceful Blow and optional follow | #283 | blocks #498 → #337 |
| [#452](https://github.com/subculture-collective/edda/issues/452) | Resolve Brutal Strike Hamstring Blow expiry and replacement | #283 | blocks #498 → #337 |
| [#453](https://github.com/subculture-collective/edda/issues/453) | Confirm Relentless Rage zero HP replacement and DC recovery | #283 | blocks #498 → #337 |
| [#454](https://github.com/subculture-collective/edda/issues/454) | Expose level13 expanded Brutal Strike effect domain | #283 | blocks #498 → #337 |
| [#455](https://github.com/subculture-collective/edda/issues/455) | Resolve Staggering Blow next save and opportunity suppression | #283 | blocks #498 → #337 |
| [#456](https://github.com/subculture-collective/edda/issues/456) | Resolve Sundering Blow next other creature attack bonus | #283 | blocks #498 → #337 |
| [#457](https://github.com/subculture-collective/edda/issues/457) | Confirm Persistent Rage initiative recovery and revised duration | #283 | blocks #498 → #337 |
| [#458](https://github.com/subculture-collective/edda/issues/458) | Resolve level17 Brutal Strike two different effects and 2d10 | #283 | blocks #498 → #337 |
| [#459](https://github.com/subculture-collective/edda/issues/459) | Confirm Indomitable Might Strength total replacement | #283 | blocks #498 → #337 |
| [#460](https://github.com/subculture-collective/edda/issues/460) | Offer qualified Barbarian level19 Epic Boon or other feat | #283 | blocks #498 → #337 |
| [#461](https://github.com/subculture-collective/edda/issues/461) | Apply Primal Champion Strength Constitution increases once | #283 | blocks #498 → #337 |
| [#462](https://github.com/subculture-collective/edda/issues/462) | Resolve Berserker Frenzy first Strength hit damage | #283 | blocks #498 → #337 |
| [#463](https://github.com/subculture-collective/edda/issues/463) | Apply Mindless Rage Charmed Frightened immunity and ending | #283 | blocks #498 → #337 |
| [#464](https://github.com/subculture-collective/edda/issues/464) | Confirm Berserker Retaliation damage-triggered melee reaction | #283 | blocks #498 → #337 |
| [#465](https://github.com/subculture-collective/edda/issues/465) | Confirm Intimidating Presence chosen emanation saves | #283 | blocks #498 → #337 |
| [#466](https://github.com/subculture-collective/edda/issues/466) | Confirm Intimidating Presence Rage-cost restoration and rest reset | #283 | blocks #498 → #337 |
| [#467](https://github.com/subculture-collective/edda/issues/467) | Confirm Barbarian level1 creation | #283 | blocks #498 → #337 |
| [#468](https://github.com/subculture-collective/edda/issues/468) | Confirm Barbarian level1 to2 transition | #283 | blocks #498 → #337 |
| [#469](https://github.com/subculture-collective/edda/issues/469) | Confirm Barbarian level2 to3 transition | #283 | blocks #498 → #337 |
| [#470](https://github.com/subculture-collective/edda/issues/470) | Confirm Barbarian level3 to4 transition | #283 | blocks #498 → #337 |
| [#471](https://github.com/subculture-collective/edda/issues/471) | Confirm Barbarian level4 to5 transition | #283 | blocks #498 → #337 |
| [#472](https://github.com/subculture-collective/edda/issues/472) | Confirm Barbarian level5 to6 transition | #283 | blocks #498 → #337 |
| [#473](https://github.com/subculture-collective/edda/issues/473) | Confirm Barbarian level6 to7 transition | #283 | blocks #498 → #337 |
| [#474](https://github.com/subculture-collective/edda/issues/474) | Confirm Barbarian level7 to8 transition | #283 | blocks #498 → #337 |
| [#475](https://github.com/subculture-collective/edda/issues/475) | Confirm Barbarian level8 to9 transition | #283 | blocks #498 → #337 |
| [#476](https://github.com/subculture-collective/edda/issues/476) | Confirm Barbarian level9 to10 transition | #283 | blocks #498 → #337 |
| [#477](https://github.com/subculture-collective/edda/issues/477) | Confirm Barbarian level10 to11 transition | #283 | blocks #498 → #337 |
| [#478](https://github.com/subculture-collective/edda/issues/478) | Confirm Barbarian level11 to12 transition | #283 | blocks #498 → #337 |
| [#479](https://github.com/subculture-collective/edda/issues/479) | Confirm Barbarian level12 to13 transition | #283 | blocks #498 → #337 |
| [#480](https://github.com/subculture-collective/edda/issues/480) | Confirm Barbarian level13 to14 transition | #283 | blocks #498 → #337 |
| [#481](https://github.com/subculture-collective/edda/issues/481) | Confirm Barbarian level14 to15 transition | #283 | blocks #498 → #337 |
| [#482](https://github.com/subculture-collective/edda/issues/482) | Confirm Barbarian level15 to16 transition | #283 | blocks #498 → #337 |
| [#483](https://github.com/subculture-collective/edda/issues/483) | Confirm Barbarian level16 to17 transition | #283 | blocks #498 → #337 |
| [#484](https://github.com/subculture-collective/edda/issues/484) | Confirm Barbarian level17 to18 transition | #283 | blocks #498 → #337 |
| [#485](https://github.com/subculture-collective/edda/issues/485) | Confirm Barbarian level18 to19 transition | #283 | blocks #498 → #337 |
| [#486](https://github.com/subculture-collective/edda/issues/486) | Confirm Barbarian level19 to20 transition | #283 | blocks #498 → #337 |
| [#487](https://github.com/subculture-collective/edda/issues/487) | Resolve shared Incapacitated and Unconscious condition admission | #269 | blocks #498 → #337 |
| [#488](https://github.com/subculture-collective/edda/issues/488) | Resolve shared Charmed Frightened and Prone condition consumers | #270 | blocks #498 → #337 |
| [#489](https://github.com/subculture-collective/edda/issues/489) | Resolve shared weapon mastery Cleave with one attack receipt | #278 | blocks #498 → #337 |
| [#490](https://github.com/subculture-collective/edda/issues/490) | Resolve shared weapon mastery Graze with one attack receipt | #278 | blocks #498 → #337 |
| [#491](https://github.com/subculture-collective/edda/issues/491) | Resolve shared weapon mastery Nick with one attack receipt | #278 | blocks #498 → #337 |
| [#492](https://github.com/subculture-collective/edda/issues/492) | Resolve shared weapon mastery Push with one attack receipt | #278 | blocks #498 → #337 |
| [#493](https://github.com/subculture-collective/edda/issues/493) | Resolve shared weapon mastery Sap with one attack receipt | #278 | blocks #498 → #337 |
| [#494](https://github.com/subculture-collective/edda/issues/494) | Resolve shared weapon mastery Slow with one attack receipt | #278 | blocks #498 → #337 |
| [#495](https://github.com/subculture-collective/edda/issues/495) | Resolve shared weapon mastery Topple with one attack receipt | #278 | blocks #498 → #337 |
| [#496](https://github.com/subculture-collective/edda/issues/496) | Resolve shared weapon mastery Vex with one attack receipt | #278 | blocks #498 → #337 |
| [#497](https://github.com/subculture-collective/edda/issues/497) | Confirm later-class Barbarian level1 admission | #283 | blocks #498 → #337 |

## Reproduction and limits

```sh
rtk proxy python3 scripts/verify_barbarian_scope.py --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy python3 scripts/verify_barbarian_scope.py --live --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy python3 scripts/verify_srd_barbarian.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy python3 scripts/verify_srd_berserker.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy python3 scripts/verify_srd_advancement.py /tmp/edda-wave1/SRD_CC_v5.2.1.pdf
rtk proxy python3 scripts/test_verify_scope_pagination.py
rtk proxy env GOMAXPROCS=2 go test -p 2 -short -race ./...
rtk proxy env GOMAXPROCS=2 go vet -p 2 ./...
rtk git diff --check
```

The live verifier only reads GitHub and updates its local receipt after every guard passes.
Its GraphQL node inventory includes open and closed issues. Blocker pagination reads every
page, rejecting repeated pages instead of truncating a reported larger edge set; parent and dependency graphs are checked
separately for cycles. Existing effect/feat verification retains its exact edge guards. Its graph reader retains
complete pagination and rejects repeated pages; it does not relax any #334 blocker.

Corruption probes reject wrong source, missing child, missing key, each prerequisite source
edge, each of71 required terminal-to-#498 edges, changed body hash and a dependency cycle.
Exact source snapshots retain all conditions/numbers; source comparators remain separate.
Corruption probes also reject removal of #498→#337 or the #321 planning prerequisite.
Owned #47/#283/#337 sections are append-only: fresh body read before publication, exact
prefix-plus-section readback, then owned-section checks tolerate later external prefix edits.
No #54/#282 source checklists or #304/#305/#320 closure were changed by this decision.

These checks establish source/plan/native-graph consistency, not implemented gameplay.
No runtime code, grants, coverage, database changes, production, deployment, inference,
client journeys or child implementations are supplied. Child mutation acceptance requires
isolated PostgreSQL rollback/replay/reload/concurrency and web/TUI journeys in its own PR.
Parent acceptance, hosted CI and merge remain separate from local validation. The known
linter shim is unavailable; host limits and toolchain remain unchanged.

The pagination regression probe parses query parameters and caps API calls at three. It
checks retention of blocker101 and rejection of repeated pages; repeated-page guards also
protect both live graph readers. A faulty earlier temporary probe was cancelled by the
T3 restart. Recovery confirmed the old probe/publisher absent and resumed the saved issue
IDs without recreating issues. Temporary files were restored from committed-source inputs.

## Native dependency capacity and migration

GitHub rejected a direct #337 blocker with HTTP422: “Source issue cannot have more than
100 blocked-by relations.” Parent explicitly authorized a per-class native qualification
gate, preserving every terminal and every historical #337 edge. This changes the #321
publication graph only; #334's direct feat blockers and strict verifier guards stay intact.

[Qualification #498](https://github.com/subculture-collective/edda/issues/498), Task M under
#283, requires all71 exact implementation terminals and #321. It directly blocks #337.
Its own qualification PR must prove source-linked behavior, PostgreSQL durability and
web/TUI client journeys after the implementations; creating this issue qualifies nothing.

The [migration receipt](issue-321-gate-migration.json) records the100-edge observation,
protected legacy dependencies and exact owned-edge removals. Before touching #337,
all71 terminals and #321 were attached to #498. Both #321 and #337 were read OPEN and
#337 still required #321. One newly-added #321 terminal edge was removed, #498 was
immediately added to #337, and every terminal path plus every protected legacy edge was
read back before removing the remaining #321-owned direct edges. No historical edge,
#334 prerequisite or reused terminal was removed. The verifier checks the exact terminal
set on #498, #498→#337, retained planning/source prerequisites, legacy edge preservation,
all body/Project fields, and both complete native graph cycle proofs.

Owned capacity explanations on #321/#337/#47/#283 are append-only. Source checklist
updates by other jobs are retained. Gate closure remains blocked by every implementation
and its independent qualification evidence; decision PR merge alone never enables it.

## Local verification result

Both the original run and the recovery run of the full Go short race suite passed with
`GOMAXPROCS=2` and `-p 2`; vet passed under the same limits. Barbarian, Berserker and
advancement PDF comparators passed after restoring the exact pinned PDF. Existing #334
verification and all four feat source comparators passed. Python compilation, bounded
pagination regression probes and diff whitespace checks passed. The final offline/live
verifier result is recorded with the exact graph/body/Project evidence in the native receipt.

The shared terminal reservations were sent to the parent for Bard reuse. Bard retains
its own D8/fixed5 adapter, source-specific casting/list/slot work and spell effects;
this decision owns no duplicate implementation of those rules. Parent received the #498
migration result and actual54 freed root slots before Bard gate publication proceeds.

Final gate validation also required the preexisting #351 effect and #353 Concentration
implementations as direct #337 blockers. Recovery attached both after the class migration;
these consumed two of the54 free slots. This added requirements without removing any
protected dependency. Subsequent class-gate additions can further change available slots.

Final strict live and offline verification passed: 80 exact source records, 20 rows,
189 unique keys, 71 implementation bodies plus qualification #498, all native parents/
prerequisites/Task M fields and 406 acyclic issue nodes. #498 has72 exact blockers
(71 terminals plus #321); #337 has 49 blockers in this receipt.
The unchanged #334 validator also passed against this final native graph, including
its independent missing-source-edge probes.
