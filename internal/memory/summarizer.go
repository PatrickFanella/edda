package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PatrickFanella/game-master/internal/llm"
	"github.com/PatrickFanella/game-master/internal/llmutil"
	"github.com/PatrickFanella/game-master/internal/prompt"
)


// SummaryResult holds the structured output of a turn summarization.
type SummaryResult struct {
	// Summary is a 1-3 sentence prose summary of what happened in the turn.
	Summary string `json:"summary"`
	// Location is the extracted location name or ID (empty if none).
	Location string `json:"location"`
	// NPCs lists the NPC names involved in the turn.
	NPCs []string `json:"npcs"`
	// EventType classifies the turn: combat, dialogue, exploration,
	// quest_update, discovery, trade, or other.
	EventType string `json:"event_type"`
	// Significance rates the turn's importance: low, medium, high, or critical.
	Significance string `json:"significance"`
}

// Summarizer uses an LLM provider to distill game turns into structured
// summaries suitable for long-term memory storage.
type Summarizer struct {
	llm llm.Provider
}

// NewSummarizer returns a Summarizer backed by the given LLM provider.
func NewSummarizer(provider llm.Provider) *Summarizer {
	return &Summarizer{llm: provider}
}

// SummarizeTurn asks the LLM to produce a structured summary of a single
// game turn described by the player's input, the game's narrative response,
// and any tool calls that were executed.
//
// If the LLM returns unparseable JSON, the method falls back to a best-effort
// result using the raw response as the summary.
func (s *Summarizer) SummarizeTurn(ctx context.Context, playerInput string, llmResponse string, toolCalls string) (*SummaryResult, error) {
	userContent := fmt.Sprintf(
		"Player action: %s\n\nGame response: %s\n\nTool calls: %s",
		playerInput, llmResponse, toolCalls,
	)

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: prompt.Summarizer},
		{Role: llm.RoleUser, Content: userContent},
	}

	resp, err := s.llm.Complete(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("summarize turn: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return nil, fmt.Errorf("summarize turn: empty LLM response: %w", &ErrEmptyInput{})
	}

	content = llmutil.StripMarkdownFences(content)

	var result SummaryResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Fallback: use raw response as summary with safe defaults.
		return &SummaryResult{
			Summary:      content,
			EventType:    "other",
			Significance: "medium",
		}, nil
	}

	// Ensure NPCs is never nil so callers can range without checking.
	if result.NPCs == nil {
		result.NPCs = []string{}
	}

	return &result, nil
}

