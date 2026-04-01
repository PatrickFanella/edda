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

type stubUpdatePlayerStatsStore struct {
	player      *domain.PlayerCharacter
	getErr      error
	updateErr   error
	lastPlayer  uuid.UUID
	lastStats   json.RawMessage
	updateCalls int
}

func (s *stubUpdatePlayerStatsStore) GetPlayerCharacterByID(_ context.Context, _ uuid.UUID) (*domain.PlayerCharacter, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.player, nil
}

func (s *stubUpdatePlayerStatsStore) UpdatePlayerStats(_ context.Context, playerCharacterID uuid.UUID, stats json.RawMessage) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.lastPlayer = playerCharacterID
	s.lastStats = append([]byte(nil), stats...)
	s.updateCalls++
	return nil
}

func TestRegisterUpdatePlayerStats(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterUpdatePlayerStats(reg, &stubUpdatePlayerStatsStore{}); err != nil {
		t.Fatalf("register update_player_stats: %v", err)
	}

	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != updatePlayerStatsToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, updatePlayerStatsToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 3 || required[0] != "stat_name" || required[1] != "value" || required[2] != "operation" {
		t.Fatalf("required schema = %#v, want [stat_name value operation]", required)
	}
}

func TestUpdatePlayerStatsHandleSetOperation(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":10,"dexterity":12}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"stat_name": "strength",
		"value":     14,
		"operation": "set",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if got.Data["old_value"] != 10 {
		t.Fatalf("old_value = %v, want 10", got.Data["old_value"])
	}
	if got.Data["new_value"] != 14 {
		t.Fatalf("new_value = %v, want 14", got.Data["new_value"])
	}
	if store.lastPlayer != playerID {
		t.Fatalf("updated player_id = %s, want %s", store.lastPlayer, playerID)
	}

	var updated map[string]any
	if err := json.Unmarshal(store.lastStats, &updated); err != nil {
		t.Fatalf("unmarshal updated stats: %v", err)
	}
	if updated["strength"] != float64(14) {
		t.Fatalf("strength = %v, want 14", updated["strength"])
	}
}

func TestUpdatePlayerStatsHandleAddOperationWithClamp(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"dexterity":28}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"stat_name": "dexterity",
		"value":     10,
		"operation": "add",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["old_value"] != 28 {
		t.Fatalf("old_value = %v, want 28", got.Data["old_value"])
	}
	if got.Data["new_value"] != 30 {
		t.Fatalf("new_value = %v, want 30", got.Data["new_value"])
	}
}

func TestUpdatePlayerStatsHandleSubtractOperationWithClamp(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"wisdom":3}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"stat_name": "wisdom",
		"value":     10,
		"operation": "subtract",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["old_value"] != 3 {
		t.Fatalf("old_value = %v, want 3", got.Data["old_value"])
	}
	if got.Data["new_value"] != 1 {
		t.Fatalf("new_value = %v, want 1", got.Data["new_value"])
	}
}

func TestUpdatePlayerStatsHandleValidationErrors(t *testing.T) {
	playerID := uuid.New()
	baseStore := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":10}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(baseStore)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{
			name: "invalid stat name",
			args: map[string]any{"stat_name": "luck", "value": 1, "operation": "set"},
			want: "stat_name must be one of",
		},
		{
			name: "invalid operation",
			args: map[string]any{"stat_name": "strength", "value": 1, "operation": "multiply"},
			want: "operation must be one of",
		},
		{
			name: "missing stat in player stats",
			args: map[string]any{"stat_name": "wisdom", "value": 1, "operation": "set"},
			want: "does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.Handle(ctx, tc.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func TestUpdatePlayerStatsHandleRequiresPlayerContext(t *testing.T) {
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    uuid.New(),
			Stats: []byte(`{"strength":10}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)

	_, err := h.Handle(context.Background(), map[string]any{
		"stat_name": "strength",
		"value":     1,
		"operation": "add",
	})
	if err == nil {
		t.Fatal("expected missing context error")
	}
	if !strings.Contains(err.Error(), "requires current player character id in context") {
		t.Fatalf("error = %v, want context message", err)
	}
}

func TestUpdatePlayerStatsHandleStoreErrors(t *testing.T) {
	playerID := uuid.New()

	t.Run("get player error", func(t *testing.T) {
		h := NewUpdatePlayerStatsHandler(&stubUpdatePlayerStatsStore{
			getErr: errors.New("db read failed"),
		})
		ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

		_, err := h.Handle(ctx, map[string]any{
			"stat_name": "strength",
			"value":     1,
			"operation": "set",
		})
		if err == nil || !strings.Contains(err.Error(), "get player character") {
			t.Fatalf("error = %v, want get player character wrapper", err)
		}
	})

	t.Run("update stats error", func(t *testing.T) {
		h := NewUpdatePlayerStatsHandler(&stubUpdatePlayerStatsStore{
			player: &domain.PlayerCharacter{
				ID:    playerID,
				Stats: []byte(`{"strength":10}`),
			},
			updateErr: errors.New("db write failed"),
		})
		ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

		_, err := h.Handle(ctx, map[string]any{
			"stat_name": "strength",
			"value":     1,
			"operation": "set",
		})
		if err == nil || !strings.Contains(err.Error(), "update player stats") {
			t.Fatalf("error = %v, want update player stats wrapper", err)
		}
	})
}

func TestUpdatePlayerStatsHandlerWithBoundsSwapsInvalidBounds(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":18}`),
		},
	}
	h := NewUpdatePlayerStatsHandlerWithBounds(store, 30, 1)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"stat_name": "strength",
		"value":     40,
		"operation": "set",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["new_value"] != 30 {
		t.Fatalf("new_value = %v, want 30", got.Data["new_value"])
	}
}

