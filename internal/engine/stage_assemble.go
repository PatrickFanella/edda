package engine

import (
	"context"
)

// assembleStage builds the LLM message array and filters available tools
// based on the current game phase and state.
func (e *Engine) assembleStage() Stage {
	return func(ctx context.Context, tc *TurnContext) error {
		tc.Messages = e.assembler.AssembleContext(tc.State, tc.RecentLogs, tc.PlayerInput, tc.Memories...)

		tc.AllTools = e.assembler.Tools()
		tc.FilteredTools = tc.AllTools
		if e.toolFilter != nil {
			tc.FilteredTools = e.toolFilter.Filter(tc.State, tc.AllTools)
		}

		phase := DetectPhase(tc.State)
		tc.Logger.Debug("context assembled",
			"campaign_id", tc.CampaignID,
			"messages", len(tc.Messages),
			"recent_turns", len(tc.RecentLogs),
			"memories", len(tc.Memories),
			"all_tools", len(tc.AllTools),
			"filtered_tools", len(tc.FilteredTools),
			"phase", phase.String(),
		)
		return nil
	}
}
