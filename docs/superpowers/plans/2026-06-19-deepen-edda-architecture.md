# Deepen Edda Architecture Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Implement all 10 deepening opportunities from the architecture review while improving testability, AI-navigability, leverage, and locality.

**Architecture:** Deepen shared seams first, then persistence and gameplay modules, then client play shells. Each wave preserves existing behaviour while moving row-shaped, transport-shaped, and page-shaped interfaces behind deeper modules. UI and TUI become adapters over deeper play and narrative modules rather than reimplementing the same behaviour.

**Tech Stack:** Go 1.24, chi, pgx/sqlc, PostgreSQL, Bubble Tea TUI, React/Vite/TypeScript, TanStack Query, WebSocket, Docker/Taskfile.

---

## Vocabulary rules

Use these terms in issue descriptions, review notes, and code comments where architecture is discussed:

- **Module** — anything with an interface and an implementation.
- **Interface** — everything a caller must know: types, invariants, error modes, ordering, config.
- **Implementation** — the code inside.
- **Depth** — leverage at the interface.
- **Deep** — high leverage behind a small interface.
- **Shallow** — interface nearly as complex as implementation.
- **Seam** — where an interface lives.
- **Adapter** — concrete thing satisfying an interface at a seam.
- **Leverage** — what callers get from depth.
- **Locality** — what maintainers get from depth.

Avoid replacing these with “component,” “service,” “API,” or “boundary” in architecture-facing artifacts.

---

## Program order

Do not start with the large frontend `CampaignPlayPage` extraction. That would preserve today’s shallow backend and transport seams. The safe order is:

1. Route composition + Auth session
2. API representation + Frontend backend access
3. Game state persistence
4. World creation + Quest mutation
5. Narrative turn
6. TUI play shell + Campaign play session frontend

Each wave should end with all relevant tests passing and a review-quality pass.

---

## Wave 0: Baseline and safety rails

**Purpose:** Freeze current behaviour with tests and documentation before refactoring.

**Files:**

- Modify: `README.md`
- Create: `CONTEXT.md`
- Create: `docs/architecture/deepening-map.md`
- Review: `Taskfile.yml`
- Review: `cmd/server/main_test.go`
- Review: `internal/tools/*_test.go`
- Review: `tui/*_test.go`
- Review: `frontend/package.json`

### Sprint 0.1: Establish domain context

- [ ] Create `CONTEXT.md` with the domain terms already used in the codebase: Campaign, Startup wizard, World creation, Campaign play, Narrative turn, Quest, Objective, World fact, Location, NPC, Save point, Campaign time, Rules mode, Combat, Session log.
- [ ] Add `docs/architecture/deepening-map.md` listing the 10 target modules and the intended wave order.
- [ ] Fix the visible merge-conflict markers in `README.md:70-93` before using README as a reference document.
- [ ] Run `task test`.
- [ ] If the frontend has a test command in `frontend/package.json`, run it and record the baseline result in `docs/architecture/deepening-map.md`.
- [ ] Commit with message: `docs: map architecture deepening plan`.

### Sprint 0.2: Add characterization tests where coverage is missing

- [ ] Add route characterization tests for public health, API health, auth-enabled routes, and auth-disabled routes in `cmd/server/main_test.go` or a new focused route test file.
- [ ] Add frontend tests for `frontend/src/api/client.ts` if no test harness exists; otherwise document the missing harness as a blocker for Wave 2.
- [ ] Add one WebSocket event parser test around `decodeWebSocketEvent` by moving parser logic to a testable module if needed.
- [ ] Run backend tests: `task test`.
- [ ] Run frontend tests if configured.
- [ ] Commit with message: `test: characterize current architecture seams`.

**Wave 0 gate:** Current behaviour is documented, tests pass or known test-harness gaps are recorded, and the plan has a project-local context file.

---

## Wave 1: Route composition and Auth session

Implements opportunities 9 and 4.

### Target depth

- Route composition becomes a deeper module that owns route grouping, auth mode selection, middleware policy, and handler wiring.
- Auth session becomes a deeper module that owns register, login, current-user lookup, logout, token/cookie issuance, token extraction, and authenticated-user derivation.

### Files

