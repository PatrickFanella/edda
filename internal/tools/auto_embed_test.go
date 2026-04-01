package tools

import (
	"context"
	"errors"
	"testing"

	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// --- spy stubs ---

type spyEmbedder struct {
	called    bool
	text      string
	vec       []float32
	err       error
}

func (s *spyEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	s.called = true
	s.text = text
	return s.vec, s.err
}

func (s *spyEmbedder) BatchEmbed(_ context.Context, texts []string) ([][]float32, error) {
	return nil, errors.New("not implemented")
}

type spyAutoEmbedStore struct {
	called bool
	params statedb.CreateMemoryParams
	mem    statedb.Memory
	err    error
}

func (s *spyAutoEmbedStore) CreateMemory(_ context.Context, arg statedb.CreateMemoryParams) (statedb.Memory, error) {
	s.called = true
	s.params = arg
	return s.mem, s.err
}

// --- helpers ---

func dummyVec() []float32 {
	v := make([]float32, 768)
	v[0] = 1.0
	return v
}

func successResult() *ToolResult {
	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"id":          "00000000-0000-0000-0000-000000000001",
			"campaign_id": "00000000-0000-0000-0000-000000000099",
			"name":        "Thorn",
			"description": "A shadowy rogue",
		},
	}
}

func okHandler(result *ToolResult) func(context.Context, map[string]any) (*ToolResult, error) {
	return func(_ context.Context, _ map[string]any) (*ToolResult, error) {
		return result, nil
	}
}

func errHandler() func(context.Context, map[string]any) (*ToolResult, error) {
	return func(_ context.Context, _ map[string]any) (*ToolResult, error) {
		return nil, errors.New("tool exploded")
	}
}

// --- tests ---

