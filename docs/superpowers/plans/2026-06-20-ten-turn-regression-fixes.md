# Ten-Turn Regression Fixes Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Fix the five regressions exposed by the local 10-turn Edda campaign run: unreliable new-location movement, duplicate-ish locations, embedding dimension mismatch, missing quest persistence, and hidden visible facts.

**Architecture:** Put durable-state truth behind deterministic seams instead of relying on prompt obedience. Add small domain helpers for canonical location identity and durable-claim auditing, make visible fact persistence explicit, validate memory vector dimensions at startup, and repair/regenerate turns before `session_logs` persistence when narrative claims do not match applied tools.

**Tech Stack:** Go, sqlc, PostgreSQL + pgvector, goose migrations, Edda engine/tool pipeline, `rtk go test`.

---

## Files and responsibilities

- `internal/domain/location_names.go` / `internal/domain/location_names_test.go` — canonical location-name normalization used by world creation and play tools.
- `internal/tools/create_location.go` / `internal/tools/create_location_test.go` — reuse same-ish locations, make `move_player_here` results truthful, expose movement state data only after persisted movement.
- `internal/world/orchestrator.go` / `internal/world/orchestrator_test.go` — resolve skeleton starting locations by canonical name, not exact string only.
- `internal/config/config.go` / `internal/config/config_test.go` — add `llm.ollama.embeddingdimension` configuration.
- `internal/memory/schema.go` / `internal/memory/schema_test.go` — inspect DB vector dimensions and validate against configured embedder dimension.
- `cmd/server/main.go` — pass configured embedding dimension to the Ollama embedder and fail fast on mismatch.
- `migrations/00035_resize_memories_embedding_to_768.sql` — repair existing local/dev memory schema mismatch by clearing derived memory rows and resizing the vector column to 768.
- `queries/world_facts.sql` — make `CreateFact` accept `player_known` so visibility is persisted in the insert.
- `internal/tools/establish_fact.go` / `internal/tools/establish_fact_test.go` — visible play facts default to `player_known=true`; hidden facts remain possible.
- `internal/world/store_adapter.go` / world tests — startup skeleton facts stay hidden by setting `player_known=false`.
- `internal/engine/durable_claims.go` / `internal/engine/durable_claims_test.go` — detect durable narrative claims that lack matching applied state tools.
- `internal/engine/turn_processor.go` / `internal/engine/turn_processor_test.go` — add one repair pass before returning narrative/applied tools.
- `internal/engine/runtime.go` / engine integration tests — run durable-claim audit before `persistStage`; refresh session-log location data after movement.
- `internal/prompt/gamemaster.txt` / `internal/prompt/gamemaster_test.go` — align prompt wording with deterministic rules.

---

### Task 1: Add canonical location names

**Files:**
- Create: `internal/domain/location_names.go`
- Create: `internal/domain/location_names_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/domain/location_names_test.go`:

```go
package domain

import "testing"

func TestCanonicalLocationNameCollapsesObviousVariants(t *testing.T) {
	cases := map[string]string{
		"Lower Tiers Corridor":       "lower tier corridor",
		"Lower Tier Corridor":        "lower tier corridor",
		"  The Core Reactor  ":       "core reactor",
		"core-reactor":               "core reactor",
		"The   Forgotten   Vaults":   "forgotten vault",
		"Forgotten Vault":            "forgotten vault",
		"Abyssal Tiers":              "abyssal tier",
	}
	for input, want := range cases {
		if got := CanonicalLocationName(input); got != want {
			t.Fatalf("CanonicalLocationName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSameCanonicalLocationName(t *testing.T) {
	if !SameCanonicalLocationName("Lower Tiers Corridor", "Lower Tier Corridor") {
		t.Fatal("expected obvious singular/plural variant to match")
	}
	if SameCanonicalLocationName("Core Archive", "Core Reactor") {
		t.Fatal("expected distinct core locations not to match")
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
rtk go test ./internal/domain -run 'TestCanonicalLocationName|TestSameCanonicalLocationName' -count=1
```

Expected: FAIL because `CanonicalLocationName` is undefined.

- [ ] **Step 3: Implement canonicalization**

Create `internal/domain/location_names.go`:

```go
package domain

import (
	"regexp"
	"strings"
)

var nonLocationNameWord = regexp.MustCompile(`[^a-z0-9]+`)

// CanonicalLocationName returns a conservative key for detecting obvious
// same-location variants inside one campaign. It is not a broad fuzzy matcher.
func CanonicalLocationName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = nonLocationNameWord.ReplaceAllString(name, " ")
	parts := strings.Fields(name)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "the" || part == "a" || part == "an" {
			continue
		}
		out = append(out, singularizeLocationToken(part))
	}
	return strings.Join(out, " ")
}

func SameCanonicalLocationName(a, b string) bool {
	ca := CanonicalLocationName(a)
	cb := CanonicalLocationName(b)
	return ca != "" && ca == cb
}

func singularizeLocationToken(token string) string {
	if len(token) <= 3 {
		return token
	}
	if strings.HasSuffix(token, "ies") && len(token) > 4 {
		return strings.TrimSuffix(token, "ies") + "y"
	}
	if strings.HasSuffix(token, "s") && !strings.HasSuffix(token, "ss") {
		return strings.TrimSuffix(token, "s")
	}
	return token
}
```

