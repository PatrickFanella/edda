package engine

import (
	"context"
	"fmt"
	"time"
)

// processStage calls the LLM via the TurnProcessor, collecting raw narrative
// and applied tool calls. Deterministic completion/persistence happens in the
// turn completion stage.
func (e *Engine) processStage() Stage {
	return func(ctx context.Context, tc *TurnContext) error {
		narrative, applied, err := e.processor.ProcessWithRecoveryWithOptions(ctx, tc.Messages, tc.FilteredTools, TurnProcessorOptions{CandidateCallback: func(text string) error { return e.recordRawCandidate(ctx, text) }})
		if err != nil {
			tc.Logger.Error("process turn failed during turn processor",
				"campaign_id", tc.CampaignID,
				"duration_ms", time.Since(tc.Started).Milliseconds(),
				"error", err,
			)
			return fmt.Errorf("process turn: %w", err)
		}

		tc.Narrative = narrative
		tc.Applied = applied

		return e.recordCandidate(ctx, tc)
	}
}
