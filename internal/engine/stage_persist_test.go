package engine

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/game"
)

func TestPersistStageUsesPostTurnLocation(t *testing.T) {
	defaultLocationID := uuid.New()
	newLocationID := uuid.New()
	state := &game.GameState{Player: domain.PlayerCharacter{CurrentLocationID: &defaultLocationID}}
	fake := &fakeStateManager{state: state}
	e := &Engine{state: fake}
	tc := &TurnContext{
		CampaignID:  uuid.New(),
		PlayerInput: "move north",
		Narrative:   "done",
		State:       state,
		Applied: []AppliedToolCall{{
			Tool:   "move_player",
			Result: json.RawMessage(`{"location_id":"` + newLocationID.String() + `"}`),
		}},
		Logger: testLogger(),
	}

	if err := e.persistStage()(context.Background(), tc); err != nil {
		t.Fatalf("persistStage() error = %v", err)
	}
	if len(fake.savedLogs) != 1 {
		t.Fatalf("saved log count = %d, want 1", len(fake.savedLogs))
	}
	if fake.savedLogs[0].LocationID == nil || *fake.savedLogs[0].LocationID != newLocationID {
		t.Fatalf("saved location = %v, want %v", fake.savedLogs[0].LocationID, newLocationID)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