- [ ] **Step 4: Verify tests pass**

Run:

```bash
rtk go test ./internal/domain -run 'TestCanonicalLocationName|TestSameCanonicalLocationName' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/location_names.go internal/domain/location_names_test.go
git commit -m "fix: canonicalize location names"
```

---

### Task 2: Reuse duplicate-ish locations and resolve starting locations canonically

**Files:**
- Modify: `internal/tools/create_location.go:246-255`
- Modify: `internal/tools/create_location_test.go`
- Modify: `internal/world/orchestrator.go:194-218`
- Modify: `internal/world/persist_adapter_orchestrator_test.go`

- [ ] **Step 1: Add create-location reuse test**

In `internal/tools/create_location_test.go`, add a test next to the existing exact reuse test:

```go
func TestCreateLocationReusesCanonicalNameVariant(t *testing.T) {
	currentLocationID := uuid.New()
	campaignID := uuid.New()
	existingLocationID := uuid.New()
	store := newStubLocationStore(t)
	store.locations[currentLocationID] = statedb.Location{ID: pgUUID(currentLocationID), CampaignID: pgUUID(campaignID), Name: "The Ark"}
	store.locations[existingLocationID] = statedb.Location{
		ID:           pgUUID(existingLocationID),
		CampaignID:   pgUUID(campaignID),
		Name:         "Lower Tier Corridor",
		Description:  pgtype.Text{String: "A narrow corridor.", Valid: true},
		Region:       pgtype.Text{String: "Lower Tiers", Valid: true},
		LocationType: pgtype.Text{String: "corridor", Valid: true},
		Properties:   []byte(`{"light":"dim"}`),
	}

	handler := NewCreateLocationHandler(store, nil, nil)
	ctx := WithCurrentLocationID(context.Background(), currentLocationID)
	result, err := handler.Handle(ctx, map[string]any{
		"name":          "Lower Tiers Corridor",
		"description":   "A second spelling of the same corridor.",
		"region":        "Lower Tiers",
		"location_type": "corridor",
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if result.Data["id"] != existingLocationID.String() {
		t.Fatalf("expected existing location id %s, got %#v", existingLocationID, result.Data["id"])
	}
	if result.Data["reused"] != true {
		t.Fatalf("expected reused=true, got %#v", result.Data["reused"])
	}
	if store.createdLocations != 0 {
		t.Fatalf("expected no new location rows, got %d", store.createdLocations)
	}
}
```

Adjust helper field names to the actual stub names in `create_location_test.go`; keep the assertions identical.

- [ ] **Step 2: Add starting-location resolution test**

In `internal/world/persist_adapter_orchestrator_test.go`, add:

```go
func TestResolveStartingLocationUsesCanonicalName(t *testing.T) {
	ctx := context.Background()
	campaignID := pgUUID(uuid.New())
	locationID := uuid.New()
	q := &stubQuerier{
		listLocationsByCampaignFn: func(_ context.Context, _ pgtype.UUID) ([]statedb.Location, error) {
			return []statedb.Location{{
				ID:         pgUUID(locationID),
				CampaignID: campaignID,
				Name:       "The Core Reactor",
			}}, nil
		},
	}

	got, err := resolveStartingLocation(ctx, q, campaignID, "Core Reactor")
	if err != nil {
		t.Fatalf("resolveStartingLocation returned error: %v", err)
	}
	if got != locationID {
		t.Fatalf("got %s, want %s", got, locationID)
	}
}
```

- [ ] **Step 3: Run tests to verify failure**

Run:

```bash
rtk go test ./internal/tools ./internal/world -run 'TestCreateLocationReusesCanonicalNameVariant|TestResolveStartingLocationUsesCanonicalName' -count=1
```

Expected: FAIL because matching is still exact.

- [ ] **Step 4: Implement canonical matching**

In `internal/tools/create_location.go`, import domain if not already imported and replace exact matching:

```go
for _, location := range existing {
	if domain.SameCanonicalLocationName(location.Name, name) {
		return location, true, nil
	}
}
```

In `internal/world/orchestrator.go`, import `git.subcult.tv/subculture-collective/edda/internal/domain` and replace the exact comparison:

```go
if domain.SameCanonicalLocationName(loc.Name, name) {
	if loc.ID.Valid {
		match = loc.ID.Bytes
		count++
	}
}
```

Keep the ambiguity behavior: if canonical matching finds 0 or more than 1 rows, return the existing “expected exactly one location named …” error.

- [ ] **Step 5: Verify tests pass**

Run:

