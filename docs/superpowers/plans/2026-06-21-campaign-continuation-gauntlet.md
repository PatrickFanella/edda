# Campaign Continuation Gauntlet Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Make long Edda campaign play resilient enough to continue the active campaign and add a repeatable mechanics gauntlet that proves combat, HP/status, inventory, quest progress, movement, facts, memory recall, save/resume, and API/frontend state-change behavior.

**Architecture:** Keep the game engine's tool-driven state authority. Add a small retry wrapper around explicitly transient initial LLM failures, fix turn status observability through request-local callbacks, then add two separate QA lanes: a deterministic mechanics/API gauntlet on disposable campaigns and a live-model campaign smoke test. Do not hide failures: every gauntlet step should record the submitted action or tool fixture, HTTP result, `state_changes`, API snapshots, DB counts, and explicit pass/fail checks.

**Tech Stack:** Go engine/API, Ollama-compatible LLM provider through Switchyard, PostgreSQL/pgvector, Python stdlib scripts for local QA, existing `rtk go test`, `task generate`, `task migrate`, Docker Compose Postgres.

---

## Current campaign to continue

- Local base URL: `http://localhost:18080`
- Current playtest campaign: `c5c3b22c-ab69-4162-9c69-c662078e1bee`
- Character: `Ilya Voss`
- Current location after the last run: `Maintenance Tunnel Alpha-7`
- Known active issue from playtest: one transient Switchyard/Ollama 502 timed out the initial LLM call, but retrying later worked.
- Important operational constraint: turns can take 1-5+ minutes. Local test clients must use at least a 900s timeout, or use streaming/async status instead of short blocking HTTP.
- Do **not** run destructive/mechanics gauntlets directly against this active campaign. The gauntlet intentionally mutates HP, inventory, quests, movement, combat, and facts. Use a fresh disposable campaign, a DB snapshot/restore, or a cloned campaign first.

---

## Critical review corrections

This plan was reviewed before execution. These corrections are binding:

1. **Separate deterministic mechanics testing from live-model smoke testing.** The gauntlet must not depend only on natural-language model compliance. First prove mechanics/API behavior with deterministic tests or forced fixtures, then run a smaller live-model smoke test.
2. **Never mutate the active Ilya Voss campaign with the gauntlet by default.** Continue that campaign only with normal cautious play actions after health checks, or snapshot/clone/restore first.
3. **Use real tool result field names for `state_changes`.** Before adding projection rows, inspect each tool result struct/test fixture. Do not invent `id` when the real result uses `item_id`, `objective_id`, etc.
4. **Make status emission request-local.** `ProcessTurnStream` currently mutates shared processor callback state; do not add more status events on top of that without first removing cross-request leakage risk.
5. **Retry only explicit transient/provider-capacity failures.** Do not retry every `ErrConnection`; include `ErrRateLimit` with `Retry-After`, explicit `ErrTransient`, `ErrTimeout`, and connection errors that are known transport failures.

---

## File structure

### Retry and status visibility

- Modify: `internal/engine/turn_processor.go`
  - Add initial-provider-call retry/backoff for explicit transient classes: `llm.ErrTransient`, `llm.ErrTimeout`, retryable transport `llm.ErrConnection`, and `llm.ErrRateLimit` honoring `RetryAfter`.
  - Emit request-local status messages for retry attempts and duration checkpoints.
- Modify: `internal/engine/turn_processor_test.go`
  - Add tests for initial LLM retry success, retry exhaustion, and non-retryable auth/model errors.
- Modify: `internal/engine/runtime.go`
  - Make streaming status callbacks request-local before adding more phase statuses. Do not mutate shared `e.processor.StatusCallback` for concurrent requests.
  - Add status messages around memory/context stages after callback isolation exists.
- Modify: `internal/engine/runtime_test.go`
  - Assert streaming status order includes gather/context/think/tool/finalize.

### State-change coverage for mechanics gauntlet

- Modify: `internal/engine/state_changes.go`
  - Add missing state-change projections for inventory, quest progress/completion, XP/level, NPC updates, combat resolution, and time advancement using real tool result ID fields.
- Modify: `internal/engine/state_changes_test.go`
  - Add table cases for every gauntlet tool.

### QA gauntlet script and documentation

- Create: `scripts/edda_mechanics_gauntlet.py`
  - Local HTTP script that creates a disposable campaign by default, submits forced mechanics actions, fetches API snapshots, runs DB assertions for local mode, and writes redacted JSONL/Markdown artifacts.
  - It must refuse to run against a supplied existing campaign unless `--i-understand-this-mutates-campaign` is set.
- Modify: `docs/play-edda-logging.md`
  - Document the gauntlet and continuation workflow.
- Create: `docs/playtest/README.md`
  - Explain how to continue a campaign, how to run the gauntlet, what pass/fail means, and what known model-broker instability looks like.

### API/integration coverage

- Modify: `internal/handlers/integration_test.go`
  - Add API-level state-change and endpoint consistency tests for inventory, quest progress, movement, facts, save/resume, and combat where feasible with deterministic fakes.
- Modify: `internal/handlers/websocket_test.go`
  - Add status/result envelope assertions for longer turns and state changes.
- Modify: `internal/saves/handlers_test.go`
  - Add save/resume precondition coverage if the gauntlet exposes missing behavior.

---

## Task 1: Add retry/backoff for transient initial LLM failures

