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

type stubAbilityStore struct {
	player    *domain.PlayerCharacter
	getErr    error
	updateErr error

	lastPlayerID  uuid.UUID
	lastAbilities json.RawMessage
}

func (s *stubAbilityStore) GetPlayerCharacterByID(_ context.Context, _ uuid.UUID) (*domain.PlayerCharacter, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.player, nil
}

func (s *stubAbilityStore) UpdatePlayerAbilities(_ context.Context, playerCharacterID uuid.UUID, abilities json.RawMessage) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.lastPlayerID = playerCharacterID
	s.lastAbilities = append([]byte(nil), abilities...)
	return nil
}

func TestRegisterAddAbility(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterAddAbility(reg, &stubAbilityStore{}); err != nil {
		t.Fatalf("register add_ability: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != addAbilityToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, addAbilityToolName)
	}
}

func TestRegisterRemoveAbility(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterRemoveAbility(reg, &stubAbilityStore{}); err != nil {
		t.Fatalf("register remove_ability: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != removeAbilityToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, removeAbilityToolName)
	}
}

func TestAddAbilityHandleSuccess(t *testing.T) {
	playerID := uuid.New()
	store := &stubAbilityStore{
		player: &domain.PlayerCharacter{
			ID:        playerID,
			Abilities: []byte(`[]`),
		},
	}
	h := NewAddAbilityHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"name":        "Second Wind",
		"description": "Recover stamina",
		"type":        "active",
		"cooldown":    2,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !got.Success {
		t.Fatal("expected success")
	}
	if store.lastPlayerID != playerID {
		t.Fatalf("updated player id = %s, want %s", store.lastPlayerID, playerID)
	}

	var abilities []map[string]any
	if err := json.Unmarshal(store.lastAbilities, &abilities); err != nil {
		t.Fatalf("unmarshal abilities: %v", err)
	}
	if len(abilities) != 1 {
		t.Fatalf("abilities length = %d, want 1", len(abilities))
	}
	if abilities[0]["name"] != "Second Wind" {
		t.Fatalf("ability name = %v, want Second Wind", abilities[0]["name"])
	}
	if abilities[0]["type"] != "active" {
		t.Fatalf("ability type = %v, want active", abilities[0]["type"])
	}
	if abilities[0]["cooldown"] != float64(2) {
		t.Fatalf("ability cooldown = %v, want 2", abilities[0]["cooldown"])
	}
}

func TestAddAbilityHandleTrimsNameAndDescription(t *testing.T) {
	playerID := uuid.New()
	store := &stubAbilityStore{
		player: &domain.PlayerCharacter{
			ID:        playerID,
			Abilities: []byte(`[]`),
		},
	}
	h := NewAddAbilityHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"name":        "  Second Wind  ",
		"description": "  Recover stamina  ",
		"type":        "active",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["name"] != "Second Wind" {
		t.Fatalf("name = %v, want trimmed value", got.Data["name"])
	}
	if got.Data["description"] != "Recover stamina" {
		t.Fatalf("description = %v, want trimmed value", got.Data["description"])
	}
}

func TestAddAbilityHandleDuplicateRejected(t *testing.T) {
	playerID := uuid.New()
	store := &stubAbilityStore{
		player: &domain.PlayerCharacter{
			ID:        playerID,
			Abilities: []byte(`[{"name":"Second Wind","description":"Recover stamina","type":"active"}]`),
		},
	}
	h := NewAddAbilityHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"name":        "second wind",
		"description": "Duplicate",
		"type":        "active",
	})
	if err == nil {
		t.Fatal("expected duplicate ability error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want duplicate message", err)
	}
}

func TestRemoveAbilityHandleSuccessAndNotFound(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("success", func(t *testing.T) {
		store := &stubAbilityStore{
			player: &domain.PlayerCharacter{
				ID: playerID,
				Abilities: []byte(`[
					{"name":"Second Wind","description":"Recover stamina","type":"active","cooldown":2},
					{"name":"Toughness","description":"Passive resilience","type":"passive"}
				]`),
			},
		}
		h := NewRemoveAbilityHandler(store)
		got, err := h.Handle(ctx, map[string]any{
			"ability_name": "Second Wind",
		})
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if !got.Success {
			t.Fatal("expected success")
		}

		var abilities []map[string]any
		if err := json.Unmarshal(store.lastAbilities, &abilities); err != nil {
			t.Fatalf("unmarshal abilities: %v", err)
		}
		if len(abilities) != 1 {
			t.Fatalf("abilities length = %d, want 1", len(abilities))
		}
		if abilities[0]["name"] != "Toughness" {
			t.Fatalf("remaining ability name = %v, want Toughness", abilities[0]["name"])
		}
	})

	t.Run("not found", func(t *testing.T) {
		store := &stubAbilityStore{
			player: &domain.PlayerCharacter{
				ID:        playerID,
				Abilities: []byte(`[{"name":"Toughness","description":"Passive resilience","type":"passive"}]`),
			},
		}
		h := NewRemoveAbilityHandler(store)
		_, err := h.Handle(ctx, map[string]any{
			"ability_name": "Second Wind",
		})
		if err == nil {
			t.Fatal("expected not found error")
		}
		if !strings.Contains(err.Error(), "was not found") {
			t.Fatalf("error = %v, want not-found message", err)
		}
	})
}

