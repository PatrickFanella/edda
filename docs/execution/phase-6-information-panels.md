---
title: Phase 6 Execution Paths — Information Panels
type: tracking
created: 2026-04-04
tags: [tracking, phase-6, execution, frontend, spoiler-safety]
---

# Phase 6: Information Panels

> 53 issues across 5 tracks. Adds sidebar panels for NPCs, relationships, world facts, codex, and quests — all with spoiler protection. A shared pattern: add `player_known`/`player_aware` columns, update tools with `reveal_to_player`, add filtered endpoints, build frontend panels.
> Created: 2026-04-04

## Phase Entry Criteria

Phase 5 complete: sidebar refresh works after turns, streaming functional, tool filtering active. (**Required.**)

## Phase Exit Criteria

- Player sees only information their character has discovered
- NPC panel shows encountered NPCs with dialogue history
- Relationships tab visualizes known connections
- World facts panel shows discovered lore
- Codex panel shows discovered world systems
- Quest panel supports notes, pinning, and version history

---

## Sprint 1: Spoiler Infrastructure (migrations + tool updates)

> All panels share the same spoiler-safety pattern. Do the schema changes first.

### Track A: Spoiler Columns & Tool Updates

> Database migrations and tool parameter additions. All independent.
> Depends on: Nothing