func TestUpdatePlayerStatsHandleRejectsNonIntegerStoredStat(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":10.5}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"stat_name": "strength",
		"value":     1,
		"operation": "add",
	})
	if err == nil {
		t.Fatal("expected non-integer stat error")
	}
	if !strings.Contains(err.Error(), "not a valid integer") {
		t.Fatalf("error = %v, want integer validation message", err)
	}
}

func TestUpdatePlayerStatsHandleRejectsOutOfRangeStoredStat(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":1e100}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"stat_name": "strength",
		"value":     1,
		"operation": "add",
	})
	if err == nil {
		t.Fatal("expected out-of-range stat error")
	}
	if !strings.Contains(err.Error(), "not a valid integer") {
		t.Fatalf("error = %v, want integer validation message", err)
	}
}

func TestUpdatePlayerStatsHandleMultipleOperationsSequential(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":10,"dexterity":8}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	// First call: set strength = 20.
	got, err := h.Handle(ctx, map[string]any{
		"stat_name": "strength",
		"value":     20,
		"operation": "set",
	})
	if err != nil {
		t.Fatalf("first Handle: %v", err)
	}
	if got.Data["new_value"] != 20 {
		t.Fatalf("first new_value = %v, want 20", got.Data["new_value"])
	}

	// Rebuild store.player.Stats from what was persisted so the second call sees updated data.
	store.player.Stats = append([]byte(nil), store.lastStats...)

	// Second call: add 3 to dexterity.
	got, err = h.Handle(ctx, map[string]any{
		"stat_name": "dexterity",
		"value":     3,
		"operation": "add",
	})
	if err != nil {
		t.Fatalf("second Handle: %v", err)
	}
	if got.Data["new_value"] != 11 {
		t.Fatalf("second new_value = %v, want 11", got.Data["new_value"])
	}

	// Verify both stats are present and correct in the final persisted payload.
	var final map[string]any
	if err := json.Unmarshal(store.lastStats, &final); err != nil {
		t.Fatalf("unmarshal final stats: %v", err)
	}
	if final["strength"] != float64(20) {
		t.Fatalf("final strength = %v, want 20", final["strength"])
	}
	if final["dexterity"] != float64(11) {
		t.Fatalf("final dexterity = %v, want 11", final["dexterity"])
	}
	if store.updateCalls != 2 {
		t.Fatalf("updateCalls = %d, want 2", store.updateCalls)
	}
}

func TestUpdatePlayerStatsHandleMissingArgs(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":10}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{
			name: "missing stat_name",
			args: map[string]any{"value": 1, "operation": "set"},
			want: "stat_name is required",
		},
		{
			name: "missing value",
			args: map[string]any{"stat_name": "strength", "operation": "set"},
			want: "value is required",
		},
		{
			name: "missing operation",
			args: map[string]any{"stat_name": "strength", "value": 1},
			want: "operation is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.Handle(ctx, tc.args)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func TestUpdatePlayerStatsHandleNilPlayer(t *testing.T) {
	playerID := uuid.New()
	// store returns nil player, no error.
	store := &stubUpdatePlayerStatsStore{player: nil}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"stat_name": "strength",
		"value":     1,
		"operation": "set",
	})
	if err == nil {
		t.Fatal("expected nil-player error")
	}
	if !strings.Contains(err.Error(), "current player character does not exist") {
		t.Fatalf("error = %v, want 'current player character does not exist'", err)
	}
}

func TestUpdatePlayerStatsHandleEmptyStats(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: nil, // empty / unset stats
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"stat_name": "strength",
		"value":     10,
		"operation": "set",
	})
	if err == nil {
		t.Fatal("expected error for missing stat in empty stats")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %v, want 'does not exist'", err)
	}
}

func TestUpdatePlayerStatsHandleCaseInsensitiveStatName(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerStatsStore{
		player: &domain.PlayerCharacter{
			ID:    playerID,
			Stats: []byte(`{"strength":10}`),
		},
	}
	h := NewUpdatePlayerStatsHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"stat_name": " STRENGTH ", // leading/trailing space + uppercase
		"value":     15,
		"operation": "set",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["new_value"] != 15 {
		t.Fatalf("new_value = %v, want 15", got.Data["new_value"])
	}

	var updated map[string]any
	if err := json.Unmarshal(store.lastStats, &updated); err != nil {
		t.Fatalf("unmarshal updated stats: %v", err)
	}
	if updated["strength"] != float64(15) {
		t.Fatalf("strength = %v, want 15", updated["strength"])
	}
}

var _ UpdatePlayerStatsStore = (*stubUpdatePlayerStatsStore)(nil)