func TestAbilityHandlersValidationAndWrappedErrors(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("add requires player context", func(t *testing.T) {
		h := NewAddAbilityHandler(&stubAbilityStore{
			player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
		})
		_, err := h.Handle(context.Background(), map[string]any{
			"name":        "Dash",
			"description": "Move quickly",
			"type":        "active",
		})
		if err == nil || !strings.Contains(err.Error(), "requires current player character id in context") {
			t.Fatalf("error = %v, want context error", err)
		}
	})

	t.Run("add invalid type", func(t *testing.T) {
		h := NewAddAbilityHandler(&stubAbilityStore{
			player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
		})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Dash",
			"description": "Move quickly",
			"type":        "toggle",
		})
		if err == nil || !strings.Contains(err.Error(), "type must be one of") {
			t.Fatalf("error = %v, want type validation error", err)
		}
	})

	t.Run("add rejects whitespace name", func(t *testing.T) {
		h := NewAddAbilityHandler(&stubAbilityStore{
			player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
		})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "   ",
			"description": "Move quickly",
			"type":        "active",
		})
		if err == nil || !strings.Contains(err.Error(), "name cannot be empty or whitespace") {
			t.Fatalf("error = %v, want whitespace name validation", err)
		}
	})

	t.Run("add rejects whitespace description", func(t *testing.T) {
		h := NewAddAbilityHandler(&stubAbilityStore{
			player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
		})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Dash",
			"description": "   ",
			"type":        "active",
		})
		if err == nil || !strings.Contains(err.Error(), "description cannot be empty or whitespace") {
			t.Fatalf("error = %v, want whitespace description validation", err)
		}
	})

	t.Run("add rejects negative cooldown", func(t *testing.T) {
		h := NewAddAbilityHandler(&stubAbilityStore{
			player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
		})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Dash",
			"description": "Move quickly",
			"type":        "active",
			"cooldown":    -1,
		})
		if err == nil || !strings.Contains(err.Error(), "cooldown must be greater than or equal to 0") {
			t.Fatalf("error = %v, want cooldown validation error", err)
		}
	})

	t.Run("remove wrapped update error", func(t *testing.T) {
		h := NewRemoveAbilityHandler(&stubAbilityStore{
			player: &domain.PlayerCharacter{
				ID:        playerID,
				Abilities: []byte(`[{"name":"Dash","description":"Move quickly","type":"active"}]`),
			},
			updateErr: errors.New("write fail"),
		})
		_, err := h.Handle(ctx, map[string]any{"ability_name": "Dash"})
		if err == nil || !strings.Contains(err.Error(), "update player abilities: write fail") {
			t.Fatalf("error = %v, want wrapped update error", err)
		}
	})

	t.Run("remove rejects whitespace ability_name", func(t *testing.T) {
		h := NewRemoveAbilityHandler(&stubAbilityStore{
			player: &domain.PlayerCharacter{
				ID:        playerID,
				Abilities: []byte(`[{"name":"Dash","description":"Move quickly","type":"active"}]`),
			},
		})
		_, err := h.Handle(ctx, map[string]any{"ability_name": "   "})
		if err == nil || !strings.Contains(err.Error(), "ability_name cannot be empty or whitespace") {
			t.Fatalf("error = %v, want whitespace ability_name validation", err)
		}
	})
}

func TestAddAbilityHandleNilPlayer(t *testing.T) {
	playerID := uuid.New()
	h := NewAddAbilityHandler(&stubAbilityStore{player: nil})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err := h.Handle(ctx, map[string]any{
		"name":        "Dash",
		"description": "Move quickly",
		"type":        "active",
	})
	if err == nil || !strings.Contains(err.Error(), "current player character does not exist") {
		t.Fatalf("error = %v, want nil player error", err)
	}
}

func TestAddAbilityHandleGetPlayerError(t *testing.T) {
	playerID := uuid.New()
	h := NewAddAbilityHandler(&stubAbilityStore{getErr: errors.New("db failure")})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err := h.Handle(ctx, map[string]any{
		"name":        "Dash",
		"description": "Move quickly",
		"type":        "active",
	})
	if err == nil || !strings.Contains(err.Error(), "get player character") {
		t.Fatalf("error = %v, want get player character error", err)
	}
}

