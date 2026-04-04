---
title: Phase 7 Execution Paths — Game Systems
type: tracking
created: 2026-04-04
tags: [tracking, phase-7, execution, game-design, combat, rules]
---

# Phase 7: Game Systems

> 48 issues across 4 tracks. Adds the rules engine, combat UI, time/calendar, and save/resume systems. These are the mechanics that make the game feel like a game.
> Created: 2026-04-04

## Phase Entry Criteria

Phase 6 complete: information panels working, spoiler protection active, sidebar refreshes after turns. (**Required.**)

## Phase Exit Criteria

- Campaign creation includes rules mode selection (narrative / light / crunch)
- Combat UI exists for crunch mode; narrative mode stays text-only
- Time passes in-game; NPCs and environments reference it
- Players can save, resume, and start over
- Feats and skills available in crunch mode

---

## Sprint 1: Rules Engine Foundation

### Track A: Rules Mode & Schema

> Database schema for feats, skills, and the rules_mode campaign setting.
> Depends on: Nothing

| #   | Issue                                                            | Title                                      | Size | Blocker | Model  | Notes            |
| --- | ---------------------------------------------------------------- | ------------------------------------------ | :--: | ------- | ------ | ---------------- |
| 1   | [#344](https://github.com/PatrickFanella/game-master/issues/344) | Add rules_mode column to campaigns         |  S   | None    | Sonnet | Migration + sqlc |
| 2   | [#350](https://github.com/PatrickFanella/game-master/issues/350) | Create feat_definitions table              |  S   | None    | Sonnet | Migration + sqlc |
| 3   | [#353](https://github.com/PatrickFanella/game-master/issues/353) | Create character_feats table               |  S   | None    | Sonnet | Migration + sqlc |
| 4   | [#355](https://github.com/PatrickFanella/game-master/issues/355) | Create skill_definitions table             |  S   | None    | Sonnet | Migration + sqlc |
| 5   | [#357](https://github.com/PatrickFanella/game-master/issues/357) | Create character_skills table              |  S   | None    | Sonnet | Migration + sqlc |
| 6   | [#347](https://github.com/PatrickFanella/game-master/issues/347) | Add rules_mode to campaign creation wizard |  M   | #344    | Sonnet | Frontend step    |

**All migrations (#344, #350, #353, #355, #357) in parallel. Then #347.**

### Track B: Feats & Skills Implementation

> Tools, seeding, and integration for crunch mode.
> Depends on: Track A schema

| #   | Issue                                                            | Title                                   | Size | Blocker    | Model  | Notes                         |
| --- | ---------------------------------------------------------------- | --------------------------------------- | :--: | ---------- | ------ | ----------------------------- |
| 1   | [#363](https://github.com/PatrickFanella/game-master/issues/363) | Seed default feats for crunch mode      |  M   | #350       | Opus   | LLM-generated per genre       |
| 2   | [#367](https://github.com/PatrickFanella/game-master/issues/367) | Seed default skills for crunch mode     |  M   | #355       | Opus   | Genre-appropriate skills      |
| 3   | [#373](https://github.com/PatrickFanella/game-master/issues/373) | Implement grant_feat tool               |  M   | #350, #353 | Sonnet | Prerequisite validation       |
| 4   | [#374](https://github.com/PatrickFanella/game-master/issues/374) | Implement allocate_skill tool           |  M   | #355, #357 | Sonnet | Point validation              |
| 5   | [#380](https://github.com/PatrickFanella/game-master/issues/380) | Integrate feat bonuses into skill_check |  M   | #373       | Opus   | Modifier lookups              |
| 6   | [#392](https://github.com/PatrickFanella/game-master/issues/392) | Adapt system prompt by rules_mode       |  M   | #344       | Opus   | 3 prompt variants             |
| 7   | [#395](https://github.com/PatrickFanella/game-master/issues/395) | Adapt tool filtering by rules_mode      |  M   | #344       | Sonnet | Narrative: fewer tools        |
| 8   | [#384](https://github.com/PatrickFanella/game-master/issues/384) | Feat browser frontend component         |  M   | #373       | Sonnet | Character panel (crunch only) |
| 9   | [#387](https://github.com/PatrickFanella/game-master/issues/387) | Skill tree frontend component           |  M   | #374       | Sonnet | Skill list + allocation       |

**#363, #367 (parallel, need Opus for LLM generation). #373, #374 (parallel). #380 after #373. #392, #395 anytime after #344. #384, #387 after their tools.**

---

## Sprint 2: Combat UI

### Track C: Combat System

> Dedicated combat view for crunch mode.
> Depends on: Track A (#344 for rules_mode)

| #   | Issue                                                            | Title                                | Size | Blocker                | Model  | Notes                       |
| --- | ---------------------------------------------------------------- | ------------------------------------ | :--: | ---------------------- | ------ | --------------------------- |
| 1   | [#381](https://github.com/PatrickFanella/game-master/issues/381) | Add combat_active flag to game state |  M   | None                   | Sonnet | Schema + state gathering    |
| 2   | [#386](https://github.com/PatrickFanella/game-master/issues/386) | Add WebSocket combat events          |  M   | #381                   | Opus   | New event types             |
| 3   | [#390](https://github.com/PatrickFanella/game-master/issues/390) | Build initiative tracker component   |  M   | #386                   | Sonnet | Turn order bar              |
| 4   | [#393](https://github.com/PatrickFanella/game-master/issues/393) | Build combatant cards component      |  M   | #386                   | Sonnet | HP bars, conditions         |
| 5   | [#397](https://github.com/PatrickFanella/game-master/issues/397) | Build combat action bar              |  M   | #386                   | Sonnet | Attack/Defend/Spell buttons |
| 6   | [#401](https://github.com/PatrickFanella/game-master/issues/401) | Build combat log sub-panel           |  M   | #386                   | Sonnet | Scrollable combat narrative |
| 7   | [#403](https://github.com/PatrickFanella/game-master/issues/403) | Implement combat view transition     |  L   | #390, #393, #397, #401 | Opus   | State-driven view switch    |
| 8   | [#408](https://github.com/PatrickFanella/game-master/issues/408) | Combat-only tool filtering           |  S   | #381                   | Sonnet | Ties into Epic #289         |
| 9   | [#410](https://github.com/PatrickFanella/game-master/issues/410) | Rules mode toggle for combat UI      |  S   | #344, #403             | Sonnet | Narrative stays text-only   |

**#381 first. #386 and #408 after #381. #390, #393, #397, #401 in parallel after #386. #403 after all components. #410 last.**

---

## Sprint 3: Time & Save Systems

### Track D: Time/Calendar + Save/Resume

> Two independent systems that can be built in parallel.
> Depends on: Nothing

**Time/Calendar sub-track:**

| #   | Issue                                                            | Title                                      | Size | Blocker | Model  | Notes                  |
| --- | ---------------------------------------------------------------- | ------------------------------------------ | :--: | ------- | ------ | ---------------------- |
| 1   | [#418](https://github.com/PatrickFanella/game-master/issues/418) | Create campaign_time table + sqlc          |  S   | None    | Sonnet | Migration              |
| 2   | [#420](https://github.com/PatrickFanella/game-master/issues/420) | Initialize campaign_time in world creation |  M   | #418    | Sonnet | Orchestrator hook      |
| 3   | [#425](https://github.com/PatrickFanella/game-master/issues/425) | Implement advance_time tool                |  M   | #418    | Sonnet | Hour/day math          |
| 4   | [#427](https://github.com/PatrickFanella/game-master/issues/427) | Implement rest tool                        |  M   | #425    | Sonnet | Time + HP heal         |
| 5   | [#431](https://github.com/PatrickFanella/game-master/issues/431) | Update move_player to deduct travel time   |  M   | #425    | Sonnet | Connection travel_time |
| 6   | [#434](https://github.com/PatrickFanella/game-master/issues/434) | Add time to game state serialization       |  S   | #418    | Sonnet | LLM sees current time  |
| 7   | [#437](https://github.com/PatrickFanella/game-master/issues/437) | Campaign time REST endpoint                |  S   | #418    | Sonnet | GET endpoint           |
| 8   | [#440](https://github.com/PatrickFanella/game-master/issues/440) | Time display widget frontend               |  M   | #437    | Sonnet | Clock in header        |
| 9   | [#443](https://github.com/PatrickFanella/game-master/issues/443) | Time-of-day system prompt guidance         |  S   | #434    | Sonnet | Prompt update          |

**#418 first. Then #420, #425, #434, #437 in parallel. Then #427, #431, #440, #443.**

**Save/Resume sub-track:**

| #   | Issue                                                            | Title                                     | Size | Blocker | Model  | Notes                      |
| --- | ---------------------------------------------------------------- | ----------------------------------------- | :--: | ------- | ------ | -------------------------- |
| 1   | [#447](https://github.com/PatrickFanella/game-master/issues/447) | Create save_points table + sqlc           |  S   | None    | Sonnet | Migration                  |
| 2   | [#454](https://github.com/PatrickFanella/game-master/issues/454) | Implement auto-save after each turn       |  M   | #447    | Sonnet | Rolling 3 saves            |
| 3   | [#458](https://github.com/PatrickFanella/game-master/issues/458) | Implement manual save endpoint            |  S   | #447    | Sonnet | POST handler               |
| 4   | [#462](https://github.com/PatrickFanella/game-master/issues/462) | Implement list saves endpoint             |  S   | #447    | Sonnet | GET handler                |
| 5   | [#467](https://github.com/PatrickFanella/game-master/issues/467) | Implement proper campaign resume endpoint |  M   | None    | Opus   | Last narrative + choices   |
| 6   | [#475](https://github.com/PatrickFanella/game-master/issues/475) | Update frontend resume to seed narrative  |  M   | #467    | Sonnet | Load last state on mount   |
| 7   | [#481](https://github.com/PatrickFanella/game-master/issues/481) | Implement start over (restart) endpoint   |  M   | None    | Opus   | Destructive: clear + regen |
| 8   | [#484](https://github.com/PatrickFanella/game-master/issues/484) | Save button in frontend header            |  S   | #458    | Sonnet | Icon + name dialog         |
| 9   | [#485](https://github.com/PatrickFanella/game-master/issues/485) | Save list in campaign menu                |  M   | #462    | Sonnet | List + load option         |
| 10  | [#487](https://github.com/PatrickFanella/game-master/issues/487) | Start over button with confirmation       |  S   | #481    | Sonnet | Destructive dialog         |

**#447 first. #454, #458, #462 after #447 (parallel). #467 and #481 independent (parallel). Frontend after endpoints.**

**Phase 7 total: 48 issues. Three sprints, heavily parallelizable within each sprint.**

---
