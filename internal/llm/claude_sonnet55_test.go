package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClaudeSonnet55ThinkingReplayEffortAndFallback(t *testing.T) {
	var bodies []map[string]any
	var betas []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		bodies = append(bodies, body)
		betas = append(betas, r.Header.Get("anthropic-beta"))
		w.Header().Set("Content-Type", "application/json")
		if len(bodies) == 1 {
			_, _ = io.WriteString(w, `{"content":[
				{"type":"thinking","thinking":"","signature":"sig-1"},
				{"type":"tool_use","id":"toolu_1","name":"move_player","input":{"to":"inn"}},
				{"type":"tool_use","id":"toolu_2","name":"roll","input":{}}
			],"stop_reason":"tool_use","usage":{"input_tokens":5,"output_tokens":3}}`)
			return
		}
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"You reach the inn."}],"stop_reason":"end_turn","usage":{"input_tokens":5,"output_tokens":3}}`)
	}))
	defer srv.Close()

	c := NewClaudeClient(srv.URL, "test-key", "").WithEffort("high")
	history := []Message{{Role: RoleSystem, Content: "GM"}, {Role: RoleUser, Content: "Go to the inn."}}
	resp, err := c.Complete(context.Background(), history, []Tool{{Name: "move_player", Parameters: map[string]any{"type": "object"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ToolCalls) != 2 || resp.ToolCalls[0].ID != "toolu_1" {
		t.Fatalf("tool calls = %+v", resp.ToolCalls)
	}

	first := bodies[0]
	if first["model"] != "claude-sonnet-5-5" || first["fallbacks"] != "default" || betas[0] != claudeFallbackBeta {
		t.Errorf("model/fallbacks/beta = %v/%v/%q", first["model"], first["fallbacks"], betas[0])
	}
	if oc, _ := first["output_config"].(map[string]any); oc["effort"] != "high" {
		t.Errorf("output_config = %v, want effort high", first["output_config"])
	}
	if _, ok := first["thinking"]; ok {
		t.Errorf("thinking should be left to the model default, got %v", first["thinking"])
	}
	if first["max_tokens"].(float64) < 16000 {
		t.Errorf("max_tokens = %v, want at least 16000", first["max_tokens"])
	}

	follow := append(append([]Message{}, history...),
		Message{Role: RoleAssistant, ToolCalls: resp.ToolCalls},
		Message{Role: RoleTool, ToolCallID: "toolu_1", Content: `{"ok":true}`},
		Message{Role: RoleTool, ToolCallID: "toolu_2", Content: `{"roll":4}`},
	)
	resp, err = c.Complete(context.Background(), follow, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "You reach the inn." {
		t.Fatalf("content = %q", resp.Content)
	}

	msgs := bodies[1]["messages"].([]any)
	if len(msgs) != 3 {
		t.Fatalf("follow-up has %d messages, want user, assistant, one user turn with both results", len(msgs))
	}
	assistant := msgs[1].(map[string]any)["content"].([]any)
	if b := assistant[0].(map[string]any); b["type"] != "thinking" || b["signature"] != "sig-1" {
		t.Errorf("thinking block not replayed unchanged: %v", b)
	}
	results := msgs[2].(map[string]any)["content"].([]any)
	if len(results) != 2 {
		t.Errorf("tool results = %v, want both in one user message", results)
	}
}

func TestClaudeRefusalIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[],"stop_reason":"refusal","stop_details":{"type":"refusal","category":"general_harms"},"usage":{}}`)
	}))
	defer srv.Close()
	_, err := NewClaudeClient(srv.URL, "k", "").Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil)
	var malformed *ErrMalformedResponse
	if !errors.As(err, &malformed) || !strings.Contains(err.Error(), "general_harms") {
		t.Fatalf("err = %v, want a refusal error naming the category", err)
	}
}

func TestClaudeOlderModelSkipsServerFallback(t *testing.T) {
	var body map[string]any
	var beta string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		beta = r.Header.Get("anthropic-beta")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{}}`)
	}))
	defer srv.Close()
	if _, err := NewClaudeClient(srv.URL, "k", "claude-haiku-4-5").Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["fallbacks"]; ok || beta != "" {
		t.Errorf("fallbacks=%v beta=%q, want neither for claude-haiku-4-5", body["fallbacks"], beta)
	}
	if _, ok := body["output_config"]; ok {
		t.Errorf("output_config sent without a configured effort: %v", body["output_config"])
	}
}
