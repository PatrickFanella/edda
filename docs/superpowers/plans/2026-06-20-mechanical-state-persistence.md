# Mechanical State Persistence Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Make Edda persist mechanical campaign consequences whenever the GM narrates durable events: quests, NPCs, world facts, locations, player status/HP, and combat state.

**Architecture:** Keep the current tool-driven authority model: durable state changes must come from tool calls, not free prose. Fix the seams that prevent tools from being available/logged/reported, then add a small deterministic state-change projection from applied tool results for API/UI visibility. Avoid a broad narrative parser until tool-call persistence is reliable.

**Tech Stack:** Go, existing `internal/engine`, `internal/tools`, `internal/game`, sqlc Postgres stores, `rtk go test`.

---

## File Structure

- Modify `internal/engine/tool_filter.go` — allow first quest creation and player status updates in exploration.
- Modify `internal/engine/tool_filter_test.go` — update filter expectations and guard tool-count budget.
- Create `internal/engine/state_changes.go` — project `AppliedToolCall` result JSON into `[]StateChange`.
- Create `internal/engine/state_changes_test.go` — unit-test projection for NPCs, facts, quests, movement, status, HP, combat.
- Modify `internal/engine/runtime.go` — populate `TurnResult.StateChanges` from applied tools.
- Modify `internal/tools/establish_fact.go` — make memory embedding best-effort after the fact is already persisted.
- Modify `internal/tools/establish_fact_test.go` — prove embedding failure does not create a failed/duplicate tool turn.
- Add `internal/tools/update_player_hp.go` and `internal/tools/update_player_hp_test.go` — provide a non-combat HP persistence tool for hazards.
- Modify `internal/engine/tools.go` — register `update_player_hp` as a base/progression-safe tool.
- Modify `internal/prompt/gamemaster.txt` and tests — make the GM explicitly call tools before claiming durable changes.
- Add or update local diagnostic smoke script/docs after implementation.

---

## Task 1: Fix Tool Availability for First Quest and Exploration Status

**Why:** `create_quest` is currently hidden until a quest/NPC already exists, which blocks the first quest. `update_player_status` is effectively hidden in normal exploration unless near level-up or combat, which blocks poison/cursed/resting/etc. consequences from hazards.

**Files:**
- Modify: `internal/engine/tool_filter.go:123-143`
- Modify: `internal/engine/tool_filter_test.go:273-314`

- [ ] **Step 1: Write failing tests**

In `internal/engine/tool_filter_test.go`, replace `TestFilter_NoQuestToolsWhenNoQuestsOrNPCs` with:

```go
func TestFilter_QuestCreationAvailableWithoutExistingQuestOrNPC(t *testing.T) {
	f := NewPhaseToolFilter(testRegistry())
	state := &game.GameState{
		Player:    domain.PlayerCharacter{Status: "active"},
		RulesMode: "light",
	}
	got := f.Filter(state, allTestTools())

	hasAll(t, got, "create_quest")
	hasNone(t, got, "update_quest", "complete_objective", "branch_quest", "link_quest_entity")
}

func TestFilter_PlayerStatusAvailableDuringExploration(t *testing.T) {
	f := NewPhaseToolFilter(testRegistry())
	state := &game.GameState{
		Player:    domain.PlayerCharacter{Status: "active", Level: 1, Experience: 0},
		RulesMode: "light",
	}
	got := f.Filter(state, allTestTools())

	hasAll(t, got, "update_player_status")
}
```

- [ ] **Step 2: Run the failing filter tests**

Run:

```bash
rtk go test ./internal/engine -run 'TestFilter_(QuestCreationAvailableWithoutExistingQuestOrNPC|PlayerStatusAvailableDuringExploration)' -count=1
```

Expected: both fail before implementation.

- [ ] **Step 3: Implement minimal filter change**

In `internal/engine/tool_filter.go`, change quest/progression filtering to:

```go
		case tools.CategoryQuest:
			if tool.Name == "create_quest" {
				allowed[tool.Name] = struct{}{}
				continue
			}
			if len(state.ActiveQuests) > 0 || len(state.NearbyNPCs) > 0 {
				allowed[tool.Name] = struct{}{}
			}
```

Then replace the special-case block at lines 135-143 with:

```go
	// Handle tools with dual category membership.
	if phase == PhaseCombat {
		allowed["initiate_combat"] = struct{}{}
	}
	allowed["update_player_status"] = struct{}{}
	if nearLevelThreshold(state) {
		allowed["update_player_stats"] = struct{}{}
	}
```

- [ ] **Step 4: Verify**

Run:

```bash
rtk go test ./internal/engine -run TestFilter -count=1
```

