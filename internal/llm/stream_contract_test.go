package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenRouterStreamErrorDoesNotCompleteOrDispatchCalls(t *testing.T) {
	for _, terminal := range []string{
		`data: {"error":{"message":"failed","code":503},"choices":[{"delta":{},"finish_reason":"error"}]}`,
		`data: {"error":{"message":"failed","code":503}}`,
		`data: {"choices":[{"delta":{},"finish_reason":"error"}]}`,
	} {
		t.Run(terminal, func(t *testing.T) {
			wire := `data: {"choices":[{"delta":{"content":"Abandoned","tool_calls":[{"index":0,"id":"tool","function":{"name":"update_player_hp","arguments":"{\"hp\":1}"}}]}}]}` + "\n\n" + terminal + "\n\ndata: [DONE]\n\n"
			ch := make(chan StreamChunk, 10)
			(&OpenRouterClient{}).parseSSEStream(context.Background(), strings.NewReader(wire), ch)
			close(ch)
			for chunk := range ch {
				if chunk.Done || chunk.ToolCallDelta != nil {
					t.Fatalf("failed stream accepted: %+v", chunk)
				}
			}
		})
	}
}

func TestClaudeStreamRefusalDoesNotComplete(t *testing.T) {
	wire := "event: content_block_delta\ndata: {\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"I cannot help with that.\"}}\n\nevent: message_delta\ndata: {\"delta\":{\"stop_reason\":\"refusal\"}}\n\nevent: message_stop\ndata: {}\n\n"
	ch := make(chan StreamChunk, 10)
	(&ClaudeClient{}).parseSSEStream(context.Background(), strings.NewReader(wire), ch)
	close(ch)
	for chunk := range ch {
		if chunk.Done {
			t.Fatal("refusal accepted as completed narration")
		}
	}
}

func TestClaudeStreamThinkingReplay(t *testing.T) {
	wire := strings.Join([]string{
		`event: content_block_start`, `data: {"index":0,"content_block":{"type":"thinking","thinking":""}}`, ``,
		`event: content_block_delta`, `data: {"index":0,"delta":{"type":"thinking_delta","thinking":"Private reasoning"}}`, ``,
		`event: content_block_delta`, `data: {"index":0,"delta":{"type":"signature_delta","signature":"signed-thinking"}}`, ``,
		`event: content_block_stop`, `data: {"index":0}`, ``,
		`event: content_block_start`, `data: {"index":1,"content_block":{"type":"redacted_thinking","data":"encrypted"}}`, ``,
		`event: content_block_stop`, `data: {"index":1}`, ``,
		`event: content_block_start`, `data: {"index":2,"content_block":{"type":"text","text":""}}`, ``,
		`event: content_block_delta`, `data: {"index":2,"delta":{"type":"text_delta","text":"Checking."}}`, ``,
		`event: content_block_stop`, `data: {"index":2}`, ``,
		`event: content_block_start`, `data: {"index":3,"content_block":{"type":"tool_use","id":"call-1","name":"probe","input":{}}}`, ``,
		`event: content_block_delta`, `data: {"index":3,"delta":{"type":"input_json_delta","partial_json":"{\"name\":\"gate\"}"}}`, ``,
		`event: content_block_stop`, `data: {"index":3}`, ``,
		`event: message_delta`, `data: {"delta":{"stop_reason":"tool_use"}}`, ``,
		`event: message_stop`, `data: {}`, ``,
	}, "\n")
	var followup map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		if req["stream"] == true {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte(wire))
			return
		}
		followup = req
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"Done."}],"stop_reason":"end_turn"}`))
	}))
	defer server.Close()
	c := NewClaudeClient(server.URL, "test", "test")
	stream, err := c.Stream(context.Background(), []Message{{Role: RoleUser, Content: "Inspect"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var calls []ToolCall
	var text string
	done := false
	for chunk := range stream {
		text += chunk.ContentDelta
		if chunk.ToolCallDelta != nil {
			calls = append(calls, *chunk.ToolCallDelta)
		}
		done = done || chunk.Done
	}
	if !done || text != "Checking." || len(calls) != 1 {
		t.Fatalf("bad stream %q %+v done=%v", text, calls, done)
	}
	_, err = c.Complete(context.Background(), []Message{{Role: RoleUser, Content: "Inspect"}, {Role: RoleAssistant, Content: text, ToolCalls: calls}, {Role: RoleTool, ToolCallID: calls[0].ID, Content: "gate"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	messages := followup["messages"].([]any)
	if len(messages) != 3 {
		t.Fatalf("messages %+v", messages)
	}
	blocks := messages[1].(map[string]any)["content"].([]any)
	if len(blocks) != 4 {
		t.Fatalf("thinking replay missing: %+v", blocks)
	}
	// Decode the outgoing raw blocks to inspect thinking, signatures and tool input.
	raw, err := json.Marshal(blocks)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Private reasoning", "signed-thinking", "encrypted", "gate"} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("replay omitted %q: %s", want, raw)
		}
	}
}

func TestClaudeMixedMalformedToolsNeverCompleteOrCache(t *testing.T) {
	wire := strings.Join([]string{
		`event: content_block_start`, `data: {"index":0,"content_block":{"type":"tool_use","id":"valid","name":"probe","input":{}}}`, ``,
		`event: content_block_delta`, `data: {"index":0,"delta":{"type":"input_json_delta","partial_json":"{\"name\":\"gate\"}"}}`, ``,
		`event: content_block_stop`, `data: {"index":0}`, ``,
		`event: content_block_start`, `data: {"index":1,"content_block":{"type":"tool_use","id":"bad","name":"probe","input":{}}}`, ``,
		`event: content_block_delta`, `data: {"index":1,"delta":{"type":"input_json_delta","partial_json":"{invalid"}}`, ``,
		`event: content_block_stop`, `data: {"index":1}`, ``,
		`event: message_stop`, `data: {}`, ``,
	}, "\n")
	client := &ClaudeClient{}
	ch := make(chan StreamChunk, 10)
	client.parseSSEStream(context.Background(), strings.NewReader(wire), ch)
	close(ch)
	for chunk := range ch {
		if chunk.Done {
			t.Fatal("partially malformed response completed")
		}
	}
	if saved := client.replay.get("valid"); saved != nil {
		t.Fatalf("malformed response cached: %s", saved)
	}
}
