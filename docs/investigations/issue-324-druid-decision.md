# Druid and Circle of the Land fixture decision (#324)

Druid remains unavailable. This decision publishes bounded implementation contracts; it adds no runtime behavior, grants, source catalogue changes, coverage claim or Full Rules promotion.

The audit is against base `9dd5cac8b90eee89dfaa0b3555e8fc107c18601a` in the T3-prepared `feat/druid-playable-fixtures` checkout. GitHub `origin` is authoritative; the Gitea mirror has pushes disabled. The pinned local English SRD 5.2.1 PDF has SHA-256 `8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87`.

The canonical Druid catalogue has 76 records; Circle of the Land has 27. Each identity, field, prerequisite operand, nested option, source reference and numeric value is retained in [the exact matrix](issue-324-druid-scope.json). All 20 Druid class rows are paired with the global advancement row; multiclass metadata is separately pinned. Counts are cross-checks, never substitutes for identity/value equality. Direct source comparison covers PDF 41–43, 46, advancement23 and multiclass24–25; additional clause guards cover casting105–106, rests185/187 and conditions184/186–187. The 25 consumed spell identities are compared directly with full raw PDF107–175 source operands; Fireball extraction explicitly excludes the adjacent Otherworldly Steed block.

## Executable evidence and admission

`internal/rules/srd521/creation.go:SupportedBuildForConcept` admits the fixed Human Soldier Fighter and Human Acolyte Cleric fixtures and rejects Druid. `advancement.go:AdvanceFighter` implements Fighter1→2 only. The Druid and Land source tests validate catalogues and composed references; they do not execute Wild Shape, spellcasting, subclass grants or level transitions. No Druid action handler was found in `internal/game`, `internal/engine` or `internal/combat`. Typed character projection, authoritative action receipts, mitigation/concentration and web/TUI callers provide reusable seams. They provide no Druid entitlement.

The matrix gives every record its engine, web, TUI and qualification evidence and hashes the inspected files. Every new behavior terminal requires an explicitly isolated entitlement for higher-level fixtures. Every entry or level application consumes implemented behavior and applies grants once. The 20 applications and later-class Druid1 are separate contracts. Production admission requires independent qualification, source-linked action fixtures, PostgreSQL rollback/retry/replay/reload/concurrency and desktop/phone/TUI availability/disclosure proof. None of that new feature execution proof is supplied by this planning PR.

## Ownership and source decisions

Druid D8/fixed5 consumes the reusable XP/PB/CON transaction #429, not Barbarian D12/fixed7 or its eligibility. First-class traits/equipment differ from later-class entry; latter requires WIS13 and every current primary threshold, grants no new starting skills/saves/tools/weapons/equipment, and uses later-level HP. Preparation is per-class; combined slots never broaden the prepared spell level predicate. #58 owns sheet projection; #273 owns casting modes. #502 is Bard instrument routing, #556 Cleric Holy Symbol routing; Druid Focus has its own source-specific adapter. #503 validates four-class list membership, independently of any spell effect.

Druidic messages expose presence versus deciphered content separately. Magician adds one Druid cantrip and max(1,WISmod) to INT Arcana/Nature only. Warden grants Martial weapons/Medium armor; neither grants weapon mastery. No #489–496 mastery leaf is consumed without a source grant. Improvements at4/8/12/16 and level19 feats require #384 selection, #389 caps and all exact eligible #334 effect terminals, with original prerequisites/repeatability/source provenance retained. No blanket feat availability follows from a source choice.

Wild Shape uses a single canonical resource shared with Companion, Resurgence and Land effects. Known forms retain Beast type, exact CR/Fly/level bounds, versioned stat provenance and GM permission for external sources. Recommendations grant no entitlement. Retained Druid statistics differ from replaced Beast statistics; retained skills/saves use own PB or higher Beast modifier. Speech is retained but casting is prohibited until Beast Spells; transformation alone retains prior Concentration, while Incapacitated still breaks it. Equipment does not resize; worn practicality is an explicit GM fact, merged items have no effect, limbs constrain handling. Ending restores source-owned state once without healing. Wild Companion consumes #181/#184/#185 ownership/lifecycle and #183 qualification; a planning owner does not establish executable familiar support.

