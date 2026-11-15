package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenRouterStreamAssemblesInterleavedToolArguments(t *testing.T) {
	wire := strings.Join([]string{
		`data: {"choices":[{"delta":{"content":"Checking. ","tool_calls":[{"index":0,"id":"a","function":{"name":"probe","arguments":"{\"name\":"}},{"index":1,"id":"b","function":{"name":"probe","arguments":"{\"name\":\"second\"}"}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"first\"}"}}]},"finish_reason":"tool_calls"}]}`,
	}, "\n\n")
	ch := make(chan StreamChunk, 4)
	(&OpenRouterClient{}).parseSSEStream(context.Background(), strings.NewReader(wire), ch)
	close(ch)
	var chunks []StreamChunk
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}
	if len(chunks) != 4 || chunks[0].ContentDelta != "Checking. " || !chunks[3].Done {
		t.Fatalf("chunks = %+v", chunks)
	}
	for i, want := range []string{"first", "second"} {
		call := chunks[i+1].ToolCallDelta
		if call == nil || call.Name != "probe" || call.Arguments["name"] != want || call.ID != []string{"a", "b"}[i] {
			t.Fatalf("call %d = %+v", i, call)
		}
	}
}

func TestOpenRouterStreamMalformedToolArgumentsNeverComplete(t *testing.T) {
	wire := `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"a","function":{"name":"probe","arguments":"{"}}]},"finish_reason":"tool_calls"}]}`
	ch := make(chan StreamChunk, 2)
	(&OpenRouterClient{}).parseSSEStream(context.Background(), strings.NewReader(wire), ch)
	if len(ch) != 0 {
		t.Fatal("malformed tool arguments emitted a call or completion")
	}
}

func newTestOpenRouterServer(t *testing.T, servedBy string, gotBody *map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != openRouterChatPath {
			t.Errorf("path = %q, want %q", r.URL.Path, openRouterChatPath)
		}
		if err := json.NewDecoder(r.Body).Decode(gotBody); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"` + servedBy + `","choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
}

func TestOpenRouterSendsFallbackModelsAndReasoningEffort(t *testing.T) {
	var body map[string]any
	srv := newTestOpenRouterServer(t, "openai/gpt-6.1-terra", &body)
	defer srv.Close()

	c := NewOpenRouterClient("sk-or-test", "anthropic/claude-sonnet-5.5").
		WithFallbackModels([]string{" openai/gpt-6.1-terra ", "", "anthropic/claude-sonnet-5.5"}).
		WithReasoningEffort("high")
	c.baseURL = srv.URL

	resp, err := c.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("content = %q, want ok", resp.Content)
	}
	if body["model"] != "anthropic/claude-sonnet-5.5" {
		t.Fatalf("model = %v", body["model"])
	}
	models, _ := body["models"].([]any)
	if len(models) != 2 || models[0] != "anthropic/claude-sonnet-5.5" || models[1] != "openai/gpt-6.1-terra" {
		t.Fatalf("models = %v, want [primary, gpt-6.1-terra]", body["models"])
	}
	reasoning, _ := body["reasoning"].(map[string]any)
	if reasoning["effort"] != "high" {
		t.Fatalf("reasoning = %v, want effort high", body["reasoning"])
	}
}

func TestOpenRouterOmitsFallbackFieldsWhenUnset(t *testing.T) {
	var body map[string]any
	srv := newTestOpenRouterServer(t, "openai/gpt-4o-mini", &body)
	defer srv.Close()

	c := NewOpenRouterClient("sk-or-test", "openai/gpt-4o-mini")
	c.baseURL = srv.URL

	if _, err := c.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if _, ok := body["models"]; ok {
		t.Fatalf("models sent without fallbacks: %v", body["models"])
	}
	if _, ok := body["reasoning"]; ok {
		t.Fatalf("reasoning sent without effort: %v", body["reasoning"])
	}
}
