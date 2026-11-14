package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func writeTestCreds(t *testing.T, expiresIn int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "chatgpt-auth.json")
	creds := ChatGPTCredentials{
		ClientID:     "oaiapp_test",
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		ExpiresIn:    expiresIn,
		Scopes:       []string{"openid", ChatGPTPlanScope},
		SavedAt:      time.Now().UTC(),
	}
	if err := WriteChatGPTCredentials(path, creds); err != nil {
		t.Fatal(err)
	}
	return path
}

func sse(w http.ResponseWriter, events ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	for _, e := range events {
		_, _ = fmt.Fprintf(w, "event: x\ndata: %s\n\n", e)
	}
}

const completedEvent = `{"type":"response.completed","response":{"status":"completed","usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}}`

func TestChatGPTCompleteRequestShapeAndToolCall(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access-1" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		sse(w,
			`{"type":"response.output_item.done","item":{"type":"reasoning","id":"rs_1","encrypted_content":"opaque"}}`,
			`{"type":"response.output_item.done","item":{"type":"function_call","id":"fc_1","call_id":"call_1","name":"move_player","namespace":"edda","arguments":"{\"to\":\"inn\"}"}}`,
			completedEvent,
		)
	}))
	defer srv.Close()

	tokens, err := LoadChatGPTTokenStore(writeTestCreds(t, 3600))
	if err != nil {
		t.Fatal(err)
	}
	c := NewChatGPTClient(tokens, "gpt-6.1-terra", "high")
	c.baseURL = srv.URL

	resp, err := c.Complete(context.Background(), []Message{
		{Role: RoleSystem, Content: "You are the GM."},
		{Role: RoleUser, Content: "Go to the inn."},
	}, []Tool{{Name: "move_player", Description: "Move", Parameters: map[string]any{"type": "object"}}})
	if err != nil {
		t.Fatal(err)
	}

	if got["store"] != false || got["stream"] != true {
		t.Errorf("store/stream = %v/%v, want false/true", got["store"], got["stream"])
	}
	if r, _ := got["reasoning"].(map[string]any); r["effort"] != "high" {
		t.Errorf("reasoning = %v", got["reasoning"])
	}
	for _, field := range []string{"temperature", "max_output_tokens", "previous_response_id", "metadata"} {
		if _, ok := got[field]; ok {
			t.Errorf("request has unsupported field %q", field)
		}
	}
	input := got["input"].([]any)
	if role := input[0].(map[string]any)["role"]; role != "developer" {
		t.Errorf("system message sent with role %v, want developer", role)
	}
	ns := got["tools"].([]any)[0].(map[string]any)
	if ns["type"] != "namespace" || ns["tools"].([]any)[0].(map[string]any)["name"] != "move_player" {
		t.Errorf("tools = %v, want one namespace holding move_player", got["tools"])
	}

	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].ID != "call_1" || resp.ToolCalls[0].Arguments["to"] != "inn" {
		t.Fatalf("tool calls = %+v", resp.ToolCalls)
	}
	if resp.FinishReason != "tool_calls" || resp.Usage.TotalTokens != 15 {
		t.Errorf("finish/usage = %q/%+v", resp.FinishReason, resp.Usage)
	}

	// The follow-up must send the reasoning item back before the tool result.
	body, err := c.buildRequest([]Message{
		{Role: RoleUser, Content: "Go to the inn."},
		{Role: RoleAssistant, ToolCalls: resp.ToolCalls},
		{Role: RoleTool, ToolCallID: "call_1", Content: `{"ok":true}`},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var follow struct {
		Input []map[string]any `json:"input"`
	}
	_ = json.Unmarshal(body, &follow)
	types := []string{}
	for _, item := range follow.Input {
		typ, _ := item["type"].(string)
		if typ == "" {
			typ = "message"
		}
		types = append(types, typ)
	}
	if strings.Join(types, ",") != "message,reasoning,function_call,function_call_output" {
		t.Errorf("follow-up input types = %v", types)
	}
	if follow.Input[1]["encrypted_content"] != "opaque" {
		t.Errorf("reasoning item not replayed verbatim: %v", follow.Input[1])
	}
}

func TestChatGPTRebuildsToolCallsFromOtherProvider(t *testing.T) {
	tokens, err := LoadChatGPTTokenStore(writeTestCreds(t, 3600))
	if err != nil {
		t.Fatal(err)
	}
	c := NewChatGPTClient(tokens, "", "")
	body, err := c.buildRequest([]Message{
		{Role: RoleAssistant, Content: "Walking.", ToolCalls: []ToolCall{{ID: "toolu_1", Name: "move_player", Arguments: map[string]any{"to": "inn"}}}},
		{Role: RoleTool, ToolCallID: "toolu_1", Content: "{}"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{`"role":"assistant","content":"Walking."`, `"type":"function_call","call_id":"toolu_1","name":"move_player","namespace":"edda","arguments":"{\"to\":\"inn\"}"`, `"type":"function_call_output","call_id":"toolu_1"`} {
		if !strings.Contains(s, want) {
			t.Errorf("request missing %s\n%s", want, s)
		}
	}
	if strings.Contains(s, `"reasoning"`) {
		t.Errorf("request should not have reasoning: %s", s)
	}
}

func TestChatGPTRefreshesAfter401(t *testing.T) {
	var refreshes int
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/token":
			refreshes++
			_ = r.ParseForm()
			if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "refresh-1" ||
				r.Form.Get("client_id") != "oaiapp_test" || r.Form.Get("resource") != ChatGPTResource {
				t.Errorf("refresh form = %v", r.Form)
			}
			_, _ = io.WriteString(w, `{"access_token":"access-2","refresh_token":"refresh-2","expires_in":3600,"scope":"openid chatgpt.tokens.use.direct"}`)
		case "/responses":
			if r.Header.Get("Authorization") != "Bearer access-2" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"detail":"expired"}`)
				return
			}
			sse(w, `{"type":"response.output_item.done","item":{"type":"message","content":[{"type":"output_text","text":"Hello"}]}}`, completedEvent)
		}
	}))
	defer srv.Close()

	path := writeTestCreds(t, 3600)
	tokens, err := LoadChatGPTTokenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	tokens.tokenURL = srv.URL + "/token"
	c := NewChatGPTClient(tokens, "", "")
	c.baseURL = srv.URL

	resp, err := c.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "Hello" || refreshes != 1 {
		t.Fatalf("content=%q refreshes=%d", resp.Content, refreshes)
	}
	saved, err := ReadChatGPTCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if saved.RefreshToken != "refresh-2" || saved.AccessToken != "access-2" {
		t.Errorf("rotated tokens not saved: %+v", saved)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Errorf("credential file mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestChatGPTUsageLimitIsRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse(w, `{"type":"response.output_text.delta","delta":"Hel"}`,
			`{"type":"response.failed","response":{"status":"failed","error":{"code":"subscription_sharing_usage_limit_exceeded","message":"limit"}}}`)
	}))
	defer srv.Close()
	tokens, err := LoadChatGPTTokenStore(writeTestCreds(t, 3600))
	if err != nil {
		t.Fatal(err)
	}
	c := NewChatGPTClient(tokens, "", "")
	c.baseURL = srv.URL
	_, err = c.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil)
	var rl *ErrRateLimit
	if !errors.As(err, &rl) {
		t.Fatalf("err = %v, want ErrRateLimit", err)
	}
}

func TestChatGPTStreamEndingEarlyIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse(w, `{"type":"response.output_text.delta","delta":"Hel"}`)
	}))
	defer srv.Close()
	tokens, err := LoadChatGPTTokenStore(writeTestCreds(t, 3600))
	if err != nil {
		t.Fatal(err)
	}
	c := NewChatGPTClient(tokens, "", "")
	c.baseURL = srv.URL
	if _, err := c.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil); err == nil {
		t.Fatal("want an error when the stream ends before response.completed")
	}
}

func TestLoadChatGPTTokenStoreRejectsMissingPlanScope(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	if err := WriteChatGPTCredentials(path, ChatGPTCredentials{ClientID: "oaiapp_x", RefreshToken: "r", Scopes: []string{"openid"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadChatGPTTokenStore(path); err == nil || !strings.Contains(err.Error(), ChatGPTPlanScope) {
		t.Fatalf("err = %v, want missing-scope error", err)
	}
}

type stubProvider struct {
	resp  *Response
	err   error
	calls int
}

func (s *stubProvider) Complete(context.Context, []Message, []Tool) (*Response, error) {
	s.calls++
	return s.resp, s.err
}

func (s *stubProvider) Stream(context.Context, []Message, []Tool) (<-chan StreamChunk, error) {
	s.calls++
	return nil, s.err
}

func TestFallbackProvider(t *testing.T) {
	primary := &stubProvider{err: &ErrConnection{Err: errors.New("402 insufficient credits")}}
	secondary := &stubProvider{resp: &Response{Content: "from fallback"}}
	f := &FallbackProvider{Primary: primary, Secondary: secondary, PrimaryName: "openrouter", SecondaryName: "chatgpt"}

	resp, err := f.Complete(context.Background(), nil, nil)
	if err != nil || resp.Content != "from fallback" {
		t.Fatalf("resp=%v err=%v", resp, err)
	}

	primary.err, primary.resp = nil, &Response{Content: "from primary"}
	resp, _ = f.Complete(context.Background(), nil, nil)
	if resp.Content != "from primary" || secondary.calls != 1 {
		t.Fatalf("healthy primary should not reach fallback: %q, fallback calls %d", resp.Content, secondary.calls)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	primary.err = context.Canceled
	if _, err := f.Complete(ctx, nil, nil); !errors.Is(err, context.Canceled) || secondary.calls != 1 {
		t.Fatalf("canceled request should not fall back: err=%v fallback calls %d", err, secondary.calls)
	}
}