**Files:**
- Modify: `internal/engine/turn_processor.go`
- Modify: `internal/engine/turn_processor_test.go`

- [ ] **Step 1: Add failing tests for transient initial LLM retry**

Append tests near the existing `TestTurnProcessor_*` retry tests in `internal/engine/turn_processor_test.go`:

```go
func TestTurnProcessor_InitialTransientFailureRetriesThenSucceeds(t *testing.T) {
	reg, _ := buildProcessorTestRegistry(t, 0)
	validator := tools.NewValidator(reg)

	provider := newMockProvider(t,
		struct {
			resp *llm.Response
			err  error
		}{resp: nil, err: &llm.ErrTransient{URL: "http://broker/api/chat", StatusCode: 502, Err: errors.New("bad gateway")}},
		struct {
			resp *llm.Response
			err  error
		}{resp: &llm.Response{Content: "The broker recovered.", ToolCalls: nil}, err: nil},
	)

	tp := NewTurnProcessor(provider, reg, validator, nil)
	narrative, applied, err := tp.ProcessWithRecovery(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "continue"}}, reg.List())

	if err != nil {
		t.Fatalf("ProcessWithRecovery returned error after retry: %v", err)
	}
	if narrative != "The broker recovered." {
		t.Fatalf("narrative = %q", narrative)
	}
	if len(applied) != 0 {
		t.Fatalf("len(applied) = %d, want 0", len(applied))
	}
	if provider.callCount != 2 {
		t.Fatalf("provider.callCount = %d, want 2", provider.callCount)
	}
}

func TestTurnProcessor_InitialTransientFailureExhaustsRetry(t *testing.T) {
	reg, _ := buildProcessorTestRegistry(t, 0)
	validator := tools.NewValidator(reg)
	errTransient := &llm.ErrTimeout{URL: "http://broker/api/chat", Err: context.DeadlineExceeded}

	provider := newMockProvider(t,
		struct {
			resp *llm.Response
			err  error
		}{resp: nil, err: errTransient},
		struct {
			resp *llm.Response
			err  error
		}{resp: nil, err: errTransient},
	)

	tp := NewTurnProcessor(provider, reg, validator, nil)
	_, _, err := tp.ProcessWithRecovery(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "continue"}}, reg.List())

	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
	if provider.callCount != 2 {
		t.Fatalf("provider.callCount = %d, want 2", provider.callCount)
	}
}

func TestTurnProcessor_InitialRateLimitHonorsRetryAfter(t *testing.T) {
	reg, _ := buildProcessorTestRegistry(t, 0)
	validator := tools.NewValidator(reg)

	provider := newMockProvider(t,
		struct {
			resp *llm.Response
			err  error
		}{resp: nil, err: &llm.ErrRateLimit{URL: "http://broker/api/chat", StatusCode: 503, RetryAfter: time.Millisecond, HasRetryAfter: true, Err: errors.New("queue full")}},
		struct {
			resp *llm.Response
			err  error
		}{resp: &llm.Response{Content: "The broker queue cleared."}, err: nil},
	)

	tp := NewTurnProcessor(provider, reg, validator, nil)
	narrative, _, err := tp.ProcessWithRecovery(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "continue"}}, reg.List())
	if err != nil {
		t.Fatalf("ProcessWithRecovery returned error after rate-limit retry: %v", err)
	}
	if narrative != "The broker queue cleared." {
		t.Fatalf("narrative = %q", narrative)
	}
	if provider.callCount != 2 {
		t.Fatalf("provider.callCount = %d, want 2", provider.callCount)
	}
}

func TestTurnProcessor_InitialAuthFailureDoesNotRetry(t *testing.T) {
	reg, _ := buildProcessorTestRegistry(t, 0)
	validator := tools.NewValidator(reg)

	provider := newMockProvider(t,
		struct {
			resp *llm.Response
			err  error
		}{resp: nil, err: &llm.ErrAuth{URL: "http://broker/api/chat", StatusCode: 401, Err: errors.New("unauthorized")}},
	)

	tp := NewTurnProcessor(provider, reg, validator, nil)
	_, _, err := tp.ProcessWithRecovery(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "continue"}}, reg.List())

	if err == nil {
		t.Fatal("expected non-retryable auth error")
	}
	if provider.callCount != 1 {
		t.Fatalf("provider.callCount = %d, want 1", provider.callCount)
	}
}
```

- [ ] **Step 2: Run tests and confirm they fail**

Run:

```bash
rtk go test ./internal/engine -run 'TestTurnProcessor_Initial.*Failure' -count=1
```

Expected: transient retry tests fail because `ProcessWithRecovery` currently returns after the first initial provider error.

- [ ] **Step 3: Implement the minimal retry helper**

In `internal/engine/turn_processor.go`, add imports if missing:

```go
import (
	"errors"
	"math/rand"
	"strings"
)
```

Add constants and helper functions near `NewTurnProcessor`:

