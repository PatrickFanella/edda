package game

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/tools"
)

type progressionService struct {
	queries statedb.Querier
}

func NewProgressionService(q statedb.Querier) *progressionService {
	return &progressionService{queries: q}
}

func (s *progressionService) GetPlayerCharacterByID(ctx context.Context, playerCharacterID uuid.UUID) (*domain.PlayerCharacter, error) {
	return getPlayerCharacterByID(ctx, s.queries, playerCharacterID)
}

func (s *progressionService) UpdatePlayerExperience(ctx context.Context, playerCharacterID uuid.UUID, experience, level int) error {
	_, err := s.queries.UpdatePlayerExperience(ctx, statedb.UpdatePlayerExperienceParams{
		ID:         dbutil.ToPgtype(playerCharacterID),
		Experience: int32(experience),
		Level:      int32(level),
	})
	return err
}

func (s *progressionService) UpdatePlayerLevel(ctx context.Context, playerCharacterID uuid.UUID, level int) error {
	_, err := s.queries.UpdatePlayerLevel(ctx, statedb.UpdatePlayerLevelParams{
		ID:    dbutil.ToPgtype(playerCharacterID),
		Level: int32(level),
	})
	return err
}

func (s *progressionService) UpdatePlayerStats(ctx context.Context, playerCharacterID uuid.UUID, stats json.RawMessage) error {
	_, err := s.queries.UpdatePlayerStats(ctx, statedb.UpdatePlayerStatsParams{
		ID:    dbutil.ToPgtype(playerCharacterID),
		Stats: stats,
	})
	return err
}

func (s *progressionService) UpdatePlayerAbilities(ctx context.Context, playerCharacterID uuid.UUID, abilities json.RawMessage) error {
	_, err := s.queries.UpdatePlayerAbilities(ctx, statedb.UpdatePlayerAbilitiesParams{
		ID:        dbutil.ToPgtype(playerCharacterID),
		Abilities: abilities,
	})
	return err
}

var _ tools.AddExperienceStore = (*progressionService)(nil)
var _ tools.LevelUpStore = (*progressionService)(nil)
