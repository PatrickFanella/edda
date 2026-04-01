package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/PatrickFanella/game-master/internal/llm"
)

// stubLLMProvider is a test double for llm.Provider that returns canned
// responses from Complete and always errors on Stream.
type stubLLMProvider struct {
	completeContent string
	completeErr     error
	// capturedMessages stores messages from the last Complete call so tests
	// can inspect what was sent.
	capturedMessages []llm.Message
}

func (s *stubLLMProvider) Complete(_ context.Context, messages []llm.Message, _ []llm.Tool) (*llm.Response, error) {
	s.capturedMessages = messages
	if s.completeErr != nil {
		return nil, s.completeErr
	}
	return &llm.Response{Content: s.completeContent}, nil
}

func (s *stubLLMProvider) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("stream not supported in stub")
}

func TestSummarizeTurn_Success(t *testing.T) {
	stub := &stubLLMProvider{
		completeContent: `{"summary":"The party defeated the goblin.","location":"Dark Cave","npcs":["Gruk"],"event_type":"combat","significance":"high"}`,
	}
	sum := NewSummarizer(stub)

	res, err := sum.SummarizeTurn(context.Background(), "attack goblin", "You swing your sword.", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "The party defeated the goblin." {
		t.Errorf("summary = %q, want %q", res.Summary, "The party defeated the goblin.")
	}
	if res.Location != "Dark Cave" {
		t.Errorf("location = %q, want %q", res.Location, "Dark Cave")
	}
	if len(res.NPCs) != 1 || res.NPCs[0] != "Gruk" {
		t.Errorf("npcs = %v, want [Gruk]", res.NPCs)
	}
	if res.EventType != "combat" {
		t.Errorf("event_type = %q, want %q", res.EventType, "combat")
	}
	if res.Significance != "high" {
		t.Errorf("significance = %q, want %q", res.Significance, "high")
	}
}

func TestSummarizeTurn_MalformedJSON(t *testing.T) {
	stub := &stubLLMProvider{
		completeContent: "I cannot produce JSON right now, sorry.",
	}
	sum := NewSummarizer(stub)

	res, err := sum.SummarizeTurn(context.Background(), "look around", "You see a forest.", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "I cannot produce JSON right now, sorry." {
		t.Errorf("summary = %q, want raw fallback", res.Summary)
	}
	if res.EventType != "other" {
		t.Errorf("event_type = %q, want %q", res.EventType, "other")
	}
	if res.Significance != "medium" {
		t.Errorf("significance = %q, want %q", res.Significance, "medium")
	}
}

func TestSummarizeTurn_LLMError(t *testing.T) {
	stub := &stubLLMProvider{
		completeErr: errors.New("provider unavailable"),
	}
	sum := NewSummarizer(stub)

	_, err := sum.SummarizeTurn(context.Background(), "go north", "...", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, stub.completeErr) {
		t.Errorf("error = %v, want wrapping of %v", err, stub.completeErr)
	}
}

func TestSummarizeTurn_EmptyInput(t *testing.T) {
	stub := &stubLLMProvider{
		completeContent: `{"summary":"Nothing happened.","location":"","npcs":[],"event_type":"other","significance":"low"}`,
	}
	sum := NewSummarizer(stub)

	res, err := sum.SummarizeTurn(context.Background(), "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "Nothing happened." {
		t.Errorf("summary = %q", res.Summary)
	}
}

func TestSummarizeTurn_AllEventTypes(t *testing.T) {
	types := []string{"combat", "dialogue", "exploration", "quest_update", "discovery", "trade", "other"}
	for _, et := range types {
		t.Run(et, func(t *testing.T) {
			stub := &stubLLMProvider{
				completeContent: `{"summary":"x","location":"","npcs":[],"event_type":"` + et + `","significance":"low"}`,
			}
			sum := NewSummarizer(stub)

			res, err := sum.SummarizeTurn(context.Background(), "input", "response", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.EventType != et {
				t.Errorf("event_type = %q, want %q", res.EventType, et)
			}
		})
	}
}

func TestSummarizeTurn_PromptContainsInput(t *testing.T) {
	stub := &stubLLMProvider{
		completeContent: `{"summary":"ok","location":"","npcs":[],"event_type":"other","significance":"low"}`,
	}
	sum := NewSummarizer(stub)

	_, err := sum.SummarizeTurn(context.Background(), "cast fireball", "The goblin burns.", "spell_damage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stub.capturedMessages) < 2 {
		t.Fatalf("expected >=2 messages, got %d", len(stub.capturedMessages))
	}
	user := stub.capturedMessages[1]
	if user.Role != llm.RoleUser {
		t.Fatalf("message[1] role = %q, want user", user.Role)
	}
	for _, want := range []string{"cast fireball", "The goblin burns.", "spell_damage"} {
		if !contains(user.Content, want) {
			t.Errorf("user message missing %q", want)
		}
	}
}

func TestSummarizeTurn_MarkdownFences(t *testing.T) {
	stub := &stubLLMProvider{
		completeContent: "```json\n{\"summary\":\"Fenced.\",\"location\":\"\",\"npcs\":[],\"event_type\":\"other\",\"significance\":\"low\"}\n```",
	}
	sum := NewSummarizer(stub)

	res, err := sum.SummarizeTurn(context.Background(), "x", "y", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "Fenced." {
		t.Errorf("summary = %q, want %q", res.Summary, "Fenced.")
	}
}

func TestSummarizeTurn_PartialJSON(t *testing.T) {
	// Missing significance and npcs fields — should zero-value, not error.
	stub := &stubLLMProvider{
		completeContent: `{"summary":"Partial.","event_type":"dialogue"}`,
	}
	sum := NewSummarizer(stub)

	res, err := sum.SummarizeTurn(context.Background(), "x", "y", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "Partial." {
		t.Errorf("summary = %q", res.Summary)
	}
	if res.Significance != "" {
		t.Errorf("significance = %q, want zero value", res.Significance)
	}
	if res.NPCs == nil {
		t.Error("npcs should be non-nil empty slice")
	}
}

func TestSummarizeTurn_EmptyLLMResponse(t *testing.T) {
	stub := &stubLLMProvider{
		completeContent: "",
	}
	sum := NewSummarizer(stub)

	_, err := sum.SummarizeTurn(context.Background(), "x", "y", "")
	if err == nil {
		t.Fatal("expected error for empty LLM response")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