- Modify: `cmd/server/main.go`
- Create: `internal/server/routes.go`
- Create: `internal/server/routes_test.go`
- Modify: `internal/auth/handlers.go`
- Modify: `internal/auth/middleware.go`
- Modify: `internal/auth/db_adapter.go`
- Create: `internal/auth/session.go`
- Create: `internal/auth/session_test.go`
- Modify: `frontend/src/context/AuthContext.tsx`
- Modify: `frontend/src/api/auth.ts`
- Modify: `frontend/src/api/client.ts`
- Update docs: `docs/ios-api-handoff.md`

### Sprint 1.1: Extract route composition without behaviour change

- [ ] Move `newRouterWithProvider` and `registerAPIRoutes` behaviour from `cmd/server/main.go` into `internal/server/routes.go`.
- [ ] Keep the existing exported paths unchanged: `/healthz`, `/api/healthz`, `/api/v1/auth/*`, `/api/v1/campaigns/*`.
- [ ] Keep `cmd/server/main.go` as the entrypoint and dependency assembly only.
- [ ] Add tests in `internal/server/routes_test.go` that verify route presence and auth mode behaviour.
- [ ] Run `go test ./cmd/server ./internal/server ./internal/auth`.
- [ ] Commit with message: `refactor: deepen server route composition`.

### Sprint 1.2: Deepen Auth session backend

- [ ] Introduce an Auth session module in `internal/auth/session.go` that owns session lifecycle rules.
- [ ] Move token generation, cookie setting/clearing, current-user lookup, and request-token extraction behind this module.
- [ ] Keep `AuthHandlers` as HTTP adapters that translate requests/responses.
- [ ] Keep `JWTMiddleware` and `NoOpMiddleware` as adapters over the session decision path.
- [ ] Remove duplicate auth DB shapes where possible; if sqlc cannot satisfy the interface directly, keep one small DB adapter in `internal/auth/db_adapter.go`.
- [ ] Add tests in `internal/auth/session_test.go` for register/login/current-user/logout decision paths, invalid token, missing token, malformed auth header, WebSocket cookie token, and no-op mode.
- [ ] Run `go test ./internal/auth ./internal/server ./cmd/server`.
- [ ] Commit with message: `refactor: deepen auth session module`.

### Sprint 1.3: Align frontend auth adapter

- [ ] Update `frontend/src/api/auth.ts` so register/login/me/logout map to the backend Auth session semantics.
- [ ] Update `frontend/src/context/AuthContext.tsx` so session restoration, logout, and token persistence are concentrated there or in a new session helper.
- [ ] Keep browser token/cookie behaviour compatible with `docs/ios-api-handoff.md`.
- [ ] Add or update frontend auth tests.
- [ ] Run backend auth tests and frontend tests.
- [ ] Commit with message: `refactor: align frontend auth session adapter`.

**Wave 1 gate:** Protected routes reject unauthenticated users, auth-enabled and no-op modes both work, WebSocket cookie auth still works, and frontend session restoration still works.

---

## Wave 2: API representation and Frontend backend access

Implements opportunities 10 and 5.

### Target depth

- API representation owns exposed fields, defaulting, optionality, JSON shape stability, and conversion from persistence rows.
- Frontend backend access owns HTTP/WebSocket transport, auth credential attachment, backend URL resolution, error normalization, and conversion into frontend-facing read models.

### Files

- Modify: `pkg/api/types.go`
- Modify: `pkg/api/types_test.go`
- Modify: `internal/handlers/convert.go`
- Create: `internal/handlers/convert_test.go`
- Modify: `frontend/src/api/types.ts`
- Modify: `frontend/src/api/client.ts`
- Create: `frontend/src/api/backend.ts`
- Create: `frontend/src/api/websocket.ts`
- Modify: `frontend/src/api/*.ts`
- Modify: `frontend/src/hooks/useWebSocket.ts`
- Update: `docs/ios-api-handoff.md`

### Sprint 2.1: Add API representation contract tests

- [ ] Add tests for conversion of Campaign, Character, Quest, Objective, World fact, Location, NPC, Save point, Campaign time, and Turn response representations.
- [ ] Add JSON fixture tests that compare Go response shapes against checked-in golden JSON.
- [ ] Run `go test ./pkg/api ./internal/handlers`.
- [ ] Commit with message: `test: lock api representation shapes`.

### Sprint 2.2: Deepen backend representation module

