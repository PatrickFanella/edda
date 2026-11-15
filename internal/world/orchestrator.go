package world

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/rules"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
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
	SpawnPackage     *CharacterSpawnPackage
	RulesMode        string
	Pool             db.DBTX
	UserID           uuid.UUID
}

// Orchestrator chains the full new-campaign pipeline:
// 1. Draft world skeleton and opening scene in memory (no database transaction).
// 2. Persist campaign, world, character, spawn package and scene in one transaction.
// Failed or interrupted creation leaves no partial campaign to resume.
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

	transactionStore, ok := o.queries.(interface {
		RunInTransaction(context.Context, func(statedb.Querier, statedb.DBTX) error) error
	})
	if !ok {
		return nil, fmt.Errorf("orchestrator: transactional creation unavailable")
	}
	report("Forging the world…")
	skeleton, err := NewSkeletonGenerator(o.llm, nil).Draft(ctx, uuid.Nil, input.Profile)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: generate skeleton: %w", err)
	}
	report("Setting the scene…")
	scene, err := NewSceneGenerator(o.llm, nil).Draft(ctx, uuid.Nil, input.Profile, skeleton)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: generate scene: %w", err)
	}
	var result *OrchestratorResult
	err = transactionStore.RunInTransaction(ctx, func(queries statedb.Querier, tx statedb.DBTX) error {
		report("Creating campaign…")
		campaign, err := queries.CreateCampaign(ctx, statedb.CreateCampaignParams{
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
			return fmt.Errorf("orchestrator: create campaign: %w", err)
		}
		campaignUUID := campaignID(campaign)

		store := NewSkeletonStoreAdapter(queries)
		if err := NewSkeletonGenerator(o.llm, store).Persist(ctx, campaignUUID, skeleton); err != nil {
			return fmt.Errorf("orchestrator: persist skeleton: %w", err)
		}
		startingLocationID, err := resolveStartingLocation(ctx, queries, campaign.ID, skeleton.StartingLocationName)
		if err != nil {
			logger().Error("orchestrator starting location resolution failed", "campaign_id", campaignUUID, "starting_location", skeleton.StartingLocationName, "error", err)
			return fmt.Errorf("orchestrator: resolve starting location: %w", err)
		}
		logger().Debug("orchestrator resolved starting location",
			"campaign_id", campaignUUID,
			"starting_location", skeleton.StartingLocationName,
			"starting_location_id", startingLocationID,
		)

		report("Bringing your character to life…")
		var locPtr *uuid.UUID
		if startingLocationID != uuid.Nil {
			locPtr = &startingLocationID
		}
		characterRow, err := PersistCharacterProfile(ctx, queries, campaignUUID, input.UserID, input.CharacterProfile, locPtr)
		if err != nil {
			logger().Error("orchestrator character persistence failed", "campaign_id", campaignUUID, "error", err)
			return fmt.Errorf("orchestrator: persist character: %w", err)
		}
		logger().Info("orchestrator character persisted", "campaign_id", campaignUUID, "character_id", dbCharacterID(characterRow), "character", characterRow.Name)
		if err := ApplySpawnPackage(ctx, queries, campaignUUID, dbCharacterID(characterRow), input.SpawnPackage); err != nil {
			logger().Error("orchestrator spawn package application failed", "campaign_id", campaignUUID, "character_id", dbCharacterID(characterRow), "error", err)
			return fmt.Errorf("orchestrator: apply spawn package: %w", err)
		}

		if err := NewSceneGenerator(o.llm, store).Persist(ctx, campaignUUID, scene); err != nil {
			return fmt.Errorf("orchestrator: persist scene: %w", err)
		}
		rulesMode, err := orchestrateStartupWorldSetup(ctx, input.RulesMode, tx, campaignUUID)
		if err != nil {
			return fmt.Errorf("orchestrator: initialize starting state: %w", err)
		}
		campaign.RulesMode = rulesMode
		result = &OrchestratorResult{Campaign: campaign, Scene: scene, StartingLocationID: startingLocationID}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("orchestrator: persist world: %w", err)
	}
	logger().Info("orchestrator completed",
		"campaign_id", campaignID(result.Campaign),
		"duration_ms", time.Since(started).Milliseconds(),
		"starting_location_id", result.StartingLocationID,
		"choices", len(scene.Choices),
	)
	return result, nil
}

// orchestrateStartupWorldSetup runs required writes on the creation transaction.
// The returned mode comes from PostgreSQL, so a success response matches the save.
func orchestrateStartupWorldSetup(ctx context.Context, rulesMode string, pool db.DBTX, campaignUUID uuid.UUID) (string, error) {
	if pool == nil {
		return "", fmt.Errorf("starting-state database unavailable")
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO campaign_time (campaign_id, day, hour, minute) VALUES ($1, 1, 8, 0) ON CONFLICT (campaign_id) DO NOTHING", campaignUUID,
	); err != nil {
		return "", fmt.Errorf("initialize campaign clock: %w", err)
	}
	if rulesMode == "" {
		rulesMode = "narrative"
	}
	var savedMode string
	if err := pool.QueryRow(ctx,
		"UPDATE campaigns SET rules_mode = $1 WHERE id = $2 RETURNING rules_mode", rulesMode, campaignUUID,
	).Scan(&savedMode); err != nil {
		return "", fmt.Errorf("save campaign rules mode: %w", err)
	}
	if savedMode == "crunch" {
		if err := rules.SeedDefaultFeats(ctx, pool, campaignUUID); err != nil {
			return "", fmt.Errorf("initialize starting feats: %w", err)
		}
		if err := rules.SeedDefaultSkills(ctx, pool, campaignUUID); err != nil {
			return "", fmt.Errorf("initialize starting skills: %w", err)
		}
	}
	return savedMode, nil
}

// campaignID extracts a uuid.UUID from a statedb.Campaign's pgtype.UUID ID.
func campaignID(c statedb.Campaign) uuid.UUID {
	if !c.ID.Valid {
		return uuid.Nil
	}
	return c.ID.Bytes
}

// resolveStartingLocation queries locations for the campaign and returns the
// ID of the exact matching location name.
func resolveStartingLocation(ctx context.Context, q statedb.Querier, campaignPgID pgtype.UUID, name string) (uuid.UUID, error) {
	if name == "" {
		return uuid.Nil, fmt.Errorf("starting location is required")
	}
	locations, err := q.ListLocationsByCampaign(ctx, campaignPgID)
	if err != nil {
		return uuid.Nil, err
	}
	var match uuid.UUID
	count := 0
	for _, loc := range locations {
		if domain.SameCanonicalLocationName(loc.Name, name) {
			if loc.ID.Valid {
				match = loc.ID.Bytes
				count++
			}
		}
	}
	if count != 1 {
		return uuid.Nil, fmt.Errorf("expected exactly one location named %q, found %d", name, count)
	}
	return match, nil
}

func dbCharacterID(pc statedb.PlayerCharacter) uuid.UUID {
	if !pc.ID.Valid {
		return uuid.Nil
	}
	return pc.ID.Bytes
}
