package engine

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock implementation – compile-time interface check
// ---------------------------------------------------------------------------

// mockEngine is a minimal stub used to verify that the GameEngine interface
// can be satisfied.
type mockEngine struct{}

var _ GameEngine = (*mockEngine)(nil)

func (m *mockEngine) ProcessTurn(_ context.Context, _ uuid.UUID, _ string) (*TurnResult, error) {
	return &TurnResult{}, nil
}

func (m *mockEngine) GetGameState(_ context.Context, _ uuid.UUID) (*GameState, error) {
	return &GameState{}, nil
}

func (m *mockEngine) NewCampaign(_ context.Context, _ uuid.UUID) (*domain.Campaign, error) {
	return &domain.Campaign{}, nil
}

func (m *mockEngine) LoadCampaign(_ context.Context, _ uuid.UUID) error {
	return nil
}

// ---------------------------------------------------------------------------
// Type construction tests
// ---------------------------------------------------------------------------

func TestTurnResult_FieldsAccessible(t *testing.T) {
	tr := TurnResult{
		Narrative: "You enter a dark cave.",
		AppliedToolCalls: []AppliedToolCall{
			{
				Tool:      "skill_check",
				Arguments: json.RawMessage(`{"skill":"perception","dc":15}`),
				Result:    json.RawMessage(`{"success":true}`),
			},
		},
		Choices: []Choice{
			{ID: "explore", Text: "Explore deeper into the cave"},
			{ID: "retreat", Text: "Retreat to the entrance"},
		},
		StateChanges: []StateChange{
			{
				Entity:   "location",
				EntityID: uuid.New(),
				Field:    "visited",
				OldValue: json.RawMessage(`false`),
				NewValue: json.RawMessage(`true`),
			},
		},
	}

	if tr.Narrative == "" {
		t.Error("expected non-empty narrative")
	}
	if len(tr.AppliedToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(tr.AppliedToolCalls))
	}
	if tr.AppliedToolCalls[0].Tool != "skill_check" {
		t.Errorf("expected tool 'skill_check', got %q", tr.AppliedToolCalls[0].Tool)
	}
	if len(tr.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(tr.Choices))
	}
	if len(tr.StateChanges) != 1 {
		t.Fatalf("expected 1 state change, got %d", len(tr.StateChanges))
	}
}

func TestGameState_FieldsAccessible(t *testing.T) {
	locID := uuid.New()
	pcID := uuid.New()

	gs := GameState{
		CurrentLocation: domain.Location{
			ID:   locID,
			Name: "Dark Cave",
		},
		PlayerCharacter: domain.PlayerCharacter{
			ID:   pcID,
			Name: "Elara",
		},
		NPCsPresent: []domain.NPC{
			{ID: uuid.New(), Name: "Goblin Scout", Alive: true},
		},
		ActiveQuests: []domain.Quest{
			{ID: uuid.New(), Title: "Find the Lost Amulet", Status: domain.QuestStatusActive},
		},
	}

	if gs.CurrentLocation.ID != locID {
		t.Error("location ID mismatch")
	}
	if gs.PlayerCharacter.ID != pcID {
		t.Error("player character ID mismatch")
	}
	if len(gs.NPCsPresent) != 1 {
		t.Fatalf("expected 1 NPC, got %d", len(gs.NPCsPresent))
	}
	if gs.NPCsPresent[0].Name != "Goblin Scout" {
		t.Errorf("expected NPC name 'Goblin Scout', got %q", gs.NPCsPresent[0].Name)
	}
	if len(gs.ActiveQuests) != 1 {
		t.Fatalf("expected 1 quest, got %d", len(gs.ActiveQuests))
	}
	if gs.ActiveQuests[0].Status != domain.QuestStatusActive {
		t.Errorf("expected quest status 'active', got %q", gs.ActiveQuests[0].Status)
	}
}

func TestMockEngine_SatisfiesInterface(t *testing.T) {
	var eng GameEngine = &mockEngine{}

	ctx := context.Background()
	cID := uuid.New()
	uID := uuid.New()

	tr, err := eng.ProcessTurn(ctx, cID, "look around")
	if err != nil {
		t.Fatalf("ProcessTurn: %v", err)
	}
	if tr == nil {
		t.Fatal("ProcessTurn returned nil TurnResult")
	}

	gs, err := eng.GetGameState(ctx, cID)
	if err != nil {
		t.Fatalf("GetGameState: %v", err)
	}
	if gs == nil {
		t.Fatal("GetGameState returned nil GameState")
	}

	c, err := eng.NewCampaign(ctx, uID)
	if err != nil {
		t.Fatalf("NewCampaign: %v", err)
	}
	if c == nil {
		t.Fatal("NewCampaign returned nil Campaign")
	}

	if err := eng.LoadCampaign(ctx, cID); err != nil {
		t.Fatalf("LoadCampaign: %v", err)
	}
}