- [ ] Move scattered row-to-response policy into focused conversion functions in `internal/handlers/convert.go` or split by representation area if the file becomes hard to navigate.
- [ ] Ensure defaulting and optionality rules are tested in `internal/handlers/convert_test.go`.
- [ ] Keep route handlers thin: load state, call representation conversion, write JSON.
- [ ] Update `pkg/api/types.go` comments for invariants callers must know.
- [ ] Run `go test ./pkg/api ./internal/handlers ./cmd/server`.
- [ ] Commit with message: `refactor: deepen api representation module`.

### Sprint 2.3: Deepen frontend backend access module

- [ ] Create `frontend/src/api/backend.ts` for HTTP transport, base URL resolution, JSON parsing, auth credential attachment, and error normalization.
- [ ] Create `frontend/src/api/websocket.ts` for WebSocket URL construction and event decoding.
- [ ] Update `frontend/src/api/client.ts` to become a compatibility adapter or remove it after callers migrate.
- [ ] Update `frontend/src/hooks/useWebSocket.ts` to consume the WebSocket backend access module instead of building URLs and decoding events internally.
- [ ] Update all `frontend/src/api/*.ts` files to use the one backend access seam.
- [ ] Add frontend tests for HTTP error normalization, auth header attachment, WebSocket URL construction, and event decoding.
- [ ] Run frontend tests and `go test ./...`.
- [ ] Commit with message: `refactor: deepen frontend backend access module`.

**Wave 2 gate:** No new raw `fetch` usage outside the backend access module, WebSocket URL construction exists in one place, and API representation tests protect web/iOS shape stability.

---

## Wave 3: Game state persistence

Implements opportunity 6.

### Target depth

Game state persistence owns domain actions and invariants instead of exposing `statedb.Querier`, `statedb.*Params`, `pgtype`, and row sequencing to tool handlers.

### Files

- Modify: `internal/game/world_service.go`
- Modify: `internal/game/location_service.go`
- Create: `internal/game/state_store.go`
- Create: `internal/game/state_store_test.go`
- Modify: `internal/tools/revise_fact.go`
- Modify: `internal/tools/reveal_location.go`
- Modify: `internal/tools/move_player.go`
- Modify: `internal/tools/update_quest.go`
- Modify: `queries/*.sql` only when a deeper operation belongs in SQL
- Review: `internal/state/sqlc/querier.go`

### Sprint 3.1: Define persistence action inventory

- [ ] List every direct `statedb.Querier` use in `internal/tools`, `internal/world`, `internal/handlers`, and `internal/game`.
- [ ] Group uses into domain actions: revise fact, reveal location, move player, create/update quest, update relationship, persist world entities, create session log, create save point, advance campaign time.
- [ ] Record the grouping in `docs/architecture/deepening-map.md`.
- [ ] Commit with message: `docs: inventory game state persistence seams`.

### Sprint 3.2: Introduce Game state persistence module

- [ ] Create `internal/game/state_store.go` with domain-action methods and keep sqlc behind the implementation.
- [ ] Add unit tests with fake adapters for visibility rules, supersession rules, ordering rules, and campaign ownership checks.
- [ ] Keep existing callers working by migrating one action at a time.
- [ ] Run `go test ./internal/game`.
- [ ] Commit with message: `refactor: add game state persistence module`.

### Sprint 3.3: Migrate high-churn tools to Game state persistence

- [ ] Migrate `internal/tools/revise_fact.go` so fact supersession and player-known rules live in the Game state persistence module.
- [ ] Migrate `internal/tools/reveal_location.go` and `internal/tools/move_player.go` so location visibility, visited state, travel effects, and campaign-time effects live behind the same module.
- [ ] Keep tool modules as adapters that parse LLM arguments and render tool results.
- [ ] Update tests for each tool to assert outcomes through fake Game state persistence adapters.
- [ ] Run `go test ./internal/tools ./internal/game`.
- [ ] Commit with message: `refactor: move world tool persistence behind game state module`.

**Wave 3 gate:** New gameplay mutations do not call sqlc directly unless they are inside the Game state persistence implementation.

---

## Wave 4: World creation and Quest mutation

Implements opportunities 3 and 2.

### Target depth

- World creation owns the full new-campaign journey: profile + character + selected proposal/name in, persisted playable world + opening scene out.
- Quest mutation owns quest status transitions, description changes, objective appends, subquest cascades, and result shaping.

### Files