```go
const initialLLMMaxAttempts = 2

func initialLLMBackoff(attempt int, err error) time.Duration {
	var rateLimit *llm.ErrRateLimit
	if errors.As(err, &rateLimit) && rateLimit.HasRetryAfter && rateLimit.RetryAfter > 0 {
		return rateLimit.RetryAfter
	}
	base := time.Duration(750*attempt) * time.Millisecond
	jitter := time.Duration(rand.Intn(250)) * time.Millisecond
	return base + jitter
}

func isRetryableInitialLLMError(err error) bool {
	if err == nil {
		return false
	}
	var transient *llm.ErrTransient
	if errors.As(err, &transient) {
		return true
	}
	var timeout *llm.ErrTimeout
	if errors.As(err, &timeout) {
		return true
	}
	var rateLimit *llm.ErrRateLimit
	if errors.As(err, &rateLimit) {
		return rateLimit.StatusCode == 429 || rateLimit.StatusCode == 503 || rateLimit.HasRetryAfter
	}
	var conn *llm.ErrConnection
	if errors.As(err, &conn) {
		msg := strings.ToLower(conn.Error())
		return strings.Contains(msg, "connection reset") || strings.Contains(msg, "connection refused") || strings.Contains(msg, "timeout") || strings.Contains(msg, "temporary")
	}
	return false
}
```

Replace lines like the current direct initial call in `ProcessWithRecovery`:

```go
resp, err := tp.completeInitialWithRetry(ctx, messages, availableTools)
if err != nil {
	tp.logger.Error("turn processor initial llm call failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
	return "", nil, fmt.Errorf("initial LLM call failed: %w", err)
}
```

Add method:

```go
func (tp *TurnProcessor) completeInitialWithRetry(ctx context.Context, messages []llm.Message, availableTools []llm.Tool) (*llm.Response, error) {
	var lastErr error
	for attempt := 1; attempt <= initialLLMMaxAttempts; attempt++ {
		resp, err := tp.provider.Complete(ctx, messages, availableTools)
		if err == nil {
			if attempt > 1 {
				tp.logger.Info("initial llm call recovered after retry", "attempt", attempt)
			}
			return resp, nil
		}
		lastErr = err
		if attempt == initialLLMMaxAttempts || !isRetryableInitialLLMError(err) {
			return nil, err
		}
		backoff := initialLLMBackoff(attempt, err)
		tp.logger.Warn("initial llm call failed; retrying", "attempt", attempt, "next_attempt", attempt+1, "backoff_ms", backoff.Milliseconds(), "error", err)
		tp.emitStatus(api.StatusPayload{Stage: "retrying", Description: "Model provider timed out; retrying..."})
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, lastErr
}
```

- [ ] **Step 4: Run focused retry tests**

Run:

```bash
gofmt -w internal/engine/turn_processor.go internal/engine/turn_processor_test.go
rtk go test ./internal/engine -run 'TestTurnProcessor_Initial.*Failure|TestTurnProcessor_RetrySucceeds' -count=1
```

Expected: all pass.

- [ ] **Step 5: Commit only Task 1**

```bash
git add internal/engine/turn_processor.go internal/engine/turn_processor_test.go
git commit -m "fix: retry transient initial llm failures"
```

---

## Task 2: Make turn status request-local, then add long-turn timing visibility

**Files:**
- Modify: `internal/engine/runtime.go`
- Modify: `internal/engine/runtime_test.go`
- Modify: `internal/handlers/websocket_test.go`
- Optional Modify: `frontend/src/components/layout/ThinkingIndicator.tsx`

- [ ] **Step 1: Add a regression test for concurrent stream status isolation**

In `internal/engine/runtime_test.go`, add a test with two concurrent `ProcessTurnStream` calls using fake providers that block on separate channels. Each stream should receive only its own statuses. The test should fail if one request's status callback is overwritten by the other request.

Use this assertion shape:

```go
if slices.Contains(streamAStages, "stream-b-only") {
	t.Fatalf("stream A received stream B status: %v", streamAStages)
}
if slices.Contains(streamBStages, "stream-a-only") {
	t.Fatalf("stream B received stream A status: %v", streamBStages)
}
```

- [ ] **Step 2: Refactor status callback to be request-local**

Do not mutate shared `e.processor.StatusCallback` inside `ProcessTurnStream`. Instead, introduce a request-scoped processing option or method. One acceptable shape:

```go
type TurnProcessorOptions struct {
	StatusCallback func(api.StatusPayload)
}

func (tp *TurnProcessor) ProcessWithRecoveryWithOptions(
	ctx context.Context,
	messages []llm.Message,
	availableTools []llm.Tool,
	opts TurnProcessorOptions,
) (string, []AppliedToolCall, error) {
	clone := *tp
	clone.StatusCallback = opts.StatusCallback
	return clone.ProcessWithRecovery(ctx, messages, availableTools)
}
```

Then have `ProcessTurnStream` call this request-scoped method instead of setting a field on the shared processor. Keep `ProcessWithRecovery` for existing tests/callers.

- [ ] **Step 3: Add a status-capture test for streaming phases**

In `internal/engine/runtime_test.go`, add a test using the existing stream test harness. The assertion should check that streamed status envelopes include at least:

```go
wantStages := []string{"gathering", "thinking", "finalizing"}
```

If the current fake provider emits a tool call, also assert:

```go
wantStages = append(wantStages, "tools", "tool_execution")
```

- [ ] **Step 4: Emit more explicit runtime phase statuses**

In `internal/engine/runtime.go`, around gather/memory/assemble/persist phases, emit statuses through the existing stream callback path:

```go
emit(api.StatusPayload{Stage: "gathering", Description: "Loading campaign state..."})
emit(api.StatusPayload{Stage: "memory", Description: "Retrieving relevant memory..."})
emit(api.StatusPayload{Stage: "context", Description: "Assembling turn context..."})
emit(api.StatusPayload{Stage: "finalizing", Description: "Persisting turn results..."})
```

