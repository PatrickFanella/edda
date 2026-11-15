package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/memory"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
)

type memorySchemaStub struct {
	dimension int
	err       error
	calls     int
}

func (s *memorySchemaStub) QueryRow(context.Context, string, ...any) pgx.Row {
	s.calls++
	return s
}

func (s *memorySchemaStub) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	*dest[0].(*int) = s.dimension
	return nil
}

// Only persistence is doubled: real tool handlers, HTTP embedders, searcher,
// tier-3 retrieval, assembly and turn processing run through production code.
type campaignMemoryStore struct {
	campaignID uuid.UUID
	facts      []statedb.WorldFact
	memories   []tools.CreateMemoryParams
	searches   int
}

func (s *campaignMemoryStore) GetLocationByID(context.Context, pgtype.UUID) (statedb.Location, error) {
	return statedb.Location{CampaignID: db.ToPgUUID(s.campaignID)}, nil
}

func (s *campaignMemoryStore) CreateFact(_ context.Context, p statedb.CreateFactParams) (statedb.WorldFact, error) {
	f := statedb.WorldFact{ID: db.ToPgUUID(uuid.New()), CampaignID: p.CampaignID, Fact: p.Fact, Category: p.Category, PlayerKnown: p.PlayerKnown}
	s.facts = append(s.facts, f)
	return f, nil
}

func (s *campaignMemoryStore) SetFactPlayerKnown(context.Context, pgtype.UUID) error { return nil }

func (s *campaignMemoryStore) CreateMemory(_ context.Context, p tools.CreateMemoryParams) error {
	s.memories = append(s.memories, p)
	return nil
}

func (s *campaignMemoryStore) SearchMemoriesBySimilarity(_ context.Context, p statedb.SearchMemoriesBySimilarityParams) ([]statedb.SearchMemoriesBySimilarityRow, error) {
	s.searches++
	var rows []statedb.SearchMemoriesBySimilarityRow
	for _, m := range s.memories {
		if m.CampaignID != db.FromPgUUID(p.CampaignID) {
			continue
		}
		// Fixtures use unit basis vectors, so 1-dot is their cosine distance.
		dot := float64(0)
		for i, q := range p.QueryEmbedding.Slice() {
			dot += float64(q * m.Embedding[i])
		}
		rows = append(rows, statedb.SearchMemoriesBySimilarityRow{
			ID: db.ToPgUUID(uuid.New()), CampaignID: p.CampaignID,
			Content: m.Content, MemoryType: m.MemoryType, Distance: 1 - dot,
			CreatedAt: pgtype.Timestamptz{Time: time.Unix(1, 0), Valid: true},
		})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Distance < rows[j].Distance })
	if len(rows) > int(p.LimitCount) {
		rows = rows[:p.LimitCount]
	}
	return rows, nil
}

func independentMemoryConfig(endpoint string) config.LLMConfig {
	return config.LLMConfig{
		Provider: "claude", Claude: config.ClaudeConfig{APIKey: "gm-test-key"},
		// Deliberately unusable legacy settings must not affect explicit memory.
		Ollama:          config.OllamaConfig{Endpoint: "invalid", EmbeddingModel: "wrong-model", APIKey: "wrong-key"},
		MemoryEmbedding: &config.MemoryEmbeddingConfig{Enabled: true, Provider: "ollama", Endpoint: endpoint, Model: "fixture-embed", Dimension: 3, APIKey: "embedding-test-key"},
	}
}

