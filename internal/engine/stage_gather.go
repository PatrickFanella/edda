package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/tools"
)

// gatherStage loads the game state and recent session logs for the campaign.
// It also enriches tc.Ctx with campaign, player, and location IDs so that
// downstream stages (especially tool execution) can access them.
func (e *Engine) gatherStage() Stage {
	return func(ctx context.Context, tc *TurnContext) error {
		state, err := e.state.GatherState(ctx, tc.CampaignID)
		if err != nil {
			tc.Logger.Error("process turn failed during state gather",
				"campaign_id", tc.CampaignID,
				"duration_ms", time.Since(tc.Started).Milliseconds(),
				"error", err,
			)
			return fmt.Errorf("gather state: %w", err)
		}
		tc.State = state
		tc.Logger.Debug("state gathered",
			"campaign_id", tc.CampaignID,
			"player_id", state.Player.ID,
			"has_location", state.Player.CurrentLocationID != nil,
		)

		// Enrich context with campaign/player/location IDs for tool execution.
		enriched := tools.WithCurrentCampaignID(ctx, tc.CampaignID)
		if state.Player.ID != (uuid.UUID{}) {
			enriched = tools.WithCurrentPlayerCharacterID(enriched, state.Player.ID)
		}
		if state.Player.CurrentLocationID != nil {
			enriched = tools.WithCurrentLocationID(enriched, *state.Player.CurrentLocationID)
		}
		tc.Ctx = enriched

		recentTurns, err := e.state.ListRecentSessionLogs(enriched, tc.CampaignID, recentTurnLimit)
		if err != nil {
			tc.Logger.Error("process turn failed during session-log fetch",
				"campaign_id", tc.CampaignID,
				"duration_ms", time.Since(tc.Started).Milliseconds(),
				"error", err,
			)
			return fmt.Errorf("list recent session logs: %w", err)
		}
		tc.RecentLogs = recentTurns

		return nil
	}
}
