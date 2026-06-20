package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
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
			LocationID:  finalLocationIDFromApplied(tc.State.Player.CurrentLocationID, tc.Applied),
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

func finalLocationIDFromApplied(defaultLocationID *uuid.UUID, applied []AppliedToolCall) *uuid.UUID {
	locationID := defaultLocationID
	for _, call := range applied {
		var data map[string]any
		if err := json.Unmarshal(call.Result, &data); err != nil {
			continue
		}

		switch call.Tool {
		case "move_player":
			if id, ok := uuidStringField(data, "location_id"); ok {
				locationID = &id
			}
		case "create_location":
			if moved, _ := data["move_player_here"].(bool); moved {
				if id, ok := uuidStringField(data, "location_id"); ok {
					locationID = &id
					continue
				}
				if id, ok := uuidStringField(data, "id"); ok {
					locationID = &id
				}
			}
		}
	}
	return locationID
}

func uuidStringField(data map[string]any, key string) (uuid.UUID, bool) {
	raw, ok := data[key]
	if !ok {
		return uuid.UUID{}, false
	}
	str, ok := raw.(string)
	if !ok {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(str)
	return id, err == nil
}