Keep HTTP `/action` response unchanged. This task is about WebSocket/status observability and logs, not changing the action schema.

- [ ] **Step 5: Add duration logs for slow phases**

In `runtime.go`, wrap phase calls with a small helper:

```go
func logPhaseDuration(logger *slog.Logger, phase string, started time.Time, attrs ...any) {
	duration := time.Since(started)
	level := slog.LevelDebug
	if duration > 30*time.Second {
		level = slog.LevelWarn
	}
	logger.Log(context.Background(), level, "turn phase completed", append([]any{"phase", phase, "duration_ms", duration.Milliseconds()}, attrs...)...)
}
```

If using `context.Background()` in helper is undesirable, pass `ctx` into the helper.

- [ ] **Step 6: Run status tests**

```bash
gofmt -w internal/engine/runtime.go internal/engine/runtime_test.go internal/handlers/websocket_test.go
rtk go test ./internal/engine ./internal/handlers -run 'TestProcessTurnStream|TestWebSocket.*Status' -count=1
```

Expected: pass with status phases visible.

- [ ] **Step 7: Commit only Task 2**

```bash
git add internal/engine/runtime.go internal/engine/runtime_test.go internal/handlers/websocket_test.go frontend/src/components/layout/ThinkingIndicator.tsx
git commit -m "chore: surface long turn status"
```

If the frontend file is not changed, omit it from `git add`.

---

## Task 3: Expand state-change projection for gauntlet tools

**Files:**
- Modify: `internal/engine/state_changes.go`
- Modify: `internal/engine/state_changes_test.go`

- [ ] **Step 1: Add failing tests for currently invisible mutators**

Before writing any projection test row, inspect the real tool result payload in the handler and existing tests. Required real result ID fields from current code:

```text
add_item / update_item / remove_item: item_id
complete_objective: objective_id plus quest_id
update_quest: id
add_experience / level_up: player_character_id
update_npc: id
resolve_combat: combat_state_id if present in its result; inspect before mapping
advance_time: campaign_id if present in its result; inspect before mapping
```

Do not write synthetic tests with `id` for item/objective tools.

Add table rows to `TestStateChangesFromAppliedToolCalls` for these tool/result shapes:

```go
{"add_item", applied("add_item", `{"item_id":"`+itemID.String()+`","player_character_id":"`+playerID.String()+`","name":"choir-spanner"}`), "inventory_item", itemID, "created"},
{"update_item", applied("update_item", `{"item_id":"`+itemID.String()+`","player_character_id":"`+playerID.String()+`","name":"charged choir-spanner"}`), "inventory_item", itemID, "updated"},
{"remove_item", applied("remove_item", `{"item_id":"`+itemID.String()+`","player_character_id":"`+playerID.String()+`"}`), "inventory_item", itemID, "removed"},
{"update_quest", applied("update_quest", `{"id":"`+questID.String()+`","status":"active"}`), "quest", questID, "updated"},
{"complete_objective", applied("complete_objective", `{"objective_id":"`+objectiveID.String()+`","quest_id":"`+questID.String()+`","objective_completed":true}`), "objective", objectiveID, "completed"},
{"add_experience", applied("add_experience", `{"player_character_id":"`+playerID.String()+`","xp":50}`), "player_character", playerID, "experience_updated"},
{"resolve_combat", applied("resolve_combat", `{"combat_state_id":"`+combatID.String()+`","outcome":"victory"}`), "combat", combatID, "resolved"},
{"update_npc", applied("update_npc", `{"id":"`+npcID.String()+`","name":"Captain Virel"}`), "npc", npcID, "updated"},
{"advance_time", applied("advance_time", `{"campaign_id":"`+campaignID.String()+`","day":1,"hour":9}`), "campaign_time", campaignID, "updated"},
```

Use IDs already declared in the test or add them at test start.

- [ ] **Step 2: Run tests and confirm failures**

```bash
rtk go test ./internal/engine -run TestStateChangesFromAppliedToolCalls -count=1
```

Expected: new rows fail because mappings do not yet exist.

- [ ] **Step 3: Add projection mappings**

In `internal/engine/state_changes.go`, extend the switch:

```go
case "add_item", "create_item":
	return one(entityStateChange("inventory_item", data, "item_id", "created"))
case "update_item", "modify_item":
	return one(entityStateChange("inventory_item", data, "item_id", "updated"))
case "remove_item":
	return one(entityStateChange("inventory_item", data, "item_id", "removed"))
case "update_quest":
	return one(entityStateChange("quest", data, "id", "updated"))
case "complete_objective":
	return one(entityStateChange("objective", data, "objective_id", "completed"))
case "add_experience", "level_up":
	return one(entityStateChange("player_character", data, "player_character_id", "experience_updated"))
case "resolve_combat":
	return one(entityStateChange("combat", data, "combat_state_id", "resolved"))
case "update_npc":
	return one(entityStateChange("npc", data, "id", "updated"))
case "advance_time":
	return one(entityStateChange("campaign_time", data, "campaign_id", "updated"))
```

If a real tool result uses a different ID field, inspect that tool and adjust the row and mapping to match reality.

- [ ] **Step 4: Run tests**

```bash
gofmt -w internal/engine/state_changes.go internal/engine/state_changes_test.go
rtk go test ./internal/engine -run TestStateChangesFromAppliedToolCalls -count=1
```

