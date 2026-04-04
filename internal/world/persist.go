package world

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

func characterLogger() *slog.Logger {
	return slog.Default().WithGroup("world")
}

// PersistCharacterProfile creates a player character record from the given
// CharacterProfile. It maps the profile's descriptive fields into the
// database schema's stat-oriented columns with sensible defaults for a
// new level-1 character.
func PersistCharacterProfile(
	ctx context.Context,
	queries statedb.Querier,
	campaignID, userID uuid.UUID,
	profile *CharacterProfile,
	startingLocationID *uuid.UUID,
) (statedb.PlayerCharacter, error) {
	if profile == nil {
		err := fmt.Errorf("persist character profile: nil profile")
		characterLogger().Error("character persistence failed", "campaign_id", campaignID, "user_id", userID, "error", err)
		return statedb.PlayerCharacter{}, err
	}
	characterLogger().Info("persisting character profile", "campaign_id", campaignID, "user_id", userID, "character", profile.Name)

	stats, err := json.Marshal(map[string]any{
		"concept":     profile.Concept,
		"personality": profile.Personality,
		"motivations": profile.Motivations,
		"strengths":   profile.Strengths,
		"weaknesses":  profile.Weaknesses,
	})
	if err != nil {
		characterLogger().Error("character stats marshal failed", "campaign_id", campaignID, "character", profile.Name, "error", err)
		return statedb.PlayerCharacter{}, fmt.Errorf("persist character profile: marshal stats: %w", err)
	}

	abilities, err := json.Marshal(profile.Strengths)
	if err != nil {
		characterLogger().Error("character abilities marshal failed", "campaign_id", campaignID, "character", profile.Name, "error", err)
		return statedb.PlayerCharacter{}, fmt.Errorf("persist character profile: marshal abilities: %w", err)
	}

	var locationID pgtype.UUID
	if startingLocationID != nil {
		locationID = dbutil.ToPgtype(*startingLocationID)
	}

	description := fmt.Sprintf("%s. %s", profile.Concept, profile.Background)
	pc, err := queries.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{
		CampaignID:        dbutil.ToPgtype(campaignID),
		UserID:            dbutil.ToPgtype(userID),
		Name:              profile.Name,
		Description:       pgtype.Text{String: description, Valid: true},
		Stats:             stats,
		Hp:                10,
		MaxHp:             10,
		Experience:        0,
		Level:             1,
		Status:            "active",
		Abilities:         abilities,
		CurrentLocationID: locationID,
	})
	if err != nil {
		characterLogger().Error("character persistence failed", "campaign_id", campaignID, "character", profile.Name, "error", err)
		return statedb.PlayerCharacter{}, err
	}
	characterLogger().Info("character persisted", "campaign_id", campaignID, "character_id", dbCharacterID(pc), "character", pc.Name)
	return pc, nil
}
