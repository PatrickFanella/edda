---
title: Phase 3 Execution Paths — Game Systems & API
type: tracking
created: 2026-03-21
tags: [tracking, phase-3, execution, game-systems]
---

# Phase 3: Game Systems & API

> 77 issues across 9 tracks. **39 ready** (when Phase 2 completes), 38 blocked by internal dependencies.
> Updated: 2026-03-21

## Summary

| Track | Name               | Total  | Ready  | Blocked | Epic | Parallel group | Models              |
| ----- | ------------------ | :----: | :----: | :-----: | ---- | -------------- | ------------------- |
| A     | Embedding & Memory |   7    |   4    |    3    | #7   | Alpha          | Mixed               |
| B     | Context Management |   8    |   0    |    8    | #8   | Beta (after A) | Mixed (Opus-heavy)  |
| C     | Campaign Creation  |   7    |   4    |    3    | #9   | Alpha          | Claude Opus 4.6     |
| D     | Player Character   |   9    |   6    |    3    | #10  | Alpha          | Mixed               |
| E     | World Generation   |   9    |   6    |    3    | #11  | Alpha          | gpt-5.3-codex       |
| F     | Quest System       |   8    |   6    |    2    | #12  | Alpha          | gpt-5.3-codex       |
| G     | Combat             |   10   |   2    |    8    | #13  | Beta (after D) | Claude Opus 4.6     |
| H     | Expanded World     |   9    |   5    |    4    | #14  | Beta (after E) | gpt-5.3-codex       |
| I     | REST API           |   10   |   5    |    5    | #15  | Alpha          | Mixed               |
|       | **Total**          | **77** | **38** | **39**  |      |                |                     |

**Parallel groups:**

- **Alpha** (Tracks A, C, D, E, F, I): All can start immediately when Phase 2 completes. Six independent tracks running simultaneously.
- **Beta** (Tracks B, G, H): Blocked by Alpha tracks. Track B needs Track A. Track G needs Track D. Track H needs Track E.

**Phase entry criteria:** Phase 2 complete (turn pipeline working, player can type and get LLM responses).

**Phase exit criteria:** All game systems implemented. Embedding + semantic memory working. Campaign creation, player management, world generation, quests, combat all functional. REST API serving game state. Expanded worldbuilding tools available.

---

## Track G: Combat

> Structured narrative combat system behind pluggable interface.
> Depends on: Track D (Player Character — needs stats/HP)