func TestAutoEmbedder_WrapsSuccessfulTool(t *testing.T) {
	emb := &spyEmbedder{vec: dummyVec()}
	store := &spyAutoEmbedStore{}
	ae := NewAutoEmbedder(emb, store, nil)

	res := successResult()
	wrapped := ae.Wrap("create_npc", okHandler(res))

	got, err := wrapped(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != res {
		t.Fatal("expected original result returned")
	}
	if !emb.called {
		t.Fatal("embedder was not called")
	}
	if emb.text != "Thorn. A shadowy rogue" {
		t.Fatalf("summary = %q, want %q", emb.text, "Thorn. A shadowy rogue")
	}
	if !store.called {
		t.Fatal("store.CreateMemory was not called")
	}
	if store.params.MemoryType != "npc" {
		t.Fatalf("memory_type = %q, want %q", store.params.MemoryType, "npc")
	}
	if store.params.Content != "Thorn. A shadowy rogue" {
		t.Fatalf("content = %q, want %q", store.params.Content, "Thorn. A shadowy rogue")
	}
}

func TestAutoEmbedder_ToolFailureSkipsEmbed(t *testing.T) {
	emb := &spyEmbedder{vec: dummyVec()}
	store := &spyAutoEmbedStore{}
	ae := NewAutoEmbedder(emb, store, nil)

	wrapped := ae.Wrap("create_npc", errHandler())

	_, err := wrapped(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from handler")
	}
	if emb.called {
		t.Fatal("embedder should not be called on handler error")
	}
	if store.called {
		t.Fatal("store should not be called on handler error")
	}
}

func TestAutoEmbedder_NilEmbedderIsNoop(t *testing.T) {
	store := &spyAutoEmbedStore{}
	ae := NewAutoEmbedder(nil, store, nil)

	res := successResult()
	wrapped := ae.Wrap("create_npc", okHandler(res))

	got, err := wrapped(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != res {
		t.Fatal("expected original result returned")
	}
	if store.called {
		t.Fatal("store should not be called when embedder is nil")
	}
}

func TestAutoEmbedder_EmbedErrorDoesNotFailTool(t *testing.T) {
	emb := &spyEmbedder{err: errors.New("embed failure")}
	store := &spyAutoEmbedStore{}
	ae := NewAutoEmbedder(emb, store, nil)

	res := successResult()
	wrapped := ae.Wrap("create_npc", okHandler(res))

	got, err := wrapped(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != res {
		t.Fatal("expected original result despite embed error")
	}
	if store.called {
		t.Fatal("store should not be called when embedding fails")
	}
}

func TestAutoEmbedder_StoreErrorDoesNotFailTool(t *testing.T) {
	emb := &spyEmbedder{vec: dummyVec()}
	store := &spyAutoEmbedStore{err: errors.New("store failure")}
	ae := NewAutoEmbedder(emb, store, nil)

	res := successResult()
	wrapped := ae.Wrap("create_npc", okHandler(res))

	got, err := wrapped(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != res {
		t.Fatal("expected original result despite store error")
	}
	if !store.called {
		t.Fatal("store should have been called")
	}
}

func TestAutoEmbedder_UnknownToolSkipsEmbed(t *testing.T) {
	emb := &spyEmbedder{vec: dummyVec()}
	store := &spyAutoEmbedStore{}
	ae := NewAutoEmbedder(emb, store, nil)

	res := successResult()
	wrapped := ae.Wrap("unknown_tool", okHandler(res))

	got, err := wrapped(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != res {
		t.Fatal("expected original result returned")
	}
	if emb.called {
		t.Fatal("embedder should not be called for unknown tool")
	}
	if store.called {
		t.Fatal("store should not be called for unknown tool")
	}
}

func TestComposeSummary_VariousInputs(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want string
	}{
		{
			name: "name and description",
			data: map[string]any{"name": "Thorn", "description": "A rogue"},
			want: "Thorn. A rogue",
		},
		{
			name: "title only",
			data: map[string]any{"title": "The Lost Crown"},
			want: "The Lost Crown",
		},
		{
			name: "content and fact",
			data: map[string]any{"content": "Ancient lore", "fact": "Dragons exist"},
			want: "Ancient lore. Dragons exist",
		},
		{
			name: "all fields",
			data: map[string]any{
				"name": "X", "title": "Y", "description": "Z",
				"content": "C", "fact": "F",
			},
			want: "X. Y. Z. C. F",
		},
		{
			name: "empty data",
			data: map[string]any{"id": "123"},
			want: "",
		},
		{
			name: "nil data",
			data: nil,
			want: "",
		},
		{
			name: "empty strings ignored",
			data: map[string]any{"name": "", "description": "Some desc"},
			want: "Some desc",
		},
		{
			name: "non-string values ignored",
			data: map[string]any{"name": 42, "description": "Valid"},
			want: "Valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := composeSummary(tt.data)
			if got != tt.want {
				t.Fatalf("composeSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAutoEmbedder_MemoryTypeMapping(t *testing.T) {
	expected := map[string]string{
		"create_npc":             "npc",
		"create_location":        "location",
		"create_faction":         "faction",
		"create_lore":            "lore",
		"establish_fact":         "world_fact",
		"create_language":        "language",
		"create_belief_system":   "belief_system",
		"create_culture":         "culture",
		"create_city":            "city",
		"create_economic_system": "economic_system",
	}

	for toolName, wantType := range expected {
		t.Run(toolName, func(t *testing.T) {
			emb := &spyEmbedder{vec: dummyVec()}
			store := &spyAutoEmbedStore{}
			ae := NewAutoEmbedder(emb, store, nil)

			res := successResult()
			wrapped := ae.Wrap(toolName, okHandler(res))

			_, err := wrapped(context.Background(), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !store.called {
				t.Fatalf("store not called for tool %q", toolName)
			}
			if store.params.MemoryType != wantType {
				t.Fatalf("memory_type = %q, want %q", store.params.MemoryType, wantType)
			}
		})
	}
}

func TestAutoEmbedder_NilResultDataSkipsEmbed(t *testing.T) {
	emb := &spyEmbedder{vec: dummyVec()}
	store := &spyAutoEmbedStore{}
	ae := NewAutoEmbedder(emb, store, nil)

	res := &ToolResult{Success: true, Data: nil}
	wrapped := ae.Wrap("create_npc", okHandler(res))

	got, err := wrapped(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != res {
		t.Fatal("expected original result")
	}
	if emb.called {
		t.Fatal("embedder should not be called when Data is nil")
	}
}

func TestAutoEmbedder_EmptySummarySkipsEmbed(t *testing.T) {
	emb := &spyEmbedder{vec: dummyVec()}
	store := &spyAutoEmbedStore{}
	ae := NewAutoEmbedder(emb, store, nil)

	res := &ToolResult{
		Success: true,
		Data:    map[string]any{"id": "some-id", "campaign_id": "some-uuid"},
	}
	wrapped := ae.Wrap("create_npc", okHandler(res))

	got, err := wrapped(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != res {
		t.Fatal("expected original result")
	}
	if emb.called {
		t.Fatal("embedder should not be called when summary is empty")
	}
}
