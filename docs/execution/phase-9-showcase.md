---
title: Phase 9 Execution Paths — Showcase Features
type: tracking
created: 2026-04-04
tags: [tracking, phase-9, execution, export, replay]
---

# Phase 9: Showcase Features

> 20 issues across 2 tracks. Replay mode and campaign export — features that let players share, archive, and relive their campaigns.
> Created: 2026-04-04

## Phase Entry Criteria

Phase 8 complete: journal view with session summaries working, auth functional, audio system operational. (**Required.**)

## Phase Exit Criteria

- Replay mode plays back any campaign turn-by-turn with controls
- Campaign export produces PDF, markdown transcript, character sheet, and JSON dump
- All exports downloadable from the frontend

---

## Sprint 1: Replay Mode

### Track A: Campaign Replay

> Read-only playback of completed or in-progress campaigns.
> Depends on: Session history endpoint (already exists)

| #   | Issue                                                            | Title                                      | Size | Blocker | Model  | Notes                              |
| --- | ---------------------------------------------------------------- | ------------------------------------------ | :--: | ------- | ------ | ---------------------------------- |
| 1   | [#439](https://github.com/PatrickFanella/game-master/issues/439) | Create replay page route                   |  S   | None    | Sonnet | React Router + stub page           |
| 2   | [#444](https://github.com/PatrickFanella/game-master/issues/444) | Build replay playback engine               |  L   | #439    | Opus   | Timer state machine, speed control |
| 3   | [#445](https://github.com/PatrickFanella/game-master/issues/445) | Build replay controls component            |  M   | #444    | Sonnet | Play/pause, speed, skip            |
| 4   | [#448](https://github.com/PatrickFanella/game-master/issues/448) | Build replay timeline scrubber             |  M   | #444    | Sonnet | Draggable progress bar             |
| 5   | [#461](https://github.com/PatrickFanella/game-master/issues/461) | Implement typewriter rendering in replay   |  M   | #444    | Sonnet | Animated text at playback speed    |
| 6   | [#469](https://github.com/PatrickFanella/game-master/issues/469) | Replay sidebar with point-in-time state    |  L   | #444    | Opus   | Reconstruct state per turn         |
| 7   | [#474](https://github.com/PatrickFanella/game-master/issues/474) | Add "Watch Replay" button to campaign list |  S   | #439    | Sonnet | Navigation button                  |
| 8   | [#482](https://github.com/PatrickFanella/game-master/issues/482) | Optional: Replay sharing with URL          |  M   | #439    | Sonnet | Shareable link (if auth exists)    |

**#439 first. Then #444 (core engine). Then #445, #448, #461 in parallel (UI components). Then #469 (state reconstruction — complex). #474 anytime after #439. #482 last (stretch goal).**

---

## Sprint 2: Campaign Export

### Track B: Export System

> Multi-format export: JSON data dump, markdown transcript, character sheet, styled PDF.
> Depends on: Journal summaries (Phase 8 Track C) for PDF story section

| #   | Issue                                                            | Title                                          | Size | Blocker   | Model  | Notes                            |
| --- | ---------------------------------------------------------------- | ---------------------------------------------- | :--: | --------- | ------ | -------------------------------- |
| 1   | [#486](https://github.com/PatrickFanella/game-master/issues/486) | Implement JSON export endpoint                 |  M   | None      | Sonnet | Aggregate all campaign tables    |
| 2   | [#488](https://github.com/PatrickFanella/game-master/issues/488) | Implement markdown transcript export           |  M   | None      | Sonnet | Template-based, turn-by-turn     |
| 3   | [#489](https://github.com/PatrickFanella/game-master/issues/489) | Implement markdown character sheet export      |  S   | None      | Sonnet | Single character document        |
| 4   | [#490](https://github.com/PatrickFanella/game-master/issues/490) | Implement PDF export endpoint                  |  L   | None      | Opus   | Go PDF library, Art Deco styling |
| 5   | [#491](https://github.com/PatrickFanella/game-master/issues/491) | Add export button with format selection dialog |  M   | #486-#490 | Sonnet | Frontend modal + download        |
| 6   | [#492](https://github.com/PatrickFanella/game-master/issues/492) | Add loading indicator for PDF export           |  S   | #491      | Sonnet | Progress during generation       |

**#486, #488, #489, #490 all in parallel (each is an independent endpoint). Then #491 (frontend). Then #492.**

**Phase 9 total: 14 issues (plus 6 from earlier numbering = 20 total). Two clean sprints, both highly parallelizable.**

---

## Phase Summary

This is the final planned phase. After Phase 9, the project has:

- Intelligent tool filtering reducing LLM cognitive load
- Spoiler-safe information panels (NPCs, relationships, facts, codex, map)
- Full rules engine with narrative/light/crunch modes
- Dedicated combat UI
- Time/calendar system
- Save/resume/restart
- XP and leveling with feats and skills
- Streaming typewriter narrative
- Ambient audio with controls
- Thinking indicators with context
- Session journal with auto-summaries
- Multi-user authentication
- Campaign replay mode
- Multi-format campaign export

**Total issues across Phases 5-9: ~200 sub-issues from 24 epics.**

---
