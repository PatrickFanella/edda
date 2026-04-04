---
title: Phase 5 Execution Paths — Gameplay Quality
type: tracking
created: 2026-04-04
tags: [tracking, phase-5, execution, gameplay]
---

# Phase 5: Gameplay Quality

> 42 issues across 5 tracks. Focused on making the core gameplay loop reliable, responsive, and immersive before adding new features.
> Created: 2026-04-04

## Phase Entry Criteria

Core gameplay loop working end-to-end via web frontend: campaign creation wizard, turn processing, WebSocket streaming, session history loading. (**Met.**)

## Phase Exit Criteria

- Tool filtering reduces per-turn tool count to 10-15
- Sidebar panels refresh after every turn
- Streaming typewriter effect works in frontend
- Semantic memory enriches LLM context
- Scroll behavior is non-disruptive
- XP/leveling system functional

---

## Sprint 1: Engine Reliability (fix what's broken before adding features)

### Track A: Tool Filtering

> Reduce 41 tools to ~10-15 per turn based on game phase.
> Depends on: Nothing

| #   | Issue                                                            | Title                                          | Size | Blocker   | Model  | Notes                                    |
| --- | ---------------------------------------------------------------- | ---------------------------------------------- | :--: | --------- | ------ | ---------------------------------------- |
| 1   | [#313](https://github.com/PatrickFanella/game-master/issues/313) | Define ToolFilter interface and GamePhase enum |  S   | None      | Sonnet | Architecture: interface + enum types     |
| 2   | [#314](https://github.com/PatrickFanella/game-master/issues/314) | Implement base tool set selection              |  S   | #313      | Sonnet | Always-on ~10 tools                      |
| 3   | [#315](https://github.com/PatrickFanella/game-master/issues/315) | Combat phase detection and tool injection      |  S   | #313      | Sonnet | combat_active → combat tools             |
| 4   | [#316](https://github.com/PatrickFanella/game-master/issues/316) | Exploration phase detection and tool injection |  S   | #313      | Sonnet | Add world-building tools                 |
| 5   | [#317](https://github.com/PatrickFanella/game-master/issues/317) | Quest tool injection                           |  S   | #313      | Sonnet | Active quests → quest tools              |
| 6   | [#318](https://github.com/PatrickFanella/game-master/issues/318) | Progression tool injection                     |  S   | #313      | Sonnet | Near level threshold → progression tools |
| 7   | [#319](https://github.com/PatrickFanella/game-master/issues/319) | Wire ToolFilter into engine.ProcessTurn        |  M   | #314-#318 | Opus   | Integration point                        |
| 8   | [#321](https://github.com/PatrickFanella/game-master/issues/321) | Update system prompt for all capabilities      |  S   | None      | Sonnet | Prompt still describes everything        |

**#313 first. Then #314-#318 in parallel. Then #319. #321 anytime.**

### Track B: Semantic Memory

> Wire existing embedding infrastructure into the server engine.
> Depends on: Nothing (all infrastructure exists)

| #   | Issue                                                            | Title                                       | Size | Blocker    | Model  | Notes                               |
| --- | ---------------------------------------------------------------- | ------------------------------------------- | :--: | ---------- | ------ | ----------------------------------- |
| 1   | [#389](https://github.com/PatrickFanella/game-master/issues/389) | Wire Tier3Retriever into server engine      |  M   | None       | Opus   | Architectural wiring                |
| 2   | [#422](https://github.com/PatrickFanella/game-master/issues/422) | Add embedding model configuration           |  S   | None       | Sonnet | Config field + env var              |
| 3   | [#399](https://github.com/PatrickFanella/game-master/issues/399) | Implement auto-embed after turn completion  |  L   | #389, #422 | Opus   | Async goroutine + summarizer prompt |
| 4   | [#406](https://github.com/PatrickFanella/game-master/issues/406) | Inject retrieved memories into turn context |  M   | #389       | Sonnet | Hook exists, just needs calling     |
| 5   | [#412](https://github.com/PatrickFanella/game-master/issues/412) | Implement search_memory tool                |  M   | #389       | Sonnet | New tool for GM to search past      |
| 6   | [#415](https://github.com/PatrickFanella/game-master/issues/415) | Add search_memory to system prompt          |  S   | #412       | Sonnet | Prompt update                       |

**#389 and #422 first (parallel). Then #399, #406, #412 in parallel. Then #415.**

### Track C: Streaming & UX Polish

> Frontend responsiveness improvements.
> Depends on: Nothing

| #   | Issue                                                            | Title                                            | Size | Blocker | Model  | Notes                    |
| --- | ---------------------------------------------------------------- | ------------------------------------------------ | :--: | ------- | ------ | ------------------------ |
| 1   | [#402](https://github.com/PatrickFanella/game-master/issues/402) | Render streaming chunks progressively            |  M   | None    | Sonnet | Core streaming fix       |
| 2   | [#404](https://github.com/PatrickFanella/game-master/issues/404) | Add blinking cursor during streaming             |  S   | #402    | Sonnet | CSS animation            |
| 3   | [#407](https://github.com/PatrickFanella/game-master/issues/407) | Remove thinking placeholder on first chunk       |  S   | #402    | Sonnet | Replace placeholder text |
| 4   | [#409](https://github.com/PatrickFanella/game-master/issues/409) | Smooth transition from streaming to final result |  S   | #402    | Sonnet | No flicker               |
| 5   | [#364](https://github.com/PatrickFanella/game-master/issues/364) | Track scroll position in narrative panel         |  S   | None    | Sonnet | onScroll handler         |
| 6   | [#370](https://github.com/PatrickFanella/game-master/issues/370) | Only auto-scroll when user is at bottom          |  S   | #364    | Sonnet | Conditional scroll       |
| 7   | [#377](https://github.com/PatrickFanella/game-master/issues/377) | Floating "New response" indicator                |  M   | #364    | Sonnet | Gold pill badge          |
| 8   | [#383](https://github.com/PatrickFanella/game-master/issues/383) | Identify React Query cache keys                  |  S   | None    | Sonnet | Research task            |
| 9   | [#388](https://github.com/PatrickFanella/game-master/issues/388) | Implement useRefreshAfterTurn hook               |  M   | #383    | Sonnet | Cache invalidation       |
| 10  | [#394](https://github.com/PatrickFanella/game-master/issues/394) | Wire useRefreshAfterTurn into play page          |  S   | #388    | Sonnet | Integration              |
| 11  | [#398](https://github.com/PatrickFanella/game-master/issues/398) | Selective invalidation from state_changes        |  M   | #388    | Sonnet | Optimization             |

**Three parallel sub-tracks: streaming (#402→#404→#407→#409), scroll (#364→#370→#377), refresh (#383→#388→#394→#398).**

---

## Sprint 2: XP & Progression

### Track D: Leveling System

> XP awards, level-up mechanics, and frontend display.
> Depends on: Nothing

| #   | Issue                                                            | Title                                     | Size | Blocker | Model  | Notes                     |
| --- | ---------------------------------------------------------------- | ----------------------------------------- | :--: | ------- | ------ | ------------------------- |
| 1   | [#472](https://github.com/PatrickFanella/game-master/issues/472) | Define XP threshold curve                 |  S   | None    | Sonnet | Pure logic                |
| 2   | [#473](https://github.com/PatrickFanella/game-master/issues/473) | Add XP award guidelines to system prompt  |  S   | None    | Sonnet | Prompt update             |
| 3   | [#476](https://github.com/PatrickFanella/game-master/issues/476) | Auto-level-up detection in add_experience |  M   | #472    | Sonnet | Tool enhancement          |
| 4   | [#478](https://github.com/PatrickFanella/game-master/issues/478) | HP increase on level-up                   |  S   | #472    | Sonnet | level_up tool enhancement |
| 5   | [#479](https://github.com/PatrickFanella/game-master/issues/479) | XP progress bar in character panel        |  M   | None    | Sonnet | Frontend component        |
| 6   | [#480](https://github.com/PatrickFanella/game-master/issues/480) | Level-up celebration notification         |  S   | #479    | Sonnet | Toast on level_up event   |

**#472 and #473 first (parallel). Then #476, #478 (parallel). #479 anytime. #480 after #479.**

### Track E: Quick Wins

> Small but impactful UX fixes.
> Depends on: Nothing

| #   | Issue                                                            | Title                                         | Size | Blocker | Model  | Notes                   |
| --- | ---------------------------------------------------------------- | --------------------------------------------- | :--: | ------- | ------ | ----------------------- |
| 1   | [#324](https://github.com/PatrickFanella/game-master/issues/324) | Equipped badge and rarity colors on inventory |  S   | None    | Sonnet | Frontend polish         |
| 2   | [#328](https://github.com/PatrickFanella/game-master/issues/328) | Refresh inventory after turn results          |  S   | None    | Sonnet | Cache invalidation      |
| 3   | [#349](https://github.com/PatrickFanella/game-master/issues/349) | Add delete button to campaign list            |  S   | None    | Sonnet | UI button               |
| 4   | [#354](https://github.com/PatrickFanella/game-master/issues/354) | Confirmation dialog for deletion              |  S   | #349    | Sonnet | Modal                   |
| 5   | [#358](https://github.com/PatrickFanella/game-master/issues/358) | Wire delete to API and refresh                |  S   | #354    | Sonnet | API call + state update |

**All quick and parallelizable. #349→#354→#358 is the only chain.**

**Phase 5 total: 42 issues. Tracks A-C are Sprint 1 (max parallelism). Tracks D-E are Sprint 2.**

---
