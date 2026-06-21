# Edda architecture deepening map

Wave order: 0 baseline/safety rails, 1 routes + auth, 2 API representation + frontend backend access, 3 game state persistence, 4 world creation + quest mutation, 5 narrative turn, 6 TUI play shell + campaign play session frontend, 7 cleanup/docs.

| Opportunity | Module | Seam / Adapter | Benefit | Plan wave |
| --- | --- | --- | --- | --- |
| 1 | Campaign play session frontend | `CampaignPlayPage.tsx` becomes adapter over `useCampaignPlaySession` | Deeper play-state ownership, better leverage and locality | Wave 6 |
| 2 | Quest mutation backend | `internal/game/quest_mutation.go` behind tool adapter | Quest rules centralize; deletion test moves complexity out of tool call sites | Wave 4 |
| 3 | World creation backend | `internal/world/*` orchestrator/persist seams | Startup wizard becomes deeper and more durable | Wave 4 |
| 4 | Auth session backend | `internal/auth/session.go` behind handler/middleware adapters | Session lifecycle becomes one Module with clear Interface | Wave 1 |
| 5 | Frontend backend access | `frontend/src/api/backend.ts` and `websocket.ts` with client adapters | One transport seam for fetch/WebSocket behavior | Wave 2 |
| 6 | Game state persistence | `internal/game/state_store.go` behind tool adapters | Removes raw sqlc shapes from gameplay mutations | Wave 3 |
| 7 | TUI play shell | `tui/app.go`/router become adapters over play shell | Tab lifecycle and streaming rules live in one place | Wave 6 |
| 8 | Narrative turn | `internal/game/narrative_turn.go` behind REST/WebSocket/TUI adapters | Streaming turn flow gains depth and test surface | Wave 5 |
| 9 | Route composition | `internal/server/routes.go` behind `cmd/server/main.go` | Route grouping and auth policy move behind a small interface | Wave 1 |
| 10 | API representation | `pkg/api/types.go` + converters | Response shape stability and explicit optionality/defaulting | Wave 2 |

## Baseline verification

- `task test` — run during Sprint 0.1 baseline.
- Frontend test command — none configured in `frontend/package.json` (only build/lint/dev/preview exist), so no frontend test baseline command was available.

## Sprint 0.2 blockers / gaps

- Frontend characterization tests for `frontend/src/api/client.ts` and `frontend/src/hooks/websocketParser.ts` were not added because the frontend repo has no configured test harness or `test` script in `frontend/package.json`, so they cannot be executed yet without adding a new runner.
- This missing frontend harness remains after Wave 2: backend API representation is covered by Go goldens, and frontend backend-access code has static verification, but frontend transport/parser behavior still needs an executable test runner before future frontend refactors rely on characterization coverage.

## Wave 2 verification

- API representation goldens were added for public response shapes in `pkg/api`, `internal/handlers`, and `internal/saves`.
- Frontend backend access is centralized in `frontend/src/api/backend.ts` and `frontend/src/api/websocket.ts`; raw `fetch(` usage is isolated to `backend.ts`, and `useWebSocket.ts` no longer builds URLs directly.
- `pnpm build` could not run locally because `frontend/node_modules` is missing (`tsc: command not found`).

## Wave 3 Sprint 3.1 game state persistence inventory

The Game state persistence Module should hide `statedb.Querier`, `statedb.*Params`, `pgtype`, row sequencing, and rollout-tolerant raw SQL from gameplay mutation callers. Current direct persistence knowledge groups into these domain actions:

