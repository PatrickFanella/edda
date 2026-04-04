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

// campaignNameResponse is the expected JSON shape returned by the LLM.
type campaignNameResponse struct {
	Name string `json:"name"`
}

// GenerateCampaignName asks the LLM to produce a campaign name that fits the
// given profile. It returns a single evocative name.
func GenerateCampaignName(ctx context.Context, provider llm.Provider, profile *CampaignProfile) (string, error) {
	started := time.Now()
	if profile == nil {
		err := fmt.Errorf("generate campaign name: profile is nil")
		logger().Error("campaign naming failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return "", err
	}

	themes := strings.Join(profile.Themes, ", ")
	promptText := prompt.BuildCampaignNamePrompt(
		profile.Genre,
		profile.Tone,
		themes,
		profile.WorldType,
	)
	logger().Info("campaign naming started",
		"genre", profile.Genre,
		"tone", profile.Tone,
		"themes", len(profile.Themes),
		"prompt_len", len(promptText),
	)

	resp, err := provider.Complete(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: promptText},
	}, nil)
	if err != nil {
		logger().Error("campaign naming failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return "", fmt.Errorf("generate campaign name: llm call: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		err := fmt.Errorf("generate campaign name: empty LLM response")
		logger().Error("campaign naming failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return "", err
	}

	content = llmutil.StripMarkdownFences(content)

	var parsed campaignNameResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		logger().Error("campaign naming parse failed",
			"duration_ms", time.Since(started).Milliseconds(),
			"response_len", len(content),
			"error", err,
		)
		return "", fmt.Errorf("generate campaign name: parse response: %w", err)
	}

	name := strings.TrimSpace(parsed.Name)
	if name == "" {
		err := fmt.Errorf("generate campaign name: name is empty")
		logger().Error("campaign naming failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return "", err
	}

	logger().Info("campaign naming completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"name", name,
	)
	return name, nil
}