- Modify: `internal/world/orchestrator.go`
- Modify: `internal/world/skeleton.go`
- Modify: `internal/world/scene.go`
- Modify: `internal/world/persist.go`
- Modify: `internal/world/store_adapter.go`
- Modify: `internal/world/*_test.go`
- Modify: `internal/tools/update_quest.go`
- Create: `internal/game/quest_mutation.go`
- Create: `internal/game/quest_mutation_test.go`
- Modify: `frontend/src/lib/startupWorkflow.ts`
- Modify: `frontend/src/pages/CampaignCreatePage.tsx`
- Modify: `tui/campaign/*.go`

### Sprint 4.1: Deepen World creation backend

- [ ] Move campaign row creation, skeleton persistence, starting-location resolution, character persistence, opening-scene generation, and session-log persistence behind the World creation module.
- [ ] Ensure opening scene and suggested choices are persisted so refresh/resume does not depend on frontend route state.
- [ ] Add rollback or transaction tests for failures at each stage.
- [ ] Keep progress reporting as an adapter concern: CLI/TUI/web can observe stages but should not own creation order.
- [ ] Run `go test ./internal/world ./internal/game ./cmd/server`.
- [ ] Commit with message: `refactor: deepen world creation module`.

### Sprint 4.2: Align web and TUI startup adapters

- [ ] Update `frontend/src/lib/startupWorkflow.ts` and `frontend/src/pages/CampaignCreatePage.tsx` to consume the deeper World creation result.
- [ ] Update `tui/campaign/*.go` to consume the same world creation behaviour.
- [ ] Remove duplicated startup seed logic where persisted opening scene can now be loaded normally.
- [ ] Update `docs/ios-api-handoff.md` for any startup response changes.
- [ ] Run backend tests, frontend tests, and TUI tests.
- [ ] Commit with message: `refactor: align startup adapters with world creation module`.

### Sprint 4.3: Deepen Quest mutation backend

- [ ] Create `internal/game/quest_mutation.go` for quest mutation rules.
- [ ] Move status transition, description replacement, objective ordering, and subquest cascade logic out of `internal/tools/update_quest.go`.
- [ ] Keep `UpdateQuestHandler` as an adapter that parses tool arguments and formats `ToolResult`.
- [ ] Add tests for active → completed, active → failed, completed parent cascades active children to abandoned, failed parent cascades active children to failed, objective append ordering, and no-op rejection.
- [ ] Run `go test ./internal/game ./internal/tools`.
- [ ] Commit with message: `refactor: deepen quest mutation module`.

**Wave 4 gate:** Campaign creation persists a playable world durably, and quest mutation tests verify outcomes instead of sqlc call sequencing.

---

## Wave 5: Narrative turn

Implements opportunity 8.

### Target depth

Narrative turn owns the lifecycle from player action to streaming GM response to final choices/status. Frontend, WebSocket, REST fallback, and TUI become adapters over this behaviour.

### Files

- Review/modify: `internal/engine/*`
- Review/modify: `internal/handlers/actions.go`
- Review/modify: `internal/handlers/websocket*.go`
- Create: `internal/game/narrative_turn.go`
- Create: `internal/game/narrative_turn_test.go`
- Modify: `frontend/src/hooks/useNarrative.ts`
- Modify: `frontend/src/hooks/narrativeReducer.ts`
- Modify: `frontend/src/hooks/useWebSocket.ts`
- Modify: `tui/app.go`
- Modify: `tui/narrative/narrative.go`

### Sprint 5.1: Locate and characterize current turn execution

- [ ] Map REST action, WebSocket action, engine processing, tool execution, persistence, streaming chunk, result event, and suggested choice flow.
- [ ] Add characterization tests around non-streaming and streaming turn results.
- [ ] Add a concurrency test for simultaneous streamed turns if the current implementation allows it.
- [ ] Run `go test ./internal/engine ./internal/handlers ./internal/game`.
- [ ] Commit with message: `test: characterize narrative turn flow`.

### Sprint 5.2: Introduce Narrative turn module

- [ ] Create `internal/game/narrative_turn.go` to own per-turn context, tool execution, stream event production, result persistence, and final choices.
- [ ] Ensure per-turn stream state is local to the turn implementation, not shared across requests.
- [ ] Keep REST and WebSocket handlers as adapters over Narrative turn.
- [ ] Add unit tests with fake LLM and fake Game state persistence adapters.
- [ ] Run `go test ./internal/game ./internal/handlers ./internal/engine`.
- [ ] Commit with message: `refactor: deepen narrative turn module`.

