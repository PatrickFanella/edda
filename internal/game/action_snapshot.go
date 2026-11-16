package game

import (
	"context"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ReadActionSnapshot loads exactly the inputs fingerprinted by preparation.
// A nil owner is for trusted engine callers; API callers supply the owner.
// Locking is used only inside a short confirmation transaction.
func ReadActionSnapshot(ctx context.Context, database statedb.DBTX, campaignID, actorID, ownerID uuid.UUID, lock bool) (domain.Campaign, domain.PlayerCharacter, error) {
	var owner any
	if ownerID != uuid.Nil {
		owner = dbutil.ToPgtype(ownerID)
	}
	locking := ""
	if lock {
		// Campaign first prevents cross-caster lock inversion between engines.
		var id uuid.UUID
		if err := database.QueryRow(ctx, `SELECT id FROM campaigns WHERE id=$1 AND ($2::uuid IS NULL OR created_by=$2) FOR UPDATE`, campaignID, owner).Scan(&id); err != nil {
			return domain.Campaign{}, domain.PlayerCharacter{}, err
		}
		locking = " FOR UPDATE OF p"
	}
	campaign := domain.Campaign{ID: campaignID}
	pc := domain.PlayerCharacter{ID: actorID, CampaignID: campaignID}
	var location pgtype.UUID
	// Owner filtering is repeated in the same snapshot as the selected actor.
	// No world/NPC/fact table is read and arbitrary actors never disclose existence.
	err := database.QueryRow(ctx, `SELECT c.rules_version,c.status,c.updated_at,
 c.created_by,p.user_id,p.name,p.stats,p.hp,p.max_hp,p.experience,p.level,p.status,p.current_location_id,p.updated_at
 FROM campaigns c JOIN player_characters p ON p.campaign_id=c.id
 WHERE c.id=$1 AND p.id=$2 AND ($3::uuid IS NULL OR (c.created_by=$3 AND p.user_id=$3))`+locking, dbutil.ToPgtype(campaignID), dbutil.ToPgtype(actorID), owner).Scan(
		&campaign.RulesVersion, &campaign.Status, &campaign.UpdatedAt, &campaign.CreatedBy, &pc.UserID, &pc.Name, &pc.Stats, &pc.HP, &pc.MaxHP, &pc.Experience, &pc.Level, &pc.Status, &location, &pc.UpdatedAt)
	if location.Valid {
		id := dbutil.FromPgtype(location)
		pc.CurrentLocationID = &id
	}
	return campaign, pc, err
}
