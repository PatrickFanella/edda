package world

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/llm"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// OrchestratorResult holds everything the TUI needs after world creation
// completes: the persisted campaign, the opening scene, and the starting
// location ID so the App can orient the player.
type OrchestratorResult struct {
	Campaign           statedb.Campaign
	Scene              *SceneResult
	StartingLocationID uuid.UUID
}

// OrchestratorInput bundles the player's choices from the creation wizard.
type OrchestratorInput struct {
	Name             string // LLM-generated or picked from proposals
	Summary          string // campaign description/summary
	Profile          *CampaignProfile
	CharacterProfile *CharacterProfile
	UserID           uuid.UUID
}

// Orchestrator chains the full new-campaign pipeline:
// 1. Insert campaign row
// 2. Generate world skeleton → persist factions/locations/NPCs/facts
// 3. Persist player character
// 4. Generate opening scene → persist session log
//
// Progress is reported through the callback so the TUI can display stage messages.
type Orchestrator struct {
	llm     llm.Provider
	queries statedb.Querier
}

// NewOrchestrator creates an Orchestrator wired to the given LLM and DB querier.
func NewOrchestrator(provider llm.Provider, queries statedb.Querier) *Orchestrator {
	return &Orchestrator{llm: provider, queries: queries}
}

// Run executes the full creation pipeline. The progress callback is called
// before each stage with a human-readable description; it may be nil.
func (o *Orchestrator) Run(ctx context.Context, input OrchestratorInput, progress func(stage string)) (*OrchestratorResult, error) {
	started := time.Now()
	if input.Profile == nil {
		err := fmt.Errorf("orchestrator: campaign profile is nil")
		logger().Error("orchestrator failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}
	if input.CharacterProfile == nil {
		err := fmt.Errorf("orchestrator: character profile is nil")
		logger().Error("orchestrator failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	logger().Info("orchestrator started",
		"campaign", input.Name,
		"user_id", input.UserID,
		"themes", len(input.Profile.Themes),
		"character", input.CharacterProfile.Name,
	)
	report := func(s string) {
		logger().Info("orchestrator stage", "stage", s, "campaign", input.Name)
		if progress != nil {
			progress(s)
		}
	}

	report("Creating campaign…")
	campaign, err := o.queries.CreateCampaign(ctx, statedb.CreateCampaignParams{
		Name:                input.Name,
		Description:         pgtype.Text{String: input.Summary, Valid: input.Summary != ""},
		Genre:               pgtype.Text{String: input.Profile.Genre, Valid: input.Profile.Genre != ""},
		Tone:                pgtype.Text{String: input.Profile.Tone, Valid: input.Profile.Tone != ""},
		Themes:              input.Profile.Themes,
		WorldType:           pgtype.Text{String: input.Profile.WorldType, Valid: input.Profile.WorldType != ""},
		DangerLevel:         pgtype.Text{String: input.Profile.DangerLevel, Valid: input.Profile.DangerLevel != ""},
		PoliticalComplexity: pgtype.Text{String: input.Profile.PoliticalComplexity, Valid: input.Profile.PoliticalComplexity != ""},
		Status:              "active",
		CreatedBy:           pgtype.UUID{Bytes: input.UserID, Valid: input.UserID != uuid.Nil},
	})
	if err != nil {
		logger().Error("orchestrator campaign creation failed", "campaign", input.Name, "error", err)
		return nil, fmt.Errorf("orchestrator: create campaign: %w", err)
	}
	campaignUUID := campaignID(campaign)

	report("Forging the world…")
	store := NewSkeletonStoreAdapter(o.queries)
	skeletonGen := NewSkeletonGenerator(o.llm, store)
	skeleton, err := skeletonGen.Generate(ctx, campaignUUID, input.Profile)
	if err != nil {
		logger().Error("orchestrator skeleton generation failed", "campaign_id", campaignUUID, "error", err)
		return nil, fmt.Errorf("orchestrator: generate skeleton: %w", err)
	}

	var startingLocationID uuid.UUID
	if skeleton.StartingLocationName != "" {
		startingLocationID = resolveStartingLocation(ctx, o.queries, campaign.ID, skeleton.StartingLocationName)
		logger().Debug("orchestrator resolved starting location",
			"campaign_id", campaignUUID,
			"starting_location", skeleton.StartingLocationName,
			"starting_location_id", startingLocationID,
		)
	}

	report("Bringing your character to life…")
	var locPtr *uuid.UUID
	if startingLocationID != uuid.Nil {
		locPtr = &startingLocationID
	}
	characterRow, err := PersistCharacterProfile(ctx, o.queries, campaignUUID, input.UserID, input.CharacterProfile, locPtr)
	if err != nil {
		logger().Error("orchestrator character persistence failed", "campaign_id", campaignUUID, "error", err)
		return nil, fmt.Errorf("orchestrator: persist character: %w", err)
	}
	logger().Info("orchestrator character persisted", "campaign_id", campaignUUID, "character_id", dbCharacterID(characterRow), "character", characterRow.Name)

	report("Setting the scene…")
	sceneGen := NewSceneGenerator(o.llm, store)
	scene, err := sceneGen.Generate(ctx, campaignUUID, input.Profile, skeleton)
	if err != nil {
		logger().Error("orchestrator scene generation failed", "campaign_id", campaignUUID, "error", err)
		return nil, fmt.Errorf("orchestrator: generate scene: %w", err)
	}

	result := &OrchestratorResult{
		Campaign:           campaign,
		Scene:              scene,
		StartingLocationID: startingLocationID,
	}
	logger().Info("orchestrator completed",
		"campaign_id", campaignUUID,
		"duration_ms", time.Since(started).Milliseconds(),
		"starting_location_id", startingLocationID,
		"choices", len(scene.Choices),
	)
	return result, nil
}

// campaignID extracts a uuid.UUID from a statedb.Campaign's pgtype.UUID ID.
func campaignID(c statedb.Campaign) uuid.UUID {
	if !c.ID.Valid {
		return uuid.Nil
	}
	return c.ID.Bytes
}

// resolveStartingLocation queries locations for the campaign and returns the
// ID of the location matching name. Returns uuid.Nil if not found.
func resolveStartingLocation(ctx context.Context, q statedb.Querier, campaignPgID pgtype.UUID, name string) uuid.UUID {
	locations, err := q.ListLocationsByCampaign(ctx, campaignPgID)
	if err != nil {
		return uuid.Nil
	}
	for _, loc := range locations {
		if loc.Name == name {
			if loc.ID.Valid {
				return loc.ID.Bytes
			}
		}
	}
	return uuid.Nil
}

func dbCharacterID(pc statedb.PlayerCharacter) uuid.UUID {
	if !pc.ID.Valid {
		return uuid.Nil
	}
	return pc.ID.Bytes
}