Expected: pass.

- [ ] **Step 5: Commit only Task 3**

```bash
git add internal/engine/state_changes.go internal/engine/state_changes_test.go
git commit -m "fix: project state changes for mechanics tools"
```

---

## Task 4: Add a scriptable mechanics gauntlet

**Files:**
- Create: `scripts/edda_mechanics_gauntlet.py`
- Modify: `docs/play-edda-logging.md`
- Create: `docs/playtest/README.md`

> **Implementation note:** The large script snippets below are historical planning scaffolding. The checked-in implementation in `scripts/edda_mechanics_gauntlet.py` is authoritative. If the snippet and the checked-in script disagree, follow the checked-in script. In particular, the checked-in script splits `disposable_mode` from `local_base`, refuses non-local disposable campaign creation unless `--i-understand-this-creates-remote-disposable-campaign` is supplied, refuses existing-campaign mutation unless `--i-understand-this-mutates-campaign` is supplied, and only allows `--db-counts` against a local base URL.

- [ ] **Step 1: Create the gauntlet script skeleton**

Create `scripts/edda_mechanics_gauntlet.py`:

```python
#!/usr/bin/env python3
"""Run a disposable local Edda live-model mechanics smoke test and write redacted artifacts.

By default this creates a fresh disposable campaign. Supplying --campaign-id is
dangerous because the script intentionally mutates durable campaign state; it is
refused unless --i-understand-this-mutates-campaign is also provided.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path
from typing import Any


DEFAULT_BASE_URL = "http://localhost:18080"
DEFAULT_TIMEOUT = 900


def now_slug() -> str:
    return dt.datetime.now(dt.UTC).strftime("%Y%m%dT%H%M%SZ")


class Client:
    def __init__(self, base_url: str, token: str | None = None, timeout: int = DEFAULT_TIMEOUT):
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.timeout = timeout

    def request(self, method: str, path: str, body: Any | None = None) -> Any:
        headers = {"Accept": "application/json", "Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        data = None if body is None else json.dumps(body, ensure_ascii=False).encode("utf-8")
        req = urllib.request.Request(self.base_url + path, data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                raw = resp.read().decode("utf-8")
                return json.loads(raw) if raw else None
        except urllib.error.HTTPError as exc:
            raw = exc.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"{method} {path} HTTP {exc.code}: {raw}") from exc


def append_jsonl(path: Path, record: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as f:
        f.write(json.dumps(record, ensure_ascii=False, sort_keys=True) + "\n")


def db_scalar(sql: str) -> str:
    return subprocess.check_output(
        ["docker", "compose", "exec", "-T", "postgres", "psql", "-U", "edda", "-d", "edda", "-t", "-A", "-c", sql],
        text=True,
    ).strip()


def db_snapshot(campaign_id: str) -> dict[str, int]:
    safe_campaign_id = campaign_id.replace("'", "''")
    sql = f"""
select json_build_object(
  'session_logs', (select count(*) from session_logs where campaign_id='{safe_campaign_id}'),
  'locations', (select count(*) from locations where campaign_id='{safe_campaign_id}'),
  'quests', (select count(*) from quests where campaign_id='{safe_campaign_id}'),
  'known_facts', (select count(*) from world_facts where campaign_id='{safe_campaign_id}' and player_known=true and superseded_by is null),
  'items', (select count(*) from items where campaign_id='{safe_campaign_id}')
)::text;
"""
    return json.loads(db_scalar(sql))
```

- [ ] **Step 2: Add account/campaign creation and login helpers**

Append:

```python
def register(client: Client) -> tuple[str, str, str]:
    email = f"edda-gauntlet-{int(time.time())}-{uuid.uuid4().hex[:6]}@local.test"
    password = "local-gauntlet-password"
    payload = client.request("POST", "/api/v1/auth/register", {"email": email, "password": password, "name": "Gauntlet Runner"})
    return email, password, payload["token"]


def login_for_campaign(client: Client, campaign_id: str, password: str) -> str:
    email = db_scalar(f"select u.email from users u join campaigns c on c.created_by=u.id where c.id='{campaign_id}';")
    return client.request("POST", "/api/v1/auth/login", {"email": email, "password": password})["token"]


def create_campaign(client: Client) -> str:
    payload = {
        "name": "Mechanics Gauntlet Ossuary",
        "summary": "A compact test campaign designed to exercise Edda mechanical state changes.",
        "rules_mode": "light",
        "profile": {
            "genre": "science-fantasy survival",
            "tone": "clear, tactical, testable",
            "themes": ["mechanical consequence", "danger", "escape"],
            "world_type": "sealed test station",
            "danger_level": "high",
            "political_complexity": "low",
        },
        "character_profile": {
            "name": "Testa Vey",
            "concept": "human systems scout",
            "background": "field tester",
            "personality": "cautious and practical",
            "motivations": ["survive", "map the station"],
            "strengths": ["steady hands", "situational awareness"],
            "weaknesses": ["over-cautious"],
        },
    }
    resp = client.request("POST", "/api/v1/campaigns/start/world", payload)
    campaign_id = (resp.get("campaign") or {}).get("id") or resp.get("campaign_id") or resp.get("id")
    if not campaign_id:
        raise RuntimeError(f"world creation response had no campaign id: {resp}")
    return campaign_id
```