```bash
rtk go test ./internal/tools ./internal/world -run 'TestCreateLocationReusesCanonicalNameVariant|TestCreateLocation|TestResolveStartingLocationUsesCanonicalName|TestResolveStartingLocation' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/tools/create_location.go internal/tools/create_location_test.go internal/world/orchestrator.go internal/world/persist_adapter_orchestrator_test.go
git commit -m "fix: reuse canonical location variants"
```

---

### Task 3: Make embedding dimensions configurable and validated

**Files:**
- Modify: `internal/config/config.go:26-34,98-112`
- Create: `internal/config/config_test.go` if missing
- Create: `internal/memory/schema.go`
- Create: `internal/memory/schema_test.go`
- Modify: `cmd/server/main.go:88-104`

- [ ] **Step 1: Write config tests**

In `internal/config/config_test.go`, add:

```go
func TestLoadDefaultsEmbeddingDimension(t *testing.T) {
	t.Setenv("EDDA_LLM_OLLAMA_EMBEDDINGDIMENSION", "")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.LLM.Ollama.EmbeddingDimension != 768 {
		t.Fatalf("EmbeddingDimension = %d, want 768", cfg.LLM.Ollama.EmbeddingDimension)
	}
}

func TestLoadEmbeddingDimensionFromEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_OLLAMA_EMBEDDINGDIMENSION", "1536")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.LLM.Ollama.EmbeddingDimension != 1536 {
		t.Fatalf("EmbeddingDimension = %d, want 1536", cfg.LLM.Ollama.EmbeddingDimension)
	}
}
```

- [ ] **Step 2: Write memory schema tests**

Create `internal/memory/schema_test.go`:

```go
package memory

import (
	"context"
	"errors"
	"testing"
)

type dimensionQuerierFunc func(context.Context, string, ...any) DimensionRow

type DimensionRow interface { Scan(...any) error }

type fakeDimensionDB struct { dim int; err error }

func (f fakeDimensionDB) QueryRow(_ context.Context, _ string, _ ...any) dimensionRow {
	return dimensionRow{dim: f.dim, err: f.err}
}

type dimensionRow struct { dim int; err error }

func (r dimensionRow) Scan(dest ...any) error {
	if r.err != nil { return r.err }
	*(dest[0].(*int)) = r.dim
	return nil
}

func TestValidateMemoryEmbeddingDimension(t *testing.T) {
	if err := ValidateMemoryEmbeddingDimension(context.Background(), fakeDimensionDB{dim: 768}, 768); err != nil {
		t.Fatalf("expected matching dimensions to pass: %v", err)
	}
	err := ValidateMemoryEmbeddingDimension(context.Background(), fakeDimensionDB{dim: 1536}, 768)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
	if !errors.Is(err, ErrMemoryEmbeddingDimensionMismatch) {
		t.Fatalf("expected ErrMemoryEmbeddingDimensionMismatch, got %v", err)
	}
}
```

If the exact fake type does not compile, keep the production interface below and adjust the fake to satisfy it.

- [ ] **Step 3: Run tests to verify failure**

Run:

```bash
rtk go test ./internal/config ./internal/memory -run 'TestLoad.*EmbeddingDimension|TestValidateMemoryEmbeddingDimension' -count=1
```

Expected: FAIL because fields/functions do not exist.

- [ ] **Step 4: Add config field and defaults**

In `internal/config/config.go`, update `OllamaConfig`:

```go
EmbeddingDimension int `koanf:"embeddingdimension"`
```

Add default:

```go
"llm.ollama.embeddingdimension": 768,
```

Extend `Validate`:

```go
if c.LLM.Provider == "ollama" && c.LLM.Ollama.EmbeddingDimension <= 0 {
	return errors.New("ollama embedding dimension must be positive")
}
```

- [ ] **Step 5: Add memory dimension validation helper**

Create `internal/memory/schema.go`:

```go
package memory

import (
	"context"
	"errors"
	"fmt"
)

var ErrMemoryEmbeddingDimensionMismatch = errors.New("memory embedding dimension mismatch")

type DimensionDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) DimensionRow
}

type DimensionRow interface {
	Scan(dest ...any) error
}

func MemoryEmbeddingDimension(ctx context.Context, db DimensionDB) (int, error) {
	const query = `SELECT atttypmod::int FROM pg_attribute WHERE attrelid = 'memories'::regclass AND attname = 'embedding'`
	var dim int
	if err := db.QueryRow(ctx, query).Scan(&dim); err != nil {
		return 0, err
	}
	return dim, nil
}

func ValidateMemoryEmbeddingDimension(ctx context.Context, db DimensionDB, configured int) error {
	actual, err := MemoryEmbeddingDimension(ctx, db)
	if err != nil {
		return fmt.Errorf("read memories embedding dimension: %w", err)
	}
	if actual != configured {
		return fmt.Errorf("%w: database memories.embedding is vector(%d), configured embedder dimension is %d", ErrMemoryEmbeddingDimensionMismatch, actual, configured)
	}
	return nil
}
```

- [ ] **Step 6: Wire server startup**

In `cmd/server/main.go`, before constructing the embedder in the Ollama branch:

