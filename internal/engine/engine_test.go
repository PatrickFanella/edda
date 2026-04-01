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

func (m *mockEngine) ProcessTurnStream(_ context.Context, _ uuid.UUID, _ string) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent)
	close(ch)
	return ch, nil
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

func TestTurnResult_JSONRoundTrip(t *testing.T) {
	tr := &TurnResult{
		Narrative: "The door creaks open.",
		AppliedToolCalls: []AppliedToolCall{
			{
				Tool:      "skill_check",
				Arguments: json.RawMessage(`{"skill":"perception","dc":15}`),
				Result:    json.RawMessage(`{"success":true}`),
			},
		},
		Choices: []Choice{
			{ID: "enter", Text: "Enter the room"},
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

	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("failed to marshal TurnResult: %v", err)
	}

	var got TurnResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal TurnResult: %v", err)
	}

	if got.Narrative != tr.Narrative {
		t.Errorf("narrative mismatch: got %q, want %q", got.Narrative, tr.Narrative)
	}
	if len(got.AppliedToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(got.AppliedToolCalls))
	}
	if got.AppliedToolCalls[0].Tool != "skill_check" {
		t.Errorf("tool name mismatch: got %q", got.AppliedToolCalls[0].Tool)
	}
	if len(got.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(got.Choices))
	}
	if len(got.StateChanges) != 1 {
		t.Fatalf("expected 1 state change, got %d", len(got.StateChanges))
	}
}

func TestGameState_JSONRoundTrip(t *testing.T) {
	gs := &GameState{
		CurrentLocation: domain.Location{
			ID:   uuid.New(),
			Name: "Dark Cave",
		},
		PlayerCharacter: domain.PlayerCharacter{
			ID:   uuid.New(),
			Name: "Elara",
		},
		NPCsPresent: []domain.NPC{
			{ID: uuid.New(), Name: "Goblin Scout", Alive: true},
		},
		ActiveQuests: []domain.Quest{
			{ID: uuid.New(), Title: "Find the Lost Amulet", Status: domain.QuestStatusActive},
		},
	}

	data, err := json.Marshal(gs)
	if err != nil {
		t.Fatalf("failed to marshal GameState: %v", err)
	}

	var got GameState
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal GameState: %v", err)
	}

	if got.CurrentLocation.Name != "Dark Cave" {
		t.Errorf("location name mismatch: got %q", got.CurrentLocation.Name)
	}
	if got.PlayerCharacter.Name != "Elara" {
		t.Errorf("player character name mismatch: got %q", got.PlayerCharacter.Name)
	}
	if len(got.NPCsPresent) != 1 {
		t.Fatalf("expected 1 NPC, got %d", len(got.NPCsPresent))
	}
	if got.NPCsPresent[0].Name != "Goblin Scout" {
		t.Errorf("NPC name mismatch: got %q", got.NPCsPresent[0].Name)
	}
	if len(got.ActiveQuests) != 1 {
		t.Fatalf("expected 1 quest, got %d", len(got.ActiveQuests))
	}
	if got.ActiveQuests[0].Title != "Find the Lost Amulet" {
		t.Errorf("quest title mismatch: got %q", got.ActiveQuests[0].Title)
	}
}
