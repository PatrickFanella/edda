package game

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/tools"
)

var _ tools.StatModifierResolver = (*statModifierResolver)(nil)

func TestStatModifierResolverReturnsModifier(t *testing.T) {
	q := newMockQuerier()
	pcID := uuid.New()
	q.playerCharacter = statedb.PlayerCharacter{
		ID:    dbutil.ToPgtype(pcID),
		Stats: []byte(`{"strength":16,"dexterity":12}`),
	}
	resolver := NewStatModifierResolver(q)

	// strength 16 → modifier (16-10)/2 = 3
	mod, err := resolver.GetStatModifier(context.Background(), pcID, "strength")
	if err != nil {
		t.Fatalf("GetStatModifier() error = %v", err)
	}
	if mod != 3 {
		t.Fatalf("modifier = %d, want 3 for strength 16", mod)
	}

	// dexterity 12 → modifier (12-10)/2 = 1
	mod, err = resolver.GetStatModifier(context.Background(), pcID, "dexterity")
	if err != nil {
		t.Fatalf("GetStatModifier() error = %v", err)
	}
	if mod != 1 {
		t.Fatalf("modifier = %d, want 1 for dexterity 12", mod)
	}
}

func TestStatModifierResolverCaseInsensitive(t *testing.T) {
	q := newMockQuerier()
	pcID := uuid.New()
	q.playerCharacter = statedb.PlayerCharacter{
		ID:    dbutil.ToPgtype(pcID),
		Stats: []byte(`{"Strength":14}`),
	}
	resolver := NewStatModifierResolver(q)

	mod, err := resolver.GetStatModifier(context.Background(), pcID, "strength")
	if err != nil {
		t.Fatalf("GetStatModifier() error = %v", err)
	}
	if mod != 2 {
		t.Fatalf("modifier = %d, want 2 for Strength 14", mod)
	}
}

func TestStatModifierResolverMissingStatReturnsZero(t *testing.T) {
	q := newMockQuerier()
	pcID := uuid.New()
	q.playerCharacter = statedb.PlayerCharacter{
		ID:    dbutil.ToPgtype(pcID),
		Stats: []byte(`{"strength":10}`),
	}
	resolver := NewStatModifierResolver(q)

	mod, err := resolver.GetStatModifier(context.Background(), pcID, "wisdom")
	if err != nil {
		t.Fatalf("GetStatModifier() error = %v", err)
	}
	if mod != 0 {
		t.Fatalf("modifier = %d, want 0 for missing stat", mod)
	}
}

func TestStatModifierResolverEmptyStatsReturnsZero(t *testing.T) {
	q := newMockQuerier()
	pcID := uuid.New()
	q.playerCharacter = statedb.PlayerCharacter{
		ID: dbutil.ToPgtype(pcID),
	}
	resolver := NewStatModifierResolver(q)

	mod, err := resolver.GetStatModifier(context.Background(), pcID, "strength")
	if err != nil {
		t.Fatalf("GetStatModifier() error = %v", err)
	}
	if mod != 0 {
		t.Fatalf("modifier = %d, want 0 for empty stats", mod)
	}
}

func TestStatModifierResolverLowScoreNegativeModifier(t *testing.T) {
	q := newMockQuerier()
	pcID := uuid.New()
	q.playerCharacter = statedb.PlayerCharacter{
		ID:    dbutil.ToPgtype(pcID),
		Stats: []byte(`{"strength":8}`),
	}
	resolver := NewStatModifierResolver(q)

	// strength 8 → modifier (8-10)/2 = -1
	mod, err := resolver.GetStatModifier(context.Background(), pcID, "strength")
	if err != nil {
		t.Fatalf("GetStatModifier() error = %v", err)
	}
	if mod != -1 {
		t.Fatalf("modifier = %d, want -1 for strength 8", mod)
	}
}
