package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/domain"
)

type stubAddExperienceStore struct {
	player             *domain.PlayerCharacter
	getErr             error
	updateErr          error
	lastPlayerID       uuid.UUID
	lastExperience     int
	lastLevel          int
	updateCallCount    int
}

func (s *stubAddExperienceStore) GetPlayerCharacterByID(_ context.Context, _ uuid.UUID) (*domain.PlayerCharacter, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.player, nil
}

func (s *stubAddExperienceStore) UpdatePlayerExperience(_ context.Context, playerCharacterID uuid.UUID, experience, level int) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.lastPlayerID = playerCharacterID
	s.lastExperience = experience
	s.lastLevel = level
	s.updateCallCount++
	return nil
}

func TestRegisterAddExperience(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterAddExperience(reg, &stubAddExperienceStore{}); err != nil {
		t.Fatalf("register add_experience: %v", err)
	}

	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != addExperienceToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, addExperienceToolName)
	}
}

func TestAddExperienceHandleAccumulatesExperienceAndFlagsLevelUp(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddExperienceStore{
		player: &domain.PlayerCharacter{
			ID:         playerID,
			Experience: 980,
			Level:      1,
		},
	}
	h := NewAddExperienceHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"amount": 50,
		"reason": "defeating the bandit",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if store.lastExperience != 1030 {
		t.Fatalf("updated experience = %d, want 1030", store.lastExperience)
	}
	if store.lastLevel != 1 {
		t.Fatalf("updated level = %d, want 1", store.lastLevel)
	}
	if got.Data["level_up_available"] != true {
		t.Fatalf("level_up_available = %v, want true", got.Data["level_up_available"])
	}
	if got.Narrative != "You gained 50 XP for defeating the bandit." {
		t.Fatalf("narrative = %q", got.Narrative)
	}
}

func TestAddExperienceHandleUsesConfigurableThreshold(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddExperienceStore{
		player: &domain.PlayerCharacter{
			ID:         playerID,
			Experience: 95,
			Level:      1,
		},
	}
	h := NewAddExperienceHandlerWithThreshold(store, func(_ int) int { return 100 })
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"amount": 5,
		"reason": "a quest milestone",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["level_up_available"] != true {
		t.Fatalf("level_up_available = %v, want true", got.Data["level_up_available"])
	}
}

func TestAddExperienceHandleUsesCumulativeDefaultThreshold(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddExperienceStore{
		player: &domain.PlayerCharacter{
			ID:         playerID,
			Experience: 2999,
			Level:      2,
		},
	}
	h := NewAddExperienceHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"amount": 1,
		"reason": "finishing a battle",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["level_up_available"] != true {
		t.Fatalf("level_up_available = %v, want true", got.Data["level_up_available"])
	}
}

func TestAddExperienceHandleValidationAndStoreErrors(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("requires player context", func(t *testing.T) {
		h := NewAddExperienceHandler(&stubAddExperienceStore{})
		_, err := h.Handle(context.Background(), map[string]any{
			"amount": 1,
			"reason": "test",
		})
		if err == nil || !strings.Contains(err.Error(), "requires current player character id in context") {
			t.Fatalf("error = %v, want missing context", err)
		}
	})

	t.Run("rejects non-positive amount", func(t *testing.T) {
		h := NewAddExperienceHandler(&stubAddExperienceStore{
			player: &domain.PlayerCharacter{ID: playerID, Level: 1},
		})
		_, err := h.Handle(ctx, map[string]any{
			"amount": 0,
			"reason": "test",
		})
		if err == nil || !strings.Contains(err.Error(), "amount must be greater than 0") {
			t.Fatalf("error = %v, want amount validation", err)
		}
	})

	t.Run("get player wrapped error", func(t *testing.T) {
		h := NewAddExperienceHandler(&stubAddExperienceStore{
			getErr: errors.New("db read failed"),
		})
		_, err := h.Handle(ctx, map[string]any{
			"amount": 5,
			"reason": "test",
		})
		if err == nil || !strings.Contains(err.Error(), "get player character") {
			t.Fatalf("error = %v, want get player character wrapper", err)
		}
	})

	t.Run("update experience wrapped error", func(t *testing.T) {
		h := NewAddExperienceHandler(&stubAddExperienceStore{
			player: &domain.PlayerCharacter{ID: playerID, Experience: 10, Level: 1},
			updateErr: errors.New("db write failed"),
		})
		_, err := h.Handle(ctx, map[string]any{
			"amount": 5,
			"reason": "test",
		})
		if err == nil || !strings.Contains(err.Error(), "update player experience") {
			t.Fatalf("error = %v, want update player experience wrapper", err)
		}
	})
}

