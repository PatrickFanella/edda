package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	pgvector "github.com/pgvector/pgvector-go"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/tools"
)

// worldService consolidates world-building persistence for the create_language
// tool and memory storage.
type worldService struct {
	queries statedb.Querier
}

// NewWorldService creates a service that satisfies both tools.LanguageStore
// and tools.MemoryStore.
func NewWorldService(q statedb.Querier) *worldService {
	return &worldService{queries: q}
}

// --- tools.LanguageStore methods ---

func (s *worldService) CreateLanguage(ctx context.Context, params tools.CreateLanguageParams) (uuid.UUID, error) {
	lang, err := s.queries.CreateLanguage(ctx, statedb.CreateLanguageParams{
		CampaignID:         dbutil.ToPgtype(params.CampaignID),
		Name:               params.Name,
		Description:        pgtype.Text{String: params.Description, Valid: true},
		Phonology:          params.PhonologicalRules,
		Naming:             params.NamingConventions,
		Vocabulary:         params.SampleVocabulary,
		SpokenByFactionIds: dbutil.UUIDsToPgtype(params.SpokenByFactionIDs),
		SpokenByCultureIds: dbutil.UUIDsToPgtype(params.SpokenByCultureIDs),
	})
	if err != nil {
		return uuid.Nil, err
	}
	return dbutil.FromPgtype(lang.ID), nil
}

func (s *worldService) FactionBelongsToCampaign(ctx context.Context, factionID, campaignID uuid.UUID) (bool, error) {
	faction, err := s.queries.GetFactionByID(ctx, dbutil.ToPgtype(factionID))
	if err != nil {
		return false, fmt.Errorf("get faction: %w", err)
	}
	return faction.CampaignID == dbutil.ToPgtype(campaignID), nil
}

func (s *worldService) CultureBelongsToCampaign(ctx context.Context, cultureID, campaignID uuid.UUID) (bool, error) {
	culture, err := s.queries.GetCultureByID(ctx, dbutil.ToPgtype(cultureID))
	if err != nil {
		return false, fmt.Errorf("get culture: %w", err)
	}
	return culture.CampaignID == dbutil.ToPgtype(campaignID), nil
}

// --- tools.MemoryStore methods ---

func (s *worldService) CreateMemory(ctx context.Context, params tools.CreateMemoryParams) error {
	_, err := s.queries.CreateMemory(ctx, statedb.CreateMemoryParams{
		CampaignID: dbutil.ToPgtype(params.CampaignID),
		Content:    params.Content,
		Embedding:  pgvector.NewVector(params.Embedding),
		MemoryType: params.MemoryType,
		Metadata:   params.Metadata,
	})
	return err
}