```go
if err := memory.ValidateMemoryEmbeddingDimension(ctx, pool, cfg.LLM.Ollama.EmbeddingDimension); err != nil {
	logger.Errorf("memory schema validation: %v", err)
	return 1
}
```

Pass dimension into the embedder:

```go
memory.WithOllamaEmbedderDimension(cfg.LLM.Ollama.EmbeddingDimension),
```

- [ ] **Step 7: Verify tests pass**

Run:

```bash
rtk go test ./internal/config ./internal/memory ./cmd/server -run 'TestLoad.*EmbeddingDimension|TestValidateMemoryEmbeddingDimension|TestNewOllamaEmbedderDefaults' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go internal/memory/schema.go internal/memory/schema_test.go cmd/server/main.go
git commit -m "fix: validate memory embedding dimension"
```

---

### Task 4: Resize local/dev memory schema to 768

**Files:**
- Create: `migrations/00035_resize_memories_embedding_to_768.sql`
- Modify: `.env.example`
- Modify: `.env.production.example`

- [ ] **Step 1: Add migration**

Create `migrations/00035_resize_memories_embedding_to_768.sql`:

```sql
-- +goose Up
-- Memories are derived recall cache. Existing rows may have embeddings from a
-- different model width, so clear them before resizing the vector column.
DROP INDEX IF EXISTS idx_memories_embedding_hnsw;
DELETE FROM memories;
ALTER TABLE memories
  ALTER COLUMN embedding TYPE vector(768)
  USING NULL::vector(768);
CREATE INDEX idx_memories_embedding_hnsw
  ON memories USING hnsw (embedding vector_cosine_ops)
  WITH (m = 16, ef_construction = 64);

-- +goose Down
DROP INDEX IF EXISTS idx_memories_embedding_hnsw;
DELETE FROM memories;
ALTER TABLE memories
  ALTER COLUMN embedding TYPE vector(1536)
  USING NULL::vector(1536);
CREATE INDEX idx_memories_embedding_hnsw
  ON memories USING hnsw (embedding vector_cosine_ops)
  WITH (m = 16, ef_construction = 64);
```

- [ ] **Step 2: Document env var**

Add to `.env.example` and `.env.production.example` near the embedding model:

```dotenv
EDDA_LLM_OLLAMA_EMBEDDINGDIMENSION=768
```

If production uses a 1536-dimensional embedding model, set the production example to `1536` and add a comment naming the expected model.

- [ ] **Step 3: Run generation/migrations locally**

Run:

```bash
task migrate
```

Expected: migration 35 applies successfully.

- [ ] **Step 4: Verify DB dimension**

Run:

```bash
docker compose exec -T postgres psql -U edda -d edda -c "SELECT atttypmod AS embedding_typmod FROM pg_attribute WHERE attrelid = 'memories'::regclass AND attname = 'embedding';"
```

Expected: `embedding_typmod` is `768`.

- [ ] **Step 5: Commit**

```bash
git add migrations/00035_resize_memories_embedding_to_768.sql .env.example .env.production.example
git commit -m "fix: align memory vector dimension with local embedder"
```

---

### Task 5: Persist fact visibility in the insert

**Files:**
- Modify: `queries/world_facts.sql:1-13`
- Run generated updates in `internal/state/sqlc/world_facts.sql.go`, `internal/state/sqlc/querier.go`
- Modify: `internal/tools/establish_fact.go:112-155`
- Modify: `internal/tools/establish_fact_test.go`
- Modify: `internal/world/store_adapter.go:106-113`
- Modify tests/stubs that construct `statedb.CreateFactParams`

- [ ] **Step 1: Update SQL query**

Change `queries/world_facts.sql` `CreateFact` to:

```sql
-- name: CreateFact :one
INSERT INTO world_facts (
  campaign_id,
  fact,
  category,
  source,
  player_known
) VALUES (
  sqlc.arg(campaign_id),
  sqlc.arg(fact),
  sqlc.arg(category),
  sqlc.arg(source),
  sqlc.arg(player_known)
)
RETURNING id, campaign_id, fact, category, source, superseded_by, created_at, player_known;
```

- [ ] **Step 2: Regenerate sqlc**

Run:

```bash
task generate
```

Expected: generated `CreateFactParams` gains `PlayerKnown bool`.

- [ ] **Step 3: Update visible play fact creation**

In `internal/tools/establish_fact.go`, parse optional reveal flag with visible default:

```go
revealToPlayer := true
if _, exists := args["reveal_to_player"]; exists {
	parsed, err := parseBoolArg(args, "reveal_to_player")
	if err != nil {
		return nil, err
	}
	revealToPlayer = parsed
}
worldFact, err := h.factStore.CreateFact(ctx, statedb.CreateFactParams{
	CampaignID:   currentLocation.CampaignID,
	Fact:         fact,
	Category:     category,
	Source:       establishedSource,
	PlayerKnown:  revealToPlayer,
})
```

Remove the best-effort `SetFactPlayerKnown` block. Include `player_known` in every success and warning payload:

```go
"player_known": revealToPlayer,
```

