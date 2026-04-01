package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// getPlayerCharacterByID looks up a player character by ID, returning nil
// when not found.
func getPlayerCharacterByID(ctx context.Context, q statedb.Querier, id uuid.UUID) (*domain.PlayerCharacter, error) {
	pc, err := q.GetPlayerCharacterByID(ctx, dbutil.ToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	domainPC := playerCharacterToDomain(pc)
	return &domainPC, nil
}

// listNPCsByCampaign returns all NPCs in a campaign as domain objects.
func listNPCsByCampaign(ctx context.Context, q statedb.Querier, campaignID uuid.UUID) ([]domain.NPC, error) {
	npcs, err := q.ListNPCsByCampaign(ctx, dbutil.ToPgtype(campaignID))
	if err != nil {
		return nil, fmt.Errorf("list npcs by campaign: %w", err)
	}
	out := make([]domain.NPC, 0, len(npcs))
	for _, npc := range npcs {
		out = append(out, npcToDomain(npc))
	}
	return out, nil
}
