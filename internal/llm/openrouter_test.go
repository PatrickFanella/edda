package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
