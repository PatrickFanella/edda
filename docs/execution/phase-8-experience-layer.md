---
title: Phase 8 Execution Paths — Experience Layer
type: tracking
created: 2026-04-04
tags: [tracking, phase-8, execution, ux, audio, journal]
---

# Phase 8: Experience Layer

> 38 issues across 4 tracks. Adds the features that transform the app from functional to delightful: ambient audio, thinking indicators, journal view, and authentication.
> Created: 2026-04-04

## Phase Entry Criteria

Phase 7 complete: rules engine functional, combat UI working, time/calendar active, save/resume operational. (**Required.**)

## Phase Exit Criteria

- Ambient music and SFX play during gameplay with per-layer controls
- Thinking indicator shows what the engine is doing during processing
- Journal summarizes past sessions into readable chapters
- Authentication supports multiple user accounts

---

## Sprint 1: Feedback & Immersion

### Track A: Thinking Indicator

> Show players what's happening during turn processing and world generation.
> Depends on: Nothing (backend WebSocket already exists)

| #   | Issue                                                            | Title                                             | Size | Blocker    | Model  | Notes                                 |
| --- | ---------------------------------------------------------------- | ------------------------------------------------- | :--: | ---------- | ------ | ------------------------------------- |
| 1   | [#411](https://github.com/PatrickFanella/game-master/issues/411) | Add WebSocket status events for tool execution    |  M   | None       | Opus   | New event type from turn_processor    |
| 2   | [#414](https://github.com/PatrickFanella/game-master/issues/414) | Add WebSocket status events for processing stages |  M   | None       | Opus   | Events from engine.ProcessTurn        |
| 3   | [#417](https://github.com/PatrickFanella/game-master/issues/417) | Map tool names to human-readable descriptions     |  S   | None       | Sonnet | Frontend mapping config               |
| 4   | [#433](https://github.com/PatrickFanella/game-master/issues/433) | Process status events in useWebSocket             |  M   | #411, #414 | Sonnet | Handle new event type                 |
| 5   | [#421](https://github.com/PatrickFanella/game-master/issues/421) | Build contextual thinking indicator component     |  M   | #433, #417 | Sonnet | Animated status line                  |
| 6   | [#429](https://github.com/PatrickFanella/game-master/issues/429) | Add world generation progress streaming           |  L   | #414       | Opus   | SSE or staged response for BuildWorld |

**#411, #414, #417 in parallel. Then #433. Then #421. #429 after #414 (independent of frontend).**

### Track B: Ambient Audio

> Three-layer audio system: ambient, music, SFX.
> Depends on: Nothing

| #   | Issue                                                            | Title                            | Size | Blocker          | Model  | Notes                                               |
| --- | ---------------------------------------------------------------- | -------------------------------- | :--: | ---------------- | ------ | --------------------------------------------------- |
| 1   | [#369](https://github.com/PatrickFanella/game-master/issues/369) | Create useAudio hook             |  M   | None             | Sonnet | Three layers, localStorage persistence              |
| 2   | [#376](https://github.com/PatrickFanella/game-master/issues/376) | Define audio asset mapping       |  S   | None             | Sonnet | JSON config: location→ambient, tone→music, tool→sfx |
| 3   | [#379](https://github.com/PatrickFanella/game-master/issues/379) | Source royalty-free audio assets |  M   | None             | Sonnet | 5 ambient + 4 music + 5 sfx                         |
| 4   | [#385](https://github.com/PatrickFanella/game-master/issues/385) | Implement ambient audio layer    |  M   | #369, #376, #379 | Sonnet | Crossfade on location change                        |
| 5   | [#391](https://github.com/PatrickFanella/game-master/issues/391) | Implement music layer            |  M   | #369, #376, #379 | Sonnet | State-based track selection                         |
| 6   | [#396](https://github.com/PatrickFanella/game-master/issues/396) | Implement SFX layer              |  M   | #369, #376, #379 | Sonnet | Tool call → sound                                   |
| 7   | [#400](https://github.com/PatrickFanella/game-master/issues/400) | Build audio controls UI          |  M   | #369             | Sonnet | Volume sliders, mute toggles                        |
| 8   | [#405](https://github.com/PatrickFanella/game-master/issues/405) | Handle browser autoplay policy   |  S   | #369             | Sonnet | Require user interaction                            |

**#369, #376, #379 in parallel (foundation). Then #385, #391, #396 in parallel (layers). #400 and #405 after #369.**

---

## Sprint 2: Journal & Auth

### Track C: Journal View

> Session summaries and player notes.
> Depends on: Nothing

| #   | Issue                                                            | Title                                  | Size | Blocker    | Model  | Notes                               |
| --- | ---------------------------------------------------------------- | -------------------------------------- | :--: | ---------- | ------ | ----------------------------------- |
| 1   | [#446](https://github.com/PatrickFanella/game-master/issues/446) | Create session_summaries table         |  S   | None       | Sonnet | Migration + sqlc                    |
| 2   | [#451](https://github.com/PatrickFanella/game-master/issues/451) | Create player_journal_entries table    |  S   | None       | Sonnet | Migration + sqlc                    |
| 3   | [#457](https://github.com/PatrickFanella/game-master/issues/457) | Implement summarization service        |  L   | #446       | Opus   | Uses summarizer.txt prompt          |
| 4   | [#463](https://github.com/PatrickFanella/game-master/issues/463) | Implement auto-summarization trigger   |  M   | #457       | Sonnet | Every 10 turns                      |
| 5   | [#465](https://github.com/PatrickFanella/game-master/issues/465) | Add journal REST endpoints             |  M   | #446, #451 | Sonnet | GET summaries + CRUD entries        |
| 6   | [#471](https://github.com/PatrickFanella/game-master/issues/471) | Add manual summarize endpoint          |  S   | #457       | Sonnet | POST force-summarize                |
| 7   | [#477](https://github.com/PatrickFanella/game-master/issues/477) | Journal frontend API types + functions |  S   | #465       | Sonnet | TypeScript                          |
| 8   | [#483](https://github.com/PatrickFanella/game-master/issues/483) | Build journal panel component          |  M   | #477       | Sonnet | Chapters + notes + summarize button |

**#446 and #451 first (parallel). Then #457, #465 (parallel). Then #463, #471, #477 (parallel). Then #483.**

### Track D: Authentication

> Multi-user support with session-based auth.
> Depends on: Nothing (but should be one of the last features to avoid breaking existing flows)

| #   | Issue                                                            | Title                              | Size | Blocker          | Model  | Notes                             |
| --- | ---------------------------------------------------------------- | ---------------------------------- | :--: | ---------------- | ------ | --------------------------------- |
| 1   | [#413](https://github.com/PatrickFanella/game-master/issues/413) | Add auth columns to users table    |  S   | None             | Sonnet | email + password_hash migration   |
| 2   | [#416](https://github.com/PatrickFanella/game-master/issues/416) | Create sessions table              |  S   | None             | Sonnet | Migration + sqlc                  |
| 3   | [#419](https://github.com/PatrickFanella/game-master/issues/419) | Implement password hashing utility |  S   | None             | Sonnet | bcrypt functions                  |
| 4   | [#423](https://github.com/PatrickFanella/game-master/issues/423) | Implement register endpoint        |  M   | #413, #419       | Sonnet | POST /auth/register               |
| 5   | [#424](https://github.com/PatrickFanella/game-master/issues/424) | Implement login endpoint           |  M   | #413, #416, #419 | Sonnet | POST /auth/login + session cookie |
| 6   | [#426](https://github.com/PatrickFanella/game-master/issues/426) | Implement logout endpoint          |  S   | #416             | Sonnet | POST /auth/logout                 |
| 7   | [#428](https://github.com/PatrickFanella/game-master/issues/428) | Implement session middleware       |  M   | #416             | Opus   | Replace NoOpMiddleware            |
| 8   | [#430](https://github.com/PatrickFanella/game-master/issues/430) | Implement GET /auth/me endpoint    |  S   | #428             | Sonnet | Current user info                 |
| 9   | [#432](https://github.com/PatrickFanella/game-master/issues/432) | Add session cleanup job            |  S   | #416             | Sonnet | Goroutine periodic cleanup        |
| 10  | [#435](https://github.com/PatrickFanella/game-master/issues/435) | Build login page frontend          |  M   | #424             | Sonnet | Email + password form             |
| 11  | [#436](https://github.com/PatrickFanella/game-master/issues/436) | Build register page frontend       |  M   | #423             | Sonnet | Name + email + password form      |
| 12  | [#438](https://github.com/PatrickFanella/game-master/issues/438) | Implement auth context in frontend |  M   | #430             | Sonnet | React context + provider          |
| 13  | [#441](https://github.com/PatrickFanella/game-master/issues/441) | Add protected routes               |  S   | #438             | Sonnet | Redirect if unauthenticated       |
| 14  | [#442](https://github.com/PatrickFanella/game-master/issues/442) | Add user menu to header            |  S   | #438             | Sonnet | Name + logout                     |

**#413, #416, #419 in parallel (foundation). Then #423, #424, #426, #428 in parallel (endpoints). Then #430, #432 (parallel). Then frontend: #435, #436 (parallel) → #438 → #441, #442 (parallel).**

**Phase 8 total: 38 issues. Tracks A+B are Sprint 1 (immersion). Tracks C+D are Sprint 2 (infrastructure).**

---
