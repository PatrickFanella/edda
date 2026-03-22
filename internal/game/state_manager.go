package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/PatrickFanella/game-master/internal/domain"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// pgStateManager implements StateManager using pgx and sqlc.
type pgStateManager struct {
	db      statedb.DBTX
	queries statedb.Querier
}

// NewStateManager creates a new StateManager backed by the given database connection.
func NewStateManager(db statedb.DBTX) StateManager {
	return &pgStateManager{
		db:      db,
		queries: statedb.New(db),
	}
}

// newStateManagerWithQuerier is used for testing with a mock Querier.
func newStateManagerWithQuerier(q statedb.Querier) *pgStateManager {
	return &pgStateManager{queries: q}
}

func (sm *pgStateManager) GetOrCreateDefaultUser(ctx context.Context) (*domain.User, error) {
	const defaultName = "Player"

	u, err := sm.queries.GetUserByName(ctx, defaultName)
	if err == nil {
		return userToDomain(u), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get default user: %w", err)
	}

	u, err = sm.queries.CreateUser(ctx, defaultName)
	if err != nil {
		return nil, fmt.Errorf("create default user: %w", err)
	}
	return userToDomain(u), nil
}

func (sm *pgStateManager) CreateCampaign(ctx context.Context, params CreateCampaignParams) (*domain.Campaign, error) {
	return nil, fmt.Errorf("CreateCampaign: not yet implemented (requires campaign queries)")
}

func (sm *pgStateManager) LoadCampaign(ctx context.Context, id uuid.UUID) (*GameState, error) {
	return nil, fmt.Errorf("LoadCampaign: not yet implemented (requires campaign, character, location, npc, quest queries)")
}

func (sm *pgStateManager) GetGameState(ctx context.Context, campaignID uuid.UUID) (*GameState, error) {
	return nil, fmt.Errorf("GetGameState: not yet implemented (requires campaign, character, location, npc, quest queries)")
}

func (sm *pgStateManager) SaveSessionLog(ctx context.Context, log domain.SessionLog) error {
	return fmt.Errorf("SaveSessionLog: not yet implemented (requires session_log queries)")
}
