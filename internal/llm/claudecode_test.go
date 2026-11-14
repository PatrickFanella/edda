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

func TestClaudeCodeClientRequestAndToolCalls(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/run" || r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("path=%s auth=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = io.WriteString(w, `{"structured_output":{"text":"You walk in.","tool_calls":[{"name":"move_player","arguments":{"to":"inn"}}]},"input_tokens":100,"output_tokens":20}`)
	}))
	defer srv.Close()

	c := NewClaudeCodeClient(srv.URL+"/", "tok")
	resp, err := c.Complete(context.Background(), []Message{
		{Role: RoleSystem, Content: "You are the GM."},
		{Role: RoleUser, Content: "Go to the inn."},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "toolu_9", Name: "move_player", Arguments: map[string]any{"to": "gate"}}}},
		{Role: RoleTool, ToolCallID: "toolu_9", Content: `{"ok":true}`},
	}, []Tool{{Name: "move_player", Description: "Move", Parameters: map[string]any{"type": "object"}}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "You walk in." || len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Arguments["to"] != "inn" || resp.ToolCalls[0].ID == "" {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.FinishReason != "tool_calls" || resp.Usage.TotalTokens != 120 {
		t.Errorf("finish/usage = %q/%+v", resp.FinishReason, resp.Usage)
	}
	system, _ := got["system"].(string)
	prompt, _ := got["prompt"].(string)
	if !strings.HasPrefix(system, "You are the GM.") || !strings.Contains(system, "response schema") {
		t.Errorf("system = %q", system)
	}
	for _, want := range []string{`"name":"move_player"`, "<user>\nGo to the inn.\n</user>", `<tool_call id="toolu_9" name="move_player">{"to":"gate"}</tool_call>`, `<tool_result id="toolu_9">`} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q:\n%s", want, prompt)
		}
	}
	if got["json_schema"] == nil {
		t.Error("request has no json_schema")
	}
}

func TestClaudeCodeClientRejectsUnknownTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"structured_output":{"text":"","tool_calls":[{"name":"rm_rf","arguments":{}}]}}`)
	}))
	defer srv.Close()
	_, err := NewClaudeCodeClient(srv.URL, "tok").Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, []Tool{{Name: "move_player"}})
	var malformed *ErrMalformedResponse
	if !errors.As(err, &malformed) {
		t.Fatalf("err = %v, want ErrMalformedResponse", err)
	}
}

func TestClaudeCodeClientErrorStatuses(t *testing.T) {
	for _, tc := range []struct {
		status int
		check  func(error) bool
	}{
		{http.StatusTooManyRequests, func(err error) bool { var e *ErrRateLimit; return errors.As(err, &e) }},
		{http.StatusUnauthorized, func(err error) bool { var e *ErrAuth; return errors.As(err, &e) }},
		{http.StatusBadGateway, func(err error) bool { var e *ErrTransient; return errors.As(err, &e) }},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", tc.status)
		}))
		_, err := NewClaudeCodeClient(srv.URL, "tok").Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil)
		srv.Close()
		if !tc.check(err) {
			t.Errorf("status %d: err = %T %v", tc.status, err, err)
		}
	}
}