| #   | Issue                                                            | Title                                        | Size | Blocker    | Status  | Model             | Notes                         |
| --- | ---------------------------------------------------------------- | -------------------------------------------- | :--: | ---------- | ------- | ----------------- | ----------------------------- |
| 1   | [#139](https://github.com/PatrickFanella/game-master/issues/139) | Define CombatResolver interface              |  M   | Track D    | BLOCKED | Claude Opus 4.6   |                               |
| 2   | [#144](https://github.com/PatrickFanella/game-master/issues/144) | Implement skill_check tool                   |  S   | Phase 2    | READY   | gpt-5.3-codex     | Pure logic, no track dep      |
| 3   | [#140](https://github.com/PatrickFanella/game-master/issues/140) | Implement initiative ordering system         |  M   | #139       | BLOCKED | gpt-5.3-codex     |                               |
| 4   | [#141](https://github.com/PatrickFanella/game-master/issues/141) | Implement combatant state management         |  M   | #139       | BLOCKED | Claude Sonnet 4.6 |                               |
| 5   | [#142](https://github.com/PatrickFanella/game-master/issues/142) | Implement initiate_combat tool               |  M   | #139       | BLOCKED | gpt-5.3-codex     |                               |
| 6   | [#143](https://github.com/PatrickFanella/game-master/issues/143) | Implement combat_round tool                  |  L   | #140, #141 | BLOCKED | Claude Opus 4.6   |                               |
| 7   | [#145](https://github.com/PatrickFanella/game-master/issues/145) | Implement apply_damage/apply_condition tools |  M   | #141       | BLOCKED | gpt-5.3-codex     |                               |
| 8   | [#146](https://github.com/PatrickFanella/game-master/issues/146) | Implement resolve_combat tool                |  M   | #141       | BLOCKED | Claude Sonnet 4.6 |                               |
| 9   | [#147](https://github.com/PatrickFanella/game-master/issues/147) | Implement default narrative CombatResolver   |  L   | #140-#146  | BLOCKED | Claude Opus 4.6   | Ties everything together      |
| 10  | [#148](https://github.com/PatrickFanella/game-master/issues/148) | Interface contract tests for CombatResolver  |  L   | #147       | BLOCKED | Claude Opus 4.6   | Reusable for future rule sets |

```mermaid
graph TD
    TD[Track D: Player complete] --> 139[#139 CombatResolver interface]
    P2[Phase 2] --> 144[#144 skill_check]
    139 --> 140[#140 Initiative]
    139 --> 141[#141 Combatant state]
    139 --> 142[#142 initiate_combat]
    140 --> 143[#143 combat_round]
    141 --> 143
    141 --> 145[#145 damage/condition]
    141 --> 146[#146 resolve_combat]
    140 --> 147[#147 Default resolver]
    141 --> 147
    142 --> 147
    143 --> 147
    145 --> 147
    146 --> 147
    147 --> 148[#148 Contract tests]

    style 144 fill:#22c55e
    style 139 fill:#ef4444
    style 140 fill:#ef4444
    style 141 fill:#ef4444
    style 142 fill:#ef4444
    style 143 fill:#ef4444
    style 145 fill:#ef4444
    style 146 fill:#ef4444
    style 147 fill:#ef4444
    style 148 fill:#ef4444
```

**Parallelizable after #139:** #140, #141, #142 in parallel. After #141: #143, #145, #146 in parallel.

---

## Track H: Expanded World Generation

> Deep worldbuilding tools: languages, belief systems, economies, cultures, cities.
> Depends on: Track E (World Generation — core entity tools)

| #   | Issue                                                            | Title                                       | Size | Blocker   | Status  | Model             | Notes                   |
| --- | ---------------------------------------------------------------- | ------------------------------------------- | :--: | --------- | ------- | ----------------- | ----------------------- |
| 1   | [#155](https://github.com/PatrickFanella/game-master/issues/155) | Migration: create expanded world tables     |  S   | Phase 1   | READY   | gpt-5.3-codex     | Can start early         |
| 2   | [#156](https://github.com/PatrickFanella/game-master/issues/156) | sqlc queries: expanded world tables         |  S   | #155      | BLOCKED | gpt-5.3-codex     |                         |
| 3   | [#149](https://github.com/PatrickFanella/game-master/issues/149) | Implement create_language tool              |  M   | Track E   | BLOCKED | gpt-5.3-codex     |                         |
| 4   | [#150](https://github.com/PatrickFanella/game-master/issues/150) | Implement create_belief_system tool         |  M   | Track E   | BLOCKED | gpt-5.3-codex     |                         |
| 5   | [#151](https://github.com/PatrickFanella/game-master/issues/151) | Implement create_economic_system tool       |  M   | Track E   | BLOCKED | gpt-5.3-codex     |                         |
| 6   | [#152](https://github.com/PatrickFanella/game-master/issues/152) | Implement create_culture tool               |  M   | Track E   | BLOCKED | gpt-5.3-codex     |                         |
| 7   | [#153](https://github.com/PatrickFanella/game-master/issues/153) | Implement create_city tool                  |  M   | Track E   | BLOCKED | gpt-5.3-codex     |                         |
| 8   | [#154](https://github.com/PatrickFanella/game-master/issues/154) | Implement language naming integration       |  M   | #149      | BLOCKED | Claude Sonnet 4.6 | Uses language phonology |
| 9   | [#157](https://github.com/PatrickFanella/game-master/issues/157) | Unit tests: expanded world generation tools |  L   | #149-#153 | BLOCKED | gpt-5.3-codex     |                         |

```mermaid
graph TD
    P1[Phase 1] --> 155[#155 Migrations]
    155 --> 156[#156 sqlc queries]
    TE[Track E complete] --> 149[#149 create_language]
    TE --> 150[#150 create_belief_system]
    TE --> 151[#151 create_economic_system]
    TE --> 152[#152 create_culture]
    TE --> 153[#153 create_city]
    149 --> 154[#154 Naming integration]
    149 --> 157[#157 Unit tests]
    150 --> 157
    151 --> 157
    152 --> 157
    153 --> 157

    style 155 fill:#22c55e
    style 156 fill:#ef4444
    style 149 fill:#ef4444
    style 150 fill:#ef4444
    style 151 fill:#ef4444
    style 152 fill:#ef4444
    style 153 fill:#ef4444
    style 154 fill:#ef4444
    style 157 fill:#ef4444
```

**Note:** #155 and #156 (migrations + queries) can start in Phase 1 or Phase 2 since they only need the database. Start them early to avoid blocking the tools.

**Parallelizable after Track E:** #149-#153 all in parallel.

---

## Track I: REST API

> HTTP API wrapping the game engine for future web/mobile clients.
> Depends on: Phase 2 (Epic #6 turn pipeline)

| #   | Issue                                                            | Title                                       | Size | Blocker    | Status  | Model             | Notes                  |
| --- | ---------------------------------------------------------------- | ------------------------------------------- | :--: | ---------- | ------- | ----------------- | ---------------------- |
| 1   | [#159](https://github.com/PatrickFanella/game-master/issues/159) | Define shared API types in pkg/api          |  M   | Phase 2    | READY   | gpt-5.3-codex     |                        |
| 2   | [#158](https://github.com/PatrickFanella/game-master/issues/158) | Create cmd/server entry point + chi router  |  M   | Phase 2    | READY   | gpt-5.3-codex     |                        |
| 3   | [#166](https://github.com/PatrickFanella/game-master/issues/166) | Implement auth middleware interface (no-op) |  S   | Phase 2    | READY   | gpt-5.3-codex     |                        |
| 4   | [#160](https://github.com/PatrickFanella/game-master/issues/160) | Implement campaign REST endpoints           |  M   | #158, #159 | BLOCKED | gpt-5.3-codex     |                        |
| 5   | [#161](https://github.com/PatrickFanella/game-master/issues/161) | Implement character REST endpoints          |  S   | #158, #159 | BLOCKED | gpt-5.3-codex     |                        |
| 6   | [#162](https://github.com/PatrickFanella/game-master/issues/162) | Implement location and NPC REST endpoints   |  M   | #158, #159 | BLOCKED | gpt-5.3-codex     |                        |
| 7   | [#163](https://github.com/PatrickFanella/game-master/issues/163) | Implement quest REST endpoints              |  S   | #158, #159 | BLOCKED | gpt-5.3-codex     |                        |
| 8   | [#164](https://github.com/PatrickFanella/game-master/issues/164) | Implement POST /action endpoint             |  M   | #158       | BLOCKED | Claude Sonnet 4.6 | Core gameplay endpoint |
| 9   | [#165](https://github.com/PatrickFanella/game-master/issues/165) | Implement WebSocket streaming endpoint      |  L   | #158       | BLOCKED | Claude Sonnet 4.6 |                        |
| 10  | [#167](https://github.com/PatrickFanella/game-master/issues/167) | HTTP integration tests for API              |  L   | #160-#165  | BLOCKED | Claude Sonnet 4.6 | testcontainers         |

```mermaid
graph TD
    P2[Phase 2 complete] --> 159[#159 API types]
    P2 --> 158[#158 cmd/server + chi]
    P2 --> 166[#166 Auth middleware]
    158 --> 160[#160 Campaign endpoints]
    159 --> 160
    158 --> 161[#161 Character endpoints]
    159 --> 161
    158 --> 162[#162 Location/NPC endpoints]
    159 --> 162
    158 --> 163[#163 Quest endpoints]
    159 --> 163
    158 --> 164[#164 POST /action]
    158 --> 165[#165 WebSocket streaming]
    160 --> 167[#167 Integration tests]
    161 --> 167
    162 --> 167
    163 --> 167
    164 --> 167
    165 --> 167

    style 159 fill:#22c55e
    style 158 fill:#22c55e
    style 166 fill:#22c55e
    style 160 fill:#ef4444
    style 161 fill:#ef4444
    style 162 fill:#ef4444
    style 163 fill:#ef4444
    style 164 fill:#ef4444
    style 165 fill:#ef4444
    style 167 fill:#ef4444
```

**Parallelizable:** #158, #159, #166 start simultaneously. Then #160-#165 all in parallel.

---

## Phase 3 Execution Order

```
Sprint 1:  Alpha group — 6 tracks in parallel
           ├── Track A: Embedding (#91, #92, #94, #97 → #93, #95 → #96)
           ├── Track C: Campaign (#107, #111, #106, #108 → #109 → #110, #112)
           ├── Track D: Player (#113-#118 → #119, #120, #121)
           ├── Track E: World (#122-#127 → #128, #129, #130)
           ├── Track F: Quest (#131-#136 → #137, #138)
           └── Track I: API (#158, #159, #166 → #160-#165 → #167)
           Also: Track H #155, #156 (migrations — start early)

Sprint 2:  Beta group — 3 tracks unblocked by Alpha
           ├── Track B: Context Mgmt (#98-#103 → #101 → #104, #105)
           ├── Track G: Combat (#139 → #140-#142 → #143, #145, #146 → #147 → #148)
           └── Track H: Expanded World (#149-#153 → #154, #157)
           Gate: full game loop with memory, combat, deep worldbuilding, API
```

---

## Full Project Dependency Graph

```mermaid
graph TD
    P1A[Phase 1: Scaffold] --> P1BEF[Phase 1: DB + LLM + TUI]
    P1BEF --> P2[Phase 2: Turn Pipeline]
    P1BEF --> P2C[Phase 2: Claude Provider]
    P2 --> P3A[Track A: Embedding]
    P2 --> P3C[Track C: Campaign]
    P2 --> P3D[Track D: Player]
    P2 --> P3E[Track E: World Gen]
    P2 --> P3F[Track F: Quests]
    P2 --> P3I[Track I: API]
    P3A --> P3B[Track B: Context Mgmt]
    P3D --> P3G[Track G: Combat]
    P3E --> P3H[Track H: Expanded World]

    style P1A fill:#22c55e
    style P1BEF fill:#ef4444
    style P2 fill:#ef4444
    style P2C fill:#ef4444
    style P3A fill:#ef4444
    style P3B fill:#ef4444
    style P3C fill:#ef4444
    style P3D fill:#ef4444
    style P3E fill:#ef4444
    style P3F fill:#ef4444
    style P3G fill:#ef4444
    style P3H fill:#ef4444
    style P3I fill:#ef4444
```
