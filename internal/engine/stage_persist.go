package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/PatrickFanella/game-master/internal/domain"
)

// persistStage saves the session log, snapshots quests, triggers auto-save,
// and triggers auto-summarization.
func (e *Engine) persistStage() Stage {
	return func(ctx context.Context, tc *TurnContext) error {
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
			LocationID:  tc.State.Player.CurrentLocationID,
		}
		if err := e.state.SaveSessionLog(ctx, log); err != nil {
			tc.Logger.Error("process turn failed during session-log save",
				"campaign_id", tc.CampaignID,
				"duration_ms", time.Since(tc.Started).Milliseconds(),
				"error", err,
			)
			return fmt.Errorf("save session log: %w", err)
		}

		e.snapshotQuestsIfNeeded(ctx, tc.CampaignID, tc.Applied)
		e.autoSaveIfNeeded(ctx, tc.CampaignID, tc.TurnNumber)
		e.autoSummarizeIfNeeded(ctx, tc.CampaignID, tc.TurnNumber)

		return nil
	}
}