- [ ] **Step 3: Add gauntlet action list and checks**

Append:

```python
GAUNTLET_ACTIONS = [
    ("quest_create", "Create and track a clear quest: escape this station through a safe control room. Use the quest journal if appropriate.", ["quest:created"]),
    ("move_new_location", "Move into a brand-new test corridor now. If it is new, create it and move me there in this turn.", ["location:moved"]),
    ("known_fact", "Confirm and record this known fact: the test corridor contains active blue conduit veins. Make it visible to me.", ["world_fact:created"]),
    ("hp_status", "A harmless but real electrical shock grazes me; reduce my HP by 1 and mark the injury if appropriate.", ["player_character:hp_updated"]),
    ("inventory", "Secure one practical item from the corridor and add it to my inventory: a calibrated test spanner.", ["inventory_item:created"]),
    ("existing_move", "Return through the connected route to the previous room using the existing connection if possible.", ["location:moved"]),
    ("quest_progress", "Mark progress on the active escape quest: we found a safe corridor route toward the control room.", ["objective:completed", "quest:updated"]),
    ("combat_start", "A weak hostile test drone blocks the route. Start combat only if it is unavoidable.", ["combat:started"]),
    ("combat_resolve", "Resolve the weak drone encounter cautiously, avoiding lethal risk if possible.", ["combat:resolved"]),
    ("save_resume", "Pause at a defensible point and prepare to save progress.", []),
]


def changes(resp: dict[str, Any]) -> list[str]:
    return [f"{c.get('entity_type')}:{c.get('change_type')}" for c in resp.get("state_changes", [])]


def snapshot(client: Client, campaign_id: str) -> dict[str, Any]:
    parts = {}
    for name, path in {
        "character": "/character",
        "locations": "/locations",
        "quests": "/quests",
        "inventory": "/character/inventory",
        "facts": "/facts",
        "history": "/history",
        "time": "/time",
    }.items():
        parts[name] = client.request("GET", f"/api/v1/campaigns/{campaign_id}{path}")
    return parts


def run_actions(client: Client, campaign_id: str, artifact_dir: Path) -> list[dict[str, Any]]:
    results = []
    for index, (label, action, expected_any) in enumerate(GAUNTLET_ACTIONS, 1):
        started = time.time()
        before_db = db_snapshot(campaign_id)
        try:
            resp = client.request("POST", f"/api/v1/campaigns/{campaign_id}/action", {"input": action})
            observed = changes(resp)
            snap = snapshot(client, campaign_id)
            after_db = db_snapshot(campaign_id)
            ok = not expected_any or any(e in observed for e in expected_any)
            record = {"step": index, "label": label, "ok": ok, "seconds": round(time.time() - started, 1), "expected_any": expected_any, "changes": observed, "narrative": (resp.get("narrative") or "")[:800], "snapshot": summarize_snapshot(snap), "db_before": before_db, "db_after": after_db}
        except Exception as exc:
            record = {"step": index, "label": label, "ok": False, "error": repr(exc), "seconds": round(time.time() - started, 1)}
        append_jsonl(artifact_dir / "gauntlet.jsonl", record)
        results.append(record)
        print(f"{index:02d} {label}: {'PASS' if record.get('ok') else 'FAIL'} {record.get('changes', record.get('error'))}", flush=True)
    return results


def summarize_snapshot(snap: dict[str, Any]) -> dict[str, Any]:
    char = snap.get("character") or {}
    return {
        "hp": [char.get("hp"), char.get("max_hp")],
        "status": char.get("status"),
        "location": char.get("current_location_id"),
        "locations": len(snap.get("locations") or []),
        "quests": len(snap.get("quests") or []),
        "inventory": len(snap.get("inventory") or []),
        "facts": len(snap.get("facts") or []),
    }
```

- [ ] **Step 4: Add CLI entry point**

Append:

```python
def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default=os.environ.get("EDDA_BASE_URL", DEFAULT_BASE_URL))
    parser.add_argument("--campaign-id", default="")
    parser.add_argument("--password", default="local-gauntlet-password")
    parser.add_argument("--artifact-dir", default=f"logs/edda/gauntlet-{now_slug()}")
    parser.add_argument("--i-understand-this-mutates-campaign", action="store_true")
    args = parser.parse_args()

    client = Client(args.base_url)
    if args.campaign_id:
        if not args.i_understand_this_mutates_campaign:
            raise SystemExit("Refusing to mutate an existing campaign. Re-run with --i-understand-this-mutates-campaign only after snapshot/clone/restore preparation.")
        client.token = login_for_campaign(client, args.campaign_id, args.password)
        campaign_id = args.campaign_id
    else:
        _email, _password, token = register(client)
        client.token = token
        campaign_id = create_campaign(client)

    artifact_dir = Path(args.artifact_dir)
    artifact_dir.mkdir(parents=True, exist_ok=True)
    results = run_actions(client, campaign_id, artifact_dir)
    failed = [r for r in results if not r.get("ok")]
    summary = {"campaign_id": campaign_id, "passed": len(results) - len(failed), "failed": len(failed), "artifact_dir": str(artifact_dir)}
    (artifact_dir / "summary.json").write_text(json.dumps(summary, indent=2), encoding="utf-8")
    print(json.dumps(summary, indent=2))
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
```

- [ ] **Step 5: Run the gauntlet against local API**