func TestAddExperienceHandleNegativeAmount(t *testing.T) {
	playerID := uuid.New()
	h := NewAddExperienceHandler(&stubAddExperienceStore{
		player: &domain.PlayerCharacter{ID: playerID, Level: 1},
	})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"amount": -5,
		"reason": "test",
	})
	if err == nil || !strings.Contains(err.Error(), "amount must be greater than 0") {
		t.Fatalf("error = %v, want amount must be greater than 0", err)
	}
}

func TestAddExperienceHandleLargeXPValue(t *testing.T) {
	playerID := uuid.New()
	const largeAmount = 999999999
	store := &stubAddExperienceStore{
		player: &domain.PlayerCharacter{ID: playerID, Experience: 0, Level: 1},
	}
	h := NewAddExperienceHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"amount": largeAmount,
		"reason": "massive battle reward",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["new_experience"] != largeAmount {
		t.Fatalf("new_experience = %v, want %d", got.Data["new_experience"], largeAmount)
	}
	if got.Data["level_up_available"] != true {
		t.Fatalf("level_up_available = %v, want true", got.Data["level_up_available"])
	}
}

func TestAddExperienceHandleNilPlayer(t *testing.T) {
	playerID := uuid.New()
	// store returns (nil, nil) — player not found, no error
	h := NewAddExperienceHandler(&stubAddExperienceStore{player: nil})
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"amount": 10,
		"reason": "test",
	})
	if err == nil || !strings.Contains(err.Error(), "current player character does not exist") {
		t.Fatalf("error = %v, want current player character does not exist", err)
	}
}

func TestAddExperienceHandleLevelZeroNormalized(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddExperienceStore{
		player: &domain.PlayerCharacter{ID: playerID, Experience: 900, Level: 0},
	}
	h := NewAddExperienceHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"amount": 100,
		"reason": "test",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["current_level"] != 1 {
		t.Fatalf("current_level = %v, want 1 (level 0 normalized to 1)", got.Data["current_level"])
	}
}

func TestAddExperienceHandleMissingArgs(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddExperienceStore{
		player: &domain.PlayerCharacter{ID: playerID, Level: 1},
	}
	h := NewAddExperienceHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	cases := []struct {
		name    string
		args    map[string]any
		wantErr string
	}{
		{
			name:    "missing amount",
			args:    map[string]any{"reason": "test"},
			wantErr: "amount is required",
		},
		{
			name:    "missing reason",
			args:    map[string]any{"amount": 5},
			wantErr: "reason is required",
		},
		{
			name:    "non-integer amount",
			args:    map[string]any{"amount": "five", "reason": "test"},
			wantErr: "amount must be an integer",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.Handle(ctx, tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestAddExperienceHandleExperienceToNextField(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddExperienceStore{
		player: &domain.PlayerCharacter{ID: playerID, Experience: 500, Level: 1},
	}
	// fixed threshold of 1000 for level 1 so the assertion is deterministic
	h := NewAddExperienceHandlerWithThreshold(store, func(_ int) int { return 1000 })
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"amount": 200,
		"reason": "dungeon clear",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	// new_experience = 700; threshold = 1000; experience_to_next = 300
	if got.Data["new_experience"] != 700 {
		t.Fatalf("new_experience = %v, want 700", got.Data["new_experience"])
	}
	if got.Data["experience_to_next"] != 300 {
		t.Fatalf("experience_to_next = %v, want 300", got.Data["experience_to_next"])
	}
}

var _ AddExperienceStore = (*stubAddExperienceStore)(nil)