Land choice occurs on completed Long Rest and is not inferred from map terrain. Exactly one current table applies; its3/5/7/9 spells are conditional and count as Druid spells without consuming preparation count. All24 named table spell effects have exact separate spell-owned contracts. Source-list Aid #614 is separately consumed; it is not Land's Aid. Land's Aid is one Magic action/use and one atomic chosen-target damage/healing batch, scaling2d6/3d6/4d6 at3/10/14. Natural Recovery has independent free-cast and expended-slot recovery locks; Short Rest recovery uses ceil(level/2) total slot levels, each below6. Ward's Poisoned immunity is distinct from current-land damage resistance. Sanctuary grants Half Cover to self/allies but only the current Ward resistance to allies; it never grants their Poisoned immunity. Cube placement, move, membership and endings have separate explicit contracts.

Shared #487/#488/#504 retain their exact scopes. New #727 admits only Paralyzed/Poisoned/Restrained projections and consumes #487 for Incapacitated; #504 retains trusted condition ending. Wall of Stone, Polymorph and Web are split by bounded source behavior. Polymorph does not borrow Wild Shape's retained mental statistics, class features or speech. World geometry, tree knowledge, Beast facts, object burning and authorized collapse consume authoritative world facts, never model inference or hidden state disclosures.

## Native qualification and publication

Top gate #630 has native parent #324, preserving #283's child capacity. Feature gate #631 contains the exact 64 Druid feature/entry/level leaves as prerequisites; shared gate #632 contains 31 new shared leaves and 46 reused leaves. Every one of 141 mandatory leaves has the native path `terminal → subordinate gate → #630 → #337`. #630 additionally requires decision324 and existing qualification183/279/280. No new terminal directly blocks337. New Druid-specific terminals retain native parent630; shared spells parent191 and condition admission parent270. Gate containment and dependency graphs are validated separately: a parent gate blocked by its contained children is valid.

The saved full graph preserves every observed historical337 prerequisite (including334, all 18 feat children and reused direct terminals), native parents, existing bodies and source prerequisites. Existing Barbarian498/Bard501 contracts stay intact. Native parents, exact bodies/titles/prerequisites, Project Task/Size, source keys/options, evidence hashes and gate paths are read back in [the receipt](issue-324-link-verification.json). The verifier is read-only and bounds API calls/pages, parses exact query parameters and rejects repeated pages/cursors. Offline corruption probes remove every saved new edge individually and also corrupt source pins/records/operands, spell text, body, metadata and both cycle types.

The parent has serially published the four exact supplements for47/54/283/337, preserving every freshly read prior prefix. The refreshed receipt seals each complete published body, and strict `--require-supplements` acceptance passes. Future pure appends must retain the sealed prefixes. The #47 supplement is below 1,000 characters and every combined body is checked against 65,536. Planning closure never changes class availability.

## Verification and limits

- Passed once: `GOMAXPROCS=2 go test -p 2 -short -race ./...` with default timeouts/assertions; SRD package 221.261s, all packages pass.
- Passed: `go vet -p 2 ./...`.
- Source comparators: Druid exact fields and 4 contrary controls; Land 27 complete records and 185 contrary controls; advancement 20 exact identities/XP/PB; multiclass 12 classes / 180 PDF cells; additional spell/rest/casting/condition clauses.
- Strict publication and corruption proof: `python3 scripts/verify_druid_scope.py --pdf /tmp/edda-wave1/SRD_CC_v5.2.1.pdf --probes`; live refresh adds `--live`. Final acceptance adds `--require-supplements`.
- Existing feat guards remain unchanged. No frontend/runtime changes; no new feature execution, device journey, production admission, deployment or Full Rules evidence is claimed.
- `golangci-lint version` cannot run: the active Go1.27.3 mise shim has no lint binary. No toolchain install or host limit change was made.

