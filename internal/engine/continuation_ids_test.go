package engine

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
)

// recordingProvider returns responses in order and keeps the messages of each call.
type recordingProvider struct {
	responses []*llm.Response
	calls     [][]llm.Message
}

func (p *recordingProvider) Complete(_ context.Context, messages []llm.Message, _ []llm.Tool) (*llm.Response, error) {
	p.calls = append(p.calls, append([]llm.Message(nil), messages...))
	return p.responses[len(p.calls)-1], nil
}

func (p *recordingProvider) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	panic("not used")
}

// Hosted providers reject a tool result whose ID matches no tool call, so the
// continuation must answer each call under the call's own ID.
func TestTurnProcessor_ContinuationAnswersToolCallsByID(t *testing.T) {
	reg, _ := buildProcessorTestRegistry(t, 0)
	provider := &recordingProvider{responses: []*llm.Response{
		{ToolCalls: []llm.ToolCall{
			{ID: "toolu_1", Name: "mock_tool", Arguments: map[string]any{"name": "Ari"}},
			{ID: "toolu_2", Name: "mock_tool", Arguments: map[string]any{"name": "Bo"}},
		}},
		{Content: "Ari and Bo are greeted."},
	}}
	tp := NewTurnProcessor(provider, reg, tools.NewValidator(reg), nil)

	narrative, _, err := tp.ProcessWithRecovery(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "greet them"}}, reg.List())
	if err != nil {
		t.Fatalf("ProcessWithRecovery() error = %v", err)
	}
	if narrative != "Ari and Bo are greeted." {
		t.Fatalf("narrative = %q", narrative)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("provider calls = %d, want initial + continuation", len(provider.calls))
	}
	cont := provider.calls[1]
	results := cont[len(cont)-2:]
	for i, want := range []string{"toolu_1", "toolu_2"} {
		if results[i].Role != llm.RoleTool || results[i].ToolCallID != want {
			t.Fatalf("result %d = role %q id %q, want tool result for %q", i, results[i].Role, results[i].ToolCallID, want)
		}
	}
	if results[0].Content != `{"greeted":"Ari"}` || results[1].Content != `{"greeted":"Bo"}` {
		t.Fatalf("results = %q / %q", results[0].Content, results[1].Content)
	}
}