func TestSemanticMemoryNonOllamaGMRecordsAndRecallsOldFact(t *testing.T) {
	const fact = "The amber bell opens the north vault."
	const otherCampaignFact = "A private fact from another campaign."
	campaignID := uuid.New()
	store := &campaignMemoryStore{campaignID: campaignID}
	var inputMu sync.Mutex
	var embedInputs []string
	embedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/broker/api/embed" || r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer embedding-test-key" {
			t.Error("embedding request did not use the independent path and credentials")
		}
		var req struct{ Model, Input string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if req.Model != "fixture-embed" {
			t.Errorf("embedding model = %q", req.Model)
		}
		inputMu.Lock()
		embedInputs = append(embedInputs, req.Input)
		inputMu.Unlock()
		_, _ = w.Write([]byte(`{"embeddings":[[1,0,0]]}`))
	}))
	defer embedServer.Close()

	// Use the real Claude adapter against a deterministic local Messages API.
	gmRequests := make(chan string, 4)
	var gmCalls int
	gmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.Header.Get("x-api-key") != "gm-test-key" {
			t.Error("expected a Claude GM request with separate credentials")
		}
		var req json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		gmRequests <- string(req)
		gmCalls++
		if gmCalls == 1 {
			_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"You study the bell."},{"type":"tool_use","id":"fact-1","name":"establish_fact","input":{"fact":"The amber bell opens the north vault.","category":"mechanism"}}],"stop_reason":"tool_use"}`))
			return
		}
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"You wait in the tower."}],"stop_reason":"end_turn"}`))
	}))
	defer gmServer.Close()
	provider := llm.NewClaudeClient(gmServer.URL, "gm-test-key", "fixture-gm")
	cfg := independentMemoryConfig(embedServer.URL + "/broker")
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	opts, err := SemanticMemoryOptions(context.Background(), cfg, &memorySchemaStub{dimension: 3}, store, logger)
	if err != nil {
		t.Fatal(err)
	}
	state := &fakeStateManager{state: makePipelineState(campaignID)}
	reg := tools.NewRegistry()
	e := newPipelineTestEngine(state, provider, reg)
	for _, opt := range opts {
		opt(e)
	}
	if e.embedder == nil || e.searcher == nil || e.tier3 == nil {
		t.Fatal("startup must wire recording, search_memory and tier-3 context together")
	}
	if err := tools.RegisterEstablishFact(reg, store, store, e.embedder); err != nil {
		t.Fatal(err)
	}
	if err := tools.RegisterSearchMemory(reg, e.searcher); err != nil {
		t.Fatal(err)
	}
	// The validator snapshots tool definitions, as it does after production
	// startup has registered all tools.
	e.processor = NewTurnProcessor(provider, reg, tools.NewValidator(reg), logger)
	firstTurn, err := e.ProcessTurn(context.Background(), campaignID, "Record what the amber bell does.")
	if err != nil || len(firstTurn.AppliedToolCalls) != 1 || firstTurn.AppliedToolCalls[0].Tool != "establish_fact" {
		t.Fatalf("record fact through Claude turn: result=%+v, error=%v", firstTurn, err)
	}
	<-gmRequests // The first turn predates the stored memory.
	if len(store.memories) != 1 || len(store.facts) != 1 || !strings.Contains(store.memories[0].Content, fact) {
		t.Fatal("fact and its semantic memory must both be recorded")
	}
	// Older fact is absent from current state and the most recent ten turns.
	for i := 11; i <= 20; i++ {
		state.recentLogs = append(state.recentLogs, makeRecentDomainLog(campaignID, i, "Look at the clouds.", "Clouds pass overhead."))
	}
	// Add enough unrelated memories to exclude the old fact unless vectors rank it.
	for i := 0; i < 6; i++ {
		store.memories = append([]tools.CreateMemoryParams{{CampaignID: campaignID, Content: "Unrelated weather.", Embedding: []float32{0, 1, 0}, MemoryType: "world_fact"}}, store.memories...)
	}
	store.memories = append(store.memories, tools.CreateMemoryParams{CampaignID: uuid.New(), Content: otherCampaignFact, Embedding: []float32{1, 0, 0}, MemoryType: "world_fact"})
	turn, err := e.ProcessTurn(context.Background(), campaignID, "How do I open the north vault?")
	if err != nil || turn.Narrative != "You wait in the tower." {
		t.Fatalf("ProcessTurn() = %+v, %v", turn, err)
	}
	var request struct {
		System   string
		Messages []struct {
			Role    string
			Content []struct{ Text string }
		}
	}
	if err := json.Unmarshal([]byte(<-gmRequests), &request); err != nil {
		t.Fatal(err)
	}
	var retrieved string
	for _, msg := range request.Messages {
		for _, block := range msg.Content {
			if msg.Role == "user" && strings.Contains(block.Text, "## Retrieved Memories") {
				retrieved = block.Text
			}
		}
	}
	if !strings.Contains(retrieved, fact) || !strings.Contains(retrieved, "relevance: 1.00") || strings.Contains(retrieved, otherCampaignFact) || strings.Contains(request.System, fact) {
		t.Fatal("Claude turn context must contain the old relevant memory as an untrusted reference within its campaign")
	}
	if len(state.savedLogs) != 2 || state.savedLogs[1].TurnNumber != 21 {
		t.Fatal("memory-enabled turns did not persist normally")
	}
	search, err := reg.Execute(tools.WithCurrentCampaignID(context.Background(), campaignID), "search_memory", map[string]any{"query": "north vault", "limit": float64(1)})
	if err != nil || !search.Success {
		t.Fatalf("search_memory = %+v, %v", search, err)
	}
	searchJSON, err := json.Marshal(search.Data)
	if err != nil || !strings.Contains(string(searchJSON), fact) {
		t.Fatalf("search_memory did not recall recorded fact: %v", err)
	}
	inputMu.Lock()
	defer inputMu.Unlock()
	if len(embedInputs) != 4 || !strings.Contains(embedInputs[1], fact) || !strings.Contains(embedInputs[2], "How do I open the north vault?") || !strings.Contains(embedInputs[2], "Signal Tower") || store.searches != 3 {
		t.Fatalf("expected recording, composite turn query and explicit search; inputs=%v, searches=%d", embedInputs, store.searches)
	}
}

