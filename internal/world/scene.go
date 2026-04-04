package world

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/domain"
	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/llmutil"
	"github.com/PatrickFanella/game-master/internal/prompt"
)

// SceneResult holds the opening scene narrative and initial player choices
// produced by the SceneGenerator.
type SceneResult struct {
	Narrative string   // Opening scene prose
	Choices   []string // 2-4 initial player choices
}

// SceneStore persists session logs during scene generation.
type SceneStore interface {
	SaveSessionLog(ctx context.Context, log domain.SessionLog) error
}

// SceneGenerator produces an opening scene from a campaign profile and world
// skeleton by prompting an LLM and persisting the result as a session log.
type SceneGenerator struct {
	llm   llm.Provider
	store SceneStore
}

// NewSceneGenerator returns a generator wired to the given LLM and store.
func NewSceneGenerator(provider llm.Provider, store SceneStore) *SceneGenerator {
	return &SceneGenerator{llm: provider, store: store}
}

// sceneLLMResponse is the expected JSON shape returned by the LLM.
type sceneLLMResponse struct {
	Narrative string   `json:"narrative"`
	Choices   []string `json:"choices"`
}

// Generate asks the LLM to produce an opening scene for the given campaign,
// persists the narrative as a session log, and returns the scene result.
func (g *SceneGenerator) Generate(ctx context.Context, campaignID uuid.UUID, profile *CampaignProfile, skeleton *WorldSkeleton) (*SceneResult, error) {
	started := time.Now()
	if profile == nil {
		err := fmt.Errorf("generate scene: profile is nil")
		logger().Error("scene generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}
	if skeleton == nil {
		err := fmt.Errorf("generate scene: skeleton is nil")
		logger().Error("scene generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	logger().Info("scene generation started",
		"campaign_id", campaignID,
		"starting_location", skeleton.StartingLocationName,
		"npcs", len(skeleton.NPCs),
		"world_facts", len(skeleton.WorldFacts),
	)
	promptText := buildScenePrompt(profile, skeleton)
	logger().Debug("scene prompt built", "campaign_id", campaignID, "prompt_len", len(promptText))

	resp, err := g.llm.Complete(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: promptText},
	}, nil)
	if err != nil {
		logger().Error("scene generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, fmt.Errorf("generate scene: llm call: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		err := fmt.Errorf("generate scene: empty LLM response")
		logger().Error("scene generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	content = llmutil.StripMarkdownFences(content)

	var parsed sceneLLMResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		logger().Error("scene generation parse failed",
			"campaign_id", campaignID,
			"duration_ms", time.Since(started).Milliseconds(),
			"response_len", len(content),
			"error", err,
		)
		return nil, fmt.Errorf("generate scene: parse response: %w", err)
	}

	if strings.TrimSpace(parsed.Narrative) == "" {
		err := fmt.Errorf("generate scene: narrative is empty")
		logger().Error("scene generation failed", "campaign_id", campaignID, "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	log := domain.SessionLog{
		ID:          uuid.New(),
		CampaignID:  campaignID,
		TurnNumber:  1,
		PlayerInput: "[scene_generation]",
		InputType:   domain.InputTypeNarrative,
		LLMResponse: parsed.Narrative,
	}
	if err := g.store.SaveSessionLog(ctx, log); err != nil {
		logger().Error("scene session log persistence failed", "campaign_id", campaignID, "error", err)
		return nil, fmt.Errorf("generate scene: save session log: %w", err)
	}

	result := &SceneResult{
		Narrative: parsed.Narrative,
		Choices:   parsed.Choices,
	}
	logger().Info("scene generation completed",
		"campaign_id", campaignID,
		"duration_ms", time.Since(started).Milliseconds(),
		"narrative_len", len(result.Narrative),
		"choices", len(result.Choices),
	)
	return result, nil
}

// buildScenePrompt constructs the system prompt for opening scene generation.
func buildScenePrompt(profile *CampaignProfile, skeleton *WorldSkeleton) string {
	themes := "none specified"
	if len(profile.Themes) > 0 {
		themes = strings.Join(profile.Themes, ", ")
	}

	// Resolve starting location details from the skeleton.
	startingLocation := "unknown"
	for _, loc := range skeleton.Locations {
		if loc.Name == skeleton.StartingLocationName {
			startingLocation = loc.Name + " \u2014 " + loc.Description
			break
		}
	}

	// Collect NPCs at the starting location.
	var npcsAtStart []string
	for _, npc := range skeleton.NPCs {
		if npc.Location != skeleton.StartingLocationName {
			continue
		}
		npcsAtStart = append(npcsAtStart, npc.Name+" ("+npc.Description+")")
	}
	npcList := "none"
	if len(npcsAtStart) > 0 {
		npcList = strings.Join(npcsAtStart, ", ")
	}

	// Collect world facts.
	var facts []string
	for _, f := range skeleton.WorldFacts {
		facts = append(facts, f.Fact)
	}
	factList := "none"
	if len(facts) > 0 {
		factList = strings.Join(facts, "; ")
	}

	return prompt.BuildScenePrompt(profile.Genre, profile.Tone, themes, startingLocation, npcList, factList)
}
