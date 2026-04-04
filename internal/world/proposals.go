package world

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/llmutil"
	"github.com/PatrickFanella/game-master/internal/prompt"
)

// CampaignProposal is a single campaign option presented to the player during
// campaign creation. It pairs a human-readable name and summary with the full
// CampaignProfile that would be used to generate the world.
type CampaignProposal struct {
	Name    string          `json:"name"`
	Summary string          `json:"summary"`
	Profile CampaignProfile `json:"profile"`
}

// proposalLLMResponse is the expected JSON object shape returned by the LLM.
type proposalLLMResponse struct {
	Proposals []proposalEntry `json:"proposals"`
}

// proposalProfilePayload supports providers that nest the profile under a
// separate object instead of returning flat proposal fields.
type proposalProfilePayload struct {
	Genre               string   `json:"genre"`
	Tone                string   `json:"tone"`
	Themes              []string `json:"themes"`
	WorldType           string   `json:"world_type"`
	DangerLevel         string   `json:"danger_level"`
	PoliticalComplexity string   `json:"political_complexity"`
}

// proposalEntry is the per-proposal shape the LLM may produce. It accepts both
// the flat fields requested by the prompt and an alternate nested `profile`
// object so we can recover from small schema drift.
type proposalEntry struct {
	Name                string                  `json:"name"`
	Summary             string                  `json:"summary"`
	Genre               string                  `json:"genre"`
	Tone                string                  `json:"tone"`
	Themes              []string                `json:"themes"`
	WorldType           string                  `json:"world_type"`
	DangerLevel         string                  `json:"danger_level"`
	PoliticalComplexity string                  `json:"political_complexity"`
	Profile             *proposalProfilePayload `json:"profile"`
}

// ProposalGenerator produces campaign proposals by prompting an LLM with the
// player's stated preferences.
type ProposalGenerator struct {
	llm llm.Provider
}

// NewProposalGenerator returns a generator wired to the given LLM provider.
func NewProposalGenerator(provider llm.Provider) *ProposalGenerator {
	return &ProposalGenerator{llm: provider}
}

// Generate produces campaign proposals from the given player preferences by
// prompting the LLM and parsing its structured response. It returns an error
// if the LLM response is empty, cannot be parsed, or contains no proposals.
func (g *ProposalGenerator) Generate(ctx context.Context, genre, settingStyle, tone string) ([]CampaignProposal, error) {
	started := time.Now()
	logger().Info("proposal generation started",
		"genre", genre,
		"setting_style", settingStyle,
		"tone", tone,
	)
	promptText := prompt.BuildProposalsPrompt(genre, settingStyle, tone)
	logger().Debug("proposal prompt built", "prompt_len", len(promptText))

	resp, err := g.llm.Complete(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: promptText},
	}, nil)
	if err != nil {
		logger().Error("proposal generation failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, fmt.Errorf("generate proposals: llm call: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		err := fmt.Errorf("generate proposals: empty LLM response")
		logger().Error("proposal generation failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	entries, err := parseProposalEntries(content)
	if err != nil {
		logger().Error("proposal generation parse failed",
			"duration_ms", time.Since(started).Milliseconds(),
			"response_len", len(content),
			"error", err,
		)
		return nil, fmt.Errorf("generate proposals: parse response: %w", err)
	}
	if len(entries) == 0 {
		err := fmt.Errorf("generate proposals: LLM returned 0 proposals")
		logger().Error("proposal generation failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return nil, err
	}

	proposals := make([]CampaignProposal, len(entries))
	for i, entry := range entries {
		profile := entry.profilePayload()
		proposals[i] = CampaignProposal{
			Name:    strings.TrimSpace(entry.Name),
			Summary: strings.TrimSpace(entry.Summary),
			Profile: CampaignProfile{
				Genre:               strings.TrimSpace(profile.Genre),
				Tone:                strings.TrimSpace(profile.Tone),
				Themes:              append([]string(nil), profile.Themes...),
				WorldType:           strings.TrimSpace(profile.WorldType),
				DangerLevel:         strings.TrimSpace(profile.DangerLevel),
				PoliticalComplexity: strings.TrimSpace(profile.PoliticalComplexity),
			},
		}
	}

	logger().Info("proposal generation completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"proposals", len(proposals),
	)
	return proposals, nil
}

func parseProposalEntries(content string) ([]proposalEntry, error) {
	normalized := strings.TrimSpace(llmutil.StripMarkdownFences(content))

	if entries, err := decodeProposalEntries(normalized); err == nil {
		return entries, nil
	}

	extracted := extractLikelyJSON(normalized)
	if extracted == normalized {
		return nil, fmt.Errorf("no parseable JSON object found")
	}

	entries, err := decodeProposalEntries(extracted)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func decodeProposalEntries(raw string) ([]proposalEntry, error) {
	var object proposalLLMResponse
	if err := json.Unmarshal([]byte(raw), &object); err == nil && len(object.Proposals) > 0 {
		return object.Proposals, nil
	}

	var array []proposalEntry
	if err := json.Unmarshal([]byte(raw), &array); err == nil && len(array) > 0 {
		return array, nil
	}

	var objectErr error
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		objectErr = err
	}
	if objectErr != nil {
		return nil, objectErr
	}
	return object.Proposals, nil
}

func (e proposalEntry) profilePayload() proposalProfilePayload {
	if e.Profile != nil {
		return proposalProfilePayload{
			Genre:               firstNonEmpty(e.Genre, e.Profile.Genre),
			Tone:                firstNonEmpty(e.Tone, e.Profile.Tone),
			Themes:              firstNonEmptyThemes(e.Themes, e.Profile.Themes),
			WorldType:           firstNonEmpty(e.WorldType, e.Profile.WorldType),
			DangerLevel:         firstNonEmpty(e.DangerLevel, e.Profile.DangerLevel),
			PoliticalComplexity: firstNonEmpty(e.PoliticalComplexity, e.Profile.PoliticalComplexity),
		}
	}
	return proposalProfilePayload{
		Genre:               e.Genre,
		Tone:                e.Tone,
		Themes:              e.Themes,
		WorldType:           e.WorldType,
		DangerLevel:         e.DangerLevel,
		PoliticalComplexity: e.PoliticalComplexity,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonEmptyThemes(candidates ...[]string) []string {
	for _, themes := range candidates {
		if len(themes) > 0 {
			return themes
		}
	}
	return nil
}

func extractLikelyJSON(s string) string {
	start := strings.IndexAny(s, "[{")
	if start == -1 {
		return s
	}

	objectEnd := strings.LastIndex(s, "}")
	arrayEnd := strings.LastIndex(s, "]")
	end := max(objectEnd, arrayEnd)
	if end < start {
		return s
	}

	return strings.TrimSpace(s[start : end+1])
}