### Sprint 5.3: Align client narrative adapters

- [ ] Update `frontend/src/hooks/useNarrative.ts` and `frontend/src/hooks/narrativeReducer.ts` to consume normalized Narrative turn events from the backend access module.
- [ ] Update `tui/app.go` and `tui/narrative/narrative.go` so TUI turn streaming follows the same lifecycle concepts as the backend Narrative turn module.
- [ ] Add frontend reducer tests for action sent, chunk received, result received, error received, connection lost, and choices updated.
- [ ] Run frontend tests and TUI tests.
- [ ] Commit with message: `refactor: align narrative adapters with turn module`.

**Wave 5 gate:** Streaming and non-streaming narrative turns share one core implementation, concurrent turns do not race through shared turn state, and clients consume normalized turn events.

---

## Wave 6: TUI play shell and Campaign play session frontend

Implements opportunities 7 and 1.

### Target depth

- TUI play shell owns tab lifecycle, global key handling, active-view routing, size propagation, and narrative streaming.
- Campaign play session frontend owns campaign play state, startup seed merging, narrative turn flow, combat presentation decisions, refresh-after-turn behaviour, save/start-over transitions, and panel update signals.

### Files

- Modify: `tui/app.go`
- Modify: `tui/router.go`
- Modify: `tui/view.go`
- Modify: `tui/narrative/narrative.go`
- Modify: `tui/*_test.go`
- Modify: `frontend/src/pages/CampaignPlayPage.tsx`
- Create: `frontend/src/play/useCampaignPlaySession.ts`
- Create: `frontend/src/play/CampaignPlayShell.tsx`
- Create: `frontend/src/play/PlayTabContent.tsx`
- Create: `frontend/src/play/NarrativeTab.tsx`
- Create: `frontend/src/play/CampaignPlayActions.tsx`
- Modify: `frontend/src/hooks/useRefreshAfterTurn.ts`
- Modify: `frontend/src/hooks/useDiscoveryBadge.ts`

### Sprint 6.1: Deepen TUI play shell

- [ ] Move tab lifecycle, active-view selection, size propagation, global key conflict policy, and narrative streaming coordination into a focused TUI play shell module.
- [ ] Stop direct mutation of router internals from `tui/app.go`.
- [ ] Keep individual TUI views focused on their own display/input rules.
- [ ] Update tests in `tui/app_test.go`, `tui/router_test.go`, and `tui/narrative/narrative_test.go`.
- [ ] Run `go test ./tui/...`.
- [ ] Commit with message: `refactor: deepen tui play shell module`.

### Sprint 6.2: Extract Campaign play session state

- [ ] Create `frontend/src/play/useCampaignPlaySession.ts` to own campaign state, active tab state, startup seed merging, narrative state, combat state, saves visibility, export/start-over dialog state, level-up banner state, discovery badges, and refresh-after-turn behaviour.
- [ ] Keep `CampaignPlayPage.tsx` as route parameter loading and high-level composition only.
- [ ] Add tests for startup seed merge, existing opening scene detection, level-up message derivation, world unread badge clearing, and start-over success invalidation.
- [ ] Run frontend tests.
- [ ] Commit with message: `refactor: extract campaign play session module`.

### Sprint 6.3: Extract Campaign play rendering adapters

- [ ] Create `frontend/src/play/CampaignPlayShell.tsx` for shell composition.
- [ ] Create `frontend/src/play/PlayTabContent.tsx` for tab dispatch.
- [ ] Create `frontend/src/play/NarrativeTab.tsx` for narrative/combat presentation.
- [ ] Create `frontend/src/play/CampaignPlayActions.tsx` for save/export/start-over actions.
- [ ] Reduce `frontend/src/pages/CampaignPlayPage.tsx` to route loading, error/loading state, and shell invocation.
- [ ] Add render tests for narrative mode, combat mode, saves toggle, and tab dispatch.
- [ ] Run frontend tests and a browser smoke test if available.
- [ ] Commit with message: `refactor: deepen campaign play frontend module`.

**Wave 6 gate:** `CampaignPlayPage.tsx` is no longer the owner of play-session rules, TUI and web both use deeper narrative/play concepts, and behaviour remains unchanged from the user’s perspective.

---

## Wave 7: Cleanup, documentation, and deletion tests

**Purpose:** Remove leftover shallow modules and verify the new depth actually improves leverage and locality.

