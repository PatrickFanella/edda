package engine

import (
	"context"
	"fmt"
	"time"
)

// processStage calls the LLM via the TurnProcessor, extracting narrative,
// applied tool calls, choices, and combat state from the response.
func (e *Engine) processStage() Stage {
	return func(ctx context.Context, tc *TurnContext) error {
		narrative, applied, err := e.processor.ProcessWithRecovery(ctx, tc.Messages, tc.FilteredTools)
		if err != nil {
			tc.Logger.Error("process turn failed during turn processor",
				"campaign_id", tc.CampaignID,
				"duration_ms", time.Since(tc.Started).Milliseconds(),
				"error", err,
			)
			return fmt.Errorf("process turn: %w", err)
		}

		tc.Narrative, tc.Choices = extractChoices(narrative)
		tc.Applied = applied

		// Derive combat state from pre-turn state and applied tool calls.
		tc.CombatActive = tc.State.CombatActive
		for _, atc := range applied {
			switch atc.Tool {
			case "initiate_combat":
				tc.CombatActive = true
			case "resolve_combat":
				tc.CombatActive = false
			}
		}

		return nil
	}
}
