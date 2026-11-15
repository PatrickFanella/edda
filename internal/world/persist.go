package world

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
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

// PersistSupportedCharacter saves the supported SRD creation fixture
// with its derived vitals and catalogue identifier for action handling.
func PersistSupportedCharacter(ctx context.Context, queries statedb.Querier, campaignID, userID uuid.UUID, profile *CharacterProfile, build *srd521.SupportedBuild, startingLocationID *uuid.UUID) (statedb.PlayerCharacter, error) {
	if profile == nil {
		return statedb.PlayerCharacter{}, fmt.Errorf("persist supported SRD character: nil profile")
	}
	stats, err := build.Stats()
	if err != nil {
		return statedb.PlayerCharacter{}, fmt.Errorf("persist supported SRD character: %w", err)
	}
	// Keep the player's descriptive choices alongside the versioned mechanics.
	var envelope map[string]any
	if err := json.Unmarshal(stats, &envelope); err != nil {
		return statedb.PlayerCharacter{}, fmt.Errorf("parse SRD stats: %w", err)
	}
	envelope["concept"] = profile.Concept
	envelope["personality"] = profile.Personality
	envelope["motivations"] = profile.Motivations
	envelope["strengths"] = profile.Strengths
	envelope["weaknesses"] = profile.Weaknesses
	stats, err = json.Marshal(envelope)
	if err != nil {
		return statedb.PlayerCharacter{}, fmt.Errorf("encode SRD profile: %w", err)
	}
	var locationID pgtype.UUID
	if startingLocationID != nil {
		locationID = dbutil.ToPgtype(*startingLocationID)
	}
	features := []string{"Second Wind", "Weapon Mastery: Longsword, Greatsword, Javelin", "Fighting Style: Defense", "Origin Feat: " + build.OriginFeat, "Species Feat: " + build.SpeciesFeat, "Tool Proficiency: " + build.Tool}
	if build.Class == "Cleric" {
		features = []string{"Spellcasting: Wisdom", "Divine Order: Protector", "Origin Feat: " + build.OriginFeat, "Species Feat: " + build.SpeciesFeat, "Tool Proficiency: " + build.Tool}
	}
	abilities, err := json.Marshal(features)
	if err != nil {
		return statedb.PlayerCharacter{}, fmt.Errorf("marshal supported SRD abilities: %w", err)
	}
	hp := int32(12)
	if build.Class == "Cleric" {
		hp = 10
	}
	pc, err := queries.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: dbutil.ToPgtype(campaignID), UserID: dbutil.ToPgtype(userID), Name: profile.Name, Description: pgtype.Text{String: fmt.Sprintf("%s. %s", profile.Concept, profile.Background), Valid: true}, Stats: stats, Hp: hp, MaxHp: hp, Level: 1, Status: "active", Abilities: abilities, CurrentLocationID: locationID})
	if err != nil {
		return statedb.PlayerCharacter{}, fmt.Errorf("create supported SRD character: %w", err)
	}
	type equipmentEntry struct {
		name, itemType, properties string
		equipped                   bool
		quantity                   int32
	}
	equipment := []equipmentEntry{
		{"Chain Mail", "armor", `{"srd521_id":"chain-mail"}`, true, 1},
		{"Shield", "armor", `{"srd521_id":"shield"}`, true, 1},
		{"Longsword", "weapon", `{"srd521_id":"longsword"}`, true, 1},
		{"Greatsword", "weapon", `{"srd521_id":"greatsword"}`, false, 1},
		{"Flail", "weapon", `{"srd521_id":"flail"}`, false, 1},
		{"Javelin", "weapon", `{"srd521_id":"javelin"}`, false, 8},
		{"Dungeoneer's Pack", "misc", `{"srd521_id":"dungeoneers-pack"}`, false, 1},
		{"Coin Pouch (29 GP)", "misc", `{"srd521_gold":29}`, false, 1},
	}
	if build.Class == "Cleric" {
		equipment = equipment[:0]
		for _, name := range build.Equipment {
			quantity := int32(1)
			if name == "Parchment (10)" {
				quantity = 10
			}
			equipment = append(equipment, equipmentEntry{name, "misc", `{}`, false, quantity})
		}
		equipment = append(equipment, equipmentEntry{"Coin Pouch (118 GP)", "misc", `{"srd521_gold":118}`, false, 1})
	}
	for _, item := range equipment {
		if _, err = queries.CreateItem(ctx, statedb.CreateItemParams{CampaignID: dbutil.ToPgtype(campaignID), PlayerCharacterID: pc.ID, Name: item.name, ItemType: item.itemType, Rarity: "common", Equipped: item.equipped, Quantity: item.quantity, Properties: []byte(item.properties)}); err != nil {
			return statedb.PlayerCharacter{}, fmt.Errorf("create supported SRD equipment %s: %w", item.name, err)
		}
	}
	return pc, nil
}
