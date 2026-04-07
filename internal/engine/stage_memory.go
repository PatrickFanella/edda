package engine

import (
	"context"
)

// memoryStage retrieves tier-3 semantic memories relevant to the player's input.
// Memory retrieval failure is non-fatal; the turn proceeds without memories.
func (e *Engine) memoryStage() Stage {
	return func(ctx context.Context, tc *TurnContext) error {
		if e.tier3 == nil {
			return nil
		}
		memories, err := e.tier3.Retrieve(ctx, tc.CampaignID, tc.PlayerInput, tc.State)
		if err != nil {
			tc.Logger.Warn("tier3 memory retrieval failed",
				"campaign_id", tc.CampaignID,
				"error", err,
			)
			return nil // non-fatal
		}
		tc.Memories = memories
		tc.Logger.Debug("tier3 memories retrieved",
			"campaign_id", tc.CampaignID,
			"count", len(memories),
		)
		return nil
	}
}
