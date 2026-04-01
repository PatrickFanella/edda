package world

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/llmutil"
	"github.com/PatrickFanella/game-master/internal/prompt"
)

// WorldSkeleton is the initial world state generated from a CampaignProfile.
type WorldSkeleton struct {
	Factions           []SkeletonFaction
	Locations          []SkeletonLocation
	NPCs               []SkeletonNPC
	WorldFacts         []SkeletonFact
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
	if profile == nil {
		return nil, fmt.Errorf("generate skeleton: profile is nil")
	}

	promptText := buildSkeletonPrompt(profile)

	resp, err := g.llm.Complete(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: promptText},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("generate skeleton: llm call: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return nil, fmt.Errorf("generate skeleton: empty LLM response")
	}

	content = llmutil.StripMarkdownFences(content)

	var parsed skeletonLLMResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("generate skeleton: parse response: %w", err)
	}

	// Persist factions, build name→ID map.
	factionIDs := make(map[string]uuid.UUID, len(parsed.Factions))
	for _, f := range parsed.Factions {
		id, err := g.store.CreateFaction(ctx, campaignID, f)
		if err != nil {
			return nil, fmt.Errorf("generate skeleton: create faction %q: %w", f.Name, err)
		}
		factionIDs[f.Name] = id
	}

	// Persist locations, build name→ID map.
	locationIDs := make(map[string]uuid.UUID, len(parsed.Locations))
	for _, l := range parsed.Locations {
		id, err := g.store.CreateLocation(ctx, campaignID, l)
		if err != nil {
			return nil, fmt.Errorf("generate skeleton: create location %q: %w", l.Name, err)
		}
		locationIDs[l.Name] = id
	}

	// Persist NPCs, resolving faction/location references where possible.
	for _, n := range parsed.NPCs {
		var factionID, locationID *uuid.UUID
		if fid, ok := factionIDs[n.Faction]; ok {
			factionID = &fid
		}
		if lid, ok := locationIDs[n.Location]; ok {
			locationID = &lid
		}
		if _, err := g.store.CreateNPC(ctx, campaignID, n, factionID, locationID); err != nil {
			return nil, fmt.Errorf("generate skeleton: create npc %q: %w", n.Name, err)
		}
	}

	// Persist world facts.
	for _, f := range parsed.WorldFacts {
		if _, err := g.store.CreateWorldFact(ctx, campaignID, f); err != nil {
			return nil, fmt.Errorf("generate skeleton: create world fact: %w", err)
		}
	}

	// Resolve starting location name.
	startName := parsed.StartingLocation

	return &WorldSkeleton{
		Factions:             parsed.Factions,
		Locations:            parsed.Locations,
		NPCs:                 parsed.NPCs,
		WorldFacts:           parsed.WorldFacts,
		StartingLocationName: startName,
	}, nil
}

func buildSkeletonPrompt(p *CampaignProfile) string {
	themes := "none specified"
	if len(p.Themes) > 0 {
		themes = strings.Join(p.Themes, ", ")
	}
	return prompt.BuildSkeletonPrompt(p.Genre, p.Tone, themes, p.WorldType, p.DangerLevel, p.PoliticalComplexity)
}