| #   | Issue                                                            | Title                                               | Size | Blocker | Model  | Notes              |
| --- | ---------------------------------------------------------------- | --------------------------------------------------- | :--: | ------- | ------ | ------------------ |
| 1   | [#351](https://github.com/PatrickFanella/game-master/issues/351) | Add player_known to world_facts                     |  S   | None    | Sonnet | Migration only     |
| 2   | [#336](https://github.com/PatrickFanella/game-master/issues/336) | Add player_aware to entity_relationships            |  S   | None    | Sonnet | Migration only     |
| 3   | [#325](https://github.com/PatrickFanella/game-master/issues/325) | Add player_known to languages                       |  S   | None    | Sonnet | Migration only     |
| 4   | [#327](https://github.com/PatrickFanella/game-master/issues/327) | Add player_known to cultures                        |  S   | None    | Sonnet | Migration only     |
| 5   | [#329](https://github.com/PatrickFanella/game-master/issues/329) | Add player_known to belief_systems                  |  S   | None    | Sonnet | Migration only     |
| 6   | [#330](https://github.com/PatrickFanella/game-master/issues/330) | Add player_known to economic_systems                |  S   | None    | Sonnet | Migration only     |
| 7   | [#352](https://github.com/PatrickFanella/game-master/issues/352) | Add player_visited/player_known to locations        |  S   | None    | Sonnet | Migration only     |
| 8   | [#356](https://github.com/PatrickFanella/game-master/issues/356) | Update establish_fact with reveal_to_player         |  S   | #351    | Sonnet | Tool param         |
| 9   | [#361](https://github.com/PatrickFanella/game-master/issues/361) | Update revise_fact to preserve player_known         |  S   | #351    | Sonnet | Inherit visibility |
| 10  | [#338](https://github.com/PatrickFanella/game-master/issues/338) | Update establish_relationship with reveal_to_player |  S   | #336    | Sonnet | Tool param         |
| 11  | [#333](https://github.com/PatrickFanella/game-master/issues/333) | Update create_language with reveal_to_player        |  S   | #325    | Sonnet | Tool param         |
| 12  | [#334](https://github.com/PatrickFanella/game-master/issues/334) | Update create_culture with reveal_to_player         |  S   | #327    | Sonnet | Tool param         |
| 13  | [#335](https://github.com/PatrickFanella/game-master/issues/335) | Update create_belief_system with reveal_to_player   |  S   | #329    | Sonnet | Tool param         |
| 14  | [#337](https://github.com/PatrickFanella/game-master/issues/337) | Update create_economic_system with reveal_to_player |  S   | #330    | Sonnet | Tool param         |
| 15  | [#359](https://github.com/PatrickFanella/game-master/issues/359) | Update move_player to set player_visited            |  S   | #352    | Sonnet | Tool hook          |
| 16  | [#360](https://github.com/PatrickFanella/game-master/issues/360) | Add reveal_location tool                            |  S   | #352    | Sonnet | New small tool     |

**All migrations (#351, #336, #325, #327, #329, #330, #352) run in parallel. Tool updates follow their migration.**

---

## Sprint 2: Backend Endpoints

### Track B: NPC & Relationships Endpoints

> Filtered endpoints for spoiler-safe NPC and relationship data.
> Depends on: Track A migrations

| #   | Issue                                                            | Title                               | Size | Blocker | Model  | Notes                      |
| --- | ---------------------------------------------------------------- | ----------------------------------- | :--: | ------- | ------ | -------------------------- |
| 1   | [#320](https://github.com/PatrickFanella/game-master/issues/320) | Encountered NPCs endpoint           |  M   | None    | Sonnet | Join npcs + session_logs   |
| 2   | [#322](https://github.com/PatrickFanella/game-master/issues/322) | NPC dialogue history endpoint       |  M   | #320    | Sonnet | Filter session_logs by NPC |
| 3   | [#340](https://github.com/PatrickFanella/game-master/issues/340) | Player-aware relationships endpoint |  M   | #336    | Sonnet | Filtered query             |

**#320 and #340 in parallel. #322 after #320.**

### Track C: Facts, Codex & Map Endpoints

> Filtered endpoints for world data.
> Depends on: Track A migrations

| #   | Issue                                                            | Title                                                     | Size | Blocker   | Model  | Notes                                 |
| --- | ---------------------------------------------------------------- | --------------------------------------------------------- | :--: | --------- | ------ | ------------------------------------- |
| 1   | [#366](https://github.com/PatrickFanella/game-master/issues/366) | Player-known facts endpoint                               |  M   | #351      | Sonnet | Filtered + non-superseded             |
| 2   | [#339](https://github.com/PatrickFanella/game-master/issues/339) | Codex endpoints (languages, cultures, beliefs, economies) |  L   | #325-#330 | Sonnet | 4 endpoints, same pattern             |
| 3   | [#362](https://github.com/PatrickFanella/game-master/issues/362) | Map data endpoint                                         |  M   | #352      | Sonnet | Visited/known locations + connections |

**All three in parallel.**

### Track D: Quest Backend

> Quest notes, history, and prompt updates.
> Depends on: Nothing

| #   | Issue                                                            | Title                                               | Size | Blocker | Model  | Notes                    |
| --- | ---------------------------------------------------------------- | --------------------------------------------------- | :--: | ------- | ------ | ------------------------ |
| 1   | [#449](https://github.com/PatrickFanella/game-master/issues/449) | Create quest_notes migration + sqlc                 |  S   | None    | Sonnet | New table                |
| 2   | [#450](https://github.com/PatrickFanella/game-master/issues/450) | Create quest_history migration + sqlc               |  S   | None    | Sonnet | New table                |
| 3   | [#452](https://github.com/PatrickFanella/game-master/issues/452) | Quest notes API endpoints                           |  M   | #449    | Sonnet | CRUD handlers            |
| 4   | [#453](https://github.com/PatrickFanella/game-master/issues/453) | Quest history API endpoint                          |  M   | #450    | Sonnet | GET handler              |
| 5   | [#455](https://github.com/PatrickFanella/game-master/issues/455) | Trigger history snapshot on tool calls              |  M   | #450    | Opus   | Hook into turn_processor |
| 6   | [#470](https://github.com/PatrickFanella/game-master/issues/470) | Update system prompt for proactive quest management |  S   | None    | Sonnet | Prompt update            |

**#449 and #450 first (parallel). Then #452, #453, #455 (parallel). #470 anytime.**

---

## Sprint 3: Frontend Panels

### Track E: NPC Panel

> Frontend for NPC discovery and dialogue.
> Depends on: Track B endpoints

| #   | Issue                                                            | Title                            | Size | Blocker    | Model  | Notes                |
| --- | ---------------------------------------------------------------- | -------------------------------- | :--: | ---------- | ------ | -------------------- |
| 1   | [#323](https://github.com/PatrickFanella/game-master/issues/323) | NPC frontend API functions       |  S   | #320, #322 | Sonnet | TypeScript API calls |
| 2   | [#326](https://github.com/PatrickFanella/game-master/issues/326) | NPC list panel component         |  M   | #323       | Sonnet | Sidebar tab          |
| 3   | [#331](https://github.com/PatrickFanella/game-master/issues/331) | NPC detail view with dialogue    |  M   | #323       | Sonnet | Expanded view        |
| 4   | [#332](https://github.com/PatrickFanella/game-master/issues/332) | "New NPC met" toast notification |  S   | #326       | Sonnet | Event detection      |

**#323 first. Then #326 and #331 in parallel. Then #332.**

### Track F: Relationships Panel

> Frontend for relationship visualization.
> Depends on: Track B endpoint #340

| #   | Issue                                                            | Title                                       | Size | Blocker | Model  | Notes            |
| --- | ---------------------------------------------------------------- | ------------------------------------------- | :--: | ------- | ------ | ---------------- |
| 1   | [#342](https://github.com/PatrickFanella/game-master/issues/342) | Relationships frontend API types + function |  S   | #340    | Sonnet | TypeScript       |
| 2   | [#345](https://github.com/PatrickFanella/game-master/issues/345) | Relationships list panel                    |  M   | #342    | Sonnet | Sidebar tab      |
| 3   | [#348](https://github.com/PatrickFanella/game-master/issues/348) | Optional: Force-directed relationship graph |  L   | #342    | Opus   | d3 visualization |

**#342 → #345 → #348 (linear).**

### Track G: Facts & Codex Panels

> Frontend for world knowledge.
> Depends on: Track C endpoints

| #   | Issue                                                            | Title                                | Size | Blocker | Model  | Notes       |
| --- | ---------------------------------------------------------------- | ------------------------------------ | :--: | ------- | ------ | ----------- |
| 1   | [#371](https://github.com/PatrickFanella/game-master/issues/371) | Facts frontend API types + function  |  S   | #366    | Sonnet | TypeScript  |
| 2   | [#378](https://github.com/PatrickFanella/game-master/issues/378) | Facts/lore panel with categories     |  M   | #371    | Sonnet | Sidebar tab |
| 3   | [#382](https://github.com/PatrickFanella/game-master/issues/382) | "New lore discovered" notification   |  S   | #378    | Sonnet | Badge       |
| 4   | [#341](https://github.com/PatrickFanella/game-master/issues/341) | Codex frontend API types + functions |  S   | #339    | Sonnet | TypeScript  |
| 5   | [#343](https://github.com/PatrickFanella/game-master/issues/343) | Codex panel with sub-navigation      |  M   | #341    | Sonnet | Sidebar tab |
| 6   | [#346](https://github.com/PatrickFanella/game-master/issues/346) | "New discovery" notification badge   |  S   | #343    | Sonnet | Badge       |

**Facts sub-track (#371→#378→#382) and codex sub-track (#341→#343→#346) run in parallel.**

### Track H: Quest Frontend

> Quest notes, history, pinning UI.
> Depends on: Track D endpoints

| #   | Issue                                                            | Title                              | Size | Blocker | Model  | Notes                          |
| --- | ---------------------------------------------------------------- | ---------------------------------- | :--: | ------- | ------ | ------------------------------ |
| 1   | [#456](https://github.com/PatrickFanella/game-master/issues/456) | Quest notes frontend types + API   |  S   | #452    | Sonnet | TypeScript                     |
| 2   | [#459](https://github.com/PatrickFanella/game-master/issues/459) | Quest history frontend types + API |  S   | #453    | Sonnet | TypeScript                     |
| 3   | [#460](https://github.com/PatrickFanella/game-master/issues/460) | Quest notes UI component           |  M   | #456    | Sonnet | Textarea per quest             |
| 4   | [#464](https://github.com/PatrickFanella/game-master/issues/464) | Quest history UI component         |  M   | #459    | Sonnet | Expandable versions            |
| 5   | [#466](https://github.com/PatrickFanella/game-master/issues/466) | Pinned objectives UI               |  M   | None    | Sonnet | Floating banner + localStorage |
| 6   | [#468](https://github.com/PatrickFanella/game-master/issues/468) | Quest state change notifications   |  S   | None    | Sonnet | Toast on quest events          |

**#456 and #459 first (parallel). Then #460 and #464 (parallel). #466 and #468 anytime (no backend deps).**

### Track I: Map Panel

> Location visualization.
> Depends on: Track C endpoint #362

| #   | Issue                                                            | Title                             | Size | Blocker | Model  | Notes                      |
| --- | ---------------------------------------------------------------- | --------------------------------- | :--: | ------- | ------ | -------------------------- |
| 1   | [#365](https://github.com/PatrickFanella/game-master/issues/365) | Map frontend API types + function |  S   | #362    | Sonnet | TypeScript                 |
| 2   | [#368](https://github.com/PatrickFanella/game-master/issues/368) | Map panel with node graph         |  L   | #365    | Opus   | d3/cytoscape visualization |
| 3   | [#372](https://github.com/PatrickFanella/game-master/issues/372) | Click-to-view location details    |  S   | #368    | Sonnet | Detail panel               |
| 4   | [#375](https://github.com/PatrickFanella/game-master/issues/375) | Optional: generate_local_map tool |  M   | None    | Sonnet | Stretch goal               |

**#365 → #368 → #372 (linear). #375 independent.**

**Phase 6 total: 53 issues. Sprint 1 is all parallel migrations. Sprint 2 is parallel endpoints. Sprint 3 is parallel frontend panels.**

---