```bash
chmod +x scripts/edda_mechanics_gauntlet.py
EDDA_BASE_URL=http://localhost:18080 scripts/edda_mechanics_gauntlet.py
```

Expected: The first run may expose real gaps. Do not treat a failed gauntlet as a failed implementation; file the exact failing step in the report and continue to Task 5 for targeted fixes.

- [ ] **Step 6: Document usage**

Add to `docs/play-edda-logging.md`:

```markdown
## Mechanics gauntlet

Run a local mechanics QA pass:

```bash
EDDA_BASE_URL=http://localhost:18080 scripts/edda_mechanics_gauntlet.py
```

Mutate an existing campaign only after snapshot/clone/restore preparation:

```bash
EDDA_BASE_URL=http://localhost:18080 EDDA_TOKEN=placeholder-token scripts/edda_mechanics_gauntlet.py \
  --campaign-id placeholder-campaign-id \
  --token "$EDDA_TOKEN" \
  --i-understand-this-mutates-campaign
```

This is destructive and not for active campaigns unless they have been snapshotted or cloned first.

Artifacts are written under `logs/edda/gauntlet-*` and are gitignored.
```

Create `docs/playtest/README.md` with:

```markdown
# Edda Playtest Runbook

## Before continuing a campaign

1. Confirm local API health: `curl http://localhost:18080/api/healthz`.
2. Confirm broker health: `curl http://10.0.0.50:11434/api/models`.
3. Use 900s client timeouts for blocking `/action` calls.
4. Expect transient 502s from the model broker; retry once after health recovers.

## Mechanics gauntlet