## Exact terminal index

| Issue | Native parent | Delivery contract |
|---|---|---|
| #633 | #630 | Apply Druid D8 or fixed-five through shared advancement |
| #634 | #630 | Confirm first-class Druid traits and two skills |
| #635 | #630 | Confirm either Druid starting equipment package |
| #636 | #630 | Validate Druid Wisdom and reduced later-class traits |
| #637 | #630 | Route Druid Wisdom and Druidic Focus casts through receipts |
| #638 | #630 | Confirm Druid cantrips and prepared spell changes |
| #639 | #630 | Project Druid nine slot columns and completed-rest recovery |
| #640 | #630 | Resolve Druidic language and concealed messages |
| #641 | #630 | Confirm Magician or Warden at Druid entry |
| #642 | #630 | Resolve Magician extra cantrip and Intelligence bonus |
| #643 | #630 | Track Wild Shape capacity and completed-rest recovery |
| #644 | #630 | Confirm source-linked known Beast forms and replacement |
| #645 | #630 | Confirm Wild Shape Bonus Action and temporary HP |
| #646 | #630 | Project retained Druid and replaced Beast statistics |
| #647 | #630 | Confirm shape equipment and limb restrictions |
| #648 | #630 | Enforce shaped casting prohibition without dropping prior spells |
| #649 | #630 | End Wild Shape on reuse early exit incapacity death or timeout |
| #650 | #630 | Cast Wild Companion with either canonical resource |
| #651 | #630 | Confirm Circle of the Land at Druid three |
| #652 | #630 | Offer qualified Druid feats at four eight twelve sixteen |
| #653 | #630 | Convert one slot to one Wild Shape use at zero |
| #654 | #630 | Convert one Wild Shape use to level-one slot |
| #655 | #630 | Confirm exactly one Elemental Fury option |
| #656 | #630 | Add Wisdom to Druid cantrip damage only |
| #657 | #630 | Confirm Primal Strike once on own-turn hit |
| #658 | #630 | Increase eligible chosen Potent cantrip range |
| #659 | #630 | Scale selected Primal Strike to two dice |
| #660 | #630 | Allow Beast Spells except costly or consumed materials |
| #661 | #630 | Offer Druid nineteen Epic Boon or other qualified feat |
| #662 | #630 | Recover one Wild Shape use on zero-use Initiative |
| #663 | #630 | Convert unexpended Wild Shape uses to one slot |
| #664 | #630 | Project Archdruid ten-to-one bodily aging |
| #665 | #630 | Select current land on completed Long Rest |
| #666 | #630 | Apply exact arid Circle spell table |
| #667 | #630 | Apply exact polar Circle spell table |
| #668 | #630 | Apply exact temperate Circle spell table |
| #669 | #630 | Apply exact tropical Circle spell table |
| #670 | #630 | Resolve Lands Aid chosen sphere damage and healing |
| #671 | #630 | Cast one prepared Circle spell without slot |
| #672 | #630 | Recover bounded expended slots on completed Short Rest |
| #673 | #630 | Apply Natures Ward immunity and current-land resistance |
| #674 | #630 | Create bounded Natures Sanctuary ground cube |
| #675 | #630 | Move and expire Natures Sanctuary with current membership |
| #676 | #191 | Resolve Blur exact source effect |
| #677 | #191 | Resolve Burning Hands exact source effect |
| #678 | #191 | Resolve Fire Bolt exact source effect |
| #679 | #191 | Resolve Fireball exact source effect |
| #680 | #191 | Resolve Blight exact source effect |
| #681 | #191 | Resolve Wall Of Stone exact source effect |
| #682 | #191 | Resolve Fog Cloud exact source effect |
| #683 | #191 | Resolve Hold Person exact source effect |
| #684 | #191 | Resolve Ray Of Frost exact source effect |
| #685 | #191 | Resolve Sleet Storm exact source effect |
| #686 | #191 | Resolve Ice Storm exact source effect |
| #687 | #191 | Resolve Cone Of Cold exact source effect |
| #688 | #191 | Resolve Misty Step exact source effect |
| #689 | #191 | Resolve Shocking Grasp exact source effect |
| #690 | #191 | Resolve Sleep exact source effect |
| #691 | #191 | Resolve Lightning Bolt exact source effect |
| #692 | #191 | Resolve Freedom Of Movement exact source effect |
| #693 | #191 | Resolve Tree Stride exact source effect |
| #694 | #191 | Resolve Acid Splash exact source effect |
| #695 | #191 | Resolve Ray Of Sickness exact source effect |
| #696 | #191 | Resolve Web exact source effect |
| #697 | #191 | Resolve Stinking Cloud exact source effect |
| #698 | #191 | Resolve Polymorph exact source effect |
| #699 | #191 | Resolve Insect Plague exact source effect |
| #700 | #191 | Resolve Speak With Animals exact source effect |
| #701 | #630 | Confirm Druid first-character level one |
| #702 | #630 | Confirm Druid level1 to2 transition |
| #703 | #630 | Confirm Druid level2 to3 transition |
| #704 | #630 | Confirm Druid level3 to4 transition |
| #705 | #630 | Confirm Druid level4 to5 transition |
| #706 | #630 | Confirm Druid level5 to6 transition |
| #707 | #630 | Confirm Druid level6 to7 transition |
| #708 | #630 | Confirm Druid level7 to8 transition |
| #709 | #630 | Confirm Druid level8 to9 transition |
| #710 | #630 | Confirm Druid level9 to10 transition |
| #711 | #630 | Confirm Druid level10 to11 transition |
| #712 | #630 | Confirm Druid level11 to12 transition |
| #713 | #630 | Confirm Druid level12 to13 transition |
| #714 | #630 | Confirm Druid level13 to14 transition |
| #715 | #630 | Confirm Druid level14 to15 transition |
| #716 | #630 | Confirm Druid level15 to16 transition |
| #717 | #630 | Confirm Druid level16 to17 transition |
| #718 | #630 | Confirm Druid level17 to18 transition |
| #719 | #630 | Confirm Druid level18 to19 transition |
| #720 | #630 | Confirm Druid level19 to20 transition |
| #721 | #630 | Confirm later-class Druid first-level application |
| #722 | #191 | Resolve Wall of Stone panel damage and authorized collapse |
| #723 | #191 | Resolve Wall of Stone concentration permanence |
| #724 | #191 | Project Polymorph Beast stats anatomy and melded gear |
| #725 | #191 | Resolve Polymorph temporary HP and exact ending |
| #726 | #191 | Resolve Web escape burning and source contribution ending |
| #727 | #270 | Admit exact Paralyzed Poisoned and Restrained condition projections |

