package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClaudeClientDefaults(t *testing.T) {
	client := NewClaudeClient("", "sk-ant-test", "")

	if client.baseURL != defaultClaudeBaseURL {
		t.Fatalf("baseURL = %q, want %q", client.baseURL, defaultClaudeBaseURL)
	}
	if client.model != defaultClaudeModel {
		t.Fatalf("model = %q, want %q", client.model, defaultClaudeModel)
	}
	if client.anthropicVersion != defaultAnthropicVersion {
		t.Fatalf("anthropicVersion = %q, want %q", client.anthropicVersion, defaultAnthropicVersion)
	}
	if client.client == nil {
		t.Fatal("http client must be configured")
	}
}

func TestClaudeClientCompleteRequestHeadersAndPayload(t *testing.T) {
	var gotReq claudeMessagesRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != claudeMessagesPath {
			t.Fatalf("path = %q, want %q", r.URL.Path, claudeMessagesPath)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("x-api-key"); got != "sk-ant-test" {
			t.Fatalf("x-api-key = %q, want %q", got, "sk-ant-test")
		}
		if got := r.Header.Get("anthropic-version"); got != defaultAnthropicVersion {
			t.Fatalf("anthropic-version = %q, want %q", got, defaultAnthropicVersion)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q, want application/json", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		_, _ = w.Write([]byte(`{
"content":[
  {"type":"text","text":"assistant reply"},
  {"type":"tool_use","id":"tool_1","name":"lookup_weather","input":{"city":"Paris"}}
],
"stop_reason":"end_turn",
"usage":{"input_tokens":11,"output_tokens":7}
}`))
	}))
	defer server.Close()

	client := NewClaudeClient(server.URL+"/", "sk-ant-test", "claude-test")
	resp, err := client.Complete(context.Background(), []Message{
		{Role: RoleSystem, Content: "be concise"},
		{Role: RoleUser, Content: "weather?"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "tool_1", Name: "lookup_weather", Arguments: map[string]any{"city": "Paris"}}}},
		{Role: RoleTool, ToolCallID: "tool_1", Content: `{"temp":"20C"}`},
	}, []Tool{{
		Name:        "lookup_weather",
		Description: "weather by city",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{"type": "string"},
			},
			"required": []string{"city"},
		},
	}})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if gotReq.Model != "claude-test" {
		t.Fatalf("model = %q, want claude-test", gotReq.Model)
	}
	if gotReq.System != "be concise" {
		t.Fatalf("system = %q, want %q", gotReq.System, "be concise")
	}
	if gotReq.MaxTokens != defaultClaudeMaxTokens {
		t.Fatalf("max_tokens = %d, want %d", gotReq.MaxTokens, defaultClaudeMaxTokens)
	}
	if len(gotReq.Messages) != 3 {
		t.Fatalf("messages length = %d, want 3", len(gotReq.Messages))
	}
	if gotReq.Messages[0].Role != claudeRoleUser || gotReq.Messages[0].Content[0].Type != claudeContentTypeText {
		t.Fatalf("first message not marshaled as user text: %#v", gotReq.Messages[0])
	}
	if gotReq.Messages[1].Role != claudeRoleAssistant || gotReq.Messages[1].Content[0].Type != claudeContentTypeToolUse {
		t.Fatalf("assistant tool call not marshaled correctly: %#v", gotReq.Messages[1])
	}
	if gotReq.Messages[2].Role != claudeRoleUser || gotReq.Messages[2].Content[0].Type != claudeContentTypeToolResp {
		t.Fatalf("tool result not marshaled correctly: %#v", gotReq.Messages[2])
	}
	if len(gotReq.Tools) != 1 || gotReq.Tools[0].Name != "lookup_weather" {
		t.Fatalf("tools not marshaled correctly: %#v", gotReq.Tools)
	}

	if resp.Content != "assistant reply" {
		t.Fatalf("response content = %q, want assistant reply", resp.Content)
	}
	if resp.FinishReason != "end_turn" {
		t.Fatalf("finish reason = %q, want end_turn", resp.FinishReason)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("tool calls length = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ID != "tool_1" || resp.ToolCalls[0].Name != "lookup_weather" {
		t.Fatalf("tool call = %#v, want id=tool_1 name=lookup_weather", resp.ToolCalls[0])
	}
	if resp.Usage.PromptTokens != 11 || resp.Usage.CompletionTokens != 7 || resp.Usage.TotalTokens != 18 {
		t.Fatalf("usage = %#v, want prompt=11 completion=7 total=18", resp.Usage)
	}
}