Expected: all filter tests pass and `TestFilter_ReducesToolCount` remains under its budget.

---

## Task 2: Add API-Visible State Changes from Applied Tools

**Why:** State may persist, but API `state_changes` stays empty because `TurnResult.StateChanges` is never populated.

**Files:**
- Create: `internal/engine/state_changes.go`
- Create: `internal/engine/state_changes_test.go`
- Modify: `internal/engine/runtime.go:137-142`

- [ ] **Step 1: Create failing projection tests**

Create `internal/engine/state_changes_test.go`:

```go
package engine

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func applied(tool string, result string) AppliedToolCall {
	return AppliedToolCall{Tool: tool, Result: json.RawMessage(result)}
}

func TestStateChangesFromAppliedToolCalls_CreateNPC(t *testing.T) {
	id := uuid.New()
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("create_npc", `{"id":"`+id.String()+`","name":"Seraphine Maw"}`),
	})
	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if changes[0].Entity != "npc" || changes[0].EntityID != id || changes[0].Field != "created" {
		t.Fatalf("change = %#v", changes[0])
	}
}

func TestStateChangesFromAppliedToolCalls_CreateQuest(t *testing.T) {
	id := uuid.New()
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("create_quest", `{"id":"`+id.String()+`","title":"Escape the Hollow Spire"}`),
	})
	if len(changes) != 1 || changes[0].Entity != "quest" || changes[0].Field != "created" {
		t.Fatalf("changes = %#v", changes)
	}
}

func TestStateChangesFromAppliedToolCalls_MovePlayer(t *testing.T) {
	locID := uuid.New()
	pcID := uuid.New()
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("move_player", `{"location_id":"`+locID.String()+`","player_character_id":"`+pcID.String()+`"}`),
	})
	if len(changes) != 1 || changes[0].Entity != "character" || changes[0].EntityID != pcID || changes[0].Field != "location_updated" {
		t.Fatalf("changes = %#v", changes)
	}
}

func TestStateChangesFromAppliedToolCalls_UnknownOrMalformedSkipped(t *testing.T) {
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("unknown_tool", `{}`),
		applied("create_npc", `{bad json`),
	})
	if len(changes) != 0 {
		t.Fatalf("changes = %#v, want none", changes)
	}
}
```

- [ ] **Step 2: Run failing projection tests**

```bash
rtk go test ./internal/engine -run TestStateChangesFromAppliedToolCalls -count=1
```

Expected: compile failure because `StateChangesFromAppliedToolCalls` does not exist.

- [ ] **Step 3: Implement projection helper**

Create `internal/engine/state_changes.go`:

```go
package engine

import (
	"encoding/json"

	"github.com/google/uuid"
)

func StateChangesFromAppliedToolCalls(applied []AppliedToolCall) []StateChange {
	changes := make([]StateChange, 0, len(applied))
	for _, call := range applied {
		change, ok := stateChangeFromAppliedToolCall(call)
		if ok {
			changes = append(changes, change)
		}
	}
	return changes
}

func stateChangeFromAppliedToolCall(call AppliedToolCall) (StateChange, bool) {
	var data map[string]any
	if err := json.Unmarshal(call.Result, &data); err != nil {
		return StateChange{}, false
	}

	switch call.Tool {
	case "create_npc":
		return entityCreated("npc", data)
	case "establish_fact":
		return entityCreated("world_fact", data)
	case "create_quest":
		return entityCreated("quest", data)
	case "create_location", "reveal_location":
		return entityCreated("location", data)
	case "move_player":
		return characterLocationUpdated(data)
	case "update_player_status":
		return characterUpdated("status_updated", data)
	case "update_player_hp":
		return characterUpdated("hp_updated", data)
	case "initiate_combat":
		return combatUpdated(data)
	default:
		return StateChange{}, false
	}
}

func entityCreated(entity string, data map[string]any) (StateChange, bool) {
	id, ok := uuidField(data, "id")
	if !ok {
		return StateChange{}, false
	}
	return StateChange{Entity: entity, EntityID: id, Field: "created", NewValue: mustJSON(data)}, true
}

func characterLocationUpdated(data map[string]any) (StateChange, bool) {
	id, ok := uuidField(data, "player_character_id")
	if !ok {
		return StateChange{}, false
	}
	return StateChange{Entity: "character", EntityID: id, Field: "location_updated", NewValue: mustJSON(data)}, true
}

func characterUpdated(field string, data map[string]any) (StateChange, bool) {
	id, ok := uuidField(data, "player_character_id")
	if !ok {
		return StateChange{}, false
	}
	return StateChange{Entity: "character", EntityID: id, Field: field, NewValue: mustJSON(data)}, true
}

func combatUpdated(data map[string]any) (StateChange, bool) {
	id, ok := uuidField(data, "combat_id")
	if !ok {
		if fallback, fallbackOK := uuidField(data, "campaign_id"); fallbackOK {
			id = fallback
			ok = true
		}
	}
	if !ok {
		return StateChange{}, false
	}
	return StateChange{Entity: "combat", EntityID: id, Field: "started", NewValue: mustJSON(data)}, true
}

func uuidField(data map[string]any, key string) (uuid.UUID, bool) {
	value, ok := data[key].(string)
	if !ok || value == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func mustJSON(value any) json.RawMessage {
	buf, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return json.RawMessage(buf)
}
```

- [ ] **Step 4: Wire into runtime**

In `internal/engine/runtime.go`, change result construction to:

```go
	result := &TurnResult{
		Narrative:        tc.Narrative,
		AppliedToolCalls: tc.Applied,
		Choices:          tc.Choices,
		StateChanges:     StateChangesFromAppliedToolCalls(tc.Applied),
		CombatActive:     tc.CombatActive,
	}
```

- [ ] **Step 5: Verify**

```bash
rtk go test ./internal/engine -run 'TestStateChangesFromAppliedToolCalls|TestTurnProcessor|TestFilter' -count=1
```

Expected: pass.

---

## Task 3: Make Post-Persist Embedding Best-Effort

**Why:** `establish_fact` persisted the fact, then failed while embedding memory, so the turn processor treated the tool as failed and omitted it from the applied tool log. A post-write side effect must not turn a durable write into a failed tool call.

**Files:**
- Modify: `internal/tools/establish_fact.go:130-133`
- Modify: `internal/tools/establish_fact_test.go`

- [ ] **Step 1: Write failing test**

Add a test in `internal/tools/establish_fact_test.go` using the existing stubs. The test should create a handler with a working fact store, a memory store, and an embedder that returns an error, then assert `Handle` returns success and includes `memory_warning` in data.

Core assertion:

```go
result, err := handler.Handle(ctx, args)
if err != nil {
	t.Fatalf("Handle returned error after fact persisted: %v", err)
}
if result == nil || !result.Success {
	t.Fatalf("result = %#v, want success", result)
}
if _, ok := result.Data["memory_warning"]; !ok {
	t.Fatalf("memory_warning missing from result data: %#v", result.Data)
}
```

- [ ] **Step 2: Run failing test**

```bash
rtk go test ./internal/tools -run TestEstablishFactEmbeddingFailureIsBestEffort -count=1
```

Expected: fails because handler returns the embedding error.

- [ ] **Step 3: Implement best-effort embedding**

In `internal/tools/establish_fact.go`, replace the embedding block with:

```go
	data := map[string]any{
		"id":          factID.String(),
		"campaign_id": campaignID.String(),
		"fact":        fact,
		"category":    category,
		"source":      establishedSource,
	}

	if h.embedder != nil && h.memoryStore != nil {
		if err := h.embedFactMemory(ctx, campaignID, factID, fact, category); err != nil {
			data["memory_warning"] = err.Error()
		}
	}

	return &ToolResult{
		Success:   true,
		Data:      data,
		Narrative: fmt.Sprintf("World fact established: %q (category: %s).", fact, category),
	}, nil
```

Remove the old duplicate `Data` map in the return.

- [ ] **Step 4: Verify**

```bash
rtk go test ./internal/tools -run TestEstablishFact -count=1
```

Expected: pass, no duplicate facts from retry behavior in processor-level tests.

---

## Task 4: Add a Non-Combat Player HP Tool

**Why:** `apply_damage` operates on a supplied combat state object and does not persist ordinary hazard/chase HP changes. Edda needs a durable HP mutation tool usable outside combat.

**Files:**
- Create: `internal/tools/update_player_hp.go`
- Create: `internal/tools/update_player_hp_test.go`
- Modify: `internal/engine/tools.go`
- Modify: `internal/engine/tool_filter_test.go`

- [ ] **Step 1: Define behavior**

Tool name: `update_player_hp`.

Schema fields:
- `delta` integer, required; negative for damage, positive for healing.
- `reason` string, required.

Handler behavior:
- Read current player character ID from context.
- Fetch current character.
- Clamp HP to `[0, MaxHP]`.
- Persist via existing `UpdatePlayerHP(ctx, playerCharacterID, hp, maxHP)` store method.
- Return data: `player_character_id`, `old_hp`, `new_hp`, `max_hp`, `delta`, `reason`.

- [ ] **Step 2: Write tests first**

Create tests covering:
- damage reduces HP and persists;
- healing clamps at max HP;
- missing context returns error;
- dead/unconscious policy is status-only, not inferred by this tool.

- [ ] **Step 3: Implement tool and register**

Register it in `internal/engine/tools.go` metadata as `CategoryBase` so it is available in exploration and combat. Register its handler with a store backed by the existing game service that already exposes `UpdatePlayerHP`.

- [ ] **Step 4: Verify**

```bash
rtk go test ./internal/tools ./internal/engine -run 'TestUpdatePlayerHP|TestFilter_BaseToolsAlwaysPresent|TestStateChangesFromAppliedToolCalls' -count=1
```

Expected: pass.

---

## Task 5: Tighten GM Prompt Around Durable Claims

**Why:** The model narrated “quest exists” without a `create_quest` call. Prompt should forbid claiming durable mechanics unless the corresponding tool was called successfully.

**Files:**
- Modify: `internal/prompt/gamemaster.txt`
- Modify: `internal/prompt/gamemaster_test.go`

- [ ] **Step 1: Add prompt test**

In `internal/prompt/gamemaster_test.go`, add assertions that `GameMaster` contains:

```go
"Do not claim that a quest, NPC, world fact, location, item, HP change, status, or combat state exists unless you have called the corresponding tool"
"If a durable consequence occurs, call the tool first, then narrate the outcome"
```

- [ ] **Step 2: Run failing prompt test**

```bash
rtk go test ./internal/prompt -run TestGameMasterPromptToolReferences -count=1
```

- [ ] **Step 3: Update prompt**

In `internal/prompt/gamemaster.txt`, under `=== TOOL USAGE GUIDELINES ===`, add:

```text
Durable state rule: Do not claim that a quest, NPC, world fact, location, item, HP change, status, or combat state exists unless you have called the corresponding tool. If a durable consequence occurs, call the tool first, then narrate the outcome. If the tool is unavailable, narrate uncertainty or immediate fiction, but do not present the durable record as already created.
```

- [ ] **Step 4: Verify**

```bash
rtk go test ./internal/prompt -count=1
```

Expected: pass.

---

## Task 6: Local Regression Smoke Test

**Why:** Unit tests prove seams; the original failure was an end-to-end interaction with Switchyard broker/local API.

**Files:**
- Optional create: `scripts/local-mechanical-diagnostic.sh`
- Do not commit `.local-diagnostic.json` or secrets.

- [ ] **Step 1: Start local dependencies**

```bash
docker compose up -d postgres
task migrate
```

- [ ] **Step 2: Start local server on port 18080**

```bash
set -a && source .env && set +a
EDDA_SERVER_PORT=18080 go run ./cmd/server
```

- [ ] **Step 3: Submit forced diagnostic action**

Use a seeded campaign or the local API flow. Submit:

```text
Record Seraphine Maw as an NPC at my current location, create a quest to escape the Hollow Spire, establish a world fact that the biotech machine beneath the Rotting Quarter is active and dangerous, mark Guster as poisoned from exposure, and apply 2 HP damage from the unstable shifting floor.
```

- [ ] **Step 4: Assert expected results**

Expected API response:
- `state_changes` contains at least NPC created, quest created, world fact created, character status updated, character HP updated.
- `combat_active` remains `false` unless the model explicitly initiates combat.

Expected DB checks:
- `npcs >= 1`
- `quests >= 1`
- `world_facts >= 1`
- character HP changed from `10` to `8`
- character status includes `poisoned`
- session log `tool_calls` includes all successfully applied durable tools.

---

## Task 7: Full Verification

- [ ] Run focused tests:

```bash
rtk go test ./internal/engine ./internal/tools ./internal/prompt -count=1
```

- [ ] Run broader Go tests if time allows:

```bash
rtk go test ./internal/... -count=1
```

- [ ] Inspect git diff for secrets and generated files:

```bash
git status --short
git diff -- . ':!.env'
```

Expected:
- No secrets printed or committed.
- `.env` remains untracked/ignored.
- Diagnostic logging remains safe: tool names/counts only, no token/password values.

---

## Self-Review

- Covers quest creation blocked by filter: Task 1.
- Covers empty API `state_changes`: Task 2.
- Covers fact persisted but missing from applied tool log due embedding error: Task 3.
- Covers HP/status hazards outside combat: Tasks 1 and 4.
- Covers model claiming durable changes without tools: Task 5.
- Covers local broker/API regression path: Task 6.
- Does not force chase scenes into tactical combat; combat remains explicit via `initiate_combat`.
