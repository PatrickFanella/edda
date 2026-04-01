package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/domain"
)

type stubUpdatePlayerStatusStore struct {
	player      *domain.PlayerCharacter
	getErr      error
	updateErr   error
	lastPlayer  uuid.UUID
	lastStatus  string
	updateCalls int
}

func (s *stubUpdatePlayerStatusStore) GetPlayerCharacterByID(_ context.Context, playerCharacterID uuid.UUID) (*domain.PlayerCharacter, error) {
	s.lastPlayer = playerCharacterID
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.player, nil
}

func (s *stubUpdatePlayerStatusStore) UpdatePlayerStatus(_ context.Context, playerCharacterID uuid.UUID, status string) error {
	s.lastPlayer = playerCharacterID
	s.lastStatus = status
	s.updateCalls++
	if s.updateErr != nil {
		return s.updateErr
	}
	return nil
}

func TestRegisterUpdatePlayerStatus(t *testing.T) {
	reg := NewRegistry()
	store := &stubUpdatePlayerStatusStore{}
	if err := RegisterUpdatePlayerStatus(reg, store); err != nil {
		t.Fatalf("register update_player_status: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != updatePlayerStatusToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, updatePlayerStatusToolName)
	}
}

func TestUpdatePlayerStatusStacksStatuses(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: `[{"status":"poisoned","duration":{"unit":"turns","value":"2"}}]`,
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	result, err := h.Handle(ctx, map[string]any{
		"status": "cursed",
		"duration": map[string]any{
			"unit":  "in_game_time",
			"value": "10 minutes",
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success=true")
	}
	if store.updateCalls != 1 {
		t.Fatalf("UpdatePlayerStatus calls = %d, want 1", store.updateCalls)
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("persisted statuses count = %d, want 2", len(statuses))
	}
	if statuses[0].Status != "poisoned" || statuses[1].Status != "cursed" {
		t.Fatalf("persisted statuses = %+v, want poisoned and cursed", statuses)
	}
}

func TestUpdatePlayerStatusRefreshesExistingDuration(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: `[{"status":"poisoned","duration":{"unit":"turns","value":"1"}}]`,
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"status": "poisoned",
		"duration": map[string]any{
			"unit":  "turns",
			"value": "4",
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("persisted statuses count = %d, want 1", len(statuses))
	}
	if statuses[0].Duration == nil || statuses[0].Duration.Value != "4" {
		t.Fatalf("refreshed duration = %+v, want value 4", statuses[0].Duration)
	}
}

func TestUpdatePlayerStatusHealthyClearsNegativeStatuses(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: `[{"status":"poisoned"},{"status":"cursed"},{"status":"resting"}]`,
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{"status": "healthy"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("persisted statuses count = %d, want 2", len(statuses))
	}
	if statuses[0].Status != "resting" || statuses[1].Status != "healthy" {
		t.Fatalf("persisted statuses = %+v, want resting and healthy", statuses)
	}
}

func TestUpdatePlayerStatusDeadIsTerminal(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: `[{"status":"dead"}]`,
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{"status": "resting"})
	if err == nil || !strings.Contains(err.Error(), "dead player character") {
		t.Fatalf("expected dead terminal error, got %v", err)
	}
	if store.updateCalls != 0 {
		t.Fatalf("UpdatePlayerStatus calls = %d, want 0", store.updateCalls)
	}
}

func TestUpdatePlayerStatusDeadSetsGameOver(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{ID: playerID, Status: "resting"},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	result, err := h.Handle(ctx, map[string]any{"status": "dead"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got, _ := result.Data["game_over"].(bool); !got {
		t.Fatalf("game_over = %v, want true", result.Data["game_over"])
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 1 || statuses[0].Status != "dead" {
		t.Fatalf("persisted statuses = %+v, want only dead", statuses)
	}
	if statuses[0].Duration != nil {
		t.Fatalf("dead status duration = %+v, want nil", statuses[0].Duration)
	}
}

func TestUpdatePlayerStatusPreservesCombatMode(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: "in_combat",
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	result, err := h.Handle(ctx, map[string]any{
		"status": "poisoned",
		"duration": map[string]any{
			"unit":  "turns",
			"value": "2",
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got, _ := result.Data["mode"].(string); got != "in_combat" {
		t.Fatalf("mode = %q, want in_combat", got)
	}

	var persisted playerStatusState
	if err := json.Unmarshal([]byte(store.lastStatus), &persisted); err != nil {
		t.Fatalf("unmarshal persisted status object: %v", err)
	}
	if persisted.Mode != "in_combat" {
		t.Fatalf("persisted mode = %q, want in_combat", persisted.Mode)
	}
	if len(persisted.Conditions) != 1 || persisted.Conditions[0].Status != "poisoned" {
		t.Fatalf("persisted conditions = %+v, want one poisoned", persisted.Conditions)
	}
}

func TestUpdatePlayerStatusResponseDurationMatchesPersisted(t *testing.T) {
	cases := []string{"healthy", "dead"}
	for _, statusName := range cases {
		t.Run(statusName, func(t *testing.T) {
			playerID := uuid.New()
			store := &stubUpdatePlayerStatusStore{
				player: &domain.PlayerCharacter{
					ID:     playerID,
					Status: `[{"status":"poisoned","duration":{"unit":"turns","value":"3"}}]`,
				},
			}
			h := NewUpdatePlayerStatusHandler(store)
			ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

			result, err := h.Handle(ctx, map[string]any{
				"status": statusName,
				"duration": map[string]any{
					"unit":  "turns",
					"value": "99",
				},
			})
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if _, exists := result.Data["duration"]; exists {
				t.Fatalf("response duration should be omitted for %s, got %v", statusName, result.Data["duration"])
			}

			var statuses []playerStatusEntry
			if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
				t.Fatalf("unmarshal persisted status: %v", err)
			}
			for _, entry := range statuses {
				if entry.Status == statusName && entry.Duration != nil {
					t.Fatalf("persisted %s duration = %+v, want nil", statusName, entry.Duration)
				}
			}
			if statusName == "dead" {
				if len(statuses) != 1 || statuses[0].Status != "dead" {
					t.Fatalf("persisted statuses = %+v, want only dead", statuses)
				}
			}
		})
	}
}

func TestUpdatePlayerStatusErrors(t *testing.T) {
	playerID := uuid.New()
	h := NewUpdatePlayerStatusHandler(&stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{ID: playerID},
	})

	_, err := h.Handle(context.Background(), map[string]any{"status": "poisoned"})
	if err == nil || !strings.Contains(err.Error(), "current player character id in context") {
		t.Fatalf("expected missing context error, got %v", err)
	}

	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err = h.Handle(ctx, map[string]any{"status": "unknown"})
	if err == nil || !strings.Contains(err.Error(), "status must be one of") {
		t.Fatalf("expected status validation error, got %v", err)
	}

	_, err = h.Handle(ctx, map[string]any{
		"status": "poisoned",
		"duration": map[string]any{
			"unit":  "days",
			"value": "2",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duration.unit must be one of") {
		t.Fatalf("expected duration unit validation error, got %v", err)
	}

	_, err = h.Handle(ctx, map[string]any{
		"status":   "poisoned",
		"duration": "bad",
	})
	if err == nil || !strings.Contains(err.Error(), "duration must be an object") {
		t.Fatalf("expected duration object validation error, got %v", err)
	}

	_, err = h.Handle(ctx, map[string]any{
		"status": "poisoned",
		"duration": map[string]any{
			"unit":  "turns",
			"value": "   ",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duration.value") {
		t.Fatalf("expected duration value validation error, got %v", err)
	}
}

func TestUpdatePlayerStatusStoreErrors(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	getErrStore := &stubUpdatePlayerStatusStore{getErr: errors.New("boom")}
	_, err := NewUpdatePlayerStatusHandler(getErrStore).Handle(ctx, map[string]any{"status": "poisoned"})
	if err == nil || !strings.Contains(err.Error(), "get player character") {
		t.Fatalf("expected get error, got %v", err)
	}

	updateErrStore := &stubUpdatePlayerStatusStore{
		player:    &domain.PlayerCharacter{ID: playerID},
		updateErr: errors.New("write failed"),
	}
	_, err = NewUpdatePlayerStatusHandler(updateErrStore).Handle(ctx, map[string]any{"status": "poisoned"})
	if err == nil || !strings.Contains(err.Error(), "update player status") {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestUpdatePlayerStatusHandleNilPlayer(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{player: nil}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{"status": "poisoned"})
	if err == nil || !strings.Contains(err.Error(), "current player character does not exist") {
		t.Fatalf("expected nil player error, got %v", err)
	}
}

func TestUpdatePlayerStatusHandleMissingStatus(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{ID: playerID},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "status is required") {
		t.Fatalf("expected missing status error, got %v", err)
	}
}

func TestUpdatePlayerStatusHandleNewStatusRemovesHealthy(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: `[{"status":"healthy"}]`,
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"status": "poisoned",
		"duration": map[string]any{
			"unit":  "turns",
			"value": "2",
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("persisted statuses count = %d, want 1", len(statuses))
	}
	if statuses[0].Status != "poisoned" {
		t.Fatalf("persisted status = %q, want poisoned", statuses[0].Status)
	}
}

func TestUpdatePlayerStatusHandleDeadClearsAllPrior(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: `[{"status":"poisoned"},{"status":"cursed"},{"status":"resting"}]`,
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	result, err := h.Handle(ctx, map[string]any{"status": "dead"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got, _ := result.Data["game_over"].(bool); !got {
		t.Fatalf("game_over = %v, want true", result.Data["game_over"])
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 1 || statuses[0].Status != "dead" {
		t.Fatalf("persisted statuses = %+v, want only dead", statuses)
	}
}

func TestUpdatePlayerStatusHandleAddStatusWithoutDuration(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{ID: playerID},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{"status": "cursed"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("persisted statuses count = %d, want 1", len(statuses))
	}
	if statuses[0].Status != "cursed" {
		t.Fatalf("persisted status = %q, want cursed", statuses[0].Status)
	}
	if statuses[0].Duration != nil {
		t.Fatalf("persisted duration = %+v, want nil", statuses[0].Duration)
	}
}

func TestUpdatePlayerStatusHandleRefreshKeepsOldDurationWhenNewOmitted(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatusStore{
		player: &domain.PlayerCharacter{
			ID:     playerID,
			Status: `[{"status":"poisoned","duration":{"unit":"turns","value":"3"}}]`,
		},
	}
	h := NewUpdatePlayerStatusHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{"status": "poisoned"})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var statuses []playerStatusEntry
	if err := json.Unmarshal([]byte(store.lastStatus), &statuses); err != nil {
		t.Fatalf("unmarshal persisted status: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("persisted statuses count = %d, want 1", len(statuses))
	}
	if statuses[0].Duration == nil || statuses[0].Duration.Unit != "turns" || statuses[0].Duration.Value != "3" {
		t.Fatalf("persisted duration = %+v, want turns/3", statuses[0].Duration)
	}
}

var _ UpdatePlayerStatusStore = (*stubUpdatePlayerStatusStore)(nil)

func TestParsePersistedStatusStateCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantMode  string
		wantCount int
	}{
		{name: "empty", raw: "", wantMode: "", wantCount: 0},
		{name: "legacy mode", raw: "in_combat", wantMode: "in_combat", wantCount: 0},
		{name: "legacy array", raw: `[{"status":"cursed"}]`, wantMode: "", wantCount: 1},
		{name: "new object", raw: `{"mode":"active","conditions":[{"status":"poisoned"}]}`, wantMode: "active", wantCount: 1},
		{name: "legacy scalar condition", raw: "resting", wantMode: "", wantCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePersistedStatusState(tt.raw)
			if err != nil {
				t.Fatalf("parsePersistedStatusState: %v", err)
			}
			if got.Mode != tt.wantMode {
				t.Fatalf("mode = %q, want %q", got.Mode, tt.wantMode)
			}
			if len(got.Conditions) != tt.wantCount {
				t.Fatalf("conditions count = %d, want %d", len(got.Conditions), tt.wantCount)
			}
		})
	}

	_, err := parsePersistedStatusState("{bad json")
	if err == nil {
		t.Fatal("expected parse error for invalid object JSON")
	}
}

func TestMarshalPersistedStatusStateCompatibility(t *testing.T) {
	arrayPayload, err := marshalPersistedStatusState(playerStatusState{
		Conditions: []playerStatusEntry{{Status: "resting"}},
	})
	if err != nil {
		t.Fatalf("marshalPersistedStatusState array: %v", err)
	}
	if !strings.HasPrefix(arrayPayload, "[") {
		t.Fatalf("array payload = %q, want JSON array", arrayPayload)
	}

	objectPayload, err := marshalPersistedStatusState(playerStatusState{
		Mode:       "in_combat",
		Conditions: []playerStatusEntry{{Status: "poisoned"}},
	})
	if err != nil {
		t.Fatalf("marshalPersistedStatusState object: %v", err)
	}
	if !strings.HasPrefix(objectPayload, "{") {
		t.Fatalf("object payload = %q, want JSON object", objectPayload)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(objectPayload), &decoded); err != nil {
		t.Fatalf("unmarshal object payload: %v", err)
	}
	if decoded["mode"] != "in_combat" {
		t.Fatalf("decoded mode = %v, want in_combat", decoded["mode"])
	}
}