**Files:**

- Review: `internal/game/*`
- Review: `internal/world/*`
- Review: `internal/tools/*`
- Review: `internal/handlers/*`
- Review: `frontend/src/pages/CampaignPlayPage.tsx`
- Review: `frontend/src/hooks/*`
- Review: `frontend/src/api/*`
- Review: `tui/*`
- Update: `CONTEXT.md`
- Update: `docs/architecture/deepening-map.md`
- Update: `docs/ios-api-handoff.md`

### Sprint 7.1: Run deletion tests on the new modules

- [ ] For each new/deepened module, write a short deletion-test note in `docs/architecture/deepening-map.md`: if deleted, where would complexity reappear?
- [ ] Remove adapters that still have only one fake use and no real variation unless they are test adapters with clear leverage.
- [ ] Remove pass-through functions that no longer concentrate behaviour.
- [ ] Commit with message: `refactor: remove shallow architecture leftovers`.

### Sprint 7.2: Update docs and client handoff

- [ ] Update `CONTEXT.md` with any new canonical domain terms created during implementation.
- [ ] Update `docs/ios-api-handoff.md` to match the final Auth session, API representation, World creation, Narrative turn, and Save/Resume behaviour.
- [ ] Update `README.md` architecture layout if new packages were introduced.
- [ ] Run doc consistency review.
- [ ] Commit with message: `docs: update architecture and api handoff`.

### Sprint 7.3: Final verification

- [ ] Run `task test`.
- [ ] Run frontend tests.
- [ ] Run TUI tests: `go test ./tui/...`.
- [ ] Run server tests: `go test ./cmd/server ./internal/server ./internal/auth ./internal/handlers`.
- [ ] Run a smoke flow: register/login, create campaign, build world, play first turn, mutate quest, save, reload/resume.
- [ ] Run `review-quality` on the full diff.
- [ ] Commit with message: `chore: verify architecture deepening program`.

**Wave 7 gate:** All 10 opportunities are implemented, tests pass, docs match code, and deletion-test notes show that new modules concentrate complexity rather than merely moving files.

---

## Cross-wave testing policy

Use this minimum suite before merging each wave:

```bash
task test
go test ./tui/...
```

If frontend tests exist, also run the project’s frontend test command from `frontend/package.json`.

For waves touching frontend runtime behaviour, also run one manual or automated smoke flow:

1. Register or login.
2. Create a campaign through the startup wizard.
3. Enter the play view.
4. Send one narrative action.
5. Verify streamed or final GM response appears.
6. Verify quests/world panels refresh after the turn.
7. Save and reload/resume.

---

## Review gates by risk domain

- Auth/session or route policy: use a security-focused review before merge.
- Persistence changes: use a data/integrity review before merge.
- Frontend play/session extraction: use TypeScript/React review before merge.
- TUI shell changes: use Go/TUI review before merge.
- Final full-program review: use `review-quality` and an architecture simplification pass.

---

## Suggested branch strategy

Use one branch per wave:

- `architecture/wave-0-baseline`
- `architecture/wave-1-routes-auth`
- `architecture/wave-2-api-access`
- `architecture/wave-3-game-state`
- `architecture/wave-4-world-quests`
- `architecture/wave-5-narrative-turn`
- `architecture/wave-6-play-shells`
- `architecture/wave-7-cleanup-docs`

Merge waves in order. Do not parallelize waves that depend on previous seams. Within a wave, parallelize only after file ownership is explicit.

---

## Completion checklist

- [ ] Opportunity 1: Campaign play session frontend module implemented.
- [ ] Opportunity 2: Quest mutation backend module implemented.
- [ ] Opportunity 3: World creation module implemented.
- [ ] Opportunity 4: Auth session module implemented.
- [ ] Opportunity 5: Frontend backend access module implemented.
- [ ] Opportunity 6: Game state persistence module implemented.
- [ ] Opportunity 7: TUI play shell module implemented.
- [ ] Opportunity 8: Narrative turn module implemented.
- [ ] Opportunity 9: Route composition module implemented.
- [ ] Opportunity 10: API representation module implemented.
- [ ] `CONTEXT.md` updated with all new terms.
- [ ] `docs/ios-api-handoff.md` matches final behaviour.
- [ ] Deletion-test notes recorded for all deepened modules.
- [ ] Full backend, frontend, TUI, and smoke tests pass.
