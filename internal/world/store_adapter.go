package world

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// worldGenerationSource is the provenance marker stored with facts created
// during skeleton generation.
const worldGenerationSource = "world_generation"

// SkeletonStoreAdapter wraps a statedb.Querier and satisfies both the
// SkeletonStore and SceneStore interfaces, translating between domain types
// and sqlc parameter types.
type SkeletonStoreAdapter struct {
	q statedb.Querier
}

// Compile-time interface checks.
var (
	_ SkeletonStore = (*SkeletonStoreAdapter)(nil)
	_ SceneStore    = (*SkeletonStoreAdapter)(nil)
)

// NewSkeletonStoreAdapter returns an adapter backed by the given querier.
func NewSkeletonStoreAdapter(q statedb.Querier) *SkeletonStoreAdapter {
	return &SkeletonStoreAdapter{q: q}
}

// optionalText converts a string to pgtype.Text, marking it invalid when empty.
func optionalText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// optionalUUID converts a nullable uuid pointer to pgtype.UUID.
// A nil pointer yields a zero-value (invalid) pgtype.UUID.
func optionalUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return dbutil.ToPgtype(*id)
}

// CreateFaction persists a skeleton faction and returns its generated ID.
func (a *SkeletonStoreAdapter) CreateFaction(ctx context.Context, campaignID uuid.UUID, f SkeletonFaction) (uuid.UUID, error) {
	row, err := a.q.CreateFaction(ctx, statedb.CreateFactionParams{
		CampaignID:  dbutil.ToPgtype(campaignID),
		Name:        f.Name,
		Description: optionalText(f.Description),
		Agenda:      optionalText(f.Agenda),
		Territory:   optionalText(f.Territory),
	})
	if err != nil {
		logger().Error("create faction failed", "campaign_id", campaignID, "faction", f.Name, "error", err)
		return uuid.Nil, err
	}
	logger().Debug("created faction", "campaign_id", campaignID, "faction", f.Name, "faction_id", dbutil.FromPgtype(row.ID))
	return dbutil.FromPgtype(row.ID), nil
}

// CreateLocation persists a skeleton location and returns its generated ID.
func (a *SkeletonStoreAdapter) CreateLocation(ctx context.Context, campaignID uuid.UUID, l SkeletonLocation) (uuid.UUID, error) {
	row, err := a.q.CreateLocation(ctx, statedb.CreateLocationParams{
		CampaignID:   dbutil.ToPgtype(campaignID),
		Name:         l.Name,
		Description:  optionalText(l.Description),
		Region:       optionalText(l.Region),
		LocationType: optionalText(l.LocationType),
	})
	if err != nil {
		logger().Error("create location failed", "campaign_id", campaignID, "location", l.Name, "error", err)
		return uuid.Nil, err
	}
	logger().Debug("created location", "campaign_id", campaignID, "location", l.Name, "location_id", dbutil.FromPgtype(row.ID))
	return dbutil.FromPgtype(row.ID), nil
}

// CreateNPC persists a skeleton NPC and returns its generated ID.
// Nil faction or location IDs are stored as invalid pgtype.UUIDs.
// Alive defaults to true and disposition to 50 (neutral).
func (a *SkeletonStoreAdapter) CreateNPC(ctx context.Context, campaignID uuid.UUID, n SkeletonNPC, factionID, locationID *uuid.UUID) (uuid.UUID, error) {
	row, err := a.q.CreateNPC(ctx, statedb.CreateNPCParams{
		CampaignID:  dbutil.ToPgtype(campaignID),
		Name:        n.Name,
		Description: optionalText(n.Description),
		Personality: optionalText(n.Personality),
		Disposition: 50,
		FactionID:   optionalUUID(factionID),
		LocationID:  optionalUUID(locationID),
		Alive:       true,
	})
	if err != nil {
		logger().Error("create npc failed", "campaign_id", campaignID, "npc", n.Name, "error", err)
		return uuid.Nil, err
	}
	logger().Debug("created npc", "campaign_id", campaignID, "npc", n.Name, "npc_id", dbutil.FromPgtype(row.ID))
	return dbutil.FromPgtype(row.ID), nil
}

// CreateWorldFact persists a skeleton world fact and returns its generated ID.
func (a *SkeletonStoreAdapter) CreateWorldFact(ctx context.Context, campaignID uuid.UUID, f SkeletonFact) (uuid.UUID, error) {
	row, err := a.q.CreateFact(ctx, statedb.CreateFactParams{
		CampaignID: dbutil.ToPgtype(campaignID),
		Fact:       f.Fact,
		Category:   f.Category,
		Source:     worldGenerationSource,
	})
	if err != nil {
		logger().Error("create world fact failed", "campaign_id", campaignID, "category", f.Category, "error", err)
		return uuid.Nil, err
	}
	logger().Debug("created world fact", "campaign_id", campaignID, "category", f.Category, "fact_id", dbutil.FromPgtype(row.ID))
	return dbutil.FromPgtype(row.ID), nil
}

// SaveSessionLog persists a domain session log through the querier.
func (a *SkeletonStoreAdapter) SaveSessionLog(ctx context.Context, log domain.SessionLog) error {
	_, err := a.q.CreateSessionLog(ctx, statedb.CreateSessionLogParams{
		CampaignID:   dbutil.ToPgtype(log.CampaignID),
		TurnNumber:   int32(log.TurnNumber),
		PlayerInput:  log.PlayerInput,
		InputType:    string(log.InputType),
		LlmResponse:  log.LLMResponse,
		ToolCalls:    log.ToolCalls,
		LocationID:   optionalUUID(log.LocationID),
		NpcsInvolved: dbutil.UUIDsToPgtype(log.NPCsInvolved),
	})
	if err != nil {
		logger().Error("save session log failed", "campaign_id", log.CampaignID, "turn", log.TurnNumber, "error", err)
		return err
	}
	logger().Debug("saved session log", "campaign_id", log.CampaignID, "turn", log.TurnNumber)
	return nil
}