| Domain action | Current files | Caller knowledge leaking across the Seam | Suggested migration order |
| --- | --- | --- | --- |
| Revise world fact | `internal/tools/revise_fact.go`, `internal/game/world_service.go`, `internal/world/store_adapter.go`, `queries/world_facts.sql` | Callers know `SupersedeFact` ordering, reject already-superseded facts, copy `player_known` manually, and preserve fact category/source from the old row. | 1 |
| Reveal location | `internal/tools/reveal_location.go`, `internal/game/location_service.go`, `queries/locations.sql` | Callers load a location, then separately set `player_known`; visibility is a row flag rather than a domain action. | 2 |
| Move player | `internal/tools/move_player.go`, `internal/game/location_service.go`, `internal/game/loader.go`, `queries/locations.sql`, `queries/location_connections.sql` | Callers validate current location/context IDs, connection existence, player location update, destination visited side effect, and optional campaign-time raw SQL. | 2 |
| Create/update quest and objectives | `internal/tools/create_quest.go`, `internal/tools/complete_objective.go`, `internal/tools/update_quest.go`, `internal/tools/create_subquest.go`, `internal/game/world_service.go`, `internal/handlers/quests.go`, `queries/quests.sql`, `queries/quest_objectives.sql` | Callers normalize objective ordering, select objectives by ID or description, treat already-complete objectives as non-fatal, auto-complete when all objectives complete, and cascade parent status to subquests. | 3 |
| NPCs, relationships, items, combat | `internal/game/npc_service.go`, `internal/game/combat_service.go`, `internal/game/world_service.go`, `internal/tools/update_npc.go`, `internal/tools/create_npc.go`, `internal/tools/resolve_combat.go`, relationship/faction/city/economic tools, `internal/handlers/npcs.go`, `internal/handlers/encountered.go` | Callers convert `pgx.ErrNoRows` to `nil`, infer source entities from context, create combat/session-log side effects, ensure player character exists before item creation, and manage relationship row shapes. | 4 |
| Session log | `internal/world/store_adapter.go`, `internal/game/state_manager.go`, `internal/game/npc_service.go`, `internal/game/combat_service.go`, `internal/handlers/campaigns.go`, `internal/handlers/encountered.go`, `queries/session_logs.sql` | Callers choose turn numbers, input types, empty `ToolCalls`, recent-log limits, and dialogue/combat log shapes. | 4 |
| Persist world entities | `internal/world/orchestrator.go`, `internal/world/persist.go`, `internal/world/store_adapter.go`, `internal/world/persist_adapter_orchestrator_test.go` | World creation callers know campaign creation order, skeleton persistence, starting-location name resolution, character defaults, session-log persistence, and many `pgtype` mappings. | 5 |
| Save point / campaign time | `internal/game/state_manager.go`, `internal/game/loader.go`, `internal/tools/advance_time.go`, `internal/tools/move_player.go`, `internal/handlers/startup.go`, `queries/campaign_time.sql` | Campaign-time callers still use rollout-tolerant raw SQL and tolerate missing rows/table; save/time JSON is locked separately by `internal/saves` goldens. | 6 |

Sprint 3.2 should start with world facts because `revise_fact` is already isolated and covered by tool tests, then migrate location reveal/move, then quests/objectives. NPCs/session logs/world creation have broader fan-out and should wait until the simpler Game state persistence Interface proves useful by the deletion test.

Behaviour currently protected by tests includes `internal/tools/revise_fact_test.go`, `internal/game/location_service_test.go`, `internal/tools/move_player_test.go`, `internal/tools/reveal_location_test.go`, `internal/tools/create_quest_test.go`, `internal/tools/complete_objective_test.go`, `internal/tools/update_quest_test.go`, `internal/game/npc_service_test.go`, `internal/game/combat_service_test.go`, `internal/game/state_manager_test.go`, and `internal/world/persist_adapter_orchestrator_test.go`.

## Wave 3 Sprint 3.2 verification

- Added the first Game state persistence vertical slice in `internal/game/state_store.go`: `ReviseWorldFact` is a campaign-scoped domain action that returns domain-shaped results/errors and keeps `statedb`, `pgtype`, and `pgx.ErrNoRows` inside the Implementation.
- Moved `revise_fact` to a tool Adapter over that Module. The tool still emits the same successful result shape, now accepts campaign ownership from tool context, rejects explicit `campaign_id` mismatches, and treats post-mutation memory embedding failure as a success with a warning instead of a failed retry trap.
- Hardened `SupersedeFact` in `queries/world_facts.sql`: it is campaign-scoped, locks the previous active fact with `FOR UPDATE`, re-checks `superseded_by IS NULL` during update, and carries `player_known OR reveal` into the inserted fact in the same SQL statement.
- Verification passed: `go test ./internal/state/sqlc ./internal/game ./internal/tools` and `task test`.
- `task generate` could not complete because sqlc reports an existing unrelated query issue in `queries/save_points.sql`: `column reference "campaign_id" is ambiguous`. The generated `internal/state/sqlc/world_facts.sql.go` was updated for the world-fact query shape, but full regeneration needs the save-point query fixed first.