- [ ] **Step 4: Keep startup skeleton facts hidden**

In `internal/world/store_adapter.go`, set:

```go
PlayerKnown: false,
```

inside `CreateWorldFact`.

- [ ] **Step 5: Update tests**

In `internal/tools/establish_fact_test.go`, add:

```go
func TestEstablishFactDefaultsToPlayerKnown(t *testing.T) {
	store := newStubFactStore(t)
	handler := NewEstablishFactHandler(store, nil, nil)
	ctx := WithCurrentLocationID(context.Background(), store.currentLocationID)

	result, err := handler.Handle(ctx, map[string]any{
		"fact":     "The visible machine is awake.",
		"category": "hazard",
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !store.lastCreateFactParams.PlayerKnown {
		t.Fatal("expected visible established fact to be player_known")
	}
	if result.Data["player_known"] != true {
		t.Fatalf("expected result player_known=true, got %#v", result.Data["player_known"])
	}
}
```

Add a second test for hidden facts:

```go
func TestEstablishFactCanRemainHidden(t *testing.T) {
	store := newStubFactStore(t)
	handler := NewEstablishFactHandler(store, nil, nil)
	ctx := WithCurrentLocationID(context.Background(), store.currentLocationID)

	_, err := handler.Handle(ctx, map[string]any{
		"fact":             "The GM-only backstory remains secret.",
		"category":         "lore",
		"reveal_to_player": false,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if store.lastCreateFactParams.PlayerKnown {
		t.Fatal("expected reveal_to_player=false to create hidden fact")
	}
}
```

Adjust stub field names to match the current test file.

- [ ] **Step 6: Verify**

Run:

```bash
rtk go test ./internal/tools ./internal/world ./internal/state/sqlc -run 'TestEstablishFact.*PlayerKnown|TestSkeletonStoreAdapter_CreateWorldFactSetsSource|TestListPlayerKnownFacts' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add queries/world_facts.sql internal/state/sqlc/world_facts.sql.go internal/state/sqlc/querier.go internal/tools/establish_fact.go internal/tools/establish_fact_test.go internal/world/store_adapter.go internal/world/*test.go
git commit -m "fix: persist visible world facts as player known"
```

---

### Task 6: Add durable-claim audit

**Files:**
- Create: `internal/engine/durable_claims.go`
- Create: `internal/engine/durable_claims_test.go`

- [ ] **Step 1: Write tests**

Create `internal/engine/durable_claims_test.go`:

```go
package engine

import "testing"

func TestAuditDurableClaimsFlagsMissingMovement(t *testing.T) {
	issues := AuditDurableClaims("You arrive at the Lower Tier Corridor.", nil, []string{"move_player", "create_location"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimMovement {
		t.Fatalf("expected movement issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsAcceptsCreateLocationMovePlayerHere(t *testing.T) {
	applied := []AppliedToolCall{{Tool: "create_location", Result: []byte(`{"move_player_here":true,"location_id":"00000000-0000-0000-0000-000000000001"}`)}}
	issues := AuditDurableClaims("You arrive at the Lower Tier Corridor.", applied, []string{"move_player", "create_location"})
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsMissingQuest(t *testing.T) {
	issues := AuditDurableClaims("A new quest is added to your journal: Find the Core.", nil, []string{"create_quest"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimQuest {
		t.Fatalf("expected quest issue, got %#v", issues)
	}
}

func TestAuditDurableClaimsFlagsMissingFact(t *testing.T) {
	issues := AuditDurableClaims("You now know the Core Archive is unstable.", nil, []string{"establish_fact"})
	if len(issues) != 1 || issues[0].Kind != DurableClaimFact {
		t.Fatalf("expected fact issue, got %#v", issues)
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
rtk go test ./internal/engine -run TestAuditDurableClaims -count=1
```

Expected: FAIL because audit symbols do not exist.

- [ ] **Step 3: Implement audit helper**

Create `internal/engine/durable_claims.go`:

