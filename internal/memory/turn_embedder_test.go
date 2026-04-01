package memory_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/memory"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// --- stubs -----------------------------------------------------------

type fakeEmbedder struct {
	vec []float32
	err error
}

func (f *fakeEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return f.vec, f.err
}

func (f *fakeEmbedder) BatchEmbed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = f.vec
	}
	return out, f.err
}

type fakeStore struct {
	called bool
	params statedb.CreateMemoryParams
	mem    statedb.Memory
	err    error
}

func (f *fakeStore) CreateMemory(_ context.Context, arg statedb.CreateMemoryParams) (statedb.Memory, error) {
	f.called = true
	f.params = arg
	return f.mem, f.err
}

// --- helpers ---------------------------------------------------------

func testVec() []float32 { return []float32{0.1, 0.2, 0.3} }

func fullInput() memory.TurnSummaryInput {
	locID := uuid.New()
	return memory.TurnSummaryInput{
		CampaignID:   uuid.New(),
		PlayerInput:  "I search the room",
		Narrative:    "You find a hidden door",
		ToolsUsed:    []string{"search", "perception_check"},
		StateChanges: []string{"door_revealed", "xp_gained"},
		LocationID:   &locID,
		NPCsInvolved: []uuid.UUID{uuid.New(), uuid.New()},
		InGameTime:   "Dawn of Day 3",
	}
}

// --- tests -----------------------------------------------------------

func TestEmbedTurn_Success(t *testing.T) {
	vec := testVec()
	emb := &fakeEmbedder{vec: vec}
	store := &fakeStore{}
	te := memory.NewTurnEmbedder(emb, store)

	input := fullInput()
	if err := te.EmbedTurn(context.Background(), input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !store.called {
		t.Fatal("store.CreateMemory was not called")
	}

	p := store.params
	if p.MemoryType != "turn_summary" {
		t.Errorf("MemoryType = %q, want %q", p.MemoryType, "turn_summary")
	}
	if p.CampaignID != dbutil.ToPgtype(input.CampaignID) {
		t.Error("CampaignID mismatch")
	}
	if p.LocationID != dbutil.ToPgtype(*input.LocationID) {
		t.Error("LocationID mismatch")
	}
	if len(p.NpcsInvolved) != len(input.NPCsInvolved) {
		t.Errorf("NpcsInvolved len = %d, want %d", len(p.NpcsInvolved), len(input.NPCsInvolved))
	}
	if !p.InGameTime.Valid || p.InGameTime.String != input.InGameTime {
		t.Errorf("InGameTime = %v, want %q", p.InGameTime, input.InGameTime)
	}
	if got := p.Embedding.Slice(); len(got) != len(vec) {
		t.Errorf("Embedding dimension = %d, want %d", len(got), len(vec))
	}
}

func TestEmbedTurn_MinimalInput(t *testing.T) {
	emb := &fakeEmbedder{vec: testVec()}
	store := &fakeStore{}
	te := memory.NewTurnEmbedder(emb, store)

	input := memory.TurnSummaryInput{
		CampaignID:  uuid.New(),
		PlayerInput: "look around",
		Narrative:   "Nothing happens",
	}
	if err := te.EmbedTurn(context.Background(), input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := store.params
	if p.LocationID.Valid {
		t.Error("LocationID should be invalid for nil input")
	}
	if len(p.NpcsInvolved) != 0 {
		t.Errorf("NpcsInvolved len = %d, want 0", len(p.NpcsInvolved))
	}
	if p.InGameTime.Valid {
		t.Error("InGameTime should be invalid for empty input")
	}

	// Metadata should still have empty arrays, not null.
	var meta struct {
		ToolsUsed    []string `json:"tools_used"`
		StateChanges []string `json:"state_changes"`
	}
	if err := json.Unmarshal(p.Metadata, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if meta.ToolsUsed == nil || len(meta.ToolsUsed) != 0 {
		t.Errorf("tools_used = %v, want []", meta.ToolsUsed)
	}
	if meta.StateChanges == nil || len(meta.StateChanges) != 0 {
		t.Errorf("state_changes = %v, want []", meta.StateChanges)
	}
}

func TestEmbedTurn_EmbedError(t *testing.T) {
	embedErr := errors.New("model unavailable")
	emb := &fakeEmbedder{err: embedErr}
	store := &fakeStore{}
	te := memory.NewTurnEmbedder(emb, store)

	err := te.EmbedTurn(context.Background(), fullInput())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, embedErr) {
		t.Errorf("error = %v, want wrapped %v", err, embedErr)
	}
	if store.called {
		t.Error("store should not be called when embedding fails")
	}
}

func TestEmbedTurn_StoreError(t *testing.T) {
	storeErr := errors.New("db write failed")
	emb := &fakeEmbedder{vec: testVec()}
	store := &fakeStore{err: storeErr}
	te := memory.NewTurnEmbedder(emb, store)

	err := te.EmbedTurn(context.Background(), fullInput())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, storeErr) {
		t.Errorf("error = %v, want wrapped %v", err, storeErr)
	}
}

func TestEmbedTurn_SummaryContent(t *testing.T) {
	emb := &fakeEmbedder{vec: testVec()}
	store := &fakeStore{}
	te := memory.NewTurnEmbedder(emb, store)

	input := memory.TurnSummaryInput{
		CampaignID:   uuid.New(),
		PlayerInput:  "open chest",
		Narrative:    "gold coins spill out",
		ToolsUsed:    []string{"loot"},
		StateChanges: []string{"gold+100"},
	}
	if err := te.EmbedTurn(context.Background(), input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "Player: open chest\nOutcome: gold coins spill out\nActions: loot\nChanges: gold+100"
	if store.params.Content != want {
		t.Errorf("summary =\n%q\nwant\n%q", store.params.Content, want)
	}
}

func TestEmbedTurn_MetadataJSON(t *testing.T) {
	emb := &fakeEmbedder{vec: testVec()}
	store := &fakeStore{}
	te := memory.NewTurnEmbedder(emb, store)

	input := fullInput()
	if err := te.EmbedTurn(context.Background(), input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var meta struct {
		ToolsUsed    []string `json:"tools_used"`
		StateChanges []string `json:"state_changes"`
	}
	if err := json.Unmarshal(store.params.Metadata, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if len(meta.ToolsUsed) != len(input.ToolsUsed) {
		t.Errorf("tools_used len = %d, want %d", len(meta.ToolsUsed), len(input.ToolsUsed))
	}
	for i, tool := range input.ToolsUsed {
		if meta.ToolsUsed[i] != tool {
			t.Errorf("tools_used[%d] = %q, want %q", i, meta.ToolsUsed[i], tool)
		}
	}
	if len(meta.StateChanges) != len(input.StateChanges) {
		t.Errorf("state_changes len = %d, want %d", len(meta.StateChanges), len(input.StateChanges))
	}
	for i, sc := range input.StateChanges {
		if meta.StateChanges[i] != sc {
			t.Errorf("state_changes[%d] = %q, want %q", i, meta.StateChanges[i], sc)
		}
	}
}
