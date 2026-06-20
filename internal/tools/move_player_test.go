package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
)

type stubMovePlayerStore struct {
	result  *domain.MovePlayerResult
	err     error
	lastCmd domain.MovePlayerCommand
}

func (s *stubMovePlayerStore) MovePlayer(_ context.Context, cmd domain.MovePlayerCommand) (*domain.MovePlayerResult, error) {
	s.lastCmd = cmd
	return s.result, s.err
}

func TestRegisterMovePlayer(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterMovePlayer(reg, &stubMovePlayerStore{}); err != nil {
		t.Fatalf("register move_player: %v", err)
	}

	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != movePlayerToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, movePlayerToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 1 || required[0] != "location_id" {
		t.Fatalf("required schema = %#v, want [location_id]", required)
	}
}

func TestMovePlayerHandleConnectedLocation(t *testing.T) {
	currentLocationID := uuid.New()
	targetLocationID := uuid.New()
	playerID := uuid.New()
	campaignID := uuid.New()
	store := &stubMovePlayerStore{
		result: &domain.MovePlayerResult{
			ToLocationName:        "Ancient Gate",
			ToLocationDescription: "A ruined gate covered in glowing runes.",
			TravelTime:            "1 hour",
			Day:                   1,
			Hour:                  9,
			Minute:                0,
		},
	}
	h := NewMovePlayerHandler(store)
	ctx := WithCurrentCampaignID(WithCurrentPlayerCharacterID(WithCurrentLocationID(context.Background(), currentLocationID), playerID), campaignID)

	got, err := h.Handle(ctx, map[string]any{
		"location_id": targetLocationID.String(),
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if store.lastCmd.PlayerCharacterID != playerID || store.lastCmd.TargetLocationID != targetLocationID || store.lastCmd.CurrentLocationID != currentLocationID || store.lastCmd.CampaignID != campaignID {
		t.Fatalf("unexpected command: %+v", store.lastCmd)
	}
	if got.Data["name"] != "Ancient Gate" {
		t.Fatalf("result name = %v, want Ancient Gate", got.Data["name"])
	}
	if got.Data["description"] != "A ruined gate covered in glowing runes." {
		t.Fatalf("result description = %v, want expected description", got.Data["description"])
	}
}

func TestMovePlayerHandleMissingContext(t *testing.T) {
	store := &stubMovePlayerStore{result: &domain.MovePlayerResult{ToLocationName: "Ancient Gate"}}
	h := NewMovePlayerHandler(store)
	_, err := h.Handle(context.Background(), map[string]any{"location_id": uuid.New().String()})
	if err == nil || !strings.Contains(err.Error(), "current location id") {
		t.Fatalf("err = %v, want missing-context error", err)
	}
}

func TestMovePlayerHandleUnconnectedLocation(t *testing.T) {
	currentLocationID := uuid.New()
	targetLocationID := uuid.New()
	playerID := uuid.New()
	campaignID := uuid.New()
	store := &stubMovePlayerStore{err: errors.New("target location is not connected to current location")}
	h := NewMovePlayerHandler(store)
	ctx := WithCurrentCampaignID(WithCurrentPlayerCharacterID(WithCurrentLocationID(context.Background(), currentLocationID), playerID), campaignID)

	_, err := h.Handle(ctx, map[string]any{
		"location_id": targetLocationID.String(),
	})
	if err == nil {
		t.Fatal("expected unconnected location error")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Fatalf("error = %v, want unconnected-location message", err)
	}
}

func TestMovePlayerHandleToleratedWarnings(t *testing.T) {
	currentLocationID := uuid.New()
	targetLocationID := uuid.New()
	playerID := uuid.New()
	campaignID := uuid.New()
	store := &stubMovePlayerStore{result: &domain.MovePlayerResult{ToLocationName: "Frost Bridge", ToLocationDescription: "A narrow bridge", VisitedWarning: "visited failed", TimeWarning: "time skipped"}}
	h := NewMovePlayerHandler(store)
	ctx := WithCurrentCampaignID(WithCurrentPlayerCharacterID(WithCurrentLocationID(context.Background(), currentLocationID), playerID), campaignID)
	got, err := h.Handle(ctx, map[string]any{"location_id": targetLocationID.String()})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(got.Narrative, "warning") {
		t.Fatalf("narrative = %q, want warnings", got.Narrative)
	}
}

func TestMovePlayerHandleNonexistentLocation(t *testing.T) {
	currentLocationID := uuid.New()
	targetLocationID := uuid.New()
	playerID := uuid.New()
	campaignID := uuid.New()
	store := &stubMovePlayerStore{err: errors.New("target location not found in campaign")}
	h := NewMovePlayerHandler(store)
	ctx := WithCurrentCampaignID(WithCurrentPlayerCharacterID(WithCurrentLocationID(context.Background(), currentLocationID), playerID), campaignID)

	_, err := h.Handle(ctx, map[string]any{
		"location_id": targetLocationID.String(),
	})
	if err == nil {
		t.Fatal("expected nonexistent location error")
	}
	if !strings.Contains(err.Error(), "target location not found in campaign") {
		t.Fatalf("error = %v, want campaign-scoped not-found message", err)
	}
}

var _ MovePlayerStore = (*stubMovePlayerStore)(nil)
