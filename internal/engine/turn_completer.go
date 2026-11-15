package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
)

// TurnCompleter owns the deterministic post-LLM turn completion work: turning
// raw narrative plus applied tool calls into one persisted, player-facing turn.
type TurnCompleter struct {
	engine   *Engine
	provider llm.Provider
}

func NewTurnCompleter(engine *Engine, provider llm.Provider) *TurnCompleter {
	return &TurnCompleter{engine: engine, provider: provider}
}

func (c *TurnCompleter) CompleteAndPersist(ctx context.Context, tc *TurnContext) error {
	if c == nil || c.engine == nil {
		return fmt.Errorf("turn completer is not configured")
	}

	cleaned, choices, err := extractChoicesStrict(tc.Narrative)
	if err != nil {
		tc.Logger.Error("process turn failed during choice extraction",
			"campaign_id", tc.CampaignID,
			"duration_ms", time.Since(tc.Started).Milliseconds(),
			"error", err,
		)
		return fmt.Errorf("complete turn: %w", err)
	}

	// If no choices were parsed from the narrative, ask the model to generate some.
	if len(choices) == 0 && c.provider != nil {
		generated, genErr := c.generateChoices(ctx, tc.Narrative)
		if genErr == nil && len(generated) > 0 {
			choices = generated
		}
	}

	tc.Narrative, tc.Choices = cleaned, choices
	tc.CombatActive = combatActiveAfterApplied(tc.State.CombatActive, tc.Applied)
	tc.StateChanges = StateChangesFromAppliedToolCalls(tc.Applied)

	toolCallsJSON, err := marshalAppliedToolCalls(tc.Applied)
	if err != nil {
		tc.Logger.Error("process turn failed during tool-call marshal",
			"campaign_id", tc.CampaignID,
			"duration_ms", time.Since(tc.Started).Milliseconds(),
			"error", err,
		)
		return fmt.Errorf("marshal applied tool calls: %w", err)
	}
	tc.ToolCallsJSON = toolCallsJSON

	tc.TurnNumber = nextTurnNumber(tc.RecentLogs)
	log := domain.SessionLog{
		CampaignID:  tc.CampaignID,
		TurnNumber:  tc.TurnNumber,
		PlayerInput: tc.PlayerInput,
		InputType:   domain.Classify(tc.PlayerInput),
		LLMResponse: tc.Narrative,
		ToolCalls:   toolCallsJSON,
		LocationID:  finalLocationIDFromApplied(tc.State.Player.CurrentLocationID, tc.Applied),
	}
	if err := c.engine.persistCompletedTurn(ctx, tc, log); err != nil {
		tc.Logger.Error("process turn failed during session-log save",
			"campaign_id", tc.CampaignID,
			"duration_ms", time.Since(tc.Started).Milliseconds(),
			"error", err,
		)
		return fmt.Errorf("save session log: %w", err)
	}

	c.engine.snapshotQuestsIfNeeded(ctx, tc.CampaignID, tc.Applied)
	c.engine.autoSaveIfNeeded(ctx, tc.CampaignID, tc.TurnNumber)
	c.engine.autoSummarizeIfNeeded(ctx, tc.CampaignID, tc.TurnNumber)

	return nil
}

// generateChoices asks the model to produce 3-5 choices based on the narrative.
func (c *TurnCompleter) generateChoices(ctx context.Context, narrative string) ([]Choice, error) {
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "You are a choice generator for a tabletop RPG. Given a narrative scene, produce 3-5 distinct, concrete player choices. Each choice should be a single imperative sentence. Number them 1. 2. 3. etc. Return ONLY the numbered list, no other text."},
		{Role: llm.RoleUser, Content: fmt.Sprintf("Generate 3-5 player choices for this scene:\n\n%s", narrative)},
	}

	resp, err := c.provider.Complete(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("generate choices: %w", err)
	}

	return parseGeneratedChoices(resp.Content), nil
}

// parseGeneratedChoices extracts numbered choices from model output.
func parseGeneratedChoices(text string) []Choice {
	var choices []Choice
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Match "1. text" or "1) text" patterns.
		for _, prefix := range []string{fmt.Sprintf("%d.", i+1), fmt.Sprintf("%d)", i+1)} {
			if after, ok := strings.CutPrefix(line, prefix); ok {
				text := strings.TrimSpace(after)
				if text != "" {
					choices = append(choices, Choice{ID: fmt.Sprintf("%d", i+1), Text: text})
				}
				break
			}
		}
	}
	return choices
}

func combatActiveAfterApplied(initial bool, applied []AppliedToolCall) bool {
	combatActive := initial
	for _, atc := range applied {
		switch atc.Tool {
		case "initiate_combat":
			combatActive = true
		case "resolve_combat":
			combatActive = false
		}
	}
	return combatActive
}