func TestSemanticMemoryStartupReportsDisabledAndInvalidSettings(t *testing.T) {
	for _, tt := range []struct {
		name, want string
		cfg        config.LLMConfig
		db         memorySchemaStub
		wantErr    bool
	}{
		{"disabled Ollama", "disabled", config.LLMConfig{Provider: "ollama", MemoryEmbedding: &config.MemoryEmbeddingConfig{}}, memorySchemaStub{}, false},
		{"unconfigured Claude", "no embedding provider configured", config.LLMConfig{Provider: "claude"}, memorySchemaStub{}, false},
		{"unsupported", "unsupported", config.LLMConfig{MemoryEmbedding: &config.MemoryEmbeddingConfig{Enabled: true, Provider: "openai"}}, memorySchemaStub{}, true},
		{"schema mismatch", "memory embedding dimension mismatch", independentMemoryConfig("http://localhost:11434"), memorySchemaStub{dimension: 768}, true},
		{"schema unavailable", "database offline", independentMemoryConfig("http://localhost:11434"), memorySchemaStub{err: errors.New("database offline")}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			opts, err := SemanticMemoryOptions(context.Background(), tt.cfg, &tt.db, nil, slog.New(slog.NewTextHandler(&logs, nil)))
			if len(opts) != 0 {
				t.Fatal("disabled or invalid memory must not attach partial capabilities")
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("startup error = %v, want %q", err, tt.want)
				}
			} else if err != nil || !strings.Contains(logs.String(), tt.want) || tt.db.calls != 0 {
				t.Fatalf("disabled startup: err=%v, logs=%q, database calls=%d", err, logs.String(), tt.db.calls)
			}
		})
	}
}

func TestSemanticMemoryUnavailableDoesNotUndoFactOrBlockTurn(t *testing.T) {
	for _, tt := range []struct {
		name, body, want string
		status           int
	}{
		{"offline", "unavailable", "status 503", http.StatusServiceUnavailable},
		{"unsupported API", "not found", "status 404", http.StatusNotFound},
		{"wrong model dimension", `{"embeddings":[[1,0]]}`, "dimension mismatch", http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = fmt.Fprint(w, tt.body)
			}))
			defer srv.Close()
			campaignID := uuid.New()
			store := &campaignMemoryStore{campaignID: campaignID}
			var logs bytes.Buffer
			opts, err := SemanticMemoryOptions(context.Background(), independentMemoryConfig(srv.URL), &memorySchemaStub{dimension: 3}, store, slog.New(slog.NewTextHandler(&logs, nil)))
			if err != nil {
				t.Fatal(err)
			}
			provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: "You pause by the tower."}})
			state := &fakeStateManager{state: makePipelineState(campaignID)}
			reg := tools.NewRegistry()
			e := newPipelineTestEngine(state, provider, reg)
			for _, opt := range opts {
				opt(e)
			}
			if err := tools.RegisterEstablishFact(reg, store, store, e.embedder); err != nil {
				t.Fatal(err)
			}
			ctx := tools.WithCurrentLocationID(context.Background(), state.state.CurrentLocation.ID)
			result, err := reg.Execute(ctx, "establish_fact", map[string]any{"fact": "The bell opens the vault.", "category": "mechanism"})
			if err != nil || !result.Success || !strings.Contains(fmt.Sprint(result.Data["memory_warning"]), tt.want) || len(store.facts) != 1 || len(store.memories) != 0 {
				t.Fatalf("durable fact should succeed with a memory warning: result=%+v, err=%v", result, err)
			}
			if _, err := e.ProcessTurn(context.Background(), campaignID, "Wait."); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(logs.String(), "tier3 memory retrieval failed") || !strings.Contains(logs.String(), tt.want) || len(state.savedLogs) != 1 {
				t.Fatalf("retrieval failure must be reported while the turn persists: %s", logs.String())
			}
		})
	}
}

var _ memory.MemorySearchStore = (*campaignMemoryStore)(nil)
var _ tools.MemoryStore = (*campaignMemoryStore)(nil)