Use `scripts/edda_mechanics_gauntlet.py` for a repeatable pass over durable mechanics.
By default it creates a disposable campaign. Do not run it against a real campaign unless the campaign has been cloned or snapshotted.
Treat any missing `state_changes` or API/DB mismatch as a bug candidate.
```

- [ ] **Step 7: Commit only Task 4**

```bash
git add scripts/edda_mechanics_gauntlet.py docs/play-edda-logging.md docs/playtest/README.md
git commit -m "test: add mechanics gauntlet runbook"
```

---

## Task 5: Fix gaps exposed by first gauntlet run

**Files:**
- Modify based on failing gauntlet steps. Likely files:
  - `internal/engine/state_changes.go`
  - `internal/tools/*`
  - `internal/prompt/gamemaster.txt`
  - `internal/handlers/*`

- [ ] **Step 1: Read gauntlet failures**

Open the latest artifact:

```bash
python - <<'PY'
import json, pathlib
latest = sorted(pathlib.Path('logs/edda').glob('gauntlet-*'))[-1]
print(latest)
for line in (latest / 'gauntlet.jsonl').read_text().splitlines():
    row = json.loads(line)
    if not row.get('ok'):
        print(json.dumps(row, indent=2)[:2000])
PY
```

- [ ] **Step 2: For each failed step, classify root cause**

Use this decision table:

```text
No tool call emitted        -> prompt/tool availability issue
Tool call emitted, DB wrong -> tool/store bug
DB right, API wrong         -> handler/query/visibility bug
API right, state_changes missing -> state_changes projection bug
HTTP 500/502               -> provider retry/runtime issue
Frontend stale only         -> frontend invalidation/parser issue
```

- [ ] **Step 3: Add one failing test per real bug**

Examples:

```bash
rtk go test ./internal/engine -run TestStateChangesFromAppliedToolCalls -count=1
rtk go test ./internal/tools -run TestUpdatePlayerHP -count=1
rtk go test ./internal/handlers -run TestAction -count=1
```

- [ ] **Step 4: Fix only the root cause for that step**

Do not bundle multiple gauntlet failures into one opaque patch. Use one commit per root cause.

- [ ] **Step 5: Re-run the failing gauntlet step or whole gauntlet**

```bash
EDDA_BASE_URL=http://localhost:18080 scripts/edda_mechanics_gauntlet.py
```

Expected: either all steps pass or the remaining failures are documented as new follow-up issues with exact artifacts.

---

## Task 6: Add API integration tests for critical mechanics

**Files:**
- Modify: `internal/handlers/integration_test.go`
- Modify: `internal/handlers/websocket_test.go`

- [ ] **Step 1: Add deterministic action integration test scaffold**

In `internal/handlers/integration_test.go`, add a test that uses a fake engine/provider returning predetermined applied tool results. The test should assert API response includes state changes for:

```go
[]string{
	"quest:created",
	"player_character:location_updated",
	"location:moved",
	"world_fact:created",
	"player_character:hp_updated",
	"inventory_item:created",
	"combat:started",
}
```

- [ ] **Step 2: Add WebSocket result envelope test**

In `internal/handlers/websocket_test.go`, extend existing websocket tests so a fake result with `state_changes` verifies the outgoing result envelope preserves:

```go
combat_active
state_changes
choices
narrative
```

- [ ] **Step 3: Run integration tests**

```bash
gofmt -w internal/handlers/integration_test.go internal/handlers/websocket_test.go
rtk go test ./internal/handlers -run 'Test.*StateChanges|Test.*WebSocket.*Result' -count=1
```

- [ ] **Step 4: Commit only Task 6**

```bash
git add internal/handlers/integration_test.go internal/handlers/websocket_test.go
git commit -m "test: cover mechanics state changes at API boundary"
```

---

## Task 7: Continue the current campaign with non-destructive guardrails

**Files:**
- No code files required unless failures are found.
- Artifacts: `logs/edda/gauntlet-*` and `logs/edda/edda-turns.jsonl` remain gitignored.

- [ ] **Step 1: Health checks**

```bash
curl -fsS http://localhost:18080/api/healthz
curl -fsS http://10.0.0.50:11434/api/models >/tmp/edda-models.json
```

Expected: API returns `{"engine_ready":true,"status":"ok"}` and broker lists the Ollama models.

- [ ] **Step 2: Snapshot the active campaign before any continuation test**

Create a local DB snapshot for rollback before touching `c5c3b22c-ab69-4162-9c69-c662078e1bee`:

```bash
mkdir -p /tmp/opencode/edda-backups
docker compose exec -T postgres pg_dump -U edda -d edda \
  --format=custom \
  --file=/tmp/edda-before-ilya-continuation.dump
docker cp "$(docker compose ps -q postgres)":/tmp/edda-before-ilya-continuation.dump /tmp/opencode/edda-backups/
```

Expected: `/tmp/opencode/edda-backups/edda-before-ilya-continuation.dump` exists.

- [ ] **Step 3: Continue Ilya Voss campaign with ordinary cautious play, not the gauntlet**

Submit 3-5 normal campaign-appropriate turns with a 900s timeout. Do not intentionally force HP loss, inventory mutation, quest completion, or combat in the active campaign unless the user explicitly asks.

Use a small helper or the existing logged-turn script:

```bash
EDDA_BASE_URL=http://localhost:18080 \
EDDA_TOKEN="$EDDA_TOKEN" \
scripts/edda_logged_turn.py c5c3b22c-ab69-4162-9c69-c662078e1bee \
  "I pause in Maintenance Tunnel Alpha-7, check my condition and exits, then choose the safest next threshold without touching unstable machinery."
```

If authentication must be recreated, log in using the local test account that owns this campaign. Do not print or commit the password/token.

- [ ] **Step 4: Verify final state after continuing**

Run:

```bash
docker compose exec -T postgres psql -U edda -d edda -t -A -c "
select 'session_logs', count(*) from session_logs where campaign_id='c5c3b22c-ab69-4162-9c69-c662078e1bee'
union all select 'locations', count(*) from locations where campaign_id='c5c3b22c-ab69-4162-9c69-c662078e1bee'
union all select 'quests', count(*) from quests where campaign_id='c5c3b22c-ab69-4162-9c69-c662078e1bee'
union all select 'known_facts', count(*) from world_facts where campaign_id='c5c3b22c-ab69-4162-9c69-c662078e1bee' and player_known=true and superseded_by is null;
"
```

Expected: counts increase only where the ordinary play turns actually changed durable state. HP/status should not change unless the narrative/tool call explicitly caused it.

- [ ] **Step 5: Report findings**

Report:

```markdown
## Continued campaign report
- Campaign ID:
- Turns attempted:
- Turns succeeded:
- Transient provider failures:
- New state changes:
- Final HP/status/location:
- Quest changes:
- Known facts:
- Bugs found:
```

- [ ] **Step 6: Run the destructive mechanics gauntlet only on a disposable campaign**

```bash
EDDA_BASE_URL=http://localhost:18080 scripts/edda_mechanics_gauntlet.py
```

Expected: the script creates a fresh campaign and does not mutate `c5c3b22c-ab69-4162-9c69-c662078e1bee`.

---

## Task 8: Full verification before merging/deploying

**Files:**
- No new files expected.

- [ ] **Step 1: Run generation**

```bash
task generate
```

Expected: sqlc generation succeeds and no unexpected generated drift appears.

- [ ] **Step 2: Run full Go suite**

```bash
rtk go test ./... -count=1
```

Expected: all packages pass.

- [ ] **Step 3: Run migrations against local DB**

```bash
task migrate
```

Expected: no pending migrations or migration success.

- [ ] **Step 4: Run the gauntlet**

```bash
EDDA_BASE_URL=http://localhost:18080 scripts/edda_mechanics_gauntlet.py
```

Expected: all required steps pass, or remaining failures are documented with exact artifacts and root-cause classification.

- [ ] **Step 5: Commit verification artifacts?**

Do not commit `logs/edda/*`. Commit only code/docs/test changes:

```bash
git status --short
```

Expected: no untracked log artifacts are staged.

---

## Self-review

- Spec coverage:
  - Transient broker 502 retry: Task 1.
  - Long-turn visibility: Task 2.
  - Mechanics not yet tested: Tasks 3-6.
  - Continue campaign safely: Task 7.
  - Full verification: Task 8.
- Placeholder scan: no `TBD`, no intentionally vague implementation steps.
- Type consistency:
  - Uses existing `llm.ErrTransient`, `llm.ErrTimeout`, `llm.ErrConnection` from `internal/llm/errors.go`.
  - Uses existing `api.StatusPayload` and `TurnProcessor.StatusCallback` paths.
  - Uses existing `StateChangesFromAppliedToolCalls` shape and `entity:change` strings seen in playtest reports.

---

## Review gates

Before execution:
- Ask `oracle` to review Task 1 and Task 2 for retry/status design risk.
- Ask `oracle` or a test-focused reviewer to review Task 4 gauntlet scope after the first script draft.

Before final report:
- Run `task generate` and `rtk go test ./... -count=1`.
- Run the gauntlet once against local API.
- Report any transient broker failures separately from Edda state bugs.