```go
package engine

import (
	"encoding/json"
	"strings"
)

type DurableClaimKind string

const (
	DurableClaimMovement DurableClaimKind = "movement"
	DurableClaimQuest    DurableClaimKind = "quest"
	DurableClaimFact     DurableClaimKind = "fact"
)

type DurableClaimIssue struct {
	Kind    DurableClaimKind
	Message string
}

func AuditDurableClaims(narrative string, applied []AppliedToolCall, advertised []string) []DurableClaimIssue {
	lower := strings.ToLower(narrative)
	appliedTools := map[string]bool{}
	createLocationMoved := false
	for _, call := range applied {
		appliedTools[call.Tool] = true
		if call.Tool == "create_location" {
			var data map[string]any
			if err := json.Unmarshal(call.Result, &data); err == nil {
				createLocationMoved, _ = data["move_player_here"].(bool)
			}
		}
	}

	issues := []DurableClaimIssue{}
	if claimsMovement(lower) && !appliedTools["move_player"] && !createLocationMoved {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimMovement, Message: "narrative claims player movement without persisted movement tool"})
	}
	if claimsQuest(lower) && !appliedTools["create_quest"] && !appliedTools["update_quest"] && !appliedTools["complete_objective"] {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimQuest, Message: "narrative claims quest journal change without quest tool"})
	}
	if claimsFact(lower) && !appliedTools["establish_fact"] && !appliedTools["revise_fact"] {
		issues = append(issues, DurableClaimIssue{Kind: DurableClaimFact, Message: "narrative claims known durable fact without fact tool"})
	}
	return issues
}

func claimsMovement(s string) bool {
	phrases := []string{"you arrive", "you step into", "you enter", "you move into", "you pass into", "you emerge into"}
	return containsAny(s, phrases)
}

func claimsQuest(s string) bool {
	phrases := []string{"quest is added", "new quest", "quest added", "added to your journal", "objective is added", "objective added"}
	return containsAny(s, phrases)
}

func claimsFact(s string) bool {
	phrases := []string{"you now know", "it is clear that", "you learn that", "you confirm that", "you establish that"}
	return containsAny(s, phrases)
}

func containsAny(s string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(s, phrase) {
			return true
		}
	}
	return false
}
```

This intentionally catches explicit durable language only. It should not infer quests from generic goals.

- [ ] **Step 4: Verify tests pass**

Run:

```bash
rtk go test ./internal/engine -run TestAuditDurableClaims -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/engine/durable_claims.go internal/engine/durable_claims_test.go
git commit -m "fix: audit durable narrative claims"
```

---

### Task 7: Add one repair pass before persistence

**Files:**
- Modify: `internal/engine/turn_processor.go:156-168`
- Modify: `internal/engine/turn_processor_test.go`
- Modify: `internal/prompt/gamemaster.txt:44-46`
- Modify: `internal/prompt/gamemaster_test.go`

- [ ] **Step 1: Add turn processor repair test**

In `internal/engine/turn_processor_test.go`, add a fake provider test using the existing test provider pattern:

```go
func TestProcessWithRecoveryRepairsDurableClaimBeforeReturning(t *testing.T) {
	provider := &sequenceProvider{responses: []llm.Response{
		{Content: "You arrive at the Lower Tier Corridor.", ToolCalls: nil},
		{Content: "You find a route toward the Lower Tier Corridor, but you have not committed to entering it yet.", ToolCalls: nil},
	}}
	processor := NewTurnProcessor(provider, tools.NewRegistry(), tools.NewValidator(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	narrative, applied, err := processor.ProcessWithRecovery(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "move to a new corridor"}}, []llm.Tool{tools.CreateLocationTool(), tools.MovePlayerTool()})
	if err != nil {
		t.Fatalf("ProcessWithRecovery returned error: %v", err)
	}
	if len(applied) != 0 {
		t.Fatalf("expected no tools from repair response, got %#v", applied)
	}
	if strings.Contains(strings.ToLower(narrative), "you arrive") {
		t.Fatalf("expected repaired narrative to avoid arrival claim, got %q", narrative)
	}
}
```