## Record reconciliation

Each identity below is metadata only and Druid remains unavailable. Complete source values and evidence are in the JSON matrix; owners below refer to exact issues above.

| Source record identity | Behavior owners | Application |
|---|---|---|
| `srd-5.2.1:character/class/druid` | #633, #636, #634 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/skills` | #634 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/starting-equipment` | #635 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/1` | #651 | level-1 |
| `srd-5.2.1:character/feature/druid/spellcasting` | #637, #638 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/druidic` | #640 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/primal-order` | #642, #641 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/2` | #651 | level-2 |
| `srd-5.2.1:character/feature/druid/wild-shape` | #644, #648, #649, #645, #647, #643, #646 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/wild-companion` | #650 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/3` | #651 | level-3 |
| `srd-5.2.1:character/feature/druid/druid-subclass` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/4` | #651 | level-4 |
| `srd-5.2.1:character/feature/druid/ability-score-improvement` | #652 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/5` | #651 | level-5 |
| `srd-5.2.1:character/feature/druid/wild-resurgence` | #653 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/6` | #651 | level-6 |
| `srd-5.2.1:character/level/druid/7` | #651 | level-7 |
| `srd-5.2.1:character/feature/druid/elemental-fury` | #655, #657 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/8` | #651 | level-8 |
| `srd-5.2.1:character/level/druid/9` | #651 | level-9 |
| `srd-5.2.1:character/level/druid/10` | #651 | level-10 |
| `srd-5.2.1:character/level/druid/11` | #651 | level-11 |
| `srd-5.2.1:character/level/druid/12` | #651 | level-12 |
| `srd-5.2.1:character/level/druid/13` | #651 | level-13 |
| `srd-5.2.1:character/level/druid/14` | #651 | level-14 |
| `srd-5.2.1:character/level/druid/15` | #651 | level-15 |
| `srd-5.2.1:character/feature/druid/improved-elemental-fury` | #655, #658, #659, #657 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/16` | #651 | level-16 |
| `srd-5.2.1:character/level/druid/17` | #651 | level-17 |
| `srd-5.2.1:character/level/druid/18` | #651 | level-18 |
| `srd-5.2.1:character/feature/druid/beast-spells` | #660 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/19` | #651 | level-19 |
| `srd-5.2.1:character/feature/druid/epic-boon` | #661 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/level/druid/20` | #651 | level-20 |
| `srd-5.2.1:character/feature/druid/archdruid` | #662, #664, #642, #663 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/cantrips` | #503, #638 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/prepared-spells` | #503, #638 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/prepared-spells-slot-level` | #638 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-1` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-2` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-3` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-4` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-5` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-6` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-7` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-8` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/spell-slot-9` | #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/primal-order` | #641 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/magician` | #503, #642 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/warden` | #641 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/magician-selected` | #642 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/warden-selected` | #641 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/wild-shape` | #643 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/known-forms` | #644 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/known-form-eligibility` | #644 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/shape-equipment` | #635, #645, #647 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/wild-companion-cost` | #650, #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/druid-subclass` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/ability-score-improvement` | #652, #384 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/ability-score-improvement-qualification` | #384 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/epic-boon` | #661, #384 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/epic-boon-qualification` | #384 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/wild-resurgence-conversion` | #654, #653, #639 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/wild-resurgence-empty` | #653 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/wild-resurgence-slot-conversion` | #654 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/elemental-fury` | #655, #656, #657 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/potent-spellcasting` | #656 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/primal-strike` | #657 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/potent-spellcasting-selected` | #656 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/primal-strike-selected` | #657 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/improved-fury-choice` | #655 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/beast-spells-components` | #660 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/nature-magician-uses` | #642, #663 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/nature-magician` | #642 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/nature-magician-unexpended` | #663 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/subclass/druid/circle-of-the-land` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/circle-of-the-land-level` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/circle-of-the-land-spells` | #665 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/circle-of-the-land-spells-level` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/choice/druid/circle-of-the-land/land` | #665 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/spells-arid` | #666, #680, #676, #677, #678, #679, #681 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/spells-arid-selected` | #665 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/spells-polar` | #667, #687, #682, #683, #686, #684, #685 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/spells-polar-selected` | #665 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/spells-temperate` | #668, #692, #691, #688, #689, #690, #693 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/spells-temperate-selected` | #665 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/spells-tropical` | #669, #694, #699, #698, #695, #697, #696 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/spells-tropical-selected` | #665 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/lands-aid` | #670 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/lands-aid-level` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/lands-aid-wild-shape` | #670 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/natural-recovery` | #671, #672 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/natural-recovery-level` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/natural-recovery-operands` | #672 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/circle-of-the-land/natural-recovery-free-cast` | #671 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/resource/druid/circle-of-the-land/natural-recovery-slot-recovery` | #672 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/natures-ward` | #673 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/natures-ward-level` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/natures-ward-current-land` | #673 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/feature/druid/circle-of-the-land/natures-sanctuary` | #674, #675, #673 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/natures-sanctuary-level` | #651 | feature contract; entry/level grants separate |
| `srd-5.2.1:character/prerequisite/druid/circle-of-the-land/natures-sanctuary-operands` | #674 | feature contract; entry/level grants separate |