func TestAddAbilityHandleUpdateError(t *testing.T) {
	playerID := uuid.New()
	h := NewAddAbilityHandler(&stubAbilityStore{
		player:    &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
		updateErr: errors.New("write fail"),
	})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err := h.Handle(ctx, map[string]any{
		"name":        "Dash",
		"description": "Move quickly",
		"type":        "active",
	})
	if err == nil || !strings.Contains(err.Error(), "update player abilities") {
		t.Fatalf("error = %v, want update player abilities error", err)
	}
}

func TestAddAbilityHandleZeroCooldown(t *testing.T) {
	playerID := uuid.New()
	h := NewAddAbilityHandler(&stubAbilityStore{
		player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
	})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	got, err := h.Handle(ctx, map[string]any{
		"name":        "Dash",
		"description": "Move quickly",
		"type":        "active",
		"cooldown":    0,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	cd, ok := got.Data["cooldown"].(*int)
	if !ok || cd == nil || *cd != 0 {
		t.Fatalf("cooldown = %v, want pointer to 0", got.Data["cooldown"])
	}
}

func TestAddAbilityHandleNoCooldown(t *testing.T) {
	playerID := uuid.New()
	h := NewAddAbilityHandler(&stubAbilityStore{
		player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
	})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	got, err := h.Handle(ctx, map[string]any{
		"name":        "Dash",
		"description": "Move quickly",
		"type":        "active",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["cooldown"] != (*int)(nil) {
		t.Fatalf("cooldown = %v, want nil", got.Data["cooldown"])
	}
}

func TestAddAbilityHandlePassiveType(t *testing.T) {
	playerID := uuid.New()
	h := NewAddAbilityHandler(&stubAbilityStore{
		player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
	})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	got, err := h.Handle(ctx, map[string]any{
		"name":        "Resilience",
		"description": "Passive toughness",
		"type":        "passive",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["type"] != "passive" {
		t.Fatalf("type = %v, want passive", got.Data["type"])
	}
}

func TestRemoveAbilityHandleNilPlayer(t *testing.T) {
	playerID := uuid.New()
	h := NewRemoveAbilityHandler(&stubAbilityStore{player: nil})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err := h.Handle(ctx, map[string]any{"ability_name": "Dash"})
	if err == nil || !strings.Contains(err.Error(), "current player character does not exist") {
		t.Fatalf("error = %v, want nil player error", err)
	}
}

func TestRemoveAbilityHandleGetPlayerError(t *testing.T) {
	playerID := uuid.New()
	h := NewRemoveAbilityHandler(&stubAbilityStore{getErr: errors.New("db failure")})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err := h.Handle(ctx, map[string]any{"ability_name": "Dash"})
	if err == nil || !strings.Contains(err.Error(), "get player character") {
		t.Fatalf("error = %v, want get player character error", err)
	}
}

func TestAddAbilityHandleMissingArgs(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	h := NewAddAbilityHandler(&stubAbilityStore{
		player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
	})

	cases := []struct {
		name    string
		args    map[string]any
		wantErr string
	}{
		{
			name: "missing name",
			args: map[string]any{
				"description": "Move quickly",
				"type":        "active",
			},
			wantErr: "name is required",
		},
		{
			name: "missing description",
			args: map[string]any{
				"name": "Dash",
				"type": "active",
			},
			wantErr: "description is required",
		},
		{
			name: "missing type",
			args: map[string]any{
				"name":        "Dash",
				"description": "Move quickly",
			},
			wantErr: "type is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.Handle(ctx, tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestRemoveAbilityHandleMissingAbilityName(t *testing.T) {
	playerID := uuid.New()
	h := NewRemoveAbilityHandler(&stubAbilityStore{
		player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`[]`)},
	})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err := h.Handle(ctx, map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "ability_name is required") {
		t.Fatalf("error = %v, want ability_name is required", err)
	}
}

func TestAddAbilityHandleCorruptedAbilitiesJSON(t *testing.T) {
	playerID := uuid.New()
	h := NewAddAbilityHandler(&stubAbilityStore{
		player: &domain.PlayerCharacter{ID: playerID, Abilities: []byte(`not json`)},
	})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	_, err := h.Handle(ctx, map[string]any{
		"name":        "Dash",
		"description": "Move quickly",
		"type":        "active",
	})
	if err == nil || !strings.Contains(err.Error(), "unmarshal player abilities") {
		t.Fatalf("error = %v, want unmarshal player abilities error", err)
	}
}

var _ AddAbilityStore = (*stubAbilityStore)(nil)
var _ RemoveAbilityStore = (*stubAbilityStore)(nil)