If `sequenceProvider` is not present, create a small fake in the test file with `Complete(ctx, messages, tools) (llm.Response, error)` that pops responses.

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
rtk go test ./internal/engine -run TestProcessWithRecoveryRepairsDurableClaimBeforeReturning -count=1
```

Expected: FAIL because no repair pass exists.

- [ ] **Step 3: Implement repair call**

Add helper in `turn_processor.go` after tool execution and before final return:

```go
issues := AuditDurableClaims(narrative, applied, advertisedToolNames(availableTools))
if len(issues) > 0 {
	repaired, repairErr := tp.requestDurableClaimRepair(ctx, messages, narrative, issues, availableTools)
	if repairErr != nil {
		tp.logger.Error("durable claim repair failed", "issues", len(issues), "error", repairErr)
	} else if repaired.Content != "" {
		narrative = repaired.Content
		for _, tc := range repaired.ToolCalls {
			result, execErr := tp.attemptToolCall(ctx, tc, allowed)
			if execErr != nil {
				tp.logger.Error("durable claim repair tool failed", "tool", tc.Name, "error", execErr)
				continue
			}
			if atc, encErr := buildAppliedToolCall(tc, result); encErr == nil {
				applied = append(applied, atc)
			}
		}
	}
}
```

Add the helper:

```go
func (tp *TurnProcessor) requestDurableClaimRepair(ctx context.Context, original []llm.Message, narrative string, issues []DurableClaimIssue, availableTools []llm.Tool) (llm.Response, error) {
	content := "The previous narrative made durable state claims that were not backed by successful tools. Either call the missing durable tools now, or rewrite the narrative so those claims are provisional and not yet true. Issues:\n"
	for _, issue := range issues {
		content += "- " + string(issue.Kind) + ": " + issue.Message + "\n"
	}
	content += "Previous narrative:\n" + narrative
	messages := append([]llm.Message{}, original...)
	messages = append(messages, llm.Message{Role: llm.RoleSystem, Content: content})
	return tp.provider.Complete(ctx, messages, availableTools)
}
```

After the repair call, run `AuditDurableClaims` again. If issues remain, replace the narrative with a safe generic line:

```go
narrative = "You pause at the threshold, the next step still uncertain. The situation remains unresolved until you commit to a concrete action."
```

Do not persist an unbacked durable claim.

- [ ] **Step 4: Tighten prompt wording**

In `internal/prompt/gamemaster.txt`, replace the `Tool availability` sentence that says quest tools appear only with quests/NPCs nearby. Use:

```text
**Tool availability:** Not all tools are available at all times. The system selects tools based on the current game context. If you can see create_quest, use it whenever the narrative explicitly adds a quest or objective. If you can see create_location with move_player_here, use it when the player enters a newly introduced place. If the needed tool is unavailable, keep the event provisional.
```

- [ ] **Step 5: Verify**

Run:

```bash
rtk go test ./internal/engine ./internal/prompt -run 'TestProcessWithRecoveryRepairsDurableClaimBeforeReturning|TestAuditDurableClaims|TestGameMasterPrompt' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/engine/turn_processor.go internal/engine/turn_processor_test.go internal/prompt/gamemaster.txt internal/prompt/gamemaster_test.go
git commit -m "fix: repair unbacked durable narrative claims"
```

---

### Task 8: Refresh final turn state before session-log persistence

**Files:**
- Modify: `internal/engine/runtime.go`
- Modify: `internal/engine/stage_persist.go`
- Modify: `internal/engine/runtime_test.go` or existing integration test file

- [ ] **Step 1: Add regression test**

Add an engine-level test with a fake state store and fake provider where `create_location(move_player_here=true)` applies, then assert the persisted session log location matches the moved-to location. Use existing runtime tests as scaffolding. The core assertion must be:

```go
if savedLog.LocationID != movedLocationID {
	t.Fatalf("session log location = %s, want moved location %s", savedLog.LocationID, movedLocationID)
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
rtk go test ./internal/engine -run TestPersistStageUsesPostTurnLocation -count=1
```

Expected: FAIL because `persistStage` uses pre-turn state.

- [ ] **Step 3: Derive final location from applied tool calls**

Add helper in `runtime.go` or a focused file:

```go
func finalLocationIDFromApplied(defaultID uuid.UUID, applied []AppliedToolCall) uuid.UUID {
	finalID := defaultID
	for _, call := range applied {
		if call.Tool != "move_player" && call.Tool != "create_location" {
			continue
		}
		var data map[string]any
		if err := json.Unmarshal(call.Result, &data); err != nil {
			continue
		}
		if call.Tool == "create_location" {
			moved, _ := data["move_player_here"].(bool)
			if !moved {
				continue
			}
		}
		if raw, ok := data["location_id"].(string); ok {
			if id, err := uuid.Parse(raw); err == nil {
				finalID = id
			}
		}
	}
	return finalID
}
```

In `persistStage`, use this helper when building `domain.SessionLog.LocationID`.

- [ ] **Step 4: Verify**

Run:

```bash
rtk go test ./internal/engine -run 'TestPersistStageUsesPostTurnLocation|TestFinalLocationIDFromApplied' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/engine/runtime.go internal/engine/stage_persist.go internal/engine/*test.go
git commit -m "fix: persist turn logs with final location"
```

---

### Task 9: Strengthen `create_location(move_player_here=true)` semantics

**Files:**
- Modify: `internal/tools/create_location.go:181-243,275-297`
- Modify: `internal/tools/create_location_test.go`

- [ ] **Step 1: Add movement-failure test**

In `internal/tools/create_location_test.go`, add:

```go
func TestCreateLocationMovePlayerHereFailsWhenMovementFails(t *testing.T) {
	currentLocationID := uuid.New()
	playerID := uuid.New()
	campaignID := uuid.New()
	store := newStubLocationStore(t)
	store.locations[currentLocationID] = statedb.Location{ID: pgUUID(currentLocationID), CampaignID: pgUUID(campaignID), Name: "The Ark"}
	store.updatePlayerLocationErr = errors.New("database unavailable")

	handler := NewCreateLocationHandler(store, nil, nil)
	ctx := WithCurrentPlayerCharacterID(WithCurrentLocationID(context.Background(), currentLocationID), playerID)
	_, err := handler.Handle(ctx, map[string]any{
		"name":             "Lower Tier Corridor",
		"description":      "A narrow corridor.",
		"region":           "Lower Tiers",
		"location_type":    "corridor",
		"move_player_here": true,
	})
	if err == nil {
		t.Fatal("expected movement failure to fail the tool")
	}
}
```

- [ ] **Step 2: Run test to verify current behavior**

Run:

```bash
rtk go test ./internal/tools -run TestCreateLocationMovePlayerHereFailsWhenMovementFails -count=1
```

Expected: PASS if current code already returns error on movement failure. If it fails because movement is warning-only, fix in step 3.

- [ ] **Step 3: Make movement failure a hard error**

Ensure `movePlayerHere` returns an error from `UpdatePlayerLocation`, and `Handle` returns that error before connection/memory work:

```go
movementData, movementWarning, err := h.movePlayerHere(ctx, movePlayerHere, playerCharacterID, locationID)
if err != nil {
	return nil, err
}
```

Keep `SetLocationPlayerVisited` as warning-only, but add `player_known` support if locations have such a column. If locations do not have player-known semantics, keep `visited_marked` as the durable discoverability flag.

- [ ] **Step 4: Verify result data includes movement fields only after success**

Add assertion to the existing successful `move_player_here` test:

```go
if result.Data["move_player_here"] != true {
	t.Fatalf("expected move_player_here=true, got %#v", result.Data["move_player_here"])
}
if result.Data["player_character_id"] != playerID.String() {
	t.Fatalf("expected player id in result, got %#v", result.Data["player_character_id"])
}
```

- [ ] **Step 5: Verify**

Run:

```bash
rtk go test ./internal/tools -run 'TestCreateLocation.*Move|TestCreateLocation.*Reuse|TestCreateLocation.*Warning' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/tools/create_location.go internal/tools/create_location_test.go
git commit -m "fix: make create location movement truthful"
```

---

### Task 10: Full local regression and verification

**Files:**
- No production files expected.
- Temporary script may be created under `/tmp/opencode` only.

- [ ] **Step 1: Run focused package tests**

Run:

```bash
rtk go test ./internal/domain ./internal/tools ./internal/world ./internal/config ./internal/memory ./internal/engine ./internal/prompt ./internal/state/sqlc -count=1
```

Expected: all selected packages pass.

- [ ] **Step 2: Run full suite**

Run:

```bash
rtk go test ./... -count=1
```

Expected: full suite passes.

- [ ] **Step 3: Restart local API with latest code**

Run:

```bash
pkill -f 'go run ./cmd/server' || true
pkill -f '/tmp/go-build.*/exe/server' || true
setsid go run ./cmd/server > .logs/local-api-final-regression.log 2>&1 < /dev/null &
```

Then poll:

```bash
curl -fsS http://localhost:18080/api/healthz
```

Expected: `{"engine_ready":true,"status":"ok"}`.

- [ ] **Step 4: Run a 10-turn local regression**

Use the same attribute path as the previous run. Include actions that force each fixed area:

1. Search the starting area for useful items.
2. Inspect an obvious hazard and record what is learned.
3. State a clear objective: find the safest route to the core or exit.
4. Move into a newly introduced corridor, explicitly entering it.
5. Ask for the next landmark and proceed if safe.
6. Establish a visible environmental fact.
7. Create/accept a concrete quest objective from the situation.
8. Move to another newly introduced threshold.
9. Check known facts and inventory.
10. Pause and summarize the current goal.

Expected mechanical outcomes:

- `current_location_id` changes after new-location movement.
- No duplicate same-ish location names for obvious variants.
- No `expected 1536 dimensions, not 768` warnings.
- At least one quest row exists when narrative says a quest/objective is tracked.
- `/facts` returns visible play facts, while hidden skeleton facts remain hidden.
- Session log location for movement turns matches final location.

- [ ] **Step 5: Inspect DB and API state**

Run targeted checks:

```bash
docker compose exec -T postgres psql -U edda -d edda -c "SELECT atttypmod AS embedding_typmod FROM pg_attribute WHERE attrelid = 'memories'::regclass AND attname = 'embedding';"
docker compose exec -T postgres psql -U edda -d edda -c "SELECT name, count(*) FROM locations GROUP BY name HAVING count(*) > 1;"
docker compose exec -T postgres psql -U edda -d edda -c "SELECT player_known, count(*) FROM world_facts GROUP BY player_known ORDER BY player_known;"
```

Expected: dimension `768`; no duplicate exact names; known facts count includes visible play facts.

- [ ] **Step 6: Commit final verification notes if docs changed**

If a regression script or docs file was added intentionally:

```bash
git add <intentional-doc-or-script-paths>
git commit -m "test: document ten-turn mechanical regression"
```

If no docs/script changed, do not create an empty commit.

---

## Self-review checklist

- Spec coverage:
  - Movement/new location: Tasks 6, 7, 8, 9, 10.
  - Duplicate locations and brittle starting location: Tasks 1, 2, 10.
  - Embedding dimension mismatch: Tasks 3, 4, 10.
  - Quest non-persistence: Tasks 6, 7, 10.
  - Facts visibility mismatch: Task 5, 10.
- Placeholder scan: every code-changing step includes concrete file paths, snippets, commands, and expected results.
- Type consistency: `player_known`, `move_player_here`, `location_id`, and `player_character_id` names match current tool/result conventions.

## Review gates

- Use `data-vault` after Tasks 4 and 5 because migrations and fact visibility affect persisted data.
- Use `oracle` after Tasks 6-9 because durable-claim repair changes engine behavior and can over-constrain narrative.
- Use `test-warden` only if the 10-turn regression remains flaky after deterministic unit/integration coverage.
