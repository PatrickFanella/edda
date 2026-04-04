package world

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/llmutil"
	"github.com/PatrickFanella/game-master/internal/prompt"
)

// WorldSkeleton is the initial world state generated from a CampaignProfile.
type WorldSkeleton struct {
	Factions             []SkeletonFaction
	Locations            []SkeletonLocation
	NPCs                 []SkeletonNPC
	WorldFacts           []SkeletonFact
	StartingLocationName string // name of the starting location chosen by the LLM
}

// SkeletonFaction describes a faction to seed into a new campaign world.
type SkeletonFaction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Agenda      string `json:"agenda"`
	Territory   string `json:"territory"`
}

// SkeletonLocation describes a location to seed into a new campaign world.
type SkeletonLocation struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Region       string `json:"region"`
	LocationType string `json:"location_type"`
}

// SkeletonNPC describes an NPC to seed into a new campaign world.
type SkeletonNPC struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Personality string `json:"personality"`
	Faction     string `json:"faction"`  // name reference
	Location    string `json:"location"` // name reference
}

// SkeletonFact describes a world fact to seed into a new campaign world.
type SkeletonFact struct {
	Fact     string `json:"fact"`
	Category string `json:"category"`
}

// skeletonLLMResponse is the expected JSON shape returned by the LLM.
type skeletonLLMResponse struct {
	Factions         []SkeletonFaction  `json:"factions"`
	Locations        []SkeletonLocation `json:"locations"`
	NPCs             []SkeletonNPC      `json:"npcs"`
	WorldFacts       []SkeletonFact     `json:"world_facts"`
	StartingLocation string             `json:"starting_location"`
}

// SkeletonStore persists skeleton entities during world generation.
type SkeletonStore interface {
	CreateFaction(ctx context.Context, campaignID uuid.UUID, f SkeletonFaction) (uuid.UUID, error)
	CreateLocation(ctx context.Context, campaignID uuid.UUID, l SkeletonLocation) (uuid.UUID, error)
	CreateNPC(ctx context.Context, campaignID uuid.UUID, n SkeletonNPC, factionID, locationID *uuid.UUID) (uuid.UUID, error)
	CreateWorldFact(ctx context.Context, campaignID uuid.UUID, f SkeletonFact) (uuid.UUID, error)
}

// SkeletonGenerator produces an initial world skeleton from a campaign profile
// by prompting an LLM and persisting the results through a SkeletonStore.
type SkeletonGenerator struct {
	llm   llm.Provider
	store SkeletonStore
}

// NewSkeletonGenerator returns a generator wired to the given LLM and store.
func NewSkeletonGenerator(provider llm.Provider, store SkeletonStore) *SkeletonGenerator {
	return &SkeletonGenerator{llm: provider, store: store}
}

// Generate asks the LLM to produce a world skeleton for the given campaign
// profile, persists all entities through the store, and returns the populated
// skeleton with resolved IDs.
func (g *SkeletonGenerator) Generate(ctx context.Context, campaignID uuid.UUID, profile *CampaignProfile) (*WorldSkeleton, error) {
	started := time.Now()
	if profile == nil {
		err := fmt.Errorf("generate skeleton: profile is nil")
		logger().Error("skeleton generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	logger().Info("skeleton generation started",
		"campaign_id", campaignID,
		"genre", profile.Genre,
		"tone", profile.Tone,
		"themes", len(profile.Themes),
	)
	promptText := buildSkeletonPrompt(profile)
	logger().Debug("skeleton prompt built", "campaign_id", campaignID, "prompt_len", len(promptText))

	resp, err := g.llm.Complete(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: promptText},
	}, nil)
	if err != nil {
		logger().Error("skeleton generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, fmt.Errorf("generate skeleton: llm call: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		err := fmt.Errorf("generate skeleton: empty LLM response")
		logger().Error("skeleton generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	content = llmutil.StripMarkdownFences(content)

	var parsed skeletonLLMResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		logger().Error("skeleton generation parse failed",
			"campaign_id", campaignID,
			"duration_ms", time.Since(started).Milliseconds(),
			"response_len", len(content),
			"error", err,
		)
		return nil, fmt.Errorf("generate skeleton: parse response: %w", err)
	}

	logger().Info("skeleton response parsed",
		"campaign_id", campaignID,
		"factions", len(parsed.Factions),
		"locations", len(parsed.Locations),
		"npcs", len(parsed.NPCs),
		"world_facts", len(parsed.WorldFacts),
		"starting_location", parsed.StartingLocation,
	)

	factionIDs := make(map[string]uuid.UUID, len(parsed.Factions))
	for _, f := range parsed.Factions {
		id, err := g.store.CreateFaction(ctx, campaignID, f)
		if err != nil {
			logger().Error("skeleton faction persistence failed", "campaign_id", campaignID, "faction", f.Name, "error", err)
			return nil, fmt.Errorf("generate skeleton: create faction %q: %w", f.Name, err)
		}
		factionIDs[f.Name] = id
	}

	locationIDs := make(map[string]uuid.UUID, len(parsed.Locations))
	for _, l := range parsed.Locations {
		id, err := g.store.CreateLocation(ctx, campaignID, l)
		if err != nil {
			logger().Error("skeleton location persistence failed", "campaign_id", campaignID, "location", l.Name, "error", err)
			return nil, fmt.Errorf("generate skeleton: create location %q: %w", l.Name, err)
		}
		locationIDs[l.Name] = id
	}

	for _, n := range parsed.NPCs {
		var factionID, locationID *uuid.UUID
		if fid, ok := factionIDs[n.Faction]; ok {
			factionID = &fid
		}
		if lid, ok := locationIDs[n.Location]; ok {
			locationID = &lid
		}
		if _, err := g.store.CreateNPC(ctx, campaignID, n, factionID, locationID); err != nil {
			logger().Error("skeleton npc persistence failed", "campaign_id", campaignID, "npc", n.Name, "error", err)
			return nil, fmt.Errorf("generate skeleton: create npc %q: %w", n.Name, err)
		}
	}

	for _, f := range parsed.WorldFacts {
		if _, err := g.store.CreateWorldFact(ctx, campaignID, f); err != nil {
			logger().Error("skeleton fact persistence failed", "campaign_id", campaignID, "category", f.Category, "error", err)
			return nil, fmt.Errorf("generate skeleton: create world fact: %w", err)
		}
	}

	result := &WorldSkeleton{
		Factions:             parsed.Factions,
		Locations:            parsed.Locations,
		NPCs:                 parsed.NPCs,
		WorldFacts:           parsed.WorldFacts,
		StartingLocationName: parsed.StartingLocation,
	}

	logger().Info("skeleton generation completed",
		"campaign_id", campaignID,
		"duration_ms", time.Since(started).Milliseconds(),
		"factions", len(result.Factions),
		"locations", len(result.Locations),
		"npcs", len(result.NPCs),
		"world_facts", len(result.WorldFacts),
	)
	return result, nil
}

func buildSkeletonPrompt(p *CampaignProfile) string {
	themes := "none specified"
	if len(p.Themes) > 0 {
		themes = strings.Join(p.Themes, ", ")
	}
	return prompt.BuildSkeletonPrompt(p.Genre, p.Tone, themes, p.WorldType, p.DangerLevel, p.PoliticalComplexity)
}
