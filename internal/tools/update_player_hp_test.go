package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
)

type stubUpdatePlayerHPStore struct {
	player      *domain.PlayerCharacter
	getErr      error
	updateErr   error
	lastPlayer  uuid.UUID
	lastHP      int
	lastMaxHP   int
	updateCalls int
}

func (s *stubUpdatePlayerHPStore) GetPlayerCharacterByID(_ context.Context, playerCharacterID uuid.UUID) (*domain.PlayerCharacter, error) {
	s.lastPlayer = playerCharacterID
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.player, nil
}

func (s *stubUpdatePlayerHPStore) UpdatePlayerHP(_ context.Context, playerCharacterID uuid.UUID, hp, maxHP int) error {
	s.lastPlayer = playerCharacterID
	s.lastHP = hp
	s.lastMaxHP = maxHP
	s.updateCalls++
	if s.updateErr != nil {
		return s.updateErr
	}
	return nil
}

func (s *stubUpdatePlayerHPStore) UpdatePlayerCurrentHP(_ context.Context, playerCharacterID uuid.UUID, hp int) error {
	s.lastPlayer = playerCharacterID
	s.lastHP = hp
	s.updateCalls++
	if s.updateErr != nil {
		return s.updateErr
	}
	return nil
}

func TestRegisterUpdatePlayerHP(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterUpdatePlayerHP(reg, &stubUpdatePlayerHPStore{}); err != nil {
		t.Fatalf("register update_player_hp: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != updatePlayerHPToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, updatePlayerHPToolName)
	}
}

func TestUpdatePlayerHPHandlePreservesCurrentMax(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerHPStore{player: &domain.PlayerCharacter{ID: playerID, HP: 12, MaxHP: 20}}
	h := NewUpdatePlayerHPHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{"hp": 7})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["hp"] != 7 || got.Data["max_hp"] != 20 {
		t.Fatalf("result data = %#v, want hp=7 max_hp=20", got.Data)
	}
	if store.lastHP != 7 || store.lastMaxHP != 0 {
		t.Fatalf("persisted hp/max_hp = %d/%d, want hp-only 7/0", store.lastHP, store.lastMaxHP)
	}
}

func TestUpdatePlayerHPHandleRejectsExplicitMaxHP(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerHPStore{player: &domain.PlayerCharacter{ID: playerID, HP: 12, MaxHP: 20}}
	h := NewUpdatePlayerHPHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{"hp": 10, "max_hp": 18})
	if err == nil {
		t.Fatal("expected max_hp to be rejected by additionalProperties=false contract")
	}
	if store.updateCalls != 0 {
		t.Fatalf("updateCalls = %d, want 0", store.updateCalls)
	}
}

func TestUpdatePlayerHPHandleValidationErrors(t *testing.T) {
	playerID := uuid.New()
	store := &stubUpdatePlayerHPStore{player: &domain.PlayerCharacter{ID: playerID, HP: 12, MaxHP: 20}}
	h := NewUpdatePlayerHPHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	for name, args := range map[string]map[string]any{
		"negative hp":  {"hp": -1},
		"hp above max": {"hp": 21},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := h.Handle(ctx, args)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestUpdatePlayerHPHandleRequiresPlayerContext(t *testing.T) {
	h := NewUpdatePlayerHPHandler(&stubUpdatePlayerHPStore{player: &domain.PlayerCharacter{ID: uuid.New(), HP: 1, MaxHP: 1}})
	_, err := h.Handle(context.Background(), map[string]any{"hp": 1})
	if err == nil || !strings.Contains(err.Error(), "requires current player character id in context") {
		t.Fatalf("err = %v, want missing-context error", err)
	}
}

func TestUpdatePlayerHPHandleStoreErrors(t *testing.T) {
	playerID := uuid.New()
	t.Run("get player error", func(t *testing.T) {
		h := NewUpdatePlayerHPHandler(&stubUpdatePlayerHPStore{getErr: errors.New("db read failed")})
		ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
		_, err := h.Handle(ctx, map[string]any{"hp": 1})
		if err == nil || !strings.Contains(err.Error(), "get player character") {
			t.Fatalf("err = %v, want get player character wrapper", err)
		}
	})

	t.Run("update hp error", func(t *testing.T) {
		h := NewUpdatePlayerHPHandler(&stubUpdatePlayerHPStore{player: &domain.PlayerCharacter{ID: playerID, HP: 3, MaxHP: 5}, updateErr: errors.New("db write failed")})
		ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
		_, err := h.Handle(ctx, map[string]any{"hp": 1})
		if err == nil || !strings.Contains(err.Error(), "update player hp") {
			t.Fatalf("err = %v, want update player hp wrapper", err)
		}
	})
}