## Wave 3 Sprint 3.3 verification

- Added campaign-scoped location actions behind the Game state persistence Module: `RevealLocation` and `MovePlayer` now live in `internal/game/state_store.go` with command/result types in `internal/domain/location_actions.go`.
- `internal/tools/reveal_location.go` and `internal/tools/move_player.go` are now tool Adapters: they parse arguments/current campaign-location-player context and format `ToolResult`; location visibility, movement validation, visited marking, and campaign-time effects live behind the StateStore Interface.
- `MovePlayer` validates player character, current location, target location, and connection within the campaign. Best-effort effects remain non-fatal and surface as `visited_warning` / `time_warning` result fields.
- Raw campaign-time SQL was moved out of `internal/tools/move_player.go` and into `internal/game/state_store.go`; engine wiring passes the existing time DB into `game.NewStateStore` when available.
- Verification passed: `go test ./internal/game ./internal/tools ./internal/engine` and `task test`.

## Wave 4 Sprint 4.1 verification

- World creation now owns startup playable-world setup: `internal/handlers/startup.go` passes `rules_mode` and DB access to `internal/world.Orchestrator`, and the handler no longer initializes campaign time, updates rules mode, or seeds crunch rules directly.
- Starting-location resolution is strict in the World creation Module: missing, empty, or duplicate starting-location names now fail before character persistence instead of creating a locationless character.
- Opening scene choices are persisted in `session_logs.tool_calls` using a stable array payload: `[{"type":"opening_choices","choices":[...]}]`, while the public BuildWorld response remains `campaign + opening_scene`.
- LLM calls remain outside any explicit DB transaction. The Sprint 4.1 slice did not introduce full transactional rollback for campaign + skeleton + character + scene persistence; that remains a deeper follow-up before treating World creation as fully atomic.
- Verification passed: `go test ./internal/world ./internal/game ./cmd/server` and `task test`.

## Wave 4 Sprint 4.2 verification

- Session history now exposes persisted opening choices through an additive `choices` field on `SessionLogEntry`; raw `tool_calls` remains hidden behind API representation conversion.
- `internal/handlers/convert.go` extracts only the stable opening-choices payload shape from `session_logs.tool_calls`: `[{"type":"opening_choices","choices":[...]}]`.
- The web narrative adapter maps history choices back into GM entries so the existing startup-seed fallback can deduplicate against persisted opening choices.
- `StartupPlaySeed` and TUI `seedOpeningScene` intentionally remain in place until both web and TUI can load narrative plus choices from persisted history without losing first-turn options.
- Verification passed: `go test ./pkg/api ./internal/handlers ./internal/world` and `task test`.

## Wave 4 Sprint 4.3 verification

- Broke the `internal/game` → `internal/tools` dependency by moving domain-shaped tool DTOs and narrow store Interfaces into `internal/domain/tool_types.go`, while leaving compatibility aliases in `internal/tools`.
- No `internal/game` file imports `internal/tools`; this removed the package-cycle blocker for deeper gameplay Modules.
- Quest mutation now lives behind a domain-shaped Interface in `internal/domain/quest_mutation.go` and a Game state persistence Implementation in `internal/game/quest_mutation.go`.
- `internal/tools/update_quest.go` and `internal/tools/complete_objective.go` are tool Adapters: they parse tool args/current campaign context, call the Quest mutation Module, and format the existing `ToolResult` shapes.
- Quest mutation enforces campaign scoping, preserves objective append ordering, recursive active-subquest cascades for completed/failed quest status, normalized objective selection, already-complete no-op behavior, and auto-complete semantics. Auto-complete now uses the same completed-quest cascade path as `update_quest`.
- Verification passed: `go test ./internal/game ./internal/tools ./internal/engine`, `task test`, and a grep confirmed no `internal/game` imports of `internal/tools`.

## Notes

- Use the architecture vocabulary exactly in later review notes and docs: Module, Interface, Implementation, Depth, Deep/Shallow, Seam, Adapter, Leverage, Locality, deletion test, interface as test surface, one adapter hypothetical/two adapters real.
